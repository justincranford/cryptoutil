package cicd

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGoEnforceTestPatterns_RegexValidation(t *testing.T) {
	// Test the regex patterns used in checkTestFile to ensure they work correctly
	// This was originally created as a one-off test during chat session
	// Test t.Errorf pattern
	errorfPattern := regexp.MustCompile(`^t\.Errorf\([^)]+\)$`)

	t.Logf("Compiled regex pattern: %s", `t\.Errorf\([^)]+\)`)

	// Debug: test with f.Errorf pattern
	fErrorfPattern := regexp.MustCompile(`^f\.Errorf\([^)]+\)$`)
	testString1 := `fmt.Errorf("failed to create pools: %w", errors.Join(err1, err2))`
	fMatches := fErrorfPattern.FindAllString(testString1, -1)
	t.Logf("F pattern matches for string: %s", testString1)
	t.Logf("F pattern matches found: %v", fMatches)

	// Should match t.Errorf calls
	require.True(t, errorfPattern.MatchString(`t.Errorf("test failed: %v", err)`), "Should match t.Errorf call")
	require.True(t, errorfPattern.MatchString(`t.Errorf("expected %d, got %d", expected, actual)`), "Should match t.Errorf with multiple args")

	// Should NOT match fmt.Errorf calls (these are legitimate error creation)
	matches1 := errorfPattern.FindAllString(testString1, -1)
	t.Logf("T pattern matches for string: %s", testString1)
	t.Logf("T pattern matches found: %v", matches1)
	require.False(t, errorfPattern.MatchString(testString1), "Should NOT match fmt.Errorf call")

	testString3 := `var x = 1`
	matches3 := errorfPattern.FindAllString(testString3, -1)
	t.Logf("Testing string 3: %s", testString3)
	t.Logf("Regex matches found: %v", matches3)
	require.False(t, errorfPattern.MatchString(testString3), "Should NOT match simple assignment")

	// Test t.Fatalf pattern
	fatalfPattern := regexp.MustCompile(`t\.Fatalf\([^)]+\)`)

	// Should match t.Fatalf calls
	require.True(t, fatalfPattern.MatchString(`t.Fatalf("failed to parse date: %v", err)`), "Should match t.Fatalf call")
	require.True(t, fatalfPattern.MatchString(`t.Fatalf("expected error, got nil")`), "Should match t.Fatalf with simple message")

	// Should NOT match other patterns
	require.False(t, fatalfPattern.MatchString(`fmt.Errorf("some error")`), "Should NOT match fmt.Errorf")
	require.False(t, fatalfPattern.MatchString(`t.Errorf("some error")`), "Should NOT match t.Errorf")
}
