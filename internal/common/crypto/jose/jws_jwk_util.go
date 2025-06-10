package jose

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/json"
	"fmt"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilKeygen "cryptoutil/internal/common/crypto/keygen"
	cryptoutilUtil "cryptoutil/internal/common/util"

	"github.com/cloudflare/circl/sign/ed448"
	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

func GenerateJwsJwkForAlg(alg *joseJwa.SignatureAlgorithm) (*googleUuid.UUID, joseJwk.Key, []byte, error) {
	kid, err := googleUuid.NewV7()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create uuid v7: %w", err)
	}
	key, err := validateJwsJwkHeaders(&kid, alg, nil, true) // true => generates enc key of the correct length
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid JWS JWK headers: %w", err)
	}
	return CreateJwsJwkFromKey(&kid, alg, key)
}

func CreateJwsJwkFromKey(kid *googleUuid.UUID, alg *joseJwa.SignatureAlgorithm, key cryptoutilKeygen.Key) (*googleUuid.UUID, joseJwk.Key, []byte, error) {
	_, err := validateJwsJwkHeaders(kid, alg, key, false)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid JWS JWK headers: %w", err)
	}
	var jwk joseJwk.Key
	switch typedKey := key.(type) {
	case cryptoutilKeygen.SecretKey: // HMAC
		if jwk, err = joseJwk.Import([]byte(typedKey)); err != nil {
			return nil, nil, nil, fmt.Errorf("failed to import key material into JWS JWK: %w", err)
		}
		if err = jwk.Set(joseJwk.KeyTypeKey, KtyOCT); err != nil {
			return nil, nil, nil, fmt.Errorf("failed to set 'kty' header to 'oct' in JWS JWK: %w", err)
		}
	case *cryptoutilKeygen.KeyPair: // RSA, ECDSA, EdDSA
		if jwk, err = joseJwk.Import(typedKey.Private); err != nil {
			return nil, nil, nil, fmt.Errorf("failed to import key pair into JWS JWK: %w", err)
		}
		switch typedKey.Private.(type) {
		case *rsa.PrivateKey: // RSA
			if err = jwk.Set(joseJwk.KeyTypeKey, KtyRSA); err != nil {
				return nil, nil, nil, fmt.Errorf("failed to set 'kty' header to 'rsa' in JWS JWK: %w", err)
			}
		case *ecdsa.PrivateKey: // ECDSA, ECDH
			if err = jwk.Set(joseJwk.KeyTypeKey, KtyEC); err != nil {
				return nil, nil, nil, fmt.Errorf("failed to set 'kty' header to 'ec' in JWS JWK: %w", err)
			}
		case ed25519.PrivateKey, ed448.PrivateKey: // ED25519, ED448
			if err = jwk.Set(joseJwk.KeyTypeKey, KtyOKP); err != nil {
				return nil, nil, nil, fmt.Errorf("failed to set 'kty' header to 'okp' in JWS JWK: %w", err)
			}
		default:
			return nil, nil, nil, fmt.Errorf("failed to set 'kty' header in JWS JWK: unexpected key type %T", key)
		}
	default:
		return nil, nil, nil, fmt.Errorf("unsupported key type %T for JWS JWK", key)
	}

	if err = jwk.Set(joseJwk.KeyIDKey, kid.String()); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to set `kid` header in JWS JWK: %w", err)
	}
	if err = jwk.Set(joseJwk.AlgorithmKey, *alg); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to set `alg` header in JWS JWK: %w", err)
	}
	if err = jwk.Set(joseJwk.KeyUsageKey, "sig"); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to set `use` header in JWS JWK: %w", err)
	}
	if err = jwk.Set(joseJwk.KeyOpsKey, OpsSigVer); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to set `ops` header in JWS JWK: %w", err)
	}

	encodedJwk, err := json.Marshal(jwk)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to serialize JWS JWK: %w", err)
	}

	return kid, jwk, encodedJwk, nil
}

func validateJwsJwkHeaders(kid *googleUuid.UUID, alg *joseJwa.SignatureAlgorithm, key cryptoutilKeygen.Key, isNilRawKeyOk bool) (cryptoutilKeygen.Key, error) {
	if err := cryptoutilUtil.ValidateUUID(kid, "invalid JWS JWK kid"); err != nil {
		return nil, fmt.Errorf("JWS JWK kid must be valid: %w", err)
	} else if alg == nil {
		return nil, fmt.Errorf("JWS JWK alg must be non-nil")
	} else if !isNilRawKeyOk && key == nil {
		return nil, fmt.Errorf("JWS JWK key material must be non-nil")
	}
	switch *alg {
	case AlgRS512, AlgPS512:
		return validateOrGenerateJwsRsaJwk(key, alg, 4096)
	case AlgRS384, AlgPS384:
		return validateOrGenerateJwsRsaJwk(key, alg, 3072)
	case AlgRS256, AlgPS256:
		return validateOrGenerateJwsRsaJwk(key, alg, 2048)
	case AlgES256:
		return validateOrGenerateJwsEcdsaJwk(key, alg, elliptic.P521())
	case AlgES384:
		return validateOrGenerateJwsEcdsaJwk(key, alg, elliptic.P384())
	case AlgES512:
		return validateOrGenerateJwsEcdsaJwk(key, alg, elliptic.P256())
	case AlgEdDSA:
		return validateOrGenerateJwsEddsaJwk(key, alg, "Ed25519")
	case AlgHS512:
		return validateOrGenerateJwsHmacJwk(key, alg, 512)
	case AlgHS384:
		return validateOrGenerateJwsHmacJwk(key, alg, 384)
	case AlgHS256:
		return validateOrGenerateJwsHmacJwk(key, alg, 256)
	default:
		return nil, fmt.Errorf("unsupported JWS JWK alg: %s", alg)
	}
}

func validateOrGenerateJwsRsaJwk(key cryptoutilKeygen.Key, alg *joseJwa.SignatureAlgorithm, keyBitsLength int) (*cryptoutilKeygen.KeyPair, error) {
	if key == nil {
		generatedKey, err := cryptoutilKeygen.GenerateRSAKeyPair(keyBitsLength)
		if err != nil {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but failed to generate RSA %d key: %w", *alg, keyBitsLength, err)
		}
		return generatedKey, nil
	} else {
		keyPair, ok := key.(*cryptoutilKeygen.KeyPair)
		if !ok {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but unsupported key type %T; use *cryptoutilKeygen.KeyPair", *alg, key)
		}
		rsaPrivateKey, ok := keyPair.Private.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but invalid key type %T; use *rsa.PrivateKey", *alg, keyPair.Private)
		} else if rsaPrivateKey == nil {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but invalid nil RSA private key", *alg)
		}
		rsaPublicKey, ok := keyPair.Public.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but invalid key type %T; use *rsa.PublicKey", *alg, keyPair.Public)
		} else if rsaPublicKey == nil {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but invalid nil RSA public key", *alg)
		}
		return keyPair, nil
	}
}

func validateOrGenerateJwsEcdsaJwk(key cryptoutilKeygen.Key, alg *joseJwa.SignatureAlgorithm, curve elliptic.Curve) (*cryptoutilKeygen.KeyPair, error) {
	if key == nil {
		generatedKey, err := cryptoutilKeygen.GenerateECDSAKeyPair(curve)
		if err != nil {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but failed to generate ECDSA %s key pair: %w", *alg, curve, err)
		}
		return generatedKey, nil
	} else {
		keyPair, ok := key.(*cryptoutilKeygen.KeyPair)
		if !ok {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but unsupported key type %T; use *cryptoutilKeygen.KeyPair", *alg, key)
		}
		rsaPrivateKey, ok := keyPair.Private.(*ecdsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but invalid key type %T; use *ecdsa.PrivateKey", *alg, keyPair.Private)
		} else if rsaPrivateKey == nil {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but invalid nil ECDSA private key", *alg)
		}
		rsaPublicKey, ok := keyPair.Public.(*ecdsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but invalid key type %T; use *ecdsa.PublicKey", *alg, keyPair.Public)
		} else if rsaPublicKey == nil {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but invalid nil ECDSA public key", *alg)
		}
		return keyPair, nil
	}
}

func validateOrGenerateJwsEddsaJwk(key cryptoutilKeygen.Key, alg *joseJwa.SignatureAlgorithm, curve string) (*cryptoutilKeygen.KeyPair, error) {
	if key == nil {
		generatedKey, err := cryptoutilKeygen.GenerateEDDSAKeyPair(curve)
		if err != nil {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but failed to generate Ed29919 key pair: %w", *alg, err)
		}
		return generatedKey, nil
	} else {
		keyPair, ok := key.(*cryptoutilKeygen.KeyPair)
		if !ok {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but unsupported key type %T; use *cryptoutilKeygen.KeyPair", *alg, key)
		}
		rsaPrivateKey, ok := keyPair.Private.(ed25519.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but invalid key type %T; use ed25519.PrivateKey", *alg, keyPair.Private)
		} else if rsaPrivateKey == nil {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but invalid nil Ed29919 private key", *alg)
		}
		rsaPublicKey, ok := keyPair.Public.(ed25519.PublicKey)
		if !ok {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but invalid key type %T; use ed25519.PublicKey", *alg, keyPair.Public)
		} else if rsaPublicKey == nil {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but invalid nil Ed29919 public key", *alg)
		}
		return keyPair, nil
	}
}

func validateOrGenerateJwsHmacJwk(key cryptoutilKeygen.Key, alg *joseJwa.SignatureAlgorithm, keyBitsLength int) (cryptoutilKeygen.SecretKey, error) {
	if key == nil {
		generatedKey, err := cryptoutilKeygen.GenerateHMACKey(keyBitsLength)
		if err != nil {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but failed to generate AES %d key: %w", *alg, keyBitsLength, err)
		}
		return generatedKey, nil
	} else {
		hmacKey, ok := key.(cryptoutilKeygen.SecretKey)
		if !ok {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but invalid key type %T; use cryptoKeygen.SecretKey", *alg, key)
		} else if hmacKey == nil {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but invalid nil key bytes", *alg)
		} else if len(hmacKey) != keyBitsLength/8 {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but invalid key length %d; use AES %d", *alg, len(hmacKey), keyBitsLength)
		}
		return hmacKey, nil
	}
}

func ExtractAlgFromJwsJwk(jwk joseJwk.Key, i int) (*joseJwa.SignatureAlgorithm, error) {
	if jwk == nil {
		return nil, fmt.Errorf("JWK %d invalid: %w", i, cryptoutilAppErr.ErrCantBeNil)
	}

	var alg joseJwa.SignatureAlgorithm
	err := jwk.Get(joseJwk.AlgorithmKey, &alg) // Example: RS256, RS384, RS512, ES256, ES384, ES512, PS256, PS384, PS512, EdDSA
	if err != nil {
		return nil, fmt.Errorf("can't get JWK %d 'alg' attribute: %w", i, err)
	}

	return &alg, nil
}
