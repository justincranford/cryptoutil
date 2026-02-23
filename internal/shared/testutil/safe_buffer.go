// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

package testutil

import (
	"bytes"
	"fmt"
	"sync"
)

// SafeBuffer is a thread-safe bytes.Buffer for concurrent read/write in tests.
// Use this when a goroutine writes to a buffer while the test goroutine reads it
// (e.g., polling for server startup messages via require.Eventually).
type SafeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

// Write appends p to the buffer (thread-safe).
func (sb *SafeBuffer) Write(p []byte) (n int, err error) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	n, err = sb.buf.Write(p)
	if err != nil {
		return n, fmt.Errorf("safe buffer write failed: %w", err)
	}

	return n, nil
}

// String returns the buffer contents as a string (thread-safe).
func (sb *SafeBuffer) String() string {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	return sb.buf.String()
}

// Reset clears the buffer (thread-safe).
func (sb *SafeBuffer) Reset() {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	sb.buf.Reset()
}

// Len returns the number of bytes in the buffer (thread-safe).
func (sb *SafeBuffer) Len() int {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	return sb.buf.Len()
}
