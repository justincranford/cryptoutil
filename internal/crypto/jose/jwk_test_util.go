package jose

import (
	"sync"
	"testing"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

func GenerateAes256KeysForTest(t *testing.T, count int, kekAlg joseJwa.KeyEncryptionAlgorithm) []joseJwk.Key {
	type jwkOrErr struct {
		key joseJwk.Key
		err error
	}

	jwkOrErrs := make(chan jwkOrErr, count)
	var wg sync.WaitGroup
	for range count {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, jwk, _, err := GenerateAesJWK(kekAlg)
			jwkOrErrs <- jwkOrErr{key: jwk, err: err}
		}()
	}
	wg.Wait()
	close(jwkOrErrs)

	jwks := make([]joseJwk.Key, 0, count)
	for res := range jwkOrErrs {
		require.NoError(t, res.err, "Expected no error")
		jwks = append(jwks, res.key)
	}
	return jwks
}
