package ens

import "testing"

func TestUnpackPayload(t *testing.T) {
	type args struct {
		payload []byte
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{
			name: "1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			UnpackPayload(tt.args.payload)
		})
	}
}
