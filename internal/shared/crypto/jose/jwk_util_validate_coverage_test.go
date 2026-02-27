// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	"crypto"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"
	"testing"

	cryptoutilOpenapiModel "cryptoutil/api/model"
	joseCert "github.com/lestrrat-go/jwx/v3/cert"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

// mockUnknownKey implements joseJwk.Key with an unknown underlying type.
// This allows testing default branches in type switches that check for
// specific JWK key types (ECDSAPrivateKey, RSAPublicKey, etc.).
type mockUnknownKey struct{}

func (m mockUnknownKey) Has(string) bool                          { return false }
func (m mockUnknownKey) Get(string, any) error                    { return fmt.Errorf("mock key: not implemented") }
func (m mockUnknownKey) Set(string, any) error                    { return nil }
func (m mockUnknownKey) Remove(string) error                      { return nil }
func (m mockUnknownKey) Validate() error                          { return nil }
func (m mockUnknownKey) Thumbprint(crypto.Hash) ([]byte, error)   { return nil, nil }
func (m mockUnknownKey) Keys() []string                           { return nil }
func (m mockUnknownKey) Clone() (joseJwk.Key, error)              { return m, nil }
func (m mockUnknownKey) PublicKey() (joseJwk.Key, error)          { return m, nil }
func (m mockUnknownKey) KeyType() joseJwa.KeyType                 { return joseJwa.InvalidKeyType() }
func (m mockUnknownKey) KeyUsage() (string, bool)                 { return "", false }
func (m mockUnknownKey) KeyOps() (joseJwk.KeyOperationList, bool) { return nil, false }
func (m mockUnknownKey) Algorithm() (joseJwa.KeyAlgorithm, bool)  { return nil, false }
func (m mockUnknownKey) KeyID() (string, bool)                    { return "", false }
func (m mockUnknownKey) X509URL() (string, bool)                  { return "", false }
func (m mockUnknownKey) X509CertChain() (*joseCert.Chain, bool)   { return nil, false }
func (m mockUnknownKey) X509CertThumbprint() (string, bool)       { return "", false }
func (m mockUnknownKey) X509CertThumbprintS256() (string, bool)   { return "", false }

// mockUnknownKeyWithJWSAlg returns a valid SignatureAlgorithm from Algorithm()
// so that ExtractAlgFromJWSJWK succeeds, allowing IsSignJWK/IsVerifyJWK to
// reach the type switch default branch.
type mockUnknownKeyWithJWSAlg struct{ mockUnknownKey }

func (m mockUnknownKeyWithJWSAlg) Algorithm() (joseJwa.KeyAlgorithm, bool) {
	return joseJwa.ES256(), true
}

// mockUnknownKeyWithJWEHeaders returns valid enc and alg headers from Get()
// so that ExtractAlgEncFromJWEJWK succeeds, allowing IsEncryptJWK/IsDecryptJWK
// to reach the type switch default branch.
type mockUnknownKeyWithJWEHeaders struct{ mockUnknownKey }

func (m mockUnknownKeyWithJWEHeaders) Get(name string, dst any) error {
	switch name {
	case cryptoutilSharedMagic.JoseKeyUseEnc:
		if p, ok := dst.(*joseJwa.ContentEncryptionAlgorithm); ok {
			*p = joseJwa.A256GCM()

			return nil
		}
	case joseJwk.AlgorithmKey:
		if p, ok := dst.(*joseJwa.KeyEncryptionAlgorithm); ok {
			*p = joseJwa.A256KW()

			return nil
		}
	}

	return fmt.Errorf("mock key: not implemented for %s", name)
}

// TestIsPublicJWK_UnknownKeyType tests the default branch for IsPublicJWK.
func TestIsPublicJWK_UnknownKeyType(t *testing.T) {
	t.Parallel()

	result, err := IsPublicJWK(mockUnknownKey{})
	require.Error(t, err)
	require.False(t, result)
	require.Contains(t, err.Error(), "unsupported JWK type")
}

// TestIsPrivateJWK_UnknownKeyType tests the default branch for IsPrivateJWK.
func TestIsPrivateJWK_UnknownKeyType(t *testing.T) {
	t.Parallel()

	result, err := IsPrivateJWK(mockUnknownKey{})
	require.Error(t, err)
	require.False(t, result)
	require.Contains(t, err.Error(), "unsupported JWK type")
}

// TestIsAsymmetricJWK_UnknownKeyType tests the default branch for IsAsymmetricJWK.
func TestIsAsymmetricJWK_UnknownKeyType(t *testing.T) {
	t.Parallel()

	result, err := IsAsymmetricJWK(mockUnknownKey{})
	require.Error(t, err)
	require.False(t, result)
	require.Contains(t, err.Error(), "unsupported JWK type")
}

// TestIsSymmetricJWK_UnknownKeyType tests the default branch for IsSymmetricJWK.
func TestIsSymmetricJWK_UnknownKeyType(t *testing.T) {
	t.Parallel()

	result, err := IsSymmetricJWK(mockUnknownKey{})
	require.Error(t, err)
	require.False(t, result)
	require.Contains(t, err.Error(), "unsupported JWK type")
}

// TestIsEncryptJWK_UnknownKeyType tests the default branch for IsEncryptJWK.
// Uses mockUnknownKeyWithJWEHeaders so ExtractAlgEncFromJWEJWK succeeds,
// allowing the type switch default branch to be reached.
func TestIsEncryptJWK_UnknownKeyType(t *testing.T) {
	t.Parallel()

	result, err := IsEncryptJWK(mockUnknownKeyWithJWEHeaders{})
	require.Error(t, err)
	require.False(t, result)
	require.Contains(t, err.Error(), "unsupported JWK type")
}

// TestIsDecryptJWK_UnknownKeyType tests the default branch for IsDecryptJWK.
func TestIsDecryptJWK_UnknownKeyType(t *testing.T) {
	t.Parallel()

	result, err := IsDecryptJWK(mockUnknownKeyWithJWEHeaders{})
	require.Error(t, err)
	require.False(t, result)
	require.Contains(t, err.Error(), "unsupported JWK type")
}

// TestIsSignJWK_UnknownKeyType tests the default branch for IsSignJWK.
// Uses mockUnknownKeyWithJWSAlg so ExtractAlgFromJWSJWK succeeds,
// allowing the type switch default branch to be reached.
func TestIsSignJWK_UnknownKeyType(t *testing.T) {
	t.Parallel()

	result, err := IsSignJWK(mockUnknownKeyWithJWSAlg{})
	require.Error(t, err)
	require.False(t, result)
	require.Contains(t, err.Error(), "unsupported JWK type")
}

// TestIsVerifyJWK_UnknownKeyType tests the default branch for IsVerifyJWK.
func TestIsVerifyJWK_UnknownKeyType(t *testing.T) {
	t.Parallel()

	result, err := IsVerifyJWK(mockUnknownKeyWithJWSAlg{})
	require.Error(t, err)
	require.False(t, result)
	require.Contains(t, err.Error(), "unsupported JWK type")
}

// TestGetGenerateAlgorithmTestProbability_UnknownAlgorithm tests the default branch.
func TestGetGenerateAlgorithmTestProbability_UnknownAlgorithm(t *testing.T) {
	t.Parallel()

	prob := GetGenerateAlgorithmTestProbability(cryptoutilOpenapiModel.GenerateAlgorithm("unknown"))
	require.InDelta(t, cryptoutilSharedMagic.TestProbAlways, prob, 0.001)
}

// TestGetElasticKeyAlgorithmTestProbability_UnknownAlgorithm tests the default branch.
func TestGetElasticKeyAlgorithmTestProbability_UnknownAlgorithm(t *testing.T) {
	t.Parallel()

	prob := GetElasticKeyAlgorithmTestProbability(cryptoutilOpenapiModel.ElasticKeyAlgorithm("unknown"))
	require.InDelta(t, cryptoutilSharedMagic.TestProbAlways, prob, 0.001)
}

// mockGenerateAlgorithmKeyAlg implements joseJwa.KeyAlgorithm with a custom string
// representation that maps to a GenerateAlgorithm value (e.g., "EC/P256").
type mockGenerateAlgorithmKeyAlg string

func (m mockGenerateAlgorithmKeyAlg) String() string     { return string(m) }
func (m mockGenerateAlgorithmKeyAlg) IsDeprecated() bool { return false }

// mockKeyWithGenerateAlg is a mock key whose Algorithm() returns a KeyAlgorithm
// whose String() maps to a value in the generateAlgorithms map.
type mockKeyWithGenerateAlg struct{ mockUnknownKey }

func (m mockKeyWithGenerateAlg) Algorithm() (joseJwa.KeyAlgorithm, bool) {
	return mockGenerateAlgorithmKeyAlg(cryptoutilOpenapiModel.ECP256), true
}

// TestExtractAlg_HappyPath covers the successful return path in ExtractAlg.
// Uses a mock key whose Algorithm().String() returns "EC/P256" which IS in the
// generateAlgorithms map, unlike JOSE algorithms (ES256) which are NOT.
func TestExtractAlg_HappyPath(t *testing.T) {
	t.Parallel()

	result, err := ExtractAlg(mockKeyWithGenerateAlg{})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, cryptoutilOpenapiModel.ECP256, *result)
}

// TestSignBytes_IsSignJWKError tests SignBytes when IsSignJWK returns an error (nil key).
func TestSignBytes_IsSignJWKError(t *testing.T) {
	t.Parallel()

	_, _, err := SignBytes([]joseJwk.Key{nil}, []byte("data"))
	require.Error(t, err)
}

// TestVerifyBytes_IsVerifyJWKError tests VerifyBytes when IsVerifyJWK returns an error (nil key).
func TestVerifyBytes_IsVerifyJWKError(t *testing.T) {
	t.Parallel()

	_, err := VerifyBytes([]joseJwk.Key{nil}, []byte("data"))
	require.Error(t, err)
}

// TestEncryptBytesWithContext_IsEncryptJWKError tests error when IsEncryptJWK returns an error.
func TestEncryptBytesWithContext_IsEncryptJWKError(t *testing.T) {
	t.Parallel()

	_, _, err := EncryptBytesWithContext([]joseJwk.Key{nil}, []byte("data"), nil)
	require.Error(t, err)
}

// TestDecryptBytesWithContext_IsDecryptJWKError tests error when IsDecryptJWK returns an error.
func TestDecryptBytesWithContext_IsDecryptJWKError(t *testing.T) {
	t.Parallel()

	_, err := DecryptBytesWithContext([]joseJwk.Key{nil}, []byte("data"), nil)
	require.Error(t, err)
}
