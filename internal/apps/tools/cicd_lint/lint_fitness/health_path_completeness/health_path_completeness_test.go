// Copyright (c) 2025 Justin Cranford

package health_path_completeness

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilLintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// testLogger returns a no-op logger for testing.
func testLogger() *cryptoutilCmdCicdCommon.Logger {
	return cryptoutilCmdCicdCommon.NewLogger("test")
}

// allHealthPaths returns a copy of the required health paths for test use.
func allHealthPaths() []string {
	return []string{
		cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath + "/health",
		cryptoutilSharedMagic.DefaultPublicBrowserAPIContextPath + "/health",
		cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminLivezRequestPath,
		cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminReadyzRequestPath,
		cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminShutdownRequestPath,
	}
}

// setupServiceDir creates a tmp service directory with a Go file containing all required paths.
func setupServiceDir(t *testing.T, tmpDir, psID string, paths []string) {
	t.Helper()

	svcDir := filepath.Join(tmpDir, "internal", "apps", psID)

	err := os.MkdirAll(svcDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute)
	require.NoError(t, err)

	content := "package " + strings.ReplaceAll(psID, "-", "") + "\n\nconst usageText = `\n"
	for _, p := range paths {
		content += "  " + p + "\n"
	}

	content += "`\n"

	err = os.WriteFile(filepath.Join(svcDir, psID+"_usage.go"), []byte(content), cryptoutilSharedMagic.FilePermissionsDefault)
	require.NoError(t, err)
}

// setupAllValidServiceDirs creates tmp service dirs with all required health paths for all non-skeleton PS-IDs.
func setupAllValidServiceDirs(t *testing.T, tmpDir string) {
	t.Helper()

	for _, ps := range cryptoutilLintFitnessRegistry.AllProductServices() {
		if ps.PSID == cryptoutilSharedMagic.SkeletonTemplateServiceID {
			continue
		}

		setupServiceDir(t, tmpDir, ps.InternalAppsDir[:len(ps.InternalAppsDir)-1], allHealthPaths())
	}
}

// TestCheck_DelegatesToCheckInDir verifies Check() delegates to CheckInDir with workspace root.
// Not parallel because it needs to chdir.
func TestCheck_DelegatesToCheckInDir(t *testing.T) {
	tmpDir := t.TempDir()
	setupAllValidServiceDirs(t, tmpDir)

	origDir, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	defer func() {
		_ = os.Chdir(origDir)
	}()

	err = Check(testLogger())
	require.NoError(t, err)
}

// TestCheckInDir_RealWorkspace verifies the check passes on the actual workspace.
func TestCheckInDir_RealWorkspace(t *testing.T) {
	t.Parallel()

	workspaceRoot := filepath.Join("..", "..", "..", "..", "..", "..")

	err := CheckInDir(testLogger(), workspaceRoot, os.ReadDir, os.ReadFile)
	require.NoError(t, err)
}

// TestCheckInDir_AllPathsPresent verifies that a service with all required paths passes.
func TestCheckInDir_AllPathsPresent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupAllValidServiceDirs(t, tmpDir)

	err := CheckInDir(testLogger(), tmpDir, os.ReadDir, os.ReadFile)
	require.NoError(t, err)
}

// TestCheckInDir_MissingPath tests each missing path generates the correct violation.
func TestCheckInDir_MissingPath(t *testing.T) {
	t.Parallel()

	paths := allHealthPaths()

	tests := []struct {
		name        string
		missingPath string
	}{
		{"missing_service_health", paths[0]},
		{"missing_browser_health", paths[1]},
		{"missing_livez", paths[2]},
		{"missing_readyz", paths[3]},
		{"missing_shutdown", paths[4]},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()

			// Set up all services with all paths except the one being tested for sm-kms.
			presentPaths := make([]string, 0, len(paths)-1)
			for _, p := range paths {
				if p != tt.missingPath {
					presentPaths = append(presentPaths, p)
				}
			}

			// Set up sm-kms with missing path; all others with all paths.
			for _, ps := range cryptoutilLintFitnessRegistry.AllProductServices() {
				if ps.PSID == cryptoutilSharedMagic.SkeletonTemplateServiceID {
					continue
				}

				psDir := ps.InternalAppsDir[:len(ps.InternalAppsDir)-1]
				if ps.PSID == cryptoutilSharedMagic.OTLPServiceSMKMS {
					setupServiceDir(t, tmpDir, psDir, presentPaths)
				} else {
					setupServiceDir(t, tmpDir, psDir, allHealthPaths())
				}
			}

			err := CheckInDir(testLogger(), tmpDir, os.ReadDir, os.ReadFile)
			require.Error(t, err)
			require.Contains(t, err.Error(), "1 health path completeness violations")
		})
	}
}

// TestCheckInDir_WrongPath verifies wrong path string (wrong prefix) triggers violation.
func TestCheckInDir_WrongPath(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Provide wrong paths (like original sm-im had).
	wrongPaths := []string{
		"/health",            // wrong: should be /service/api/v1/health or /browser/api/v1/health
		"/admin/v1/livez",    // wrong: should be /admin/api/v1/livez
		"/admin/v1/readyz",   // wrong: should be /admin/api/v1/readyz
		"/admin/v1/shutdown", // wrong: should be /admin/api/v1/shutdown
	}

	for _, ps := range cryptoutilLintFitnessRegistry.AllProductServices() {
		if ps.PSID == cryptoutilSharedMagic.SkeletonTemplateServiceID {
			continue
		}

		psDir := ps.InternalAppsDir[:len(ps.InternalAppsDir)-1]
		if ps.PSID == cryptoutilSharedMagic.OTLPServiceSMKMS {
			setupServiceDir(t, tmpDir, psDir, wrongPaths)
		} else {
			setupServiceDir(t, tmpDir, psDir, allHealthPaths())
		}
	}

	err := CheckInDir(testLogger(), tmpDir, os.ReadDir, os.ReadFile)
	require.Error(t, err)
	require.Contains(t, err.Error(), "health path completeness violations")
}

// TestCheckInDir_EmptyServiceDir verifies empty service directory produces violations.
func TestCheckInDir_EmptyServiceDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	for _, ps := range cryptoutilLintFitnessRegistry.AllProductServices() {
		if ps.PSID == cryptoutilSharedMagic.SkeletonTemplateServiceID {
			continue
		}

		psDir := ps.InternalAppsDir[:len(ps.InternalAppsDir)-1]
		if ps.PSID == cryptoutilSharedMagic.OTLPServiceSMKMS {
			// Create empty directory.
			svcDir := filepath.Join(tmpDir, "internal", "apps", psDir)
			err := os.MkdirAll(svcDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute)
			require.NoError(t, err)
		} else {
			setupServiceDir(t, tmpDir, psDir, allHealthPaths())
		}
	}

	err := CheckInDir(testLogger(), tmpDir, os.ReadDir, os.ReadFile)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("%d health path completeness violations", len(allHealthPaths())))
}

// TestCheckInDir_ServiceDirMissing verifies missing service directory returns error.
func TestCheckInDir_ServiceDirMissing(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Only set up some services — sm-kms dir does not exist.
	for _, ps := range cryptoutilLintFitnessRegistry.AllProductServices() {
		if ps.PSID == cryptoutilSharedMagic.SkeletonTemplateServiceID {
			continue
		}

		if ps.PSID == cryptoutilSharedMagic.OTLPServiceSMKMS {
			continue // sm-kms dir intentionally missing.
		}

		psDir := ps.InternalAppsDir[:len(ps.InternalAppsDir)-1]
		setupServiceDir(t, tmpDir, psDir, allHealthPaths())
	}

	err := CheckInDir(testLogger(), tmpDir, os.ReadDir, os.ReadFile)
	require.Error(t, err)
	require.Contains(t, err.Error(), "checking service")
}

// TestCheckInDir_SkeletonTemplateSkipped verifies skeleton-template is excluded from checks.
func TestCheckInDir_SkeletonTemplateSkipped(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Set up all services except skeleton-template.
	for _, ps := range cryptoutilLintFitnessRegistry.AllProductServices() {
		if ps.PSID == cryptoutilSharedMagic.SkeletonTemplateServiceID {
			continue
		}

		psDir := ps.InternalAppsDir[:len(ps.InternalAppsDir)-1]
		setupServiceDir(t, tmpDir, psDir, allHealthPaths())
	}

	// Do NOT create skeleton-template dir — check should pass anyway.
	err := CheckInDir(testLogger(), tmpDir, os.ReadDir, os.ReadFile)
	require.NoError(t, err)
}

// TestCheckInDir_InvalidGoFile verifies non-Go files in service dir are skipped.
func TestCheckInDir_NonGoFilesSkipped(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	for _, ps := range cryptoutilLintFitnessRegistry.AllProductServices() {
		if ps.PSID == cryptoutilSharedMagic.SkeletonTemplateServiceID {
			continue
		}

		psDir := ps.InternalAppsDir[:len(ps.InternalAppsDir)-1]
		if ps.PSID == cryptoutilSharedMagic.OTLPServiceSMKMS {
			svcDir := filepath.Join(tmpDir, "internal", "apps", psDir)
			err := os.MkdirAll(svcDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute)
			require.NoError(t, err)

			// Write all paths to a non-Go file — should NOT be scanned.
			content := strings.Join(allHealthPaths(), "\n")
			err = os.WriteFile(filepath.Join(svcDir, "README.md"), []byte(content), cryptoutilSharedMagic.FilePermissionsDefault)
			require.NoError(t, err)

			// Write a .go file WITHOUT the paths.
			err = os.WriteFile(filepath.Join(svcDir, "empty.go"), []byte("package smkms\n"), cryptoutilSharedMagic.FilePermissionsDefault)
			require.NoError(t, err)
		} else {
			setupServiceDir(t, tmpDir, psDir, allHealthPaths())
		}
	}

	err := CheckInDir(testLogger(), tmpDir, os.ReadDir, os.ReadFile)
	require.Error(t, err, "non-Go file with paths should not satisfy the requirement")
}

// TestCheckInDir_ReadDirError verifies ReadDir error returns wrapped error.
func TestCheckInDir_ReadDirError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupAllValidServiceDirs(t, tmpDir)

	err := CheckInDir(testLogger(), tmpDir, func(name string) ([]os.DirEntry, error) {
		if strings.HasSuffix(filepath.ToSlash(name), "/"+cryptoutilSharedMagic.OTLPServiceSMKMS) {
			return nil, fmt.Errorf("injected ReadDir error")
		}

		return os.ReadDir(name)
	}, os.ReadFile)
	require.Error(t, err)
	require.Contains(t, err.Error(), "injected ReadDir error")
}

// TestCheckInDir_ReadFileError verifies ReadFile error returns wrapped error.
func TestCheckInDir_ReadFileError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupAllValidServiceDirs(t, tmpDir)

	err := CheckInDir(testLogger(), tmpDir, os.ReadDir, func(name string) ([]byte, error) {
		if filepath.Base(name) == "sm-kms_usage.go" {
			return nil, fmt.Errorf("injected ReadFile error")
		}

		return os.ReadFile(name)
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "injected ReadFile error")
}
