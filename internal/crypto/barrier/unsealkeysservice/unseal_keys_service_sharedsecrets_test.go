package unsealkeysservice

import (
	"cryptoutil/internal/crypto/keygen"
	"testing"

	"github.com/stretchr/testify/require"
)

const sharedSecretCount = 10

func TestNewUnsealKeysServiceSharedSecrets_HappyPath(t *testing.T) {
	unsealKeys, err := keygen.GenerateSharedSecrets(sharedSecretCount, 32)
	require.NoError(t, err)
	unsealKeysService, err := NewUnsealKeysServiceSharedSecrets(unsealKeys, sharedSecretCount-1)
	require.NoError(t, err)
	require.NotNil(t, unsealKeysService)
}

func TestNewUnsealKeysServiceSharedSecrets_SadPath_EmptySharedSecrets(t *testing.T) {
	unsealKeysService, err := NewUnsealKeysServiceSharedSecrets([][]byte{}, 1)
	require.Error(t, err)
	require.Nil(t, unsealKeysService)
	require.EqualError(t, err, "shared secrets can't be zero")
}

func TestNewUnsealKeysServiceSharedSecrets_SadPath_NilSharedSecrets(t *testing.T) {
	unsealKeysService, err := NewUnsealKeysServiceSharedSecrets(nil, 1)
	require.Error(t, err)
	require.Nil(t, unsealKeysService)
	require.EqualError(t, err, "shared secrets can't be nil")
}

func TestNewUnsealKeysServiceSharedSecrets_SadPath_NilSharedSecret(t *testing.T) {
	unsealKeysService, err := NewUnsealKeysServiceSharedSecrets([][]byte{nil}, 1)
	require.Error(t, err)
	require.Nil(t, unsealKeysService)
	require.EqualError(t, err, "shared secret 0 can't be nil")
}
