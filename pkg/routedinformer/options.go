package routedinformer

import (
	"time"

	"go.uber.org/zap"
	"k8s.io/client-go/informers/internalinterfaces"
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

// WithTweakListOptions sets the list-options tweak applied when building the informer (before Run).
// Use this for server-side filtering (e.g. label selector) instead of in-process Filter.
func WithTweakListOptions(tweak internalinterfaces.TweakListOptionsFunc) Option {
	return func(i *RoutedInformer) {
		i.tweakListOptions = tweak
	}
}
