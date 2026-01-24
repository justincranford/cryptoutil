// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	"context"
	"crypto/ecdh"
	"crypto/elliptic"
	"errors"
	"fmt"

	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedPool "cryptoutil/internal/shared/pool"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

// JWKGenService provides pooled JWK key generation for various algorithms.
type JWKGenService struct {
	telemetryService      *cryptoutilSharedTelemetry.TelemetryService
	RSA4096KeyGenPool     *cryptoutilSharedPool.ValueGenPool[*cryptoutilSharedCryptoKeygen.KeyPair]  // 512-bytes
	RSA3072KeyGenPool     *cryptoutilSharedPool.ValueGenPool[*cryptoutilSharedCryptoKeygen.KeyPair]  // 384-bytes
	RSA2048KeyGenPool     *cryptoutilSharedPool.ValueGenPool[*cryptoutilSharedCryptoKeygen.KeyPair]  // 256-bytes
	ECDSAP521KeyGenPool   *cryptoutilSharedPool.ValueGenPool[*cryptoutilSharedCryptoKeygen.KeyPair]  // 65.125-bytes
	ECDSAP384KeyGenPool   *cryptoutilSharedPool.ValueGenPool[*cryptoutilSharedCryptoKeygen.KeyPair]  // 48-bytes
	ECDSAP256KeyGenPool   *cryptoutilSharedPool.ValueGenPool[*cryptoutilSharedCryptoKeygen.KeyPair]  // 32-bytes
	ECDHP521KeyGenPool    *cryptoutilSharedPool.ValueGenPool[*cryptoutilSharedCryptoKeygen.KeyPair]  // 65.125-bytes
	ECDHP384KeyGenPool    *cryptoutilSharedPool.ValueGenPool[*cryptoutilSharedCryptoKeygen.KeyPair]  // 48-bytes
	ECDHP256KeyGenPool    *cryptoutilSharedPool.ValueGenPool[*cryptoutilSharedCryptoKeygen.KeyPair]  // 32-bytes
	ED25519KeyGenPool     *cryptoutilSharedPool.ValueGenPool[*cryptoutilSharedCryptoKeygen.KeyPair]  // 32-bytes
	AES256KeyGenPool      *cryptoutilSharedPool.ValueGenPool[cryptoutilSharedCryptoKeygen.SecretKey] // 32-bytes A256GCM, A256KW, A256GCMKW
	AES192KeyGenPool      *cryptoutilSharedPool.ValueGenPool[cryptoutilSharedCryptoKeygen.SecretKey] // 24-bytes A192GCM, A192KW, A192GCMKW
	AES128KeyGenPool      *cryptoutilSharedPool.ValueGenPool[cryptoutilSharedCryptoKeygen.SecretKey] // 16-bytes A128GCM, A128KW, A128GCMKW
	AES256HS512KeyGenPool *cryptoutilSharedPool.ValueGenPool[cryptoutilSharedCryptoKeygen.SecretKey] // 32-bytes A256CBC + 32-bytes HS512 (half of 64-bytes)
	AES192HS384KeyGenPool *cryptoutilSharedPool.ValueGenPool[cryptoutilSharedCryptoKeygen.SecretKey] // 24-bytes A192CBC + 24-bytes HS384 (half of 48-bytes)
	AES128HS256KeyGenPool *cryptoutilSharedPool.ValueGenPool[cryptoutilSharedCryptoKeygen.SecretKey] // 16-bytes A128CBC + 16-bytes HS256 (half of 32-bytes)
	HMAC512KeyGenPool     *cryptoutilSharedPool.ValueGenPool[cryptoutilSharedCryptoKeygen.SecretKey] // 64-bytes HS512
	HMAC384KeyGenPool     *cryptoutilSharedPool.ValueGenPool[cryptoutilSharedCryptoKeygen.SecretKey] // 48-bytes HS384
	HMAC256KeyGenPool     *cryptoutilSharedPool.ValueGenPool[cryptoutilSharedCryptoKeygen.SecretKey] // 32-bytes HS256
	UUIDv7KeyGenPool      *cryptoutilSharedPool.ValueGenPool[*googleUuid.UUID]
}

// NewJWKGenService creates a new JWKGenService with pooled key generation.
func NewJWKGenService(ctx context.Context, telemetryService *cryptoutilSharedTelemetry.TelemetryService, verbose bool) (*JWKGenService, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context must be non-nil")
	} else if telemetryService == nil {
		return nil, fmt.Errorf("telemetry service must be non-nil")
	}

	rsa4096KeyGenPool, err1 := cryptoutilSharedPool.NewValueGenPool(cryptoutilSharedPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService RSA 4096", cryptoutilSharedMagic.DefaultPoolConfigRSA4096.NumWorkers, cryptoutilSharedMagic.DefaultPoolConfigRSA4096.MaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateRSAKeyPairFunction(cryptoutilSharedMagic.RSAKeySize4096), verbose))
	rsa3072KeyGenPool, err2 := cryptoutilSharedPool.NewValueGenPool(cryptoutilSharedPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService RSA 3072", cryptoutilSharedMagic.DefaultPoolConfigRSA3072.NumWorkers, cryptoutilSharedMagic.DefaultPoolConfigRSA3072.MaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateRSAKeyPairFunction(cryptoutilSharedMagic.RSAKeySize3072), verbose))
	rsa2048KeyGenPool, err3 := cryptoutilSharedPool.NewValueGenPool(cryptoutilSharedPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService RSA 2048", cryptoutilSharedMagic.DefaultPoolConfigRSA2048.NumWorkers, cryptoutilSharedMagic.DefaultPoolConfigRSA2048.MaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateRSAKeyPairFunction(cryptoutilSharedMagic.RSAKeySize2048), verbose))
	ecdsaP521KeyGenPool, err4 := cryptoutilSharedPool.NewValueGenPool(cryptoutilSharedPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService ECDSA-P521", cryptoutilSharedMagic.DefaultPoolConfigECDSAP521.NumWorkers, cryptoutilSharedMagic.DefaultPoolConfigECDSAP521.MaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPairFunction(elliptic.P521()), verbose))
	ecdsaP384KeyGenPool, err5 := cryptoutilSharedPool.NewValueGenPool(cryptoutilSharedPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService ECDSA-P384", cryptoutilSharedMagic.DefaultPoolConfigECDSAP384.NumWorkers, cryptoutilSharedMagic.DefaultPoolConfigECDSAP384.MaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPairFunction(elliptic.P384()), verbose))
	ecdsaP256KeyGenPool, err6 := cryptoutilSharedPool.NewValueGenPool(cryptoutilSharedPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService ECDSA-P256", cryptoutilSharedMagic.DefaultPoolConfigECDSAP256.NumWorkers, cryptoutilSharedMagic.DefaultPoolConfigECDSAP256.MaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPairFunction(elliptic.P256()), verbose))
	ecdhP521KeyGenPool, err7 := cryptoutilSharedPool.NewValueGenPool(cryptoutilSharedPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService ECDH-P521", cryptoutilSharedMagic.DefaultPoolConfigECDHP521.NumWorkers, cryptoutilSharedMagic.DefaultPoolConfigECDHP521.MaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateECDHKeyPairFunction(ecdh.P521()), verbose))
	ecdhP384KeyGenPool, err8 := cryptoutilSharedPool.NewValueGenPool(cryptoutilSharedPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService ECDH-P384", cryptoutilSharedMagic.DefaultPoolConfigECDHP384.NumWorkers, cryptoutilSharedMagic.DefaultPoolConfigECDHP384.MaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateECDHKeyPairFunction(ecdh.P384()), verbose))
	ecdhP256KeyGenPool, err9 := cryptoutilSharedPool.NewValueGenPool(cryptoutilSharedPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService ECDH-P256", cryptoutilSharedMagic.DefaultPoolConfigECDHP256.NumWorkers, cryptoutilSharedMagic.DefaultPoolConfigECDHP256.MaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateECDHKeyPairFunction(ecdh.P256()), verbose))
	ed25519KeyGenPool, err10 := cryptoutilSharedPool.NewValueGenPool(cryptoutilSharedPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService Ed25519", cryptoutilSharedMagic.DefaultPoolConfigED25519.NumWorkers, cryptoutilSharedMagic.DefaultPoolConfigED25519.MaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateEDDSAKeyPairFunction("Ed25519"), verbose))
	aes256KeyGenPool, err11 := cryptoutilSharedPool.NewValueGenPool(cryptoutilSharedPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService AES-256-GCM", cryptoutilSharedMagic.DefaultPoolConfigAES256.NumWorkers, cryptoutilSharedMagic.DefaultPoolConfigAES256.MaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateAESKeyFunction(cryptoutilSharedMagic.AESKeySize256), verbose))
	aes192KeyGenPool, err12 := cryptoutilSharedPool.NewValueGenPool(cryptoutilSharedPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService AES-192-GCM", cryptoutilSharedMagic.DefaultPoolConfigAES192.NumWorkers, cryptoutilSharedMagic.DefaultPoolConfigAES192.MaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateAESKeyFunction(cryptoutilSharedMagic.AESKeySize192), verbose))
	aes128KeyGenPool, err13 := cryptoutilSharedPool.NewValueGenPool(cryptoutilSharedPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService AES-128-GCM", cryptoutilSharedMagic.DefaultPoolConfigAES128.NumWorkers, cryptoutilSharedMagic.DefaultPoolConfigAES128.MaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateAESKeyFunction(cryptoutilSharedMagic.AESKeySize128), verbose))
	aes256HS512KeyGenPool, err14 := cryptoutilSharedPool.NewValueGenPool(cryptoutilSharedPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService AES-256-CBC HS-512", cryptoutilSharedMagic.DefaultPoolConfigAES256HS512.NumWorkers, cryptoutilSharedMagic.DefaultPoolConfigAES256HS512.MaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateAESHSKeyFunction(cryptoutilSharedMagic.HMACKeySize512), verbose))
	aes192HS384KeyGenPool, err15 := cryptoutilSharedPool.NewValueGenPool(cryptoutilSharedPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService AES-192-CBC HS-384", cryptoutilSharedMagic.DefaultPoolConfigAES192HS384.NumWorkers, cryptoutilSharedMagic.DefaultPoolConfigAES192HS384.MaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateAESHSKeyFunction(cryptoutilSharedMagic.HMACKeySize384), verbose))
	aes128HS256KeyGenPool, err16 := cryptoutilSharedPool.NewValueGenPool(cryptoutilSharedPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService AES-128-CBC HS-256", cryptoutilSharedMagic.DefaultPoolConfigAES128HS256.NumWorkers, cryptoutilSharedMagic.DefaultPoolConfigAES128HS256.MaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateAESHSKeyFunction(cryptoutilSharedMagic.HMACKeySize256), verbose))
	hmac512KeyGenPool, err17 := cryptoutilSharedPool.NewValueGenPool(cryptoutilSharedPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService HMAC-512", cryptoutilSharedMagic.DefaultPoolConfigHMAC512.NumWorkers, cryptoutilSharedMagic.DefaultPoolConfigHMAC512.MaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateHMACKeyFunction(cryptoutilSharedMagic.HMACKeySize512), verbose))
	hmac384KeyGenPool, err18 := cryptoutilSharedPool.NewValueGenPool(cryptoutilSharedPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService HMAC-384", cryptoutilSharedMagic.DefaultPoolConfigHMAC384.NumWorkers, cryptoutilSharedMagic.DefaultPoolConfigHMAC384.MaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateHMACKeyFunction(cryptoutilSharedMagic.HMACKeySize384), verbose))
	hmac256KeyGenPool, err19 := cryptoutilSharedPool.NewValueGenPool(cryptoutilSharedPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService HMAC-256", cryptoutilSharedMagic.DefaultPoolConfigHMAC256.NumWorkers, cryptoutilSharedMagic.DefaultPoolConfigHMAC256.MaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateHMACKeyFunction(cryptoutilSharedMagic.HMACKeySize256), verbose))
	uuidV7KeyGenPool, err20 := cryptoutilSharedPool.NewValueGenPool(cryptoutilSharedPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService UUIDv7", cryptoutilSharedMagic.DefaultPoolConfigUUIDv7.NumWorkers, cryptoutilSharedMagic.DefaultPoolConfigUUIDv7.MaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedUtilRandom.GenerateUUIDv7Function(), verbose))

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil || err7 != nil || err8 != nil || err9 != nil || err10 != nil || err11 != nil || err12 != nil || err13 != nil || err14 != nil || err15 != nil || err16 != nil || err17 != nil || err18 != nil || err19 != nil || err20 != nil {
		return nil, fmt.Errorf("failed to create pools: %w", errors.Join(err1, err2, err3, err4, err5, err6, err7, err8, err9, err10, err11, err12, err13, err14, err15, err16, err17, err18, err19, err20))
	}

	return &JWKGenService{
		telemetryService:      telemetryService,
		RSA4096KeyGenPool:     rsa4096KeyGenPool,
		RSA3072KeyGenPool:     rsa3072KeyGenPool,
		RSA2048KeyGenPool:     rsa2048KeyGenPool,
		ECDSAP521KeyGenPool:   ecdsaP521KeyGenPool,
		ECDSAP384KeyGenPool:   ecdsaP384KeyGenPool,
		ECDSAP256KeyGenPool:   ecdsaP256KeyGenPool,
		ECDHP521KeyGenPool:    ecdhP521KeyGenPool,
		ECDHP384KeyGenPool:    ecdhP384KeyGenPool,
		ECDHP256KeyGenPool:    ecdhP256KeyGenPool,
		ED25519KeyGenPool:     ed25519KeyGenPool,
		AES256KeyGenPool:      aes256KeyGenPool,
		AES192KeyGenPool:      aes192KeyGenPool,
		AES128KeyGenPool:      aes128KeyGenPool,
		AES256HS512KeyGenPool: aes256HS512KeyGenPool,
		AES192HS384KeyGenPool: aes192HS384KeyGenPool,
		AES128HS256KeyGenPool: aes128HS256KeyGenPool,
		HMAC512KeyGenPool:     hmac512KeyGenPool,
		HMAC384KeyGenPool:     hmac384KeyGenPool,
		HMAC256KeyGenPool:     hmac256KeyGenPool,
		UUIDv7KeyGenPool:      uuidV7KeyGenPool,
	}, nil
}

// Shutdown gracefully shuts down the JWKGenService and its key generation pools.
func (s *JWKGenService) Shutdown() {
	s.telemetryService.Slogger.Debug("stopping JWKGenService")
	cryptoutilSharedPool.CancelAllNotNil([]*cryptoutilSharedPool.ValueGenPool[*cryptoutilSharedCryptoKeygen.KeyPair]{
		s.RSA4096KeyGenPool,
		s.RSA3072KeyGenPool,
		s.RSA2048KeyGenPool,
		s.ECDSAP521KeyGenPool,
		s.ECDSAP384KeyGenPool,
		s.ECDSAP256KeyGenPool,
		s.ECDHP521KeyGenPool,
		s.ECDHP384KeyGenPool,
		s.ECDHP256KeyGenPool,
		s.ED25519KeyGenPool,
	})
	cryptoutilSharedPool.CancelAllNotNil([]*cryptoutilSharedPool.ValueGenPool[cryptoutilSharedCryptoKeygen.SecretKey]{
		s.AES256KeyGenPool,
		s.AES192KeyGenPool,
		s.AES128KeyGenPool,
		s.AES256HS512KeyGenPool,
		s.AES192HS384KeyGenPool,
		s.AES128HS256KeyGenPool,
		s.HMAC512KeyGenPool,
		s.HMAC384KeyGenPool,
		s.HMAC256KeyGenPool,
	})
	cryptoutilSharedPool.CancelNotNil(s.UUIDv7KeyGenPool)
}

// GenerateJWEJWK generates a JWE JWK using the pooled key generation service.
func (s *JWKGenService) GenerateJWEJWK(enc *joseJwa.ContentEncryptionAlgorithm, alg *joseJwa.KeyEncryptionAlgorithm) (*googleUuid.UUID, joseJwk.Key, joseJwk.Key, []byte, []byte, error) {
	switch *alg {
	case AlgDir:
		switch *enc {
		case EncA256GCM:
			return CreateJWEJWKFromKey(s.UUIDv7KeyGenPool.Get(), enc, alg, s.AES256KeyGenPool.Get())
		case EncA192GCM:
			return CreateJWEJWKFromKey(s.UUIDv7KeyGenPool.Get(), enc, alg, s.AES192KeyGenPool.Get())
		case EncA128GCM:
			return CreateJWEJWKFromKey(s.UUIDv7KeyGenPool.Get(), enc, alg, s.AES128KeyGenPool.Get())
		case EncA256CBCHS512:
			return CreateJWEJWKFromKey(s.UUIDv7KeyGenPool.Get(), enc, alg, s.AES256HS512KeyGenPool.Get())
		case EncA192CBCHS384:
			return CreateJWEJWKFromKey(s.UUIDv7KeyGenPool.Get(), enc, alg, s.AES192HS384KeyGenPool.Get())
		case EncA128CBCHS256:
			return CreateJWEJWKFromKey(s.UUIDv7KeyGenPool.Get(), enc, alg, s.AES128HS256KeyGenPool.Get())
		default:
			return nil, nil, nil, nil, nil, fmt.Errorf("unsupported JWE JWK enc %s", *enc)
		}

	case AlgA256KW, AlgA256GCMKW:
		return CreateJWEJWKFromKey(s.UUIDv7KeyGenPool.Get(), enc, alg, s.AES256KeyGenPool.Get())
	case AlgA192KW, AlgA192GCMKW:
		return CreateJWEJWKFromKey(s.UUIDv7KeyGenPool.Get(), enc, alg, s.AES192KeyGenPool.Get())
	case AlgA128KW, AlgA128GCMKW:
		return CreateJWEJWKFromKey(s.UUIDv7KeyGenPool.Get(), enc, alg, s.AES128KeyGenPool.Get())

	case AlgRSAOAEP512:
		return CreateJWEJWKFromKey(s.UUIDv7KeyGenPool.Get(), enc, alg, s.RSA4096KeyGenPool.Get())
	case AlgRSAOAEP384:
		return CreateJWEJWKFromKey(s.UUIDv7KeyGenPool.Get(), enc, alg, s.RSA3072KeyGenPool.Get())
	case AlgRSAOAEP256:
		return CreateJWEJWKFromKey(s.UUIDv7KeyGenPool.Get(), enc, alg, s.RSA2048KeyGenPool.Get())
	case AlgRSAOAEP:
		return CreateJWEJWKFromKey(s.UUIDv7KeyGenPool.Get(), enc, alg, s.RSA2048KeyGenPool.Get())
	case AlgRSA15:
		return CreateJWEJWKFromKey(s.UUIDv7KeyGenPool.Get(), enc, alg, s.RSA2048KeyGenPool.Get())

	case AlgECDHES, AlgECDHESA256KW:
		return CreateJWEJWKFromKey(s.UUIDv7KeyGenPool.Get(), enc, alg, s.ECDHP521KeyGenPool.Get())
	case AlgECDHESA192KW:
		return CreateJWEJWKFromKey(s.UUIDv7KeyGenPool.Get(), enc, alg, s.ECDHP384KeyGenPool.Get())
	case AlgECDHESA128KW:
		return CreateJWEJWKFromKey(s.UUIDv7KeyGenPool.Get(), enc, alg, s.ECDHP256KeyGenPool.Get())

	default:
		return nil, nil, nil, nil, nil, fmt.Errorf("unsupported JWE JWK alg %s", *alg)
	}
}

// GenerateJWSJWK generates a JWS JWK using the pooled key generation service.
func (s *JWKGenService) GenerateJWSJWK(alg joseJwa.SignatureAlgorithm) (*googleUuid.UUID, joseJwk.Key, joseJwk.Key, []byte, []byte, error) {
	switch alg.String() {
	case "PS512":
		return CreateJWSJWKFromKey(s.UUIDv7KeyGenPool.Get(), &alg, s.RSA4096KeyGenPool.Get())
	case "PS384":
		return CreateJWSJWKFromKey(s.UUIDv7KeyGenPool.Get(), &alg, s.RSA3072KeyGenPool.Get())
	case "PS256":
		return CreateJWSJWKFromKey(s.UUIDv7KeyGenPool.Get(), &alg, s.RSA2048KeyGenPool.Get())
	case "RS512":
		return CreateJWSJWKFromKey(s.UUIDv7KeyGenPool.Get(), &alg, s.RSA4096KeyGenPool.Get())
	case "RS384":
		return CreateJWSJWKFromKey(s.UUIDv7KeyGenPool.Get(), &alg, s.RSA3072KeyGenPool.Get())
	case "RS256":
		return CreateJWSJWKFromKey(s.UUIDv7KeyGenPool.Get(), &alg, s.RSA2048KeyGenPool.Get())
	case "ES512":
		return CreateJWSJWKFromKey(s.UUIDv7KeyGenPool.Get(), &alg, s.ECDSAP521KeyGenPool.Get())
	case "ES384":
		return CreateJWSJWKFromKey(s.UUIDv7KeyGenPool.Get(), &alg, s.ECDSAP384KeyGenPool.Get())
	case "ES256":
		return CreateJWSJWKFromKey(s.UUIDv7KeyGenPool.Get(), &alg, s.ECDSAP256KeyGenPool.Get())
	case "EdDSA":
		return CreateJWSJWKFromKey(s.UUIDv7KeyGenPool.Get(), &alg, s.ED25519KeyGenPool.Get())
	case "HS512":
		return CreateJWSJWKFromKey(s.UUIDv7KeyGenPool.Get(), &alg, s.HMAC512KeyGenPool.Get())
	case "HS384":
		return CreateJWSJWKFromKey(s.UUIDv7KeyGenPool.Get(), &alg, s.HMAC384KeyGenPool.Get())
	case "HS256":
		return CreateJWSJWKFromKey(s.UUIDv7KeyGenPool.Get(), &alg, s.HMAC256KeyGenPool.Get())
	default:
		return nil, nil, nil, nil, nil, fmt.Errorf("unsupported JWS JWK alg: %s", alg)
	}
}

// GenerateJWK generates a JWK for the specified algorithm using pooled key generation.
func (s *JWKGenService) GenerateJWK(alg *cryptoutilOpenapiModel.GenerateAlgorithm) (*googleUuid.UUID, joseJwk.Key, joseJwk.Key, []byte, []byte, error) {
	switch *alg {
	case cryptoutilOpenapiModel.RSA4096:
		return CreateJWKFromKey(s.UUIDv7KeyGenPool.Get(), alg, s.RSA4096KeyGenPool.Get())
	case cryptoutilOpenapiModel.RSA3072:
		return CreateJWKFromKey(s.UUIDv7KeyGenPool.Get(), alg, s.RSA3072KeyGenPool.Get())
	case cryptoutilOpenapiModel.RSA2048:
		return CreateJWKFromKey(s.UUIDv7KeyGenPool.Get(), alg, s.RSA2048KeyGenPool.Get())
	case cryptoutilOpenapiModel.ECP521:
		return CreateJWKFromKey(s.UUIDv7KeyGenPool.Get(), alg, s.ECDSAP521KeyGenPool.Get())
	case cryptoutilOpenapiModel.ECP384:
		return CreateJWKFromKey(s.UUIDv7KeyGenPool.Get(), alg, s.ECDSAP384KeyGenPool.Get())
	case cryptoutilOpenapiModel.ECP256:
		return CreateJWKFromKey(s.UUIDv7KeyGenPool.Get(), alg, s.ECDSAP256KeyGenPool.Get())
	case cryptoutilOpenapiModel.OKPEd25519:
		return CreateJWKFromKey(s.UUIDv7KeyGenPool.Get(), alg, s.ED25519KeyGenPool.Get())
	case cryptoutilOpenapiModel.Oct512:
		return CreateJWKFromKey(s.UUIDv7KeyGenPool.Get(), alg, s.AES256HS512KeyGenPool.Get())
	case cryptoutilOpenapiModel.Oct384:
		return CreateJWKFromKey(s.UUIDv7KeyGenPool.Get(), alg, s.AES192HS384KeyGenPool.Get())
	case cryptoutilOpenapiModel.Oct256:
		return CreateJWKFromKey(s.UUIDv7KeyGenPool.Get(), alg, s.AES128HS256KeyGenPool.Get())
	case cryptoutilOpenapiModel.Oct192:
		return CreateJWKFromKey(s.UUIDv7KeyGenPool.Get(), alg, s.AES192KeyGenPool.Get())
	case cryptoutilOpenapiModel.Oct128:
		return CreateJWKFromKey(s.UUIDv7KeyGenPool.Get(), alg, s.AES128KeyGenPool.Get())
	default:
		return nil, nil, nil, nil, nil, fmt.Errorf("unsupported JWK alg: %v", alg)
	}
}

// GenerateUUIDv7 generates a UUID v7 using the pooled generation service.
func (s *JWKGenService) GenerateUUIDv7() *googleUuid.UUID {
	return s.UUIDv7KeyGenPool.Get()
}
