// Copyright (c) 2025 Justin Cranford

package hash

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"
	"os"
	"strings"
)

// PepperConfig defines version-specific pepper configuration.
// MANDATORY per OWASP Password Storage Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#peppering
//
// Pattern: PBKDF2(password||pepper, salt, iterations, keyLength)
// Storage: Docker/Kubernetes secrets (NEVER in DB/source code)
// Rotation: Requires version bump + re-hash all records (lazy migration).
type PepperConfig struct {
	// Version identifier (e.g., "1", "2", "3") matching PBKDF2Params.Version
	Version string

	// SecretPath is the Docker/K8s secret file path (e.g., "/run/secrets/hash_pepper_v3.secret")
	// Supports "file://" prefix for consistency with other config patterns
	SecretPath string
}

// LoadPepperFromSecret loads pepper from Docker/K8s secret file.
// Supports "file://path" or direct path patterns.
//
// Security Requirements:
//   - File MUST be readable (400 or 440 permissions recommended)
//   - File MUST be in /run/secrets/ (Docker) or /var/run/secrets/kubernetes.io/ (K8s)
//   - Content SHOULD be base64-encoded 32-byte random value (256 bits)
//
// Returns trimmed secret content (removes trailing newlines/whitespace).
func LoadPepperFromSecret(secretPath string) (string, error) {
	if secretPath == "" {
		return "", fmt.Errorf("secret path is empty")
	}

	// Strip "file://" prefix if present (consistency with other config patterns)
	actualPath := secretPath
	if strings.HasPrefix(secretPath, cryptoutilSharedMagic.FileURIScheme) {
		actualPath = strings.TrimPrefix(secretPath, cryptoutilSharedMagic.FileURIScheme)
	}

	// Read secret file
	content, err := os.ReadFile(actualPath)
	if err != nil {
		return "", fmt.Errorf("failed to read pepper secret from %q: %w", actualPath, err)
	}

	// Trim whitespace (Docker secrets often have trailing newlines)
	pepper := strings.TrimSpace(string(content))

	if pepper == "" {
		return "", fmt.Errorf("pepper secret file %q is empty", actualPath)
	}

	return pepper, nil
}

// ConfigurePeppers loads peppers from Docker/K8s secrets and updates parameter sets in registry.
//
// CRITICAL: This MUST be called during service initialization before any hash operations.
// Pepper configuration is MANDATORY per OWASP requirements.
//
// Example usage:
//
//	registry := GetGlobalRegistry()
//	peppers := []PepperConfig{
//	    {Version: "1", SecretPath: "/run/secrets/hash_pepper_v1.secret"},
//	    {Version: "2", SecretPath: "/run/secrets/hash_pepper_v2.secret"},
//	    {Version: "3", SecretPath: "file:///run/secrets/hash_pepper_v3.secret"},
//	}
//	if err := ConfigurePeppers(registry, peppers); err != nil {
//	    log.Fatalf("Failed to configure peppers: %v", err)
//	}
func ConfigurePeppers(registry *ParameterSetRegistry, peppers []PepperConfig) error {
	if registry == nil {
		return fmt.Errorf("registry is nil")
	}

	for _, pepperCfg := range peppers {
		if pepperCfg.Version == "" {
			return fmt.Errorf("pepper config has empty version")
		}

		if pepperCfg.SecretPath == "" {
			return fmt.Errorf("pepper config for version %q has empty secret path", pepperCfg.Version)
		}

		// Load pepper from secret file
		pepper, err := LoadPepperFromSecret(pepperCfg.SecretPath)
		if err != nil {
			return fmt.Errorf("failed to load pepper for version %q: %w", pepperCfg.Version, err)
		}

		// Update parameter set with pepper
		params, err := registry.GetParameterSet(pepperCfg.Version)
		if err != nil {
			return fmt.Errorf("failed to get parameter set for version %q: %w", pepperCfg.Version, err)
		}

		// CRITICAL: Set pepper in parameter set (OWASP requirement)
		params.Pepper = pepper
	}

	return nil
}
