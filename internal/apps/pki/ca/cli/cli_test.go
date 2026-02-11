// Copyright (c) 2025 Justin Cranford

package cli

import (
	"bytes"
	"context"
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	rsa "crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewCLI(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		out    *bytes.Buffer
		errOut *bytes.Buffer
	}{
		{
			name:   "with buffers",
			out:    new(bytes.Buffer),
			errOut: new(bytes.Buffer),
		},
		{
			name:   "with nil out",
			out:    nil,
			errOut: new(bytes.Buffer),
		},
		{
			name:   "with nil errOut",
			out:    new(bytes.Buffer),
			errOut: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cli := NewCLI(tc.out, tc.errOut)
			require.NotNil(t, cli)
		})
	}
}

// Key type constants for tests.
const (
	keyTypeRSA     = "rsa"
	keyTypeECDSA   = "ecdsa"
	keyTypeEd25519 = "ed25519"
)

func TestCLI_GenerateKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cli := NewCLI(nil, nil)

	tests := []struct {
		name     string
		opts     *KeyGenOptions
		wantType string
		wantErr  bool
	}{
		{
			name:     "nil options defaults to ECDSA P-256",
			opts:     nil,
			wantType: keyTypeECDSA,
			wantErr:  false,
		},
		{
			name:     "RSA 2048",
			opts:     &KeyGenOptions{Algorithm: "RSA", KeySize: 2048},
			wantType: keyTypeRSA,
			wantErr:  false,
		},
		{
			name:     "RSA default size",
			opts:     &KeyGenOptions{Algorithm: "rsa"},
			wantType: keyTypeRSA,
			wantErr:  false,
		},
		{
			name:     "ECDSA P-256",
			opts:     &KeyGenOptions{Algorithm: "ECDSA", Curve: "P-256"},
			wantType: keyTypeECDSA,
			wantErr:  false,
		},
		{
			name:     "ECDSA P-384",
			opts:     &KeyGenOptions{Algorithm: "ec", Curve: "P-384"},
			wantType: keyTypeECDSA,
			wantErr:  false,
		},
		{
			name:     "ECDSA P-521",
			opts:     &KeyGenOptions{Algorithm: "EC", Curve: "P-521"},
			wantType: keyTypeECDSA,
			wantErr:  false,
		},
		{
			name:     "Ed25519",
			opts:     &KeyGenOptions{Algorithm: "Ed25519"},
			wantType: keyTypeEd25519,
			wantErr:  false,
		},
		{
			name:     "ed25519 lowercase",
			opts:     &KeyGenOptions{Algorithm: "ed25519"},
			wantType: keyTypeEd25519,
			wantErr:  false,
		},
		{
			name:     "EdDSA alias",
			opts:     &KeyGenOptions{Algorithm: "EdDSA"},
			wantType: keyTypeEd25519,
			wantErr:  false,
		},
		{
			name:    "unsupported algorithm",
			opts:    &KeyGenOptions{Algorithm: "unknown"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			key, err := cli.GenerateKey(ctx, tc.opts, nil)
			if tc.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, key)

			switch tc.wantType {
			case keyTypeRSA:
				_, ok := key.(*rsa.PrivateKey)
				require.True(t, ok)
			case keyTypeECDSA:
				_, ok := key.(*ecdsa.PrivateKey)
				require.True(t, ok)
			case keyTypeEd25519:
				_, ok := key.(ed25519.PrivateKey)
				require.True(t, ok)
			}
		})
	}
}

func TestCLI_GenerateKey_WriteToFile(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cli := NewCLI(nil, nil)
	tmpDir := t.TempDir()

	tests := []struct {
		name   string
		format string
		ext    string
	}{
		{"PEM format", "pem", ".pem"},
		{"DER format", "der", ".der"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := filepath.Join(tmpDir, tc.format)
			err := os.MkdirAll(dir, 0o755)
			require.NoError(t, err)

			cmdOpts := &CommandOptions{
				OutputDir:    dir,
				OutputFormat: tc.format,
			}

			key, err := cli.GenerateKey(ctx, &KeyGenOptions{Algorithm: "ECDSA", Curve: "P-256"}, cmdOpts)
			require.NoError(t, err)
			require.NotNil(t, key)

			// Verify file was created.
			keyFile := filepath.Join(dir, "key"+tc.ext)
			_, err = os.Stat(keyFile)
			require.NoError(t, err)
		})
	}
}

func TestCLI_GenerateSelfSignedCA(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cli := NewCLI(nil, nil)

	tests := []struct {
		name    string
		opts    *CertGenOptions
		wantErr bool
	}{
		{
			name:    "nil key",
			opts:    nil,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := cli.GenerateSelfSignedCA(ctx, nil, tc.opts, nil)
			if tc.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestCLI_GenerateSelfSignedCA_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cli := NewCLI(nil, nil)

	key, err := cli.GenerateKey(ctx, &KeyGenOptions{Algorithm: "ECDSA", Curve: "P-256"}, nil)
	require.NoError(t, err)

	tests := []struct {
		name string
		opts *CertGenOptions
	}{
		{
			name: "with default options",
			opts: nil,
		},
		{
			name: "with custom subject",
			opts: &CertGenOptions{
				Subject: pkix.Name{
					CommonName:   "Test Root CA",
					Organization: []string{"Test Org"},
					Country:      []string{"US"},
				},
				ValidityDays: 3650,
				IsCA:         true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cert, err := cli.GenerateSelfSignedCA(ctx, key, tc.opts, nil)
			require.NoError(t, err)
			require.NotNil(t, cert)
			require.True(t, cert.IsCA)
			require.Equal(t, cert.Issuer.String(), cert.Subject.String())
		})
	}
}

func TestCLI_GenerateSelfSignedCA_WriteToFile(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cli := NewCLI(nil, nil)
	tmpDir := t.TempDir()

	key, err := cli.GenerateKey(ctx, &KeyGenOptions{Algorithm: "ECDSA", Curve: "P-256"}, nil)
	require.NoError(t, err)

	cmdOpts := &CommandOptions{
		OutputDir:    tmpDir,
		OutputFormat: "pem",
	}

	cert, err := cli.GenerateSelfSignedCA(ctx, key, nil, cmdOpts)
	require.NoError(t, err)
	require.NotNil(t, cert)

	// Verify file was created.
	certFile := filepath.Join(tmpDir, "ca.pem")
	_, err = os.Stat(certFile)
	require.NoError(t, err)
}

func TestCLI_GenerateIntermediateCA(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cli := NewCLI(nil, nil)

	// Generate root CA.
	rootKey, err := cli.GenerateKey(ctx, &KeyGenOptions{Algorithm: "ECDSA", Curve: "P-256"}, nil)
	require.NoError(t, err)

	rootCert, err := cli.GenerateSelfSignedCA(ctx, rootKey, nil, nil)
	require.NoError(t, err)

	// Generate intermediate key.
	intermediateKey, err := cli.GenerateKey(ctx, &KeyGenOptions{Algorithm: "ECDSA", Curve: "P-256"}, nil)
	require.NoError(t, err)

	tests := []struct {
		name       string
		key        any
		parentCert *x509.Certificate
		parentKey  any
		wantErr    bool
	}{
		{
			name:       "nil key",
			key:        nil,
			parentCert: rootCert,
			parentKey:  rootKey,
			wantErr:    true,
		},
		{
			name:       "nil parent cert",
			key:        intermediateKey,
			parentCert: nil,
			parentKey:  rootKey,
			wantErr:    true,
		},
		{
			name:       "nil parent key",
			key:        intermediateKey,
			parentCert: rootCert,
			parentKey:  nil,
			wantErr:    true,
		},
		{
			name:       "valid intermediate",
			key:        intermediateKey,
			parentCert: rootCert,
			parentKey:  rootKey,
			wantErr:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cert, err := cli.GenerateIntermediateCA(ctx, tc.key, tc.parentCert, tc.parentKey, nil, nil)
			if tc.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, cert)
			require.True(t, cert.IsCA)
			require.Equal(t, rootCert.Subject.String(), cert.Issuer.String())
		})
	}
}

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
				ValidityDays: 365,
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

	for i := 0; i < 100; i++ {
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
	rsaKey, err := cli.GenerateKey(ctx, &KeyGenOptions{Algorithm: "RSA", KeySize: 2048}, nil)
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
	edKey, err := cli.GenerateKey(ctx, &KeyGenOptions{Algorithm: "Ed25519"}, nil)
	require.NoError(t, err)

	edPub := publicKey(edKey)
	_, ok = edPub.(ed25519.PublicKey)
	require.True(t, ok)

	// Test unsupported type.
	unknownPub := publicKey("not a key")
	require.Nil(t, unknownPub)
}
