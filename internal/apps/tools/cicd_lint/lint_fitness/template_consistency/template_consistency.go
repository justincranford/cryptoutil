// Copyright (c) 2025 Justin Cranford

// Package template_consistency verifies that deployments/skeleton-template/secrets/
// uses hyphenated filenames (not underscores). This enforces the canonical secret
// naming convention so the template serves as a correct reference for new services.
// ENG-HANDBOOK.md Section 9.11.1 Fitness Sub-Linter Catalog.
package template_consistency

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
)

const (
	defaultTemplateSecretsPath = "deployments/skeleton-template/secrets"
	suffixSecret               = ".secret"
	suffixNever                = ".secret.never"
)

// Check runs the template-consistency check from the current working directory.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir checks that deployments/skeleton-template/secrets/ under rootDir
// only contains hyphenated filenames (no underscores in the base name).
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking skeleton-template secret naming consistency...")

	violations, err := FindViolationsInDir(rootDir)
	if err != nil {
		return fmt.Errorf("failed to check template consistency: %w", err)
	}

	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Fprintf(os.Stderr, "  underscore in secret filename: %s\n", v)
		}

		return fmt.Errorf("template-consistency: found %d secret file%s with underscore in name", len(violations), pluralS(len(violations)))
	}

	logger.Log("template-consistency: all skeleton-template secret names use hyphens")

	return nil
}

// FindViolationsInDir scans deployments/skeleton-template/secrets/ under rootDir
// and returns relative paths of secret files whose base name contains underscores.
func FindViolationsInDir(rootDir string) ([]string, error) {
	secretsDir := filepath.Join(rootDir, filepath.FromSlash(defaultTemplateSecretsPath))

	entries, err := os.ReadDir(secretsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("skeleton-template secrets directory not found: %s", secretsDir)
		}

		return nil, fmt.Errorf("failed to read skeleton-template secrets directory: %w", err)
	}

	var violations []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		if hasUnderscoreInBase(name) {
			violations = append(violations, name)
		}
	}

	sort.Strings(violations)

	return violations, nil
}

// hasUnderscoreInBase reports whether the base name of a secret file (i.e. the
// portion before the .secret or .secret.never suffix) contains an underscore.
func hasUnderscoreInBase(filename string) bool {
	var base string

	switch {
	case strings.HasSuffix(filename, suffixNever):
		base = strings.TrimSuffix(filename, suffixNever)
	case strings.HasSuffix(filename, suffixSecret):
		base = strings.TrimSuffix(filename, suffixSecret)
	default:
		return false
	}

	return strings.Contains(base, "_")
}

// pluralS returns an empty string for count==1, "s" otherwise.
func pluralS(count int) string {
	if count == 1 {
		return ""
	}

	return "s"
}
