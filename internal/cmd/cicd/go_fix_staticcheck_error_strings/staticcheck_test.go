package go_fix_staticcheck_error_strings

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmd "cryptoutil/internal/cmd/cicd/common"
)

func TestFix_EmptyDirectory(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-fix-empty")

	processed, modified, issuesFixed, err := Fix(logger, tmpDir)
	require.NoError(t, err)
	require.Equal(t, 0, processed)
	require.Equal(t, 0, modified)
	require.Equal(t, 0, issuesFixed)
}

func TestFix_NoGoFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-fix-no-go")

	// Create non-Go files.
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Test"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "config.yaml"), []byte("key: value"), 0o600))

	processed, modified, issuesFixed, err := Fix(logger, tmpDir)
	require.NoError(t, err)
	require.Equal(t, 0, processed)
	require.Equal(t, 0, modified)
	require.Equal(t, 0, issuesFixed)
}

func TestFix_NoErrorStrings(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-fix-no-errors")

	// Create Go file without error strings.
	goFile := filepath.Join(tmpDir, "clean.go")
	content := `package test

func Add(a, b int) int {
	return a + b
}
`
	require.NoError(t, os.WriteFile(goFile, []byte(content), 0o600))

	processed, modified, issuesFixed, err := Fix(logger, tmpDir)
	require.NoError(t, err)
	require.Equal(t, 1, processed)
	require.Equal(t, 0, modified)
	require.Equal(t, 0, issuesFixed)
}

func TestFix_ErrorStringWithUppercase(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-fix-uppercase")

	goFile := filepath.Join(tmpDir, "errors.go")
	content := `package test

import "errors"

var ErrInvalid = errors.New("Invalid input provided")
`
	require.NoError(t, os.WriteFile(goFile, []byte(content), 0o600))

	processed, modified, issuesFixed, err := Fix(logger, tmpDir)
	require.NoError(t, err)
	require.Equal(t, 1, processed)
	require.Equal(t, 1, modified)
	require.Equal(t, 1, issuesFixed)

	// Verify the fix.
	fixed, err := os.ReadFile(goFile)
	require.NoError(t, err)
	require.Contains(t, string(fixed), `errors.New("invalid input provided")`)
	require.NotContains(t, string(fixed), `errors.New("Invalid input provided")`)
}

func TestFix_ErrorStringWithAcronym(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-fix-acronym")

	goFile := filepath.Join(tmpDir, "http_errors.go")
	content := `package test

import "errors"

var ErrHTTPFailed = errors.New("HTTP request failed")
var ErrJSONInvalid = errors.New("JSON parsing error")
var ErrURLMalformed = errors.New("URL is malformed")
`
	require.NoError(t, os.WriteFile(goFile, []byte(content), 0o600))

	processed, modified, issuesFixed, err := Fix(logger, tmpDir)
	require.NoError(t, err)
	require.Equal(t, 1, processed)
	require.Equal(t, 0, modified) // No changes because acronyms are allowed.
	require.Equal(t, 0, issuesFixed)

	// Verify no changes.
	fixed, err := os.ReadFile(goFile)
	require.NoError(t, err)
	require.Equal(t, content, string(fixed))
}

func TestFix_MultipleErrorStrings(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-fix-multiple")

	goFile := filepath.Join(tmpDir, "multi_errors.go")
	content := `package test

import "errors"

var ErrOne = errors.New("First error occurred")
var ErrTwo = errors.New("Second error occurred")
var ErrThree = errors.New("Third error occurred")
`
	require.NoError(t, os.WriteFile(goFile, []byte(content), 0o600))

	processed, modified, issuesFixed, err := Fix(logger, tmpDir)
	require.NoError(t, err)
	require.Equal(t, 1, processed)
	require.Equal(t, 1, modified)
	require.Equal(t, 3, issuesFixed)

	// Verify all fixes.
	fixed, err := os.ReadFile(goFile)
	require.NoError(t, err)
	require.Contains(t, string(fixed), `errors.New("first error occurred")`)
	require.Contains(t, string(fixed), `errors.New("second error occurred")`)
	require.Contains(t, string(fixed), `errors.New("third error occurred")`)
}

func TestFix_MixedAcronymsAndUppercase(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-fix-mixed")

	goFile := filepath.Join(tmpDir, "mixed.go")
	content := `package test

import "errors"

var ErrHTTP = errors.New("HTTP connection failed")
var ErrGeneric = errors.New("Generic error occurred")
var ErrJSON = errors.New("JSON decode error")
var ErrBad = errors.New("Bad request received")
`
	require.NoError(t, os.WriteFile(goFile, []byte(content), 0o600))

	processed, modified, issuesFixed, err := Fix(logger, tmpDir)
	require.NoError(t, err)
	require.Equal(t, 1, processed)
	require.Equal(t, 1, modified)
	require.Equal(t, 2, issuesFixed) // Only "Generic" and "Bad" should be fixed.

	// Verify selective fixes.
	fixed, err := os.ReadFile(goFile)
	require.NoError(t, err)
	require.Contains(t, string(fixed), `errors.New("HTTP connection failed")`) // Unchanged.
	require.Contains(t, string(fixed), `errors.New("generic error occurred")`) // Fixed.
	require.Contains(t, string(fixed), `errors.New("JSON decode error")`)      // Unchanged.
	require.Contains(t, string(fixed), `errors.New("bad request received")`)   // Fixed.
}

func TestFix_FmtErrorf(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-fix-fmt")

	goFile := filepath.Join(tmpDir, "fmt_errors.go")
	content := `package test

import "fmt"

var ErrFmt = fmt.Errorf("Failed to process request")
`
	require.NoError(t, os.WriteFile(goFile, []byte(content), 0o600))

	processed, modified, issuesFixed, err := Fix(logger, tmpDir)
	require.NoError(t, err)
	require.Equal(t, 1, processed)
	require.Equal(t, 1, modified)
	require.Equal(t, 1, issuesFixed)

	// Verify the fix.
	fixed, err := os.ReadFile(goFile)
	require.NoError(t, err)
	require.Contains(t, string(fixed), `fmt.Errorf("failed to process request")`)
}

func TestFix_TestFilesSkipped(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-fix-skip-test")

	// Create test file with uppercase error.
	testFile := filepath.Join(tmpDir, "errors_test.go")
	content := `package test

import "errors"

var ErrTest = errors.New("Test error occurred")
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0o600))

	processed, modified, issuesFixed, err := Fix(logger, tmpDir)
	require.NoError(t, err)
	require.Equal(t, 0, processed) // Test files should be skipped.
	require.Equal(t, 0, modified)
	require.Equal(t, 0, issuesFixed)
}

func TestFix_NestedDirectories(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-fix-nested")

	// Create nested directory structure.
	subDir := filepath.Join(tmpDir, "sub", "nested")
	require.NoError(t, os.MkdirAll(subDir, 0o755))

	// Create files in different levels.
	file1 := filepath.Join(tmpDir, "root.go")
	file2 := filepath.Join(tmpDir, "sub", "mid.go")
	file3 := filepath.Join(subDir, "deep.go")

	content := `package test
import "errors"
var Err = errors.New("Error occurred")
`
	require.NoError(t, os.WriteFile(file1, []byte(content), 0o600))
	require.NoError(t, os.WriteFile(file2, []byte(content), 0o600))
	require.NoError(t, os.WriteFile(file3, []byte(content), 0o600))

	processed, modified, issuesFixed, err := Fix(logger, tmpDir)
	require.NoError(t, err)
	require.Equal(t, 3, processed)
	require.Equal(t, 3, modified)
	require.Equal(t, 3, issuesFixed)
}

func TestFix_InvalidDirectory(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmd.NewLogger("test-fix-invalid")

	processed, modified, issuesFixed, err := Fix(logger, "/nonexistent/path/to/nowhere")
	require.Error(t, err)
	require.Equal(t, 0, processed)
	require.Equal(t, 0, modified)
	require.Equal(t, 0, issuesFixed)
}

func TestFix_ConstErrorStrings(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-fix-const")

	goFile := filepath.Join(tmpDir, "const_errors.go")
	content := `package test

import "errors"

const (
	errMsg = "Error message one"
)

var ErrConst = errors.New("Constant error occurred")
`
	require.NoError(t, os.WriteFile(goFile, []byte(content), 0o600))

	processed, modified, issuesFixed, err := Fix(logger, tmpDir)
	require.NoError(t, err)
	require.Equal(t, 1, processed)
	require.Equal(t, 1, modified)
	require.Equal(t, 1, issuesFixed)

	// Verify the fix (only the errors.New() call should be fixed).
	fixed, err := os.ReadFile(goFile)
	require.NoError(t, err)
	require.Contains(t, string(fixed), `errors.New("constant error occurred")`)
}
