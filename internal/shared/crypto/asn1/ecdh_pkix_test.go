// Copyright (c) 2025-2026 Justin Cranford.
//

package asn1

import (
	"crypto/ecdh"
	crand "crypto/rand"
	"crypto/x509"
	stdasn1 "encoding/asn1"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseObjectIdentifier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		want    string
		wantErr string
	}{
		{name: "valid first arc zero", value: testECDHOIDZero, want: testECDHOIDZero},
		{name: "valid first arc one", value: testECDHOIDOne, want: testECDHOIDOne},
		{name: "valid first arc two", value: testECDHOIDTwo, want: testECDHOIDTwo},
		{name: "missing second arc", value: "1", wantErr: "invalid object identifier"},
		{name: "empty arc", value: "1..2", wantErr: testECDHInvalidObjectIdentifierArcError},
		{name: "non-numeric arc", value: "1.x", wantErr: testECDHInvalidObjectIdentifierArcError},
		{name: "negative arc", value: "1.-1", wantErr: "negative object identifier arc"},
		{name: "first arc too large", value: "3.0", wantErr: "invalid first object identifier arc"},
		{name: "second arc too large", value: "1.40", wantErr: "invalid second object identifier arc"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseObjectIdentifier(tc.value)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, got.String())
		})
	}
}

func TestMarshalECDHPrivateKeyInfo_Errors(t *testing.T) {
	t.Parallel()

	key, err := ecdh.P256().GenerateKey(crand.Reader)
	require.NoError(t, err)

	validDER, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)

	var validInfo pkcs8PrivateKeyInfo

	_, err = stdasn1.Unmarshal(validDER, &validInfo)
	require.NoError(t, err)

	x25519Key, err := ecdh.X25519().GenerateKey(crand.Reader)
	require.NoError(t, err)
	x25519DER, err := x509.MarshalPKCS8PrivateKey(x25519Key)
	require.NoError(t, err)

	missingParametersInfo := validInfo
	missingParametersInfo.Algorithm.Parameters = stdasn1.RawValue{}
	missingParametersDER, err := stdasn1.Marshal(missingParametersInfo)
	require.NoError(t, err)

	tests := []struct {
		name          string
		marshalFn     func(any) ([]byte, error)
		parseOIDFn    func(string) (stdasn1.ObjectIdentifier, error)
		marshalASN1Fn func(any) ([]byte, error)
		wantErr       string
	}{
		{
			name: testECDHMarshalErrorCase,
			marshalFn: func(any) ([]byte, error) {
				return nil, errors.New(testECDHInjectedMarshalError)
			},
			wantErr: "marshal PKCS#8 ECDH private key",
		},
		{
			name: "malformed PKCS8",
			marshalFn: func(any) ([]byte, error) {
				return []byte{0xff}, nil
			},
			wantErr: "unmarshal PKCS#8 ECDH private key",
		},
		{
			name: "trailing PKCS8 data",
			marshalFn: func(any) ([]byte, error) {
				return append(validDER, 0), nil
			},
			wantErr: "PKCS#8 ECDH private key has trailing data",
		},
		{
			name: testECDHUnexpectedAlgorithmOIDCase,
			marshalFn: func(any) ([]byte, error) {
				return x25519DER, nil
			},
			wantErr: "unexpected ECDH private-key algorithm OID",
		},
		{
			name: testECDHMissingCurveParametersCase,
			marshalFn: func(any) ([]byte, error) {
				return missingParametersDER, nil
			},
			wantErr: "PKCS#8 ECDH private key is missing curve parameters",
		},
		{
			name: testECDHInvalidECDHOIDCase,
			parseOIDFn: func(string) (stdasn1.ObjectIdentifier, error) {
				return nil, errors.New(testECDHInjectedOIDParseError)
			},
			wantErr: testECDHParseECDHAlgorithmOIDError,
		},
		{
			name: testECDHASN1MarshalErrorCase,
			marshalASN1Fn: func(any) ([]byte, error) {
				return nil, errors.New(testECDHInjectedASN1MarshalError)
			},
			wantErr: "marshal id-ecDH PKCS#8 private key",
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

			marshalPKCS8Fn := tc.marshalFn
			if marshalPKCS8Fn == nil {
				marshalPKCS8Fn = x509.MarshalPKCS8PrivateKey
			}

			_, err := marshalECDHPrivateKeyInfo(key, marshalPKCS8Fn, parseOIDFn, marshalASN1Fn)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestDEREncode_ECDHMarshalErrors(t *testing.T) {
	t.Parallel()

	privateKey, err := ecdh.P256().GenerateKey(crand.Reader)
	require.NoError(t, err)

	tests := []struct {
		name string
		key  any
	}{
		{name: "private key", key: privateKey},
		{name: "public key", key: privateKey.PublicKey()},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, _, err := derEncodeWithFns(tc.key, func(any) ([]byte, error) {
				return nil, errors.New(testECDHInjectedMarshalError)
			}, func(any) ([]byte, error) {
				return nil, errors.New(testECDHInjectedMarshalError)
			})
			require.ErrorContains(t, err, "encode failed")
		})
	}
}

func TestMarshalECDHPublicKeyInfo_Errors(t *testing.T) {
	t.Parallel()

	key, err := ecdh.P256().GenerateKey(crand.Reader)
	require.NoError(t, err)

	validDER, err := x509.MarshalPKIXPublicKey(key.PublicKey())
	require.NoError(t, err)

	var validInfo pkixPublicKeyInfo

	_, err = stdasn1.Unmarshal(validDER, &validInfo)
	require.NoError(t, err)

	x25519Key, err := ecdh.X25519().GenerateKey(crand.Reader)
	require.NoError(t, err)
	x25519DER, err := x509.MarshalPKIXPublicKey(x25519Key.PublicKey())
	require.NoError(t, err)

	missingParametersInfo := validInfo
	missingParametersInfo.Algorithm.Parameters = stdasn1.RawValue{}
	missingParametersDER, err := stdasn1.Marshal(missingParametersInfo)
	require.NoError(t, err)

	tests := []struct {
		name          string
		marshalFn     func(any) ([]byte, error)
		parseOIDFn    func(string) (stdasn1.ObjectIdentifier, error)
		marshalASN1Fn func(any) ([]byte, error)
		wantErr       string
	}{
		{
			name: testECDHMarshalErrorCase,
			marshalFn: func(any) ([]byte, error) {
				return nil, errors.New(testECDHInjectedMarshalError)
			},
			wantErr: "marshal PKIX ECDH public key",
		},
		{
			name: "malformed PKIX",
			marshalFn: func(any) ([]byte, error) {
				return []byte{0xff}, nil
			},
			wantErr: "unmarshal PKIX ECDH public key",
		},
		{
			name: "trailing PKIX data",
			marshalFn: func(any) ([]byte, error) {
				return append(validDER, 0), nil
			},
			wantErr: "PKIX ECDH public key has trailing data",
		},
		{
			name: testECDHUnexpectedAlgorithmOIDCase,
			marshalFn: func(any) ([]byte, error) {
				return x25519DER, nil
			},
			wantErr: "unexpected ECDH public-key algorithm OID",
		},
		{
			name: testECDHMissingCurveParametersCase,
			marshalFn: func(any) ([]byte, error) {
				return missingParametersDER, nil
			},
			wantErr: "PKIX ECDH public key is missing curve parameters",
		},
		{
			name: testECDHInvalidECDHOIDCase,
			parseOIDFn: func(string) (stdasn1.ObjectIdentifier, error) {
				return nil, errors.New(testECDHInjectedOIDParseError)
			},
			wantErr: testECDHParseECDHAlgorithmOIDError,
		},
		{
			name: testECDHASN1MarshalErrorCase,
			marshalASN1Fn: func(any) ([]byte, error) {
				return nil, errors.New(testECDHInjectedASN1MarshalError)
			},
			wantErr: "marshal id-ecDH PKIX public key",
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

			marshalPKIXFn := tc.marshalFn
			if marshalPKIXFn == nil {
				marshalPKIXFn = x509.MarshalPKIXPublicKey
			}

			_, err := marshalECDHPublicKeyInfo(key.PublicKey(), marshalPKIXFn, parseOIDFn, marshalASN1Fn)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}
