// Copyright (c) 2025 Justin Cranford

package template_drift

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// CheckDockerfile verifies all PS-ID Dockerfiles match the canonical template.
func CheckDockerfile(logger *cryptoutilCmdCicdCommon.Logger) error {
	return checkDockerfileInDir(logger, ".", instantiate)
}

func checkDockerfileInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string, instFn instantiateFn) error {
	logger.Log("Checking Dockerfile template drift...")

	var errs []string

	for _, ps := range cryptoutilRegistry.AllProductServices() {
		params := buildParams(ps.PSID)

		expected, err := instFn("Dockerfile.tmpl", params)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %s", ps.PSID, err))

			continue
		}

		actualPath := filepath.Join(rootDir, "deployments", ps.PSID, "Dockerfile")

		actual, err := os.ReadFile(actualPath)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %s", ps.PSID, err))

			continue
		}

		if diff := compareExact(expected, string(actual)); diff != "" {
			errs = append(errs, fmt.Sprintf("%s: Dockerfile drift:\n%s", ps.PSID, diff))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("template-dockerfile violations:\n%s", strings.Join(errs, "\n"))
	}

	logger.Log("template-dockerfile: all Dockerfiles match canonical template")

	return nil
}

// CheckCompose verifies all PS-ID compose.yml files match the canonical template.
// pki-ca uses superset comparison (allows domain-specific extra volume mounts).
func CheckCompose(logger *cryptoutilCmdCicdCommon.Logger) error {
	return checkComposeInDir(logger, ".", instantiate)
}

func checkComposeInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string, instFn instantiateFn) error {
	logger.Log("Checking compose.yml template drift...")

	var errs []string

	for _, ps := range cryptoutilRegistry.AllProductServices() {
		params := buildParams(ps.PSID)

		expected, err := instFn("compose.yml.tmpl", params)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %s", ps.PSID, err))

			continue
		}

		actualPath := filepath.Join(rootDir, "deployments", ps.PSID, "compose.yml")

		actual, err := os.ReadFile(actualPath)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %s", ps.PSID, err))

			continue
		}

		var diff string
		if ps.PSID == pkiCAPSID {
			diff = compareSupersetOrdered(normalizeCommentAlignment(expected), normalizeCommentAlignment(string(actual)))
		} else {
			diff = compareExact(normalizeCommentAlignment(expected), normalizeCommentAlignment(string(actual)))
		}

		if diff != "" {
			errs = append(errs, fmt.Sprintf("%s: compose.yml drift:\n%s", ps.PSID, diff))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("template-compose violations:\n%s", strings.Join(errs, "\n"))
	}

	logger.Log("template-compose: all compose.yml files match canonical template")

	return nil
}

// CheckConfigCommon verifies all PS-ID common config overlays match the canonical template.
// pki-ca uses prefix comparison (allows domain-specific CRL additions at end).
func CheckConfigCommon(logger *cryptoutilCmdCicdCommon.Logger) error {
	return checkConfigCommonInDir(logger, ".", instantiate)
}

func checkConfigCommonInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string, instFn instantiateFn) error {
	logger.Log("Checking config-common template drift...")

	var errs []string

	for _, ps := range cryptoutilRegistry.AllProductServices() {
		params := buildParams(ps.PSID)

		expected, err := instFn("config-common.yml.tmpl", params)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %s", ps.PSID, err))

			continue
		}

		actualPath := filepath.Join(rootDir, "deployments", ps.PSID, "config", ps.PSID+"-app-common.yml")

		actual, err := os.ReadFile(actualPath)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %s", ps.PSID, err))

			continue
		}

		var diff string
		if ps.PSID == pkiCAPSID {
			diff = comparePrefix(expected, string(actual))
		} else {
			diff = compareExact(expected, string(actual))
		}

		if diff != "" {
			errs = append(errs, fmt.Sprintf("%s: config-common drift:\n%s", ps.PSID, diff))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("template-config-common violations:\n%s", strings.Join(errs, "\n"))
	}

	logger.Log("template-config-common: all common configs match canonical template")

	return nil
}

// CheckConfigSQLite verifies all PS-ID SQLite config overlays match the canonical template.
func CheckConfigSQLite(logger *cryptoutilCmdCicdCommon.Logger) error {
	return checkConfigSQLiteInDir(logger, ".", instantiate)
}

func checkConfigSQLiteInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string, instFn instantiateFn) error {
	logger.Log("Checking config-sqlite template drift...")

	var errs []string

	for _, ps := range cryptoutilRegistry.AllProductServices() {
		basePort := cryptoutilRegistry.PublicPort(ps.PSID)
		instances := []struct {
			num    int
			port   int
			suffix string
		}{
			{num: 1, port: basePort + cryptoutilRegistry.ComposeVariantOffsetSQLite1, suffix: cryptoutilRegistry.DeploymentConfigSuffixSQLite1},
			{num: 2, port: basePort + cryptoutilRegistry.ComposeVariantOffsetSQLite2, suffix: cryptoutilRegistry.DeploymentConfigSuffixSQLite2},
		}

		for _, inst := range instances {
			params := buildInstanceParams(ps.PSID, inst.num, inst.port)

			expected, err := instFn("config-sqlite.yml.tmpl", params)
			if err != nil {
				errs = append(errs, fmt.Sprintf("%s instance %d: %s", ps.PSID, inst.num, err))

				continue
			}

			actualPath := filepath.Join(rootDir, "deployments", ps.PSID, "config", ps.PSID+inst.suffix)

			actual, err := os.ReadFile(actualPath)
			if err != nil {
				errs = append(errs, fmt.Sprintf("%s instance %d: %s", ps.PSID, inst.num, err))

				continue
			}

			if diff := compareExact(expected, string(actual)); diff != "" {
				errs = append(errs, fmt.Sprintf("%s instance %d: config-sqlite drift:\n%s", ps.PSID, inst.num, diff))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("template-config-sqlite violations:\n%s", strings.Join(errs, "\n"))
	}

	logger.Log("template-config-sqlite: all SQLite config overlays match canonical template")

	return nil
}

// CheckConfigPostgreSQL verifies all PS-ID PostgreSQL config overlays match the canonical template.
func CheckConfigPostgreSQL(logger *cryptoutilCmdCicdCommon.Logger) error {
	return checkConfigPostgreSQLInDir(logger, ".", instantiate)
}

func checkConfigPostgreSQLInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string, instFn instantiateFn) error {
	logger.Log("Checking config-postgresql template drift...")

	var errs []string

	for _, ps := range cryptoutilRegistry.AllProductServices() {
		basePort := cryptoutilRegistry.PublicPort(ps.PSID)
		instances := []struct {
			num    int
			port   int
			suffix string
		}{
			{num: 1, port: basePort + cryptoutilRegistry.ComposeVariantOffsetPostgres1, suffix: cryptoutilRegistry.DeploymentConfigSuffixPostgresql1},
			{num: 2, port: basePort + cryptoutilRegistry.ComposeVariantOffsetPostgres2, suffix: cryptoutilRegistry.DeploymentConfigSuffixPostgresql2},
		}

		for _, inst := range instances {
			params := buildInstanceParams(ps.PSID, inst.num, inst.port)

			expected, err := instFn("config-postgresql.yml.tmpl", params)
			if err != nil {
				errs = append(errs, fmt.Sprintf("%s instance %d: %s", ps.PSID, inst.num, err))

				continue
			}

			actualPath := filepath.Join(rootDir, "deployments", ps.PSID, "config", ps.PSID+inst.suffix)

			actual, err := os.ReadFile(actualPath)
			if err != nil {
				errs = append(errs, fmt.Sprintf("%s instance %d: %s", ps.PSID, inst.num, err))

				continue
			}

			if diff := compareExact(expected, string(actual)); diff != "" {
				errs = append(errs, fmt.Sprintf("%s instance %d: config-postgresql drift:\n%s", ps.PSID, inst.num, diff))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("template-config-postgresql violations:\n%s", strings.Join(errs, "\n"))
	}

	logger.Log("template-config-postgresql: all PostgreSQL config overlays match canonical template")

	return nil
}

// CheckStandaloneConfig verifies all PS-ID standalone configs start with the canonical prefix.
// Domain-specific additions after the prefix are allowed.
func CheckStandaloneConfig(logger *cryptoutilCmdCicdCommon.Logger) error {
	return checkStandaloneConfigInDir(logger, ".", instantiate)
}

func checkStandaloneConfigInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string, instFn instantiateFn) error {
	logger.Log("Checking standalone config template drift...")

	var errs []string

	for _, ps := range cryptoutilRegistry.AllProductServices() {
		params := buildParams(ps.PSID)

		expected, err := instFn("standalone-config.yml.tmpl", params)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %s", ps.PSID, err))

			continue
		}

		actualPath := filepath.Join(rootDir, cryptoutilSharedMagic.CICDConfigsDir, ps.PSID, ps.PSID+"-framework.yml")

		actual, err := os.ReadFile(actualPath)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %s", ps.PSID, err))

			continue
		}

		if diff := comparePrefix(expected, string(actual)); diff != "" {
			errs = append(errs, fmt.Sprintf("%s: standalone config drift:\n%s", ps.PSID, diff))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("template-standalone-config violations:\n%s", strings.Join(errs, "\n"))
	}

	logger.Log("template-standalone-config: all standalone configs match canonical prefix")

	return nil
}
