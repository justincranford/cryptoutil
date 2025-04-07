package jose

import (
	"crypto/rand"
	"fmt"

	"encoding/json"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

var (
	KtyOct              = joseJwa.OctetSeq()                             // KeyType
	AlgDIRECT           = joseJwa.DIRECT()                               // KeyEncryptionAlgorithm
	AlgA256GCMKW        = joseJwa.A256GCMKW()                            // KeyEncryptionAlgorithm
	AlgA256GCM          = joseJwa.A256GCM()                              // ContentEncryptionAlgorithm
	OpsEncDec           = joseJwk.KeyOperationList{"encrypt", "decrypt"} // []KeyOperation
	ErrCantBeNil        = fmt.Errorf("jwk can't be nil")
	ErrKidCantBeNilUuid = fmt.Errorf("jwk kid can't be nil uuid")
	ErrKidCantBeMaxUuid = fmt.Errorf("jwk kid can't be max uuid")
)

func ExtractKidUuid(jwk joseJwk.Key) (googleUuid.UUID, error) {
	if jwk == nil {
		return googleUuid.Nil, ErrCantBeNil
	}
	var err error
	var kidString string
	if err = jwk.Get(joseJwk.KeyIDKey, &kidString); err != nil {
		return googleUuid.Nil, fmt.Errorf("failed to get `kid` header: %w", err)
	}
	var kidUuid googleUuid.UUID
	if kidUuid, err = googleUuid.Parse(kidString); err != nil {
		return googleUuid.Nil, fmt.Errorf("failed to parse `kid` as UUID: %w", err)
	}
	if err = ValidateKid(kidUuid); err != nil {
		return googleUuid.Nil, fmt.Errorf("invalid `kid`: %w", err)
	}
	return kidUuid, nil
}

func ValidateKid(kidUuid googleUuid.UUID) error {
	switch kidUuid {
	case googleUuid.Nil:
		return ErrKidCantBeNilUuid
	case googleUuid.Max:
		return ErrKidCantBeMaxUuid
	default:
		return nil
	}
}

func GenerateAesJWK(kekAlg joseJwa.KeyEncryptionAlgorithm) (joseJwk.Key, []byte, error) {
	rawKey := make([]byte, 32)
	_, err := rand.Read(rawKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate raw AES 256 key: %w", err)
	}
	return CreateAesJWK(kekAlg, rawKey)
}

func CreateAesJWK(kekAlg joseJwa.KeyEncryptionAlgorithm, rawkey []byte) (joseJwk.Key, []byte, error) {
	switch kekAlg {
	case AlgDIRECT, AlgA256GCMKW:
	default:
		return nil, nil, fmt.Errorf("unsupported algorithm; only use %s or %s", AlgDIRECT, AlgA256GCMKW)
	}

	aesJwk, err := joseJwk.Import(rawkey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to import raw AES 256 key: %w", err)
	}
	if err = aesJwk.Set(joseJwk.KeyIDKey, googleUuid.Must(googleUuid.NewV7()).String()); err != nil {
		return nil, nil, fmt.Errorf("failed to set `kid` header: %w", err)
	}
	if err = aesJwk.Set(joseJwk.AlgorithmKey, kekAlg); err != nil {
		return nil, nil, fmt.Errorf("failed to set `alg` header: %w", err)
	}
	if err = aesJwk.Set(joseJwk.KeyUsageKey, "enc"); err != nil {
		return nil, nil, fmt.Errorf("failed to set `enc` header: %w", err)
	}
	if err = aesJwk.Set(joseJwk.KeyOpsKey, OpsEncDec); err != nil {
		return nil, nil, fmt.Errorf("failed to set `ops` header: %w", err)
	}
	if err = aesJwk.Set(joseJwk.KeyTypeKey, KtyOct); err != nil {
		return nil, nil, fmt.Errorf("failed to set 'kty': %w", err)
	}

	encodedAesJwk, err := json.Marshal(aesJwk)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to serialize key AES 256 JWK: %w", err)
	}

	return aesJwk, encodedAesJwk, nil
}

func EncryptBytes(encryptionKey joseJwk.Key, clearBytes []byte) (*joseJwe.Message, []byte, error) {
	if encryptionKey == nil {
		return nil, nil, ErrCantBeNil
	}

	var alg joseJwa.KeyAlgorithm
	err := encryptionKey.Get(joseJwk.AlgorithmKey, &alg)
	if err != nil {
		return nil, nil, fmt.Errorf("algorithm not found in JWK: %w", err)
	}

	encodedJweMessage, err := joseJwe.Encrypt(clearBytes, joseJwe.WithKey(alg, encryptionKey))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt clearBytes: %w", err)
	}

	jweMessage, err := joseJwe.Parse(encodedJweMessage)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse encrypted object: %w", err)
	}

	return jweMessage, encodedJweMessage, nil
}

func DecryptBytes(decryptionKey joseJwk.Key, encryptedBytes []byte) ([]byte, error) {
	if decryptionKey == nil {
		return nil, ErrCantBeNil
	}

	var alg joseJwa.KeyAlgorithm
	err := decryptionKey.Get(joseJwk.AlgorithmKey, &alg)
	if err != nil {
		return nil, fmt.Errorf("algorithm not found in JWK: %w", err)
	}

	var msg joseJwe.Message
	decryptedBytes, err := joseJwe.Decrypt(encryptedBytes, joseJwe.WithKey(alg, decryptionKey), joseJwe.WithMessage(&msg))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt JWE: %w", err)
	}

	return decryptedBytes, nil
}

func EncryptKey(kek joseJwk.Key, key joseJwk.Key) (*joseJwe.Message, []byte, error) {
	encodedKey, err := json.Marshal(key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode JWK: %w", err)
	}
	return EncryptBytes(kek, []byte(encodedKey))
}

func DecryptKey(kek joseJwk.Key, encryptedBytes []byte) (joseJwk.Key, error) {
	decryptedBytes, err := DecryptBytes(kek, encryptedBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt JWK: %w", err)
	}
	parsedKey, err := joseJwk.ParseKey(decryptedBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWK: %w", err)
	}
	return parsedKey, nil
}

func JSONHeadersString(jweMessage *joseJwe.Message) (string, error) {
	jweHeadersString, err := json.Marshal(jweMessage.ProtectedHeaders())
	if err != nil {
		return "", fmt.Errorf("failed to marshall JWE headers: %w", err)
	}
	return string(jweHeadersString), err
}
