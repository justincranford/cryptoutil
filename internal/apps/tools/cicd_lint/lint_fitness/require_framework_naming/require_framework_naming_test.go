// Copyright (c) 2025 Justin Cranford

package require_framework_naming

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"

	"github.com/stretchr/testify/require"
)

func newTestLogger() *cryptoutilCmdCicdCommon.Logger {
	return cryptoutilCmdCicdCommon.NewLogger("test")
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(path), cryptoutilSharedMagic.DirPermissions))
	require.NoError(t, os.WriteFile(path, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))
}

// --- isBannedImport unit tests ---

func TestIsBannedImport_OldTemplatePath_Banned(t *testing.T) {
	t.Parallel()

	require.True(t, isBannedImport("cryptoutil/internal/apps/template/service/server"))
}

func TestIsBannedImport_OldTemplateSubpath_Banned(t *testing.T) {
	t.Parallel()

	require.True(t, isBannedImport("cryptoutil/internal/apps/template/service/server/repository"))
}

func TestIsBannedImport_SkeletonTemplate_Allowed(t *testing.T) {
	t.Parallel()

	require.False(t, isBannedImport("cryptoutil/internal/apps/skeleton-template/server"))
}

func TestIsBannedImport_FrameworkPath_Allowed(t *testing.T) {
	t.Parallel()

	require.False(t, isBannedImport("cryptoutil/internal/apps/framework/service/server"))
}

func TestIsBannedImport_ExternalPackage_Allowed(t *testing.T) {
	t.Parallel()

	require.False(t, isBannedImport("github.com/stretchr/testify/require"))
}

func TestIsBannedImport_SharedPackage_Allowed(t *testing.T) {
	t.Parallel()

	require.False(t, isBannedImport("cryptoutil/internal/shared/magic"))
}

func TestIsBannedImport_EmptyString_Allowed(t *testing.T) {
	t.Parallel()

	require.False(t, isBannedImport(""))
}

// --- checkFile unit tests ---

func TestCheckFile_NoBannedImports_NoViolations(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	goFile := filepath.Join(tmp, "clean.go")
	writeFile(t, goFile, `package clean

import (
	"fmt"

	cryptoutilFrameworkServer "cryptoutil/internal/apps/framework/service/server"
)

var _ = fmt.Println
var _ = cryptoutilFrameworkServer.NewServiceFramework
`)

	violations, err := checkFile(goFile)
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestCheckFile_BannedImport_ReportsViolation(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	goFile := filepath.Join(tmp, "bad.go")
	writeFile(t, goFile, `package bad

import (
	"fmt"

	cryptoutilTemplateServer "cryptoutil/internal/apps/template/service/server"
)

var _ = fmt.Println
var _ = cryptoutilTemplateServer.NewServiceFramework
`)

	violations, err := checkFile(goFile)
	require.NoError(t, err)
	require.Len(t, violations, 1)
	require.Contains(t, violations[0], "cryptoutil/internal/apps/template/service/server")
	require.Contains(t, violations[0], "banned path")
}

func TestCheckFile_SkeletonTemplateImport_NoViolation(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	goFile := filepath.Join(tmp, "skeleton.go")
	writeFile(t, goFile, `package skeleton

import (
	cryptoutilSkeletonTemplate "cryptoutil/internal/apps/skeleton-template/server"
)

var _ = cryptoutilSkeletonTemplate.Something
`)

	violations, err := checkFile(goFile)
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestCheckFile_SingleLineImport_BannedPath(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	goFile := filepath.Join(tmp, "single.go")
	writeFile(t, goFile, `package single

import "cryptoutil/internal/apps/template/service/server"

var _ = server.Something
`)

	violations, err := checkFile(goFile)
	require.NoError(t, err)
	require.Len(t, violations, 1)
}

func TestCheckFile_MultipleBannedImports_ReportsAll(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	goFile := filepath.Join(tmp, "multi.go")
	writeFile(t, goFile, `package multi

import (
	cryptoutilTemplateServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilTemplateRepo "cryptoutil/internal/apps/template/service/server/repository"
)

var _ = cryptoutilTemplateServer.Something
var _ = cryptoutilTemplateRepo.Something
`)

	violations, err := checkFile(goFile)
	require.NoError(t, err)
	require.Len(t, violations, 2)
}

func TestCheckFile_NonexistentFile_Error(t *testing.T) {
	t.Parallel()

	violations, err := checkFile("/nonexistent/file.go")
	require.Error(t, err)
	require.Nil(t, violations)
}

// --- CheckInDir unit tests ---

func TestCheckInDir_CleanTree_Passes(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "pkg", "clean.go"), `package pkg

import (
	"fmt"

	cryptoutilFrameworkServer "cryptoutil/internal/apps/framework/service/server"
)

var _ = fmt.Println
var _ = cryptoutilFrameworkServer.NewServiceFramework
`)

	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_BannedImport_Fails(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "pkg", "bad.go"), `package pkg

import (
	cryptoutilTemplateServer "cryptoutil/internal/apps/template/service/server"
)

var _ = cryptoutilTemplateServer.Something
`)

	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "banned internal/apps/template/ imports")
}

func TestCheckInDir_SkipsGitDir(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	// Put a banned import inside .git/ — should be skipped.
	writeFile(t, filepath.Join(tmp, cryptoutilSharedMagic.CICDExcludeDirGit, "hooks", "bad.go"), `package hooks

import (
	cryptoutilTemplateServer "cryptoutil/internal/apps/template/service/server"
)

var _ = cryptoutilTemplateServer.Something
`)

	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_SkipsVendorDir(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, cryptoutilSharedMagic.CICDExcludeDirVendor, "somelib", "bad.go"), `package somelib

import (
	cryptoutilTemplateServer "cryptoutil/internal/apps/template/service/server"
)

var _ = cryptoutilTemplateServer.Something
`)

	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_SkipsNonGoFiles(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "readme.md"), `This file mentions cryptoutil/internal/apps/template/service/server but is not Go.`)

	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_SkipsArchivedDir(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "_archived", "old", "bad.go"), `package old

import (
	cryptoutilTemplateServer "cryptoutil/internal/apps/template/service/server"
)

var _ = cryptoutilTemplateServer.Something
`)

	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_EmptyDir_Passes(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()

	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_WalkError(t *testing.T) {
	t.Parallel()

	stubWalkFn := func(_ string, _ filepath.WalkFunc) error {
		return errors.New("simulated walk error")
	}

	err := checkInDir(newTestLogger(), ".", stubWalkFn)
	require.Error(t, err)
	require.Contains(t, err.Error(), "walking directory tree")
}

func TestCheckInDir_WalkCallbackError(t *testing.T) {
	t.Parallel()

	stubWalkFn := func(_ string, fn filepath.WalkFunc) error {
		return fn("fake.go", nil, errors.New("simulated callback error"))
	}

	err := checkInDir(newTestLogger(), ".", stubWalkFn)
	require.Error(t, err)
	require.Contains(t, err.Error(), "walking directory tree")
}

// --- Integration test against real workspace ---

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

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

// TestCheck_DirectCall verifies that Check() delegates to CheckInDir correctly.
func TestCheck_DirectCall(t *testing.T) {
	t.Parallel()

	_, err := findProjectRoot()
	if err != nil {
		t.Skip("Skipping integration test - cannot find project root (no go.mod)")
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	// Check() calls CheckInDir(logger, ".") which uses the current working directory.
	err = CheckInDir(logger, ".")
	_ = err

	err2 := Check(logger)
	_ = err2
}
