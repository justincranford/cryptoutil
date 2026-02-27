// Copyright (c) 2025 Justin Cranford

package cli

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	rsa "crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCLI_GenerateEndEntityCert(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cli := NewCLI(nil, nil)

	// Generate CA.
	caKey, err := cli.GenerateKey(ctx, &KeyGenOptions{Algorithm: "ECDSA", Curve: "P-256"}, nil)
	require.NoError(t, err)

	caCert, err := cli.GenerateSelfSignedCA(ctx, caKey, nil, nil)
	require.NoError(t, err)

	// Generate end-entity key.
	eeKey, err := cli.GenerateKey(ctx, &KeyGenOptions{Algorithm: "ECDSA", Curve: "P-256"}, nil)
	require.NoError(t, err)

	tests := []struct {
		name    string
		key     any
		caCert  *x509.Certificate
		caKey   any
		opts    *CertGenOptions
		wantErr bool
	}{
		{
			name:    "nil key",
			key:     nil,
			caCert:  caCert,
			caKey:   caKey,
			opts:    &CertGenOptions{Subject: pkix.Name{CommonName: "test"}},
			wantErr: true,
		},
		{
			name:    "nil CA cert",
			key:     eeKey,
			caCert:  nil,
			caKey:   caKey,
			opts:    &CertGenOptions{Subject: pkix.Name{CommonName: "test"}},
			wantErr: true,
		},
		{
			name:    "nil CA key",
			key:     eeKey,
			caCert:  caCert,
			caKey:   nil,
			opts:    &CertGenOptions{Subject: pkix.Name{CommonName: "test"}},
			wantErr: true,
		},
		{
			name:    "nil options",
			key:     eeKey,
			caCert:  caCert,
			caKey:   caKey,
			opts:    nil,
			wantErr: true,
		},
		{
			name:   "valid TLS server cert",
			key:    eeKey,
			caCert: caCert,
			caKey:  caKey,
			opts: &CertGenOptions{
				Subject:      pkix.Name{CommonName: "test.example.com"},
				DNSNames:     []string{"test.example.com"},
				ValidityDays: cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year,
				ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cert, err := cli.GenerateEndEntityCert(ctx, tc.key, tc.caCert, tc.caKey, tc.opts, nil)
			if tc.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, cert)
			require.False(t, cert.IsCA)
			require.Equal(t, caCert.Subject.String(), cert.Issuer.String())
		})
	}
}

func TestCLI_ValidateCertificate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cli := NewCLI(nil, nil)

	// Generate CA and certificate.
	caKey, err := cli.GenerateKey(ctx, &KeyGenOptions{Algorithm: "ECDSA", Curve: "P-256"}, nil)
	require.NoError(t, err)

	caCert, err := cli.GenerateSelfSignedCA(ctx, caKey, nil, nil)
	require.NoError(t, err)

	eeKey, err := cli.GenerateKey(ctx, &KeyGenOptions{Algorithm: "ECDSA", Curve: "P-256"}, nil)
	require.NoError(t, err)

	eeCert, err := cli.GenerateEndEntityCert(ctx, eeKey, caCert, caKey, &CertGenOptions{
		Subject:  pkix.Name{CommonName: "test.example.com"},
		DNSNames: []string{"test.example.com"},
	}, nil)
	require.NoError(t, err)

	// Create root pool.
	roots := x509.NewCertPool()
	roots.AddCert(caCert)

	tests := []struct {
		name    string
		cert    *x509.Certificate
		roots   *x509.CertPool
		wantErr bool
	}{
		{
			name:    "nil certificate",
			cert:    nil,
			roots:   roots,
			wantErr: true,
		},
		{
			name:    "valid certificate",
			cert:    eeCert,
			roots:   roots,
			wantErr: false,
		},
		{
			name:    "unknown issuer",
			cert:    eeCert,
			roots:   x509.NewCertPool(),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := cli.ValidateCertificate(ctx, tc.cert, tc.roots, nil)
			if tc.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestGenerateSerialNumber(t *testing.T) {
	t.Parallel()

	// Generate multiple serial numbers and verify uniqueness.
	serials := make(map[string]bool)

	for i := 0; i < cryptoutilSharedMagic.JoseJAMaxMaterials; i++ {
		serial, err := generateSerialNumber()
		require.NoError(t, err)
		require.NotNil(t, serial)
		require.True(t, serial.Sign() >= 0, "serial number should be non-negative")

		serialStr := serial.String()
		require.False(t, serials[serialStr], "serial numbers should be unique")

		serials[serialStr] = true
	}
}

func TestSanitizeFilename(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with spaces", "with_spaces"},
		{"with/slash", "with_slash"},
		{"with\\backslash", "with_backslash"},
		{"special!@#$chars", "specialchars"},
		{"dots.and-dashes_ok", "dots.and-dashes_ok"},
		{"MixedCase123", "MixedCase123"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			result := sanitizeFilename(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestPublicKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cli := NewCLI(nil, nil)

	// Test RSA.
	rsaKey, err := cli.GenerateKey(ctx, &KeyGenOptions{Algorithm: cryptoutilSharedMagic.KeyTypeRSA, KeySize: cryptoutilSharedMagic.DefaultMetricsBatchSize}, nil)
	require.NoError(t, err)

	rsaPub := publicKey(rsaKey)
	_, ok := rsaPub.(*rsa.PublicKey)
	require.True(t, ok)

	// Test ECDSA.
	ecKey, err := cli.GenerateKey(ctx, &KeyGenOptions{Algorithm: "ECDSA", Curve: "P-256"}, nil)
	require.NoError(t, err)

	ecPub := publicKey(ecKey)
	_, ok = ecPub.(*ecdsa.PublicKey)
	require.True(t, ok)

	// Test Ed25519.
	edKey, err := cli.GenerateKey(ctx, &KeyGenOptions{Algorithm: cryptoutilSharedMagic.EdCurveEd25519}, nil)
	require.NoError(t, err)

	edPub := publicKey(edKey)
	_, ok = edPub.(ed25519.PublicKey)
	require.True(t, ok)

	// Test unsupported type.
	unknownPub := publicKey("not a key")
	require.Nil(t, unknownPub)
}
