// Copyright (c) 2025 Justin Cranford

package no_unit_test_real_server

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// listenViolationContent contains Go code that calls app.Listen() in a unit test.
// Split across two package-level vars to avoid false-positive self-detection by the linter.
var (
	listenViolationPrefix  = "package handler_test\n\nfunc TestFoo(t *testing.T) {\n\tapp := fiber.New()\n\tgo app."
	listenViolationSuffix  = `Listen(":8080")` + "\n}\n"
	listenViolationContent = listenViolationPrefix + listenViolationSuffix
)

var listenAndServeViolationContent = "package handler_test\n\nfunc TestFoo(t *testing.T) {\n\thttp." +
	`ListenAndServe(":8080", nil)` + "\n}\n"

var listenAndServeTLSViolationContent = "package handler_test\n\nfunc TestFoo(t *testing.T) {\n\thttp." +
	`ListenAndServeTLS(":8080", "cert", "key", nil)` + "\n}\n"

func TestCheckFile_Violations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fileContent string
		wantIssues  bool
	}{
		{
			name:        "app.Listen violation",
			fileContent: listenViolationContent,
			wantIssues:  true,
		},
		{
			name:        "ListenAndServe violation",
			fileContent: listenAndServeViolationContent,
			wantIssues:  true,
		},
		{
			name:        "ListenAndServeTLS violation",
			fileContent: listenAndServeTLSViolationContent,
			wantIssues:  true,
		},
		{
			name:        "no violation - app.Test",
			fileContent: "package handler_test\n\nfunc TestFoo(t *testing.T) {\n\tapp := fiber.New()\n\tresp, _ := app.Test(req, -1)\n\t_ = resp\n}\n",
			wantIssues:  false,
		},
		{
			name:        "no violation - comment",
			fileContent: "package handler_test\n\n// app should use app.Test() for testing.\nfunc TestFoo(t *testing.T) {}\n",
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

func TestCheckFiles_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckFiles(logger, []string{})

	require.NoError(t, err)
}

func TestCheckFiles_WithViolation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "handler_test.go")
	err := os.WriteFile(testFile, []byte(listenViolationContent), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = CheckFiles(logger, []string{testFile})

	require.Error(t, err)
	require.Contains(t, err.Error(), "violation")
}

func TestCheckInDir_SkipsIntegrationAndE2E(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
	}{
		{"integration test", "handler_integration_test.go"},
		{"e2e test", "handler_e2e_test.go"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, tc.filename)
			err := os.WriteFile(testFile, []byte(listenViolationContent), cryptoutilSharedMagic.CacheFilePermissions)
			require.NoError(t, err)

			logger := cryptoutilCmdCicdCommon.NewLogger("test")
			err = CheckInDir(logger, tmpDir)

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
		{"skips .git dir", cryptoutilSharedMagic.CICDExcludeDirGit},
		{"skips vendor dir", cryptoutilSharedMagic.CICDExcludeDirVendor},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			subDir := filepath.Join(tmpDir, tc.dirName)
			err := os.MkdirAll(subDir, cryptoutilSharedMagic.DirPermissions)
			require.NoError(t, err)

			ignoredFile := filepath.Join(subDir, "something_test.go")
			err = os.WriteFile(ignoredFile, []byte(listenViolationContent), cryptoutilSharedMagic.CacheFilePermissions)
			require.NoError(t, err)

			logger := cryptoutilCmdCicdCommon.NewLogger("test")
			err = CheckInDir(logger, tmpDir)

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

	logger := cryptoutilCmdCicdCommon.NewLogger("test-no-unit-test-real-server")

	err = Check(logger)
	require.NoError(t, err)
}
