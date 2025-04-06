package jose

import (
	"crypto/rand"
	"fmt"

	"encoding/json"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwe"
	"github.com/lestrrat-go/jwx/v3/jwk"
)

var (
	KtyOct       = jwa.OctetSeq()                             // KeyType
	AlgDIRECT    = jwa.DIRECT()                               // KeyEncryptionAlgorithm
	AlgA256GCMKW = jwa.A256GCMKW()                            // KeyEncryptionAlgorithm
	AlgA256GCM   = jwa.A256GCM()                              // ContentEncryptionAlgorithm
	OpsEncDec    = jwk.KeyOperationList{"encrypt", "decrypt"} // []KeyOperation
)

func GenerateAesJWK(alg jwa.KeyEncryptionAlgorithm) (jwk.Key, []byte, error) {
	switch alg {
	case AlgDIRECT, AlgA256GCMKW:
	default:
		return nil, nil, fmt.Errorf("unsupported algorithm; only use %s or %s", AlgDIRECT, AlgA256GCMKW)
	}

	rawkey := make([]byte, 32)
	_, err := rand.Read(rawkey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate raw AES 256 key: %w", err)
	}

	aesJwk, err := jwk.Import(rawkey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to import raw AES 256 key: %w", err)
	}
	if err = aesJwk.Set(jwk.KeyIDKey, uuid.Must(uuid.NewV7()).String()); err != nil {
		return nil, nil, fmt.Errorf("failed to set `kid` header: %w", err)
	}
	if err = aesJwk.Set(jwk.AlgorithmKey, alg); err != nil {
		return nil, nil, fmt.Errorf("failed to set `alg` header: %w", err)
	}
	if err = aesJwk.Set(jwk.KeyUsageKey, "enc"); err != nil {
		return nil, nil, fmt.Errorf("failed to set `enc` header: %w", err)
	}
	if err = aesJwk.Set(jwk.KeyOpsKey, OpsEncDec); err != nil {
		return nil, nil, fmt.Errorf("failed to set `ops` header: %w", err)
	}
	if err = aesJwk.Set(jwk.KeyTypeKey, KtyOct); err != nil {
		return nil, nil, fmt.Errorf("failed to set 'kty': %w", err)
	}

	encodedAesJwk, err := json.Marshal(aesJwk)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to serialize key AES 256 JWK: %w", err)
	}

	return aesJwk, encodedAesJwk, nil
}

func EncryptBytes(encryptionKey jwk.Key, clearBytes []byte) (*jwe.Message, []byte, error) {
	if encryptionKey == nil {
		return nil, nil, fmt.Errorf("nil JWK key provided")
	}

	var alg jwa.KeyAlgorithm
	err := encryptionKey.Get(jwk.AlgorithmKey, &alg)
	if err != nil {
		return nil, nil, fmt.Errorf("algorithm not found in JWK: %w", err)
	}

	encodedJweMessage, err := jwe.Encrypt(clearBytes, jwe.WithKey(alg, encryptionKey))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt clearBytes: %w", err)
	}

	jweMessage, err := jwe.Parse(encodedJweMessage)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse encrypted object: %w", err)
	}

	return jweMessage, encodedJweMessage, nil
}

func DecryptBytes(decryptionKey jwk.Key, encryptedBytes []byte) ([]byte, error) {
	if decryptionKey == nil {
		return nil, fmt.Errorf("nil JWK key provided")
	}

	var alg jwa.KeyAlgorithm
	err := decryptionKey.Get(jwk.AlgorithmKey, &alg)
	if err != nil {
		return nil, fmt.Errorf("algorithm not found in JWK: %w", err)
	}

	var msg jwe.Message
	decryptedBytes, err := jwe.Decrypt(encryptedBytes, jwe.WithKey(alg, decryptionKey), jwe.WithMessage(&msg))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt JWE: %w", err)
	}

	return decryptedBytes, nil
}

func EncryptKey(kek jwk.Key, key jwk.Key) (*jwe.Message, []byte, error) {
	encodedKey, err := json.Marshal(key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode JWK: %w", err)
	}
	return EncryptBytes(kek, []byte(encodedKey))
}

func DecryptKey(kek jwk.Key, encryptedBytes []byte) (jwk.Key, error) {
	decryptedBytes, err := DecryptBytes(kek, encryptedBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt JWK: %w", err)
	}
	parsedKey, err := jwk.ParseKey(decryptedBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWK: %w", err)
	}
	return parsedKey, nil
}

func JSONHeadersString(jweMessage *jwe.Message) (string, error) {
	jweHeadersString, err := json.Marshal(jweMessage.ProtectedHeaders())
	if err != nil {
		return "", fmt.Errorf("failed to marshall JWE headers: %w", err)
	}
	return string(jweHeadersString), err
}
