package breaker

import (
	"errors"
	"reflect"
	"strings"
	"sync"
	"unsafe"

	"github.com/mgiaccone/tripswitch/internal/coreutil"
)

var (
	_circuits     = make(map[string]*entry)
	_circuitsLock sync.Mutex
	_defaultOpts  []Option
)

var (
	// ErrRequiredName is returned when a circuit name is required.
	ErrRequiredName = errors.New("missing circuit name")

	// ErrRequiredRetrier is returned when a retrier is required.
	ErrRequiredRetrier = errors.New("missing retrier")

	// ErrDuplicateCircuit is returned when attempting to configure
	// a circuit that already exists.
	ErrDuplicateCircuit = errors.New("duplicate circuit")

	// ErrTypeMismatch is returned when a circuit configured for a type
	// is used with a different generic type.
	ErrTypeMismatch = errors.New("circuit breaker type mismatch")
)

type entry struct {
	typ reflect.Type
	ptr unsafe.Pointer
}

// Configure sets custom options for a named circuit breaker.
func Configure[T any](name string, opts ...Option) error {
	return ConfigureWithRetrier[T](name, &nopRetrier[T]{}, opts...)
}

// ConfigureWithRetrier sets a retrier and custom options for a named circuit breaker.
func ConfigureWithRetrier[T any](name string, retrier Retrier[T], opts ...Option) error {
	if len(strings.TrimSpace(name)) == 0 {
		return ErrRequiredName
	}

	if retrier == nil {
		return ErrRequiredRetrier
	}

	_circuitsLock.Lock()
	defer _circuitsLock.Unlock()

	if _, exists := _circuits[name]; exists {
		return ErrDuplicateCircuit
	}

	cb := NewCircuitBreakerWithRetrier[T](retrier, opts...)

	_circuits[name] = &entry{
		typ: reflect.TypeOf(cb),
		ptr: unsafe.Pointer(reflect.ValueOf(cb).Pointer()),
	}

	return nil
}

// DefaultOptions overrides the default options.
func DefaultOptions(opts ...Option) {
	_defaultOpts = opts
}

// MustConfigure sets custom options for a named circuit breaker.
// It triggers a panic in case of error.
func MustConfigure[T any](name string, opts ...Option) {
	coreutil.MustErr(ConfigureWithRetrier[T](name, &nopRetrier[T]{}, opts...))
}

// MustWithRetrier sets a retrier and custom options for a named circuit breaker.
// It triggers a panic in case of error.
func MustWithRetrier[T any](name string, retrier Retrier[T], opts ...Option) {
	coreutil.MustErr(ConfigureWithRetrier[T](name, retrier, opts...))
}

// Do wraps a function execution with a named circuit breaker.
func Do[T any](name string, fn ProtectedFunc[T]) (res T, err error) {
	cb, err := getOrCreateEntry[T](name)
	if err != nil {
		// nolint:gocritic
		return *new(T), err
	}

	return cb.Do(fn)
}

func getOrCreateEntry[T any](name string) (*CircuitBreaker[T], error) {
	_circuitsLock.Lock()
	defer _circuitsLock.Unlock()

	v, exists := _circuits[name]
	if exists {
		vTyp := reflect.TypeOf((*CircuitBreaker[T])(v.ptr))
		if vTyp.String() != v.typ.String() {
			return nil, ErrTypeMismatch
		}
	}

	cb := (*CircuitBreaker[T])(v.ptr)
	if cb == nil {
		cb = NewCircuitBreaker[T]()
		_circuits[name] = &entry{
			typ: reflect.TypeOf(cb),
			ptr: unsafe.Pointer(reflect.ValueOf(cb).Pointer()),
		}
	}

	return cb, nil
}
