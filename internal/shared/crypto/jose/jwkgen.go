// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	"crypto/ecdh"
	"crypto/elliptic"
	"fmt"

	cryptoutilKeyGen "cryptoutil/internal/shared/crypto/keygen"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

func GenerateRSAJWKFunction(rsaBits int) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateRSAJWK(rsaBits) }
}

func GenerateRSAJWK(rsaBits int) (joseJwk.Key, error) {
	raw, err := cryptoutilKeyGen.GenerateRSAKeyPair(rsaBits)

	return BuildJWK(KtyRSA, raw.Private, err)
}

func GenerateECDSAJWKFunction(ecdsaCurve elliptic.Curve) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateECDSAJWK(ecdsaCurve) }
}

func GenerateECDSAJWK(ecdsaCurve elliptic.Curve) (joseJwk.Key, error) {
	raw, err := cryptoutilKeyGen.GenerateECDSAKeyPair(ecdsaCurve)

	return BuildJWK(KtyEC, raw.Private, err)
}

func GenerateECDHJWKFunction(ecdhCurve ecdh.Curve) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateECDHJWK(ecdhCurve) }
}

func GenerateECDHJWK(ecdhCurve ecdh.Curve) (joseJwk.Key, error) {
	raw, err := cryptoutilKeyGen.GenerateECDHKeyPair(ecdhCurve)

	return BuildJWK(KtyEC, raw.Private, err)
}

func GenerateEDDSAJWKFunction(edCurve string) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateEDDSAJWK(edCurve) }
}

func GenerateEDDSAJWK(edCurve string) (joseJwk.Key, error) {
	raw, err := cryptoutilKeyGen.GenerateEDDSAKeyPair(edCurve)

	return BuildJWK(KtyOKP, raw.Private, err)
}

func GenerateAESJWKFunction(aesBits int) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateAESJWK(aesBits) }
}

func GenerateAESJWK(aesBits int) (joseJwk.Key, error) {
	aesSecretKeyBytes, err := cryptoutilKeyGen.GenerateAESKey(aesBits)

	return BuildJWK(KtyOCT, []byte(aesSecretKeyBytes), err)
}

func GenerateAESHSJWKFunction(aesHsBits int) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateAESHSJWK(aesHsBits) }
}

func GenerateAESHSJWK(aesHsBits int) (joseJwk.Key, error) {
	aesHsSecretKeyBytes, err := cryptoutilKeyGen.GenerateAESHSKey(aesHsBits)

	return BuildJWK(KtyOCT, []byte(aesHsSecretKeyBytes), err)
}

func GenerateHMACJWKFunction(hmacBits int) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateHMACJWK(hmacBits) }
}

func GenerateHMACJWK(hmacBits int) (joseJwk.Key, error) {
	hmacSecretKeyBytes, err := cryptoutilKeyGen.GenerateHMACKey(hmacBits)

	return BuildJWK(KtyOCT, []byte(hmacSecretKeyBytes), err)
}

func BuildJWK(kty joseJwa.KeyType, raw any, err error) (joseJwk.Key, error) {
	if err != nil {
		return nil, fmt.Errorf("failed to generate %s: %w", kty, err)
	}

	jwk, err := joseJwk.Import(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to import %s: %w", kty, err)
	}

	if err = jwk.Set(joseJwk.KeyTypeKey, kty); err != nil {
		return nil, fmt.Errorf("failed to set 'kty' for %s: %w", kty, err)
	}

	kid, err := googleUuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to create uuid v7 for %s: %w", kty, err)
	}

	if err = jwk.Set(joseJwk.KeyIDKey, kid.String()); err != nil {
		return nil, fmt.Errorf("failed to set `kid` for %s: %w", kty, err)
	}

	return jwk, nil
}
