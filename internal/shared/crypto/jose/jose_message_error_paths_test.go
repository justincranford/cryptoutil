// Copyright (c) 2025-2026 Justin Cranford.
//
//

package crypto

import (
	"crypto/ecdh"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	"errors"
	"io"
	"testing"

	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	joseJws "github.com/lestrrat-go/jwx/v3/jws"
	"github.com/stretchr/testify/require"
)

const (
	joseMessagePayload         = "payload"
	joseMalformedJWE           = "malformed JWE"
	joseMalformedJWS           = "malformed JWS"
	joseInvalidClearBytes      = "invalid clearBytes"
	joseNilPayloadTestCase     = "nil payload"
	joseEmptyPayloadTestCase   = "empty payload"
	joseInvalidJWEMessageBytes = "invalid jweMessageBytes"
	joseNilMessageTestCase     = "nil message"
	joseEmptyMessageTestCase   = "empty message"
	joseInvalidJWSMessageBytes = "invalid jwsMessageBytes"
)

var (
	errJWKMarshal      = errors.New("injected JWK marshal failure")
	errCurveGeneration = errors.New("injected curve generation failure")
)

type failingECDHCurve struct {
	ecdh.Curve
}

func (failingECDHCurve) GenerateKey(io.Reader) (*ecdh.PrivateKey, error) {
	return nil, errCurveGeneration
}

type marshalErrorJWK struct {
	mockUnknownKey
}

func (marshalErrorJWK) MarshalJSON() ([]byte, error) {
	return nil, errJWKMarshal
}

func TestEncryptKey_MarshalCEKError(t *testing.T) {
	t.Parallel()

	_, _, err := EncryptKey(nil, marshalErrorJWK{})
	require.ErrorIs(t, err, errJWKMarshal)
	require.ErrorContains(t, err, "failed to encode CEK")
}

func TestDecryptBytes_MalformedMessage(t *testing.T) {
	t.Parallel()

	_, decryptJWK, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)

	_, err = DecryptBytes([]joseJwk.Key{decryptJWK}, []byte(joseMalformedJWE))
	require.ErrorContains(t, err, "failed to parse JWE message bytes")
}

func TestDecryptKey_MalformedMessage(t *testing.T) {
	t.Parallel()

	_, decryptJWK, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)

	_, err = DecryptKey([]joseJwk.Key{decryptJWK}, []byte(joseMalformedJWE))
	require.ErrorContains(t, err, "failed to decrypt CDK bytes")
}

func TestDecryptBytes_WrongKey(t *testing.T) {
	t.Parallel()

	_, encryptJWK, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)

	_, decryptJWK, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)

	_, encrypted, err := EncryptBytes([]joseJwk.Key{encryptJWK}, []byte(joseMessagePayload))
	require.NoError(t, err)

	_, err = DecryptBytes([]joseJwk.Key{decryptJWK}, encrypted)
	require.ErrorContains(t, err, "failed to decrypt JWE message bytes")
}

func TestSignBytes_SignError(t *testing.T) {
	t.Parallel()

	_, signingJWK, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)
	require.NoError(t, signingJWK.Set(joseJwk.AlgorithmKey, joseJwa.RS256()))

	_, _, err = SignBytes([]joseJwk.Key{signingJWK}, []byte(joseMessagePayload))
	require.ErrorContains(t, err, "failed to sign clearBytes")
}

func TestVerifyBytes_WrongKey(t *testing.T) {
	t.Parallel()

	_, signingJWK, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)
	_, _, verificationJWK, _, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)

	_, signed, err := SignBytes([]joseJwk.Key{signingJWK}, []byte(joseMessagePayload))
	require.NoError(t, err)

	_, err = VerifyBytes([]joseJwk.Key{verificationJWK}, signed)
	require.Error(t, err)
}

func TestJWEKeyValidation_GenerationErrors(t *testing.T) {
	t.Parallel()

	enc := EncA256GCM
	alg := AlgA256KW

	_, err := validateOrGenerateJWEAESJWK(
		nil,
		&enc,
		&alg,
		0,
		&enc,
	)
	require.ErrorContains(t, err, "failed to generate AES")

	_, err = validateOrGenerateJWERSAJWK(
		nil,
		&enc,
		&alg,
		0,
		&enc,
	)
	require.ErrorContains(t, err, "failed to generate RSA")
}

func TestJWSKeyValidation_GenerationErrors(t *testing.T) {
	t.Parallel()

	rsaBits := 0
	_, err := validateOrGenerateJWSRSAJWK(nil, AlgRS256, rsaBits)
	require.ErrorContains(t, err, "failed to generate RSA")

	_, err = validateOrGenerateJWSEddsaJWK(nil, AlgEdDSA, cryptoutilSharedMagic.DevSetupTestInvalidName)
	require.ErrorContains(t, err, "failed to generate Ed29919")

	hmacBits := 0
	_, err = validateOrGenerateJWSHMACJWK(nil, AlgHS256, hmacBits)
	require.ErrorContains(t, err, "failed to generate AES")
}

func TestExtractAlgFromJWSJWK_KeyEncryptionAlgorithm(t *testing.T) {
	t.Parallel()

	_, encryptJWK, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)

	_, err = ExtractAlgFromJWSJWK(encryptJWK, 0)
	require.ErrorContains(t, err, "is not a signature algorithm")
}

func TestIsEncryptAndDecryptJWK_EdDSAKey(t *testing.T) {
	t.Parallel()

	_, privateJWK, publicJWK, _, _, err := GenerateJWSJWKForAlg(&AlgEdDSA)
	require.NoError(t, err)

	for _, key := range []joseJwk.Key{privateJWK, publicJWK} {
		require.NoError(t, key.Set(cryptoutilSharedMagic.JoseKeyUseEnc, EncA256GCM))
		require.NoError(t, key.Set(joseJwk.AlgorithmKey, AlgA256KW))

		encrypt, err := IsEncryptJWK(key)
		require.NoError(t, err)
		require.False(t, encrypt)

		decrypt, err := IsDecryptJWK(key)
		require.NoError(t, err)
		require.False(t, decrypt)
	}
}

func TestVerifyBytes_MalformedMessage(t *testing.T) {
	t.Parallel()

	_, signingJWK, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)

	_, err = VerifyBytes([]joseJwk.Key{signingJWK}, []byte(joseMalformedJWS))
	require.ErrorContains(t, err, "failed to parse JWS message bytes")
}

func TestLogJWSInfo_PublicHeader(t *testing.T) {
	t.Parallel()

	_, signingJWK, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)

	publicHeaders := joseJws.NewHeaders()
	require.NoError(t, publicHeaders.Set(cryptoutilSharedMagic.PublicLabel, "value"))

	signed, err := joseJws.Sign(
		[]byte(joseMessagePayload),
		joseJws.WithJSON(),
		joseJws.WithKey(
			joseJwa.HS256(),
			signingJWK,
			joseJws.WithPublicHeaders(publicHeaders),
		),
	)
	require.NoError(t, err)

	message, err := joseJws.Parse(signed)
	require.NoError(t, err)
	require.NoError(t, LogJWSInfo(message))
}

type signatureAlgorithmSetErrorKey struct {
	mockUnknownKey
}

func (signatureAlgorithmSetErrorKey) Get(name string, dst any) error {
	if name != joseJwk.AlgorithmKey {
		return errors.New("unexpected JWK attribute")
	}

	algorithm, ok := dst.(*joseJwa.SignatureAlgorithm)
	if !ok {
		return errors.New("unexpected algorithm destination")
	}

	*algorithm = joseJwa.HS256()

	return nil
}

func (signatureAlgorithmSetErrorKey) Set(string, any) error {
	return errJWKMarshal
}

func TestEnsureSignatureAlgorithmType_SetError(t *testing.T) {
	t.Parallel()

	err := EnsureSignatureAlgorithmType(signatureAlgorithmSetErrorKey{})
	require.ErrorIs(t, err, errJWKMarshal)
	require.ErrorContains(t, err, "failed to set typed algorithm")
}

func TestEncryptBytesWithContext_RejectsEmptyPayloads(t *testing.T) {
	t.Parallel()

	_, encryptJWK, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)

	tests := []struct {
		name       string
		clearBytes []byte
		wantErr    string
	}{
		{name: joseNilPayloadTestCase, clearBytes: nil, wantErr: joseInvalidClearBytes},
		{name: joseEmptyPayloadTestCase, clearBytes: []byte{}, wantErr: joseInvalidClearBytes},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, _, err := EncryptBytesWithContext([]joseJwk.Key{encryptJWK}, tc.clearBytes, nil)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestDecryptBytesWithContext_RejectsEmptyMessages(t *testing.T) {
	t.Parallel()

	_, _, decryptJWK, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)

	tests := []struct {
		name    string
		message []byte
		wantErr string
	}{
		{name: joseNilMessageTestCase, message: nil, wantErr: joseInvalidJWEMessageBytes},
		{name: joseEmptyMessageTestCase, message: []byte{}, wantErr: joseInvalidJWEMessageBytes},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := DecryptBytesWithContext([]joseJwk.Key{decryptJWK}, tc.message, nil)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestSignBytes_RejectsEmptyPayloads(t *testing.T) {
	t.Parallel()

	_, signingJWK, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)

	tests := []struct {
		name       string
		clearBytes []byte
		wantErr    string
	}{
		{name: joseNilPayloadTestCase, clearBytes: nil, wantErr: joseInvalidClearBytes},
		{name: joseEmptyPayloadTestCase, clearBytes: []byte{}, wantErr: joseInvalidClearBytes},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, _, err := SignBytes([]joseJwk.Key{signingJWK}, tc.clearBytes)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestVerifyBytes_RejectsEmptyMessages(t *testing.T) {
	t.Parallel()

	_, _, verificationJWK, _, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)

	tests := []struct {
		name    string
		message []byte
		wantErr string
	}{
		{name: joseNilMessageTestCase, message: nil, wantErr: joseInvalidJWSMessageBytes},
		{name: joseEmptyMessageTestCase, message: []byte{}, wantErr: joseInvalidJWSMessageBytes},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := VerifyBytes([]joseJwk.Key{verificationJWK}, tc.message)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestEncryptAndSignBytes_RejectJWKsWithoutKeyIDs(t *testing.T) {
	t.Parallel()

	_, encryptJWK, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)
	require.NoError(t, encryptJWK.Remove(joseJwk.KeyIDKey))

	_, _, err = EncryptBytes([]joseJwk.Key{encryptJWK}, []byte(joseMessagePayload))
	require.ErrorContains(t, err, "JWK 0 invalid")

	_, signingJWK, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)
	require.NoError(t, signingJWK.Remove(joseJwk.KeyIDKey))

	_, _, err = SignBytes([]joseJwk.Key{signingJWK}, []byte(joseMessagePayload))
	require.ErrorContains(t, err, "JWK 0 invalid")
}

func TestExtractKty_GetError(t *testing.T) {
	t.Parallel()

	_, err := ExtractKty(mockUnknownKey{})
	require.ErrorContains(t, err, "failed to get kty header")
}

func TestJWEHeadersString_MarshalError(t *testing.T) {
	t.Parallel()

	_, encryptJWK, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)
	message, _, err := EncryptBytes([]joseJwk.Key{encryptJWK}, []byte(joseMessagePayload))
	require.NoError(t, err)
	require.NoError(t, message.ProtectedHeaders().Set("invalid-value", make(chan struct{})))

	_, err = JWEHeadersString(message)
	require.ErrorContains(t, err, "failed to marshall JWE headers")
}

func TestBuildJWK_ImportError(t *testing.T) {
	t.Parallel()

	_, err := BuildJWK(KtyOCT, make(chan struct{}), nil)
	require.ErrorContains(t, err, "failed to import")
}

func TestEncryptBytes_RejectsMultipleKeyAlgorithms(t *testing.T) {
	t.Parallel()

	_, encryptJWK1, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)
	_, encryptJWK2, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA128GCM, &AlgA256KW)
	require.NoError(t, err)
	_, _, err = EncryptBytes([]joseJwk.Key{encryptJWK1, encryptJWK2}, []byte(joseMessagePayload))
	require.ErrorContains(t, err, "only one unique 'enc' attribute")

	_, encryptJWK3, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)
	_, encryptJWK4, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA128KW)
	require.NoError(t, err)
	_, _, err = EncryptBytes([]joseJwk.Key{encryptJWK3, encryptJWK4}, []byte(joseMessagePayload))
	require.ErrorContains(t, err, "only one unique 'alg' attribute")
}

func TestDecryptBytes_RejectsMultipleKeyAlgorithms(t *testing.T) {
	t.Parallel()

	_, encryptJWK, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)
	_, encrypted, err := EncryptBytes([]joseJwk.Key{encryptJWK}, []byte(joseMessagePayload))
	require.NoError(t, err)

	_, decryptJWK1, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)
	_, decryptJWK2, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA128GCM, &AlgA256KW)
	require.NoError(t, err)
	_, err = DecryptBytes([]joseJwk.Key{decryptJWK1, decryptJWK2}, encrypted)
	require.ErrorContains(t, err, "only one unique 'enc' attribute")

	_, decryptJWK3, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)
	_, decryptJWK4, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA128KW)
	require.NoError(t, err)
	_, err = DecryptBytes([]joseJwk.Key{decryptJWK3, decryptJWK4}, encrypted)
	require.ErrorContains(t, err, "only one unique 'alg' attribute")
}

func TestSignBytesAndVerifyBytes_RejectMultipleAlgorithms(t *testing.T) {
	t.Parallel()

	_, hmacJWK, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)
	_, ecdsaJWK, _, _, _, err := GenerateJWSJWKForAlg(&AlgES256)
	require.NoError(t, err)

	_, _, err = SignBytes([]joseJwk.Key{hmacJWK, ecdsaJWK}, []byte(joseMessagePayload))
	require.ErrorContains(t, err, "only one unique 'alg' attribute")

	_, signed, err := SignBytes([]joseJwk.Key{hmacJWK}, []byte(joseMessagePayload))
	require.NoError(t, err)
	_, _, ecdsaVerifyJWK, _, _, err := GenerateJWSJWKForAlg(&AlgES256)
	require.NoError(t, err)
	_, err = VerifyBytes([]joseJwk.Key{hmacJWK, ecdsaVerifyJWK}, signed)
	require.ErrorContains(t, err, "only one unique 'alg' attribute")
}

func TestJWSHeadersString_MarshalError(t *testing.T) {
	t.Parallel()

	_, signingJWK, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)
	message, _, err := SignBytes([]joseJwk.Key{signingJWK}, []byte(joseMessagePayload))
	require.NoError(t, err)
	require.NoError(t, message.Signatures()[0].ProtectedHeaders().Set("invalid-value", make(chan struct{})))

	_, err = JWSHeadersString(message)
	require.ErrorContains(t, err, "failed to marshal JWS headers")
}

func TestJWSKeyValidation_ECDSAErrors(t *testing.T) {
	t.Parallel()

	_, err := validateOrGenerateJWSEcdsaJWK(
		&cryptoutilSharedCryptoKeygen.KeyPair{
			Private: &ecdsa.PrivateKey{},
			Public:  "invalid public key",
		},
		AlgES256,
		elliptic.P256(),
	)
	require.ErrorContains(t, err, "invalid key type")
}

func TestJWEKeyValidation_ECDHGenerationError(t *testing.T) {
	t.Parallel()

	enc := EncA256GCM
	alg := AlgECDHES
	_, err := validateOrGenerateJWEEcdhJWK(
		nil,
		&enc,
		&alg,
		failingECDHCurve{Curve: ecdh.P521()},
		&enc,
	)
	require.ErrorIs(t, err, errCurveGeneration)
}
