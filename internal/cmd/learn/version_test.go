// Copyright (c) 2025 Justin Cranford
//
//

package learn_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	learn "cryptoutil/internal/cmd/learn"
	"cryptoutil/internal/cmd/learn/im"
)

// captureOutput captures stdout during function execution.
func captureOutput(t *testing.T, fn func()) string {
	t.Helper()

	originalStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)

	os.Stdout = w

	outChan := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outChan <- buf.String()
	}()

	fn()

	err = w.Close()
	require.NoError(t, err)
	os.Stdout = originalStdout

	return <-outChan
}

// TestPrintIMVersion tests the IM version output.
func TestPrintIMVersion(t *testing.T) {
	output := captureOutput(t, func() {
		exitCode := im.IM([]string{"version"})
		require.Equal(t, 0, exitCode)
	})

	require.Contains(t, output, "learn-im service")
	require.Contains(t, output, "cryptoutil learn product")
}

// TestPrintLearnVersion tests the Learn version output.
func TestPrintLearnVersion(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "version_command",
			args: []string{"version"},
		},
		{
			name: "version_flag_long",
			args: []string{"--version"},
		},
		{
			name: "version_flag_short",
			args: []string{"-v"},
		},
	}

	for _, tt := range tests {
		// Capture range variable for parallel tests.
		t.Run(tt.name, func(t *testing.T) {
			// Remove t.Parallel() - output/output capture has race condition with parallel tests.
			// TODO: Investigate safer capture method that works with t.Parallel().
			output := captureOutput(t, func() {
				exitCode := learn.Learn(tt.args)
				require.Equal(t, 0, exitCode)
			})

			combinedOutput := output + output
			require.Contains(t, combinedOutput, "learn product")
		})
	}
}
