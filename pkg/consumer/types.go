package consumer

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
)

// Message is the type passed to the handler. Headers are available via context
// using the HeadersFromContext helper.
type Message struct {
	Data   []byte
	RawMsg *nats.Msg // optional: for Ack/Nak when handler needs to ack explicitly
}

// HandlerFunc is the signature for a route handler. The context contains
// request-scoped values including NATS headers (see HeadersFromContext).
// msg.Data is the message payload; msg.RawMsg is set when Run wraps the handler
// so the handler may call Ack/Nak if needed.
type HandlerFunc func(ctx context.Context, msg *Message) error

// OptionFunc applies options to a route config.
type OptionFunc func(*RouteConfig)

// RouteConfig holds the configuration for a single route, stored when Route() is called.
type RouteConfig struct {
	Channel       string
	QueueGroup    string
	Handler       HandlerFunc
	TaskTimeout   time.Duration
	RetryCount    int
}

// HeadersFromContext returns the NATS message headers stored in the context,
// or nil if not set.
func HeadersFromContext(ctx context.Context) nats.Header {
	v := ctx.Value(headerContextKey)
	if v == nil {
		return nil
	}
	if h, ok := v.(nats.Header); ok {
		return h
	}
	return nil
}

// contextWithHeaders returns a new context with the given headers.
func contextWithHeaders(ctx context.Context, h nats.Header) context.Context {
	if h == nil {
		return ctx
	}
	return context.WithValue(ctx, headerContextKey, nats.Header(h))
}
