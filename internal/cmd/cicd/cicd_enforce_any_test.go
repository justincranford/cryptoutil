package cicd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
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
	testFile1 := writeTempFile(t, tempDir, "test1.go", content1)

	// Process the file
	replacements1, err := processGoFile(testFile1)
	require.NoError(t, err, "processGoFile failed")

	require.Equal(t, 4, replacements1, "Expected 4 replacements")

	// Verify the content was modified correctly
	modifiedContent1, err := os.ReadFile(testFile1)
	require.NoError(t, err, "Failed to read modified file")

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
	testFile2 := writeTempFile(t, tempDir, "test2.go", content2)

	// Process the file
	replacements2, err := processGoFile(testFile2)
	require.NoError(t, err, "processGoFile failed")

	require.Equal(t, 0, replacements2, "Expected 0 replacements")

	// Verify the content was not modified
	modifiedContent2, err := os.ReadFile(testFile2)
	require.NoError(t, err, "Failed to read modified file")

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
	testFile3 := writeTempFile(t, tempDir, "test3.go", content3)

	// Process the file
	replacements3, err := processGoFile(testFile3)
	require.NoError(t, err, "processGoFile failed")

	require.Equal(t, 3, replacements3, "Expected 3 replacements (in comment, string, and code)")

	// Verify the content was modified (currently replaces everywhere due to simple regex)
	modifiedContent3, err := os.ReadFile(testFile3)
	require.NoError(t, err, "Failed to read modified file")

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
	// Note: This test cannot easily test runGoEnforceAny() directly because it calls os.Exit(1)
	// when files are modified. Instead, we test the core logic by simulating what it does.
	tempDir := t.TempDir()

	// Create test Go files with interface{}
	content1 := testPackageMain + `
func main() {
	var x interface{}
}
`
	testFile1 := writeTempFile(t, tempDir, "test1.go", content1)

	content2 := testPackageMain + testTypeMyStructInterface + `
`
	testFile2 := writeTempFile(t, tempDir, "test2.go", content2)

	// Simulate the file discovery logic from runGoEnforceAny
	var goFiles []string

	err := filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			goFiles = append(goFiles, path)
		}

		return nil
	})
	require.NoError(t, err, "Failed to walk temp dir")

	require.Len(t, goFiles, 2, "Expected 2 Go files")

	// Process each file
	filesModified := 0
	totalReplacements := 0

	for _, filePath := range goFiles {
		replacements, err := processGoFile(filePath)
		require.NoError(t, err, "Error processing %s", filePath)

		if replacements > 0 {
			filesModified++
			totalReplacements += replacements
		}
	}

	require.Equal(t, 2, filesModified, "Expected 2 files modified")

	require.Equal(t, 2, totalReplacements, "Expected 2 total replacements")

	// Verify files were actually modified
	modifiedContent1, err := os.ReadFile(testFile1)
	require.NoError(t, err, "Failed to read modified file")

	require.Contains(t, string(modifiedContent1), "var x any", "File 1 was not modified correctly. Content: %s", string(modifiedContent1))

	modifiedContent2, err := os.ReadFile(testFile2)
	require.NoError(t, err, "Failed to read modified file")

	require.Contains(t, string(modifiedContent2), "Data any", "File 2 was not modified correctly. Content: %s", string(modifiedContent2))
}
