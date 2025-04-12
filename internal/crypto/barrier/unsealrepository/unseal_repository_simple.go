package unsealrepository

import (
	"fmt"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

type UnsealRepositorySimple struct {
	unsealJwks []joseJwk.Key
}

func (u *UnsealRepositorySimple) UnsealJwks() []joseJwk.Key {
	return u.unsealJwks
}

func NewUnsealRepositorySimple(unsealJwks []joseJwk.Key) (UnsealRepository, error) {
	if unsealJwks == nil {
		return nil, fmt.Errorf("unsealJwks can't be nil")
	} else if len(unsealJwks) == 0 {
		return nil, fmt.Errorf("unsealJwks can't be empty")
	}
	return &UnsealRepositorySimple{unsealJwks: unsealJwks}, nil
}
