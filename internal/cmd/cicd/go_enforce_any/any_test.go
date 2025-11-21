// Copyright (c) 2025 Justin Cranford

package go_enforce_any

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"cryptoutil/internal/cmd/cicd/common"
	cryptoutilTestutil "cryptoutil/internal/common/testutil"
)

const (
	testPackageMain           = "package main"
	testImportFmt             = `import "fmt"`
	testFuncMainStart         = "\nfunc main() {"
	testFuncMainEnd           = "\n}\n"
	testTypeMyStructInterface = `
type MyStruct struct {
	Data interface{}
}
`
	testStrAssignmentInterface = `
	str := "interface{} in string should not be replaced"`
)

func TestProcessGoFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		content          string
		wantReplacements int
		wantError        bool
		wantContains     string
	}{
		{
			name: "file with interface{} that should be replaced",
			content: testPackageMain + `

` + testImportFmt + `

func main() {
	var x interface{}
	fmt.Println(x)
}` + testTypeMyStructInterface + `
func process(data interface{}) interface{} {
	return data
}
`,
			wantReplacements: 4,
			wantError:        false,
			wantContains: testPackageMain + `

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
`,
		},
		{
			name: "file with no interface{} unchanged",
			content: testPackageMain + `

` + testImportFmt + testFuncMainStart + `
	var x int
	fmt.Println(x)
` + testFuncMainEnd,
			wantReplacements: 0,
			wantError:        false,
			wantContains: testPackageMain + `

` + testImportFmt + testFuncMainStart + `
	var x int
	fmt.Println(x)
` + testFuncMainEnd,
		},
		{
			name: "file with interface{} in comments and strings (regex limitation)",
			content: testPackageMain + `
// This is a comment with interface{} that will be replaced (regex limitation)` + testFuncMainStart + `
	var x interface{}` + testStrAssignmentInterface + `
	fmt.Println(x, str)
` + testFuncMainEnd,
			wantReplacements: 3,
			wantError:        false,
			wantContains: `package main
// This is a comment with any that will be replaced (regex limitation)
func main() {
	var x any
	str := "any in string should not be replaced"
	fmt.Println(x, str)

}
`,
		},
		{
			name:             "empty file",
			content:          "",
			wantReplacements: 0,
			wantError:        false,
			wantContains:     "",
		},
		{
			name:             "only comments",
			content:          "// This is a comment\n/* Block comment */\n",
			wantReplacements: 0,
			wantError:        false,
			wantContains:     "// This is a comment\n/* Block comment */\n",
		},
		{
			name: "already using any",
			content: `package test
var x any = 42
`,
			wantReplacements: 0,
			wantError:        false,
			wantContains: `package test
var x any = 42
`,
		},
		{
			name: "multiple any on same line",
			content: `package test
func convert(a any, b any) (any, any) {
	return a, b
}
`,
			wantReplacements: 0,
			wantError:        false,
			wantContains: `package test
func convert(a any, b any) (any, any) {
	return a, b
}
`,
		},
	}

	for _, tc := range tests {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			testFile := cryptoutilTestutil.WriteTempFile(t, tempDir, "test.go", tc.content)

			replacements, err := processGoFile(testFile)

			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err, "processGoFile failed")
				require.Equal(t, tc.wantReplacements, replacements, "Expected %d replacements", tc.wantReplacements)

				// Verify the content matches expected
				modifiedContent := cryptoutilTestutil.ReadTestFile(t, testFile)
				require.Equal(t, tc.wantContains, string(modifiedContent), "File content doesn't match expected output")
			}
		})
	}
}

func TestEnforce(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		files        map[string]string
		wantError    bool
		wantAnyStr   bool // expect "any" in at least one file
	}{
		{
			name: "files with interface{} should return error after modification",
			files: map[string]string{
				"test1.go": testPackageMain + `
func main() {
	var x interface{}
}
`,
				"test2.go": testPackageMain + testTypeMyStructInterface,
			},
			wantError:  true,
			wantAnyStr: true,
		},
		{
			name: "files without interface{} should not return error",
			files: map[string]string{
				"clean.go": testPackageMain + `
func main() {
	var x any
}
`,
			},
			wantError:  false,
			wantAnyStr: true,
		},
	}

	for _, tc := range tests {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			var filePaths []string

			// Create test files
			for filename, content := range tc.files {
				filePath := cryptoutilTestutil.WriteTempFile(t, tempDir, filename, content)
				filePaths = append(filePaths, filePath)
			}

			// Change to temp directory
			oldWd, err := os.Getwd()
			require.NoError(t, err)

			defer func() {
				_ = os.Chdir(oldWd) // Best effort cleanup
			}()

			require.NoError(t, os.Chdir(tempDir))

			// Test Enforce
			logger := common.NewLogger("test")
			err = Enforce(logger, filePaths)

			if tc.wantError {
				require.Error(t, err, "Should return error when files are modified")
				require.Contains(t, err.Error(), "modified", "Error should indicate files were modified")
			} else {
				require.NoError(t, err, "Should not return error for clean files")
			}

			// Verify expected content in modified files
			if tc.wantAnyStr {
				foundAny := false
				for _, filePath := range filePaths {
					modifiedContent := cryptoutilTestutil.ReadTestFile(t, filePath)
					if strings.Contains(string(modifiedContent), "any") {
						foundAny = true

						break
					}
				}
				require.True(t, foundAny, "At least one file should contain 'any'")
			}
		})
	}
}