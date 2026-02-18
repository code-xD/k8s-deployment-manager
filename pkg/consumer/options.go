package consumer

import "time"

// TaskTimeout sets the per-message task timeout. Handler will be run with
// a context that times out after this duration. Default: 1 minute.
func TaskTimeout(d time.Duration) OptionFunc {
	return func(c *RouteConfig) {
		if d > 0 {
			c.TaskTimeout = d
		}
	}
}

// RetryCount sets how many times to retry the handler on error before NAK.
// Default: 1 (no retries beyond the first attempt).
func RetryCount(n int) OptionFunc {
	return func(c *RouteConfig) {
		if n >= 0 {
			c.RetryCount = n
		}
	}
}
