// Copyright (c) 2025 Justin Cranford
//
//

package example

import (
	"crypto/ecdh"
	crand "crypto/rand"
	rsa "crypto/rsa"
	json "encoding/json"
	"testing"
	"time"

	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

type testCaseJWE struct {
	raw any // Non-public raw key => Secret Key (AES) or Private Key (ECDH, RSA)
	enc joseJwa.ContentEncryptionAlgorithm
	alg joseJwa.KeyEncryptionAlgorithm
}

func Test_Import_Encrypt_Decrypt(t *testing.T) {
	testCasesJWE := []testCaseJWE{
		generateJWETestCaseECDH(t, ecdh.P256(), joseJwa.A256GCM(), joseJwa.ECDH_ES_A256KW()),
		generateJWETestCaseRSA(t, 2048, joseJwa.A256GCM(), joseJwa.RSA_OAEP_256()),
		generateJWETestCaseAES(t, 256, joseJwa.A256GCM(), joseJwa.A256KW()),
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

func generateJWETestCaseECDH(t *testing.T, ecdhCurve ecdh.Curve, enc joseJwa.ContentEncryptionAlgorithm, alg joseJwa.KeyEncryptionAlgorithm) testCaseJWE {
	t.Helper()

	ecdhPrivateKey, err := ecdhCurve.GenerateKey(crand.Reader)
	require.NoError(t, err, "failed to generate raw ECDH private key for JWE test case")

	return testCaseJWE{raw: ecdhPrivateKey, enc: enc, alg: alg}
}

func generateJWETestCaseRSA(t *testing.T, keyLengthBits int, enc joseJwa.ContentEncryptionAlgorithm, alg joseJwa.KeyEncryptionAlgorithm) testCaseJWE {
	t.Helper()

	rsaPrivateKey, err := rsa.GenerateKey(crand.Reader, keyLengthBits)
	require.NoError(t, err, "failed to generate raw RSA private key for JWE test case")

	return testCaseJWE{raw: rsaPrivateKey, enc: enc, alg: alg}
}

func generateJWETestCaseAES(t *testing.T, keyLengthBits int, enc joseJwa.ContentEncryptionAlgorithm, alg joseJwa.KeyEncryptionAlgorithm) testCaseJWE {
	t.Helper()

	aesSecretKey := make([]byte, keyLengthBits/8)
	_, err := crand.Read(aesSecretKey)
	require.NoError(t, err, "failed to generate raw AES secret key for JWE test case")

	return testCaseJWE{raw: aesSecretKey, enc: enc, alg: alg}
}

func Import(t *testing.T, raw any, enc joseJwa.ContentEncryptionAlgorithm, alg joseJwa.KeyEncryptionAlgorithm) joseJwk.Key {
	t.Helper()

	nonPublicJWK, err := joseJwk.Import(raw)
	require.NoError(t, err, "failed to import raw key into JWK")

	kid, err := googleUuid.NewV7()
	require.NoError(t, err, "failed to generate UUIDv7 for recipient JWK 'kid'")

	err = nonPublicJWK.Set(joseJwk.KeyIDKey, kid.String())
	require.NoError(t, err, "failed to set 'kid' in recipient JWK")
	err = nonPublicJWK.Set(joseJwk.AlgorithmKey, alg)
	require.NoError(t, err, "failed to set 'alg' in recipient JWK")
	err = nonPublicJWK.Set("enc", enc)
	require.NoError(t, err, "failed to set 'enc' in recipient JWK")
	err = nonPublicJWK.Set("iat", time.Now().UTC().Unix())
	require.NoError(t, err, "failed to set 'iat' in recipient JWK")
	err = nonPublicJWK.Set("exp", time.Now().UTC().Unix()+(365*24*60*60)) // 365 days expiration (in seconds)
	require.NoError(t, err, "failed to set 'exp' in recipient JWK")
	err = nonPublicJWK.Set(joseJwk.KeyUsageKey, joseJwk.ForEncryption.String())
	require.NoError(t, err, "failed to set 'use' in recipient JWK")
	err = nonPublicJWK.Set(joseJwk.KeyOpsKey, joseJwk.KeyOperationList{joseJwk.KeyOpEncrypt, joseJwk.KeyOpDecrypt})
	require.NoError(t, err, "failed to set 'key_ops' in recipient JWK")

	nonPublicJWKBytes, err := json.Marshal(nonPublicJWK)
	require.NoError(t, err, "failed to marshal recipient JWK")
	t.Logf("JWE JWK:\n%s", string(nonPublicJWKBytes))

	return nonPublicJWK
}

func encrypt(t *testing.T, recipientJWK joseJwk.Key, plaintext []byte) *joseJwe.Message {
	t.Helper()
	require.NotEmpty(t, plaintext, "plaintext can't be empty")

	isEncryptJWK, err := cryptoutilSharedCryptoJose.IsEncryptJWK(recipientJWK)
	require.NoError(t, err, "failed to validate recipient JWK")
	require.True(t, isEncryptJWK, "recipient JWK must be an encrypt JWK")

	jweProtectedHeaders := joseJwe.NewHeaders()
	err = jweProtectedHeaders.Set("iat", time.Now().UTC().Unix())
	require.NoError(t, err, "failed to set 'iat' header in JWE protected headers")

	jweEncryptOptions := make([]joseJwe.EncryptOption, 0, 2)
	jweEncryptOptions = append(jweEncryptOptions, joseJwe.WithProtectedHeaders(jweProtectedHeaders))

	kid, enc, alg := getKidEncAlgFromJWK(t, recipientJWK)

	jweProtectedHeaders = joseJwe.NewHeaders()
	if err := jweProtectedHeaders.Set(joseJwk.KeyIDKey, kid); err != nil {
		require.NoError(t, err, "failed to set kid header")
	}

	if err := jweProtectedHeaders.Set("enc", enc); err != nil {
		require.NoError(t, err, "failed to set enc header")
	}

	if err := jweProtectedHeaders.Set(joseJwk.AlgorithmKey, alg); err != nil {
		require.NoError(t, err, "failed to set alg header")
	}

	jweEncryptOptions = append(jweEncryptOptions, joseJwe.WithKey(alg, recipientJWK, joseJwe.WithPerRecipientHeaders(jweProtectedHeaders)))

	jweMessageBytes, err := joseJwe.Encrypt(plaintext, jweEncryptOptions...)
	require.NoError(t, err, "failed to encrypt plaintext")
	t.Logf("JWE Message:\n%s", string(jweMessageBytes))

	jweMessage, err := joseJwe.Parse(jweMessageBytes)
	require.NoError(t, err, "failed to parse JWE message bytes")

	return jweMessage
}

func decrypt(t *testing.T, recipientJWK joseJwk.Key, jweMessage *joseJwe.Message) []byte {
	t.Helper()
	require.NotEmpty(t, jweMessage, "JWE message can't be empty")

	isDecryptJWK, err := cryptoutilSharedCryptoJose.IsDecryptJWK(recipientJWK)
	require.NoError(t, err, "failed to validate recipient JWK")
	require.True(t, isDecryptJWK, "recipient JWK must be a decrypt JWK")

	jweMessageBytes, err := jweMessage.MarshalJSON()
	require.NoError(t, err, "failed to marshal JWE message to JSON")

	_, _, alg := getKidEncAlgFromJWK(t, recipientJWK)
	jweDecryptOptions := []joseJwe.DecryptOption{joseJwe.WithKey(alg, recipientJWK)}

	decryptedBytes, err := joseJwe.Decrypt(jweMessageBytes, jweDecryptOptions...)
	require.NoError(t, err, "failed to decrypt JWE message bytes")

	return decryptedBytes
}

// getKidEncAlgFromJWK extracts 'kid', 'enc', and 'alg' headers from recipient JWK. All 3 are assumed to be present in the JWK.
func getKidEncAlgFromJWK(t *testing.T, recipientJWK joseJwk.Key) (string, joseJwa.ContentEncryptionAlgorithm, joseJwa.KeyEncryptionAlgorithm) {
	t.Helper()

	var kid string

	err := recipientJWK.Get(joseJwk.KeyIDKey, &kid)
	require.NoError(t, err, "failed to get 'kid' from recipient JWK")

	var enc joseJwa.ContentEncryptionAlgorithm

	err = recipientJWK.Get("enc", &enc) // EX: A256GCM, A256CBC-HS512, dir
	if err != nil {
		var encString string // Workaround: get 'enc' as string and convert to ContentEncryptionAlgorithm

		err = recipientJWK.Get("enc", &encString)
		require.NoError(t, err, "failed to get 'enc' from recipient JWK")

		enc = joseJwa.NewContentEncryptionAlgorithm(encString) // Convert string to ContentEncryptionAlgorithm
	}

	var alg joseJwa.KeyEncryptionAlgorithm

	err = recipientJWK.Get(joseJwk.AlgorithmKey, &alg) // EX: A256KW, A256GCMKW, RSA_OAEP_512, RSA1_5, ECDH_ES_A256KW
	require.NoError(t, err, "failed to get 'alg' from recipient JWK")

	return kid, enc, alg
}
