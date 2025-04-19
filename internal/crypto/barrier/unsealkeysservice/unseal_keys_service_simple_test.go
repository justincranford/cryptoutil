package unsealkeysservice

import (
	"testing"

	cryptoutilJose "cryptoutil/internal/crypto/jose"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

func TestNewUnsealKeysServiceSimple_HappyPath(t *testing.T) {
	const newConst = 10
	unsealKeys := make([]joseJwk.Key, 0, newConst)
	for range newConst {
		unsealJwk, _, _, err := cryptoutilJose.GenerateAesJWK(cryptoutilJose.AlgDIRECT)
		require.NoError(t, err, "Expected no error")
		unsealKeys = append(unsealKeys, unsealJwk)
	}

	unsealKeysService, err := NewUnsealKeysServiceSimple(unsealKeys)
	require.NoError(t, err, "Expected no error")
	require.NotNil(t, unsealKeysService, "Repository should not be nil")
}

func TestNewUnsealKeysServiceSimple_SadPath_NilInput(t *testing.T) {
	unsealKeysService, err := NewUnsealKeysServiceSimple(nil)
	require.Error(t, err, "Expected error for nil input")
	require.Nil(t, unsealKeysService, "Repository should be nil for nil input")
	require.EqualError(t, err, "unsealJwks can't be nil", "Unexpected error message")
}

func TestNewUnsealKeysServiceSimple_SadPath_EmptyInput(t *testing.T) {
	unsealKeysService, err := NewUnsealKeysServiceSimple([]joseJwk.Key{})
	require.Error(t, err, "Expected error for empty input")
	require.Nil(t, unsealKeysService, "Repository should be nil for empty input")
	require.EqualError(t, err, "unsealJwks can't be empty", "Unexpected error message")
}
