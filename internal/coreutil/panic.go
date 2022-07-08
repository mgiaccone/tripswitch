package coreutil

import (
	"errors"
	"fmt"
)

var (
	// ErrPanicRecovered is a panic recovered error.
	ErrPanicRecovered = errors.New("panic recovered")
)

// MustErr trigger a panic if it detects an error.
func MustErr(err error) {
	if err != nil {
		panic(err)
	}
}

// MustResErr trigger a panic if it detects an error.
func MustResErr[T any](_ T, err error) {
	if err != nil {
		panic(err)
	}
}

// RecoverPanic is an utility function to handler the recovery from a panic.
func RecoverPanic() {
	if r := recover(); r != nil {
		// TODO: implement recovered panic output
		fmt.Println("Recovering from panic:", r)
	}
}
