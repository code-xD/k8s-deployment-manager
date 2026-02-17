package common

import (
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// Consumer handles consuming messages from NATS
type Consumer struct {
	js     nats.JetStreamContext
	logger *zap.Logger
}

// NewConsumer creates a new NATS consumer
func NewConsumer(nats *NATS) *Consumer {
	return &Consumer{
		js:     nats.JS,
		logger: nats.logger,
	}
}

// SubscribeWithQueue subscribes to a NATS subject with a queue group
func (c *Consumer) SubscribeWithQueue(
	subject string,
	queueGroup string,
	handler func(msg *nats.Msg) error,
) (*nats.Subscription, error) {
	sub, err := c.js.QueueSubscribe(subject, queueGroup, func(msg *nats.Msg) {
		c.logger.Info("Message received",
			zap.String("subject", subject),
			zap.String("queue_group", queueGroup),
			zap.Int("payload_size", len(msg.Data)),
		)

		if err := handler(msg); err != nil {
			c.logger.Error("Error processing message",
				zap.String("subject", subject),
				zap.String("queue_group", queueGroup),
				zap.Error(err),
			)
			// Optionally nack the message if processing fails
			if err := msg.Nak(); err != nil {
				c.logger.Error("Failed to nack message", zap.Error(err))
			}
			return
		}

		// Acknowledge the message after successful processing
		if err := msg.Ack(); err != nil {
			c.logger.Error("Failed to ack message", zap.Error(err))
		}
	})
	if err != nil {
		return nil, fmt.Errorf(
			"failed to subscribe to subject %s with queue group %s: %w",
			subject,
			queueGroup,
			err,
		)
	}

	c.logger.Info("Subscribed to subject with queue group",
		zap.String("subject", subject),
		zap.String("queue_group", queueGroup),
	)

	return sub, nil
}

// UnmarshalMessage unmarshals a NATS message into the provided type
func UnmarshalMessage[T any](msg *nats.Msg) (*T, error) {
	var data T
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}
	return &data, nil
}
