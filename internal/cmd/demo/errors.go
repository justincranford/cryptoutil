// Copyright (c) 2025 Justin Cranford
//
//

// Package demo provides error aggregation for demo CLI.
package demo

import (
	json "encoding/json"
	"errors"
	"fmt"
	"strings"
)

// DemoResult holds the result of a demo execution.
type DemoResult struct {
	// Success is true if all steps completed without errors.
	Success bool `json:"success"`

	// TotalSteps is the total number of steps executed.
	TotalSteps int `json:"total_steps"`

	// PassedSteps is the number of steps that passed.
	PassedSteps int `json:"passed_steps"`

	// FailedSteps is the number of steps that failed.
	FailedSteps int `json:"failed_steps"`

	// SkippedSteps is the number of steps that were skipped.
	SkippedSteps int `json:"skipped_steps"`

	// Errors contains all errors encountered during execution.
	Errors []DemoError `json:"errors,omitempty"`

	// Duration is the execution duration in milliseconds.
	DurationMS int64 `json:"duration_ms"`
}

// ExitCode returns the appropriate exit code based on the result.
func (r *DemoResult) ExitCode() int {
	if r.Success {
		return ExitSuccess
	}

	if r.PassedSteps > 0 {
		return ExitPartialFailure
	}

	return ExitFailure
}

// DemoError represents an error that occurred during demo execution.
type DemoError struct {
	// Step is the step name where the error occurred.
	Step string `json:"step"`

	// Phase is the phase name (e.g., "kms", "identity", "integration").
	Phase string `json:"phase"`

	// Message is the error message.
	Message string `json:"message"`

	// Details contains additional error details.
	Details string `json:"details,omitempty"`

	// Cause is the underlying cause if available.
	Cause *DemoError `json:"cause,omitempty"`
}

// Error implements the error interface.
func (e DemoError) Error() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("[%s/%s] %s", e.Phase, e.Step, e.Message))

	if e.Details != "" {
		sb.WriteString(": ")
		sb.WriteString(e.Details)
	}

	return sb.String()
}

// Unwrap returns the underlying cause.
func (e DemoError) Unwrap() error {
	if e.Cause == nil {
		return nil
	}

	return *e.Cause
}

// NewDemoError creates a new DemoError.
func NewDemoError(phase, step, message string) DemoError {
	return DemoError{
		Phase:   phase,
		Step:    step,
		Message: message,
	}
}

// WithDetails adds details to the error.
func (e DemoError) WithDetails(details string) DemoError {
	e.Details = details

	return e
}

// WithCause adds a cause to the error.
func (e DemoError) WithCause(cause error) DemoError {
	var demoErr DemoError
	if errors.As(cause, &demoErr) {
		e.Cause = &demoErr
	} else if cause != nil {
		e.Cause = &DemoError{
			Message: cause.Error(),
		}
	}

	return e
}

// ErrorAggregator collects errors during demo execution.
type ErrorAggregator struct {
	errors []DemoError
	phase  string
}

// NewErrorAggregator creates a new error aggregator.
func NewErrorAggregator(phase string) *ErrorAggregator {
	return &ErrorAggregator{
		errors: make([]DemoError, 0),
		phase:  phase,
	}
}

// Add adds an error to the aggregator.
func (a *ErrorAggregator) Add(step, message string, err error) {
	demoErr := NewDemoError(a.phase, step, message)

	if err != nil {
		demoErr = demoErr.WithCause(err)
	}

	a.errors = append(a.errors, demoErr)
}

// AddError adds a DemoError directly.
func (a *ErrorAggregator) AddError(err DemoError) {
	a.errors = append(a.errors, err)
}

// HasErrors returns true if any errors were collected.
func (a *ErrorAggregator) HasErrors() bool {
	return len(a.errors) > 0
}

// Errors returns all collected errors.
func (a *ErrorAggregator) Errors() []DemoError {
	return a.errors
}

// Count returns the number of errors.
func (a *ErrorAggregator) Count() int {
	return len(a.errors)
}

// ToResult creates a DemoResult from the aggregator.
func (a *ErrorAggregator) ToResult(passed, skipped int) *DemoResult {
	failed := len(a.errors)
	total := passed + failed + skipped

	return &DemoResult{
		Success:      failed == 0,
		TotalSteps:   total,
		PassedSteps:  passed,
		FailedSteps:  failed,
		SkippedSteps: skipped,
		Errors:       a.errors,
	}
}

// OutputFormatter formats demo output.
type OutputFormatter struct {
	format OutputFormat
}

// NewOutputFormatter creates a new output formatter.
func NewOutputFormatter(format OutputFormat) *OutputFormatter {
	return &OutputFormatter{format: format}
}

// FormatResult formats a DemoResult based on the output format.
func (f *OutputFormatter) FormatResult(result *DemoResult) string {
	switch f.format {
	case OutputJSON:
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Sprintf(`{"error": "failed to marshal result: %v"}`, err)
		}

		return string(data)

	case OutputStructured:
		return f.formatStructured(result)

	default:
		return f.formatHuman(result)
	}
}

func (f *OutputFormatter) formatStructured(result *DemoResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("level=info msg=\"demo completed\" success=%t total=%d passed=%d failed=%d skipped=%d duration_ms=%d\n",
		result.Success, result.TotalSteps, result.PassedSteps, result.FailedSteps, result.SkippedSteps, result.DurationMS))

	for _, err := range result.Errors {
		sb.WriteString(fmt.Sprintf("level=error msg=\"step failed\" phase=%s step=%s error=%q\n",
			err.Phase, err.Step, err.Message))
	}

	return sb.String()
}

func (f *OutputFormatter) formatHuman(result *DemoResult) string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("📊 Demo Summary\n")
	sb.WriteString("================\n")
	sb.WriteString(fmt.Sprintf("Duration: %dms\n", result.DurationMS))
	sb.WriteString(fmt.Sprintf("Steps: %d total, %d passed, %d failed, %d skipped\n",
		result.TotalSteps, result.PassedSteps, result.FailedSteps, result.SkippedSteps))

	if result.FailedSteps > 0 {
		sb.WriteString("\n❌ Failed Steps:\n")

		for _, err := range result.Errors {
			sb.WriteString(fmt.Sprintf("  - [%s/%s] %s\n", err.Phase, err.Step, err.Message))
		}
	}

	sb.WriteString("\n")

	if result.Success {
		sb.WriteString("✅ Demo completed successfully!\n")
	} else if result.FailedSteps > 0 && result.PassedSteps > 0 {
		sb.WriteString("⚠️ Demo completed with some failures\n")
	} else {
		sb.WriteString("❌ Demo failed\n")
	}

	return sb.String()
}
