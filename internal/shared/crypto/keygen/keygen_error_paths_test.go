// Copyright (c) 2025 Justin Cranford
//
//

package keygen

import (
	"crypto/ecdh"
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"errors"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	circlEd448 "github.com/cloudflare/circl/sign/ed448"
	"github.com/stretchr/testify/require"
)

// TestGenerateRSAKeyPair_RandError tests the RSA key generation error path via seam injection.
// Go 1.24+ ignores the rand io.Reader in rsa.GenerateKey; function-level seam is required.
func TestGenerateRSAKeyPair_RandError(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected RSA error")

	_, err := generateRSAKeyPairInternal(cryptoutilSharedMagic.RSAKeySize2048, func(_ int) (*rsa.PrivateKey, error) { return nil, injectedErr })
	require.ErrorIs(t, err, injectedErr)
}

// TestGenerateECDSAKeyPair_RandError tests the ECDSA key generation error path via seam injection.
// Go 1.24+ ignores the rand io.Reader in ecdsa.GenerateKey; function-level seam is required.
func TestGenerateECDSAKeyPair_RandError(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected ECDSA error")

	_, err := generateECDSAKeyPairInternal(elliptic.P256(), func(_ elliptic.Curve) (*ecdsa.PrivateKey, error) { return nil, injectedErr })
	require.ErrorIs(t, err, injectedErr)
}

// TestGenerateECDHKeyPair_RandError tests the ECDH key generation error path via seam injection.
// Go 1.24+ ignores the rand io.Reader in ecdh.Curve.GenerateKey; function-level seam is required.
func TestGenerateECDHKeyPair_RandError(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected ECDH error")

	_, err := generateECDHKeyPairInternal(ecdh.P256(), func(_ ecdh.Curve) (*ecdh.PrivateKey, error) { return nil, injectedErr })
	require.ErrorIs(t, err, injectedErr)
}

func TestGenerateEDDSAKeyPair_Ed448RandError(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected Ed448 error")

	_, err := generateEDDSAKeyPairInternal(EdCurveEd448, func() (circlEd448.PublicKey, circlEd448.PrivateKey, error) {
		return nil, nil, injectedErr
	}, func() (ed25519.PublicKey, ed25519.PrivateKey, error) { return nil, nil, nil })
	require.ErrorIs(t, err, injectedErr)
}

func TestGenerateEDDSAKeyPair_Ed25519RandError(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected Ed25519 error")

	_, err := generateEDDSAKeyPairInternal(EdCurveEd25519, func() (circlEd448.PublicKey, circlEd448.PrivateKey, error) { return nil, nil, nil }, func() (ed25519.PublicKey, ed25519.PrivateKey, error) {
		return nil, nil, injectedErr
	})
	require.ErrorIs(t, err, injectedErr)
}

func TestGenerateAESKey_GenerateBytesError(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected GenerateBytes error")

	_, err := generateAESKeyInternal(cryptoutilSharedMagic.TLSSelfSignedCertSerialNumberBits, func(_ int) ([]byte, error) { return nil, injectedErr })
	require.ErrorIs(t, err, injectedErr)
}

func TestGenerateAESHSKey_GenerateBytesError(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected GenerateBytes error")

	_, err := generateAESHSKeyInternal(cryptoutilSharedMagic.MaxUnsealSharedSecrets, func(_ int) ([]byte, error) { return nil, injectedErr })
	require.ErrorIs(t, err, injectedErr)
}

func TestGenerateHMACKey_GenerateBytesError(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected GenerateBytes error")

	_, err := generateHMACKeyInternal(cryptoutilSharedMagic.MaxUnsealSharedSecrets, func(_ int) ([]byte, error) { return nil, injectedErr })
	require.ErrorIs(t, err, injectedErr)
}

// Compile-time check: unused import crand is used via crand.Reader below if needed.
var _ = crand.Reader
