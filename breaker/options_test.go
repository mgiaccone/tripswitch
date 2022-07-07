package breaker

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// func Test_config_apply(t *testing.T) {
// 	type fields struct {
// 		circuits                map[string]*circuit
// 		failThreshold    int32
// 		successThreshold int32
// 		waitInterval     time.Duration
// 		stateChangeFunc         StateChangeFunc
// 	}
// 	type args struct {
// 		opts []Option
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
// 				failThreshold:    tt.fields.failThreshold,
// 				successThreshold: tt.fields.successThreshold,
// 				waitInterval:     tt.fields.waitInterval,
// 				stateChangeFunc:         tt.fields.stateChangeFunc,
// 			}
// 			if err := c.applyOpts(tt.args.opts...); (err != nil) != tt.wantErr {
// 				t.Errorf("applyOpts() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

func TestWithFailThreshold(t *testing.T) {
	var cfg config
	want := 3
	WithFailThreshold(want)(&cfg)
	require.Equal(t, int32(want), cfg.failThreshold,
		"TestWithFailThreshold(): got = %v, want = %v", cfg.failThreshold, want)
}

func TestWithSuccessThreshold(t *testing.T) {
	var cfg config
	want := 3
	WithSuccessThreshold(want)(&cfg)
	require.Equal(t, int32(want), cfg.successThreshold,
		"WithSuccessThreshold(): got = %v, want = %v", cfg.successThreshold, want)
}

func TestWithWaitInterval(t *testing.T) {
	var cfg config
	want := 99 * time.Millisecond
	WithWaitInterval(want)(&cfg)
	require.Equal(t, want, cfg.waitInterval,
		"WithWaitInterval(): got = %v, want = %v", cfg.waitInterval, want)
}

func TestWithStateChangeFunc(t *testing.T) {
	var cfg config
	want := func(oldState, newState CircuitState) {}
	WithStateChangeFunc(want)(&cfg)
	require.Equal(t, reflect.ValueOf(want).Pointer(), reflect.ValueOf(cfg.stateChangeFunc).Pointer(),
		"WithStateChangeFunc(): cfg = %v, want = %v", cfg.stateChangeFunc, want)
}
