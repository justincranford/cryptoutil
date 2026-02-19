// Copyright (c) 2025 Justin Cranford

//go:build e2e

package test

import (
	"fmt"
	"os"
	"strings"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
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
func Log(logger *Logger, format string, args ...any) {
	if logger == nil {
		return
	}

	message := fmt.Sprintf("[%s] [%v] %s\n",
		time.Now().UTC().Format("15:04:05"),
		time.Since(logger.startTime).Round(time.Second),
		fmt.Sprintf(format, args...))

	// Write to console
	fmt.Print(message)

	// Write to log file if available
	if logger.logFile != nil {
		if _, err := logger.logFile.WriteString(message); err != nil {
			// If we can't write to the log file, at least write to console
			fmt.Printf("⚠️ Failed to write to log file: %v\n", err)
		}
	}
}

// LogCommand provides structured logging for commands with their output.
func LogCommand(logger *Logger, description, command, output string) {
	if logger == nil {
		return
	}

	Log(logger, "📋 [%s] %s", description, command)

	if output != "" {
		Log(logger, "📋 [%s] Output: %s", description, strings.TrimSpace(output))
	}
}

// LogTestStep provides structured logging for test steps with timestamp and elapsed time.
func LogTestStep(logger *Logger, name, description string) {
	if logger == nil {
		return
	}

	Log(logger, "📋 %s: %s", name, description)
}

// LogTestStepCompletion provides structured logging for test step completion with status and timing.
func LogTestStepCompletion(logger *Logger, statusEmoji, name, result string, duration time.Duration) {
	if logger == nil {
		return
	}

	Log(logger, "%s %s: %s (took %v)", statusEmoji, name, result, duration.Round(time.Millisecond))
}

// LogTestSetup provides structured logging for test setup with timestamp and elapsed time.
func LogTestSetup(logger *Logger, testName string) {
	if logger == nil {
		return
	}

	Log(logger, "📋 Setting up test: %s", testName)
}

// LogTestCleanup provides structured logging for test cleanup with timestamp and elapsed time.
func LogTestCleanup(logger *Logger, testName string) {
	if logger == nil {
		return
	}

	Log(logger, "🧹 Cleaning up test: %s", testName)
}

// LogStep starts tracking a new test step and logs it.
func LogStep(summary *TestSummary, logger *Logger, name, description string) {
	step := TestStep{
		Name:        name,
		StartTime:   time.Now().UTC(),
		Description: description,
	}
	summary.Steps = append(summary.Steps, step)

	// Only log to fixture if logger exists (it won't exist during very early setup)
	if logger != nil {
		LogTestStep(logger, name, description)
	}
}

// CompleteStep marks the current step as completed with a status and logs it.
func CompleteStep(summary *TestSummary, logger *Logger, status, result string) {
	if len(summary.Steps) == 0 {
		return
	}

	step := &summary.Steps[len(summary.Steps)-1]
	step.EndTime = time.Now().UTC()
	step.Duration = step.EndTime.Sub(step.StartTime)
	step.Status = status

	summary.TotalSteps++

	switch status {
	case "PASS":
		summary.PassedSteps++
	case "FAIL":
		summary.FailedSteps++
	case "SKIP":
		summary.SkippedSteps++
	}

	statusEmoji := GetStatusEmoji(status)

	// Only log if logger exists
	if logger != nil {
		LogTestStepCompletion(logger, statusEmoji, step.Name, result, step.Duration)
	}
}

// GetStatusEmoji returns the appropriate emoji for a test status.
func GetStatusEmoji(status string) string {
	switch status {
	case cryptoutilSharedMagic.TestStatusFail:
		return cryptoutilSharedMagic.TestStatusEmojiFail
	case cryptoutilSharedMagic.TestStatusSkip:
		return cryptoutilSharedMagic.TestStatusEmojiSkip
	default:
		return cryptoutilSharedMagic.TestStatusEmojiPass
	}
}

// LogSummaryReport logs a summary report.
func LogSummaryReport(logger *Logger, report string) {
	Log(logger, "%s", report)
}
