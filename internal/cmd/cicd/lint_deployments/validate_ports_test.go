package lint_deployments

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testComposeServicePort = "services:\n  my-service:\n    ports:\n      - \"8700:8080\"\n"

// createDeploymentWithCompose creates a deployment dir with a compose file.
func createDeploymentWithCompose(t *testing.T, composeContent string) string {
t.Helper()

dir := t.TempDir()
require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), []byte(composeContent), 0o600))

return dir
}

func TestValidatePorts_ServiceLevelValid(t *testing.T) {
t.Parallel()

compose := "services:\n  my-service:\n    ports:\n      - \"8700:8080\"\n      - \"8701:8080\"\n"
dir := createDeploymentWithCompose(t, compose)

result, err := ValidatePorts(dir, "cipher-im", DeploymentTypeProductService)
require.NoError(t, err)
require.NotNil(t, result)
assert.True(t, result.Valid)
assert.Empty(t, result.Errors)
}

func TestValidatePorts_ServiceLevelOutOfRange(t *testing.T) {
t.Parallel()

compose := "services:\n  my-service:\n    ports:\n      - \"18700:8080\"\n"
dir := createDeploymentWithCompose(t, compose)

result, err := ValidatePorts(dir, "cipher-im", DeploymentTypeProductService)
require.NoError(t, err)
require.NotNil(t, result)
assert.False(t, result.Valid)
assert.True(t, containsSubstring(result.Errors, "outside PRODUCT-SERVICE range"))
}

func TestValidatePorts_ProductLevelValid(t *testing.T) {
t.Parallel()

compose := "services:\n  cipher-im-sqlite:\n    ports:\n      - \"18700:8080\"\n  cipher-im-pg-1:\n    ports:\n      - \"18701:8080\"\n"
dir := createDeploymentWithCompose(t, compose)

result, err := ValidatePorts(dir, "cipher", DeploymentTypeProduct)
require.NoError(t, err)
require.NotNil(t, result)
assert.True(t, result.Valid)
}

func TestValidatePorts_ProductLevelOutOfRange(t *testing.T) {
t.Parallel()

compose := testComposeServicePort
dir := createDeploymentWithCompose(t, compose)

result, err := ValidatePorts(dir, "cipher", DeploymentTypeProduct)
require.NoError(t, err)
require.NotNil(t, result)
assert.False(t, result.Valid)
assert.True(t, containsSubstring(result.Errors, "outside PRODUCT range"))
}

func TestValidatePorts_SuiteLevelValid(t *testing.T) {
t.Parallel()

compose := "services:\n  sm-kms-sqlite:\n    ports:\n      - \"28000:8080\"\n"
dir := createDeploymentWithCompose(t, compose)

result, err := ValidatePorts(dir, "cryptoutil", DeploymentTypeSuite)
require.NoError(t, err)
require.NotNil(t, result)
assert.True(t, result.Valid)
}

func TestValidatePorts_SuiteLevelOutOfRange(t *testing.T) {
t.Parallel()

compose := "services:\n  sm-kms-sqlite:\n    ports:\n      - \"8000:8080\"\n"
dir := createDeploymentWithCompose(t, compose)

result, err := ValidatePorts(dir, "cryptoutil", DeploymentTypeSuite)
require.NoError(t, err)
require.NotNil(t, result)
assert.False(t, result.Valid)
assert.True(t, containsSubstring(result.Errors, "outside SUITE range"))
}

func TestValidatePorts_InfrastructurePortsSkipped(t *testing.T) {
t.Parallel()

compose := "services:\n  postgres:\n    ports:\n      - \"5432:5432\"\n  grafana:\n    ports:\n      - \"3000:3000\"\n  otel:\n    ports:\n      - \"4317:4317\"\n      - \"4318:4318\"\n      - \"14317:4317\"\n      - \"14318:4318\"\n      - \"13133:13133\"\n"
dir := createDeploymentWithCompose(t, compose)

result, err := ValidatePorts(dir, "cipher-im", DeploymentTypeProductService)
require.NoError(t, err)
require.NotNil(t, result)
assert.True(t, result.Valid)
}

func TestValidatePorts_ConfigPortValidation(t *testing.T) {
t.Parallel()

compose := testComposeServicePort
dir := createDeploymentWithCompose(t, compose)

// Add config directory with a config file.
configDir := filepath.Join(dir, "config")
require.NoError(t, os.MkdirAll(configDir, 0o755))
require.NoError(t, os.WriteFile(filepath.Join(configDir, "config-sqlite.yml"),
[]byte("bind-public-port: 8700\n"), 0o600))

result, err := ValidatePorts(dir, "cipher-im", DeploymentTypeProductService)
require.NoError(t, err)
require.NotNil(t, result)
assert.True(t, result.Valid)
}

func TestValidatePorts_ConfigPortOutOfRange(t *testing.T) {
t.Parallel()

compose := testComposeServicePort
dir := createDeploymentWithCompose(t, compose)

configDir := filepath.Join(dir, "config")
require.NoError(t, os.MkdirAll(configDir, 0o755))
require.NoError(t, os.WriteFile(filepath.Join(configDir, "config-sqlite.yml"),
[]byte("bind-public-port: 18700\n"), 0o600))

result, err := ValidatePorts(dir, "cipher-im", DeploymentTypeProductService)
require.NoError(t, err)
require.NotNil(t, result)
assert.False(t, result.Valid)
assert.True(t, containsSubstring(result.Errors, "bind-public-port"))
}

func TestValidatePorts_PathNotFound(t *testing.T) {
t.Parallel()

result, err := ValidatePorts(filepath.Join(t.TempDir(), "nonexistent"), "x", DeploymentTypeProductService)
require.NoError(t, err)
require.NotNil(t, result)
assert.False(t, result.Valid)
assert.True(t, containsSubstring(result.Errors, "does not exist"))
}

func TestValidatePorts_PathIsFile(t *testing.T) {
t.Parallel()

f := filepath.Join(t.TempDir(), "file.txt")
require.NoError(t, os.WriteFile(f, []byte("data"), 0o600))

result, err := ValidatePorts(f, "x", DeploymentTypeProductService)
require.NoError(t, err)
require.NotNil(t, result)
assert.False(t, result.Valid)
assert.True(t, containsSubstring(result.Errors, "not a directory"))
}

func TestValidatePorts_NoComposeFile(t *testing.T) {
t.Parallel()

dir := t.TempDir()
result, err := ValidatePorts(dir, "cipher-im", DeploymentTypeProductService)
require.NoError(t, err)
require.NotNil(t, result)
// No compose.yml -> no port validation errors (compose existence checked by other validators).
assert.True(t, result.Valid)
}

func TestValidatePorts_InvalidComposeYAML(t *testing.T) {
t.Parallel()

dir := createDeploymentWithCompose(t, "invalid: [yaml: {broken")

result, err := ValidatePorts(dir, "cipher-im", DeploymentTypeProductService)
require.NoError(t, err)
require.NotNil(t, result)
assert.True(t, result.Valid)
assert.True(t, containsSubstring(result.Warnings, "Cannot parse"))
}

func TestValidatePorts_NonNumericPortSkipped(t *testing.T) {
t.Parallel()

compose := "services:\n  my-service:\n    ports:\n      - \"abc:8080\"\n      - \"8700:8080\"\n"
dir := createDeploymentWithCompose(t, compose)

result, err := ValidatePorts(dir, "cipher-im", DeploymentTypeProductService)
require.NoError(t, err)
require.NotNil(t, result)
assert.True(t, result.Valid)
}

func TestValidatePorts_ContainerOnlyPortSkipped(t *testing.T) {
t.Parallel()

compose := "services:\n  my-service:\n    ports:\n      - \"8080\"\n      - \"8700:8080\"\n"
dir := createDeploymentWithCompose(t, compose)

result, err := ValidatePorts(dir, "cipher-im", DeploymentTypeProductService)
require.NoError(t, err)
require.NotNil(t, result)
assert.True(t, result.Valid)
}

func TestValidatePorts_ConfigDirWithNonYAML(t *testing.T) {
t.Parallel()

compose := testComposeServicePort
dir := createDeploymentWithCompose(t, compose)

configDir := filepath.Join(dir, "config")
require.NoError(t, os.MkdirAll(configDir, 0o755))
require.NoError(t, os.WriteFile(filepath.Join(configDir, "README.md"),
[]byte("# readme\n"), 0o600))

result, err := ValidatePorts(dir, "cipher-im", DeploymentTypeProductService)
require.NoError(t, err)
require.NotNil(t, result)
assert.True(t, result.Valid)
}

func TestValidatePorts_ConfigInvalidYAML(t *testing.T) {
t.Parallel()

compose := testComposeServicePort
dir := createDeploymentWithCompose(t, compose)

configDir := filepath.Join(dir, "config")
require.NoError(t, os.MkdirAll(configDir, 0o755))
require.NoError(t, os.WriteFile(filepath.Join(configDir, "broken.yml"),
[]byte("invalid: [yaml: {broken"), 0o600))

result, err := ValidatePorts(dir, "cipher-im", DeploymentTypeProductService)
require.NoError(t, err)
require.NotNil(t, result)
assert.True(t, result.Valid)
}

func TestValidatePorts_ConfigNonIntPort(t *testing.T) {
t.Parallel()

compose := testComposeServicePort
dir := createDeploymentWithCompose(t, compose)

configDir := filepath.Join(dir, "config")
require.NoError(t, os.MkdirAll(configDir, 0o755))
require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.yml"),
[]byte("bind-public-port: \"not-a-number\"\n"), 0o600))

result, err := ValidatePorts(dir, "cipher-im", DeploymentTypeProductService)
require.NoError(t, err)
require.NotNil(t, result)
assert.True(t, result.Valid) // String not parseable, skip validation.
}

func TestValidatePorts_ConfigUnreadableFile(t *testing.T) {
	t.Parallel()

	compose := testComposeServicePort
	dir := createDeploymentWithCompose(t, compose)

	configDir := filepath.Join(dir, "config")
	require.NoError(t, os.MkdirAll(configDir, 0o755))

	// Create a symlink to a non-existent target so ReadFile fails.
	require.NoError(t, os.Symlink("/nonexistent/file.yml", filepath.Join(configDir, "broken.yml")))

	result, err := ValidatePorts(dir, "cipher-im", DeploymentTypeProductService)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Valid) // Unreadable files are silently skipped.
}

func TestValidatePorts_ConfigDirWithSubdirectory(t *testing.T) {
	t.Parallel()

	compose := testComposeServicePort
	dir := createDeploymentWithCompose(t, compose)

	configDir := filepath.Join(dir, "config")
	require.NoError(t, os.MkdirAll(filepath.Join(configDir, "subdir"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "valid.yml"),
		[]byte("bind-public-port: 8700\n"), 0o600))

	result, err := ValidatePorts(dir, "cipher-im", DeploymentTypeProductService)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Valid) // Subdirectory skipped, valid port accepted.
}

func TestValidateConfigPortRanges_UnreadableDir(t *testing.T) {
	t.Parallel()

	result := &PortValidationResult{Valid: true}
	validateConfigPortRanges("/nonexistent/path", "cipher-im", DeploymentTypeProductService, result)
	assert.True(t, result.Valid) // ReadDir error is silently handled.
}

func TestGetPortRange(t *testing.T) {
t.Parallel()

tests := []struct {
name     string
level    string
wantMin  int
wantMax  int
}{
{name: "service", level: DeploymentTypeProductService, wantMin: servicePortMin, wantMax: servicePortMax},
{name: "product", level: DeploymentTypeProduct, wantMin: productPortMin, wantMax: productPortMax},
{name: "suite", level: DeploymentTypeSuite, wantMin: suitePortMin, wantMax: suitePortMax},
{name: "unknown defaults to service", level: "unknown", wantMin: servicePortMin, wantMax: servicePortMax},
}

for _, tc := range tests {
t.Run(tc.name, func(t *testing.T) {
t.Parallel()

gotMin, gotMax := getPortRange(tc.level)
assert.Equal(t, tc.wantMin, gotMin)
assert.Equal(t, tc.wantMax, gotMax)
})
}
}

func TestIsInfrastructurePort(t *testing.T) {
t.Parallel()

tests := []struct {
name string
port int
want bool
}{
{name: "postgres", port: 5432, want: true},
{name: "grafana", port: 3000, want: true},
{name: "otlp-grpc", port: 4317, want: true},
{name: "otlp-http", port: 4318, want: true},
{name: "otel-health", port: 13133, want: true},
{name: "otlp-grpc-fwd", port: 14317, want: true},
{name: "otlp-http-fwd", port: 14318, want: true},
{name: "service-port", port: 8700, want: false},
{name: "zero", port: 0, want: false},
}

for _, tc := range tests {
t.Run(tc.name, func(t *testing.T) {
t.Parallel()

assert.Equal(t, tc.want, isInfrastructurePort(tc.port))
})
}
}

func TestFormatPortValidationResult(t *testing.T) {
t.Parallel()

tests := []struct {
name     string
result   *PortValidationResult
contains []string
}{
{
name:     "passing",
result:   &PortValidationResult{Path: "/deploy", Valid: true},
contains: []string{"PASS", "/deploy"},
},
{
name: "failing",
result: &PortValidationResult{
Path: "/deploy", Valid: false,
Errors:   []string{"port out of range"},
Warnings: []string{"cannot parse"},
},
contains: []string{"FAIL", "port out of range", "cannot parse"},
},
}

for _, tc := range tests {
t.Run(tc.name, func(t *testing.T) {
t.Parallel()

output := FormatPortValidationResult(tc.result)
for _, s := range tc.contains {
assert.Contains(t, output, s)
}
})
}
}

func TestValidatePorts_RealCipherIM(t *testing.T) {
t.Parallel()

deplPath := filepath.Join(".", "..", "..", "..", "..", "deployments", "cipher-im")
info, err := os.Stat(deplPath)

if err != nil || !info.IsDir() {
t.Skip("Real cipher-im deployment not found - skipping integration test")
}

result, err := ValidatePorts(deplPath, "cipher-im", DeploymentTypeProductService)
require.NoError(t, err)
require.NotNil(t, result)
assert.True(t, result.Valid, "Real cipher-im should pass port validation. Errors: %v", result.Errors)
}
