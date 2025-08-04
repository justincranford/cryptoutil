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

type tc struct {
	key any
	enc jwa.ContentEncryptionAlgorithm
	alg jwa.KeyEncryptionAlgorithm
}

func Test_Import_Encrypt(t *testing.T) {
	aesTC := aesTestCase(t, 256, jwa.A256GCM(), jwa.A256KW())
	rsaTC := rsaTestCase(t, 2048, jwa.A256GCM(), jwa.RSA_OAEP_256())
	ecdhTC := ecdhTestCase(t, ecdh.P256(), jwa.A256GCM(), jwa.ECDH_ES_A256KW())

	// Skipping ECDH test due to issues with the library

	plaintext := []byte("Hello, World!")
	for _, tc := range []tc{ecdhTC, rsaTC, aesTC} {
		t.Run(tc.alg.String(), func(t *testing.T) {
			jweJWK := importJWK(t, tc.key, tc.enc, tc.alg)

			jweMessage, jweMessageBytes := encrypt(t, jweJWK, plaintext)
			t.Logf("JWE Message:\n%s", string(jweMessageBytes))

			decrypted := decrypt(t, jweJWK, jweMessage)

			require.Equal(t, plaintext, decrypted, "decrypted text should match original plaintext")
			t.Logf("Successfully encrypted and decrypted with %s", tc.alg.String())
		})
	}
}

func aesTestCase(t *testing.T, keyLengthBits int, enc jwa.ContentEncryptionAlgorithm, alg jwa.KeyEncryptionAlgorithm) tc {
	aesSecretKey := make([]byte, keyLengthBits/8)
	_, err := rand.Read(aesSecretKey)
	require.NoError(t, err, "failed to generate AES secret key")
	return tc{key: aesSecretKey, enc: enc, alg: alg}
}

func rsaTestCase(t *testing.T, keyLengthBits int, enc jwa.ContentEncryptionAlgorithm, alg jwa.KeyEncryptionAlgorithm) tc {
	rsaPrivateKey, err := rsa.GenerateKey(rand.Reader, keyLengthBits)
	require.NoError(t, err, "failed to generate RSA private key")
	return tc{key: rsaPrivateKey, enc: enc, alg: alg}
}

func ecdhTestCase(t *testing.T, ecdhCurve ecdh.Curve, enc jwa.ContentEncryptionAlgorithm, alg jwa.KeyEncryptionAlgorithm) tc {
	ecdhPrivateKey, err := ecdhCurve.GenerateKey(rand.Reader)
	require.NoError(t, err, "failed to generate ECDH private key")
	return tc{key: ecdhPrivateKey, enc: enc, alg: alg}
}

func importJWK(t *testing.T, key any, enc jwa.ContentEncryptionAlgorithm, alg jwa.KeyEncryptionAlgorithm) jwk.Key {
	recipientJWK, err := jwk.Import(key)
	require.NoError(t, err, "failed to import key into JWK")

	kid, err := uuid.NewV7()
	require.NoError(t, err, "failed to generate UUID for JWK 'kid'")

	err = recipientJWK.Set(jwk.KeyIDKey, kid.String())
	require.NoError(t, err, "failed to set 'kid' in recipient JWK")
	err = recipientJWK.Set(jwk.AlgorithmKey, alg)
	require.NoError(t, err, "failed to set 'alg' in recipient JWK")
	err = recipientJWK.Set("enc", enc)
	require.NoError(t, err, "failed to set 'enc' in recipient JWK")
	err = recipientJWK.Set("iat", time.Now().UTC().Unix())
	require.NoError(t, err, "failed to set 'iat' in recipient JWK")
	err = recipientJWK.Set("exp", time.Now().UTC().Unix()+(365*24*60*60)) // 365 days expiration (in seconds)
	require.NoError(t, err, "failed to set 'exp' in recipient JWK")
	err = recipientJWK.Set(jwk.KeyUsageKey, jwk.ForEncryption.String())
	require.NoError(t, err, "failed to set 'use' in recipient JWK")
	err = recipientJWK.Set(jwk.KeyOpsKey, jwk.KeyOperationList{jwk.KeyOpEncrypt, jwk.KeyOpDecrypt})
	require.NoError(t, err, "failed to set 'key_ops' in recipient JWK")

	recipientJWKBytes, err := json.Marshal(recipientJWK)
	require.NoError(t, err, "failed to marshal recipient JWK")
	t.Logf("JWE JWK:\n%s", string(recipientJWKBytes))

	return recipientJWK
}

func encrypt(t *testing.T, recipientJWK jwk.Key, clearBytes []byte) (*jwe.Message, []byte) {
	require.NotEmpty(t, clearBytes, "clearBytes can't be empty")

	jweProtectedHeaders := jwe.NewHeaders()
	err := jweProtectedHeaders.Set("iat", time.Now().UTC().Unix())
	require.NoError(t, err, "failed to set 'iat' header in JWE protected headers")

	jweEncryptOptions := make([]jwe.EncryptOption, 0, 2)
	jweEncryptOptions = append(jweEncryptOptions, jwe.WithProtectedHeaders(jweProtectedHeaders))

	kid, enc, alg := extractKidEncAlg(t, recipientJWK)

	jweProtectedHeaders = jwe.NewHeaders()
	jweProtectedHeaders.Set(jwk.KeyIDKey, kid)
	jweProtectedHeaders.Set("enc", enc)
	jweProtectedHeaders.Set(jwk.AlgorithmKey, alg)
	jweEncryptOptions = append(jweEncryptOptions, jwe.WithKey(alg, recipientJWK, jwe.WithPerRecipientHeaders(jweProtectedHeaders)))

	jweMessageBytes, err := jwe.Encrypt(clearBytes, jweEncryptOptions...)
	require.NoError(t, err, fmt.Errorf("failed to encrypt clearBytes: %w", err))

	jweMessage, err := jwe.Parse(jweMessageBytes)
	require.NoError(t, err, fmt.Errorf("failed to parse JWE message bytes: %w", err))

	return jweMessage, jweMessageBytes
}

func decrypt(t *testing.T, recipientJWK jwk.Key, jweMessage *jwe.Message) []byte {
	require.NotEmpty(t, jweMessage, "JWE message can't be empty")

	var alg jwa.KeyEncryptionAlgorithm
	err := jweMessage.ProtectedHeaders().Get(jwk.AlgorithmKey, &alg)
	require.NoError(t, err, "failed to get algorithm from key")

	jweMessageBytes, err := jweMessage.MarshalJSON()
	require.NoError(t, err, "failed to marshal JWE message to JSON")

	jweDecryptOptions := []jwe.DecryptOption{jwe.WithKey(alg, recipientJWK)}

	decryptedBytes, err := jwe.Decrypt(jweMessageBytes, jweDecryptOptions...)
	require.NoError(t, err, "failed to decrypt JWE message bytes")

	return decryptedBytes
}

// extractKidEncAlg extracts 'kid', 'enc', and 'alg' headers from recipient JWK. All 3 are assumed to be present in the JWK.
func extractKidEncAlg(t *testing.T, recipientJWK jwk.Key) (string, jwa.ContentEncryptionAlgorithm, jwa.KeyEncryptionAlgorithm) {
	var kid string
	err := recipientJWK.Get(jwk.KeyIDKey, &kid)
	require.NoError(t, err, "failed to get 'kid' from recipient JWK")

	var enc jwa.ContentEncryptionAlgorithm
	err = recipientJWK.Get("enc", &enc) // EX: A256GCM, A256CBC-HS512, dir
	if err != nil {
		var encString string // Workaround: get 'enc' as string and convert to ContentEncryptionAlgorithm
		err = recipientJWK.Get("enc", &encString)
		require.NoError(t, err, "failed to get 'enc' from recipient JWK")
		enc = jwa.NewContentEncryptionAlgorithm(encString) // Convert string to ContentEncryptionAlgorithm
	}

	var alg jwa.KeyEncryptionAlgorithm
	err = recipientJWK.Get(jwk.AlgorithmKey, &alg) // EX: A256KW, A256GCMKW, RSA_OAEP_512, RSA1_5, ECDH_ES_A256KW
	require.NoError(t, err, "failed to get 'alg' from recipient JWK")
	return kid, enc, alg
}
