package routedinformer

import (
	"time"

	"go.uber.org/zap"
)

// Option configures RoutedInformer.
type Option func(*RoutedInformer)

// WithResyncPeriod sets the resync period for the informer. Default is 0 (no periodic resync).
func WithResyncPeriod(d time.Duration) Option {
	return func(i *RoutedInformer) {
		i.resyncPeriod = d
	}
}

// WithLogger sets the logger. If not provided, a no-op logger is used.
func WithLogger(logger *zap.Logger) Option {
	return func(i *RoutedInformer) {
		if logger != nil {
			i.logger = logger
		}
	}
}

// WithTaskTimeout sets the timeout for each handler invocation. Handler receives a context
// that is cancelled after this duration. Default is 0 (no timeout).
func WithTaskTimeout(d time.Duration) Option {
	return func(i *RoutedInformer) {
		i.taskTimeout = d
	}
}

// WithFilter adds a filter to the stack. Filters are applied in order; all must pass
// for the event to be routed to the handler.
func WithFilter(f Filter) Option {
	return func(i *RoutedInformer) {
		if f != nil {
			i.filters = append(i.filters, f)
		}
	}
}
