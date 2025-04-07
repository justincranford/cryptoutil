package barrierpersistence

import (
	"cryptoutil/internal/crypto/jose"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

func mockJWKKey() joseJwk.Key {
	aesJwk, _, _ := jose.GenerateAesJWK(jose.AlgA256GCMKW)
	return aesJwk
}
