package unsealkeysservice

import (
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

// RequireNewSimpleForTest no validation
func RequireNewSimpleForTest(unsealJwks []joseJwk.Key) UnsealKeysService {
	return &UnsealKeysServiceSimple{unsealJwks: unsealJwks}
}
