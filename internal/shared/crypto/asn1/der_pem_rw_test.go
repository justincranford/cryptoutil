// Copyright (c) 2025 Justin Cranford

//nolint:wrapcheck,thelper,errcheck // Test code doesn't need to wrap crypto errors or use t.Helper()
package asn1

import (
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestPEMWrite(t *testing.T) {
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
			name: "ECDSA public key",
			keyGen: func() (any, error) {
				priv, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
				if err != nil {
					return nil, err
				}

				return &priv.PublicKey, nil
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

			// Create temporary file path with subdirectory.
			tmpDir := t.TempDir()
			filename := filepath.Join(tmpDir, "subdir", "test_key.pem")

			// Write key to file (should create subdirectory).
			err = PEMWrite(key, filename)
			require.NoError(t, err)

			// Verify file exists.
			info, err := os.Stat(filename)
			require.NoError(t, err)
			require.NotZero(t, info.Size())

			// Verify directory was created.
			dirInfo, err := os.Stat(filepath.Dir(filename))
			require.NoError(t, err)
			require.True(t, dirInfo.IsDir())
		})
	}
}

// TestPEMWrite_UnsupportedType tests PEMWrite with unsupported types.
func TestPEMWrite_UnsupportedType(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "invalid.pem")

	err := PEMWrite("invalid type", filename)
	require.Error(t, err)
	require.Contains(t, err.Error(), "encode failed")
}

// TestDERWrite tests writing keys to DER files.
func TestDERWrite(t *testing.T) {
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
			name: "ECDSA public key",
			keyGen: func() (any, error) {
				priv, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
				if err != nil {
					return nil, err
				}

				return &priv.PublicKey, nil
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

			// Create temporary file path with subdirectory.
			tmpDir := t.TempDir()
			filename := filepath.Join(tmpDir, "subdir", "test_key.der")

			// Write key to file (should create subdirectory).
			err = DERWrite(key, filename)
			require.NoError(t, err)

			// Verify file exists.
			info, err := os.Stat(filename)
			require.NoError(t, err)
			require.NotZero(t, info.Size())

			// Verify directory was created.
			dirInfo, err := os.Stat(filepath.Dir(filename))
			require.NoError(t, err)
			require.True(t, dirInfo.IsDir())
		})
	}
}

// TestDERWrite_UnsupportedType tests DERWrite with unsupported types.
func TestDERWrite_UnsupportedType(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "invalid.der")

	err := DERWrite("invalid type", filename)
	require.Error(t, err)
	require.Contains(t, err.Error(), "encode failed")
}

// TestRoundTrip_PEMFileOperations tests complete PEM file write/read cycle.
func TestRoundTrip_PEMFileOperations(t *testing.T) {
	t.Parallel()

	// Generate RSA key.
	originalKey, err := rsa.GenerateKey(crand.Reader, 2048)
	require.NoError(t, err)

	// Create temporary file.
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "roundtrip_key.pem")

	// Write key to file.
	err = PEMWrite(originalKey, filename)
	require.NoError(t, err)

	// Read key from file.
	readKey, err := PEMRead(filename)
	require.NoError(t, err)

	// Verify key type.
	require.IsType(t, &rsa.PrivateKey{}, readKey)
	readKeyTyped := readKey.(*rsa.PrivateKey)

	// Compare key material (not struct pointers).
	// Note: Cannot use require.Equal() because fips field is a pointer
	// that differs between instances even with identical key material.
	require.Equal(t, originalKey.N, readKeyTyped.N, "modulus mismatch")
	require.Equal(t, originalKey.E, readKeyTyped.E, "public exponent mismatch")
	require.Equal(t, originalKey.D, readKeyTyped.D, "private exponent mismatch")
	require.Equal(t, len(originalKey.Primes), len(readKeyTyped.Primes), "prime count mismatch")

	for i := range originalKey.Primes {
		require.Equal(t, originalKey.Primes[i], readKeyTyped.Primes[i], "prime[%d] mismatch", i)
	}
}

// TestRoundTrip_DERFileOperations tests complete DER file write/read cycle.
func TestRoundTrip_DERFileOperations(t *testing.T) {
	t.Parallel()

	// Generate ECDSA key.
	originalKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	// Create temporary file.
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "roundtrip_key.der")

	// Write key to file.
	err = DERWrite(originalKey, filename)
	require.NoError(t, err)

	// Read key from file.
	readKey, derType, err := DERRead(filename)
	require.NoError(t, err)
	require.Equal(t, cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey, derType)

	// Verify key type.
	require.IsType(t, &ecdsa.PrivateKey{}, readKey)
}

// TestCertificateRequest_EncodeDecode tests CSR encoding and decoding.
func TestCertificateRequest_EncodeDecode(t *testing.T) {
	t.Parallel()

	// Generate key for CSR.
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	// Create CSR template.
	csrTemplate := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: "Test CSR",
		},
	}

	// Create CSR.
	csrDER, err := x509.CreateCertificateRequest(crand.Reader, csrTemplate, privateKey)
	require.NoError(t, err)

	csr, err := x509.ParseCertificateRequest(csrDER)
	require.NoError(t, err)

	// Test PEM encoding/decoding.
	pemBytes, err := PEMEncode(csr)
	require.NoError(t, err)

	decodedCSR, err := PEMDecode(pemBytes)
	require.NoError(t, err)
	require.IsType(t, &x509.CertificateRequest{}, decodedCSR)

	// Test DER encoding/decoding.
	derBytes, pemType, err := DEREncode(csr)
	require.NoError(t, err)
	require.Equal(t, cryptoutilSharedMagic.StringPEMTypeCSR, pemType)

	decodedCSR2, err := DERDecode(derBytes, cryptoutilSharedMagic.StringPEMTypeCSR)
	require.NoError(t, err)
	require.IsType(t, &x509.CertificateRequest{}, decodedCSR2)
}
