package consumer

import (
	"context"
	"fmt"
	"sync"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// NATSConsumer subscribes to NATS JetStream subjects with queue groups,
// runs handlers with configurable timeout and retries, and supports graceful shutdown.
type NATSConsumer struct {
	js     nats.JetStreamContext
	conn   *nats.Conn
	logger *zap.Logger

	mu     sync.Mutex
	wg     sync.WaitGroup
	routes []*RouteConfig
	subs   []*nats.Subscription
}

// NewNATSConsumer creates a new NATS consumer. js and conn must be non-nil;
// the consumer will drain conn on Shutdown.
func NewNATSConsumer(js nats.JetStreamContext, conn *nats.Conn, logger *zap.Logger) *NATSConsumer {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &NATSConsumer{
		js:     js,
		conn:   conn,
		logger: logger,
		routes: nil,
		subs:   nil,
	}
}

// Route registers a route: when Run() is called, the consumer will subscribe to
// channel with the given queue group and invoke handler for each message.
// Options (e.g. TaskTimeout, RetryCount) are applied to this route only.
// Route must be called before Run().
func (c *NATSConsumer) Route(channel, queueGroupName string, handler HandlerFunc, opts ...OptionFunc) {
	cfg := &RouteConfig{
		Channel:     channel,
		QueueGroup:  queueGroupName,
		Handler:     handler,
		TaskTimeout: defaultTaskTimeout,
		RetryCount:  defaultRetryCount,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	c.mu.Lock()
	c.routes = append(c.routes, cfg)
	c.mu.Unlock()
}

// Run subscribes to all registered routes and starts consuming. Each handler
// is run with a context that includes message headers and a timeout (TaskTimeout).
// Run blocks until Shutdown is called or the connection is closed.
func (c *NATSConsumer) Run() error {
	c.mu.Lock()
	routes := make([]*RouteConfig, len(c.routes))
	copy(routes, c.routes)
	c.mu.Unlock()

	if len(routes) == 0 {
		return fmt.Errorf("consumer: no routes registered; call Route before Run")
	}

	for _, r := range routes {
		sub, err := c.subscribe(r)
		if err != nil {
			// best-effort cleanup of already-subscribed
			_ = c.drainSubs()
			return err
		}
		c.mu.Lock()
		c.subs = append(c.subs, sub)
		c.mu.Unlock()

		c.logger.Info("Subscribed to channel with queue group",
			zap.String("channel", r.Channel),
			zap.String("queue_group", r.QueueGroup),
		)
	}

	return nil
}

// subscribe creates a single subscription for the given route and returns it.
func (c *NATSConsumer) subscribe(r *RouteConfig) (*nats.Subscription, error) {
	handler := c.wrapHandler(r)
	sub, err := c.js.QueueSubscribe(r.Channel, r.QueueGroup, func(msg *nats.Msg) {
		c.logger.Debug("Message received",
			zap.String("subject", r.Channel),
			zap.String("queue_group", r.QueueGroup),
			zap.Int("payload_size", len(msg.Data)),
		)
		handler(msg)
	})
	if err != nil {
		return nil, fmt.Errorf("subscribe to %s (queue %s): %w", r.Channel, r.QueueGroup, err)
	}
	return sub, nil
}

// wrapHandler returns a function that runs the route handler with timeout and
// retries, injects headers into context, and Ack/Nak the message.
func (c *NATSConsumer) wrapHandler(r *RouteConfig) func(*nats.Msg) {
	return func(msg *nats.Msg) {
		c.wg.Add(1)
		defer c.wg.Done()

		ctx := contextWithHeaders(context.Background(), msg.Header)
		ctx, cancel := context.WithTimeout(ctx, r.TaskTimeout)
		defer cancel()

		m := &Message{Data: msg.Data, RawMsg: msg}
		var lastErr error
		for attempt := 0; attempt <= r.RetryCount; attempt++ {
			// If this is the last allowed attempt, flag the context that it is the last attempt.
			if attempt == r.RetryCount {
				ctx = context.WithValue(ctx, lastAttemptContextKey, true)
			}

			lastErr = r.Handler(ctx, m)
			if lastErr == nil {
				if err := msg.Ack(); err != nil {
					c.logger.Error("Failed to ack message",
						zap.String("channel", r.Channel),
						zap.String("queue_group", r.QueueGroup),
						zap.Error(err),
					)
				}
				return
			}
			if attempt < r.RetryCount {
				c.logger.Debug("Handler error, retrying",
					zap.String("channel", r.Channel),
					zap.Int("attempt", attempt+1),
					zap.Error(lastErr),
				)
			}
		}

		c.logger.Error("Error processing message after retries",
			zap.String("channel", r.Channel),
			zap.String("queue_group", r.QueueGroup),
			zap.Error(lastErr),
		)
		// Note: In case of retry failures, we ACK the message. In future, implement DLQ shift here.
		if err := msg.Ack(); err != nil {
			c.logger.Error("Failed to ack message after retries", zap.Error(err))
		}
	}
}

// Shutdown gracefully stops the consumer: it drains all subscriptions (no new
// messages), waits for in-flight handler tasks to complete (via WaitGroup), then
// drains the NATS connection. If ctx is cancelled before the wait completes,
// Shutdown proceeds to close the connection and returns ctx.Err().
func (c *NATSConsumer) Shutdown(ctx context.Context) error {
	if err := c.drainSubs(); err != nil {
		c.logger.Warn("Error draining subscriptions", zap.Error(err))
	}

	// Wait for in-flight handlers to finish, or for ctx to be cancelled.
	waitDone := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(waitDone)
	}()
	select {
	case <-ctx.Done():
		c.logger.Warn("Shutdown context cancelled while waiting for in-flight tasks", zap.Error(ctx.Err()))
	case <-waitDone:
	}

	done := make(chan struct{})
	go func() {
		if c.conn != nil {
			_ = c.conn.Drain()
		}
		close(done)
	}()

	select {
	case <-ctx.Done():
		if c.conn != nil {
			c.conn.Close()
		}
		return ctx.Err()
	case <-done:
		return nil
	}
}

func (c *NATSConsumer) drainSubs() error {
	c.mu.Lock()
	subs := make([]*nats.Subscription, len(c.subs))
	copy(subs, c.subs)
	c.subs = nil
	c.mu.Unlock()

	var err error
	for _, sub := range subs {
		if sub == nil {
			continue
		}
		if drainErr := sub.Drain(); drainErr != nil {
			if err == nil {
				err = drainErr
			}
			c.logger.Warn("Error draining subscription", zap.Error(drainErr))
		}
	}
	return err
}
