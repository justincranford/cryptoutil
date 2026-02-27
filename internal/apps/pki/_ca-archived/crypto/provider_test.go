// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"crypto"
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	rsa "crypto/rsa"
	sha256 "crypto/sha256"
	"crypto/x509"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSoftwareProvider_GenerateKeyPair(t *testing.T) {
	t.Parallel()

	provider := NewSoftwareProvider()

	tests := []struct {
		name        string
		spec        KeySpec
		wantErr     bool
		errContains string
		checkType   func(t *testing.T, kp *KeyPair)
	}{
		{
			name:    "RSA-2048",
			spec:    KeySpec{Type: KeyTypeRSA, RSABits: cryptoutilSharedMagic.DefaultMetricsBatchSize},
			wantErr: false,
			checkType: func(t *testing.T, kp *KeyPair) {
				t.Helper()

				_, ok := kp.PrivateKey.(*rsa.PrivateKey)
				require.True(t, ok, "expected RSA private key")
				require.Equal(t, KeyTypeRSA, kp.Type)
				require.Equal(t, "RSA-2048", kp.Algorithm)
			},
		},
		{
			name:    "RSA-4096",
			spec:    KeySpec{Type: KeyTypeRSA, RSABits: cryptoutilSharedMagic.RSA4096KeySize},
			wantErr: false,
			checkType: func(t *testing.T, kp *KeyPair) {
				t.Helper()

				_, ok := kp.PrivateKey.(*rsa.PrivateKey)
				require.True(t, ok, "expected RSA private key")
				require.Equal(t, "RSA-4096", kp.Algorithm)
			},
		},
		{
			name:        "RSA-1024-too-small",
			spec:        KeySpec{Type: KeyTypeRSA, RSABits: cryptoutilSharedMagic.DefaultLogsBatchSize},
			wantErr:     true,
			errContains: "at least",
		},
		{
			name:    "ECDSA-P256",
			spec:    KeySpec{Type: KeyTypeECDSA, ECDSACurve: "P-256"},
			wantErr: false,
			checkType: func(t *testing.T, kp *KeyPair) {
				t.Helper()

				_, ok := kp.PrivateKey.(*ecdsa.PrivateKey)
				require.True(t, ok, "expected ECDSA private key")
				require.Equal(t, KeyTypeECDSA, kp.Type)
				require.Equal(t, "ECDSA-P-256", kp.Algorithm)
			},
		},
		{
			name:    "ECDSA-P384",
			spec:    KeySpec{Type: KeyTypeECDSA, ECDSACurve: "P-384"},
			wantErr: false,
			checkType: func(t *testing.T, kp *KeyPair) {
				t.Helper()

				_, ok := kp.PrivateKey.(*ecdsa.PrivateKey)
				require.True(t, ok, "expected ECDSA private key")
				require.Equal(t, "ECDSA-P-384", kp.Algorithm)
			},
		},
		{
			name:    "ECDSA-P521",
			spec:    KeySpec{Type: KeyTypeECDSA, ECDSACurve: "P-521"},
			wantErr: false,
			checkType: func(t *testing.T, kp *KeyPair) {
				t.Helper()

				_, ok := kp.PrivateKey.(*ecdsa.PrivateKey)
				require.True(t, ok, "expected ECDSA private key")
			},
		},
		{
			name:        "ECDSA-invalid-curve",
			spec:        KeySpec{Type: KeyTypeECDSA, ECDSACurve: "P-128"},
			wantErr:     true,
			errContains: "unsupported ECDSA curve",
		},
		{
			name:    "EdDSA-Ed25519",
			spec:    KeySpec{Type: KeyTypeEdDSA, EdDSACurve: cryptoutilSharedMagic.EdCurveEd25519},
			wantErr: false,
			checkType: func(t *testing.T, kp *KeyPair) {
				t.Helper()

				_, ok := kp.PrivateKey.(ed25519.PrivateKey)
				require.True(t, ok, "expected Ed25519 private key")
				require.Equal(t, KeyTypeEdDSA, kp.Type)
				require.Equal(t, cryptoutilSharedMagic.EdCurveEd25519, kp.Algorithm)
			},
		},
		{
			name:        "unsupported-type",
			spec:        KeySpec{Type: "INVALID"},
			wantErr:     true,
			errContains: "unsupported key type",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			kp, err := provider.GenerateKeyPair(tc.spec)

			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)
				require.Nil(t, kp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, kp)
				require.NotNil(t, kp.PublicKey)
				require.NotNil(t, kp.PrivateKey)

				if tc.checkType != nil {
					tc.checkType(t, kp)
				}
			}
		})
	}
}

func TestSoftwareProvider_SignAndVerify(t *testing.T) {
	t.Parallel()

	provider := NewSoftwareProvider()
	message := []byte("test message to sign")
	digest := sha256.Sum256(message)

	tests := []struct {
		name string
		spec KeySpec
		opts crypto.SignerOpts
	}{
		{
			name: "RSA-SHA256",
			spec: KeySpec{Type: KeyTypeRSA, RSABits: cryptoutilSharedMagic.DefaultMetricsBatchSize},
			opts: crypto.SHA256,
		},
		{
			name: "ECDSA-P256",
			spec: KeySpec{Type: KeyTypeECDSA, ECDSACurve: "P-256"},
			opts: crypto.SHA256,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Generate key pair.
			kp, err := provider.GenerateKeyPair(tc.spec)
			require.NoError(t, err)

			// Sign.
			signature, err := provider.Sign(kp.PrivateKey, digest[:], tc.opts)
			require.NoError(t, err)
			require.NotEmpty(t, signature)

			// Verify.
			err = provider.Verify(kp.PublicKey, digest[:], signature, tc.opts)
			require.NoError(t, err)
		})
	}
}

func TestSoftwareProvider_SignNotSigner(t *testing.T) {
	t.Parallel()

	provider := NewSoftwareProvider()

	// Use a type that doesn't implement crypto.Signer.
	_, err := provider.Sign("not-a-key", []byte("test"), crypto.SHA256)
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not implement crypto.Signer")
}

func TestSoftwareProvider_VerifyInvalidSignature(t *testing.T) {
	t.Parallel()

	provider := NewSoftwareProvider()
	message := []byte("test message")
	digest := sha256.Sum256(message)
	invalidSignature := []byte("invalid signature")

	tests := []struct {
		name string
		spec KeySpec
		opts crypto.SignerOpts
	}{
		{
			name: cryptoutilSharedMagic.KeyTypeRSA,
			spec: KeySpec{Type: KeyTypeRSA, RSABits: cryptoutilSharedMagic.DefaultMetricsBatchSize},
			opts: crypto.SHA256,
		},
		{
			name: "ECDSA",
			spec: KeySpec{Type: KeyTypeECDSA, ECDSACurve: "P-256"},
			opts: crypto.SHA256,
		},
		{
			name: cryptoutilSharedMagic.JoseAlgEdDSA,
			spec: KeySpec{Type: KeyTypeEdDSA, EdDSACurve: cryptoutilSharedMagic.EdCurveEd25519},
			opts: nil, // EdDSA uses pre-hashed, but for this test we pass the full message.
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			kp, err := provider.GenerateKeyPair(tc.spec)
			require.NoError(t, err)

			err = provider.Verify(kp.PublicKey, digest[:], invalidSignature, tc.opts)
			require.Error(t, err)
		})
	}
}

func TestSoftwareProvider_VerifyUnsupportedKey(t *testing.T) {
	t.Parallel()

	provider := NewSoftwareProvider()

	err := provider.Verify("not-a-key", []byte("digest"), []byte(cryptoutilSharedMagic.JoseKeyUseSig), crypto.SHA256)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported public key type")
}

func TestSoftwareProvider_GetSignatureAlgorithm(t *testing.T) {
	t.Parallel()

	provider := NewSoftwareProvider()

	tests := []struct {
		name     string
		spec     KeySpec
		expected x509.SignatureAlgorithm
	}{
		{
			name:     "RSA-2048",
			spec:     KeySpec{Type: KeyTypeRSA, RSABits: cryptoutilSharedMagic.DefaultMetricsBatchSize},
			expected: x509.SHA256WithRSA,
		},
		{
			name:     "RSA-3072",
			spec:     KeySpec{Type: KeyTypeRSA, RSABits: cryptoutilSharedMagic.RSA3072KeySize},
			expected: x509.SHA384WithRSA,
		},
		{
			name:     "RSA-4096",
			spec:     KeySpec{Type: KeyTypeRSA, RSABits: cryptoutilSharedMagic.RSA4096KeySize},
			expected: x509.SHA512WithRSA,
		},
		{
			name:     "ECDSA-P256",
			spec:     KeySpec{Type: KeyTypeECDSA, ECDSACurve: "P-256"},
			expected: x509.ECDSAWithSHA256,
		},
		{
			name:     "ECDSA-P384",
			spec:     KeySpec{Type: KeyTypeECDSA, ECDSACurve: "P-384"},
			expected: x509.ECDSAWithSHA384,
		},
		{
			name:     "ECDSA-P521",
			spec:     KeySpec{Type: KeyTypeECDSA, ECDSACurve: "P-521"},
			expected: x509.ECDSAWithSHA512,
		},
		{
			name:     "EdDSA-Ed25519",
			spec:     KeySpec{Type: KeyTypeEdDSA, EdDSACurve: cryptoutilSharedMagic.EdCurveEd25519},
			expected: x509.PureEd25519,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			kp, err := provider.GenerateKeyPair(tc.spec)
			require.NoError(t, err)

			sigAlg, err := provider.GetSignatureAlgorithm(kp.PublicKey)
			require.NoError(t, err)
			require.Equal(t, tc.expected, sigAlg)
		})
	}
}

func TestSoftwareProvider_GetSignatureAlgorithmUnsupported(t *testing.T) {
	t.Parallel()

	provider := NewSoftwareProvider()

	_, err := provider.GetSignatureAlgorithm("not-a-key")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported public key type")
}

func TestParseKeySpecFromConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		algorithm   string
		curveOrSize string
		wantErr     bool
		errContains string
		checkSpec   func(t *testing.T, spec KeySpec)
	}{
		{
			name:        "RSA-2048",
			algorithm:   cryptoutilSharedMagic.KeyTypeRSA,
			curveOrSize: "2048",
			wantErr:     false,
			checkSpec: func(t *testing.T, spec KeySpec) {
				t.Helper()
				require.Equal(t, KeyTypeRSA, spec.Type)
				require.Equal(t, cryptoutilSharedMagic.DefaultMetricsBatchSize, spec.RSABits)
			},
		},
		{
			name:        "RSA-invalid",
			algorithm:   cryptoutilSharedMagic.KeyTypeRSA,
			curveOrSize: "invalid",
			wantErr:     true,
			errContains: "invalid RSA key size",
		},
		{
			name:        "ECDSA-P256",
			algorithm:   "ECDSA",
			curveOrSize: "P-256",
			wantErr:     false,
			checkSpec: func(t *testing.T, spec KeySpec) {
				t.Helper()
				require.Equal(t, KeyTypeECDSA, spec.Type)
				require.Equal(t, "P-256", spec.ECDSACurve)
			},
		},
		{
			name:        "EdDSA-Ed25519",
			algorithm:   cryptoutilSharedMagic.JoseAlgEdDSA,
			curveOrSize: cryptoutilSharedMagic.EdCurveEd25519,
			wantErr:     false,
			checkSpec: func(t *testing.T, spec KeySpec) {
				t.Helper()
				require.Equal(t, KeyTypeEdDSA, spec.Type)
				require.Equal(t, cryptoutilSharedMagic.EdCurveEd25519, spec.EdDSACurve)
			},
		},
		{
			name:        "unsupported",
			algorithm:   "INVALID",
			curveOrSize: "256",
			wantErr:     true,
			errContains: "unsupported algorithm",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			spec, err := ParseKeySpecFromConfig(tc.algorithm, tc.curveOrSize)

			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)

				if tc.checkSpec != nil {
					tc.checkSpec(t, spec)
				}
			}
		})
	}
}
