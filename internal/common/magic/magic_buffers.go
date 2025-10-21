// Package magic provides commonly used magic numbers and values as named constants.
// This file contains buffer sizes and memory allocation constants.
// This package centralizes magic values to avoid linter violations and improve code maintainability.
// All constants are grouped by category for better organization.
package magic

// Buffer sizes and memory allocations.
const (
	// BufferSize1KB - 1KB buffer size, common memory allocation.
	BufferSize1KB = 1024
	// BufferSize2KB - 2KB buffer size, RSA-2048 key size.
	BufferSize2KB = 2048
	// BufferSize4KB - 4KB buffer size, RSA-4096 key size.
	BufferSize4KB = 4096
)
