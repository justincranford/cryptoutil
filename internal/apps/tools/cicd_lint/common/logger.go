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
type Logger struct {
	startTime time.Time
	operation string
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

// Log outputs a message with duration and timestamp information.
func (l *Logger) Log(message string) {
	now := time.Now().UTC()
	fmt.Fprintf(os.Stderr, "[CICD] dur=%v now=%s: %s\n",
		now.Sub(l.startTime),
		now.Format(cryptoutilSharedMagic.TimeFormat),
		message)
}

// LogError outputs an error message with duration and timestamp.
func (l *Logger) LogError(err error) {
	now := time.Now().UTC()
	fmt.Fprintf(os.Stderr, "[CICD] dur=%v now=%s ERROR: %v\n",
		now.Sub(l.startTime),
		now.Format(cryptoutilSharedMagic.TimeFormat),
		err)
}

// LogWithPrefix outputs a message with a custom prefix.
func (l *Logger) LogWithPrefix(prefix, message string) {
	now := time.Now().UTC()
	fmt.Fprintf(os.Stderr, "[CICD] dur=%v now=%s %s: %s\n",
		now.Sub(l.startTime),
		now.Format(cryptoutilSharedMagic.TimeFormat),
		prefix,
		message)
}

// Duration returns the elapsed time since the logger was created.
func (l *Logger) Duration() time.Duration {
	return time.Since(l.startTime)
}

// Operation returns the operation name for this logger.
func (l *Logger) Operation() string {
	return l.operation
}
