package integration

// boolPtr returns a pointer to the given bool value.
// Used for domain model fields that are *bool type.
func boolPtr(b bool) *bool {
	return &b
}
