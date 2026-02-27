// Copyright (c) 2025 Justin Cranford

package crypto

import (
	"crypto/ed25519"
	crand "crypto/rand"
	rsa "crypto/rsa"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"

	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

// ====================== JWS validate typed nil tests =========================

// TestValidateOrGenerateJWSRSAJWK_TypedNilPrivateKey tests the typed nil
// private key path that passes type assertion but fails nil check.
func TestValidateOrGenerateJWSRSAJWK_TypedNilPrivateKey(t *testing.T) {
	t.Parallel()

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: (*rsa.PrivateKey)(nil),
		Public:  &rsa.PublicKey{},
	}

	validated, err := validateOrGenerateJWSRSAJWK(keyPair, joseJwa.RS256(), cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid nil RSA private key")
}

// TestValidateOrGenerateJWSRSAJWK_TypedNilPublicKey tests the typed nil
// public key path that passes type assertion but fails nil check.
func TestValidateOrGenerateJWSRSAJWK_TypedNilPublicKey(t *testing.T) {
	t.Parallel()

	privateKey, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: privateKey,
		Public:  (*rsa.PublicKey)(nil),
	}

	validated, err := validateOrGenerateJWSRSAJWK(keyPair, joseJwa.RS256(), cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid nil RSA public key")
}

// TestValidateOrGenerateJWSEddsaJWK_TypedNilPrivateKey tests the typed nil
// EdDSA private key path.
func TestValidateOrGenerateJWSEddsaJWK_TypedNilPrivateKey(t *testing.T) {
	t.Parallel()

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: ed25519.PrivateKey(nil),
		Public:  ed25519.PublicKey{},
	}

	validated, err := validateOrGenerateJWSEddsaJWK(keyPair, joseJwa.EdDSA(), cryptoutilSharedMagic.EdCurveEd25519)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid nil Ed29919 private key")
}

// TestValidateOrGenerateJWSEddsaJWK_TypedNilPublicKey tests the typed nil
// EdDSA public key path.
func TestValidateOrGenerateJWSEddsaJWK_TypedNilPublicKey(t *testing.T) {
	t.Parallel()

	_, privateKey, err := ed25519.GenerateKey(crand.Reader)
	require.NoError(t, err)

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: privateKey,
		Public:  ed25519.PublicKey(nil),
	}

	validated, err := validateOrGenerateJWSEddsaJWK(keyPair, joseJwa.EdDSA(), cryptoutilSharedMagic.EdCurveEd25519)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid nil Ed29919 public key")
}

// ====================== JWS wrong key type tests =============================

// TestValidateOrGenerateJWSRSAJWK_WrongPrivateKeyType tests passing an EdDSA
// private key where RSA is expected.
func TestValidateOrGenerateJWSRSAJWK_WrongPrivateKeyType(t *testing.T) {
	t.Parallel()

	_, edPriv, err := ed25519.GenerateKey(crand.Reader)
	require.NoError(t, err)

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: edPriv,
		Public:  &rsa.PublicKey{},
	}

	validated, err := validateOrGenerateJWSRSAJWK(keyPair, joseJwa.RS256(), cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid key type")
}

// TestValidateOrGenerateJWSRSAJWK_WrongPublicKeyType tests passing an EdDSA
// public key where RSA public key is expected.
func TestValidateOrGenerateJWSRSAJWK_WrongPublicKeyType(t *testing.T) {
	t.Parallel()

	rsaKey, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	edPub, _, err := ed25519.GenerateKey(crand.Reader)
	require.NoError(t, err)

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: rsaKey,
		Public:  edPub,
	}

	validated, err := validateOrGenerateJWSRSAJWK(keyPair, joseJwa.RS256(), cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid key type")
}

// TestValidateOrGenerateJWSEddsaJWK_WrongPublicKeyType tests passing an RSA
// public key where EdDSA public key is expected.
func TestValidateOrGenerateJWSEddsaJWK_WrongPublicKeyType(t *testing.T) {
	t.Parallel()

	_, edPriv, err := ed25519.GenerateKey(crand.Reader)
	require.NoError(t, err)

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: edPriv,
		Public:  &rsa.PublicKey{},
	}

	validated, err := validateOrGenerateJWSEddsaJWK(keyPair, joseJwa.EdDSA(), cryptoutilSharedMagic.EdCurveEd25519)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid key type")
}

// TestValidateOrGenerateJWSEddsaJWK_WrongPrivateKeyType tests passing an RSA
// private key where EdDSA private key is expected.
func TestValidateOrGenerateJWSEddsaJWK_WrongPrivateKeyType(t *testing.T) {
	t.Parallel()

	rsaKey, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: rsaKey,
		Public:  ed25519.PublicKey{},
	}

	validated, err := validateOrGenerateJWSEddsaJWK(keyPair, joseJwa.EdDSA(), cryptoutilSharedMagic.EdCurveEd25519)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid key type")
}

// ====================== Extract function edge cases ==========================

// TestExtractAlgFromJWSJWK_NonSignatureAlg tests that a JWK with a
// key encryption algorithm (not signature) returns an error.
func TestExtractAlgFromJWSJWK_NonSignatureAlg(t *testing.T) {
	t.Parallel()

	jwk, err := joseJwk.Import([]byte("0123456789abcdef"))
	require.NoError(t, err)
	require.NoError(t, jwk.Set(joseJwk.AlgorithmKey, joseJwa.A128KW()))

	_, extractErr := ExtractAlgFromJWSJWK(jwk, 0)
	require.Error(t, extractErr)
	require.Contains(t, extractErr.Error(), "not a signature algorithm")
}

// TestExtractAlgFromJWSJWK_MissingAlg tests a JWK with no algorithm set.
func TestExtractAlgFromJWSJWK_MissingAlg(t *testing.T) {
	t.Parallel()

	jwk, err := joseJwk.Import([]byte("0123456789abcdef"))
	require.NoError(t, err)

	_, extractErr := ExtractAlgFromJWSJWK(jwk, 0)
	require.Error(t, extractErr)
	require.Contains(t, extractErr.Error(), "missing algorithm")
}

// TestExtractAlgFromJWSJWK_NilKey tests that nil JWK returns an error.
func TestExtractAlgFromJWSJWK_NilKey(t *testing.T) {
	t.Parallel()

	_, err := ExtractAlgFromJWSJWK(nil, 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "can't be nil")
}

// TestExtractAlgEncFromJWEJWK_NilKey tests that nil JWK returns an error.
func TestExtractAlgEncFromJWEJWK_NilKey(t *testing.T) {
	t.Parallel()

	_, _, err := ExtractAlgEncFromJWEJWK(nil, 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "can't be nil")
}

// TestExtractAlgEncFromJWEJWK_MissingEncAttr tests a JWK missing the enc attribute.
func TestExtractAlgEncFromJWEJWK_MissingEncAttr(t *testing.T) {
	t.Parallel()

	jwk, err := joseJwk.Import([]byte("0123456789abcdef"))
	require.NoError(t, err)

	require.NoError(t, jwk.Set(joseJwk.AlgorithmKey, joseJwa.A128KW()))

	_, _, extractErr := ExtractAlgEncFromJWEJWK(jwk, 0)
	require.Error(t, extractErr)
	require.Contains(t, extractErr.Error(), "'enc' attribute")
}

// TestExtractAlgEncFromJWEJWK_MissingAlgAttr tests a JWK with enc but no alg.
func TestExtractAlgEncFromJWEJWK_MissingAlgAttr(t *testing.T) {
	t.Parallel()

	jwk, err := joseJwk.Import([]byte("0123456789abcdef"))
	require.NoError(t, err)

	require.NoError(t, jwk.Set(cryptoutilSharedMagic.JoseKeyUseEnc, cryptoutilSharedMagic.JoseEncA256GCM))

	_, _, extractErr := ExtractAlgEncFromJWEJWK(jwk, 0)
	require.Error(t, extractErr)
	require.Contains(t, extractErr.Error(), "'alg' attribute")
}

// ====================== Message util edge cases ==============================

// TestSignBytes_DifferentAlgorithms tests that SignBytes rejects JWKs with
// different signature algorithms.
func TestSignBytes_DifferentAlgorithms(t *testing.T) {
	t.Parallel()

	alg1 := joseJwa.RS256()
	_, jwk1, _, _, _, err := GenerateJWSJWKForAlg(&alg1)
	require.NoError(t, err)

	alg2 := joseJwa.EdDSA()
	_, jwk2, _, _, _, err := GenerateJWSJWKForAlg(&alg2)
	require.NoError(t, err)

	_, _, signErr := SignBytes([]joseJwk.Key{jwk1, jwk2}, []byte("test data"))
	require.Error(t, signErr)
	require.Contains(t, signErr.Error(), "only one unique 'alg' attribute is allowed")
}
