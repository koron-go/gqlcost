package gqlcost

// ComplexityRange provides valid complexity min and max values.
type ComplexityRange struct {
	Min int
	Max int
}

func (r ComplexityRange) outside(v int) bool {
	if r.Min == 0 && r.Max == 0 {
		return false
	}
	return v < r.Min || v > r.Max
}
