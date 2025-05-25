package jose

import (
	"crypto/ecdh"
	"crypto/elliptic"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

func GenerateRSAJwkFunction(rsaBits int) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateRSAJwk(rsaBits) }
}

func GenerateECDSAJwkFunction(ecdsaCurve elliptic.Curve) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateECDSAJwk(ecdsaCurve) }
}

func GenerateECDHJwkFunction(ecdhCurve ecdh.Curve) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateECDHJwk(ecdhCurve) }
}

func GenerateEDDSAJwkFunction(edCurve string) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateEDDSAJwk(edCurve) }
}

func GenerateAESJwkFunction(aesBits int) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateAESJwk(aesBits) }
}

func GenerateAESHSJwkFunction(aesHsBits int) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateAESHSJwk(aesHsBits) }
}

func GenerateHMACJwkFunction(hmacBits int) func() (joseJwk.Key, error) {
	return func() (joseJwk.Key, error) { return GenerateHMACJwk(hmacBits) }
}
