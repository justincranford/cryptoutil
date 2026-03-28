// Copyright (c) 2025 Justin Cranford

package unseal_secret_content

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestCheck_RealWorkspace(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	// Run from project root (6 levels up from this test file).
	err := CheckInDir(logger, filepath.Join("..", "..", "..", "..", "..", ".."))
	if err != nil {
		t.Fatalf("unexpected violation in real workspace: %v", err)
	}
}

func TestFindViolationsInDir_ValidDeployments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupFiles map[string]string
	}{
		{
			name: "valid service deployment",
			setupFiles: map[string]string{
				"deployments/sm-kms/secrets/unseal-1of5.secret": "sm-kms-unseal-key-1-of-5-" + hexA(),
				"deployments/sm-kms/secrets/unseal-2of5.secret": "sm-kms-unseal-key-2-of-5-" + hexB(),
				"deployments/sm-kms/secrets/unseal-3of5.secret": "sm-kms-unseal-key-3-of-5-" + hexC(),
				"deployments/sm-kms/secrets/unseal-4of5.secret": "sm-kms-unseal-key-4-of-5-" + hexD(),
				"deployments/sm-kms/secrets/unseal-5of5.secret": "sm-kms-unseal-key-5-of-5-" + hexE(),
			},
		},
		{
			name: "valid product deployment",
			setupFiles: map[string]string{
				"deployments/sm/secrets/unseal-1of5.secret": "sm-unseal-key-1-of-5-" + hexA(),
				"deployments/sm/secrets/unseal-2of5.secret": "sm-unseal-key-2-of-5-" + hexB(),
				"deployments/sm/secrets/unseal-3of5.secret": "sm-unseal-key-3-of-5-" + hexC(),
				"deployments/sm/secrets/unseal-4of5.secret": "sm-unseal-key-4-of-5-" + hexD(),
				"deployments/sm/secrets/unseal-5of5.secret": "sm-unseal-key-5-of-5-" + hexE(),
			},
		},
		{
			name: "valid suite deployment",
			setupFiles: map[string]string{
				"deployments/cryptoutil/secrets/unseal-1of5.secret": "cryptoutil-unseal-key-1-of-5-" + hexA(),
				"deployments/cryptoutil/secrets/unseal-2of5.secret": "cryptoutil-unseal-key-2-of-5-" + hexB(),
				"deployments/cryptoutil/secrets/unseal-3of5.secret": "cryptoutil-unseal-key-3-of-5-" + hexC(),
				"deployments/cryptoutil/secrets/unseal-4of5.secret": "cryptoutil-unseal-key-4-of-5-" + hexD(),
				"deployments/cryptoutil/secrets/unseal-5of5.secret": "cryptoutil-unseal-key-5-of-5-" + hexE(),
			},
		},
		{
			name: "deployment without secrets dir is skipped",
			setupFiles: map[string]string{
				"deployments/no-secrets/.gitkeep": "",
			},
		},
		{
			name: "deployment with no unseal files is valid",
			setupFiles: map[string]string{
				"deployments/shared-postgres/secrets/postgres.secret": "password123",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootDir := setupTempDir(t, tc.setupFiles)

			violations, err := FindViolationsInDir(rootDir)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(violations) > 0 {
				t.Errorf("expected no violations, got %d: %v", len(violations), violations)
			}
		})
	}
}

func TestFindViolationsInDir_InvalidContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupFiles     map[string]string
		wantViolations int
		wantSubstring  string
	}{
		{
			name: "wrong prefix",
			setupFiles: map[string]string{
				"deployments/pki-ca/secrets/unseal-1of5.secret": "kms-unseal-key-1-of-5-" + hexA(),
			},
			wantViolations: 1,
			wantSubstring:  "prefix",
		},
		{
			name: "wrong shard number",
			setupFiles: map[string]string{
				"deployments/sm-kms/secrets/unseal-1of5.secret": "sm-kms-unseal-key-3-of-5-" + hexA(),
			},
			wantViolations: 1,
			wantSubstring:  "shard number",
		},
		{
			name: "generic dev prefix",
			setupFiles: map[string]string{
				"deployments/dev/secrets/unseal-1of5.secret": "dev-unseal-key-1-of-5-" + hexA(),
			},
			wantViolations: 1,
			wantSubstring:  "generic",
		},
		{
			name: "duplicate hex values",
			setupFiles: map[string]string{
				"deployments/sm-kms/secrets/unseal-1of5.secret": "sm-kms-unseal-key-1-of-5-" + hexA(),
				"deployments/sm-kms/secrets/unseal-2of5.secret": "sm-kms-unseal-key-2-of-5-" + hexA(),
			},
			wantViolations: 1,
			wantSubstring:  "duplicate hex",
		},
		{
			name: "empty file",
			setupFiles: map[string]string{
				"deployments/sm-kms/secrets/unseal-1of5.secret": "",
			},
			wantViolations: 1,
			wantSubstring:  "empty",
		},
		{
			name: "invalid format - no hex",
			setupFiles: map[string]string{
				"deployments/sm-kms/secrets/unseal-1of5.secret": "sm-kms-unseal-key-1-of-5",
			},
			wantViolations: 1,
			wantSubstring:  "does not match pattern",
		},
		{
			name: "invalid format - short hex",
			setupFiles: map[string]string{
				"deployments/sm-kms/secrets/unseal-1of5.secret": "sm-kms-unseal-key-1-of-5-abcd1234",
			},
			wantViolations: 1,
			wantSubstring:  "does not match pattern",
		},
		{
			name: "invalid format - uppercase hex",
			setupFiles: map[string]string{
				"deployments/sm-kms/secrets/unseal-1of5.secret": "sm-kms-unseal-key-1-of-5-" + hexAUpper(),
			},
			wantViolations: 1,
			wantSubstring:  "does not match pattern",
		},
		{
			name: "content with trailing newline is valid",
			setupFiles: map[string]string{
				"deployments/sm-kms/secrets/unseal-1of5.secret": "sm-kms-unseal-key-1-of-5-" + hexA() + "\n",
			},
			wantViolations: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootDir := setupTempDir(t, tc.setupFiles)

			violations, err := FindViolationsInDir(rootDir)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(violations) != tc.wantViolations {
				t.Errorf("expected %d violations, got %d: %v", tc.wantViolations, len(violations), violations)
			}

			if tc.wantSubstring != "" && len(violations) > 0 {
				found := false

				for _, v := range violations {
					if containsSubstring(v, tc.wantSubstring) {
						found = true

						break
					}
				}

				if !found {
					t.Errorf("expected violation containing %q, got: %v", tc.wantSubstring, violations)
				}
			}
		})
	}
}

func TestFindViolationsInDir_MissingDeploymentsDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	_, err := FindViolationsInDir(tmpDir)
	if err == nil {
		t.Fatal("expected error for missing deployments/ directory")
	}
}

func TestCheckInDir_WithViolation(t *testing.T) {
	t.Parallel()

	files := map[string]string{
		"deployments/sm-kms/secrets/unseal-1of5.secret": "wrong-prefix-unseal-key-1-of-5-" + hexA(),
	}
	rootDir := setupTempDir(t, files)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, rootDir)
	if err == nil {
		t.Fatal("expected error from CheckInDir with violations")
	}
}

func TestCheckInDir_NoViolation(t *testing.T) {
	t.Parallel()

	files := map[string]string{
		"deployments/sm-kms/secrets/unseal-1of5.secret": "sm-kms-unseal-key-1-of-5-" + hexA(),
	}
	rootDir := setupTempDir(t, files)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, rootDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// setupTempDir creates a temporary directory with the given file structure.
func setupTempDir(t *testing.T, files map[string]string) string {
	t.Helper()

	rootDir := t.TempDir()

	for relPath, content := range files {
		absPath := filepath.Join(rootDir, filepath.FromSlash(relPath))

		if err := os.MkdirAll(filepath.Dir(absPath), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute); err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}

		if err := os.WriteFile(absPath, []byte(content), cryptoutilSharedMagic.CacheFilePermissions); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}

	return rootDir
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}

// Unique 64-char hex values for testing.
func hexA() string {
	return "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
}

func hexB() string {
	return "b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3"
}

func hexC() string {
	return "c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4"
}

func hexD() string {
	return "d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5"
}

func hexE() string {
	return "e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6"
}

func hexAUpper() string {
	return "A1B2C3D4E5F6A1B2C3D4E5F6A1B2C3D4E5F6A1B2C3D4E5F6A1B2C3D4E5F6A1B2"
}
