package tripswitch

import (
	"reflect"
	"testing"
	"time"
)

func TestWithCircuit(t *testing.T) {
	type args struct {
		name             string
		failThreshold    int
		successThreshold int
		waitInterval     time.Duration
	}
	tests := []struct {
		name string
		args args
		want CircuitBreakerOption
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithCircuit(tt.args.name, tt.args.failThreshold, tt.args.successThreshold, tt.args.waitInterval); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithCircuit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithFailThreshold(t *testing.T) {
	type args struct {
		threshold int
	}
	tests := []struct {
		name string
		args args
		want CircuitBreakerOption
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithFailThreshold(tt.args.threshold); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithFailThreshold() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithStateChangeFunc(t *testing.T) {
	type args struct {
		fn StateChangeFunc
	}
	tests := []struct {
		name string
		args args
		want CircuitBreakerOption
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithStateChangeFunc(tt.args.fn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithStateChangeFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithSuccessThreshold(t *testing.T) {
	type args struct {
		threshold int
	}
	tests := []struct {
		name string
		args args
		want CircuitBreakerOption
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithSuccessThreshold(tt.args.threshold); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithSuccessThreshold() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithWaitInterval(t *testing.T) {
	type args struct {
		interval time.Duration
	}
	tests := []struct {
		name string
		args args
		want CircuitBreakerOption
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithWaitInterval(tt.args.interval); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithWaitInterval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_config_applyOpts(t *testing.T) {
	type fields struct {
		circuits                map[string]*circuit
		defaultFailThreshold    int32
		defaultSuccessThreshold int32
		defaultWaitInterval     time.Duration
		stateChangeFunc         StateChangeFunc
	}
	type args struct {
		opts []CircuitBreakerOption
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &config{
				circuits:                tt.fields.circuits,
				defaultFailThreshold:    tt.fields.defaultFailThreshold,
				defaultSuccessThreshold: tt.fields.defaultSuccessThreshold,
				defaultWaitInterval:     tt.fields.defaultWaitInterval,
				stateChangeFunc:         tt.fields.stateChangeFunc,
			}
			if err := c.applyOpts(tt.args.opts...); (err != nil) != tt.wantErr {
				t.Errorf("applyOpts() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
