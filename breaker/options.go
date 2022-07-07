package breaker

import (
	"time"
)

// CircuitBreakeOption represents a functional option applicable
// to a circuit breaker.
type CircuitBreakeOption func(cfg *config)

type config struct {
	failThreshold    int32
	stateChangeFunc  StateChangeFunc
	successThreshold int32
	waitInterval     time.Duration
}

func (c *config) applyOpts(opts ...CircuitBreakeOption) {
	for _, apply := range opts {
		apply(c)
	}
}

// WithFailThreshold overrides the default value for the number of failes
// executions required to trip the circuit breaker to its CircuitOpen state.
func WithFailThreshold(threshold int) CircuitBreakeOption {
	return func(cfg *config) {
		cfg.failThreshold = int32(threshold)
	}
}

// WithStateChangeFunc attaches a function that will receive notifications
// of circuit breaker state changes.
func WithStateChangeFunc(fn StateChangeFunc) CircuitBreakeOption {
	return func(cfg *config) {
		cfg.stateChangeFunc = fn
	}
}

// WithSuccessThreshold overrides the default value for the number of successful
// executions required to restore the circuit breaker to its CircuitClosed state.
func WithSuccessThreshold(threshold int) CircuitBreakeOption {
	return func(cfg *config) {
		cfg.successThreshold = int32(threshold)
	}
}

// WithWaitInterval overrides the default value for the time the circuit breaker
// will wait before entering the CircuitHalfOpen state.
func WithWaitInterval(interval time.Duration) CircuitBreakeOption {
	return func(cfg *config) {
		cfg.waitInterval = interval
	}
}
