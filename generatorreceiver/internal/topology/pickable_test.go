package topology

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_pickBasedOnWeight(t *testing.T) {
	type args struct {
		ps []P
	}
	tests := []struct {
		name string
		args args
		want P
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, pickBasedOnWeight(tt.args.ps), "pickBasedOnWeight(%v)", tt.args.ps)
		})
	}
}
