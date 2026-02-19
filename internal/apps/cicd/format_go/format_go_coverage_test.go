// Copyright (c) 2025 Justin Cranford

package format_go

import (
"os"
"path/filepath"
"testing"

"github.com/stretchr/testify/require"

cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

// testGoContentWithInterfaceBraces is Go content with interface{} in a temp file
// that will NOT match the format-go self-exclusion pattern (not in format_go dir).
// CRITICAL: Uses interface{} in string literal intentionally - this is test data, not production code.
const testGoContentWithInterfaceBraces = "package server\n\nfunc Handle(data interface{}) interface{} {\n\treturn data\n}\n"

// TestEnforceAny_WithModificationsViaEnforceAny calls enforceAny directly with a file
// containing interface{} to cover the filesModified > 0 code path.
func TestEnforceAny_WithModificationsViaEnforceAny(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
testFile := filepath.Join(tmpDir, "server.go")

err := os.WriteFile(testFile, []byte(testGoContentWithInterfaceBraces), 0o600)
require.NoError(t, err)

logger := cryptoutilCmdCicdCommon.NewLogger("test")
filesByExtension := map[string][]string{
"go": {testFile},
}

err = enforceAny(logger, filesByExtension)
require.Error(t, err, "enforceAny should return error when files were modified")
require.Contains(t, err.Error(), "modified")
}

// TestFormat_WithEnforceAnyModification calls Format with a file containing interface{}
// to cover the simple formatter error path and the len(errors) > 0 final return.
func TestFormat_WithEnforceAnyModification(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
testFile := filepath.Join(tmpDir, "server.go")

err := os.WriteFile(testFile, []byte(testGoContentWithInterfaceBraces), 0o600)
require.NoError(t, err)

logger := cryptoutilCmdCicdCommon.NewLogger("test")
filesByExtension := map[string][]string{
"go": {testFile},
}

err = Format(logger, filesByExtension)
require.Error(t, err, "Format should return error when formatters made modifications")
require.Contains(t, err.Error(), "completed with modifications")
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

result := isGoVersionSupported(tc.version)
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
changed, fixes, err := fixCopyLoopVarInFile(logger, testFile)

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
changed, fixes, err := fixCopyLoopVarInFile(logger, testFile)

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
changed, fixes, err := fixCopyLoopVarInFile(logger, testFile)

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
changed, fixes, err := fixCopyLoopVarInFile(logger, testFile)

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
changed, fixes, err := fixCopyLoopVarInFile(logger, testFile)

require.NoError(t, err)
require.False(t, changed, "File should not be changed (RHS is index expression, not Ident)")
require.Equal(t, 0, fixes, "Should have no fixes when RHS is not Ident")
}

// TestProcessGoFile_WriteError tests that processGoFile returns an error when
// the file has interface{} but cannot be written (read-only file).
func TestProcessGoFile_WriteError(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
testFile := filepath.Join(tmpDir, "readonly.go")

// Create file with interface{} that needs replacement.
err := os.WriteFile(testFile, []byte(testGoContentWithInterfaceBraces), 0o600)
require.NoError(t, err)

// Make file read-only so WriteFile will fail.
err = os.Chmod(testFile, 0o400)
require.NoError(t, err)

replacements, err := processGoFile(testFile)
require.Error(t, err, "processGoFile should fail when file is read-only")
require.Equal(t, 0, replacements, "Should return 0 replacements on error")
require.Contains(t, err.Error(), "failed to write file")

// Restore write permission for cleanup.
_ = os.Chmod(testFile, 0o600)
}

// TestProcessGoFileForTimeNowUTC_ReadError tests the read file error path.
func TestProcessGoFileForTimeNowUTC_ReadError(t *testing.T) {
t.Parallel()

replacements, err := processGoFileForTimeNowUTC("/nonexistent/path/to/test.go")
require.Error(t, err, "processGoFileForTimeNowUTC should fail for nonexistent file")
require.Equal(t, 0, replacements)
require.Contains(t, err.Error(), "failed to read file")
}

// testGoContentUTCOnVariable is Go content with t.UTC() where t is a time.Time variable.
// This triggers the first-pass innerCall type assertion failure (Ident receiver, not CallExpr).
const testGoContentUTCOnVariable = `package main

import "time"

func main() {
var t time.Time
_ = t.UTC()
}
`

// TestProcessGoFileForTimeNowUTC_UTCOnVariable tests the first-pass type assertion failure
// for x.UTC() where x is an identifier (not a call expression).
func TestProcessGoFileForTimeNowUTC_UTCOnVariable(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
testFile := filepath.Join(tmpDir, "test.go")

err := os.WriteFile(testFile, []byte(testGoContentUTCOnVariable), 0o600)
require.NoError(t, err)

replacements, err := processGoFileForTimeNowUTC(testFile)
require.NoError(t, err, "Should not error on file with t.UTC() on variable")
require.Equal(t, 0, replacements, "No replacements needed (already correct or not time.Now())")
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
changed, fixes, err := fixCopyLoopVarInFile(logger, testFile)

// Restore permissions for cleanup.
_ = os.Chmod(testFile, 0o600)

require.Error(t, err, "fixCopyLoopVarInFile should fail when file is read-only")
require.False(t, changed, "Should return false when error occurs")
require.Greater(t, fixes, 0, "Should have detected fixes before write error")
require.Contains(t, err.Error(), "failed to create file")
}
