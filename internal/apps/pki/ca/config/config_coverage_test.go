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

// writeConfigFile writes config content to a temp file for testing.
func writeConfigFile(t *testing.T, content string) string {
t.Helper()

tmpDir := t.TempDir()
path := filepath.Join(tmpDir, "config.yaml")

err := os.WriteFile(path, []byte(content), 0o600)
require.NoError(t, err)

return path
}

// TestLoadCAConfig_EmptyCommonName tests that ValidateCAConfig returns an error
// when the CA subject common name is empty.
func TestLoadCAConfig_EmptyCommonName(t *testing.T) {
t.Parallel()

content := `ca:
  name: "test-ca"
  type: "root"
  subject:
    common_name: ""
  key:
    algorithm: "ECDSA"
    curve_or_size: "P-256"
  validity:
    days: 3650
`

path := writeConfigFile(t, content)

_, err := LoadCAConfig(path)

require.Error(t, err)
require.Contains(t, err.Error(), "subject common name is required")
}

// TestLoadCAConfig_InvalidValidityDays tests that ValidateCAConfig returns an error
// when the validity days is zero or negative.
func TestLoadCAConfig_InvalidValidityDays(t *testing.T) {
t.Parallel()

content := `ca:
  name: "test-ca"
  type: "root"
  subject:
    common_name: "Test CA"
  key:
    algorithm: "ECDSA"
    curve_or_size: "P-256"
  validity:
    days: 0
`

path := writeConfigFile(t, content)

_, err := LoadCAConfig(path)

require.Error(t, err)
require.Contains(t, err.Error(), "validity days must be positive")
}

// TestLoadCAConfig_InvalidEdDSACurve tests that validateKeyConfig returns an error
// when an invalid EdDSA curve is specified.
func TestLoadCAConfig_InvalidEdDSACurve(t *testing.T) {
t.Parallel()

content := `ca:
  name: "test-ca"
  type: "root"
  subject:
    common_name: "Test CA"
  key:
    algorithm: "EdDSA"
    curve_or_size: "X25519"
  validity:
    days: 3650
`

path := writeConfigFile(t, content)

_, err := LoadCAConfig(path)

require.Error(t, err)
require.Contains(t, err.Error(), "invalid EdDSA curve")
}

// TestLoadCAConfig_InvalidKeyAlgorithm tests that validateKeyConfig returns an error
// when an unrecognized key algorithm is specified.
func TestLoadCAConfig_InvalidKeyAlgorithm(t *testing.T) {
t.Parallel()

content := `ca:
  name: "test-ca"
  type: "root"
  subject:
    common_name: "Test CA"
  key:
    algorithm: "HMAC"
    curve_or_size: "SHA256"
  validity:
    days: 3650
`

path := writeConfigFile(t, content)

_, err := LoadCAConfig(path)

require.Error(t, err)
require.Contains(t, err.Error(), "invalid algorithm")
}
