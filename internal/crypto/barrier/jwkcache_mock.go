package barrier

import (
	"cryptoutil/internal/crypto/jose"

	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

func mockJWKKey() (googleUuid.UUID, joseJwk.Key) {
	aesJwk, _, _ := jose.GenerateAesJWK(jose.AlgA256GCMKW)

	var kid string
	_ = aesJwk.Get(joseJwk.KeyIDKey, &kid)

	uuid, _ := googleUuid.Parse(kid)

	return uuid, aesJwk
}
