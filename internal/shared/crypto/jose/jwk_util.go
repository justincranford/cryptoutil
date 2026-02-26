// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	rsa "crypto/rsa"
	"fmt"
	"time"

	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
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
	algStrHS256 = cryptoutilSharedMagic.JoseAlgHS256
	algStrHS384 = cryptoutilSharedMagic.JoseAlgHS384
	algStrHS512 = cryptoutilSharedMagic.JoseAlgHS512
	algStrRS256 = cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm
	algStrRS384 = cryptoutilSharedMagic.JoseAlgRS384
	algStrRS512 = cryptoutilSharedMagic.JoseAlgRS512
	algStrPS256 = cryptoutilSharedMagic.JoseAlgPS256
	algStrPS384 = cryptoutilSharedMagic.JoseAlgPS384
	algStrPS512 = cryptoutilSharedMagic.JoseAlgPS512
	algStrES256 = cryptoutilSharedMagic.JoseAlgES256
	algStrES384 = cryptoutilSharedMagic.JoseAlgES384
	algStrES512 = cryptoutilSharedMagic.JoseAlgES512
	algStrEdDSA = cryptoutilSharedMagic.JoseAlgEdDSA

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

	if err = cryptoutilSharedUtilRandom.ValidateUUID(&kidUUID, ErrInvalidJWKKidUUID); err != nil {
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
	kid := googleUuid.Must(googleUuid.NewV7())

	key, err := validateJWKHeaders2(&kid, alg, nil, true) // true => generates enc key of the correct length
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("invalid JWK headers: %w", err)
	}

	return CreateJWKFromKey(&kid, alg, key)
}

// CreateJWKFromKey creates a JWK from an existing cryptographic key.
func CreateJWKFromKey(kid *googleUuid.UUID, alg *cryptoutilOpenapiModel.GenerateAlgorithm, key cryptoutilSharedCryptoKeygen.Key) (*googleUuid.UUID, joseJwk.Key, joseJwk.Key, []byte, []byte, error) {
	now := time.Now().UTC().Unix()

	_, err := validateJWKHeaders2(kid, alg, key, false)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("invalid JWK headers: %w", err)
	}

	var nonPublicJWK joseJwk.Key

	switch typedKey := key.(type) {
	case cryptoutilSharedCryptoKeygen.SecretKey: // HMAC // pragma: allowlist secret
		if nonPublicJWK, err = jwkImport([]byte(typedKey)); err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to import key material into JWK: %w", err)
		}

		if err = jwkKeySet(nonPublicJWK, joseJwk.KeyTypeKey, KtyOCT); err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'kty' header to 'oct' in JWK: %w", err)
		}

		// Set algorithm and use based on the generate algorithm
		switch *alg {
		case cryptoutilOpenapiModel.Oct256:
			if err = jwkKeySet(nonPublicJWK, joseJwk.AlgorithmKey, AlgHS256); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'alg' header to 'HS256' in JWK: %w", err)
			}

			if err = jwkKeySet(nonPublicJWK, joseJwk.KeyUsageKey, joseJwk.ForSignature); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'use' header to 'sig' in JWK: %w", err)
			}

			if err = jwkKeySet(nonPublicJWK, joseJwk.KeyOpsKey, OpsSigVer); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'key_ops' header in JWK: %w", err)
			}
		case cryptoutilOpenapiModel.Oct384:
			if err = jwkKeySet(nonPublicJWK, joseJwk.AlgorithmKey, AlgHS384); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'alg' header to 'HS384' in JWK: %w", err)
			}

			if err = jwkKeySet(nonPublicJWK, joseJwk.KeyUsageKey, joseJwk.ForSignature); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'use' header to 'sig' in JWK: %w", err)
			}

			if err = jwkKeySet(nonPublicJWK, joseJwk.KeyOpsKey, OpsSigVer); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'key_ops' header in JWK: %w", err)
			}
		case cryptoutilOpenapiModel.Oct512:
			if err = jwkKeySet(nonPublicJWK, joseJwk.AlgorithmKey, AlgHS512); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'alg' header to 'HS512' in JWK: %w", err)
			}

			if err = jwkKeySet(nonPublicJWK, joseJwk.KeyUsageKey, joseJwk.ForSignature); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'use' header to 'sig' in JWK: %w", err)
			}

			if err = jwkKeySet(nonPublicJWK, joseJwk.KeyOpsKey, OpsSigVer); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'key_ops' header in JWK: %w", err)
			}
		case cryptoutilOpenapiModel.Oct128, cryptoutilOpenapiModel.Oct192:
			// AES keys, set encryption algorithm
			switch *alg {
			case cryptoutilOpenapiModel.Oct128:
				if err = jwkKeySet(nonPublicJWK, joseJwk.AlgorithmKey, cryptoutilSharedMagic.JoseEncA128GCM); err != nil {
					return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'alg' header to 'A128GCM' in JWK: %w", err)
				}
			case cryptoutilOpenapiModel.Oct192:
				if err = jwkKeySet(nonPublicJWK, joseJwk.AlgorithmKey, cryptoutilSharedMagic.JoseEncA192GCM); err != nil {
					return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'alg' header to 'A192GCM' in JWK: %w", err)
				}
			default:
				return nil, nil, nil, nil, nil, fmt.Errorf("unexpected algorithm %s for secret key in AES switch", *alg)
			}

			if err = jwkKeySet(nonPublicJWK, joseJwk.KeyUsageKey, cryptoutilSharedMagic.JoseKeyUseEnc); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'use' header to 'enc' in JWK: %w", err)
			}
		default:
			return nil, nil, nil, nil, nil, fmt.Errorf("unexpected algorithm %s for secret key", *alg)
		}
	case *cryptoutilSharedCryptoKeygen.KeyPair: // RSA, ECDSA, EdDSA
		if nonPublicJWK, err = jwkImport(typedKey.Private); err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to import key pair into JWK: %w", err)
		}

		switch typedKey.Private.(type) {
		case *rsa.PrivateKey: // RSA
			if err = jwkKeySet(nonPublicJWK, joseJwk.KeyTypeKey, KtyRSA); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'kty' header to 'rsa' in JWK: %w", err)
			}
		case *ecdsa.PrivateKey: // ECDSA, ECDH
			if err = jwkKeySet(nonPublicJWK, joseJwk.KeyTypeKey, KtyEC); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'kty' header to 'ec' in JWK: %w", err)
			}
		case ed25519.PrivateKey, ed448.PrivateKey: // ED25519, ED448
			if err = jwkKeySet(nonPublicJWK, joseJwk.KeyTypeKey, KtyOKP); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'kty' header to 'okp' in JWK: %w", err)
			}
		default:
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'kty' header in JWK: unexpected key type %T", key)
		}
	default:
		return nil, nil, nil, nil, nil, fmt.Errorf("unsupported key type %T for JWK", key)
	}

	if err = jwkKeySet(nonPublicJWK, joseJwk.KeyIDKey, kid.String()); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to set `kid` header in JWK: %w", err)
	}

	if err = jwkKeySet(nonPublicJWK, cryptoutilSharedMagic.ClaimIat, now); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to set `iat` header in JWK: %w", err)
	}

	clearNonPublicJWKBytes, err := jsonMarshalFunc(nonPublicJWK)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to serialize JWK: %w", err)
	}

	var publicJWK joseJwk.Key

	var clearPublicJWKBytes []byte

	if _, ok := key.(*cryptoutilSharedCryptoKeygen.KeyPair); ok { // RSA, EC, ED
		publicJWK, err = jwkPublicKey(nonPublicJWK)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to get public JWE JWK from private JWE JWK: %w", err)
		}

		clearPublicJWKBytes, err = jsonMarshalFunc(publicJWK)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to serialize public JWE JWK: %w", err)
		}
	}

	return kid, nonPublicJWK, publicJWK, clearNonPublicJWKBytes, clearPublicJWKBytes, nil
}

func validateJWKHeaders2(kid *googleUuid.UUID, alg *cryptoutilOpenapiModel.GenerateAlgorithm, key cryptoutilSharedCryptoKeygen.Key, isNilRawKeyOk bool) (cryptoutilSharedCryptoKeygen.Key, error) {
	if err := cryptoutilSharedUtilRandom.ValidateUUID(kid, ErrInvalidJWKKidUUID); err != nil {
		return nil, fmt.Errorf("JWK kid must be valid: %w", err)
	} else if alg == nil {
		return nil, fmt.Errorf("JWK alg must be non-nil")
	} else if !isNilRawKeyOk && key == nil {
		return nil, fmt.Errorf("JWK key material must be non-nil")
	}

	switch *alg {
	case cryptoutilOpenapiModel.RSA4096:
		return validateOrGenerateRSAJWK(key, cryptoutilSharedMagic.RSAKeySize4096)
	case cryptoutilOpenapiModel.RSA3072:
		return validateOrGenerateRSAJWK(key, cryptoutilSharedMagic.RSAKeySize3072)
	case cryptoutilOpenapiModel.RSA2048:
		return validateOrGenerateRSAJWK(key, cryptoutilSharedMagic.RSAKeySize2048)
	case cryptoutilOpenapiModel.ECP521:
		return validateOrGenerateEcdsaJWK(key, elliptic.P521())
	case cryptoutilOpenapiModel.ECP384:
		return validateOrGenerateEcdsaJWK(key, elliptic.P384())
	case cryptoutilOpenapiModel.ECP256:
		return validateOrGenerateEcdsaJWK(key, elliptic.P256())
	case cryptoutilOpenapiModel.OKPEd25519:
		return validateOrGenerateEddsaJWK(key, cryptoutilSharedMagic.EdCurveEd25519)
	case cryptoutilOpenapiModel.Oct512:
		return validateOrGenerateHMACJWK(key, cryptoutilSharedMagic.HMACKeySize512)
	case cryptoutilOpenapiModel.Oct384:
		return validateOrGenerateHMACJWK(key, cryptoutilSharedMagic.HMACKeySize384)
	case cryptoutilOpenapiModel.Oct256:
		return validateOrGenerateAESJWK(key, cryptoutilSharedMagic.AESKeySize256)
	case cryptoutilOpenapiModel.Oct192:
		return validateOrGenerateAESJWK(key, cryptoutilSharedMagic.AESKeySize192)
	case cryptoutilOpenapiModel.Oct128:
		return validateOrGenerateAESJWK(key, cryptoutilSharedMagic.AESKeySize128)
	default:
		return nil, fmt.Errorf("unsupported JWK alg: %v", alg)
	}
}
