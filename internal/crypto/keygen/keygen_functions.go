package keygen

import (
	"crypto/ecdh"
	"crypto/elliptic"
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
