package unsealrepository

import (
	"testing"

	cryptoutilJose "cryptoutil/internal/crypto/jose"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type UnsealKeyRepositoryMock struct {
	mock.Mock
}

func (m *UnsealKeyRepositoryMock) UnsealJwks() []joseJwk.Key {
	args := m.Called()
	return args.Get(0).([]joseJwk.Key)
}

func NewUnsealKeyRepositoryMock(t *testing.T, numUnsealJwks int) (*UnsealKeyRepositoryMock, []joseJwk.Key, error) {
	unsealKeys := make([]joseJwk.Key, 0, numUnsealJwks)
	for _ = range numUnsealJwks {
		unsealJwk, _, err := cryptoutilJose.GenerateAesJWK(cryptoutilJose.AlgA256GCMKW)
		assert.NoError(t, err)
		unsealKeys = append(unsealKeys, unsealJwk)
	}
	mockUnsealKeyRepository := &UnsealKeyRepositoryMock{}
	mockUnsealKeyRepository.On("UnsealJwks").Return(unsealKeys)
	return mockUnsealKeyRepository, unsealKeys, nil
}
