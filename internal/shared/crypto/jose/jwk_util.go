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

	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilKeyGen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"

	"github.com/cloudflare/circl/sign/ed448"
	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

// JWK algorithm and key type constants and error values.
var (
	// ErrInvalidJWKKidUUID indicates the JWK key ID is not a valid UUID.
	ErrInvalidJWKKidUUID = "invalid JWK kid UUID"

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

	// String constants for algorithm names to avoid goconst warnings.
	algStrHS256 = "HS256"
	algStrHS384 = "HS384"
	algStrHS512 = "HS512"
	algStrRS256 = "RS256"
	algStrRS384 = "RS384"
	algStrRS512 = "RS512"
	algStrPS256 = "PS256"
	algStrPS384 = "PS384"
	algStrPS512 = "PS512"
	algStrES256 = "ES256"
	algStrES384 = "ES384"
	algStrES512 = "ES512"
	algStrEdDSA = "EdDSA"

	OpsEncDec = joseJwk.KeyOperationList{joseJwk.KeyOpEncrypt, joseJwk.KeyOpDecrypt} // []KeyOperation
	OpsSigVer = joseJwk.KeyOperationList{joseJwk.KeyOpSign, joseJwk.KeyOpVerify}     // []KeyOperation
	OpsEnc    = joseJwk.KeyOperationList{joseJwk.KeyOpEncrypt}                       // []KeyOperation
	OpsVer    = joseJwk.KeyOperationList{joseJwk.KeyOpVerify}                        // []KeyOperation
)

// ExtractKidUUID extracts the key ID as a UUID from a JWK.
func ExtractKidUUID(jwk joseJwk.Key) (*googleUuid.UUID, error) {
	if jwk == nil {
		return nil, fmt.Errorf("invalid jwk: %w", cryptoutilSharedApperr.ErrCantBeNil)
	}

	var err error

	var kidString string
	if err = jwk.Get(joseJwk.KeyIDKey, &kidString); err != nil {
		return nil, fmt.Errorf("failed to get kid header: %w", err)
	}

	var kidUUID googleUuid.UUID

	if kidUUID, err = googleUuid.Parse(kidString); err != nil {
		return nil, fmt.Errorf("failed to parse kid as UUID: %w", err)
	}

	if err = cryptoutilSharedUtilRandom.ValidateUUID(&kidUUID, &ErrInvalidJWKKidUUID); err != nil {
		return nil, fmt.Errorf("failed to validate kid UUID: %w", err)
	}

	return &kidUUID, nil
}

// ExtractAlg extracts the algorithm from a JWK.
func ExtractAlg(jwk joseJwk.Key) (*cryptoutilOpenapiModel.GenerateAlgorithm, error) {
	if jwk == nil {
		return nil, fmt.Errorf("invalid jwk: %w", cryptoutilSharedApperr.ErrCantBeNil)
	}

	alg, ok := jwk.Algorithm()
	if !ok {
		return nil, fmt.Errorf("failed to get alg header: missing algorithm")
	}

	algString := alg.String()

	generateAlg, err := ToGenerateAlgorithm(&algString)
	if err != nil {
		return nil, fmt.Errorf("failed to map to generate alg: %w", err)
	}

	return generateAlg, nil
}

// ExtractKty extracts the key type from a JWK.
func ExtractKty(jwk joseJwk.Key) (*joseJwa.KeyType, error) {
	if jwk == nil {
		return nil, fmt.Errorf("invalid jwk: %w", cryptoutilSharedApperr.ErrCantBeNil)
	}

	var err error

	var kty joseJwa.KeyType
	if err = jwk.Get(joseJwk.KeyTypeKey, &kty); err != nil {
		return nil, fmt.Errorf("failed to get kty header: %w", err)
	}

	return &kty, nil
}

// GenerateJWKForAlg generates a JWK for the specified algorithm.
func GenerateJWKForAlg(alg *cryptoutilOpenapiModel.GenerateAlgorithm) (*googleUuid.UUID, joseJwk.Key, joseJwk.Key, []byte, []byte, error) {
	kid, err := googleUuid.NewV7()
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to create uuid v7: %w", err)
	}

	key, err := validateJWKHeaders2(&kid, alg, nil, true) // true => generates enc key of the correct length
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("invalid JWK headers: %w", err)
	}

	return CreateJWKFromKey(&kid, alg, key)
}

// CreateJWKFromKey creates a JWK from an existing cryptographic key.
func CreateJWKFromKey(kid *googleUuid.UUID, alg *cryptoutilOpenapiModel.GenerateAlgorithm, key cryptoutilKeyGen.Key) (*googleUuid.UUID, joseJwk.Key, joseJwk.Key, []byte, []byte, error) {
	now := time.Now().UTC().Unix()

	_, err := validateJWKHeaders2(kid, alg, key, false)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("invalid JWK headers: %w", err)
	}

	var nonPublicJWK joseJwk.Key

	switch typedKey := key.(type) {
	case cryptoutilKeyGen.SecretKey: // HMAC // pragma: allowlist secret
		if nonPublicJWK, err = joseJwk.Import([]byte(typedKey)); err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to import key material into JWK: %w", err)
		}

		if err = nonPublicJWK.Set(joseJwk.KeyTypeKey, KtyOCT); err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'kty' header to 'oct' in JWK: %w", err)
		}

		// Set algorithm and use based on the generate algorithm
		switch *alg {
		case cryptoutilOpenapiModel.Oct256:
			if err = nonPublicJWK.Set(joseJwk.AlgorithmKey, AlgHS256); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'alg' header to 'HS256' in JWK: %w", err)
			}

			if err = nonPublicJWK.Set(joseJwk.KeyUsageKey, joseJwk.ForSignature); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'use' header to 'sig' in JWK: %w", err)
			}

			if err = nonPublicJWK.Set(joseJwk.KeyOpsKey, OpsSigVer); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'key_ops' header in JWK: %w", err)
			}
		case cryptoutilOpenapiModel.Oct384:
			if err = nonPublicJWK.Set(joseJwk.AlgorithmKey, AlgHS384); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'alg' header to 'HS384' in JWK: %w", err)
			}

			if err = nonPublicJWK.Set(joseJwk.KeyUsageKey, joseJwk.ForSignature); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'use' header to 'sig' in JWK: %w", err)
			}

			if err = nonPublicJWK.Set(joseJwk.KeyOpsKey, OpsSigVer); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'key_ops' header in JWK: %w", err)
			}
		case cryptoutilOpenapiModel.Oct512:
			if err = nonPublicJWK.Set(joseJwk.AlgorithmKey, AlgHS512); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'alg' header to 'HS512' in JWK: %w", err)
			}

			if err = nonPublicJWK.Set(joseJwk.KeyUsageKey, joseJwk.ForSignature); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'use' header to 'sig' in JWK: %w", err)
			}

			if err = nonPublicJWK.Set(joseJwk.KeyOpsKey, OpsSigVer); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'key_ops' header in JWK: %w", err)
			}
		case cryptoutilOpenapiModel.Oct128, cryptoutilOpenapiModel.Oct192:
			// AES keys, set encryption algorithm
			switch *alg {
			case cryptoutilOpenapiModel.Oct128:
				if err = nonPublicJWK.Set(joseJwk.AlgorithmKey, "A128GCM"); err != nil {
					return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'alg' header to 'A128GCM' in JWK: %w", err)
				}
			case cryptoutilOpenapiModel.Oct192:
				if err = nonPublicJWK.Set(joseJwk.AlgorithmKey, "A192GCM"); err != nil {
					return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'alg' header to 'A192GCM' in JWK: %w", err)
				}
			default:
				return nil, nil, nil, nil, nil, fmt.Errorf("unexpected algorithm %s for secret key in AES switch", *alg)
			}

			if err = nonPublicJWK.Set(joseJwk.KeyUsageKey, "enc"); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'use' header to 'enc' in JWK: %w", err)
			}
		default:
			return nil, nil, nil, nil, nil, fmt.Errorf("unexpected algorithm %s for secret key", *alg)
		}
	case *cryptoutilKeyGen.KeyPair: // RSA, ECDSA, EdDSA
		if nonPublicJWK, err = joseJwk.Import(typedKey.Private); err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to import key pair into JWK: %w", err)
		}

		switch typedKey.Private.(type) {
		case *rsa.PrivateKey: // RSA
			if err = nonPublicJWK.Set(joseJwk.KeyTypeKey, KtyRSA); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'kty' header to 'rsa' in JWK: %w", err)
			}
		case *ecdsa.PrivateKey: // ECDSA, ECDH
			if err = nonPublicJWK.Set(joseJwk.KeyTypeKey, KtyEC); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'kty' header to 'ec' in JWK: %w", err)
			}
		case ed25519.PrivateKey, ed448.PrivateKey: // ED25519, ED448
			if err = nonPublicJWK.Set(joseJwk.KeyTypeKey, KtyOKP); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'kty' header to 'okp' in JWK: %w", err)
			}
		default:
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'kty' header in JWK: unexpected key type %T", key)
		}
	default:
		return nil, nil, nil, nil, nil, fmt.Errorf("unsupported key type %T for JWK", key)
	}

	if err = nonPublicJWK.Set(joseJwk.KeyIDKey, kid.String()); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to set `kid` header in JWK: %w", err)
	}

	if err = nonPublicJWK.Set("iat", now); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to set `iat` header in JWK: %w", err)
	}

	clearNonPublicJWKBytes, err := json.Marshal(nonPublicJWK)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to serialize JWK: %w", err)
	}

	var publicJWK joseJwk.Key

	var clearPublicJWKBytes []byte

	if _, ok := key.(*cryptoutilKeyGen.KeyPair); ok { // RSA, EC, ED
		publicJWK, err = nonPublicJWK.PublicKey()
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to get public JWE JWK from private JWE JWK: %w", err)
		}

		clearPublicJWKBytes, err = json.Marshal(publicJWK)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to serialize public JWE JWK: %w", err)
		}
	}

	return kid, nonPublicJWK, publicJWK, clearNonPublicJWKBytes, clearPublicJWKBytes, nil
}

func validateJWKHeaders2(kid *googleUuid.UUID, alg *cryptoutilOpenapiModel.GenerateAlgorithm, key cryptoutilKeyGen.Key, isNilRawKeyOk bool) (cryptoutilKeyGen.Key, error) {
	if err := cryptoutilSharedUtilRandom.ValidateUUID(kid, &ErrInvalidJWKKidUUID); err != nil {
		return nil, fmt.Errorf("JWK kid must be valid: %w", err)
	} else if alg == nil {
		return nil, fmt.Errorf("JWK alg must be non-nil")
	} else if !isNilRawKeyOk && key == nil {
		return nil, fmt.Errorf("JWK key material must be non-nil")
	}

	switch *alg {
	case cryptoutilOpenapiModel.RSA4096:
		return validateOrGenerateRSAJWK(key, cryptoutilMagic.RSAKeySize4096)
	case cryptoutilOpenapiModel.RSA3072:
		return validateOrGenerateRSAJWK(key, cryptoutilMagic.RSAKeySize3072)
	case cryptoutilOpenapiModel.RSA2048:
		return validateOrGenerateRSAJWK(key, cryptoutilMagic.RSAKeySize2048)
	case cryptoutilOpenapiModel.ECP521:
		return validateOrGenerateEcdsaJWK(key, elliptic.P521())
	case cryptoutilOpenapiModel.ECP384:
		return validateOrGenerateEcdsaJWK(key, elliptic.P384())
	case cryptoutilOpenapiModel.ECP256:
		return validateOrGenerateEcdsaJWK(key, elliptic.P256())
	case cryptoutilOpenapiModel.OKPEd25519:
		return validateOrGenerateEddsaJWK(key, "Ed25519")
	case cryptoutilOpenapiModel.Oct512:
		return validateOrGenerateHMACJWK(key, cryptoutilMagic.HMACKeySize512)
	case cryptoutilOpenapiModel.Oct384:
		return validateOrGenerateHMACJWK(key, cryptoutilMagic.HMACKeySize384)
	case cryptoutilOpenapiModel.Oct256:
		return validateOrGenerateAESJWK(key, cryptoutilMagic.AESKeySize256)
	case cryptoutilOpenapiModel.Oct192:
		return validateOrGenerateAESJWK(key, cryptoutilMagic.AESKeySize192)
	case cryptoutilOpenapiModel.Oct128:
		return validateOrGenerateAESJWK(key, cryptoutilMagic.AESKeySize128)
	default:
		return nil, fmt.Errorf("unsupported JWK alg: %v", alg)
	}
}

func validateOrGenerateRSAJWK(key cryptoutilKeyGen.Key, keyBitsLength int) (*cryptoutilKeyGen.KeyPair, error) {
	if key == nil {
		generatedKey, err := cryptoutilKeyGen.GenerateRSAKeyPair(keyBitsLength)
		if err != nil {
			return nil, fmt.Errorf("failed to generate RSA %d key: %w", keyBitsLength, err)
		}

		return generatedKey, nil
	}

	keyPair, ok := key.(*cryptoutilKeyGen.KeyPair)
	if !ok {
		return nil, fmt.Errorf("unsupported key type %T; use *cryptoutilKeyGen.KeyPair", key)
	}

	rsaPrivateKey, ok := keyPair.Private.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("invalid key type %T; use *rsa.PrivateKey", keyPair.Private)
	}

	if rsaPrivateKey == nil { // pragma: allowlist secret
		return nil, fmt.Errorf("invalid nil RSA private key")
	}

	rsaPublicKey, ok := keyPair.Public.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("invalid key type %T; use *rsa.PublicKey", keyPair.Public)
	}

	if rsaPublicKey == nil {
		return nil, fmt.Errorf("invalid nil RSA public key")
	}

	return keyPair, nil
}

func validateOrGenerateEcdsaJWK(key cryptoutilKeyGen.Key, curve elliptic.Curve) (*cryptoutilKeyGen.KeyPair, error) {
	if key == nil {
		generatedKey, err := cryptoutilKeyGen.GenerateECDSAKeyPair(curve)
		if err != nil {
			return nil, fmt.Errorf("failed to generate ECDSA %s key pair: %w", curve, err)
		}

		return generatedKey, nil
	}

	keyPair, ok := key.(*cryptoutilKeyGen.KeyPair)
	if !ok {
		return nil, fmt.Errorf("unsupported key type %T; use *cryptoutilKeyGen.KeyPair", key)
	}

	rsaPrivateKey, ok := keyPair.Private.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("invalid key type %T; use *ecdsa.PrivateKey", keyPair.Private)
	}

	if rsaPrivateKey == nil { // pragma: allowlist secret
		return nil, fmt.Errorf("invalid nil ECDSA private key")
	}

	rsaPublicKey, ok := keyPair.Public.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("invalid key type %T; use *ecdsa.PublicKey", keyPair.Public)
	}

	if rsaPublicKey == nil {
		return nil, fmt.Errorf("invalid nil ECDSA public key")
	}

	return keyPair, nil
}

func validateOrGenerateEddsaJWK(key cryptoutilKeyGen.Key, curve string) (*cryptoutilKeyGen.KeyPair, error) {
	if key == nil {
		generatedKey, err := cryptoutilKeyGen.GenerateEDDSAKeyPair(curve)
		if err != nil {
			return nil, fmt.Errorf("failed to generate Ed29919 key pair: %w", err)
		}

		return generatedKey, nil
	}

	keyPair, ok := key.(*cryptoutilKeyGen.KeyPair)
	if !ok {
		return nil, fmt.Errorf("unsupported key type %T; use *cryptoutilKeyGen.KeyPair", key)
	}

	rsaPrivateKey, ok := keyPair.Private.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("invalid key type %T; use ed25519.PrivateKey", keyPair.Private)
	}

	if rsaPrivateKey == nil { // pragma: allowlist secret
		return nil, fmt.Errorf("invalid nil Ed29919 private key")
	}

	rsaPublicKey, ok := keyPair.Public.(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("invalid key type %T; use ed25519.PublicKey", keyPair.Public)
	}

	if rsaPublicKey == nil {
		return nil, fmt.Errorf("invalid nil Ed29919 public key")
	}

	return keyPair, nil
}

func validateOrGenerateHMACJWK(key cryptoutilKeyGen.Key, keyBitsLength int) (cryptoutilKeyGen.SecretKey, error) { // pragma: allowlist secret
	if key == nil {
		generatedKey, err := cryptoutilKeyGen.GenerateHMACKey(keyBitsLength)
		if err != nil {
			return nil, fmt.Errorf("failed to generate HMAC %d key: %w", keyBitsLength, err)
		}

		return generatedKey, nil
	}

	hmacKey, ok := key.(cryptoutilKeyGen.SecretKey) // pragma: allowlist secret
	if !ok {
		return nil, fmt.Errorf("invalid key type %T; use cryptoKeygen.SecretKey", key) // pragma: allowlist secret
	}

	if hmacKey == nil {
		return nil, fmt.Errorf("invalid nil key bytes")
	}

	if len(hmacKey) != keyBitsLength/cryptoutilMagic.BitsToBytes {
		return nil, fmt.Errorf("invalid key length %d; use HMAC %d", len(hmacKey), keyBitsLength)
	}

	return hmacKey, nil
}

func validateOrGenerateAESJWK(key cryptoutilKeyGen.Key, keyBitsLength int) (cryptoutilKeyGen.SecretKey, error) { // pragma: allowlist secret
	if key == nil {
		generatedKey, err := cryptoutilKeyGen.GenerateAESKey(keyBitsLength)
		if err != nil {
			return nil, fmt.Errorf("failed to generate AES %d key: %w", keyBitsLength, err)
		}

		return generatedKey, nil
	}

	aesKey, ok := key.(cryptoutilKeyGen.SecretKey) // pragma: allowlist secret
	if !ok {
		return nil, fmt.Errorf("invalid key type %T; use cryptoKeygen.SecretKey", key) // pragma: allowlist secret
	}

	if aesKey == nil {
		return nil, fmt.Errorf("invalid nil key bytes")
	}

	if len(aesKey) != keyBitsLength/cryptoutilMagic.BitsToBytes {
		return nil, fmt.Errorf("invalid key length %d; use AES %d", len(aesKey), keyBitsLength)
	}

	return aesKey, nil
}

// IsPublicJWK returns true if the JWK is a public key.
func IsPublicJWK(jwk joseJwk.Key) (bool, error) {
	if jwk == nil {
		return false, fmt.Errorf("invalid jwk: %w", cryptoutilSharedApperr.ErrCantBeNil)
	}

	switch jwk.(type) {
	case joseJwk.RSAPrivateKey, joseJwk.ECDSAPrivateKey, joseJwk.OKPPrivateKey, joseJwk.SymmetricKey:
		return false, nil
	case joseJwk.RSAPublicKey, joseJwk.ECDSAPublicKey, joseJwk.OKPPublicKey:
		return true, nil
	default:
		return false, fmt.Errorf("unsupported JWK type %T", jwk)
	}
}

// IsPrivateJWK returns true if the JWK is a private key.
func IsPrivateJWK(jwk joseJwk.Key) (bool, error) {
	if jwk == nil {
		return false, fmt.Errorf("invalid jwk: %w", cryptoutilSharedApperr.ErrCantBeNil)
	}

	switch jwk.(type) {
	case joseJwk.RSAPrivateKey, joseJwk.ECDSAPrivateKey, joseJwk.OKPPrivateKey:
		return true, nil
	case joseJwk.RSAPublicKey, joseJwk.ECDSAPublicKey, joseJwk.OKPPublicKey, joseJwk.SymmetricKey:
		return false, nil
	default:
		return false, fmt.Errorf("unsupported JWK type %T", jwk)
	}
}

// IsAsymmetricJWK returns true if the JWK is an asymmetric key.
func IsAsymmetricJWK(jwk joseJwk.Key) (bool, error) {
	if jwk == nil {
		return false, fmt.Errorf("invalid jwk: %w", cryptoutilSharedApperr.ErrCantBeNil)
	}

	switch jwk.(type) {
	case joseJwk.RSAPrivateKey, joseJwk.RSAPublicKey, joseJwk.ECDSAPrivateKey, joseJwk.ECDSAPublicKey, joseJwk.OKPPrivateKey, joseJwk.OKPPublicKey:
		return true, nil
	case joseJwk.SymmetricKey:
		return false, nil
	default:
		return false, fmt.Errorf("unsupported JWK type %T", jwk)
	}
}

// IsSymmetricJWK returns true if the JWK is a symmetric key.
func IsSymmetricJWK(jwk joseJwk.Key) (bool, error) {
	if jwk == nil {
		return false, fmt.Errorf("invalid jwk: %w", cryptoutilSharedApperr.ErrCantBeNil)
	}

	switch jwk.(type) {
	case joseJwk.RSAPrivateKey, joseJwk.RSAPublicKey, joseJwk.ECDSAPrivateKey, joseJwk.ECDSAPublicKey, joseJwk.OKPPrivateKey, joseJwk.OKPPublicKey:
		return false, nil
	case joseJwk.SymmetricKey:
		return true, nil
	default:
		return false, fmt.Errorf("unsupported JWK type %T", jwk)
	}
}

// IsEncryptJWK returns true if the JWK can be used for encryption.
func IsEncryptJWK(jwk joseJwk.Key) (bool, error) {
	if jwk == nil {
		return false, fmt.Errorf("JWK invalid: %w", cryptoutilSharedApperr.ErrCantBeNil)
	}

	_, _, err := ExtractAlgEncFromJWEJWK(jwk, 0)
	if err != nil {
		// Missing enc/alg headers means not an encryption key
		return false, nil //nolint:nilerr // Missing headers = incompatible key type, not an error
	}

	// At this point, JWK is confirmed to have an enc header
	switch jwk.(type) {
	case joseJwk.ECDSAPrivateKey:
		return false, nil // private key can be used for encrypt, but shouldn't be used in practice
	case joseJwk.RSAPrivateKey:
		return false, nil // private key can be used for encrypt, but shouldn't be used in practice
	case joseJwk.ECDSAPublicKey:
		return true, nil // jwx uses ECDSAPrivateKey for both ECDSA or ECDH, but encrypt alg header narrows it down to ECDH
	case joseJwk.RSAPublicKey:
		return true, nil
	case joseJwk.SymmetricKey:
		return true, nil // jwx SymmetricKey can be AES and HMAC, but enc header narrows it down to AES only
	case joseJwk.OKPPrivateKey, joseJwk.OKPPublicKey:
		return false, nil // Ed25519/Ed448 are signing only
	default:
		return false, fmt.Errorf("unsupported JWK type %T", jwk)
	}
}

// IsDecryptJWK returns true if the JWK can be used for decryption.
func IsDecryptJWK(jwk joseJwk.Key) (bool, error) {
	if jwk == nil {
		return false, fmt.Errorf("JWK invalid: %w", cryptoutilSharedApperr.ErrCantBeNil)
	}

	_, _, err := ExtractAlgEncFromJWEJWK(jwk, 0)
	if err != nil {
		// Missing enc/alg headers means not a decryption key
		return false, nil //nolint:nilerr // Missing headers = incompatible key type, not an error
	}

	// At this point, JWK is confirmed to have an enc header
	switch jwk.(type) {
	case joseJwk.ECDSAPrivateKey:
		return true, nil // jwx uses ECDSAPrivateKey for both ECDSA or ECDH, but encrypt alg header narrows it down to ECDH
	case joseJwk.RSAPrivateKey:
		return true, nil
	case joseJwk.ECDSAPublicKey:
		return false, nil
	case joseJwk.RSAPublicKey:
		return false, nil
	case joseJwk.SymmetricKey:
		return true, nil // jwx SymmetricKey can be AES and HMAC, but enc header narrows it down to AES only
	case joseJwk.OKPPrivateKey, joseJwk.OKPPublicKey:
		return false, nil // Ed25519/Ed448 are signing only
	default:
		return false, fmt.Errorf("unsupported JWK type %T", jwk)
	}
}

// IsSignJWK returns true if the JWK can be used for signing.
func IsSignJWK(jwk joseJwk.Key) (bool, error) {
	if jwk == nil {
		return false, fmt.Errorf("JWK invalid: %w", cryptoutilSharedApperr.ErrCantBeNil)
	}

	_, err := ExtractAlgFromJWSJWK(jwk, 0)
	if err != nil {
		// Missing alg header means not a signing key
		return false, nil //nolint:nilerr // Missing headers = incompatible key type, not an error
	}

	switch jwk.(type) {
	case joseJwk.ECDSAPrivateKey:
		return true, nil // jwx uses ECDSAPrivateKey for both ECDSA or ECDH, but signature alg header narrows it down to ECDSA
	case joseJwk.RSAPrivateKey:
		return true, nil
	case joseJwk.OKPPrivateKey:
		return true, nil
	case joseJwk.ECDSAPublicKey:
		return false, nil
	case joseJwk.RSAPublicKey:
		return false, nil
	case joseJwk.OKPPublicKey:
		return false, nil
	case joseJwk.SymmetricKey:
		return true, nil // jwx SymmetricKey can be AES and HMAC, but enc header narrows it down to HMAC only
	default:
		return false, fmt.Errorf("unsupported JWK type %T", jwk)
	}
}

// IsVerifyJWK returns true if the JWK can be used for signature verification.
func IsVerifyJWK(jwk joseJwk.Key) (bool, error) {
	if jwk == nil {
		return false, fmt.Errorf("JWK invalid: %w", cryptoutilSharedApperr.ErrCantBeNil)
	}

	_, err := ExtractAlgFromJWSJWK(jwk, 0)
	if err != nil {
		// Missing alg header means not a verification key
		return false, nil //nolint:nilerr // Missing headers = incompatible key type, not an error
	}

	switch jwk.(type) {
	case joseJwk.ECDSAPrivateKey:
		return false, nil // jwx uses ECDSAPrivateKey for both ECDSA or ECDH, but signature alg header narrows it down to ECDSA
	case joseJwk.RSAPrivateKey:
		return false, nil
	case joseJwk.OKPPrivateKey:
		return false, nil
	case joseJwk.RSAPublicKey:
		return true, nil
	case joseJwk.ECDSAPublicKey:
		return true, nil
	case joseJwk.OKPPublicKey:
		return true, nil
	case joseJwk.SymmetricKey:
		return true, nil // jwx SymmetricKey can be AES and HMAC, but enc header narrows it down to HMAC only
	default:
		return false, fmt.Errorf("unsupported JWK type %T", jwk)
	}
}

// EnsureSignatureAlgorithmType ensures that the JWK's algorithm field is properly typed as joseJwa.SignatureAlgorithm
// instead of a string. This is necessary after JSON parsing which converts types to strings.
func EnsureSignatureAlgorithmType(jwk joseJwk.Key) error {
	if jwk == nil {
		return fmt.Errorf("JWK invalid: %w", cryptoutilSharedApperr.ErrCantBeNil)
	}

	// Get the algorithm as a string first
	var algString string

	err := jwk.Get(joseJwk.AlgorithmKey, &algString)
	if err != nil {
		return fmt.Errorf("failed to get algorithm from JWK: %w", err)
	}

	// Convert string to proper SignatureAlgorithm type
	var alg joseJwa.SignatureAlgorithm

	switch algString {
	case algStrHS256:
		alg = AlgHS256
	case algStrHS384:
		alg = AlgHS384
	case algStrHS512:
		alg = AlgHS512
	case algStrRS256:
		alg = AlgRS256
	case algStrRS384:
		alg = AlgRS384
	case algStrRS512:
		alg = AlgRS512
	case algStrPS256:
		alg = AlgPS256
	case algStrPS384:
		alg = AlgPS384
	case algStrPS512:
		alg = AlgPS512
	case algStrES256:
		alg = AlgES256
	case algStrES384:
		alg = AlgES384
	case algStrES512:
		alg = AlgES512
	case algStrEdDSA:
		alg = AlgEdDSA
	default:
		return fmt.Errorf("unsupported signature algorithm: %s", algString)
	}

	// Set the properly typed algorithm back on the JWK
	err = jwk.Set(joseJwk.AlgorithmKey, alg)
	if err != nil {
		return fmt.Errorf("failed to set typed algorithm on JWK: %w", err)
	}

	return nil
}
