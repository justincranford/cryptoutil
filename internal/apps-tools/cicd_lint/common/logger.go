// Copyright (c) 2025 Justin Cranford
//
//

package common

import (
	"fmt"
	"os"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Logger provides structured logging for CICD commands with timing information.
// When quiet is true, Log and LogWithPrefix are no-ops; LogError always writes.
type Logger struct {
	startTime time.Time
	operation string
	quiet     bool
}

// NewLogger creates a new logger for the specified operation.
// It logs the start time immediately.
func NewLogger(operation string) *Logger {
	start := time.Now().UTC()
	fmt.Fprintf(os.Stderr, "[CICD] start=%s operation=%s\n",
		start.Format(cryptoutilSharedMagic.TimeFormat),
		operation)

	return &Logger{
		startTime: start,
		operation: operation,
	}
}

// NewQuietLogger creates a silent logger that suppresses Log and LogWithPrefix output.
// LogError always writes regardless of quiet mode so errors remain visible.
func NewQuietLogger(operation string) *Logger {
	return &Logger{
		startTime: time.Now().UTC(),
		operation: operation,
		quiet:     true,
	}
}

// Log outputs a message with duration and timestamp information.
// No-op when logger is in quiet mode.
func (l *Logger) Log(message string) {
	if l.quiet {
		return
	}

	now := time.Now().UTC()
	fmt.Fprintf(os.Stderr, "[CICD] dur=%v now=%s: %s\n",
		now.Sub(l.startTime),
		now.Format(cryptoutilSharedMagic.TimeFormat),
		message)
}

// LogError outputs an error message with duration and timestamp.
// Always writes even in quiet mode.
func (l *Logger) LogError(err error) {
	now := time.Now().UTC()
	fmt.Fprintf(os.Stderr, "[CICD] dur=%v now=%s ERROR: %v\n",
		now.Sub(l.startTime),
		now.Format(cryptoutilSharedMagic.TimeFormat),
		err)
}

// LogWithPrefix outputs a message with a custom prefix.
// No-op when logger is in quiet mode.
func (l *Logger) LogWithPrefix(prefix, message string) {
	if l.quiet {
		return
	}

	now := time.Now().UTC()
	fmt.Fprintf(os.Stderr, "[CICD] dur=%v now=%s %s: %s\n",
		now.Sub(l.startTime),
		now.Format(cryptoutilSharedMagic.TimeFormat),
		prefix,
		message)
}

// IsQuiet returns true if this logger suppresses verbose output.
func (l *Logger) IsQuiet() bool {
	return l.quiet
}

// Duration returns the elapsed time since the logger was created.
func (l *Logger) Duration() time.Duration {
	return time.Since(l.startTime)
}

// Operation returns the operation name for this logger.
func (l *Logger) Operation() string {
	return l.operation
}
