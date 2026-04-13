// Copyright (c) 2025 Justin Cranford

// Package compose_entrypoint_uniformity validates that every PS-ID Docker Compose app
// service uses the canonical shell-form command structure defined in ENG-HANDBOOK.md §12.3.5.
//
// Canonical command format (all 10 PS-IDs × 4 variants = 40 service definitions):
//
//	/bin/sh -c "exec /app/{PS-ID} server --config=/certs/tls-config.yml
//	  --config=/app/config/{PS-ID}-app-framework-common.yml
//	  --config=/app/config/{PS-ID}-app-framework-{variant}.yml
//	  --config=/app/config/{PS-ID}-app-domain-common.yml
//	  --config=/app/config/{PS-ID}-app-domain-{variant}.yml
//	  --config=/app/otel/otel.yml
//	  --bind-public-port=8080 -u {DATABASE_URL} $$SUITE_ARGS"
//
// Database URL by variant:
//   - sqlite-1:      sqlite://file::memory:?cache=shared
//   - sqlite-2:      sqlite://file::memory:?cache=shared
//   - postgresql-1:  file:///run/secrets/postgres-url.secret
//   - postgresql-2:  file:///run/secrets/postgres-url.secret
package compose_entrypoint_uniformity

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
)

// DSN constants for the two database backends, matching the canonical values in §12.3.5.
const (
	dsnSQLite   = "sqlite://file::memory:?cache=shared"
	dsnPostgres = "file:///run/secrets/postgres-url.secret"
)

// variantDSN maps each compose variant to its expected database URL.
var variantDSN = map[string]string{
	lintFitnessRegistry.ComposeVariantSQLite1:   dsnSQLite,
	lintFitnessRegistry.ComposeVariantSQLite2:   dsnSQLite,
	lintFitnessRegistry.ComposeVariantPostgres1: dsnPostgres,
	lintFitnessRegistry.ComposeVariantPostgres2: dsnPostgres,
}

// orderedVariants lists the 4 app variants in canonical order.
var orderedVariants = []string{
	lintFitnessRegistry.ComposeVariantSQLite1,
	lintFitnessRegistry.ComposeVariantSQLite2,
	lintFitnessRegistry.ComposeVariantPostgres1,
	lintFitnessRegistry.ComposeVariantPostgres2,
}

// composeFile represents the top-level structure of a compose.yml file.
type composeFile struct {
	Services map[string]composeService `yaml:"services"`
}

// composeService captures only the command field from a compose service definition.
// Command can be either a string (shell-form) or []string (exec-form).
type composeService struct {
	Command yamlStringOrSlice `yaml:"command"`
}

// yamlStringOrSlice handles YAML values that can be either a string or []string.
type yamlStringOrSlice struct {
	Value string
}

// UnmarshalYAML handles both string and []string YAML values.
func (y *yamlStringOrSlice) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		y.Value = value.Value

		return nil
	}

	if value.Kind == yaml.SequenceNode {
		var items []string
		if err := value.Decode(&items); err != nil {
			return fmt.Errorf("decode sequence: %w", err)
		}

		y.Value = strings.Join(items, " ")

		return nil
	}

	return fmt.Errorf("expected string or sequence, got %v", value.Kind)
}

// Check validates compose entrypoint uniformity from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates compose entrypoint uniformity under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	return checkInDir(logger, rootDir, os.ReadFile)
}

// checkInDir is the internal implementation that accepts a readFileFn for testing.
func checkInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string, readFileFn func(string) ([]byte, error)) error {
	logger.Log("Checking compose entrypoint uniformity (canonical command array §12.3.5)...")

	var violations []string

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		v := checkCompose(rootDir, ps.PSID, readFileFn)
		violations = append(violations, v...)
	}

	if len(violations) > 0 {
		return fmt.Errorf("compose entrypoint uniformity violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("compose-entrypoint-uniformity: all 40 app service command arrays match canonical pattern")

	return nil
}

// expectedCommand returns the canonical shell-form command string for a PS-ID and variant.
func expectedCommand(psID, variant string) string {
	return fmt.Sprintf(
		`/bin/sh -c "exec /app/%s server --config=/certs/tls-config.yml --config=/app/config/%s-app-framework-common.yml --config=/app/config/%s-app-framework-%s.yml --config=/app/config/%s-app-domain-common.yml --config=/app/config/%s-app-domain-%s.yml --config=/app/otel/otel.yml --bind-public-port=8080 -u %s $$SUITE_ARGS"`,
		psID, psID, psID, variant, psID, psID, variant, variantDSN[variant],
	)
}

// checkCompose validates all 4 app-service command strings in one PS-ID compose file.
func checkCompose(rootDir, psID string, readFileFn func(string) ([]byte, error)) []string {
	composePath := filepath.Join(rootDir, "deployments", psID, "compose.yml")

	data, err := readFileFn(composePath)
	if err != nil {
		return []string{fmt.Sprintf("%s: cannot read deployments/%s/compose.yml: %v", psID, psID, err)}
	}

	var cf composeFile
	if err := yaml.Unmarshal(data, &cf); err != nil {
		return []string{fmt.Sprintf("%s: cannot parse deployments/%s/compose.yml: %v", psID, psID, err)}
	}

	var violations []string

	for _, variant := range orderedVariants {
		svcName := lintFitnessRegistry.ComposeServiceName(psID, variant)

		svc, ok := cf.Services[svcName]
		if !ok {
			violations = append(violations, fmt.Sprintf(
				"%s/%s: service %q not found in deployments/%s/compose.yml",
				psID, variant, svcName, psID,
			))

			continue
		}

		want := expectedCommand(psID, variant)

		if svc.Command.Value != want {
			violations = append(violations, fmt.Sprintf(
				"%s/%s: service %q command mismatch\n  got:  %s\n  want: %s",
				psID, variant, svcName, svc.Command.Value, want,
			))
		}
	}

	return violations
}
