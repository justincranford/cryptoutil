// Copyright (c) 2025-2026 Justin Cranford.
package magic_constant_location

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestCheckAndFindViolationsErrorPaths(t *testing.T) {
	t.Parallel()

	require.NoError(t, Check(cryptoutilCmdCicdCommon.NewLogger("magic-constant-location")))

	_, err := FindViolationsInDir(filepath.Join(t.TempDir(), "missing"))
	require.ErrorContains(t, err, "failed to walk directory tree")

	err = CheckInDir(
		cryptoutilCmdCicdCommon.NewLogger("magic-constant-location"),
		filepath.Join(t.TempDir(), "missing"),
	)
	require.ErrorContains(t, err, "failed to check magic-constant-location")
}

func TestSkipDirRules(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	tests := []struct {
		name string
		path string
		want error
	}{
		{name: "magic package", path: filepath.Join(root, "internal", "shared", "magic"), want: filepath.SkipDir},
		{name: "git directory", path: filepath.Join(root, cryptoutilSharedMagic.CICDExcludeDirGit), want: filepath.SkipDir},
		{name: "vendor directory", path: filepath.Join(root, cryptoutilSharedMagic.CICDExcludeDirVendor), want: filepath.SkipDir},
		{name: "testdata directory", path: filepath.Join(root, "pkg", "testdata"), want: filepath.SkipDir},
		{name: "root api directory", path: filepath.Join(root, "api"), want: filepath.SkipDir},
		{name: "regular directory", path: filepath.Join(root, "pkg"), want: nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := skipDir(root, tc.path)
			require.ErrorIs(t, err, tc.want)
		})
	}
}

func TestScanFileForSuspiciousConsts_ErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{name: "integer overflow is ignored", content: "const tooLarge = 999999999999999999999999999999999999999999999999999999999999999\n"},
		{name: "scanner token too long", content: strings.Repeat("x", 70*1024), wantErr: "failed to scan"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "constants.go")
			require.NoError(t, os.WriteFile(path, []byte(tc.content), cryptoutilSharedMagic.CacheFilePermissions))

			violations, err := scanFileForSuspiciousConsts(path, "constants.go")

			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
				require.Empty(t, violations)
			}
		})
	}

	_, err := scanFileForSuspiciousConsts(
		filepath.Join(t.TempDir(), "missing.go"),
		"missing.go",
	)
	require.ErrorContains(t, err, "failed to open")
}

func TestFindViolationsInDir_ErrorPaths(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	filePath := filepath.Join(root, "constants.go")
	require.NoError(t, os.WriteFile(filePath, []byte("package constants\n"), cryptoutilSharedMagic.CacheFilePermissions))
	info, err := os.Stat(filePath)
	require.NoError(t, err)

	_, err = findViolationsInDir(
		root,
		func(_ string, walkFn filepath.WalkFunc) error {
			return walkFn(filePath, info, nil)
		},
		func(string, string) (string, error) {
			return "", errors.New("injected relative path failure")
		},
	)
	require.ErrorContains(t, err, "failed to get relative path")

	largeSourcePath := filepath.Join(root, "large.go")
	require.NoError(t, os.WriteFile(
		largeSourcePath,
		[]byte("package constants\n"+strings.Repeat("x", 70*1024)),
		cryptoutilSharedMagic.CacheFilePermissions,
	))

	_, err = FindViolationsInDir(root)
	require.ErrorContains(t, err, "failed to scan")
}

func TestSkipDirRelativePathError(t *testing.T) {
	t.Parallel()

	err := skipDirWithRel("root", "outside", func(string, string) (string, error) {
		return "", errors.New("injected relative path failure")
	})

	require.ErrorContains(t, err, "failed to compute relative path")
}
