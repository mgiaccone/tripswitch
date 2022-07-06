package tripswitch

import (
	"errors"
	"fmt"
)

var (
	// ErrCircuitOpen
	ErrCircuitOpen = errors.New("circuit open")
)

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
