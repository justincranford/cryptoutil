// Copyright (c) 2025-2026 Justin Cranford.
package test_presence

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const testFailedToWalkAppsErrorMessage = "failed to walk internal/apps"

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

func TestCheck_RealWorkspace(t *testing.T) {
	t.Parallel()

	root, err := findProjectRoot()
	if err != nil {
		t.Skip("Skipping integration test - cannot find project root (no go.mod)")
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = CheckInDir(logger, root)
	require.NoError(t, err)
}

func TestCheckInDir_ReturnsMissingAppsDirectory(t *testing.T) {
	t.Parallel()

	err := CheckInDir(cryptoutilCmdCicdCommon.NewLogger(t.Name()), t.TempDir())

	require.ErrorContains(t, err, "internal/apps directory not found")
}

func TestCheckInDir_AllValid(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	pkgDir := filepath.Join(tmpDir, "internal", "apps", "myproduct")
	require.NoError(t, os.MkdirAll(pkgDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "logic.go"), []byte("package myproduct\nfunc Hello() {}"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "logic_test.go"), []byte("package myproduct"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_MissingTestFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	pkgDir := filepath.Join(tmpDir, "internal", "apps", "myproduct")
	require.NoError(t, os.MkdirAll(pkgDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "logic.go"), []byte("package myproduct\nfunc Hello() {}"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "test presence violations")
	require.Contains(t, err.Error(), "missing test files")
}

func TestCheckInDir_ExcludedMagicDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	magicDir := filepath.Join(tmpDir, "internal", "apps", "shared", "magic")
	require.NoError(t, os.MkdirAll(magicDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(magicDir, "constants.go"), []byte("package magic\nfunc X() {}"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_ExcludedUnifiedDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	unifiedDir := filepath.Join(tmpDir, "internal", "apps", cryptoutilSharedMagic.SkeletonProductName, "ja", "unified")
	require.NoError(t, os.MkdirAll(unifiedDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(unifiedDir, "server.go"), []byte("package unified\nfunc X() {}"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_ExcludedArchivedDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	archivedDir := filepath.Join(tmpDir, "internal", "apps", cryptoutilSharedMagic.SMProductName, "_ca-archived", cryptoutilSharedMagic.CMD_SERVER)
	require.NoError(t, os.MkdirAll(archivedDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(archivedDir, "server.go"), []byte("package server\nfunc X() {}"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_ExcludedTestingDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testingDir := filepath.Join(tmpDir, "internal", "apps", cryptoutilSharedMagic.SkeletonTemplateServiceName, "service", "testing", "helpers")
	require.NoError(t, os.MkdirAll(testingDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(testingDir, "helpers.go"), []byte("package helpers\nfunc X() {}"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_ExcludedCmdMainDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cmdMainDir := filepath.Join(tmpDir, "internal", "apps", cryptoutilSharedMagic.IdentityProductName, "cmd", "main", cryptoutilSharedMagic.AuthzServiceName)
	require.NoError(t, os.MkdirAll(cmdMainDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(cmdMainDir, "authz.go"), []byte("package authz\nfunc X() {}"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_GeneratedFilesIgnored(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	pkgDir := filepath.Join(tmpDir, "internal", "apps", "myproduct")
	require.NoError(t, os.MkdirAll(pkgDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "server.gen.go"), []byte("package myproduct\nfunc X() {}"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "models_gen.go"), []byte("package myproduct\nfunc X() {}"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_EmbedOnlyPackageSkipped(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	pkgDir := filepath.Join(tmpDir, "internal", "apps", "myproduct", "repository")
	require.NoError(t, os.MkdirAll(pkgDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	// File with only embed.FS variable declaration (no func declarations).
	content := "package repository\nimport \"embed\"\n//go:embed migrations/*.sql\nvar MigrationsFS embed.FS\n"
	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "migrations.go"), []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_NoAppsDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "internal/apps directory not found")
}

func TestCheckDirForGoFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		files      map[string]string
		missingDir bool
		readDirErr bool
		wantLogic  bool
		wantTests  bool
		wantErr    string
	}{
		{name: "empty directory"},
		{name: "test file only", files: map[string]string{"logic_test.go": "package pkg"}, wantTests: true},
		{name: "production logic", files: map[string]string{"logic.go": "package pkg\nfunc Hello() {}"}, wantLogic: true},
		{name: "nonexistent directory", missingDir: true, wantErr: testPresenceReadDirectoryErrorPrefix},
		{name: "read directory error", readDirErr: true, wantErr: testPresenceReadDirectoryErrorPrefix},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := filepath.Join(t.TempDir(), "package")
			if !tc.missingDir {
				require.NoError(t, os.MkdirAll(dir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

				for name, content := range tc.files {
					require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), cryptoutilSharedMagic.CacheFilePermissions))
				}
			}

			operations := productionTestPresenceOperations()
			if tc.readDirErr {
				operations.readDirFn = func(string) ([]os.DirEntry, error) {
					return nil, fmt.Errorf("injected read directory error")
				}
			}

			hasLogic, hasTests, err := checkDirForGoFiles(dir, operations)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantLogic, hasLogic)
			require.Equal(t, tc.wantTests, hasTests)
		})
	}
}

func TestFileHasLogic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		content   []byte
		readErr   bool
		wantLogic bool
		wantErr   string
	}{
		{name: "function declaration", content: []byte("package pkg\nfunc Hello() {}"), wantLogic: true},
		{name: "variable only", content: []byte("package pkg\nvar X = 5")},
		{name: "read failure", readErr: true, wantErr: "failed to read Go source file"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			operations := productionTestPresenceOperations()
			operations.readFileFn = func(string) ([]byte, error) {
				if tc.readErr {
					return nil, fmt.Errorf("injected Go file read error")
				}

				return tc.content, nil
			}

			hasLogic, err := fileHasLogic("logic.go", operations)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantLogic, hasLogic)
		})
	}
}

func TestCheckInDir_MultiplePackages(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Package with logic and tests.
	pkg1 := filepath.Join(tmpDir, "internal", "apps", "prod1")
	require.NoError(t, os.MkdirAll(pkg1, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(pkg1, "code.go"), []byte("package prod1\nfunc X() {}"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(pkg1, "code_test.go"), []byte("package prod1"), cryptoutilSharedMagic.CacheFilePermissions))

	// Package with logic but missing tests.
	pkg2 := filepath.Join(tmpDir, "internal", "apps", "prod2")
	require.NoError(t, os.MkdirAll(pkg2, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(pkg2, "code.go"), []byte("package prod2\nfunc Y() {}"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "prod2")
}

func TestCheck_FromProjectRoot(t *testing.T) {
	t.Parallel()

	root, err := findProjectRoot()
	if err != nil {
		t.Skip("Skipping - cannot find project root")
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = CheckInDir(logger, root)
	require.NoError(t, err)
}

func TestCheckInDir_OperationPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(*testing.T, *testPresenceOperations, string)
		wantErr string
	}{
		{
			name: "empty apps directory succeeds",
		},
		{
			name: "stat error is reported",
			setup: func(_ *testing.T, operations *testPresenceOperations, _ string) {
				operations.statFn = func(string) (os.FileInfo, error) {
					return nil, fmt.Errorf("injected apps stat error")
				}
			},
			wantErr: "failed to stat internal/apps directory",
		},
		{
			name: "walk error is reported",
			setup: func(_ *testing.T, operations *testPresenceOperations, _ string) {
				operations.walkFn = func(string, filepath.WalkFunc) error {
					return fmt.Errorf("injected walk error")
				}
			},
			wantErr: testFailedToWalkAppsErrorMessage,
		},
		{
			name: "walk callback error is reported",
			setup: func(_ *testing.T, operations *testPresenceOperations, _ string) {
				operations.walkFn = func(root string, walkFn filepath.WalkFunc) error {
					info, err := os.Stat(root)
					if err != nil {
						return err
					}

					return walkFn(root, info, fmt.Errorf("injected callback error"))
				}
			},
			wantErr: testFailedToWalkAppsErrorMessage,
		},
		{
			name: "read directory error is reported",
			setup: func(t *testing.T, operations *testPresenceOperations, appsDir string) {
				pkgDir := filepath.Join(appsDir, "myservice")
				require.NoError(t, os.MkdirAll(pkgDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

				operations.readDirFn = func(string) ([]os.DirEntry, error) {
					return nil, fmt.Errorf("injected directory read error")
				}
			},
			wantErr: testPresenceInspectGoFilesErrorPrefix,
		},
		{
			name: "source file read error is reported",
			setup: func(t *testing.T, operations *testPresenceOperations, appsDir string) {
				pkgDir := filepath.Join(appsDir, "myservice")
				require.NoError(t, os.MkdirAll(pkgDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
				require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "logic.go"), []byte("package myservice\nfunc Handle() {}"), cryptoutilSharedMagic.CacheFilePermissions))

				operations.readFileFn = func(string) ([]byte, error) {
					return nil, fmt.Errorf("injected source file read error")
				}
			},
			wantErr: testPresenceInspectGoFilesErrorPrefix,
		},
		{
			name: "relative path error is reported",
			setup: func(t *testing.T, operations *testPresenceOperations, appsDir string) {
				pkgDir := filepath.Join(appsDir, "myservice")
				require.NoError(t, os.MkdirAll(pkgDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
				require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "logic.go"), []byte("package myservice\nfunc Handle() {}"), cryptoutilSharedMagic.CacheFilePermissions))

				operations.relFn = func(string, string) (string, error) {
					return "", fmt.Errorf("injected relative path error")
				}
			},
			wantErr: "failed to get path relative to workspace root",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootDir := t.TempDir()
			appsDir := filepath.Join(rootDir, "internal", "apps")
			require.NoError(t, os.MkdirAll(appsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

			operations := productionTestPresenceOperations()
			if tc.setup != nil {
				tc.setup(t, &operations, appsDir)
			}

			err := checkInDirWithOperations(cryptoutilCmdCicdCommon.NewLogger("test"), rootDir, operations)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
		})
	}
}
