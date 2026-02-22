package insecure_skip_verify

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	lintGoCommon "cryptoutil/internal/apps/cicd/lint_go/common"
)

func TestCheckInsecureSkipVerify_Clean(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	cleanFile := filepath.Join(tempDir, "clean.go")
	content := `package main

import (
	"crypto/tls"
)

func main() {
	config := &tls.Config{
		MinVersion: tls.VersionTLS13,
	}
	println(config)
}
`

	err := os.WriteFile(cleanFile, []byte(content), 0o600)
	require.NoError(t, err)

	violations, err := CheckFileForInsecureSkipVerify(cleanFile)
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestCheckInsecureSkipVerify_Violation(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	badFile := filepath.Join(tempDir, "bad.go")
	// Use concatenation to avoid triggering the linter on this test file.
	content := "package main\n\nimport (\n\t\"crypto/tls\"\n)\n\nfunc main() {\n\tconfig := &tls.Config{\n\t\t" + "Insecure" + "SkipVerify: true,\n\t}\n\tprintln(config)\n}\n"

	err := os.WriteFile(badFile, []byte(content), 0o600)
	require.NoError(t, err)

	violations, err := CheckFileForInsecureSkipVerify(badFile)
	require.NoError(t, err)
	require.Len(t, violations, 1)
	require.Contains(t, violations[0].Issue, "disables TLS certificate verification")
}

func TestCheckInsecureSkipVerify_WithNolint(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	cleanFile := filepath.Join(tempDir, "clean.go")
	// Use concatenation to avoid triggering the linter on this test file.
	content := "package main\n\nimport (\n\t\"crypto/tls\"\n)\n\nfunc main() {\n\tconfig := &tls.Config{\n\t\t" + "Insecure" + "SkipVerify: true, //nolint:all\n\t}\n\tprintln(config)\n}\n"

	err := os.WriteFile(cleanFile, []byte(content), 0o600)
	require.NoError(t, err)

	violations, err := CheckFileForInsecureSkipVerify(cleanFile)
	require.NoError(t, err)
	require.Empty(t, violations) // Should be skipped due to nolint.
}

func TestCheckInsecureSkipVerify_FileNotFound(t *testing.T) {
	t.Parallel()

	_, err := CheckFileForInsecureSkipVerify("/nonexistent/path.go")
	require.Error(t, err)
}

func TestPrintInsecureSkipVerifyViolations(t *testing.T) {
	t.Parallel()

	violations := []lintGoCommon.CryptoViolation{
		{File: "file2.go", Line: 5, Content: "TLS config", Issue: "disables TLS certificate verification"},
	}

	// Just verify the print function does not panic.
	lintGoCommon.PrintCryptoViolations("InsecureSkipVerify", violations)
}

// TestCheckInsecureSkipVerifyInDir_WalkError verifies that CheckInDir
// returns error when FindInsecureSkipVerifyViolationsInDir returns a walk error.
func TestCheckInsecureSkipVerifyInDir_WalkError(t *testing.T) {
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
	require.Contains(t, err.Error(), "failed to check InsecureSkipVerify")
}

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

func TestFindInsecureSkipVerifyViolationsInDir_VendorDirSkipped(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	vendorDir := filepath.Join(tmpDir, "vendor")
	require.NoError(t, os.MkdirAll(vendorDir, 0o700))

	vendorFile := filepath.Join(vendorDir, "bad.go")
	content := []byte("package vendor\n\nfunc bad() bool { return true } // InsecureSkipVerify: true\n")
	require.NoError(t, os.WriteFile(vendorFile, content, 0o600))

	violations, err := FindInsecureSkipVerifyViolationsInDir(tmpDir)
	require.NoError(t, err)
	require.Empty(t, violations, "vendor/ directory should be skipped")
}

func TestFindInsecureSkipVerifyViolationsInDir_NonGoFileSkipped(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	txtFile := filepath.Join(tmpDir, "config.txt")
	require.NoError(t, os.WriteFile(txtFile, []byte("InsecureSkipVerify: true\n"), 0o600))

	violations, err := FindInsecureSkipVerifyViolationsInDir(tmpDir)
	require.NoError(t, err)
	require.Empty(t, violations, "non-.go file should be skipped")
}

func TestFindInsecureSkipVerifyViolationsInDir_NonExistentRoot(t *testing.T) {
	t.Parallel()

	_, err := FindInsecureSkipVerifyViolationsInDir("/nonexistent/path/that/does/not/exist")
	require.Error(t, err, "Non-existent root should return an error")
}

func TestCheckFileForInsecureSkipVerify_ScannerError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create a Go file with a line longer than bufio.MaxScanTokenSize (64KB) to trigger scanner.Err().
	longLine := "// " + strings.Repeat("x", 70000) + "\n"
	goFile := filepath.Join(tempDir, "main.go")
	require.NoError(t, os.WriteFile(goFile, []byte("package main\n"+longLine), 0o600))

	_, err := CheckFileForInsecureSkipVerify(goFile)
	require.Error(t, err)
	require.Contains(t, err.Error(), "error reading file")
}

func TestFindInsecureSkipVerifyViolationsInDir_CheckFileError(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("chmod 0o000 does not work on Windows")
	}

	tempDir := t.TempDir()

	// Create a Go file that is unreadable, so CheckFileForInsecureSkipVerify returns error during Walk.
	goFile := filepath.Join(tempDir, "main.go")
	require.NoError(t, os.WriteFile(goFile, []byte("package main\n"), 0o600))
	require.NoError(t, os.Chmod(goFile, 0o000))

	defer func() { _ = os.Chmod(goFile, 0o600) }()

	_, err := FindInsecureSkipVerifyViolationsInDir(tempDir)
	require.Error(t, err)
}
