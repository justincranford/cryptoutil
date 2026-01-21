// Copyright (c) 2025 Justin Cranford
//
//

//nolint:errcheck // Test infrastructure uses os.Pipe/Close/ReadFrom without error checking
package common

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewLogger(t *testing.T) {
	// Note: Not using t.Parallel() to avoid stderr capture conflicts
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe() //nolint:errcheck // Test infrastructure
	os.Stderr = w

	logger := NewLogger("test-operation")

	_ = w.Close() //nolint:errcheck // Test infrastructure

	os.Stderr = oldStderr

	var buf bytes.Buffer

	_, _ = buf.ReadFrom(r) //nolint:errcheck // Test infrastructure
	output := buf.String()

	require.NotNil(t, logger, "Logger should not be nil")
	require.Equal(t, "test-operation", logger.Operation(), "Operation should match")
	require.Contains(t, output, "[CICD] start=", "Output should contain start marker")
	require.Contains(t, output, "operation=test-operation", "Output should contain operation name")
}

func TestLogger_Log(t *testing.T) {
	// Note: Not using t.Parallel() to avoid stderr capture conflicts
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	logger := NewLogger("test-log")

	time.Sleep(10 * time.Millisecond) // Ensure some duration
	logger.Log("test message")

	w.Close()

	os.Stderr = oldStderr

	var buf bytes.Buffer

	_, _ = buf.ReadFrom(r)
	output := buf.String()

	require.Contains(t, output, "[CICD]", "Output should contain CICD marker")
	require.Contains(t, output, "dur=", "Output should contain duration")
	require.Contains(t, output, "now=", "Output should contain timestamp")
	require.Contains(t, output, "test message", "Output should contain message")
}

func TestLogger_LogError(t *testing.T) {
	// Note: Not using t.Parallel() to avoid stderr capture conflicts
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	logger := NewLogger("test-error")
	testErr := errors.New("test error message")
	logger.LogError(testErr)

	w.Close()

	os.Stderr = oldStderr

	var buf bytes.Buffer

	_, _ = buf.ReadFrom(r)
	output := buf.String()

	require.Contains(t, output, "ERROR:", "Output should contain ERROR marker")
	require.Contains(t, output, "test error message", "Output should contain error message")
}

func TestLogger_LogWithPrefix(t *testing.T) {
	// Note: Not using t.Parallel() to avoid stderr capture conflicts
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	logger := NewLogger("test-prefix")
	logger.LogWithPrefix("CUSTOM", "custom message")

	w.Close()

	os.Stderr = oldStderr

	var buf bytes.Buffer

	_, _ = buf.ReadFrom(r)
	output := buf.String()

	require.Contains(t, output, "CUSTOM:", "Output should contain custom prefix")
	require.Contains(t, output, "custom message", "Output should contain message")
}

func TestLogger_Duration(t *testing.T) {
	t.Parallel()

	logger := NewLogger("test-duration")

	time.Sleep(50 * time.Millisecond)

	duration := logger.Duration()

	require.GreaterOrEqual(t, duration, 50*time.Millisecond, "Duration should be at least 50ms")
	require.Less(t, duration, 200*time.Millisecond, "Duration should be less than 200ms")
}

func TestLogger_Operation(t *testing.T) {
	// Note: Not using t.Parallel() because test modifies global os.Stderr
	tests := []struct {
		name      string
		operation string
	}{
		{"simple operation", "test-op"},
		{"complex operation", "complex-test-operation-name"},
		{"with spaces", "test operation"},
		{"empty", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Not using t.Parallel() - see parent comment
			// Capture stderr to prevent test output pollution
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			logger := NewLogger(tc.operation)

			w.Close()

			os.Stderr = oldStderr

			var buf bytes.Buffer

			_, _ = buf.ReadFrom(r) // Drain pipe

			require.Equal(t, tc.operation, logger.Operation(), "Operation should match input")
		})
	}
}

func TestLogger_MultipleLogs(t *testing.T) {
	// Note: Not using t.Parallel() to avoid stderr capture conflicts
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	logger := NewLogger("test-multiple")
	logger.Log("message 1")
	time.Sleep(10 * time.Millisecond)
	logger.Log("message 2")
	time.Sleep(10 * time.Millisecond)
	logger.Log("message 3")

	w.Close()

	os.Stderr = oldStderr

	var buf bytes.Buffer

	_, _ = buf.ReadFrom(r)
	output := buf.String()

	lines := strings.Split(output, "\n")
	logLines := 0

	for _, line := range lines {
		if strings.Contains(line, "[CICD]") {
			logLines++
		}
	}

	// Should have: 1 start + 3 log messages = 4 lines
	require.GreaterOrEqual(t, logLines, 4, "Should have at least 4 log lines")
}

func TestLogger_ConcurrentLogging(t *testing.T) {
	// Note: Not using t.Parallel() to avoid stderr capture conflicts
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	logger := NewLogger("test-concurrent")

	// Spawn multiple goroutines logging concurrently
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			logger.Log(fmt.Sprintf("concurrent message %d", id))

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	w.Close()

	os.Stderr = oldStderr

	var buf bytes.Buffer

	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Verify all messages appear (may be interleaved)
	for i := 0; i < 10; i++ {
		expectedMsg := fmt.Sprintf("concurrent message %d", i)
		require.Contains(t, output, expectedMsg, "Should contain message %d", i)
	}
}
