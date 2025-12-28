// Copyright (c) 2025 Justin Cranford
//
//

package learn_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	learn "cryptoutil/internal/cmd/learn"
	"cryptoutil/internal/cmd/learn/im"
)

// TestPrintIMVersion tests the IM version output.
func TestPrintIMVersion(t *testing.T) {
	stdout, _ := captureOutput(t, func() {
		exitCode := im.IM([]string{"version"})
		require.Equal(t, 0, exitCode)
	})

	require.Contains(t, stdout, "learn-im service")
	require.Contains(t, stdout, "cryptoutil learn product")
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
			// Remove t.Parallel() - stdout/stderr capture has race condition with parallel tests.
			// TODO: Investigate safer capture method that works with t.Parallel().
			stdout, stderr := captureOutput(t, func() {
				exitCode := learn.Learn(tt.args)
				require.Equal(t, 0, exitCode)
			})

			combinedOutput := stdout + stderr
			require.Contains(t, combinedOutput, "learn product")
		})
	}
}
