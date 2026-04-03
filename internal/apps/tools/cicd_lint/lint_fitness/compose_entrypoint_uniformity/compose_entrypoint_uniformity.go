// Copyright (c) 2025 Justin Cranford

// Package compose_entrypoint_uniformity validates that every PS-ID Docker Compose app
// service uses the canonical command array structure defined in ARCHITECTURE.md §12.3.5.
//
// Canonical command format (all 10 PS-IDs × 4 variants = 40 service definitions):
//
//	["server", "--bind-public-port=8080", "--config=/certs/tls-config.yml",
//	 "--config=/app/config/{PS-ID}-app-{variant}.yml",
//	 "--config=/app/config/{PS-ID}-app-common.yml",
//	 "--config=/app/otel/otel.yml",
//	 "-u", "{DATABASE_URL}"]
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
	dsnSQLite    = "sqlite://file::memory:?cache=shared"
	dsnPostgres  = "file:///run/secrets/postgres-url.secret"
	bindPublic   = "--bind-public-port=8080"
	tlsConfig    = "--config=/certs/tls-config.yml"
	otelConfig   = "--config=/app/otel/otel.yml"
	argSubcmd    = "server"
	argDSNFlag   = "-u"
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
type composeService struct {
	Command []string `yaml:"command"`
}

// Injectable function for testing the read-error path.
var readFileFn = os.ReadFile

// Check validates compose entrypoint uniformity from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates compose entrypoint uniformity under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking compose entrypoint uniformity (canonical command array §12.3.5)...")

	var violations []string

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		v := checkCompose(rootDir, ps.PSID)
		violations = append(violations, v...)
	}

	if len(violations) > 0 {
		return fmt.Errorf("compose entrypoint uniformity violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("compose-entrypoint-uniformity: all 40 app service command arrays match canonical pattern")

	return nil
}

// expectedCommand returns the canonical 8-element command array for a PS-ID and variant.
func expectedCommand(psID, variant string) []string {
	return []string{
		argSubcmd,
		bindPublic,
		tlsConfig,
		fmt.Sprintf("--config=/app/config/%s-app-%s.yml", psID, variant),
		fmt.Sprintf("--config=/app/config/%s-app-common.yml", psID),
		otelConfig,
		argDSNFlag,
		variantDSN[variant],
	}
}

// checkCompose validates all 4 app-service command arrays in one PS-ID compose file.
func checkCompose(rootDir, psID string) []string {
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

		if !commandsEqual(svc.Command, want) {
			violations = append(violations, fmt.Sprintf(
				"%s/%s: service %q command mismatch\n  got:  %v\n  want: %v",
				psID, variant, svcName, svc.Command, want,
			))
		}
	}

	return violations
}

// commandsEqual returns true if a and b are identical element-by-element.
func commandsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
