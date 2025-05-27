package jose

import (
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"encoding/json"
	"fmt"

	"github.com/cloudflare/circl/sign/ed448"

	cryptoutilKeygen "cryptoutil/internal/common/crypto/keygen"
	cryptoutilPool "cryptoutil/internal/common/pool"
	cryptoutilUtil "cryptoutil/internal/common/util"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

func GenerateJweJwkForEncAndAlg(enc *joseJwa.ContentEncryptionAlgorithm, alg *joseJwa.KeyEncryptionAlgorithm) (*googleUuid.UUID, joseJwk.Key, []byte, error) {
	kid, err := googleUuid.NewV7()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create uuid v7: %w", err)
	}
	key, err := validateJweJwkHeaders(&kid, enc, alg, nil, true) // true => generates enc key of the correct length
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid JWE JWK headers: %w", err)
	}
	return CreateJweJwkFromKey(&kid, enc, alg, key)
}

func GenerateJweJwkFromKeyPool(enc *joseJwa.ContentEncryptionAlgorithm, alg *joseJwa.KeyEncryptionAlgorithm, kidGenPool *cryptoutilPool.ValueGenPool[*googleUuid.UUID], keyGenPool *cryptoutilPool.ValueGenPool[cryptoutilKeygen.Key]) (*googleUuid.UUID, joseJwk.Key, []byte, error) {
	kidUuid := kidGenPool.Get()
	key := keyGenPool.Get()
	return CreateJweJwkFromKey(kidUuid, enc, alg, &key)
}

func CreateJweJwkFromKey(kid *googleUuid.UUID, enc *joseJwa.ContentEncryptionAlgorithm, alg *joseJwa.KeyEncryptionAlgorithm, rawKey *cryptoutilKeygen.Key) (*googleUuid.UUID, joseJwk.Key, []byte, error) {
	_, err := validateJweJwkHeaders(kid, enc, alg, rawKey, false)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid JWE JWK headers: %w", err)
	}
	var jwk joseJwk.Key
	if rawKey == nil {
		return nil, nil, nil, fmt.Errorf("JWE JWK key must be non-nil")
	} else if rawKey.Secret != nil {
		if rawKey.Private != nil {
			return nil, nil, nil, fmt.Errorf("invalid mix of non-nil Secret and non-nil Private: %w", err)
		} else if rawKey.Public != nil {
			return nil, nil, nil, fmt.Errorf("invalid mix of non-nil Secret and non-nil Public: %w", err)
		}
		jwk, err = joseJwk.Import(rawKey.Secret) // []byte, OctetSeq (AES/HMAC)
	} else if rawKey.Private != nil {
		if rawKey.Public == nil {
			return nil, nil, nil, fmt.Errorf("invalid mix of non-nil Private and nil Public: %w", err)
		} else if rawKey.Secret != nil {
			return nil, nil, nil, fmt.Errorf("invalid mix of non-nil Private and non-nil Secret: %w", err)
		}
		jwk, err = joseJwk.Import(rawKey.Private) // RSA, EC, ED
	} else {
		return nil, nil, nil, fmt.Errorf("missing Secret and Private: %w", err)
	}
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to import key material into JWE JWK: %w", err)
	}
	if err = jwk.Set(joseJwk.KeyIDKey, kid.String()); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to set `kid` header in JWE JWK: %w", err)
	}
	if err = jwk.Set(joseJwk.AlgorithmKey, *alg); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to set `alg` header in JWE JWK: %w", err)
	}
	if err = jwk.Set("enc", *enc); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to set `alg` header in JWE JWK: %w", err)
	}
	if err = jwk.Set(joseJwk.KeyUsageKey, "enc"); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to set `enc` header in JWE JWK: %w", err)
	}
	if err = jwk.Set(joseJwk.KeyOpsKey, OpsEncDec); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to set `ops` header in JWK: %w", err)
	}
	if rawKey.Secret != nil {
		switch rawKey.Secret.(type) {
		case []byte: // AES, AESCBC-HS, HMAC
			if err = jwk.Set(joseJwk.KeyTypeKey, KtyOCT); err != nil {
				return nil, nil, nil, fmt.Errorf("failed to set 'kty' header to 'oct' in JWE JWK: %w", err)
			}
		default:
			return nil, nil, nil, fmt.Errorf("failed to set 'kty' header in JWE JWK: unexpected key type %T", rawKey.Secret)
		}
	} else if rawKey.Private != nil {
		switch rawKey.Private.(type) {
		case *rsa.PrivateKey: // RSA
			if err = jwk.Set(joseJwk.KeyTypeKey, KtyRSA); err != nil {
				return nil, nil, nil, fmt.Errorf("failed to set 'kty' header to 'rsa' in JWE JWK: %w", err)
			}
		case *ecdsa.PrivateKey, *ecdh.PrivateKey: // ECDSA, ECDH
			if err = jwk.Set(joseJwk.KeyTypeKey, KtyEC); err != nil {
				return nil, nil, nil, fmt.Errorf("failed to set 'kty' header to 'ec' in JWE JWK: %w", err)
			}
		case ed25519.PrivateKey, ed448.PrivateKey: // ED25519, ED448
			if err = jwk.Set(joseJwk.KeyTypeKey, KtyOKP); err != nil {
				return nil, nil, nil, fmt.Errorf("failed to set 'kty' header to 'okp' in JWE JWK: %w", err)
			}
		default:
			return nil, nil, nil, fmt.Errorf("failed to set 'kty' header in JWE JWK: unexpected key type %T", rawKey.Secret)
		}
	}

	encodedJwk, err := json.Marshal(jwk)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to serialize JWE JWK: %w", err)
	}

	return kid, jwk, encodedJwk, nil
}

func validateJweJwkHeaders(kid *googleUuid.UUID, enc *joseJwa.ContentEncryptionAlgorithm, alg *joseJwa.KeyEncryptionAlgorithm, rawKey *cryptoutilKeygen.Key, isNilRawKeyOk bool) (*cryptoutilKeygen.Key, error) {
	if err := cryptoutilUtil.ValidateUUID(kid, "invalid JWE JWK kid"); err != nil {
		return nil, fmt.Errorf("JWE JWK kid must be valid: %w", err)
	} else if alg == nil {
		return nil, fmt.Errorf("JWE JWK alg must be non-nil")
	} else if enc == nil {
		return nil, fmt.Errorf("JWE JWK enc must be non-nil")
	} else if !isNilRawKeyOk && rawKey == nil {
		return nil, fmt.Errorf("JWE JWK key must be non-nil")
	}
	encKeyBitsLength, err := EncToBitsLength(enc)
	if err != nil {
		return nil, fmt.Errorf("JWE JWK length error: %w", err)
	}

	switch *alg {
	case AlgDir:
		return validateOrGenerateJweAesJwk(rawKey, enc, alg, encKeyBitsLength, &EncA256GCM, &EncA256CBC_HS512, &EncA192GCM, &EncA192CBC_HS384, &EncA128GCM, &EncA128CBC_HS256)

	case AlgA256KW, AlgA256GCMKW:
		return validateOrGenerateJweAesJwk(rawKey, enc, alg, 256, &EncA256GCM, &EncA256CBC_HS512, &EncA192GCM, &EncA192CBC_HS384, &EncA128GCM, &EncA128CBC_HS256)
	case AlgA192KW, AlgA192GCMKW:
		return validateOrGenerateJweAesJwk(rawKey, enc, alg, 192, &EncA192GCM, &EncA192CBC_HS384, &EncA128GCM, &EncA128CBC_HS256)
	case AlgA128KW, AlgA128GCMKW:
		return validateOrGenerateJweAesJwk(rawKey, enc, alg, 128, &EncA128GCM, &EncA128CBC_HS256)

	case AlgRSAOAEP512:
		return validateOrGenerateJweRsaJwk(rawKey, enc, alg, 4096, &EncA256GCM, &EncA256CBC_HS512, &EncA192GCM, &EncA192CBC_HS384, &EncA128GCM, &EncA128CBC_HS256)
	case AlgRSAOAEP384:
		return validateOrGenerateJweRsaJwk(rawKey, enc, alg, 3072, &EncA192GCM, &EncA192CBC_HS384, &EncA128GCM, &EncA128CBC_HS256)
	case AlgRSAOAEP256, AlgRSA15, AlgRSAOAEP:
		return validateOrGenerateJweRsaJwk(rawKey, enc, alg, 2048, &EncA128GCM, &EncA128CBC_HS256)

	case AlgECDHES, AlgECDHESA256KW:
		return validateOrGenerateJweEcdhJwk(rawKey, enc, alg, ecdh.P521(), &EncA256GCM, &EncA256CBC_HS512, &EncA192GCM, &EncA192CBC_HS384, &EncA128GCM, &EncA128CBC_HS256)
	case AlgECDHESA192KW:
		return validateOrGenerateJweEcdhJwk(rawKey, enc, alg, ecdh.P384(), &EncA192GCM, &EncA192CBC_HS384, &EncA128GCM, &EncA128CBC_HS256)
	case AlgECDHESA128KW:
		return validateOrGenerateJweEcdhJwk(rawKey, enc, alg, ecdh.P256(), &EncA128GCM, &EncA128CBC_HS256)

	default:
		return nil, fmt.Errorf("unsupported JWE JWK alg %s", *alg)
	}
}

func validateOrGenerateJweAesJwk(key *cryptoutilKeygen.Key, enc *joseJwa.ContentEncryptionAlgorithm, alg *joseJwa.KeyEncryptionAlgorithm, keyBitsLength int, allowedEncs ...*joseJwa.ContentEncryptionAlgorithm) (*cryptoutilKeygen.Key, error) {
	if !cryptoutilUtil.Contains(allowedEncs, enc) {
		return nil, fmt.Errorf("valid JWE JWK alg %s, but enc %s not allowed; use one of %v", *alg, *enc, allowedEncs)
	} else if key == nil {
		aesKeyBytes, err := cryptoutilUtil.GenerateBytes(keyBitsLength / 8)
		if err != nil {
			return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but failed to generate AES %d key: %w", *enc, *alg, keyBitsLength, err)
		}
		key = &cryptoutilKeygen.Key{Secret: aesKeyBytes}
	} else {
		aesKeyBytes, ok := key.Secret.([]byte)
		if !ok {
			return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but unsupported key type %T; use []byte", *enc, *alg, key.Secret)
		} else if aesKeyBytes == nil {
			return nil, fmt.Errorf("valid enc %s and alg %s, but invalid nil key bytes", *enc, *alg)
		} else if len(aesKeyBytes) != keyBitsLength/8 {
			return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but invalid key length %d; use AES %d", *enc, *alg, len(aesKeyBytes), keyBitsLength)
		}
	}
	return key, nil
}

func validateOrGenerateJweRsaJwk(key *cryptoutilKeygen.Key, enc *joseJwa.ContentEncryptionAlgorithm, alg *joseJwa.KeyEncryptionAlgorithm, keyBitsLength int, allowedEncs ...*joseJwa.ContentEncryptionAlgorithm) (*cryptoutilKeygen.Key, error) {
	if !cryptoutilUtil.Contains(allowedEncs, enc) {
		return nil, fmt.Errorf("valid JWE JWK alg %s, but enc %s not allowed; use one of %v", *alg, *enc, allowedEncs)
	} else if key == nil {
		generatedKey, err := cryptoutilKeygen.GenerateRSAKeyPair(keyBitsLength)
		if err != nil {
			return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but failed to generate RSA %d key: %w", *enc, *alg, keyBitsLength, err)
		}
		key = &generatedKey
	} else {
		rsaPrivateKey, ok := key.Private.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but unsupported key type %T; use *rsa.PrivateKey", *enc, *alg, key.Private)
		} else if rsaPrivateKey == nil {
			return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but invalid nil RSA private key", *enc, *alg)
		}
		rsaPublicKey, ok := key.Public.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but unsupported key type %T; use *rsa.PublicKey", *enc, *alg, key.Public)
		} else if rsaPublicKey == nil {
			return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but invalid nil RSA public key", *enc, *alg)
		}
	}
	return key, nil
}

func validateOrGenerateJweEcdhJwk(key *cryptoutilKeygen.Key, enc *joseJwa.ContentEncryptionAlgorithm, alg *joseJwa.KeyEncryptionAlgorithm, ecdhCurve ecdh.Curve, allowedEncs ...*joseJwa.ContentEncryptionAlgorithm) (*cryptoutilKeygen.Key, error) {
	if !cryptoutilUtil.Contains(allowedEncs, enc) {
		return nil, fmt.Errorf("valid JWE JWK alg %s, but enc %s not allowed; use one of %v", *alg, *enc, allowedEncs)
	} else if key == nil {
		generatedEcdhKeyPair, err := cryptoutilKeygen.GenerateECDHKeyPair(ecdhCurve)
		if err != nil {
			return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but failed to generate ECDH %s key: %w", *enc, *alg, ecdhCurve, err)
		}
		key = &generatedEcdhKeyPair
	} else {
		ecdhPrivateKey, ok := key.Private.(*ecdh.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but unsupported key type %T; use *ecdh.PrivateKey", *enc, *alg, key.Private)
		} else if ecdhPrivateKey == nil {
			return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but unsupported nil ECDH private key", *enc, *alg)
		}
		ecdhPublicKey, ok := key.Public.(*ecdh.PublicKey)
		if !ok {
			return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but unsupported key type %T; use *ecdh.PublicKey", *enc, *alg, key.Public)
		} else if ecdhPublicKey == nil {
			return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but unsupported nil ECDH public key", *enc, *alg)
		}
	}
	return key, nil
}

func EncToBitsLength(enc *joseJwa.ContentEncryptionAlgorithm) (int, error) {
	switch *enc {
	case EncA256GCM:
		return 256, nil
	case EncA192GCM:
		return 192, nil
	case EncA128GCM:
		return 128, nil
	case EncA256CBC_HS512:
		return 512, nil
	case EncA192CBC_HS384:
		return 384, nil
	case EncA128CBC_HS256:
		return 256, nil
	default:
		return 0, fmt.Errorf("unsupported JWE JWK enc %s", *enc)
	}
}
