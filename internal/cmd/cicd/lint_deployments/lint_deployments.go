package lint_deployments

import (
"fmt"
"os"
"path/filepath"
"sort"
"strings"
)

// DeploymentStructure defines expected directory structure for each deployment type.
type DeploymentStructure struct {
Name              string
RequiredDirs      []string
RequiredFiles     []string
OptionalFiles     []string
RequiredSecrets   []string
AllowedExtensions []string
}

// GetExpectedStructures returns validation rules for different deployment types.
// See: docs/ARCHITECTURE-TODO.md for architectural documentation (pending Section 12.4).
func GetExpectedStructures() map[string]DeploymentStructure {
return map[string]DeploymentStructure{
"PRODUCT-SERVICE": {
Name:          "PRODUCT-SERVICE deployment (e.g., jose-ja, cipher-im)",
RequiredDirs:  []string{"secrets", "config"},
RequiredFiles: []string{"compose.yml", "Dockerfile"},
OptionalFiles: []string{"compose.demo.yml", "otel-collector-config.yaml", "README.md"},
RequiredSecrets: []string{
"unseal_1of5.secret", "unseal_2of5.secret", "unseal_3of5.secret",
"unseal_4of5.secret", "unseal_5of5.secret", "hash_pepper_v3.secret",
"postgres_url.secret", "postgres_username.secret",
"postgres_password.secret", "postgres_database.secret",
},
AllowedExtensions: []string{".yml", ".yaml", ".secret", ".md"},
},
"template": {
Name:          "Template deployment (deployments/template/)",
RequiredDirs:  []string{"secrets"},
RequiredFiles: []string{"compose.yml"},
RequiredSecrets: []string{
"unseal_1of5.secret", "unseal_2of5.secret", "unseal_3of5.secret",
"unseal_4of5.secret", "unseal_5of5.secret", "hash_pepper_v3.secret",
"postgres_url.secret", "postgres_username.secret",
"postgres_password.secret", "postgres_database.secret",
},
AllowedExtensions: []string{".yml", ".yaml", ".secret", ".md"},
},
"infrastructure": {
Name:              "Infrastructure deployment (postgres, citus, telemetry)",
RequiredDirs:      []string{},
RequiredFiles:     []string{"compose.yml"},
OptionalFiles:     []string{"init-db.sql", "init-citus.sql", "README.md"},
RequiredSecrets:   []string{}, // Infrastructure secrets are optional
AllowedExtensions: []string{".yml", ".yaml", ".sql", ".md"},
},
}
}

// ValidationResult holds validation outcome for a directory.
type ValidationResult struct {
Path           string
Type           string
Valid          bool
MissingDirs    []string
MissingFiles   []string
MissingSecrets []string
Errors         []string
Warnings       []string
}

// ValidateDeploymentStructure validates a deployment directory against expected structure.
func ValidateDeploymentStructure(basePath string, deploymentName string, structType string) (*ValidationResult, error) {
structures := GetExpectedStructures()
expected, ok := structures[structType]
if !ok {
return nil, fmt.Errorf("unknown structure type: %s", structType)
}

result := &ValidationResult{
Path:  basePath,
Type:  structType,
Valid: true,
}

// Check required directories
for _, dir := range expected.RequiredDirs {
dirPath := filepath.Join(basePath, dir)
if _, err := os.Stat(dirPath); os.IsNotExist(err) {
result.MissingDirs = append(result.MissingDirs, dir)
result.Valid = false
}
}

// Check required files
for _, file := range expected.RequiredFiles {
filePath := filepath.Join(basePath, file)
if _, err := os.Stat(filePath); os.IsNotExist(err) {
result.MissingFiles = append(result.MissingFiles, file)
result.Valid = false
}
}

// Check required secrets
if len(expected.RequiredSecrets) > 0 {
secretsPath := filepath.Join(basePath, "secrets")
if _, err := os.Stat(secretsPath); err == nil {
for _, secret := range expected.RequiredSecrets {
secretPath := filepath.Join(secretsPath, secret)
if _, err := os.Stat(secretPath); os.IsNotExist(err) {
result.MissingSecrets = append(result.MissingSecrets, secret)
result.Valid = false
}
}
}
}

return result, nil
}

// ValidateAllDeployments validates all deployments in the given root directory.
func ValidateAllDeployments(deploymentsRoot string) ([]ValidationResult, error) {
var results []ValidationResult

// Service deployments (PRODUCT-SERVICE pattern)
serviceNames := []string{
"jose-ja", "cipher-im", "pki-ca", "sm-kms",
"identity-authz", "identity-idp", "identity-rp", "identity-rs", "identity-spa",
}

for _, svc := range serviceNames {
svcPath := filepath.Join(deploymentsRoot, svc)
if _, err := os.Stat(svcPath); err == nil {
result, err := ValidateDeploymentStructure(svcPath, svc, "PRODUCT-SERVICE")
if err != nil {
return nil, fmt.Errorf("failed to validate %s: %w", svc, err)
}
results = append(results, *result)
}
}

// Template deployment
templatePath := filepath.Join(deploymentsRoot, "template")
if _, err := os.Stat(templatePath); err == nil {
result, err := ValidateDeploymentStructure(templatePath, "template", "template")
if err != nil {
return nil, fmt.Errorf("failed to validate template: %w", err)
}
results = append(results, *result)
}

// Infrastructure deployments
infraNames := []string{"postgres", "citus", "telemetry", "compose"}
for _, infra := range infraNames {
infraPath := filepath.Join(deploymentsRoot, infra)
if _, err := os.Stat(infraPath); err == nil {
result, err := ValidateDeploymentStructure(infraPath, infra, "infrastructure")
if err != nil {
return nil, fmt.Errorf("failed to validate %s: %w", infra, err)
}
results = append(results, *result)
}
}

return results, nil
}

// FormatResults formats validation results for human-readable output.
func FormatResults(results []ValidationResult) string {
var sb strings.Builder

validCount := 0
for _, r := range results {
if r.Valid {
validCount++
}
}

sb.WriteString(fmt.Sprintf("Validated %d deployments: %d valid, %d with issues\n\n", len(results), validCount, len(results)-validCount))

// Sort results: invalid first, then by path
sorted := make([]ValidationResult, len(results))
copy(sorted, results)
sort.Slice(sorted, func(i, j int) bool {
if sorted[i].Valid != sorted[j].Valid {
return !sorted[i].Valid // Invalid first
}
return sorted[i].Path < sorted[j].Path
})

for _, r := range sorted {
status := "✅ VALID"
if !r.Valid {
status = "❌ INVALID"
}

sb.WriteString(fmt.Sprintf("%s %s (%s)\n", status, filepath.Base(r.Path), r.Type))

if len(r.MissingDirs) > 0 {
sb.WriteString(fmt.Sprintf("  Missing directories: %s\n", strings.Join(r.MissingDirs, ", ")))
}
if len(r.MissingFiles) > 0 {
sb.WriteString(fmt.Sprintf("  Missing files: %s\n", strings.Join(r.MissingFiles, ", ")))
}
if len(r.MissingSecrets) > 0 {
sb.WriteString(fmt.Sprintf("  Missing secrets: %s\n", strings.Join(r.MissingSecrets, ", ")))
}
if len(r.Errors) > 0 {
for _, err := range r.Errors {
sb.WriteString(fmt.Sprintf("  ERROR: %s\n", err))
}
}
if len(r.Warnings) > 0 {
for _, warn := range r.Warnings {
sb.WriteString(fmt.Sprintf("  WARN: %s\n", warn))
}
}
}

return sb.String()
}
