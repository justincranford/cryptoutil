// Copyright (c) 2025 Justin Cranford

package non_fips_algorithms

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"


	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestCheckNonFIPS(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-nonfips")

	// Should pass (no violations in actual codebase after Phase 1).
	err := Check(logger)
	require.NoError(t, err, "Non-FIPS check should pass after bcrypt removal")
}

func TestCheckFileForNonFIPS_Clean(t *testing.T) {
	t.Parallel()

	// Create temp file with FIPS-approved code.
	tmpDir := t.TempDir()
	cleanFile := filepath.Join(tmpDir, "clean.go")

	cleanCode := `package main

import (
	"crypto/sha256"
	"golang.org/x/crypto/pbkdf2"
)

func main() {
	// FIPS-approved: SHA-256
	hash := sha256.Sum256([]byte("data"))

	// FIPS-approved: PBKDF2
	key := pbkdf2.Key([]byte("password"), []byte("salt"), 600000, 32, sha256.New)
}
`

	err := os.WriteFile(cleanFile, []byte(cleanCode), 0o600)
	require.NoError(t, err)

	// Check file.
	violations := CheckFileForNonFIPS(cleanFile)
	require.Empty(t, violations, "FIPS-approved code should have 0 violations")
}

func TestCheckFileForNonFIPS_Bcrypt(t *testing.T) {
	t.Parallel()

	// Create temp file with bcrypt (banned).
	tmpDir := t.TempDir()
	bannedFile := filepath.Join(tmpDir, "banned.go")

	bannedCode := `package main

import (
	"golang.org/x/crypto/bcrypt"
)

func main() {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
}
`

	err := os.WriteFile(bannedFile, []byte(bannedCode), 0o600)
	require.NoError(t, err)

	// Check file.
	violations := CheckFileForNonFIPS(bannedFile)
	require.NotEmpty(t, violations, "Bcrypt usage should be detected")
	require.Contains(t, strings.Join(violations, "\n"), "bcrypt", "Violation should mention bcrypt")
	require.Contains(t, strings.Join(violations, "\n"), "PBKDF2", "Violation should suggest PBKDF2")
}

func TestCheckFileForNonFIPS_MD5(t *testing.T) {
	t.Parallel()

	// Create temp file with MD5 (banned).
	tmpDir := t.TempDir()
	bannedFile := filepath.Join(tmpDir, "md5.go")

	bannedCode := `package main

import (
	"crypto/md5"
)

func main() {
	hash := md5.New()
	sum := md5.Sum([]byte("data"))
}
`

	err := os.WriteFile(bannedFile, []byte(bannedCode), 0o600)
	require.NoError(t, err)

	// Check file.
	violations := CheckFileForNonFIPS(bannedFile)
	require.NotEmpty(t, violations, "MD5 usage should be detected")
	require.Contains(t, strings.Join(violations, "\n"), "md5", "Violation should mention md5")
	require.Contains(t, strings.Join(violations, "\n"), "SHA-256", "Violation should suggest SHA-256")
}

func TestCheckFileForNonFIPS_SHA1(t *testing.T) {
	t.Parallel()

	// Create temp file with SHA-1 (banned).
	tmpDir := t.TempDir()
	bannedFile := filepath.Join(tmpDir, "sha1.go")

	bannedCode := `package main

import (
	"crypto/sha1"
)

func main() {
	hash := sha1.New()
	sum := sha1.Sum([]byte("data"))
}
`

	err := os.WriteFile(bannedFile, []byte(bannedCode), 0o600)
	require.NoError(t, err)

	// Check file.
	violations := CheckFileForNonFIPS(bannedFile)
	require.NotEmpty(t, violations, "SHA-1 usage should be detected")
	require.Contains(t, strings.Join(violations, "\n"), "sha1", "Violation should mention sha1")
	require.Contains(t, strings.Join(violations, "\n"), "SHA-256", "Violation should suggest SHA-256")
}

func TestCheckFileForNonFIPS_DES(t *testing.T) {
	t.Parallel()

	// Create temp file with DES (banned).
	tmpDir := t.TempDir()
	bannedFile := filepath.Join(tmpDir, "des.go")

	bannedCode := `package main

import (
	"crypto/des"
)

func main() {
	cipher, _ := des.NewCipher([]byte("12345678"))
}
`

	err := os.WriteFile(bannedFile, []byte(bannedCode), 0o600)
	require.NoError(t, err)

	// Check file.
	violations := CheckFileForNonFIPS(bannedFile)
	require.NotEmpty(t, violations, "DES usage should be detected")
	require.Contains(t, strings.Join(violations, "\n"), "des", "Violation should mention des")
	require.Contains(t, strings.Join(violations, "\n"), "AES", "Violation should suggest AES")
}

func TestCheckFileForNonFIPS_MultipleViolations(t *testing.T) {
	t.Parallel()

	// Create temp file with multiple violations.
	tmpDir := t.TempDir()
	bannedFile := filepath.Join(tmpDir, "multiple.go")

	bannedCode := `package main

import (
	"crypto/md5"
	"crypto/sha1"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	md5Hash := md5.New()
	sha1Hash := sha1.New()
	bcryptHash, _ := bcrypt.GenerateFromPassword([]byte("password"), 10)
}
`

	err := os.WriteFile(bannedFile, []byte(bannedCode), 0o600)
	require.NoError(t, err)

	// Check file.
	violations := CheckFileForNonFIPS(bannedFile)
	require.NotEmpty(t, violations, "Multiple violations should be detected")

	violationsStr := strings.Join(violations, "\n")
	require.Contains(t, violationsStr, "md5", "Should detect md5")
	require.Contains(t, violationsStr, "sha1", "Should detect sha1")
	require.Contains(t, violationsStr, "bcrypt", "Should detect bcrypt")
}

func TestPrintNonFIPSViolations(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test redirects os.Stderr which is global.
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	violations := map[string][]string{
		"file1.go": {"Line 5: Found 'bcrypt' (non-FIPS) - use PBKDF2-HMAC-SHA256 instead"},
		"file2.go": {"Line 10: Found 'md5.New' (non-FIPS) - use SHA-256/384/512 instead"},
	}

	PrintNonFIPSViolations(violations)

	_ = w.Close()

	os.Stderr = oldStderr

	var buf [8192]byte

	n, _ := r.Read(buf[:])
	output := string(buf[:n])

	require.Contains(t, output, "non-FIPS algorithm violations")
	require.Contains(t, output, "file1.go")
	require.Contains(t, output, "file2.go")
	require.Contains(t, output, "bcrypt")
	require.Contains(t, output, "md5")
	require.Contains(t, output, "FIPS 140-3")
}

// Test constants for repeated string literals.
const (

	testPackageMainDef = "package main\n"
)

func TestFindGoFiles_WithTempDir(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory.

	// Save current directory.
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	// Create temp directory.
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))

	// Create Go files.
	require.NoError(t, os.WriteFile("main.go", []byte(testPackageMainDef), 0o600))
	require.NoError(t, os.WriteFile("util.go", []byte(testPackageMainDef), 0o600))
	require.NoError(t, os.WriteFile("main_test.go", []byte(testPackageMainDef), 0o600))

	// Create excluded directories.
	require.NoError(t, os.MkdirAll("vendor", 0o755))
	require.NoError(t, os.WriteFile("vendor/vendored.go", []byte("package vendor\n"), 0o600))

	// Test - should find main.go and util.go, but NOT test files, vendor files.
	files, err := FindGoFiles()
	require.NoError(t, err)
	require.Len(t, files, 2)
}

func TestCheckNonFIPS_WithTempDir(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory.

	// Save current directory.
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	// Create temp directory.
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))

	// Create clean Go file without banned algorithms.
	cleanContent := "package main\n\nimport (\n\t\"crypto/sha256\"\n)\n\nfunc main() { sha256.New() }\n"
	require.NoError(t, os.WriteFile("main.go", []byte(cleanContent), 0o600))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	// Test - should pass with FIPS-compliant code.
	err = Check(logger)
	require.NoError(t, err)
}

func TestCheckNonFIPS_WithViolations(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory and redirects stderr.

	// Save current directory.
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	// Create temp directory.
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))

	// Create Go file with banned algorithm (bcrypt).
	badContent := "package main\n\nimport \"golang.org/x/crypto/bcrypt\"\n\nfunc main() { bcrypt.GenerateFromPassword(nil, 0) }\n"
	require.NoError(t, os.WriteFile("bad.go", []byte(badContent), 0o600))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	// Test - should fail with violations.
	err = Check(logger)
	require.Error(t, err)
	require.Contains(t, err.Error(), "non-FIPS algorithm violations")
}

func TestFindGoFiles_ErrorPath(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory.
	if runtime.GOOS == cryptoutilSharedMagic.OSNameWindows {
		t.Skip("os.Chmod does not enforce POSIX permissions on Windows")
	}

	// Save current directory.
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	// Create temp directory.
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))

	// Create a subdirectory that will trigger walk error.
	subDir := "subdir"
	require.NoError(t, os.MkdirAll(subDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "file.go"), []byte("package main\n"), 0o600))

	// Make subdirectory unreadable.
	require.NoError(t, os.Chmod(subDir, 0o000))

	defer func() {
		// Restore permissions for cleanup.
		_ = os.Chmod(filepath.Join(tempDir, subDir), 0o755)
	}()

	// Test - should get error from walking directory.
	_, err = FindGoFiles()
	require.Error(t, err)
}

func TestCheckFileForNonFIPS_ReadFileError(t *testing.T) {
	t.Parallel()

	// Passing a nonexistent path should return an error message in the violations.
	violations := CheckFileForNonFIPS("/nonexistent/path/to/file.go")
	require.Len(t, violations, 1)
	require.Contains(t, violations[0], "Error reading file")
}

func TestCheckFileForNonFIPS_NolintComment(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create a Go file with a banned algorithm and a nolint comment on the same line.
	goFile := filepath.Join(tempDir, "main.go")
	content := "package main\n\nimport \"crypto/md5\" //nolint:gosec // Required for legacy\n"
	require.NoError(t, os.WriteFile(goFile, []byte(content), 0o600))

	violations := CheckFileForNonFIPS(goFile)
	require.Empty(t, violations, "nolint comment should suppress violation")
}

func TestCheck_FindGoFilesError(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory.
	if runtime.GOOS == "windows" {
		t.Skip("chmod 0o000 does not work on Windows")
	}

	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { require.NoError(t, os.Chdir(origDir)) }()

	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))

	// Create a subdirectory that will trigger walk error.
	subDir := "lockdir"
	require.NoError(t, os.MkdirAll(subDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "file.go"), []byte("package main\n"), 0o600))
	require.NoError(t, os.Chmod(subDir, 0o000))

	defer func() { _ = os.Chmod(filepath.Join(tempDir, subDir), 0o755) }()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to find Go files")
}
