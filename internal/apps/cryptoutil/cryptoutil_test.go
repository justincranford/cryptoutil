// Copyright (c) 2025 Justin Cranford
//
//

package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrintUsage(t *testing.T) {
	t.Parallel()
	// Test that printUsage doesn't panic
	require.NotPanics(t, func() {
		printUsage("test-executable")
	})
}
