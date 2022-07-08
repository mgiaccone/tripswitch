package breaker

import (
	"time"
)

// Option represents a functional option applicable to a circuit breaker.
type Option func(cfg *config)

type config struct {
	failThreshold    int32
	stateChangeFunc  StateChangeFunc
	successThreshold int32
	waitInterval     time.Duration
}

func newConfig(opts ...Option) config {
	cfg := config{
		stateChangeFunc: func(oldState, newState CircuitState) {
			// nop by default
		},
		failThreshold:    _defaultFailThreshold,
		successThreshold: _defaultSuccessThreshold,
		waitInterval:     _defaultWaitInterval,
	}
	cfg.applyOpts(opts...)

	return cfg
}

func (c *config) applyOpts(opts ...Option) {
	for _, apply := range opts {
		apply(c)
	}
}

// WithFailThreshold overrides the default value for the number of failes
// executions required to trip the circuit breaker to its CircuitOpen state.
func WithFailThreshold(threshold int) Option {
	return func(cfg *config) {
		cfg.failThreshold = int32(threshold)
	}
}

// WithStateChangeFunc attaches a function that will receive notifications
// of circuit breaker state changes.
func WithStateChangeFunc(fn StateChangeFunc) Option {
	return func(cfg *config) {
		cfg.stateChangeFunc = fn
	}
}

// WithSuccessThreshold overrides the default value for the number of successful
// executions required to restore the circuit breaker to its CircuitClosed state.
func WithSuccessThreshold(threshold int) Option {
	return func(cfg *config) {
		cfg.successThreshold = int32(threshold)
	}
}

// WithWaitInterval overrides the default value for the time the circuit breaker
// will wait before entering the CircuitHalfOpen state.
func WithWaitInterval(interval time.Duration) Option {
	return func(cfg *config) {
		cfg.waitInterval = interval
	}
}
