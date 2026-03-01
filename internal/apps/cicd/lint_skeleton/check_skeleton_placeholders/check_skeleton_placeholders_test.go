// Copyright (c) 2025 Justin Cranford

package check_skeleton_placeholders

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	goSkeletonFuncContent = "package myservice\n\nfunc NewSkeletonService() {}\n"
	goSkeletonTypeContent = "package myservice\n\ntype SkeletonServer struct{}\n"
)

// Note: This test file is in internal/apps/cicd/ which is excluded from lint-skeleton scans.
// Therefore skeleton-related words can appear here without triggering violations.

func TestFindViolations_CleanFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	cleanFile := filepath.Join(tempDir, "clean.go")
	content := "package myservice\n\nfunc NewMyService() {}\n"
	err := os.WriteFile(cleanFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	violations, err := FindViolations(tempDir)
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestFindViolations_WithSkeletonLowercase(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	badFile := filepath.Join(tempDir, "bad.go")
	content := goSkeletonFuncContent
	err := os.WriteFile(badFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	violations, err := FindViolations(tempDir)
	require.NoError(t, err)
	require.NotEmpty(t, violations)
	require.Contains(t, violations[0].Content, cryptoutilSharedMagic.SkeletonProductNameTitleCase)
}

func TestFindViolations_WithSkeletonWord(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	badFile := filepath.Join(tempDir, "bad.go")
	content := goSkeletonTypeContent
	err := os.WriteFile(badFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	violations, err := FindViolations(tempDir)
	require.NoError(t, err)
	require.NotEmpty(t, violations)
	require.Equal(t, badFile, violations[0].File)
	require.Equal(t, 3, violations[0].Line)
	require.Equal(t, cryptoutilSharedMagic.SkeletonProductNameTitleCase, violations[0].Word)
}

func TestFindViolations_WithSkeletonAllCaps(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	badFile := filepath.Join(tempDir, "bad.go")
	content := "package myservice\n\nconst SKELETON_PORT = 8080\n"
	err := os.WriteFile(badFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	violations, err := FindViolations(tempDir)
	require.NoError(t, err)
	require.NotEmpty(t, violations)
	require.Equal(t, cryptoutilSharedMagic.SkeletonProductNameUpperCase, violations[0].Word)
}

func TestFindViolations_TestFilesSkipped(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	testFile := filepath.Join(tempDir, "bad_test.go")
	content := "package myservice\n\nfunc TestSkeletonService(t *testing.T) {}\n"
	err := os.WriteFile(testFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	violations, err := FindViolations(tempDir)
	require.NoError(t, err)
	require.Empty(t, violations, "test files should be excluded from validation")
}

func TestFindViolations_SkeletonDirExcluded(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	skeletonDir := filepath.Join(tempDir, "internal", "apps", cryptoutilSharedMagic.SkeletonProductName, cryptoutilSharedMagic.SkeletonTemplateServiceName)
	err := os.MkdirAll(skeletonDir, 0o700)
	require.NoError(t, err)

	skeletonFile := filepath.Join(skeletonDir, "template.go")
	content := "package template\n\n// Skeleton template service\nfunc Skeleton() {}\n"
	err = os.WriteFile(skeletonFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	violations, err := FindViolations(tempDir)
	require.NoError(t, err)
	require.Empty(t, violations, "files in internal/apps/skeleton/ must be excluded")
}

func TestFindViolations_CmdSkeletonTemplateExcluded(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	cmdDir := filepath.Join(tempDir, "cmd", cryptoutilSharedMagic.OTLPServiceSkeletonTemplate)
	err := os.MkdirAll(cmdDir, 0o700)
	require.NoError(t, err)

	mainFile := filepath.Join(cmdDir, "main.go")
	content := "package main\n\n// Entry point for skeleton-template service\nfunc main() {}\n"
	err = os.WriteFile(mainFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	violations, err := FindViolations(tempDir)
	require.NoError(t, err)
	require.Empty(t, violations, "files in cmd/skeleton-template/ must be excluded")
}

func TestFindViolations_CmdSkeletonExcluded(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	cmdDir := filepath.Join(tempDir, "cmd", cryptoutilSharedMagic.SkeletonProductName)
	err := os.MkdirAll(cmdDir, 0o700)
	require.NoError(t, err)

	mainFile := filepath.Join(cmdDir, "main.go")
	content := "package main\n\n// Skeleton product entry point\nimport \"cryptoutil/internal/apps/skeleton\"\n"
	err = os.WriteFile(mainFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	violations, err := FindViolations(tempDir)
	require.NoError(t, err)
	require.Empty(t, violations, "files in cmd/skeleton/ must be excluded")
}

func TestFindViolations_VendorDirExcluded(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	vendorDir := filepath.Join(tempDir, cryptoutilSharedMagic.CICDExcludeDirVendor, "somelib")
	err := os.MkdirAll(vendorDir, 0o700)
	require.NoError(t, err)

	vendorFile := filepath.Join(vendorDir, "skeleton.go")
	content := "package somelib\n\ntype SkeletonServer struct{}\n"
	err = os.WriteFile(vendorFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	violations, err := FindViolations(tempDir)
	require.NoError(t, err)
	require.Empty(t, violations, "vendor directory must be excluded")
}

func TestFindViolations_MagicDirExcluded(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	magicDir := filepath.Join(tempDir, "internal", "shared", "magic")
	err := os.MkdirAll(magicDir, 0o700)
	require.NoError(t, err)

	magicFile := filepath.Join(magicDir, "magic_skeleton.go")
	content := "package magic\n\nconst SkeletonPort = 8900\n"
	err = os.WriteFile(magicFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	violations, err := FindViolations(tempDir)
	require.NoError(t, err)
	require.Empty(t, violations, "internal/shared/magic/ must be excluded")
}

func TestFindViolations_CryptoutilDirExcluded(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	cryptoutilDir := filepath.Join(tempDir, "internal", "apps", cryptoutilSharedMagic.DefaultOTLPServiceDefault)
	err := os.MkdirAll(cryptoutilDir, 0o700)
	require.NoError(t, err)

	file := filepath.Join(cryptoutilDir, "cryptoutil.go")
	content := "package cryptoutil\n\nimport \"cryptoutil/internal/apps/skeleton\"\n"
	err = os.WriteFile(file, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	violations, err := FindViolations(tempDir)
	require.NoError(t, err)
	require.Empty(t, violations, "internal/apps/cryptoutil/ must be excluded")
}

func TestFindViolations_CicdDirExcluded(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	cicdDir := filepath.Join(tempDir, "internal", "apps", "cicd")
	err := os.MkdirAll(cicdDir, 0o700)
	require.NoError(t, err)

	file := filepath.Join(cicdDir, "cicd.go")
	content := "package cicd\n\nconst SkeletonCmd = \"lint-skeleton\"\n"
	err = os.WriteFile(file, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	violations, err := FindViolations(tempDir)
	require.NoError(t, err)
	require.Empty(t, violations, "internal/apps/cicd/ must be excluded")
}

func TestFindViolations_TemplateDirExcluded(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	templateDir := filepath.Join(tempDir, "internal", "apps", cryptoutilSharedMagic.SkeletonTemplateServiceName)
	err := os.MkdirAll(templateDir, 0o700)
	require.NoError(t, err)

	file := filepath.Join(templateDir, "server.go")
	content := "package template\n\n// Supported services: skeleton-template\n"
	err = os.WriteFile(file, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	violations, err := FindViolations(tempDir)
	require.NoError(t, err)
	require.Empty(t, violations, "internal/apps/template/ must be excluded")
}

func TestFindViolations_NonGoFilesIgnored(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	yamlFile := filepath.Join(tempDir, "myfile.yaml")
	content := "name: Skeleton-service\n"
	err := os.WriteFile(yamlFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	violations, err := FindViolations(tempDir)
	require.NoError(t, err)
	require.Empty(t, violations, "non-.go files must be ignored")
}

func TestFindViolations_OnlyReportedOnce(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	badFile := filepath.Join(tempDir, "bad.go")
	content := "package myservice\n\n// line with skeleton and Skeleton keywords\n"
	err := os.WriteFile(badFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	violations, err := FindViolations(tempDir)
	require.NoError(t, err)
	require.Len(t, violations, 1, "each line should only be reported once")
}

func TestCheckInDir_PassesClean(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	cleanFile := filepath.Join(tempDir, "clean.go")
	err := os.WriteFile(cleanFile, []byte("package myservice\n\nfunc NewMyService() {}\n"), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = CheckInDir(logger, tempDir)
	require.NoError(t, err)
}

func TestCheckInDir_FailsWithViolation(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	badFile := filepath.Join(tempDir, "bad.go")
	content := goSkeletonTypeContent
	err := os.WriteFile(badFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = CheckInDir(logger, tempDir)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "[ValidateSkeleton]"))
	require.True(t, strings.Contains(err.Error(), "ARCHITECTURE.md Section 5.1"))
}

func TestCheckInDir_InvalidRootDir(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, "/nonexistent/path/that/does/not/exist")
	require.Error(t, err)
}

func TestCheck_DelegatesCheckInDir(t *testing.T) {
	// NOTE: Cannot use t.Parallel() — test changes working directory.
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { require.NoError(t, os.Chdir(origDir)) }()

	tmpDir := t.TempDir()
	require.NoError(t, os.Chdir(tmpDir))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger)
	require.NoError(t, err)
}

func TestScanFile_EmptyFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	emptyFile := filepath.Join(tempDir, "empty.go")
	err := os.WriteFile(emptyFile, []byte(""), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	violations, err := scanFile(emptyFile)
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestScanFile_MultipleViolationsInFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	badFile := filepath.Join(tempDir, "multi.go")
	content := "package myservice\n\nvar a = \"skeleton\"\nvar b = \"Skeleton\"\n"
	err := os.WriteFile(badFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	violations, err := scanFile(badFile)
	require.NoError(t, err)
	require.Len(t, violations, 2, "each line with violation should be reported separately")
}

func TestScanFile_FileNotFound(t *testing.T) {
	t.Parallel()

	_, err := scanFile("/nonexistent/path/file.go")
	require.Error(t, err)
}

func TestFindViolations_WalkError(t *testing.T) {
	t.Parallel()

	// Use a non-directory path as rootDir to cause WalkDir to fail.
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "notadir.go")
	err := os.WriteFile(tempFile, []byte("package test\n"), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	// WalkDir on a file still works (treats it as a single entry), no error expected.
	// This exercises the non-directory walk path.
	violations, err := FindViolations(tempFile)
	require.NoError(t, err)
	require.Empty(t, violations, "scanning a single clean .go file directly should find no violations")
}

func TestFindViolations_ReadError(t *testing.T) {
	t.Parallel()

	// Create a temp dir with a .go file we then make unreadable to trigger a read error.
	tempDir := t.TempDir()
	badFile := filepath.Join(tempDir, "unreadable.go")
	err := os.WriteFile(badFile, []byte("package test\n"), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	// Remove read permissions so scanFile fails to open it.
	require.NoError(t, os.Chmod(badFile, 0o000))

	defer func() { _ = os.Chmod(badFile, cryptoutilSharedMagic.CICDOutputFilePermissions) }()

	_, findErr := FindViolations(tempDir)
	require.Error(t, findErr, "should fail when a file cannot be opened")
}

func TestPrintViolations_OutputsErrors(_ *testing.T) {
	// Exercise printViolations for code coverage — it writes to stderr.
	violations := []Violation{
		{File: "fake.go", Line: 1, Word: cryptoutilSharedMagic.SkeletonProductNameTitleCase, Content: "type SkeletonService struct{}"},
	}

	printViolations(violations)
}

func TestFindViolations_MultipleFilesOneViolation(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	cleanFile := filepath.Join(tempDir, "clean.go")
	err := os.WriteFile(cleanFile, []byte("package myservice\n\nfunc Ok() {}\n"), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	badFile := filepath.Join(tempDir, "bad.go")
	content := goSkeletonFuncContent
	err = os.WriteFile(badFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	violations, err := FindViolations(tempDir)
	require.NoError(t, err)
	require.Len(t, violations, 1)
	require.Equal(t, badFile, violations[0].File)
}

// TestFindViolations_ErrorWrapping ensures WalkDir errors are returned (not swallowed).
func TestFindViolations_ErrorWrapping(t *testing.T) {
	t.Parallel()

	_, err := FindViolations("/path/that/does/not/exist/at/all")
	// filepath.Abs on a non-existent path succeeds (it just resolves the path string).
	// The WalkDir call will fail when it tries to open the directory.
	require.Error(t, err, "walking a non-existent directory should return an error")

	_ = errors.New("expected error")
}

// TestFindViolations_AbsError exercises the filepath.Abs error path via test seam injection.
// NOT parallel: modifies a package-level seam variable — see ARCHITECTURE.md Section 10.2.4.
func TestFindViolations_AbsError(t *testing.T) {
	orig := filepathAbs

	defer func() { filepathAbs = orig }()

	filepathAbs = func(_ string) (string, error) { return "", errors.New("mock abs error") }

	_, err := FindViolations(".")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to resolve rootDir")
}

// TestFindViolations_RelError exercises the filepath.Rel error path via test seam injection.
// NOT parallel: modifies a package-level seam variable — see ARCHITECTURE.md Section 10.2.4.
func TestFindViolations_RelError(t *testing.T) {
	orig := filepathRel

	defer func() { filepathRel = orig }()

	filepathRel = func(_, _ string) (string, error) { return "", errors.New("mock rel error") }

	tempDir := t.TempDir()
	goFile := filepath.Join(tempDir, "dummy.go")
	require.NoError(t, os.WriteFile(goFile, []byte("package test\n"), cryptoutilSharedMagic.CacheFilePermissions))

	_, err := FindViolations(tempDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to compute relative path")
}
