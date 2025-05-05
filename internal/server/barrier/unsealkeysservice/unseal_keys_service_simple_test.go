package unsealkeysservice

import (
	cryptoutilJose "cryptoutil/internal/common/crypto/jose"
	"testing"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

const jwkCount = 10

func TestNewUnsealKeysServiceSimple_HappyPath(t *testing.T) {
	unsealKeys := cryptoutilJose.GenerateAes256KeysForTest(t, jwkCount, &cryptoutilJose.AlgA256KW, &cryptoutilJose.EncA256GCM)
	unsealKeysService, err := NewUnsealKeysServiceSimple(unsealKeys)
	require.NoError(t, err)
	require.NotNil(t, unsealKeysService)
}

func TestNewUnsealKeysServiceSimple_SadPath_EmptyInput(t *testing.T) {
	unsealKeysService, err := NewUnsealKeysServiceSimple([]joseJwk.Key{})
	require.Error(t, err)
	require.Nil(t, unsealKeysService)
	require.EqualError(t, err, "unsealJwks can't be empty", "Unexpected error message")
}

func TestNewUnsealKeysServiceSimple_SadPath_NilInput(t *testing.T) {
	unsealKeysService, err := NewUnsealKeysServiceSimple(nil)
	require.Error(t, err)
	require.Nil(t, unsealKeysService)
	require.EqualError(t, err, "unsealJwks can't be nil", "Unexpected error message")
}
