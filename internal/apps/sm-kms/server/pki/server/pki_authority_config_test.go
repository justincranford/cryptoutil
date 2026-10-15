// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

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
