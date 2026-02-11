// Copyright (c) 2025 Justin Cranford
//
//

package realm

import (
	"os"
	"path/filepath"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Test UUIDs generated once per test run for consistency.
var (
	realmTestID1 = googleUuid.Must(googleUuid.NewV7()).String()
	userTestID1  = googleUuid.Must(googleUuid.NewV7()).String()
)

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()

	require.NotNil(t, config)
	require.Equal(t, "1.0", config.Version)
	require.Empty(t, config.Realms)
	require.Equal(t, cryptoutilSharedMagic.PBKDF2DefaultAlgorithm, config.Defaults.PasswordPolicy.Algorithm)
	require.Equal(t, cryptoutilSharedMagic.PBKDF2DefaultIterations, config.Defaults.PasswordPolicy.Iterations)
	require.Equal(t, cryptoutilSharedMagic.PBKDF2DefaultSaltBytes, config.Defaults.PasswordPolicy.SaltBytes)
	require.Equal(t, cryptoutilSharedMagic.PBKDF2DefaultHashBytes, config.Defaults.PasswordPolicy.HashBytes)
}

func TestDefaultPasswordPolicy(t *testing.T) {
	t.Parallel()

	policy := DefaultPasswordPolicy()

	require.Equal(t, "SHA-256", policy.Algorithm)
	require.Equal(t, 600000, policy.Iterations)
	require.Equal(t, 32, policy.SaltBytes)
	require.Equal(t, 32, policy.HashBytes)
}

func TestGenerateTenantID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "generates valid UUIDv4"},
		{name: "generates unique UUIDs"},
		{name: "generates properly formatted UUID"},
	}

	generatedIDs := make(map[string]bool)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			id := GenerateTenantID()
			require.NotEmpty(t, id)
			require.Len(t, id, 36) // UUID format: 8-4-4-4-12 = 36 chars.
			require.Contains(t, id, "-")

			// Each call should generate unique ID.
			require.False(t, generatedIDs[id], "expected unique ID, got duplicate")

			generatedIDs[id] = true
		})
	}
}

func TestRealmConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty config is valid",
			config:  Config{},
			wantErr: false,
		},
		{
			name: "valid single realm",
			config: Config{
				Realms: []RealmConfig{
					{
						ID:      realmTestID1,
						Name:    "test-realm",
						Type:    RealmTypeFile,
						Enabled: true,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing realm id",
			config: Config{
				Realms: []RealmConfig{
					{
						Name:    "test-realm",
						Type:    RealmTypeFile,
						Enabled: true,
					},
				},
			},
			wantErr: true,
			errMsg:  "id is required",
		},
		{
			name: "invalid realm id format",
			config: Config{
				Realms: []RealmConfig{
					{
						ID:      "not-a-uuid",
						Name:    "test-realm",
						Type:    RealmTypeFile,
						Enabled: true,
					},
				},
			},
			wantErr: true,
			errMsg:  "must be valid UUID",
		},
		{
			name: "missing realm name",
			config: Config{
				Realms: []RealmConfig{
					{
						ID:      realmTestID1,
						Type:    RealmTypeFile,
						Enabled: true,
					},
				},
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "duplicate realm id",
			config: Config{
				Realms: []RealmConfig{
					{
						ID:      realmTestID1,
						Name:    "realm1",
						Type:    RealmTypeFile,
						Enabled: true,
					},
					{
						ID:      realmTestID1,
						Name:    "realm2",
						Type:    RealmTypeFile,
						Enabled: true,
					},
				},
			},
			wantErr: true,
			errMsg:  "duplicate id",
		},
		{
			name: "duplicate realm name",
			config: Config{
				Realms: []RealmConfig{
					{
						ID:      realmTestID1,
						Name:    "same-name",
						Type:    RealmTypeFile,
						Enabled: true,
					},
					{
						ID:      userTestID1,
						Name:    "same-name",
						Type:    RealmTypeFile,
						Enabled: true,
					},
				},
			},
			wantErr: true,
			errMsg:  "duplicate name",
		},
		{
			name: "invalid realm type",
			config: Config{
				Realms: []RealmConfig{
					{
						ID:      realmTestID1,
						Name:    "test-realm",
						Type:    RealmType("invalid"),
						Enabled: true,
					},
				},
			},
			wantErr: true,
			errMsg:  "invalid type",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.config.Validate()
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLoadConfig_NonexistentFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	config, err := LoadConfig(tempDir)

	require.NoError(t, err)
	require.NotNil(t, config)
	require.Equal(t, "1.0", config.Version)
}

func TestLoadConfig_ValidFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "realms.yml")

	yamlContent := `
version: "1.0"
realms:
  - id: "` + realmTestID1 + `"
    name: "demo-realm"
    description: "Demo realm for testing"
    type: "file"
    enabled: true
    users:
      - id: "` + userTestID1 + `"
        username: "admin"
        password_hash: "$` + cryptoutilSharedMagic.PBKDF2DefaultHashName + `$600000$salt$hash"
        email: "admin@example.com"
        roles: ["admin"]
        enabled: true
    roles:
      - name: "admin"
        description: "Administrator role"
        permissions: ["kms:admin", "kms:read", "kms:write"]
defaults:
  password_policy:
    algorithm: "SHA-256"
    iterations: 600000
    salt_bytes: 32
    hash_bytes: 32
`

	err := os.WriteFile(configPath, []byte(yamlContent), 0o600)
	require.NoError(t, err)

	config, err := LoadConfig(tempDir)

	require.NoError(t, err)
	require.NotNil(t, config)
	require.Equal(t, "1.0", config.Version)
	require.Len(t, config.Realms, 1)
	require.Equal(t, "demo-realm", config.Realms[0].Name)
	require.Equal(t, RealmTypeFile, config.Realms[0].Type)
	require.True(t, config.Realms[0].Enabled)
	require.Len(t, config.Realms[0].Users, 1)
	require.Equal(t, "admin", config.Realms[0].Users[0].Username)
	require.Len(t, config.Realms[0].Roles, 1)
	require.Equal(t, "admin", config.Realms[0].Roles[0].Name)
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "realms.yml")

	invalidYAML := `
version: "1.0"
realms:
  - id: [invalid yaml
`

	err := os.WriteFile(configPath, []byte(invalidYAML), 0o600)
	require.NoError(t, err)

	config, err := LoadConfig(tempDir)

	require.Error(t, err)
	require.Nil(t, config)
	require.Contains(t, err.Error(), "failed to parse realms.yml")
}

func TestIsValidRealmType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		realmType RealmType
		valid     bool
	}{
		{name: "file type", realmType: RealmTypeFile, valid: true},
		{name: "database type", realmType: RealmTypeDatabase, valid: true},
		{name: "ldap type", realmType: RealmTypeLDAP, valid: true},
		{name: "oidc type", realmType: RealmTypeOIDC, valid: true},
		{name: "empty type", realmType: RealmType(""), valid: false},
		{name: "invalid type", realmType: RealmType("invalid"), valid: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.valid, isValidRealmType(tc.realmType))
		})
	}
}
