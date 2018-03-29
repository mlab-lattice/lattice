package expected

// Int32Range is a range of int32s, inclusive for both the Min and Max
type Int32Range struct {
	Min int32
	Max int32
}

func (r *Int32Range) Contains(i int32) bool {
	return i >= r.Min && i <= r.Max
}
