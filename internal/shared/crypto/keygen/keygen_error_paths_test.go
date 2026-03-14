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
cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
"errors"
"testing"

circlEd448 "github.com/cloudflare/circl/sign/ed448"
"github.com/stretchr/testify/require"
)

// TestGenerateRSAKeyPair_RandError tests the RSA key generation error path via seam injection.
// Go 1.24+ ignores the rand io.Reader in rsa.GenerateKey; function-level seam is required.
func TestGenerateRSAKeyPair_RandError(t *testing.T) {
// Sequential: modifies package-level keygenRSAFn seam.
injectedErr := errors.New("injected RSA error")
orig := keygenRSAFn

keygenRSAFn = func(_ int) (*rsa.PrivateKey, error) { return nil, injectedErr }

defer func() { keygenRSAFn = orig }()

_, err := GenerateRSAKeyPair(cryptoutilSharedMagic.RSAKeySize2048)
require.ErrorIs(t, err, injectedErr)
}

// TestGenerateECDSAKeyPair_RandError tests the ECDSA key generation error path via seam injection.
// Go 1.24+ ignores the rand io.Reader in ecdsa.GenerateKey; function-level seam is required.
func TestGenerateECDSAKeyPair_RandError(t *testing.T) {
// Sequential: modifies package-level keygenECDSAFn seam.
injectedErr := errors.New("injected ECDSA error")
orig := keygenECDSAFn

keygenECDSAFn = func(_ elliptic.Curve) (*ecdsa.PrivateKey, error) { return nil, injectedErr }

defer func() { keygenECDSAFn = orig }()

_, err := GenerateECDSAKeyPair(elliptic.P256())
require.ErrorIs(t, err, injectedErr)
}

// TestGenerateECDHKeyPair_RandError tests the ECDH key generation error path via seam injection.
// Go 1.24+ ignores the rand io.Reader in ecdh.Curve.GenerateKey; function-level seam is required.
func TestGenerateECDHKeyPair_RandError(t *testing.T) {
// Sequential: modifies package-level keygenECDHFn seam.
injectedErr := errors.New("injected ECDH error")
orig := keygenECDHFn

keygenECDHFn = func(_ ecdh.Curve) (*ecdh.PrivateKey, error) { return nil, injectedErr }

defer func() { keygenECDHFn = orig }()

_, err := GenerateECDHKeyPair(ecdh.P256())
require.ErrorIs(t, err, injectedErr)
}

func TestGenerateEDDSAKeyPair_Ed448RandError(t *testing.T) {
// Sequential: modifies package-level keygenEdDSAEd448Fn seam.
injectedErr := errors.New("injected Ed448 error")
orig := keygenEdDSAEd448Fn

keygenEdDSAEd448Fn = func() (circlEd448.PublicKey, circlEd448.PrivateKey, error) {
return nil, nil, injectedErr
}

defer func() { keygenEdDSAEd448Fn = orig }()

_, err := GenerateEDDSAKeyPair(EdCurveEd448)
require.ErrorIs(t, err, injectedErr)
}

func TestGenerateEDDSAKeyPair_Ed25519RandError(t *testing.T) {
// Sequential: modifies package-level keygenEdDSAEd25519Fn seam.
injectedErr := errors.New("injected Ed25519 error")
orig := keygenEdDSAEd25519Fn

keygenEdDSAEd25519Fn = func() (ed25519.PublicKey, ed25519.PrivateKey, error) {
return nil, nil, injectedErr
}

defer func() { keygenEdDSAEd25519Fn = orig }()

_, err := GenerateEDDSAKeyPair(EdCurveEd25519)
require.ErrorIs(t, err, injectedErr)
}

func TestGenerateAESKey_GenerateBytesError(t *testing.T) {
// Sequential: modifies package-level keygenGenerateBytesFn seam.
injectedErr := errors.New("injected GenerateBytes error")
orig := keygenGenerateBytesFn

keygenGenerateBytesFn = func(_ int) ([]byte, error) { return nil, injectedErr }

defer func() { keygenGenerateBytesFn = orig }()

_, err := GenerateAESKey(cryptoutilSharedMagic.TLSSelfSignedCertSerialNumberBits)
require.ErrorIs(t, err, injectedErr)
}

func TestGenerateAESHSKey_GenerateBytesError(t *testing.T) {
// Sequential: modifies package-level keygenGenerateBytesFn seam.
injectedErr := errors.New("injected GenerateBytes error")
orig := keygenGenerateBytesFn

keygenGenerateBytesFn = func(_ int) ([]byte, error) { return nil, injectedErr }

defer func() { keygenGenerateBytesFn = orig }()

_, err := GenerateAESHSKey(cryptoutilSharedMagic.MaxUnsealSharedSecrets)
require.ErrorIs(t, err, injectedErr)
}

func TestGenerateHMACKey_GenerateBytesError(t *testing.T) {
// Sequential: modifies package-level keygenGenerateBytesFn seam.
injectedErr := errors.New("injected GenerateBytes error")
orig := keygenGenerateBytesFn

keygenGenerateBytesFn = func(_ int) ([]byte, error) { return nil, injectedErr }

defer func() { keygenGenerateBytesFn = orig }()

_, err := GenerateHMACKey(cryptoutilSharedMagic.MaxUnsealSharedSecrets)
require.ErrorIs(t, err, injectedErr)
}

// Compile-time check: unused import crand is used via crand.Reader below if needed.
var _ = crand.Reader
