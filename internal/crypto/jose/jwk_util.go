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
	var cekAlg *joseJwa.ContentEncryptionAlgorithm
	switch *kekAlg {
	case AlgKekDIRECT:
		rawKey = make([]byte, 32)
		cekAlg = &AlgCekA256GCM
	case AlgKekA256GCMKW:
		rawKey = make([]byte, 32)
		cekAlg = &AlgCekA256GCM
	case AlgKekA192GCMKW:
		rawKey = make([]byte, 24)
		cekAlg = &AlgCekA192GCM
	case AlgKekA128GCMKW:
		rawKey = make([]byte, 16)
		cekAlg = &AlgCekA128GCM
	default:
		return nil, nil, nil, fmt.Errorf("unsupported KEK algorithm; only use %s, %s, %s, or %s", AlgKekDIRECT, AlgKekA256GCMKW, AlgKekA192GCMKW, AlgKekA128GCMKW)
	}
	_, err := rand.Read(rawKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate raw AES %d key: %w", len(rawKey)/8, err)
	}
	kekKidUuid := googleUuid.Must(googleUuid.NewV7())
	return CreateAesJWKFromBytes(&kekKidUuid, kekAlg, cekAlg, rawKey)
}

func GenerateAesJWKFromPool(kekAlg *joseJwa.KeyEncryptionAlgorithm, aesKeyGenPool *cryptoutilKeygen.KeyGenPool) (*googleUuid.UUID, joseJwk.Key, []byte, error) {
	rawKey, ok := aesKeyGenPool.Get().Private.([]byte)
	if !ok {
		return nil, nil, nil, fmt.Errorf("failed to generate raw AES 256 key from pool")
	}
	var cekAlg *joseJwa.ContentEncryptionAlgorithm
	switch *kekAlg {
	case AlgKekDIRECT:
		if len(rawKey) == 32 {
			cekAlg = &AlgCekA256GCM
		} else if len(rawKey) == 24 {
			cekAlg = &AlgCekA192GCM
		} else if len(rawKey) == 16 {
			cekAlg = &AlgCekA128GCM
		} else {
			return nil, nil, nil, fmt.Errorf("unsupported key pool %s key length %d for KEK algorithm: %s", aesKeyGenPool.Name(), len(rawKey), AlgKekDIRECT)
		}
	case AlgKekA256GCMKW:
		if len(rawKey) != 32 {
			return nil, nil, nil, fmt.Errorf("unsupported key pool %s key length %d for KEK algorithm: %s", aesKeyGenPool.Name(), len(rawKey), AlgKekA256GCMKW)
		}
		cekAlg = &AlgCekA256GCM
	case AlgKekA192GCMKW:
		if len(rawKey) != 24 {
			return nil, nil, nil, fmt.Errorf("unsupported key pool %s key length %d for KEK algorithm: %s", aesKeyGenPool.Name(), len(rawKey), AlgKekA192GCMKW)
		}
		cekAlg = &AlgCekA192GCM
	case AlgKekA128GCMKW:
		if len(rawKey) != 16 {
			return nil, nil, nil, fmt.Errorf("unsupported key pool %s key length %d for KEK algorithm: %s", aesKeyGenPool.Name(), len(rawKey), AlgKekA128GCMKW)
		}
		cekAlg = &AlgCekA128GCM
	default:
		return nil, nil, nil, fmt.Errorf("unsupported KEK algorithm; only use %s, %s, %s, or %s", AlgKekDIRECT, AlgKekA256GCMKW, AlgKekA192GCMKW, AlgKekA128GCMKW)
	}
	kekKidUuid := googleUuid.Must(googleUuid.NewV7())
	return CreateAesJWKFromBytes(&kekKidUuid, kekAlg, cekAlg, rawKey)
}

func CreateAesJWKFromBytes(kekKidUuid *googleUuid.UUID, kekAlg *joseJwa.KeyEncryptionAlgorithm, cekAlg *joseJwa.ContentEncryptionAlgorithm, rawkey []byte) (*googleUuid.UUID, joseJwk.Key, []byte, error) {
	if err := cryptoutilUtil.ValidateUUID(kekKidUuid, "invalid kid"); err != nil {
		return nil, nil, nil, fmt.Errorf("kid uuid must be valid")
	} else if kekAlg == nil {
		return nil, nil, nil, fmt.Errorf("kek alg must be non-nil")
	} else if cekAlg == nil {
		return nil, nil, nil, fmt.Errorf("cek alg must be non-nil")
	} else if rawkey == nil {
		return nil, nil, nil, fmt.Errorf("raw key must be non-nil")
	}
	switch (*kekAlg).String() {
	case AlgKekDIRECT.String():
		if !(len(rawkey) == 32 || len(rawkey) == 24 || len(rawkey) == 16) {
			return nil, nil, nil, fmt.Errorf("invalid raw key length %d for alg=dir, must be 32-bytes, 24-bytes, or 16-bytes", len(rawkey))
		} else if !((*cekAlg).String() == AlgCekA256GCM.String() || (*cekAlg).String() == AlgCekA192GCM.String() || (*cekAlg).String() == AlgCekA128GCM.String() || (*cekAlg).String() == AlgCekA256CBC_HS512.String() || (*cekAlg).String() == AlgCekA192CBC_HS384.String() || (*cekAlg).String() == AlgCekA128CBC_HS256.String()) {
			return nil, nil, nil, fmt.Errorf("invalid raw key enc=%s for alg=dir, must be A256GCM, A192GCM, A128GCM, A256CBC_HS512, A192CBC_HS384, or A128CBC_HS256", *cekAlg)
		}
	case AlgKekA256GCMKW.String():
		if len(rawkey) != 32 {
			return nil, nil, nil, fmt.Errorf("invalid raw key length %d for alg=A256GCMKW, must be 32-bytes", len(rawkey))
		} else if !((*cekAlg).String() != AlgCekA256GCM.String() || (*cekAlg).String() != AlgCekA256CBC_HS512.String()) {
			return nil, nil, nil, fmt.Errorf("invalid raw key enc=%s for alg=A256GCMKW, must be A256GCM or A256CBC_HS512", *cekAlg)
		}
	case AlgKekA192GCMKW.String():
		if len(rawkey) != 24 {
			return nil, nil, nil, fmt.Errorf("invalid raw key length %d for alg=A192GCMKW, must be 24-bytes", len(rawkey))
		} else if !((*cekAlg).String() != AlgCekA192GCM.String() || (*cekAlg).String() != AlgCekA192CBC_HS384.String()) {
			return nil, nil, nil, fmt.Errorf("invalid raw key enc=%s for alg=A192GCMKW, must be A192GCM or A192CBC_HS384", *cekAlg)
		}
	case AlgKekA128GCMKW.String():
		if len(rawkey) != 16 {
			return nil, nil, nil, fmt.Errorf("invalid raw key length %d for alg=A128GCMKW, must be 16-bytes", len(rawkey))
		} else if !((*cekAlg).String() != AlgCekA128GCM.String() || (*cekAlg).String() != AlgCekA128CBC_HS256.String()) {
			return nil, nil, nil, fmt.Errorf("invalid raw key enc=%s for alg=A128GCMKW, must be A128GCM or A128CBC_HS256", *cekAlg)
		}
	default:
		return nil, nil, nil, fmt.Errorf("unsupported KEK algorithm; only use %s, %s, %s, or %s", AlgKekDIRECT, AlgKekA256GCMKW, AlgKekA192GCMKW, AlgKekA128GCMKW)
	}

	aesJwk, err := joseJwk.Import(rawkey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to import raw AES raw key: %w", err)
	}
	if err = aesJwk.Set(joseJwk.KeyIDKey, kekKidUuid.String()); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to set `kid` header: %w", err)
	}
	if err = aesJwk.Set(joseJwk.AlgorithmKey, *kekAlg); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to set `alg` header: %w", err)
	}
	if err = aesJwk.Set("enc", *cekAlg); err != nil {
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
