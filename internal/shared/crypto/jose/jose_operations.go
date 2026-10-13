// Copyright (c) 2025-2026 Justin Cranford.
//
//

package crypto

import (
	json "encoding/json"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

// joseOperations contains per-call JOSE library dependencies that tests may replace.
type joseOperations struct {
	setJWK       func(joseJwk.Key, string, any) error
	importJWK    func(any) (joseJwk.Key, error)
	marshalJSON  func(any) ([]byte, error)
	publicKeyJWK func(joseJwk.Key) (joseJwk.Key, error)
}

func newJOSEOperations() joseOperations {
	return joseOperations{
		setJWK:       func(key joseJwk.Key, name string, value any) error { return key.Set(name, value) },
		importJWK:    joseJwk.Import,
		marshalJSON:  json.Marshal,
		publicKeyJWK: func(key joseJwk.Key) (joseJwk.Key, error) { return key.PublicKey() },
	}
}
