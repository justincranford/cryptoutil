package jose

import (
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudflare/circl/sign/ed448"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilKeygen "cryptoutil/internal/common/crypto/keygen"
	cryptoutilUtil "cryptoutil/internal/common/util"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

// type generateJweJwkResult struct {
// 	kid                       *googleUuid.UUID
// 	privateOrSecretJwk        joseJwk.Key
// 	publicKey                 joseJwk.Key
// 	clearPrivateOrSecretJwk []byte
// 	clearPublicKey          []byte
// }

// func GenerateJweJwkForEncAndAlgFunction(enc *joseJwa.ContentEncryptionAlgorithm, alg *joseJwa.KeyEncryptionAlgorithm) func() (*generateJweJwkResult, error) {
// 	return func() (*generateJweJwkResult, error) {
// 		kid1, err := googleUuid.NewV7()
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to create JWE JWK kid UUID: %w", err)
// 		}
// 		var key cryptoutilKeygen.Key
// 		switch *alg {
// 		case AlgDir:
// 			switch *enc {
// 			case EncA256GCM:
// 				key, err = cryptoutilKeygen.GenerateAESKey(256)
// 			case EncA192GCM:
// 				key, err = cryptoutilKeygen.GenerateAESKey(192)
// 			case EncA128GCM:
// 				key, err = cryptoutilKeygen.GenerateAESKey(128)
// 			case EncA256CBC_HS512:
// 				key, err = cryptoutilKeygen.GenerateAESHSKey(512)
// 			case EncA192CBC_HS384:
// 				key, err = cryptoutilKeygen.GenerateAESHSKey(384)
// 			case EncA128CBC_HS256:
// 				key, err = cryptoutilKeygen.GenerateAESHSKey(256)
// 			default:
// 				return nil, fmt.Errorf("unsupported JWE JWK enc %s", *enc)
// 			}

// 		case AlgA256KW, AlgA256GCMKW:
// 			key, err = cryptoutilKeygen.GenerateAESKey(256)
// 		case AlgA192KW, AlgA192GCMKW:
// 			key, err = cryptoutilKeygen.GenerateAESKey(192)
// 		case AlgA128KW, AlgA128GCMKW:
// 			key, err = cryptoutilKeygen.GenerateAESKey(128)

// 		case AlgRSAOAEP512:
// 			key, err = cryptoutilKeygen.GenerateRSAKeyPair(4096)
// 		case AlgRSAOAEP384:
// 			key, err = cryptoutilKeygen.GenerateRSAKeyPair(3072)
// 		case AlgRSAOAEP256:
// 			key, err = cryptoutilKeygen.GenerateRSAKeyPair(2048)
// 		case AlgRSAOAEP:
// 			key, err = cryptoutilKeygen.GenerateRSAKeyPair(2048)
// 		case AlgRSA15:
// 			key, err = cryptoutilKeygen.GenerateRSAKeyPair(2048)

// 		case AlgECDHES, AlgECDHESA256KW:
// 			key, err = cryptoutilKeygen.GenerateECDHKeyPair(ecdh.P521())
// 		case AlgECDHESA192KW:
// 			key, err = cryptoutilKeygen.GenerateECDHKeyPair(ecdh.P384())
// 		case AlgECDHESA128KW:
// 			key, err = cryptoutilKeygen.GenerateECDHKeyPair(ecdh.P256())

// 		default:
// 			return nil, fmt.Errorf("unsupported JWE JWK alg %s", *alg)
// 		}
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to generate private or secret key: %w", err)
// 		}
// 		kid, privateOrSecretJwk, publicKey, clearPrivateOrSecretJwk, clearPublicKey, err := CreateJweJwkFromKey(&kid1, enc, alg, key)
// 		generateJweJwkResult := generateJweJwkResult{
// 			kid:                       kid,
// 			privateOrSecretJwk:        privateOrSecretJwk,
// 			publicKey:                 publicKey,
// 			clearPrivateOrSecretJwk: clearPrivateOrSecretJwk,
// 			clearPublicKey:          clearPublicKey,
// 		}
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to create JWE JWK: %w", err)
// 		}
// 		return &generateJweJwkResult, nil
// 	}
// }

func GenerateJweJwkForEncAndAlg(enc *joseJwa.ContentEncryptionAlgorithm, alg *joseJwa.KeyEncryptionAlgorithm) (*googleUuid.UUID, joseJwk.Key, joseJwk.Key, []byte, []byte, error) {
	kid, err := googleUuid.NewV7()
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to create uuid v7: %w", err)
	}
	key, err := validateJweJwkHeaders(&kid, enc, alg, nil, true) // true => generates enc key of the correct length
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("invalid JWE JWK headers: %w", err)
	}
	return CreateJweJwkFromKey(&kid, enc, alg, key)
}

func CreateJweJwkFromKey(kid *googleUuid.UUID, enc *joseJwa.ContentEncryptionAlgorithm, alg *joseJwa.KeyEncryptionAlgorithm, key cryptoutilKeygen.Key) (*googleUuid.UUID, joseJwk.Key, joseJwk.Key, []byte, []byte, error) {
	now := time.Now().UTC().Unix()
	_, err := validateJweJwkHeaders(kid, enc, alg, key, false)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("invalid JWE JWK headers: %w", err)
	}
	var privateOrSecretJwk joseJwk.Key
	switch typedKey := key.(type) {
	case cryptoutilKeygen.SecretKey: // AES, AES-HS, HMAC
		if privateOrSecretJwk, err = joseJwk.Import([]byte(typedKey)); err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to import key material into JWE JWK: %w", err)
		}
		if err = privateOrSecretJwk.Set(joseJwk.KeyTypeKey, KtyOCT); err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'kty' header to 'oct' in JWE JWK: %w", err)
		}
	case *cryptoutilKeygen.KeyPair: // RSA, EC, ED
		if privateOrSecretJwk, err = joseJwk.Import(typedKey.Private); err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to import key pair into JWE JWK: %w", err)
		}
		switch typedKey.Private.(type) {
		case *rsa.PrivateKey: // RSA
			if err = privateOrSecretJwk.Set(joseJwk.KeyTypeKey, KtyRSA); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'kty' header to 'rsa' in JWE JWK: %w", err)
			}
		case *ecdsa.PrivateKey, *ecdh.PrivateKey: // ECDSA, ECDH
			if err = privateOrSecretJwk.Set(joseJwk.KeyTypeKey, KtyEC); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'kty' header to 'ec' in JWE JWK: %w", err)
			}
		case ed25519.PrivateKey, ed448.PrivateKey: // EdDSA
			if err = privateOrSecretJwk.Set(joseJwk.KeyTypeKey, KtyOKP); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'kty' header to 'okp' in JWE JWK: %w", err)
			}
		default:
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'kty' header in JWE JWK: unexpected key type %T", key)
		}
	default:
		return nil, nil, nil, nil, nil, fmt.Errorf("unsupported key type %T for JWE JWK", key)
	}

	if err = privateOrSecretJwk.Set(joseJwk.KeyIDKey, kid.String()); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to set `kid` header in JWE JWK: %w", err)
	}
	if err = privateOrSecretJwk.Set(joseJwk.AlgorithmKey, *alg); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to set `alg` header in JWE JWK: %w", err)
	}
	if err = privateOrSecretJwk.Set("enc", *enc); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to set `alg` header in JWE JWK: %w", err)
	}
	if err = privateOrSecretJwk.Set("iat", now); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to set `iat` header in JWE JWK: %w", err)
	}
	if err = privateOrSecretJwk.Set(joseJwk.KeyUsageKey, joseJwk.ForEncryption.String()); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to set `enc` header in JWE JWK: %w", err)
	}
	if err = privateOrSecretJwk.Set(joseJwk.KeyOpsKey, OpsEncDec); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to set `ops` header in JWE JWK: %w", err)
	}

	clearPrivateOrSecretJwk, err := json.Marshal(privateOrSecretJwk)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to serialize private or secret JWE JWK: %w", err)
	}

	var publicJwk joseJwk.Key
	var clearPublicJwk []byte
	if _, ok := key.(*cryptoutilKeygen.KeyPair); ok { // RSA, EC, ED
		publicJwk, err = privateOrSecretJwk.PublicKey()
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to get public JWE JWK from private JWE JWK: %w", err)
		}
		if err = publicJwk.Set(joseJwk.KeyOpsKey, OpsEnc); err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to set `ops` header in JWE JWK: %w", err)
		}
		clearPublicJwk, err = json.Marshal(publicJwk)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to serialize public JWE JWK: %w", err)
		}
	}

	return kid, privateOrSecretJwk, publicJwk, clearPrivateOrSecretJwk, clearPublicJwk, nil
}

func validateJweJwkHeaders(kid *googleUuid.UUID, enc *joseJwa.ContentEncryptionAlgorithm, alg *joseJwa.KeyEncryptionAlgorithm, key cryptoutilKeygen.Key, isNilRawKeyOk bool) (cryptoutilKeygen.Key, error) {
	if err := cryptoutilUtil.ValidateUUID(kid, "invalid JWE JWK kid"); err != nil {
		return nil, fmt.Errorf("JWE JWK kid must be valid: %w", err)
	} else if alg == nil {
		return nil, fmt.Errorf("JWE JWK alg must be non-nil")
	} else if enc == nil {
		return nil, fmt.Errorf("JWE JWK enc must be non-nil")
	} else if !isNilRawKeyOk && key == nil {
		return nil, fmt.Errorf("JWE JWK key must be non-nil")
	}
	encKeyBitsLength, err := EncToBitsLength(enc)
	if err != nil {
		return nil, fmt.Errorf("JWE JWK length error: %w", err)
	}

	switch *alg {
	case AlgDir:
		return validateOrGenerateJweAesJwk(key, enc, alg, encKeyBitsLength, &EncA256GCM, &EncA256CBC_HS512, &EncA192GCM, &EncA192CBC_HS384, &EncA128GCM, &EncA128CBC_HS256)

	case AlgA256KW, AlgA256GCMKW:
		return validateOrGenerateJweAesJwk(key, enc, alg, 256, &EncA256GCM, &EncA256CBC_HS512, &EncA192GCM, &EncA192CBC_HS384, &EncA128GCM, &EncA128CBC_HS256)
	case AlgA192KW, AlgA192GCMKW:
		return validateOrGenerateJweAesJwk(key, enc, alg, 192, &EncA256GCM, &EncA256CBC_HS512, &EncA192GCM, &EncA192CBC_HS384, &EncA128GCM, &EncA128CBC_HS256)
	case AlgA128KW, AlgA128GCMKW:
		return validateOrGenerateJweAesJwk(key, enc, alg, 128, &EncA256GCM, &EncA256CBC_HS512, &EncA192GCM, &EncA192CBC_HS384, &EncA128GCM, &EncA128CBC_HS256)

	case AlgRSAOAEP512:
		return validateOrGenerateJweRsaJwk(key, enc, alg, 4096, &EncA256GCM, &EncA256CBC_HS512, &EncA192GCM, &EncA192CBC_HS384, &EncA128GCM, &EncA128CBC_HS256)
	case AlgRSAOAEP384:
		return validateOrGenerateJweRsaJwk(key, enc, alg, 3072, &EncA256GCM, &EncA256CBC_HS512, &EncA192GCM, &EncA192CBC_HS384, &EncA128GCM, &EncA128CBC_HS256)
	case AlgRSAOAEP256:
		return validateOrGenerateJweRsaJwk(key, enc, alg, 2048, &EncA256GCM, &EncA256CBC_HS512, &EncA192GCM, &EncA192CBC_HS384, &EncA128GCM, &EncA128CBC_HS256)
	case AlgRSAOAEP:
		return validateOrGenerateJweRsaJwk(key, enc, alg, 2048, &EncA256GCM, &EncA256CBC_HS512, &EncA192GCM, &EncA192CBC_HS384, &EncA128GCM, &EncA128CBC_HS256)
	case AlgRSA15:
		return validateOrGenerateJweRsaJwk(key, enc, alg, 2048, &EncA256GCM, &EncA256CBC_HS512, &EncA192GCM, &EncA192CBC_HS384, &EncA128GCM, &EncA128CBC_HS256)

	case AlgECDHES, AlgECDHESA256KW:
		return validateOrGenerateJweEcdhJwk(key, enc, alg, ecdh.P521(), &EncA256GCM, &EncA256CBC_HS512, &EncA192GCM, &EncA192CBC_HS384, &EncA128GCM, &EncA128CBC_HS256)
	case AlgECDHESA192KW:
		return validateOrGenerateJweEcdhJwk(key, enc, alg, ecdh.P384(), &EncA256GCM, &EncA256CBC_HS512, &EncA192GCM, &EncA192CBC_HS384, &EncA128GCM, &EncA128CBC_HS256)
	case AlgECDHESA128KW:
		return validateOrGenerateJweEcdhJwk(key, enc, alg, ecdh.P256(), &EncA256GCM, &EncA256CBC_HS512, &EncA192GCM, &EncA192CBC_HS384, &EncA128GCM, &EncA128CBC_HS256)

	default:
		return nil, fmt.Errorf("unsupported JWE JWK alg %s", *alg)
	}
}

func validateOrGenerateJweAesJwk(key cryptoutilKeygen.Key, enc *joseJwa.ContentEncryptionAlgorithm, alg *joseJwa.KeyEncryptionAlgorithm, keyBitsLength int, allowedEncs ...*joseJwa.ContentEncryptionAlgorithm) (cryptoutilKeygen.SecretKey, error) {
	if !cryptoutilUtil.Contains(allowedEncs, enc) {
		return nil, fmt.Errorf("valid JWE JWK alg %s, but enc %s not allowed; use one of %v", *alg, *enc, allowedEncs)
	} else if key == nil {
		var keyBytes cryptoutilKeygen.SecretKey
		var err error
		switch *alg {
		case AlgA256KW, AlgA256GCMKW, AlgA192KW, AlgA192GCMKW, AlgA128KW, AlgA128GCMKW:
			keyBytes, err = cryptoutilKeygen.GenerateAESKey(keyBitsLength)

		case AlgDir:
			switch *enc {
			case EncA256GCM, EncA192GCM, EncA128GCM:
				keyBytes, err = cryptoutilKeygen.GenerateAESKey(keyBitsLength)
			case EncA256CBC_HS512, EncA192CBC_HS384, EncA128CBC_HS256:
				keyBytes, err = cryptoutilKeygen.GenerateAESHSKey(keyBitsLength)
			default:
				return nil, fmt.Errorf("valid JWE JWK alg %s, but invalid enc %s", *enc, *alg)
			}
		}
		if err != nil {
			return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but failed to generate AES %d key: %w", *enc, *alg, keyBitsLength, err)
		}
		return keyBytes, nil
	} else {
		keyBytes, ok := key.(cryptoutilKeygen.SecretKey)
		if !ok {
			return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but unsupported key type %T; use []byte", *enc, *alg, key)
		} else if keyBytes == nil {
			return nil, fmt.Errorf("valid enc %s and alg %s, but invalid nil key bytes", *enc, *alg)
		} else if len(keyBytes) != keyBitsLength/8 {
			return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but invalid key length %d; use AES %d", *enc, *alg, len(keyBytes), keyBitsLength)
		}
		return keyBytes, nil
	}
}

func validateOrGenerateJweRsaJwk(key cryptoutilKeygen.Key, enc *joseJwa.ContentEncryptionAlgorithm, alg *joseJwa.KeyEncryptionAlgorithm, keyBitsLength int, allowedEncs ...*joseJwa.ContentEncryptionAlgorithm) (*cryptoutilKeygen.KeyPair, error) {
	if !cryptoutilUtil.Contains(allowedEncs, enc) {
		return nil, fmt.Errorf("valid JWE JWK alg %s, but enc %s not allowed; use one of %v", *alg, *enc, allowedEncs)
	} else if key == nil {
		generatedKeyPair, err := cryptoutilKeygen.GenerateRSAKeyPair(keyBitsLength)
		if err != nil {
			return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but failed to generate RSA %d key: %w", *enc, *alg, keyBitsLength, err)
		}
		return generatedKeyPair, nil
	} else {
		keyPair, ok := key.(*cryptoutilKeygen.KeyPair)
		if !ok {
			return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but unsupported key type %T; use *cryptoutilKeygen.Key", *enc, *alg, key)
		}
		rsaPrivateKey, ok := keyPair.Private.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but unsupported key type %T; use *rsa.PrivateKey", *enc, *alg, keyPair.Private)
		} else if rsaPrivateKey == nil {
			return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but invalid nil RSA private key", *enc, *alg)
		}
		rsaPublicKey, ok := keyPair.Public.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but unsupported key type %T; use *rsa.PublicKey", *enc, *alg, keyPair.Public)
		} else if rsaPublicKey == nil {
			return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but invalid nil RSA public key", *enc, *alg)
		}
		return keyPair, nil
	}
}

func validateOrGenerateJweEcdhJwk(key cryptoutilKeygen.Key, enc *joseJwa.ContentEncryptionAlgorithm, alg *joseJwa.KeyEncryptionAlgorithm, ecdhCurve ecdh.Curve, allowedEncs ...*joseJwa.ContentEncryptionAlgorithm) (*cryptoutilKeygen.KeyPair, error) {
	if !cryptoutilUtil.Contains(allowedEncs, enc) {
		return nil, fmt.Errorf("valid JWE JWK alg %s, but enc %s not allowed; use one of %v", *alg, *enc, allowedEncs)
	} else if key == nil {
		generatedKeyPair, err := cryptoutilKeygen.GenerateECDHKeyPair(ecdhCurve)
		if err != nil {
			return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but failed to generate ECDH %s key: %w", *enc, *alg, ecdhCurve, err)
		}
		return generatedKeyPair, nil
	} else {
		keyPair, ok := key.(*cryptoutilKeygen.KeyPair)
		if !ok {
			return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but unsupported key type %T; use *cryptoutilKeygen.Key", *enc, *alg, key)
		}
		ecdhPrivateKey, ok := keyPair.Private.(*ecdh.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but unsupported key type %T; use *ecdh.PrivateKey", *enc, *alg, keyPair.Private)
		} else if ecdhPrivateKey == nil {
			return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but unsupported nil ECDH private key", *enc, *alg)
		}
		ecdhPublicKey, ok := keyPair.Public.(*ecdh.PublicKey)
		if !ok {
			return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but unsupported key type %T; use *ecdh.PublicKey", *enc, *alg, keyPair.Public)
		} else if ecdhPublicKey == nil {
			return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but unsupported nil ECDH public key", *enc, *alg)
		}
		return keyPair, nil
	}
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

func ExtractAlgEncFromJweJwk(jwk joseJwk.Key, i int) (*joseJwa.ContentEncryptionAlgorithm, *joseJwa.KeyEncryptionAlgorithm, error) {
	if jwk == nil {
		return nil, nil, fmt.Errorf("JWK %d invalid: %w", i, cryptoutilAppErr.ErrCantBeNil)
	}

	var enc joseJwa.ContentEncryptionAlgorithm
	err := jwk.Get("enc", &enc) // Example: A256GCM, A192GCM, A128GCM, A256CBC-HS512, A192CBC-HS384, A128CBC-HS256
	if err != nil {
		// Workaround: If JWK was serialized (for encryption) and parsed (after decryption), 'enc' header incorrect gets parsed as string, so try getting as string converting it to joseJwa.ContentEncryptionAlgorithm
		var encString string
		err = jwk.Get("enc", &encString)
		if err != nil {
			return nil, nil, fmt.Errorf("can't get JWK %d 'enc' attribute: %w", i, err)
		}
		enc = joseJwa.NewContentEncryptionAlgorithm(encString)
	}

	var alg joseJwa.KeyEncryptionAlgorithm
	err = jwk.Get(joseJwk.AlgorithmKey, &alg) // Example: A256KW, A192KW, A128KW, A256GCMKW, A192GCMKW, A128GCMKW, dir
	if err != nil {
		return nil, nil, fmt.Errorf("can't get JWK %d 'alg' attribute: %w", i, err)
	}

	return &enc, &alg, nil
}
