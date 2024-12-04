package keypairgen

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"errors"
)

// GenerateRSAKeyPair returns a function to generate RSA key pairs of a specified bit length
func GenerateRSAKeyPair(bits int) func() (KeyPair, error) {
	return func() (KeyPair, error) {
		return rsa.GenerateKey(rand.Reader, bits)
	}
}

// GenerateECKeyPair returns a function to generate EC key pairs for a given curve
func GenerateECKeyPair(curve elliptic.Curve) func() (KeyPair, error) {
	return func() (KeyPair, error) {
		priv, err := ecdsa.GenerateKey(curve, rand.Reader)
		if err != nil {
			return nil, err
		}
		return priv, nil
	}
}

// GenerateEDKeyPair returns a function to generate ED key pairs (currently supports Ed25519)
func GenerateEDKeyPair(curve string) func() (KeyPair, error) {
	return func() (KeyPair, error) {
		switch curve {
		case "Ed25519":
			_, priv, err := ed25519.GenerateKey(rand.Reader)
			if err != nil {
				return nil, err
			}
			return priv, nil
		default:
			return nil, errors.New("unsupported ED curve")
		}
	}
}
