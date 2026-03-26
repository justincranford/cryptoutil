// Copyright (c) 2025 Justin Cranford

package dockerfile_labels

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

func TestFindViolationsInDir_ValidDockerfiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupFiles map[string]string
	}{
		{
			name: "correct PS-ID in title",
			setupFiles: map[string]string{
				"deployments/sm-kms/Dockerfile": dockerfileWithLabels("cryptoutil-sm-kms", "SM KMS Server"),
			},
		},
		{
			name: "correct suite title",
			setupFiles: map[string]string{
				"deployments/cryptoutil-suite/Dockerfile": dockerfileWithLabels("CryptoUtil Suite", "CryptoUtil Suite Server"),
			},
		},
		{
			name: "deployment without Dockerfile is skipped",
			setupFiles: map[string]string{
				"deployments/shared-postgres/.gitkeep": "",
			},
		},
		{
			name: "title is case-insensitive",
			setupFiles: map[string]string{
				"deployments/sm-kms/Dockerfile": dockerfileWithLabels("CRYPTOUTIL-SM-KMS", "Description"),
			},
		},
		{
			name: "multi-line labels",
			setupFiles: map[string]string{
				"deployments/sm-kms/Dockerfile": multiLineLabelDockerfile("cryptoutil-sm-kms", "SM KMS Server"),
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

func TestFindViolationsInDir_InvalidDockerfiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupFiles     map[string]string
		wantViolations int
		wantSubstring  string
	}{
		{
			name: "wrong title - does not contain deployment name",
			setupFiles: map[string]string{
				"deployments/sm-kms/Dockerfile": dockerfileWithLabels("CA Server", "Some description"),
			},
			wantViolations: 1,
			wantSubstring:  "does not contain deployment name",
		},
		{
			name: "missing title label",
			setupFiles: map[string]string{
				"deployments/sm-kms/Dockerfile": dockerfileWithOnlyDescription("Some description"),
			},
			wantViolations: 1,
			wantSubstring:  "missing required label",
		},
		{
			name: "missing description label",
			setupFiles: map[string]string{
				"deployments/sm-kms/Dockerfile": dockerfileWithOnlyTitle("cryptoutil-sm-kms"),
			},
			wantViolations: 1,
			wantSubstring:  "missing required label",
		},
		{
			name: "no labels at all",
			setupFiles: map[string]string{
				"deployments/sm-kms/Dockerfile": dockerfileNoLabels(),
			},
			wantViolations: 2,
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
		"deployments/sm-kms/Dockerfile": dockerfileWithLabels("Wrong Title", "Description"),
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
		"deployments/sm-kms/Dockerfile": dockerfileWithLabels("cryptoutil-sm-kms", "SM KMS Server"),
	}
	rootDir := setupTempDir(t, files)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, rootDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExtractLabels_MultiLine(t *testing.T) {
	t.Parallel()

	content := multiLineLabelDockerfile("cryptoutil-test", "Test Server")
	tmpFile := filepath.Join(t.TempDir(), "Dockerfile")

	if err := os.WriteFile(tmpFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	labels, err := extractLabels(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	title, ok := labels["org.opencontainers.image.title"]
	if !ok {
		t.Fatal("expected title label to be present")
	}

	if !strings.Contains(strings.ToLower(title), "cryptoutil-test") {
		t.Errorf("expected title to contain 'cryptoutil-test', got %q", title)
	}
}

func TestTitleContainsDeploymentName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		title          string
		deploymentName string
		want           bool
	}{
		{name: "exact match", title: cryptoutilSharedMagic.OTLPServiceSMKMS, deploymentName: cryptoutilSharedMagic.OTLPServiceSMKMS, want: true},
		{name: "prefixed match", title: "cryptoutil-" + cryptoutilSharedMagic.OTLPServiceSMKMS, deploymentName: cryptoutilSharedMagic.OTLPServiceSMKMS, want: true},
		{name: "case insensitive", title: "CryptoUtil-Suite", deploymentName: "cryptoutil-suite", want: true},
		{name: "no match", title: "CA Server", deploymentName: cryptoutilSharedMagic.OTLPServiceSMKMS, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := titleContainsDeploymentName(tc.title, tc.deploymentName)
			if got != tc.want {
				t.Errorf("titleContainsDeploymentName(%q, %q) = %v, want %v", tc.title, tc.deploymentName, got, tc.want)
			}
		})
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

func dockerfileWithLabels(title, description string) string {
	return "FROM alpine:3.19\n" +
		"LABEL org.opencontainers.image.title=\"" + title + "\"\n" +
		"LABEL org.opencontainers.image.description=\"" + description + "\"\n"
}

func multiLineLabelDockerfile(title, description string) string {
	return "FROM alpine:3.19\n" +
		"LABEL org.opencontainers.image.title=\"" + title + "\" \\\n" +
		"      org.opencontainers.image.description=\"" + description + "\"\n"
}

func dockerfileWithOnlyTitle(title string) string {
	return "FROM alpine:3.19\n" +
		"LABEL org.opencontainers.image.title=\"" + title + "\"\n"
}

func dockerfileWithOnlyDescription(description string) string {
	return "FROM alpine:3.19\n" +
		"LABEL org.opencontainers.image.description=\"" + description + "\"\n"
}

func dockerfileNoLabels() string {
	return "FROM alpine:3.19\nRUN echo hello\n"
}
