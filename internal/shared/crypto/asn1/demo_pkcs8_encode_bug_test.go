// Copyright (c) 2025 Justin Cranford
//
//

package asn1

import (
	"crypto/ecdh"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var ecdhTestCurves = []struct {
	Name  string
	Curve ecdh.Curve
}{
	{"ECDH X25519", ecdh.X25519()}, // PASS
	{"ECDH P256", ecdh.P256()},     // FAIL => encode+decode returns ECDSA instead of ECDH
	{"ECDH P384", ecdh.P384()},     // FAIL => encode+decode returns ECDSA instead of ECDH
	{"ECDH P521", ecdh.P521()},     // FAIL => encode+decode returns ECDSA instead of ECDH
}

var ecdsaTestCurves = []struct {
	Name  string
	Curve elliptic.Curve
}{
	{"ECDSA P224", elliptic.P224()}, // PASS
	{"ECDSA P256", elliptic.P256()}, // PASS
	{"ECDSA P384", elliptic.P384()}, // PASS
	{"ECDSA P521", elliptic.P521()}, // PASS
}

func TestEncodeDecodeECDH(t *testing.T) {
	t.Parallel()
	t.Skip("Blocked by bug: https://github.com/golang/go/issues/71919")

	for _, curve := range ecdhTestCurves {
		t.Run(curve.Name, func(t *testing.T) {
			original, err := curve.Curve.GenerateKey(crand.Reader)
			require.NoError(t, err)

			require.IsType(t, &ecdh.PrivateKey{}, original)

			decoded, err := pkcs8EncodeDecode(t, original)
			require.NoError(t, err)

			require.IsType(t, &ecdh.PrivateKey{}, decoded)
		})
	}
}

func TestEncodeDecodeECDSA(t *testing.T) {
	t.Parallel()
	for _, curve := range ecdsaTestCurves {
		t.Run(curve.Name, func(t *testing.T) {
			original, err := ecdsa.GenerateKey(curve.Curve, crand.Reader)
			require.NoError(t, err)

			require.IsType(t, &ecdsa.PrivateKey{}, original)

			decoded, err := pkcs8EncodeDecode(t, original)
			require.NoError(t, err)

			require.IsType(t, &ecdsa.PrivateKey{}, decoded)
		})
	}
}

func pkcs8EncodeDecode(t *testing.T, key any) (any, error) {
	t.Helper()

	encodedBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("encode failed: %w", err)
	}

	pemBytes := pem.EncodeToMemory(&pem.Block{Bytes: encodedBytes, Type: "PRIVATE KEY"})
	t.Logf("PKCS#8 PEM of private Key :\n%s", string(pemBytes))

	decodedKey, err := x509.ParsePKCS8PrivateKey(encodedBytes)
	if err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	return decodedKey, nil
}
