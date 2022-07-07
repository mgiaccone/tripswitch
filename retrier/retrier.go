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

// Retriable is a short-hand function to wrap an error into a RetriableError.
func Retriable(err error) error {
	return &RetriableError{Err: err}
}

// RetriableError represents and error that signal the retries that the function execution should be retried.
type RetriableError struct {
	Err error
}

// Error implements the Error interface.
func (e *RetriableError) Error() string {
	return fmt.Sprintf("retriable: %s", e.Err)
}

// Unwrap implements the Unwrap interface.
func (e *RetriableError) Unwrap() error {
	return e.Err
}
