// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadCAConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		content     string
		wantErr     bool
		errContains string
	}{
		{
			name: "valid_root_ca",
			content: `ca:
  name: "test-root-ca"
  description: "Test Root CA"
  type: "root"
  parent: ""
  subject:
    common_name: "Test Root CA"
    organization: "Test Org"
  key:
    algorithm: "ECDSA"
    curve_or_size: "P-256"
  validity:
    days: 3650
  max_path_length: 2
`,
			wantErr: false,
		},
		{
			name: "valid_intermediate_ca",
			content: `ca:
  name: "test-intermediate-ca"
  description: "Test Intermediate CA"
  type: "intermediate"
  parent: "test-root-ca"
  subject:
    common_name: "Test Intermediate CA"
  key:
    algorithm: "RSA"
    curve_or_size: "2048"
  validity:
    days: 1825
  max_path_length: 1
`,
			wantErr: false,
		},
		{
			name: "missing_name",
			content: `ca:
  type: "root"
  subject:
    common_name: "Test CA"
  key:
    algorithm: "ECDSA"
    curve_or_size: "P-256"
  validity:
    days: 365
`,
			wantErr:     true,
			errContains: "CA name is required",
		},
		{
			name: "invalid_type",
			content: `ca:
  name: "test-ca"
  type: "invalid"
  subject:
    common_name: "Test CA"
  key:
    algorithm: "ECDSA"
    curve_or_size: "P-256"
  validity:
    days: 365
`,
			wantErr:     true,
			errContains: "invalid CA type",
		},
		{
			name: "intermediate_without_parent",
			content: `ca:
  name: "test-ca"
  type: "intermediate"
  parent: ""
  subject:
    common_name: "Test CA"
  key:
    algorithm: "ECDSA"
    curve_or_size: "P-256"
  validity:
    days: 365
`,
			wantErr:     true,
			errContains: "parent CA is required",
		},
		{
			name: "invalid_ecdsa_curve",
			content: `ca:
  name: "test-ca"
  type: "root"
  subject:
    common_name: "Test CA"
  key:
    algorithm: "ECDSA"
    curve_or_size: "P-128"
  validity:
    days: 365
`,
			wantErr:     true,
			errContains: "invalid ECDSA curve",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create temp file.
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "ca-config.yaml")
			err := os.WriteFile(tmpFile, []byte(tc.content), 0o600)
			require.NoError(t, err)

			// Load config.
			config, err := LoadCAConfig(tmpFile)

			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)
				require.Nil(t, config)
			} else {
				require.NoError(t, err)
				require.NotNil(t, config)
			}
		})
	}
}

func TestLoadProfileConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		content     string
		wantErr     bool
		errContains string
	}{
		{
			name: "valid_tls_server_profile",
			content: `profile:
  name: "tls-server"
  description: "TLS Server Certificate"
  validity:
    max_days: 398
    min_days: 1
    default_days: 365
  key:
    allowed_algorithms:
      - algorithm: "ECDSA"
        allowed_curves:
          - "P-256"
          - "P-384"
    default_algorithm: "ECDSA"
    default_curve_or_size: "P-256"
  key_usage:
    - "digitalSignature"
  extended_key_usage:
    required:
      - "serverAuth"
    optional: []
  subject:
    require_common_name: true
  san:
    allow_dns_names: true
    allow_ip_addresses: true
  extensions:
    required:
      - "keyUsage"
    optional: []
  basic_constraints:
    is_ca: false
  signature:
    preferred:
      - "ECDSA-SHA256"
    forbidden:
      - "MD5"
`,
			wantErr: false,
		},
		{
			name: "missing_name",
			content: `profile:
  validity:
    max_days: 365
    default_days: 30
  key:
    allowed_algorithms:
      - algorithm: "RSA"
        min_size: 2048
`,
			wantErr:     true,
			errContains: "profile name is required",
		},
		{
			name: "default_exceeds_max",
			content: `profile:
  name: "test"
  validity:
    max_days: 30
    default_days: 365
  key:
    allowed_algorithms:
      - algorithm: "RSA"
`,
			wantErr:     true,
			errContains: "default validity days cannot exceed max",
		},
		{
			name: "no_algorithms",
			content: `profile:
  name: "test"
  validity:
    max_days: 365
    default_days: 30
  key:
    allowed_algorithms: []
`,
			wantErr:     true,
			errContains: "at least one allowed algorithm",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create temp file.
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "profile.yaml")
			err := os.WriteFile(tmpFile, []byte(tc.content), 0o600)
			require.NoError(t, err)

			// Load config.
			config, err := LoadProfileConfig(tmpFile)

			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)
				require.Nil(t, config)
			} else {
				require.NoError(t, err)
				require.NotNil(t, config)
			}
		})
	}
}

func TestLoadCAConfig_FileNotFound(t *testing.T) {
	t.Parallel()

	config, err := LoadCAConfig("/nonexistent/path/config.yaml")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read CA config file")
	require.Nil(t, config)
}

func TestLoadProfileConfig_FileNotFound(t *testing.T) {
	t.Parallel()

	config, err := LoadProfileConfig("/nonexistent/path/profile.yaml")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read profile config file")
	require.Nil(t, config)
}

func TestLoadCAConfig_InvalidYAML(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.yaml")
	err := os.WriteFile(tmpFile, []byte("{{invalid yaml"), 0o600)
	require.NoError(t, err)

	config, err := LoadCAConfig(tmpFile)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse CA config file")
	require.Nil(t, config)
}

func TestLoadProfileConfig_InvalidYAML(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.yaml")
	err := os.WriteFile(tmpFile, []byte("{{invalid yaml"), 0o600)
	require.NoError(t, err)

	config, err := LoadProfileConfig(tmpFile)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse profile config file")
	require.Nil(t, config)
}
