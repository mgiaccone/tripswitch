package tripswitch

// func TestCircuitBreaker_getOrCreateCircuit(t *testing.T) {
// 	tests := []struct {
// 		name            string
// 		circuits        map[string]*circuit
// 		testNames       []string
// 		wantFn          func(t *testing.T, initial map[string]*circuit, current map[string]*circuit)
// 		wantNotifyCount int
// 	}{
// 		{
// 			name:      "success empty initial circuits",
// 			testNames: []string{"test"},
// 			wantFn: func(t *testing.T, initial map[string]*circuit, current map[string]*circuit) {
// 				require.Empty(t, initial, "getOrCreateCircuit(): want init empty")
//
// 				got := current["test"]
// 				require.NotNil(t, got, "getOrCreateCircuit(): got = %v, want not nil", got)
// 			},
// 			wantNotifyCount: 1,
// 		},
// 		{
// 			name: "success with existing circuit",
// 			circuits: map[string]*circuit{
// 				"test": {name: "test"},
// 			},
// 			testNames:       []string{"test"},
// 			wantNotifyCount: 0,
// 			wantFn: func(t *testing.T, initial map[string]*circuit, current map[string]*circuit) {
// 				expected := initial["test"]
// 				got := current["test"]
// 				require.Equal(t, expected, got, "getOrCreateCircuit(): got = %v, want = %v", got, expected)
// 			},
// 		},
// 		{
// 			name: "success with mixed state",
// 			circuits: map[string]*circuit{
// 				"test_0": {name: "test_0"},
// 			},
// 			testNames: []string{"test_0", "test_1"},
// 			wantFn: func(t *testing.T, initial map[string]*circuit, current map[string]*circuit) {
// 				// existing
// 				expected := initial["test_0"]
// 				got := current["test_0"]
// 				require.Equal(t, expected, got, "getOrCreateCircuit(): got = %v, want = %v", got, expected)
//
// 				// new
// 				got = current["test_1"]
// 				require.NotNil(t, got, "getOrCreateCircuit(): got = %v, want not nil", got)
// 			},
// 			wantNotifyCount: 1,
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		tt := tt
// 		t.Run(tt.name, func(t *testing.T) {
// 			cb := &CircuitBreaker[any]{
// 				circuits: tt.circuits,
// 			}
//
// 			notifyCount := 0
// 			notifyFn := func(name string, oldState, newState CircuitState) {
// 				notifyCount++
// 			}
//
// 			for _, name := range tt.testNames {
// 				got := cb.getOrCreateCircuit(name, notifyFn)
// 				require.NotNil(t, got, "getOrCreateCircuit(): want not nil, got = %v", got)
// 			}
//
// 			if tt.wantFn != nil {
// 				tt.wantFn(t, tt.circuits, cb.circuits)
// 			}
// 			require.Equal(t, tt.wantNotifyCount, notifyCount, "recordFailure() - notifyCount: got = %v, want = %v", notifyCount, tt.wantNotifyCount)
// 		})
// 	}
// }
//
