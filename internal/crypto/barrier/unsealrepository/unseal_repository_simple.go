package unsealrepository

import (
	"fmt"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

type UnsealKeyRepositorySimple struct {
	unsealJwks []joseJwk.Key
}

func (u *UnsealKeyRepositorySimple) UnsealJwks() []joseJwk.Key {
	return u.unsealJwks
}

func NewUnsealKeyRepositorySimple(unsealJwks []joseJwk.Key) (UnsealKeyRepository, error) {
	if unsealJwks == nil {
		return nil, fmt.Errorf("unsealJwks can't be nil")
	} else if len(unsealJwks) == 0 {
		return nil, fmt.Errorf("unsealJwks can't be empty")
	}
	return &UnsealKeyRepositorySimple{unsealJwks: unsealJwks}, nil
}
