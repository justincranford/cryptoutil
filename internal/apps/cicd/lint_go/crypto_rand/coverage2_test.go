// Copyright (c) 2025 Justin Cranford

package crypto_rand

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

// TestCheckMainGoFile_ReadError verifies that lintGoCmdMainPattern.CheckMainGoFile returns error when file cannot be read.

// TestCheckCryptoRandInDir_WithViolation verifies that CheckInDir returns
// an error when a Go file uses math/rand.
func TestCheckCryptoRandInDir_WithViolation(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()
	violationFile := filepath.Join(tmpDir, "badrand.go")

	content := []byte(`package foo

import "math/rand"

func getNum() int {
	return rand.Intn(100)
}
`)

	err := os.WriteFile(violationFile, content, 0o600)
	require.NoError(t, err)

	err = CheckInDir(logger, tmpDir)
	require.Error(t, err, "Should return error when math/rand is used")
	require.Contains(t, err.Error(), "math/rand violations")
}

// TestCheckCryptoRandInDir_Clean verifies that CheckInDir returns nil
// when no math/rand violations are found.
func TestCheckCryptoRandInDir_Clean(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()
	cleanFile := filepath.Join(tmpDir, "clean.go")

	content := []byte("package foo\n")

	err := os.WriteFile(cleanFile, content, 0o600)
	require.NoError(t, err)

	err = CheckInDir(logger, tmpDir)
	require.NoError(t, err, "Should pass when no math/rand is used")
}

// TestFindMathRandViolationsInDir_SkipsTestFiles verifies that test files are skipped.
func TestFindMathRandViolationsInDir_SkipsTestFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "pkg_test.go")
	content := []byte(`package foo

import "math/rand"

func TestHelper() int {
	return rand.Intn(100)
}
`)

	err := os.WriteFile(testFile, content, 0o600)
	require.NoError(t, err)

	violations, err := FindMathRandViolationsInDir(tmpDir)
	require.NoError(t, err)
	require.Empty(t, violations, "Test files should be skipped for math/rand check")
}

// TestFindMathRandViolationsInDir_SkipsNolintFiles verifies that files with nolint are skipped.
func TestFindMathRandViolationsInDir_SkipsNolintFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	nolintFile := filepath.Join(tmpDir, "pkg.go")
	content := []byte(`package foo

import "math/rand" //nolint:gosec

func get() int {
	return rand.Intn(100)
}
`)

	err := os.WriteFile(nolintFile, content, 0o600)
	require.NoError(t, err)

	violations, err := FindMathRandViolationsInDir(tmpDir)
	require.NoError(t, err)
	require.Empty(t, violations, "Files with nolint should be skipped")
}

// TestCheckCryptoRandInDir_SkipsTestHelperDirs verifies that test helper directories
// are excluded from the math/rand check.
func TestCheckCryptoRandInDir_SkipsTestHelperDirs(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	// Place violations only inside test helper directories (should be skipped).
	testingDir := filepath.Join(tmpDir, "testing")
	err := os.MkdirAll(testingDir, 0o755)
	require.NoError(t, err)

	violationFile := filepath.Join(testingDir, "helper.go")
	content := []byte(`package testing

import "math/rand"

func get() int { return rand.Intn(100) }
`)

	err = os.WriteFile(violationFile, content, 0o600)
	require.NoError(t, err)

	err = CheckInDir(logger, tmpDir)
	require.NoError(t, err, "testing/ directory should be excluded from math/rand check")
}

// TestFindMathRandViolationsInDir_WalkDirError verifies that FindMathRandViolationsInDir
// returns error when a subdirectory is inaccessible during the walk.
func TestFindMathRandViolationsInDir_WalkDirError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create an inaccessible subdirectory.
	badDir := filepath.Join(tmpDir, "baddir")
	err := os.MkdirAll(badDir, 0o000)
	require.NoError(t, err)

	t.Cleanup(func() { _ = os.Chmod(badDir, 0o700) })

	violations, err := FindMathRandViolationsInDir(tmpDir)
	require.Error(t, err, "Should error when a subdirectory cannot be accessed")
	require.Nil(t, violations)
}
