package certificate

// prepend adds an element to the beginning of a slice and returns the new slice
func prepend[T any](slice []T, item T) []T {
	return append([]T{item}, slice...)
}
