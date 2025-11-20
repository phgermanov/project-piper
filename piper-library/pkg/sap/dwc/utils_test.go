//go:build unit
// +build unit

package dwc

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMergeMaps(t *testing.T) {
	t.Parallel()
	type args struct {
		a map[string]interface{}
		b map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{
			name: "merge maps with different fields",
			args: args{
				a: map[string]interface{}{
					"x": "foo",
					"z": map[string]interface{}{
						"foo": "bar",
					},
				},
				b: map[string]interface{}{
					"y": "bar",
				},
			},
			want: map[string]interface{}{
				"x": "foo",
				"y": "bar",
				"z": map[string]interface{}{
					"foo": "bar",
				},
			},
		},
		{
			name: "merge maps with same fields - latest should override value",
			args: args{
				a: map[string]interface{}{
					"x": "foo",
					"y": map[string]interface{}{
						"x": "foo",
					},
				},
				b: map[string]interface{}{
					"x": "bar",
					"y": map[string]interface{}{
						"x": "bar",
					},
				},
			},
			want: map[string]interface{}{
				"x": "bar",
				"y": map[string]interface{}{
					"x": "bar",
				},
			},
		},
		{
			name: "merge maps with overlapping fields - latest should override value",
			args: args{
				a: map[string]interface{}{
					"x": "foo",
					"y": "bar",
					"a": map[string]interface{}{
						"x": "foo",
						"y": "bar",
					},
				},
				b: map[string]interface{}{
					"x": "bar",
					"z": "foo",
					"a": map[string]interface{}{
						"y": "foo",
						"z": "bar",
					},
				},
			},
			want: map[string]interface{}{
				"x": "bar",
				"y": "bar",
				"z": "foo",
				"a": map[string]interface{}{
					"x": "foo",
					"y": "foo",
					"z": "bar",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, MergeMaps(tt.args.a, tt.args.b), "MergeMaps(%v, %v)", tt.args.a, tt.args.b)
		})
	}
}
