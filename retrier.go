package tripswitch

var (
	_ Retrier[any] = (*nopRetrier[any])(nil)
	// 	_ Retrier[any] = (*BackoffRetrier[any])(nil)
	// 	_ Retrier[any] = (*ConstantRetrier[any])(nil)
)

// Retrier is the interface representing a retrier.
type Retrier[T any] interface {
	Do(fn ProtectedFunc[T]) (T, error)
}

// NopRetrier returns a new instance of a pass-through retrier.
func NopRetrier[T any]() Retrier[T] {
	return &nopRetrier[T]{}
}

type nopRetrier[T any] struct {
}

// Do implement the ProtectedFunc interface.
func (r *nopRetrier[T]) Do(fn ProtectedFunc[T]) (T, error) {
	return fn()
}

// type BackoffRetrier[T any] struct {
// }
//
// func (r *BackoffRetrier[T]) Do(fn ProtectedFunc[T]) (T, error) {
// 	return *new(T), nil
// }
//
// type ConstantRetrier[T any] struct {
// }
//
// func (r *ConstantRetrier[T]) Do(fn ProtectedFunc[T]) (T, error) {
// 	return *new(T), nil
// }
