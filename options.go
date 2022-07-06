package tripswitch

import (
	"fmt"
	"strings"
	"time"
)

// CircuitBreakerOption represents a functional option applicable to the circuit breaker.
type CircuitBreakerOption func(cfg *config) error

type config struct {
	circuits                map[string]*circuit
	defaultFailThreshold    int32
	defaultSuccessThreshold int32
	defaultWaitInterval     time.Duration
	stateChangeFunc         StateChangeFunc
}

func (c *config) apply(opts ...CircuitBreakerOption) error {
	for _, apply := range opts {
		if err := apply(c); err != nil {
			return err
		}
	}
	return nil
}

// WithCircuit set custom options for a named circuit.
func WithCircuit(name string, failThreshold, successThreshold int, waitInterval time.Duration) CircuitBreakerOption {
	return func(cfg *config) error {
		if name = strings.TrimSpace(name); name == "" {
			return ErrNameRequired
		}

		if cfg.circuits == nil {
			cfg.circuits = make(map[string]*circuit)
		}

		c, exists := cfg.circuits[name]
		if exists {
			return fmt.Errorf("%q: %w", name, ErrCircuitConflict)
		}

		c = &circuit{
			name:             name,
			state:            Closed,
			failThreshold:    int32(failThreshold),
			successThreshold: int32(successThreshold),
			waitInterval:     waitInterval,
		}

		cfg.circuits[name] = c
		return nil
	}
}

func WithFailThreshold(threshold int) CircuitBreakerOption {
	return func(cfg *config) error {
		cfg.defaultFailThreshold = int32(threshold)
		return nil
	}
}

func WithSuccessThreshold(threshold int) CircuitBreakerOption {
	return func(cfg *config) error {
		cfg.defaultSuccessThreshold = int32(threshold)
		return nil
	}
}

func WithWaitInterval(interval time.Duration) CircuitBreakerOption {
	return func(cfg *config) error {
		cfg.defaultWaitInterval = interval
		return nil
	}
}

func WithStateChangeFunc(fn StateChangeFunc) CircuitBreakerOption {
	return func(cfg *config) error {
		cfg.stateChangeFunc = fn
		return nil
	}
}
