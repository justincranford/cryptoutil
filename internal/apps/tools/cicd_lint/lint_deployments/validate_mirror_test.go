package lint_deployments

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestMapDeploymentToConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		deployment string
		want       string
	}{
		{name: "identity-authz service", deployment: cryptoutilSharedMagic.OTLPServiceIdentityAuthz, want: cryptoutilSharedMagic.OTLPServiceIdentityAuthz},
		{name: "identity product", deployment: cryptoutilSharedMagic.IdentityProductName, want: cryptoutilSharedMagic.IdentityProductName},
		{name: "sm-im service", deployment: cryptoutilSharedMagic.OTLPServiceSMIM, want: cryptoutilSharedMagic.OTLPServiceSMIM},
		{name: "jose-ja service", deployment: cryptoutilSharedMagic.OTLPServiceJoseJA, want: cryptoutilSharedMagic.OTLPServiceJoseJA},
		{name: "jose product", deployment: cryptoutilSharedMagic.JoseProductName, want: cryptoutilSharedMagic.JoseProductName},
		{name: "pki product", deployment: cryptoutilSharedMagic.PKIProductName, want: cryptoutilSharedMagic.PKIProductName},
		{name: "pki-ca service", deployment: cryptoutilSharedMagic.OTLPServicePKICA, want: cryptoutilSharedMagic.OTLPServicePKICA},
		{name: "sm product", deployment: "sm", want: "sm"},
		{name: "sm-kms service", deployment: cryptoutilSharedMagic.OTLPServiceSMKMS, want: cryptoutilSharedMagic.OTLPServiceSMKMS},
		{name: "suite mapping", deployment: "cryptoutil-suite", want: cryptoutilSharedMagic.DefaultOTLPServiceDefault},
		{name: "single segment identity", deployment: "newproduct", want: "newproduct"},
		{name: "product-service identity", deployment: "newproduct-service", want: "newproduct-service"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := mapDeploymentToConfig(tc.deployment)
			require.Equal(t, tc.want, got, "mapDeploymentToConfig(%q)", tc.deployment)
		})
	}
}

func TestGetSubdirectories(t *testing.T) {
	t.Parallel()

	t.Run("nonexistent directory", func(t *testing.T) {
		t.Parallel()

		dirs, err := getSubdirectories("/nonexistent/path")
		if err == nil {
			t.Error("expected error for nonexistent directory")
		}

		if dirs != nil {
			t.Error("expected nil dirs on error")
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		dirs, err := getSubdirectories(tmpDir)
		require.NoError(t, err)

		require.Empty(t, dirs, "expected 0 subdirectories")
	})

	t.Run("directories and files", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		createTestDir(t, tmpDir, "subdir1")
		createTestDir(t, tmpDir, "subdir2")
		createTestFile(t, tmpDir, "file.txt", "")

		dirs, err := getSubdirectories(tmpDir)
		require.NoError(t, err)

		require.Len(t, dirs, 2)
	})
}

func TestValidateStructuralMirror(t *testing.T) {
	t.Parallel()

	t.Run("both directories empty", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		deploymentsDir := tmpDir + "/deployments"
		configsDir := tmpDir + "/configs"
		createTestDir(t, tmpDir, "deployments")
		createTestDir(t, tmpDir, cryptoutilSharedMagic.CICDConfigsDir)

		result, err := ValidateStructuralMirror(deploymentsDir, configsDir)
		require.NoError(t, err)

		if !result.Valid {
			t.Error("expected valid for empty directories")
		}

		require.Empty(t, result.MissingMirrors)
	})

	t.Run("nonexistent deployments directory", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		configsDir := tmpDir + "/configs"
		createTestDir(t, tmpDir, cryptoutilSharedMagic.CICDConfigsDir)

		_, err := ValidateStructuralMirror("/nonexistent", configsDir)
		if err == nil {
			t.Error("expected error for nonexistent deployments dir")
		}
	})

	t.Run("nonexistent configs directory", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		deploymentsDir := tmpDir + "/deployments"
		createTestDir(t, tmpDir, "deployments")

		_, err := ValidateStructuralMirror(deploymentsDir, "/nonexistent")
		if err == nil {
			t.Error("expected error for nonexistent configs dir")
		}
	})

	t.Run("excluded directories skipped", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		deploymentsDir := tmpDir + "/deployments"
		configsDir := tmpDir + "/configs"
		createTestDir(t, tmpDir, "deployments")
		createTestDir(t, tmpDir, cryptoutilSharedMagic.CICDConfigsDir)

		// Create excluded deployment directories.
		createTestDir(t, deploymentsDir, "shared-postgres")

		createTestDir(t, deploymentsDir, "shared-telemetry")
		createTestDir(t, deploymentsDir, "archived")
		createTestDir(t, deploymentsDir, cryptoutilSharedMagic.SkeletonTemplateServiceName)

		result, err := ValidateStructuralMirror(deploymentsDir, configsDir)
		require.NoError(t, err)

		if !result.Valid {
			t.Error("expected valid when only excluded dirs exist")
		}

		require.Len(t, result.Excluded, cryptoutilSharedMagic.CICDDeploymentMirrorExcludedCount)

		require.Empty(t, result.MissingMirrors)
	})

	t.Run("missing config directory detected", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		deploymentsDir := tmpDir + "/deployments"
		configsDir := tmpDir + "/configs"
		createTestDir(t, tmpDir, "deployments")
		createTestDir(t, tmpDir, cryptoutilSharedMagic.CICDConfigsDir)

		// Create deployment dir without matching config dir.
		createTestDir(t, deploymentsDir, cryptoutilSharedMagic.OTLPServiceSMIM)

		result, err := ValidateStructuralMirror(deploymentsDir, configsDir)
		require.NoError(t, err)

		if result.Valid {
			t.Error("expected invalid when config mirror missing")
		}

		require.Len(t, result.MissingMirrors, 1)

		require.Equal(t, cryptoutilSharedMagic.OTLPServiceSMIM, result.MissingMirrors[0])
	})

	t.Run("matching config directory passes", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		deploymentsDir := tmpDir + "/deployments"
		configsDir := tmpDir + "/configs"
		createTestDir(t, tmpDir, "deployments")
		createTestDir(t, tmpDir, cryptoutilSharedMagic.CICDConfigsDir)

		// Create deployment and matching config.
		createTestDir(t, deploymentsDir, cryptoutilSharedMagic.OTLPServiceSMIM)
		createTestDir(t, configsDir, cryptoutilSharedMagic.OTLPServiceSMIM)

		result, err := ValidateStructuralMirror(deploymentsDir, configsDir)
		require.NoError(t, err)

		require.True(t, result.Valid, "expected valid, got errors: %v, missing: %v", result.Errors, result.MissingMirrors)
	})

	t.Run("orphaned config directory warns", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		deploymentsDir := tmpDir + "/deployments"
		configsDir := tmpDir + "/configs"
		createTestDir(t, tmpDir, "deployments")
		createTestDir(t, tmpDir, cryptoutilSharedMagic.CICDConfigsDir)

		// Create config without matching deployment.
		createTestDir(t, configsDir, "orphaned-service")

		result, err := ValidateStructuralMirror(deploymentsDir, configsDir)
		require.NoError(t, err)

		if !result.Valid {
			t.Error("orphaned configs should not invalidate result")
		}

		require.Len(t, result.Orphans, 1)

		require.Equal(t, "orphaned-service", result.Orphans[0])
	})

	t.Run("product and service both have configs", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		deploymentsDir := tmpDir + "/deployments"
		configsDir := tmpDir + "/configs"
		createTestDir(t, tmpDir, "deployments")
		createTestDir(t, tmpDir, cryptoutilSharedMagic.CICDConfigsDir)

		// Both identity and identity-authz have their own config dirs (flat layout).
		createTestDir(t, deploymentsDir, cryptoutilSharedMagic.IdentityProductName)
		createTestDir(t, deploymentsDir, cryptoutilSharedMagic.OTLPServiceIdentityAuthz)
		createTestDir(t, configsDir, cryptoutilSharedMagic.IdentityProductName)
		createTestDir(t, configsDir, cryptoutilSharedMagic.OTLPServiceIdentityAuthz)

		result, err := ValidateStructuralMirror(deploymentsDir, configsDir)
		require.NoError(t, err)

		require.True(t, result.Valid, "expected valid when both product and service have config dirs, errors: %v, missing: %v", result.Errors, result.MissingMirrors)

		require.Empty(t, result.MissingMirrors)
	})

	t.Run("identity mapping pki-ca to pki-ca", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		deploymentsDir := tmpDir + "/deployments"
		configsDir := tmpDir + "/configs"
		createTestDir(t, tmpDir, "deployments")
		createTestDir(t, tmpDir, cryptoutilSharedMagic.CICDConfigsDir)

		createTestDir(t, deploymentsDir, cryptoutilSharedMagic.OTLPServicePKICA)
		createTestDir(t, configsDir, cryptoutilSharedMagic.OTLPServicePKICA)

		result, err := ValidateStructuralMirror(deploymentsDir, configsDir)
		require.NoError(t, err)

		require.True(t, result.Valid, "expected valid for pki-ca -> pki-ca identity mapping, errors: %v, missing: %v", result.Errors, result.MissingMirrors)
	})

	t.Run("warnings for orphaned configs", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		deploymentsDir := tmpDir + "/deployments"
		configsDir := tmpDir + "/configs"
		createTestDir(t, tmpDir, "deployments")
		createTestDir(t, tmpDir, cryptoutilSharedMagic.CICDConfigsDir)

		createTestDir(t, configsDir, "orphan1")
		createTestDir(t, configsDir, "orphan2")

		result, err := ValidateStructuralMirror(deploymentsDir, configsDir)
		require.NoError(t, err)

		require.GreaterOrEqual(t, len(result.Warnings), 2)
	})
}

func TestFormatMirrorResult(t *testing.T) {
	t.Parallel()

	t.Run("valid result with excluded", func(t *testing.T) {
		t.Parallel()

		result := &MirrorResult{
			Valid:          true,
			MissingMirrors: []string{},
			Orphans:        []string{},
			Excluded:       []string{"shared-postgres"},
		}

		output := FormatMirrorResult(result)

		require.Contains(t, output, cryptoutilSharedMagic.TestStatusPass)
		require.Contains(t, output, "Excluded (1)")
		require.Contains(t, output, "shared-postgres")
		require.NotContains(t, output, "Errors")
		require.NotContains(t, output, "Warnings")
	})

	t.Run("invalid result with all sections", func(t *testing.T) {
		t.Parallel()

		result := &MirrorResult{
			Valid:          false,
			MissingMirrors: []string{"sm", cryptoutilSharedMagic.JoseProductName},
			Orphans:        []string{"orphan1"},
			Excluded:       []string{cryptoutilSharedMagic.SkeletonTemplateServiceName},
			Errors:         []string{"some error"},
			Warnings:       []string{"orphaned: orphan1"},
		}

		output := FormatMirrorResult(result)

		require.Contains(t, output, cryptoutilSharedMagic.TestStatusFail)
		require.Contains(t, output, "Excluded (1)")
		require.Contains(t, output, cryptoutilSharedMagic.SkeletonTemplateServiceName)
		require.Contains(t, output, "Errors (1)")
		require.Contains(t, output, "some error")
		require.Contains(t, output, "Warnings (1)")
		require.Contains(t, output, "orphaned: orphan1")
		require.Contains(t, output, "missing=2")
		require.Contains(t, output, "orphans=1")
	})

	t.Run("empty result no optional sections", func(t *testing.T) {
		t.Parallel()

		result := &MirrorResult{
			Valid: true,
		}

		output := FormatMirrorResult(result)

		require.Contains(t, output, cryptoutilSharedMagic.TestStatusPass)
		require.NotContains(t, output, "Excluded")
		require.NotContains(t, output, "Errors")
		require.NotContains(t, output, "Warnings")
	})
}

// TestValidateStructuralMirror_ExcludedWithConfigs tests excluded deployments in orphan check.
func TestValidateStructuralMirror_ExcludedWithConfigs(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	deploymentsDir := filepath.Join(tmpDir, "deployments")
	configsDir := filepath.Join(tmpDir, cryptoutilSharedMagic.CICDConfigsDir)

	require.NoError(t, os.MkdirAll(deploymentsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))
	require.NoError(t, os.MkdirAll(configsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))

	// Create excluded deployment + a matching config deployment.
	createTestDir(t, deploymentsDir, "shared-postgres")
	createTestDir(t, deploymentsDir, cryptoutilSharedMagic.OTLPServiceJoseJA)
	createTestDir(t, configsDir, cryptoutilSharedMagic.OTLPServiceJoseJA)

	// Add an orphan config to trigger the orphan check loop which includes excluded dirs.
	createTestDir(t, configsDir, "orphaned")

	result, err := ValidateStructuralMirror(deploymentsDir, configsDir)
	require.NoError(t, err)
	require.True(t, result.Valid)
	require.Len(t, result.Excluded, 1, "shared-postgres should be excluded")
	require.Len(t, result.Orphans, 1, "orphaned config should be reported")
}

// TestValidateStructuralMirror_MatchedAndOrphaned verifies orphan detection correctly
// distinguishes matched config dirs from unmatched ones.
func TestValidateStructuralMirror_MatchedAndOrphaned(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	deploymentsDir := filepath.Join(tmpDir, "deployments")
	configsDir := filepath.Join(tmpDir, cryptoutilSharedMagic.CICDConfigsDir)

	require.NoError(t, os.MkdirAll(deploymentsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.MkdirAll(configsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	// Create matched pair: sm-im -> sm-im (1:1 identity mapping).
	require.NoError(t, os.MkdirAll(filepath.Join(deploymentsDir, cryptoutilSharedMagic.OTLPServiceSMIM), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.MkdirAll(filepath.Join(configsDir, cryptoutilSharedMagic.OTLPServiceSMIM), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	// Create orphaned config (no matching deployment).
	require.NoError(t, os.MkdirAll(filepath.Join(configsDir, "orphan-svc"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	result, err := ValidateStructuralMirror(deploymentsDir, configsDir)
	require.NoError(t, err)

	// sm-im should NOT be in orphans (it matches deployment sm-im).
	require.NotContains(t, result.Orphans, cryptoutilSharedMagic.OTLPServiceSMIM, "matched config should not be orphaned")
	// orphan-svc should be in orphans.
	require.Contains(t, result.Orphans, "orphan-svc", "unmatched config should be orphaned")
	require.Len(t, result.Orphans, 1, "only unmatched config should be orphaned")
}

// TestValidateStructuralMirror_UnreadableDeployments tests error when deployment dirs unreadable.
func TestValidateStructuralMirror_UnreadableDeployments(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == cryptoutilSharedMagic.OSNameWindows {
		t.Skip("os.Chmod 0o000 does not restrict access on Windows NTFS")
	}

	tmpDir := t.TempDir()
	deploymentsDir := filepath.Join(tmpDir, "deployments")
	configsDir := filepath.Join(tmpDir, cryptoutilSharedMagic.CICDConfigsDir)

	require.NoError(t, os.MkdirAll(deploymentsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))
	require.NoError(t, os.MkdirAll(configsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))

	// Make deployments dir unreadable.
	require.NoError(t, os.Chmod(deploymentsDir, 0o000))

	t.Cleanup(func() {
		_ = os.Chmod(deploymentsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute)
	})

	_, err := ValidateStructuralMirror(deploymentsDir, configsDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to list deployment directories")
}

// TestValidateStructuralMirror_UnreadableConfigs tests error when config dirs unreadable.
func TestValidateStructuralMirror_UnreadableConfigs(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == cryptoutilSharedMagic.OSNameWindows {
		t.Skip("os.Chmod 0o000 does not restrict access on Windows NTFS")
	}

	tmpDir := t.TempDir()
	deploymentsDir := filepath.Join(tmpDir, "deployments")
	configsDir := filepath.Join(tmpDir, cryptoutilSharedMagic.CICDConfigsDir)

	require.NoError(t, os.MkdirAll(deploymentsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))
	require.NoError(t, os.MkdirAll(configsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))

	// Make configs dir unreadable.
	require.NoError(t, os.Chmod(configsDir, 0o000))

	t.Cleanup(func() {
		_ = os.Chmod(configsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute)
	})

	_, err := ValidateStructuralMirror(deploymentsDir, configsDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to list config directories")
}
