// Copyright (c) 2025 Justin Cranford

// Package secret_content validates that all non-unseal secret files across
// deployments/ have correct content format per ENG-HANDBOOK.md Section 13.3.
// This file implements the schema loader for secret-schemas.yaml.
package secret_content

import (
	_ "embed"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed secret-schemas.yaml
var secretSchemasYAML []byte

// SecretSchema describes a single secret file format rule.
type SecretSchema struct {
	// Filename is the exact secret filename (e.g., "hash-pepper-v3.secret").
	Filename string `yaml:"filename"`
	// Tiers lists which deployment tiers this rule applies to.
	// Valid values: "service", "product", "suite".
	Tiers []string `yaml:"tiers"`
	// ValuePattern is a regex template with {PREFIX}, {PREFIX_US}, {B64URL43} placeholders.
	// Empty for .secret.never files.
	ValuePattern string `yaml:"value_pattern,omitempty"`
	// NeverContent is the exact string expected in .secret.never marker files.
	// Empty for regular .secret files.
	NeverContent string `yaml:"never_content,omitempty"`
	// Description is a human-readable explanation of the rule.
	Description string `yaml:"description"`
}

// SecretSchemas is a slice of all secret schema rules.
type SecretSchemas []SecretSchema

// b64URL43 is the regex component for 43-char base64url nonces.
const b64URL43 = `[A-Za-z0-9_-]{43}`

// LoadSecretSchemas parses the embedded secret-schemas.yaml.
func LoadSecretSchemas() (SecretSchemas, error) {
	var schemas SecretSchemas
	if err := yaml.Unmarshal(secretSchemasYAML, &schemas); err != nil {
		return nil, fmt.Errorf("cannot parse secret-schemas.yaml: %w", err)
	}

	return schemas, nil
}

// ForTier returns only the schema rules applicable to the given tier.
func (s SecretSchemas) ForTier(tier string) SecretSchemas {
	var result SecretSchemas

	for _, schema := range s {
		for _, t := range schema.Tiers {
			if t == tier {
				result = append(result, schema)

				break
			}
		}
	}

	return result
}

// ExpandPattern substitutes {PREFIX}, {PREFIX_US}, and {B64URL43} in a
// value_pattern template, returning a ready-to-compile regex.
func ExpandPattern(pattern, prefix, prefixUS string) string {
	r := strings.NewReplacer(
		"{PREFIX}", prefix,
		"{PREFIX_US}", prefixUS,
		"{B64URL43}", b64URL43,
	)

	return r.Replace(pattern)
}
