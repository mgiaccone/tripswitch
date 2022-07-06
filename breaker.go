package tripswitch

import (
	"fmt"
	"strings"
	"sync"
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
	Unknown CircuitState = iota
	Closed
	HalfOpen
	Open
)

// String implements the Stringer interface.
func (s CircuitState) String() string {
	switch s {
	case Closed:
		return "closed"
	case HalfOpen:
		return "half-open"
	case Open:
		return "open"
	default:
		return "undefined"
	}
}

type circuit struct {
	name             string
	failCount        int32
	failThreshold    int32
	state            CircuitState
	successCount     int32
	successThreshold int32
	waitInterval     time.Duration
}

type recoverCircuitEvent struct {
	circuit  *circuit
	notifyFn notifyStateChangeFunc
}

type stateChangeEvent struct {
	name     string
	oldState CircuitState
	newState CircuitState
}

type notifyStateChangeFunc func(name string, oldState, newState CircuitState)

type recoverFunc func(c *circuit, notifyFn notifyStateChangeFunc)

// ProtectedFunc represents the function to be protected by the circuit breaker.
type ProtectedFunc[T any] func() (T, error)

// StateChangeFunc represents the function to handle state change notifications.
type StateChangeFunc func(name string, oldState, newState CircuitState)

// CircuitBreaker is the struct implementing the circuit breaker logic.
type CircuitBreaker[T any] struct {
	circuits                map[string]*circuit
	circuitsLock            sync.RWMutex
	defaultFailThreshold    int32
	defaultSuccessThreshold int32
	defaultWaitInterval     time.Duration
	notifyStateChangeCh     chan stateChangeEvent
	recoverCircuitCh        chan recoverCircuitEvent
	retrier                 Retrier[T]
	stateChangeFunc         StateChangeFunc
}

// MustCircuitBreaker creates a new instance of a circuit breaker or panics in case of a configuration error.
func MustCircuitBreaker[T any](retrier Retrier[T], opts ...CircuitBreakerOption) *CircuitBreaker[T] {
	cb, err := NewCircuitBreaker[T](retrier, opts...)
	if err != nil {
		panic(err)
	}
	return cb
}

// NewCircuitBreaker creates a new instance of a circuit breaker.
func NewCircuitBreaker[T any](retrier Retrier[T], opts ...CircuitBreakerOption) (*CircuitBreaker[T], error) {
	if retrier == nil {
		return nil, ErrRetrierRequired
	}

	cfg := config{
		circuits:                make(map[string]*circuit),
		defaultFailThreshold:    _defaultFailThreshold,
		defaultSuccessThreshold: _defaultSuccessThreshold,
		defaultWaitInterval:     _defaultWaitInterval,
	}
	if err := cfg.apply(opts...); err != nil {
		return nil, fmt.Errorf("apply options: %w", err)
	}

	cb := CircuitBreaker[T]{
		circuits:                cfg.circuits,
		defaultFailThreshold:    cfg.defaultFailThreshold,
		defaultSuccessThreshold: cfg.defaultSuccessThreshold,
		defaultWaitInterval:     cfg.defaultWaitInterval,
		notifyStateChangeCh:     make(chan stateChangeEvent),
		recoverCircuitCh:        make(chan recoverCircuitEvent),
		retrier:                 retrier,
		stateChangeFunc:         cfg.stateChangeFunc,
	}

	go cb.processEvents()

	return &cb, nil
}

// Do execute a function managed by the named circuit breaker.
func (cb *CircuitBreaker[T]) Do(name string, fn ProtectedFunc[T]) (T, error) {
	if name = strings.TrimSpace(name); name == "" {
		return *new(T), ErrNameRequired
	}

	// wrap function with the circuit breaker
	execFn := func() (T, error) {
		c := cb.getOrCreateCircuit(name, cb.notifyStateChange)

		// fails immediately if the circuit state is Open
		if CircuitState(atomic.LoadInt32((*int32)(&c.state))) == Open {
			return *new(T), ErrCircuitOpen
		}

		res, err := wrapWithRetrier(cb.retrier, fn)()
		if err != nil {
			recordFailure(c, cb.scheduleCircuitRecovery, cb.notifyStateChange)
			return res, err
		}

		recordSuccess(c, cb.notifyStateChange)
		return res, err
	}

	return execFn()
}

// State returns the state of a circuit or Unknown if the circuit does not exist.
func (cb *CircuitBreaker[T]) State(name string) CircuitState {
	cb.circuitsLock.RLock()
	defer cb.circuitsLock.RUnlock()

	if cb.circuits == nil {
		return Unknown
	}

	c, exists := cb.circuits[name]
	if !exists {
		return Unknown
	}

	return CircuitState(atomic.LoadInt32((*int32)(&c.state)))
}

func (cb *CircuitBreaker[T]) processEvents() {
	for {
		select {
		case msg, ok := <-cb.notifyStateChangeCh:
			if !ok {
				// TODO: log error instead of panic
				panic("error reading from notifyStateChange")
			}

			if cb.stateChangeFunc == nil {
				continue
			}
			cb.stateChangeFunc(msg.name, msg.oldState, msg.newState)
		case msg, ok := <-cb.recoverCircuitCh:
			if !ok {
				// TODO: log error instead of panic
				panic("error reading from recoverCircuit channel")
			}
			go recoverCircuit(msg.circuit, msg.notifyFn)
		}
	}
}

// getOrCreateCircuit returns a circuit. If the circuit does not exist,
// it will create a new one with the default configuration.
func (cb *CircuitBreaker[T]) getOrCreateCircuit(name string, notifyFn notifyStateChangeFunc) *circuit {
	cb.circuitsLock.Lock()
	defer cb.circuitsLock.Unlock()

	if cb.circuits == nil {
		cb.circuits = make(map[string]*circuit)
	}

	c, exists := cb.circuits[name]
	if !exists {
		c = &circuit{
			name:             name,
			state:            Closed,
			failThreshold:    cb.defaultFailThreshold,
			successThreshold: cb.defaultSuccessThreshold,
			waitInterval:     cb.defaultWaitInterval,
		}
		cb.circuits[name] = c
		notifyFn(name, Unknown, Closed)
	}

	return c
}

func (cb *CircuitBreaker[T]) notifyStateChange(name string, oldState, newState CircuitState) {
	cb.notifyStateChangeCh <- stateChangeEvent{
		name:     name,
		oldState: oldState,
		newState: newState,
	}
}

func (cb *CircuitBreaker[T]) scheduleCircuitRecovery(c *circuit, notifyFn notifyStateChangeFunc) {
	cb.recoverCircuitCh <- recoverCircuitEvent{
		circuit:  c,
		notifyFn: notifyFn,
	}
}

// recordFailure handles a failed function execution.
// If the current state is Closed and the failure counter reached the threshold, it will set the circuit breaker state to Open.
// Otherwise, it resets the success counter and sets the state to Open when the current state is HalfOpen.
func recordFailure(c *circuit, recoverFn recoverFunc, notifyFn notifyStateChangeFunc) {
	switch CircuitState(atomic.LoadInt32((*int32)(&c.state))) {
	case Open:
		// TODO: update stats
	case Closed:
		// TODO: update stats
		if atomic.AddInt32(&c.failCount, 1) < c.failThreshold {
			return
		}

		if atomic.CompareAndSwapInt32((*int32)(&c.state), int32(Closed), int32(Open)) {
			notifyFn(c.name, Closed, Open)
			recoverFn(c, notifyFn)
		}
	case HalfOpen:
		// TODO: update stats
		if atomic.CompareAndSwapInt32((*int32)(&c.state), int32(HalfOpen), int32(Open)) {
			atomic.AddInt32(&c.failCount, 1)
			atomic.StoreInt32(&c.successCount, 0)

			notifyFn(c.name, HalfOpen, Open)
			recoverFn(c, notifyFn)
		}
	default:
		// ignore other states
		return
	}
}

// recordSuccess handles a successful function execution.
// If the current state is HalfOpen, it resets the circuit breaker.
func recordSuccess(c *circuit, notifyFn notifyStateChangeFunc) {
	switch CircuitState(atomic.LoadInt32((*int32)(&c.state))) {
	case Open:
		// TODO: update stats
	case Closed:
		// TODO: update stats
	case HalfOpen:
		// TODO: update stats
		if atomic.AddInt32(&c.successCount, 1) < c.successThreshold {
			return
		}

		if atomic.CompareAndSwapInt32((*int32)(&c.state), int32(HalfOpen), int32(Closed)) {
			atomic.StoreInt32(&c.failCount, 0)
			atomic.StoreInt32(&c.successCount, 0)
			notifyFn(c.name, HalfOpen, Closed)
		}
	default:
		// ignore other states
		return
	}
}

// recoverCircuit waits for the configured interval before attempting to reopen the circuit.
// If the current state is Open, it sets a timer to setting the state to HalfOpen.
func recoverCircuit(c *circuit, notifyFn notifyStateChangeFunc) {
	if CircuitState(atomic.LoadInt32((*int32)(&c.state))) != Open {
		return
	}

	t := time.NewTimer(c.waitInterval)
	defer t.Stop()

	<-t.C

	if atomic.CompareAndSwapInt32((*int32)(&c.state), int32(Open), int32(HalfOpen)) {
		atomic.StoreInt32(&c.failCount, 0)
		atomic.StoreInt32(&c.successCount, 0)
		notifyFn(c.name, Open, HalfOpen)
	}
	// TODO: update stats
}
