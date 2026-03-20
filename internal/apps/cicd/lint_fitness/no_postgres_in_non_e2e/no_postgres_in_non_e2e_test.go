// Copyright (c) 2025 Justin Cranford
package no_postgres_in_non_e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// postgresRunContainerViolationPrefix and postgresRunContainerViolationSuffix are split to avoid
// false-positive self-detection by this linter when scanning its own test file.
var (
	postgresRunContainerViolationPrefix = "package repo_test\n\nfunc TestFoo(t *testing.T) {\n\tcontainer, err := postgres"
	postgresRunContainerViolationSuffix = ".RunContainer(ctx)\n\t_ = container\n\t_ = err\n}\n"
	postgresRunContainerViolation       = postgresRunContainerViolationPrefix + postgresRunContainerViolationSuffix
)

var (
	postgresModuleRunViolationPrefix = "package repo_test\n\nfunc TestFoo(t *testing.T) {\n\tcontainer, err := postgres"
	postgresModuleRunViolationSuffix = "Module.Run(ctx, \"postgres:18-alpine\")\n\t_ = container\n\t_ = err\n}\n"
	postgresModuleRunViolation       = postgresModuleRunViolationPrefix + postgresModuleRunViolationSuffix
)

var (
	postgresDirectRunViolationPrefix = "package repo_test\n\nfunc TestFoo(t *testing.T) {\n\tcontainer, err := postgres"
	postgresDirectRunViolationSuffix = ".Run(ctx, \"postgres:18-alpine\")\n\t_ = container\n\t_ = err\n}\n"
	postgresDirectRunViolation       = postgresDirectRunViolationPrefix + postgresDirectRunViolationSuffix
)

var (
	newPostgresHelperViolationPrefix = "package repo_test\n\nfunc TestFoo(t *testing.T) {\n\tdb := cryptoutilTestingTestdb"
	newPostgresHelperViolationSuffix = ".NewPostgresTestContainer(ctx, t)\n\t_ = db\n}\n"
	newPostgresHelperViolation       = newPostgresHelperViolationPrefix + newPostgresHelperViolationSuffix
)

var (
	requireNewPostgresViolationPrefix = "package repo_test\n\nfunc TestFoo(t *testing.T) {\n\tdb := cryptoutilTestingTestdb"
	requireNewPostgresViolationSuffix = ".RequireNewPostgresTestContainer(ctx, t, &m{})\n\t_ = db\n}\n"
	requireNewPostgresViolation       = requireNewPostgresViolationPrefix + requireNewPostgresViolationSuffix
)

var cleanUnitTestContent = "package repo_test\n\nfunc TestFoo(t *testing.T) {\n\tdb := testdb.NewInMemorySQLiteDB(t)\n\t_ = db\n}\n"

var commentedOutViolationContent = "package repo_test\n\nfunc TestFoo(t *testing.T) {\n\t// postgres.RunContainer(ctx) -- do not use\n}\n"

var e2eBuildTagContent = "//go:build e2e\n\npackage repo_test\n\nfunc TestFoo(t *testing.T) {\n\tcontainer, err := postgres.RunContainer(ctx)\n\t_ = container\n\t_ = err\n}\n"

func TestCheckFile_Violations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		content   string
		wantCount int
	}{
		{
			name:      "postgres.RunContainer violation",
			content:   postgresRunContainerViolation,
			wantCount: 1,
		},
		{
			name:      "postgresModule.Run violation",
			content:   postgresModuleRunViolation,
			wantCount: 1,
		},
		{
			name:      "postgres.Run violation",
			content:   postgresDirectRunViolation,
			wantCount: 1,
		},
		{
			name:      "NewPostgresTestContainer violation",
			content:   newPostgresHelperViolation,
			wantCount: 1,
		},
		{
			name:      "RequireNewPostgresTestContainer violation",
			content:   requireNewPostgresViolation,
			wantCount: 1,
		},
		{
			name:      "clean SQLite unit test - no violation",
			content:   cleanUnitTestContent,
			wantCount: 0,
		},
		{
			name:      "commented-out postgres call - no violation",
			content:   commentedOutViolationContent,
			wantCount: 0,
		},
		{
			name:      "file with //go:build e2e tag - allowed",
			content:   e2eBuildTagContent,
			wantCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpFile, err := os.CreateTemp(t.TempDir(), "*_test.go")
			require.NoError(t, err)

			_, err = tmpFile.WriteString(tc.content)
			require.NoError(t, err)

			require.NoError(t, tmpFile.Close())

			violations := CheckFile(tmpFile.Name())
			require.Len(t, violations, tc.wantCount)
		})
	}
}

func TestCheckFiles_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckFiles(logger, []string{})
	require.NoError(t, err)
}

func TestCheckFiles_WithViolation(t *testing.T) {
	t.Parallel()

	tmpFile, err := os.CreateTemp(t.TempDir(), "*_test.go")
	require.NoError(t, err)

	_, err = tmpFile.WriteString(postgresRunContainerViolation)
	require.NoError(t, err)

	require.NoError(t, tmpFile.Close())

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = CheckFiles(logger, []string{tmpFile.Name()})
	require.Error(t, err)
	require.Contains(t, err.Error(), "violation")
}

func TestCheckInDir_SkipsE2EFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "foo_e2e_test.go")

	err := os.WriteFile(filePath, []byte(postgresRunContainerViolation), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = CheckInDir(logger, dir)
	require.NoError(t, err)
}

func TestCheckInDir_FlagsIntegrationFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "foo_integration_test.go")

	err := os.WriteFile(filePath, []byte(postgresRunContainerViolation), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = CheckInDir(logger, dir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "violation")
}

func TestCheckInDir_SkipsAllowedPathFragments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		dirPath  string
		filename string
	}{
		{"testdb package", "testing/testdb", "foo_test.go"},
		{"container package", "shared/container", "foo_test.go"},
		{"lint_fitness package", "lint_fitness/some_linter", "foo_test.go"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			subDir := filepath.Join(dir, filepath.FromSlash(tc.dirPath))

			err := os.MkdirAll(subDir, cryptoutilSharedMagic.DirPermissions)
			require.NoError(t, err)

			filePath := filepath.Join(subDir, tc.filename)

			err = os.WriteFile(filePath, []byte(postgresRunContainerViolation), cryptoutilSharedMagic.CacheFilePermissions)
			require.NoError(t, err)

			logger := cryptoutilCmdCicdCommon.NewLogger("test")
			err = CheckInDir(logger, dir)
			require.NoError(t, err)
		})
	}
}

func TestCheckInDir_FlagsBusinesslogicFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	subDir := filepath.Join(dir, filepath.FromSlash("service/server/businesslogic"))

	err := os.MkdirAll(subDir, cryptoutilSharedMagic.DirPermissions)
	require.NoError(t, err)

	filePath := filepath.Join(subDir, "foo_test.go")

	err = os.WriteFile(filePath, []byte(postgresRunContainerViolation), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = CheckInDir(logger, dir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "violation")
}

func TestCheckInDir_SkipsGitAndVendorDirs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		dirName string
	}{
		{"git dir", cryptoutilSharedMagic.CICDExcludeDirGit},
		{"vendor dir", cryptoutilSharedMagic.CICDExcludeDirVendor},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			subDir := filepath.Join(dir, tc.dirName)

			err := os.MkdirAll(subDir, cryptoutilSharedMagic.DirPermissions)
			require.NoError(t, err)

			filePath := filepath.Join(subDir, "foo_test.go")

			err = os.WriteFile(filePath, []byte(postgresRunContainerViolation), cryptoutilSharedMagic.CacheFilePermissions)
			require.NoError(t, err)

			logger := cryptoutilCmdCicdCommon.NewLogger("test")
			err = CheckInDir(logger, dir)
			require.NoError(t, err)
		})
	}
}

func TestHasE2EBuildTag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name:    "has //go:build e2e",
			content: "//go:build e2e\n\npackage foo\n",
			want:    true,
		},
		{
			name:    "has //go:build integration,e2e",
			content: "//go:build integration,e2e\n\npackage foo\n",
			want:    true,
		},
		{
			name:    "no build tag",
			content: "package foo\n\nfunc TestFoo() {}\n",
			want:    false,
		},
		{
			name:    "build tag after first maxBuildTagLines lines - not detected",
			content: fmt.Sprintf("%s\n//go:build e2e\n", generateLines(maxBuildTagLines+1)),
			want:    false,
		},
		{
			name:    "unrelated build tag",
			content: "//go:build integration\n\npackage foo\n",
			want:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := hasE2EBuildTag(tc.content)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestCheckFile_ReadError(t *testing.T) {
	t.Parallel()

	violations := CheckFile("/nonexistent/path/totally_missing_test.go")

	require.Len(t, violations, 1)
	require.Contains(t, violations[0], "error reading file")
}

func TestCheckInDir_WalkError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, "/nonexistent/directory/that/does/not/exist")

	require.Error(t, err)
}

// generateLines creates a string with n empty comment lines.
func generateLines(n int) string {
	result := ""

	for i := range n {
		result += fmt.Sprintf("// line %d\n", i+1)
	}

	return result
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}

		dir = parent
	}
}

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheck_Integration(t *testing.T) {
	root, err := findProjectRoot()
	if err != nil {
		t.Skip("Skipping - cannot find project root")
	}

	origDir, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(root))

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-no-postgres-in-non-e2e")

	err = Check(logger)
	require.NoError(t, err)
}
