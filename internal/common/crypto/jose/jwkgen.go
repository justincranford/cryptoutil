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

	"cryptoutil/internal/common/crypto/keygen"

	"github.com/cloudflare/circl/sign/ed448"
	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

func GenerateRSAJwk(rsaBits int) (joseJwk.Key, error) {
	raw, err := rsa.GenerateKey(rand.Reader, rsaBits)
	return buildJwk(KtyRsa, raw, err)
}

func GenerateECDSAJwk(ecdsaCurve elliptic.Curve) (joseJwk.Key, error) {
	raw, err := ecdsa.GenerateKey(ecdsaCurve, rand.Reader)
	return buildJwk(KtyEC, raw, err)
}

func GenerateECDHJwk(ecdhCurve ecdh.Curve) (joseJwk.Key, error) {
	raw, err := ecdhCurve.GenerateKey(rand.Reader)
	return buildJwk(KtyEC, raw, err)
}

func GenerateEDDSAJwk(edCurve string) (joseJwk.Key, error) {
	switch edCurve {
	case "Ed448":
		_, raw, err := ed448.GenerateKey(rand.Reader)
		return buildJwk(KtyOkp, raw, err)
	case "Ed25519":
		_, raw, err := ed25519.GenerateKey(rand.Reader)
		return buildJwk(KtyOkp, raw, err)
	default:
		return nil, errors.New("unsupported Ed curve")
	}
}

func GenerateAESJwk(aesBits int) (joseJwk.Key, error) {
	if aesBits != 128 && aesBits != 192 && aesBits != 256 {
		return nil, fmt.Errorf("invalid AES key size: %d (must be 128, 192, or 256 bits)", aesBits)
	}
	raw, err := keygen.GenerateBytes(aesBits / 8)
	return buildJwk(KtyOct, raw, err)
}

func GenerateAESHSJwk(aesHsBits int) (joseJwk.Key, error) {
	if aesHsBits != 256 && aesHsBits != 384 && aesHsBits != 512 {
		return nil, fmt.Errorf("invalid AES HAMC-SHA2 key size: %d (must be 256, 384, or 512 bits)", aesHsBits)
	}
	raw, err := keygen.GenerateBytes(aesHsBits / 8)
	return buildJwk(KtyOct, raw, err)
}

func GenerateHMACJwk(hmacBits int) (joseJwk.Key, error) {
	if hmacBits < 256 {
		return nil, fmt.Errorf("invalid HMAC key size: %d (must be 256 bits or higher)", hmacBits)
	}
	raw, err := keygen.GenerateBytes(hmacBits / 8)
	return buildJwk(KtyOct, raw, err)
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
