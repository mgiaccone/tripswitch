package retrier

import (
	"fmt"

	"github.com/mgiaccone/tripswitch/internal/coreutil"
)

var (
	// ErrPanicRecovered is a panic recovered error.
	ErrPanicRecovered = coreutil.ErrPanicRecovered
)

// ProtectedFunc represents the function to be protected by the circuit breaker.
type ProtectedFunc[T any] func() (T, error)

// Unrecoverable is a short-hand function to wrap an error into a UnrecoverableError.
// When used, the retrier will stop retrying and just return the original error.
func Unrecoverable(err error) error {
	return &UnrecoverableError{Err: err}
}

// UnrecoverableError represents and error that signal the retries that the function execution should be retried.
type UnrecoverableError struct {
	Err error
}

// Error implements the Error interface.
func (e *UnrecoverableError) Error() string {
	return fmt.Sprintf("retriable: %s", e.Err)
}

// Unwrap implements the Unwrap interface.
func (e *UnrecoverableError) Unwrap() error {
	return e.Err
}
