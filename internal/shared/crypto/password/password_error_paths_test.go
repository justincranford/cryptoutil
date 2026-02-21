// Copyright (c) 2025 ZREV Enterprises LLC. All rights reserved.
// Use of this source code is governed by the MIT License.

package password

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashPassword_Error(t *testing.T) {
	injectedErr := errors.New("injected hash error")
	orig := passwordHashFn

	passwordHashFn = func(_ string) (string, error) { return "", injectedErr }

	defer func() { passwordHashFn = orig }()

	_, err := HashPassword("password123")

	require.ErrorIs(t, err, injectedErr)
}

func TestVerifyPassword_BcryptInvalidHash(t *testing.T) {
	t.Parallel()

	// "$2a$" prefix makes DetectHashType return "bcrypt", but the hash is too short
	// so bcrypt.CompareHashAndPassword returns ErrHashTooShort (not ErrMismatchedHashAndPassword).
	_, _, err := VerifyPassword("password", "$2a$")

	require.Error(t, err)
	require.Contains(t, err.Error(), "legacy hash verification failed")
}

func TestVerifyPassword_PBKDF2VerifyError(t *testing.T) {
	injectedErr := errors.New("injected verify error")
	orig := passwordVerifyFn

	passwordVerifyFn = func(_, _ string) (bool, error) { return false, injectedErr }

	defer func() { passwordVerifyFn = orig }()

	// Use a valid pbkdf2 hash prefix so DetectHashType returns "pbkdf2".
	_, _, err := VerifyPassword("password", "$pbkdf2-sha256$600000$aGVsbG8$aGVsbG8")

	require.ErrorIs(t, err, injectedErr)
}

func TestVerifyPassword_UnknownHashType(t *testing.T) {
	t.Parallel()

	// Pass a hash string that DetectHashType returns "unknown" for.
	_, _, err := VerifyPassword("password", "unknown-hash-prefix-value")

	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown hash type")
}
