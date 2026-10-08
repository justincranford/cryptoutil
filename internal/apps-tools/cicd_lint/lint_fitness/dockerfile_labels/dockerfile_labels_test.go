// Copyright (c) 2025-2026 Justin Cranford.
package dockerfile_labels

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilCmdCicdRegistry "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheck_RealWorkspace(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	// Run from project root (6 levels up from this test file).
	err := CheckInDir(logger, filepath.Join("..", "..", "..", "..", ".."))
	if err != nil {
		require.FailNowf(t, cryptoutilSharedMagic.CICDTestFatalfFailureMessage, "unexpected violation in real workspace: %v", err)
	}
}

func TestFindViolationsInDir_ValidDockerfiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupFiles map[string]string
	}{
		{
			name: "correct PS-ID title and entrypoint for sm-kms",
			setupFiles: map[string]string{
				"deployments/sm-kms/Dockerfile": dockerfileWithLabelsAndEntrypoint(
					cryptoutilCmdCicdRegistry.OTLPServiceName(cryptoutilSharedMagic.OTLPServiceSMKMS),
					"SM KMS Server",
					cryptoutilCmdCicdRegistry.DockerfileEntrypoint(cryptoutilSharedMagic.OTLPServiceSMKMS),
				),
			},
		},
		{
			name: "correct PS-ID title and entrypoint for pki-ca",
			setupFiles: map[string]string{
				"deployments/pki-ca/Dockerfile": dockerfileWithLabelsAndEntrypoint(
					cryptoutilCmdCicdRegistry.OTLPServiceName(cryptoutilSharedMagic.OTLPServicePKICA),
					"JOSE JWK Authority Server",
					cryptoutilCmdCicdRegistry.DockerfileEntrypoint(cryptoutilSharedMagic.OTLPServicePKICA),
				),
			},
		},
		{
			name: "correct PS-ID title and entrypoint for identity-authz",
			setupFiles: map[string]string{
				"deployments/identity-authz/Dockerfile": dockerfileWithLabelsAndEntrypoint(
					cryptoutilCmdCicdRegistry.OTLPServiceName(cryptoutilSharedMagic.OTLPServiceIdentityAuthz),
					"Identity AuthZ Server",
					cryptoutilCmdCicdRegistry.DockerfileEntrypoint(cryptoutilSharedMagic.OTLPServiceIdentityAuthz),
				),
			},
		},
		{
			name: "deployment without Dockerfile is skipped",
			setupFiles: map[string]string{
				"deployments/shared-postgres/.gitkeep": "",
			},
		},
		{
			name: "PS-ID title is case-insensitive exact match",
			setupFiles: map[string]string{
				"deployments/sm-kms/Dockerfile": dockerfileWithLabelsAndEntrypoint(
					"CRYPTOUTIL-SM-KMS",
					"Description",
					cryptoutilCmdCicdRegistry.DockerfileEntrypoint(cryptoutilSharedMagic.OTLPServiceSMKMS),
				),
			},
		},
		{
			name: "multi-line labels with correct PS-ID entrypoint",
			setupFiles: map[string]string{
				"deployments/sm-kms/Dockerfile": multiLineLabelDockerfileWithEntrypoint(
					cryptoutilCmdCicdRegistry.OTLPServiceName(cryptoutilSharedMagic.OTLPServiceSMKMS),
					"SM KMS Server",
					cryptoutilCmdCicdRegistry.DockerfileEntrypoint(cryptoutilSharedMagic.OTLPServiceSMKMS),
				),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootDir := setupTempDir(t, tc.setupFiles)

			violations, err := FindViolationsInDir(rootDir)
			if err != nil {
				require.FailNowf(t, cryptoutilSharedMagic.CICDTestFatalfFailureMessage, "unexpected error: %v", err)
			}

			if len(violations) > 0 {
				assert.Failf(t, cryptoutilSharedMagic.CICDTestErrorfFailureMessage, "expected no violations, got %d: %v", len(violations), violations)
			}
		})
	}
}

func TestFindViolationsInDir_InvalidDockerfiles(t *testing.T) {
	t.Parallel()

	correctSMKMSEntrypoint := cryptoutilCmdCicdRegistry.DockerfileEntrypoint(cryptoutilSharedMagic.OTLPServiceSMKMS)
	correctSMKMSTitle := cryptoutilCmdCicdRegistry.OTLPServiceName(cryptoutilSharedMagic.OTLPServiceSMKMS)

	tests := []struct {
		name           string
		setupFiles     map[string]string
		wantViolations int
		wantSubstring  string
	}{
		{
			name: "wrong title for PS-ID - must exactly match registry value",
			setupFiles: map[string]string{
				"deployments/sm-kms/Dockerfile": dockerfileWithLabelsAndEntrypoint("CA Server", "Some description", correctSMKMSEntrypoint),
			},
			wantViolations: 1,
			wantSubstring:  "must exactly match registry-derived value",
		},
		{
			name: "wrong title for PS-ID - substring no longer accepted",
			setupFiles: map[string]string{
				"deployments/sm-kms/Dockerfile": dockerfileWithLabelsAndEntrypoint("sm-kms service", "Some description", correctSMKMSEntrypoint),
			},
			wantViolations: 1,
			wantSubstring:  "must exactly match registry-derived value",
		},
		{
			name: "wrong ENTRYPOINT for PS-ID",
			setupFiles: map[string]string{
				"deployments/sm-kms/Dockerfile": dockerfileWithLabelsAndEntrypoint(correctSMKMSTitle, "Some description", []string{"/app/wrong-binary"}),
			},
			wantViolations: 1,
			wantSubstring:  "must match registry-declared entrypoint",
		},
		{
			name: "missing ENTRYPOINT for PS-ID",
			setupFiles: map[string]string{
				"deployments/sm-kms/Dockerfile": dockerfileWithLabels(correctSMKMSTitle, "Some description"),
			},
			wantViolations: 1,
			wantSubstring:  "missing ENTRYPOINT instruction",
		},
		{
			name: "missing title label",
			setupFiles: map[string]string{
				"deployments/sm-kms/Dockerfile": dockerfileWithOnlyDescription("Some description"),
			},
			wantViolations: 2,
			wantSubstring:  "missing required label",
		},
		{
			name: "missing description label",
			setupFiles: map[string]string{
				"deployments/sm-kms/Dockerfile": dockerfileWithOnlyTitle(correctSMKMSTitle),
			},
			wantViolations: 2,
			wantSubstring:  "missing required label",
		},
		{
			name: "no labels at all",
			setupFiles: map[string]string{
				"deployments/sm-kms/Dockerfile": dockerfileNoLabels(),
			},
			wantViolations: 3,
		},
		{
			name: "non PS-ID deployment Dockerfile is rejected",
			setupFiles: map[string]string{
				"deployments/cryptoutil/Dockerfile": dockerfileWithLabels("CA Server", "Some description"),
			},
			wantViolations: 1,
			wantSubstring:  "non-PS-ID deployment MUST NOT define Dockerfile",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootDir := setupTempDir(t, tc.setupFiles)

			violations, err := FindViolationsInDir(rootDir)
			if err != nil {
				require.FailNowf(t, cryptoutilSharedMagic.CICDTestFatalfFailureMessage, "unexpected error: %v", err)
			}

			if len(violations) != tc.wantViolations {
				assert.Failf(t, cryptoutilSharedMagic.CICDTestErrorfFailureMessage, "expected %d violations, got %d: %v", tc.wantViolations, len(violations), violations)
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
					assert.Failf(t, cryptoutilSharedMagic.CICDTestErrorfFailureMessage, "expected violation containing %q, got: %v", tc.wantSubstring, violations)
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
		"deployments/sm-kms/Dockerfile": dockerfileWithLabelsAndEntrypoint(
			cryptoutilCmdCicdRegistry.OTLPServiceName(cryptoutilSharedMagic.OTLPServiceSMKMS),
			"SM KMS Server",
			cryptoutilCmdCicdRegistry.DockerfileEntrypoint(cryptoutilSharedMagic.OTLPServiceSMKMS),
		),
	}
	rootDir := setupTempDir(t, files)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, rootDir)
	if err != nil {
		require.FailNowf(t, cryptoutilSharedMagic.CICDTestFatalfFailureMessage, "unexpected error: %v", err)
	}
}

func TestExtractLabelsAndEntrypoint_MultiLine(t *testing.T) {
	t.Parallel()

	content := multiLineLabelDockerfileWithEntrypoint(
		cryptoutilCmdCicdRegistry.OTLPServiceName(cryptoutilSharedMagic.OTLPServicePKICA),
		"JOSE JWK Authority",
		cryptoutilCmdCicdRegistry.DockerfileEntrypoint(cryptoutilSharedMagic.OTLPServicePKICA),
	)
	tmpFile := filepath.Join(t.TempDir(), "Dockerfile")

	if err := os.WriteFile(tmpFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions); err != nil {
		require.FailNowf(t, cryptoutilSharedMagic.CICDTestFatalfFailureMessage, "failed to write temp file: %v", err)
	}

	labels, entrypoint, err := extractLabelsAndEntrypoint(tmpFile)
	if err != nil {
		require.FailNowf(t, cryptoutilSharedMagic.CICDTestFatalfFailureMessage, "unexpected error: %v", err)
	}

	title, ok := labels["org.opencontainers.image.title"]
	if !ok {
		t.Fatal("expected title label to be present")
	}

	if !strings.EqualFold(title, cryptoutilCmdCicdRegistry.OTLPServiceName(cryptoutilSharedMagic.OTLPServicePKICA)) {
		assert.Failf(t, cryptoutilSharedMagic.CICDTestErrorfFailureMessage, "expected title to match OTLP service name, got %q", title)
	}

	if len(entrypoint) == 0 {
		t.Error("expected entrypoint to be parsed")
	}
}

// setupTempDir creates a temporary directory with the given file structure.
func setupTempDir(t *testing.T, files map[string]string) string {
	t.Helper()

	rootDir := t.TempDir()

	for relPath, content := range files {
		absPath := filepath.Join(rootDir, filepath.FromSlash(relPath))

		if err := os.MkdirAll(filepath.Dir(absPath), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute); err != nil {
			require.FailNowf(t, cryptoutilSharedMagic.CICDTestFatalfFailureMessage, "failed to create directory: %v", err)
		}

		if err := os.WriteFile(absPath, []byte(content), cryptoutilSharedMagic.CacheFilePermissions); err != nil {
			require.FailNowf(t, cryptoutilSharedMagic.CICDTestFatalfFailureMessage, "failed to write file: %v", err)
		}
	}

	return rootDir
}

func dockerfileWithLabelsAndEntrypoint(title, description string, entrypoint []string) string {
	var sb strings.Builder

	sb.WriteString("FROM alpine:latest\n")
	sb.WriteString("LABEL org.opencontainers.image.title=\"" + title + "\"\n")
	sb.WriteString("LABEL org.opencontainers.image.description=\"" + description + "\"\n")

	if len(entrypoint) > 0 {
		sb.WriteString("ENTRYPOINT [")

		for i, arg := range entrypoint {
			if i > 0 {
				sb.WriteString(", ")
			}

			sb.WriteString("\"" + arg + "\"")
		}

		sb.WriteString("]\n")
	}

	return sb.String()
}

func dockerfileWithLabels(title, description string) string {
	return "FROM alpine:latest\n" +
		"LABEL org.opencontainers.image.title=\"" + title + "\"\n" +
		"LABEL org.opencontainers.image.description=\"" + description + "\"\n"
}

func multiLineLabelDockerfileWithEntrypoint(title, description string, entrypoint []string) string {
	var sb strings.Builder

	sb.WriteString("FROM alpine:latest\n")
	sb.WriteString("LABEL org.opencontainers.image.title=\"" + title + "\" \\\n")
	sb.WriteString("      org.opencontainers.image.description=\"" + description + "\"\n")

	if len(entrypoint) > 0 {
		sb.WriteString("ENTRYPOINT [")

		for i, arg := range entrypoint {
			if i > 0 {
				sb.WriteString(", ")
			}

			sb.WriteString("\"" + arg + "\"")
		}

		sb.WriteString("]\n")
	}

	return sb.String()
}

func dockerfileWithOnlyTitle(title string) string {
	return "FROM alpine:latest\n" +
		"LABEL org.opencontainers.image.title=\"" + title + "\"\n"
}

func dockerfileWithOnlyDescription(description string) string {
	return "FROM alpine:latest\n" +
		"LABEL org.opencontainers.image.description=\"" + description + "\"\n"
}

func dockerfileNoLabels() string {
	return "FROM alpine:latest\nRUN echo hello\n"
}

func TestCheck_ReturnsError_WhenNoDeployments(t *testing.T) {
	t.Parallel()

	// Cannot call Check() directly (it uses "." as rootDir which requires workspace root).
	// Instead test the same code path via CheckInDir with a dir that has no deployments/.
	tmpDir := t.TempDir()
	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, tmpDir)
	if err == nil {
		t.Fatal("expected error when deployments/ is missing")
	}
}

func TestParseEntrypointLine_UnquotedToken(t *testing.T) {
	t.Parallel()

	// Unquoted token inside brackets — should break and return nil/empty.
	result := parseEntrypointLine("[unquoted]")
	if len(result) != 0 {
		assert.Failf(t, cryptoutilSharedMagic.CICDTestErrorfFailureMessage, "expected empty result for unquoted token, got %v", result)
	}
}

func TestCheck_DirectCall(t *testing.T) {
	t.Parallel()

	// Call Check() directly -- it uses "." as rootDir.
	// In test context, "." is the package dir (no deployments/), so it returns an error.
	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := Check(logger)
	if err == nil {
		t.Fatal("expected error when running Check() from package directory (no deployments/)")
	}
}

func TestParseLabelsFromLine_UnquotedWithSpace(t *testing.T) {
	t.Parallel()

	labels := make(map[string]string)
	// Unquoted value followed by a space and another key=value pair.
	parseLabelsFromLine("key1=val1 key2=val2", labels)

	if labels["key1"] != "val1" {
		assert.Failf(t, cryptoutilSharedMagic.CICDTestErrorfFailureMessage, "expected key1=%q, got %q", "val1", labels["key1"])
	}

	if labels["key2"] != "val2" {
		assert.Failf(t, cryptoutilSharedMagic.CICDTestErrorfFailureMessage, "expected key2=%q, got %q", "val2", labels["key2"])
	}
}

func TestParseLabelsFromLine_UnclosedQuotedValue(t *testing.T) {
	t.Parallel()

	labels := make(map[string]string)
	// Quoted value with no closing quote -- takes the rest of the line as value.
	parseLabelsFromLine(`key="unclosed value`, labels)

	got, ok := labels["key"]
	if !ok {
		t.Fatal("expected key to be present")
	}

	if got != "unclosed value" {
		assert.Failf(t, cryptoutilSharedMagic.CICDTestErrorfFailureMessage, "expected %q, got %q", "unclosed value", got)
	}
}

func TestParseEntrypointLine_UnclosedQuotedToken(t *testing.T) {
	t.Parallel()

	// Quoted token with no closing quote inside brackets -- should break and return empty.
	result := parseEntrypointLine(`["/unclosed`)
	if len(result) != 0 {
		assert.Failf(t, cryptoutilSharedMagic.CICDTestErrorfFailureMessage, "expected empty result for unclosed quoted token, got %v", result)
	}
}

func TestValidateDockerfileLabels_ReadError(t *testing.T) {
	t.Parallel()

	// Point to a non-existent file -- extractLabelsAndEntrypoint will fail to open.
	psidTitleMap := map[string]string{}
	psidEntrypointMap := map[string][]string{}

	violations := validateDockerfileLabels("/nonexistent/path/Dockerfile", "nonexistent", psidTitleMap, psidEntrypointMap)
	if len(violations) == 0 {
		t.Fatal("expected violation for unreadable Dockerfile")
	}

	if !strings.Contains(violations[0], "failed to read Dockerfile") {
		assert.Failf(t, cryptoutilSharedMagic.CICDTestErrorfFailureMessage, "expected 'failed to read Dockerfile' in violation, got: %s", violations[0])
	}
}

func TestParseEntrypointLine_UnclosedQuoteInsideBrackets(t *testing.T) {
	t.Parallel()

	// Has outer brackets but quoted token inside has no closing quote -- should break and return empty.
	// e.g. ["/unclosed-value] - has [ and ] but no closing " inside
	result := parseEntrypointLine(`["/unclosed-value]`)
	if len(result) != 0 {
		assert.Failf(t, cryptoutilSharedMagic.CICDTestErrorfFailureMessage, "expected empty result for unclosed quoted token inside brackets, got %v", result)
	}
}
