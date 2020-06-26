package gqlcost

import "testing"

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
