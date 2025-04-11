package unsealrepository

import (
	cryptoutilJose "cryptoutil/internal/crypto/jose"
	"testing"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/assert"
)

func TestNewUnsealKeyRepositorySimple_HappyPath(t *testing.T) {
	const newConst = 10
	unsealKeys := make([]joseJwk.Key, 0, newConst)
	for _ = range newConst {
		unsealJwk, _, err := cryptoutilJose.GenerateAesJWK(cryptoutilJose.AlgDIRECT)
		assert.NoError(t, err, "Expected no error")
		unsealKeys = append(unsealKeys, unsealJwk)
	}

	unsealKeyRepository, err := NewUnsealKeyRepositorySimple(unsealKeys)
	assert.NoError(t, err, "Expected no error")
	assert.NotNil(t, unsealKeyRepository, "Repository should not be nil")
	assert.Equal(t, unsealKeys, unsealKeyRepository.UnsealJwks(), "Expected returned JWKs to match input JWKs")
}

func TestNewUnsealKeyRepositorySimple_SadPath_NilInput(t *testing.T) {
	unsealKeyRepository, err := NewUnsealKeyRepositorySimple(nil)
	assert.Error(t, err, "Expected error for nil input")
	assert.Nil(t, unsealKeyRepository, "Repository should be nil for nil input")
	assert.EqualError(t, err, "unsealJwks can't be nil", "Unexpected error message")
}

func TestNewUnsealKeyRepositorySimple_SadPath_EmptyInput(t *testing.T) {
	unsealKeyRepository, err := NewUnsealKeyRepositorySimple([]joseJwk.Key{})
	assert.Error(t, err, "Expected error for empty input")
	assert.Nil(t, unsealKeyRepository, "Repository should be nil for empty input")
	assert.EqualError(t, err, "unsealJwks can't be empty", "Unexpected error message")
}
