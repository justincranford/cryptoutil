package unsealrepository

import (
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

// TODO Support different sources of unseal JWKs (e.g. SysInfo, HSM, 3rd-party KMS, secret key sharing, etc)

type UnsealKeyRepository struct {
	unsealJwks []joseJwk.Key
}
