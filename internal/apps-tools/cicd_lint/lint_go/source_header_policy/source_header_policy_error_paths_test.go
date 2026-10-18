// Copyright (c) 2025-2026 Justin Cranford.
package source_header_policy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const (
	testMITLicenseHeader = "// SPDX-License-Identifier: MIT"
	testHeaderComment    = "// header\n"
	testInvalidYear      = "year"
)

func TestCheck(t *testing.T) {
	t.Parallel()

	require.NoError(t, Check(cryptoutilCmdCicdCommon.NewLogger("source-header-policy-test")))
}

func TestFindViolations_ErrorPathsAndExcludedDirectories(t *testing.T) {
	t.Parallel()

	missingRoot := filepath.Join(t.TempDir(), "missing")
	_, err := findViolations(missingRoot, time.Now().UTC().Year())
	require.ErrorContains(t, err, "failed to walk source files")
	err = checkInDirWithYear(
		cryptoutilCmdCicdCommon.NewLogger("source-header-policy-test"),
		missingRoot,
		time.Now().UTC().Year(),
	)
	require.ErrorContains(t, err, "source-header-policy")

	rootDir := t.TempDir()
	vendorDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirVendor)
	gitDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGit)

	require.NoError(t, os.MkdirAll(vendorDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))
	require.NoError(t, os.MkdirAll(gitDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))

	badHeader := strings.Join([]string{
		"// Copyright (c) 2028-2027 Justin Cranford.",
		testMITLicenseHeader,
		"",
		testPackageExample,
		"",
	}, "\n")
	for _, dir := range []string{vendorDir, gitDir} {
		require.NoError(t, os.WriteFile(filepath.Join(dir, "ignored.go"), []byte(badHeader), cryptoutilSharedMagic.CacheFilePermissions))
	}

	violations, err := findViolations(rootDir, time.Now().UTC().Year())
	require.NoError(t, err)
	require.Empty(t, violations)

	require.NoError(t, os.WriteFile(
		filepath.Join(rootDir, "source.go"),
		[]byte(strings.Join([]string{
			"// SPDX-License-Identifier: AGPL-3.0-only",
			"",
			testPackageExample,
		}, "\n")),
		cryptoutilSharedMagic.CacheFilePermissions,
	))

	currentYear := time.Now().UTC().Year()

	require.NoError(t, os.WriteFile(
		filepath.Join(rootDir, "bad_header.go"),
		[]byte(badHeader),
		cryptoutilSharedMagic.CacheFilePermissions,
	))
	currentViolations, err := findViolations(rootDir, currentYear)
	require.NoError(t, err)
	require.Len(t, currentViolations, 3)
	require.Contains(t, strings.Join(currentViolations, "\n"), "copyright year range start")

	_, err = findViolationsWithFileChecker(rootDir, time.Now().UTC().Year(), func(string, int) ([]string, error) {
		return nil, os.ErrPermission
	})
	require.ErrorIs(t, err, os.ErrPermission)
}

func TestCheckFile_OpenAndHeaderReadErrors(t *testing.T) {
	t.Parallel()

	_, err := checkFile(filepath.Join(t.TempDir(), "missing.go"), time.Now().UTC().Year())
	require.ErrorContains(t, err, "failed to open")

	path := filepath.Join(t.TempDir(), "oversized_header.go")
	require.NoError(t, os.WriteFile(path, []byte("//"+strings.Repeat("x", 70_000)), cryptoutilSharedMagic.CacheFilePermissions))

	_, err = checkFile(path, time.Now().UTC().Year())
	require.ErrorContains(t, err, "failed to read header section")
	require.ErrorContains(t, err, "token too long")

	parseErrorPath := filepath.Join(t.TempDir(), "parse_error.go")
	parseErrorContent := strings.Join([]string{
		"// Copyright (c) 2025 Justin Cranford.",
		testSpdxLicenseIdentifierAgpl30Only,
		"",
		testPackageExample,
		"",
	}, "\n")
	require.NoError(t, os.WriteFile(
		parseErrorPath,
		[]byte(parseErrorContent),
		cryptoutilSharedMagic.CacheFilePermissions,
	))
	violations, err := checkFileWithCopyrightYearParser(
		parseErrorPath,
		time.Now().UTC().Year(),
		func(string, string) (int, int, error) {
			return 0, 0, os.ErrInvalid
		},
	)
	require.NoError(t, err)
	require.Len(t, violations, 1)
	require.Contains(t, violations[0], "invalid copyright header years")
}

func TestReadHeaderSection_StopsBeforePackageOrCodeAndAcceptsEOF(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		want    string
	}{
		{name: "package declaration", content: testHeaderComment + "\npackage example\n", want: testHeaderComment},
		{name: "non-comment code", content: testHeaderComment + "\nimport \"example\"\n", want: testHeaderComment},
		{name: "empty file", content: "", want: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "header.go")
			require.NoError(t, os.WriteFile(path, []byte(tc.content), cryptoutilSharedMagic.CacheFilePermissions))
			file, err := os.Open(path)
			require.NoError(t, err)

			defer func() { _ = file.Close() }()

			header, err := readHeaderSection(file)
			require.NoError(t, err)
			require.Equal(t, tc.want, header)
		})
	}
}

func TestParseCopyrightYears_InvalidInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		startYear string
		endYear   string
		wantError string
	}{
		{name: "invalid start", startYear: testInvalidYear, wantError: "invalid start year"},
		{name: "invalid end", startYear: "2025", endYear: testInvalidYear, wantError: "invalid end year"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, _, err := parseCopyrightYears(tc.startYear, tc.endYear)
			require.ErrorContains(t, err, tc.wantError)
		})
	}
}
