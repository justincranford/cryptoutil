// Copyright (c) 2025 Justin Cranford
//
//

// Package demo provides progress display utilities for demo CLI.
package demo

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// ProgressDisplay manages progress output during demo execution.
type ProgressDisplay struct {
	writer    io.Writer
	noColor   bool
	quiet     bool
	verbose   bool
	startTime time.Time
	mu        sync.Mutex
	stepCount int
	stepTotal int
}

// NewProgressDisplay creates a new progress display.
func NewProgressDisplay(config *Config) *ProgressDisplay {
	return &ProgressDisplay{
		writer:    os.Stdout,
		noColor:   config.NoColor,
		quiet:     config.Quiet,
		verbose:   config.Verbose,
		startTime: time.Now().UTC(),
		stepCount: 0,
		stepTotal: 0,
	}
}

// SetTotalSteps sets the total number of steps for progress tracking.
func (p *ProgressDisplay) SetTotalSteps(total int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.stepTotal = total
}

// StartStep begins a new step with progress indicator.
func (p *ProgressDisplay) StartStep(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.stepCount++

	if p.quiet {
		return
	}

	progress := ""
	if p.stepTotal > 0 {
		progress = fmt.Sprintf("[%d/%d] ", p.stepCount, p.stepTotal)
	}

	if p.noColor {
		_, _ = fmt.Fprintf(p.writer, "%s%s...\n", progress, name)
	} else {
		_, _ = fmt.Fprintf(p.writer, "⏳ %s%s...\n", progress, name)
	}
}

// CompleteStep marks a step as completed.
func (p *ProgressDisplay) CompleteStep(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.quiet {
		return
	}

	if p.noColor {
		_, _ = fmt.Fprintf(p.writer, "  [OK] %s\n", name)
	} else {
		_, _ = fmt.Fprintf(p.writer, "  ✅ %s\n", name)
	}
}

// FailStep marks a step as failed.
func (p *ProgressDisplay) FailStep(name string, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.noColor {
		_, _ = fmt.Fprintf(p.writer, "  [FAIL] %s: %v\n", name, err)
	} else {
		_, _ = fmt.Fprintf(p.writer, "  ❌ %s: %v\n", name, err)
	}
}

// SkipStep marks a step as skipped.
func (p *ProgressDisplay) SkipStep(name string, reason string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.quiet {
		return
	}

	if p.noColor {
		_, _ = fmt.Fprintf(p.writer, "  [SKIP] %s: %s\n", name, reason)
	} else {
		_, _ = fmt.Fprintf(p.writer, "  ⏭️ %s: %s\n", name, reason)
	}
}

// Info prints an informational message.
func (p *ProgressDisplay) Info(format string, args ...any) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.quiet {
		return
	}

	if p.noColor {
		_, _ = fmt.Fprintf(p.writer, "[INFO] "+format+"\n", args...)
	} else {
		_, _ = fmt.Fprintf(p.writer, "ℹ️ "+format+"\n", args...)
	}
}

// Debug prints a debug message (only in verbose mode).
func (p *ProgressDisplay) Debug(format string, args ...any) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.verbose {
		return
	}

	if p.noColor {
		_, _ = fmt.Fprintf(p.writer, "[DEBUG] "+format+"\n", args...)
	} else {
		_, _ = fmt.Fprintf(p.writer, "🔍 "+format+"\n", args...)
	}
}

// Warn prints a warning message.
func (p *ProgressDisplay) Warn(format string, args ...any) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.noColor {
		_, _ = fmt.Fprintf(p.writer, "[WARN] "+format+"\n", args...)
	} else {
		_, _ = fmt.Fprintf(p.writer, "⚠️ "+format+"\n", args...)
	}
}

// Error prints an error message.
func (p *ProgressDisplay) Error(format string, args ...any) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.noColor {
		_, _ = fmt.Fprintf(p.writer, "[ERROR] "+format+"\n", args...)
	} else {
		_, _ = fmt.Fprintf(p.writer, "❌ "+format+"\n", args...)
	}
}

// PrintSummary prints a summary of the demo execution.
func (p *ProgressDisplay) PrintSummary(result *DemoResult) {
	p.mu.Lock()
	defer p.mu.Unlock()

	duration := time.Since(p.startTime)

	_, _ = fmt.Fprintln(p.writer)

	if p.noColor {
		_, _ = fmt.Fprintln(p.writer, "=== Demo Summary ===")
	} else {
		_, _ = fmt.Fprintln(p.writer, "📊 Demo Summary")
		_, _ = fmt.Fprintln(p.writer, "================")
	}

	_, _ = fmt.Fprintf(p.writer, "Duration: %s\n", duration.Round(time.Millisecond))
	_, _ = fmt.Fprintf(p.writer, "Steps: %d total, %d passed, %d failed, %d skipped\n",
		result.TotalSteps, result.PassedSteps, result.FailedSteps, result.SkippedSteps)

	if result.FailedSteps > 0 {
		_, _ = fmt.Fprintln(p.writer)

		if p.noColor {
			_, _ = fmt.Fprintln(p.writer, "Failed Steps:")
		} else {
			_, _ = fmt.Fprintln(p.writer, "❌ Failed Steps:")
		}

		for _, err := range result.Errors {
			_, _ = fmt.Fprintf(p.writer, "  - %s\n", err)
		}
	}

	_, _ = fmt.Fprintln(p.writer)

	if result.Success {
		if p.noColor {
			_, _ = fmt.Fprintln(p.writer, "[SUCCESS] Demo completed successfully!")
		} else {
			_, _ = fmt.Fprintln(p.writer, "✅ Demo completed successfully!")
		}
	} else if result.FailedSteps > 0 && result.PassedSteps > 0 {
		if p.noColor {
			_, _ = fmt.Fprintln(p.writer, "[PARTIAL] Demo completed with some failures")
		} else {
			_, _ = fmt.Fprintln(p.writer, "⚠️ Demo completed with some failures")
		}
	} else {
		if p.noColor {
			_, _ = fmt.Fprintln(p.writer, "[FAILURE] Demo failed")
		} else {
			_, _ = fmt.Fprintln(p.writer, "❌ Demo failed")
		}
	}
}

// Spinner provides a simple spinner for long-running operations.
type Spinner struct {
	frames   []string
	index    int
	interval time.Duration
	stop     chan struct{}
	running  bool
	mu       sync.Mutex
}

// NewSpinner creates a new spinner.
func NewSpinner() *Spinner {
	return &Spinner{
		frames:   []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		interval: cryptoutilSharedMagic.DefaultSpinnerInterval,
		stop:     make(chan struct{}),
	}
}

// Start starts the spinner.
func (s *Spinner) Start(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return
	}

	s.running = true
	s.stop = make(chan struct{})

	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-s.stop:
				fmt.Print("\r\033[K") // Clear line

				return
			case <-ticker.C:
				s.mu.Lock()
				frame := s.frames[s.index%len(s.frames)]
				s.index++
				s.mu.Unlock()

				fmt.Printf("\r%s %s", frame, message)
			}
		}
	}()
}

// Stop stops the spinner.
func (s *Spinner) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.running = false

	close(s.stop)
}
