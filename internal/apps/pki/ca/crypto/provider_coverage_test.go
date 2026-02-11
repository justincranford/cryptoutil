// Copyright (c) 2025 Justin Cranford

package crypto

import (
	"crypto"
	"crypto/ed25519"
	crand "crypto/rand"
	rsa "crypto/rsa"
	sha256 "crypto/sha256"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testEd25519Curve = "Ed25519"
	testRSAAlgorithm = "RSA"
)

// TestEdDSACurves tests Ed25519 and Ed448 curve generation.
func TestEdDSACurves(t *testing.T) {
	t.Parallel()

	provider := NewSoftwareProvider()

	tests := []struct {
		name        string
		curve       string
		wantErr     bool
		errContains string
	}{
		{
			name:    "Ed25519",
			curve:   testEd25519Curve,
			wantErr: false,
		},
		{
			name:    "Ed448",
			curve:   "Ed448",
			wantErr: false,
		},
		{
			name:        "Invalid-curve",
			curve:       "InvalidCurve",
			wantErr:     true,
			errContains: "unsupported Ed curve",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			kp, err := provider.generateEdDSAKeyPair(tt.curve)

			if tt.wantErr {
				require.Error(t, err)

				if tt.errContains != "" {
					require.Contains(t, err.Error(), tt.errContains)
				}

				return
			}

			require.NoError(t, err)
			require.NotNil(t, kp)
			require.Equal(t, KeyTypeEdDSA, kp.Type)

			if tt.curve == testEd25519Curve {
				_, ok := kp.PrivateKey.(ed25519.PrivateKey)
				require.True(t, ok, "expected Ed25519 private key")
			}
		})
	}
}

// TestVerifyEdDSAFailures tests EdDSA verification failure paths.
func TestVerifyEdDSAFailures(t *testing.T) {
	t.Parallel()

	provider := NewSoftwareProvider()

	// Generate Ed25519 key pair
	kp, err := provider.generateEdDSAKeyPair(testEd25519Curve)
	require.NoError(t, err)

	pub, ok := kp.PublicKey.(ed25519.PublicKey)
	require.True(t, ok)

	// Test verification with invalid signature
	digest := sha256.Sum256([]byte("test message"))
	invalidSignature := make([]byte, ed25519.SignatureSize)

	_, err = crand.Read(invalidSignature)
	require.NoError(t, err)

	err = provider.verifyEdDSA(pub, digest[:], invalidSignature)
	require.Error(t, err)
	require.Contains(t, err.Error(), "EdDSA signature verification failed")
}

// TestVerifyRSAFailures tests RSA verification failure paths.
func TestVerifyRSAFailures(t *testing.T) {
	t.Parallel()

	provider := NewSoftwareProvider()

	// Generate RSA key pair
	kp, err := provider.generateRSAKeyPair(2048)
	require.NoError(t, err)

	priv, ok := kp.PrivateKey.(*rsa.PrivateKey)
	require.True(t, ok)

	// Test verification with invalid signature
	digest := sha256.Sum256([]byte("test message"))
	invalidSignature := make([]byte, 256)

	_, err = crand.Read(invalidSignature)
	require.NoError(t, err)

	err = provider.verifyRSA(&priv.PublicKey, digest[:], invalidSignature, crypto.SHA256)
	require.Error(t, err)
	require.Contains(t, err.Error(), "RSA signature verification failed")

	// Test verification with zero hash function
	validSignature, err := rsa.SignPKCS1v15(crand.Reader, priv, crypto.SHA256, digest[:])
	require.NoError(t, err)

	err = provider.verifyRSA(&priv.PublicKey, digest[:], validSignature, crypto.Hash(0))
	require.Error(t, err)
	require.Contains(t, err.Error(), "hash function required")
}

// TestSignFailures tests Sign function failure paths.
func TestSignFailures(t *testing.T) {
	t.Parallel()

	provider := NewSoftwareProvider()

	// Test with non-Signer private key (raw []byte)
	fakeKey := []byte("not a real key")
	digest := sha256.Sum256([]byte("test message"))

	_, err := provider.Sign(fakeKey, digest[:], crypto.SHA256)
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not implement crypto.Signer")
}

// TestGenerateRSAKeyPairSizes tests different RSA key sizes for coverage.
func TestGenerateRSAKeyPairSizes(t *testing.T) {
	t.Parallel()

	provider := NewSoftwareProvider()

	tests := []struct {
		name    string
		bits    int
		wantErr bool
	}{
		{
			name:    "RSA-2048",
			bits:    2048,
			wantErr: false,
		},
		{
			name:    "RSA-3072",
			bits:    3072,
			wantErr: false,
		},
		{
			name:    "RSA-4096",
			bits:    4096,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			kp, err := provider.generateRSAKeyPair(tt.bits)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, kp)

			priv, ok := kp.PrivateKey.(*rsa.PrivateKey)
			require.True(t, ok)
			require.Equal(t, tt.bits, priv.N.BitLen())
		})
	}
}

// TestGenerateECDSAKeyPairCurves tests different ECDSA curves for coverage.
func TestGenerateECDSAKeyPairCurves(t *testing.T) {
	t.Parallel()

	provider := NewSoftwareProvider()

	tests := []struct {
		name    string
		curve   string
		wantErr bool
	}{
		{
			name:    "P-256",
			curve:   "P-256",
			wantErr: false,
		},
		{
			name:    "P-384",
			curve:   "P-384",
			wantErr: false,
		},
		{
			name:    "P-521",
			curve:   "P-521",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			kp, err := provider.generateECDSAKeyPair(tt.curve)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, kp)
			require.Equal(t, KeyTypeECDSA, kp.Type)
		})
	}
}

// TestGetSignatureAlgorithm tests GetSignatureAlgorithm for different key types.
func TestGetSignatureAlgorithm(t *testing.T) {
	t.Parallel()

	provider := NewSoftwareProvider()

	tests := []struct {
		name    string
		keySize int
		keyType string
		wantErr bool
	}{
		{
			name:    "RSA-2048",
			keySize: 2048,
			keyType: testRSAAlgorithm,
			wantErr: false,
		},
		{
			name:    "RSA-3072",
			keySize: 3072,
			keyType: testRSAAlgorithm,
			wantErr: false,
		},
		{
			name:    "RSA-4096",
			keySize: 4096,
			keyType: testRSAAlgorithm,
			wantErr: false,
		},
		{
			name:    "ECDSA-P256",
			keySize: 256,
			keyType: "ECDSA-P256",
			wantErr: false,
		},
		{
			name:    "ECDSA-P384",
			keySize: 384,
			keyType: "ECDSA-P384",
			wantErr: false,
		},
		{
			name:    "ECDSA-P521",
			keySize: 521,
			keyType: "ECDSA-P521",
			wantErr: false,
		},
		{
			name:    "Ed25519",
			keySize: 0,
			keyType: testEd25519Curve,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var publicKey crypto.PublicKey

			var err error

			switch tt.keyType {
			case testRSAAlgorithm:
				kp, genErr := provider.generateRSAKeyPair(tt.keySize)
				require.NoError(t, genErr)

				priv, ok := kp.PrivateKey.(*rsa.PrivateKey)
				require.True(t, ok)

				publicKey = &priv.PublicKey
			case "ECDSA-P256":
				kp, genErr := provider.generateECDSAKeyPair("P-256")
				require.NoError(t, genErr)

				publicKey = kp.PublicKey
			case "ECDSA-P384":
				kp, genErr := provider.generateECDSAKeyPair("P-384")
				require.NoError(t, genErr)

				publicKey = kp.PublicKey
			case "ECDSA-P521":
				kp, genErr := provider.generateECDSAKeyPair("P-521")
				require.NoError(t, genErr)

				publicKey = kp.PublicKey
			case testEd25519Curve:
				kp, genErr := provider.generateEdDSAKeyPair(testEd25519Curve)
				require.NoError(t, genErr)

				publicKey = kp.PublicKey
			}

			alg, err := provider.GetSignatureAlgorithm(publicKey)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.NotZero(t, alg, "algorithm should not be zero")
		})
	}
}

// TestGetSignatureAlgorithmUnsupportedKey tests GetSignatureAlgorithm with unsupported key type.
func TestGetSignatureAlgorithmUnsupportedKey(t *testing.T) {
	t.Parallel()

	provider := NewSoftwareProvider()

	// Test with unsupported key type (raw byte slice)
	invalidKey := []byte("not a real public key")

	_, err := provider.GetSignatureAlgorithm(invalidKey)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported public key type")
}
