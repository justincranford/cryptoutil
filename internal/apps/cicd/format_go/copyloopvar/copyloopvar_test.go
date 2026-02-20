// Copyright (c) 2025 Justin Cranford

package copyloopvar

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

const testGoContentInvalid = "package main\n\nfunc main() {\n\tthis is not valid go code\n}\n"

func TestIsGoVersionSupported(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		version   string
		supported bool
	}{
		{"go1.22", "go1.22", true},
		{"go1.22.0", "go1.22.0", true},
		{"go1.25.5", "go1.25.5", true},
		{"go1.21", "go1.21", false},
		{"go1.21.5", "go1.21.5", false},
		{"go1.20", "go1.20", false},
		{"invalid", "invalid", false},
		{"empty", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := IsGoVersionSupported(tc.version)
			require.Equal(t, tc.supported, result)
		})
	}
}

func TestIsLoopVarCopy(t *testing.T) {
	t.Parallel()

	// This is a unit test for the isLoopVarCopy function.
	// We test the function logic indirectly through fixCopyLoopVarInFile.
	// Direct AST testing would be more complex.
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	// File with no loop var copy.
	content := `package main

func main() {
	items := []int{1, 2, 3}
	for _, v := range items {
		println(v)
	}
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	changed, fixes, err := FixInFile(logger, testFile)

	require.NoError(t, err)
	require.False(t, changed, "File should not be changed")
	require.Equal(t, 0, fixes, "Should have no fixes")
}

const testGoContentWithLoopVarCopy = `package main

func main() {
	items := []int{1, 2, 3}
	for _, v := range items {
		v := v
		println(v)
	}
}
`

func TestFixCopyLoopVarInFile_WithLoopVarCopy(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	// File with loop var copy pattern.
	err := os.WriteFile(testFile, []byte(testGoContentWithLoopVarCopy), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	changed, fixes, err := FixInFile(logger, testFile)

	require.NoError(t, err)
	require.True(t, changed, "File should be changed")
	require.Equal(t, 1, fixes, "Should have 1 fix")

	// Verify the file was modified.
	modifiedContent, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.NotContains(t, string(modifiedContent), "v := v", "File should not contain loop var copy")
}

func TestFixCopyLoopVarInFile_ParseError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "invalid.go")

	// Invalid Go syntax.
	err := os.WriteFile(testFile, []byte(testGoContentInvalid), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	changed, fixes, err := FixInFile(logger, testFile)

	require.Error(t, err)
	require.False(t, changed)
	require.Equal(t, 0, fixes)
	require.Contains(t, err.Error(), "failed to parse file")
}

func TestFixCopyLoopVarInFile_EmptyBody(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	// Loop with empty body.
	content := `package main

func main() {
	items := []int{1, 2, 3}
	for _, v := range items {
	}
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	changed, fixes, err := FixInFile(logger, testFile)

	require.NoError(t, err)
	require.False(t, changed)
	require.Equal(t, 0, fixes)
}

func TestFixCopyLoopVar_UnsupportedVersion(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	processed, modified, issuesFixed, err := Fix(logger, tmpDir, "go1.21")

	require.NoError(t, err)
	require.Equal(t, 0, processed, "Should process 0 files")
	require.Equal(t, 0, modified, "Should modify 0 files")
	require.Equal(t, 0, issuesFixed, "Should fix 0 issues")
}

func TestFixCopyLoopVar_SupportedVersion(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create a Go file with loop var copy.
	testFile := filepath.Join(tmpDir, "test.go")
	err := os.WriteFile(testFile, []byte(testGoContentWithLoopVarCopy), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	processed, modified, issuesFixed, err := Fix(logger, tmpDir, "go1.22")

	require.NoError(t, err)
	require.Equal(t, 1, processed, "Should process 1 file")
	require.Equal(t, 1, modified, "Should modify 1 file")
	require.Equal(t, 1, issuesFixed, "Should fix 1 issue")
}

func TestFixCopyLoopVar_EmptyDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	processed, modified, issuesFixed, err := Fix(logger, tmpDir, "go1.22")

	require.NoError(t, err)
	require.Equal(t, 0, processed, "Should process 0 files")
	require.Equal(t, 0, modified, "Should modify 0 files")
	require.Equal(t, 0, issuesFixed, "Should fix 0 issues")
}

// TestIsGoVersionSupported_AtoisErrors tests strconv.Atoi error paths in isGoVersionSupported.
func TestIsGoVersionSupported_AtoisErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		version   string
		supported bool
	}{
		{"major_atoi_error", "abc.1", false},
		{"minor_atoi_error", "1.abc", false},
		{"too_few_parts", "1", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := IsGoVersionSupported(tc.version)
			require.Equal(t, tc.supported, result)
		})
	}
}

// testGoContentKeyCopy is a Go file with a range-over-map loop using key copy.
const testGoContentKeyCopy = `package main

func main() {
m := map[string]int{"a": 1, "b": 2}
for k := range m {
k := k
_ = k
}
}
`

// testGoContentLhsNotRhs is a Go file with a loop where lhs := rhs but lhs != rhs name.
const testGoContentLhsNotRhs = `package main

func main() {
items := []int{1, 2, 3}
for _, v := range items {
x := v
println(x)
}
}
`

// testGoContentOuterVarShadow is a Go file where x := x in range but x is not a range variable.
const testGoContentOuterVarShadow = `package main

func main() {
x := 42
items := []int{1, 2, 3}
for _, v := range items {
x := x
println(x, v)
}
}
`

// TestFixCopyLoopVarInFile_KeyCopy tests the key-copy path in isLoopVarCopy.
func TestFixCopyLoopVarInFile_KeyCopy(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	err := os.WriteFile(testFile, []byte(testGoContentKeyCopy), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	changed, fixes, err := FixInFile(logger, testFile)

	require.NoError(t, err)
	require.True(t, changed, "File should be changed (key copy should be removed)")
	require.Equal(t, 1, fixes, "Should have 1 fix for key copy")
}

// TestFixCopyLoopVarInFile_LhsNotRhs tests isLoopVarCopy returning false when lhs != rhs.
func TestFixCopyLoopVarInFile_LhsNotRhs(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	err := os.WriteFile(testFile, []byte(testGoContentLhsNotRhs), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	changed, fixes, err := FixInFile(logger, testFile)

	require.NoError(t, err)
	require.False(t, changed, "File should not be changed (lhs != rhs is not a copy)")
	require.Equal(t, 0, fixes, "Should have no fixes")
}

// TestFixCopyLoopVarInFile_OuterVarShadow tests isLoopVarCopy returning false
// when x := x is inside a range but x is not a range variable (outer scope shadow).
func TestFixCopyLoopVarInFile_OuterVarShadow(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	err := os.WriteFile(testFile, []byte(testGoContentOuterVarShadow), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	changed, fixes, err := FixInFile(logger, testFile)

	require.NoError(t, err)
	require.False(t, changed, "File should not be changed (x is not a range variable)")
	require.Equal(t, 0, fixes, "Should have no fixes")
}

// testGoContentSelfAssign has v = v (assignment not :=) in a range loop.
const testGoContentSelfAssign = `package main

func main() {
items := []int{1, 2, 3}
for _, v := range items {
v = v
println(v)
}
}
`

// testGoContentRhsIsIndex has v := items[0] inside a range (RHS is not an Ident).
const testGoContentRhsIsIndex = `package main

func main() {
items := []int{1, 2, 3}
for _, v := range items {
v := items[0]
println(v)
}
}
`

// TestFixCopyLoopVarInFile_SelfAssignment tests isLoopVarCopy returning false
// when the statement uses = (ASSIGN) not := (DEFINE).
func TestFixCopyLoopVarInFile_SelfAssignment(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	err := os.WriteFile(testFile, []byte(testGoContentSelfAssign), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	changed, fixes, err := FixInFile(logger, testFile)

	require.NoError(t, err)
	require.False(t, changed, "File should not be changed (v = v uses ASSIGN not DEFINE)")
	require.Equal(t, 0, fixes, "Should have no fixes for plain assignment")
}

// TestFixCopyLoopVarInFile_RhsIsNotIdent tests isLoopVarCopy returning false
// when RHS is not an identifier (e.g., index expression).
func TestFixCopyLoopVarInFile_RhsIsNotIdent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	err := os.WriteFile(testFile, []byte(testGoContentRhsIsIndex), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	changed, fixes, err := FixInFile(logger, testFile)

	require.NoError(t, err)
	require.False(t, changed, "File should not be changed (RHS is index expression, not Ident)")
	require.Equal(t, 0, fixes, "Should have no fixes when RHS is not Ident")
}

// TestFixCopyLoopVarInFile_ReadOnlyFile tests that fixCopyLoopVarInFile
// returns an error when the file has loop var copies but cannot be written.
func TestFixCopyLoopVarInFile_ReadOnlyFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	// Create file with loop var copy content.
	err := os.WriteFile(testFile, []byte(testGoContentWithLoopVarCopy), 0o600)
	require.NoError(t, err)

	// Make file read-only so os.Create will fail during write.
	err = os.Chmod(testFile, 0o400)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	changed, fixes, err := FixInFile(logger, testFile)

	// Restore permissions for cleanup.
	_ = os.Chmod(testFile, 0o600)

	require.Error(t, err, "fixCopyLoopVarInFile should fail when file is read-only")
	require.False(t, changed, "Should return false when error occurs")
	require.Greater(t, fixes, 0, "Should have detected fixes before write error")
	require.Contains(t, err.Error(), "failed to create file")
}

func TestFix_WalkDirError(t *testing.T) {
	// Non-parallel: modifies directory permissions.
	tmpDir := t.TempDir()
	// Create an unreadable subdirectory - Walk will call callback with OS error.
	badSubDir := filepath.Join(tmpDir, "locked")
	require.NoError(t, os.MkdirAll(badSubDir, 0o700))
	require.NoError(t, os.Chmod(badSubDir, 0o000))
	t.Cleanup(func() { _ = os.Chmod(badSubDir, 0o700) })

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	_, _, _, err := Fix(logger, tmpDir, "1.22")
	require.Error(t, err, "Should fail when subdirectory is unreadable")
	require.Contains(t, err.Error(), "failed to walk directory")
}

func TestFix_WithSyntaxErrorFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	// Create a .go file with syntax errors - FixInFile will fail to parse it.
	goFile := filepath.Join(tmpDir, "bad_syntax.go")
	require.NoError(t, os.WriteFile(goFile, []byte(testGoContentInvalid), 0o600))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	_, _, _, err := Fix(logger, tmpDir, "1.22")
	require.Error(t, err, "Should fail when a .go file has syntax errors")
	require.Contains(t, err.Error(), "failed to process")
}

func TestIsLoopVarCopy_MultipleAssignTargets(t *testing.T) {
	t.Parallel()

	// Parse a range where the first statement is a multi-assignment (a, b := v, v).
	src := `package foo
func f() {
	for _, v := range []int{1} {
		a, b := v, v
		_ = a
		_ = b
	}
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, 0)
	require.NoError(t, err)

	var (
		rangeStmt *ast.RangeStmt
		assign    *ast.AssignStmt
	)

	ast.Inspect(file, func(n ast.Node) bool {
		if rs, ok := n.(*ast.RangeStmt); ok {
			rangeStmt = rs
			if len(rs.Body.List) > 0 {
				if as, ok2 := rs.Body.List[0].(*ast.AssignStmt); ok2 {
					assign = as
				}
			}
		}

		return true
	})

	require.NotNil(t, rangeStmt)
	require.NotNil(t, assign)
	// Multi-assign (a, b := v, v) should not be identified as a loop var copy.
	require.False(t, IsLoopVarCopy(rangeStmt, assign), "Multi-assign should return false")
}
