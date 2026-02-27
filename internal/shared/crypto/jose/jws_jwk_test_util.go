// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

// GenerateJWSJWKsForTest generates multiple JWS JWKs for testing.
func GenerateJWSJWKsForTest(t *testing.T, count int, alg *joseJwa.SignatureAlgorithm) ([]joseJwk.Key, []joseJwk.Key, error) {
	t.Helper()

	type jwkOrErr struct {
		nonPublicJWK joseJwk.Key
		publicJWK    joseJwk.Key
		err          error
	}

	jwkOrErrs := make(chan jwkOrErr, count)

	var wg sync.WaitGroup
	for range count {
		wg.Add(1)

		go func() {
			defer wg.Done()

			_, nonPublicJWK, publicJWK, _, _, err := GenerateJWSJWKForAlg(alg)
			jwkOrErrs <- jwkOrErr{nonPublicJWK: nonPublicJWK, publicJWK: publicJWK, err: err}
		}()
	}

	wg.Wait()
	close(jwkOrErrs) //nolint:errcheck

	nonPublicJWKs := make([]joseJwk.Key, 0, count)
	publicJWKs := make([]joseJwk.Key, 0, count)
	errs := make([]error, 0, count)

	for res := range jwkOrErrs {
		if res.err != nil {
			errs = append(errs, res.err)
		} else {
			nonPublicJWKs = append(nonPublicJWKs, res.nonPublicJWK)
			publicJWKs = append(publicJWKs, res.publicJWK)
		}
	}

	if len(errs) > 0 {
		return nil, nil, fmt.Errorf("unexpected %d errors: %w", len(errs), errors.Join(errs...))
	}

	return nonPublicJWKs, publicJWKs, nil
}
