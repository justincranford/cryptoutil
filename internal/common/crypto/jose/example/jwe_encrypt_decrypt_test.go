package example

import (
	"crypto/ecdh"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	cryptoutilJose "cryptoutil/internal/common/crypto/jose"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwe"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

type testCaseJWE struct {
	raw any // Non-public raw key => Secret Key (AES) or Private Key (ECDH, RSA)
	enc jwa.ContentEncryptionAlgorithm
	alg jwa.KeyEncryptionAlgorithm
}

func Test_Import_Encrypt_Decrypt(t *testing.T) {
	testCasesJWE := []testCaseJWE{
		generateJWETestCaseECDH(t, ecdh.P256(), jwa.A256GCM(), jwa.ECDH_ES_A256KW()),
		generateJWETestCaseRSA(t, 2048, jwa.A256GCM(), jwa.RSA_OAEP_256()),
		generateJWETestCaseAES(t, 256, jwa.A256GCM(), jwa.A256KW()),
	}

	plaintext := []byte("Hello, World!")
	for _, testCaseJWE := range testCasesJWE {
		t.Run(testCaseJWE.alg.String(), func(t *testing.T) {
			nonPublicJWK := Import(t, testCaseJWE.raw, testCaseJWE.enc, testCaseJWE.alg)
			publicJWK, err := nonPublicJWK.PublicKey()
			require.NoError(t, err, "failed to get public key from non-public JWK")
			encrypted := encrypt(t, publicJWK, plaintext)
			decrypted := decrypt(t, nonPublicJWK, encrypted)
			require.Equal(t, plaintext, decrypted, "decrypted must match original")
		})
	}
}

func generateJWETestCaseECDH(t *testing.T, ecdhCurve ecdh.Curve, enc jwa.ContentEncryptionAlgorithm, alg jwa.KeyEncryptionAlgorithm) testCaseJWE {
	t.Helper()
	ecdhPrivateKey, err := ecdhCurve.GenerateKey(rand.Reader)
	require.NoError(t, err, "failed to generate raw ECDH private key for JWE test case")
	return testCaseJWE{raw: ecdhPrivateKey, enc: enc, alg: alg}
}

func generateJWETestCaseRSA(t *testing.T, keyLengthBits int, enc jwa.ContentEncryptionAlgorithm, alg jwa.KeyEncryptionAlgorithm) testCaseJWE {
	t.Helper()
	rsaPrivateKey, err := rsa.GenerateKey(rand.Reader, keyLengthBits)
	require.NoError(t, err, "failed to generate raw RSA private key for JWE test case")
	return testCaseJWE{raw: rsaPrivateKey, enc: enc, alg: alg}
}

func generateJWETestCaseAES(t *testing.T, keyLengthBits int, enc jwa.ContentEncryptionAlgorithm, alg jwa.KeyEncryptionAlgorithm) testCaseJWE {
	t.Helper()
	aesSecretKey := make([]byte, keyLengthBits/8)
	_, err := rand.Read(aesSecretKey)
	require.NoError(t, err, "failed to generate raw AES secret key for JWE test case")
	return testCaseJWE{raw: aesSecretKey, enc: enc, alg: alg}
}

func Import(t *testing.T, raw any, enc jwa.ContentEncryptionAlgorithm, alg jwa.KeyEncryptionAlgorithm) jwk.Key {
	t.Helper()
	nonPublicJWK, err := jwk.Import(raw)
	require.NoError(t, err, "failed to import raw key into JWK")

	kid, err := uuid.NewV7()
	require.NoError(t, err, "failed to generate UUIDv7 for recipient JWK 'kid'")

	err = nonPublicJWK.Set(jwk.KeyIDKey, kid.String())
	require.NoError(t, err, "failed to set 'kid' in recipient JWK")
	err = nonPublicJWK.Set(jwk.AlgorithmKey, alg)
	require.NoError(t, err, "failed to set 'alg' in recipient JWK")
	err = nonPublicJWK.Set("enc", enc)
	require.NoError(t, err, "failed to set 'enc' in recipient JWK")
	err = nonPublicJWK.Set("iat", time.Now().UTC().Unix())
	require.NoError(t, err, "failed to set 'iat' in recipient JWK")
	err = nonPublicJWK.Set("exp", time.Now().UTC().Unix()+(365*24*60*60)) // 365 days expiration (in seconds)
	require.NoError(t, err, "failed to set 'exp' in recipient JWK")
	err = nonPublicJWK.Set(jwk.KeyUsageKey, jwk.ForEncryption.String())
	require.NoError(t, err, "failed to set 'use' in recipient JWK")
	err = nonPublicJWK.Set(jwk.KeyOpsKey, jwk.KeyOperationList{jwk.KeyOpEncrypt, jwk.KeyOpDecrypt})
	require.NoError(t, err, "failed to set 'key_ops' in recipient JWK")

	nonPublicJWKBytes, err := json.Marshal(nonPublicJWK)
	require.NoError(t, err, "failed to marshal recipient JWK")
	t.Logf("JWE JWK:\n%s", string(nonPublicJWKBytes))

	return nonPublicJWK
}

func encrypt(t *testing.T, recipientJWK jwk.Key, plaintext []byte) *jwe.Message {
	t.Helper()
	require.NotEmpty(t, plaintext, "plaintext can't be empty")
	isEncryptJWK, err := cryptoutilJose.IsEncryptJWK(recipientJWK)
	require.NoError(t, err, "failed to validate recipient JWK")
	require.True(t, isEncryptJWK, "recipient JWK must be an encrypt JWK")

	jweProtectedHeaders := jwe.NewHeaders()
	err = jweProtectedHeaders.Set("iat", time.Now().UTC().Unix())
	require.NoError(t, err, "failed to set 'iat' header in JWE protected headers")

	jweEncryptOptions := make([]jwe.EncryptOption, 0, 2)
	jweEncryptOptions = append(jweEncryptOptions, jwe.WithProtectedHeaders(jweProtectedHeaders))

	kid, enc, alg := getKidEncAlgFromJWK(t, recipientJWK)

	jweProtectedHeaders = jwe.NewHeaders()
	if err := jweProtectedHeaders.Set(jwk.KeyIDKey, kid); err != nil {
		require.NoError(t, err, "failed to set kid header")
	}
	if err := jweProtectedHeaders.Set("enc", enc); err != nil {
		require.NoError(t, err, "failed to set enc header")
	}
	if err := jweProtectedHeaders.Set(jwk.AlgorithmKey, alg); err != nil {
		require.NoError(t, err, "failed to set alg header")
	}
	jweEncryptOptions = append(jweEncryptOptions, jwe.WithKey(alg, recipientJWK, jwe.WithPerRecipientHeaders(jweProtectedHeaders)))

	jweMessageBytes, err := jwe.Encrypt(plaintext, jweEncryptOptions...)
	require.NoError(t, err, fmt.Errorf("failed to encrypt plaintext: %w", err))
	t.Logf("JWE Message:\n%s", string(jweMessageBytes))

	jweMessage, err := jwe.Parse(jweMessageBytes)
	require.NoError(t, err, fmt.Errorf("failed to parse JWE message bytes: %w", err))

	return jweMessage
}

func decrypt(t *testing.T, recipientJWK jwk.Key, jweMessage *jwe.Message) []byte {
	t.Helper()
	require.NotEmpty(t, jweMessage, "JWE message can't be empty")
	isDecryptJWK, err := cryptoutilJose.IsDecryptJWK(recipientJWK)
	require.NoError(t, err, "failed to validate recipient JWK")
	require.True(t, isDecryptJWK, "recipient JWK must be a decrypt JWK")

	jweMessageBytes, err := jweMessage.MarshalJSON()
	require.NoError(t, err, "failed to marshal JWE message to JSON")

	_, _, alg := getKidEncAlgFromJWK(t, recipientJWK)
	jweDecryptOptions := []jwe.DecryptOption{jwe.WithKey(alg, recipientJWK)}

	decryptedBytes, err := jwe.Decrypt(jweMessageBytes, jweDecryptOptions...)
	require.NoError(t, err, "failed to decrypt JWE message bytes")

	return decryptedBytes
}

// getKidEncAlgFromJWK extracts 'kid', 'enc', and 'alg' headers from recipient JWK. All 3 are assumed to be present in the JWK.
func getKidEncAlgFromJWK(t *testing.T, recipientJWK jwk.Key) (string, jwa.ContentEncryptionAlgorithm, jwa.KeyEncryptionAlgorithm) {
	t.Helper()
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
