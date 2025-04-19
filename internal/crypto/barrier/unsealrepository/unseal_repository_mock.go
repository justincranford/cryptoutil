package unsealrepository

import (
	"testing"

	cryptoutilJose "cryptoutil/internal/crypto/jose"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type UnsealRepositoryMock struct {
	mock.Mock
}

func (m *UnsealRepositoryMock) unsealJwks() []joseJwk.Key {
	args := m.Called()
	return args.Get(0).([]joseJwk.Key)
}

func (u *UnsealRepositoryMock) EncryptKey(clearRootKey joseJwk.Key) ([]byte, error) {
	return encryptKey(u.unsealJwks(), clearRootKey)
}

func (u *UnsealRepositoryMock) DecryptKey(encryptedRootKeyBytes []byte) (joseJwk.Key, error) {
	return decryptKey(u.unsealJwks(), encryptedRootKeyBytes)
}

func (u *UnsealRepositoryMock) Shutdown() {
}

func NewUnsealRepositoryMock(t *testing.T, numUnsealJwks int) (*UnsealRepositoryMock, []joseJwk.Key, error) {
	unsealKeys := make([]joseJwk.Key, 0, numUnsealJwks)
	for range numUnsealJwks {
		unsealJwk, _, _, err := cryptoutilJose.GenerateAesJWK(cryptoutilJose.AlgA256GCMKW)
		assert.NoError(t, err)
		unsealKeys = append(unsealKeys, unsealJwk)
	}
	mockUnsealRepository := &UnsealRepositoryMock{}
	mockUnsealRepository.On("unsealJwks").Return(unsealKeys)
	return mockUnsealRepository, unsealKeys, nil
}
