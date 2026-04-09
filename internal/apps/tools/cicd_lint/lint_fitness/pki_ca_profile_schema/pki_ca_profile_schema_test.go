// Copyright (c) 2025 Justin Cranford

package pki_ca_profile_schema

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// -----------------------------------------------------------------------
// Test helpers
// -----------------------------------------------------------------------

// buildProfileRoot creates a temp root dir with a pkica profiles directory.
func buildProfileRoot(t *testing.T) string {
	t.Helper()

	rootDir := t.TempDir()
	profilesDir := filepath.Join(rootDir, filepath.FromSlash(cryptoutilSharedMagic.CICDPKICAProfilesDir))
	require.NoError(t, os.MkdirAll(profilesDir, cryptoutilSharedMagic.CacheFilePermissions))

	return rootDir
}

// writeProfileYAML writes a YAML file in the profiles directory.
func writeProfileYAML(t *testing.T, rootDir, name, content string) string {
	t.Helper()

	profilesDir := filepath.Join(rootDir, filepath.FromSlash(cryptoutilSharedMagic.CICDPKICAProfilesDir))
	path := filepath.Join(profilesDir, name)
	require.NoError(t, os.WriteFile(path, []byte(content), cryptoutilSharedMagic.FilePermissionsDefault))

	return path
}

// minValidProfile returns a minimal valid profile YAML.
func minValidProfile(name string) string {
	return fmt.Sprintf(`profile:
  name: %q
  description: "Test profile for %s"
  validity:
    max_days: 365
    min_days: 1
    default_days: 90
  key:
    allowed_algorithms:
      - algorithm: "RSA"
        min_size: 2048
        max_size: 4096
    default_algorithm: "RSA"
    default_curve_or_size: 2048
  key_usage:
    - "digitalSignature"
  extended_key_usage:
    required:
      - "serverAuth"
    optional: []
`, name, name)
}

func newTestLogger(t *testing.T) *cryptoutilCmdCicdCommon.Logger {
	t.Helper()

	return cryptoutilCmdCicdCommon.NewLogger("test-pki-ca-profile-schema")
}

// -----------------------------------------------------------------------
// CheckInDir - happy paths
// -----------------------------------------------------------------------

func TestCheckInDir_HappyPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(t *testing.T, rootDir string)
	}{
		{
			name:  "no profiles",
			setup: func(_ *testing.T, _ string) {},
		},
		{
			name: "valid RSA profile",
			setup: func(t *testing.T, rootDir string) {
				t.Helper()
				writeProfileYAML(t, rootDir, "tls-server.yaml", minValidProfile("tls-server"))
			},
		},
		{
			name: "Ed25519 null curve",
			setup: func(t *testing.T, rootDir string) {
				t.Helper()
				writeProfileYAML(t, rootDir, "ssh-host.yaml", `profile:
  name: "ssh-host"
  description: "SSH Host Certificate"
  validity: {max_days: 365, min_days: 1, default_days: 90}
  key:
    allowed_algorithms:
      - algorithm: "Ed25519"
    default_algorithm: "Ed25519"
    default_curve_or_size: null
  key_usage: ["digitalSignature"]
  extended_key_usage: {required: [], optional: []}
`)
			},
		},
		{
			name: "min one day short lived",
			setup: func(t *testing.T, rootDir string) {
				t.Helper()
				writeProfileYAML(t, rootDir, "k8s-workload.yaml", `profile:
  name: "k8s-workload"
  description: "Short-lived Kubernetes Workload Certificate"
  validity: {max_days: 1, min_days: 1, default_days: 1}
  key:
    allowed_algorithms:
      - algorithm: "ECDSA"
        allowed_curves: ["P-256"]
    default_algorithm: "ECDSA"
    default_curve_or_size: "P-256"
  key_usage: ["digitalSignature"]
  extended_key_usage: {required: [], optional: []}
`)
			},
		},
		{
			name: "non-YAML file skipped",
			setup: func(t *testing.T, rootDir string) {
				t.Helper()

				profilesDir := filepath.Join(rootDir, filepath.FromSlash(cryptoutilSharedMagic.CICDPKICAProfilesDir))
				require.NoError(t, os.WriteFile(filepath.Join(profilesDir, "profile-schema.json"), []byte("{}"), cryptoutilSharedMagic.FilePermissionsDefault))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootDir := buildProfileRoot(t)
			tc.setup(t, rootDir)
			logger := newTestLogger(t)

			err := CheckInDir(logger, rootDir, os.ReadFile, filepath.WalkDir)

			require.NoError(t, err)
		})
	}
}

// -----------------------------------------------------------------------
// CheckInDir - violations
// -----------------------------------------------------------------------

func TestCheckInDir_Violations(t *testing.T) {
	t.Parallel()

	rsaBase := `profile:
  name: "test"
  description: "Test"
  validity: {max_days: 365, min_days: 1, default_days: 90}
  key:
    allowed_algorithms:
      - algorithm: "RSA"
    default_algorithm: "RSA"
    default_curve_or_size: 2048
  key_usage: ["digitalSignature"]
  extended_key_usage: {required: [], optional: []}`

	tests := []struct {
		name    string
		file    string
		yaml    string
		wantErr string
	}{
		{
			name:    "missing profile field",
			file:    "bad.yaml",
			yaml:    "validity:\n  max_days: 365\n",
			wantErr: "violation",
		},
		{
			name: "empty profile name",
			file: "nameless.yaml",
			yaml: `profile:
  name: ""
  description: "Test"
  validity: {max_days: 365, min_days: 1, default_days: 90}
  key:
    allowed_algorithms:
      - algorithm: "RSA"
        min_size: 2048
        max_size: 4096
    default_algorithm: "RSA"
    default_curve_or_size: 2048
  key_usage: ["digitalSignature"]
  extended_key_usage: {required: [], optional: []}`,
			wantErr: "profile.name",
		},
		{
			name: "invalid max days less than min days",
			file: "bad-validity.yaml",
			yaml: `profile:
  name: "test"
  description: "Test"
  validity: {max_days: 5, min_days: 10, default_days: 7}
  key:
    allowed_algorithms:
      - algorithm: "RSA"
    default_algorithm: "RSA"
    default_curve_or_size: 2048
  key_usage: ["digitalSignature"]
  extended_key_usage: {required: [], optional: []}`,
			wantErr: "max_days",
		},
		{
			name: "default days out of range",
			file: "bad-default.yaml",
			yaml: `profile:
  name: "test"
  description: "Test"
  validity: {max_days: 365, min_days: 1, default_days: 500}
  key:
    allowed_algorithms:
      - algorithm: "RSA"
    default_algorithm: "RSA"
    default_curve_or_size: 2048
  key_usage: ["digitalSignature"]
  extended_key_usage: {required: [], optional: []}`,
			wantErr: "default_days",
		},
		{
			name: "max days exceeds cap",
			file: "too-long.yaml",
			yaml: `profile:
  name: "too-long"
  description: "Test"
  validity: {max_days: 99999, min_days: 1, default_days: 90}
  key:
    allowed_algorithms:
      - algorithm: "RSA"
    default_algorithm: "RSA"
    default_curve_or_size: 2048
  key_usage: ["digitalSignature"]
  extended_key_usage: {required: [], optional: []}`,
			wantErr: "absolute cap",
		},
		{
			name: "empty allowed algorithms",
			file: "no-alg.yaml",
			yaml: `profile:
  name: "test"
  description: "Test"
  validity: {max_days: 365, min_days: 1, default_days: 90}
  key:
    allowed_algorithms: []
    default_algorithm: "RSA"
    default_curve_or_size: 2048
  key_usage: ["digitalSignature"]
  extended_key_usage: {required: [], optional: []}`,
			wantErr: "at least one",
		},
		{
			name: "unknown algorithm",
			file: "bad-alg.yaml",
			yaml: `profile:
  name: "test"
  description: "Test"
  validity: {max_days: 365, min_days: 1, default_days: 90}
  key:
    allowed_algorithms:
      - algorithm: "DSA"
    default_algorithm: "DSA"
    default_curve_or_size: 1024
  key_usage: ["digitalSignature"]
  extended_key_usage: {required: [], optional: []}`,
			wantErr: "known algorithm",
		},
		{
			name: "empty key usage",
			file: "no-ku.yaml",
			yaml: `profile:
  name: "test"
  description: "Test"
  validity: {max_days: 365, min_days: 1, default_days: 90}
  key:
    allowed_algorithms:
      - algorithm: "RSA"
    default_algorithm: "RSA"
    default_curve_or_size: 2048
  key_usage: []
  extended_key_usage: {required: [], optional: []}`,
			wantErr: "key_usage",
		},
		{
			name: "unknown key usage",
			file: "unknown-ku.yaml",
			yaml: `profile:
  name: "test"
  description: "Test"
  validity: {max_days: 365, min_days: 1, default_days: 90}
  key:
    allowed_algorithms:
      - algorithm: "RSA"
    default_algorithm: "RSA"
    default_curve_or_size: 2048
  key_usage: ["superPower"]
  extended_key_usage: {required: [], optional: []}`,
			wantErr: "unknown value",
		},
		{
			name: "missing extended key usage",
			file: "no-eku.yaml",
			yaml: `profile:
  name: "test"
  description: "Test"
  validity: {max_days: 365, min_days: 1, default_days: 90}
  key:
    allowed_algorithms:
      - algorithm: "RSA"
    default_algorithm: "RSA"
    default_curve_or_size: 2048
  key_usage: ["digitalSignature"]`,
			wantErr: "extended_key_usage",
		},
		{
			name: "SAN negative max entries",
			file: "bad-san.yaml",
			yaml: rsaBase + `
  san:
    allow_dns_names: true
    allow_ip_addresses: false
    allow_email_addresses: false
    allow_uris: false
    require_at_least_one: true
    max_entries: -1`,
			wantErr: "max_entries",
		},
		{
			name: "SAN missing required fields",
			file: "partial-san.yaml",
			yaml: rsaBase + `
  san:
    allow_dns_names: true`,
			wantErr: "san.",
		},
		{
			name:    "invalid YAML skips with error",
			file:    "bad-syntax.yaml",
			yaml:    "!!! not: valid: yaml: [unparseable",
			wantErr: "failed to parse",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootDir := buildProfileRoot(t)
			writeProfileYAML(t, rootDir, tc.file, tc.yaml)
			logger := newTestLogger(t)

			err := CheckInDir(logger, rootDir, os.ReadFile, filepath.WalkDir)

			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

// -----------------------------------------------------------------------
// CheckInDir - seam injection tests
// -----------------------------------------------------------------------

func TestCheckInDir_WalkError(t *testing.T) {
	t.Parallel()

	rootDir := buildProfileRoot(t)
	logger := newTestLogger(t)

	err := CheckInDir(logger, rootDir, os.ReadFile, func(_ string, _ fs.WalkDirFunc) error {
		return errors.New("injected walk error")
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "injected walk error")
}

func TestCheckInDir_ReadFileError(t *testing.T) {
	t.Parallel()

	// Use real walk to find the file, then inject read error.
	rootDir := buildProfileRoot(t)
	writeProfileYAML(t, rootDir, "any.yaml", minValidProfile("any"))
	logger := newTestLogger(t)

	err := CheckInDir(logger, rootDir, func(_ string) ([]byte, error) {
		return nil, errors.New("injected read error")
	}, filepath.WalkDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "injected read error")
}

// -----------------------------------------------------------------------
// CheckInDir - zero min_days with ECDSA+P-256
// -----------------------------------------------------------------------

func TestCheckInDir_ZeroMinDaysIsError(t *testing.T) {
	t.Parallel()

	rootDir := buildProfileRoot(t)
	writeProfileYAML(t, rootDir, "zero-min.yaml", `profile:
  name: "zero-min"
  description: "Invalid zero min_days"
  validity: {max_days: 30, min_days: 0, default_days: 1}
  key:
    allowed_algorithms:
      - algorithm: "ECDSA"
        allowed_curves: ["P-256"]
    default_algorithm: "ECDSA"
    default_curve_or_size: "P-256"
  key_usage: ["digitalSignature"]
  extended_key_usage: {required: [], optional: []}
`)
	logger := newTestLogger(t)

	err := CheckInDir(logger, rootDir, os.ReadFile, filepath.WalkDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "min_days must be >= 1")
}

// Copyright (c) 2025 Justin Cranford
