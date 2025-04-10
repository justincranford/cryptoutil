package jose

import (
	"crypto/rand"
	"fmt"

	"encoding/json"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

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
