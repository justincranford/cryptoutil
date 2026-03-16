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
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
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
			opts:     &KeyGenOptions{Algorithm: cryptoutilSharedMagic.KeyTypeRSA, KeySize: cryptoutilSharedMagic.DefaultMetricsBatchSize},
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
			name:     cryptoutilSharedMagic.EdCurveEd25519,
			opts:     &KeyGenOptions{Algorithm: cryptoutilSharedMagic.EdCurveEd25519},
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
			opts:     &KeyGenOptions{Algorithm: cryptoutilSharedMagic.JoseAlgEdDSA},
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
			err := os.MkdirAll(dir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute)
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
