package xload

// Mutable is a fragment that wraps a value. This type should be used when the operation
// changes the fragment's value instead of returning a result.
type Mutable struct {
	Value interface{}
}

// NewMutable wraps the input into a mutable.
func NewMutable(in interface{}) Mutable {
	return Mutable{Value: in}
}

// Pick is a noop that returns the input.
func (m Mutable) Pick(in interface{}) interface{} {
	return in
}
