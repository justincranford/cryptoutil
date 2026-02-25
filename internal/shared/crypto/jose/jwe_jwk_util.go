// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	"crypto/ecdh"
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	rsa "crypto/rsa"
	"fmt"
	"time"

	"github.com/cloudflare/circl/sign/ed448"

	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtil "cryptoutil/internal/shared/util"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

// ErrInvalidJWEJWKKidUUID indicates the JWE JWK key ID is not a valid UUID.
var ErrInvalidJWEJWKKidUUID = "invalid JWE JWK kid UUID"

// GenerateJWEJWKForEncAndAlg generates a new JWE JWK for the specified encryption and algorithm.
func GenerateJWEJWKForEncAndAlg(enc *joseJwa.ContentEncryptionAlgorithm, alg *joseJwa.KeyEncryptionAlgorithm) (*googleUuid.UUID, joseJwk.Key, joseJwk.Key, []byte, []byte, error) {
	kid := googleUuid.Must(googleUuid.NewV7())

	key, err := validateJWEJWKHeaders(&kid, enc, alg, nil, true) // true => generates enc key of the correct length
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("invalid JWE JWK headers: %w", err)
	}

	return CreateJWEJWKFromKey(&kid, enc, alg, key)
}

// CreateJWEJWKFromKey creates a JWE JWK from an existing cryptographic key.
func CreateJWEJWKFromKey(kid *googleUuid.UUID, enc *joseJwa.ContentEncryptionAlgorithm, alg *joseJwa.KeyEncryptionAlgorithm, key cryptoutilSharedCryptoKeygen.Key) (*googleUuid.UUID, joseJwk.Key, joseJwk.Key, []byte, []byte, error) {
	now := time.Now().UTC().Unix()

	_, err := validateJWEJWKHeaders(kid, enc, alg, key, false)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("invalid JWE JWK headers: %w", err)
	}

	var nonPublicJWK joseJwk.Key

	switch typedKey := key.(type) {
	case cryptoutilSharedCryptoKeygen.SecretKey: // AES, AES-HS, HMAC
		if nonPublicJWK, err = jwkImport([]byte(typedKey)); err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to import key material into JWE JWK: %w", err)
		}

		if err = jwkKeySet(nonPublicJWK, joseJwk.KeyTypeKey, KtyOCT); err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'kty' header to 'oct' in JWE JWK: %w", err)
		}
	case *cryptoutilSharedCryptoKeygen.KeyPair: // RSA, EC, ED
		if nonPublicJWK, err = jwkImport(typedKey.Private); err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to import key pair into JWE JWK: %w", err)
		}

		switch typedKey.Private.(type) {
		case *rsa.PrivateKey: // RSA
			if err = jwkKeySet(nonPublicJWK, joseJwk.KeyTypeKey, KtyRSA); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'kty' header to 'rsa' in JWE JWK: %w", err)
			}
		case *ecdsa.PrivateKey, *ecdh.PrivateKey: // ECDSA, ECDH
			if err = jwkKeySet(nonPublicJWK, joseJwk.KeyTypeKey, KtyEC); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'kty' header to 'ec' in JWE JWK: %w", err)
			}
		case ed25519.PrivateKey, ed448.PrivateKey: // EdDSA
			if err = jwkKeySet(nonPublicJWK, joseJwk.KeyTypeKey, KtyOKP); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'kty' header to 'okp' in JWE JWK: %w", err)
			}
		default:
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to set 'kty' header in JWE JWK: unexpected key type %T", key)
		}
	default:
		return nil, nil, nil, nil, nil, fmt.Errorf("unsupported key type %T for JWE JWK", key)
	}

	if err = jwkKeySet(nonPublicJWK, joseJwk.KeyIDKey, kid.String()); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to set `kid` header in JWE JWK: %w", err)
	}

	if err = jwkKeySet(nonPublicJWK, joseJwk.AlgorithmKey, *alg); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to set `alg` header in JWE JWK: %w", err)
	}

	if err = jwkKeySet(nonPublicJWK, "enc", *enc); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to set `alg` header in JWE JWK: %w", err)
	}

	if err = jwkKeySet(nonPublicJWK, "iat", now); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to set `iat` header in JWE JWK: %w", err)
	}

	if err = jwkKeySet(nonPublicJWK, joseJwk.KeyUsageKey, joseJwk.ForEncryption.String()); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to set `enc` header in JWE JWK: %w", err)
	}

	if err = jwkKeySet(nonPublicJWK, joseJwk.KeyOpsKey, OpsEncDec); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to set `ops` header in JWE JWK: %w", err)
	}

	clearNonPublicJWKBytes, err := jsonMarshalFunc(nonPublicJWK)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to serialize private or secret JWE JWK: %w", err)
	}

	var publicJWK joseJwk.Key

	var clearPublicJWKBytes []byte

	if _, ok := key.(*cryptoutilSharedCryptoKeygen.KeyPair); ok { // RSA, EC, ED
		publicJWK, err = jwkPublicKey(nonPublicJWK)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to get public JWE JWK from private JWE JWK: %w", err)
		}

		if err = jwkKeySet(publicJWK, joseJwk.KeyOpsKey, OpsEnc); err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to set `ops` header in JWE JWK: %w", err)
		}

		clearPublicJWKBytes, err = jsonMarshalFunc(publicJWK)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to serialize public JWE JWK: %w", err)
		}
	}

	return kid, nonPublicJWK, publicJWK, clearNonPublicJWKBytes, clearPublicJWKBytes, nil
}

func validateJWEJWKHeaders(kid *googleUuid.UUID, enc *joseJwa.ContentEncryptionAlgorithm, alg *joseJwa.KeyEncryptionAlgorithm, key cryptoutilSharedCryptoKeygen.Key, isNilRawKeyOk bool) (cryptoutilSharedCryptoKeygen.Key, error) {
	if err := cryptoutilSharedUtilRandom.ValidateUUID(kid, ErrInvalidJWEJWKKidUUID); err != nil {
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
		return validateOrGenerateJWEAESJWK(key, enc, alg, encKeyBitsLength, &EncA256GCM, &EncA256CBCHS512, &EncA192GCM, &EncA192CBCHS384, &EncA128GCM, &EncA128CBCHS256)

	case AlgA256KW, AlgA256GCMKW:
		return validateOrGenerateJWEAESJWK(key, enc, alg, cryptoutilSharedMagic.JWEA256KeySize, &EncA256GCM, &EncA256CBCHS512, &EncA192GCM, &EncA192CBCHS384, &EncA128GCM, &EncA128CBCHS256)
	case AlgA192KW, AlgA192GCMKW:
		return validateOrGenerateJWEAESJWK(key, enc, alg, cryptoutilSharedMagic.JWEA192KeySize, &EncA256GCM, &EncA256CBCHS512, &EncA192GCM, &EncA192CBCHS384, &EncA128GCM, &EncA128CBCHS256)
	case AlgA128KW, AlgA128GCMKW:
		return validateOrGenerateJWEAESJWK(key, enc, alg, cryptoutilSharedMagic.JWEA128KeySize, &EncA256GCM, &EncA256CBCHS512, &EncA192GCM, &EncA192CBCHS384, &EncA128GCM, &EncA128CBCHS256)

	case AlgRSAOAEP512:
		return validateOrGenerateJWERSAJWK(key, enc, alg, cryptoutilSharedMagic.RSAKeySize4096, &EncA256GCM, &EncA256CBCHS512, &EncA192GCM, &EncA192CBCHS384, &EncA128GCM, &EncA128CBCHS256)
	case AlgRSAOAEP384:
		return validateOrGenerateJWERSAJWK(key, enc, alg, cryptoutilSharedMagic.RSAKeySize3072, &EncA256GCM, &EncA256CBCHS512, &EncA192GCM, &EncA192CBCHS384, &EncA128GCM, &EncA128CBCHS256)
	case AlgRSAOAEP256:
		return validateOrGenerateJWERSAJWK(key, enc, alg, cryptoutilSharedMagic.RSAKeySize2048, &EncA256GCM, &EncA256CBCHS512, &EncA192GCM, &EncA192CBCHS384, &EncA128GCM, &EncA128CBCHS256)
	case AlgRSAOAEP:
		return validateOrGenerateJWERSAJWK(key, enc, alg, cryptoutilSharedMagic.RSAKeySize2048, &EncA256GCM, &EncA256CBCHS512, &EncA192GCM, &EncA192CBCHS384, &EncA128GCM, &EncA128CBCHS256)
	case AlgRSA15:
		return validateOrGenerateJWERSAJWK(key, enc, alg, cryptoutilSharedMagic.RSAKeySize2048, &EncA256GCM, &EncA256CBCHS512, &EncA192GCM, &EncA192CBCHS384, &EncA128GCM, &EncA128CBCHS256)

	case AlgECDHES, AlgECDHESA256KW:
		return validateOrGenerateJWEEcdhJWK(key, enc, alg, ecdh.P521(), &EncA256GCM, &EncA256CBCHS512, &EncA192GCM, &EncA192CBCHS384, &EncA128GCM, &EncA128CBCHS256)
	case AlgECDHESA192KW:
		return validateOrGenerateJWEEcdhJWK(key, enc, alg, ecdh.P384(), &EncA256GCM, &EncA256CBCHS512, &EncA192GCM, &EncA192CBCHS384, &EncA128GCM, &EncA128CBCHS256)
	case AlgECDHESA128KW:
		return validateOrGenerateJWEEcdhJWK(key, enc, alg, ecdh.P256(), &EncA256GCM, &EncA256CBCHS512, &EncA192GCM, &EncA192CBCHS384, &EncA128GCM, &EncA128CBCHS256)

	default:
		return nil, fmt.Errorf("unsupported JWE JWK alg %s", *alg)
	}
}

func validateOrGenerateJWEAESJWK(key cryptoutilSharedCryptoKeygen.Key, enc *joseJwa.ContentEncryptionAlgorithm, alg *joseJwa.KeyEncryptionAlgorithm, keyBitsLength int, allowedEncs ...*joseJwa.ContentEncryptionAlgorithm) (cryptoutilSharedCryptoKeygen.SecretKey, error) {
	if !cryptoutilSharedUtil.Contains(allowedEncs, enc) {
		return nil, fmt.Errorf("valid JWE JWK alg %s, but enc %s not allowed; use one of %v", *alg, *enc, allowedEncs)
	}

	if key == nil {
		var keyBytes cryptoutilSharedCryptoKeygen.SecretKey

		var err error

		switch *alg {
		case AlgA256KW, AlgA256GCMKW, AlgA192KW, AlgA192GCMKW, AlgA128KW, AlgA128GCMKW:
			keyBytes, err = cryptoutilSharedCryptoKeygen.GenerateAESKey(keyBitsLength)

		case AlgDir:
			switch *enc {
			case EncA256GCM, EncA192GCM, EncA128GCM:
				keyBytes, err = cryptoutilSharedCryptoKeygen.GenerateAESKey(keyBitsLength)
			case EncA256CBCHS512, EncA192CBCHS384, EncA128CBCHS256:
				keyBytes, err = cryptoutilSharedCryptoKeygen.GenerateAESHSKey(keyBitsLength)
			default:
				return nil, fmt.Errorf("valid JWE JWK alg %s, but invalid enc %s", *enc, *alg)
			}
		}

		if err != nil {
			return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but failed to generate AES %d key: %w", *enc, *alg, keyBitsLength, err)
		}

		return keyBytes, nil
	}

	keyBytes, ok := key.(cryptoutilSharedCryptoKeygen.SecretKey)
	if !ok {
		return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but unsupported key type %T; use []byte", *enc, *alg, key)
	}

	if keyBytes == nil {
		return nil, fmt.Errorf("valid enc %s and alg %s, but invalid nil key bytes", *enc, *alg)
	}

	if len(keyBytes) != keyBitsLength/8 {
		return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but invalid key length %d; use AES %d", *enc, *alg, len(keyBytes), keyBitsLength)
	}

	return keyBytes, nil
}

func validateOrGenerateJWERSAJWK(key cryptoutilSharedCryptoKeygen.Key, enc *joseJwa.ContentEncryptionAlgorithm, alg *joseJwa.KeyEncryptionAlgorithm, keyBitsLength int, allowedEncs ...*joseJwa.ContentEncryptionAlgorithm) (*cryptoutilSharedCryptoKeygen.KeyPair, error) {
	if !cryptoutilSharedUtil.Contains(allowedEncs, enc) {
		return nil, fmt.Errorf("valid JWE JWK alg %s, but enc %s not allowed; use one of %v", *alg, *enc, allowedEncs)
	}

	if key == nil {
		generatedKeyPair, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(keyBitsLength)
		if err != nil {
			return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but failed to generate RSA %d key: %w", *enc, *alg, keyBitsLength, err)
		}

		return generatedKeyPair, nil
	}

	keyPair, ok := key.(*cryptoutilSharedCryptoKeygen.KeyPair)
	if !ok {
		return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but unsupported key type %T; use *cryptoutilKeyGen.Key", *enc, *alg, key)
	}

	rsaPrivateKey, ok := keyPair.Private.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but unsupported key type %T; use *rsa.PrivateKey", *enc, *alg, keyPair.Private)
	}

	if rsaPrivateKey == nil { // pragma: allowlist secret
		return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but invalid nil RSA private key", *enc, *alg)
	}

	rsaPublicKey, ok := keyPair.Public.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but unsupported key type %T; use *rsa.PublicKey", *enc, *alg, keyPair.Public)
	}

	if rsaPublicKey == nil {
		return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but invalid nil RSA public key", *enc, *alg)
	}

	return keyPair, nil
}

func validateOrGenerateJWEEcdhJWK(key cryptoutilSharedCryptoKeygen.Key, enc *joseJwa.ContentEncryptionAlgorithm, alg *joseJwa.KeyEncryptionAlgorithm, ecdhCurve ecdh.Curve, allowedEncs ...*joseJwa.ContentEncryptionAlgorithm) (*cryptoutilSharedCryptoKeygen.KeyPair, error) {
	if !cryptoutilSharedUtil.Contains(allowedEncs, enc) {
		return nil, fmt.Errorf("valid JWE JWK alg %s, but enc %s not allowed; use one of %v", *alg, *enc, allowedEncs)
	}

	if key == nil {
		generatedKeyPair, err := cryptoutilSharedCryptoKeygen.GenerateECDHKeyPair(ecdhCurve)
		if err != nil {
			return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but failed to generate ECDH %s key: %w", *enc, *alg, ecdhCurve, err)
		}

		return generatedKeyPair, nil
	}

	keyPair, ok := key.(*cryptoutilSharedCryptoKeygen.KeyPair)
	if !ok {
		return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but unsupported key type %T; use *cryptoutilKeyGen.Key", *enc, *alg, key)
	}

	ecdhPrivateKey, ok := keyPair.Private.(*ecdh.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but unsupported key type %T; use *ecdh.PrivateKey", *enc, *alg, keyPair.Private)
	}

	if ecdhPrivateKey == nil { // pragma: allowlist secret
		return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but unsupported nil ECDH private key", *enc, *alg)
	}

	ecdhPublicKey, ok := keyPair.Public.(*ecdh.PublicKey)
	if !ok {
		return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but unsupported key type %T; use *ecdh.PublicKey", *enc, *alg, keyPair.Public)
	}

	if ecdhPublicKey == nil {
		return nil, fmt.Errorf("valid JWE JWK enc %s and alg %s, but unsupported nil ECDH public key", *enc, *alg)
	}

	return keyPair, nil
}

// EncToBitsLength returns the key size in bits for a content encryption algorithm.
func EncToBitsLength(enc *joseJwa.ContentEncryptionAlgorithm) (int, error) {
	switch *enc {
	case EncA256GCM:
		return cryptoutilSharedMagic.JWEA256KeySize, nil
	case EncA192GCM:
		return cryptoutilSharedMagic.JWEA192KeySize, nil
	case EncA128GCM:
		return cryptoutilSharedMagic.JWEA128KeySize, nil
	case EncA256CBCHS512:
		return cryptoutilSharedMagic.JWEA512KeySize, nil
	case EncA192CBCHS384:
		return cryptoutilSharedMagic.JWEA384KeySize, nil
	case EncA128CBCHS256:
		return cryptoutilSharedMagic.JWEA256KeySize, nil
	default:
		return 0, fmt.Errorf("unsupported JWE JWK enc %s", *enc)
	}
}

// ExtractAlgEncFromJWEJWK extracts the encryption algorithm and key encryption algorithm from a JWE JWK.
func ExtractAlgEncFromJWEJWK(jwk joseJwk.Key, i int) (*joseJwa.ContentEncryptionAlgorithm, *joseJwa.KeyEncryptionAlgorithm, error) {
	if jwk == nil {
		return nil, nil, fmt.Errorf("JWK %d invalid: %w", i, cryptoutilSharedApperr.ErrCantBeNil)
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

// IsJWEAlg returns true if the algorithm is a JWE key encryption algorithm.
func IsJWEAlg(alg *joseJwa.KeyAlgorithm, i int) (bool, error) {
	if alg == nil {
		return false, fmt.Errorf("alg %d invalid: %w", i, cryptoutilSharedApperr.ErrCantBeNil)
	}

	_, ok := (*alg).(joseJwa.KeyEncryptionAlgorithm)

	return ok, nil
}
