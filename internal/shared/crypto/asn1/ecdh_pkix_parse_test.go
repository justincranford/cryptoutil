// Copyright (c) 2025-2026 Justin Cranford.
//

package asn1

import (
	"crypto/ecdh"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	stdasn1 "encoding/asn1"
	"errors"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestParseECDHPrivateKey_Errors(t *testing.T) {
	t.Parallel()

	key, err := ecdh.P256().GenerateKey(crand.Reader)
	require.NoError(t, err)

	validDER, err := marshalECDHPrivateKeyInfo(key, x509.MarshalPKCS8PrivateKey, parseObjectIdentifier, stdasn1.Marshal)
	require.NoError(t, err)

	var validInfo pkcs8PrivateKeyInfo

	_, err = stdasn1.Unmarshal(validDER, &validInfo)
	require.NoError(t, err)

	invalidInnerKeyInfo := validInfo
	invalidInnerKeyInfo.PrivateKey = []byte{0xff}
	invalidInnerKeyDER, err := stdasn1.Marshal(invalidInnerKeyInfo)
	require.NoError(t, err)

	trailingDER := append(validDER, 0)

	unsupportedCurve, err := ecdsa.GenerateKey(elliptic.P224(), crand.Reader)
	require.NoError(t, err)
	unsupportedCurveDER, err := x509.MarshalPKCS8PrivateKey(unsupportedCurve)
	require.NoError(t, err)

	var unsupportedCurveInfo pkcs8PrivateKeyInfo

	_, err = stdasn1.Unmarshal(unsupportedCurveDER, &unsupportedCurveInfo)
	require.NoError(t, err)
	unsupportedCurveInfo.Algorithm.Algorithm, err = parseObjectIdentifier(cryptoutilSharedMagic.StringOIDPublicKeyECDH)
	require.NoError(t, err)
	unsupportedCurveDER, err = stdasn1.Marshal(unsupportedCurveInfo)
	require.NoError(t, err)

	ordinaryECDSAKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)
	ordinaryECDSADER, err := x509.MarshalPKCS8PrivateKey(ordinaryECDSAKey)
	require.NoError(t, err)

	tests := []struct {
		name          string
		der           []byte
		parseOIDFn    func(string) (stdasn1.ObjectIdentifier, error)
		marshalASN1Fn func(any) ([]byte, error)
		parsePKCS8Fn  func([]byte) (any, error)
		wantErr       string
	}{
		{name: "invalid private ASN1", der: []byte{0xff}, wantErr: testECDHInvalidASN1SyntaxError},
		{name: testECDHTrailingDataCase, der: trailingDER, wantErr: "PKCS#8 private key has trailing data"},
		{name: "invalid ECDH inner key", der: invalidInnerKeyDER, wantErr: "parse id-ecDH PKCS#8 private key"},
		{name: testECDHUnsupportedCurveCase, der: unsupportedCurveDER, wantErr: "convert id-ecDH private key"},
		{name: testECDHOrdinaryECDSACompatibilityCase, der: ordinaryECDSADER},
		{
			name: testECDHInvalidECAlgorithmOIDCase,
			der:  validDER,
			parseOIDFn: func(string) (stdasn1.ObjectIdentifier, error) {
				return nil, errors.New(testECDHInjectedOIDParseError)
			},
			wantErr: testECDHParseECAlgorithmOIDError,
		},
		{
			name: testECDHASN1MarshalErrorCase,
			der:  validDER,
			marshalASN1Fn: func(any) ([]byte, error) {
				return nil, errors.New(testECDHInjectedASN1MarshalError)
			},
			wantErr: "marshal id-ecPublicKey PKCS#8 private key",
		},
		{
			name: testECDHUnexpectedParsedKeyTypeCase,
			der:  validDER,
			parsePKCS8Fn: func([]byte) (any, error) {
				return struct{}{}, nil
			},
			wantErr: "unexpected private-key type for id-ecDH",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			parseOIDFn := tc.parseOIDFn
			if parseOIDFn == nil {
				parseOIDFn = parseObjectIdentifier
			}

			marshalASN1Fn := tc.marshalASN1Fn
			if marshalASN1Fn == nil {
				marshalASN1Fn = stdasn1.Marshal
			}

			parsePKCS8Fn := tc.parsePKCS8Fn
			if parsePKCS8Fn == nil {
				parsePKCS8Fn = x509.ParsePKCS8PrivateKey
			}

			got, err := parsePKCS8PrivateKeyWithFns(tc.der, parseOIDFn, marshalASN1Fn, parsePKCS8Fn)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.IsType(t, &ecdsa.PrivateKey{}, got)
		})
	}
}

func TestParseECDHPublicKey_Errors(t *testing.T) {
	t.Parallel()

	key, err := ecdh.P256().GenerateKey(crand.Reader)
	require.NoError(t, err)

	validDER, err := marshalECDHPublicKeyInfo(key.PublicKey(), x509.MarshalPKIXPublicKey, parseObjectIdentifier, stdasn1.Marshal)
	require.NoError(t, err)

	var validInfo pkixPublicKeyInfo

	_, err = stdasn1.Unmarshal(validDER, &validInfo)
	require.NoError(t, err)

	invalidInnerKeyInfo := validInfo
	invalidInnerKeyInfo.PublicKey = stdasn1.BitString{Bytes: []byte{0xff}, BitLength: cryptoutilSharedMagic.BitsPerByte}
	invalidInnerKeyDER, err := stdasn1.Marshal(invalidInnerKeyInfo)
	require.NoError(t, err)

	trailingDER := append(validDER, 0)

	unsupportedCurve, err := ecdsa.GenerateKey(elliptic.P224(), crand.Reader)
	require.NoError(t, err)
	unsupportedCurveDER, err := x509.MarshalPKIXPublicKey(&unsupportedCurve.PublicKey)
	require.NoError(t, err)

	var unsupportedCurveInfo pkixPublicKeyInfo

	_, err = stdasn1.Unmarshal(unsupportedCurveDER, &unsupportedCurveInfo)
	require.NoError(t, err)
	unsupportedCurveInfo.Algorithm.Algorithm, err = parseObjectIdentifier(cryptoutilSharedMagic.StringOIDPublicKeyECDH)
	require.NoError(t, err)
	unsupportedCurveDER, err = stdasn1.Marshal(unsupportedCurveInfo)
	require.NoError(t, err)

	ordinaryECDSAKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)
	ordinaryECDSADER, err := x509.MarshalPKIXPublicKey(&ordinaryECDSAKey.PublicKey)
	require.NoError(t, err)

	tests := []struct {
		name          string
		der           []byte
		parseOIDFn    func(string) (stdasn1.ObjectIdentifier, error)
		marshalASN1Fn func(any) ([]byte, error)
		parsePKIXFn   func([]byte) (any, error)
		wantErr       string
	}{
		{name: "invalid public ASN1", der: []byte{0xff}, wantErr: testECDHInvalidASN1SyntaxError},
		{name: testECDHTrailingDataCase, der: trailingDER, wantErr: "PKIX public key has trailing data"},
		{name: "invalid ECDH public-key bits", der: invalidInnerKeyDER, wantErr: "parse id-ecDH PKIX public key"},
		{name: testECDHUnsupportedCurveCase, der: unsupportedCurveDER, wantErr: "convert id-ecDH public key"},
		{name: testECDHOrdinaryECDSACompatibilityCase, der: ordinaryECDSADER},
		{
			name: testECDHInvalidECAlgorithmOIDCase,
			der:  validDER,
			parseOIDFn: func(string) (stdasn1.ObjectIdentifier, error) {
				return nil, errors.New(testECDHInjectedOIDParseError)
			},
			wantErr: testECDHParseECAlgorithmOIDError,
		},
		{
			name: testECDHASN1MarshalErrorCase,
			der:  validDER,
			marshalASN1Fn: func(any) ([]byte, error) {
				return nil, errors.New(testECDHInjectedASN1MarshalError)
			},
			wantErr: "marshal id-ecPublicKey PKIX public key",
		},
		{
			name: testECDHUnexpectedParsedKeyTypeCase,
			der:  validDER,
			parsePKIXFn: func([]byte) (any, error) {
				return struct{}{}, nil
			},
			wantErr: "unexpected public-key type for id-ecDH",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			parseOIDFn := tc.parseOIDFn
			if parseOIDFn == nil {
				parseOIDFn = parseObjectIdentifier
			}

			marshalASN1Fn := tc.marshalASN1Fn
			if marshalASN1Fn == nil {
				marshalASN1Fn = stdasn1.Marshal
			}

			parsePKIXFn := tc.parsePKIXFn
			if parsePKIXFn == nil {
				parsePKIXFn = x509.ParsePKIXPublicKey
			}

			got, err := parsePKIXPublicKeyWithFns(tc.der, parseOIDFn, marshalASN1Fn, parsePKIXFn)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.IsType(t, &ecdsa.PublicKey{}, got)
		})
	}
}
