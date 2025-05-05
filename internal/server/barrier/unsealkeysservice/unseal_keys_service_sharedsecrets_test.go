package unsealkeysservice

import (
	"cryptoutil/internal/common/crypto/keygen"
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

func TestSharedSecretsCountGreaterThan256(t *testing.T) {
	sharedSecretsM := make([][]byte, 257)
	for i := range sharedSecretsM {
		sharedSecretsM[i] = make([]byte, 32)
	}
	_, err := NewUnsealKeysServiceSharedSecrets(sharedSecretsM, 1)
	require.Error(t, err)
	require.Equal(t, "shared secrets can't be greater than 256", err.Error())
}

func TestChooseNZero(t *testing.T) {
	sharedSecretsM := [][]byte{
		make([]byte, 32),
	}
	_, err := NewUnsealKeysServiceSharedSecrets(sharedSecretsM, 0)
	require.Error(t, err)
	require.Equal(t, "n can't be zero", err.Error())
}

func TestChooseNNegative(t *testing.T) {
	sharedSecretsM := [][]byte{
		make([]byte, 32),
	}
	_, err := NewUnsealKeysServiceSharedSecrets(sharedSecretsM, -1)
	require.Error(t, err)
	require.Equal(t, "n can't be negative", err.Error())
}

func TestChooseNGreaterThanCount(t *testing.T) {
	sharedSecretsM := [][]byte{
		make([]byte, 32),
		make([]byte, 32),
	}
	_, err := NewUnsealKeysServiceSharedSecrets(sharedSecretsM, 3)
	require.Error(t, err)
	require.Equal(t, "n can't be greater than shared secrets count", err.Error())
}

func TestSharedSecretNil(t *testing.T) {
	sharedSecretsM := [][]byte{
		make([]byte, 32),
		nil,
	}
	_, err := NewUnsealKeysServiceSharedSecrets(sharedSecretsM, 1)
	require.Error(t, err)
	require.Equal(t, "shared secret 1 can't be nil", err.Error())
}

func TestSharedSecretLengthLessThan32(t *testing.T) {
	sharedSecretsM := [][]byte{
		make([]byte, 32),
		make([]byte, 31),
	}
	_, err := NewUnsealKeysServiceSharedSecrets(sharedSecretsM, 1)
	require.Error(t, err)
	require.Equal(t, "shared secret 1 length can't be greater than 32", err.Error())
}

func TestSharedSecretLengthGreaterThan64(t *testing.T) {
	sharedSecretsM := [][]byte{
		make([]byte, 32),
		make([]byte, 65),
	}
	_, err := NewUnsealKeysServiceSharedSecrets(sharedSecretsM, 1)
	require.Error(t, err)
	require.Equal(t, "shared secret 1 length can't be greater than 64", err.Error())
}
