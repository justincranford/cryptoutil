// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	"crypto/ecdh"
	"crypto/elliptic"
	"fmt"

	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

// GenerateRSAJWKFunction returns a function that generates RSA JWKs with the specified bit size.
func GenerateRSAJWKFunction(rsaBits int) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateRSAJWK(rsaBits) }
}

// GenerateRSAJWK generates an RSA JWK with the specified bit size.
func GenerateRSAJWK(rsaBits int) (joseJwk.Key, error) {
	raw, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(rsaBits)

	return BuildJWK(KtyRSA, raw.Private, err)
}

// GenerateECDSAJWKFunction returns a function that generates ECDSA JWKs with the specified curve.
func GenerateECDSAJWKFunction(ecdsaCurve elliptic.Curve) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateECDSAJWK(ecdsaCurve) }
}

// GenerateECDSAJWK generates an ECDSA JWK with the specified curve.
func GenerateECDSAJWK(ecdsaCurve elliptic.Curve) (joseJwk.Key, error) {
	raw, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(ecdsaCurve)

	return BuildJWK(KtyEC, raw.Private, err)
}

// GenerateECDHJWKFunction returns a function that generates ECDH JWKs with the specified curve.
func GenerateECDHJWKFunction(ecdhCurve ecdh.Curve) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateECDHJWK(ecdhCurve) }
}

// GenerateECDHJWK generates an ECDH JWK with the specified curve.
func GenerateECDHJWK(ecdhCurve ecdh.Curve) (joseJwk.Key, error) {
	raw, err := cryptoutilSharedCryptoKeygen.GenerateECDHKeyPair(ecdhCurve)

	return BuildJWK(KtyEC, raw.Private, err)
}

// GenerateEDDSAJWKFunction returns a function that generates EdDSA JWKs with the specified curve.
func GenerateEDDSAJWKFunction(edCurve string) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateEDDSAJWK(edCurve) }
}

// GenerateEDDSAJWK generates an EdDSA JWK with the specified curve.
func GenerateEDDSAJWK(edCurve string) (joseJwk.Key, error) {
	raw, err := cryptoutilSharedCryptoKeygen.GenerateEDDSAKeyPair(edCurve)

	return BuildJWK(KtyOKP, raw.Private, err)
}

// GenerateAESJWKFunction returns a function that generates AES JWKs with the specified bit size.
func GenerateAESJWKFunction(aesBits int) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateAESJWK(aesBits) }
}

// GenerateAESJWK generates an AES JWK with the specified bit size.
func GenerateAESJWK(aesBits int) (joseJwk.Key, error) {
	aesSecretKeyBytes, err := cryptoutilSharedCryptoKeygen.GenerateAESKey(aesBits)

	return BuildJWK(KtyOCT, []byte(aesSecretKeyBytes), err)
}

// GenerateAESHSJWKFunction returns a function that generates AES-HS JWKs with the specified bit size.
func GenerateAESHSJWKFunction(aesHsBits int) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateAESHSJWK(aesHsBits) }
}

// GenerateAESHSJWK generates an AES-HS JWK with the specified bit size.
func GenerateAESHSJWK(aesHsBits int) (joseJwk.Key, error) {
	aesHsSecretKeyBytes, err := cryptoutilSharedCryptoKeygen.GenerateAESHSKey(aesHsBits)

	return BuildJWK(KtyOCT, []byte(aesHsSecretKeyBytes), err)
}

// GenerateHMACJWKFunction returns a function that generates HMAC JWKs with the specified bit size.
func GenerateHMACJWKFunction(hmacBits int) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateHMACJWK(hmacBits) }
}

// GenerateHMACJWK generates an HMAC JWK with the specified bit size.
func GenerateHMACJWK(hmacBits int) (joseJwk.Key, error) {
	hmacSecretKeyBytes, err := cryptoutilSharedCryptoKeygen.GenerateHMACKey(hmacBits)

	return BuildJWK(KtyOCT, []byte(hmacSecretKeyBytes), err)
}

// BuildJWK builds a JWK from raw key material.
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
