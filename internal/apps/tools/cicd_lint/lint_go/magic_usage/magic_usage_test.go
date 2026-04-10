// Copyright (c) 2025 Justin Cranford

package magic_usage

import (
	"os"
	"path/filepath"
	"runtime"
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

// setupMagicUsageDirs creates a magic dir and a separate root dir for usage tests.
func setupMagicUsageDirs(t *testing.T) (magicDir, rootDir string) {
	t.Helper()

	magicDir = t.TempDir()
	rootDir = t.TempDir()

	return magicDir, rootDir
}

func TestCheckMagicUsageInDir_NoErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setupFn func(t *testing.T) (string, string)
	}{
		{
			name: "no violations",
			setupFn: func(t *testing.T) (string, string) {
				t.Helper()
				magicDir, rootDir := setupMagicUsageDirs(t)
				writeMagicFile(t, magicDir, "magic.go", "package magic\n\nconst ProtocolHTTPS = \"https\"\n")
				require.NoError(t, os.WriteFile(filepath.Join(rootDir, "handler.go"), []byte("package app\n\nimport \"fmt\"\n\nfunc greet() { fmt.Println(\"hello\") }\n"), cryptoutilSharedMagic.CacheFilePermissions))

				return magicDir, rootDir
			},
		},
		{
			name: "trivial string not flagged",
			setupFn: func(t *testing.T) (string, string) {
				t.Helper()
				magicDir, rootDir := setupMagicUsageDirs(t)
				writeMagicFile(t, magicDir, "magic.go", "package magic\n\nconst EmptyString = \"\"\nconst Dot = \".\"\n")
				require.NoError(t, os.WriteFile(filepath.Join(rootDir, "app.go"), []byte("package app\n\nfunc f() string { return \".\" }\n"), cryptoutilSharedMagic.CacheFilePermissions))

				return magicDir, rootDir
			},
		},
		{
			name: "trivial int not flagged",
			setupFn: func(t *testing.T) (string, string) {
				t.Helper()
				magicDir, rootDir := setupMagicUsageDirs(t)
				writeMagicFile(t, magicDir, "magic.go", "package magic\n\nconst Zero = 0\nconst One  = 1\n")
				require.NoError(t, os.WriteFile(filepath.Join(rootDir, "app.go"), []byte("package app\n\nfunc count() int { return 0 }\n"), cryptoutilSharedMagic.CacheFilePermissions))

				return magicDir, rootDir
			},
		},
		{
			name: "test const only matches test file",
			setupFn: func(t *testing.T) (string, string) {
				t.Helper()
				magicDir, rootDir := setupMagicUsageDirs(t)
				writeMagicFile(t, magicDir, "magic_testing.go", "package magic\n\nconst TestRateLimit = 500\n")
				require.NoError(t, os.WriteFile(filepath.Join(rootDir, "server.go"), []byte("package app\n\nconst localLimit = 500\n"), cryptoutilSharedMagic.CacheFilePermissions))
				require.NoError(t, os.WriteFile(filepath.Join(rootDir, "server_test.go"), []byte("package app\n\nconst wantLimit = 500\n"), cryptoutilSharedMagic.CacheFilePermissions))

				return magicDir, rootDir
			},
		},
		{
			name: "empty magic package",
			setupFn: func(t *testing.T) (string, string) {
				t.Helper()
				magicDir, rootDir := setupMagicUsageDirs(t)
				writeMagicFile(t, magicDir, "magic.go", "package magic\n")
				require.NoError(t, os.WriteFile(filepath.Join(rootDir, "app.go"), []byte("package app\n\nconst x = \"hello\"\n"), cryptoutilSharedMagic.CacheFilePermissions))

				return magicDir, rootDir
			},
		},
		{
			name: "generated file skipped",
			setupFn: func(t *testing.T) (string, string) {
				t.Helper()
				magicDir, rootDir := setupMagicUsageDirs(t)
				writeMagicFile(t, magicDir, "magic.go", "package magic\n\nconst ProtocolHTTPS = \"https\"\n")
				require.NoError(t, os.WriteFile(filepath.Join(rootDir, "openapi.gen.go"), []byte("package app\n\nfunc genFunc() string { return \"https\" }\n"), cryptoutilSharedMagic.CacheFilePermissions))

				return magicDir, rootDir
			},
		},
		{
			name: "magic dir inside root",
			setupFn: func(t *testing.T) (string, string) {
				t.Helper()
				rootDir := t.TempDir()
				magicDir := filepath.Join(rootDir, "magic")
				require.NoError(t, os.MkdirAll(magicDir, 0o700))
				writeMagicFile(t, magicDir, "magic.go", "package magic\n\nconst Timeout = 30\n")
				require.NoError(t, os.WriteFile(filepath.Join(rootDir, "app.go"), []byte("package app\n\nfunc f() {}\n"), cryptoutilSharedMagic.CacheFilePermissions))

				return magicDir, rootDir
			},
		},
		{
			name: "vendor dir skipped",
			setupFn: func(t *testing.T) (string, string) {
				t.Helper()
				magicDir, rootDir := setupMagicUsageDirs(t)
				writeMagicFile(t, magicDir, "magic.go", "package magic\n\nconst Timeout = 30\n")

				vendorDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirVendor, "somepkg")
				require.NoError(t, os.MkdirAll(vendorDir, 0o700))
				require.NoError(t, os.WriteFile(filepath.Join(vendorDir, "pkg.go"), []byte("package somepkg\n\nconst x = 30\n"), cryptoutilSharedMagic.CacheFilePermissions))

				return magicDir, rootDir
			},
		},
		{
			name: "unparseable go file skipped",
			setupFn: func(t *testing.T) (string, string) {
				t.Helper()
				magicDir, rootDir := setupMagicUsageDirs(t)
				writeMagicFile(t, magicDir, "magic.go", "package magic\n\nconst Timeout = 30\n")
				require.NoError(t, os.WriteFile(filepath.Join(rootDir, "broken.go"), []byte("package INVALID {{{"), cryptoutilSharedMagic.CacheFilePermissions))

				return magicDir, rootDir
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			magicDir, rootDir := tc.setupFn(t)
			logger := cryptoutilCmdCicdCommon.NewLogger("magic-usage-test")

			err := CheckMagicUsageInDir(logger, magicDir, rootDir, filepath.Abs, filepath.Walk)

			require.NoError(t, err)
		})
	}
}

func TestCheckMagicUsageInDir_Violations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupFn     func(t *testing.T) (string, string)
		wantContain string
	}{
		{
			name: "literal violation",
			setupFn: func(t *testing.T) (string, string) {
				t.Helper()
				magicDir, rootDir := setupMagicUsageDirs(t)
				writeMagicFile(t, magicDir, "magic.go", "package magic\n\nconst ProtocolHTTPS = \"https\"\n")
				require.NoError(t, os.WriteFile(filepath.Join(rootDir, "client.go"), []byte("package app\n\nfunc scheme() string { return \"https\" }\n"), cryptoutilSharedMagic.CacheFilePermissions))

				return magicDir, rootDir
			},
			wantContain: "literal-use",
		},
		{
			name: "const redefine",
			setupFn: func(t *testing.T) (string, string) {
				t.Helper()
				magicDir, rootDir := setupMagicUsageDirs(t)
				writeMagicFile(t, magicDir, "magic.go", "package magic\n\nconst ProtocolHTTPS = \"https\"\n")
				require.NoError(t, os.WriteFile(filepath.Join(rootDir, "localconst.go"), []byte("package app\n\nconst localHTTPS = \"https\"\n"), cryptoutilSharedMagic.CacheFilePermissions))

				return magicDir, rootDir
			},
			wantContain: "const-redefine-string",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			magicDir, rootDir := tc.setupFn(t)
			logger := cryptoutilCmdCicdCommon.NewLogger("magic-usage-test")

			err := CheckMagicUsageInDir(logger, magicDir, rootDir, filepath.Abs, filepath.Walk)

			require.Error(t, err)
			require.ErrorContains(t, err, tc.wantContain)
		})
	}
}

func TestCheckMagicUsageInDir_PathErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupFn     func(t *testing.T) (string, string)
		wantContain string
	}{
		{
			name: "invalid magic dir",
			setupFn: func(_ *testing.T) (string, string) {
				return "/nonexistent/magic", "."
			},
			wantContain: "failed to parse magic package",
		},
		{
			name: "nonexistent root dir",
			setupFn: func(t *testing.T) (string, string) {
				t.Helper()
				magicDir := t.TempDir()
				writeMagicFile(t, magicDir, "magic.go", "package magic\n\nconst Timeout = 30\n")

				return magicDir, "/nonexistent/root/dir"
			},
			wantContain: "walk errors",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			magicDir, rootDir := tc.setupFn(t)
			logger := cryptoutilCmdCicdCommon.NewLogger("magic-usage-test")

			err := CheckMagicUsageInDir(logger, magicDir, rootDir, filepath.Abs, filepath.Walk)

			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantContain)
		})
	}
}

func TestCheck_UsesMagicDefaultDir(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	// Check() calls CheckMagicUsageInDir with MagicDefaultDir="internal/shared/magic".
	// When run from the package test directory, that relative path does not exist,
	// so Check() returns an error. This exercises the Check() code path.
	err := Check(logger)
	require.Error(t, err, "Check() should fail when MagicDefaultDir does not exist relative to CWD")
	require.Contains(t, err.Error(), "failed to parse magic package")
}

func TestCheckMagicUsageInDir_WalkError(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == cryptoutilSharedMagic.OSNameWindows {
		t.Skip("os.Chmod 0o000 does not restrict access on Windows NTFS")
	}

	magicDir, rootDir := setupMagicUsageDirs(t)
	writeMagicFile(t, magicDir, "magic.go", "package magic\n\nconst Timeout = 30\n")

	// Create an unreadable subdirectory to trigger walk error accumulation.
	badSubDir := filepath.Join(rootDir, "locked")
	require.NoError(t, os.MkdirAll(badSubDir, 0o700))
	require.NoError(t, os.Chmod(badSubDir, 0o000))
	t.Cleanup(func() { _ = os.Chmod(badSubDir, 0o700) })

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckMagicUsageInDir(logger, magicDir, rootDir, filepath.Abs, filepath.Walk)
	require.Error(t, err, "Walk errors should be returned")
	require.Contains(t, err.Error(), "walk errors")
}

// Sequential: uses os.Chdir (global process state).
func TestCheckMagicUsageInDir_AbsErrorDeletedCWD(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes and deletes CWD.
	if runtime.GOOS == cryptoutilSharedMagic.OSNameWindows {
		t.Skip("deleting CWD not supported on Windows")
	}

	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { require.NoError(t, os.Chdir(origDir)) }()

	// Create a magic directory with real constants (absolute path for ParseMagicDir).
	magicDir := t.TempDir()
	writeMagicFile(t, magicDir, "magic.go", "package magic\n\nconst Timeout = 30\n")

	// Create a temporary directory, chdir into it, then delete it to break Getwd().
	lostDir, err := os.MkdirTemp("", "lost-cwd-*")
	require.NoError(t, err)
	require.NoError(t, os.Chdir(lostDir))
	require.NoError(t, os.RemoveAll(lostDir))

	// Now filepath.Abs on any relative path will fail because Getwd() fails.
	// Pass a relative rootDir to trigger the Abs error path.
	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = CheckMagicUsageInDir(logger, magicDir, "relative/root", filepath.Abs, filepath.Walk)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot resolve")
}
