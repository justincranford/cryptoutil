package lint_deployments

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Validator name constants for identification in aggregated results.
const (
	validatorNameNaming          = "naming"
	validatorNameKebabCase       = "kebab-case"
	validatorNameSchema          = "schema"
	validatorNameTemplatePattern = "template-pattern"
	validatorNamePorts           = "ports"
	validatorNameTelemetry       = "telemetry"
	validatorNameAdmin           = "admin"
	validatorNameSecrets         = "secrets"
)

// ValidatorResult holds the outcome of a single validator run.
type ValidatorResult struct {
	Name     string
	Target   string
	Passed   bool
	Output   string
	Duration time.Duration
}

// AllValidationResult aggregates results from all validators.
type AllValidationResult struct {
	Results       []ValidatorResult
	TotalDuration time.Duration
}

// AllPassed returns true if every validator result passed.
func (r *AllValidationResult) AllPassed() bool {
	for i := range r.Results {
		if !r.Results[i].Passed {
			return false
		}
	}

	return true
}

// addResult appends a validator result to the aggregated results.
func (r *AllValidationResult) addResult(name, target string, passed bool, output string, dur time.Duration) {
	r.Results = append(r.Results, ValidatorResult{
		Name:     name,
		Target:   target,
		Passed:   passed,
		Output:   output,
		Duration: dur,
	})
}

// deploymentEntry holds metadata for a discovered deployment directory.
type deploymentEntry struct {
	path  string
	name  string
	level string
}

// ValidateAll runs all 8 validators sequentially with aggregated error reporting.
// It continues on failure, collecting ALL results before returning.
// See: Decision 11:E (sequential + aggregated error reporting).
func ValidateAll(deploymentsDir, configsDir string) *AllValidationResult {
	start := time.Now().UTC()
	result := &AllValidationResult{}

	// 1. Naming validation on deployments/ and configs/.
	runNamingValidation(deploymentsDir, configsDir, result)

	// 2. KebabCase validation on configs/.
	runKebabCaseValidation(configsDir, result)

	// 3. Schema validation on each config YAML file.
	runSchemaValidation(configsDir, result)

	// 4. Template pattern validation on deployments/template/.
	runTemplatePatternValidation(deploymentsDir, result)

	// 5. Discover deployment directories for per-deployment validators.
	deployments := discoverDeploymentDirs(deploymentsDir)

	// 6. Ports validation on each non-infrastructure deployment.
	runPortsValidation(deployments, result)

	// 7. Telemetry validation on configs/.
	runTelemetryValidation(configsDir, result)

	// 8. Admin validation on each deployment.
	runAdminValidation(deployments, result)

	// 9. Secrets validation on each deployment.
	runSecretsValidation(deployments, result)

	result.TotalDuration = time.Since(start)

	return result
}

// runNamingValidation runs ValidateNaming on both deployments and configs directories.
func runNamingValidation(deploymentsDir, configsDir string, result *AllValidationResult) {
	for _, dir := range []string{deploymentsDir, configsDir} {
		start := time.Now().UTC()
		nr, _ := ValidateNaming(dir)
		dur := time.Since(start)

		result.addResult(validatorNameNaming, dir, nr.Valid, FormatNamingValidationResult(nr), dur)
	}
}

// runKebabCaseValidation runs ValidateKebabCase on the configs directory.
func runKebabCaseValidation(configsDir string, result *AllValidationResult) {
	start := time.Now().UTC()
	kr, _ := ValidateKebabCase(configsDir)
	dur := time.Since(start)

	result.addResult(validatorNameKebabCase, configsDir, kr.Valid, FormatKebabCaseValidationResult(kr), dur)
}

// isServiceTemplateConfig returns true if the file matches the service template
// config naming pattern (config-*.yml). Only these files use the flat kebab-case
// format validated by ValidateSchema. Other configs (e.g., ca-server.yml,
// identity profiles, policies) use nested YAML with domain-specific schemas.
func isServiceTemplateConfig(path string) bool {
	base := filepath.Base(path)

	return strings.HasPrefix(base, "config-") && (strings.HasSuffix(base, ".yml") || strings.HasSuffix(base, ".yaml"))
}

// runSchemaValidation discovers config YAML files and runs ValidateSchema on each.
// Only validates service template config files (config-*.yml) which use the flat
// kebab-case format. Other configs have domain-specific schemas.
func runSchemaValidation(configsDir string, result *AllValidationResult) {
	configFiles := discoverConfigFiles(configsDir)

	for _, cf := range configFiles {
		if !isServiceTemplateConfig(cf) {
			continue
		}

		start := time.Now().UTC()
		sr, _ := ValidateSchema(cf)
		dur := time.Since(start)

		result.addResult(validatorNameSchema, cf, sr.Valid, FormatSchemaValidationResult(sr), dur)
	}
}

// runTemplatePatternValidation runs ValidateTemplatePattern on deployments/template/.
func runTemplatePatternValidation(deploymentsDir string, result *AllValidationResult) {
	templatePath := filepath.Join(deploymentsDir, "template")

	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return
	}

	start := time.Now().UTC()
	tr, _ := ValidateTemplatePattern(templatePath)
	dur := time.Since(start)

	result.addResult(validatorNameTemplatePattern, templatePath, tr.Valid, FormatTemplatePatternResult(tr), dur)
}

// runPortsValidation runs ValidatePorts on each non-infrastructure deployment.
func runPortsValidation(deployments []deploymentEntry, result *AllValidationResult) {
	for _, d := range deployments {
		if d.level == DeploymentTypeInfrastructure || d.level == DeploymentTypeTemplate {
			continue
		}

		start := time.Now().UTC()
		pr, _ := ValidatePorts(d.path, d.name, d.level)
		dur := time.Since(start)

		result.addResult(validatorNamePorts, d.path, pr.Valid, FormatPortValidationResult(pr), dur)
	}
}

// runTelemetryValidation runs ValidateTelemetry on the configs directory.
func runTelemetryValidation(configsDir string, result *AllValidationResult) {
	start := time.Now().UTC()
	tr, _ := ValidateTelemetry(configsDir)
	dur := time.Since(start)

	result.addResult(validatorNameTelemetry, configsDir, tr.Valid, FormatTelemetryValidationResult(tr), dur)
}

// runAdminValidation runs ValidateAdmin on each deployment directory.
func runAdminValidation(deployments []deploymentEntry, result *AllValidationResult) {
	for _, d := range deployments {
		start := time.Now().UTC()
		ar, _ := ValidateAdmin(d.path)
		dur := time.Since(start)

		result.addResult(validatorNameAdmin, d.path, ar.Valid, FormatAdminValidationResult(ar), dur)
	}
}

// runSecretsValidation runs ValidateSecrets on each deployment directory.
// runSecretsValidation runs ValidateSecrets on each non-infrastructure deployment.
// Infrastructure deployments (compose, shared-telemetry) are skipped because they
// intentionally use inline credentials for local dev services (e.g., Grafana).
func runSecretsValidation(deployments []deploymentEntry, result *AllValidationResult) {
	for _, d := range deployments {
		if d.level == DeploymentTypeInfrastructure {
			continue
		}

		start := time.Now().UTC()
		sr, _ := ValidateSecrets(d.path)
		dur := time.Since(start)

		result.addResult(validatorNameSecrets, d.path, sr.Valid, FormatSecretValidationResult(sr), dur)
	}
}

// discoverDeploymentDirs finds and classifies all deployment subdirectories.
func discoverDeploymentDirs(deploymentsDir string) []deploymentEntry {
	var deployments []deploymentEntry

	entries, err := os.ReadDir(deploymentsDir)
	if err != nil {
		return deployments
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		level := classifyDeployment(name)
		deployments = append(deployments, deploymentEntry{
			path:  filepath.Join(deploymentsDir, name),
			name:  name,
			level: level,
		})
	}

	return deployments
}

// classifyDeployment determines the deployment type from the directory name.
func classifyDeployment(name string) string {
	serviceNames := map[string]bool{
		"jose-ja": true, "sm-im": true, "pki-ca": true, "sm-kms": true,
		"identity-authz": true, "identity-idp": true, "identity-rp": true,
		"identity-rs": true, "identity-spa": true,
	}

	productNames := map[string]bool{
		"identity": true, "sm": true, "pki": true, "jose": true,
	}

	switch {
	case serviceNames[name]:
		return DeploymentTypeProductService
	case productNames[name]:
		return DeploymentTypeProduct
	case name == "cryptoutil-suite":
		return DeploymentTypeSuite
	case name == "template":
		return DeploymentTypeTemplate
	default:
		return DeploymentTypeInfrastructure
	}
}

// discoverConfigFiles walks the configs directory and returns all YAML file paths.
func discoverConfigFiles(configsDir string) []string {
	var files []string

	_ = filepath.WalkDir(configsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			return nil
		}

		if isYAMLFile(path) {
			files = append(files, path)
		}

		return nil
	})

	return files
}

// FormatAllValidationResult produces a human-readable summary of all validator results.
func FormatAllValidationResult(result *AllValidationResult) string {
	var sb strings.Builder

	sb.WriteString("=== Validate All: Aggregated Results ===\n\n")

	passedCount := 0
	failedCount := 0

	for i := range result.Results {
		vr := &result.Results[i]
		status := statusPass

		if !vr.Passed {
			status = statusFail
			failedCount++
		} else {
			passedCount++
		}

		sb.WriteString(fmt.Sprintf("[%s] %s (%s) [%s]\n", status, vr.Name, vr.Target, vr.Duration.Round(time.Millisecond)))
	}

	sb.WriteString("\n--- Summary ---\n")
	sb.WriteString(fmt.Sprintf("Total:    %d validators\n", len(result.Results)))
	sb.WriteString(fmt.Sprintf("Passed:   %d\n", passedCount))
	sb.WriteString(fmt.Sprintf("Failed:   %d\n", failedCount))
	sb.WriteString(fmt.Sprintf("Duration: %s\n", result.TotalDuration.Round(time.Millisecond)))

	if result.AllPassed() {
		sb.WriteString("\nResult: ALL VALIDATORS PASSED\n")
	} else {
		sb.WriteString("\nResult: VALIDATION FAILED\n")
		sb.WriteString("\nFailed validators:\n")

		for i := range result.Results {
			vr := &result.Results[i]
			if !vr.Passed {
				sb.WriteString(fmt.Sprintf("  - %s (%s)\n", vr.Name, vr.Target))
			}
		}
	}

	return sb.String()
}
