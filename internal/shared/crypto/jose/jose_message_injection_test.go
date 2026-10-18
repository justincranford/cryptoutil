// Copyright (c) 2025-2026 Justin Cranford.
//
//

package crypto

import (
	"errors"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	joseJws "github.com/lestrrat-go/jwx/v3/jws"
	"github.com/stretchr/testify/require"
)

const (
	joseSetIATHeaderError = "failed to set iat header"
	joseSetKIDHeaderError = "failed to set kid header"
	joseSetALGHeaderError = "failed to set alg header"
	joseParseTestCase     = "parse"
	joseSetIATTestCase    = "set iat"
	joseSetKIDTestCase    = "set kid"
	joseSetALGTestCase    = "set alg"
)

var errInjectedMessageOperation = errors.New("injected message operation error")

func TestEncryptBytesWithContext_InjectedOperationErrors(t *testing.T) {
	t.Parallel()

	_, encryptJWK, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)

	tests := []struct {
		name        string
		headerName  string
		context     []byte
		failEncrypt bool
		failParse   bool
		wantErr     string
	}{
		{name: joseSetIATTestCase, headerName: cryptoutilSharedMagic.ClaimIat, wantErr: joseSetIATHeaderError},
		{name: "set aad", headerName: joseJwe.AuthenticatedDataKey, context: []byte(joseMessagePayload), wantErr: "failed to set aad header"},
		{name: "set kid", headerName: joseJwk.KeyIDKey, wantErr: "failed to set kid header"},
		{name: "set enc", headerName: cryptoutilSharedMagic.JoseKeyUseEnc, wantErr: "failed to set enc header"},
		{name: "set alg", headerName: joseJwk.AlgorithmKey, wantErr: "failed to set alg header"},
		{name: "encrypt", failEncrypt: true, wantErr: "failed to encrypt clearBytes"},
		{name: joseParseTestCase, failParse: true, wantErr: "failed to parse JWE message bytes"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			setHeader := setJWEHeader
			if tc.headerName != "" {
				setHeader = func(setter jweHeaderSetter, name string, value any) error {
					if name == tc.headerName {
						return errInjectedMessageOperation
					}

					return setter(name, value)
				}
			}

			encrypt := joseJwe.Encrypt
			if tc.failEncrypt {
				encrypt = func([]byte, ...joseJwe.EncryptOption) ([]byte, error) {
					return nil, errInjectedMessageOperation
				}
			}

			parse := joseJwe.Parse

			if tc.failParse {
				encrypt = func([]byte, ...joseJwe.EncryptOption) ([]byte, error) {
					return []byte(joseMessagePayload), nil
				}
				parse = func([]byte, ...joseJwe.ParseOption) (*joseJwe.Message, error) {
					return nil, errInjectedMessageOperation
				}
			}

			_, _, err := encryptBytesWithContext(
				[]joseJwk.Key{encryptJWK},
				[]byte(joseMessagePayload),
				tc.context,
				setHeader,
				encrypt,
				parse,
			)
			require.ErrorIs(t, err, errInjectedMessageOperation)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestSignBytes_InjectedOperationErrors(t *testing.T) {
	t.Parallel()

	_, signingJWK, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)

	tests := []struct {
		name       string
		headerName string
		failSign   bool
		failParse  bool
		wantErr    string
	}{
		{name: joseSetIATTestCase, headerName: cryptoutilSharedMagic.ClaimIat, wantErr: joseSetIATHeaderError},
		{name: joseSetKIDTestCase, headerName: joseJwk.KeyIDKey, wantErr: joseSetKIDHeaderError},
		{name: joseSetALGTestCase, headerName: joseJwk.AlgorithmKey, wantErr: joseSetALGHeaderError},
		{name: "sign", failSign: true, wantErr: "failed to sign clearBytes"},
		{name: joseParseTestCase, failParse: true, wantErr: "failed to parse JWS message bytes"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			setHeader := setJWSHeader
			if tc.headerName != "" {
				setHeader = func(setter jwsHeaderSetter, name string, value any) error {
					if name == tc.headerName {
						return errInjectedMessageOperation
					}

					return setter(name, value)
				}
			}

			sign := joseJws.Sign
			if tc.failSign {
				sign = func([]byte, ...joseJws.SignOption) ([]byte, error) {
					return nil, errInjectedMessageOperation
				}
			}

			parse := joseJws.Parse

			if tc.failParse {
				sign = func([]byte, ...joseJws.SignOption) ([]byte, error) {
					return []byte(joseMessagePayload), nil
				}
				parse = func([]byte, ...joseJws.ParseOption) (*joseJws.Message, error) {
					return nil, errInjectedMessageOperation
				}
			}

			_, _, err := signBytes(
				[]joseJwk.Key{signingJWK},
				[]byte(joseMessagePayload),
				setHeader,
				sign,
				parse,
			)
			require.ErrorIs(t, err, errInjectedMessageOperation)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestJWEAlgorithmsFromJWSKey_RejectsSigningKey(t *testing.T) {
	t.Parallel()

	_, signingJWK, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)

	_, _, err = EncryptBytes([]joseJwk.Key{signingJWK}, []byte(joseMessagePayload))
	require.Error(t, err)
	_, err = DecryptBytes([]joseJwk.Key{signingJWK}, []byte(joseMalformedJWE))
	require.Error(t, err)
}
