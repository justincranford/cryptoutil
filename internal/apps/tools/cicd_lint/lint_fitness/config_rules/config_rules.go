// Copyright (c) 2025 Justin Cranford

// Package config_rules provides supplementary rule-based linters for config files.
// These complement the template_drift linters by enforcing cross-cutting structural
// rules: kebab-case key naming, header identity, instance minimality, and common
// config completeness.
// See ENG-HANDBOOK.md Section 9.11.1 Fitness Sub-Linter Catalog.
package config_rules

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// kebabCaseRegexp matches valid kebab-case YAML keys.
var kebabCaseRegexp = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

// requiredCommonKeys lists config keys that MUST appear in every common config overlay.
var requiredCommonKeys = []string{
	"bind-public-address",
	"tls-cert-file",
	"tls-key-file",
	"unseal-mode",
	"unseal-files",
	"browser-username-file",
	"browser-password-file",
	"service-username-file",
	"service-password-file",
	"allowed-ips",
	"allowed-cidrs",
	"csrf-token-single-use-token",
}

// allowedInstanceKeys lists the ONLY keys permitted in instance config overlays.
var allowedInstanceKeys = map[string]bool{
	"cors-origins":  true,
	"otlp-service":  true,
	"otlp-hostname": true,
	"database-url":  true,
}

// CheckKeyNaming validates all YAML keys in config files are kebab-case.
func CheckKeyNaming(logger *cryptoutilCmdCicdCommon.Logger) error {
	return checkKeyNamingInDir(logger, ".")
}

func checkKeyNamingInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking config YAML key naming...")

	var errs []string

	// Check deployment config overlays only.
	// Standalone configs (configs/PS-ID/PS-ID.yml) are excluded because they contain
	// domain-specific nested keys that legitimately use underscores (e.g., pki-ca's
	// ca.subject.common_name). The kebab-case rule applies to service framework keys.
	for _, ps := range cryptoutilRegistry.AllProductServices() {
		configDir := filepath.Join(rootDir, "deployments", ps.PSID, "config")

		files, err := filepath.Glob(filepath.Join(configDir, "*.yml"))
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: glob error: %s", ps.PSID, err))

			continue
		}

		for _, f := range files {
			if violations := checkFileKeyNaming(f); len(violations) > 0 {
				relPath, _ := filepath.Rel(rootDir, f)
				errs = append(errs, fmt.Sprintf("%s: %s", relPath, strings.Join(violations, "; ")))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("config-key-naming violations:\n%s", strings.Join(errs, "\n"))
	}

	logger.Log("config-key-naming: all config YAML keys are kebab-case")

	return nil
}

// checkFileKeyNaming parses a YAML file and returns non-kebab-case key violations.
func checkFileKeyNaming(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return []string{fmt.Sprintf("read error: %s", err)}
	}

	var node yaml.Node

	if err := yaml.Unmarshal(data, &node); err != nil {
		return []string{fmt.Sprintf("parse error: %s", err)}
	}

	return collectNonKebabKeys(&node, "")
}

// collectNonKebabKeys recursively finds non-kebab-case mapping keys.
func collectNonKebabKeys(node *yaml.Node, prefix string) []string {
	if node == nil {
		return nil
	}

	var violations []string

	switch node.Kind {
	case yaml.DocumentNode:
		for _, child := range node.Content {
			violations = append(violations, collectNonKebabKeys(child, prefix)...)
		}
	case yaml.MappingNode:
		for i := 0; i+1 < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valNode := node.Content[i+1]

			key := keyNode.Value
			fullKey := key

			if prefix != "" {
				fullKey = prefix + "." + key
			}

			if !kebabCaseRegexp.MatchString(key) {
				violations = append(violations, fmt.Sprintf("non-kebab-case key %q", fullKey))
			}

			violations = append(violations, collectNonKebabKeys(valNode, fullKey)...)
		}
	case yaml.SequenceNode:
		for _, child := range node.Content {
			violations = append(violations, collectNonKebabKeys(child, prefix)...)
		}
	case yaml.ScalarNode, yaml.AliasNode:
		// Leaf nodes — no keys to check.
	}

	return violations
}

// CheckHeaderIdentity validates that config file headers reference the correct PS-ID.
func CheckHeaderIdentity(logger *cryptoutilCmdCicdCommon.Logger) error {
	return checkHeaderIdentityInDir(logger, ".")
}

func checkHeaderIdentityInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking config header identity...")

	var errs []string

	// Check deployment config overlays.
	for _, ps := range cryptoutilRegistry.AllProductServices() {
		configDir := filepath.Join(rootDir, "deployments", ps.PSID, "config")

		files, err := filepath.Glob(filepath.Join(configDir, ps.PSID+"-app-*.yml"))
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: glob error: %s", ps.PSID, err))

			continue
		}

		for _, f := range files {
			if violation := checkFileHeader(f, ps.PSID); violation != "" {
				relPath, _ := filepath.Rel(rootDir, f)
				errs = append(errs, fmt.Sprintf("%s: %s", relPath, violation))
			}
		}
	}

	// Check standalone configs (framework + domain).
	for _, ps := range cryptoutilRegistry.AllProductServices() {
		for _, suffix := range []string{"-framework.yml", "-domain.yml"} {
			f := filepath.Join(rootDir, cryptoutilSharedMagic.CICDConfigsDir, ps.PSID, ps.PSID+suffix)

			if violation := checkFileHeader(f, ps.PSID); violation != "" {
				relPath, _ := filepath.Rel(rootDir, f)
				errs = append(errs, fmt.Sprintf("%s: %s", relPath, violation))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("config-header-identity violations:\n%s", strings.Join(errs, "\n"))
	}

	logger.Log("config-header-identity: all config headers reference correct PS-ID")

	return nil
}

// checkFileHeader reads the first two lines and verifies the expected PS-ID appears
// in at least one of them. Deployment configs have the PS-ID on line 1; standalone
// configs have descriptive names on line 1 and the PS-ID on line 2 (e.g.,
// "# Local development config for sm-kms.").
func checkFileHeader(path, expectedPSID string) string {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Sprintf("read error: %s", err)
	}

	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		return "file is empty"
	}

	firstLine := scanner.Text()
	if !strings.HasPrefix(firstLine, "#") {
		return fmt.Sprintf("first line is not a comment: %q", firstLine)
	}

	if strings.Contains(firstLine, expectedPSID) {
		return ""
	}

	// Check second line (standalone configs reference PS-ID here).
	if scanner.Scan() {
		secondLine := scanner.Text()
		if strings.Contains(secondLine, expectedPSID) {
			return ""
		}
	}

	return fmt.Sprintf("header does not reference PS-ID %q: %q", expectedPSID, firstLine)
}

// CheckInstanceMinimal validates instance config overlays only contain allowed keys.
func CheckInstanceMinimal(logger *cryptoutilCmdCicdCommon.Logger) error {
	return checkInstanceMinimalInDir(logger, ".")
}

func checkInstanceMinimalInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking instance config minimality...")

	var errs []string

	for _, ps := range cryptoutilRegistry.AllProductServices() {
		configDir := filepath.Join(rootDir, "deployments", ps.PSID, "config")

		// Instance configs match: {ps-id}-app-{sqlite|postgresql}-{N}.yml.
		patterns := []string{
			filepath.Join(configDir, ps.PSID+"-app-sqlite-*.yml"),
			filepath.Join(configDir, ps.PSID+"-app-postgresql-*.yml"),
		}

		for _, pattern := range patterns {
			files, err := filepath.Glob(pattern)
			if err != nil {
				errs = append(errs, fmt.Sprintf("%s: glob error: %s", ps.PSID, err))

				continue
			}

			for _, f := range files {
				if violations := checkInstanceKeys(f); len(violations) > 0 {
					relPath, _ := filepath.Rel(rootDir, f)
					errs = append(errs, fmt.Sprintf("%s: %s", relPath, strings.Join(violations, "; ")))
				}
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("config-instance-minimal violations:\n%s", strings.Join(errs, "\n"))
	}

	logger.Log("config-instance-minimal: all instance configs are minimal")

	return nil
}

// checkInstanceKeys parses a YAML file and returns keys not in the allowed set.
func checkInstanceKeys(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return []string{fmt.Sprintf("read error: %s", err)}
	}

	var config map[string]any

	if err := yaml.Unmarshal(data, &config); err != nil {
		return []string{fmt.Sprintf("parse error: %s", err)}
	}

	var violations []string

	for key := range config {
		if !allowedInstanceKeys[key] {
			violations = append(violations, fmt.Sprintf("unexpected key %q (only cors-origins, otlp-service, otlp-hostname, database-url allowed)", key))
		}
	}

	return violations
}

// CheckCommonComplete validates common config overlays contain all required keys.
func CheckCommonComplete(logger *cryptoutilCmdCicdCommon.Logger) error {
	return checkCommonCompleteInDir(logger, ".")
}

func checkCommonCompleteInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking common config completeness...")

	var errs []string

	for _, ps := range cryptoutilRegistry.AllProductServices() {
		f := filepath.Join(rootDir, "deployments", ps.PSID, "config", ps.PSID+"-app-framework-common.yml")

		if violations := checkCommonKeys(f); len(violations) > 0 {
			relPath, _ := filepath.Rel(rootDir, f)
			errs = append(errs, fmt.Sprintf("%s: %s", relPath, strings.Join(violations, "; ")))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("config-common-complete violations:\n%s", strings.Join(errs, "\n"))
	}

	logger.Log("config-common-complete: all common configs have required keys")

	return nil
}

// checkCommonKeys parses a YAML file and returns missing required keys.
func checkCommonKeys(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return []string{fmt.Sprintf("read error: %s", err)}
	}

	var config map[string]any

	if err := yaml.Unmarshal(data, &config); err != nil {
		return []string{fmt.Sprintf("parse error: %s", err)}
	}

	var violations []string

	for _, key := range requiredCommonKeys {
		if _, ok := config[key]; !ok {
			violations = append(violations, fmt.Sprintf("missing required key %q", key))
		}
	}

	return violations
}
