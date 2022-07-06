package tripswitch

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// func Test_config_apply(t *testing.T) {
// 	type fields struct {
// 		circuits                map[string]*circuit
// 		defaultFailThreshold    int32
// 		defaultSuccessThreshold int32
// 		defaultWaitInterval     time.Duration
// 		stateChangeFunc         StateChangeFunc
// 	}
// 	type args struct {
// 		opts []CircuitBreakerOption
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			c := &config{
// 				circuits:                tt.fields.circuits,
// 				defaultFailThreshold:    tt.fields.defaultFailThreshold,
// 				defaultSuccessThreshold: tt.fields.defaultSuccessThreshold,
// 				defaultWaitInterval:     tt.fields.defaultWaitInterval,
// 				stateChangeFunc:         tt.fields.stateChangeFunc,
// 			}
// 			if err := c.apply(tt.args.opts...); (err != nil) != tt.wantErr {
// 				t.Errorf("apply() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

func TestWithCircuit(t *testing.T) {
	type args struct {
		name             string
		failThreshold    int
		successThreshold int
		waitInterval     time.Duration
	}
	tests := []struct {
		name    string
		args    args
		want    *config
		wantErr error
	}{
		{
			name: "fail with name required error",
			args: args{
				name: "",
			},
			want:    nil,
			wantErr: ErrNameRequired,
		},
		{
			name: "success",
			args: args{
				name:             "test_1",
				failThreshold:    3,
				successThreshold: 4,
				waitInterval:     5,
			},
			want: &config{
				circuits: map[string]*circuit{
					"test_1": {
						name:             "test_1",
						state:            Closed,
						failThreshold:    int32(3),
						successThreshold: int32(4),
						waitInterval:     5,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var cfg config

			fn := WithCircuit(tt.args.name, tt.args.failThreshold, tt.args.successThreshold, tt.args.waitInterval)
			err := fn(&cfg)

			require.ErrorIs(t, err, tt.wantErr, "WithCircuit(): error = %v, wantErr = %v", err, tt.wantErr)
			if tt.wantErr == nil {
				require.Equal(t, tt.want, &cfg, "WithCircuit(): cfg = %v, want = %v", cfg, tt.want)
			}
		})
	}
}

func TestWithFailThreshold(t *testing.T) {
	var cfg config
	want := 3
	err := WithFailThreshold(want)(&cfg)
	require.NoError(t, err, "TestWithFailThreshold(): unexpected error")
	require.Equal(t, int32(want), cfg.defaultFailThreshold,
		"TestWithFailThreshold(): got = %v, want = %v", cfg.defaultFailThreshold, want)
}

func TestWithSuccessThreshold(t *testing.T) {
	var cfg config
	want := 3
	err := WithSuccessThreshold(want)(&cfg)
	require.NoError(t, err, "WithSuccessThreshold(): unexpected error")
	require.Equal(t, int32(want), cfg.defaultSuccessThreshold,
		"WithSuccessThreshold(): got = %v, want = %v", cfg.defaultSuccessThreshold, want)
}

func TestWithWaitInterval(t *testing.T) {
	var cfg config
	want := 99 * time.Millisecond
	err := WithWaitInterval(want)(&cfg)
	require.NoError(t, err, "WithWaitInterval(): unexpected error")
	require.Equal(t, want, cfg.defaultWaitInterval,
		"WithWaitInterval(): got = %v, want = %v", cfg.defaultWaitInterval, want)
}

func TestWithStateChangeFunc(t *testing.T) {
	var cfg config
	want := func(name string, oldState, newState CircuitState) {}
	err := WithStateChangeFunc(want)(&cfg)
	require.NoError(t, err, "WithStateChangeFunc(): unexpected error")
	require.Equal(t, reflect.ValueOf(want).Pointer(), reflect.ValueOf(cfg.stateChangeFunc).Pointer(),
		"WithStateChangeFunc(): cfg = %v, want = %v", cfg.stateChangeFunc, want)
}
