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
	"math/big"
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// TestPEMEncodes tests encoding multiple certificates to PEM format.
func TestPEMEncodes(t *testing.T) {
	t.Parallel()

	// Generate test certificates.
	privateKey1, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	privateKey2, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "Test Certificate",
		},
	}

	cert1DER, err := x509.CreateCertificate(crand.Reader, template, template, &privateKey1.PublicKey, privateKey1)
	require.NoError(t, err)

	cert1, err := x509.ParseCertificate(cert1DER)
	require.NoError(t, err)

	template.SerialNumber = big.NewInt(2)
	cert2DER, err := x509.CreateCertificate(crand.Reader, template, template, &privateKey2.PublicKey, privateKey2)
	require.NoError(t, err)

	cert2, err := x509.ParseCertificate(cert2DER)
	require.NoError(t, err)

	// Test encoding multiple certificates.
	pemBytesList, err := PEMEncodes([]*x509.Certificate{cert1, cert2})
	require.NoError(t, err)
	require.Len(t, pemBytesList, 2)

	// Verify each PEM can be decoded.
	decoded1, err := PEMDecode(pemBytesList[0])
	require.NoError(t, err)
	require.IsType(t, &x509.Certificate{}, decoded1)

	decoded2, err := PEMDecode(pemBytesList[1])
	require.NoError(t, err)
	require.IsType(t, &x509.Certificate{}, decoded2)
}

// TestPEMEncodes_UnsupportedType tests PEMEncodes with unsupported types.
func TestPEMEncodes_UnsupportedType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       any
		expectedErr string
	}{
		{
			name:        "string type",
			input:       "invalid",
			expectedErr: "unsupported type: string",
		},
		{
			name:        "int type",
			input:       123,
			expectedErr: "unsupported type: int",
		},
		{
			name:        "nil",
			input:       nil,
			expectedErr: "unsupported type: <nil>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := PEMEncodes(tt.input)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

// TestDEREncodes tests encoding multiple certificates to DER format.
func TestDEREncodes(t *testing.T) {
	t.Parallel()

	// Generate test certificates.
	privateKey1, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	privateKey2, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "Test Certificate",
		},
	}

	cert1DER, err := x509.CreateCertificate(crand.Reader, template, template, &privateKey1.PublicKey, privateKey1)
	require.NoError(t, err)

	cert1, err := x509.ParseCertificate(cert1DER)
	require.NoError(t, err)

	template.SerialNumber = big.NewInt(2)
	cert2DER, err := x509.CreateCertificate(crand.Reader, template, template, &privateKey2.PublicKey, privateKey2)
	require.NoError(t, err)

	cert2, err := x509.ParseCertificate(cert2DER)
	require.NoError(t, err)

	// Test encoding multiple certificates.
	derBytesList, err := DEREncodes([]*x509.Certificate{cert1, cert2})
	require.NoError(t, err)
	require.Len(t, derBytesList, 2)

	// Verify each DER can be decoded.
	decoded1, err := DERDecode(derBytesList[0], cryptoutilSharedMagic.StringPEMTypeCertificate)
	require.NoError(t, err)
	require.IsType(t, &x509.Certificate{}, decoded1)

	decoded2, err := DERDecode(derBytesList[1], cryptoutilSharedMagic.StringPEMTypeCertificate)
	require.NoError(t, err)
	require.IsType(t, &x509.Certificate{}, decoded2)
}

// TestDEREncodes_UnsupportedType tests DEREncodes with unsupported types.
func TestDEREncodes_UnsupportedType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       any
		expectedErr string
	}{
		{
			name:        "string type",
			input:       "invalid",
			expectedErr: "unsupported type: string",
		},
		{
			name:        "int type",
			input:       123,
			expectedErr: "unsupported type: int",
		},
		{
			name:        "nil",
			input:       nil,
			expectedErr: "unsupported type: <nil>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := DEREncodes(tt.input)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

// TestDEREncode_AllKeyTypes tests DER encoding for all supported key types.
func TestDEREncode_AllKeyTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		keyGen          func() (any, error) //nolint:wrapcheck // Test helpers don't need to wrap crypto errors
		expectedPEMType string
	}{
		{
			name: "RSA private key",
			keyGen: func() (any, error) {
				return rsa.GenerateKey(crand.Reader, 2048)
			},
			expectedPEMType: cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey,
		},
		{
			name: "RSA public key",
			keyGen: func() (any, error) {
				priv, err := rsa.GenerateKey(crand.Reader, 2048)
				if err != nil {
					return nil, err
				}

				return &priv.PublicKey, nil
			},
			expectedPEMType: cryptoutilSharedMagic.StringPEMTypePKIXPublicKey,
		},
		{
			name: "ECDSA private key",
			keyGen: func() (any, error) {
				return ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
			},
			expectedPEMType: cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey,
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
			expectedPEMType: cryptoutilSharedMagic.StringPEMTypePKIXPublicKey,
		},
		{
			name: "EdDSA private key",
			keyGen: func() (any, error) {
				_, priv, err := ed25519.GenerateKey(crand.Reader)

				return priv, err
			},
			expectedPEMType: cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey,
		},
		{
			name: "EdDSA public key",
			keyGen: func() (any, error) {
				pub, _, err := ed25519.GenerateKey(crand.Reader)

				return pub, err
			},
			expectedPEMType: cryptoutilSharedMagic.StringPEMTypePKIXPublicKey,
		},
		{
			name: "byte slice (secret key)",
			keyGen: func() (any, error) {
				secretKey := make([]byte, 32)
				_, err := crand.Read(secretKey)

				return secretKey, err
			},
			expectedPEMType: cryptoutilSharedMagic.StringPEMTypeSecretKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			key, err := tt.keyGen()
			require.NoError(t, err)

			derBytes, pemType, err := DEREncode(key)
			require.NoError(t, err)
			require.NotEmpty(t, derBytes)
			require.Equal(t, tt.expectedPEMType, pemType)
		})
	}
}

// TestDEREncode_UnsupportedType tests DEREncode with unsupported types.
func TestDEREncode_UnsupportedType(t *testing.T) {
	t.Parallel()

	_, _, err := DEREncode("invalid type")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not supported")
}

// TestDERDecode_AllKeyTypes tests DER decoding for all supported key types.
func TestDERDecode_AllKeyTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		keyGen     func() (any, error)
		pemType    string
		verifyType func(*testing.T, any)
	}{
		{
			name: "RSA private key (PKCS8)",
			keyGen: func() (any, error) {
				return rsa.GenerateKey(crand.Reader, 2048)
			},
			pemType: cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey,
			verifyType: func(t *testing.T, key any) {
				require.IsType(t, &rsa.PrivateKey{}, key)
			},
		},
		{
			name: "RSA public key (PKIX)",
			keyGen: func() (any, error) {
				priv, err := rsa.GenerateKey(crand.Reader, 2048)
				if err != nil {
					return nil, err
				}

				return &priv.PublicKey, nil
			},
			pemType: cryptoutilSharedMagic.StringPEMTypePKIXPublicKey,
			verifyType: func(t *testing.T, key any) {
				require.IsType(t, &rsa.PublicKey{}, key)
			},
		},
		{
			name: "ECDSA private key (PKCS8)",
			keyGen: func() (any, error) {
				return ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
			},
			pemType: cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey,
			verifyType: func(t *testing.T, key any) {
				require.IsType(t, &ecdsa.PrivateKey{}, key)
			},
		},
		{
			name: "ECDSA public key (PKIX)",
			keyGen: func() (any, error) {
				priv, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
				if err != nil {
					return nil, err
				}

				return &priv.PublicKey, nil
			},
			pemType: cryptoutilSharedMagic.StringPEMTypePKIXPublicKey,
			verifyType: func(t *testing.T, key any) {
				require.IsType(t, &ecdsa.PublicKey{}, key)
			},
		},
		{
			name: "EdDSA private key (PKCS8)",
			keyGen: func() (any, error) {
				_, priv, err := ed25519.GenerateKey(crand.Reader)

				return priv, err
			},
			pemType: cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey,
			verifyType: func(t *testing.T, key any) {
				require.IsType(t, ed25519.PrivateKey{}, key)
			},
		},
		{
			name: "EdDSA public key (PKIX)",
			keyGen: func() (any, error) {
				pub, _, err := ed25519.GenerateKey(crand.Reader)

				return pub, err
			},
			pemType: cryptoutilSharedMagic.StringPEMTypePKIXPublicKey,
			verifyType: func(t *testing.T, key any) {
				require.IsType(t, ed25519.PublicKey{}, key)
			},
		},
		{
			name: "Secret key (byte slice)",
			keyGen: func() (any, error) {
				secretKey := make([]byte, 32)
				_, err := crand.Read(secretKey)

				return secretKey, err
			},
			pemType: cryptoutilSharedMagic.StringPEMTypeSecretKey,
			verifyType: func(t *testing.T, key any) {
				require.IsType(t, []byte{}, key)
			},
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

			// Decode and verify type.
			decodedKey, err := DERDecode(derBytes, tt.pemType)
			require.NoError(t, err)
			tt.verifyType(t, decodedKey)
		})
	}
}

// TestDERDecode_UnsupportedType tests DERDecode with unsupported PEM type.
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
