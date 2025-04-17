package unsealrepository

import (
	"crypto/rand"
	"fmt"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

type UnsealRepositorySharedSecrets struct {
	unsealJwks []joseJwk.Key
}

func (u *UnsealRepositorySharedSecrets) UnsealJwks() []joseJwk.Key {
	return u.unsealJwks
}

func (u *UnsealRepositorySharedSecrets) Shutdown() {
	u.unsealJwks = nil
}

func NewUnsealRepositorySharedSecrets(m, chooseN, secretBytesLength int) (UnsealRepository, error) {
	if m == 0 {
		return nil, fmt.Errorf("m can't be zero")
	} else if m < 0 {
		return nil, fmt.Errorf("m can't be negative")
	} else if m >= 255 {
		return nil, fmt.Errorf("m can't be greater than 255")
	} else if chooseN == 0 {
		return nil, fmt.Errorf("n can't be zero")
	} else if chooseN < 0 {
		return nil, fmt.Errorf("n can't be negative")
	} else if chooseN > m {
		return nil, fmt.Errorf("n can't be greater than m")
	} else if secretBytesLength < 32 {
		return nil, fmt.Errorf("secretBytesLength can't be greater than 32")
	} else if secretBytesLength > 64 {
		return nil, fmt.Errorf("secretBytesLength can't be greater than 64")
	}

	sharedSecrets := make([][]byte, m)
	for i := range m {
		sharedSecrets[i] = make([]byte, secretBytesLength)
		if _, err := rand.Read(sharedSecrets[i]); err != nil {
			return nil, fmt.Errorf("failed to generate shared secret: %w", err)
		}
	}

	unsealJwks, err := deriveJwksFromMChooseNCombinations(sharedSecrets, chooseN)
	if err != nil {
		return nil, fmt.Errorf("failed to create unseal JWKs: %w", err)
	}
	return &UnsealRepositorySharedSecrets{unsealJwks: unsealJwks}, nil
}
