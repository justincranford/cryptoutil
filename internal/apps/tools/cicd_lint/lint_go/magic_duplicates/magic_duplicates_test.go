// Copyright (c) 2025 Justin Cranford

package magic_duplicates

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"

	"github.com/stretchr/testify/require"
)

// writeMagicFile creates a file inside dir with the given content.
func writeMagicFile(t *testing.T, dir, name, content string) {
	t.Helper()

	err := os.WriteFile(filepath.Join(dir, name), []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)
}

// writeGoFile creates a .go file inside a subdirectory of dir.
func writeGoFile(t *testing.T, dir, subPkg, name, content string) {
	t.Helper()

	pkgDir := filepath.Join(dir, subPkg)
	require.NoError(t, os.MkdirAll(pkgDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))

	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, name), []byte(content), cryptoutilSharedMagic.CacheFilePermissions))
}

func TestCheckMagicDuplicatesInDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		files         map[string]string
		useInvalidDir bool
		wantErr       string
	}{
		{
			name:  "no duplicates",
			files: map[string]string{"magic_strings.go": "package magic\n\nconst (\nProtocolHTTPS = \"https\"\nSchemeHTTP    = \"http\"\n)\n"},
		},
		{
			name:  "with duplicates logged not errored",
			files: map[string]string{"magic_net.go": "package magic\n\nconst (\nProtocolHTTPS = \"https\"\nSchemeHTTPS   = \"https\"\n)\n"},
		},
		{
			name:  "trivial ints still detected",
			files: map[string]string{"magic_sizes.go": "package magic\n\nconst (\nDefaultSizeA = 5\nDefaultSizeB = 5\n)\n"},
		},
		{
			name:          "invalid directory",
			useInvalidDir: true,
			wantErr:       "failed to parse magic package",
		},
		{
			name:  "single constant",
			files: map[string]string{"magic_only.go": "package magic\n\nconst Alone = \"solo\"\n"},
		},
		{
			name:  "empty package",
			files: map[string]string{"magic.go": "package magic\n"},
		},
		{
			name: "multi file duplicates",
			files: map[string]string{
				"magic_a.go": "package magic\n\nconst AlgoRSA = \"RSA\"\n",
				"magic_b.go": "package magic\n\nconst AlgorithmRSA = \"RSA\"\n",
			},
		},
		{
			name: "multiple duplicate groups",
			files: map[string]string{
				"magic_multi.go": "package magic\n\nconst (\n\tProtoHTTPS  = \"https\"\n\tSchemeHTTPS = \"https\"\n\tSizeA       = 42\n\tSizeB       = 42\n\tAlgoRSA     = \"RSA\"\n\tAlgoRSA2    = \"RSA\"\n)\n",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var dir string
			if tc.useInvalidDir {
				dir = "/nonexistent/path/that/does/not/exist"
			} else {
				dir = t.TempDir()
				for name, content := range tc.files {
					writeMagicFile(t, dir, name, content)
				}
			}

			logger := cryptoutilCmdCicdCommon.NewLogger("magic-dup-test")
			err := CheckMagicDuplicatesInDir(logger, dir)

			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestCheck_UsesMagicDefaultDir(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	// Check() calls CheckMagicDuplicatesInDir with MagicDefaultDir="internal/shared/magic".
	// When run from the package test directory, that relative path does not exist,
	// so Check() returns an error. This exercises the Check() code path.
	err := Check(logger)
	require.Error(t, err, "Check() should fail when MagicDefaultDir does not exist relative to CWD")
	require.Contains(t, err.Error(), "failed to parse magic package")
}

func TestCheckCrossFileDuplicatesInDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(t *testing.T) (magicDir, rootDir string)
	}{
		{
			name: "no duplicates across files",
			setup: func(t *testing.T) (string, string) {
				t.Helper()
				rootDir := t.TempDir()
				writeGoFile(t, rootDir, "pkg/a", "a.go", "package a\n\nconst AlgoRSA = \"RSA\"\n")
				writeGoFile(t, rootDir, "pkg/b", "b.go", "package b\n\nconst ProtoHTTPS = \"https\"\n")

				return t.TempDir(), rootDir
			},
		},
		{
			name: "finds duplicates logged not errored",
			setup: func(t *testing.T) (string, string) {
				t.Helper()
				rootDir := t.TempDir()
				writeGoFile(t, rootDir, "pkg/a", "a.go", "package a\n\nconst ProtoHTTPS = \"https\"\n")
				writeGoFile(t, rootDir, "pkg/b", "b.go", "package b\n\nconst SchemeHTTPS = \"https\"\n")

				return t.TempDir(), rootDir
			},
		},
		{
			name: "same file not cross duplicate",
			setup: func(t *testing.T) (string, string) {
				t.Helper()
				rootDir := t.TempDir()
				writeGoFile(t, rootDir, "pkg/a", "a.go", "package a\n\nconst (\nProtoHTTPS  = \"https\"\nSchemeHTTPS = \"https\"\n)\n")

				return t.TempDir(), rootDir
			},
		},
		{
			name: "skips magic directory",
			setup: func(t *testing.T) (string, string) {
				t.Helper()
				rootDir := t.TempDir()
				magicDir := filepath.Join(rootDir, "shared", "magic")
				require.NoError(t, os.MkdirAll(magicDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))
				require.NoError(t, os.WriteFile(
					filepath.Join(magicDir, "magic_net.go"),
					[]byte("package magic\n\nconst ProtoHTTPS = \"https\"\n"),
					cryptoutilSharedMagic.CacheFilePermissions,
				))
				writeGoFile(t, rootDir, "pkg/a", "a.go", "package a\n\nconst SchemeHTTPS = \"https\"\n")

				return magicDir, rootDir
			},
		},
		{
			name: "non-string constants",
			setup: func(t *testing.T) (string, string) {
				t.Helper()
				tmpDir := t.TempDir()
				magicDir := filepath.Join(tmpDir, "magic")
				require.NoError(t, os.MkdirAll(magicDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
				writeMagicFile(t, tmpDir, "consts.go", "package root\nconst (\n\tA = 42\n\tB = 3.14\n)\n")

				return magicDir, tmpDir
			},
		},
		{
			name: "unparseable file skipped gracefully",
			setup: func(t *testing.T) (string, string) {
				t.Helper()
				tmpDir := t.TempDir()
				magicDir := filepath.Join(tmpDir, "magic")
				require.NoError(t, os.MkdirAll(magicDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
				writeMagicFile(t, tmpDir, "bad.go", "THIS IS NOT VALID GO CODE @@@@!!")

				return magicDir, tmpDir
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			magicDir, rootDir := tc.setup(t)
			logger := cryptoutilCmdCicdCommon.NewLogger("cross-dup-test")
			err := CheckCrossFileDuplicatesInDir(logger, magicDir, rootDir, filepath.Abs, filepath.Walk)
			require.NoError(t, err)
		})
	}
}

func TestCheckCrossFileDuplicatesInDir_ErrorPaths(t *testing.T) {
	t.Parallel()

	failAbsOnCall := func(n int) func(string) (string, error) {
		callCount := 0

		return func(path string) (string, error) {
			callCount++
			if callCount == n {
				return "", os.ErrInvalid
			}

			return filepath.Abs(path)
		}
	}

	tests := []struct {
		name    string
		dirs    [2]string
		absFn   func(string) (string, error)
		walkFn  func(string, filepath.WalkFunc) error
		wantErr string
	}{
		{
			name:    "abs magic dir error",
			dirs:    [2]string{"/some/magic", "/some/root"},
			absFn:   failAbsOnCall(1),
			walkFn:  filepath.Walk,
			wantErr: "cannot resolve magic dir",
		},
		{
			name:    "abs root dir error",
			dirs:    [2]string{"/some/magic", "/some/root"},
			absFn:   failAbsOnCall(2),
			walkFn:  filepath.Walk,
			wantErr: "cannot resolve root dir",
		},
		{
			name:  "walk function error",
			absFn: filepath.Abs,
			walkFn: func(_ string, _ filepath.WalkFunc) error {
				return os.ErrPermission
			},
			wantErr: "directory walk failed",
		},
		{
			name:  "walk callback error",
			absFn: filepath.Abs,
			walkFn: func(root string, fn filepath.WalkFunc) error {
				_ = fn(filepath.Join(root, "bad.go"), nil, fmt.Errorf("injected walk callback error"))

				return nil
			},
			wantErr: "walk errors",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			magicDir, rootDir := tc.dirs[0], tc.dirs[1]
			if magicDir == "" {
				magicDir = t.TempDir()
			}

			if rootDir == "" {
				rootDir = t.TempDir()
			}

			logger := cryptoutilCmdCicdCommon.NewLogger("cross-dup-test")
			err := CheckCrossFileDuplicatesInDir(logger, magicDir, rootDir, tc.absFn, tc.walkFn)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestCheckCrossFileDuplicatesInDir_WalkFileErr(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	magicDir := filepath.Join(tmpDir, "magic")
	rootDir := tmpDir

	require.NoError(t, os.MkdirAll(magicDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	// Create a sub-dir that will become unreadable, triggering a walk file error.
	subDir := filepath.Join(rootDir, "subpkg")
	require.NoError(t, os.MkdirAll(subDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "constants.go"), []byte("package subpkg\nconst X = \"hello\"\n"), cryptoutilSharedMagic.CacheFilePermissions))
	// Make dir unreadable to trigger a walk error inside the walk callback.
	require.NoError(t, os.Chmod(subDir, 0o000))

	t.Cleanup(func() { _ = os.Chmod(subDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute) })

	logger := cryptoutilCmdCicdCommon.NewLogger("cross-dup-test")
	err := CheckCrossFileDuplicatesInDir(logger, magicDir, rootDir, filepath.Abs, filepath.Walk)
	// On Windows chmod 000 doesn't work the same way; the test is best-effort.
	// If it doesn't produce an error, that's acceptable.
	if err != nil {
		require.Contains(t, err.Error(), "walk errors")
	}
}

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheck_ProjectRoot(t *testing.T) {
	root, err := findProjectRoot()
	if err != nil {
		t.Skip("Skipping - cannot find project root")
	}

	orig, err := os.Getwd()
	require.NoError(t, err)

	t.Cleanup(func() { _ = os.Chdir(orig) })

	require.NoError(t, os.Chdir(root))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger)
	require.NoError(t, err)
}

// findProjectRoot finds the project root by walking up to find go.mod.
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
			return "", fmt.Errorf("go.mod not found")
		}

		dir = parent
	}
}
