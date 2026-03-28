// Copyright (c) 2025 Justin Cranford

package codegen_config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestCheck_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Check(logger, map[string][]string{})

	require.NoError(t, err)
}

func TestCheck_ValidConfig(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	apiDir := filepath.Join(tmpDir, "api", "test-service")
	err := os.MkdirAll(apiDir, cryptoutilSharedMagic.CICDTempDirPermissions)
	require.NoError(t, err)

	configFile := filepath.Join(apiDir, "openapi-gen_config_server.yaml")
	content := `package: server
output: server/server.gen.go
generate:
  fiber-server: true
output-options:
  additional-initialisms:
    - IDS
    - JWT
    - JWK
    - JWE
    - JWS
    - OIDC
    - SAML
    - AES
    - GCM
    - CBC
    - RSA
    - EC
    - HMAC
    - SHA
    - TLS
    - IP
    - AI
    - ML
    - KEM
    - PEM
    - DER
    - DSA
    - IKM
`

	err = os.WriteFile(configFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger, map[string][]string{"yaml": {configFile}})

	require.NoError(t, err)
}

func TestCheck_MissingInitialisms(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	apiDir := filepath.Join(tmpDir, "api", "test-service")
	err := os.MkdirAll(apiDir, cryptoutilSharedMagic.CICDTempDirPermissions)
	require.NoError(t, err)

	configFile := filepath.Join(apiDir, "openapi-gen_config_server.yaml")
	content := `package: server
output-options:
  additional-initialisms:
    - JWT
    - JWK
`

	err = os.WriteFile(configFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger, map[string][]string{"yaml": {configFile}})

	require.Error(t, err)
	require.Contains(t, err.Error(), "codegen-config")
}

func TestCheck_NonAPIFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	configFile := filepath.Join(tmpDir, "openapi-gen_config_server.yaml")
	content := `package: server
`

	err := os.WriteFile(configFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger, map[string][]string{"yaml": {configFile}})

	require.NoError(t, err, "non-api config files should be skipped")
}

func TestCheck_WithComments(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	apiDir := filepath.Join(tmpDir, "api", "test-service")
	err := os.MkdirAll(apiDir, cryptoutilSharedMagic.CICDTempDirPermissions)
	require.NoError(t, err)

	configFile := filepath.Join(apiDir, "openapi-gen_config_models.yaml")
	content := `package: model
output-options:
  additional-initialisms:
    - IDS    # Intrusion Detection System
    - JWT    # JSON Web Token
    - JWK    # JSON Web Key
    - JWE    # JSON Web Encryption
    - JWS    # JSON Web Signature
    - OIDC   # OpenID Connect
    - SAML   # Security Assertion Markup Language
    - AES    # Advanced Encryption Standard
    - GCM    # Galois/Counter Mode
    - CBC    # Cipher Block Chaining
    - RSA    # Rivest-Shamir-Adleman
    - EC     # Elliptic Curve
    - HMAC   # Hash-based Message Authentication Code
    - SHA    # Secure Hash Algorithm
    - TLS    # Transport Layer Security
    - IP     # Internet Protocol
    - AI     # Artificial Intelligence
    - ML     # Machine Learning
    - KEM    # Key Encapsulation Mechanism
    - PEM    # Privacy Enhanced Mail
    - DER    # Distinguished Encoding Rules
    - DSA    # Digital Signature Algorithm
    - IKM    # Input Keying Material
`

	err = os.WriteFile(configFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger, map[string][]string{"yaml": {configFile}})

	require.NoError(t, err, "initialisms with trailing comments should be recognized")
}

func TestFindMissingInitialisms_NonexistentFile(t *testing.T) {
	t.Parallel()

	_, err := findMissingInitialisms("/nonexistent/file.yaml")

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to open file")
}
