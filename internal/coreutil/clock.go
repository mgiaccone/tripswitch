package coreutil

import (
	"time"
)

type Clock interface {
	Now() time.Time
}

type SystemClock struct {
}

func NewSystemClock() *SystemClock {
	return &SystemClock{}
}

func (c *SystemClock) Now() time.Time {
	return time.Now()
}
