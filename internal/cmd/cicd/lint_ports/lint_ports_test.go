// Copyright (c) 2025 Justin Cranford

package lint_ports

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/cmd/cicd/common"

	"github.com/stretchr/testify/require"
)

func TestLint_NoLegacyPorts(t *testing.T) {
	t.Parallel()

	// Create a temp directory with clean files.
	tempDir := t.TempDir()

	// Create a Go file with standardized ports only.
	goFile := filepath.Join(tempDir, "main.go")
	err := os.WriteFile(goFile, []byte(`package main

const port = 8070 // cipher-im standardized port
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

const port = 8888 // legacy cipher-im port
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

	// Create a regular file with a cipher-im legacy port.
	// Use 8890 which is NOT an OTEL collector port.
	normalFile := filepath.Join(tempDir, "config.go")
	err := os.WriteFile(normalFile, []byte(`package main

const port = 8890 // legacy cipher-im port
`), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go": {normalFile},
	}

	err = Lint(logger, filesByExtension)
	require.Error(t, err) // Should fail - 8890 is legacy cipher-im port
}

func TestLint_Detects8888InRegularFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create a regular file (not in telemetry/observability path) with port 8888.
	// 8888 is both a cipher-im legacy port AND an OTEL collector port.
	// It should be detected as legacy in non-observability files.
	normalFile := filepath.Join(tempDir, "main.go")
	err := os.WriteFile(normalFile, []byte(`package main

const port = 8888
`), 0o600)
	require.NoError(t, err)

	// Verify the file path does NOT contain OTEL-related terms.
	require.False(t, isOtelRelatedFile(normalFile), "Test file path should NOT be telemetry-related")

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

func TestAllLegacyPorts(t *testing.T) {
	t.Parallel()

	ports := AllLegacyPorts()

	// Verify known legacy ports are included.
	require.Contains(t, ports, uint16(8888)) // cipher-im legacy
	require.Contains(t, ports, uint16(8889)) // cipher-im legacy
	require.Contains(t, ports, uint16(8890)) // cipher-im legacy
	require.Contains(t, ports, uint16(9443)) // jose-ja legacy
	require.Contains(t, ports, uint16(8092)) // jose-ja legacy
	require.Contains(t, ports, uint16(8443)) // pki-ca legacy
}

func TestAllValidPublicPorts(t *testing.T) {
	t.Parallel()

	ports := AllValidPublicPorts()

	// Verify standardized ports are included.
	require.Contains(t, ports, uint16(8070))  // cipher-im
	require.Contains(t, ports, uint16(8071))  // cipher-im
	require.Contains(t, ports, uint16(8072))  // cipher-im
	require.Contains(t, ports, uint16(8060))  // jose-ja
	require.Contains(t, ports, uint16(8050))  // pki-ca
	require.Contains(t, ports, uint16(8080))  // sm-kms
	require.Contains(t, ports, uint16(8081))  // sm-kms
	require.Contains(t, ports, uint16(8082))  // sm-kms
	require.Contains(t, ports, uint16(18000)) // identity-authz
	require.Contains(t, ports, uint16(18100)) // identity-idp
	require.Contains(t, ports, uint16(18200)) // identity-rs
	require.Contains(t, ports, uint16(18300)) // identity-rp
	require.Contains(t, ports, uint16(18400)) // identity-spa
}

func TestIsOtelCollectorPort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		port uint16
		want bool
	}{
		{name: "OTEL internal metrics", port: 8888, want: true},
		{name: "OTEL Prometheus", port: 8889, want: true},
		{name: "cipher-im standardized", port: 8070, want: false},
		{name: "jose-ja standardized", port: 8060, want: false},
		{name: "random port", port: 12345, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := IsOtelCollectorPort(tt.port)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestIsOtelRelatedFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filePath string
		want     bool
	}{
		{name: "otel in path", filePath: "/path/to/otel_config.go", want: true},
		{name: "opentelemetry in path", filePath: "/path/opentelemetry/main.go", want: true},
		{name: "telemetry in path", filePath: "/internal/telemetry/setup.go", want: true},
		{name: "regular go file", filePath: "/internal/server/main.go", want: false},
		{name: "config yaml", filePath: "/configs/app.yml", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := isOtelRelatedFile(tt.filePath)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestGetServiceForLegacyPort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		port uint16
		want string
	}{
		{name: "cipher-im 8888", port: 8888, want: "cipher-im"},
		{name: "cipher-im 8889", port: 8889, want: "cipher-im"},
		{name: "cipher-im 8890", port: 8890, want: "cipher-im"},
		{name: "jose-ja 9443", port: 9443, want: "jose-ja"},
		{name: "jose-ja 8092", port: 8092, want: "jose-ja"},
		{name: "pki-ca 8443", port: 8443, want: "pki-ca"},
		{name: "unknown port", port: 12345, want: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := getServiceForLegacyPort(tt.port)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestServicePorts_AllServicesPresent(t *testing.T) {
	t.Parallel()

	expectedServices := []string{
		"cipher-im",
		"jose-ja",
		"pki-ca",
		"sm-kms",
		"identity-authz",
		"identity-idp",
		"identity-rs",
		"identity-rp",
		"identity-spa",
	}

	for _, svc := range expectedServices {
		t.Run(svc, func(t *testing.T) {
			t.Parallel()

			cfg, ok := ServicePorts[svc]
			require.True(t, ok, "Service %s should be in ServicePorts", svc)
			require.Equal(t, svc, cfg.Name)
			require.Equal(t, StandardAdminPort, cfg.AdminPort)
			require.NotEmpty(t, cfg.PublicPorts)
		})
	}
}

// TestLint_FileOpenError tests that checkFile handles file open errors gracefully.
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
const validPort = 8080
const anotherValid = 8070
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

const port1 = 8080  // standard port
const port2 = 8081  // standard port
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
func TestIsOtelRelatedContent(t *testing.T) {
t.Parallel()

tests := []struct {
name    string
content string
want    bool
}{
{name: "otel in constant name", content: "PortOtelCollectorReceivedMetrics uint16 = 8889", want: true},
{name: "telemetry in comment", content: "// OpenTelemetry metrics port", want: true},
{name: "opentelemetry in text", content: "// Use OpenTelemetry for observability", want: true},
{name: "OTEL uppercase", content: "const OTEL_PORT = 8888", want: true},
{name: "no otel terms", content: "const port = 8080", want: false},
{name: "cipher-im port", content: "const cipherPort = 8888", want: false},
{name: "empty line", content: "", want: false},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
t.Parallel()

got := isOtelRelatedContent(tt.content)
require.Equal(t, tt.want, got)
})
}
}

// TestLint_SkipsCollectorPortsInMagicFile tests that collector ports are skipped
// when the line content contains related terms (even if file path doesn't).
// NOTE: Function name avoids "otel/telemetry" to prevent t.TempDir() path matching.
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
require.False(t, isOtelRelatedFile(magicFile), "File path should NOT be otel-related")

logger := cryptoutilCmdCicdCommon.NewLogger("test")
filesByExtension := map[string][]string{
"go": {magicFile},
}

// Should pass because line content contains OTEL-related terms.
err = Lint(logger, filesByExtension)
require.NoError(t, err, "OTEL ports in OTEL-related content should NOT be flagged")
}
