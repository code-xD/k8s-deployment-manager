package common

import (
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// Producer handles publishing messages to NATS
type Producer struct {
	js     nats.JetStreamContext
	logger *zap.Logger
}

// NewProducer creates a new NATS producer
func NewProducer(nats *NATS) *Producer {
	return &Producer{
		js:     nats.JS,
		logger: nats.logger,
	}
}

// Publish publishes a message to a NATS subject
func (p *Producer) Publish(subject string, data interface{}) error {
	// Marshal data to JSON
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Publish message
	ack, err := p.js.Publish(subject, payload)
	if err != nil {
		return fmt.Errorf("failed to publish message to subject %s: %w", subject, err)
	}

	p.logger.Info("Message published successfully",
		zap.String("subject", subject),
		zap.Int("payload_size", len(payload)),
		zap.Any("nats_ack", map[string]interface{}{
			"stream":    ack.Stream,
			"sequence":  ack.Sequence,
			"domain":    ack.Domain,
			"duplicate": ack.Duplicate,
		}),
	)

	return nil
}

// PublishWithHeader publishes a message to a NATS subject with metadata headers
func (p *Producer) PublishWithHeader(subject string, data interface{}, header nats.Header) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	msg := &nats.Msg{
		Subject: subject,
		Data:    payload,
		Header:  header,
	}

	ack, err := p.js.PublishMsg(msg)
	if err != nil {
		return fmt.Errorf("failed to publish message to subject %s: %w", subject, err)
	}

	p.logger.Info("Message published successfully with headers",
		zap.String("subject", subject),
		zap.Int("payload_size", len(payload)),
		zap.Any("nats_ack", map[string]interface{}{
			"stream":    ack.Stream,
			"sequence":  ack.Sequence,
			"domain":    ack.Domain,
			"duplicate": ack.Duplicate,
		}),
	)

	return nil
}
