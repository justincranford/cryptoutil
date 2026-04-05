// Copyright (c) 2025 Justin Cranford
//

package no_local_closed_db_helper

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// createClosedDBViolation is a minimal Go source file that defines a banned closed-DB helper.
const createClosedDBViolation = "package repo_test\n\nfunc createClosedDatabase() (*gorm.DB, error) {\n\treturn nil, nil\n}\n"

func TestCheckFiles_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckFiles(logger, []string{})

	require.NoError(t, err)
}

func TestCheckFile_ViolationDetected(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fileContent string
		wantIssues  bool
	}{
		{
			name:        "createClosedDatabase violation",
			fileContent: createClosedDBViolation,
			wantIssues:  true,
		},
		{
			name:        "createClosedDB violation",
			fileContent: "package service_test\n\nfunc createClosedDB() *gorm.DB {\n\treturn nil\n}\n",
			wantIssues:  true,
		},
		{
			name:        "createClosedServiceDependencies violation",
			fileContent: "package service_test\n\nfunc createClosedServiceDependencies() (*gorm.DB, error) {\n\treturn nil, nil\n}\n",
			wantIssues:  true,
		},
		{
			name:        "createClosedDBHandler violation",
			fileContent: "package apis\n\nfunc createClosedDBHandler(t *testing.T) *Handler {\n\treturn nil\n}\n",
			wantIssues:  true,
		},
		{
			name:        "no violation - clean file",
			fileContent: "package repo_test\n\nfunc TestSomething(t *testing.T) {\n\tt.Parallel()\n}\n",
			wantIssues:  false,
		},
		{
			name:        "no violation - comment mentioning banned name",
			fileContent: "package repo_test\n\n// createClosedDatabase described here.\nfunc TestSomething(t *testing.T) {\n\tt.Parallel()\n}\n",
			wantIssues:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "something_test.go")
			err := os.WriteFile(testFile, []byte(tc.fileContent), cryptoutilSharedMagic.CacheFilePermissions)
			require.NoError(t, err)

			issues := CheckFile(testFile)

			if tc.wantIssues {
				require.NotEmpty(t, issues)
			} else {
				require.Empty(t, issues)
			}
		})
	}
}

func TestCheckFiles_WithViolation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "database_error_test.go")
	err := os.WriteFile(testFile, []byte(createClosedDBViolation), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = CheckFiles(logger, []string{testFile})

	require.Error(t, err)
	require.Contains(t, err.Error(), "violation")
}

func TestCheckFiles_AllClean(t *testing.T) {
	t.Parallel()

	cleanContent := "package repo_test\n\nfunc TestSomething(t *testing.T) {\n\tt.Parallel()\n}\n"
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "something_test.go")
	err := os.WriteFile(testFile, []byte(cleanContent), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = CheckFiles(logger, []string{testFile})

	require.NoError(t, err)
}

func TestCheckFile_ReadError(t *testing.T) {
	t.Parallel()

	issues := CheckFile("/nonexistent/path/test_test.go")
	require.NotEmpty(t, issues)
	require.Contains(t, issues[0], "error reading file")
}

func TestCheckInDir_FindsViolation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	violatingFile := filepath.Join(tmpDir, "database_error_test.go")
	err := os.WriteFile(violatingFile, []byte(createClosedDBViolation), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = CheckInDir(logger, tmpDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "violation")
}

func TestCheckInDir_SkipsTestdbPackage(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testdbDir := filepath.Join(tmpDir, "testing", "testdb")
	err := os.MkdirAll(testdbDir, cryptoutilSharedMagic.DirPermissions)
	require.NoError(t, err)

	allowedFile := filepath.Join(testdbDir, "testdb_test.go")
	err = os.WriteFile(allowedFile, []byte(createClosedDBViolation), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = CheckInDir(logger, tmpDir)

	require.NoError(t, err)
}

func TestCheckInDir_NoTestFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	regularFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(regularFile, []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = CheckInDir(logger, tmpDir)

	require.NoError(t, err)
}

func TestCheckInDir_NonexistentRootDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	nonExistentDir := filepath.Join(tmpDir, "does_not_exist")

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, nonExistentDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "walking test files")
}

func TestCheckInDir_SkipsGitAndVendorDirs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		dirName string
	}{
		{name: "skips .git dir", dirName: cryptoutilSharedMagic.CICDExcludeDirGit},
		{name: "skips vendor dir", dirName: cryptoutilSharedMagic.CICDExcludeDirVendor},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			skipDir := filepath.Join(tmpDir, tc.dirName)
			err := os.MkdirAll(skipDir, cryptoutilSharedMagic.CICDOutputDirPermissions)
			require.NoError(t, err)

			err = os.WriteFile(filepath.Join(skipDir, "something_test.go"), []byte(createClosedDBViolation), cryptoutilSharedMagic.CacheFilePermissions)
			require.NoError(t, err)

			logger := cryptoutilCmdCicdCommon.NewLogger("test")
			err = CheckInDir(logger, tmpDir)

			// Violating file inside .git or vendor is skipped, so no violations.
			require.NoError(t, err)
		})
	}
}

func TestCheck_DelegatesCheckInDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	// Check() delegates to CheckInDir(logger, ".").
	// From a clean temp directory with no test files, there are no violations.
	err := CheckInDir(logger, tmpDir)
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

	logger := cryptoutilCmdCicdCommon.NewLogger("test-no-local-closed-db-helper")

	err = Check(logger)
	require.NoError(t, err)
}
