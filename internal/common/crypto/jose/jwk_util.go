package jose

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"time"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilKeyGen "cryptoutil/internal/common/crypto/keygen"
	cryptoutilUtil "cryptoutil/internal/common/util"
	cryptoutilOpenapiModel "cryptoutil/internal/openapi/model"

	"github.com/cloudflare/circl/sign/ed448"
	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

var (
	ErrInvalidJwkKidUuid = "invalid JWK kid UUID"

	KtyOCT = joseJwa.OctetSeq() // KeyType
	KtyRSA = joseJwa.RSA()      // KeyType
	KtyEC  = joseJwa.EC()       // KeyType
	KtyOKP = joseJwa.OKP()      // KeyType

	EncA256GCM      = joseJwa.A256GCM()                                // ContentEncryptionAlgorithm
	EncA192GCM      = joseJwa.A192GCM()                                // ContentEncryptionAlgorithm
	EncA128GCM      = joseJwa.A128GCM()                                // ContentEncryptionAlgorithm
	EncA256CBCHS512 = joseJwa.A256CBC_HS512()                          // ContentEncryptionAlgorithm
	EncA192CBCHS384 = joseJwa.A192CBC_HS384()                          // ContentEncryptionAlgorithm
	EncA128CBCHS256 = joseJwa.A128CBC_HS256()                          // ContentEncryptionAlgorithm
	EncInvalid      = joseJwa.NewContentEncryptionAlgorithm("invalid") // ContentEncryptionAlgorithm

	AlgA256KW       = joseJwa.A256KW()                             // KeyEncryptionAlgorithm
	AlgA192KW       = joseJwa.A192KW()                             // KeyEncryptionAlgorithm
	AlgA128KW       = joseJwa.A128KW()                             // KeyEncryptionAlgorithm
	AlgA256GCMKW    = joseJwa.A256GCMKW()                          // KeyEncryptionAlgorithm
	AlgA192GCMKW    = joseJwa.A192GCMKW()                          // KeyEncryptionAlgorithm
	AlgA128GCMKW    = joseJwa.A128GCMKW()                          // KeyEncryptionAlgorithm
	AlgRSAOAEP512   = joseJwa.RSA_OAEP_512()                       // KeyEncryptionAlgorithm
	AlgRSAOAEP384   = joseJwa.RSA_OAEP_384()                       // KeyEncryptionAlgorithm
	AlgRSAOAEP256   = joseJwa.RSA_OAEP_256()                       // KeyEncryptionAlgorithm
	AlgRSAOAEP      = joseJwa.RSA_OAEP()                           // KeyEncryptionAlgorithm
	AlgRSA15        = joseJwa.RSA1_5()                             // KeyEncryptionAlgorithm
	AlgECDHESA256KW = joseJwa.ECDH_ES_A256KW()                     // KeyEncryptionAlgorithm
	AlgECDHESA192KW = joseJwa.ECDH_ES_A192KW()                     // KeyEncryptionAlgorithm
	AlgECDHESA128KW = joseJwa.ECDH_ES_A128KW()                     // KeyEncryptionAlgorithm
	AlgECDHES       = joseJwa.ECDH_ES()                            // KeyEncryptionAlgorithm
	AlgDir          = joseJwa.DIRECT()                             // KeyEncryptionAlgorithm
	AlgEncInvalid   = joseJwa.NewKeyEncryptionAlgorithm("invalid") // KeyEncryptionAlgorithm

	AlgRS512      = joseJwa.RS512()                          // SignatureAlgorithm
	AlgRS384      = joseJwa.RS384()                          // SignatureAlgorithm
	AlgRS256      = joseJwa.RS256()                          // SignatureAlgorithm
	AlgPS512      = joseJwa.PS512()                          // SignatureAlgorithm
	AlgPS384      = joseJwa.PS384()                          // SignatureAlgorithm
	AlgPS256      = joseJwa.PS256()                          // SignatureAlgorithm
	AlgES512      = joseJwa.ES512()                          // SignatureAlgorithm
	AlgES384      = joseJwa.ES384()                          // SignatureAlgorithm
	AlgES256      = joseJwa.ES256()                          // SignatureAlgorithm
	AlgHS512      = joseJwa.HS512()                          // SignatureAlgorithm
	AlgHS384      = joseJwa.HS384()                          // SignatureAlgorithm
	AlgHS256      = joseJwa.HS256()                          // SignatureAlgorithm
	AlgEdDSA      = joseJwa.EdDSA()                          // SignatureAlgorithm
	AlgSigInvalid = joseJwa.NewSignatureAlgorithm("invalid") // SignatureAlgorithm

	OpsEncDec = joseJwk.KeyOperationList{joseJwk.KeyOpEncrypt, joseJwk.KeyOpDecrypt} // []KeyOperation
	OpsSigVer = joseJwk.KeyOperationList{joseJwk.KeyOpSign, joseJwk.KeyOpVerify}     // []KeyOperation
	OpsEnc    = joseJwk.KeyOperationList{joseJwk.KeyOpEncrypt}                       // []KeyOperation
	OpsVer    = joseJwk.KeyOperationList{joseJwk.KeyOpVerify}                        // []KeyOperation
)

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
	if err = cryptoutilUtil.ValidateUUID(&kidUuid, &ErrInvalidJwkKidUuid); err != nil {
		return nil, err
	}
	return &kidUuid, nil
}

func GenerateJwkForAlg(alg *cryptoutilOpenapiModel.GenerateAlgorithm) (*googleUuid.UUID, joseJwk.Key, joseJwk.Key, []byte, []byte, error) {
	kid, err := googleUuid.NewV7()
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to create uuid v7: %w", err)
	}
	key, err := validateJwkHeaders2(&kid, alg, nil, true) // true => generates enc key of the correct length
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("invalid JWK headers: %w", err)
	}
	return CreateJwkFromKey(&kid, alg, key)
}

func CreateJwkFromKey(kid *googleUuid.UUID, alg *cryptoutilOpenapiModel.GenerateAlgorithm, key cryptoutilKeyGen.Key) (*googleUuid.UUID, joseJwk.Key, joseJwk.Key, []byte, []byte, error) {
	now := time.Now().UTC().Unix()
	_, err := validateJwkHeaders2(kid, alg, key, false)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("invalid JWK headers: %w", err)
	}
	var nonPublicJwk joseJwk.Key
	switch typedKey := key.(type) {
	case cryptoutilKeyGen.SecretKey: // HMAC
		if nonPublicJwk, err = joseJwk.Import([]byte(typedKey)); err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to import key material into JWK: %w", err)
		}
		if err = nonPublicJwk.Set(joseJwk.KeyTypeKey, KtyOCT); err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'kty' header to 'oct' in JWK: %w", err)
		}
	case *cryptoutilKeyGen.KeyPair: // RSA, ECDSA, EdDSA
		if nonPublicJwk, err = joseJwk.Import(typedKey.Private); err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to import key pair into JWK: %w", err)
		}
		switch typedKey.Private.(type) {
		case *rsa.PrivateKey: // RSA
			if err = nonPublicJwk.Set(joseJwk.KeyTypeKey, KtyRSA); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'kty' header to 'rsa' in JWK: %w", err)
			}
		case *ecdsa.PrivateKey: // ECDSA, ECDH
			if err = nonPublicJwk.Set(joseJwk.KeyTypeKey, KtyEC); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'kty' header to 'ec' in JWK: %w", err)
			}
		case ed25519.PrivateKey, ed448.PrivateKey: // ED25519, ED448
			if err = nonPublicJwk.Set(joseJwk.KeyTypeKey, KtyOKP); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'kty' header to 'okp' in JWK: %w", err)
			}
		default:
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'kty' header in JWK: unexpected key type %T", key)
		}
	default:
		return nil, nil, nil, nil, nil, fmt.Errorf("unsupported key type %T for JWK", key)
	}

	if err = nonPublicJwk.Set(joseJwk.KeyIDKey, kid.String()); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to set `kid` header in JWK: %w", err)
	}
	if err = nonPublicJwk.Set("iat", now); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to set `iat` header in JWK: %w", err)
	}

	clearNonPublicJwkBytes, err := json.Marshal(nonPublicJwk)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to serialize JWK: %w", err)
	}

	var publicJwk joseJwk.Key
	var clearPublicJwkBytes []byte
	if _, ok := key.(*cryptoutilKeyGen.KeyPair); ok { // RSA, EC, ED
		publicJwk, err = nonPublicJwk.PublicKey()
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to get public JWE JWK from private JWE JWK: %w", err)
		}
		clearPublicJwkBytes, err = json.Marshal(publicJwk)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to serialize public JWE JWK: %w", err)
		}
	}

	return kid, nonPublicJwk, publicJwk, clearNonPublicJwkBytes, clearPublicJwkBytes, nil
}

func validateJwkHeaders2(kid *googleUuid.UUID, alg *cryptoutilOpenapiModel.GenerateAlgorithm, key cryptoutilKeyGen.Key, isNilRawKeyOk bool) (cryptoutilKeyGen.Key, error) {
	if err := cryptoutilUtil.ValidateUUID(kid, &ErrInvalidJwkKidUuid); err != nil {
		return nil, fmt.Errorf("JWK kid must be valid: %w", err)
	} else if alg == nil {
		return nil, fmt.Errorf("JWK alg must be non-nil")
	} else if !isNilRawKeyOk && key == nil {
		return nil, fmt.Errorf("JWK key material must be non-nil")
	}
	switch *alg {
	case cryptoutilOpenapiModel.RSA4096:
		return validateOrGenerateRsaJwk(key, 4096)
	case cryptoutilOpenapiModel.RSA3072:
		return validateOrGenerateRsaJwk(key, 3072)
	case cryptoutilOpenapiModel.RSA2048:
		return validateOrGenerateRsaJwk(key, 2048)
	case cryptoutilOpenapiModel.ECP521:
		return validateOrGenerateEcdsaJwk(key, elliptic.P521())
	case cryptoutilOpenapiModel.ECP384:
		return validateOrGenerateEcdsaJwk(key, elliptic.P384())
	case cryptoutilOpenapiModel.ECP256:
		return validateOrGenerateEcdsaJwk(key, elliptic.P256())
	case cryptoutilOpenapiModel.OKPEd25519:
		return validateOrGenerateEddsaJwk(key, "Ed25519")
	case cryptoutilOpenapiModel.Oct512:
		return validateOrGenerateHmacJwk(key, 512)
	case cryptoutilOpenapiModel.Oct384:
		return validateOrGenerateHmacJwk(key, 384)
	case cryptoutilOpenapiModel.Oct256:
		return validateOrGenerateHmacJwk(key, 256)
	case cryptoutilOpenapiModel.Oct192:
		return validateOrGenerateHmacJwk(key, 192)
	case cryptoutilOpenapiModel.Oct128:
		return validateOrGenerateHmacJwk(key, 128)
	// case cryptoutilOpenapiModel.OctUUIDv7:
	default:
		return nil, fmt.Errorf("unsupported JWK alg: %v", alg)
	}
}

func validateOrGenerateRsaJwk(key cryptoutilKeyGen.Key, keyBitsLength int) (*cryptoutilKeyGen.KeyPair, error) {
	if key == nil {
		generatedKey, err := cryptoutilKeyGen.GenerateRSAKeyPair(keyBitsLength)
		if err != nil {
			return nil, fmt.Errorf("failed to generate RSA %d key: %w", keyBitsLength, err)
		}
		return generatedKey, nil
	} else {
		keyPair, ok := key.(*cryptoutilKeyGen.KeyPair)
		if !ok {
			return nil, fmt.Errorf("unsupported key type %T; use *cryptoutilKeyGen.KeyPair", key)
		}
		rsaPrivateKey, ok := keyPair.Private.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("invalid key type %T; use *rsa.PrivateKey", keyPair.Private)
		} else if rsaPrivateKey == nil {
			return nil, fmt.Errorf("invalid nil RSA private key")
		}
		rsaPublicKey, ok := keyPair.Public.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("invalid key type %T; use *rsa.PublicKey", keyPair.Public)
		} else if rsaPublicKey == nil {
			return nil, fmt.Errorf("invalid nil RSA public key")
		}
		return keyPair, nil
	}
}

func validateOrGenerateEcdsaJwk(key cryptoutilKeyGen.Key, curve elliptic.Curve) (*cryptoutilKeyGen.KeyPair, error) {
	if key == nil {
		generatedKey, err := cryptoutilKeyGen.GenerateECDSAKeyPair(curve)
		if err != nil {
			return nil, fmt.Errorf("failed to generate ECDSA %s key pair: %w", curve, err)
		}
		return generatedKey, nil
	} else {
		keyPair, ok := key.(*cryptoutilKeyGen.KeyPair)
		if !ok {
			return nil, fmt.Errorf("unsupported key type %T; use *cryptoutilKeyGen.KeyPair", key)
		}
		rsaPrivateKey, ok := keyPair.Private.(*ecdsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("invalid key type %T; use *ecdsa.PrivateKey", keyPair.Private)
		} else if rsaPrivateKey == nil {
			return nil, fmt.Errorf("invalid nil ECDSA private key")
		}
		rsaPublicKey, ok := keyPair.Public.(*ecdsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("invalid key type %T; use *ecdsa.PublicKey", keyPair.Public)
		} else if rsaPublicKey == nil {
			return nil, fmt.Errorf("invalid nil ECDSA public key")
		}
		return keyPair, nil
	}
}

func validateOrGenerateEddsaJwk(key cryptoutilKeyGen.Key, curve string) (*cryptoutilKeyGen.KeyPair, error) {
	if key == nil {
		generatedKey, err := cryptoutilKeyGen.GenerateEDDSAKeyPair(curve)
		if err != nil {
			return nil, fmt.Errorf("failed to generate Ed29919 key pair: %w", err)
		}
		return generatedKey, nil
	} else {
		keyPair, ok := key.(*cryptoutilKeyGen.KeyPair)
		if !ok {
			return nil, fmt.Errorf("unsupported key type %T; use *cryptoutilKeyGen.KeyPair", key)
		}
		rsaPrivateKey, ok := keyPair.Private.(ed25519.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("invalid key type %T; use ed25519.PrivateKey", keyPair.Private)
		} else if rsaPrivateKey == nil {
			return nil, fmt.Errorf("invalid nil Ed29919 private key")
		}
		rsaPublicKey, ok := keyPair.Public.(ed25519.PublicKey)
		if !ok {
			return nil, fmt.Errorf("invalid key type %T; use ed25519.PublicKey", keyPair.Public)
		} else if rsaPublicKey == nil {
			return nil, fmt.Errorf("invalid nil Ed29919 public key")
		}
		return keyPair, nil
	}
}

func validateOrGenerateHmacJwk(key cryptoutilKeyGen.Key, keyBitsLength int) (cryptoutilKeyGen.SecretKey, error) {
	if key == nil {
		generatedKey, err := cryptoutilKeyGen.GenerateHMACKey(keyBitsLength)
		if err != nil {
			return nil, fmt.Errorf("failed to generate AES %d key: %w", keyBitsLength, err)
		}
		return generatedKey, nil
	} else {
		hmacKey, ok := key.(cryptoutilKeyGen.SecretKey)
		if !ok {
			return nil, fmt.Errorf("invalid key type %T; use cryptoKeygen.SecretKey", key)
		} else if hmacKey == nil {
			return nil, fmt.Errorf("invalid nil key bytes")
		} else if len(hmacKey) != keyBitsLength/8 {
			return nil, fmt.Errorf("invalid key length %d; use AES %d", len(hmacKey), keyBitsLength)
		}
		return hmacKey, nil
	}
}
