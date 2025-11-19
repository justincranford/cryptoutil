// Copyright (c) 2025 Justin Cranford
//
//

package cicd

import (
	"testing"

	cryptoutilMagic "cryptoutil/internal/common/magic"

	"github.com/stretchr/testify/require"
)

// Test regex patterns for validation.
var (
	testErrorfValidationPattern  = cryptoutilMagic.TestErrorfValidationPattern
	testFErrorfValidationPattern = cryptoutilMagic.TestFErrorfValidationPattern
	testFatalfValidationPattern  = cryptoutilMagic.TestFatalfValidationPattern
)

func TestGoEnforceTestPatterns_RegexValidation(t *testing.T) {
	// Test the regex patterns used in checkTestFile to ensure they work correctly
	// This was originally created as a one-off test during chat session
	// Test t.Errorf pattern
	t.Logf("Compiled regex pattern: %s", `t\.Errorf\([^)]+\)`)

	// Debug: test with f.Errorf pattern
	testString1 := `fmt.Errorf("failed to create pools: %w", errors.Join(err1, err2))`
	fMatches := testFErrorfValidationPattern.FindAllString(testString1, -1)
	t.Logf("F pattern matches for string: %s", testString1)
	t.Logf("F pattern matches found: %v", fMatches)

	// Should match t.Errorf calls
	require.True(t, testErrorfValidationPattern.MatchString(`t.Errorf("test failed: %v", err)`), "Should match t.Errorf call")
	require.True(t, testErrorfValidationPattern.MatchString(`t.Errorf("expected %d, got %d", expected, actual)`), "Should match t.Errorf with multiple args")

	// Should NOT match fmt.Errorf calls (these are legitimate error creation)
	matches1 := testErrorfValidationPattern.FindAllString(testString1, -1)
	t.Logf("T pattern matches for string: %s", testString1)
	t.Logf("T pattern matches found: %v", matches1)
	require.False(t, testErrorfValidationPattern.MatchString(testString1), "Should NOT match fmt.Errorf call")

	testString3 := `var x = 1`
	matches3 := testErrorfValidationPattern.FindAllString(testString3, -1)
	t.Logf("Testing string 3: %s", testString3)
	t.Logf("Regex matches found: %v", matches3)
	require.False(t, testErrorfValidationPattern.MatchString(testString3), "Should NOT match simple assignment")

	// Test t.Fatalf pattern

	// Should match t.Fatalf calls
	require.True(t, testFatalfValidationPattern.MatchString(`t.Fatalf("failed to parse date: %v", err)`), "Should match t.Fatalf call")
	require.True(t, testFatalfValidationPattern.MatchString(`t.Fatalf("expected error, got nil")`), "Should match t.Fatalf with simple message")

	// Should NOT match other patterns
	require.False(t, testFatalfValidationPattern.MatchString(`fmt.Errorf("some error")`), "Should NOT match fmt.Errorf")
	require.False(t, testFatalfValidationPattern.MatchString(`t.Errorf("some error")`), "Should NOT match t.Errorf")
}
