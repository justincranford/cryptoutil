// Copyright (c) 2025 Justin Cranford
//
//

package cicd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilTestutil "cryptoutil/internal/common/testutil"
	"cryptoutil/internal/cmd/cicd/common"
)

const (
	testTypeMyStructInterface = `
type MyStruct struct {
	Data interface{}
}
`
	testStrAssignmentInterface = `
	str := "interface{} in string should not be replaced"`
)

func TestGoEnforceAny_ProcessGoFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Test case 1: File with interface{} that should be replaced
	content1 := testPackageMain + `

` + testImportFmt + `

func main() {
	var x interface{}
	fmt.Println(x)
}` + testTypeMyStructInterface + `
func process(data interface{}) interface{} {
	return data
}
`
	testFile1 := cryptoutilTestutil.WriteTempFile(t, tempDir, "test1.go", content1)

	// Process the file
	replacements1, err := processGoFile(testFile1)
	require.NoError(t, err, "processGoFile failed")

	require.Equal(t, 4, replacements1, "Expected 4 replacements")

	// Verify the content was modified correctly
	modifiedContent1 := cryptoutilTestutil.ReadTestFile(t, testFile1)

	expectedContent1 := testPackageMain + `

` + testImportFmt + `

func main() {
	var x any
	fmt.Println(x)
}` + `
type MyStruct struct {
	Data any
}
` + `
func process(data any) any {
	return data
}
`
	require.Equal(t, expectedContent1, string(modifiedContent1), "File content doesn't match expected output.\nGot:\n%s\nExpected:\n%s", string(modifiedContent1), expectedContent1)

	// Test case 2: File with no any (should not be modified)
	content2 := testPackageMain + `

` + testImportFmt + testFuncMainStart + `
	var x int
	fmt.Println(x)
` + testFuncMainEnd
	testFile2 := cryptoutilTestutil.WriteTempFile(t, tempDir, "test2.go", content2)

	// Process the file
	replacements2, err := processGoFile(testFile2)
	require.NoError(t, err, "processGoFile failed")

	require.Equal(t, 0, replacements2, "Expected 0 replacements")

	// Verify the content was not modified
	modifiedContent2 := cryptoutilTestutil.ReadTestFile(t, testFile2)

	expectedContent2 := testPackageMain + `

` + testImportFmt + testFuncMainStart + `
	var x int
	fmt.Println(x)
` + testFuncMainEnd
	require.Equal(t, expectedContent2, string(modifiedContent2), "File content was not modified as expected.\nGot:\n%s\nExpected:\n%s", string(modifiedContent2), expectedContent2)

	// Test case 3: File with interface{} in comments and strings (currently replaced - limitation of simple regex)
	content3 := testPackageMain + `
// This is a comment with interface{} that should not be replaced` + testFuncMainStart + `
	var x interface{}` + testStrAssignmentInterface + `
	fmt.Println(x, str)
` + testFuncMainEnd
	testFile3 := cryptoutilTestutil.WriteTempFile(t, tempDir, "test3.go", content3)

	// Process the file
	replacements3, err := processGoFile(testFile3)
	require.NoError(t, err, "processGoFile failed")

	require.Equal(t, 3, replacements3, "Expected 3 replacements (in comment, string, and code)")

	// Verify the content was modified (currently replaces everywhere due to simple regex)
	modifiedContent3 := cryptoutilTestutil.ReadTestFile(t, testFile3)

	expectedContent3 := `package main
// This is a comment with any that should not be replaced
func main() {
	var x any
	str := "any in string should not be replaced"
	fmt.Println(x, str)
}
`
	require.Equal(t, expectedContent3, string(modifiedContent3), "File content doesn't match expected output.\nGot:\n%s\nExpected:\n%s", string(modifiedContent3), expectedContent3)
}

func TestGoEnforceAny_RunGoEnforceAny(t *testing.T) {
	tempDir := t.TempDir()

	// Create test Go files with interface{}
	content1 := testPackageMain + `
func main() {
	var x interface{}
}
`
	testFile1 := cryptoutilTestutil.WriteTempFile(t, tempDir, "test1.go", content1)

	content2 := testPackageMain + testTypeMyStructInterface + `
`
	testFile2 := cryptoutilTestutil.WriteTempFile(t, tempDir, "test2.go", content2)

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()

	require.NoError(t, os.Chdir(tempDir))

	// Test that goEnforceAny returns an error when files are modified
	logger := common.NewLogger("test")
	err = goEnforceAny(logger, []string{testFile1, testFile2})
	require.Error(t, err, "Should return error when files are modified")
	require.Contains(t, err.Error(), "modified", "Error should indicate files were modified")

	// Verify files were actually modified
	modifiedContent1 := cryptoutilTestutil.ReadTestFile(t, testFile1)
	require.Contains(t, string(modifiedContent1), "var x any", "File 1 was not modified correctly")

	modifiedContent2 := cryptoutilTestutil.ReadTestFile(t, testFile2)
	require.Contains(t, string(modifiedContent2), "Data any", "File 2 was not modified correctly")
}
