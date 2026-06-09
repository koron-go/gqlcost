package gqlcost

import (
	"math"
	"testing"
)

func TestIsNullish(t *testing.T) {
	s := "hello"
	tests := []struct {
		v    interface{}
		want bool
	}{
		{nil, true},
		{0, true},
		{1, false},
		{-1, false},
		{0.0, false},
		{math.NaN(), true},
		{"", false},
		{"x", false},
		{(*string)(nil), true},
		{&s, false},
		{uint(0), false},
		{uint(1), false},
	}
	for _, tc := range tests {
		got := isNullish(tc.v)
		if got != tc.want {
			t.Errorf("isNullish(%v) = %v, want %v", tc.v, got, tc.want)
		}
	}
}
