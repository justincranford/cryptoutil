// Copyright (c) 2025 Justin Cranford

package dockerfile_healthcheck

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilCmdCicdRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestCheck_RealWorkspace(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	// Run from project root (6 levels up from this test file).
	err := CheckInDir(logger, filepath.Join("..", "..", "..", "..", "..", ".."))
	require.NoError(t, err, "unexpected violation in real workspace")
}

func TestFindViolationsInDir_ValidDockerfiles(t *testing.T) {
	t.Parallel()

	psID := cryptoutilSharedMagic.OTLPServiceSMKMS
	entrypoint := cryptoutilCmdCicdRegistry.DockerfileEntrypoint(psID)
	binaryPath := "/app/" + psID

	tests := []struct {
		name       string
		setupFiles map[string]string
	}{
		{
			name: "correct PS-ID livez healthcheck",
			setupFiles: map[string]string{
				"deployments/sm-kms/Dockerfile": dockerfileWithHealthcheck(
					entrypoint,
					binaryPath+" livez || exit 1",
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
			name: "no HEALTHCHECK instruction is acceptable for non-PS-ID",
			setupFiles: map[string]string{
				"deployments/shared-telemetry/Dockerfile": dockerfileNoHealthcheck(),
			},
		},
		{
			name: "suite deployment with correct healthcheck",
			setupFiles: map[string]string{
				"deployments/cryptoutil/Dockerfile": dockerfileWithHealthcheck(
					[]string{"/app/cryptoutil"},
					"/app/cryptoutil livez || exit 1",
				),
			},
		},
		{
			name: "multi-line healthcheck with continuation",
			setupFiles: map[string]string{
				"deployments/sm-kms/Dockerfile": dockerfileWithMultiLineHealthcheck(
					entrypoint,
					binaryPath+" livez || exit 1",
				),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootDir := setupTempDir(t, tc.setupFiles)

			violations, err := FindViolationsInDir(rootDir)
			require.NoError(t, err)
			require.Empty(t, violations, "expected no violations but got: %v", violations)
		})
	}
}

func TestFindViolationsInDir_InvalidDockerfiles(t *testing.T) {
	t.Parallel()

	psID := cryptoutilSharedMagic.OTLPServiceSMKMS
	entrypoint := cryptoutilCmdCicdRegistry.DockerfileEntrypoint(psID)

	tests := []struct {
		name            string
		setupFiles      map[string]string
		wantViolations  int
		wantContains    []string
		wantNotContains []string
	}{
		{
			name: "wget in healthcheck",
			setupFiles: map[string]string{
				"deployments/sm-kms/Dockerfile": dockerfileWithHealthcheck(
					entrypoint,
					"wget --spider https://127.0.0.1:9090/admin/api/v1/livez || exit 1",
				),
			},
			wantViolations: 2,
			wantContains:   []string{"banned tool \"wget\"", "HEALTHCHECK CMD should be"},
		},
		{
			name: "curl in healthcheck",
			setupFiles: map[string]string{
				"deployments/sm-kms/Dockerfile": dockerfileWithHealthcheck(
					entrypoint,
					"curl -sf https://127.0.0.1:9090/admin/api/v1/livez || exit 1",
				),
			},
			wantViolations: 2,
			wantContains:   []string{"banned tool \"curl\"", "HEALTHCHECK CMD should be"},
		},
		{
			name: "wrong binary in healthcheck",
			setupFiles: map[string]string{
				"deployments/sm-kms/Dockerfile": dockerfileWithHealthcheck(
					entrypoint,
					"/app/wrong-binary livez || exit 1",
				),
			},
			wantViolations: 1,
			wantContains:   []string{"HEALTHCHECK CMD should be"},
		},
		{
			name: "readyz instead of livez",
			setupFiles: map[string]string{
				"deployments/sm-kms/Dockerfile": dockerfileWithHealthcheck(
					entrypoint,
					"/app/sm-kms readyz || exit 1",
				),
			},
			wantViolations: 1,
			wantContains:   []string{"HEALTHCHECK CMD should be"},
		},
		{
			name: "missing exit 1 suffix",
			setupFiles: map[string]string{
				"deployments/sm-kms/Dockerfile": dockerfileWithHealthcheck(
					entrypoint,
					"/app/sm-kms livez",
				),
			},
			wantViolations: 1,
			wantContains:   []string{"HEALTHCHECK CMD should be"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootDir := setupTempDir(t, tc.setupFiles)

			violations, err := FindViolationsInDir(rootDir)
			require.NoError(t, err)
			require.Len(t, violations, tc.wantViolations, "violations: %v", violations)

			for _, want := range tc.wantContains {
				found := false

				for _, v := range violations {
					if containsSubstring(v, want) {
						found = true

						break
					}
				}

				require.True(t, found, "expected violation containing %q in %v", want, violations)
			}

			for _, notWant := range tc.wantNotContains {
				for _, v := range violations {
					require.NotContains(t, v, notWant)
				}
			}
		})
	}
}

func TestFindViolationsInDir_MissingDeploymentsDir(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	_, err := FindViolationsInDir(rootDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read deployments/ directory")
}

func TestExtractHealthcheckCMD_FileNotFound(t *testing.T) {
	t.Parallel()

	cmd, err := extractHealthcheckCMD("/nonexistent/Dockerfile")
	require.Error(t, err)
	require.Empty(t, cmd)
}

func TestContainsTool_Boundary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cmd  string
		tool string
		want bool
	}{
		{name: "wget standalone", cmd: "wget --spider url", tool: "wget", want: true},
		{name: "wget as substring", cmd: "powget data", tool: "wget", want: false},
		{name: "curl standalone", cmd: "curl -sf url", tool: "curl", want: true},
		{name: "curl as substring", cmd: "securly data", tool: "curl", want: false},
		{name: "tool at end", cmd: "something wget", tool: "wget", want: true},
		{name: "tool with slash prefix", cmd: "/usr/bin/wget url", tool: "wget", want: true},
		{name: "empty cmd", cmd: "", tool: "wget", want: false},
		{name: "case insensitive", cmd: "WGET --spider url", tool: "wget", want: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := containsTool(tc.cmd, tc.tool)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestJoinHealthcheckLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		lines []string
		want  string
	}{
		{
			name:  "single line",
			lines: []string{"HEALTHCHECK CMD /app/sm-kms livez || exit 1"},
			want:  "HEALTHCHECK CMD /app/sm-kms livez || exit 1",
		},
		{
			name: "multi-line with continuations",
			lines: []string{
				"HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \\",
				"CMD /app/sm-kms livez || exit 1",
			},
			want: "HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 CMD /app/sm-kms livez || exit 1",
		},
		{
			name:  "empty input",
			lines: nil,
			want:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := joinHealthcheckLines(tc.lines)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestCheckInDir_NoViolations(t *testing.T) {
	t.Parallel()

	psID := cryptoutilSharedMagic.OTLPServiceSMKMS
	entrypoint := cryptoutilCmdCicdRegistry.DockerfileEntrypoint(psID)
	binaryPath := "/app/" + psID

	rootDir := setupTempDir(t, map[string]string{
		"deployments/sm-kms/Dockerfile": dockerfileWithHealthcheck(
			entrypoint,
			binaryPath+" livez || exit 1",
		),
	})

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, rootDir)
	require.NoError(t, err)
}

func TestCheckInDir_WithViolations(t *testing.T) {
	t.Parallel()

	psID := cryptoutilSharedMagic.OTLPServiceSMKMS
	entrypoint := cryptoutilCmdCicdRegistry.DockerfileEntrypoint(psID)

	rootDir := setupTempDir(t, map[string]string{
		"deployments/sm-kms/Dockerfile": dockerfileWithHealthcheck(
			entrypoint,
			"wget --spider https://127.0.0.1:9090/admin/api/v1/livez || exit 1",
		),
	})

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, rootDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "dockerfile healthcheck violations")
}

// --- Helpers ---

func containsSubstring(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}

func setupTempDir(t *testing.T, files map[string]string) string {
	t.Helper()

	rootDir := t.TempDir()

	for relPath, content := range files {
		fullPath := filepath.Join(rootDir, filepath.FromSlash(relPath))

		err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm)
		require.NoError(t, err)

		err = os.WriteFile(fullPath, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
		require.NoError(t, err)
	}

	return rootDir
}

func dockerfileWithHealthcheck(entrypoint []string, healthcheckCMD string) string {
	var ep string

	if len(entrypoint) > 0 {
		parts := make([]string, len(entrypoint))

		for i, p := range entrypoint {
			parts[i] = `"` + p + `"`
		}

		ep = "ENTRYPOINT [" + joinWithComma(parts) + "]"
	}

	return "FROM alpine:latest\n" +
		"HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \\\n" +
		"    CMD " + healthcheckCMD + "\n" +
		ep + "\n"
}

func dockerfileWithMultiLineHealthcheck(entrypoint []string, healthcheckCMD string) string {
	var ep string

	if len(entrypoint) > 0 {
		parts := make([]string, len(entrypoint))

		for i, p := range entrypoint {
			parts[i] = `"` + p + `"`
		}

		ep = "ENTRYPOINT [" + joinWithComma(parts) + "]"
	}

	return "FROM alpine:latest\n" +
		"HEALTHCHECK \\\n" +
		"    --interval=30s \\\n" +
		"    --timeout=10s \\\n" +
		"    --start-period=30s \\\n" +
		"    --retries=3 \\\n" +
		"    CMD " + healthcheckCMD + "\n" +
		ep + "\n"
}

func dockerfileNoHealthcheck() string {
	return "FROM alpine:latest\n" +
		"RUN echo hello\n"
}

func joinWithComma(parts []string) string {
	result := ""

	for i, p := range parts {
		if i > 0 {
			result += ", "
		}

		result += p
	}

	return result
}
