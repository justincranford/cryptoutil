// Copyright (c) 2025 Justin Cranford

package security

import (
	"context"
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"crypto/x509"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config *Config
	}{
		{
			name:   "with nil config uses defaults",
			config: nil,
		},
		{
			name:   "with custom config",
			config: &Config{MinRSAKeySize: cryptoutilSharedMagic.RSA4096KeySize},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			validator := NewValidator(tc.config)
			require.NotNil(t, validator)
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()

	require.Equal(t, cryptoutilSharedMagic.DefaultMetricsBatchSize, config.MinRSAKeySize)
	require.Equal(t, cryptoutilSharedMagic.MaxUnsealSharedSecrets, config.MinECKeySize)
	require.Equal(t, cryptoutilSharedMagic.TLSMaxValidityEndEntityDays, config.MaxCertValidityDays)
	require.True(t, config.RequireKeyUsage)
	require.True(t, config.RequireBasicConstraints)
	require.True(t, config.RequireSAN)
	require.True(t, config.DisallowWeakAlgorithms)
	require.True(t, config.EnforcePathLengthConstraints)
	require.True(t, config.AuditLoggingEnabled)
	require.NotEmpty(t, config.AllowedSignatureAlgorithms)
}

func TestValidator_ValidateCertificate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	validator := NewValidator(nil)

	// Generate test key.
	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	tests := []struct {
		name      string
		certFunc  func() *x509.Certificate
		wantValid bool
		wantErr   bool
	}{
		{
			name: "valid certificate",
			certFunc: func() *x509.Certificate {
				return createTestCert(t, key, false, time.Now().UTC(), time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year*cryptoutilSharedMagic.HoursPerDay*time.Hour))
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "expired certificate",
			certFunc: func() *x509.Certificate {
				return createTestCert(t, key, false, time.Now().UTC().Add(-730*cryptoutilSharedMagic.HoursPerDay*time.Hour), time.Now().UTC().Add(-cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year*cryptoutilSharedMagic.HoursPerDay*time.Hour))
			},
			wantValid: false,
			wantErr:   false,
		},
		{
			name: "certificate with excessive validity",
			certFunc: func() *x509.Certificate {
				return createTestCert(t, key, false, time.Now().UTC(), time.Now().UTC().Add(cryptoutilSharedMagic.TestDefaultRateLimitServiceIP*cryptoutilSharedMagic.HoursPerDay*time.Hour))
			},
			wantValid: false,
			wantErr:   false,
		},
		{
			name:      "nil certificate",
			certFunc:  func() *x509.Certificate { return nil },
			wantValid: false,
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cert := tc.certFunc()
			result, err := validator.ValidateCertificate(ctx, cert)

			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, tc.wantValid, result.Valid)
			}
		})
	}
}

func TestValidator_ValidateCertificate_KeySizes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name      string
		keyFunc   func() any
		config    *Config
		wantValid bool
	}{
		{
			name: "RSA 2048 with 2048 minimum",
			keyFunc: func() any {
				key, _ := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)

				return key
			},
			config:    &Config{MinRSAKeySize: cryptoutilSharedMagic.DefaultMetricsBatchSize, AllowedSignatureAlgorithms: []x509.SignatureAlgorithm{x509.SHA256WithRSA}, MaxCertValidityDays: cryptoutilSharedMagic.TLSMaxValidityEndEntityDays},
			wantValid: true,
		},
		{
			name: "RSA 2048 with 4096 minimum",
			keyFunc: func() any {
				key, _ := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)

				return key
			},
			config:    &Config{MinRSAKeySize: cryptoutilSharedMagic.RSA4096KeySize, AllowedSignatureAlgorithms: []x509.SignatureAlgorithm{x509.SHA256WithRSA}, MaxCertValidityDays: cryptoutilSharedMagic.TLSMaxValidityEndEntityDays},
			wantValid: false,
		},
		{
			name: "EC P-256 with 256 minimum",
			keyFunc: func() any {
				key, _ := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)

				return key
			},
			config:    &Config{MinECKeySize: cryptoutilSharedMagic.MaxUnsealSharedSecrets, AllowedSignatureAlgorithms: []x509.SignatureAlgorithm{x509.ECDSAWithSHA256}, MaxCertValidityDays: cryptoutilSharedMagic.TLSMaxValidityEndEntityDays},
			wantValid: true,
		},
		{
			name: "EC P-256 with 384 minimum",
			keyFunc: func() any {
				key, _ := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)

				return key
			},
			config:    &Config{MinECKeySize: cryptoutilSharedMagic.SymmetricKeySize384, AllowedSignatureAlgorithms: []x509.SignatureAlgorithm{x509.ECDSAWithSHA256}, MaxCertValidityDays: cryptoutilSharedMagic.TLSMaxValidityEndEntityDays},
			wantValid: false,
		},
		{
			name: "Ed25519 key",
			keyFunc: func() any {
				_, key, _ := ed25519.GenerateKey(crand.Reader)

				return key
			},
			config:    &Config{AllowedSignatureAlgorithms: []x509.SignatureAlgorithm{x509.PureEd25519}, MaxCertValidityDays: cryptoutilSharedMagic.TLSMaxValidityEndEntityDays},
			wantValid: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			validator := NewValidator(tc.config)
			key := tc.keyFunc()

			cert := createTestCertWithKey(t, key, false, time.Now().UTC(), time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year*cryptoutilSharedMagic.HoursPerDay*time.Hour))
			result, err := validator.ValidateCertificate(ctx, cert)

			require.NoError(t, err)
			require.Equal(t, tc.wantValid, result.Valid)
		})
	}
}

func TestValidator_ValidatePrivateKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	validator := NewValidator(nil)

	tests := []struct {
		name      string
		keyFunc   func() any
		wantValid bool
		wantErr   bool
	}{
		{
			name: "valid RSA 2048 key",
			keyFunc: func() any {
				key, _ := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)

				return key
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "valid EC P-256 key",
			keyFunc: func() any {
				key, _ := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)

				return key
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "valid Ed25519 key",
			keyFunc: func() any {
				_, key, _ := ed25519.GenerateKey(crand.Reader)

				return key
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name:      "nil key",
			keyFunc:   func() any { return nil },
			wantValid: false,
			wantErr:   true,
		},
		{
			name:      "unknown key type",
			keyFunc:   func() any { return "invalid key" },
			wantValid: true, // Unknown types generate warning but are valid.
			wantErr:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			key := tc.keyFunc()
			result, err := validator.ValidatePrivateKey(ctx, key)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantValid, result.Valid)
			}
		})
	}
}
