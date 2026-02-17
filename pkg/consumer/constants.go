package consumer

import "time"

// contextKey is a private type for context keys to avoid collisions.
type contextKey string

const (
	headerContextKey   contextKey = "nats.headers"
	defaultTaskTimeout            = time.Minute
	defaultRetryCount             = 1
)
