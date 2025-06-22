package jose

import (
	"crypto/ecdh"
	"crypto/elliptic"
	"fmt"

	cryptoutilKeyGen "cryptoutil/internal/common/crypto/keygen"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

func GenerateRSAJwkFunction(rsaBits int) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateRSAJwk(rsaBits) }
}

func GenerateRSAJwk(rsaBits int) (joseJwk.Key, error) {
	raw, err := cryptoutilKeyGen.GenerateRSAKeyPair(rsaBits)
	return buildJwk(KtyRSA, raw.Private, err)
}

func GenerateECDSAJwkFunction(ecdsaCurve elliptic.Curve) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateECDSAJwk(ecdsaCurve) }
}

func GenerateECDSAJwk(ecdsaCurve elliptic.Curve) (joseJwk.Key, error) {
	raw, err := cryptoutilKeyGen.GenerateECDSAKeyPair(ecdsaCurve)
	return buildJwk(KtyEC, raw.Private, err)
}

func GenerateECDHJwkFunction(ecdhCurve ecdh.Curve) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateECDHJwk(ecdhCurve) }
}

func GenerateECDHJwk(ecdhCurve ecdh.Curve) (joseJwk.Key, error) {
	raw, err := cryptoutilKeyGen.GenerateECDHKeyPair(ecdhCurve)
	return buildJwk(KtyEC, raw.Private, err)
}

func GenerateEDDSAJwkFunction(edCurve string) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateEDDSAJwk(edCurve) }
}

func GenerateEDDSAJwk(edCurve string) (joseJwk.Key, error) {
	raw, err := cryptoutilKeyGen.GenerateEDDSAKeyPair(edCurve)
	return buildJwk(KtyOKP, raw.Private, err)
}

func GenerateAESJwkFunction(aesBits int) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateAESJwk(aesBits) }
}

func GenerateAESJwk(aesBits int) (joseJwk.Key, error) {
	aesSecretKeyBytes, err := cryptoutilKeyGen.GenerateAESKey(aesBits)
	return buildJwk(KtyOCT, aesSecretKeyBytes, err)
}

func GenerateAESHSJwkFunction(aesHsBits int) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateAESHSJwk(aesHsBits) }
}

func GenerateAESHSJwk(aesHsBits int) (joseJwk.Key, error) {
	aesHsSecretKeyBytes, err := cryptoutilKeyGen.GenerateAESHSKey(aesHsBits)
	return buildJwk(KtyOCT, aesHsSecretKeyBytes, err)
}

func GenerateHMACJwkFunction(hmacBits int) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateHMACJwk(hmacBits) }
}

func GenerateHMACJwk(hmacBits int) (joseJwk.Key, error) {
	hmacSecretKeyBytes, err := cryptoutilKeyGen.GenerateHMACKey(hmacBits)
	return buildJwk(KtyOCT, hmacSecretKeyBytes, err)
}

func buildJwk(kty joseJwa.KeyType, raw any, err error) (joseJwk.Key, error) {
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
