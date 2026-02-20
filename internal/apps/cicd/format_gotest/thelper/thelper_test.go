// Copyright (c) 2025 Justin Cranford

package thelper

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

func TestFix_EmptyDir(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	processed, modified, fixed, err := Fix(logger, tmpDir)

	require.NoError(t, err)
	require.Zero(t, processed)
	require.Zero(t, modified)
	require.Zero(t, fixed)
}

func TestFix_NonTestFile(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	// Write a non-test Go file (should be ignored by Fix).
	content := "package foo\n\nfunc helper() {}\n"
	err := os.WriteFile(filepath.Join(tmpDir, "helper.go"), []byte(content), 0o600)
	require.NoError(t, err)

	processed, modified, fixed, err := Fix(logger, tmpDir)

	require.NoError(t, err)
	require.Zero(t, processed)
	require.Zero(t, modified)
	require.Zero(t, fixed)
}

func TestFix_TestFileNoHelpers(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	content := "package foo\n\nimport \"testing\"\n\nfunc TestSomething(t *testing.T) {\n\tt.Parallel()\n}\n"
	err := os.WriteFile(filepath.Join(tmpDir, "something_test.go"), []byte(content), 0o600)
	require.NoError(t, err)

	processed, modified, fixed, err := Fix(logger, tmpDir)

	require.NoError(t, err)
	require.Equal(t, 1, processed)
	require.Zero(t, modified)
	require.Zero(t, fixed)
}

func TestFix_AddsTHelperToSetupFunc(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	// Write test file with setup helper missing t.Helper().
	content := "package foo\n\nimport \"testing\"\n\nfunc setupSomething(t *testing.T) {\n\tt.Log(\"setup\")\n}\n"
	err := os.WriteFile(filepath.Join(tmpDir, "something_test.go"), []byte(content), 0o600)
	require.NoError(t, err)

	processed, modified, fixed, err := Fix(logger, tmpDir)

	require.NoError(t, err)
	require.Equal(t, 1, processed)
	require.Equal(t, 1, modified)
	require.Equal(t, 1, fixed)

	resultBytes, readErr := os.ReadFile(filepath.Join(tmpDir, "something_test.go"))
	require.NoError(t, readErr)
	require.Contains(t, string(resultBytes), "t.Helper()")
}

func TestFix_AlreadyHasTHelper(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	content := "package foo\n\nimport \"testing\"\n\nfunc setupSomething(t *testing.T) {\n\tt.Helper()\n\tt.Log(\"setup\")\n}\n"
	err := os.WriteFile(filepath.Join(tmpDir, "something_test.go"), []byte(content), 0o600)
	require.NoError(t, err)

	processed, modified, fixed, err := Fix(logger, tmpDir)

	require.NoError(t, err)
	require.Equal(t, 1, processed)
	require.Zero(t, modified)
	require.Zero(t, fixed)
}

func TestFix_WalkError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	badDir := filepath.Join(tmpDir, "baddir")
	require.NoError(t, os.MkdirAll(badDir, 0o000))
	t.Cleanup(func() { _ = os.Chmod(badDir, 0o700) })

	_, _, _, err := Fix(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to walk directory")
}

func TestFix_InvalidGoFile(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	err := os.WriteFile(filepath.Join(tmpDir, "invalid_test.go"), []byte("this is not valid Go code!"), 0o600)
	require.NoError(t, err)

	_, _, _, err = Fix(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to process")
}

func TestIsTestHelperFunction_Patterns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		funcName  string
		wantMatch bool
	}{
		{name: "setup prefix", funcName: "setupDB", wantMatch: true},
		{name: "check prefix", funcName: "checkBalance", wantMatch: true},
		{name: "assert prefix", funcName: "assertValid", wantMatch: true},
		{name: "verify prefix", funcName: "verifyOutput", wantMatch: true},
		{name: "helper prefix", funcName: "helperCreate", wantMatch: true},
		{name: "create prefix", funcName: "createUser", wantMatch: true},
		{name: "build prefix", funcName: "buildRequest", wantMatch: true},
		{name: "mock prefix", funcName: "mockService", wantMatch: true},
		{name: "Test prefix excluded", funcName: "TestSomething", wantMatch: false},
		{name: "no matching prefix", funcName: "processData", wantMatch: false},
		{name: "uppercase Setup", funcName: "Setup", wantMatch: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			funcDecl := &ast.FuncDecl{Name: &ast.Ident{Name: tc.funcName}}
			result := isTestHelperFunction(funcDecl)
			require.Equal(t, tc.wantMatch, result)
		})
	}
}

func TestHasTHelperCall_Detection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		body     *ast.BlockStmt
		expected bool
	}{
		{name: "nil body", body: nil, expected: false},
		{name: "empty body", body: &ast.BlockStmt{List: []ast.Stmt{}}, expected: false},
		{
			name: "has t.Helper()",
			body: &ast.BlockStmt{List: []ast.Stmt{
				&ast.ExprStmt{X: &ast.CallExpr{Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "t"}, Sel: &ast.Ident{Name: "Helper"}}}},
			}},
			expected: true,
		},
		{
			name: "other method call only",
			body: &ast.BlockStmt{List: []ast.Stmt{
				&ast.ExprStmt{X: &ast.CallExpr{Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "t"}, Sel: &ast.Ident{Name: "Log"}}}},
			}},
			expected: false,
		},
		{
			name:     "non-expr statement",
			body:     &ast.BlockStmt{List: []ast.Stmt{&ast.ReturnStmt{}}},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			funcDecl := &ast.FuncDecl{Body: tc.body}
			result := hasTHelperCall(funcDecl)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestGetTestingParam_Extraction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "testing.T parameter named t",
			content:  "package foo\nimport \"testing\"\nfunc setupDB(t *testing.T) {}\n",
			expected: "t",
		},
		{
			name:     "testing.T parameter named tb",
			content:  "package foo\nimport \"testing\"\nfunc setupDB(tb *testing.T) {}\n",
			expected: "tb",
		},
		{
			name:     "testing.B parameter",
			content:  "package foo\nimport \"testing\"\nfunc setupBench(b *testing.B) {}\n",
			expected: "b",
		},
		{
			name:     "no testing parameter",
			content:  "package foo\nfunc helper(x int) {}\n",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "test_helper.go")
			err := os.WriteFile(filePath, []byte(tc.content), 0o600)
			require.NoError(t, err)

			fset := token.NewFileSet()
			node, parseErr := parser.ParseFile(fset, filePath, nil, 0)
			require.NoError(t, parseErr)

			var result string

			for _, decl := range node.Decls {
				if funcDecl, ok := decl.(*ast.FuncDecl); ok {
					result = getTestingParam(funcDecl)

					break
				}
			}

			require.Equal(t, tc.expected, result)
		})
	}
}
