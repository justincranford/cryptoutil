//go:build e2e

package test

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Logger provides structured logging for E2E tests.
type Logger struct {
	startTime time.Time
	logFile   *os.File
}

// NewLogger creates a new logger instance.
func NewLogger(startTime time.Time, logFile *os.File) *Logger {
	return &Logger{
		startTime: startTime,
		logFile:   logFile,
	}
}

// Log provides structured logging with timestamp and elapsed time.
func (l *Logger) Log(format string, args ...any) {
	message := fmt.Sprintf("[%s] [%v] %s\n",
		time.Now().Format("15:04:05"),
		time.Since(l.startTime).Round(time.Second),
		fmt.Sprintf(format, args...))

	// Write to console
	fmt.Print(message)

	// Write to log file if available
	if l.logFile != nil {
		if _, err := l.logFile.WriteString(message); err != nil {
			// If we can't write to the log file, at least write to console
			fmt.Printf("⚠️ Failed to write to log file: %v\n", err)
		}
	}
}

// LogCommand provides structured logging for commands with their output.
func (l *Logger) LogCommand(description, command, output string) {
	l.Log("📋 [%s] %s", description, command)

	if output != "" {
		l.Log("📋 [%s] Output: %s", description, strings.TrimSpace(output))
	}
}

// LogTestStep provides structured logging for test steps with timestamp and elapsed time.
func (l *Logger) LogTestStep(name, description string) {
	l.Log("📋 %s: %s", name, description)
}

// LogTestStepCompletion provides structured logging for test step completion with status and timing.
func (l *Logger) LogTestStepCompletion(statusEmoji, name, result string, duration time.Duration) {
	l.Log("%s %s: %s (took %v)", statusEmoji, name, result, duration.Round(time.Millisecond))
}
