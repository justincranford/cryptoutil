// Copyright (c) 2025 Justin Cranford

package dpop

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"encoding/base64"
	"testing"
	"time"

	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	joseJws "github.com/lestrrat-go/jwx/v3/jws"
	joseJwt "github.com/lestrrat-go/jwx/v3/jwt"
	"github.com/stretchr/testify/require"
)

// buildProofWithHeaders builds a DPoP JWT signed with the given private key, using the given headers.
func buildProofWithHeaders(t *testing.T, privateKey joseJwk.Key, token joseJwt.Token, headers joseJws.Headers) string {
	t.Helper()

	signed, err := joseJwt.Sign(token, joseJwt.WithKey(joseJwa.ES256(), privateKey, joseJws.WithProtectedHeaders(headers)))
	require.NoError(t, err)

	return string(signed)
}

// buildBaseToken builds a jwt.Token with the standard dpop claims set.
func buildBaseToken(t *testing.T) joseJwt.Token {
	t.Helper()

	token := joseJwt.New()
	require.NoError(t, token.Set(cryptoutilSharedMagic.ClaimJti, "test-jti"))
	require.NoError(t, token.Set("htm", "POST"))
	require.NoError(t, token.Set("htu", "https://example.com/token"))
	require.NoError(t, token.Set(cryptoutilSharedMagic.ClaimIat, time.Now().UTC().Unix()))

	return token
}

// buildValidHeadersWith builds JWS headers with the given publicKey and custom tweaks applied via a callback.
func buildValidHeaders(t *testing.T, publicKey joseJwk.Key) joseJws.Headers {
	t.Helper()

	headers := joseJws.NewHeaders()
	require.NoError(t, headers.Set("typ", "dpop+jwt"))
	require.NoError(t, headers.Set("alg", joseJwa.ES256()))
	require.NoError(t, headers.Set("jwk", publicKey))

	return headers
}

// generateTestKeys generates an ES256 private+public JWK pair for testing.
func generateTestKeys(t *testing.T) (joseJwk.Key, joseJwk.Key) {
	t.Helper()

	alg := joseJwa.ES256()
	_, privateKey, publicKey, _, _, err := cryptoutilSharedCryptoJose.GenerateJWSJWKForAlg(&alg)
	require.NoError(t, err)

	return privateKey, publicKey
}

func TestValidateProof_EmptyHTTPMethod(t *testing.T) {
	t.Parallel()

	privateKey, publicKey := generateTestKeys(t)
	dpopHeader := buildValidProof(t, privateKey, publicKey, "POST", "https://example.com/token", "")

	proof, err := ValidateProof(dpopHeader, "", "https://example.com/token", "")
	require.Error(t, err)
	require.ErrorContains(t, err, "HTTP method is required")
	require.Nil(t, proof)
}

func TestValidateProof_EmptyHTTPURI(t *testing.T) {
	t.Parallel()

	privateKey, publicKey := generateTestKeys(t)
	dpopHeader := buildValidProof(t, privateKey, publicKey, "POST", "https://example.com/token", "")

	proof, err := ValidateProof(dpopHeader, "POST", "", "")
	require.Error(t, err)
	require.ErrorContains(t, err, "HTTP URI is required")
	require.Nil(t, proof)
}

func TestValidateProof_InvalidJWT(t *testing.T) {
	t.Parallel()

	proof, err := ValidateProof("this-is-not-a-valid-jwt", "POST", "https://example.com/token", "")
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to parse DPoP JWT")
	require.Nil(t, proof)
}

func TestValidateProof_WrongTypHeader(t *testing.T) {
	t.Parallel()

	privateKey, publicKey := generateTestKeys(t)

	token := buildBaseToken(t)
	headers := buildValidHeaders(t, publicKey)
	require.NoError(t, headers.Set("typ", "application/jwt")) // wrong typ

	dpopHeader := buildProofWithHeaders(t, privateKey, token, headers)

	proof, err := ValidateProof(dpopHeader, "POST", "https://example.com/token", "")
	require.Error(t, err)
	require.ErrorContains(t, err, "typ header must be 'dpop+jwt'")
	require.Nil(t, proof)
}

func TestValidateProof_MissingJWKHeader(t *testing.T) {
	t.Parallel()

	// Build a raw ECDSA key and JWK key – sign without embedding JWK in header.
	rawKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	privateJWK, err := joseJwk.Import(rawKey)
	require.NoError(t, err)

	// Build headers without jwk field.
	headers := joseJws.NewHeaders()
	require.NoError(t, headers.Set("typ", "dpop+jwt"))
	require.NoError(t, headers.Set("alg", joseJwa.ES256()))
	// Intentionally omit: headers.Set("jwk", ...)

	token := buildBaseToken(t)
	dpopHeader := buildProofWithHeaders(t, privateJWK, token, headers)

	proof, err := ValidateProof(dpopHeader, "POST", "https://example.com/token", "")
	require.Error(t, err)
	require.ErrorContains(t, err, "must include jwk header")
	require.Nil(t, proof)
}

func TestValidateProof_MissingJTI(t *testing.T) {
	t.Parallel()

	privateKey, publicKey := generateTestKeys(t)

	// Build token without jti.
	token := joseJwt.New()
	require.NoError(t, token.Set("htm", "POST"))
	require.NoError(t, token.Set("htu", "https://example.com/token"))
	require.NoError(t, token.Set(cryptoutilSharedMagic.ClaimIat, time.Now().UTC().Unix()))

	headers := buildValidHeaders(t, publicKey)
	dpopHeader := buildProofWithHeaders(t, privateKey, token, headers)

	proof, err := ValidateProof(dpopHeader, "POST", "https://example.com/token", "")
	require.Error(t, err)
	require.ErrorContains(t, err, "must include jti claim")
	require.Nil(t, proof)
}

func TestValidateProof_MissingHTM(t *testing.T) {
	t.Parallel()

	privateKey, publicKey := generateTestKeys(t)

	// Build token without htm.
	token := joseJwt.New()
	require.NoError(t, token.Set(cryptoutilSharedMagic.ClaimJti, "test-jti"))
	require.NoError(t, token.Set("htu", "https://example.com/token"))
	require.NoError(t, token.Set(cryptoutilSharedMagic.ClaimIat, time.Now().UTC().Unix()))

	headers := buildValidHeaders(t, publicKey)
	dpopHeader := buildProofWithHeaders(t, privateKey, token, headers)

	proof, err := ValidateProof(dpopHeader, "POST", "https://example.com/token", "")
	require.Error(t, err)
	require.ErrorContains(t, err, "must include htm claim")
	require.Nil(t, proof)
}

func TestValidateProof_MissingHTU(t *testing.T) {
	t.Parallel()

	privateKey, publicKey := generateTestKeys(t)

	// Build token without htu.
	token := joseJwt.New()
	require.NoError(t, token.Set(cryptoutilSharedMagic.ClaimJti, "test-jti"))
	require.NoError(t, token.Set("htm", "POST"))
	require.NoError(t, token.Set(cryptoutilSharedMagic.ClaimIat, time.Now().UTC().Unix()))

	headers := buildValidHeaders(t, publicKey)
	dpopHeader := buildProofWithHeaders(t, privateKey, token, headers)

	proof, err := ValidateProof(dpopHeader, "POST", "https://example.com/token", "")
	require.Error(t, err)
	require.ErrorContains(t, err, "must include htu claim")
	require.Nil(t, proof)
}

func TestValidateProof_MissingIAT(t *testing.T) {
	t.Parallel()

	privateKey, publicKey := generateTestKeys(t)

	// Build token without iat.
	token := joseJwt.New()
	require.NoError(t, token.Set(cryptoutilSharedMagic.ClaimJti, "test-jti"))
	require.NoError(t, token.Set("htm", "POST"))
	require.NoError(t, token.Set("htu", "https://example.com/token"))
	// Intentionally omit iat

	headers := buildValidHeaders(t, publicKey)
	dpopHeader := buildProofWithHeaders(t, privateKey, token, headers)

	proof, err := ValidateProof(dpopHeader, "POST", "https://example.com/token", "")
	require.Error(t, err)
	require.ErrorContains(t, err, "must include iat claim")
	require.Nil(t, proof)
}

func TestValidateProof_FutureIAT(t *testing.T) {
	t.Parallel()

	privateKey, publicKey := generateTestKeys(t)

	// Build token with iat 10 minutes in the future.
	token := joseJwt.New()
	require.NoError(t, token.Set(cryptoutilSharedMagic.ClaimJti, "test-jti"))
	require.NoError(t, token.Set("htm", "POST"))
	require.NoError(t, token.Set("htu", "https://example.com/token"))
	require.NoError(t, token.Set(cryptoutilSharedMagic.ClaimIat, time.Now().UTC().Add(cryptoutilSharedMagic.JoseJADefaultMaxMaterials*time.Minute).Unix()))

	headers := buildValidHeaders(t, publicKey)
	dpopHeader := buildProofWithHeaders(t, privateKey, token, headers)

	proof, err := ValidateProof(dpopHeader, "POST", "https://example.com/token", "")
	require.Error(t, err)
	require.ErrorContains(t, err, "iat claim is outside acceptable time window")
	require.Nil(t, proof)
}

func TestValidateProof_OldIAT(t *testing.T) {
	t.Parallel()

	privateKey, publicKey := generateTestKeys(t)

	// Build token with iat 10 minutes in the past.
	token := joseJwt.New()
	require.NoError(t, token.Set(cryptoutilSharedMagic.ClaimJti, "test-jti"))
	require.NoError(t, token.Set("htm", "POST"))
	require.NoError(t, token.Set("htu", "https://example.com/token"))
	require.NoError(t, token.Set(cryptoutilSharedMagic.ClaimIat, time.Now().UTC().Add(-cryptoutilSharedMagic.JoseJADefaultMaxMaterials*time.Minute).Unix()))

	headers := buildValidHeaders(t, publicKey)
	dpopHeader := buildProofWithHeaders(t, privateKey, token, headers)

	proof, err := ValidateProof(dpopHeader, "POST", "https://example.com/token", "")
	require.Error(t, err)
	require.ErrorContains(t, err, "iat claim is outside acceptable time window")
	require.Nil(t, proof)
}

func TestValidateProof_MissingATHWithAccessToken(t *testing.T) {
	t.Parallel()

	privateKey, publicKey := generateTestKeys(t)

	// Build proof without ath claim, but then pass an access token to ValidateProof.
	// buildBaseToken sets jti/htm/htu/iat but NOT ath.
	token := buildBaseToken(t)
	headers := buildValidHeaders(t, publicKey)
	dpopHeader := buildProofWithHeaders(t, privateKey, token, headers)

	// Provide an access token so ValidateProof expects ath in the JWT.
	proof, err := ValidateProof(dpopHeader, "POST", "https://example.com/token", "some-access-token")
	require.Error(t, err)
	require.ErrorContains(t, err, "must include ath claim when used with access token")
	require.Nil(t, proof)
}

func TestValidateProof_NoneAlgorithm(t *testing.T) {
	t.Parallel()

	// Craft a raw JWS token with alg=none in the protected header manually.
	// The jwx library overrides alg when signing, so we construct the raw base64url
	// encoded header, payload, and empty signature to simulate a "none" algorithm token.
	//
	// Protected header: {"typ":"dpop+jwt","alg":"none"}
	rawHeader := `{"typ":"dpop+jwt","alg":"none"}`
	rawPayload := `{"jti":"test-jti","htm":"POST","htu":"https://example.com/token","iat":9999999999}`
	// JWS compact serialization: base64url(header).base64url(payload).base64url(signature)
	// For alg=none, signature is empty.
	encodedHeader := base64.RawURLEncoding.EncodeToString([]byte(rawHeader))
	encodedPayload := base64.RawURLEncoding.EncodeToString([]byte(rawPayload))
	// Empty signature for none algorithm
	noneToken := encodedHeader + "." + encodedPayload + "."

	proof, err := ValidateProof(noneToken, "POST", "https://example.com/token", "")
	require.Error(t, err)
	require.ErrorContains(t, err, "alg must not be 'none'")
	require.Nil(t, proof)
}

func TestIsDPoPBound_CNFWithoutJKT(t *testing.T) {
	t.Parallel()

	// Build a token where cnf exists but does not have a jkt key.
	alg := joseJwa.ES256()
	_, privateKey, _, _, _, err := cryptoutilSharedCryptoJose.GenerateJWSJWKForAlg(&alg)
	require.NoError(t, err)

	token := joseJwt.New()
	require.NoError(t, token.Set(cryptoutilSharedMagic.ClaimSub, "test-subject"))
	require.NoError(t, token.Set("cnf", map[string]any{"x5t#S256": "some-thumbprint"})) // cnf with no jkt

	signed, err := joseJwt.Sign(token, joseJwt.WithKey(joseJwa.ES256(), privateKey))
	require.NoError(t, err)

	bound, jkt, err := IsDPoPBound(string(signed))
	require.NoError(t, err)
	require.False(t, bound)
	require.Empty(t, jkt)
}

func TestIsDPoPBound_CNFJKTNotString(t *testing.T) {
	t.Parallel()

	// Build a token where cnf.jkt is an integer instead of a string.
	alg := joseJwa.ES256()
	_, privateKey, _, _, _, err := cryptoutilSharedCryptoJose.GenerateJWSJWKForAlg(&alg)
	require.NoError(t, err)

	token := joseJwt.New()
	require.NoError(t, token.Set(cryptoutilSharedMagic.ClaimSub, "test-subject"))
	require.NoError(t, token.Set("cnf", map[string]any{"jkt": 12345})) // non-string jkt

	signed, err := joseJwt.Sign(token, joseJwt.WithKey(joseJwa.ES256(), privateKey))
	require.NoError(t, err)

	bound, jkt, err := IsDPoPBound(string(signed))
	require.Error(t, err)
	require.ErrorContains(t, err, "cnf.jkt must be a string")
	require.False(t, bound)
	require.Empty(t, jkt)
}
