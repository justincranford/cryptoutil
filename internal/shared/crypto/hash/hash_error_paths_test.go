// Copyright (c) 2025 Justin Cranford
//
//

package hash

import (
	"errors"
	"testing"

	cryptoutilSharedCryptoDigests "cryptoutil/internal/shared/crypto/digests"

	"github.com/stretchr/testify/require"
)

func TestHashSecretHKDFFixedHigh_Error(t *testing.T) {
	injectedErr := errors.New("injected HKDF error")
	orig := hashHighFixedHKDFFn

	hashHighFixedHKDFFn = func(_ string, _, _, _ []byte, _ int) ([]byte, error) { return nil, injectedErr }

	defer func() { hashHighFixedHKDFFn = orig }()

	_, err := HashHighEntropyDeterministic("some-high-entropy-secret-value-1234567890")

	require.ErrorIs(t, err, injectedErr)
}

func TestVerifySecretHKDFFixedHigh_HashError(t *testing.T) {
	injectedErr := errors.New("injected HKDF error")
	orig := hashHighFixedHKDFFn

	hashHighFixedHKDFFn = func(_ string, _, _, _ []byte, _ int) ([]byte, error) { return nil, injectedErr }

	defer func() { hashHighFixedHKDFFn = orig }()

	// storedHash must pass format validation (hkdf-sha256-fixed-high$base64dk)
	_, err := VerifySecretHKDFFixedHigh("hkdf-sha256-fixed-high$aGVsbG8=", "some-high-entropy-secret-value-1234567890")

	require.ErrorIs(t, err, injectedErr)
}

func TestHashSecretHKDFRandom_CrandReadError(t *testing.T) {
	injectedErr := errors.New("injected crand error")
	orig := hashHighRandomCrandReadFn

	hashHighRandomCrandReadFn = func(_ []byte) (int, error) { return 0, injectedErr }

	defer func() { hashHighRandomCrandReadFn = orig }()

	_, err := HashSecretHKDFRandom("some-high-entropy-secret-value-1234567890")

	require.ErrorIs(t, err, injectedErr)
}

func TestHashSecretHKDFRandom_HKDFError(t *testing.T) {
	injectedErr := errors.New("injected HKDF error")
	orig := hashHighRandomHKDFFn

	hashHighRandomHKDFFn = func(_ string, _, _, _ []byte, _ int) ([]byte, error) { return nil, injectedErr }

	defer func() { hashHighRandomHKDFFn = orig }()

	_, err := HashSecretHKDFRandom("some-high-entropy-secret-value-1234567890")

	require.ErrorIs(t, err, injectedErr)
}

func TestVerifySecretHKDFRandom_InvalidParts(t *testing.T) {
	_, err := VerifySecretHKDFRandom("no-dollar-signs-here", "secret")

	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid HKDF hash format")
}

func TestVerifySecretHKDFRandom_WrongAlgo(t *testing.T) {
	_, err := VerifySecretHKDFRandom("wrong-algo$aGVsbG8=$aGVsbG8=", "secret")

	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported hash algorithm")
}

func TestVerifySecretHKDFRandom_InvalidSalt(t *testing.T) {
	_, err := VerifySecretHKDFRandom("hkdf-sha256$!not-valid-base64!$aGVsbG8", "secret")

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode salt")
}

func TestVerifySecretHKDFRandom_InvalidDK(t *testing.T) {
	_, err := VerifySecretHKDFRandom("hkdf-sha256$aGVsbG8$!not-valid-base64!", "secret")

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode derived key")
}

func TestVerifySecretHKDFRandom_HKDFError(t *testing.T) {
	injectedErr := errors.New("injected HKDF error")
	orig := hashHighRandomHKDFFn

	hashHighRandomHKDFFn = func(_ string, _, _, _ []byte, _ int) ([]byte, error) { return nil, injectedErr }

	defer func() { hashHighRandomHKDFFn = orig }()

	_, err := VerifySecretHKDFRandom("hkdf-sha256$aGVsbG8$aGVsbG8", "secret")

	require.ErrorIs(t, err, injectedErr)
}

func TestHashSecretHKDFFixedLow_Error(t *testing.T) {
	injectedErr := errors.New("injected HKDF error")
	orig := hashLowFixedHKDFFn

	hashLowFixedHKDFFn = func(_ string, _, _, _ []byte, _ int) ([]byte, error) { return nil, injectedErr }

	defer func() { hashLowFixedHKDFFn = orig }()

	_, err := HashLowEntropyDeterministic("username@example.com")

	require.ErrorIs(t, err, injectedErr)
}

func TestVerifySecretHKDFFixedLow_HashError(t *testing.T) {
	injectedErr := errors.New("injected HKDF error")
	orig := hashLowFixedHKDFFn

	hashLowFixedHKDFFn = func(_ string, _, _, _ []byte, _ int) ([]byte, error) { return nil, injectedErr }

	defer func() { hashLowFixedHKDFFn = orig }()

	// storedHash must pass format validation (hkdf-sha256-fixed$base64dk)
	_, err := VerifySecretHKDFFixed("hkdf-sha256-fixed$aGVsbG8=", "username@example.com")

	require.ErrorIs(t, err, injectedErr)
}

func TestHashSecretPBKDF2_Error(t *testing.T) {
	injectedErr := errors.New("injected PBKDF2 error")
	orig := hashPBKDF2WithParamsFn

	hashPBKDF2WithParamsFn = func(_ string, _ *cryptoutilSharedCryptoDigests.PBKDF2Params) (string, error) {
		return "", injectedErr
	}

	defer func() { hashPBKDF2WithParamsFn = orig }()

	_, err := HashSecretPBKDF2("password123")

	require.ErrorIs(t, err, injectedErr)
}

func TestHashSecretPBKDF2WithParams_Error(t *testing.T) {
	injectedErr := errors.New("injected PBKDF2 error")
	orig := hashPBKDF2WithParamsFn

	hashPBKDF2WithParamsFn = func(_ string, _ *cryptoutilSharedCryptoDigests.PBKDF2Params) (string, error) {
		return "", injectedErr
	}

	defer func() { hashPBKDF2WithParamsFn = orig }()

	_, err := HashSecretPBKDF2WithParams("password123", DefaultPBKDF2ParameterSet())

	require.ErrorIs(t, err, injectedErr)
}

func TestVerifySecretPBKDF2WithParams_Error(t *testing.T) {
	injectedErr := errors.New("injected verify error")
	orig := hashVerifySecretWithParamsFn

	hashVerifySecretWithParamsFn = func(_, _ string, _ *cryptoutilSharedCryptoDigests.PBKDF2Params) (bool, error) {
		return false, injectedErr
	}

	defer func() { hashVerifySecretWithParamsFn = orig }()

	_, err := VerifySecretPBKDF2WithParams("stored", "provided", DefaultPBKDF2ParameterSet())

	require.ErrorIs(t, err, injectedErr)
}
