package device

// As is a generic type assertion: it returns the device as role R, or false
// when the device doesn't implement it.
func As[R any](d Device) (R, bool) {
	r, ok := d.(R)
	return r, ok
}
