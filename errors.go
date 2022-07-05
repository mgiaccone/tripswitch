package tripswitch

import (
	"errors"
	"fmt"
)

var (
	ErrCircuitOpen     = errors.New("circuit open")
	ErrCircuitConflict = errors.New("circuit already exists")
	ErrNameRequired    = errors.New("circuit name required")
	ErrRetrierRequired = errors.New("retrier required")
)

type RetriableError struct {
	Err error
}

func Retriable(err error) error {
	return &RetriableError{Err: err}
}

func (e *RetriableError) Error() string {
	return fmt.Sprintf("retriable: %s", e.Err)
}

func (e *RetriableError) Unwrap() error {
	return e.Err
}
