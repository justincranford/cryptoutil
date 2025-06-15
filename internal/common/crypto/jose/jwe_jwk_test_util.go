package jose

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

func GenerateJweJwksForTest(t *testing.T, count int, enc *joseJwa.ContentEncryptionAlgorithm, alg *joseJwa.KeyEncryptionAlgorithm) ([]joseJwk.Key, []joseJwk.Key, error) {
	type jwkOrErr struct {
		privateOrSecretJwk joseJwk.Key
		publicJwk          joseJwk.Key
		err                error
	}

	jwkOrErrs := make(chan jwkOrErr, count)
	var wg sync.WaitGroup
	for range count {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, privateOrSecretJwk, publicJwk, _, _, err := GenerateJweJwkForEncAndAlg(enc, alg)
			jwkOrErrs <- jwkOrErr{privateOrSecretJwk: privateOrSecretJwk, publicJwk: publicJwk, err: err}
		}()
	}
	wg.Wait()
	close(jwkOrErrs)

	privateOrSecretJwks := make([]joseJwk.Key, 0, count)
	publicJwks := make([]joseJwk.Key, 0, count)
	errs := make([]error, 0, count)
	for res := range jwkOrErrs {
		if res.err != nil {
			errs = append(errs, res.err)
		} else {
			privateOrSecretJwks = append(privateOrSecretJwks, res.privateOrSecretJwk)
			publicJwks = append(publicJwks, res.publicJwk)
		}
	}
	if len(errs) > 0 {
		return nil, nil, fmt.Errorf("unexpected %d errors: %w", len(errs), errors.Join(errs...))
	}
	return privateOrSecretJwks, publicJwks, nil
}
