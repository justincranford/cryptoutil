// Copyright (c) 2025 Justin Cranford

package insecure_skip_verify

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

// TestCheckMainGoFile_ReadError verifies that lintGoCmdMainPattern.CheckMainGoFile returns error when file cannot be read.

// TestCheckInsecureSkipVerifyInDir_WithViolation verifies that CheckInDir
// returns error when InsecureSkipVerify: true is found in production code.
func TestCheckInsecureSkipVerifyInDir_WithViolation(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()
	violationFile := filepath.Join(tmpDir, "badtls.go")

	content := []byte(`package foo

import "crypto/tls"

func getClient() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true,
	}
}
`)

	err := os.WriteFile(violationFile, content, 0o600)
	require.NoError(t, err)

	err = CheckInDir(logger, tmpDir)
	require.Error(t, err, "Should return error when InsecureSkipVerify: true is used")
	require.Contains(t, err.Error(), "InsecureSkipVerify violations")
}

// TestCheckInsecureSkipVerifyInDir_Clean verifies that CheckInDir
// returns nil when no InsecureSkipVerify violations are found.
func TestCheckInsecureSkipVerifyInDir_Clean(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()
	cleanFile := filepath.Join(tmpDir, "goodtls.go")

	content := []byte("package foo\n")

	err := os.WriteFile(cleanFile, content, 0o600)
	require.NoError(t, err)

	err = CheckInDir(logger, tmpDir)
	require.NoError(t, err, "Should pass when no InsecureSkipVerify usage found")
}

// TestFindInsecureSkipVerifyViolationsInDir_SkipsTestFiles verifies test files are skipped.
func TestFindInsecureSkipVerifyViolationsInDir_SkipsTestFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "pkg_test.go")
	content := []byte(`package foo

import "crypto/tls"

func TestHelper() *tls.Config {
	return &tls.Config{InsecureSkipVerify: true}
}
`)

	err := os.WriteFile(testFile, content, 0o600)
	require.NoError(t, err)

	violations, err := FindInsecureSkipVerifyViolationsInDir(tmpDir)
	require.NoError(t, err)
	require.Empty(t, violations, "Test files should be skipped for InsecureSkipVerify check")
}

// TestCheckInsecureSkipVerifyInDir_SkipsTestHelperDirs verifies that test helper
// directories are excluded from the InsecureSkipVerify check.
func TestCheckInsecureSkipVerifyInDir_SkipsTestHelperDirs(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	testingDir := filepath.Join(tmpDir, "testing")
	err := os.MkdirAll(testingDir, 0o755)
	require.NoError(t, err)

	violationFile := filepath.Join(testingDir, "helper.go")
	content := []byte(`package testing

import "crypto/tls"

func Helper() *tls.Config { return &tls.Config{InsecureSkipVerify: true} }
`)

	err = os.WriteFile(violationFile, content, 0o600)
	require.NoError(t, err)

	err = CheckInDir(logger, tmpDir)
	require.NoError(t, err, "testing/ directory should be excluded")
}

// TestFindInsecureSkipVerifyViolationsInDir_WalkDirError verifies error on inaccessible dir.
func TestFindInsecureSkipVerifyViolationsInDir_WalkDirError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	badDir := filepath.Join(tmpDir, "baddir")
	err := os.MkdirAll(badDir, 0o000)
	require.NoError(t, err)

	t.Cleanup(func() { _ = os.Chmod(badDir, 0o700) })

	violations, err := FindInsecureSkipVerifyViolationsInDir(tmpDir)
	require.Error(t, err, "Should error when a subdirectory cannot be accessed")
	require.Nil(t, violations)
}
