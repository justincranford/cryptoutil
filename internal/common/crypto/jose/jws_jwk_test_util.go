package jose

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

func GenerateJwsJwksForTest(t *testing.T, count int, alg *joseJwa.SignatureAlgorithm) ([]joseJwk.Key, error) {
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
			_, jwk, _, err := GenerateJwsJwkForAlg(alg)
			jwkOrErrs <- jwkOrErr{key: jwk, err: err}
		}()
	}
	wg.Wait()
	close(jwkOrErrs)

	jwks := make([]joseJwk.Key, 0, count)
	errs := make([]error, 0, count)
	for res := range jwkOrErrs {
		if res.err != nil {
			errs = append(errs, res.err)
		} else {
			jwks = append(jwks, res.key)
		}
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("unexpected %d errors: %w", len(errs), errors.Join(errs...))
	}
	return jwks, nil
}
