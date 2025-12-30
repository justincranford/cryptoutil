// Copyright (c) 2025 Justin Cranford
//
//

package im

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestPrintIMVersion tests the IM version output.
func TestPrintIMVersion(t *testing.T) {
	output := captureOutput(t, func() {
		exitCode := IM([]string{"version"})
		require.Equal(t, 0, exitCode)
	})

	require.Contains(t, output, "learn-im service")
	require.Contains(t, output, "cryptoutil learn product")
}
