package gqlcost

import "fmt"

// ComplexityRange provides valid complexity min and max values.
type ComplexityRange struct {
	Min int
	Max int
}

func (r ComplexityRange) outside(v int) bool {
	if r.Min > 0 && v < r.Min {
		return true
	}
	if r.Max > 0 && v > r.Max {
		return true
	}
	return false
}

func (r ComplexityRange) message() string {
	switch {
	case r.Min > 0 && r.Max > 0:
		return fmt.Sprintf("between %d and %d", r.Min, r.Max)
	case r.Min > 0:
		return fmt.Sprintf("greater than or equal to %d", r.Min)
	case r.Max > 0:
		return fmt.Sprintf("less than or equal to %d", r.Max)
	default:
		return "unknown"
	}
}
