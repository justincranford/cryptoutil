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
	t.Parallel()

	injectedErr := errors.New("injected HKDF error")
	stubHKDFFn := func(_ string, _, _, _ []byte, _ int) ([]byte, error) { return nil, injectedErr }

	_, err := hashSecretHKDFFixedHighInternal("some-high-entropy-secret-value-1234567890", nil, stubHKDFFn)

	require.ErrorIs(t, err, injectedErr)
}

func TestVerifySecretHKDFFixedHigh_HashError(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected HKDF error")
	stubHKDFFn := func(_ string, _, _, _ []byte, _ int) ([]byte, error) { return nil, injectedErr }

	// storedHash must pass format validation (hkdf-sha256-fixed-high$base64dk)
	_, err := hashSecretHKDFFixedHighInternal("some-high-entropy-secret-value-1234567890", nil, stubHKDFFn)

	require.ErrorIs(t, err, injectedErr)
}

func TestHashSecretHKDFRandom_CrandReadError(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected crand error")
	stubCrandReadFn := func(_ []byte) (int, error) { return 0, injectedErr }
	stubHKDFFn := func(_ string, _, _, _ []byte, _ int) ([]byte, error) { return nil, nil }

	_, err := hashSecretHKDFRandomInternal("some-high-entropy-secret-value-1234567890", stubCrandReadFn, stubHKDFFn)

	require.ErrorIs(t, err, injectedErr)
}

func TestHashSecretHKDFRandom_HKDFError(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected HKDF error")
	stubCrandReadFn := func(b []byte) (int, error) { return len(b), nil } // succeed
	stubHKDFFn := func(_ string, _, _, _ []byte, _ int) ([]byte, error) { return nil, injectedErr }

	_, err := hashSecretHKDFRandomInternal("some-high-entropy-secret-value-1234567890", stubCrandReadFn, stubHKDFFn)

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
	t.Parallel()

	injectedErr := errors.New("injected HKDF error")
	stubHKDFFn := func(_ string, _, _, _ []byte, _ int) ([]byte, error) { return nil, injectedErr }

	_, err := hashVerifySecretHKDFRandomInternal("hkdf-sha256$aGVsbG8$aGVsbG8", "secret", stubHKDFFn)

	require.ErrorIs(t, err, injectedErr)
}

func TestHashSecretHKDFFixedLow_Error(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected HKDF error")
	stubHKDFFn := func(_ string, _, _, _ []byte, _ int) ([]byte, error) { return nil, injectedErr }

	_, err := hashSecretHKDFFixedInternal("username@example.com", nil, stubHKDFFn)

	require.ErrorIs(t, err, injectedErr)
}

func TestVerifySecretHKDFFixedLow_HashError(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected HKDF error")
	stubHKDFFn := func(_ string, _, _, _ []byte, _ int) ([]byte, error) { return nil, injectedErr }

	// storedHash must pass format validation (hkdf-sha256-fixed$base64dk)
	_, err := hashSecretHKDFFixedInternal("username@example.com", nil, stubHKDFFn)

	require.ErrorIs(t, err, injectedErr)
}

func TestHashSecretPBKDF2_Error(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected PBKDF2 error")
	stubPBKDF2Fn := func(_ string, _ *cryptoutilSharedCryptoDigests.PBKDF2Params) (string, error) {
		return "", injectedErr
	}

	_, err := hashSecretPBKDF2Internal("password123", stubPBKDF2Fn)

	require.ErrorIs(t, err, injectedErr)
}

func TestHashSecretPBKDF2WithParams_Error(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected PBKDF2 error")
	stubPBKDF2Fn := func(_ string, _ *cryptoutilSharedCryptoDigests.PBKDF2Params) (string, error) {
		return "", injectedErr
	}

	_, err := hashSecretPBKDF2WithParamsInternal("password123", DefaultPBKDF2ParameterSet(), stubPBKDF2Fn)

	require.ErrorIs(t, err, injectedErr)
}

func TestVerifySecretPBKDF2WithParams_Error(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected verify error")
	stubVerifyFn := func(_, _ string, _ *cryptoutilSharedCryptoDigests.PBKDF2Params) (bool, error) {
		return false, injectedErr
	}

	_, err := verifySecretPBKDF2WithParamsInternal("stored", "provided", DefaultPBKDF2ParameterSet(), stubVerifyFn)

	require.ErrorIs(t, err, injectedErr)
}
