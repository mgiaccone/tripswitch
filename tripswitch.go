package tripswitch

// ProtectedFunc represents the function to be protected by the circuit breaker.
type ProtectedFunc[T any] func() (T, error)

// var (
// 	circuits
// )
//
// func Do[T any](name string, fn ProtectedFunc[T]) (T, error) {
//
// }

// // getOrCreateCircuit returns a circuit. If the circuit does not exist,
// // it will create a new one with the default configuration.
// func (cb *CircuitBreaker[T]) getOrCreateCircuit(name string, notifyFn notifyStateChangeFunc) *circuit {
// 	cb.circuitsLock.Lock()
// 	defer cb.circuitsLock.Unlock()
//
// 	if cb.circuits == nil {
// 		cb.circuits = make(map[string]*circuit)
// 	}
//
// 	c, exists := cb.circuits[name]
// 	if !exists {
// 		c = &circuit{
// 			name:             name,
// 			state:            CircuitClosed,
// 			failThreshold:    cb.failThreshold,
// 			successThreshold: cb.successThreshold,
// 			waitInterval:     cb.waitInterval,
// 		}
// 		cb.circuits[name] = c
// 		notifyFn(name, Unknown, CircuitClosed)
// 	}
//
// 	return c
// }
