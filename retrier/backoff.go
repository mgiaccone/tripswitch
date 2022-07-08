package retrier

import (
	"errors"

	"github.com/mgiaccone/tripswitch/breaker"
	"github.com/mgiaccone/tripswitch/internal/coreutil"
)

// BackoffRetrier is an implementation of a retrier that will wait an exponential amount of time
// before retrying the execution up to the maximum amount of time.
type BackoffRetrier[T any] struct {
}

// Do implement the ProtectedFunc interface.
func (r *BackoffRetrier[T]) Do(fn ProtectedFunc[T]) (res T, err error) {
	err = ErrPanicRecovered
	defer coreutil.RecoverPanic()

	// TODO: missing implementation

	_ = err

	res, err = fn()
	if errors.Is(err, breaker.ErrCircuitOpen) {
		return res, breaker.ErrCircuitOpen
	}

	return
}
