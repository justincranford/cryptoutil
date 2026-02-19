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
