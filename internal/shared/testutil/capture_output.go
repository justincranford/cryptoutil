// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

package testutil

import (
	"bytes"
	"io"
	"os"
	"sync"
	"testing"
)

// captureOutputMutex protects concurrent access to os.Stdout/os.Stderr.
var captureOutputMutex sync.Mutex //nolint:gochecknoglobals // Required for thread-safe output capture.

// CaptureOutput captures stdout and stderr during function execution.
// Thread-safe for parallel tests.
func CaptureOutput(t *testing.T, fn func()) string {
	t.Helper()

	captureOutputMutex.Lock()
	defer captureOutputMutex.Unlock()

	// Save original stdout/stderr.
	originalStdout := os.Stdout
	originalStderr := os.Stderr

	// Create pipes to capture output.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}

	// Redirect stdout/stderr to pipe.
	os.Stdout = w
	os.Stderr = w

	// Channel to capture output.
	outputChan := make(chan string, 1)

	// Read from pipe in background.
	go func() {
		var buf bytes.Buffer

		_, _ = io.Copy(&buf, r)
		outputChan <- buf.String()
	}()

	// Execute function.
	fn()

	// Close writer and restore original stdout/stderr.
	_ = w.Close()
	os.Stdout = originalStdout
	os.Stderr = originalStderr

	// Get captured output.
	output := <-outputChan
	_ = r.Close()

	return output
}
