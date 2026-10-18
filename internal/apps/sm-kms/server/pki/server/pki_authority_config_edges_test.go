// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilCACrypto "cryptoutil/internal/apps/sm-kms/server/pki/crypto"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestPKIAuthorityConfigValidationEdges(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		configure func(*PKIAuthorityConfig)
		wantErr   string
	}{
		{
			name: "untrimmed ID",
			configure: func(config *PKIAuthorityConfig) {
				config.ID += pkiTestWhitespace
			},
			wantErr: "authority ID must be non-empty and trimmed",
		},
		{
			name: "ID is not a path segment",
			configure: func(config *PKIAuthorityConfig) {
				config.ID += "/child"
			},
			wantErr: "single URL path segment",
		},
		{
			name: "zero validity",
			configure: func(config *PKIAuthorityConfig) {
				config.Validity = 0
			},
			wantErr: "authority validity must be positive",
		},
		{
			name: "validity exceeds maximum",
			configure: func(config *PKIAuthorityConfig) {
				config.Validity = cryptoutilSharedMagic.TLSDefaultMaxCACertDuration + time.Second
			},
			wantErr: "no greater than",
		},
		{
			name: "CRL URL parse error",
			configure: func(config *PKIAuthorityConfig) {
				config.CRLURL = "://["
			},
			wantErr: pkiTestInvalidAuthorityCRLURL,
		},
		{
			name: "CRL URL missing host",
			configure: func(config *PKIAuthorityConfig) {
				config.CRLURL = "https:///pki/cas/" + config.ID + "/crl"
			},
			wantErr: pkiTestAbsoluteHTTPSURL,
		},
		{
			name: "CRL URL contains user information",
			configure: func(config *PKIAuthorityConfig) {
				config.CRLURL = "https://user@pki.example.test/pki/cas/" + config.ID + "/crl"
			},
			wantErr: pkiTestAbsoluteHTTPSURL,
		},
		{
			name: "OCSP URL contains fragment",
			configure: func(config *PKIAuthorityConfig) {
				config.OCSPURL = "https://pki.example.test/pki/ocsp/" + config.ID + "#fragment"
			},
			wantErr: pkiTestAbsoluteHTTPSURL,
		},
		{
			name: "CRL URL contains query",
			configure: func(config *PKIAuthorityConfig) {
				config.CRLURL = "https://pki.example.test/pki/cas/" + config.ID + "/crl?format=pem"
			},
			wantErr: "must not include a query",
		},
		{
			name: "RSA key is too small",
			configure: func(config *PKIAuthorityConfig) {
				config.KeySpec = cryptoutilCACrypto.KeySpec{Type: cryptoutilCACrypto.KeyTypeRSA, RSABits: cryptoutilCACrypto.MinRSAKeyBits - 1}
			},
			wantErr: "must be at least",
		},
		{
			name: "unsupported EdDSA curve",
			configure: func(config *PKIAuthorityConfig) {
				config.KeySpec = cryptoutilCACrypto.KeySpec{Type: cryptoutilCACrypto.KeyTypeEdDSA, EdDSACurve: cryptoutilSharedMagic.EdCurveEd448}
			},
			wantErr: "unsupported authority EdDSA curve",
		},
		{
			name: "unsupported key type",
			configure: func(config *PKIAuthorityConfig) {
				config.KeySpec.Type = pkiTestUnsupported
			},
			wantErr: "unsupported authority key type",
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

	for _, spec := range []cryptoutilCACrypto.KeySpec{
		{Type: cryptoutilCACrypto.KeyTypeRSA, RSABits: cryptoutilCACrypto.MinRSAKeyBits},
		{Type: cryptoutilCACrypto.KeyTypeEdDSA, EdDSACurve: cryptoutilSharedMagic.EdCurveEd25519},
	} {
		t.Run(string(spec.Type), func(t *testing.T) {
			t.Parallel()

			config := DefaultPKIAuthorityConfig()
			config.KeySpec = spec
			require.NoError(t, config.validate())
		})
	}
}

func TestPKIAuthorityYAMLToConfigAlgorithmsAndErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		algorithm string
		curve     string
		rsaBits   int
		validity  string
		wantType  cryptoutilCACrypto.KeyType
		wantErr   string
	}{
		{name: cryptoutilSharedMagic.ECDSA, algorithm: " ecdsa ", curve: pkiAuthorityCurveP256, validity: pkiTestOneHourValidity, wantType: cryptoutilCACrypto.KeyTypeECDSA},
		{name: cryptoutilSharedMagic.KeyTypeRSA, algorithm: cryptoutilSharedMagic.KeyTypeRSA, rsaBits: cryptoutilCACrypto.MinRSAKeyBits, validity: pkiTestOneHourValidity, wantType: cryptoutilCACrypto.KeyTypeRSA},
		{name: cryptoutilSharedMagic.SessionJWSAlgorithmEdDSA, algorithm: " eddsa ", curve: cryptoutilSharedMagic.EdCurveEd25519, validity: pkiTestOneHourValidity, wantType: cryptoutilCACrypto.KeyTypeEdDSA},
		{name: "invalid validity", algorithm: cryptoutilSharedMagic.ECDSA, curve: pkiAuthorityCurveP256, validity: cryptoutilSharedMagic.DevSetupTestInvalidName, wantErr: pkiTestParseValidity},
		{name: pkiTestUnsupportedAlgorithm, algorithm: pkiTestUnsupported, validity: pkiTestOneHourValidity, wantErr: pkiTestUnsupportedKeyAlgorithm},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config, err := (pkiAuthorityYAML{
				ID:           "yaml-authority",
				CommonName:   "YAML Authority",
				KeyAlgorithm: tc.algorithm,
				ECDSACurve:   tc.curve,
				RSABits:      tc.rsaBits,
				EdDSACurve:   tc.curve,
				Validity:     tc.validity,
			}).toConfig()
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantType, config.KeySpec.Type)
			require.Equal(t, time.Hour, config.Validity)
		})
	}
}

func TestPKIAuthorityConfigFilesMalformedSections(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		contents []string
		wantErr  string
	}{
		{name: "malformed YAML", contents: []string{"pki-authority: ["}, wantErr: "parse service configuration file"},
		{name: "authority section is not a mapping", contents: []string{"pki-authority: invalid"}, wantErr: "parse pki-authority section"},
		{
			name: "authority is defined more than once",
			contents: []string{
				"pki-authority: {id: first, common-name: first, key-algorithm: ECDSA, ecdsa-curve: P-256, validity: 1h}",
				"pki-authority: {id: second, common-name: second, key-algorithm: ECDSA, ecdsa-curve: P-256, validity: 1h}",
			},
			wantErr: "defined in multiple files",
		},
		{name: "distribution section is not a mapping", contents: []string{"pki-distribution: invalid"}, wantErr: "parse pki-distribution section"},
		{name: "invalid distribution URL", contents: []string{"pki-distribution: {crl-url: http://pki.example.test/crl}"}, wantErr: "invalid authority CRL URL"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			paths := make([]string, len(tc.contents))
			for i, content := range tc.contents {
				paths[i] = filepath.Join(t.TempDir(), "config.yml")
				require.NoError(t, os.WriteFile(paths[i], []byte(content), 0o600))
			}

			_, err := PKIAuthorityConfigFromConfigFiles(paths)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}
