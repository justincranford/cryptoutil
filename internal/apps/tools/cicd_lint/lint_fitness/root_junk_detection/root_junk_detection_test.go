// Copyright (c) 2025 Justin Cranford

package root_junk_detection

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func newTestLogger() *cryptoutilCmdCicdCommon.Logger {
	return cryptoutilCmdCicdCommon.NewLogger("test")
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	err := os.WriteFile(path, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)
}

func TestCheckInDir_EmptyDir_Passes(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_LegitimateFiles_Passes(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "go.mod"), "module example.com/test\n")
	writeFile(t, filepath.Join(tmp, "README.md"), "# Readme\n")
	writeFile(t, filepath.Join(tmp, "LICENSE"), "license text\n")
	writeFile(t, filepath.Join(tmp, "go.sum"), "")

	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_PyFile_Detected(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "write_helper.py"), "#!/usr/bin/env python3\n")

	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "junk file(s)")
}

func TestCheckInDir_ExeFile_Detected(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "cryptoutil.exe"), "binary")

	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "junk file(s)")
}

func TestCheckInDir_TestExeFile_Detected(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "mypackage.test.exe"), "binary")

	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "junk file(s)")
}

func TestCheckInDir_CoverageFile_Detected(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "coverage.out"), "mode: atomic\n")

	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "junk file(s)")
}

func TestCheckInDir_CoveragePrefix_Detected(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "coverage_all"), "coverage data")

	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "junk file(s)")
}

func TestCheckInDir_MultipleViolations_ReportsAll(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "helper.py"), "")
	writeFile(t, filepath.Join(tmp, "tool.exe"), "")
	writeFile(t, filepath.Join(tmp, "coverage.out"), "")

	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "3 junk file(s)")
}

func TestCheckInDir_DirectoryNamedCoverage_Ignored(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	err := os.MkdirAll(filepath.Join(tmp, "coverage_output"), cryptoutilSharedMagic.DirPermissions)
	require.NoError(t, err)

	err = CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestIsBannedRootFile_CaseInsensitive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{name: "py lowercase", filename: "script.py", want: true},
		{name: "py uppercase", filename: "SCRIPT.PY", want: true},
		{name: "exe lowercase", filename: "binary.exe", want: true},
		{name: "exe uppercase", filename: "BINARY.EXE", want: true},
		{name: "coverage prefix", filename: "coverage.out", want: true},
		{name: "coverage prefix uppercase", filename: "COVERAGE.out", want: true},
		{name: "go.mod not banned", filename: "go.mod", want: false},
		{name: "readme not banned", filename: "README.md", want: false},
		{name: "go source not banned", filename: "main.go", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := isBannedRootFile(tc.filename)
			require.Equal(t, tc.want, got)
		})
	}
}
