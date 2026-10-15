// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	pkiAuthorityTestSQLite1Distribution = `pki-distribution:
  crl-url: https://sm-kms-app-sqlite-1:8080/pki/cas/sm-kms-pki-test/crl
  ocsp-url: https://sm-kms-app-sqlite-1:8080/pki/ocsp/sm-kms-pki-test
`
)

var pkiAuthorityTestProjectRootRelative = filepath.Join("..", "..", "..", "..", "..", "..")

func TestPKIAuthorityConfigFromConfigFiles(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "pki.yml")
	require.NoError(t, os.WriteFile(path, []byte(`
pki-authority:
  id: sm-kms-pki-test
  common-name: Test Authority
  organization: [Example Org, PKI Team]
  country: [US, CA]
  key-algorithm: ECDSA
  ecdsa-curve: P-256
  validity: 48h
  max-path-length: 1
  crl-url: https://pki.example.test/pki/cas/sm-kms-pki-test/crl
  ocsp-url: https://pki.example.test/pki/ocsp/sm-kms-pki-test
`), 0o600))

	config, err := PKIAuthorityConfigFromConfigFiles([]string{path})
	require.NoError(t, err)
	require.Equal(t, "sm-kms-pki-test", config.ID)
	require.Equal(t, "Test Authority", config.Subject.CommonName)
	require.Equal(t, []string{"Example Org", "PKI Team"}, config.Subject.Organization)
	require.Equal(t, []string{"US", "CA"}, config.Subject.Country)
	require.Equal(t, 48*time.Hour, config.Validity)
	require.Equal(t, 1, config.MaxPathLength)
	require.Equal(t, pkiAuthorityCurveP256, config.KeySpec.ECDSACurve)
	require.Equal(t, "https://pki.example.test/pki/cas/sm-kms-pki-test/crl", config.CRLURL)
	require.Equal(t, "https://pki.example.test/pki/ocsp/sm-kms-pki-test", config.OCSPURL)
}

func TestPKIAuthorityConfigFromDeploymentConfigs(t *testing.T) {
	t.Parallel()

	projectRoot := pkiAuthorityTestProjectRootRelative
	deploymentConfigDir := filepath.Join(projectRoot, "deployments", cryptoutilSharedMagic.OTLPServiceSMKMS, "config")
	commonPath := filepath.Join(deploymentConfigDir, "sm-kms-app-domain-common.yml")

	tests := []struct {
		name         string
		configFiles  []string
		wantEndpoint string
	}{
		{
			name: "SQLite instance 1 publishes through its service",
			configFiles: []string{
				commonPath,
				filepath.Join(deploymentConfigDir, "sm-kms-app-domain-sqlite-1.yml"),
			},
			wantEndpoint: "https://sm-kms-app-sqlite-1:8080",
		},
		{
			name: "SQLite instance 2 publishes through its service",
			configFiles: []string{
				commonPath,
				filepath.Join(deploymentConfigDir, "sm-kms-app-domain-sqlite-2.yml"),
			},
			wantEndpoint: "https://sm-kms-app-sqlite-2:8080",
		},
		{
			name: "PostgreSQL instances publish through the shared issuer service",
			configFiles: []string{
				commonPath,
				filepath.Join(deploymentConfigDir, "sm-kms-app-domain-postgresql-1.yml"),
			},
			wantEndpoint: "https://sm-kms-app-postgres-1:8080",
		},
		{
			name: "local development publishes through the host port",
			configFiles: []string{
				filepath.Join(
					projectRoot,
					cryptoutilSharedMagic.CICDConfigsDir,
					cryptoutilSharedMagic.OTLPServiceSMKMS,
					"sm-kms-domain.yml",
				),
			},
			wantEndpoint: "https://localhost:8000",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config, err := PKIAuthorityConfigFromConfigFiles(tc.configFiles)
			require.NoError(t, err)
			require.Equal(t, cryptoutilSharedMagic.PKIDefaultAuthorityID, config.ID)
			require.Equal(t, tc.wantEndpoint+"/pki/cas/"+cryptoutilSharedMagic.PKIDefaultAuthorityID+"/crl", config.CRLURL)
			require.Equal(t, tc.wantEndpoint+"/pki/ocsp/"+cryptoutilSharedMagic.PKIDefaultAuthorityID, config.OCSPURL)
		})
	}
}

func TestPKIAuthorityConfigFromConfigFiles_DistributionOverrides(t *testing.T) {
	t.Parallel()

	commonPath := filepath.Join(t.TempDir(), "common.yml")
	require.NoError(t, os.WriteFile(commonPath, []byte(`
pki-authority:
  id: sm-kms-pki-test
  common-name: Test Authority
  organization: [Example Org]
  country: [US]
  key-algorithm: ECDSA
  ecdsa-curve: P-384
  validity: 48h
  max-path-length: 1
pki-distribution:
  crl-url: https://sm-kms-app-postgres-1:8080/pki/cas/sm-kms-pki-test/crl
  ocsp-url: https://sm-kms-app-postgres-1:8080/pki/ocsp/sm-kms-pki-test
`), 0o600))

	tests := []struct {
		name        string
		overrides   []string
		wantCRLURL  string
		wantOCSPURL string
	}{
		{
			name:        "sqlite instance overrides common PostgreSQL endpoints",
			overrides:   []string{pkiAuthorityTestSQLite1Distribution},
			wantCRLURL:  "https://sm-kms-app-sqlite-1:8080/pki/cas/sm-kms-pki-test/crl",
			wantOCSPURL: "https://sm-kms-app-sqlite-1:8080/pki/ocsp/sm-kms-pki-test",
		},
		{
			name: "later instance override wins",
			overrides: []string{
				pkiAuthorityTestSQLite1Distribution,
				`pki-distribution:
  crl-url: https://sm-kms-app-sqlite-2:8080/pki/cas/sm-kms-pki-test/crl
  ocsp-url: https://sm-kms-app-sqlite-2:8080/pki/ocsp/sm-kms-pki-test
`,
			},
			wantCRLURL:  "https://sm-kms-app-sqlite-2:8080/pki/cas/sm-kms-pki-test/crl",
			wantOCSPURL: "https://sm-kms-app-sqlite-2:8080/pki/ocsp/sm-kms-pki-test",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			paths := []string{commonPath}

			for _, content := range tc.overrides {
				path := filepath.Join(t.TempDir(), "instance.yml")
				require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
				paths = append(paths, path)
			}

			config, err := PKIAuthorityConfigFromConfigFiles(paths)
			require.NoError(t, err)
			require.Equal(t, tc.wantCRLURL, config.CRLURL)
			require.Equal(t, tc.wantOCSPURL, config.OCSPURL)
		})
	}
}

func TestPKIAuthorityConfigFromConfigFiles_ValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name: "invalid duration",
			content: `pki-authority:
  id: test
  common-name: Test
  key-algorithm: ECDSA
  ecdsa-curve: P-384
  validity: invalid
  max-path-length: 2
`,
			wantErr: "parse validity",
		},
		{
			name: "unsupported algorithm",
			content: `pki-authority:
  id: test
  common-name: Test
  key-algorithm: unsupported
  validity: 24h
  max-path-length: 2
`,
			wantErr: "unsupported key algorithm",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "pki.yml")
			require.NoError(t, os.WriteFile(path, []byte(tc.content), 0o600))

			_, err := PKIAuthorityConfigFromConfigFiles([]string{path})
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestPKIAuthorityConfigFromConfigFiles_DefaultAndErrors(t *testing.T) {
	t.Parallel()

	config, err := PKIAuthorityConfigFromConfigFiles(nil)
	require.NoError(t, err)
	require.Equal(t, DefaultPKIAuthorityConfig(), config)

	_, err = PKIAuthorityConfigFromConfigFiles([]string{filepath.Join(t.TempDir(), "missing.yml")})
	require.ErrorContains(t, err, "read PKI authority configuration file")
}

func TestPKIAuthorityConfigValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		configure func(*PKIAuthorityConfig)
		wantErr   string
	}{
		{
			name: "missing ID",
			configure: func(config *PKIAuthorityConfig) {
				config.ID = " "
			},
			wantErr: "authority ID",
		},
		{
			name: "missing common name",
			configure: func(config *PKIAuthorityConfig) {
				config.Subject.CommonName = ""
			},
			wantErr: "common name",
		},
		{
			name: "negative path length",
			configure: func(config *PKIAuthorityConfig) {
				config.MaxPathLength = -1
			},
			wantErr: "must not be negative",
		},
		{
			name: "unsupported curve",
			configure: func(config *PKIAuthorityConfig) {
				config.KeySpec.ECDSACurve = "P-224"
			},
			wantErr: "unsupported authority ECDSA curve",
		},
		{
			name: "CRL URL must use issuer-specific HTTPS path",
			configure: func(config *PKIAuthorityConfig) {
				config.CRLURL = "http://pki.example.test/pki/cas/sm-kms-pki-development-authority/crl"
			},
			wantErr: "absolute HTTPS URL",
		},
		{
			name: "OCSP URL must identify its issuer",
			configure: func(config *PKIAuthorityConfig) {
				config.OCSPURL = "https://pki.example.test/pki/ocsp/another-authority"
			},
			wantErr: "URL path must be",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := DefaultPKIAuthorityConfig()
			tc.configure(&config)

			require.ErrorContains(t, config.validate(), tc.wantErr)
		})
	}
}
