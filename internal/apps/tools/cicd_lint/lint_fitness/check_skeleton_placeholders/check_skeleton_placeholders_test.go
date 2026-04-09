// Copyright (c) 2025 Justin Cranford

package check_skeleton_placeholders

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	goSkeletonFuncContent = "package myservice\n\nfunc NewSkeletonService() {}\n"
	goSkeletonTypeContent = "package myservice\n\ntype SkeletonServer struct{}\n"
)

// Note: This test file is in internal/apps/cicd/ which is excluded from lint-skeleton scans.
// Therefore skeleton-related words can appear here without triggering violations.

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	err := os.WriteFile(path, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)
}

func mkdirAndWriteFile(t *testing.T, tempDir, subDir, filename, content string) {
	t.Helper()

	dir := filepath.Join(tempDir, subDir)
	require.NoError(t, os.MkdirAll(dir, 0o700))

	writeFile(t, filepath.Join(dir, filename), content)
}

func TestFindViolations_DetectsPlaceholders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		content  string
		wantWord string
		wantLine int
	}{
		{name: "lowercase skeleton in func name", filename: "bad.go", content: goSkeletonFuncContent, wantWord: cryptoutilSharedMagic.SkeletonProductNameTitleCase, wantLine: 3},
		{name: "titlecase Skeleton in type name", filename: "bad.go", content: goSkeletonTypeContent, wantWord: cryptoutilSharedMagic.SkeletonProductNameTitleCase, wantLine: 3},
		{name: "uppercase SKELETON in const name", filename: "bad.go", content: "package myservice\n\nconst SKELETON_PORT = 8080\n", wantWord: cryptoutilSharedMagic.SkeletonProductNameUpperCase, wantLine: 3},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			badFile := filepath.Join(tempDir, tc.filename)
			writeFile(t, badFile, tc.content)

			violations, err := FindViolations(tempDir)
			require.NoError(t, err)
			require.NotEmpty(t, violations)
			require.Equal(t, badFile, violations[0].File)
			require.Equal(t, tc.wantWord, violations[0].Word)
			require.Equal(t, tc.wantLine, violations[0].Line)
		})
	}
}

func TestFindViolations_ExcludedPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		subDir  string
		file    string
		content string
	}{
		{
			name:    "test files skipped",
			file:    "bad_test.go",
			content: "package myservice\n\nfunc TestSkeletonService(t *testing.T) {}\n",
		},
		{
			name:    "internal/apps/skeleton dir excluded",
			subDir:  filepath.Join("internal", "apps", cryptoutilSharedMagic.SkeletonProductName, cryptoutilSharedMagic.SkeletonTemplateServiceName),
			file:    "template.go",
			content: "package template\n\n// Skeleton template service\nfunc Skeleton() {}\n",
		},
		{
			name:    "cmd/skeleton-template dir excluded",
			subDir:  filepath.Join("cmd", cryptoutilSharedMagic.OTLPServiceSkeletonTemplate),
			file:    "main.go",
			content: "package main\n\n// Entry point for skeleton-template service\nfunc main() {}\n",
		},
		{
			name:    "cmd/skeleton dir excluded",
			subDir:  filepath.Join("cmd", cryptoutilSharedMagic.SkeletonProductName),
			file:    "main.go",
			content: "package main\n\n// Skeleton product entry point\nimport \"cryptoutil/internal/apps/skeleton\"\n",
		},
		{
			name:    "vendor dir excluded",
			subDir:  filepath.Join(cryptoutilSharedMagic.CICDExcludeDirVendor, "somelib"),
			file:    "skeleton.go",
			content: "package somelib\n\ntype SkeletonServer struct{}\n",
		},
		{
			name:    "internal/shared/magic dir excluded",
			subDir:  filepath.Join("internal", "shared", "magic"),
			file:    "magic_skeleton.go",
			content: "package magic\n\nconst SkeletonPort = 8900\n",
		},
		{
			name:    "internal/apps/cryptoutil dir excluded",
			subDir:  filepath.Join("internal", "apps", cryptoutilSharedMagic.DefaultOTLPServiceDefault),
			file:    "cryptoutil.go",
			content: "package cryptoutil\n\nimport \"cryptoutil/internal/apps/skeleton\"\n",
		},
		{
			name:    "internal/apps/tools/cicd_lint dir excluded",
			subDir:  filepath.Join("internal", "apps", "tools", "cicd_lint"),
			file:    "cicd.go",
			content: "package cicd_lint\n\nconst SkeletonCmd = \"lint-skeleton\"\n",
		},
		{
			name:    "internal/apps/framework dir excluded",
			subDir:  filepath.Join("internal", "apps", cryptoutilSharedMagic.FrameworkProductName),
			file:    "server.go",
			content: "package framework\n\n// Supported services: skeleton-template\n",
		},
		{
			name:    "api/skeleton-template dir excluded",
			subDir:  filepath.Join("api", cryptoutilSharedMagic.OTLPServiceSkeletonTemplate),
			file:    "generate.go",
			content: "// Package skeletontemplate provides generated OpenAPI code.\npackage skeletontemplate\n",
		},
		{
			name:    "non-go files ignored",
			file:    "myfile.yaml",
			content: "name: Skeleton-service\n",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			if tc.subDir != "" {
				mkdirAndWriteFile(t, tempDir, tc.subDir, tc.file, tc.content)
			} else {
				writeFile(t, filepath.Join(tempDir, tc.file), tc.content)
			}

			violations, err := FindViolations(tempDir)
			require.NoError(t, err)
			require.Empty(t, violations)
		})
	}
}

func TestFindViolations_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupFn   func(t *testing.T) string
		wantCount int
	}{
		{
			name: "clean file has no violations",
			setupFn: func(t *testing.T) string {
				t.Helper()

				tempDir := t.TempDir()
				writeFile(t, filepath.Join(tempDir, "clean.go"), "package myservice\n\nfunc NewMyService() {}\n")

				return tempDir
			},
			wantCount: 0,
		},
		{
			name: "line reported only once even with multiple matching words",
			setupFn: func(t *testing.T) string {
				t.Helper()

				tempDir := t.TempDir()
				writeFile(t, filepath.Join(tempDir, "bad.go"), "package myservice\n\n// line with skeleton and Skeleton keywords\n")

				return tempDir
			},
			wantCount: 1,
		},
		{
			name: "walk on single file path finds no violations",
			setupFn: func(t *testing.T) string {
				t.Helper()

				tempDir := t.TempDir()
				f := filepath.Join(tempDir, "notadir.go")
				writeFile(t, f, "package test\n")

				return f
			},
			wantCount: 0,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := tc.setupFn(t)

			violations, err := FindViolations(path)
			require.NoError(t, err)
			require.Len(t, violations, tc.wantCount)
		})
	}
}

func TestCheckInDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupFn     func(t *testing.T) string
		wantErr     bool
		wantErrMsgs []string
	}{
		{
			name: "passes with clean file",
			setupFn: func(t *testing.T) string {
				t.Helper()

				tempDir := t.TempDir()
				writeFile(t, filepath.Join(tempDir, "clean.go"), "package myservice\n\nfunc NewMyService() {}\n")

				return tempDir
			},
		},
		{
			name: "fails with skeleton violation",
			setupFn: func(t *testing.T) string {
				t.Helper()

				tempDir := t.TempDir()
				writeFile(t, filepath.Join(tempDir, "bad.go"), goSkeletonTypeContent)

				return tempDir
			},
			wantErr:     true,
			wantErrMsgs: []string{"[ValidateSkeleton]", "ENG-HANDBOOK.md Section 5.1"},
		},
		{
			name: "fails with invalid root directory",
			setupFn: func(t *testing.T) string {
				t.Helper()

				return "/nonexistent/path/that/does/not/exist"
			},
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := tc.setupFn(t)
			logger := cryptoutilCmdCicdCommon.NewLogger("test")

			err := CheckInDir(logger, dir)
			if tc.wantErr {
				require.Error(t, err)

				for _, msg := range tc.wantErrMsgs {
					require.ErrorContains(t, err, msg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheck_DelegatesCheckInDir(t *testing.T) {
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { require.NoError(t, os.Chdir(origDir)) }()

	tmpDir := t.TempDir()
	require.NoError(t, os.Chdir(tmpDir))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err = Check(logger)
	require.NoError(t, err)
}

func TestScanFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupFn   func(t *testing.T) string
		wantCount int
		wantErr   bool
	}{
		{
			name: "empty file has no violations",
			setupFn: func(t *testing.T) string {
				t.Helper()

				tempDir := t.TempDir()
				f := filepath.Join(tempDir, "empty.go")
				writeFile(t, f, "")

				return f
			},
		},
		{
			name: "multiple violations in file reported separately",
			setupFn: func(t *testing.T) string {
				t.Helper()

				tempDir := t.TempDir()
				f := filepath.Join(tempDir, "multi.go")
				writeFile(t, f, "package myservice\n\nvar a = \"skeleton\"\nvar b = \"Skeleton\"\n")

				return f
			},
			wantCount: 2,
		},
		{
			name: "nonexistent file returns error",
			setupFn: func(t *testing.T) string {
				t.Helper()

				return "/nonexistent/path/file.go"
			},
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := tc.setupFn(t)

			violations, err := scanFile(path)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Len(t, violations, tc.wantCount)
			}
		})
	}
}
