package gqlcost

import (
	"reflect"
	"strconv"
)

func toNumber(v interface{}) (int, bool) {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return 0, false
	}
	switch rv.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice:
		n := rv.Len()
		if n > 0 {
			return n, true
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n := int(rv.Int())
		if n != 0 {
			return n, true
		}
	case reflect.Uint, reflect.Uintptr, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n := int(rv.Uint())
		if n > 0 {
			return n, true
		}
	case reflect.Float32, reflect.Float64:
		n := int(rv.Float())
		if n != 0 {
			return n, true
		}
	case reflect.String:
		n, err := strconv.ParseInt(rv.String(), 10, 64)
		if err != nil {
			return 0, false
		}
		if n != 0 {
			return int(n), true
		}
	}
	return 0, false
}

func maxCost(costs []int) int {
	n := len(costs)
	switch n {
	case 0:
		return 0
	case 1:
		return costs[0]
	default:
		m := n / 2
		a, b := maxCost(costs[:m]), maxCost(costs[m:])
		if a > b {
			return a
		}
		return b
	}
}

func copyInts(src []int) []int {
	n := len(src)
	if n == 0 {
		return nil
	}
	dst := make([]int, n)
	copy(dst, src)
	return dst
}

func typName(typDef interface{}) (string, bool) {
	if x, ok := typDef.(interface{ Name() string }); ok {
		return x.Name(), true
	}
	return "", false
}

