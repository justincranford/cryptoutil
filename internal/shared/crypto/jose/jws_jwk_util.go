// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"time"

	cryptoutilAppErr "cryptoutil/internal/shared/apperr"
	cryptoutilKeyGen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilRandom "cryptoutil/internal/shared/util/random"

	"github.com/cloudflare/circl/sign/ed448"
	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

// ErrInvalidJWSJWKKidUUID indicates the JWS JWK key ID is not a valid UUID.
var ErrInvalidJWSJWKKidUUID = "invalid JWS JWK kid UUID"

const (
	algRS512 = "RS512"
	algPS512 = "PS512"
	algRS384 = "RS384"
	algPS384 = "PS384"
	algRS256 = "RS256"
	algPS256 = "PS256"
	algES512 = "ES512"
	algES384 = "ES384"
	algES256 = "ES256"
	algEdDSA = "EdDSA"
	algHS512 = "HS512"
	algHS384 = "HS384"
	algHS256 = "HS256"
)

// GenerateJWSJWKForAlg generates a JWS JWK for the specified signature algorithm.
func GenerateJWSJWKForAlg(alg *joseJwa.SignatureAlgorithm) (*googleUuid.UUID, joseJwk.Key, joseJwk.Key, []byte, []byte, error) {
	kid, err := googleUuid.NewV7()
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to create uuid v7: %w", err)
	}

	key, err := validateJWSJWKHeaders(&kid, alg, nil, true) // true => generates enc key of the correct length
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("invalid JWS JWK headers: %w", err)
	}

	return CreateJWSJWKFromKey(&kid, alg, key)
}

// CreateJWSJWKFromKey creates a JWS JWK from an existing cryptographic key.
func CreateJWSJWKFromKey(kid *googleUuid.UUID, alg *joseJwa.SignatureAlgorithm, key cryptoutilKeyGen.Key) (*googleUuid.UUID, joseJwk.Key, joseJwk.Key, []byte, []byte, error) {
	now := time.Now().UTC().Unix()

	_, err := validateJWSJWKHeaders(kid, alg, key, false)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("invalid JWS JWK headers: %w", err)
	}

	var nonPublicJWK joseJwk.Key

	switch typedKey := key.(type) {
	case cryptoutilKeyGen.SecretKey: // HMAC // pragma: allowlist secret
		if nonPublicJWK, err = joseJwk.Import([]byte(typedKey)); err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to import key material into JWS JWK: %w", err)
		}

		if err = nonPublicJWK.Set(joseJwk.KeyTypeKey, KtyOCT); err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'kty' header to 'oct' in JWS JWK: %w", err)
		}
	case *cryptoutilKeyGen.KeyPair: // RSA, ECDSA, EdDSA
		if nonPublicJWK, err = joseJwk.Import(typedKey.Private); err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to import key pair into JWS JWK: %w", err)
		}

		switch typedKey.Private.(type) {
		case *rsa.PrivateKey: // RSA
			if err = nonPublicJWK.Set(joseJwk.KeyTypeKey, KtyRSA); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'kty' header to 'rsa' in JWS JWK: %w", err)
			}
		case *ecdsa.PrivateKey: // ECDSA, ECDH
			if err = nonPublicJWK.Set(joseJwk.KeyTypeKey, KtyEC); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'kty' header to 'ec' in JWS JWK: %w", err)
			}
		case ed25519.PrivateKey, ed448.PrivateKey: // ED25519, ED448
			if err = nonPublicJWK.Set(joseJwk.KeyTypeKey, KtyOKP); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'kty' header to 'okp' in JWS JWK: %w", err)
			}
		default:
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'kty' header in JWS JWK: unexpected key type %T", key)
		}
	default:
		return nil, nil, nil, nil, nil, fmt.Errorf("unsupported key type %T for JWS JWK", key)
	}

	if err = nonPublicJWK.Set(joseJwk.KeyIDKey, kid.String()); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to set `kid` header in JWS JWK: %w", err)
	}

	if err = nonPublicJWK.Set(joseJwk.AlgorithmKey, *alg); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to set `alg` header in JWS JWK: %w", err)
	}

	if err = nonPublicJWK.Set("iat", now); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to set `iat` header in JWS JWK: %w", err)
	}

	if err = nonPublicJWK.Set(joseJwk.KeyUsageKey, joseJwk.ForSignature); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to set `use` header in JWS JWK: %w", err)
	}

	if err = nonPublicJWK.Set(joseJwk.KeyOpsKey, OpsSigVer); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to set `ops` header in JWS JWK: %w", err)
	}

	clearNonPublicJWKBytes, err := json.Marshal(nonPublicJWK)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to serialize JWS JWK: %w", err)
	}

	var publicJWK joseJwk.Key

	var clearPublicJWKBytes []byte

	if _, ok := key.(*cryptoutilKeyGen.KeyPair); ok { // RSA, EC, ED
		publicJWK, err = nonPublicJWK.PublicKey()
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to get public JWE JWK from private JWE JWK: %w", err)
		}

		if err = publicJWK.Set(joseJwk.KeyOpsKey, OpsVer); err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to set `ops` header in JWE JWK: %w", err)
		}

		clearPublicJWKBytes, err = json.Marshal(publicJWK)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to serialize public JWE JWK: %w", err)
		}
	}

	return kid, nonPublicJWK, publicJWK, clearNonPublicJWKBytes, clearPublicJWKBytes, nil
}

func validateJWSJWKHeaders(kid *googleUuid.UUID, alg *joseJwa.SignatureAlgorithm, key cryptoutilKeyGen.Key, isNilRawKeyOk bool) (cryptoutilKeyGen.Key, error) {
	if err := cryptoutilRandom.ValidateUUID(kid, &ErrInvalidJWSJWKKidUUID); err != nil {
		return nil, fmt.Errorf("JWS JWK kid must be valid: %w", err)
	} else if alg == nil {
		return nil, fmt.Errorf("JWS JWK alg must be non-nil")
	} else if !isNilRawKeyOk && key == nil {
		return nil, fmt.Errorf("JWS JWK key material must be non-nil")
	}

	switch (*alg).String() {
	case algRS512, algPS512:
		return validateOrGenerateJWSRSAJWK(key, *alg, cryptoutilMagic.RSAKeySize4096)
	case algRS384, algPS384:
		return validateOrGenerateJWSRSAJWK(key, *alg, cryptoutilMagic.RSAKeySize3072)
	case algRS256, algPS256:
		return validateOrGenerateJWSRSAJWK(key, *alg, cryptoutilMagic.RSAKeySize2048)
	case algES512:
		return validateOrGenerateJWSEcdsaJWK(key, *alg, elliptic.P521())
	case algES384:
		return validateOrGenerateJWSEcdsaJWK(key, *alg, elliptic.P384())
	case algES256:
		return validateOrGenerateJWSEcdsaJWK(key, *alg, elliptic.P256())
	case algEdDSA:
		return validateOrGenerateJWSEddsaJWK(key, *alg, "Ed25519")
	case algHS512:
		return validateOrGenerateJWSHMACJWK(key, *alg, cryptoutilMagic.HMACKeySize512)
	case algHS384:
		return validateOrGenerateJWSHMACJWK(key, *alg, cryptoutilMagic.HMACKeySize384)
	case algHS256:
		return validateOrGenerateJWSHMACJWK(key, *alg, cryptoutilMagic.HMACKeySize256)
	default:
		return nil, fmt.Errorf("unsupported JWS JWK alg: %s", alg)
	}
}

func validateOrGenerateJWSRSAJWK(key cryptoutilKeyGen.Key, alg joseJwa.SignatureAlgorithm, keyBitsLength int) (*cryptoutilKeyGen.KeyPair, error) {
	if key == nil {
		generatedKey, err := cryptoutilKeyGen.GenerateRSAKeyPair(keyBitsLength)
		if err != nil {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but failed to generate RSA %d key: %w", alg, keyBitsLength, err)
		}

		return generatedKey, nil
	} else {
		keyPair, ok := key.(*cryptoutilKeyGen.KeyPair)
		if !ok {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but unsupported key type %T; use *cryptoutilKeyGen.KeyPair", alg, key)
		}

		rsaPrivateKey, ok := keyPair.Private.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but invalid key type %T; use *rsa.PrivateKey", alg, keyPair.Private)
		} else if rsaPrivateKey == nil {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but invalid nil RSA private key", alg)
		}

		rsaPublicKey, ok := keyPair.Public.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but invalid key type %T; use *rsa.PublicKey", alg, keyPair.Public)
		} else if rsaPublicKey == nil {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but invalid nil RSA public key", alg)
		}

		return keyPair, nil
	}
}

func validateOrGenerateJWSEcdsaJWK(key cryptoutilKeyGen.Key, alg joseJwa.SignatureAlgorithm, curve elliptic.Curve) (*cryptoutilKeyGen.KeyPair, error) {
	if key == nil {
		generatedKey, err := cryptoutilKeyGen.GenerateECDSAKeyPair(curve)
		if err != nil {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but failed to generate ECDSA %s key pair: %w", alg, curve, err)
		}

		return generatedKey, nil
	} else {
		keyPair, ok := key.(*cryptoutilKeyGen.KeyPair)
		if !ok {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but unsupported key type %T; use *cryptoutilKeyGen.KeyPair", alg, key)
		}

		rsaPrivateKey, ok := keyPair.Private.(*ecdsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but invalid key type %T; use *ecdsa.PrivateKey", alg, keyPair.Private)
		} else if rsaPrivateKey == nil {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but invalid nil ECDSA private key", alg)
		}

		rsaPublicKey, ok := keyPair.Public.(*ecdsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but invalid key type %T; use *ecdsa.PublicKey", alg, keyPair.Public)
		} else if rsaPublicKey == nil {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but invalid nil ECDSA public key", alg)
		}

		return keyPair, nil
	}
}

func validateOrGenerateJWSEddsaJWK(key cryptoutilKeyGen.Key, alg joseJwa.SignatureAlgorithm, curve string) (*cryptoutilKeyGen.KeyPair, error) {
	if key == nil {
		generatedKey, err := cryptoutilKeyGen.GenerateEDDSAKeyPair(curve)
		if err != nil {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but failed to generate Ed29919 key pair: %w", alg, err)
		}

		return generatedKey, nil
	} else {
		keyPair, ok := key.(*cryptoutilKeyGen.KeyPair)
		if !ok {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but unsupported key type %T; use *cryptoutilKeyGen.KeyPair", alg, key)
		}

		rsaPrivateKey, ok := keyPair.Private.(ed25519.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but invalid key type %T; use ed25519.PrivateKey", alg, keyPair.Private)
		} else if rsaPrivateKey == nil {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but invalid nil Ed29919 private key", alg)
		}

		rsaPublicKey, ok := keyPair.Public.(ed25519.PublicKey)
		if !ok {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but invalid key type %T; use ed25519.PublicKey", alg, keyPair.Public)
		} else if rsaPublicKey == nil {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but invalid nil Ed29919 public key", alg)
		}

		return keyPair, nil
	}
}

func validateOrGenerateJWSHMACJWK(key cryptoutilKeyGen.Key, alg joseJwa.SignatureAlgorithm, keyBitsLength int) (cryptoutilKeyGen.SecretKey, error) { // pragma: allowlist secret
	if key == nil {
		generatedKey, err := cryptoutilKeyGen.GenerateHMACKey(keyBitsLength)
		if err != nil {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but failed to generate AES %d key: %w", alg, keyBitsLength, err)
		}

		return generatedKey, nil
	} else {
		hmacKey, ok := key.(cryptoutilKeyGen.SecretKey) // pragma: allowlist secret
		if !ok {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but invalid key type %T; use cryptoKeygen.SecretKey", alg, key) // pragma: allowlist secret
		} else if hmacKey == nil {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but invalid nil key bytes", alg)
		} else if len(hmacKey) != keyBitsLength/8 {
			return nil, fmt.Errorf("valid JWS JWK alg %s, but invalid key length %d; use AES %d", alg, len(hmacKey), keyBitsLength)
		}

		return hmacKey, nil
	}
}

// ExtractAlgFromJWSJWK extracts the signature algorithm from a JWS JWK.
func ExtractAlgFromJWSJWK(jwk joseJwk.Key, i int) (*joseJwa.SignatureAlgorithm, error) {
	if jwk == nil {
		return nil, fmt.Errorf("JWK %d invalid: %w", i, cryptoutilAppErr.ErrCantBeNil)
	}

	// Retrieve the algorithm via the helper which returns a generic KeyAlgorithm.
	keyAlg, ok := jwk.Algorithm()
	if !ok {
		return nil, fmt.Errorf("can't get JWK %d 'alg' attribute: missing algorithm", i)
	}

	// Ensure it's a signature algorithm (not a key encryption algorithm).
	if sigAlg, isSig := keyAlg.(joseJwa.SignatureAlgorithm); isSig {
		return &sigAlg, nil
	}

	return nil, fmt.Errorf("JWK %d 'alg' is not a signature algorithm", i)
}

// IsJWSAlg returns true if the algorithm is a JWS signature algorithm.
func IsJWSAlg(alg *joseJwa.KeyAlgorithm, i int) (bool, error) {
	if alg == nil {
		return false, fmt.Errorf("alg %d invalid: %w", i, cryptoutilAppErr.ErrCantBeNil)
	}

	_, ok := (*alg).(joseJwa.SignatureAlgorithm)

	return ok, nil
}
