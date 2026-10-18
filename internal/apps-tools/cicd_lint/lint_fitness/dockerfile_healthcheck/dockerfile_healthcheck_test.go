// Copyright (c) 2025-2026 Justin Cranford.
package dockerfile_healthcheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilCmdCicdRegistry "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const (
	testSmKMSExecutablePath = "/app/sm-kms"
	testSmKMSLivezCommand   = `["` + testSmKMSExecutablePath + `", "livez"]`
)

func TestCheck_RealWorkspace(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	// Run from project root (6 levels up from this test file).
	err := CheckInDir(logger, filepath.Join("..", "..", "..", "..", ".."))
	require.NoError(t, err, "unexpected violation in real workspace")
}

func TestFindViolationsInDir_ValidDockerfiles(t *testing.T) {
	t.Parallel()

	psID := cryptoutilSharedMagic.OTLPServiceSMKMS
	entrypoint := cryptoutilCmdCicdRegistry.DockerfileEntrypoint(psID)

	tests := []struct {
		name       string
		setupFiles map[string]string
	}{
		{
			name: "correct PS-ID livez healthcheck",
			setupFiles: map[string]string{
				testDeploymentsSmKMSDockerfile: dockerfileWithHealthcheck(
					entrypoint,
					testSmKMSLivezCommand,
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
			name: testNoHealthcheckInstructionIsAcceptableForNonPsID,
			setupFiles: map[string]string{
				"deployments/shared-telemetry/Dockerfile": dockerfileNoHealthcheck(),
			},
		},
		{
			name: "multi-line healthcheck with continuation",
			setupFiles: map[string]string{
				testDeploymentsSmKMSDockerfile: dockerfileWithMultiLineHealthcheck(
					entrypoint,
					testSmKMSLivezCommand,
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

			if tc.name == testNoHealthcheckInstructionIsAcceptableForNonPsID {
				require.Len(t, violations, 1)
				require.Contains(t, violations[0], "non-PS-ID deployment MUST NOT define Dockerfile")

				return
			}

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
				testDeploymentsSmKMSDockerfile: dockerfileWithHealthcheck(
					entrypoint,
					"wget --spider https://127.0.0.1:9090/admin/api/v1/livez || exit 1",
				),
			},
			wantViolations: 2,
			wantContains:   []string{"banned tool \"wget\"", testHealthcheckCmdShouldBe},
		},
		{
			name: "curl in healthcheck",
			setupFiles: map[string]string{
				testDeploymentsSmKMSDockerfile: dockerfileWithHealthcheck(
					entrypoint,
					"curl -sf https://127.0.0.1:9090/admin/api/v1/livez || exit 1",
				),
			},
			wantViolations: 2,
			wantContains:   []string{"banned tool \"curl\"", testHealthcheckCmdShouldBe},
		},
		{
			name: "wrong binary in healthcheck",
			setupFiles: map[string]string{
				testDeploymentsSmKMSDockerfile: dockerfileWithHealthcheck(
					entrypoint,
					"/app/wrong-binary livez || exit 1",
				),
			},
			wantViolations: 1,
			wantContains:   []string{testHealthcheckCmdShouldBe},
		},
		{
			name: "readyz instead of livez",
			setupFiles: map[string]string{
				testDeploymentsSmKMSDockerfile: dockerfileWithHealthcheck(
					entrypoint,
					testSmKMSExecutablePath+" readyz || exit 1",
				),
			},
			wantViolations: 1,
			wantContains:   []string{testHealthcheckCmdShouldBe},
		},
		{
			name: "missing exit 1 suffix",
			setupFiles: map[string]string{
				testDeploymentsSmKMSDockerfile: dockerfileWithHealthcheck(
					entrypoint,
					testSmKMSExecutablePath+" livez",
				),
			},
			wantViolations: 1,
			wantContains:   []string{testHealthcheckCmdShouldBe},
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

func TestCheckInDir_MissingDeploymentsDir(t *testing.T) {
	t.Parallel()

	err := CheckInDir(cryptoutilCmdCicdCommon.NewLogger("test"), t.TempDir())

	require.ErrorContains(t, err, "failed to check Dockerfile healthchecks")
}

func TestFindViolationsInDir_NonDirectoryEntryIsSkipped(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	deploymentsDir := filepath.Join(rootDir, "deployments")
	require.NoError(t, os.MkdirAll(deploymentsDir, cryptoutilSharedMagic.DirPermissions))
	require.NoError(t, os.WriteFile(filepath.Join(deploymentsDir, "README.md"), nil, cryptoutilSharedMagic.CacheFilePermissions))

	violations, err := FindViolationsInDir(rootDir)

	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestExtractHealthcheckCMD_FileNotFound(t *testing.T) {
	t.Parallel()

	cmd, err := extractHealthcheckCMD("/nonexistent/Dockerfile")
	require.Error(t, err)
	require.Empty(t, cmd)
}

func TestValidateDockerfileHealthcheck_NoHealthcheck(t *testing.T) {
	t.Parallel()

	dockerfilePath := filepath.Join(t.TempDir(), "Dockerfile")
	require.NoError(t, os.WriteFile(dockerfilePath, []byte("FROM scratch\n"), cryptoutilSharedMagic.CacheFilePermissions))

	require.Empty(t, validateDockerfileHealthcheck(
		dockerfilePath,
		cryptoutilSharedMagic.OTLPServiceSMKMS,
		map[string]string{cryptoutilSharedMagic.OTLPServiceSMKMS: testSmKMSExecutablePath},
	))
}

func TestValidateDockerfileHealthcheck_ReadError(t *testing.T) {
	t.Parallel()

	violations := validateDockerfileHealthcheck(
		filepath.Join(t.TempDir(), "missing-Dockerfile"),
		cryptoutilSharedMagic.OTLPServiceSMKMS,
		map[string]string{cryptoutilSharedMagic.OTLPServiceSMKMS: testSmKMSExecutablePath},
	)

	require.Len(t, violations, 1)
	require.Contains(t, violations[0], "failed to read Dockerfile")
}

func TestExtractHealthcheckCMD_ScannerError(t *testing.T) {
	t.Parallel()

	dockerfilePath := filepath.Join(t.TempDir(), "Dockerfile")
	require.NoError(t, os.WriteFile(
		dockerfilePath,
		[]byte("HEALTHCHECK "+strings.Repeat("x", 70*1024)),
		cryptoutilSharedMagic.CacheFilePermissions,
	))

	_, err := extractHealthcheckCMD(dockerfilePath)

	require.ErrorContains(t, err, "failed to scan Dockerfile")
	require.ErrorContains(t, err, "token too long")
}

// Sequential: changes the process working directory.
func TestCheck_FromWorkspaceRoot(t *testing.T) {
	rootDir := filepath.Join("..", "..", "..", "..", "..")

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(rootDir))

	defer func() {
		require.NoError(t, os.Chdir(originalDir))
	}()

	require.NoError(t, Check(cryptoutilCmdCicdCommon.NewLogger("test")))
}

func TestContainsTool_Boundary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cmd  string
		tool string
		want bool
	}{
		{name: "wget standalone", cmd: "wget --spider url", tool: valueWget, want: true},
		{name: "wget as substring", cmd: "powget data", tool: valueWget, want: false},
		{name: "wget with alphanumeric left boundary", cmd: "xwget data", tool: valueWget, want: false},
		{name: "wget with alphanumeric right boundary", cmd: "wgetx data", tool: valueWget, want: false},
		{name: "curl standalone", cmd: "curl -sf url", tool: valueCurl, want: true},
		{name: "curl as substring", cmd: "securly data", tool: valueCurl, want: false},
		{name: "tool at end", cmd: "something wget", tool: valueWget, want: true},
		{name: "tool with slash prefix", cmd: "/usr/bin/wget url", tool: valueWget, want: true},
		{name: "empty cmd", cmd: "", tool: valueWget, want: false},
		{name: "case insensitive", cmd: "WGET --spider url", tool: valueWget, want: true},
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
			lines: []string{testHealthcheckCmdAppSmKMSLivezExec},
			want:  testHealthcheckCmdAppSmKMSLivezExec,
		},
		{
			name: "multi-line with continuations",
			lines: []string{
				"HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \\",
				"CMD " + testSmKMSLivezCommand,
			},
			want: "HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 CMD " + testSmKMSLivezCommand,
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

	rootDir := setupTempDir(t, map[string]string{
		testDeploymentsSmKMSDockerfile: dockerfileWithHealthcheck(
			entrypoint,
			testSmKMSLivezCommand,
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
		testDeploymentsSmKMSDockerfile: dockerfileWithHealthcheck(
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
