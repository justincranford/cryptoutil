// Copyright (c) 2025 Justin Cranford

package crypto

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"crypto"
	"crypto/ed25519"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	sha256 "crypto/sha256"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
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
			name:    cryptoutilSharedMagic.EdCurveEd25519,
			curve:   testEd25519Curve,
			wantErr: false,
		},
		{
			name:    cryptoutilSharedMagic.EdCurveEd448,
			curve:   cryptoutilSharedMagic.EdCurveEd448,
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
	kp, err := provider.generateRSAKeyPair(cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	priv, ok := kp.PrivateKey.(*rsa.PrivateKey)
	require.True(t, ok)

	// Test verification with invalid signature
	digest := sha256.Sum256([]byte("test message"))
	invalidSignature := make([]byte, cryptoutilSharedMagic.MaxUnsealSharedSecrets)

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
			bits:    cryptoutilSharedMagic.DefaultMetricsBatchSize,
			wantErr: false,
		},
		{
			name:    "RSA-3072",
			bits:    cryptoutilSharedMagic.RSA3072KeySize,
			wantErr: false,
		},
		{
			name:    "RSA-4096",
			bits:    cryptoutilSharedMagic.RSA4096KeySize,
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
			keySize: cryptoutilSharedMagic.DefaultMetricsBatchSize,
			keyType: testRSAAlgorithm,
			wantErr: false,
		},
		{
			name:    "RSA-3072",
			keySize: cryptoutilSharedMagic.RSA3072KeySize,
			keyType: testRSAAlgorithm,
			wantErr: false,
		},
		{
			name:    "RSA-4096",
			keySize: cryptoutilSharedMagic.RSA4096KeySize,
			keyType: testRSAAlgorithm,
			wantErr: false,
		},
		{
			name:    "ECDSA-P256",
			keySize: cryptoutilSharedMagic.MaxUnsealSharedSecrets,
			keyType: "ECDSA-P256",
			wantErr: false,
		},
		{
			name:    "ECDSA-P384",
			keySize: cryptoutilSharedMagic.SymmetricKeySize384,
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
			name:    cryptoutilSharedMagic.EdCurveEd25519,
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

// errSignFailed is the sentinel error returned by failingSignerImpl and keygen error injections.
var errSignFailed = errors.New("test: sign operation failed")

// failingSignerImpl is a crypto.Signer that always returns an error from Sign().
// Used to cover the error path in SoftwareProvider.Sign().
type failingSignerImpl struct{}

func (f failingSignerImpl) Public() crypto.PublicKey { return []byte("pub") }

func (f failingSignerImpl) Sign(_ io.Reader, _ []byte, _ crypto.SignerOpts) ([]byte, error) {
	return nil, errSignFailed
}

// TestVerifyEdDSA_SuccessPath tests the success return path of verifyEdDSA.
// Covers the "return nil" statement on the happy path (line 204).
func TestVerifyEdDSA_SuccessPath(t *testing.T) {
	t.Parallel()

	provider := NewSoftwareProvider()

	// Generate Ed25519 key pair.
	kp, err := provider.generateEdDSAKeyPair(testEd25519Curve)
	require.NoError(t, err)

	pub, ok := kp.PublicKey.(ed25519.PublicKey)
	require.True(t, ok)

	priv, ok := kp.PrivateKey.(ed25519.PrivateKey)
	require.True(t, ok)

	// Sign the digest using the private key directly.
	digest := sha256.Sum256([]byte("message for success path"))
	signature := ed25519.Sign(priv, digest[:])

	// Verify the signature - must succeed (covers success "return nil" path).
	err = provider.verifyEdDSA(pub, digest[:], signature)
	require.NoError(t, err)
}

// TestSign_WithFailingSignerSign tests Sign when signer.Sign returns an error.
// Covers the error return in Sign after signer.Sign fails.
func TestSign_WithFailingSignerSign(t *testing.T) {
	t.Parallel()

	provider := NewSoftwareProvider()

	digest := sha256.Sum256([]byte("test sign error path"))

	_, err := provider.Sign(failingSignerImpl{}, digest[:], crypto.SHA256)
	require.Error(t, err)
	require.Contains(t, err.Error(), "signing failed")
}

// TestGenerateRSAKeyPair_KeygenError tests generateRSAKeyPair when keygen fails.
// NOTE: Must NOT use t.Parallel() - modifies package-level pkiCryptoGenerateRSAKeyPairFn.
func TestGenerateRSAKeyPair_KeygenError(t *testing.T) {
	orig := pkiCryptoGenerateRSAKeyPairFn
	pkiCryptoGenerateRSAKeyPairFn = func(_ int) (*cryptoutilSharedCryptoKeygen.KeyPair, error) {
		return nil, errSignFailed
	}

	defer func() { pkiCryptoGenerateRSAKeyPairFn = orig }()

	provider := NewSoftwareProvider()

	_, err := provider.generateRSAKeyPair(MinRSAKeyBits)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to generate RSA key pair")
}

// TestGenerateECDSAKeyPair_KeygenError tests generateECDSAKeyPair when keygen fails.
// NOTE: Must NOT use t.Parallel() - modifies package-level pkiCryptoGenerateECDSAKeyPairFn.
func TestGenerateECDSAKeyPair_KeygenError(t *testing.T) {
	orig := pkiCryptoGenerateECDSAKeyPairFn
	pkiCryptoGenerateECDSAKeyPairFn = func(_ elliptic.Curve) (*cryptoutilSharedCryptoKeygen.KeyPair, error) {
		return nil, errSignFailed
	}

	defer func() { pkiCryptoGenerateECDSAKeyPairFn = orig }()

	provider := NewSoftwareProvider()

	_, err := provider.generateECDSAKeyPair("P-256")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to generate ECDSA key pair")
}
