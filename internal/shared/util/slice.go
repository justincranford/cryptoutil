// Copyright (c) 2025 Justin Cranford
//
//

// Package util provides utility functions for common operations.
package util

// Contains checks if a slice contains the specified item.
func Contains[T comparable](slice []*T, item *T) bool {
	for _, element := range slice {
		if *element == *item {
			return true
		}
	}

	return false
}
