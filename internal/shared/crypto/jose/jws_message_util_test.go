// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	json "encoding/json"
	"fmt"
	"log"
	"testing"

	googleUuid "github.com/google/uuid"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	joseJws "github.com/lestrrat-go/jwx/v3/jws"

	"github.com/stretchr/testify/require"

	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
)

type happyPathJWSTestCase struct {
	alg          *joseJwa.SignatureAlgorithm
	expectedType joseJwa.KeyType
}

var happyPathJWSTestCases = []happyPathJWSTestCase{
	{alg: &AlgRS256, expectedType: KtyRSA}, // RSA 1.5 & SHA-256
	{alg: &AlgRS384, expectedType: KtyRSA}, // RSA 1.5 & SHA-384
	{alg: &AlgRS512, expectedType: KtyRSA}, // RSA 1.5 & SHA-512
	{alg: &AlgPS256, expectedType: KtyRSA}, // RSA 2.0 & SHA-256
	{alg: &AlgPS384, expectedType: KtyRSA}, // RSA 2.0 & SHA-384
	{alg: &AlgPS512, expectedType: KtyRSA}, // RSA 2.0 & SHA-512
	{alg: &AlgES256, expectedType: KtyEC},  // EC P-256 & SHA-256
	{alg: &AlgES384, expectedType: KtyEC},  // EC P-394 & SHA-384
	{alg: &AlgES512, expectedType: KtyEC},  // EC P-521 & SHA-512
	{alg: &AlgHS256, expectedType: KtyOCT}, // HMAC with SHA-256 & SHA-512
	{alg: &AlgHS384, expectedType: KtyOCT}, // HMAC with SHA-384 & SHA-512
	{alg: &AlgHS512, expectedType: KtyOCT}, // HMAC with SHA-512 & SHA-512
	{alg: &AlgEdDSA, expectedType: KtyOKP}, // ED25519 & SHA-256
}

func TestSignBytes_NilJWKs(t *testing.T) {
	t.Parallel()

	clearBytes := []byte("test message")
	_, _, err := SignBytes(nil, clearBytes)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWKs")
	require.ErrorIs(t, err, cryptoutilSharedApperr.ErrCantBeNil)
}

func TestSignBytes_EmptyJWKs(t *testing.T) {
	t.Parallel()

	jwks := []joseJwk.Key{}
	clearBytes := []byte("test message")
	_, _, err := SignBytes(jwks, clearBytes)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWKs")
	require.ErrorIs(t, err, cryptoutilSharedApperr.ErrCantBeEmpty)
}

func TestSignBytes_NilClearBytes(t *testing.T) {
	t.Parallel()

	_, nonPublicJWK, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)

	jwks := []joseJwk.Key{nonPublicJWK}

	_, _, err = SignBytes(jwks, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid clearBytes")
	require.ErrorIs(t, err, cryptoutilSharedApperr.ErrCantBeNil)
}

func TestSignBytes_EmptyClearBytes(t *testing.T) {
	t.Parallel()

	_, nonPublicJWK, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)

	jwks := []joseJwk.Key{nonPublicJWK}

	clearBytes := []byte{}
	_, _, err = SignBytes(jwks, clearBytes)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid clearBytes")
	require.ErrorIs(t, err, cryptoutilSharedApperr.ErrCantBeEmpty)
}

func TestSignBytes_NonSignJWK(t *testing.T) {
	t.Parallel()

	_, nonPublicJWK, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)

	jwks := []joseJwk.Key{nonPublicJWK}

	clearBytes := []byte("test message")
	_, _, err = SignBytes(jwks, clearBytes)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWK")
}

func TestSignBytes_MultipleAlgs(t *testing.T) {
	t.Parallel()

	_, nonPublicJWK1, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)

	_, nonPublicJWK2, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS512)
	require.NoError(t, err)

	jwks := []joseJwk.Key{nonPublicJWK1, nonPublicJWK2}
	clearBytes := []byte("test message")
	_, _, err = SignBytes(jwks, clearBytes)
	require.Error(t, err)
	require.Contains(t, err.Error(), "only one unique 'alg' attribute is allowed")
}

func TestVerifyBytes_NilJWKs(t *testing.T) {
	t.Parallel()

	jwsMessageBytes := []byte("dummy")
	_, err := VerifyBytes(nil, jwsMessageBytes)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWKs")
	require.ErrorIs(t, err, cryptoutilSharedApperr.ErrCantBeNil)
}

func TestVerifyBytes_EmptyJWKs(t *testing.T) {
	t.Parallel()

	jwks := []joseJwk.Key{}
	jwsMessageBytes := []byte("dummy")
	_, err := VerifyBytes(jwks, jwsMessageBytes)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWKs")
	require.ErrorIs(t, err, cryptoutilSharedApperr.ErrCantBeEmpty)
}

func TestVerifyBytes_NilMessageBytes(t *testing.T) {
	t.Parallel()

	_, nonPublicJWK, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)

	jwks := []joseJwk.Key{nonPublicJWK}

	_, err = VerifyBytes(jwks, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid jwsMessageBytes")
	require.ErrorIs(t, err, cryptoutilSharedApperr.ErrCantBeNil)
}

func TestVerifyBytes_EmptyMessageBytes(t *testing.T) {
	t.Parallel()

	_, nonPublicJWK, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)

	jwks := []joseJwk.Key{nonPublicJWK}

	jwsMessageBytes := []byte{}
	_, err = VerifyBytes(jwks, jwsMessageBytes)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid jwsMessageBytes")
	require.ErrorIs(t, err, cryptoutilSharedApperr.ErrCantBeEmpty)
}

func TestVerifyBytes_NonVerifyJWK(t *testing.T) {
	t.Parallel()

	_, nonPublicJWK, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)

	jwks := []joseJwk.Key{nonPublicJWK}

	jwsMessageBytes := []byte("dummy")
	_, err = VerifyBytes(jwks, jwsMessageBytes)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWK")
}

func TestVerifyBytes_InvalidMessageBytes(t *testing.T) {
	t.Parallel()

	_, nonPublicJWK, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)

	jwks := []joseJwk.Key{nonPublicJWK}

	jwsMessageBytes := []byte("invalid-jws-message")
	_, err = VerifyBytes(jwks, jwsMessageBytes)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse JWS message bytes")
}

func Test_HappyPath_NonJWKGenService_JWS_JWK_SignVerifyBytes(t *testing.T) {
	t.Parallel()

	for _, testCase := range happyPathJWSTestCases {
		plaintext := fmt.Appendf(nil, "Hello world alg=%s!", testCase.alg)
		t.Run(fmt.Sprintf("%v", testCase.alg), func(t *testing.T) {
			t.Parallel()

			jwsJWKKid, nonPublicJWSJWK, publicJWSJWK, clearNonPublicJWSJWKBytes, _, err := GenerateJWSJWKForAlg(testCase.alg)
			require.NoError(t, err)
			require.NotEmpty(t, jwsJWKKid)
			require.NotNil(t, nonPublicJWSJWK)

			require.NotEmpty(t, clearNonPublicJWSJWKBytes)
			log.Printf("Generated: %s", clearNonPublicJWSJWKBytes)

			requireJWSJWKHeaders(t, nonPublicJWSJWK, OpsSigVer, &testCase)

			if publicJWSJWK != nil {
				requireJWSJWKHeaders(t, publicJWSJWK, OpsVer, &testCase)
			}

			// isSignJWK, err := IsSignJWK(nonPublicJWSJWK)
			// require.NoError(t, err, "failed to validate nonPublicJWSJWK")
			// require.True(t, isSignJWK, "nonPublicJWSJWK must be an sign JWK")

			jwsMessage, encodedJWSMessage, err := SignBytes([]joseJwk.Key{nonPublicJWSJWK}, plaintext)
			require.NoError(t, err)
			require.NotEmpty(t, encodedJWSMessage)
			log.Printf("JWS Message: %s", string(encodedJWSMessage))

			requireJWSMessageHeaders(t, jwsMessage, jwsJWKKid, &testCase)

			// isVerifyJWK, err := IsVerifyJWK(publicJWSJWK)
			// require.NoError(t, err, "failed to validate publicJWSJWK")
			// require.True(t, isVerifyJWK, "publicJWSJWK must be an verify JWK")
			isSymmetric, err := IsSymmetricJWK(nonPublicJWSJWK)
			require.NoError(t, err, "failed to validate nonPublicJWSJWK")

			if isSymmetric {
				verified, err := VerifyBytes([]joseJwk.Key{nonPublicJWSJWK}, encodedJWSMessage)
				require.NoError(t, err)
				require.NotNil(t, verified)
			} else {
				verified, err := VerifyBytes([]joseJwk.Key{publicJWSJWK}, encodedJWSMessage)
				require.NoError(t, err)
				require.NotNil(t, verified)
			}
		})
	}
}

func requireJWSJWKHeaders(t *testing.T, nonPublicJWSJWK joseJwk.Key, expectedJWSJWKOps joseJwk.KeyOperationList, testCase *happyPathJWSTestCase) {
	t.Helper()

	var actualJWKAlg joseJwa.KeyAlgorithm

	require.NoError(t, nonPublicJWSJWK.Get(joseJwk.AlgorithmKey, &actualJWKAlg))
	require.Equal(t, *testCase.alg, actualJWKAlg)

	var actualJWKUse string

	require.NoError(t, nonPublicJWSJWK.Get(joseJwk.KeyUsageKey, &actualJWKUse))
	require.Equal(t, joseJwk.ForSignature.String(), actualJWKUse)

	var actualJWKOps joseJwk.KeyOperationList

	require.NoError(t, nonPublicJWSJWK.Get(joseJwk.KeyOpsKey, &actualJWKOps))
	require.Equal(t, expectedJWSJWKOps, actualJWKOps)

	var actualJWKKty joseJwa.KeyType

	require.NoError(t, nonPublicJWSJWK.Get(joseJwk.KeyTypeKey, &actualJWKKty))
	require.Equal(t, testCase.expectedType, actualJWKKty)
}

func requireJWSMessageHeaders(t *testing.T, jwsMessage *joseJws.Message, jwsJWKKid *googleUuid.UUID, testCase *happyPathJWSTestCase) {
	t.Helper()

	jwsHeaders := jwsMessage.Signatures()[0].ProtectedHeaders()
	encodedJWEHeaders, err := json.Marshal(jwsHeaders)
	require.NoError(t, err)
	log.Printf("JWS Message Headers: %v", string(encodedJWEHeaders))

	var actualJWEKid string

	require.NoError(t, jwsHeaders.Get(joseJwk.KeyIDKey, &actualJWEKid))
	require.NotEmpty(t, actualJWEKid)
	require.Equal(t, jwsJWKKid.String(), actualJWEKid)

	var actualJWSAlg joseJwa.KeyAlgorithm

	require.NoError(t, jwsHeaders.Get(joseJwk.AlgorithmKey, &actualJWSAlg))
	require.Equal(t, *testCase.alg, actualJWSAlg)
}

func Test_SadPath_SignBytes_NilKey(t *testing.T) {
	t.Parallel()

	_, _, err := SignBytes(nil, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWKs")
}

func Test_SadPath_SignBytes_EmptyKeys(t *testing.T) {
	t.Parallel()

	_, _, err := SignBytes([]joseJwk.Key{}, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWKs")
}

func Test_SadPath_SignBytes_NilClearBytes(t *testing.T) {
	t.Parallel()

	_, jwk, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)

	_, _, err = SignBytes([]joseJwk.Key{jwk}, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid clearBytes")
}

func Test_SadPath_SignBytes_EmptyClearBytes(t *testing.T) {
	t.Parallel()

	_, jwk, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)

	_, _, err = SignBytes([]joseJwk.Key{jwk}, []byte{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid clearBytes")
}

func Test_SadPath_SignBytes_InvalidJWK_NotSignKey(t *testing.T) {
	t.Parallel()

	// Generate encrypt key (not sign key) - use AES key which has enc operation.
	_, encryptJWK, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA128GCM, &AlgA128KW)
	require.NoError(t, err)

	_, _, err = SignBytes([]joseJwk.Key{encryptJWK}, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWK")
}

func Test_SadPath_SignBytes_MultipleAlgorithms(t *testing.T) {
	t.Parallel()

	// Generate two JWKs with different algorithms.
	_, jwk1, _, _, _, err := GenerateJWSJWKForAlg(&AlgRS256)
	require.NoError(t, err)

	_, jwk2, _, _, _, err := GenerateJWSJWKForAlg(&AlgES256)
	require.NoError(t, err)

	_, _, err = SignBytes([]joseJwk.Key{jwk1, jwk2}, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "only one unique 'alg' attribute is allowed")
}

func Test_SadPath_VerifyBytes_NilKey(t *testing.T) {
	t.Parallel()

	_, err := VerifyBytes(nil, []byte("ciphertext"))
	require.Error(t, err)
}

func Test_SadPath_VerifyBytes_EmptyJWKs(t *testing.T) {
	t.Parallel()

	_, err := VerifyBytes([]joseJwk.Key{}, []byte("ciphertext"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWKs")
}

func Test_SadPath_VerifyBytes_NilMessageBytes(t *testing.T) {
	t.Parallel()

	kid, nonPublicJWSJWK, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)
	require.NotNil(t, kid)
	require.NotNil(t, nonPublicJWSJWK)

	_, err = VerifyBytes([]joseJwk.Key{nonPublicJWSJWK}, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid jwsMessageBytes")
}

func Test_SadPath_VerifyBytes_EmptyMessageBytes(t *testing.T) {
	t.Parallel()

	kid, nonPublicJWSJWK, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)
	require.NotNil(t, kid)
	require.NotNil(t, nonPublicJWSJWK)

	_, err = VerifyBytes([]joseJwk.Key{nonPublicJWSJWK}, []byte{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid jwsMessageBytes")
}

func Test_SadPath_VerifyBytes_MultipleAlgorithms(t *testing.T) {
	t.Parallel()

	kid1, nonPublicJWSJWK1, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)
	require.NotNil(t, kid1)
	require.NotNil(t, nonPublicJWSJWK1)

	kid2, nonPublicJWSJWK2, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS512)
	require.NoError(t, err)
	require.NotNil(t, kid2)
	require.NotNil(t, nonPublicJWSJWK2)

	_, jwsBytes, err := SignBytes([]joseJwk.Key{nonPublicJWSJWK1}, []byte("message"))
	require.NoError(t, err)

	_, err = VerifyBytes([]joseJwk.Key{nonPublicJWSJWK1, nonPublicJWSJWK2}, jwsBytes)
	require.Error(t, err)
	require.Contains(t, err.Error(), "only one unique 'alg' attribute is allowed")
}

func Test_SadPath_VerifyBytes_InvalidJWSMessage(t *testing.T) {
	t.Parallel()

	kid, nonPublicJWSJWK, _, clearNonPublicJWSJWKBytes, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)
	require.NotNil(t, kid)
	require.NotNil(t, nonPublicJWSJWK)
	isSigntJWK, err := IsSignJWK(nonPublicJWSJWK)
	require.NoError(t, err)
	require.True(t, isSigntJWK)
	require.NotNil(t, clearNonPublicJWSJWKBytes)

	_, err = VerifyBytes([]joseJwk.Key{nonPublicJWSJWK}, []byte("this-is-not-a-valid-jws-message"))
	require.Error(t, err)
}

func Test_VerifyBytes_NonVerifyJWK(t *testing.T) {
	t.Parallel()

	enc := joseJwa.A256GCM()
	alg := joseJwa.DIRECT()
	_, nonPublicJWEJWK, _, _, _, err := GenerateJWEJWKForEncAndAlg(&enc, &alg)
	require.NoError(t, err)

	_, err = VerifyBytes([]joseJwk.Key{nonPublicJWEJWK}, []byte("any-message"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWK")
}

func Test_SadPath_GenerateJWSJWK_UnsupportedAlg(t *testing.T) {
	kid, nonPublicJWSJWK, publicJWSJWK, clearNonPublicJWSJWKBytes, clearPublicJWSJWKBytes, err := GenerateJWSJWKForAlg(&AlgSigInvalid)
	require.Error(t, err)
	require.Equal(t, "invalid JWS JWK headers: unsupported JWS JWK alg: invalid", err.Error())
	require.Nil(t, kid)
	require.Nil(t, nonPublicJWSJWK)
	require.Nil(t, publicJWSJWK)
	require.Nil(t, clearNonPublicJWSJWKBytes)
	require.Nil(t, clearPublicJWSJWKBytes)
}

func Test_SadPath_ConcurrentGenerateJWSJWK_UnsupportedAlg(t *testing.T) {
	nonPublicJWEJWKs, publicJWEJWKs, err := GenerateJWSJWKsForTest(t, 2, &AlgSigInvalid)
	require.Error(t, err)
	require.Equal(t, "unexpected 2 errors: invalid JWS JWK headers: unsupported JWS JWK alg: invalid\ninvalid JWS JWK headers: unsupported JWS JWK alg: invalid", err.Error())
	require.Nil(t, nonPublicJWEJWKs)
	require.Nil(t, publicJWEJWKs)
}

func Test_ExtractKidAlgFromJWSMessage_HappyPath(t *testing.T) {
	t.Parallel()

	// Generate JWK for signing.
	jwsJWKs, _, err := GenerateJWSJWKsForTest(t, 1, &AlgRS256)
	require.NoError(t, err)

	// Sign test data.
	plaintext := []byte("test data")
	jwsMessage, _, err := SignBytes(jwsJWKs, plaintext)
	require.NoError(t, err)

	// Test extraction.
	kid, alg, err := ExtractKidAlgFromJWSMessage(jwsMessage)
	require.NoError(t, err)
	require.NotNil(t, kid)
	require.NotNil(t, alg)

	// Verify values.
	expectedKid, err := ExtractKidUUID(jwsJWKs[0])
	require.NoError(t, err)
	require.Equal(t, expectedKid, kid)
	require.Equal(t, AlgRS256, *alg)
}

func Test_ExtractKidAlgFromJWSMessage_NoSignatures(t *testing.T) {
	t.Parallel()

	// Create JWS message without signatures.
	jwsMessage := joseJws.NewMessage()

	// Test extraction should fail.
	kid, alg, err := ExtractKidAlgFromJWSMessage(jwsMessage)
	require.Error(t, err)
	require.Nil(t, kid)
	require.Nil(t, alg)
	require.Contains(t, err.Error(), "JWS message has no signatures")
}

func Test_JWSHeadersString_HappyPath(t *testing.T) {
	t.Parallel()

	// Generate JWK for signing.
	jwsJWKs, _, err := GenerateJWSJWKsForTest(t, 1, &AlgRS256)
	require.NoError(t, err)

	// Sign test data.
	plaintext := []byte("test data")
	jwsMessage, _, err := SignBytes(jwsJWKs, plaintext)
	require.NoError(t, err)

	// Test string conversion.
	headersStr, err := JWSHeadersString(jwsMessage)
	require.NoError(t, err)
	require.NotEmpty(t, headersStr)
	require.Contains(t, headersStr, "alg")
	require.Contains(t, headersStr, "kid")
}

func Test_LogJWSInfo_NilMessage(t *testing.T) {
	t.Parallel()

	// Test with nil message.
	err := LogJWSInfo(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "jwsMessage is nil")
}

func Test_LogJWSInfo_NoSignatures(t *testing.T) {
	t.Parallel()

	// Create JWS message without signatures.
	jwsMessage := joseJws.NewMessage()

	// Test logging should fail.
	err := LogJWSInfo(jwsMessage)
	require.Error(t, err)
	require.Contains(t, err.Error(), "jwsMessage has no signatures")
}

func Test_LogJWSInfo_AllHeaders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		headers map[string]any
		alg     *joseJwa.SignatureAlgorithm
	}{
		{
			name: "RS256 with all headers",
			alg:  &AlgRS256,
			headers: map[string]any{
				joseJws.KeyIDKey:       "test-kid-" + googleUuid.NewString(),
				joseJws.TypeKey:        "JWT",
				joseJws.ContentTypeKey: "application/json",
				joseJws.JWKSetURLKey:   "https://example.com/jwks",
				joseJws.X509URLKey:     "https://example.com/x509",
				joseJws.CriticalKey:    []string{"exp", "nbf"},
				"custom-header":        "custom-value",
			},
		},
		{
			name: "ES256 with minimal headers",
			alg:  &AlgES256,
			headers: map[string]any{
				joseJws.KeyIDKey: "test-kid-" + googleUuid.NewString(),
			},
		},
		{
			name: "HS256 with type only",
			alg:  &AlgHS256,
			headers: map[string]any{
				joseJws.TypeKey: "JWT",
			},
		},
		{
			name: "EdDSA with content type",
			alg:  &AlgEdDSA,
			headers: map[string]any{
				joseJws.ContentTypeKey: "application/jose+json",
			},
		},
		{
			name:    "PS256 algorithm only",
			alg:     &AlgPS256,
			headers: map[string]any{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create test payload.
			payload := []byte(`{"sub":"test-user","iat":1234567890}`)

			// Generate signing key for algorithm.
			_, signingJWK, _, _, _, err := GenerateJWSJWKForAlg(tc.alg)
			require.NoError(t, err)
			require.NotNil(t, signingJWK)

			// Create JWS message with headers.
			jwsHeaders := joseJws.NewHeaders()
			for key, value := range tc.headers {
				err = jwsHeaders.Set(key, value)
				require.NoError(t, err)
			}

			signedMessage, err := joseJws.Sign(
				payload,
				joseJws.WithKey(*tc.alg, signingJWK, joseJws.WithProtectedHeaders(jwsHeaders)),
			)
			require.NoError(t, err)

			// Parse signed message.
			jwsMessage, err := joseJws.Parse(signedMessage)
			require.NoError(t, err)

			// Test LogJWSInfo.
			err = LogJWSInfo(jwsMessage)
			require.NoError(t, err)

			// Verify message structure.
			require.Len(t, jwsMessage.Signatures(), 1, "should have exactly one signature")
			require.Equal(t, payload, jwsMessage.Payload(), "payload should match")
		})
	}
}

// Test_JWSHeadersString_MultipleSignatures tests JWSHeadersString with multiple signatures.
func Test_JWSHeadersString_MultipleSignatures(t *testing.T) {
	t.Parallel()

	payload := []byte("test payload")

	// Generate two signing keys.
	_, jwk1, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)

	_, jwk2, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS512)
	require.NoError(t, err)

	// Sign with two keys using JSON serialization (required for multiple signatures).
	signedMessage, err := joseJws.Sign(
		payload,
		joseJws.WithKey(AlgHS256, jwk1),
		joseJws.WithKey(AlgHS512, jwk2),
		joseJws.WithJSON(), // Required for multiple signatures
	)
	require.NoError(t, err)

	// Parse signed message.
	jwsMessage, err := joseJws.Parse(signedMessage)
	require.NoError(t, err)
	require.Len(t, jwsMessage.Signatures(), 2, "should have two signatures")

	// Test JWSHeadersString with multiple signatures.
	headersStr, err := JWSHeadersString(jwsMessage)
	require.NoError(t, err)
	require.NotEmpty(t, headersStr)
	require.Contains(t, headersStr, "\n", "should contain newline separator for multiple signatures")
}
