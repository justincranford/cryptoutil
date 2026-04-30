// Copyright (c) 2025-2026 Justin Cranford.
package cmd_suite_template

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilFitnessRegistry "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// findProjectRoot traverses up from the current directory to locate go.mod.
func findProjectRoot() (string, error) {
	dir, _ := os.Getwd()

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}

		dir = parent
	}
}

// validSuiteMain returns a valid main.go content for the given suite ID.
// Suite entry points use os.Args directly (NOT os.Args[1:]).
func validSuiteMain(suiteID string) string {
	return `// Copyright (c) 2025-2026 Justin Cranford.

package main

import (
	cryptoutilSuite "cryptoutil/internal/apps/` + suiteID + `"
)

func main() {
	cryptoutilSuite.Main(os.Args)
}
`
}

// TestCheck_RealWorkspace verifies the linter passes against the actual workspace.
func TestCheck_RealWorkspace(t *testing.T) {
	t.Parallel()

	root, err := findProjectRoot()
	if err != nil {
		t.Skip("Skipping integration test - cannot find project root (no go.mod)")
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = CheckInDir(logger, root)
	require.NoError(t, err)
}

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheck_Integration(t *testing.T) {
	root, err := findProjectRoot()
	if err != nil {
		t.Skip("Skipping integration test - cannot find project root (no go.mod)")
	}

	origDir, getErr := os.Getwd()
	require.NoError(t, getErr)

	require.NoError(t, os.Chdir(root))

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger)
	require.NoError(t, err)
}

// TestCheckInDir_AllValid verifies all suites pass when synthetic main.go files are valid.
func TestCheckInDir_AllValid(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	for _, suite := range cryptoutilFitnessRegistry.AllSuites() {
		cmdDir := filepath.Join(tmpDir, "cmd", suite.ID)
		require.NoError(t, os.MkdirAll(cmdDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.WriteFile(filepath.Join(cmdDir, "main.go"), []byte(validSuiteMain(suite.ID)), cryptoutilSharedMagic.CacheFilePermissions))
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

// TestCheckInDir_MissingMainFile exercises the "main.go missing" violation path.
func TestCheckInDir_MissingMainFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	// No cmd dirs created — all main.go files are missing.

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cmd suite template violations")
}

// TestCheckCmdMainFile_MissingPackageMain exercises the "package main" violation.
func TestCheckCmdMainFile_MissingPackageMain(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mainPath := filepath.Join(tmpDir, "main.go")
	require.NoError(t, os.WriteFile(mainPath, []byte(`package notmain
import "cryptoutil/internal/apps/cryptoutil"
func main() { _ = os.Args }
`), cryptoutilSharedMagic.CacheFilePermissions))

	errs := ExportedCheckCmdMainFile(mainPath, cryptoutilSharedMagic.DefaultOTLPServiceDefault, "cryptoutil/internal/apps/cryptoutil", false)
	require.Len(t, errs, 1)
	require.Contains(t, errs[0], "missing 'package main'")
}

// TestCheckCmdMainFile_MissingImport exercises the missing import violation.
func TestCheckCmdMainFile_MissingImport(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mainPath := filepath.Join(tmpDir, "main.go")
	require.NoError(t, os.WriteFile(mainPath, []byte(`package main
func main() { _ = os.Args }
`), cryptoutilSharedMagic.CacheFilePermissions))

	errs := ExportedCheckCmdMainFile(mainPath, cryptoutilSharedMagic.DefaultOTLPServiceDefault, "cryptoutil/internal/apps/cryptoutil", false)
	require.Len(t, errs, 1)
	require.Contains(t, errs[0], "missing import")
}

// TestCheckCmdMainFile_ForbiddenArgsSlice exercises the "must not use os.Args[1:]" violation.
func TestCheckCmdMainFile_ForbiddenArgsSlice(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mainPath := filepath.Join(tmpDir, "main.go")
	require.NoError(t, os.WriteFile(mainPath, []byte(`package main
import "cryptoutil/internal/apps/cryptoutil"
func main() { _ = os.Args[1:] }
`), cryptoutilSharedMagic.CacheFilePermissions))

	// requireArgsSlice=false → file must NOT contain os.Args[1:]
	errs := ExportedCheckCmdMainFile(mainPath, cryptoutilSharedMagic.DefaultOTLPServiceDefault, "cryptoutil/internal/apps/cryptoutil", false)
	require.Len(t, errs, 1)
	require.Contains(t, errs[0], "must use 'os.Args' not 'os.Args[1:]'")
}

// TestCheckCmdMainFile_MissingArgsSlice exercises the "os.Args[1:] required" violation via the exported seam.
func TestCheckCmdMainFile_MissingArgsSlice(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mainPath := filepath.Join(tmpDir, "main.go")
	require.NoError(t, os.WriteFile(mainPath, []byte(`package main
import "cryptoutil/internal/apps/cryptoutil"
func main() {}
`), cryptoutilSharedMagic.CacheFilePermissions))

	// requireArgsSlice=true → file must contain os.Args[1:]
	errs := ExportedCheckCmdMainFile(mainPath, cryptoutilSharedMagic.DefaultOTLPServiceDefault, "cryptoutil/internal/apps/cryptoutil", true)
	require.Len(t, errs, 1)
	require.Contains(t, errs[0], "missing 'os.Args[1:]'")
}

// TestCheckCmdMainFile_Valid exercises the happy path.
func TestCheckCmdMainFile_Valid(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mainPath := filepath.Join(tmpDir, "main.go")
	require.NoError(t, os.WriteFile(mainPath, []byte(validSuiteMain(cryptoutilSharedMagic.DefaultOTLPServiceDefault)), cryptoutilSharedMagic.CacheFilePermissions))

	errs := ExportedCheckCmdMainFile(mainPath, cryptoutilSharedMagic.DefaultOTLPServiceDefault, "cryptoutil/internal/apps/cryptoutil", false)
	require.Empty(t, errs)
}
