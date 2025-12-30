// Copyright (c) 2025 Justin Cranford
//
//

package im

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// captureOutput is a test helper that captures stdout/stderr during function execution.
func captureOutput(t *testing.T, fn func()) string {
	t.Helper()

	// Create pipes to capture output.
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	rStdout, wStdout, err := os.Pipe()
	require.NoError(t, err)

	rStderr, wStderr, err := os.Pipe()
	require.NoError(t, err)

	os.Stdout = wStdout
	os.Stderr = wStderr

	// Channel for synchronization.
	done := make(chan struct{})

	// Buffer for output.
	var bufStdout, bufStderr bytes.Buffer

	// Start goroutine to read from pipes.
	go func() {
		_, _ = bufStdout.ReadFrom(rStdout)
		_, _ = bufStderr.ReadFrom(rStderr)

		close(done)
	}()

	// Execute function.
	fn()

	// Close write ends to signal EOF to read goroutine.
	_ = wStdout.Close()
	_ = wStderr.Close()

	// Wait for read goroutine to finish.
	<-done

	// Restore stdout/stderr.
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Combine stdout and stderr.
	return bufStdout.String() + bufStderr.String()
}

// TestIMClient_HelpFlag tests imClient() help flag output.
func TestIMClient_HelpFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "help_command", args: []string{"help"}},
		{name: "help_flag_long", args: []string{"--help"}},
		{name: "help_flag_short", args: []string{"-h"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output.
			output := captureOutput(t, func() {
				exitCode := imClient(tt.args)
				require.Equal(t, 0, exitCode, "Expected exit code 0 for help command")
			})

			// Verify help output contains expected keywords.
			require.Contains(t, output, "client", "Help output should contain 'client'")
			require.Contains(t, output, "Usage", "Help output should contain 'Usage'")
		})
	}
}

// TestIMInit_HelpFlag tests imInit() help flag output.
func TestIMInit_HelpFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "help_command", args: []string{"help"}},
		{name: "help_flag_long", args: []string{"--help"}},
		{name: "help_flag_short", args: []string{"-h"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output.
			output := captureOutput(t, func() {
				exitCode := imInit(tt.args)
				require.Equal(t, 0, exitCode, "Expected exit code 0 for help command")
			})

			// Verify help output contains expected keywords.
			require.Contains(t, output, "init", "Help output should contain 'init'")
			require.Contains(t, output, "Usage", "Help output should contain 'Usage'")
		})
	}
}

// TestIMHealth_HelpFlag tests imHealth() help flag output.
func TestIMHealth_HelpFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "help_command", args: []string{"help"}},
		{name: "help_flag_long", args: []string{"--help"}},
		{name: "help_flag_short", args: []string{"-h"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output.
			output := captureOutput(t, func() {
				exitCode := imHealth(tt.args)
				require.Equal(t, 0, exitCode, "Expected exit code 0 for help command")
			})

			// Verify help output contains expected keywords.
			require.Contains(t, output, "health", "Help output should contain 'health'")
			require.Contains(t, output, "Usage", "Help output should contain 'Usage'")
		})
	}
}

// TestIMLivez_HelpFlag tests imLivez() help flag output.
func TestIMLivez_HelpFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "help_command", args: []string{"help"}},
		{name: "help_flag_long", args: []string{"--help"}},
		{name: "help_flag_short", args: []string{"-h"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output.
			output := captureOutput(t, func() {
				exitCode := imLivez(tt.args)
				require.Equal(t, 0, exitCode, "Expected exit code 0 for help command")
			})

			// Verify help output contains expected keywords.
			require.Contains(t, output, "livez", "Help output should contain 'livez'")
			require.Contains(t, output, "Usage", "Help output should contain 'Usage'")
		})
	}
}

// TestIMReadyz_HelpFlag tests imReadyz() help flag output.
func TestIMReadyz_HelpFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "help_command", args: []string{"help"}},
		{name: "help_flag_long", args: []string{"--help"}},
		{name: "help_flag_short", args: []string{"-h"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output.
			output := captureOutput(t, func() {
				exitCode := imReadyz(tt.args)
				require.Equal(t, 0, exitCode, "Expected exit code 0 for help command")
			})

			// Verify help output contains expected keywords.
			require.Contains(t, output, "readyz", "Help output should contain 'readyz'")
			require.Contains(t, output, "Usage", "Help output should contain 'Usage'")
		})
	}
}

// TestIMShutdown_HelpFlag tests imShutdown() help flag output.
func TestIMShutdown_HelpFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "help_command", args: []string{"help"}},
		{name: "help_flag_long", args: []string{"--help"}},
		{name: "help_flag_short", args: []string{"-h"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output.
			output := captureOutput(t, func() {
				exitCode := imShutdown(tt.args)
				require.Equal(t, 0, exitCode, "Expected exit code 0 for help command")
			})

			// Verify help output contains expected keywords.
			require.Contains(t, output, "shutdown", "Help output should contain 'shutdown'")
			require.Contains(t, output, "Usage", "Help output should contain 'Usage'")
		})
	}
}
