// Copyright (c) 2025 Justin Cranford

package lint_ports

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	lintPortsCommon "cryptoutil/internal/apps/cicd/lint_ports/common"

	"github.com/stretchr/testify/require"
)

func TestLint_NoLegacyPorts(t *testing.T) {
	t.Parallel()

	// Create a temp directory with clean files.
	tempDir := t.TempDir()

	// Create a Go file with standardized ports only.
	goFile := filepath.Join(tempDir, "main.go")
	err := os.WriteFile(goFile, []byte(`package main

const port = 8700 // sm-im standardized port
`), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go": {goFile},
	}

	err = Lint(logger, filesByExtension)
	require.NoError(t, err)
}

func TestLint_DetectsLegacyPort(t *testing.T) {
	t.Parallel()

	// Create a temp directory with legacy port.
	tempDir := t.TempDir()

	// Create a Go file with legacy port 8888.
	goFile := filepath.Join(tempDir, "main.go")
	err := os.WriteFile(goFile, []byte(`package main

const port = 8888 // legacy sm-im port
`), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go": {goFile},
	}

	err = Lint(logger, filesByExtension)
	require.Error(t, err)
	require.Contains(t, err.Error(), "legacy port violations")
}

func TestLint_DetectsMultipleLegacyPorts(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create files with different legacy ports.
	goFile := filepath.Join(tempDir, "main.go")
	err := os.WriteFile(goFile, []byte(`package main

const cipherPort = 8888 // legacy
const josePort = 9443   // legacy
`), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go": {goFile},
	}

	err = Lint(logger, filesByExtension)
	require.Error(t, err)
	require.Contains(t, err.Error(), "2 legacy port violations")
}

func TestLint_SkipsOtelFilesForOtelPorts(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create an OTEL-related file with OTEL ports.
	otelFile := filepath.Join(tempDir, "otel_config.go")
	err := os.WriteFile(otelFile, []byte(`package main

const metricsPort = 8888 // OTEL internal metrics
const promPort = 8889    // OTEL Prometheus port
`), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go": {otelFile},
	}

	err = Lint(logger, filesByExtension)
	require.NoError(t, err) // Should pass - OTEL ports in OTEL files are OK
}

func TestLint_DetectsLegacyPortInRegularFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create a regular file with a sm-im legacy port.
	// Use 8890 which is NOT an OTEL collector port.
	normalFile := filepath.Join(tempDir, "config.go")
	err := os.WriteFile(normalFile, []byte(`package main

const port = 8890 // legacy sm-im port
`), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go": {normalFile},
	}

	err = Lint(logger, filesByExtension)
	require.Error(t, err) // Should fail - 8890 is legacy sm-im port
}

func TestLint_Detects8888InRegularFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create a regular file (not in telemetry/observability path) with port 8888.
	// 8888 is both a sm-im legacy port AND an OTEL collector port.
	// It should be detected as legacy in non-observability files.
	normalFile := filepath.Join(tempDir, "main.go")
	err := os.WriteFile(normalFile, []byte(`package main

const port = 8888
`), 0o600)
	require.NoError(t, err)

	// Verify the file path does NOT contain OTEL-related terms.
	require.False(t, lintPortsCommon.IsOtelRelatedFile(normalFile), "Test file path should NOT be telemetry-related")

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go": {normalFile},
	}

	err = Lint(logger, filesByExtension)
	require.Error(t, err, "Port 8888 should be detected as legacy in regular files")
	require.Contains(t, err.Error(), "legacy port violations")
}

func TestLint_YamlFiles(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create YAML file with legacy port.
	yamlFile := filepath.Join(tempDir, "config.yml")
	err := os.WriteFile(yamlFile, []byte(`server:
  port: 8443
`), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"yml": {yamlFile},
	}

	err = Lint(logger, filesByExtension)
	require.Error(t, err)
	require.Contains(t, err.Error(), "legacy port violations")
}

func TestLint_MarkdownFiles(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create markdown file with legacy port.
	mdFile := filepath.Join(tempDir, "README.md")
	err := os.WriteFile(mdFile, []byte(`# Server

Connect to port 9443 for JOSE.
`), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"md": {mdFile},
	}

	err = Lint(logger, filesByExtension)
	require.Error(t, err)
	require.Contains(t, err.Error(), "legacy port violations")
}

func TestLint_SkipsLintPortsDirectory(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create a file that looks like it's in lint_ports directory.
	lintPortsFile := filepath.Join(tempDir, "lint_ports", "constants.go")
	err := os.MkdirAll(filepath.Dir(lintPortsFile), 0o755)
	require.NoError(t, err)

	err = os.WriteFile(lintPortsFile, []byte(`package lint_ports

var LegacyPorts = []uint16{8888, 8889, 8890}
`), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go": {lintPortsFile},
	}

	err = Lint(logger, filesByExtension)
	require.NoError(t, err) // Should pass - lint_ports directory is excluded
}

func TestLint_FileOpenError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	nonExistentFile := filepath.Join(tempDir, "does_not_exist.go")

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go": {nonExistentFile},
	}

	// Should NOT error - file open errors are silently ignored.
	err := Lint(logger, filesByExtension)
	require.NoError(t, err, "Lint should not fail for non-existent files")
}

// TestLint_PortNumbersInText tests that text matching numbers in the port range doesn't cause issues.

func TestLint_PortNumbersInText(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "main.go")

	// File with numbers that are NOT legacy ports (testing regex matching).
	err := os.WriteFile(testFile, []byte(`package main

// Some random numbers that are not legacy ports.
const validPort = 8000
const anotherValid = 8700
const irrelevantNumber = 1234
`), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go": {testFile},
	}

	err = Lint(logger, filesByExtension)
	require.NoError(t, err, "File with valid ports should not trigger violations")
}

// TestLint_EmptyFilesByExtension tests Lint with empty input.

func TestLint_EmptyFilesByExtension(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{}

	err := Lint(logger, filesByExtension)
	require.NoError(t, err, "Empty input should not cause errors")
}

// TestLint_JsonFiles tests that Lint ignores JSON files (not in supported extensions).

func TestLint_JsonFiles(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "config.json")

	err := os.WriteFile(testFile, []byte(`{
  "port": 8888
}
`), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"json": {testFile},
	}

	// JSON files are NOT in the supported extension list, so they should be ignored.
	err = Lint(logger, filesByExtension)
	require.NoError(t, err, "JSON files are not supported by lint_ports, so no violations should be reported")
}

// TestLint_MultipleExtensions tests Lint with multiple file types having violations.

func TestLint_MultipleExtensions(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create a Go file with violation.
	goFile := filepath.Join(tempDir, "main.go")
	err := os.WriteFile(goFile, []byte(`package main
const port = 8888
`), 0o600)
	require.NoError(t, err)

	// Create a YAML file with violation.
	yamlFile := filepath.Join(tempDir, "config.yml")
	err = os.WriteFile(yamlFile, []byte(`port: 9443
`), 0o600)
	require.NoError(t, err)

	// Create another YAML file (yaml extension).
	yaml2File := filepath.Join(tempDir, "config.yaml")
	err = os.WriteFile(yaml2File, []byte(`port: 8443
`), 0o600)
	require.NoError(t, err)

	// Create a Markdown file with violation.
	mdFile := filepath.Join(tempDir, "README.md")
	err = os.WriteFile(mdFile, []byte(`# Docs
Port 8890 is used.
`), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go":   {goFile},
		"yml":  {yamlFile},
		"yaml": {yaml2File},
		"md":   {mdFile},
	}

	err = Lint(logger, filesByExtension)
	require.Error(t, err, "Multiple violations should be detected")
	require.Contains(t, err.Error(), "legacy port violations")
}

// TestLint_NonLegacyPortNumbers tests that non-legacy port numbers are not flagged.

func TestLint_NonLegacyPortNumbers(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "main.go")

	// File with port numbers in the valid range but not legacy ports.
	err := os.WriteFile(testFile, []byte(`package main

const port1 = 8000  // standard port (sm-kms)
const port2 = 8700  // standard port (sm-im)
const port3 = 9000  // random port
const port4 = 12345 // 5-digit port
`), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go": {testFile},
	}

	err = Lint(logger, filesByExtension)
	require.NoError(t, err, "Non-legacy ports should not trigger violations")
}

// TestIsOtelRelatedContent tests that OTEL-related terms in line content are detected.

func TestLint_SkipsCollectorPortsInMagicFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create a file with "magic" in name (NOT collector-related path)
	// but with collector-related content (comments or constant names contain collector terms).
	// This mirrors magic_network.go where the comment above contains "OpenTelemetry".
	magicFile := filepath.Join(tempDir, "magic_network.go")
	err := os.WriteFile(magicFile, []byte(`package magic

// Default OpenTelemetry collector internal metrics port (Prometheus).
const DefaultPublicPortInternalMetrics uint16 = 8888
// PortOtelCollectorReceivedMetrics - Default OpenTelemetry collector received metrics port.
const PortOtelCollectorReceivedMetrics uint16 = 8889
`), 0o600)
	require.NoError(t, err)
	require.False(t, lintPortsCommon.IsOtelRelatedFile(magicFile), "File path should NOT be otel-related")

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go": {magicFile},
	}

	// Should pass because line content contains OTEL-related terms.
	err = Lint(logger, filesByExtension)
	require.NoError(t, err, "OTEL ports in OTEL-related content should NOT be flagged")
}

// =============================================================================
// Host Port Range Validation Tests
// =============================================================================

func TestLint_AllThreeChecksPass(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create a valid Go file.
	goFile := filepath.Join(tempDir, "main.go")
	err := os.WriteFile(goFile, []byte(`package main

const port = 8700
`), 0o600)
	require.NoError(t, err)

	// Create a valid compose file.
	composeFile := filepath.Join(tempDir, "compose.yml")
	err = os.WriteFile(composeFile, []byte(`services:
  sm-im:
    ports:
      - "8700:8700"
    healthcheck:
      test: ["CMD", "wget", "-q", "-O", "/dev/null", "https://127.0.0.1:9090/admin/api/v1/livez"]
`), 0o600)
	require.NoError(t, err)

	// Create a valid Dockerfile.
	dockerfile := filepath.Join(tempDir, "Dockerfile")
	err = os.WriteFile(dockerfile, []byte(`FROM alpine:3.19
HEALTHCHECK CMD wget -q -O /dev/null https://127.0.0.1:9090/admin/api/v1/livez
`), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go":         {goFile},
		"yml":        {composeFile},
		"dockerfile": {dockerfile},
	}

	err = Lint(logger, filesByExtension)
	require.NoError(t, err)
}

func TestLint_LegacyPortViolation(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	goFile := filepath.Join(tempDir, "main.go")
	err := os.WriteFile(goFile, []byte(`package main

const port = 9443
`), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go": {goFile},
	}

	err = Lint(logger, filesByExtension)
	require.Error(t, err)
	require.Contains(t, err.Error(), "legacy port")
}
