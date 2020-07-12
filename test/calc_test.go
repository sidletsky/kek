package test

import (
	"testing"
)

func Test_add(t *testing.T) {
	type args struct {
		num0 int
		num1 int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "one plus one",
			args: args{
				num0: 1,
				num1: 1,
			},
			want: 2,
		},
		{
			name: "two plus two",
			args: args{
				num0: 2,
				num1: 2,
			},
			want: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := add(tt.args.num0, tt.args.num1); got != tt.want {
				t.Errorf("add() = %v, want %v", got, tt.want)
			}
		})
	}
}
