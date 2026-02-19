package lint_deployments

import (
"fmt"
"os"
"path/filepath"
"strings"

cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

"gopkg.in/yaml.v3"
)

// adminPort is the mandatory admin port for all services.
const adminPort = int(cryptoutilSharedMagic.DefaultPrivatePortCryptoutil)

// AdminValidationResult holds results from ValidateAdmin.
type AdminValidationResult struct {
Valid    bool
Errors   []string
Warnings []string
}

// ValidateAdmin validates admin endpoint configuration across a deployment.
// It checks that admin ports are not exposed in compose files and that config
// files use the correct admin bind address and port.
func ValidateAdmin(deploymentPath string) (*AdminValidationResult, error) {
result := &AdminValidationResult{Valid: true}

info, err := os.Stat(deploymentPath)
if err != nil {
result.Errors = append(result.Errors,
fmt.Sprintf("[ValidateAdmin] Deployment path not found: %s", deploymentPath))
result.Valid = false

return result, nil
}

if !info.IsDir() {
result.Errors = append(result.Errors,
fmt.Sprintf("[ValidateAdmin] Path is not a directory: %s", deploymentPath))
result.Valid = false

return result, nil
}

composePath := filepath.Join(deploymentPath, "compose.yml")
if _, statErr := os.Stat(composePath); statErr == nil {
validateAdminNotExposed(composePath, result)
}

configDir := filepath.Join(deploymentPath, "config")
if info, statErr := os.Stat(configDir); statErr == nil && info.IsDir() {
validateAdminConfigSettings(configDir, result)
}

return result, nil
}

// validateAdminNotExposed checks that admin port 9090 is NOT in compose port mappings.
func validateAdminNotExposed(composePath string, result *AdminValidationResult) {
data, err := os.ReadFile(composePath)
if err != nil {
return
}

var compose composeFile
if err := yaml.Unmarshal(data, &compose); err != nil {
return
}

for svcName, svc := range compose.Services {
for _, portMapping := range svc.Ports {
hostPort := extractHostPort(portMapping)
if hostPort == "" {
continue
}

port, ok := toInt(hostPort)
if !ok {
continue
}

if port == adminPort {
result.Errors = append(result.Errors,
fmt.Sprintf("[ValidateAdmin] Service '%s' exposes admin port %d to host in compose (SECURITY VIOLATION)",
svcName, adminPort))
result.Valid = false
}
}
}
}

// validateAdminConfigSettings checks config files for correct admin bind settings.
func validateAdminConfigSettings(configDir string, result *AdminValidationResult) {
entries, err := os.ReadDir(configDir)
if err != nil {
return
}

for _, entry := range entries {
if entry.IsDir() || !isYAMLFile(entry.Name()) {
continue
}

configPath := filepath.Join(configDir, entry.Name())

data, err := os.ReadFile(configPath)
if err != nil {
continue
}

var config map[string]any
if err := yaml.Unmarshal(data, &config); err != nil {
continue
}

validateAdminPort(config, configPath, result)
validateAdminAddress(config, configPath, result)
}
}

// validateAdminPort checks that bind-private-port is the mandatory admin port.
func validateAdminPort(config map[string]any, configPath string, result *AdminValidationResult) {
val, ok := config["bind-private-port"]
if !ok {
return
}

port, ok := toInt(val)
if !ok {
return
}

if port != adminPort {
result.Errors = append(result.Errors,
fmt.Sprintf("[ValidateAdmin] '%s': bind-private-port is %d, MUST be %d",
filepath.Base(configPath), port, adminPort))
result.Valid = false
}
}

// validateAdminAddress checks that bind-private-address is 127.0.0.1.
func validateAdminAddress(config map[string]any, configPath string, result *AdminValidationResult) {
val, ok := config["bind-private-address"]
if !ok {
return
}

strVal, ok := val.(string)
if !ok {
return
}

	if strVal != cryptoutilSharedMagic.IPv4Loopback {
		result.Errors = append(result.Errors,
			fmt.Sprintf("[ValidateAdmin] '%s': bind-private-address is %q, MUST be %q",
				filepath.Base(configPath), strVal, cryptoutilSharedMagic.IPv4Loopback))
result.Valid = false
}
}

// FormatAdminValidationResult formats the result for human-readable output.
func FormatAdminValidationResult(result *AdminValidationResult) string {
if result == nil {
return "No admin validation result"
}

var sb strings.Builder

if result.Valid {
sb.WriteString("Admin validation: PASSED\n")
} else {
sb.WriteString("Admin validation: FAILED\n")
}

for _, e := range result.Errors {
sb.WriteString(fmt.Sprintf("  ERROR: %s\n", e))
}

for _, w := range result.Warnings {
sb.WriteString(fmt.Sprintf("  WARNING: %s\n", w))
}

return sb.String()
}
