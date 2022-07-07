package breaker

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewCircuitBreaker(t *testing.T) {
	type args struct {
		opts []CircuitBreakeOption
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
				opts: nil,
			},
			wantFn: func(t *testing.T, cb *CircuitBreaker[any]) {
				require.Equal(t, _defaultFailThreshold, cb.failThreshold,
					"failThreshold: got = %v, want = %v", cb.failThreshold, _defaultFailThreshold)
				require.Equal(t, _defaultSuccessThreshold, cb.successThreshold,
					"successThreshold: got = %v, want = %v", cb.successThreshold, _defaultSuccessThreshold)
				require.Equal(t, _defaultWaitInterval, cb.waitInterval,
					"waitInterval: got = %v, want = %v", cb.waitInterval, _defaultWaitInterval)
			},
		},
		{
			name: "success with options",
			args: args{
				opts: []CircuitBreakeOption{
					WithFailThreshold(10),
					WithSuccessThreshold(99),
				},
			},
			wantFn: func(t *testing.T, cb *CircuitBreaker[any]) {
				require.Equal(t, int32(10), cb.failThreshold,
					"failThreshold: got = %v, want = %v", cb.failThreshold, 10)
				require.Equal(t, int32(99), cb.successThreshold,
					"successThreshold: got = %v, want = %v", cb.successThreshold, 99)
				require.Equal(t, _defaultWaitInterval, cb.waitInterval,
					"waitInterval: got = %v, want = %v", cb.waitInterval, _defaultWaitInterval)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cb := NewCircuitBreaker[any](tt.args.opts...)
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

	testErr := fmt.Errorf("test error")

	tests := []struct {
		name  string
		steps []testStep
	}{
		{
			name: "only successes",
			steps: []testStep{
				buildStep(0, nil, nil, CircuitClosed),
				buildStep(0, nil, nil, CircuitClosed),
			},
		},
		{
			name: "complete behaviour",
			steps: []testStep{
				/* success */ buildStep(10*time.Millisecond, nil, nil, CircuitClosed),
				/* error   */ buildStep(10*time.Millisecond, testErr, testErr, CircuitClosed), // below fail threshold
				/* error   */ buildStep(10*time.Millisecond, testErr, testErr, CircuitOpen), // trip the circuit
				/* blocked */ buildStep(10*time.Millisecond, nil, ErrCircuitOpen, CircuitOpen),
				/* blocked */ buildStep(10*time.Millisecond, nil, ErrCircuitOpen, CircuitOpen),
				/* success */ buildStep(750*time.Millisecond, nil, nil, CircuitHalfOpen), // after restore
				/* error   */ buildStep(10*time.Millisecond, testErr, testErr, CircuitOpen), // trip the circuit again
				/* blocked */ buildStep(10*time.Millisecond, nil, ErrCircuitOpen, CircuitOpen),
				/* success */ buildStep(750*time.Millisecond, nil, nil, CircuitHalfOpen), // after restore
				/* success */ buildStep(10*time.Millisecond, nil, nil, CircuitClosed), // above success threshold
				/* success */ buildStep(10*time.Millisecond, nil, nil, CircuitClosed),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			startTime := time.Now()

			stateChangeFn := func(oldState, newState CircuitState) {
				t.Logf("state change [old = %s, new = %s]", oldState, newState)
			}

			opts := []CircuitBreakeOption{
				WithFailThreshold(2),
				WithSuccessThreshold(2),
				WithWaitInterval(500 * time.Millisecond),
				WithStateChangeFunc(stateChangeFn),
			}
			cb := NewCircuitBreaker[int](opts...)

			for i, s := range tt.steps {
				if s.delay != 0 {
					time.Sleep(s.delay)
				}

				stepFn := func(i int, step testStep) ProtectedFunc[int] {
					return func() (int, error) {
						return i, s.err
					}
				}(i, s)

				execTime := time.Since(startTime).Milliseconds()
				t.Logf("execution    [step = %d, time = %dms, wantErr = %v, wantState = %v", i, execTime, s.wantErr, s.wantState)

				got, err := cb.Do(stepFn)
				t.Logf("result       [step = %d, time = %dms, got = %v, err = %v", i, execTime, got, err)

				want := i
				if errors.Is(err, ErrCircuitOpen) {
					want = 0
				}

				currentState := cb.State()
				require.Equal(t, s.wantState, currentState,
					"do(): step = %d, time = %dms, state = %v, wantState = %v, circuit = %+v", i, execTime, currentState, s.wantState, cb)
				require.Equal(t, want, got, "do(): step = %d, time = %dms, got = %v, want = %v", i, execTime, got, i)
				require.ErrorIs(t, err, s.wantErr, "do(): step = %d, time = %dms, error = %v, wantErr = %v", i, execTime, err, s.wantErr)
			}
		})
	}
}

func TestCircuitBreaker_State(t *testing.T) {
	tests := []struct {
		name  string
		state CircuitState
		want  CircuitState
	}{
		{
			name:  "return closed state for existing circuit",
			state: CircuitClosed,
			want:  CircuitClosed,
		},
		{
			name:  "return half open state for existing circuit",
			state: CircuitHalfOpen,
			want:  CircuitHalfOpen,
		},
		{
			name:  "return open state for existing circuit",
			state: CircuitOpen,
			want:  CircuitOpen,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cb := &CircuitBreaker[any]{
				state: tt.state,
			}

			got := cb.State()
			require.Equal(t, tt.want, got, "State(): got = %v, want = %v", got, tt.want)
		})
	}
}

func TestCircuitBreaker_notifyStateChange(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	wantOldState := CircuitClosed
	wantNewState := CircuitOpen

	notifyCh := make(chan stateChangeEvent)

	cb := CircuitBreaker[any]{
		notifyStateChangeCh: notifyCh,
	}

	go cb.notifyStateChange(wantOldState, wantNewState)

	var got stateChangeEvent

	select {
	case msg := <-notifyCh:
		got = msg
	case <-ctx.Done():
		require.NoError(t, ctx.Err(), "notifyStateChange() - err = %v, want no error", ctx.Err())
		return
	}

	require.Equal(t, wantOldState, got.oldState, "notifyStateChange() - oldState = %v, want = %v", got.oldState, wantOldState)
	require.Equal(t, wantNewState, got.newState, "notifyStateChange() - newState = %v, want = %v", got.newState, wantNewState)
}

func TestCircuitBreaker_scheduleRestore(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	restoreCh := make(chan restoreCircuitEvent)

	cb := CircuitBreaker[any]{
		restoreCircuitCh: restoreCh,
	}

	go cb.scheduleRestore()

	select {
	case <-restoreCh:
	case <-ctx.Done():
		require.NoError(t, ctx.Err(), "scheduleRestore() - err = %v, want no error", ctx.Err())
		return
	}
}

func Test_recordFailure(t *testing.T) {
	tests := []struct {
		name             string
		state            CircuitState
		failCount        int32
		successCount     int32
		wantState        CircuitState
		wantFailCount    int32
		wantSuccessCount int32
		wantNotifyCount  int
	}{
		{
			name:             "failure applied to open circuit",
			state:            CircuitOpen,
			failCount:        1,
			successCount:     2,
			wantState:        CircuitOpen,
			wantFailCount:    1,
			wantSuccessCount: 2,
			wantNotifyCount:  0,
		},
		{
			name:             "failure applied to half open circuit",
			state:            CircuitHalfOpen,
			failCount:        0,
			successCount:     0,
			wantState:        CircuitOpen,
			wantFailCount:    1,
			wantSuccessCount: 0,
			wantNotifyCount:  1,
		},
		{
			name:             "failure applied to closed circuit below threshold",
			state:            CircuitClosed,
			failCount:        1,
			successCount:     0,
			wantState:        CircuitClosed,
			wantFailCount:    2,
			wantSuccessCount: 0,
			wantNotifyCount:  0,
		},
		{
			name:             "failure applied to closed circuit ready to open",
			state:            CircuitClosed,
			failCount:        2,
			successCount:     0,
			wantState:        CircuitOpen,
			wantFailCount:    3,
			wantSuccessCount: 0,
			wantNotifyCount:  1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			notifyCount := 0
			notifyFn := func(oldState, newState CircuitState) {
				notifyCount++
			}

			cb := NewCircuitBreaker[any]()
			cb.state = tt.state
			cb.failCount = tt.failCount
			cb.successCount = tt.successCount
			cb.notifyStateChangeFn = notifyFn

			cb.recordFailure()

			require.Equal(t, tt.wantState, cb.state,
				"recordFailure() - state = %v, want = %v", cb.state, tt.wantState)
			require.Equal(t, tt.wantFailCount, cb.failCount,
				"recordFailure() - failCount = %v, want = %v", cb.failCount, tt.wantFailCount)
			require.Equal(t, tt.wantSuccessCount, cb.successCount,
				"recordFailure() - successCount = %v, want = %v", cb.successCount, tt.wantSuccessCount)
			require.Equal(t, tt.wantNotifyCount, notifyCount,
				"recordFailure() - notifyCount = %v, want = %v", notifyCount, tt.wantNotifyCount)
		})
	}
}

func TestCircuitBreaker_recordSuccess(t *testing.T) {
	tests := []struct {
		name             string
		state            CircuitState
		failCount        int32
		successCount     int32
		wantState        CircuitState
		wantFailCount    int32
		wantSuccessCount int32
		wantNotifyCount  int
	}{
		{
			name:             "success applied to closed circuit",
			state:            CircuitClosed,
			failCount:        1,
			successCount:     2,
			wantState:        CircuitClosed,
			wantFailCount:    1,
			wantSuccessCount: 2,
			wantNotifyCount:  0,
		},
		{
			name:             "success applied to open circuit",
			state:            CircuitOpen,
			failCount:        1,
			successCount:     2,
			wantState:        CircuitOpen,
			wantFailCount:    1,
			wantSuccessCount: 2,
			wantNotifyCount:  0,
		},
		{
			name:             "success applied to half open circuit below threshold",
			state:            CircuitHalfOpen,
			failCount:        1,
			successCount:     0,
			wantState:        CircuitHalfOpen,
			wantFailCount:    1,
			wantSuccessCount: 1,
			wantNotifyCount:  0,
		},
		{
			name:             "success applied to half open circuit ready to close",
			state:            CircuitHalfOpen,
			failCount:        1,
			successCount:     2,
			wantState:        CircuitClosed,
			wantFailCount:    0,
			wantSuccessCount: 0,
			wantNotifyCount:  1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			notifyCount := 0
			notifyFn := func(oldState, newState CircuitState) {
				notifyCount++
			}

			cb := NewCircuitBreaker[any]()
			cb.state = tt.state
			cb.failCount = tt.failCount
			cb.successCount = tt.successCount
			cb.notifyStateChangeFn = notifyFn

			cb.recordSuccess()

			require.Equal(t, tt.wantState, cb.state,
				"recordSuccess() - state = %v, want = %v", cb.state, tt.wantState)
			require.Equal(t, tt.wantFailCount, cb.failCount,
				"recordSuccess() - failCount = %v, want = %v", cb.failCount, tt.wantFailCount)
			require.Equal(t, tt.wantSuccessCount, cb.successCount,
				"recordSuccess() - successCount = %v, want = %v", cb.successCount, tt.wantSuccessCount)
			require.Equal(t, tt.wantNotifyCount, notifyCount,
				"recordSuccess() - notifyCount = %v, want = %v", notifyCount, tt.wantNotifyCount)
		})
	}
}

func Test_restoreCircuit(t *testing.T) {
	tests := []struct {
		name             string
		state            CircuitState
		failCount        int32
		successCount     int32
		wantState        CircuitState
		wantFailCount    int32
		wantSuccessCount int32
		wantNotifyCount  int
	}{
		{
			name:             "restore applied to closed circuit",
			state:            CircuitClosed,
			failCount:        1,
			successCount:     2,
			wantState:        CircuitClosed,
			wantFailCount:    1,
			wantSuccessCount: 2,
			wantNotifyCount:  0,
		},
		{
			name:             "restore applied to half open circuit",
			state:            CircuitHalfOpen,
			failCount:        1,
			successCount:     2,
			wantState:        CircuitHalfOpen,
			wantFailCount:    1,
			wantSuccessCount: 2,
			wantNotifyCount:  0,
		},
		{
			name:             "restore applied to open circuit",
			state:            CircuitOpen,
			failCount:        1,
			successCount:     2,
			wantState:        CircuitHalfOpen,
			wantFailCount:    0,
			wantSuccessCount: 0,
			wantNotifyCount:  1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			notifyCount := 0
			notifyFn := func(oldState, newState CircuitState) {
				notifyCount++
			}

			waitTime := 100 * time.Millisecond

			cb := NewCircuitBreaker[any](WithWaitInterval(waitTime))
			cb.state = tt.state
			cb.failCount = tt.failCount
			cb.successCount = tt.successCount
			cb.notifyStateChangeFn = notifyFn

			startTime := time.Now()
			cb.restoreCircuit()
			elapsed := time.Since(startTime)

			require.GreaterOrEqual(t, elapsed, waitTime,
				"restoreCircuit() - waitTime = %v, want >= %v", elapsed, waitTime)
			require.Equal(t, tt.wantState, cb.state,
				"restoreCircuit() - state = %v, want = %v", cb.state, tt.wantState)
			require.Equal(t, tt.wantFailCount, cb.failCount,
				"restoreCircuit() - failCount = %v, want = %v", cb.failCount, tt.wantFailCount)
			require.Equal(t, tt.wantSuccessCount, cb.successCount,
				"restoreCircuit() - successCount = %v, want = %v", cb.successCount, tt.wantSuccessCount)
			require.Equal(t, tt.wantNotifyCount, notifyCount,
				"restoreCircuit() - notifyCount = %v, want = %v", notifyCount, tt.wantNotifyCount)
		})
	}
}
