package crypto_rand

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	lintGoCommon "cryptoutil/internal/apps/cicd/lint_go/common"
	"github.com/stretchr/testify/require"
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

func TestCheck_DelegatesCheckInDir(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory.
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { require.NoError(t, os.Chdir(origDir)) }()

	tmpDir := t.TempDir()
	require.NoError(t, os.Chdir(tmpDir))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	// Check() delegates to CheckInDir(logger, ".").
	// From a clean temp directory with no Go files, there are no violations.
	err = Check(logger)
	require.NoError(t, err)
}

func TestFindMathRandViolationsInDir_VendorDirSkipped(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	vendorDir := filepath.Join(tmpDir, "vendor")
	require.NoError(t, os.MkdirAll(vendorDir, 0o700))

	vendorFile := filepath.Join(vendorDir, "uses_rand.go")
	content := []byte("package vendor\nimport \"math/rand\"\nvar x = rand.Intn(10)\n")
	require.NoError(t, os.WriteFile(vendorFile, content, 0o600))

	violations, err := FindMathRandViolationsInDir(tmpDir)
	require.NoError(t, err)
	require.Empty(t, violations, "vendor/ should be skipped")
}

func TestFindMathRandViolationsInDir_NonGoFileSkipped(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	txtFile := filepath.Join(tmpDir, "notes.txt")
	require.NoError(t, os.WriteFile(txtFile, []byte("math/rand usage notes\n"), 0o600))

	violations, err := FindMathRandViolationsInDir(tmpDir)
	require.NoError(t, err)
	require.Empty(t, violations, "non-.go file should be skipped")
}

func TestFindMathRandViolationsInDir_NonExistentRoot(t *testing.T) {
	t.Parallel()

	_, err := FindMathRandViolationsInDir("/nonexistent/path/that/does/not/exist")
	require.Error(t, err, "Non-existent root should return an error")
}

func TestCheckFileForMathRand_CrandAlias(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "crypto.go")
	content := []byte("package foo\n\nimport (\n\tcrand \"math/rand\"\n)\n\nfunc seed() { crand.Seed(42) }\n")
	require.NoError(t, os.WriteFile(goFile, content, 0o600))

	violations, err := CheckFileForMathRand(goFile)
	require.NoError(t, err)
	require.Empty(t, violations, "crand alias should be accepted")
}

func TestFindMathRandViolationsInDir_FileReadError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "impl.go")
	content := []byte("package foo\n\nimport \"math/rand\"\n\nvar x = rand.Intn(10)\n")
	require.NoError(t, os.WriteFile(goFile, content, 0o600))
	// Make file unreadable so CheckFileForMathRand fails to open it.
	require.NoError(t, os.Chmod(goFile, 0o000))
	t.Cleanup(func() { _ = os.Chmod(goFile, 0o600) })

	_, err := FindMathRandViolationsInDir(tmpDir)
	require.Error(t, err, "Should return error when .go file cannot be opened")
}

func TestCheckFileForMathRand_UsageNolintSkipped(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "rand_usage.go")
	// Import "math/rand" without any nolint on import line (adds import violation).
	// Usage line has //nolint:revive (not gosec/math/rand) so usage continue fires.
	content := []byte("package foo\n\nimport \"math/rand\"\n\nvar x = rand.Intn(10) //nolint:revive\n")
	require.NoError(t, os.WriteFile(goFile, content, 0o600))

	violations, err := CheckFileForMathRand(goFile)
	require.NoError(t, err)
	// Should include violation for the import line but no usage violation.
	require.Len(t, violations, 1, "Only import line should be flagged, not nolint usage line")
	require.Contains(t, violations[0].Issue, "imports math/rand")
}

func TestCheckFileForMathRand_ScannerError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "huge_line.go")

	// Create a file with a line exceeding bufio.MaxScanTokenSize (64KB) to trigger scanner.Err().
	longLine := "package foo\n// " + strings.Repeat("x", 70000) + "\n"
	require.NoError(t, os.WriteFile(goFile, []byte(longLine), 0o600))

	_, err := CheckFileForMathRand(goFile)
	require.Error(t, err)
	require.Contains(t, err.Error(), "error reading file")
}
