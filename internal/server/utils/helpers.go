package utils

// IntOrZero returns the value of p or 0 if p is nil.
func IntOrZero(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}
