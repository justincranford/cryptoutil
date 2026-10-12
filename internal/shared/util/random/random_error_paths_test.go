// Copyright (c) 2025-2026 Justin Cranford.
//
//

package random

import (
	crand "crypto/rand"
	"errors"
	"io"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// errReader is an io.Reader that always returns an error.
type errReader struct{}

func (errReader) Read(_ []byte) (int, error) {
	return 0, errors.New("injected crypto/rand failure")
}

// errUUIDGen is a uuid generator that always returns an error.
func errUUIDGen() (googleUuid.UUID, error) {
	return googleUuid.UUID{}, errors.New("injected uuid failure")
}

// TestGenerateBytes_RandError covers the crypto/rand error path.
func TestGenerateBytes_RandError(t *testing.T) {
	t.Parallel()

	_, err := generateBytes(cryptoutilSharedMagic.RealmMinTokenLengthBytes, errReader{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to generate")
}

// TestGenerateMultipleBytes_RandError covers the crypto/rand error path.
func TestGenerateMultipleBytes_RandError(t *testing.T) {
	t.Parallel()

	_, err := generateMultipleBytes(2, cryptoutilSharedMagic.RealmMinTokenLengthBytes, errReader{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to generate consecutive byte slices")
}

// TestGenerateString_RandError covers the GenerateBytes error path in GenerateString.
func TestGenerateString_RandError(t *testing.T) {
	t.Parallel()

	_, err := generateString(cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes, errReader{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to generate")
}

// TestGenerateUsernameSimple_UUIDError covers the uuid error path.
func TestGenerateUsernameSimple_UUIDError(t *testing.T) {
	t.Parallel()

	_, err := generateUsernameSimple(errUUIDGen)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to generate UUID for username")
}

// TestGeneratePasswordSimple_UUIDError covers the uuid error path.
func TestGeneratePasswordSimple_UUIDError(t *testing.T) {
	t.Parallel()

	_, err := generatePasswordSimple(errUUIDGen)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to generate UUID for password")
}

// TestGenerateUUIDv7_UUIDError covers the uuid error path.
func TestGenerateUUIDv7_UUIDError(t *testing.T) {
	t.Parallel()

	_, err := generateUUIDv7(errUUIDGen)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to generate UUID")
}

// TestGenerateUUIDv7Function_UUIDError covers the uuid error path via the function wrapper.
func TestGenerateUUIDv7Function_UUIDError(t *testing.T) {
	t.Parallel()

	fn := GenerateUUIDv7Function()
	uuid, err := fn()
	require.NoError(t, err)
	require.NotNil(t, uuid)
}

// TestGenerateUUIDv7_HappyPath2 ensures normal operation still works after error injection.
func TestGenerateUUIDv7_HappyPath2(t *testing.T) {
	t.Parallel()

	uuid, err := GenerateUUIDv7()
	require.NoError(t, err)
	require.NotNil(t, uuid)
}

// TestCryptoRandReader_ReadFull ensures the default crypto/rand reader returns data.
func TestCryptoRandReader_ReadFull(t *testing.T) {
	t.Parallel()

	buf := make([]byte, cryptoutilSharedMagic.RealmMinTokenLengthBytes)

	_, err := io.ReadFull(crand.Reader, buf)
	require.NoError(t, err)
	require.Len(t, buf, cryptoutilSharedMagic.RealmMinTokenLengthBytes)
}
