package crypto_rand

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	lintGoCommon "cryptoutil/internal/apps/cicd/lint_go/common"
)

func TestCheckCryptoRand_Clean(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	cleanFile := filepath.Join(tempDir, "clean.go")
	content := `package main

import (
	"crypto/rand"
)

func main() {
	buf := make([]byte, 32)
	rand.Read(buf)
}
`

	err := os.WriteFile(cleanFile, []byte(content), 0o600)
	require.NoError(t, err)

	violations, err := CheckFileForMathRand(cleanFile)
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestCheckCryptoRand_MathRandImport(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	badFile := filepath.Join(tempDir, "bad.go")
	content := `package main

import (
	"math/rand"
)

func main() {
	x := rand.Intn(100)
	println(x)
}
`

	err := os.WriteFile(badFile, []byte(content), 0o600)
	require.NoError(t, err)

	violations, err := CheckFileForMathRand(badFile)
	require.NoError(t, err)
	require.Len(t, violations, 2)
	require.Contains(t, violations[0].Issue, "imports math/rand")
	require.Contains(t, violations[1].Issue, "uses math/rand function")
}

func TestCheckCryptoRand_FileNotFound(t *testing.T) {
	t.Parallel()

	_, err := CheckFileForMathRand("/nonexistent/path.go")
	require.Error(t, err)
}

func TestPrintMathRandViolations(t *testing.T) {
	t.Parallel()

	violations := []lintGoCommon.CryptoViolation{
		{File: "file1.go", Line: 10, Content: "import math/rand", Issue: "imports math/rand instead of crypto/rand"},
		{File: "file1.go", Line: 20, Content: "rand.Float64()", Issue: "uses math/rand function"},
	}

	// Just verify the print function does not panic.
	lintGoCommon.PrintCryptoViolations("math/rand", violations)
}

// TestCheckCryptoRandInDir_WalkError verifies that lintGoCryptoRand.CheckInDir
// returns error when lintGoCryptoRand.FindMathRandViolationsInDir returns a walk error.

// TestCheckCryptoRandInDir_WalkError verifies that CheckInDir
// returns error when FindMathRandViolationsInDir returns a walk error.
func TestCheckCryptoRandInDir_WalkError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	// Create an inaccessible subdirectory to trigger walk error.
	badDir := filepath.Join(tmpDir, "baddir")
	err := os.MkdirAll(badDir, 0o000)
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chmod(badDir, 0o700) })

	err = CheckInDir(logger, tmpDir)
	require.Error(t, err, "Should return error when walk fails")
	require.Contains(t, err.Error(), "failed to check math/rand usage")
}

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
