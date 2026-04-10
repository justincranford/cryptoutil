// Copyright (c) 2025 Justin Cranford

package function_var_redeclaration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintGoFunctionVarRedeclaration "cryptoutil/internal/apps/tools/cicd_lint/lint_go/function_var_redeclaration"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// newLogger creates a test logger for the test.
func newLogger(t *testing.T) *cryptoutilCmdCicdCommon.Logger {
	t.Helper()

	return cryptoutilCmdCicdCommon.NewLogger(t.Name())
}

// writeFile creates a file at path with the given content.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))
}

func TestCheckInDir_NoErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setupFn func(t *testing.T) string
	}{
		{
			name: "clean code no violations",
			setupFn: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, "clean.go"), "package mypackage\n\nimport \"fmt\"\n\nvar localVar = \"hello\"\nvar number = 42\nvar computed = fmt.Sprintf(\"%d\", number)\n")

				return dir
			},
		},
		{
			name: "skips test files",
			setupFn: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, "seams_test.go"), "package mypackage_test\n\nimport \"path/filepath\"\n\nvar walkFn = filepath.Walk\n")

				return dir
			},
		},
		{
			name: "skips export_test.go",
			setupFn: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, "export_test.go"), "package mypackage\n\nimport \"path/filepath\"\n\nvar ExportedWalkFn = filepath.Walk\n")

				return dir
			},
		},
		{
			name: "skips call expressions",
			setupFn: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, "constructor.go"), "package mypackage\n\nimport \"sync\"\n\nvar mu = sync.NewMutex()\nvar once = sync.Once{}\n")

				return dir
			},
		},
		{
			name: "skips typed vars",
			setupFn: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, "typed.go"), "package p\n\nimport \"sync\"\n\nvar mu sync.Mutex\n")

				return dir
			},
		},
		{
			name: "skips non-go files",
			setupFn: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, "data.txt"), "var walkFn = filepath.Walk")
				writeFile(t, filepath.Join(dir, "Makefile"), "var walkFn = filepath.Walk")

				return dir
			},
		},
		{
			name: "skips vendor dir",
			setupFn: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				vendorDir := filepath.Join(dir, cryptoutilSharedMagic.CICDExcludeDirVendor, "somepkg")
				require.NoError(t, os.MkdirAll(vendorDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))
				writeFile(t, filepath.Join(vendorDir, "seam.go"), "package somepkg\n\nimport \"path/filepath\"\n\nvar walkFn = filepath.Walk\n")

				return dir
			},
		},
		{
			name: "skips dot dirs",
			setupFn: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				hidden := filepath.Join(dir, ".hidden", "pkg")
				require.NoError(t, os.MkdirAll(hidden, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))
				writeFile(t, filepath.Join(hidden, "seam.go"), "package pkg\n\nimport \"path/filepath\"\n\nvar walkFn = filepath.Walk\n")

				return dir
			},
		},
		{
			name: "skips underscore dirs",
			setupFn: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				underscore := filepath.Join(dir, "_internal", "pkg")
				require.NoError(t, os.MkdirAll(underscore, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))
				writeFile(t, filepath.Join(underscore, "seam.go"), "package pkg\n\nimport \"path/filepath\"\n\nvar walkFn = filepath.Walk\n")

				return dir
			},
		},
		{
			name: "parse error silently skipped",
			setupFn: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, "broken.go"), "package p THIS IS NOT VALID GO {{{")

				return dir
			},
		},
		{
			name: "nested selector not flagged",
			setupFn: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, "nested.go"), "package p\n\nimport \"net/http\"\n\nvar tlsConn = http.DefaultClient.Transport\n")

				return dir
			},
		},
		{
			name: "skips non-Fn named vars",
			setupFn: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, "defaults.go"), "package p\n\nimport \"path/filepath\"\n\nvar defaultBase = filepath.Separator\nvar configPath = filepath.Join\n")

				return dir
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := tc.setupFn(t)
			logger := newLogger(t)

			err := lintGoFunctionVarRedeclaration.CheckInDir(logger, dir, filepath.Walk)

			require.NoError(t, err)
		})
	}
}

func TestCheckInDir_Violations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupFn        func(t *testing.T) string
		wantViolations string
	}{
		{
			name: "two violations in single file",
			setupFn: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, "seam.go"), "package mypackage\n\nimport \"path/filepath\"\n\nvar walkFn = filepath.Walk\nvar absFn = filepath.Abs\n")

				return dir
			},
			wantViolations: "2 violation(s)",
		},
		{
			name: "multiple files multiple violations",
			setupFn: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, "a.go"), "package p\n\nimport \"path/filepath\"\n\nvar absFn = filepath.Abs\n")
				writeFile(t, filepath.Join(dir, "b.go"), "package p\n\nimport \"path/filepath\"\n\nvar walkFn = filepath.Walk\nvar joinFn = filepath.Join\n")

				return dir
			},
			wantViolations: "3 violation(s)",
		},
		{
			name: "var group mixed values",
			setupFn: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, "mixed.go"), "package p\n\nimport \"path/filepath\"\n\nvar (\n\twalkFn    = filepath.Walk\n\tlocalName = \"hello\"\n)\n")

				return dir
			},
			wantViolations: "1 violation(s)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := tc.setupFn(t)
			logger := newLogger(t)

			err := lintGoFunctionVarRedeclaration.CheckInDir(logger, dir, filepath.Walk)

			require.Error(t, err)
			require.Contains(t, err.Error(), "function-var-redeclaration")
			require.Contains(t, err.Error(), tc.wantViolations)
		})
	}
}

func TestCheckInDir_WalkError(t *testing.T) {
	t.Parallel()

	stubWalkFn := func(_ string, _ filepath.WalkFunc) error {
		return fmt.Errorf("injected walk error")
	}

	logger := newLogger(t)
	err := lintGoFunctionVarRedeclaration.CheckInDir(logger, ".", stubWalkFn)

	require.Error(t, err)
	require.Contains(t, err.Error(), "directory walk failed")
}

func TestCheck_NoViolationsOnCurrentCodebase(t *testing.T) {
	t.Parallel()

	// After completing the pre-work refactorings (Task 3.4 pre-work), the codebase
	// must have zero function-var redeclarations in production code.
	logger := newLogger(t)
	err := lintGoFunctionVarRedeclaration.Check(logger)

	require.NoError(t, err)
}
