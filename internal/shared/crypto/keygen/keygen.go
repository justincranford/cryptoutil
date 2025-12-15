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
	cryptoutilUtil "cryptoutil/internal/shared/util"

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
	EdCurveEd448   = cryptoutilMagic.EdCurveEd448
	EdCurveEd25519 = cryptoutilMagic.EdCurveEd25519
	ECCurveP256    = cryptoutilMagic.ECCurveP256
	ECCurveP384    = cryptoutilMagic.ECCurveP384
	ECCurveP521    = cryptoutilMagic.ECCurveP521

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
	if aesBits != aesKeySize128 && aesBits != aesKeySize192 && aesBits != aesKeySize256 {
		return nil, fmt.Errorf("invalid AES key size: %d (must be %d, %d, or %d bits)", aesBits, aesKeySize128, aesKeySize192, aesKeySize256)
	}

	aesSecretKeyBytes, err := cryptoutilUtil.GenerateBytes(aesBits / bitsToBytes)
	if err != nil {
		return nil, fmt.Errorf("generate AES %d key failed: %w", aesBits, err)
	}

	return aesSecretKeyBytes, nil
}

func GenerateAESHSKeyFunction(aesHsBits int) func() (SecretKey, error) {
	return func() (SecretKey, error) { return GenerateAESHSKey(aesHsBits) }
}

func GenerateAESHSKey(aesHsBits int) (SecretKey, error) {
	if aesHsBits != aesHsKeySize256 && aesHsBits != aesHsKeySize384 && aesHsBits != aesHsKeySize512 {
		return nil, fmt.Errorf("invalid AES HAMC-SHA2 key size: %d (must be %d, %d, or %d bits)", aesHsBits, aesHsKeySize256, aesHsKeySize384, aesHsKeySize512)
	}

	aesHsSecretKeyBytes, err := cryptoutilUtil.GenerateBytes(aesHsBits / bitsToBytes)
	if err != nil {
		return nil, fmt.Errorf("generate AES HAMC-SHA2 %d key failed: %w", aesHsBits, err)
	}

	return aesHsSecretKeyBytes, nil
}

func GenerateHMACKeyFunction(hmacBits int) func() (SecretKey, error) {
	return func() (SecretKey, error) { return GenerateHMACKey(hmacBits) }
}

func GenerateHMACKey(hmacBits int) (SecretKey, error) {
	if hmacBits < minHMACKeySize {
		return nil, fmt.Errorf("invalid HMAC key size: %d (must be %d bits or higher)", hmacBits, minHMACKeySize)
	}

	hmacSecretKeyBytes, err := cryptoutilUtil.GenerateBytes(hmacBits / bitsToBytes)
	if err != nil {
		return nil, fmt.Errorf("generate HMAC %d key failed: %w", hmacBits, err)
	}

	return hmacSecretKeyBytes, nil
}
