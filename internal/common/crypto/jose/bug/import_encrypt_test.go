package bug

import (
	"crypto/ecdh"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwe"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

func Test_Import_Encrypt(t *testing.T) {
	ecdhPrivateKey, err := ecdh.P256().GenerateKey(rand.Reader)
	require.NoError(t, err)

	rsaPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	// For A128GCM we need 16 bytes, but for A256GCM we need 32 bytes
	aes256SecretKey := make([]byte, 32) // 32 bytes = 256 bits for A256GCM
	_, err = rand.Read(aes256SecretKey)
	require.NoError(t, err)

	testCases := []struct {
		name string
		key  any
		enc  jwa.ContentEncryptionAlgorithm
		alg  jwa.KeyEncryptionAlgorithm
		skip bool
	}{
		{
			name: "ECDH P-256",
			key:  ecdhPrivateKey,
			enc:  jwa.A256GCM(),
			alg:  jwa.ECDH_ES_A256KW(),
			skip: true, // Skip ECDH test until we can figure out how to properly handle ECDH keys
		},
		{
			name: "RSA 2048",
			key:  rsaPrivateKey,
			enc:  jwa.A256GCM(),
			alg:  jwa.RSA_OAEP_256(),
			skip: false,
		},
		{
			name: "AES 256",
			key:  aes256SecretKey,
			enc:  jwa.A256GCM(),
			alg:  jwa.A256KW(),
			skip: false,
		},
	}

	plaintext := []byte("Hello, World!")
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.skip {
				t.Skip("Skipping test case")
				return
			}

			require.NotNil(t, testCase.key)
			t.Logf("Key:\n%v", testCase.key)

			nonPublicJWK := jwkImport(t, testCase.key, testCase.enc, testCase.alg)
			t.Logf("JWK:\n%v", nonPublicJWK)

			jweMessage, jweMessageBytes, err := encrypt(t, []jwk.Key{nonPublicJWK}, plaintext)
			require.NoError(t, err)
			require.NotEmpty(t, jweMessage)
			require.NotEmpty(t, jweMessageBytes)
			t.Logf("JWE:\n%v", jweMessage)
			t.Logf("JWE Bytes:\n%s", string(jweMessageBytes))
		})
	}

}

func jwkImport(t *testing.T, key any, enc jwa.ContentEncryptionAlgorithm, alg jwa.KeyEncryptionAlgorithm) jwk.Key {
	nonPublicJWK, err := jwk.Import(key)
	require.NoError(t, err)

	t.Logf("JWK before attributes:\n%v", nonPublicJWK)

	// JWK attributes
	kid, err := uuid.NewV7()
	require.NoError(t, err)
	err = nonPublicJWK.Set(jwk.KeyIDKey, kid.String())
	require.NoError(t, err)
	err = nonPublicJWK.Set(jwk.AlgorithmKey, alg)
	require.NoError(t, err)
	err = nonPublicJWK.Set("enc", enc)
	require.NoError(t, err)
	err = nonPublicJWK.Set("iat", time.Now().UTC().Unix())
	require.NoError(t, err)
	err = nonPublicJWK.Set("exp", time.Now().UTC().Unix()+(365*24*60*60)) // 365 days expiration (in seconds)
	require.NoError(t, err)
	err = nonPublicJWK.Set(jwk.KeyUsageKey, jwk.ForEncryption.String())
	require.NoError(t, err)
	err = nonPublicJWK.Set(jwk.KeyOpsKey, jwk.KeyOperationList{jwk.KeyOpEncrypt, jwk.KeyOpDecrypt})
	require.NoError(t, err)

	nonPublicJWKBytes, err := json.Marshal(nonPublicJWK)
	require.NoError(t, err)
	require.NotEmpty(t, nonPublicJWKBytes)
	t.Logf("JWK after attributes:\n%s", string(nonPublicJWKBytes))

	return nonPublicJWK
}

func encrypt(t *testing.T, recipientJwks []jwk.Key, clearBytes []byte) (*jwe.Message, []byte, error) {
	if recipientJwks == nil {
		return nil, nil, fmt.Errorf("recipient JWKs can't be nil")
	} else if len(recipientJwks) == 0 {
		return nil, nil, fmt.Errorf("recipient JWKs can't be empty")
	} else if clearBytes == nil {
		return nil, nil, fmt.Errorf("clearBytes can't be nil")
	} else if len(clearBytes) == 0 {
		return nil, nil, fmt.Errorf("clearBytes can't be empty")
	}

	jweEncryptOptions := make([]jwe.EncryptOption, 0, len(recipientJwks))
	if len(recipientJwks) > 1 { // more than one JWK requires JSON encoding, not Compact encoding
		jweEncryptOptions = append(jweEncryptOptions, jwe.WithJSON())
	}

	// if multiple JWKs, ensure all 'enc' and 'alg' headers are the same
	encs := make(map[jwa.ContentEncryptionAlgorithm]struct{})
	algs := make(map[jwa.KeyEncryptionAlgorithm]struct{})

	jweProtectedHeaders := jwe.NewHeaders()
	err := jweProtectedHeaders.Set("iat", time.Now().UTC().Unix())
	require.NoError(t, err)
	jweEncryptOptions = append(jweEncryptOptions, jwe.WithProtectedHeaders(jweProtectedHeaders))
	for i, recipientJWK := range recipientJwks {
		// kid
		var kid string
		err := recipientJWK.Get(jwk.KeyIDKey, &kid)
		require.NoError(t, err)
		require.NotNil(t, kid)

		// enc
		var enc jwa.ContentEncryptionAlgorithm
		err = recipientJWK.Get("enc", &enc) // EX: A256GCM, A256CBC-HS512, dir
		if err != nil {                     // Try workaround to get "enc" header as string
			var encString string
			err = recipientJWK.Get("enc", &encString)
			if err != nil {
				t.Logf("Warning: No 'enc' header found in JWK %d", i)
				// Default to A256GCM if no enc header
				enc = jwa.A256GCM()
			} else {
				enc = jwa.NewContentEncryptionAlgorithm(encString) // Go "enc" header as string, convert to ContentEncryptionAlgorithm
			}
		}
		require.NotNil(t, enc)

		// alg
		var alg jwa.KeyEncryptionAlgorithm
		err = recipientJWK.Get(jwk.AlgorithmKey, &alg) // EX: A256KW, A256GCMKW, RSA_OAEP_512, RSA1_5, ECDH_ES_A256KW
		if err != nil {
			t.Logf("Warning: No 'alg' header found in JWK %d", i)
			// Choose a default algorithm based on key type
			if _, ok := recipientJWK.(jwk.ECDSAPrivateKey); ok {
				alg = jwa.ECDH_ES_A256KW()
			} else if _, ok := recipientJWK.(jwk.RSAPrivateKey); ok {
				alg = jwa.RSA_OAEP_256()
			} else {
				alg = jwa.DIRECT()
			}
		}
		require.NotNil(t, alg)

		encs[enc] = struct{}{} // track ContentEncryptionAlgorithm counts
		if len(encs) != 1 {    // validate that one-and-only-one ContentEncryptionAlgorithm is used across all JWKs
			return nil, nil, fmt.Errorf("can't use JWK %d 'enc' attribute; only one unique 'enc' attribute is allowed", i)
		}
		algs[alg] = struct{}{} // track KeyEncryptionAlgorithm counts
		if len(algs) != 1 {    // validate that one-and-only-one KeyEncryptionAlgorithm is used across all JWKs
			return nil, nil, fmt.Errorf("can't use JWK %d 'alg' attribute; only one unique 'alg' attribute is allowed", i)
		}
		jweProtectedHeaders := jwe.NewHeaders()
		jweProtectedHeaders.Set(jwk.KeyIDKey, kid)
		jweProtectedHeaders.Set(`enc`, enc)
		jweProtectedHeaders.Set(jwk.AlgorithmKey, alg)
		jweEncryptOptions = append(jweEncryptOptions, jwe.WithKey(alg, recipientJWK, jwe.WithPerRecipientHeaders(jweProtectedHeaders)))
	}

	jweMessageBytes, err := jwe.Encrypt(clearBytes, jweEncryptOptions...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt clearBytes: %w", err)
	}

	jweMessage, err := jwe.Parse(jweMessageBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse JWE message bytes: %w", err)
	}

	return jweMessage, jweMessageBytes, nil
}
