// Copyright (c) 2025 Justin Cranford

package tls_minimum_version

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"

	"github.com/stretchr/testify/require"
)

func newTestLogger() *cryptoutilCmdCicdCommon.Logger {
	return cryptoutilCmdCicdCommon.NewLogger("test")
}

func writeGoFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, cryptoutilSharedMagic.DirPermissions))
	p := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(p, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	return p
}

const cleanGoContent = `package foo

import "crypto/tls"

func newConfig() *tls.Config {
return &tls.Config{MinVersion: tls.VersionTLS13}
}
`

const tls12GoContent = `package foo

import "crypto/tls"

func newConfig() *tls.Config {
return &tls.Config{MinVersion: tls.VersionTLS12}
}
`

func TestCheckInDir_CleanFile_Passes(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	writeGoFile(t, tmp, "server.go", cleanGoContent)
	err := CheckInDir(newTestLogger(), tmp, filepath.Walk, os.Open)
	require.NoError(t, err)
}

func TestCheckInDir_TLS12Production_Fails(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	writeGoFile(t, tmp, "server.go", tls12GoContent)
	err := CheckInDir(newTestLogger(), tmp, filepath.Walk, os.Open)
	require.Error(t, err)
	require.Contains(t, err.Error(), "TLS minimum version violation")
}

func TestCheckInDir_TLS12InTestFile_Skipped(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	// _test.go files are excluded from TLS check.
	writeGoFile(t, tmp, "server_test.go", tls12GoContent)
	err := CheckInDir(newTestLogger(), tmp, filepath.Walk, os.Open)
	require.NoError(t, err)
}

func TestCheckInDir_TLS12InArchivedDir_Skipped(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	archivedDir := filepath.Join(tmp, "archived")
	writeGoFile(t, archivedDir, "old.go", tls12GoContent)

	err := CheckInDir(newTestLogger(), tmp, filepath.Walk, os.Open)
	require.NoError(t, err)
}

func TestCheckInDir_CommentLine_Skipped(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	writeGoFile(t, tmp, "server.go", `package foo

// This is an old example: MinVersion: tls.VersionTLS12 should not be used.
func Good() {}
`)
	err := CheckInDir(newTestLogger(), tmp, filepath.Walk, os.Open)
	require.NoError(t, err)
}

func TestCheckInDir_MultipleViolations_AllReported(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	writeGoFile(t, tmp, "server.go", `package foo

import "crypto/tls"

func c1() *tls.Config { return &tls.Config{MinVersion: tls.VersionTLS12} }
func c2() *tls.Config { return &tls.Config{MinVersion: tls.VersionTLS12} }
`)
	err := CheckInDir(newTestLogger(), tmp, filepath.Walk, os.Open)
	require.Error(t, err)
}

func TestCheckInDir_NonGoFile_Ignored(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	p := filepath.Join(tmp, "config.yaml")
	require.NoError(t, os.WriteFile(p, []byte("MinVersion: tls.VersionTLS12\n"), cryptoutilSharedMagic.CacheFilePermissions))

	err := CheckInDir(newTestLogger(), tmp, filepath.Walk, os.Open)
	require.NoError(t, err)
}

func TestScanFileForTLSVersion_NoViolations(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	p := writeGoFile(t, tmp, "clean.go", cleanGoContent)
	violations, err := scanFileForTLSVersion(p, tmp, os.Open)
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestScanFileForTLSVersion_WithViolation(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	p := writeGoFile(t, tmp, "bad.go", tls12GoContent)
	violations, err := scanFileForTLSVersion(p, tmp, os.Open)
	require.NoError(t, err)
	require.Len(t, violations, 1)
	require.Contains(t, violations[0], "TLS MinVersion below TLS 1.3")
}

func TestScanFileForTLSVersion_NonexistentFile_Error(t *testing.T) {
	t.Parallel()

	_, err := scanFileForTLSVersion("/nonexistent/file.go", "/tmp", os.Open)
	require.Error(t, err)
}

func TestCheckInDir_VendorDir_Skipped(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	vendorDir := filepath.Join(tmp, cryptoutilSharedMagic.CICDExcludeDirVendor)
	writeGoFile(t, vendorDir, "dep.go", tls12GoContent)

	err := CheckInDir(newTestLogger(), tmp, filepath.Walk, os.Open)
	require.NoError(t, err)
}

func TestCheckInDir_WalkError(t *testing.T) {
	t.Parallel()

	err := CheckInDir(
		newTestLogger(),
		t.TempDir(),
		func(_ string, _ filepath.WalkFunc) error { return fmt.Errorf("injected walk error") },
		os.Open,
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "filesystem walk failed")
}

func TestCheckInDir_WalkCallbackError(t *testing.T) {
	t.Parallel()

	err := CheckInDir(
		newTestLogger(),
		t.TempDir(),
		func(_ string, fn filepath.WalkFunc) error {
			return fn("bad/path", nil, fmt.Errorf("injected callback error"))
		},
		os.Open,
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "filesystem walk failed")
}

func TestCheckInDir_ScanOpenError(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	writeGoFile(t, tmp, "main.go", "package main\n")

	err := CheckInDir(
		newTestLogger(),
		tmp,
		filepath.Walk,
		func(_ string) (*os.File, error) { return nil, fmt.Errorf("injected open error") },
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "filesystem walk failed")
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

	logger := cryptoutilCmdCicdCommon.NewLogger("test-tls-minimum-version")

	err = Check(logger)
	require.NoError(t, err)
}
