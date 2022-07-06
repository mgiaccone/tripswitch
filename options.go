package tripswitch

import (
	"time"
)

// CircuitOption represents a functional option applicable to a circuit.
type CircuitOption func(cfg *circuitConfig)

type circuitConfig struct {
	failThreshold    int32
	stateChangeFunc  StateChangeFunc
	successThreshold int32
	waitInterval     time.Duration
}

func (c *circuitConfig) applyOpts(opts ...CircuitOption) {
	for _, apply := range opts {
		apply(c)
	}
}

func WithFailThreshold(threshold int) CircuitOption {
	return func(cfg *circuitConfig) {
		cfg.failThreshold = int32(threshold)
	}
}

func WithStateChangeFunc(fn StateChangeFunc) CircuitOption {
	return func(cfg *circuitConfig) {
		cfg.stateChangeFunc = fn
	}
}

func WithSuccessThreshold(threshold int) CircuitOption {
	return func(cfg *circuitConfig) {
		cfg.successThreshold = int32(threshold)
	}
}

func WithWaitInterval(interval time.Duration) CircuitOption {
	return func(cfg *circuitConfig) {
		cfg.waitInterval = interval
	}
}
