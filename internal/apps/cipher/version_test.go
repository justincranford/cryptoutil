// Copyright (c) 2025 Justin Cranford
//
//

package cipher_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCipherCmd "cryptoutil/internal/apps/cipher"
	"cryptoutil/internal/apps/cipher/im"
)

// TestPrintIMVersion tests the IM version output.
func TestPrintIMVersion(t *testing.T) {
	stdout, _ := captureOutput(t, func() {
		exitCode := im.IM([]string{"version"})
		require.Equal(t, 0, exitCode)
	})

	require.Contains(t, stdout, "cipher-im service")
	require.Contains(t, stdout, "cryptoutil cipher product")
}

// TestPrintCipherVersion tests the Cipher version output.
func TestPrintCipherVersion(t *testing.T) {
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
				exitCode := cryptoutilCipherCmd.Cipher(tt.args)
				require.Equal(t, 0, exitCode)
			})

			combinedOutput := stdout + stderr
			require.Contains(t, combinedOutput, "cipher product")
		})
	}
}
