package tripswitch

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMustCircuitBreaker(t *testing.T) {
	type args struct {
		retrier Retrier[any]
		opts    []CircuitBreakerOption
	}

	tests := []struct {
		name      string
		args      args
		wantPanic bool
	}{
		{
			name: "success without options",
			args: args{
				retrier: NopRetrier[any](),
				opts:    nil,
			},
			wantPanic: false,
		},
		{
			name: "panics with options",
			args: args{
				retrier: NopRetrier[any](),
				opts: []CircuitBreakerOption{
					WithCircuit("c1", 3, 3, 5*time.Second),
					WithCircuit("c1", 3, 3, 5*time.Second),
				},
			},
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			fn := func() {
				_ = MustCircuitBreaker[any](tt.args.retrier, tt.args.opts...)
			}
			if tt.wantPanic {
				require.Panics(t, fn, "MustCircuitBreaker(): want panic")
			} else {
				require.NotPanics(t, fn, "MustCircuitBreaker(): want no panic")
			}
		})
	}
}

func TestNewCircuitBreaker(t *testing.T) {
	type args struct {
		retrier Retrier[any]
		opts    []CircuitBreakerOption
	}

	tests := []struct {
		name    string
		args    args
		wantFn  func(t *testing.T, cb *CircuitBreaker[any])
		wantErr error
	}{
		{
			name: "success without options",
			args: args{
				retrier: NopRetrier[any](),
				opts:    nil,
			},
			wantFn: func(t *testing.T, cb *CircuitBreaker[any]) {
				require.NotNil(t, cb.circuits,
					"circuits: got = %v, want not nil", cb.circuits)
				require.Equal(t, _defaultFailThreshold, cb.defaultFailThreshold,
					"defaultFailThreshold: got = %v, want = %v", cb.defaultFailThreshold, _defaultFailThreshold)
				require.Equal(t, _defaultSuccessThreshold, cb.defaultSuccessThreshold,
					"defaultSuccessThreshold: got = %v, want = %v", cb.defaultSuccessThreshold, _defaultSuccessThreshold)
				require.Equal(t, _defaultWaitInterval, cb.defaultWaitInterval,
					"defaultWaitInterval: got = %v, want = %v", cb.defaultWaitInterval, _defaultWaitInterval)
			},
		},
		{
			name: "success with options",
			args: args{
				retrier: NopRetrier[any](),
				opts: []CircuitBreakerOption{
					WithFailThreshold(10),
					WithSuccessThreshold(99),
				},
			},
			wantFn: func(t *testing.T, cb *CircuitBreaker[any]) {
				require.NotNil(t, cb.circuits,
					"circuits: got = %v, want not nil", cb.circuits)
				require.Equal(t, int32(10), cb.defaultFailThreshold,
					"defaultFailThreshold: got = %v, want = %v", cb.defaultFailThreshold, 10)
				require.Equal(t, int32(99), cb.defaultSuccessThreshold,
					"defaultSuccessThreshold: got = %v, want = %v", cb.defaultSuccessThreshold, 99)
				require.Equal(t, _defaultWaitInterval, cb.defaultWaitInterval,
					"defaultWaitInterval: got = %v, want = %v", cb.defaultWaitInterval, _defaultWaitInterval)
			},
		},
		{
			name: "fails with missing retrier",
			args: args{
				retrier: nil,
			},
			wantErr: ErrRetrierRequired,
		},
		{
			name: "fails with options",
			args: args{
				retrier: NopRetrier[any](),
				opts: []CircuitBreakerOption{
					WithCircuit("c1", 3, 3, 5*time.Second),
					WithCircuit("c1", 3, 3, 5*time.Second),
				},
			},
			wantErr: ErrCircuitConflict,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cb, err := NewCircuitBreaker[any](tt.args.retrier, tt.args.opts...)
			require.ErrorIs(t, err, tt.wantErr, "NewCircuitBreaker(): error = %v, wantErr = %v", err, tt.wantErr)
			if tt.wantFn != nil {
				tt.wantFn(t, cb)
			}
		})
	}
}

func TestCircuitBreaker_Do(t *testing.T) {
	type testStep struct {
		delay     time.Duration
		err       error
		wantErr   error
		wantState CircuitState
	}

	buildStep := func(delay time.Duration, err, wantErr error, wantState CircuitState) testStep {
		return testStep{delay: delay, err: err, wantErr: wantErr, wantState: wantState}
	}

	testErr := fmt.Errorf("err_1")

	type args struct {
		name string
	}

	tests := []struct {
		name  string
		args  args
		steps []testStep
	}{
		{
			name: "fails with missing name",
			args: args{name: ""},
			steps: []testStep{
				buildStep(0, ErrNameRequired, ErrNameRequired, Closed),
			},
		},
		{
			name: "only successes",
			args: args{name: "test_1"},
			steps: []testStep{
				buildStep(0, nil, nil, Closed),
				buildStep(0, nil, nil, Closed),
			},
		},
		{
			name: "complete behaviour",
			args: args{name: "test_2"},
			steps: []testStep{
				/* success */ buildStep(10*time.Millisecond, nil, nil, Closed),
				/* error   */ buildStep(10*time.Millisecond, testErr, testErr, Closed), // below fail threshold
				/* error   */ buildStep(10*time.Millisecond, testErr, testErr, Open), // trip the circuit
				/* blocked */ buildStep(10*time.Millisecond, nil, ErrCircuitOpen, Open),
				/* blocked */ buildStep(10*time.Millisecond, nil, ErrCircuitOpen, Open),
				/* success */ buildStep(750*time.Millisecond, nil, nil, HalfOpen), // after restore
				/* error   */ buildStep(10*time.Millisecond, testErr, testErr, Open), // trip the circuit again
				/* blocked */ buildStep(10*time.Millisecond, nil, ErrCircuitOpen, Open),
				/* success */ buildStep(750*time.Millisecond, nil, nil, HalfOpen), // after restore
				/* success */ buildStep(10*time.Millisecond, nil, nil, Closed), // above success threshold
				/* success */ buildStep(10*time.Millisecond, nil, nil, Closed),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			startTime := time.Now()
			currentState := Closed

			stateFn := func(name string, oldState, newState CircuitState) {
				currentState = newState
				execTime := time.Now().Sub(startTime).Milliseconds()
				t.Logf("state change [time = %dms, oldState = %s, oldState = %s]", execTime, oldState, newState)
			}

			opts := []CircuitBreakerOption{
				WithFailThreshold(2),
				WithSuccessThreshold(2),
				WithWaitInterval(500 * time.Millisecond),
				WithStateChangeFunc(stateFn),
			}
			cb := MustCircuitBreaker[int](NopRetrier[int](), opts...)

			for i, s := range tt.steps {
				if s.delay != 0 {
					time.Sleep(s.delay)
				}

				stepFn := func(i int, step testStep) ProtectedFunc[int] {
					return func() (int, error) {
						return i, s.err
					}
				}(i, s)

				execTime := time.Now().Sub(startTime).Milliseconds()
				t.Logf("execution    [step = %d, time = %dms, wantErr = %v, wantState = %v", i, execTime, s.wantErr, s.wantState)

				got, err := cb.Do(tt.args.name, stepFn)
				t.Logf("result       [step = %d, time = %dms, got = %v, err = %v", i, execTime, got, err)

				want := i
				if errors.Is(err, ErrCircuitOpen) {
					want = 0
				}

				require.Equal(t, s.wantState, currentState,
					"Do(): step = %d, time = %dms, state = %v, wantState = %v, circuit = %+v", i, execTime, currentState, s.wantState, cb.circuits[tt.args.name])
				require.Equal(t, want, got, "Do(): step = %d, time = %dms, got = %v, want = %v", i, execTime, got, i)
				require.ErrorIs(t, err, s.wantErr, "Do(): step = %d, time = %dms, error = %v, wantErr = %v", i, execTime, err, s.wantErr)
			}
		})
	}
}

func TestCircuitBreaker_State(t *testing.T) {
	type args struct {
		name string
	}

	tests := []struct {
		name     string
		circuits map[string]*circuit
		args     args
		want     CircuitState
	}{
		{
			name: "return unknown for nil circuits",
			want: Unknown,
		},
		{
			name: "return unknown for not existing circuit",
			circuits: map[string]*circuit{
				"test": {name: "test"},
			},
			args: args{
				name: "test_new",
			},
			want: Unknown,
		},
		{
			name: "return correct state for existing circuit",
			circuits: map[string]*circuit{
				"test": {name: "test", state: Closed},
			},
			args: args{
				name: "test",
			},
			want: Closed,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cb := &CircuitBreaker[any]{
				circuits: tt.circuits,
			}

			got := cb.State(tt.args.name)
			require.Equal(t, tt.want, got, "State(): got = %v, want = %v", got, tt.want)
		})
	}
}

func TestCircuitBreaker_getOrCreateCircuit(t *testing.T) {
	tests := []struct {
		name            string
		circuits        map[string]*circuit
		testNames       []string
		wantFn          func(t *testing.T, initial map[string]*circuit, current map[string]*circuit)
		wantNotifyCount int
	}{
		{
			name:      "success empty initial circuits",
			testNames: []string{"test"},
			wantFn: func(t *testing.T, initial map[string]*circuit, current map[string]*circuit) {
				require.Empty(t, initial, "getOrCreateCircuit(): want init empty")

				got := current["test"]
				require.NotNil(t, got, "getOrCreateCircuit(): got = %v, want not nil", got)
			},
			wantNotifyCount: 1,
		},
		{
			name: "success with existing circuit",
			circuits: map[string]*circuit{
				"test": {name: "test"},
			},
			testNames:       []string{"test"},
			wantNotifyCount: 0,
			wantFn: func(t *testing.T, initial map[string]*circuit, current map[string]*circuit) {
				expected := initial["test"]
				got := current["test"]
				require.Equal(t, expected, got, "getOrCreateCircuit(): got = %v, want = %v", got, expected)
			},
		},
		{
			name: "success with mixed state",
			circuits: map[string]*circuit{
				"test_0": {name: "test_0"},
			},
			testNames: []string{"test_0", "test_1"},
			wantFn: func(t *testing.T, initial map[string]*circuit, current map[string]*circuit) {
				// existing
				expected := initial["test_0"]
				got := current["test_0"]
				require.Equal(t, expected, got, "getOrCreateCircuit(): got = %v, want = %v", got, expected)

				// new
				got = current["test_1"]
				require.NotNil(t, got, "getOrCreateCircuit(): got = %v, want not nil", got)
			},
			wantNotifyCount: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cb := &CircuitBreaker[any]{
				circuits: tt.circuits,
			}

			notifyCount := 0
			notifyFn := func(name string, oldState, newState CircuitState) {
				notifyCount++
			}

			for _, name := range tt.testNames {
				got := cb.getOrCreateCircuit(name, notifyFn)
				require.NotNil(t, got, "getOrCreateCircuit(): want not nil, got = %v", got)
			}

			if tt.wantFn != nil {
				tt.wantFn(t, tt.circuits, cb.circuits)
			}
			require.Equal(t, tt.wantNotifyCount, notifyCount, "recordFailure() - notifyCount: got = %v, want = %v", notifyCount, tt.wantNotifyCount)
		})
	}
}

func TestCircuitBreaker_notifyStateChange(t *testing.T) {
	type args struct {
		name     string
		oldState CircuitState
		newState CircuitState
	}
	tests := []struct {
		name         string
		args         args
		setCallback  bool
		wantCount    int
		wantName     string
		wantOldState CircuitState
		wantNewState CircuitState
	}{
		{
			name: "without callback",
			args: args{
				name:     "test",
				oldState: Closed,
				newState: Open,
			},
			setCallback:  false,
			wantCount:    0,
			wantName:     "",
			wantOldState: Unknown,
			wantNewState: Unknown,
		},
		{
			name: "with callback",
			args: args{
				name:     "test",
				oldState: Closed,
				newState: Open,
			},
			setCallback:  true,
			wantCount:    1,
			wantName:     "test",
			wantOldState: Closed,
			wantNewState: Open,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cb := &CircuitBreaker[any]{}

			callCounter := 0
			callName := ""
			callOldState := Unknown
			callNewState := Unknown

			if tt.setCallback {
				cb.stateChangeFunc = func(name string, oldState, newState CircuitState) {
					callCounter++
					callName = name
					callOldState = oldState
					callNewState = newState
				}
			}

			cb.notifyStateChange(tt.args.name, tt.args.oldState, tt.args.newState)
			require.Equal(t, tt.wantCount, callCounter, "notifyStateChange(): got = %v, want = %v", callCounter, tt.wantCount)

			if tt.setCallback {
				require.Equal(t, tt.wantName, callName, "notifyStateChange() - name: got = %v, want = %v", callName, tt.wantName)
				require.Equal(t, tt.wantOldState, callOldState, "notifyStateChange() - oldState: got = %v, want = %v", callOldState, tt.wantOldState)
				require.Equal(t, tt.wantNewState, callNewState, "notifyStateChange() - newState: got = %v, want = %v", callNewState, tt.wantNewState)
			}
		})
	}
}

func Test_recordFailure(t *testing.T) {
	type args struct {
		c *circuit
	}
	tests := []struct {
		name             string
		args             args
		wantCircuit      *circuit
		wantNotifyCount  int
		wantRestoreCount int
	}{
		{
			name: "failure applied to open circuit",
			args: args{
				c: &circuit{
					name:          "test",
					state:         Open,
					failThreshold: 3,
					failCount:     0,
					successCount:  0,
				},
			},
			wantCircuit: &circuit{
				name:          "test",
				state:         Open,
				failThreshold: 3,
				failCount:     0,
				successCount:  0,
			},
			wantNotifyCount:  0,
			wantRestoreCount: 0,
		},
		{
			name: "failure applied to half open circuit",
			args: args{
				c: &circuit{
					name:          "test",
					state:         HalfOpen,
					failThreshold: 3,
					failCount:     0,
					successCount:  0,
				},
			},
			wantCircuit: &circuit{
				name:          "test",
				state:         Open,
				failThreshold: 3,
				failCount:     1,
				successCount:  0,
			},
			wantNotifyCount:  1,
			wantRestoreCount: 1,
		},
		{
			name: "failure applied to closed circuit below threshold",
			args: args{
				c: &circuit{
					name:          "test",
					state:         Closed,
					failThreshold: 3,
					failCount:     0,
					successCount:  0,
				},
			},
			wantCircuit: &circuit{
				name:          "test",
				state:         Closed,
				failThreshold: 3,
				failCount:     1,
				successCount:  0,
			},
			wantNotifyCount:  0,
			wantRestoreCount: 0,
		},
		{
			name: "failure applied to closed circuit ready to trip",
			args: args{
				c: &circuit{
					name:          "test",
					state:         Closed,
					failThreshold: 3,
					failCount:     2,
					successCount:  0,
				},
			},
			wantCircuit: &circuit{
				name:          "test",
				state:         Open,
				failThreshold: 3,
				failCount:     3,
				successCount:  0,
			},
			wantNotifyCount:  1,
			wantRestoreCount: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			restoreCount := 0
			restoreFn := func(c *circuit, notifyFn notifyStateChangeFunc) {
				restoreCount++
			}

			notifyCount := 0
			notifyFn := func(name string, oldState, newState CircuitState) {
				notifyCount++
			}

			got := tt.args.c
			recordFailure(tt.args.c, restoreFn, notifyFn)

			// sleep to allow goroutine to run
			time.Sleep(200 * time.Millisecond)

			require.Equal(t, tt.wantRestoreCount, restoreCount, "recordFailure() - restoreCount: got = %v, want = %v", restoreCount, tt.wantRestoreCount)
			require.Equal(t, tt.wantNotifyCount, notifyCount, "recordFailure() - notifyCount: got = %v, want = %v", notifyCount, tt.wantNotifyCount)
			require.Equal(t, tt.wantCircuit, got, "recordFailure() - circuit: got = %v, want = %v", got, tt.wantCircuit)
		})
	}
}

func TestCircuitBreaker_recordSuccess(t *testing.T) {
	type args struct {
		c *circuit
	}
	tests := []struct {
		name            string
		args            args
		wantCircuit     *circuit
		wantNotifyCount int
	}{
		{
			name: "success applied to open circuit",
			args: args{
				c: &circuit{
					name:             "test",
					state:            Open,
					failCount:        0,
					failThreshold:    3,
					successCount:     0,
					successThreshold: 3,
					waitInterval:     0,
				},
			},
			wantCircuit: &circuit{
				name:             "test",
				state:            Open,
				failCount:        0,
				failThreshold:    3,
				successCount:     0,
				successThreshold: 3,
				waitInterval:     0,
			},
			wantNotifyCount: 0,
		},
		{
			name: "success applied to closed circuit",
			args: args{
				c: &circuit{
					name:             "test",
					state:            Closed,
					failCount:        0,
					failThreshold:    3,
					successCount:     0,
					successThreshold: 3,
					waitInterval:     0,
				},
			},
			wantCircuit: &circuit{
				name:             "test",
				state:            Closed,
				failCount:        0,
				failThreshold:    3,
				successCount:     0,
				successThreshold: 3,
				waitInterval:     0,
			},
			wantNotifyCount: 0,
		},
		{
			name: "success applied to half open circuit below threshold",
			args: args{
				c: &circuit{
					name:             "test",
					state:            HalfOpen,
					failCount:        0,
					failThreshold:    3,
					successCount:     0,
					successThreshold: 3,
					waitInterval:     0,
				},
			},
			wantCircuit: &circuit{
				name:             "test",
				state:            HalfOpen,
				failCount:        0,
				failThreshold:    3,
				successCount:     1,
				successThreshold: 3,
				waitInterval:     0,
			},
			wantNotifyCount: 0,
		},
		{
			name: "success applied to half open circuit ready to close",
			args: args{
				c: &circuit{
					name:             "test",
					state:            HalfOpen,
					failCount:        0,
					failThreshold:    3,
					successCount:     2,
					successThreshold: 3,
					waitInterval:     0,
				},
			},
			wantCircuit: &circuit{
				name:             "test",
				state:            Closed,
				failCount:        0,
				failThreshold:    3,
				successCount:     0,
				successThreshold: 3,
				waitInterval:     0,
			},
			wantNotifyCount: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			notifyCount := 0
			notifyFn := func(name string, oldState, newState CircuitState) {
				notifyCount++
			}

			got := tt.args.c
			recordSuccess(tt.args.c, notifyFn)

			require.Equal(t, tt.wantNotifyCount, notifyCount, "recordSuccess() - notifyCount: got = %v, want = %v", notifyCount, tt.wantNotifyCount)
			require.Equal(t, tt.wantCircuit, got, "recordSuccess() - circuit: got = %v, want = %v", got, tt.wantCircuit)
		})
	}
}

func Test_restore(t *testing.T) {
	type args struct {
		c *circuit
	}
	tests := []struct {
		name            string
		args            args
		wantCircuit     *circuit
		wantNotifyCount int
	}{
		{
			name: "restore applied to closed circuit",
			args: args{
				c: &circuit{
					name:             "test",
					state:            Closed,
					failCount:        0,
					failThreshold:    3,
					successCount:     0,
					successThreshold: 3,
					waitInterval:     0,
				},
			},
			wantCircuit: &circuit{
				name:             "test",
				state:            Closed,
				failCount:        0,
				failThreshold:    3,
				successCount:     0,
				successThreshold: 3,
				waitInterval:     0,
			},
			wantNotifyCount: 0,
		},
		{
			name: "restore applied to half open circuit",
			args: args{
				c: &circuit{
					name:             "test",
					state:            HalfOpen,
					failCount:        0,
					failThreshold:    3,
					successCount:     0,
					successThreshold: 3,
					waitInterval:     0,
				},
			},
			wantCircuit: &circuit{
				name:             "test",
				state:            HalfOpen,
				failCount:        0,
				failThreshold:    3,
				successCount:     0,
				successThreshold: 3,
				waitInterval:     0,
			},
			wantNotifyCount: 0,
		},
		{
			name: "restore applied to open circuit",
			args: args{
				c: &circuit{
					name:             "test",
					state:            Open,
					failCount:        3,
					failThreshold:    3,
					successCount:     0,
					successThreshold: 3,
					waitInterval:     10 * time.Millisecond,
				},
			},
			wantCircuit: &circuit{
				name:             "test",
				state:            HalfOpen,
				failCount:        0,
				failThreshold:    3,
				successCount:     0,
				successThreshold: 3,
				waitInterval:     10 * time.Millisecond,
			},
			wantNotifyCount: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			notifyCount := 0
			notifyFn := func(name string, oldState, newState CircuitState) {
				notifyCount++
			}

			got := tt.args.c
			restore(tt.args.c, notifyFn)

			require.Equal(t, tt.wantNotifyCount, notifyCount, "restore() - notifyCount: got = %v, want = %v", notifyCount, tt.wantNotifyCount)
			require.Equal(t, tt.wantCircuit, got, "restore() - circuit: got = %v, want = %v", got, tt.wantCircuit)
		})
	}
}
