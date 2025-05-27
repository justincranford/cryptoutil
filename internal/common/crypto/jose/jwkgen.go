package jose

import (
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"

	"github.com/cloudflare/circl/sign/ed448"

	cryptoutilKeygen "cryptoutil/internal/common/crypto/keygen"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

func GenerateRSAJwkFunction(rsaBits int) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateRSAJwk(rsaBits) }
}

func GenerateRSAJwk(rsaBits int) (joseJwk.Key, error) {
	raw, err := rsa.GenerateKey(rand.Reader, rsaBits)
	return buildJwk(KtyRSA, raw, err)
}

func GenerateECDSAJwkFunction(ecdsaCurve elliptic.Curve) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateECDSAJwk(ecdsaCurve) }
}

func GenerateECDSAJwk(ecdsaCurve elliptic.Curve) (joseJwk.Key, error) {
	raw, err := ecdsa.GenerateKey(ecdsaCurve, rand.Reader)
	return buildJwk(KtyEC, raw, err)
}

func GenerateECDHJwkFunction(ecdhCurve ecdh.Curve) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateECDHJwk(ecdhCurve) }
}

func GenerateECDHJwk(ecdhCurve ecdh.Curve) (joseJwk.Key, error) {
	raw, err := ecdhCurve.GenerateKey(rand.Reader)
	return buildJwk(KtyEC, raw, err)
}

func GenerateEDDSAJwkFunction(edCurve string) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateEDDSAJwk(edCurve) }
}

func GenerateEDDSAJwk(edCurve string) (joseJwk.Key, error) {
	switch edCurve {
	case "Ed448":
		_, raw, err := ed448.GenerateKey(rand.Reader)
		return buildJwk(KtyOKP, raw, err)
	case "Ed25519":
		_, raw, err := ed25519.GenerateKey(rand.Reader)
		return buildJwk(KtyOKP, raw, err)
	default:
		return nil, errors.New("unsupported Ed curve")
	}
}

func GenerateAESJwkFunction(aesBits int) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateAESJwk(aesBits) }
}

func GenerateAESJwk(aesBits int) (joseJwk.Key, error) {
	raw, err := cryptoutilKeygen.GenerateAESKey(aesBits)
	return buildJwk(KtyOCT, raw, err)
}

func GenerateAESHSJwkFunction(aesHsBits int) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateAESHSJwk(aesHsBits) }
}

func GenerateAESHSJwk(aesHsBits int) (joseJwk.Key, error) {
	raw, err := cryptoutilKeygen.GenerateAESHSKey(aesHsBits)
	return buildJwk(KtyOCT, raw, err)
}

func GenerateHMACJwkFunction(hmacBits int) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateHMACJwk(hmacBits) }
}

func GenerateHMACJwk(hmacBits int) (joseJwk.Key, error) {
	raw, err := cryptoutilKeygen.GenerateHMACKey(hmacBits)
	return buildJwk(KtyOCT, raw, err)
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
