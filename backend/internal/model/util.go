package model

// PtrString returns a pointer to the given string.
// Useful for setting nullable *string fields from literal values.
func PtrString(s string) *string {
	return &s
}
