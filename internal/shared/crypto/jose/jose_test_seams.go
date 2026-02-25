// Copyright (c) 2025 Justin Cranford
//
//

package crypto

// Test seams: package-level function variables that wrap library calls.
// Production code calls these instead of library functions directly.
// Tests can override these to inject errors for otherwise-unreachable error paths.

import (
json "encoding/json"

cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"

joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
joseJws "github.com/lestrrat-go/jwx/v3/jws"
)

// Category 1: jwk.Set() — wraps key.Set(name, value).
var jwkKeySet = func(key joseJwk.Key, name string, value interface{}) error {
return key.Set(name, value)
}

// Category 2: joseJwk.Import() — wraps jwk.Import(raw).
var jwkImport = func(raw any) (joseJwk.Key, error) {
return joseJwk.Import(raw)
}

// Category 3: json.Marshal() — wraps json.Marshal(v).
var jsonMarshalFunc = json.Marshal

// Category 4: key.PublicKey() — wraps key.PublicKey().
var jwkPublicKey = func(key joseJwk.Key) (joseJwk.Key, error) {
return key.PublicKey()
}

// Category 6: Key generation — wraps keygen functions.
var (
generateRSAKeyPair   = cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair
generateECDSAKeyPair = cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair
generateEDDSAKeyPair = cryptoutilSharedCryptoKeygen.GenerateEDDSAKeyPair
generateHMACKey      = cryptoutilSharedCryptoKeygen.GenerateHMACKey
generateAESKey       = cryptoutilSharedCryptoKeygen.GenerateAESKey
)

// Category 8: Encrypt/Sign/Parse/Decrypt — wraps JWE/JWS operations.
var (
jweEncryptFunc = joseJwe.Encrypt
jweParseFunc   = joseJwe.Parse
jweDecryptFunc = joseJwe.Decrypt
jwsSignFunc    = joseJws.Sign
jwsParseFunc   = joseJws.Parse
)

// Category 9: jwk set Add — wraps set.AddKey().
var jwkSetAddKey = func(set joseJwk.Set, key joseJwk.Key) error {
return set.AddKey(key)
}
