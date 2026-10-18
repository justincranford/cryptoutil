// Copyright (c) 2025-2026 Justin Cranford.
package parallel

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestFixFileWithOperations_ErrorPaths(t *testing.T) {
	t.Parallel()

	injectedError := errors.New("injected file operation error")
	tests := []struct {
		name        string
		readError   error
		formatError bool
		statError   error
		writeErr    error
		wantError   string
	}{
		{name: "read error", readError: injectedError, wantError: "failed to read file"},
		{name: "format error", formatError: true, wantError: "failed to format file"},
		{name: "stat error", statError: injectedError, wantError: "failed to stat file"},
		{name: "write error", writeErr: injectedError, wantError: "failed to write file"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			filePath := filepath.Join(t.TempDir(), "sample_test.go")
			content := []byte("package example\n\nimport \"testing\"\n\nfunc TestExample(t *testing.T) {}\nfunc TestSecond(t *testing.T) {}\n")
			require.NoError(t, os.WriteFile(filePath, content, cryptoutilSharedMagic.CacheFilePermissions))

			fileOperations := parallelFileOperations{
				readFile: func(path string) ([]byte, error) {
					if tc.readError != nil {
						return nil, tc.readError
					}

					if tc.formatError {
						return []byte("package {" + strings.Repeat(" ", 2*len(content))), nil
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
			}

			_, _, err := fixFileWithOperations(filePath, fileOperations)

			require.ErrorContains(t, err, tc.wantError)

			if tc.readError != nil || tc.statError != nil || tc.writeErr != nil {
				require.ErrorIs(t, err, injectedError)
			}
		})
	}
}

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

	t.Run("processes test file", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		filePath := filepath.Join(root, "sample_test.go")
		content := []byte("package example\n\nimport \"testing\"\n\nfunc TestExample(t *testing.T) {}\n")
		require.NoError(t, os.WriteFile(filePath, content, cryptoutilSharedMagic.CacheFilePermissions))

		processed, modified, fixed, err := Fix(logger, root)

		require.NoError(t, err)
		require.Equal(t, 1, processed)
		require.Equal(t, 1, modified)
		require.Equal(t, 1, fixed)
	})
}

func TestTestingTParameter_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		source string
		want   string
	}{
		{name: "no parameters", source: "func TestNoParams() {}", want: ""},
		{name: "unnamed testing parameter", source: "func TestUnnamed(*testing.T) {}", want: ""},
		{name: "non pointer parameter", source: "func TestValue(t testing.T) {}", want: ""},
		{name: "wrong selector", source: "func TestOther(t *testing.M) {}", want: ""},
		{name: "wrong package", source: "func TestOther(t *example.T) {}", want: ""},
		{name: "first named parameter", source: "func TestNamed(t, other *testing.T) {}", want: "t"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			file, err := parser.ParseFile(token.NewFileSet(), "sample_test.go", "package example\nimport \"testing\"\n"+tc.source, 0)
			require.NoError(t, err)

			function, ok := file.Decls[len(file.Decls)-1].(*ast.FuncDecl)
			require.True(t, ok)
			require.Equal(t, tc.want, testingTParameter(function))
		})
	}
}

func TestHasParallelCall_IgnoresNestedFunctionsAndOtherReceivers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		source string
		want   bool
	}{
		{name: "nested function only", source: "func TestExample(t *testing.T) { fn := func() { t.Parallel() }; _ = fn }", want: false},
		{name: "other receiver", source: "func TestExample(t *testing.T) { other.Parallel() }", want: false},
		{name: "matching receiver", source: "func TestExample(t *testing.T) { t.Parallel() }", want: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			file, err := parser.ParseFile(token.NewFileSet(), "sample_test.go", "package example\nimport \"testing\"\n"+tc.source, 0)
			require.NoError(t, err)

			function, ok := file.Decls[len(file.Decls)-1].(*ast.FuncDecl)
			require.True(t, ok)
			require.Equal(t, tc.want, hasParallelCall(function.Body, "t"))
		})
	}
}

func TestHasSequentialComment_WindowAndPosition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		source string
		want   bool
	}{
		{name: "within window", source: "// Sequential: changes state.\nfunc TestExample(t *testing.T) {}", want: true},
		{name: "outside window", source: "// Sequential: changes state.\n\n\n\n\n\n\n\n\n\n\n\nfunc TestExample(t *testing.T) {}", want: false},
		{name: "after function", source: "func TestExample(t *testing.T) {}\n// Sequential: changes state.", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fileSet := token.NewFileSet()
			file, err := parser.ParseFile(fileSet, "sample_test.go", "package example\nimport \"testing\"\n"+tc.source, parser.ParseComments)
			require.NoError(t, err)

			function, ok := file.Decls[len(file.Decls)-1].(*ast.FuncDecl)
			require.True(t, ok)
			require.Equal(t, tc.want, hasSequentialComment(file, fileSet, function))
		})
	}
}
