// Copyright (c) 2025 Justin Cranford
package no_unit_test_real_db

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// postgresViolationPrefix and postgresViolationSuffix are split to avoid
// false-positive self-detection by this linter when scanning its own test file.
var (
	postgresViolationPrefix  = "package repo_test\n\nfunc TestFoo(t *testing.T) {\n\tcontainer, err := postgres"
	postgresViolationSuffix  = "Module.Run(ctx, \"postgres:16-alpine\")\n\t_ = container\n\t_ = err\n}\n"
	postgresViolationContent = postgresViolationPrefix + postgresViolationSuffix
)

var postgresRunContainerViolationContent = "package repo_test\n\nfunc TestFoo(t *testing.T) {\n\tcontainer, err := postgres.RunContainer(ctx)\n\t_ = container\n\t_ = err\n}\n"

var (
	newPostgresHelperViolationPrefix  = "package repo_test\n\nfunc TestFoo(t *testing.T) {\n\tdb := cryptoutilTestingTestdb"
	newPostgresHelperViolationSuffix  = ".NewPostgresTestContainer(ctx, t)\n\t_ = db\n}\n"
	newPostgresHelperViolationContent = newPostgresHelperViolationPrefix + newPostgresHelperViolationSuffix
)

var cleanUnitTestContent = "package repo_test\n\nfunc TestFoo(t *testing.T) {\n\tdb := testdb.NewInMemorySQLiteDB(t)\n\t_ = db\n}\n"

var commentedOutViolationContent = "package repo_test\n\nfunc TestFoo(t *testing.T) {\n\t// postgres.RunContainer(ctx) -- do not use\n}\n"

var testMainExemptContent = "package repo_test\n\nimport \"os\"\n\nfunc TestMain(m *testing.M) {\n\tcontainer, err := postgresModule.Run(ctx, \"postgres:latest\")\n\t_ = container\n\t_ = err\n\tos.Exit(m.Run())\n}\n"

func TestCheckFile_Violations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		content    string
		wantCount  int
		wantErrMsg string
	}{
		{
			name:       "postgresModule.Run violation",
			content:    postgresViolationContent,
			wantCount:  1,
			wantErrMsg: "real database container",
		},
		{
			name:       "postgres.RunContainer violation",
			content:    postgresRunContainerViolationContent,
			wantCount:  1,
			wantErrMsg: "real database container",
		},
		{
			name:      "NewPostgresTestContainer violation",
			content:   newPostgresHelperViolationContent,
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
			name:      "TestMain postgres - exempt (approved per ENG-HANDBOOK.md)",
			content:   testMainExemptContent,
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

			if tc.wantErrMsg != "" && len(violations) > 0 {
				require.Contains(t, violations[0], tc.wantErrMsg)
			}
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

	_, err = tmpFile.WriteString(postgresViolationContent)
	require.NoError(t, err)

	require.NoError(t, tmpFile.Close())

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = CheckFiles(logger, []string{tmpFile.Name()})
	require.Error(t, err)
	require.Contains(t, err.Error(), "violation")
}

func TestCheckInDir_SkipsIntegrationAndE2E(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
	}{
		{"integration test file", "foo_integration_test.go"},
		{"e2e test file", "foo_e2e_test.go"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			filePath := filepath.Join(dir, tc.filename)

			err := os.WriteFile(filePath, []byte(postgresViolationContent), cryptoutilSharedMagic.CacheFilePermissions)
			require.NoError(t, err)

			logger := cryptoutilCmdCicdCommon.NewLogger("test")
			err = CheckInDir(logger, dir)
			require.NoError(t, err)
		})
	}
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

			err = os.WriteFile(filePath, []byte(postgresViolationContent), cryptoutilSharedMagic.CacheFilePermissions)
			require.NoError(t, err)

			logger := cryptoutilCmdCicdCommon.NewLogger("test")
			err = CheckInDir(logger, dir)
			require.NoError(t, err)
		})
	}
}

func TestCheckInDir_AgainstCurrentCodebase(t *testing.T) {
	t.Parallel()

	// Navigate to project root so allowedPathFragments are evaluated with full paths.
	projectRoot := filepath.Join("..", "..", "..", "..", "..")

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, projectRoot)
	require.NoError(t, err)
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
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

	logger := cryptoutilCmdCicdCommon.NewLogger("test-no-unit-test-real-db")

	err = Check(logger)
	require.NoError(t, err)
}
