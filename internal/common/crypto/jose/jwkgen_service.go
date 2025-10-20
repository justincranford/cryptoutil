package jose

import (
	"context"
	"crypto/ecdh"
	"crypto/elliptic"
	"errors"
	"fmt"

	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilKeyGen "cryptoutil/internal/common/crypto/keygen"
	cryptoutilPool "cryptoutil/internal/common/pool"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilUtil "cryptoutil/internal/common/util"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

const (
	// Pool sizes for key generation pools (min, max).
	rsa4096PoolMin = 9
	rsa4096PoolMax = 9
	rsa3072PoolMin = 6
	rsa3072PoolMax = 6
	rsa2048PoolMin = 3
	rsa2048PoolMax = 3

	ecdsaP521PoolMin = 3
	ecdsaP521PoolMax = 9
	ecdsaP384PoolMin = 2
	ecdsaP384PoolMax = 6
	ecdsaP256PoolMin = 1
	ecdsaP256PoolMax = 3

	ecdhP521PoolMin = 3
	ecdhP521PoolMax = 9
	ecdhP384PoolMin = 2
	ecdhP384PoolMax = 6
	ecdhP256PoolMin = 1
	ecdhP256PoolMax = 3

	ed25519PoolMin = 1
	ed25519PoolMax = 2

	aes256PoolMin = 3
	aes256PoolMax = 9
	aes192PoolMin = 2
	aes192PoolMax = 6
	aes128PoolMin = 1
	aes128PoolMax = 3

	aes256HS512PoolMin = 3
	aes256HS512PoolMax = 9
	aes192HS384PoolMin = 2
	aes192HS384PoolMax = 6
	aes128HS256PoolMin = 1
	aes128HS256PoolMax = 3

	hmac512PoolMin = 3
	hmac512PoolMax = 9
	hmac384PoolMin = 2
	hmac384PoolMax = 6
	hmac256PoolMin = 1
	hmac256PoolMax = 3

	uuidV7PoolMin = 2
	uuidV7PoolMax = 20
)

type JWKGenService struct {
	telemetryService      *cryptoutilTelemetry.TelemetryService
	RSA4096KeyGenPool     *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair]  // 512-bytes
	RSA3072KeyGenPool     *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair]  // 384-bytes
	RSA2048KeyGenPool     *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair]  // 256-bytes
	ECDSAP521KeyGenPool   *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair]  // 65.125-bytes
	ECDSAP384KeyGenPool   *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair]  // 48-bytes
	ECDSAP256KeyGenPool   *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair]  // 32-bytes
	ECDHP521KeyGenPool    *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair]  // 65.125-bytes
	ECDHP384KeyGenPool    *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair]  // 48-bytes
	ECDHP256KeyGenPool    *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair]  // 32-bytes
	ED25519KeyGenPool     *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair]  // 32-bytes
	AES256KeyGenPool      *cryptoutilPool.ValueGenPool[cryptoutilKeyGen.SecretKey] // 32-bytes A256GCM, A256KW, A256GCMKW
	AES192KeyGenPool      *cryptoutilPool.ValueGenPool[cryptoutilKeyGen.SecretKey] // 24-bytes A192GCM, A192KW, A192GCMKW
	AES128KeyGenPool      *cryptoutilPool.ValueGenPool[cryptoutilKeyGen.SecretKey] // 16-bytes A128GCM, A128KW, A128GCMKW
	AES256HS512KeyGenPool *cryptoutilPool.ValueGenPool[cryptoutilKeyGen.SecretKey] // 32-bytes A256CBC + 32-bytes HS512 (half of 64-bytes)
	AES192HS384KeyGenPool *cryptoutilPool.ValueGenPool[cryptoutilKeyGen.SecretKey] // 24-bytes A192CBC + 24-bytes HS384 (half of 48-bytes)
	AES128HS256KeyGenPool *cryptoutilPool.ValueGenPool[cryptoutilKeyGen.SecretKey] // 16-bytes A128CBC + 16-bytes HS256 (half of 32-bytes)
	HMAC512KeyGenPool     *cryptoutilPool.ValueGenPool[cryptoutilKeyGen.SecretKey] // 64-bytes HS512
	HMAC384KeyGenPool     *cryptoutilPool.ValueGenPool[cryptoutilKeyGen.SecretKey] // 48-bytes HS384
	HMAC256KeyGenPool     *cryptoutilPool.ValueGenPool[cryptoutilKeyGen.SecretKey] // 32-bytes HS256
	UUIDv7KeyGenPool      *cryptoutilPool.ValueGenPool[*googleUuid.UUID]
}

func NewJWKGenService(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService) (*JWKGenService, error) {
	verbose := false // TODO read from settings

	if ctx == nil {
		return nil, fmt.Errorf("context must be non-nil")
	} else if telemetryService == nil {
		return nil, fmt.Errorf("telemetry service must be non-nil")
	}

	rsa4096KeyGenPool, err1 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService RSA 4096", rsa4096PoolMin, rsa4096PoolMax, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateRSAKeyPairFunction(4096), verbose))
	rsa3072KeyGenPool, err2 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService RSA 3072", rsa3072PoolMin, rsa3072PoolMax, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateRSAKeyPairFunction(3072), verbose))
	rsa2048KeyGenPool, err3 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService RSA 2048", rsa2048PoolMin, rsa2048PoolMax, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateRSAKeyPairFunction(2048), verbose))
	ecdsaP521KeyGenPool, err4 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService ECDSA-P521", ecdsaP521PoolMin, ecdsaP521PoolMax, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateECDSAKeyPairFunction(elliptic.P521()), verbose))
	ecdsaP384KeyGenPool, err5 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService ECDSA-P384", ecdsaP384PoolMin, ecdsaP384PoolMax, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateECDSAKeyPairFunction(elliptic.P384()), verbose))
	ecdsaP256KeyGenPool, err6 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService ECDSA-P256", ecdsaP256PoolMin, ecdsaP256PoolMax, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateECDSAKeyPairFunction(elliptic.P256()), verbose))
	ecdhP521KeyGenPool, err7 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService ECDH-P521", ecdhP521PoolMin, ecdhP521PoolMax, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateECDHKeyPairFunction(ecdh.P521()), verbose))
	ecdhP384KeyGenPool, err8 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService ECDH-P384", ecdhP384PoolMin, ecdhP384PoolMax, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateECDHKeyPairFunction(ecdh.P384()), verbose))
	ecdhP256KeyGenPool, err9 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService ECDH-P256", ecdhP256PoolMin, ecdhP256PoolMax, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateECDHKeyPairFunction(ecdh.P256()), verbose))
	ed25519KeyGenPool, err10 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService Ed25519", ed25519PoolMin, ed25519PoolMax, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateEDDSAKeyPairFunction("Ed25519"), verbose))
	aes256KeyGenPool, err11 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService AES-256-GCM", aes256PoolMin, aes256PoolMax, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateAESKeyFunction(256), verbose))
	aes192KeyGenPool, err12 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService AES-192-GCM", aes192PoolMin, aes192PoolMax, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateAESKeyFunction(192), verbose))
	aes128KeyGenPool, err13 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService AES-128-GCM", aes128PoolMin, aes128PoolMax, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateAESKeyFunction(128), verbose))
	aes256HS512KeyGenPool, err14 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService AES-256-CBC HS-512", aes256HS512PoolMin, aes256HS512PoolMax, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateAESHSKeyFunction(512), verbose))
	aes192HS384KeyGenPool, err15 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService AES-192-CBC HS-384", aes192HS384PoolMin, aes192HS384PoolMax, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateAESHSKeyFunction(384), verbose))
	aes128HS256KeyGenPool, err16 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService AES-128-CBC HS-256", aes128HS256PoolMin, aes128HS256PoolMax, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateAESHSKeyFunction(256), verbose))
	hmac512KeyGenPool, err17 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService HMAC-512", hmac512PoolMin, hmac512PoolMax, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateHMACKeyFunction(512), verbose))
	hmac384KeyGenPool, err18 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService HMAC-384", hmac384PoolMin, hmac384PoolMax, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateHMACKeyFunction(384), verbose))
	hmac256KeyGenPool, err19 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService HMAC-256", hmac256PoolMin, hmac256PoolMax, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateHMACKeyFunction(256), verbose))
	uuidV7KeyGenPool, err20 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JWKGenService UUIDv7", uuidV7PoolMin, uuidV7PoolMax, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilUtil.GenerateUUIDv7Function(), verbose))

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

func (s *JWKGenService) Shutdown() {
	s.telemetryService.Slogger.Debug("stopping JWKGenService")
	cryptoutilPool.CancelAllNotNil([]*cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair]{
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
	cryptoutilPool.CancelAllNotNil([]*cryptoutilPool.ValueGenPool[cryptoutilKeyGen.SecretKey]{
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
	cryptoutilPool.CancelNotNil(s.UUIDv7KeyGenPool)
}

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

func (s *JWKGenService) GenerateJWSJWK(alg *joseJwa.SignatureAlgorithm) (*googleUuid.UUID, joseJwk.Key, joseJwk.Key, []byte, []byte, error) {
	switch *alg {
	case AlgRS512, AlgPS512:
		return CreateJWSJWKFromKey(s.UUIDv7KeyGenPool.Get(), alg, s.RSA4096KeyGenPool.Get())
	case AlgRS384, AlgPS384:
		return CreateJWSJWKFromKey(s.UUIDv7KeyGenPool.Get(), alg, s.RSA3072KeyGenPool.Get())
	case AlgRS256, AlgPS256:
		return CreateJWSJWKFromKey(s.UUIDv7KeyGenPool.Get(), alg, s.RSA2048KeyGenPool.Get())
	case AlgES512:
		return CreateJWSJWKFromKey(s.UUIDv7KeyGenPool.Get(), alg, s.ECDSAP521KeyGenPool.Get())
	case AlgES384:
		return CreateJWSJWKFromKey(s.UUIDv7KeyGenPool.Get(), alg, s.ECDSAP384KeyGenPool.Get())
	case AlgES256:
		return CreateJWSJWKFromKey(s.UUIDv7KeyGenPool.Get(), alg, s.ECDSAP256KeyGenPool.Get())
	case AlgEdDSA:
		return CreateJWSJWKFromKey(s.UUIDv7KeyGenPool.Get(), alg, s.ED25519KeyGenPool.Get())
	case AlgHS512:
		return CreateJWSJWKFromKey(s.UUIDv7KeyGenPool.Get(), alg, s.HMAC512KeyGenPool.Get())
	case AlgHS384:
		return CreateJWSJWKFromKey(s.UUIDv7KeyGenPool.Get(), alg, s.HMAC384KeyGenPool.Get())
	case AlgHS256:
		return CreateJWSJWKFromKey(s.UUIDv7KeyGenPool.Get(), alg, s.HMAC256KeyGenPool.Get())
	default:
		return nil, nil, nil, nil, nil, fmt.Errorf("unsupported JWS JWK alg: %s", alg)
	}
}

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

func (s *JWKGenService) GenerateUUIDv7() *googleUuid.UUID {
	return s.UUIDv7KeyGenPool.Get()
}
