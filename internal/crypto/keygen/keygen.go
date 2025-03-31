package keygen

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
	googleUuid "github.com/google/uuid"
)

type Key struct {
	Private any
	Public  any
}

func GenerateRSAKeyPairFunction(rsaBits int) func() (Key, error) {
	return func() (Key, error) { return GenerateRSAKeyPair(rsaBits) }
}

func GenerateECDSAKeyPairFunction(ecdsaCurve elliptic.Curve) func() (Key, error) {
	return func() (Key, error) { return GenerateECDSAKeyPair(ecdsaCurve) }
}

func GenerateECDHKeyPairFunction(ecdhCurve ecdh.Curve) func() (Key, error) {
	return func() (Key, error) { return GenerateECDHKeyPair(ecdhCurve) }
}

func GenerateEDKeyPairFunction(edCurve string) func() (Key, error) {
	return func() (Key, error) { return GenerateEDKeyPair(edCurve) }
}

func GenerateAESKeyFunction(aesBits int) func() (Key, error) {
	return func() (Key, error) { return GenerateAESKey(aesBits) }
}

func GenerateHMACKeyFunction(hmacBits int) func() (Key, error) {
	return func() (Key, error) { return GenerateHMACKey(hmacBits) }
}

func GenerateUUIDv7Function() func() (Key, error) {
	return func() (Key, error) { return GenerateUUIDv7() }
}

func GenerateRSAKeyPair(rsaBits int) (Key, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if err != nil {
		return Key{}, fmt.Errorf("generate RSA key pair failed: %w", err)
	}
	return Key{Private: privateKey, Public: &privateKey.PublicKey}, nil
}

func GenerateECDSAKeyPair(ecdsaCurve elliptic.Curve) (Key, error) {
	privateKey, err := ecdsa.GenerateKey(ecdsaCurve, rand.Reader)
	if err != nil {
		return Key{}, fmt.Errorf("generate ECDSA key pair failed: %w", err)
	}
	return Key{Private: privateKey, Public: &privateKey.PublicKey}, nil
}

func GenerateECDHKeyPair(ecdhCurve ecdh.Curve) (Key, error) {
	privateKey, err := ecdhCurve.GenerateKey(rand.Reader)
	if err != nil {
		return Key{}, fmt.Errorf("generate ECDH key pair failed: %w", err)
	}
	return Key{Private: privateKey, Public: privateKey.PublicKey()}, nil
}

func GenerateEDKeyPair(edCurve string) (Key, error) {
	switch edCurve {
	case "Ed448":
		publicKey, privateKey, err := ed448.GenerateKey(rand.Reader)
		if err != nil {
			return Key{}, fmt.Errorf("generate Ed448 key pair failed: %w", err)
		}
		return Key{Private: privateKey, Public: publicKey}, nil
	case "Ed25519":
		publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return Key{}, fmt.Errorf("generate Ed25519 key pair failed: %w", err)
		}
		return Key{Private: privateKey, Public: publicKey}, nil
	default:
		return Key{}, errors.New("unsupported Ed curve")
	}
}

func GenerateAESKey(aesBits int) (Key, error) {
	if aesBits != 128 && aesBits != 192 && aesBits != 256 {
		return Key{}, fmt.Errorf("invalid AES key size: %d (must be 128, 192, or 256 bits)", aesBits)
	}
	key := make([]byte, aesBits/8)
	_, err := rand.Read(key)
	if err != nil {
		return Key{}, fmt.Errorf("generate AES %d key failed: %w", aesBits, err)
	}
	return Key{Private: key}, nil
}

func GenerateHMACKey(hmacBits int) (Key, error) {
	if hmacBits < 256 {
		return Key{}, fmt.Errorf("invalid HMAC key size: %d (must be 256 bits or higher)", hmacBits)
	}
	key := make([]byte, hmacBits/8)
	_, err := rand.Read(key)
	if err != nil {
		return Key{}, fmt.Errorf("generate HMAC %d key failed: %w", hmacBits, err)
	}
	return Key{Private: key}, nil
}

func GenerateUUIDv7() (Key, error) {
	uuid, err := googleUuid.NewV7()
	if err != nil {
		return Key{}, fmt.Errorf("failed to generate UUID: %w", err)
	}
	return Key{Private: uuid}, nil
}
