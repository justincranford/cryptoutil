// Copyright (c) 2025 Justin Cranford

// Package compose_tier_override_integrity validates tier-level compose override invariants:
// 1. Forbidden builder service definitions are not reintroduced at PRODUCT/SUITE levels.
// 2. PRODUCT/SUITE compose files override all required postgres secrets.
// 3. PRODUCT/SUITE postgres-url.secret content is synchronized with username/password/database secrets.
package compose_tier_override_integrity

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

type composeFile struct {
	Services map[string]any `yaml:"services"`
	Secrets  map[string]any `yaml:"secrets"`
}

var requiredPostgresSecretNames = []string{
	"postgres-url.secret",
	"postgres-username.secret",
	"postgres-password.secret",
	"postgres-database.secret",
}

// Check validates from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates tier-level compose override and postgres secret invariants.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking compose tier override integrity...")

	var violations []string

	tierRules := []struct {
		tierID            string
		forbiddenServices []string
	}{
		{tierID: cryptoutilSharedMagic.DefaultOTLPServiceDefault, forbiddenServices: []string{cryptoutilSharedMagic.DockerJobBuilderCryptoutil}},
		{tierID: "sm", forbiddenServices: []string{"builder-sm-kms", "builder-sm-im"}},
	}

	for _, rule := range tierRules {
		composePath := filepath.Join(rootDir, "deployments", rule.tierID, "compose.yml")

		cf, err := readComposeFile(composePath)
		if err != nil {
			violations = append(violations, fmt.Sprintf("%s: %v", rule.tierID, err))

			continue
		}

		for _, forbidden := range rule.forbiddenServices {
			if _, exists := cf.Services[forbidden]; exists {
				violations = append(violations, fmt.Sprintf(
					"%s: deployments/%s/compose.yml MUST NOT define service %q",
					rule.tierID, rule.tierID, forbidden,
				))
			}
		}

		for _, secretName := range requiredPostgresSecretNames {
			if _, exists := cf.Secrets[secretName]; !exists {
				violations = append(violations, fmt.Sprintf(
					"%s: deployments/%s/compose.yml missing required postgres secret override %q",
					rule.tierID, rule.tierID, secretName,
				))
			}
		}

		syncErr := validatePostgresURLSync(filepath.Join(rootDir, "deployments", rule.tierID, "secrets"))
		if syncErr != nil {
			violations = append(violations, fmt.Sprintf("%s: %v", rule.tierID, syncErr))
		}
	}

	if len(violations) > 0 {
		return fmt.Errorf("compose tier override integrity violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("compose-tier-override-integrity: tier-level compose overrides are valid")

	return nil
}

func readComposeFile(composePath string) (*composeFile, error) {
	data, err := os.ReadFile(composePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", composePath, err)
	}

	var cf composeFile
	if err := yaml.Unmarshal(data, &cf); err != nil {
		return nil, fmt.Errorf("cannot parse %s: %w", composePath, err)
	}

	if cf.Services == nil {
		cf.Services = make(map[string]any)
	}

	if cf.Secrets == nil {
		cf.Secrets = make(map[string]any)
	}

	return &cf, nil
}

func validatePostgresURLSync(secretsDir string) error {
	username, err := readSecretValue(filepath.Join(secretsDir, "postgres-username.secret"))
	if err != nil {
		return fmt.Errorf("cannot read postgres username secret: %w", err)
	}

	password, err := readSecretValue(filepath.Join(secretsDir, "postgres-password.secret"))
	if err != nil {
		return fmt.Errorf("cannot read postgres password secret: %w", err)
	}

	database, err := readSecretValue(filepath.Join(secretsDir, "postgres-database.secret"))
	if err != nil {
		return fmt.Errorf("cannot read postgres database secret: %w", err)
	}

	postgresURL, err := readSecretValue(filepath.Join(secretsDir, "postgres-url.secret"))
	if err != nil {
		return fmt.Errorf("cannot read postgres URL secret: %w", err)
	}

	parsedURL, err := url.Parse(postgresURL)
	if err != nil {
		return fmt.Errorf("postgres-url.secret is not a valid URL: %w", err)
	}

	if parsedURL.Scheme != cryptoutilSharedMagic.DockerServicePostgres && parsedURL.Scheme != "postgresql" {
		return fmt.Errorf("postgres-url.secret must use postgres/postgresql scheme, got %q", parsedURL.Scheme)
	}

	if parsedURL.User == nil {
		return fmt.Errorf("postgres-url.secret is missing credentials")
	}

	urlUser := parsedURL.User.Username()

	urlPass, hasPass := parsedURL.User.Password()
	if !hasPass {
		return fmt.Errorf("postgres-url.secret is missing password")
	}

	if urlUser != username {
		return fmt.Errorf("postgres-url.secret username %q does not match postgres-username.secret %q", urlUser, username)
	}

	if urlPass != password {
		return fmt.Errorf("postgres-url.secret password does not match postgres-password.secret")
	}

	pathDB := strings.TrimPrefix(parsedURL.Path, "/")
	if pathDB != database {
		return fmt.Errorf("postgres-url.secret database %q does not match postgres-database.secret %q", pathDB, database)
	}

	return nil
}

func readSecretValue(secretPath string) (string, error) {
	data, err := os.ReadFile(secretPath)
	if err != nil {
		return "", fmt.Errorf("read secret %s: %w", secretPath, err)
	}

	return strings.TrimSpace(string(data)), nil
}
