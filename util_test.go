package gqlcost

import (
	"testing"
)

func TestMaxCost(t *testing.T) {
	for _, tc := range []struct {
		costs []int
		max   int
	}{
		{[]int{1, 2, 3}, 3},
		{nil, 0},
		{[]int{}, 0},
		{[]int{9}, 9},
		{[]int{5, 4}, 5},
		{[]int{6, 7}, 7},
		{[]int{5, 2, 1, 0, 3, 4}, 5},
	} {
		act := maxCost(tc.costs)
		if act != tc.max {
			t.Fatalf("maxCost(%v) returns unexpected value: expect=%d actual=%d", tc.costs, tc.max, act)
		}
	}

}

func TestToNumber(t *testing.T) {
	tests := []struct {
		v    interface{}
		want int
		ok   bool
	}{
		{int(42), 42, true},
		{int8(8), 8, true},
		{int16(16), 16, true},
		{int32(32), 32, true},
		{int64(64), 64, true},
		{uint(1), 1, true},
		{uint8(8), 8, true},
		{uint16(16), 16, true},
		{uint32(32), 32, true},
		{uint64(64), 64, true},
		{float32(1.5), 1, true},
		{float64(2.5), 2, true},
		{"42", 42, true},
		{"abc", 0, false},
		{[]string{"a", "b"}, 2, true},
		{[]int{1}, 1, true},
		{[]int{}, 0, false},
		{map[string]int{"a": 1}, 1, true},
		{0, 0, false},
		{nil, 0, false},
	}
	for _, tc := range tests {
		got, ok := toNumber(tc.v)
		if ok != tc.ok || got != tc.want {
			t.Errorf("toNumber(%v) = (%d, %v), want (%d, %v)", tc.v, got, ok, tc.want, tc.ok)
		}
	}
}

type namedDummy struct{ name string }

func (n namedDummy) Name() string { return n.name }

func TestTypName(t *testing.T) {
	if got := typName(namedDummy{name: "Foo"}); got != "Foo" {
		t.Errorf("typName = %q, want %q", got, "Foo")
	}
	if got := typName("string"); got != "" {
		t.Errorf("typName = %q, want %q", got, "")
	}
}

