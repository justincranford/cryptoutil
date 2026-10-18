// Copyright (c) 2025-2026 Justin Cranford.
//
//

package asn1

import (
	"crypto/ecdh"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	stdasn1 "encoding/asn1"
	"fmt"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

var ecdhTestCurves = []struct {
	Name  string
	Curve ecdh.Curve
}{
	{"ECDH X25519", ecdh.X25519()},
	{"ECDH P256", ecdh.P256()},
	{"ECDH P384", ecdh.P384()},
	{"ECDH P521", ecdh.P521()},
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

	for _, curve := range ecdhTestCurves {
		t.Run(curve.Name, func(t *testing.T) {
			t.Parallel()

			original, err := curve.Curve.GenerateKey(crand.Reader)
			require.NoError(t, err)

			require.IsType(t, &ecdh.PrivateKey{}, original)

			encoded, _, err := DEREncode(original)
			require.NoError(t, err)

			if curve.Curve != ecdh.X25519() {
				var privateKeyInfo pkcs8PrivateKeyInfo

				rest, err := stdasn1.Unmarshal(encoded, &privateKeyInfo)
				require.NoError(t, err)
				require.Empty(t, rest)
				require.Equal(t, cryptoutilSharedMagic.StringOIDPublicKeyECDH, privateKeyInfo.Algorithm.Algorithm.String())
			}

			decoded, _, err := DERDecodes(encoded)
			require.NoError(t, err)

			require.IsType(t, &ecdh.PrivateKey{}, decoded)
			decodedPrivateKey, ok := decoded.(*ecdh.PrivateKey)
			require.True(t, ok)
			require.Equal(t, original.Bytes(), decodedPrivateKey.Bytes())

			publicEncoded, publicPEMType, err := DEREncode(original.PublicKey())
			require.NoError(t, err)

			if curve.Curve != ecdh.X25519() {
				var publicKeyInfo pkixPublicKeyInfo

				rest, err := stdasn1.Unmarshal(publicEncoded, &publicKeyInfo)
				require.NoError(t, err)
				require.Empty(t, rest)
				require.Equal(t, cryptoutilSharedMagic.StringOIDPublicKeyECDH, publicKeyInfo.Algorithm.Algorithm.String())
			}

			decoded, err = DERDecode(publicEncoded, publicPEMType)
			require.NoError(t, err)

			require.IsType(t, &ecdh.PublicKey{}, decoded)
			decodedPublicKey, ok := decoded.(*ecdh.PublicKey)
			require.True(t, ok)
			require.Equal(t, original.PublicKey().Bytes(), decodedPublicKey.Bytes())
		})
	}
}

func TestEncodeDecodeECDSA(t *testing.T) {
	t.Parallel()

	for _, curve := range ecdsaTestCurves {
		t.Run(curve.Name, func(t *testing.T) {
			t.Parallel()

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

	decodedKey, err := x509.ParsePKCS8PrivateKey(encodedBytes)
	if err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	return decodedKey, nil
}
