// Copyright (c) 2025 Justin Cranford
//
//

// Package realm provides realm configuration and tenant isolation for KMS.
// Realms define authentication domains, users, and access policies.
//
// Reference: Session 3 Q6-10, Session 5 Q12-14.
package realm

import (
	"fmt"
	"os"
	"path/filepath"

	googleUuid "github.com/google/uuid"
	"gopkg.in/yaml.v3"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Config holds the complete realm configuration.
type Config struct {
	// Version is the schema version of the config file.
	Version string `yaml:"version" json:"version"`

	// Realms is the list of configured realms.
	Realms []RealmConfig `yaml:"realms" json:"realms"`

	// Defaults is the default settings for all realms.
	Defaults RealmDefaults `yaml:"defaults" json:"defaults"`
}

// RealmConfig defines a single authentication realm.
type RealmConfig struct {
	// ID is the unique UUIDv4 identifier for the realm.
	ID string `yaml:"id" json:"id"`

	// Name is the human-readable name for the realm.
	Name string `yaml:"name" json:"name"`

	// Description is the optional description of the realm.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// Type is the realm type (file, database, ldap, oidc).
	Type RealmType `yaml:"type" json:"type"`

	// Enabled indicates if the realm is active.
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Users is the list of users in a file-based realm.
	Users []UserConfig `yaml:"users,omitempty" json:"users,omitempty"`

	// Roles is the hierarchical role configuration.
	Roles []RoleConfig `yaml:"roles,omitempty" json:"roles,omitempty"`

	// PasswordPolicy is the password policy for this realm.
	PasswordPolicy PasswordPolicyConfig `yaml:"password_policy,omitempty" json:"password_policy,omitempty"`
}

// RealmType defines the type of authentication realm.
type RealmType string

const (
	// RealmTypeFile is a file-based realm with users defined in YAML.
	RealmTypeFile RealmType = "file"

	// RealmTypeDatabase is a database-backed realm.
	RealmTypeDatabase RealmType = "database"

	// RealmTypeLDAP is an LDAP/AD-backed realm.
	RealmTypeLDAP RealmType = "ldap"

	// RealmTypeOIDC is an OIDC provider-backed realm.
	RealmTypeOIDC RealmType = "oidc"
)

// UserConfig defines a user in a file-based realm.
type UserConfig struct {
	// ID is the unique UUIDv7 identifier for the user.
	ID string `yaml:"id" json:"id"`

	// Username is the login username.
	Username string `yaml:"username" json:"username"`

	// PasswordHash is the PBKDF2-HMAC-SHA256 hashed password.
	// Format: $pbkdf2-sha256$iterations$salt$hash
	PasswordHash string `yaml:"password_hash" json:"password_hash"`

	// Email is the optional user email.
	Email string `yaml:"email,omitempty" json:"email,omitempty"`

	// Roles is the list of role names assigned to the user.
	Roles []string `yaml:"roles" json:"roles"`

	// Enabled indicates if the user is active.
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Metadata is optional JSON metadata with validation schema support.
	Metadata map[string]any `yaml:"metadata,omitempty" json:"metadata,omitempty"`

	// MetadataSchema is the JSON schema reference for metadata validation.
	MetadataSchema string `yaml:"metadata_schema,omitempty" json:"metadata_schema,omitempty"`
}

// RoleConfig defines a hierarchical role.
type RoleConfig struct {
	// Name is the unique role name.
	Name string `yaml:"name" json:"name"`

	// Description is the optional role description.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// Permissions is the list of permissions granted by this role.
	Permissions []string `yaml:"permissions" json:"permissions"`

	// Inherits is the list of parent role names this role inherits from.
	Inherits []string `yaml:"inherits,omitempty" json:"inherits,omitempty"`
}

// PasswordPolicyConfig defines PBKDF2 password hashing settings.
// Reference: Session 5 Q12.
type PasswordPolicyConfig struct {
	// Algorithm is the hashing algorithm (default: SHA-256).
	Algorithm string `yaml:"algorithm" json:"algorithm"`

	// Iterations is the PBKDF2 iteration count (default: 600000).
	Iterations int `yaml:"iterations" json:"iterations"`

	// SaltBytes is the salt length in bytes (default: 32).
	SaltBytes int `yaml:"salt_bytes" json:"salt_bytes"`

	// HashBytes is the derived key length in bytes (default: 32).
	HashBytes int `yaml:"hash_bytes" json:"hash_bytes"`
}

// RealmDefaults provides default values for realm configuration.
type RealmDefaults struct {
	// PasswordPolicy is the default password policy for all realms.
	PasswordPolicy PasswordPolicyConfig `yaml:"password_policy" json:"password_policy"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Version: "1.0",
		Realms:  []RealmConfig{},
		Defaults: RealmDefaults{
			PasswordPolicy: DefaultPasswordPolicy(),
		},
	}
}

// DefaultPasswordPolicy returns the default PBKDF2 password policy.
// Reference: Session 5 Q12 - SHA-256, 600K iterations, 32-byte salt.
func DefaultPasswordPolicy() PasswordPolicyConfig {
	return PasswordPolicyConfig{
		Algorithm:  cryptoutilSharedMagic.PBKDF2DefaultAlgorithm,
		Iterations: cryptoutilSharedMagic.PBKDF2DefaultIterations,
		SaltBytes:  cryptoutilSharedMagic.PBKDF2DefaultSaltBytes,
		HashBytes:  cryptoutilSharedMagic.PBKDF2DefaultHashBytes,
	}
}

// LoadConfig loads realm configuration from a YAML file.
func LoadConfig(configDir string) (*Config, error) {
	configPath := filepath.Join(configDir, "realms.yml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist.
			return DefaultConfig(), nil
		}

		return nil, fmt.Errorf("failed to read realms.yml: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse realms.yml: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid realm configuration: %w", err)
	}

	return &config, nil
}

// Validate validates the realm configuration.
func (c *Config) Validate() error {
	seenIDs := make(map[string]bool)
	seenNames := make(map[string]bool)

	for i, realm := range c.Realms {
		if realm.ID == "" {
			return fmt.Errorf("realm[%d]: id is required", i)
		}

		if _, err := googleUuid.Parse(realm.ID); err != nil {
			return fmt.Errorf("realm[%d]: id must be valid UUID: %w", i, err)
		}

		if seenIDs[realm.ID] {
			return fmt.Errorf("realm[%d]: duplicate id %s", i, realm.ID)
		}

		seenIDs[realm.ID] = true

		if realm.Name == "" {
			return fmt.Errorf("realm[%d]: name is required", i)
		}

		if seenNames[realm.Name] {
			return fmt.Errorf("realm[%d]: duplicate name %s", i, realm.Name)
		}

		seenNames[realm.Name] = true

		if !isValidRealmType(realm.Type) {
			return fmt.Errorf("realm[%d]: invalid type %s", i, realm.Type)
		}
	}

	return nil
}

// isValidRealmType checks if a realm type is valid.
func isValidRealmType(t RealmType) bool {
	switch t {
	case RealmTypeFile, RealmTypeDatabase, RealmTypeLDAP, RealmTypeOIDC:
		return true
	default:
		return false
	}
}

// GenerateTenantID generates a new UUIDv4 for tenant/realm IDs.
// Reference: Session 3 Q10, Session 4 Q6-9 - UUIDv4 for max randomness.
func GenerateTenantID() string {
	return googleUuid.New().String()
}
