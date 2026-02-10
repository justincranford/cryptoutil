// Copyright (c) 2025 Justin Cranford
//
//

package cryptoutil

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrintUsage(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	// Test that printUsage doesn't panic
	require.NotPanics(t, func() {
		printUsage(&stderr)
	})

	// Verify output contains expected content
	output := stderr.String()
	require.Contains(t, output, "Usage: cryptoutil")
	require.Contains(t, output, "Available products:")
}
