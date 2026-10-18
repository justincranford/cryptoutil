// Copyright (c) 2025-2026 Justin Cranford.
package failures

import (
	"errors"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	testErrorfFailureStatement  = `t.Errorf("failure")`
	testFatalfFailureStatement  = `t.Fatalf("failure")`
	testFailedToReadFileMessage = "failed to read file"
	testParseAfterImportsCase   = "parse after imports error"
	testReadErrorCase           = "read error"
)

func TestFix_ErrorPathsAndFileFiltering(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger(t.Name())

	t.Run("list error", func(t *testing.T) {
		t.Parallel()

		_, _, _, err := Fix(logger, filepath.Join(t.TempDir(), "missing"))

		require.ErrorContains(t, err, "failed to list files")
	})

	t.Run("ignores non-test go files", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(root, "source.go"), []byte("package example\n"), cryptoutilSharedMagic.CacheFilePermissions))

		processed, modified, fixed, err := Fix(logger, root)

		require.NoError(t, err)
		require.Zero(t, processed)
		require.Zero(t, modified)
		require.Zero(t, fixed)
	})

	t.Run("reports invalid test file", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(root, "invalid_test.go"), []byte("package example\nfunc {"), cryptoutilSharedMagic.CacheFilePermissions))

		processed, modified, fixed, err := Fix(logger, root)

		require.ErrorContains(t, err, "failed to process")
		require.Zero(t, processed)
		require.Zero(t, modified)
		require.Zero(t, fixed)
	})
}

func TestFixFileWithOperations_ErrorPaths(t *testing.T) {
	t.Parallel()

	injectedError := errors.New("injected file operation error")
	tests := []struct {
		name      string
		readError error
		statError error
		writeErr  error
		wantError string
	}{
		{name: testReadErrorCase, readError: injectedError, wantError: testFailedToReadFileMessage},
		{name: "stat error", statError: injectedError, wantError: "failed to stat file"},
		{name: "write error", writeErr: injectedError, wantError: "failed to write file"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			filePath := filepath.Join(t.TempDir(), "sample_test.go")
			content := []byte("package example\n\nimport \"testing\"\n\nfunc TestExample(t *testing.T) { " + testErrorfFailureStatement + " }\n")
			require.NoError(t, os.WriteFile(filePath, content, cryptoutilSharedMagic.CacheFilePermissions))

			fileOperations := failureFileOperations{
				readFile: func(path string) ([]byte, error) {
					if tc.readError != nil {
						return nil, tc.readError
					}

					return os.ReadFile(path)
				},
				stat: func(path string) (os.FileInfo, error) {
					if tc.statError != nil {
						return nil, tc.statError
					}

					return os.Stat(path)
				},
				writeFile: func(string, []byte, os.FileMode) error {
					return tc.writeErr
				},
				ensureImport:     ensureImport,
				ensureImportAs:   ensureImportAs,
				formatSource:     format.Source,
				collectCallEdits: collectCallEdits,
			}

			_, _, err := fixFileWithOperations(filePath, fileOperations)

			require.ErrorContains(t, err, tc.wantError)
			require.ErrorIs(t, err, injectedError)
		})
	}
}

func TestFixFileWithOperations_StageErrors(t *testing.T) {
	t.Parallel()

	injectedError := errors.New("injected stage error")
	tests := []struct {
		name      string
		statement string
		configure func(*failureFileOperations)
		wantError string
	}{
		{
			name:      "assert import error",
			statement: testErrorfFailureStatement,
			configure: func(operations *failureFileOperations) {
				operations.ensureImport = func([]byte, string) ([]byte, error) {
					return nil, injectedError
				}
			},
			wantError: injectedError.Error(),
		},
		{
			name:      "require import error",
			statement: testFatalfFailureStatement,
			configure: func(operations *failureFileOperations) {
				operations.ensureImport = func([]byte, string) ([]byte, error) {
					return nil, injectedError
				}
			},
			wantError: injectedError.Error(),
		},
		{
			name:      "magic import error",
			statement: testErrorfFailureStatement,
			configure: func(operations *failureFileOperations) {
				operations.ensureImportAs = func([]byte, string, string) ([]byte, error) {
					return nil, injectedError
				}
			},
			wantError: injectedError.Error(),
		},
		{
			name:      testParseAfterImportsCase,
			statement: testErrorfFailureStatement,
			configure: func(operations *failureFileOperations) {
				operations.ensureImportAs = func([]byte, string, string) ([]byte, error) {
					return []byte("package {"), nil
				}
			},
			wantError: "failed to parse file after imports",
		},
		{
			name:      "format error",
			statement: testErrorfFailureStatement,
			configure: func(operations *failureFileOperations) {
				operations.formatSource = func([]byte) ([]byte, error) {
					return nil, injectedError
				}
			},
			wantError: injectedError.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			filePath := filepath.Join(t.TempDir(), "sample_test.go")
			source := []byte("package example\n\nimport \"testing\"\n\nfunc TestExample(t *testing.T) { " + tc.statement + " }\n")
			require.NoError(t, os.WriteFile(filePath, source, cryptoutilSharedMagic.CacheFilePermissions))

			operations := failureFileOperations{
				readFile:  os.ReadFile,
				stat:      os.Stat,
				writeFile: os.WriteFile,
				ensureImport: func(content []byte, importPath string) ([]byte, error) {
					return ensureImport(content, importPath)
				},
				ensureImportAs:   ensureImportAs,
				formatSource:     format.Source,
				collectCallEdits: collectCallEdits,
			}
			tc.configure(&operations)

			_, _, err := fixFileWithOperations(filePath, operations)

			require.ErrorContains(t, err, tc.wantError)

			if tc.name != testParseAfterImportsCase {
				require.ErrorIs(t, err, injectedError)
			}
		})
	}
}

func TestFixFileWithOperations_NoCallEdits(t *testing.T) {
	t.Parallel()

	filePath := filepath.Join(t.TempDir(), "sample_test.go")
	source := []byte("package example\n\nimport \"testing\"\n\nfunc TestExample(t *testing.T) { " + testErrorfFailureStatement + " }\n")
	require.NoError(t, os.WriteFile(filePath, source, cryptoutilSharedMagic.CacheFilePermissions))

	operations := failureFileOperations{
		readFile:         os.ReadFile,
		stat:             os.Stat,
		writeFile:        os.WriteFile,
		ensureImport:     ensureImport,
		ensureImportAs:   ensureImportAs,
		formatSource:     format.Source,
		collectCallEdits: func(*token.FileSet, *ast.File, []byte, map[string]string) []callEdit { return nil },
	}

	changed, fixes, err := fixFileWithOperations(filePath, operations)

	require.NoError(t, err)
	require.False(t, changed)
	require.Zero(t, fixes)
}

func TestFixFile_ReadAndParseErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		content   string
		wantError string
	}{
		{name: testReadErrorCase, content: "", wantError: testFailedToReadFileMessage},
		{name: "parse error", content: "package example\nfunc {", wantError: "failed to parse file"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			filePath := filepath.Join(t.TempDir(), "sample_test.go")
			if tc.content != "" {
				require.NoError(t, os.WriteFile(filePath, []byte(tc.content), cryptoutilSharedMagic.CacheFilePermissions))
			}

			_, _, err := fixFile(filePath)

			require.ErrorContains(t, err, tc.wantError)
		})
	}
}

func TestFailureCalls_RecognizesLegacyAndConvertedMethods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statement  string
		wantErrorf bool
		wantFatalf bool
	}{
		{name: "testing Errorf", statement: testErrorfFailureStatement, wantErrorf: true},
		{name: "testing Fatalf", statement: testFatalfFailureStatement, wantFatalf: true},
		{name: "assert legacy Failf", statement: `assert.Failf(t, "testing.T.Errorf failed", "failure")`, wantErrorf: true},
		{name: "require legacy FailNowf", statement: `require.FailNowf(t, "testing.T.Fatalf failed", "failure")`, wantFatalf: true},
		{name: "wrong receiver and nonmatching calls", statement: `helper(); other.Errorf("failure"); t.Errorf(); assert.Failf(other, "testing.T.Errorf failed", "failure")`},
		{name: "nonliteral failure message", statement: `assert.Failf(t, message, "failure")`},
		{name: "unknown failure message", statement: `assert.Failf(t, "other failure", "failure")`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			source := "package example\nfunc TestExample(t *testing.T) { " + tc.statement + " }\n"
			file, err := parser.ParseFile(token.NewFileSet(), "sample_test.go", source, 0)
			require.NoError(t, err)

			hasErrorf, hasFatalf := failureCalls(file)

			require.Equal(t, tc.wantErrorf, hasErrorf)
			require.Equal(t, tc.wantFatalf, hasFatalf)
		})
	}
}

func TestFailureHelpers_InvalidStringLiteral(t *testing.T) {
	t.Parallel()

	invalidString := &ast.BasicLit{Kind: token.STRING, Value: `"invalid\q"`}
	call := &ast.CallExpr{
		Fun: &ast.SelectorExpr{X: ast.NewIdent("assert"), Sel: ast.NewIdent(cryptoutilSharedMagic.CICDTestMethodFailf)},
		Args: []ast.Expr{
			ast.NewIdent("t"),
			invalidString,
			&ast.BasicLit{Kind: token.STRING, Value: `"failure"`},
		},
	}
	file := &ast.File{
		Name: ast.NewIdent("example"),
		Decls: []ast.Decl{&ast.FuncDecl{
			Name: ast.NewIdent("TestExample"),
			Type: &ast.FuncType{Params: &ast.FieldList{}},
			Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ExprStmt{X: call}}},
		}},
	}

	hasErrorf, hasFatalf := failureCalls(file)
	require.False(t, hasErrorf)
	require.False(t, hasFatalf)
	require.Empty(t, collectCallEdits(token.NewFileSet(), file, nil, map[string]string{}))
}

func TestPackageAliases_InvalidImportPath(t *testing.T) {
	t.Parallel()

	file := &ast.File{Imports: []*ast.ImportSpec{{
		Path: &ast.BasicLit{Kind: token.STRING, Value: `"`},
	}}}

	require.Empty(t, packageAliases(file))
}

func TestCollectCallEdits_EllipsisAliasesAndIgnoredCalls(t *testing.T) {
	t.Parallel()

	fileSet := token.NewFileSet()
	source := `package example
import (
	"testing"
	check "github.com/stretchr/testify/assert"
	verify "github.com/stretchr/testify/require"
	magic "cryptoutil/internal/shared/magic"
)
func TestExample(t *testing.T) {
	t.Errorf("formatted: %s", values...)
	t.Fatalf("fatal: %s", values...)
	check.Failf(t, "testing.T.Errorf failed", "legacy")
	verify.FailNowf(t, "testing.T.Fatalf failed", "legacy")
	check.Failf(t, "other failure", "unknown message")
	check.Failf(t, message, "not a literal")
	check.Failf(other, "testing.T.Errorf failed", "wrong receiver")
	helper()
	t.Errorf()
	other.Errorf("wrong receiver")
	t.Error("not formatted")
}`
	file, err := parser.ParseFile(fileSet, "sample_test.go", source, parser.ParseComments)
	require.NoError(t, err)

	edits := collectCallEdits(fileSet, file, []byte(source), packageAliases(file))

	require.Len(t, edits, 4)
	require.Contains(t, edits[0].replacement, "magic.CICDTestErrorfFailureMessage")
	require.Contains(t, edits[1].replacement, "magic.CICDTestFatalfFailureMessage")
	require.Equal(t, "magic.CICDTestErrorfFailureMessage", edits[2].replacement)
	require.Equal(t, "magic.CICDTestFatalfFailureMessage", edits[3].replacement)
	require.Contains(t, edits[0].replacement, "...")
	require.Contains(t, edits[1].replacement, "...")
}

func TestEnsureImportAs_ErrorAndAliasCollision(t *testing.T) {
	t.Parallel()

	_, err := ensureImportAs([]byte("package {"), "example.org/target", "target")
	require.ErrorContains(t, err, "failed to parse file while adding import")

	content := []byte("package example\nimport assert \"example.org/other\"\n")
	updated, err := ensureImportAs(content, "github.com/stretchr/testify/assert", "assert")
	require.NoError(t, err)
	require.Contains(t, string(updated), `assertCICD "github.com/stretchr/testify/assert"`)
}

func TestEnsureImportAs_InlineImportBlock(t *testing.T) {
	t.Parallel()

	content := []byte(`package example
import ("testing")
`)
	updated, err := ensureImportAs(content, "example.org/target", "target")
	require.NoError(t, err)
	require.Contains(t, string(updated), "import (\"testing\"\n\t\"example.org/target\"\n)")

	_, err = parser.ParseFile(token.NewFileSet(), "sample.go", updated, parser.AllErrors)
	require.NoError(t, err)
}
