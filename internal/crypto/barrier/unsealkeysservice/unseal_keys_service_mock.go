package unsealkeysservice

import (
	"testing"

	cryptoutilJose "cryptoutil/internal/crypto/jose"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
)

type UnsealKeysServiceMock struct {
	mock.Mock
}

func (m *UnsealKeysServiceMock) unsealJwks() []joseJwk.Key {
	args := m.Called()
	return args.Get(0).([]joseJwk.Key)
}

func (u *UnsealKeysServiceMock) EncryptKey(clearRootKey joseJwk.Key) ([]byte, error) {
	return encryptKey(u.unsealJwks(), clearRootKey)
}

func (u *UnsealKeysServiceMock) DecryptKey(encryptedRootKeyBytes []byte) (joseJwk.Key, error) {
	return decryptKey(u.unsealJwks(), encryptedRootKeyBytes)
}

func (u *UnsealKeysServiceMock) Shutdown() {
}

func NewUnsealKeysServiceMock(t *testing.T, numUnsealJwks int) (*UnsealKeysServiceMock, []joseJwk.Key, error) {
	unsealKeys := make([]joseJwk.Key, 0, numUnsealJwks)
	for range numUnsealJwks {
		unsealJwk, _, _, err := cryptoutilJose.GenerateAesJWK(cryptoutilJose.AlgA256GCMKW)
		require.NoError(t, err)
		unsealKeys = append(unsealKeys, unsealJwk)
	}
	mockUnsealKeysService := &UnsealKeysServiceMock{}
	mockUnsealKeysService.On("unsealJwks").Return(unsealKeys)
	return mockUnsealKeysService, unsealKeys, nil
}
