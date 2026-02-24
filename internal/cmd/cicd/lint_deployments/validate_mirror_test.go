package lint_deployments

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapDeploymentToConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		deployment string
		want       string
	}{
		{name: "identity service", deployment: "identity-authz", want: "identity"},
		{name: "identity product", deployment: "identity", want: "identity"},
		{name: "sm-im service", deployment: "sm-im", want: "sm"},
		{name: "jose service", deployment: "jose-ja", want: "jose"},
		{name: "jose product", deployment: "jose", want: "jose"},
		{name: "pki explicit mapping", deployment: "pki", want: "ca"},
		{name: "pki-ca explicit mapping", deployment: "pki-ca", want: "ca"},
		{name: "sm explicit mapping", deployment: "sm", want: "sm"},
		{name: "sm-kms explicit mapping", deployment: "sm-kms", want: "sm"},
		{name: "single segment fallback", deployment: "newproduct", want: "newproduct"},
		{name: "product-service fallback", deployment: "newproduct-service", want: "newproduct"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := mapDeploymentToConfig(tc.deployment)
			if got != tc.want {
				t.Errorf("mapDeploymentToConfig(%q) = %q, want %q", tc.deployment, got, tc.want)
			}
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
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(dirs) != 0 {
			t.Errorf("expected 0 subdirectories, got %d", len(dirs))
		}
	})

	t.Run("directories and files", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		createTestDir(t, tmpDir, "subdir1")
		createTestDir(t, tmpDir, "subdir2")
		createTestFile(t, tmpDir, "file.txt", "")

		dirs, err := getSubdirectories(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(dirs) != 2 {
			t.Errorf("expected 2 subdirectories, got %d: %v", len(dirs), dirs)
		}
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
		createTestDir(t, tmpDir, "configs")

		result, err := ValidateStructuralMirror(deploymentsDir, configsDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !result.Valid {
			t.Error("expected valid for empty directories")
		}

		if len(result.MissingMirrors) != 0 {
			t.Errorf("expected 0 missing, got %d", len(result.MissingMirrors))
		}
	})

	t.Run("nonexistent deployments directory", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		configsDir := tmpDir + "/configs"
		createTestDir(t, tmpDir, "configs")

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
		createTestDir(t, tmpDir, "configs")

		// Create excluded deployment directories.
		createTestDir(t, deploymentsDir, "shared-postgres")
		createTestDir(t, deploymentsDir, "shared-citus")
		createTestDir(t, deploymentsDir, "shared-telemetry")
		createTestDir(t, deploymentsDir, "archived")
		createTestDir(t, deploymentsDir, "template")

		result, err := ValidateStructuralMirror(deploymentsDir, configsDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !result.Valid {
			t.Error("expected valid when only excluded dirs exist")
		}

		if len(result.Excluded) != 5 {
			t.Errorf("expected 5 excluded, got %d: %v", len(result.Excluded), result.Excluded)
		}

		if len(result.MissingMirrors) != 0 {
			t.Errorf("expected 0 missing, got %d", len(result.MissingMirrors))
		}
	})

	t.Run("missing config directory detected", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		deploymentsDir := tmpDir + "/deployments"
		configsDir := tmpDir + "/configs"
		createTestDir(t, tmpDir, "deployments")
		createTestDir(t, tmpDir, "configs")

		// Create deployment dir without matching config dir.
		createTestDir(t, deploymentsDir, "sm-im")

		result, err := ValidateStructuralMirror(deploymentsDir, configsDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Valid {
			t.Error("expected invalid when config mirror missing")
		}

		if len(result.MissingMirrors) != 1 {
			t.Fatalf("expected 1 missing, got %d", len(result.MissingMirrors))
		}

		if result.MissingMirrors[0] != "sm-im" {
			t.Errorf("expected missing mirror 'sm-im', got %q", result.MissingMirrors[0])
		}
	})

	t.Run("matching config directory passes", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		deploymentsDir := tmpDir + "/deployments"
		configsDir := tmpDir + "/configs"
		createTestDir(t, tmpDir, "deployments")
		createTestDir(t, tmpDir, "configs")

		// Create deployment and matching config.
		createTestDir(t, deploymentsDir, "sm-im")
		createTestDir(t, configsDir, "sm")

		result, err := ValidateStructuralMirror(deploymentsDir, configsDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !result.Valid {
			t.Errorf("expected valid, got errors: %v, missing: %v", result.Errors, result.MissingMirrors)
		}
	})

	t.Run("orphaned config directory warns", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		deploymentsDir := tmpDir + "/deployments"
		configsDir := tmpDir + "/configs"
		createTestDir(t, tmpDir, "deployments")
		createTestDir(t, tmpDir, "configs")

		// Create config without matching deployment.
		createTestDir(t, configsDir, "orphaned-service")

		result, err := ValidateStructuralMirror(deploymentsDir, configsDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !result.Valid {
			t.Error("orphaned configs should not invalidate result")
		}

		if len(result.Orphans) != 1 {
			t.Fatalf("expected 1 orphan, got %d", len(result.Orphans))
		}

		if result.Orphans[0] != "orphaned-service" {
			t.Errorf("expected orphan 'orphaned-service', got %q", result.Orphans[0])
		}
	})

	t.Run("deduplication of config names", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		deploymentsDir := tmpDir + "/deployments"
		configsDir := tmpDir + "/configs"
		createTestDir(t, tmpDir, "deployments")
		createTestDir(t, tmpDir, "configs")

		// Both identity and identity-authz map to identity config.
		createTestDir(t, deploymentsDir, "identity")
		createTestDir(t, deploymentsDir, "identity-authz")
		createTestDir(t, configsDir, "identity")

		result, err := ValidateStructuralMirror(deploymentsDir, configsDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !result.Valid {
			t.Errorf("expected valid when both product and service map to same config, errors: %v, missing: %v", result.Errors, result.MissingMirrors)
		}

		if len(result.MissingMirrors) != 0 {
			t.Errorf("expected 0 missing, got %d: %v", len(result.MissingMirrors), result.MissingMirrors)
		}
	})

	t.Run("explicit mapping pki to ca", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		deploymentsDir := tmpDir + "/deployments"
		configsDir := tmpDir + "/configs"
		createTestDir(t, tmpDir, "deployments")
		createTestDir(t, tmpDir, "configs")

		createTestDir(t, deploymentsDir, "pki-ca")
		createTestDir(t, configsDir, "ca")

		result, err := ValidateStructuralMirror(deploymentsDir, configsDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !result.Valid {
			t.Errorf("expected valid for pki-ca -> ca mapping, errors: %v, missing: %v", result.Errors, result.MissingMirrors)
		}
	})

	t.Run("warnings for orphaned configs", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		deploymentsDir := tmpDir + "/deployments"
		configsDir := tmpDir + "/configs"
		createTestDir(t, tmpDir, "deployments")
		createTestDir(t, tmpDir, "configs")

		createTestDir(t, configsDir, "orphan1")
		createTestDir(t, configsDir, "orphan2")

		result, err := ValidateStructuralMirror(deploymentsDir, configsDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.Warnings) < 2 {
			t.Errorf("expected at least 2 warnings, got %d", len(result.Warnings))
		}
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

		assert.Contains(t, output, "PASS")
		assert.Contains(t, output, "Excluded (1)")
		assert.Contains(t, output, "shared-postgres")
		assert.NotContains(t, output, "Errors")
		assert.NotContains(t, output, "Warnings")
	})

	t.Run("invalid result with all sections", func(t *testing.T) {
		t.Parallel()

		result := &MirrorResult{
			Valid:          false,
			MissingMirrors: []string{"sm", "jose"},
			Orphans:        []string{"orphan1"},
			Excluded:       []string{"template"},
			Errors:         []string{"some error"},
			Warnings:       []string{"orphaned: orphan1"},
		}

		output := FormatMirrorResult(result)

		assert.Contains(t, output, "FAIL")
		assert.Contains(t, output, "Excluded (1)")
		assert.Contains(t, output, "template")
		assert.Contains(t, output, "Errors (1)")
		assert.Contains(t, output, "some error")
		assert.Contains(t, output, "Warnings (1)")
		assert.Contains(t, output, "orphaned: orphan1")
		assert.Contains(t, output, "missing=2")
		assert.Contains(t, output, "orphans=1")
	})

	t.Run("empty result no optional sections", func(t *testing.T) {
		t.Parallel()

		result := &MirrorResult{
			Valid: true,
		}

		output := FormatMirrorResult(result)

		assert.Contains(t, output, "PASS")
		assert.NotContains(t, output, "Excluded")
		assert.NotContains(t, output, "Errors")
		assert.NotContains(t, output, "Warnings")
	})
}

// TestValidateStructuralMirror_ExcludedWithConfigs tests excluded deployments in orphan check.
func TestValidateStructuralMirror_ExcludedWithConfigs(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	deploymentsDir := filepath.Join(tmpDir, "deployments")
	configsDir := filepath.Join(tmpDir, "configs")

	require.NoError(t, os.MkdirAll(deploymentsDir, 0o750))
	require.NoError(t, os.MkdirAll(configsDir, 0o750))

	// Create excluded deployment + a matching config deployment.
	createTestDir(t, deploymentsDir, "shared-postgres")
	createTestDir(t, deploymentsDir, "jose-ja")
	createTestDir(t, configsDir, "jose")

	// Add an orphan config to trigger the orphan check loop which includes excluded dirs.
	createTestDir(t, configsDir, "orphaned")

	result, err := ValidateStructuralMirror(deploymentsDir, configsDir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
	assert.Len(t, result.Excluded, 1, "shared-postgres should be excluded")
	assert.Len(t, result.Orphans, 1, "orphaned config should be reported")
}

// TestValidateStructuralMirror_MatchedAndOrphaned verifies orphan detection correctly
// distinguishes matched config dirs from unmatched ones.
func TestValidateStructuralMirror_MatchedAndOrphaned(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	deploymentsDir := filepath.Join(tmpDir, "deployments")
	configsDir := filepath.Join(tmpDir, "configs")

	require.NoError(t, os.MkdirAll(deploymentsDir, 0o755))
	require.NoError(t, os.MkdirAll(configsDir, 0o755))

	// Create matched pair: sm-im -> sm.
	require.NoError(t, os.MkdirAll(filepath.Join(deploymentsDir, "sm-im"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(configsDir, "sm"), 0o755))

	// Create orphaned config (no matching deployment).
	require.NoError(t, os.MkdirAll(filepath.Join(configsDir, "orphan-svc"), 0o755))

	result, err := ValidateStructuralMirror(deploymentsDir, configsDir)
	require.NoError(t, err)

	// sm should NOT be in orphans (it matches sm-im).
	assert.NotContains(t, result.Orphans, "sm", "matched config should not be orphaned")
	// orphan-svc should be in orphans.
	assert.Contains(t, result.Orphans, "orphan-svc", "unmatched config should be orphaned")
	assert.Len(t, result.Orphans, 1, "only unmatched config should be orphaned")
}

// TestValidateStructuralMirror_UnreadableDeployments tests error when deployment dirs unreadable.
func TestValidateStructuralMirror_UnreadableDeployments(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	deploymentsDir := filepath.Join(tmpDir, "deployments")
	configsDir := filepath.Join(tmpDir, "configs")

	require.NoError(t, os.MkdirAll(deploymentsDir, 0o750))
	require.NoError(t, os.MkdirAll(configsDir, 0o750))

	// Make deployments dir unreadable.
	require.NoError(t, os.Chmod(deploymentsDir, 0o000))

	t.Cleanup(func() {
		_ = os.Chmod(deploymentsDir, 0o750)
	})

	_, err := ValidateStructuralMirror(deploymentsDir, configsDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list deployment directories")
}

// TestValidateStructuralMirror_UnreadableConfigs tests error when config dirs unreadable.
func TestValidateStructuralMirror_UnreadableConfigs(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	deploymentsDir := filepath.Join(tmpDir, "deployments")
	configsDir := filepath.Join(tmpDir, "configs")

	require.NoError(t, os.MkdirAll(deploymentsDir, 0o750))
	require.NoError(t, os.MkdirAll(configsDir, 0o750))

	// Make configs dir unreadable.
	require.NoError(t, os.Chmod(configsDir, 0o000))

	t.Cleanup(func() {
		_ = os.Chmod(configsDir, 0o750)
	})

	_, err := ValidateStructuralMirror(deploymentsDir, configsDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list config directories")
}
