package lint_deployments

import (
"os"
"path/filepath"
"testing"

"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"
)

func TestValidateTelemetry_ValidConfigs(t *testing.T) {
t.Parallel()

dir := t.TempDir()
writeConfig(t, dir, "config-1.yml", map[string]string{
"otlp":             "true",
"otlp-service":     "svc-1",
"otlp-endpoint":    "http://collector:4317",
"otlp-environment": "development",
})
writeConfig(t, dir, "config-2.yml", map[string]string{
"otlp":             "true",
"otlp-service":     "svc-2",
"otlp-endpoint":    "http://collector:4317",
"otlp-environment": "development",
})

result, err := ValidateTelemetry(dir)
require.NoError(t, err)
require.NotNil(t, result)
assert.True(t, result.Valid)
assert.Empty(t, result.Errors)
}

func TestValidateTelemetry_PathNotFound(t *testing.T) {
t.Parallel()

result, err := ValidateTelemetry("/nonexistent/path")
require.NoError(t, err)
require.NotNil(t, result)
assert.False(t, result.Valid)
assert.Len(t, result.Errors, 1)
assert.Contains(t, result.Errors[0], "not found")
}

func TestValidateTelemetry_PathIsFile(t *testing.T) {
t.Parallel()

f := filepath.Join(t.TempDir(), "file.yml")
require.NoError(t, os.WriteFile(f, []byte("test"), 0o600))

result, err := ValidateTelemetry(f)
require.NoError(t, err)
require.NotNil(t, result)
assert.False(t, result.Valid)
assert.Contains(t, result.Errors[0], "not a directory")
}

func TestValidateTelemetry_EmptyEndpoint(t *testing.T) {
t.Parallel()

dir := t.TempDir()
writeConfig(t, dir, "config.yml", map[string]string{
"otlp":         "true",
"otlp-service": "svc-1",
})

result, err := ValidateTelemetry(dir)
require.NoError(t, err)
assert.False(t, result.Valid)
assert.Contains(t, result.Errors[0], "empty")
}

func TestValidateTelemetry_InvalidEndpointURL(t *testing.T) {
t.Parallel()

dir := t.TempDir()
writeConfig(t, dir, "config.yml", map[string]string{
"otlp":          "true",
"otlp-service":  "svc-1",
"otlp-endpoint": "://bad-url",
})

result, err := ValidateTelemetry(dir)
require.NoError(t, err)
assert.False(t, result.Valid)
assert.Contains(t, result.Errors[0], "invalid")
}

func TestValidateTelemetry_EndpointMissingHost(t *testing.T) {
t.Parallel()

dir := t.TempDir()
writeConfig(t, dir, "config.yml", map[string]string{
"otlp":          "true",
"otlp-service":  "svc-1",
"otlp-endpoint": "http://:4317",
})

result, err := ValidateTelemetry(dir)
require.NoError(t, err)
assert.False(t, result.Valid)
assert.Contains(t, result.Errors[0], "missing host")
}

func TestValidateTelemetry_EndpointMissingPort(t *testing.T) {
t.Parallel()

dir := t.TempDir()
writeConfig(t, dir, "config.yml", map[string]string{
"otlp":          "true",
"otlp-service":  "svc-1",
"otlp-endpoint": "http://collector",
})

result, err := ValidateTelemetry(dir)
require.NoError(t, err)
assert.True(t, result.Valid)
assert.Contains(t, result.Warnings[0], "missing port")
}

func TestValidateTelemetry_UnusualScheme(t *testing.T) {
t.Parallel()

dir := t.TempDir()
writeConfig(t, dir, "config.yml", map[string]string{
"otlp":          "true",
"otlp-service":  "svc-1",
"otlp-endpoint": "ftp://collector:4317",
})

result, err := ValidateTelemetry(dir)
require.NoError(t, err)
assert.True(t, result.Valid)
assert.Contains(t, result.Warnings[0], "unusual scheme")
}

func TestValidateTelemetry_AcceptedSchemes(t *testing.T) {
t.Parallel()

tests := []struct {
name     string
endpoint string
}{
{name: "http", endpoint: "http://collector:4317"},
{name: "https", endpoint: "https://collector:4318"},
{name: "grpc", endpoint: "grpc://collector:4317"},
}

for _, tc := range tests {
t.Run(tc.name, func(t *testing.T) {
t.Parallel()

dir := t.TempDir()
writeConfig(t, dir, "config.yml", map[string]string{
"otlp":          "true",
"otlp-service":  "svc-1",
"otlp-endpoint": tc.endpoint,
})

result, err := ValidateTelemetry(dir)
require.NoError(t, err)
assert.True(t, result.Valid)

// No "unusual scheme" warnings for accepted schemes.
for _, w := range result.Warnings {
assert.NotContains(t, w, "unusual scheme")
}
})
}
}

func TestValidateTelemetry_DuplicateServiceNames(t *testing.T) {
t.Parallel()

dir := t.TempDir()
writeConfig(t, dir, "config-1.yml", map[string]string{
"otlp":          "true",
"otlp-service":  "same-name",
"otlp-endpoint": "http://collector:4317",
})
writeConfig(t, dir, "config-2.yml", map[string]string{
"otlp":          "true",
"otlp-service":  "same-name",
"otlp-endpoint": "http://collector:4317",
})

result, err := ValidateTelemetry(dir)
require.NoError(t, err)
assert.True(t, result.Valid) // Duplicates are warnings, not errors.
assert.NotEmpty(t, result.Warnings)
assert.Contains(t, result.Warnings[0], "Duplicate otlp-service")
}

func TestValidateTelemetry_InconsistentEndpoints(t *testing.T) {
t.Parallel()

dir := t.TempDir()
writeConfig(t, dir, "config-1.yml", map[string]string{
"otlp":          "true",
"otlp-service":  "svc-1",
"otlp-endpoint": "http://collector-a:4317",
})
writeConfig(t, dir, "config-2.yml", map[string]string{
"otlp":          "true",
"otlp-service":  "svc-2",
"otlp-endpoint": "http://collector-b:4317",
})

result, err := ValidateTelemetry(dir)
require.NoError(t, err)
assert.True(t, result.Valid) // Inconsistency is a warning.
assert.NotEmpty(t, result.Warnings)
assert.Contains(t, result.Warnings[0], "Inconsistent")
}

func TestValidateTelemetry_DisabledOTLP(t *testing.T) {
t.Parallel()

dir := t.TempDir()
writeConfig(t, dir, "config.yml", map[string]string{
"otlp": "false",
})

result, err := ValidateTelemetry(dir)
require.NoError(t, err)
assert.True(t, result.Valid)
assert.Empty(t, result.Errors)
}

func TestValidateTelemetry_NonBoolOTLP(t *testing.T) {
t.Parallel()

dir := t.TempDir()
require.NoError(t, os.WriteFile(
filepath.Join(dir, "config.yml"),
[]byte("otlp: \"yes\"\n"), 0o600))

result, err := ValidateTelemetry(dir)
require.NoError(t, err)
assert.True(t, result.Valid) // Non-bool otlp skipped entirely.
}

func TestValidateTelemetry_NoOTLPField(t *testing.T) {
t.Parallel()

dir := t.TempDir()
require.NoError(t, os.WriteFile(
filepath.Join(dir, "config.yml"),
[]byte("bind-public-port: 8080\n"), 0o600))

result, err := ValidateTelemetry(dir)
require.NoError(t, err)
assert.True(t, result.Valid)
}

func TestValidateTelemetry_InvalidYAML(t *testing.T) {
t.Parallel()

dir := t.TempDir()
require.NoError(t, os.WriteFile(
filepath.Join(dir, "config.yml"),
[]byte("{{invalid yaml"), 0o600))

result, err := ValidateTelemetry(dir)
require.NoError(t, err)
assert.True(t, result.Valid) // Invalid YAML skipped.
}

func TestValidateTelemetry_UnreadableFile(t *testing.T) {
t.Parallel()

dir := t.TempDir()
require.NoError(t, os.Symlink("/nonexistent/broken.yml",
filepath.Join(dir, "broken.yml")))

result, err := ValidateTelemetry(dir)
require.NoError(t, err)
assert.True(t, result.Valid) // Unreadable files skipped.
}

func TestValidateTelemetry_SubdirectorySkipped(t *testing.T) {
t.Parallel()

dir := t.TempDir()
require.NoError(t, os.MkdirAll(filepath.Join(dir, "subdir"), 0o755))
writeConfig(t, dir, "config.yml", map[string]string{
"otlp":          "true",
"otlp-service":  "svc-1",
"otlp-endpoint": "http://collector:4317",
})

result, err := ValidateTelemetry(dir)
require.NoError(t, err)
assert.True(t, result.Valid)
}

func TestValidateTelemetry_NonYAMLSkipped(t *testing.T) {
t.Parallel()

dir := t.TempDir()
require.NoError(t, os.WriteFile(
filepath.Join(dir, "readme.txt"),
[]byte("not yaml"), 0o600))

result, err := ValidateTelemetry(dir)
require.NoError(t, err)
assert.True(t, result.Valid)
}

func TestValidateTelemetry_UnreadableDir(t *testing.T) {
t.Parallel()

result := &TelemetryValidationResult{Valid: true}
entries := collectOTLPEntries("/nonexistent/dir", result)
assert.Nil(t, entries)
assert.True(t, result.Valid)
}

func TestValidateTelemetry_EndpointNormalization(t *testing.T) {
t.Parallel()

dir := t.TempDir()
writeConfig(t, dir, "config-1.yml", map[string]string{
"otlp":          "true",
"otlp-service":  "svc-1",
"otlp-endpoint": "http://collector:4317/",
})
writeConfig(t, dir, "config-2.yml", map[string]string{
"otlp":          "true",
"otlp-service":  "svc-2",
"otlp-endpoint": "http://collector:4317",
})

result, err := ValidateTelemetry(dir)
require.NoError(t, err)
assert.True(t, result.Valid)

// After normalization, trailing slash should not cause inconsistency.
for _, w := range result.Warnings {
assert.NotContains(t, w, "Inconsistent")
}
}

func TestNormalizeEndpoint(t *testing.T) {
t.Parallel()

tests := []struct {
name     string
input    string
expected string
}{
{name: "no trailing slash", input: "http://host:4317", expected: "http://host:4317"},
{name: "single trailing slash", input: "http://host:4317/", expected: "http://host:4317"},
{name: "multiple trailing slashes", input: "http://host:4317///", expected: "http://host:4317"},
{name: "empty", input: "", expected: ""},
}

for _, tc := range tests {
t.Run(tc.name, func(t *testing.T) {
t.Parallel()
assert.Equal(t, tc.expected, normalizeEndpoint(tc.input))
})
}
}

func TestFormatTelemetryValidationResult(t *testing.T) {
t.Parallel()

tests := []struct {
name     string
result   *TelemetryValidationResult
contains []string
}{
{
name:     "nil result",
result:   nil,
contains: []string{"No telemetry validation result"},
},
{
name:     "valid",
result:   &TelemetryValidationResult{Valid: true},
contains: []string{"PASSED"},
},
{
name: "errors and warnings",
result: &TelemetryValidationResult{
Valid:    false,
Errors:   []string{"err1"},
Warnings: []string{"warn1"},
},
contains: []string{"FAILED", "ERROR: err1", "WARNING: warn1"},
},
}

for _, tc := range tests {
t.Run(tc.name, func(t *testing.T) {
t.Parallel()

output := FormatTelemetryValidationResult(tc.result)
for _, want := range tc.contains {
assert.Contains(t, output, want)
}
})
}
}

func TestValidateTelemetry_RealSmIM(t *testing.T) {
t.Parallel()

configDir := filepath.Join("testdata", "configs", "sm", "im")
if _, err := os.Stat(configDir); err != nil {
configDir = filepath.Join("..", "..", "..", "..", "configs", "sm", "im")
}

if _, err := os.Stat(configDir); err != nil {
t.Skip("Real sm-im configs not found")
}

result, err := ValidateTelemetry(configDir)
require.NoError(t, err)
require.NotNil(t, result)

// Real configs should be valid (may have warnings).
assert.True(t, result.Valid, "Real sm-im OTLP config validation failed: %v", result.Errors)
}

// writeConfig creates a YAML config file from key-value pairs.
func writeConfig(t *testing.T, dir, name string, fields map[string]string) {
t.Helper()

var content string

for k, v := range fields {
if v == "true" || v == "false" {
content += k + ": " + v + "\n"
} else {
content += k + ": \"" + v + "\"\n"
}
}

require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
}
