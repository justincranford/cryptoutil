// Copyright (c) 2025 Justin Cranford
//
//

package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// highEntropySecretPatterns are substrings in secret file names that indicate
// high-entropy credential content (not identifiers like usernames or database names).
var highEntropySecretPatterns = []string{ //nolint:gochecknoglobals // package-level config
	"password", "passwd", "pepper", "private_key", "private-key",
	"api_key", "api-key", "secret_key", "secret-key",
	"unseal", "hash_pepper", "hash-pepper",
}

// ValidateSecretsCommand validates Docker secrets mounted at /run/secrets.
// It checks:
//   - All files have the .secret suffix
//   - High-entropy secrets meet minimum length (>=43 chars)
//   - All files are in /run/secrets/ (enforced by Docker, verified by path)
func ValidateSecretsCommand(args []string, stdout, stderr io.Writer) int {
	if IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stdout, "Usage: <service> validate-secrets\n\nValidates Docker secrets mounted at /run/secrets/.")

		return 0
	}

	secretsDir := cryptoutilSharedMagic.DockerSecretsDir

	entries, err := os.ReadDir(secretsDir)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "validate-secrets: cannot read %s: %v\n", secretsDir, err)

		return 1
	}

	var errors []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Verify .secret suffix naming convention.
		if !strings.HasSuffix(name, cryptoutilSharedMagic.CICDTemplateSecretFileSuffix) {
			errors = append(errors, fmt.Sprintf("secret file '%s' does not have required '.secret' suffix", name))

			continue
		}

		// Only check content length for high-entropy secret files.
		if !isHighEntropySecret(name) {
			continue
		}

		filePath := filepath.Join(secretsDir, name)

		content, readErr := os.ReadFile(filePath)
		if readErr != nil {
			errors = append(errors, fmt.Sprintf("cannot read secret file '%s': %v", name, readErr))

			continue
		}

		trimmed := strings.TrimSpace(string(content))

		if len(trimmed) == 0 {
			errors = append(errors, fmt.Sprintf("secret file '%s' is empty", name))

			continue
		}

		if len(trimmed) < cryptoutilSharedMagic.DockerSecretMinLength {
			errors = append(errors,
				fmt.Sprintf("secret file '%s' has %d chars (minimum: %d)",
					name, len(trimmed), cryptoutilSharedMagic.DockerSecretMinLength))
		}
	}

	if len(errors) > 0 {
		for _, e := range errors {
			_, _ = fmt.Fprintf(stderr, "validate-secrets: %s\n", e)
		}

		return 1
	}

	_, _ = fmt.Fprintln(stdout, "All secrets validated successfully")

	return 0
}

// isHighEntropySecret returns true if the secret file name indicates high-entropy credential content.
func isHighEntropySecret(name string) bool {
	lower := strings.ToLower(name)

	for _, pattern := range highEntropySecretPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	return false
}
