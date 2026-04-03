// Copyright (c) 2025 Justin Cranford
//
//

package keygen

import (
	"crypto"
	"crypto/ecdh"
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"errors"
	"fmt"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"

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
	EdCurveEd448 = cryptoutilSharedMagic.EdCurveEd448
	// EdCurveEd25519 is the Ed25519 elliptic curve identifier.
	EdCurveEd25519 = cryptoutilSharedMagic.EdCurveEd25519
	// ECCurveP256 is the P-256 elliptic curve identifier.
	ECCurveP256 = cryptoutilSharedMagic.ECCurveP256
	// ECCurveP384 is the P-384 elliptic curve identifier.
	ECCurveP384 = cryptoutilSharedMagic.ECCurveP384
	// ECCurveP521 is the P-521 elliptic curve identifier.
	ECCurveP521 = cryptoutilSharedMagic.ECCurveP521

	// AES key sizes in bits.

	// AES HMAC-SHA2 key sizes in bits.

	// Minimum HMAC key size in bits.

	// Bits to bytes conversion factor.
)

// GenerateRSAKeyPairFunction returns a function that generates an RSA key pair with the specified bit size.
func GenerateRSAKeyPairFunction(rsaBits int) func() (*KeyPair, error) {
	return func() (*KeyPair, error) { return GenerateRSAKeyPair(rsaBits) }
}

// GenerateRSAKeyPair generates an RSA key pair with the specified bit size.
func GenerateRSAKeyPair(rsaBits int) (*KeyPair, error) {
	return generateRSAKeyPairInternal(rsaBits, func(bits int) (*rsa.PrivateKey, error) { return rsa.GenerateKey(crand.Reader, bits) })
}

func generateRSAKeyPairInternal(rsaBits int, rsaFn func(int) (*rsa.PrivateKey, error)) (*KeyPair, error) {
	privateKey, err := rsaFn(rsaBits)
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
	return generateECDSAKeyPairInternal(ecdsaCurve, func(curve elliptic.Curve) (*ecdsa.PrivateKey, error) {
		return ecdsa.GenerateKey(curve, crand.Reader)
	})
}

func generateECDSAKeyPairInternal(ecdsaCurve elliptic.Curve, ecdsaFn func(elliptic.Curve) (*ecdsa.PrivateKey, error)) (*KeyPair, error) {
	privateKey, err := ecdsaFn(ecdsaCurve)
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
	return generateECDHKeyPairInternal(ecdhCurve, func(curve ecdh.Curve) (*ecdh.PrivateKey, error) { return curve.GenerateKey(crand.Reader) })
}

func generateECDHKeyPairInternal(ecdhCurve ecdh.Curve, ecdhFn func(ecdh.Curve) (*ecdh.PrivateKey, error)) (*KeyPair, error) {
	privateKey, err := ecdhFn(ecdhCurve)
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
	return generateEDDSAKeyPairInternal(edCurve,
		func() (ed448.PublicKey, ed448.PrivateKey, error) { return ed448.GenerateKey(crand.Reader) },
		func() (ed25519.PublicKey, ed25519.PrivateKey, error) { return ed25519.GenerateKey(crand.Reader) },
	)
}

func generateEDDSAKeyPairInternal(edCurve string, ed448Fn func() (ed448.PublicKey, ed448.PrivateKey, error), ed25519Fn func() (ed25519.PublicKey, ed25519.PrivateKey, error)) (*KeyPair, error) {
	switch edCurve {
	case EdCurveEd448:
		publicKey, privateKey, err := ed448Fn()
		if err != nil {
			return nil, fmt.Errorf("generate Ed448 key pair failed: %w", err)
		}

		return &KeyPair{Private: privateKey, Public: publicKey}, nil
	case EdCurveEd25519:
		publicKey, privateKey, err := ed25519Fn()
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
	return generateAESKeyInternal(aesBits, cryptoutilSharedUtilRandom.GenerateBytes)
}

func generateAESKeyInternal(aesBits int, generateBytesFn func(int) ([]byte, error)) (SecretKey, error) {
	if aesBits != cryptoutilSharedMagic.AESKeySize128 && aesBits != cryptoutilSharedMagic.AESKeySize192 && aesBits != cryptoutilSharedMagic.AESKeySize256 {
		return nil, fmt.Errorf("invalid AES key size: %d (must be %d, %d, or %d bits)", aesBits, cryptoutilSharedMagic.AESKeySize128, cryptoutilSharedMagic.AESKeySize192, cryptoutilSharedMagic.AESKeySize256)
	}

	aesSecretKeyBytes, err := generateBytesFn(aesBits / cryptoutilSharedMagic.BitsToBytes)
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
	return generateAESHSKeyInternal(aesHsBits, cryptoutilSharedUtilRandom.GenerateBytes)
}

func generateAESHSKeyInternal(aesHsBits int, generateBytesFn func(int) ([]byte, error)) (SecretKey, error) {
	if aesHsBits != cryptoutilSharedMagic.AESHSKeySize256 && aesHsBits != cryptoutilSharedMagic.AESHSKeySize384 && aesHsBits != cryptoutilSharedMagic.AESHSKeySize512 {
		return nil, fmt.Errorf("invalid AES HAMC-SHA2 key size: %d (must be %d, %d, or %d bits)", aesHsBits, cryptoutilSharedMagic.AESHSKeySize256, cryptoutilSharedMagic.AESHSKeySize384, cryptoutilSharedMagic.AESHSKeySize512)
	}

	aesHsSecretKeyBytes, err := generateBytesFn(aesHsBits / cryptoutilSharedMagic.BitsToBytes)
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
	return generateHMACKeyInternal(hmacBits, cryptoutilSharedUtilRandom.GenerateBytes)
}

func generateHMACKeyInternal(hmacBits int, generateBytesFn func(int) ([]byte, error)) (SecretKey, error) {
	if hmacBits < cryptoutilSharedMagic.MinHMACKeySize {
		return nil, fmt.Errorf("invalid HMAC key size: %d (must be %d bits or higher)", hmacBits, cryptoutilSharedMagic.MinHMACKeySize)
	}

	hmacSecretKeyBytes, err := generateBytesFn(hmacBits / cryptoutilSharedMagic.BitsToBytes)
	if err != nil {
		return nil, fmt.Errorf("generate HMAC %d key failed: %w", hmacBits, err)
	}

	return hmacSecretKeyBytes, nil
}
