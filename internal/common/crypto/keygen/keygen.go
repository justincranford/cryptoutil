package keygen

import (
	"crypto"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"

	cryptoutilUtil "cryptoutil/internal/common/util"

	"github.com/cloudflare/circl/sign/ed448"
)

type KeyPair struct {
	Private crypto.PrivateKey
	Public  crypto.PublicKey
}
type SecretKey []byte

type Key interface { // &KeyPair or SecretKey
	isKey()
}

func (k *KeyPair) isKey()  {}
func (s SecretKey) isKey() {}

const (
	EdCurveEd448   = "Ed448"
	EdCurveEd25519 = "Ed25519"
	ECCurveP256    = "P256"
	ECCurveP384    = "P384"
	ECCurveP521    = "P521"
)

func GenerateRSAKeyPairFunction(rsaBits int) func() (*KeyPair, error) {
	return func() (*KeyPair, error) { return GenerateRSAKeyPair(rsaBits) }
}

func GenerateRSAKeyPair(rsaBits int) (*KeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if err != nil {
		return nil, fmt.Errorf("generate RSA key pair failed: %w", err)
	}

	return &KeyPair{Private: privateKey, Public: &privateKey.PublicKey}, nil
}

func GenerateECDSAKeyPairFunction(ecdsaCurve elliptic.Curve) func() (*KeyPair, error) {
	return func() (*KeyPair, error) { return GenerateECDSAKeyPair(ecdsaCurve) }
}

func GenerateECDSAKeyPair(ecdsaCurve elliptic.Curve) (*KeyPair, error) {
	privateKey, err := ecdsa.GenerateKey(ecdsaCurve, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate ECDSA key pair failed: %w", err)
	}

	return &KeyPair{Private: privateKey, Public: &privateKey.PublicKey}, nil
}

func GenerateECDHKeyPairFunction(ecdhCurve ecdh.Curve) func() (*KeyPair, error) {
	return func() (*KeyPair, error) { return GenerateECDHKeyPair(ecdhCurve) }
}

func GenerateECDHKeyPair(ecdhCurve ecdh.Curve) (*KeyPair, error) {
	privateKey, err := ecdhCurve.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate ECDH key pair failed: %w", err)
	}

	return &KeyPair{Private: privateKey, Public: privateKey.PublicKey()}, nil
}

func GenerateEDDSAKeyPairFunction(edCurve string) func() (*KeyPair, error) {
	return func() (*KeyPair, error) { return GenerateEDDSAKeyPair(edCurve) }
}

func GenerateEDDSAKeyPair(edCurve string) (*KeyPair, error) {
	switch edCurve {
	case EdCurveEd448:
		publicKey, privateKey, err := ed448.GenerateKey(rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("generate Ed448 key pair failed: %w", err)
		}

		return &KeyPair{Private: privateKey, Public: publicKey}, nil
	case EdCurveEd25519:
		publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("generate Ed25519 key pair failed: %w", err)
		}

		return &KeyPair{Private: privateKey, Public: publicKey}, nil
	default:
		return nil, errors.New("unsupported Ed curve")
	}
}

func GenerateAESKeyFunction(aesBits int) func() (SecretKey, error) {
	return func() (SecretKey, error) { return GenerateAESKey(aesBits) }
}

func GenerateAESKey(aesBits int) (SecretKey, error) {
	if aesBits != 128 && aesBits != 192 && aesBits != 256 {
		return nil, fmt.Errorf("invalid AES key size: %d (must be 128, 192, or 256 bits)", aesBits)
	}

	aesSecretKeyBytes, err := cryptoutilUtil.GenerateBytes(aesBits / 8)
	if err != nil {
		return nil, fmt.Errorf("generate AES %d key failed: %w", aesBits, err)
	}

	return aesSecretKeyBytes, nil
}

func GenerateAESHSKeyFunction(aesHsBits int) func() (SecretKey, error) {
	return func() (SecretKey, error) { return GenerateAESHSKey(aesHsBits) }
}

func GenerateAESHSKey(aesHsBits int) (SecretKey, error) {
	if aesHsBits != 256 && aesHsBits != 384 && aesHsBits != 512 {
		return nil, fmt.Errorf("invalid AES HAMC-SHA2 key size: %d (must be 256, 384, or 512 bits)", aesHsBits)
	}

	aesHsSecretKeyBytes, err := cryptoutilUtil.GenerateBytes(aesHsBits / 8)
	if err != nil {
		return nil, fmt.Errorf("generate AES HAMC-SHA2 %d key failed: %w", aesHsBits, err)
	}

	return aesHsSecretKeyBytes, nil
}

func GenerateHMACKeyFunction(hmacBits int) func() (SecretKey, error) {
	return func() (SecretKey, error) { return GenerateHMACKey(hmacBits) }
}

func GenerateHMACKey(hmacBits int) (SecretKey, error) {
	if hmacBits < 256 {
		return nil, fmt.Errorf("invalid HMAC key size: %d (must be 256 bits or higher)", hmacBits)
	}

	hmacSecretKeyBytes, err := cryptoutilUtil.GenerateBytes(hmacBits / 8)
	if err != nil {
		return nil, fmt.Errorf("generate HMAC %d key failed: %w", hmacBits, err)
	}

	return hmacSecretKeyBytes, nil
}
