// Copyright (c) 2025 Justin Cranford

package keygen

import (
	"crypto/ecdh"
	"crypto/elliptic"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type errReader struct{}

func (errReader) Read(_ []byte) (int, error) { return 0, errors.New("injected reader error") }

func TestGenerateRSAKeyPair_RandError(t *testing.T) {
	orig := keygenRandReader
	keygenRandReader = errReader{}

	defer func() { keygenRandReader = orig }()

	_, err := GenerateRSAKeyPair(2048)
	require.Error(t, err)
}

func TestGenerateECDSAKeyPair_RandError(t *testing.T) {
	orig := keygenRandReader
	keygenRandReader = errReader{}

	defer func() { keygenRandReader = orig }()

	_, err := GenerateECDSAKeyPair(elliptic.P256())
	require.Error(t, err)
}

func TestGenerateECDHKeyPair_RandError(t *testing.T) {
	orig := keygenRandReader
	keygenRandReader = errReader{}

	defer func() { keygenRandReader = orig }()

	_, err := GenerateECDHKeyPair(ecdh.P256())
	require.Error(t, err)
}

func TestGenerateEDDSAKeyPair_Ed448RandError(t *testing.T) {
	orig := keygenRandReader
	keygenRandReader = errReader{}

	defer func() { keygenRandReader = orig }()

	_, err := GenerateEDDSAKeyPair(EdCurveEd448)
	require.Error(t, err)
}

func TestGenerateEDDSAKeyPair_Ed25519RandError(t *testing.T) {
	orig := keygenRandReader
	keygenRandReader = errReader{}

	defer func() { keygenRandReader = orig }()

	_, err := GenerateEDDSAKeyPair(EdCurveEd25519)
	require.Error(t, err)
}

func TestGenerateAESKey_GenerateBytesError(t *testing.T) {
	injectedErr := errors.New("injected GenerateBytes error")
	orig := keygenGenerateBytesFn
	keygenGenerateBytesFn = func(_ int) ([]byte, error) { return nil, injectedErr }

	defer func() { keygenGenerateBytesFn = orig }()

	_, err := GenerateAESKey(128)
	require.ErrorIs(t, err, injectedErr)
}

func TestGenerateAESHSKey_GenerateBytesError(t *testing.T) {
	injectedErr := errors.New("injected GenerateBytes error")
	orig := keygenGenerateBytesFn
	keygenGenerateBytesFn = func(_ int) ([]byte, error) { return nil, injectedErr }

	defer func() { keygenGenerateBytesFn = orig }()

	_, err := GenerateAESHSKey(256)
	require.ErrorIs(t, err, injectedErr)
}

func TestGenerateHMACKey_GenerateBytesError(t *testing.T) {
	injectedErr := errors.New("injected GenerateBytes error")
	orig := keygenGenerateBytesFn
	keygenGenerateBytesFn = func(_ int) ([]byte, error) { return nil, injectedErr }

	defer func() { keygenGenerateBytesFn = orig }()

	_, err := GenerateHMACKey(256)
	require.ErrorIs(t, err, injectedErr)
}
