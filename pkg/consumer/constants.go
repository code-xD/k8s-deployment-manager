package consumer

import "time"

// contextKey is a private type for context keys to avoid collisions.
type contextKey string

const (
	headerContextKey       contextKey = "nats.headers"
	lastAttemptContextKey  contextKey = "consumer.last_attempt"
	defaultTaskTimeout                = time.Minute
	defaultRetryCount                 = 1
)
