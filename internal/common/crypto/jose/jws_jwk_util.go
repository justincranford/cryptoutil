package jose

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	cryptoutilKeygen "cryptoutil/internal/common/crypto/keygen"
	cryptoutilUtil "cryptoutil/internal/common/util"
	"encoding/json"
	"fmt"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

func GenerateJwsJwkForAlg(alg *joseJwa.SignatureAlgorithm) (*googleUuid.UUID, joseJwk.Key, []byte, error) {
	kid := googleUuid.Must(googleUuid.NewV7())
	key, err := ValidateJwsJwkHeaders(&kid, alg, nil, true) // true => generates enc key of the correct length
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid JWK headers: %w", err)
	}
	return CreateJwsJwkFromKey(&kid, alg, key)
}

func GenerateJwsJwkFromKeyPool(alg *joseJwa.SignatureAlgorithm, keyGenPool *cryptoutilKeygen.KeyGenPool) (*googleUuid.UUID, joseJwk.Key, []byte, error) {
	kid := googleUuid.Must(googleUuid.NewV7())
	key := keyGenPool.Get()
	return CreateJwsJwkFromKey(&kid, alg, &key)
}

func CreateJwsJwkFromKey(kid *googleUuid.UUID, alg *joseJwa.SignatureAlgorithm, rawKey *cryptoutilKeygen.Key) (*googleUuid.UUID, joseJwk.Key, []byte, error) {
	_, err := ValidateJwsJwkHeaders(kid, alg, rawKey, false)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid JWS JWK headers: %w", err)
	}
	jwk, err := joseJwk.Import(rawKey.Private) // []byte, RSA, EC, OctetSeq (AES/HMAC)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to import key into JWK: %w", err)
	}
	if err = jwk.Set(joseJwk.KeyIDKey, kid.String()); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to set `kid` header in JWK: %w", err)
	}
	if err = jwk.Set(joseJwk.AlgorithmKey, *alg); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to set `alg` header in JWK: %w", err)
	}
	if err = jwk.Set(joseJwk.KeyUsageKey, "sig"); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to set `use` header in JWK: %w", err)
	}
	if err = jwk.Set(joseJwk.KeyOpsKey, OpsSigVer); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to set `ops` header in JWK: %w", err)
	}
	// TODO RSA, EC, OctetSeq
	// if err = jwk.Set(joseJwk.KeyTypeKey, KtyOct); err != nil {
	// 	return nil, nil, nil, fmt.Errorf("failed to set 'kty' header in JWK: %w", err)
	// }

	encodedJwk, err := json.Marshal(jwk)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to serialize JWK: %w", err)
	}

	return kid, jwk, encodedJwk, nil
}

func ValidateJwsJwkHeaders(kid *googleUuid.UUID, alg *joseJwa.SignatureAlgorithm, rawKey *cryptoutilKeygen.Key, isNilRawKeyOk bool) (*cryptoutilKeygen.Key, error) {
	if err := cryptoutilUtil.ValidateUUID(kid, "invalid kid"); err != nil {
		return nil, fmt.Errorf("kid must be valid: %w", err)
	} else if alg == nil {
		return nil, fmt.Errorf("alg must be non-nil")
	} else if !isNilRawKeyOk && rawKey == nil {
		return nil, fmt.Errorf("key must be non-nil")
	}
	switch *alg {
	case AlgRS512, AlgPS512:
		return validateOrGenerateJwsRsaJwk(nil, alg, 4096)
	case AlgRS384, AlgPS384:
		return validateOrGenerateJwsRsaJwk(nil, alg, 3072)
	case AlgRS256, AlgPS256:
		return validateOrGenerateJwsRsaJwk(nil, alg, 2048)
	case AlgES256:
		return validateOrGenerateJwsEcdsaJwk(nil, alg, elliptic.P521())
	case AlgES384:
		return validateOrGenerateJwsEcdsaJwk(nil, alg, elliptic.P384())
	case AlgES512:
		return validateOrGenerateJwsEcdsaJwk(nil, alg, elliptic.P256())
	case AlgEdDSA:
		return validateOrGenerateJwsEddsaJwk(nil, alg, "Ed25519")
	case AlgHS512:
		return validateOrGenerateJwsHmacJwk(nil, alg, 512)
	case AlgHS384:
		return validateOrGenerateJwsHmacJwk(nil, alg, 384)
	case AlgHS256:
		return validateOrGenerateJwsHmacJwk(nil, alg, 256)
	case AlgNoNoSignature:
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", alg)
	}
}

func validateOrGenerateJwsRsaJwk(key *cryptoutilKeygen.Key, alg *joseJwa.SignatureAlgorithm, keyBitsLength int) (*cryptoutilKeygen.Key, error) {
	if key == nil {
		generatedKey, err := cryptoutilKeygen.GenerateRSAKeyPair(keyBitsLength)
		if err != nil {
			return nil, fmt.Errorf("valid alg %s, but failed to generate RSA %d key: %w", *alg, keyBitsLength, err)
		}
		key = &generatedKey
	} else {
		rsaPrivateKey, ok := key.Private.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("valid alg %s, but invalid key type %T; use *rsa.PrivateKey", *alg, key.Private)
		} else if rsaPrivateKey == nil {
			return nil, fmt.Errorf("valid alg %s, but invalid nil RSA private key", *alg)
		}
		rsaPublicKey, ok := key.Public.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("valid alg %s, but invalid key type %T; use *rsa.PublicKey", *alg, key.Public)
		} else if rsaPublicKey == nil {
			return nil, fmt.Errorf("valid alg %s, but invalid nil RSA public key", *alg)
		}
	}
	return key, nil
}

func validateOrGenerateJwsEcdsaJwk(key *cryptoutilKeygen.Key, alg *joseJwa.SignatureAlgorithm, curve elliptic.Curve) (*cryptoutilKeygen.Key, error) {
	if key == nil {
		generatedKey, err := cryptoutilKeygen.GenerateECDSAKeyPair(curve)
		if err != nil {
			return nil, fmt.Errorf("valid alg %s, but failed to generate ECDSA %s key pair: %w", *alg, curve, err)
		}
		key = &generatedKey
	} else {
		rsaPrivateKey, ok := key.Private.(*ecdsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("valid alg %s, but invalid key type %T; use *ecdsa.PrivateKey", *alg, key.Private)
		} else if rsaPrivateKey == nil {
			return nil, fmt.Errorf("valid alg %s, but invalid nil ECDSA private key", *alg)
		}
		rsaPublicKey, ok := key.Public.(*ecdsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("valid alg %s, but invalid key type %T; use *ecdsa.PublicKey", *alg, key.Public)
		} else if rsaPublicKey == nil {
			return nil, fmt.Errorf("valid alg %s, but invalid nil ECDSA public key", *alg)
		}
	}
	return key, nil
}

func validateOrGenerateJwsEddsaJwk(key *cryptoutilKeygen.Key, alg *joseJwa.SignatureAlgorithm, curve string) (*cryptoutilKeygen.Key, error) {
	if key == nil {
		generatedKey, err := cryptoutilKeygen.GenerateEDDSAKeyPair(curve)
		if err != nil {
			return nil, fmt.Errorf("valid alg %s, but failed to generate Ed29919 key pair: %w", *alg, err)
		}
		key = &generatedKey
	} else {
		rsaPrivateKey, ok := key.Private.(ed25519.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("valid alg %s, but invalid key type %T; use ed25519.PrivateKey", *alg, key.Private)
		} else if rsaPrivateKey == nil {
			return nil, fmt.Errorf("valid alg %s, but invalid nil Ed29919 private key", *alg)
		}
		rsaPublicKey, ok := key.Public.(ed25519.PublicKey)
		if !ok {
			return nil, fmt.Errorf("valid alg %s, but invalid key type %T; use ed25519.PublicKey", *alg, key.Public)
		} else if rsaPublicKey == nil {
			return nil, fmt.Errorf("valid alg %s, but invalid nil Ed29919 public key", *alg)
		}
	}
	return key, nil
}

func validateOrGenerateJwsHmacJwk(key *cryptoutilKeygen.Key, alg *joseJwa.SignatureAlgorithm, keyBitsLength int) (*cryptoutilKeygen.Key, error) {
	if key == nil {
		keyBytes, err := cryptoutilKeygen.GenerateBytes(keyBitsLength / 8)
		if err != nil {
			return nil, fmt.Errorf("valid alg %s, but failed to generate AES %d key: %w", *alg, keyBitsLength, err)
		}
		key = &cryptoutilKeygen.Key{Private: keyBytes}
	} else {
		aesKey, ok := key.Private.([]byte)
		if !ok {
			return nil, fmt.Errorf("valid alg %s, but invalid key type %T; use []byte", *alg, key.Private)
		} else if aesKey == nil {
			return nil, fmt.Errorf("valid alg %s, but invalid nil key bytes", *alg)
		} else if len(aesKey) != keyBitsLength/8 {
			return nil, fmt.Errorf("valid alg %s, but invalid key length %d; use AES %d", *alg, len(aesKey), keyBitsLength)
		}
	}
	return key, nil
}
