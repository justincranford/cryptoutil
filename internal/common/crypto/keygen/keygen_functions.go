package keygen

import (
	"crypto"
	"crypto/ecdh"
	"crypto/elliptic"
)

type SecretKey any // []byte, googleUuid.UUID

type Key struct {
	Private crypto.PrivateKey
	Public  crypto.PublicKey
	Secret  SecretKey
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
	return func() (Key, error) { return GenerateEDDSAKeyPair(edCurve) }
}

func GenerateAESKeyFunction(aesBits int) func() (Key, error) {
	return func() (Key, error) { return GenerateAESKey(aesBits) }
}

func GenerateAESHSKeyFunction(aesHsBits int) func() (Key, error) {
	return func() (Key, error) { return GenerateAESHSKey(aesHsBits) }
}

func GenerateHMACKeyFunction(hmacBits int) func() (Key, error) {
	return func() (Key, error) { return GenerateHMACKey(hmacBits) }
}

func GenerateUUIDv7Function() func() (Key, error) {
	return func() (Key, error) { return GenerateUUIDv7() }
}
