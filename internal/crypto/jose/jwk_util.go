package jose

import (
	"crypto/rand"
	cryptoutilAppErr "cryptoutil/internal/apperr"
	cryptoutilKeygen "cryptoutil/internal/crypto/keygen"
	cryptoutilUtil "cryptoutil/internal/util"
	"encoding/json"
	"fmt"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

func GenerateAesJWK(kekAlg *joseJwa.KeyEncryptionAlgorithm) (*googleUuid.UUID, joseJwk.Key, []byte, error) {
	var rawKey []byte
	switch *kekAlg {
	case AlgDIRECT:
		rawKey = make([]byte, 32)
	case AlgA256GCMKW:
		rawKey = make([]byte, 32)
	case AlgA192GCMKW:
		rawKey = make([]byte, 24)
	case AlgA128GCMKW:
		rawKey = make([]byte, 16)
	default:
		return nil, nil, nil, fmt.Errorf("unsupported KEK algorithm; only use %s, %s, %s, or %s", AlgDIRECT, AlgA256GCMKW, AlgA192GCMKW, AlgA128GCMKW)
	}
	_, err := rand.Read(rawKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate raw AES %d key: %w", len(rawKey)/8, err)
	}
	kekKidUuid := googleUuid.Must(googleUuid.NewV7())
	return CreateAesJWKFromBytes(&kekKidUuid, kekAlg, rawKey)
}

func GenerateAesJWKFromPool(kekAlg *joseJwa.KeyEncryptionAlgorithm, aes256KeyGenPool *cryptoutilKeygen.KeyGenPool) (*googleUuid.UUID, joseJwk.Key, []byte, error) {
	rawKey, ok := aes256KeyGenPool.Get().Private.([]byte)
	if !ok {
		return nil, nil, nil, fmt.Errorf("failed to generate raw AES 256 key from pool")
	}
	kekKidUuid := googleUuid.Must(googleUuid.NewV7())
	return CreateAesJWKFromBytes(&kekKidUuid, kekAlg, rawKey)
}

func CreateAesJWKFromBytes(kekKidUuid *googleUuid.UUID, kekAlg *joseJwa.KeyEncryptionAlgorithm, rawkey []byte) (*googleUuid.UUID, joseJwk.Key, []byte, error) {
	if err := cryptoutilUtil.ValidateUUID(kekKidUuid, "invalid kid"); err != nil {
		return nil, nil, nil, fmt.Errorf("kid uuid must be valid")
	}
	switch *kekAlg {
	case AlgDIRECT:
		if rawkey == nil || !(len(rawkey) == 32 || len(rawkey) == 24 || len(rawkey) == 16) {
			return nil, nil, nil, fmt.Errorf("invalid raw key for alg=dir, must be 32-bytes")
		}
	case AlgA256GCMKW:
		if rawkey == nil || len(rawkey) != 32 {
			return nil, nil, nil, fmt.Errorf("invalid raw key for alg=A256GCMKW, must be 32-bytes")
		}
	case AlgA192GCMKW:
		if rawkey == nil || len(rawkey) != 24 {
			return nil, nil, nil, fmt.Errorf("invalid raw key for alg=A192GCMKW, must be 24-bytes")
		}
	case AlgA128GCMKW:
		if rawkey == nil || len(rawkey) != 16 {
			return nil, nil, nil, fmt.Errorf("invalid raw key for alg=A128GCMKW, must be 16-bytes")
		}
	default:
		return nil, nil, nil, fmt.Errorf("unsupported KEK algorithm; only use %s, %s, %s, or %s", AlgDIRECT, AlgA256GCMKW, AlgA192GCMKW, AlgA128GCMKW)
	}

	aesJwk, err := joseJwk.Import(rawkey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to import raw AES raw skey: %w", err)
	}
	if err = aesJwk.Set(joseJwk.KeyIDKey, kekKidUuid.String()); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to set `kid` header: %w", err)
	}
	if err = aesJwk.Set(joseJwk.AlgorithmKey, *kekAlg); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to set `alg` header: %w", err)
	}
	if err = aesJwk.Set(joseJwk.KeyUsageKey, "enc"); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to set `enc` header: %w", err)
	}
	if err = aesJwk.Set(joseJwk.KeyOpsKey, OpsEncDec); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to set `ops` header: %w", err)
	}
	if err = aesJwk.Set(joseJwk.KeyTypeKey, KtyOct); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to set 'kty': %w", err)
	}

	encodedAesJwk, err := json.Marshal(aesJwk)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to serialize key AES 256 JWK: %w", err)
	}

	return kekKidUuid, aesJwk, encodedAesJwk, nil
}

func ExtractKidUuid(jwk joseJwk.Key) (*googleUuid.UUID, error) {
	if jwk == nil {
		return nil, fmt.Errorf("invalid jwk: %w", cryptoutilAppErr.ErrCantBeNil)
	}
	var err error
	var kidString string
	if err = jwk.Get(joseJwk.KeyIDKey, &kidString); err != nil {
		return nil, fmt.Errorf("failed to get kid header: %w", err)
	}
	var kidUuid googleUuid.UUID
	if kidUuid, err = googleUuid.Parse(kidString); err != nil {
		return nil, fmt.Errorf("failed to parse kid as UUID: %w", err)
	}
	if err = cryptoutilUtil.ValidateUUID(&kidUuid, "invalid kid"); err != nil {
		return nil, err
	}
	return &kidUuid, nil
}
