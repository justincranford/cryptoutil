// Copyright (c) 2025 Justin Cranford

package secret_naming

import (
	"os"
	"path/filepath"
	"strings"
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

func TestFindViolationsInDir_ValidSecrets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupFiles map[string]string
	}{
		{
			name: "hyphenated secret files",
			setupFiles: map[string]string{
				"deployments/sm-kms/secrets/unseal-1of5.secret":       "content",
				"deployments/sm-kms/secrets/hash-pepper-v3.secret":    "content",
				"deployments/sm-kms/secrets/postgres-password.secret": "content",
			},
		},
		{
			name: "secret.never files are valid",
			setupFiles: map[string]string{
				"deployments/sm/secrets/browser-password.secret.never": "content",
				"deployments/sm/secrets/service-password.secret.never": "content",
			},
		},
		{
			name: "deployment without secrets dir is skipped",
			setupFiles: map[string]string{
				"deployments/shared-postgres/.gitkeep": "",
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

func TestFindViolationsInDir_InvalidSecrets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupFiles     map[string]string
		wantViolations int
		wantSubstring  string
	}{
		{
			name: "underscore in filename",
			setupFiles: map[string]string{
				"deployments/sm-kms/secrets/unseal_1of5.secret": "content",
			},
			wantViolations: 1,
			wantSubstring:  "underscore",
		},
		{
			name: "missing .secret extension",
			setupFiles: map[string]string{
				"deployments/sm-kms/secrets/password.txt": "content",
			},
			wantViolations: 1,
			wantSubstring:  "extension",
		},
		{
			name: "uppercase in filename",
			setupFiles: map[string]string{
				"deployments/sm-kms/secrets/Unseal-1of5.secret": "content",
			},
			wantViolations: 1,
			wantSubstring:  "lowercase",
		},
		{
			name: "multiple violations in one file",
			setupFiles: map[string]string{
				"deployments/sm-kms/secrets/Unseal_1of5.txt": "content",
			},
			wantViolations: 3,
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
					if strings.Contains(v, tc.wantSubstring) {
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
		"deployments/sm-kms/secrets/unseal_1of5.secret": "content",
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
		"deployments/sm-kms/secrets/unseal-1of5.secret": "content",
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
