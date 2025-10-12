package unsealkeysservice

import (
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

// RequireNewSimpleForTest no validation.
func RequireNewSimpleForTest(unsealJWKs []joseJwk.Key) UnsealKeysService {
	return &UnsealKeysServiceSimple{unsealJWKs: unsealJWKs}
}
