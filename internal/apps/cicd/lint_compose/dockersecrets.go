// Copyright (c) 2025 Justin Cranford

package lint_compose

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

// inlineCredentialPatterns matches inline credential assignments in Docker Compose files.
// These should use Docker secrets instead.
var inlineCredentialPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)^\s*(POSTGRES_PASSWORD|POSTGRES_USER|POSTGRES_DB)\s*:\s*[^${}]+$`),
	regexp.MustCompile(`(?i)^\s*(API_KEY|API_SECRET|SECRET_KEY|PRIVATE_KEY)\s*:\s*[^${}]+$`),
	regexp.MustCompile(`(?i)^\s*(PASSWORD|TOKEN|PASSPHRASE|CREDENTIAL)\s*:\s*[^${}]+$`),
	regexp.MustCompile(`(?i)^\s*(DATABASE_URL|DB_URL|CONNECTION_STRING)\s*:\s*[^${}]+$`),
	regexp.MustCompile(`(?i)^\s*(JWT_SECRET|SESSION_SECRET|ENCRYPTION_KEY)\s*:\s*[^${}]+$`),
}

// validSecretPatterns matches valid patterns that use Docker secrets.
var validSecretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)_FILE\s*:\s*/run/secrets/`),
	regexp.MustCompile(`(?i)file:///run/secrets/`),
	regexp.MustCompile(`^\s*secrets:\s*$`),
	regexp.MustCompile(`^\s*-\s*[a-zA-Z0-9_.-]+\.secret\s*$`),
	regexp.MustCompile(`^\s*#`),
}

// SecretsViolation represents an inline credential violation.
type SecretsViolation struct {
	File    string
	Line    int
	Content string
	Reason  string
}

// LintDockerSecrets checks all Docker Compose files for inline credentials.
// Returns an error if any compose file has inline credentials instead of Docker secrets.
func LintDockerSecrets(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error {
	logger.Log("Running Docker Compose lint (docker secrets check)...")

	composeFiles := findComposeFiles(filesByExtension)
	if len(composeFiles) == 0 {
		logger.Log("No compose files found")

		return nil
	}

	logger.Log(fmt.Sprintf("Found %d compose files to check for secrets", len(composeFiles)))

	var violations []SecretsViolation

	for _, file := range composeFiles {
		fileViolations, err := checkComposeFileSecrets(file)
		if err != nil {
			logger.Log(fmt.Sprintf("Warning: failed to check %s: %v", file, err))

			continue
		}

		violations = append(violations, fileViolations...)
	}

	if len(violations) > 0 {
		printSecretsViolations(violations)

		return fmt.Errorf("lint-compose-secrets failed: %d inline credential violations found", len(violations))
	}

	logger.Log("lint-compose-secrets passed: no inline credential violations")

	return nil
}

// checkComposeFileSecrets checks a single compose file for inline credentials.
func checkComposeFileSecrets(filePath string) ([]SecretsViolation, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	defer func() { _ = file.Close() }()

	var violations []SecretsViolation

	scanner := bufio.NewScanner(file)
	lineNum := 0
	inEnvironmentSection := false

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Track if we're in an environment section.
		if strings.Contains(line, "environment:") {
			inEnvironmentSection = true

			continue
		}

		// Exit environment section when we hit a new key (not indented) or top-level key.
		trimmed := strings.TrimSpace(line)
		if inEnvironmentSection && len(trimmed) > 0 &&
			!strings.HasPrefix(trimmed, "-") &&
			!strings.HasPrefix(trimmed, "#") &&
			!strings.Contains(trimmed, ":") {
			inEnvironmentSection = false
		}

		// Also exit if we see another top-level key.
		if len(line) > 0 && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			inEnvironmentSection = false
		}

		// Skip if not in environment section.
		if !inEnvironmentSection {
			continue
		}

		// Skip valid patterns (comments, secrets references).
		isValid := false

		for _, pattern := range validSecretPatterns {
			if pattern.MatchString(line) {
				isValid = true

				break
			}
		}

		if isValid {
			continue
		}

		// Check for inline credential patterns.
		for _, pattern := range inlineCredentialPatterns {
			if pattern.MatchString(line) {
				violations = append(violations, SecretsViolation{
					File:    filePath,
					Line:    lineNum,
					Content: strings.TrimSpace(line),
					Reason:  "Inline credential found - MUST use Docker secrets pattern (see 03-06.security.instructions.md)",
				})

				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return violations, nil
}

// printSecretsViolations outputs all detected secrets violations.
func printSecretsViolations(violations []SecretsViolation) {
	fmt.Println()
	fmt.Println("SECURITY VIOLATIONS: Inline credentials found")
	fmt.Println(strings.Repeat("=", lineSeparatorLength))

	for _, v := range violations {
		fmt.Printf("\nFile: %s\n", v.File)
		fmt.Printf("Line: %d\n", v.Line)
		fmt.Printf("Content: %s\n", v.Content)
		fmt.Printf("Reason: %s\n", v.Reason)
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", lineSeparatorLength))
	fmt.Println("Fix: Use Docker secrets pattern:")
	fmt.Println("   secrets:")
	fmt.Println("     postgres_password:")
	fmt.Println("       file: ./secrets/postgres_password.secret")
	fmt.Println("   services:")
	fmt.Println("     myapp:")
	fmt.Println("       secrets: [postgres_password]")
	fmt.Println("       environment:")
	fmt.Println("         POSTGRES_PASSWORD_FILE: /run/secrets/postgres_password")
	fmt.Println()
}
