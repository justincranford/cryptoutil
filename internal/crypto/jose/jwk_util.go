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

func GenerateAesJWK(alg *joseJwa.KeyEncryptionAlgorithm, enc *joseJwa.ContentEncryptionAlgorithm) (*googleUuid.UUID, joseJwk.Key, []byte, error) {
	kid := googleUuid.Must(googleUuid.NewV7())
	rawKey, err := ValidateJWKHeaders(&kid, alg, enc, nil, true) // true => if successful, it makes a raw key  []byte of the correct length
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid JWK headers: %w", err)
	}
	_, err = rand.Read(rawKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate raw key length %d for alg %s and enc %s: %w", len(rawKey)/8, *alg, *enc, err)
	}
	return CreateAesJWKFromBytes(&kid, alg, enc, rawKey)
}

func GenerateAesJWKFromPool(alg *joseJwa.KeyEncryptionAlgorithm, enc *joseJwa.ContentEncryptionAlgorithm, keyGenPool *cryptoutilKeygen.KeyGenPool) (*googleUuid.UUID, joseJwk.Key, []byte, error) {
	kid := googleUuid.Must(googleUuid.NewV7())
	rawKey, ok := keyGenPool.Get().Private.([]byte)
	if !ok {
		return nil, nil, nil, fmt.Errorf("failed to generate raw AES 256 key from pool")
	}
	return CreateAesJWKFromBytes(&kid, alg, enc, rawKey)
}

func CreateAesJWKFromBytes(kid *googleUuid.UUID, alg *joseJwa.KeyEncryptionAlgorithm, enc *joseJwa.ContentEncryptionAlgorithm, rawKey []byte) (*googleUuid.UUID, joseJwk.Key, []byte, error) {
	_, err := ValidateJWKHeaders(kid, alg, enc, rawKey, false)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid JWK headers: %w", err)
	}
	aesJwk, err := joseJwk.Import(rawKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to import raw AES raw key: %w", err)
	}
	if err = aesJwk.Set(joseJwk.KeyIDKey, kid.String()); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to set `kid` header: %w", err)
	}
	if err = aesJwk.Set(joseJwk.AlgorithmKey, *alg); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to set `alg` header: %w", err)
	}
	if err = aesJwk.Set("enc", *enc); err != nil {
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

	return kid, aesJwk, encodedAesJwk, nil
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

func ValidateJWKHeaders(kid *googleUuid.UUID, alg *joseJwa.KeyEncryptionAlgorithm, enc *joseJwa.ContentEncryptionAlgorithm, rawKey []byte, isNilRawKeyOk bool) ([]byte, error) {
	if err := cryptoutilUtil.ValidateUUID(kid, "invalid kid"); err != nil {
		return nil, fmt.Errorf("kid must be valid: %w", err)
	} else if alg == nil {
		return nil, fmt.Errorf("alg must be non-nil")
	} else if enc == nil {
		return nil, fmt.Errorf("enc must be non-nil")
	} else if !isNilRawKeyOk && rawKey == nil {
		return nil, fmt.Errorf("raw key must be non-nil")
	}
	switch *alg {
	case AlgA256KW, AlgA256GCMKW:
		switch *enc {
		case EncA256GCM, EncA256CBC_HS512:
			if rawKey == nil {
				if !isNilRawKeyOk {
					return nil, fmt.Errorf("valid alg %s and enc %s, but invalid nil key; use 32-bytes", *alg, *enc)
				}
				rawKey = make([]byte, 32)
			} else if len(rawKey) != 32 {
				return nil, fmt.Errorf("valid alg %s and enc %s, but invalid key length %d; use 32-bytes", *alg, *enc, len(rawKey))
			}
		default:
			return nil, fmt.Errorf("valid alg %s, but invalid enc %s; use %s or %s", *alg, *enc, EncA256GCM, EncA256CBC_HS512)
		}
	case AlgA192KW, AlgA192GCMKW:
		switch *enc {
		case EncA192GCM, EncA192CBC_HS384:
			if rawKey == nil {
				if !isNilRawKeyOk {
					return nil, fmt.Errorf("valid alg %s and enc %s, but invalid nil key; use 24-bytes", *alg, *enc)
				}
				rawKey = make([]byte, 24)
			} else if len(rawKey) != 24 {
				return nil, fmt.Errorf("valid alg %s and enc %s, but invalid key length %d; use 24-bytes", *alg, *enc, len(rawKey))
			}
		default:
			return nil, fmt.Errorf("valid alg %s, but invalid enc %s; use %s or %s", *alg, *enc, EncA192GCM, EncA192CBC_HS384)
		}
	case AlgA128KW, AlgA128GCMKW:
		switch *enc {
		case EncA128GCM, EncA128CBC_HS256:
			if rawKey == nil {
				if !isNilRawKeyOk {
					return nil, fmt.Errorf("valid alg %s and enc %s, but invalid nil key; use 16-bytes", *alg, *enc)
				}
				rawKey = make([]byte, 16)
			} else if len(rawKey) != 16 {
				return nil, fmt.Errorf("valid alg %s and enc %s, but invalid key length %d; use 16-bytes", *alg, *enc, len(rawKey))
			}
		default:
			return nil, fmt.Errorf("valid alg %s, but invalid enc %s; use %s or %s", *alg, *enc, EncA128GCM, EncA128CBC_HS256)
		}
	case AlgDIRECT:
		switch *enc {
		case EncA256GCM:
			if rawKey == nil {
				if !isNilRawKeyOk {
					return nil, fmt.Errorf("valid alg %s and enc %s, but invalid nil key; use 32-bytes", *alg, *enc)
				}
				rawKey = make([]byte, 32)
			} else if len(rawKey) != 32 {
				return nil, fmt.Errorf("valid alg %s and enc %s, but invalid key length %d; use 32-bytes", *alg, *enc, len(rawKey))
			}
		case EncA192GCM:
			if rawKey == nil {
				if !isNilRawKeyOk {
					return nil, fmt.Errorf("valid alg %s and enc %s, but invalid nil key; use 24-bytes", *alg, *enc)
				}
				rawKey = make([]byte, 24)
			} else if len(rawKey) != 24 {
				return nil, fmt.Errorf("valid alg %s and enc %s, but invalid key length %d; use 24-bytes", *alg, *enc, len(rawKey))
			}
		case EncA128GCM:
			if rawKey == nil {
				if !isNilRawKeyOk {
					return nil, fmt.Errorf("valid alg %s and enc %s, but invalid nil key; use 16-bytes", *alg, *enc)
				}
				rawKey = make([]byte, 16)
			} else if len(rawKey) != 16 {
				return nil, fmt.Errorf("valid alg %s and enc %s, but invalid key length %d; use 16-bytes", *alg, *enc, len(rawKey))
			}
		case EncA256CBC_HS512:
			if rawKey == nil {
				if !isNilRawKeyOk {
					return nil, fmt.Errorf("valid alg %s and enc %s, but invalid nil key; use 64-bytes", *alg, *enc)
				}
				rawKey = make([]byte, 64)
			} else if len(rawKey) != 64 {
				return nil, fmt.Errorf("valid alg %s and enc %s, but invalid key length %d; use 64-bytes", *alg, *enc, len(rawKey))
			}
		case EncA192CBC_HS384:
			if rawKey == nil {
				if !isNilRawKeyOk {
					return nil, fmt.Errorf("valid alg %s and enc %s, but invalid nil key; use 48-bytes", *alg, *enc)
				}
				rawKey = make([]byte, 48)
			} else if len(rawKey) != 48 {
				return nil, fmt.Errorf("valid alg %s and enc %s, but invalid key length %d; use 48-bytes", *alg, *enc, len(rawKey))
			}
		case EncA128CBC_HS256:
			if rawKey == nil {
				if !isNilRawKeyOk {
					return nil, fmt.Errorf("valid alg %s and enc %s, but invalid nil key; use 32-bytes", *alg, *enc)
				}
				rawKey = make([]byte, 32)
			} else if len(rawKey) != 32 {
				return nil, fmt.Errorf("valid alg %s and enc %s, but invalid key length %d; use 32-bytes", *alg, *enc, len(rawKey))
			}
		default:
			return nil, fmt.Errorf("valid alg %s, but unsupported enc %s; use %s, %s, %s, %s, %s, or %s", *alg, *enc, EncA256GCM, EncA192GCM, EncA128GCM, EncA256CBC_HS512, EncA192CBC_HS384, EncA128CBC_HS256)
		}
	default:
		return nil, fmt.Errorf("unsupported alg %s; use %s, %s, %s, %s, %s, %s, or %s", *alg, AlgA256KW, AlgA192KW, AlgA128GCMKW, AlgA256KW, AlgA192GCMKW, AlgA128GCMKW, AlgDIRECT)
	}
	return rawKey, nil
}
