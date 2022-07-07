package coreutil

import (
	"fmt"
)

// RecoverPanic is an utility function to handler the recovery from a panic.
func RecoverPanic() {
	if r := recover(); r != nil {
		// TODO: implement recovered panic output
		fmt.Println("Recovering from panic:", r)
	}
}
