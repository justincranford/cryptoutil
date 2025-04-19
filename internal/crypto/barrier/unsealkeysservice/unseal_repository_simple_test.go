package unsealkeysservice

import (
	"testing"

	cryptoutilJose "cryptoutil/internal/crypto/jose"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/assert"
)

func TestNewUnsealKeysServiceSimple_HappyPath(t *testing.T) {
	const newConst = 10
	unsealKeys := make([]joseJwk.Key, 0, newConst)
	for range newConst {
		unsealJwk, _, _, err := cryptoutilJose.GenerateAesJWK(cryptoutilJose.AlgDIRECT)
		assert.NoError(t, err, "Expected no error")
		unsealKeys = append(unsealKeys, unsealJwk)
	}

	unsealKeysService, err := NewUnsealKeysServiceSimple(unsealKeys)
	assert.NoError(t, err, "Expected no error")
	assert.NotNil(t, unsealKeysService, "Repository should not be nil")
}

func TestNewUnsealKeysServiceSimple_SadPath_NilInput(t *testing.T) {
	unsealKeysService, err := NewUnsealKeysServiceSimple(nil)
	assert.Error(t, err, "Expected error for nil input")
	assert.Nil(t, unsealKeysService, "Repository should be nil for nil input")
	assert.EqualError(t, err, "unsealJwks can't be nil", "Unexpected error message")
}

func TestNewUnsealKeysServiceSimple_SadPath_EmptyInput(t *testing.T) {
	unsealKeysService, err := NewUnsealKeysServiceSimple([]joseJwk.Key{})
	assert.Error(t, err, "Expected error for empty input")
	assert.Nil(t, unsealKeysService, "Repository should be nil for empty input")
	assert.EqualError(t, err, "unsealJwks can't be empty", "Unexpected error message")
}
