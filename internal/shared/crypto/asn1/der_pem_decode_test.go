// Copyright (c) 2025 Justin Cranford

//nolint:wrapcheck,thelper,errcheck // Test code doesn't need to wrap crypto errors or use t.Helper()
package asn1

import (
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestDERDecode_UnsupportedType(t *testing.T) {
	t.Parallel()

	_, err := DERDecode([]byte("invalid"), "UNSUPPORTED TYPE")
	require.Error(t, err)
	require.Contains(t, err.Error(), "type not supported")
}

// TestDERDecode_InvalidData tests DERDecode with invalid data.
func TestDERDecode_InvalidData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		data    []byte
		pemType string
	}{
		{
			name:    "invalid PKCS8 data",
			data:    []byte("not valid DER"),
			pemType: cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey,
		},
		{
			name:    "invalid PKIX data",
			data:    []byte("not valid DER"),
			pemType: cryptoutilSharedMagic.StringPEMTypePKIXPublicKey,
		},
		{
			name:    "invalid certificate data",
			data:    []byte("not valid DER"),
			pemType: cryptoutilSharedMagic.StringPEMTypeCertificate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := DERDecode(tt.data, tt.pemType)
			require.Error(t, err)
			require.Contains(t, err.Error(), "decode failed")
		})
	}
}

// TestDERDecodes tests automatic DER type detection.
func TestDERDecodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		keyGen       func() (any, error)
		expectedType string
	}{
		{
			name: "RSA private key",
			keyGen: func() (any, error) {
				return rsa.GenerateKey(crand.Reader, 2048)
			},
			expectedType: cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey,
		},
		{
			name: "ECDSA public key",
			keyGen: func() (any, error) {
				priv, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
				if err != nil {
					return nil, err
				}

				return &priv.PublicKey, nil
			},
			expectedType: cryptoutilSharedMagic.StringPEMTypePKIXPublicKey,
		},
		{
			name: "Secret key",
			keyGen: func() (any, error) {
				secretKey := make([]byte, 32)
				_, err := crand.Read(secretKey)

				return secretKey, err
			},
			expectedType: cryptoutilSharedMagic.StringPEMTypeSecretKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Generate and encode key.
			originalKey, err := tt.keyGen()
			require.NoError(t, err)

			derBytes, _, err := DEREncode(originalKey)
			require.NoError(t, err)

			// Decode with automatic type detection.
			decodedKey, detectedType, err := DERDecodes(derBytes)
			require.NoError(t, err)
			require.NotNil(t, decodedKey)
			require.Equal(t, tt.expectedType, detectedType)
		})
	}
}

// TestPEMDecode_InvalidData tests PEMDecode with invalid data.
func TestPEMDecode_InvalidData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "not PEM data",
			data: []byte("not PEM encoded"),
		},
		{
			name: "empty data",
			data: []byte{},
		},
		{
			name: "nil data",
			data: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := PEMDecode(tt.data)
			require.Error(t, err)
			require.Contains(t, err.Error(), "parse PEM failed")
		})
	}
}

// TestPEMRead tests reading PEM-encoded keys from files.
func TestPEMRead(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		keyGen func() (any, error)
	}{
		{
			name: "RSA private key",
			keyGen: func() (any, error) {
				return rsa.GenerateKey(crand.Reader, 2048)
			},
		},
		{
			name: "ECDSA private key",
			keyGen: func() (any, error) {
				return ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
			},
		},
		{
			name: "EdDSA private key",
			keyGen: func() (any, error) {
				_, priv, err := ed25519.GenerateKey(crand.Reader)

				return priv, err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Generate key.
			key, err := tt.keyGen()
			require.NoError(t, err)

			// Create temporary file.
			tmpDir := t.TempDir()
			filename := filepath.Join(tmpDir, "test_key.pem")

			// Write key to file.
			err = PEMWrite(key, filename)
			require.NoError(t, err)

			// Read key from file.
			readKey, err := PEMRead(filename)
			require.NoError(t, err)
			require.NotNil(t, readKey)
		})
	}
}

// TestPEMRead_NonExistentFile tests PEMRead with non-existent file.
func TestPEMRead_NonExistentFile(t *testing.T) {
	t.Parallel()

	_, err := PEMRead("/nonexistent/path/to/file.pem")
	require.Error(t, err)
	require.Contains(t, err.Error(), "read failed")
}

// TestDERRead tests reading DER-encoded keys from files.
func TestDERRead(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		keyGen       func() (any, error)
		expectedType string
	}{
		{
			name: "RSA private key",
			keyGen: func() (any, error) {
				return rsa.GenerateKey(crand.Reader, 2048)
			},
			expectedType: cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey,
		},
		{
			name: "ECDSA public key",
			keyGen: func() (any, error) {
				priv, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
				if err != nil {
					return nil, err
				}

				return &priv.PublicKey, nil
			},
			expectedType: cryptoutilSharedMagic.StringPEMTypePKIXPublicKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Generate key.
			key, err := tt.keyGen()
			require.NoError(t, err)

			// Create temporary file.
			tmpDir := t.TempDir()
			filename := filepath.Join(tmpDir, "test_key.der")

			// Write key to file.
			err = DERWrite(key, filename)
			require.NoError(t, err)

			// Read key from file.
			readKey, derType, err := DERRead(filename)
			require.NoError(t, err)
			require.NotNil(t, readKey)
			require.Equal(t, tt.expectedType, derType)
		})
	}
}

// TestDERRead_NonExistentFile tests DERRead with non-existent file.
func TestDERRead_NonExistentFile(t *testing.T) {
	t.Parallel()

	_, _, err := DERRead("/nonexistent/path/to/file.der")
	require.Error(t, err)
	require.Contains(t, err.Error(), "read failed")
}

// TestPEMWrite tests writing keys to PEM files.
