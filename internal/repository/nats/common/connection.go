package common

import (
	"errors"
	"fmt"
	"strings"

	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// DefaultStreamName is the JetStream stream used for deployment request/update subjects
const DefaultStreamName = "DEPLOYMENTS"

// NATS holds the NATS connection
// This is shared infrastructure used by all NATS repositories
type NATS struct {
	Conn   *nats.Conn
	JS     nats.JetStreamContext
	logger *zap.Logger
}

// NewNATS creates a new NATS connection
func NewNATS(cfg *dto.NatsConfig, log *zap.Logger) (*NATS, error) {
	conn, err := nats.Connect(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// Get JetStream context
	js, err := conn.JetStream()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to get JetStream context: %w", err)
	}

	natsInstance := &NATS{
		Conn:   conn,
		JS:     js,
		logger: log,
	}

	log.Info("Connected to NATS", zap.String("url", cfg.URL))
	return natsInstance, nil
}

// EnsureStream creates a JetStream stream with the given name and subjects if it does not exist.
// Idempotent: if the stream already exists (including name in use), no error is returned.
func (n *NATS) EnsureStream(streamName string, subjects []string) error {
	if streamName == "" || len(subjects) == 0 {
		return fmt.Errorf("stream name and at least one subject are required")
	}
	_, err := n.JS.AddStream(&nats.StreamConfig{
		Name:     streamName,
		Subjects: subjects,
	})
	if err != nil {
		if errors.Is(err, nats.ErrStreamNameAlreadyInUse) || strings.Contains(err.Error(), "stream name already in use") {
			n.logger.Debug("JetStream stream already exists", zap.String("stream", streamName))
			return nil
		}
		return fmt.Errorf("failed to ensure JetStream stream %s: %w", streamName, err)
	}
	n.logger.Info("JetStream stream created", zap.String("stream", streamName), zap.Strings("subjects", subjects))
	return nil
}

// Close closes the NATS connection
func (n *NATS) Close() {
	if n.Conn != nil {
		n.Conn.Close()
		n.logger.Info("NATS connection closed")
	}
}
