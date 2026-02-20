// Copyright (c) 2025 Justin Cranford

package cgo_free_sqlite

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"

	"github.com/stretchr/testify/require"
)

func TestCheckGoModForCGO_ValidFile(t *testing.T) {
	t.Parallel()

	// Create temp go.mod file without banned modules.
	tmpDir := t.TempDir()
	goModFile := filepath.Join(tmpDir, "go.mod")

	content := `module example.com/myproject

go 1.21

require (
	modernc.org/sqlite v1.29.0
	github.com/golang-migrate/migrate/v4 v4.17.0
)
`

	err := os.WriteFile(goModFile, []byte(content), 0o600)
	require.NoError(t, err)

	violations, err := CheckGoModForCGO(goModFile)
	require.NoError(t, err)
	require.Empty(t, violations, "Valid go.mod should have no violations")
}

func TestCheckGoModForCGO_BannedModule(t *testing.T) {
	t.Parallel()

	// Create temp go.mod file with banned module (direct dependency).
	tmpDir := t.TempDir()
	goModFile := filepath.Join(tmpDir, "go.mod")

	content := `module example.com/myproject

go 1.21

require (
	github.com/mattn/go-sqlite3 v1.14.19
	github.com/golang-migrate/migrate/v4/database/sqlite3 v4.17.0
)
`

	err := os.WriteFile(goModFile, []byte(content), 0o600)
	require.NoError(t, err)

	violations, err := CheckGoModForCGO(goModFile)
	require.NoError(t, err)
	require.Len(t, violations, 2, "Should detect 2 banned modules")
	require.Contains(t, strings.Join(violations, "\n"), "go-sqlite3", "Should detect banned CGO sqlite")
	require.Contains(t, strings.Join(violations, "\n"), "database/sqlite3", "Should detect banned CGO migrate")
}

func TestCheckGoModForCGO_IndirectModule(t *testing.T) {
	t.Parallel()

	// Create temp go.mod file with banned module as indirect.
	tmpDir := t.TempDir()
	goModFile := filepath.Join(tmpDir, "go.mod")

	content := `module example.com/myproject

go 1.21

require (
	github.com/mattn/go-sqlite3 v1.14.19 // indirect
)
`

	err := os.WriteFile(goModFile, []byte(content), 0o600)
	require.NoError(t, err)

	violations, err := CheckGoModForCGO(goModFile)
	require.NoError(t, err)
	require.Empty(t, violations, "Indirect dependencies should not be flagged")
}

func TestCheckGoModForCGO_FileNotFound(t *testing.T) {
	t.Parallel()

	violations, err := CheckGoModForCGO("/nonexistent/path/go.mod")
	require.Error(t, err)
	require.Nil(t, violations)
	require.Contains(t, err.Error(), "failed to open go.mod")
}

func TestCheckRequiredCGOModule_Found(t *testing.T) {
	t.Parallel()

	// Create temp go.mod file with required module.
	tmpDir := t.TempDir()
	goModFile := filepath.Join(tmpDir, "go.mod")

	content := `module example.com/myproject

go 1.21

require (
	modernc.org/sqlite v1.29.0
)
`

	err := os.WriteFile(goModFile, []byte(content), 0o600)
	require.NoError(t, err)

	found, err := CheckRequiredCGOModule(goModFile)
	require.NoError(t, err)
	require.True(t, found, "Required module should be found")
}

func TestCheckRequiredCGOModule_NotFound(t *testing.T) {
	t.Parallel()

	// Create temp go.mod file without required module.
	tmpDir := t.TempDir()
	goModFile := filepath.Join(tmpDir, "go.mod")

	content := `module example.com/myproject

go 1.21

require (
	github.com/some/other/module v1.0.0
)
`

	err := os.WriteFile(goModFile, []byte(content), 0o600)
	require.NoError(t, err)

	found, err := CheckRequiredCGOModule(goModFile)
	require.NoError(t, err)
	require.False(t, found, "Required module should not be found")
}

func TestCheckRequiredCGOModule_FileNotFound(t *testing.T) {
	t.Parallel()

	found, err := CheckRequiredCGOModule("/nonexistent/path/go.mod")
	require.Error(t, err)
	require.False(t, found)
	require.Contains(t, err.Error(), "failed to open go.mod")
}

func TestCheckGoFileForCGO_Clean(t *testing.T) {
	t.Parallel()

	// Create temp file without banned imports.
	tmpDir := t.TempDir()
	cleanFile := filepath.Join(tmpDir, "clean.go")

	content := `package main

import (
	"modernc.org/sqlite"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
)

func main() {
	// Using CGO-free sqlite
}
`

	err := os.WriteFile(cleanFile, []byte(content), 0o600)
	require.NoError(t, err)

	violations, err := CheckGoFileForCGO(cleanFile)
	require.NoError(t, err)
	require.Empty(t, violations, "Clean file should have no violations")
}

func TestCheckGoFileForCGO_BannedImport(t *testing.T) {
	t.Parallel()

	// Create temp file with banned import.
	tmpDir := t.TempDir()
	bannedFile := filepath.Join(tmpDir, "banned.go")

	content := `package main

import (
	_ "github.com/mattn/go-sqlite3"
)

func main() {
}
`

	err := os.WriteFile(bannedFile, []byte(content), 0o600)
	require.NoError(t, err)

	violations, err := CheckGoFileForCGO(bannedFile)
	require.NoError(t, err)
	require.NotEmpty(t, violations, "Banned import should be detected")
	require.Contains(t, strings.Join(violations, "\n"), "banned CGO import detected")
}

func TestCheckGoFileForCGO_BannedMigrateImport(t *testing.T) {
	t.Parallel()

	// Create temp file with banned migrate import.
	tmpDir := t.TempDir()
	bannedFile := filepath.Join(tmpDir, "banned_migrate.go")

	content := `package main

import (
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
)

func main() {
}
`

	err := os.WriteFile(bannedFile, []byte(content), 0o600)
	require.NoError(t, err)

	violations, err := CheckGoFileForCGO(bannedFile)
	require.NoError(t, err)
	require.NotEmpty(t, violations, "Banned migrate import should be detected")
	require.Contains(t, strings.Join(violations, "\n"), "banned CGO migrate import detected")
}

func TestCheckGoFileForCGO_LintGoSkipped(t *testing.T) {
	t.Parallel()

	// Create temp file in a lint_go directory (should be skipped).
	tmpDir := t.TempDir()
	lintGoDir := filepath.Join(tmpDir, "lint_go")
	require.NoError(t, os.MkdirAll(lintGoDir, 0o755))

	skippedFile := filepath.Join(lintGoDir, "lint_go.go")

	// Even with banned imports, should be skipped.
	content := `package main

import (
	_ "github.com/mattn/go-sqlite3"
)
`

	err := os.WriteFile(skippedFile, []byte(content), 0o600)
	require.NoError(t, err)

	violations, err := CheckGoFileForCGO(skippedFile)
	require.NoError(t, err)
	require.Empty(t, violations, "lint_go files should be skipped")
}

func TestCheckGoFileForCGO_FileNotFound(t *testing.T) {
	t.Parallel()

	violations, err := CheckGoFileForCGO("/nonexistent/path/file.go")
	require.Error(t, err)
	require.Nil(t, violations)
	require.Contains(t, err.Error(), "failed to open")
}

func TestPrintCGOViolations_AllTypes(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test redirects os.Stderr which is global.

	// Capture stderr to verify output.
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	goModViolations := []string{"go.mod:5: banned CGO module"}
	importViolations := []string{"file.go:10: banned CGO import"}
	hasRequired := false

	PrintCGOViolations(goModViolations, importViolations, hasRequired)

	_ = w.Close()
	os.Stderr = oldStderr

	output, _ := io.ReadAll(r)

	require.Contains(t, string(output), "CGO validation failed")
	require.Contains(t, string(output), "go.mod violations")
	require.Contains(t, string(output), "Import violations")
	require.Contains(t, string(output), "Required module missing")
}

func TestPrintCGOViolations_GoModOnly(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test redirects os.Stderr which is global.
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	PrintCGOViolations([]string{"go.mod:5: banned module"}, nil, true)

	_ = w.Close()
	os.Stderr = oldStderr

	output, _ := io.ReadAll(r)

	require.Contains(t, string(output), "go.mod violations")
	require.NotContains(t, string(output), "Import violations")
	require.NotContains(t, string(output), "Required module missing")
}

func TestPrintCGOViolations_ImportOnly(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test redirects os.Stderr which is global.
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	PrintCGOViolations(nil, []string{"file.go:10: banned import"}, true)

	_ = w.Close()
	os.Stderr = oldStderr

	output, _ := io.ReadAll(r)

	require.NotContains(t, string(output), "go.mod violations")
	require.Contains(t, string(output), "Import violations")
	require.NotContains(t, string(output), "Required module missing")
}

func TestCheckCGOFreeSQLite_WithTempDir(t *testing.T) {
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

	// Build required module string dynamically.
	var required strings.Builder
	required.WriteString("modernc.org/")
	required.WriteString("sqlite")

	// Create go.mod with required CGO-free module.
	goModContent := "module testmod\n\ngo 1.21\n\nrequire (\n\t" + required.String() + " v1.30.0\n)\n"
	require.NoError(t, os.WriteFile("go.mod", []byte(goModContent), 0o600))

	// Create clean Go file.
	require.NoError(t, os.WriteFile("main.go", []byte(testMainContent), 0o600))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	// Test - should pass with required module present.
	err = Check(logger)
	require.NoError(t, err)
}

func TestCheckCGOFreeSQLite_MissingRequired(t *testing.T) {
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

	// Create go.mod WITHOUT required CGO-free module.
	goModContent := "module testmod\n\ngo 1.21\n"
	require.NoError(t, os.WriteFile("go.mod", []byte(goModContent), 0o600))

	// Create clean Go file.
	require.NoError(t, os.WriteFile("main.go", []byte(testMainContent), 0o600))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	// Test - should fail because required module is missing.
	err = Check(logger)
	require.Error(t, err)
	require.Contains(t, err.Error(), "CGO validation failed")
}

func TestCheck_NoGoMod(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory.
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { require.NoError(t, os.Chdir(origDir)) }()

	tmpDir := t.TempDir()
	require.NoError(t, os.Chdir(tmpDir))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	// Without go.mod in the current directory, CheckGoModForCGO("go.mod") fails.
	// This covers the "failed to check go.mod" error branch in Check().
	err = Check(logger)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to check go.mod")
}

func TestCheck_WalkError(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory.
	if runtime.GOOS == osWindows {
		t.Skip("os.Chmod does not enforce POSIX permissions on Windows")
	}

	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { require.NoError(t, os.Chdir(origDir)) }()

	tmpDir := t.TempDir()
	require.NoError(t, os.Chdir(tmpDir))

	// Write go.mod with required module so CheckGoModForCGO passes.
	goModContent := "module testmod\n\ngo 1.21\n\nrequire (\n\tmodernc.org/sqlite v1.30.0\n)\n"
	require.NoError(t, os.WriteFile("go.mod", []byte(goModContent), 0o600))

	// Create chmod 0000 subdir - filepath.Walk callback receives OS error when
	// Walk tries to ReadDir the locked directory, covering the walk callback
	// error path (lines 121-123) and the Check() error path (lines 38-40).
	require.NoError(t, os.MkdirAll("locked", 0o700))
	require.NoError(t, os.Chmod("locked", 0o000))

	t.Cleanup(func() { _ = os.Chmod(filepath.Join(tmpDir, "locked"), 0o700) })

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to check Go files")
}
