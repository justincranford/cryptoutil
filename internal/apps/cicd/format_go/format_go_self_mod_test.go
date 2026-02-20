// Copyright (c) 2025 Justin Cranford

package format_go

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestEnforceAnyDoesNotModifyItself verifies that enforce_any.go is properly excluded
// from the enforcement process and will never modify its own comments or code.
//
// This test addresses historical self-modification regressions:
// - b934879b (Nov 17): Comments modified by pattern replacement
// - b0e4b6ef (Dec 16): Counting logic incorrectly used "any" instead of "interface{}"
// - 8c855a6e (Dec 16): Test data incorrectly used "any" instead of "interface{}"
//
// Protection relies on:
// 1. CICDSelfExclusionPatterns["format-go"] excluding internal/cmd/cicd/format_go/**
// 2. CRITICAL comments in processGoFile() documenting self-modification risk
// 3. Test data using interface{} as input to verify replacement works.
func TestEnforceAnyDoesNotModifyItself(t *testing.T) {
	t.Parallel()

	// Read original content of enforce_any.go.
	originalContent, err := os.ReadFile("enforce_any/enforce_any.go")
	require.NoError(t, err, "Failed to read enforce_any.go")

	// Verify the file contains critical self-modification protection markers.
	require.Contains(t, string(originalContent), "SELF-MODIFICATION PROTECTION:",
		"enforce_any.go MUST contain SELF-MODIFICATION PROTECTION comment block")
	require.Contains(t, string(originalContent), "CRITICAL: Replace interface{} with any",
		"enforce_any.go MUST contain CRITICAL comment explaining pattern replacement")
	require.Contains(t, string(originalContent), `strings.Count(originalContent, "interface{}")`,
		"enforce_any.go MUST count interface{} occurrences, NOT any")

	// Verify test data uses interface{} (not any) to properly test replacement.
	testContent, err := os.ReadFile("enforce_any/enforce_any_test.go")
	require.NoError(t, err, "Failed to read enforce_any/enforce_any_test.go")

	// Check test constants use interface{} as input data.
	require.Contains(t, string(testContent), `testGoContentWithInterfaceEmpty = "package main\n\nfunc main() {\n\tvar x interface{}`,
		"Test constants MUST use interface{} as input data, NOT any")

	// Verify test expectations check for 'any' after replacement.
	require.Contains(t, string(testContent), `"File should contain 'any' after replacement"`,
		"Tests MUST verify 'any' appears after replacement")
	require.Contains(t, string(testContent), `"File should not contain 'interface{}' after replacement"`,
		"Tests MUST verify 'interface{}' removed after replacement")
}
