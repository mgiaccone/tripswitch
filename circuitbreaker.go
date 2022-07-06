package tripswitch

import (
	"sync/atomic"
	"time"
)

const (
	_defaultFailThreshold    = int32(3)
	_defaultSuccessThreshold = int32(3)
	_defaultWaitInterval     = 10 * time.Second
)

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
	default:
		return "undefined"
	}
}

// StateChangeFunc represents the function to handle state change notifications.
type StateChangeFunc func(oldState, newState CircuitState)

type recoverCircuitEvent struct{}

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
	state               CircuitState
	successCount        int32
	successThreshold    int32
	waitInterval        time.Duration
	notifyStateChangeCh chan stateChangeEvent
	recoverCircuitCh    chan recoverCircuitEvent
	retrier             Retrier[T]
	stateChangeFunc     StateChangeFunc

	// these are used as test hooks
	notifyStateChangeFn notifyStateChangeFunc
	scheduleRecoverFn   notifyRecoverFunc
}

// NewCircuitBreaker creates a new instance of a circuit breaker.
func NewCircuitBreaker[T any](opts ...CircuitOption) *CircuitBreaker[T] {
	return NewCircuitBreakerWithRetrier[T](NopRetrier[T](), opts...)
}

func NewCircuitBreakerWithRetrier[T any](retrier Retrier[T], opts ...CircuitOption) *CircuitBreaker[T] {
	cfg := circuitConfig{
		stateChangeFunc:  func(oldState, newState CircuitState) {},
		failThreshold:    _defaultFailThreshold,
		successThreshold: _defaultSuccessThreshold,
		waitInterval:     _defaultWaitInterval,
	}
	cfg.applyOpts(opts...)

	cb := CircuitBreaker[T]{
		failThreshold:       cfg.failThreshold,
		notifyStateChangeCh: make(chan stateChangeEvent),
		recoverCircuitCh:    make(chan recoverCircuitEvent),
		retrier:             retrier,
		state:               CircuitClosed,
		stateChangeFunc:     cfg.stateChangeFunc,
		successThreshold:    cfg.successThreshold,
		waitInterval:        cfg.waitInterval,
	}
	cb.scheduleRecoverFn = cb.scheduleRecover
	cb.notifyStateChangeFn = cb.notifyStateChange

	go cb.processEvents()

	return &cb
}

// Do executes a function managed by the named circuit breaker.
func (cb *CircuitBreaker[T]) Do(fn ProtectedFunc[T]) (T, error) {
	var zeroValue T

	// wrap function with the circuit breaker
	execFn := func() (T, error) {
		// fails immediately if the circuit state is CircuitOpen
		if CircuitState(atomic.LoadInt32((*int32)(&cb.state))) == CircuitOpen {
			return zeroValue, ErrCircuitOpen
		}

		res, err := wrapWithRetrier(cb.retrier, fn)()
		if err != nil {
			cb.recordFailure()
			return res, err
		}

		cb.recordSuccess()

		return res, err
	}

	return execFn()
}

// State returns the state of the circuit.
func (cb *CircuitBreaker[T]) State() CircuitState {
	return CircuitState(atomic.LoadInt32((*int32)(&cb.state)))
}

// processEvents handles all the internal events.
func (cb *CircuitBreaker[T]) processEvents() {
	for {
		select {
		case msg := <-cb.notifyStateChangeCh:
			cb.stateChangeFunc(msg.oldState, msg.newState)
		case <-cb.recoverCircuitCh:
			go cb.recoverCircuit()
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

// scheduleRecover publishes a circuit recover request.
func (cb *CircuitBreaker[T]) scheduleRecover() {
	cb.recoverCircuitCh <- recoverCircuitEvent{}
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

// recoverCircuit waits for the configured interval before attempting to reopen the circuit.
// If the current state is CircuitOpen, it sets a timer to setting the state to CircuitHalfOpen.
func (cb *CircuitBreaker[T]) recoverCircuit() {
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
