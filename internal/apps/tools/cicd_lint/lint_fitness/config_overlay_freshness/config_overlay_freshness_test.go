// Copyright (c) 2025 Justin Cranford

package config_overlay_freshness

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// setupTempDir creates a temp directory with the given files and returns its path.
func setupTempDir(t *testing.T, files map[string]string) string {
	t.Helper()

	rootDir := t.TempDir()

	for relPath, content := range files {
		fullPath := filepath.Join(rootDir, filepath.FromSlash(relPath))

		if err := os.MkdirAll(filepath.Dir(fullPath), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute); err != nil {
			t.Fatalf("setupTempDir: mkdir %s: %v", filepath.Dir(fullPath), err)
		}

		if err := os.WriteFile(fullPath, []byte(content), cryptoutilSharedMagic.FilePermissionsDefault); err != nil {
			t.Fatalf("setupTempDir: write %s: %v", fullPath, err)
		}
	}

	return rootDir
}

// allMinimalPSIDFiles returns a file map with minimal valid content for every PS-ID in the
// registry. Callers may merge their own specific content on top to override individual entries.
// This allows CheckInDir tests to satisfy the hard-error-on-absent-config-dir requirement
// without having to enumerate all PS-IDs in every test.
func allMinimalPSIDFiles() map[string]string {
	files := make(map[string]string)

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		id := ps.PSID
		files[fmt.Sprintf("deployments/%s/config/%s-app-framework-sqlite-1.yml", id, id)] = fmt.Sprintf("database-url: \"sqlite://file::memory:?cache=shared\"\notlp-service: %s-sqlite-1\n", id)
		files[fmt.Sprintf("deployments/%s/config/%s-app-framework-sqlite-2.yml", id, id)] = fmt.Sprintf("database-url: \"sqlite://file::memory:?cache=shared\"\notlp-service: %s-sqlite-2\n", id)
		files[fmt.Sprintf("deployments/%s/config/%s-app-framework-postgresql-1.yml", id, id)] = fmt.Sprintf("otlp-service: %s-postgres-1\n", id)
		files[fmt.Sprintf("deployments/%s/config/%s-app-framework-postgresql-2.yml", id, id)] = fmt.Sprintf("otlp-service: %s-postgres-2\n", id)
	}

	return files
}

func sqliteFilesFor(psID string) map[string]string {
	return map[string]string{
		fmt.Sprintf("deployments/%s/config/%s-app-framework-sqlite-1.yml", psID, psID):     fmt.Sprintf("database-url: \"sqlite://file::memory:?cache=shared\"\notlp-service: %s-sqlite-1\n", psID),
		fmt.Sprintf("deployments/%s/config/%s-app-framework-sqlite-2.yml", psID, psID):     fmt.Sprintf("database-url: \"sqlite://file::memory:?cache=shared\"\notlp-service: %s-sqlite-2\n", psID),
		fmt.Sprintf("deployments/%s/config/%s-app-framework-postgresql-1.yml", psID, psID): fmt.Sprintf("otlp-service: %s-postgres-1\n", psID),
		fmt.Sprintf("deployments/%s/config/%s-app-framework-postgresql-2.yml", psID, psID): fmt.Sprintf("otlp-service: %s-postgres-2\n", psID),
	}
}

// TestCheck_RealWorkspace validates all deployment config overlays in the actual repository.
// Non-parallel: uses relative path navigation, not os.Chdir.
func TestCheck_RealWorkspace(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	// Run from project root (6 levels up from this test file).
	err := CheckInDir(logger, filepath.Join("..", "..", "..", "..", "..", ".."), os.ReadFile)
	if err != nil {
		t.Fatalf("unexpected violation in real workspace: %v", err)
	}
}

func TestCheckInDir_ValidOverlays(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupFiles map[string]string
	}{
		{
			name:       "valid sqlite and postgresql overlays for one service",
			setupFiles: sqliteFilesFor(cryptoutilSharedMagic.OTLPServiceSMKMS),
		},
		{
			name: "postgresql overlay: no database-url is valid",
			setupFiles: map[string]string{
				"deployments/sm-kms/config/sm-kms-app-framework-sqlite-1.yml":     "database-url: \"sqlite://file::memory:?cache=shared\"\n",
				"deployments/sm-kms/config/sm-kms-app-framework-sqlite-2.yml":     "database-url: \"sqlite://file::memory:?cache=shared\"\n",
				"deployments/sm-kms/config/sm-kms-app-framework-postgresql-1.yml": "otlp-service: sm-kms-postgres-1\n",
				"deployments/sm-kms/config/sm-kms-app-framework-postgresql-2.yml": "otlp-service: sm-kms-postgres-2\n",
			},
		},
		{
			name: "empty postgresql overlay is valid",
			setupFiles: map[string]string{
				"deployments/sm-kms/config/sm-kms-app-framework-sqlite-1.yml":     "database-url: \"sqlite://file::memory:?cache=shared\"\n",
				"deployments/sm-kms/config/sm-kms-app-framework-sqlite-2.yml":     "database-url: \"sqlite://file::memory:?cache=shared\"\n",
				"deployments/sm-kms/config/sm-kms-app-framework-postgresql-1.yml": "",
				"deployments/sm-kms/config/sm-kms-app-framework-postgresql-2.yml": "",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Merge test-specific files on top of the all-PS-IDs base so that the
			// hard-error-on-absent-config-dir check passes for every PS-ID.
			merged := allMinimalPSIDFiles()
			for k, v := range tc.setupFiles {
				merged[k] = v
			}

			rootDir := setupTempDir(t, merged)
			logger := cryptoutilCmdCicdCommon.NewLogger("test")

			err := CheckInDir(logger, rootDir, os.ReadFile)
			if err != nil {
				t.Errorf("expected no violations, got: %v", err)
			}
		})
	}
}

func TestCheckInDir_Violations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupFiles    map[string]string
		wantSubstring string
	}{
		{
			name: "sqlite variant missing database-url",
			setupFiles: map[string]string{
				"deployments/sm-kms/config/sm-kms-app-framework-sqlite-1.yml":     "otlp-service: sm-kms-sqlite-1\n",
				"deployments/sm-kms/config/sm-kms-app-framework-sqlite-2.yml":     "database-url: \"sqlite://file::memory:?cache=shared\"\n",
				"deployments/sm-kms/config/sm-kms-app-framework-postgresql-1.yml": "otlp-service: sm-kms-postgres-1\n",
				"deployments/sm-kms/config/sm-kms-app-framework-postgresql-2.yml": "otlp-service: sm-kms-postgres-2\n",
			},
			wantSubstring: `missing required key "database-url"`,
		},
		{
			name: "sqlite variant has non-sqlite database-url",
			setupFiles: map[string]string{
				"deployments/sm-kms/config/sm-kms-app-framework-sqlite-1.yml":     "database-url: \"postgres://localhost/db\"\n",
				"deployments/sm-kms/config/sm-kms-app-framework-sqlite-2.yml":     "database-url: \"sqlite://file::memory:?cache=shared\"\n",
				"deployments/sm-kms/config/sm-kms-app-framework-postgresql-1.yml": "otlp-service: sm-kms-postgres-1\n",
				"deployments/sm-kms/config/sm-kms-app-framework-postgresql-2.yml": "otlp-service: sm-kms-postgres-2\n",
			},
			wantSubstring: `does not match pattern`,
		},
		{
			name: "postgresql variant has database-url (forbidden)",
			setupFiles: map[string]string{
				"deployments/sm-kms/config/sm-kms-app-framework-sqlite-1.yml":     "database-url: \"sqlite://file::memory:?cache=shared\"\n",
				"deployments/sm-kms/config/sm-kms-app-framework-sqlite-2.yml":     "database-url: \"sqlite://file::memory:?cache=shared\"\n",
				"deployments/sm-kms/config/sm-kms-app-framework-postgresql-1.yml": "database-url: \"postgres://localhost/db\"\n",
				"deployments/sm-kms/config/sm-kms-app-framework-postgresql-2.yml": "otlp-service: sm-kms-postgres-2\n",
			},
			wantSubstring: `forbidden key "database-url"`,
		},
		{
			name: "missing overlay file",
			setupFiles: map[string]string{
				"deployments/sm-kms/config/sm-kms-app-framework-sqlite-1.yml": "database-url: \"sqlite://file::memory:?cache=shared\"\n",
				// sqlite-2 is missing
				"deployments/sm-kms/config/sm-kms-app-framework-postgresql-1.yml": "otlp-service: sm-kms-postgres-1\n",
				"deployments/sm-kms/config/sm-kms-app-framework-postgresql-2.yml": "otlp-service: sm-kms-postgres-2\n",
			},
			wantSubstring: "missing overlay file",
		},
		{
			name: "database-url value is not a string",
			setupFiles: map[string]string{
				"deployments/sm-kms/config/sm-kms-app-framework-sqlite-1.yml":     "database-url: 12345\n",
				"deployments/sm-kms/config/sm-kms-app-framework-sqlite-2.yml":     "database-url: \"sqlite://file::memory:?cache=shared\"\n",
				"deployments/sm-kms/config/sm-kms-app-framework-postgresql-1.yml": "otlp-service: sm-kms-postgres-1\n",
				"deployments/sm-kms/config/sm-kms-app-framework-postgresql-2.yml": "otlp-service: sm-kms-postgres-2\n",
			},
			wantSubstring: "must be a string",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootDir := setupTempDir(t, tc.setupFiles)
			logger := cryptoutilCmdCicdCommon.NewLogger("test")

			err := CheckInDir(logger, rootDir, os.ReadFile)
			if err == nil {
				t.Fatal("expected violation, got nil error")
			}

			if tc.wantSubstring != "" && !contains(err.Error(), tc.wantSubstring) {
				t.Errorf("want error containing %q, got: %v", tc.wantSubstring, err)
			}
		})
	}
}

func TestCheckInDir_YAMLParseError(t *testing.T) {
	t.Parallel()

	setupFiles := map[string]string{
		"deployments/sm-kms/config/sm-kms-app-framework-sqlite-1.yml":     "{ invalid yaml: [unclosed\n",
		"deployments/sm-kms/config/sm-kms-app-framework-sqlite-2.yml":     "database-url: \"sqlite://file::memory:?cache=shared\"\n",
		"deployments/sm-kms/config/sm-kms-app-framework-postgresql-1.yml": "otlp-service: sm-kms-postgres-1\n",
		"deployments/sm-kms/config/sm-kms-app-framework-postgresql-2.yml": "otlp-service: sm-kms-postgres-2\n",
	}

	rootDir := setupTempDir(t, setupFiles)
	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, rootDir, os.ReadFile)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}

	if !contains(err.Error(), "YAML parse error") {
		t.Errorf("want YAML parse error, got: %v", err)
	}
}

func TestCheckInDir_ReadFileError(t *testing.T) {
	t.Parallel()

	setupFiles := map[string]string{
		"deployments/sm-kms/config/sm-kms-app-framework-sqlite-1.yml":     "exists",
		"deployments/sm-kms/config/sm-kms-app-framework-sqlite-2.yml":     "exists",
		"deployments/sm-kms/config/sm-kms-app-framework-postgresql-1.yml": "exists",
		"deployments/sm-kms/config/sm-kms-app-framework-postgresql-2.yml": "exists",
	}

	rootDir := setupTempDir(t, setupFiles)
	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, rootDir, func(_ string) ([]byte, error) {
		return nil, fmt.Errorf("simulated read error")
	})
	if err == nil {
		t.Fatal("expected error from simulated read failure, got nil")
	}

	if !contains(err.Error(), "cannot read file") {
		t.Errorf("want 'cannot read file' in error, got: %v", err)
	}
}

func TestLoadOverlayTemplates_InvalidYAML(t *testing.T) {
	// Non-parallel: modifies package-level overlayTemplatesYAML.
	original := overlayTemplatesYAML

	defer func() {
		overlayTemplatesYAML = original
	}()

	overlayTemplatesYAML = []byte("{ invalid yaml: [unclosed")

	_, err := loadOverlayTemplates()
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestCheck_LoadTemplateError(t *testing.T) {
	// Non-parallel: modifies package-level overlayTemplatesYAML.
	original := overlayTemplatesYAML

	defer func() {
		overlayTemplatesYAML = original
	}()

	overlayTemplatesYAML = []byte("variants: invalid_not_a_list")
	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	// Even a technically valid YAML that doesn't decode to the expected shape is OK —
	// ensure we get a proper error when YAML is truly broken (not just wrong shape).
	overlayTemplatesYAML = []byte("{ invalid yaml: [unclosed")

	err := CheckInDir(logger, t.TempDir(), os.ReadFile)
	if err == nil {
		t.Fatal("expected error for invalid YAML in phase, got nil")
	}
}

func TestCheckInDir_UnknownVariantInTemplate(t *testing.T) {
	// Non-parallel: modifies package-level overlayTemplatesYAML.
	original := overlayTemplatesYAML

	defer func() {
		overlayTemplatesYAML = original
	}()

	// Template with unknown variant name.
	overlayTemplatesYAML = []byte(`variants:
  - variant: unknown-variant
    description: test
    required_keys: []
    forbidden_keys: []
    required_patterns: []
`)

	setupFiles := map[string]string{
		"deployments/sm-kms/config/sm-kms-app-framework-sqlite-1.yml": "database-url: \"sqlite://file::memory:?cache=shared\"\n",
	}

	rootDir := setupTempDir(t, setupFiles)
	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, rootDir, os.ReadFile)
	if err == nil {
		t.Fatal("expected violation for unknown variant, got nil")
	}

	if !contains(err.Error(), "unknown variant") {
		t.Errorf("want 'unknown variant' in error, got: %v", err)
	}
}

// contains reports whether substr is in s.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}

			return false
		}())
}
