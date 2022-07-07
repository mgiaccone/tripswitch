package retrier

import (
	"errors"
	"time"

	"github.com/mgiaccone/tripswitch/breaker"
	"github.com/mgiaccone/tripswitch/internal/coreutil"
)

// ConstantRetrier is an implementation of a retrier that will wait a constant amount of time
// before retrying the execution up to the maximum amount of retries.
type ConstantRetrier[T any] struct {
	clock      coreutil.Clock
	delay      time.Duration
	maxRetries int
}

// NewConstantRetrier creates a new instance of a constant retrier.
func NewConstantRetrier[T any](delay time.Duration, maxRetries int) *ConstantRetrier[T] {
	return &ConstantRetrier[T]{
		delay:      delay,
		maxRetries: maxRetries,
	}
}

// Do implement the ProtectedFunc interface.
func (r *ConstantRetrier[T]) Do(fn ProtectedFunc[T]) (res T, err error) {
	err = ErrPanicRecovered
	defer coreutil.RecoverPanic()

	_ = r.clock

	// TODO: missing implementation

	res, err = fn()
	if errors.Is(err, breaker.ErrCircuitOpen) {
		return res, breaker.ErrCircuitOpen
	}

	return
}
