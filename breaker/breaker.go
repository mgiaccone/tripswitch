package breaker

import (
	"errors"
	"sync/atomic"
	"time"

	"github.com/mgiaccone/tripswitch/internal/coreutil"
)

const (
	_defaultFailThreshold    = int32(3)
	_defaultSuccessThreshold = int32(3)
	_defaultWaitInterval     = 10 * time.Second
)

var (
	// ErrCircuitOpen is an open circuit error.
	ErrCircuitOpen = errors.New("circuit open")

	// ErrPanicRecovered is a panic recovered error.
	ErrPanicRecovered = coreutil.ErrPanicRecovered
)

// ProtectedFunc represents the function to be protected by the circuit breaker.
type ProtectedFunc[T any] func() (T, error)

// Retrier is the interface representing a retrier.
type Retrier[T any] interface {
	Do(fn ProtectedFunc[T]) (T, error)
}

// CircuitState represents the state of the circuit breaker.
type CircuitState int32

// Enumeration of circuit breaker states.
const (
	CircuitClosed CircuitState = iota << 1
	CircuitHalfOpen
	CircuitOpen
)

// String implements the Stringer interface.
func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitHalfOpen:
		return "half-open"
	case CircuitOpen:
		return "open"
	}

	return "undefined"
}

// StateChangeFunc represents the function to handle state change notifications.
type StateChangeFunc func(oldState, newState CircuitState)

type restoreCircuitEvent struct{}

type stateChangeEvent struct {
	oldState CircuitState
	newState CircuitState
}

type notifyStateChangeFunc func(oldState, newState CircuitState)

type notifyRecoverFunc func()

// CircuitBreaker is the struct implementing the circuit breaker logic.
type CircuitBreaker[T any] struct {
	failCount           int32
	failThreshold       int32
	notifyStateChangeCh chan stateChangeEvent
	state               CircuitState
	successCount        int32
	successThreshold    int32
	restoreCircuitCh    chan restoreCircuitEvent
	retrier             Retrier[T]
	stateChangeFunc     StateChangeFunc
	waitInterval        time.Duration

	// these are used as test hooks
	notifyStateChangeFn notifyStateChangeFunc
	scheduleRecoverFn   notifyRecoverFunc
}

// NewCircuitBreaker creates a new instance of a circuit breaker.
func NewCircuitBreaker[T any](opts ...Option) *CircuitBreaker[T] {
	return NewCircuitBreakerWithRetrier[T](&nopRetrier[T]{}, opts...)
}

// NewCircuitBreakerWithRetrier creates a new instance of a circuit breaker .
func NewCircuitBreakerWithRetrier[T any](retrier Retrier[T], opts ...Option) *CircuitBreaker[T] {
	cfg := config{
		stateChangeFunc: func(oldState, newState CircuitState) {
			// nop by default
		},
		failThreshold:    _defaultFailThreshold,
		successThreshold: _defaultSuccessThreshold,
		waitInterval:     _defaultWaitInterval,
	}
	cfg.applyOpts(opts...)

	cb := CircuitBreaker[T]{
		failThreshold:       cfg.failThreshold,
		notifyStateChangeCh: make(chan stateChangeEvent),
		restoreCircuitCh:    make(chan restoreCircuitEvent),
		retrier:             retrier,
		state:               CircuitClosed,
		stateChangeFunc:     cfg.stateChangeFunc,
		successThreshold:    cfg.successThreshold,
		waitInterval:        cfg.waitInterval,
	}
	cb.scheduleRecoverFn = cb.scheduleRestore
	cb.notifyStateChangeFn = cb.notifyStateChange

	go cb.processEvents()

	return &cb
}

// Do wraps a function execution with the circuit breaker.
func (cb *CircuitBreaker[T]) Do(fn ProtectedFunc[T]) (res T, err error) {
	err = ErrPanicRecovered
	defer coreutil.RecoverPanic()

	// fails immediately if the circuit state is CircuitOpen
	if CircuitState(atomic.LoadInt32((*int32)(&cb.state))) == CircuitOpen {
		return res, ErrCircuitOpen
	}

	res, err = wrapWithRetrier(cb.retrier, fn)()
	if err != nil {
		cb.recordFailure()
		return
	}
	cb.recordSuccess()

	return
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker[T]) State() CircuitState {
	return CircuitState(atomic.LoadInt32((*int32)(&cb.state)))
}

// processEvents handles all the internal events.
func (cb *CircuitBreaker[T]) processEvents() {
	for {
		select {
		case msg := <-cb.notifyStateChangeCh:
			cb.stateChangeFunc(msg.oldState, msg.newState)
		case <-cb.restoreCircuitCh:
			go cb.restoreCircuit()
		}
	}
}

// notifyStateChange publishes a state change.
func (cb *CircuitBreaker[T]) notifyStateChange(oldState, newState CircuitState) {
	cb.notifyStateChangeCh <- stateChangeEvent{
		oldState: oldState,
		newState: newState,
	}
}

// scheduleRestore publishes a circuit recover request.
func (cb *CircuitBreaker[T]) scheduleRestore() {
	cb.restoreCircuitCh <- restoreCircuitEvent{}
}

// recordFailure handles a failed function execution.
// If the current state is CircuitClosed and the failure counter reached the threshold, it will set the circuit breaker state to CircuitOpen.
// Otherwise, it resets the success counter and sets the state to CircuitOpen when the current state is CircuitHalfOpen.
func (cb *CircuitBreaker[T]) recordFailure() {
	switch CircuitState(atomic.LoadInt32((*int32)(&cb.state))) {
	case CircuitOpen:
		// TODO: update stats
	case CircuitClosed:
		// TODO: update stats
		if atomic.AddInt32(&cb.failCount, 1) < cb.failThreshold {
			return
		}

		if atomic.CompareAndSwapInt32((*int32)(&cb.state), int32(CircuitClosed), int32(CircuitOpen)) {
			cb.notifyStateChangeFn(CircuitClosed, CircuitOpen)
			cb.scheduleRecoverFn()
		}
	case CircuitHalfOpen:
		// TODO: update stats
		if atomic.CompareAndSwapInt32((*int32)(&cb.state), int32(CircuitHalfOpen), int32(CircuitOpen)) {
			atomic.AddInt32(&cb.failCount, 1)
			atomic.StoreInt32(&cb.successCount, 0)

			cb.notifyStateChangeFn(CircuitHalfOpen, CircuitOpen)
			cb.scheduleRecoverFn()
		}
	}
}

// recordSuccess handles a successful function execution.
// If the current state is CircuitHalfOpen, it resets the circuit breaker.
func (cb *CircuitBreaker[T]) recordSuccess() {
	switch CircuitState(atomic.LoadInt32((*int32)(&cb.state))) {
	case CircuitOpen:
		// TODO: update stats
	case CircuitClosed:
		// TODO: update stats
	case CircuitHalfOpen:
		// TODO: update stats
		if atomic.AddInt32(&cb.successCount, 1) < cb.successThreshold {
			return
		}

		if atomic.CompareAndSwapInt32((*int32)(&cb.state), int32(CircuitHalfOpen), int32(CircuitClosed)) {
			atomic.StoreInt32(&cb.failCount, 0)
			atomic.StoreInt32(&cb.successCount, 0)

			cb.notifyStateChangeFn(CircuitHalfOpen, CircuitClosed)
		}
	}
}

// restoreCircuit waits for the configured interval before attempting to reopen the circuit.
// If the current state is CircuitOpen, it sets a timer to setting the state to CircuitHalfOpen.
func (cb *CircuitBreaker[T]) restoreCircuit() {
	t := time.NewTimer(cb.waitInterval)
	defer t.Stop()

	<-t.C

	if atomic.CompareAndSwapInt32((*int32)(&cb.state), int32(CircuitOpen), int32(CircuitHalfOpen)) {
		atomic.StoreInt32(&cb.failCount, 0)
		atomic.StoreInt32(&cb.successCount, 0)

		cb.notifyStateChangeFn(CircuitOpen, CircuitHalfOpen)
	}
	// TODO: update stats
}

type nopRetrier[T any] struct {
}

// Do implement the ProtectedFunc interface.
// nolint:revive
func (r *nopRetrier[T]) Do(fn ProtectedFunc[T]) (T, error) {
	return fn()
}

func wrapWithRetrier[T any](r Retrier[T], fn ProtectedFunc[T]) ProtectedFunc[T] {
	return func() (T, error) {
		return r.Do(fn)
	}
}
