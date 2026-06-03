// Copyright (c) 2025-2026 Justin Cranford.
//
//

package issuer

import (
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"encoding/base64"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	invalidKeyString               = "not_a_key"
	testSigningInputJWT            = "eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiIxMjM0In0"
	errExpectedRSAPriv             = "expected RSA private key"
	errExpectedECDSAPriv           = "expected ECDSA private key"
	errExpectedRSAPub              = "expected RSA public key"
	errExpectedECDSAPub            = "expected ECDSA public key"
	errUnsupportedVerify           = "unsupported verification algorithm"
	unsupportedAlgorithmName       = "unsupported_algorithm"
	invalidES256SignatureCase      = "ES256_invalid_signature_length"
	invalidSignatureThirdByte byte = 3
	testRSAKeyID                   = "test-rsa-kid"
	testRSAPublicKeyID             = "test-rsa-pub-kid"
	testECKeyID                    = "test-ec-kid"
	testECPublicKeyID              = "test-ec-pub-kid"
)

func TestSignJWT(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		algorithm   string
		keyGen      func() any
		wantErr     bool
		errContains string
	}{
		{
			name:      "RS256_success",
			algorithm: cryptoutilSharedMagic.AlgorithmRS256,
			keyGen: func() any {
				key, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)
				require.NoError(t, err)

				return key
			},
			wantErr: false,
		},
		{
			name:      "ES256_success",
			algorithm: cryptoutilSharedMagic.AlgorithmES256,
			keyGen: func() any {
				key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
				require.NoError(t, err)

				return key
			},
			wantErr: false,
		},
		{
			name:        "RS256_wrong_key_type",
			algorithm:   cryptoutilSharedMagic.AlgorithmRS256,
			keyGen:      func() any { return invalidKeyString },
			wantErr:     true,
			errContains: errExpectedRSAPriv,
		},
		{
			name:        "ES256_wrong_key_type",
			algorithm:   cryptoutilSharedMagic.AlgorithmES256,
			keyGen:      func() any { return invalidKeyString },
			wantErr:     true,
			errContains: errExpectedECDSAPriv,
		},
		{
			name:        unsupportedAlgorithmName,
			algorithm:   cryptoutilSharedMagic.JoseAlgHS256,
			keyGen:      func() any { return []byte("secret") },
			wantErr:     true,
			errContains: "unsupported signing algorithm",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			signature, err := signJWT(testSigningInputJWT, tc.algorithm, tc.keyGen())

			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)
				require.Empty(t, signature)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, signature)

				// Verify signature is valid base64url.
				_, decErr := base64.RawURLEncoding.DecodeString(signature)
				require.NoError(t, decErr)
			}
		})
	}
}

func TestVerifyJWTSignature(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		algorithm   string
		setupKeys   func(t *testing.T) (privateKey any, publicKey any)
		wantErr     bool
		errContains string
	}{
		{
			name:      "RS256_valid_signature",
			algorithm: cryptoutilSharedMagic.AlgorithmRS256,
			setupKeys: func(_ *testing.T) (any, any) {
				key, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)
				require.NoError(t, err)

				return key, &key.PublicKey
			},
			wantErr: false,
		},
		{
			name:      "ES256_valid_signature",
			algorithm: cryptoutilSharedMagic.AlgorithmES256,
			setupKeys: func(_ *testing.T) (any, any) {
				key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
				require.NoError(t, err)

				return key, &key.PublicKey
			},
			wantErr: false,
		},
		{
			name:      "RS256_wrong_public_key_type",
			algorithm: cryptoutilSharedMagic.AlgorithmRS256,
			setupKeys: func(_ *testing.T) (any, any) {
				key, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)
				require.NoError(t, err)

				return key, invalidKeyString
			},
			wantErr:     true,
			errContains: errExpectedRSAPub,
		},
		{
			name:      "ES256_wrong_public_key_type",
			algorithm: cryptoutilSharedMagic.AlgorithmES256,
			setupKeys: func(_ *testing.T) (any, any) {
				key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
				require.NoError(t, err)

				return key, invalidKeyString
			},
			wantErr:     true,
			errContains: errExpectedECDSAPub,
		},
		{
			name:      invalidES256SignatureCase,
			algorithm: cryptoutilSharedMagic.AlgorithmES256,
			setupKeys: func(_ *testing.T) (any, any) {
				key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
				require.NoError(t, err)

				return key, &key.PublicKey
			},
			wantErr:     true,
			errContains: "invalid ECDSA signature length",
		},
		{
			name:        unsupportedAlgorithmName,
			algorithm:   cryptoutilSharedMagic.JoseAlgHS256,
			setupKeys:   func(_ *testing.T) (any, any) { return []byte("secret"), []byte("secret") },
			wantErr:     true,
			errContains: errUnsupportedVerify,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			privateKey, publicKey := tc.setupKeys(t)

			// Sign the input.
			var signatureBytes []byte

			if tc.name == invalidES256SignatureCase {
				// Use invalid signature for this test case.
				signatureBytes = []byte{1, 2, invalidSignatureThirdByte} // Too short.
			} else {
				signatureStr, err := signJWT(testSigningInputJWT, tc.algorithm, privateKey)
				if err != nil && !tc.wantErr {
					require.NoError(t, err)
				}

				if signatureStr != "" {
					var decErr error

					signatureBytes, decErr = base64.RawURLEncoding.DecodeString(signatureStr)
					require.NoError(t, decErr)
				}
			}

			// Verify the signature.
			err := verifyJWTSignature(testSigningInputJWT, signatureBytes, tc.algorithm, publicKey)

			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConvertToJWK(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		key      *SigningKey
		expected map[string]any
	}{
		{
			name: "RSA_private_key",
			key: &SigningKey{
				KeyID:     testRSAKeyID,
				Algorithm: cryptoutilSharedMagic.AlgorithmRS256,
				Key: func() any {
					key, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)
					require.NoError(t, err)

					return key
				}(),
			},
			expected: map[string]any{
				jwkFieldKeyID:     testRSAKeyID,
				jwkFieldUse:       cryptoutilSharedMagic.JoseKeyUseSig,
				jwkFieldAlgorithm: cryptoutilSharedMagic.AlgorithmRS256,
				jwkFieldKeyType:   cryptoutilSharedMagic.KeyTypeRSA,
			},
		},
		{
			name: "RSA_public_key",
			key: &SigningKey{
				KeyID:     testRSAPublicKeyID,
				Algorithm: cryptoutilSharedMagic.AlgorithmRS256,
				Key: func() any {
					key, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)
					require.NoError(t, err)

					return &key.PublicKey
				}(),
			},
			expected: map[string]any{
				jwkFieldKeyID:     testRSAPublicKeyID,
				jwkFieldUse:       cryptoutilSharedMagic.JoseKeyUseSig,
				jwkFieldAlgorithm: cryptoutilSharedMagic.AlgorithmRS256,
				jwkFieldKeyType:   cryptoutilSharedMagic.KeyTypeRSA,
			},
		},
		{
			name: "ECDSA_private_key",
			key: &SigningKey{
				KeyID:     testECKeyID,
				Algorithm: cryptoutilSharedMagic.AlgorithmES256,
				Key: func() any {
					key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
					require.NoError(t, err)

					return key
				}(),
			},
			expected: map[string]any{
				jwkFieldKeyID:     testECKeyID,
				jwkFieldUse:       cryptoutilSharedMagic.JoseKeyUseSig,
				jwkFieldAlgorithm: cryptoutilSharedMagic.AlgorithmES256,
				jwkFieldKeyType:   cryptoutilSharedMagic.KeyTypeEC,
				jwkFieldCurve:     cryptoutilSharedMagic.ECCurveP256,
			},
		},
		{
			name: "ECDSA_public_key",
			key: &SigningKey{
				KeyID:     testECPublicKeyID,
				Algorithm: cryptoutilSharedMagic.AlgorithmES256,
				Key: func() any {
					key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
					require.NoError(t, err)

					return &key.PublicKey
				}(),
			},
			expected: map[string]any{
				jwkFieldKeyID:     testECPublicKeyID,
				jwkFieldUse:       cryptoutilSharedMagic.JoseKeyUseSig,
				jwkFieldAlgorithm: cryptoutilSharedMagic.AlgorithmES256,
				jwkFieldKeyType:   cryptoutilSharedMagic.KeyTypeEC,
				jwkFieldCurve:     cryptoutilSharedMagic.ECCurveP256,
			},
		},
		{
			name: "HMAC_key_returns_nil",
			key: &SigningKey{
				KeyID:     "test-hmac-kid",
				Algorithm: cryptoutilSharedMagic.JoseAlgHS256,
				Key:       []byte("secret"),
			},
			expected: nil,
		},
		{
			name:     "nil_key",
			key:      nil,
			expected: nil,
		},
		{
			name: "unsupported_key_type",
			key: &SigningKey{
				KeyID:     "test-unsupported-kid",
				Algorithm: cryptoutilSharedMagic.UNKNOWN,
				Key:       "not_a_crypto_key",
			},
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			jwk := convertToJWK(tc.key)

			if tc.expected == nil {
				require.Nil(t, jwk)

				return
			}

			require.NotNil(t, jwk)
			require.Equal(t, tc.expected[jwkFieldKeyID], jwk[jwkFieldKeyID])
			require.Equal(t, tc.expected[jwkFieldUse], jwk[jwkFieldUse])
			require.Equal(t, tc.expected[jwkFieldAlgorithm], jwk[jwkFieldAlgorithm])
			require.Equal(t, tc.expected[jwkFieldKeyType], jwk[jwkFieldKeyType])

			if tc.expected[jwkFieldCurve] != nil {
				require.Equal(t, tc.expected[jwkFieldCurve], jwk[jwkFieldCurve])
			}

			// Verify RSA/ECDSA specific fields exist.
			if tc.expected[jwkFieldKeyType] == cryptoutilSharedMagic.KeyTypeRSA {
				require.NotNil(t, jwk["n"])
				require.NotNil(t, jwk["e"])
			}

			if tc.expected[jwkFieldKeyType] == cryptoutilSharedMagic.KeyTypeEC {
				require.NotNil(t, jwk["x"])
				require.NotNil(t, jwk["y"])
			}
		})
	}
}

func TestBase64URLEncode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "simple_bytes",
			input:    []byte{1, 2, 3, 4},
			expected: "AQIDBA",
		},
		{
			name:     "empty_bytes",
			input:    []byte{},
			expected: "",
		},
		{
			name:     "large_number",
			input:    big.NewInt(65537).Bytes(),
			expected: "AQAB",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := base64URLEncode(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}
