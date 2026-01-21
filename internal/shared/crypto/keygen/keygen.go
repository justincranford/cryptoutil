// Copyright (c) 2025 Justin Cranford
//
//

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

	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilRandom "cryptoutil/internal/shared/util/random"

	"github.com/cloudflare/circl/sign/ed448"
)

// KeyPair represents an asymmetric key pair with private and public keys.
type KeyPair struct {
	Private crypto.PrivateKey
	Public  crypto.PublicKey
}

// SecretKey represents a symmetric secret key as a byte slice.
type SecretKey []byte

// Key is an interface that represents either a KeyPair or SecretKey.
type Key interface { // &KeyPair or SecretKey
	isKey()
}

func (k *KeyPair) isKey()  {}
func (s SecretKey) isKey() {}

// Elliptic curve and key size constants.
const (
	// EdCurveEd448 is the Ed448 elliptic curve identifier.
	EdCurveEd448 = cryptoutilMagic.EdCurveEd448
	// EdCurveEd25519 is the Ed25519 elliptic curve identifier.
	EdCurveEd25519 = cryptoutilMagic.EdCurveEd25519
	// ECCurveP256 is the P-256 elliptic curve identifier.
	ECCurveP256 = cryptoutilMagic.ECCurveP256
	// ECCurveP384 is the P-384 elliptic curve identifier.
	ECCurveP384 = cryptoutilMagic.ECCurveP384
	// ECCurveP521 is the P-521 elliptic curve identifier.
	ECCurveP521 = cryptoutilMagic.ECCurveP521

	// AES key sizes in bits.
	aesKeySize128 = cryptoutilMagic.AESKeySize128
	aesKeySize192 = cryptoutilMagic.AESKeySize192
	aesKeySize256 = cryptoutilMagic.AESKeySize256

	// AES HMAC-SHA2 key sizes in bits.
	aesHsKeySize256 = cryptoutilMagic.AESHSKeySize256
	aesHsKeySize384 = cryptoutilMagic.AESHSKeySize384
	aesHsKeySize512 = cryptoutilMagic.AESHSKeySize512

	// Minimum HMAC key size in bits.
	minHMACKeySize = cryptoutilMagic.MinHMACKeySize

	// Bits to bytes conversion factor.
	bitsToBytes = cryptoutilMagic.BitsToBytes
)

// GenerateRSAKeyPairFunction returns a function that generates an RSA key pair with the specified bit size.
func GenerateRSAKeyPairFunction(rsaBits int) func() (*KeyPair, error) {
	return func() (*KeyPair, error) { return GenerateRSAKeyPair(rsaBits) }
}

// GenerateRSAKeyPair generates an RSA key pair with the specified bit size.
func GenerateRSAKeyPair(rsaBits int) (*KeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if err != nil {
		return nil, fmt.Errorf("generate RSA key pair failed: %w", err)
	}

	return &KeyPair{Private: privateKey, Public: &privateKey.PublicKey}, nil
}

// GenerateECDSAKeyPairFunction returns a function that generates an ECDSA key pair with the specified curve.
func GenerateECDSAKeyPairFunction(ecdsaCurve elliptic.Curve) func() (*KeyPair, error) {
	return func() (*KeyPair, error) { return GenerateECDSAKeyPair(ecdsaCurve) }
}

// GenerateECDSAKeyPair generates an ECDSA key pair with the specified curve.
func GenerateECDSAKeyPair(ecdsaCurve elliptic.Curve) (*KeyPair, error) {
	privateKey, err := ecdsa.GenerateKey(ecdsaCurve, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate ECDSA key pair failed: %w", err)
	}

	return &KeyPair{Private: privateKey, Public: &privateKey.PublicKey}, nil
}

// GenerateECDHKeyPairFunction returns a function that generates an ECDH key pair with the specified curve.
func GenerateECDHKeyPairFunction(ecdhCurve ecdh.Curve) func() (*KeyPair, error) {
	return func() (*KeyPair, error) { return GenerateECDHKeyPair(ecdhCurve) }
}

// GenerateECDHKeyPair generates an ECDH key pair with the specified curve.
func GenerateECDHKeyPair(ecdhCurve ecdh.Curve) (*KeyPair, error) {
	privateKey, err := ecdhCurve.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate ECDH key pair failed: %w", err)
	}

	return &KeyPair{Private: privateKey, Public: privateKey.PublicKey()}, nil
}

// GenerateEDDSAKeyPairFunction returns a function that generates an EdDSA key pair with the specified curve.
func GenerateEDDSAKeyPairFunction(edCurve string) func() (*KeyPair, error) {
	return func() (*KeyPair, error) { return GenerateEDDSAKeyPair(edCurve) }
}

// GenerateEDDSAKeyPair generates an EdDSA key pair with the specified curve (Ed25519 or Ed448).
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

// GenerateAESKeyFunction returns a function that generates an AES key with the specified bit size.
func GenerateAESKeyFunction(aesBits int) func() (SecretKey, error) {
	return func() (SecretKey, error) { return GenerateAESKey(aesBits) }
}

// GenerateAESKey generates an AES key with the specified bit size (128, 192, or 256).
func GenerateAESKey(aesBits int) (SecretKey, error) {
	if aesBits != aesKeySize128 && aesBits != aesKeySize192 && aesBits != aesKeySize256 {
		return nil, fmt.Errorf("invalid AES key size: %d (must be %d, %d, or %d bits)", aesBits, aesKeySize128, aesKeySize192, aesKeySize256)
	}

	aesSecretKeyBytes, err := cryptoutilRandom.GenerateBytes(aesBits / bitsToBytes)
	if err != nil {
		return nil, fmt.Errorf("generate AES %d key failed: %w", aesBits, err)
	}

	return aesSecretKeyBytes, nil
}

// GenerateAESHSKeyFunction returns a function that generates an AES-HMAC-SHA2 key with the specified bit size.
func GenerateAESHSKeyFunction(aesHsBits int) func() (SecretKey, error) {
	return func() (SecretKey, error) { return GenerateAESHSKey(aesHsBits) }
}

// GenerateAESHSKey generates an AES-HMAC-SHA2 key with the specified bit size (256, 384, or 512).
func GenerateAESHSKey(aesHsBits int) (SecretKey, error) {
	if aesHsBits != aesHsKeySize256 && aesHsBits != aesHsKeySize384 && aesHsBits != aesHsKeySize512 {
		return nil, fmt.Errorf("invalid AES HAMC-SHA2 key size: %d (must be %d, %d, or %d bits)", aesHsBits, aesHsKeySize256, aesHsKeySize384, aesHsKeySize512)
	}

	aesHsSecretKeyBytes, err := cryptoutilRandom.GenerateBytes(aesHsBits / bitsToBytes)
	if err != nil {
		return nil, fmt.Errorf("generate AES HAMC-SHA2 %d key failed: %w", aesHsBits, err)
	}

	return aesHsSecretKeyBytes, nil
}

// GenerateHMACKeyFunction returns a function that generates an HMAC key with the specified bit size.
func GenerateHMACKeyFunction(hmacBits int) func() (SecretKey, error) {
	return func() (SecretKey, error) { return GenerateHMACKey(hmacBits) }
}

// GenerateHMACKey generates an HMAC key with the specified bit size.
func GenerateHMACKey(hmacBits int) (SecretKey, error) {
	if hmacBits < minHMACKeySize {
		return nil, fmt.Errorf("invalid HMAC key size: %d (must be %d bits or higher)", hmacBits, minHMACKeySize)
	}

	hmacSecretKeyBytes, err := cryptoutilRandom.GenerateBytes(hmacBits / bitsToBytes)
	if err != nil {
		return nil, fmt.Errorf("generate HMAC %d key failed: %w", hmacBits, err)
	}

	return hmacSecretKeyBytes, nil
}
