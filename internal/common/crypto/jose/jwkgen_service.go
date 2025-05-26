package jose

import (
	"context"
	"crypto/ecdh"
	"crypto/elliptic"
	cryptoutilKeygen "cryptoutil/internal/common/crypto/keygen"
	"cryptoutil/internal/common/pool"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	"errors"
	"fmt"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

type JwkGenService struct {
	telemetryService      *cryptoutilTelemetry.TelemetryService
	rsa4096KeyGenPool     *pool.ValueGenPool[cryptoutilKeygen.Key] // 512-bytes
	rsa3072KeyGenPool     *pool.ValueGenPool[cryptoutilKeygen.Key] // 384-bytes
	rsa2048KeyGenPool     *pool.ValueGenPool[cryptoutilKeygen.Key] // 256-bytes
	ecdsaP521KeyGenPool   *pool.ValueGenPool[cryptoutilKeygen.Key] // 65.125-bytes
	ecdsaP384KeyGenPool   *pool.ValueGenPool[cryptoutilKeygen.Key] // 48-bytes
	ecdsaP256KeyGenPool   *pool.ValueGenPool[cryptoutilKeygen.Key] // 32-bytes
	ecdhP521KeyGenPool    *pool.ValueGenPool[cryptoutilKeygen.Key] // 65.125-bytes
	ecdhP384KeyGenPool    *pool.ValueGenPool[cryptoutilKeygen.Key] // 48-bytes
	ecdhP256KeyGenPool    *pool.ValueGenPool[cryptoutilKeygen.Key] // 32-bytes
	ed25519KeyGenPool     *pool.ValueGenPool[cryptoutilKeygen.Key] // 32-bytes
	aes256KeyGenPool      *pool.ValueGenPool[cryptoutilKeygen.Key] // 32-bytes A256GCM, A256KW, A256GCMKW
	aes192KeyGenPool      *pool.ValueGenPool[cryptoutilKeygen.Key] // 24-bytes A192GCM, A192KW, A192GCMKW
	aes128KeyGenPool      *pool.ValueGenPool[cryptoutilKeygen.Key] // 16-bytes A128GCM, A128KW, A128GCMKW
	aes256HS512KeyGenPool *pool.ValueGenPool[cryptoutilKeygen.Key] // 32-bytes A256CBC + 32-bytes HS512 (half of 64-bytes)
	aes192HS384KeyGenPool *pool.ValueGenPool[cryptoutilKeygen.Key] // 24-bytes A192CBC + 24-bytes HS384 (half of 48-bytes)
	aes128HS256KeyGenPool *pool.ValueGenPool[cryptoutilKeygen.Key] // 16-bytes A128CBC + 16-bytes HS256 (half of 32-bytes)
	hmac512KeyGenPool     *pool.ValueGenPool[cryptoutilKeygen.Key] // 64-bytes HS512
	hmac384KeyGenPool     *pool.ValueGenPool[cryptoutilKeygen.Key] // 48-bytes HS384
	hmac256KeyGenPool     *pool.ValueGenPool[cryptoutilKeygen.Key] // 32-bytes HS256
	uuidV7KeyGenPool      *pool.ValueGenPool[cryptoutilKeygen.Key]
}

func NewJwkGenService(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService) (*JwkGenService, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context must be non-nil")
	} else if telemetryService == nil {
		return nil, fmt.Errorf("telemetry service must be non-nil")
	}
	rsa4096KeyGenPoolConfig, err1 := pool.NewValueGenPoolConfig(ctx, telemetryService, "RSA 4096", 9, 9, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, cryptoutilKeygen.GenerateRSAKeyPairFunction(4096))
	rsa3072KeyGenPoolConfig, err2 := pool.NewValueGenPoolConfig(ctx, telemetryService, "RSA 3072", 6, 6, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, cryptoutilKeygen.GenerateRSAKeyPairFunction(3072))
	rsa2048KeyGenPoolConfig, err3 := pool.NewValueGenPoolConfig(ctx, telemetryService, "RSA 2048", 3, 3, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, cryptoutilKeygen.GenerateRSAKeyPairFunction(2048))
	ecdsaP521KeyGenPoolConfig, err4 := pool.NewValueGenPoolConfig(ctx, telemetryService, "Service ECDSA-P521", 3, 9, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, cryptoutilKeygen.GenerateECDSAKeyPairFunction(elliptic.P521()))
	ecdsaP384KeyGenPoolConfig, err5 := pool.NewValueGenPoolConfig(ctx, telemetryService, "Service ECDSA-P384", 2, 6, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, cryptoutilKeygen.GenerateECDSAKeyPairFunction(elliptic.P384()))
	ecdsaP256KeyGenPoolConfig, err6 := pool.NewValueGenPoolConfig(ctx, telemetryService, "Service ECDSA-P256", 1, 3, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, cryptoutilKeygen.GenerateECDSAKeyPairFunction(elliptic.P256()))
	ecdhP521KeyGenPoolConfig, err7 := pool.NewValueGenPoolConfig(ctx, telemetryService, "Service ECDH-P521", 3, 9, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, cryptoutilKeygen.GenerateECDHKeyPairFunction(ecdh.P521()))
	ecdhP384KeyGenPoolConfig, err8 := pool.NewValueGenPoolConfig(ctx, telemetryService, "Service ECSH-P384", 2, 6, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, cryptoutilKeygen.GenerateECDHKeyPairFunction(ecdh.P384()))
	ecdhP256KeyGenPoolConfig, err9 := pool.NewValueGenPoolConfig(ctx, telemetryService, "Service ECDH-P256", 1, 3, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, cryptoutilKeygen.GenerateECDHKeyPairFunction(ecdh.P256()))
	ed25519KeyGenPoolConfig, err10 := pool.NewValueGenPoolConfig(ctx, telemetryService, "Service Ed25519", 1, 2, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, cryptoutilKeygen.GenerateEDDSAKeyPairFunction("Ed25519"))
	aes256KeyGenPoolConfig, err11 := pool.NewValueGenPoolConfig(ctx, telemetryService, "Service AES-256-GCM", 3, 9, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESKeyFunction(256))
	aes192KeyGenPoolConfig, err12 := pool.NewValueGenPoolConfig(ctx, telemetryService, "Service AES-192-GCM", 2, 6, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESKeyFunction(192))
	aes128KeyGenPoolConfig, err13 := pool.NewValueGenPoolConfig(ctx, telemetryService, "Service AES-128-GCM", 1, 3, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESKeyFunction(128))
	aes256HS512KeyGenPoolConfig, err14 := pool.NewValueGenPoolConfig(ctx, telemetryService, "Service AES-256-CBC HS-512", 3, 9, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESHSKeyFunction(512))
	aes192HS384KeyGenPoolConfig, err15 := pool.NewValueGenPoolConfig(ctx, telemetryService, "Service AES-192-CBC HS-384", 2, 6, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESHSKeyFunction(384))
	aes128HS256KeyGenPoolConfig, err16 := pool.NewValueGenPoolConfig(ctx, telemetryService, "Service AES-128-CBC HS-256", 1, 3, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESHSKeyFunction(256))
	hmac512KeyGenPoolConfig, err17 := pool.NewValueGenPoolConfig(ctx, telemetryService, "Service AES-256-CBC HS-512", 3, 9, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, cryptoutilKeygen.GenerateHMACKeyFunction(512))
	hmac384KeyGenPoolConfig, err18 := pool.NewValueGenPoolConfig(ctx, telemetryService, "Service AES-192-CBC HS-384", 2, 6, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, cryptoutilKeygen.GenerateHMACKeyFunction(384))
	hmac256KeyGenPoolConfig, err19 := pool.NewValueGenPoolConfig(ctx, telemetryService, "Service AES-128-CBC HS-256", 1, 3, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, cryptoutilKeygen.GenerateHMACKeyFunction(256))
	uuidV7KeyGenPoolConfig, err20 := pool.NewValueGenPoolConfig(ctx, telemetryService, "Service UUIDv7", 2, 20, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, cryptoutilKeygen.GenerateUUIDv7Function())
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil || err7 != nil || err8 != nil || err9 != nil || err10 != nil || err11 != nil || err12 != nil || err13 != nil || err14 != nil || err15 != nil || err16 != nil || err17 != nil || err18 != nil || err19 != nil || err20 != nil {
		return nil, fmt.Errorf("failed to create pool configs: %w", errors.Join(err1, err2, err3, err4, err5, err6, err7, err8, err9, err10, err11, err12, err13, err14, err15, err16, err17, err18, err19, err20))
	}

	rsa4096KeyGenPool, err1 := pool.NewValueGenPool(rsa4096KeyGenPoolConfig)
	rsa3072KeyGenPool, err2 := pool.NewValueGenPool(rsa3072KeyGenPoolConfig)
	rsa2048KeyGenPool, err3 := pool.NewValueGenPool(rsa2048KeyGenPoolConfig)
	ecdsaP521KeyGenPool, err4 := pool.NewValueGenPool(ecdsaP521KeyGenPoolConfig)
	ecdsaP384KeyGenPool, err5 := pool.NewValueGenPool(ecdsaP384KeyGenPoolConfig)
	ecdsaP256KeyGenPool, err6 := pool.NewValueGenPool(ecdsaP256KeyGenPoolConfig)
	ecdhP521KeyGenPool, err7 := pool.NewValueGenPool(ecdhP521KeyGenPoolConfig)
	ecdhP384KeyGenPool, err8 := pool.NewValueGenPool(ecdhP384KeyGenPoolConfig)
	ecdhP256KeyGenPool, err9 := pool.NewValueGenPool(ecdhP256KeyGenPoolConfig)
	ed25519KeyGenPool, err10 := pool.NewValueGenPool(ed25519KeyGenPoolConfig)
	aes256KeyGenPool, err11 := pool.NewValueGenPool(aes256KeyGenPoolConfig)
	aes192KeyGenPool, err12 := pool.NewValueGenPool(aes192KeyGenPoolConfig)
	aes128KeyGenPool, err13 := pool.NewValueGenPool(aes128KeyGenPoolConfig)
	aes256HS512KeyGenPool, err14 := pool.NewValueGenPool(aes256HS512KeyGenPoolConfig)
	aes192HS384KeyGenPool, err15 := pool.NewValueGenPool(aes192HS384KeyGenPoolConfig)
	aes128HS256KeyGenPool, err16 := pool.NewValueGenPool(aes128HS256KeyGenPoolConfig)
	hmac512KeyGenPool, err17 := pool.NewValueGenPool(hmac512KeyGenPoolConfig)
	hmac384KeyGenPool, err18 := pool.NewValueGenPool(hmac384KeyGenPoolConfig)
	hmac256KeyGenPool, err19 := pool.NewValueGenPool(hmac256KeyGenPoolConfig)
	uuidV7KeyGenPool, err20 := pool.NewValueGenPool(uuidV7KeyGenPoolConfig)
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil || err7 != nil || err8 != nil || err9 != nil || err10 != nil || err11 != nil || err12 != nil || err13 != nil || err14 != nil || err15 != nil || err16 != nil || err17 != nil || err18 != nil || err19 != nil || err20 != nil {
		return nil, fmt.Errorf("failed to create pools: %w", errors.Join(err1, err2, err3, err4, err5, err6, err7, err8, err9, err10, err11, err12, err13, err14, err15, err16, err17, err18, err19, err20))
	}

	return &JwkGenService{
		telemetryService:      telemetryService,
		rsa4096KeyGenPool:     rsa4096KeyGenPool,
		rsa3072KeyGenPool:     rsa3072KeyGenPool,
		rsa2048KeyGenPool:     rsa2048KeyGenPool,
		ecdsaP521KeyGenPool:   ecdsaP521KeyGenPool,
		ecdsaP384KeyGenPool:   ecdsaP384KeyGenPool,
		ecdsaP256KeyGenPool:   ecdsaP256KeyGenPool,
		ecdhP521KeyGenPool:    ecdhP521KeyGenPool,
		ecdhP384KeyGenPool:    ecdhP384KeyGenPool,
		ecdhP256KeyGenPool:    ecdhP256KeyGenPool,
		ed25519KeyGenPool:     ed25519KeyGenPool,
		aes256KeyGenPool:      aes256KeyGenPool,
		aes192KeyGenPool:      aes192KeyGenPool,
		aes128KeyGenPool:      aes128KeyGenPool,
		aes256HS512KeyGenPool: aes256HS512KeyGenPool,
		aes192HS384KeyGenPool: aes192HS384KeyGenPool,
		aes128HS256KeyGenPool: aes128HS256KeyGenPool,
		hmac512KeyGenPool:     hmac512KeyGenPool,
		hmac384KeyGenPool:     hmac384KeyGenPool,
		hmac256KeyGenPool:     hmac256KeyGenPool,
		uuidV7KeyGenPool:      uuidV7KeyGenPool,
	}, nil
}

func (s *JwkGenService) Shutdown() {
	s.telemetryService.Slogger.Debug("stopping JwkGenService")
	if s.rsa4096KeyGenPool != nil {
		s.rsa4096KeyGenPool.Close()
	}
	if s.rsa3072KeyGenPool != nil {
		s.rsa3072KeyGenPool.Close()
	}
	if s.rsa2048KeyGenPool != nil {
		s.rsa2048KeyGenPool.Close()
	}
	if s.ecdsaP521KeyGenPool != nil {
		s.ecdsaP521KeyGenPool.Close()
	}
	if s.ecdsaP384KeyGenPool != nil {
		s.ecdsaP384KeyGenPool.Close()
	}
	if s.ecdsaP256KeyGenPool != nil {
		s.ecdsaP256KeyGenPool.Close()
	}
	if s.ecdhP521KeyGenPool != nil {
		s.ecdhP521KeyGenPool.Close()
	}
	if s.ecdhP384KeyGenPool != nil {
		s.ecdhP384KeyGenPool.Close()
	}
	if s.ecdhP256KeyGenPool != nil {
		s.ecdhP256KeyGenPool.Close()
	}
	if s.ed25519KeyGenPool != nil {
		s.ed25519KeyGenPool.Close()
	}
	if s.aes256KeyGenPool != nil {
		s.aes256KeyGenPool.Close()
	}
	if s.aes192KeyGenPool != nil {
		s.aes192KeyGenPool.Close()
	}
	if s.aes128KeyGenPool != nil {
		s.aes128KeyGenPool.Close()
	}
	if s.aes256HS512KeyGenPool != nil {
		s.aes256HS512KeyGenPool.Close()
	}
	if s.aes192HS384KeyGenPool != nil {
		s.aes192HS384KeyGenPool.Close()
	}
	if s.aes128HS256KeyGenPool != nil {
		s.aes128HS256KeyGenPool.Close()
	}
	if s.hmac512KeyGenPool != nil {
		s.hmac512KeyGenPool.Close()
	}
	if s.hmac384KeyGenPool != nil {
		s.hmac384KeyGenPool.Close()
	}
	if s.hmac256KeyGenPool != nil {
		s.hmac256KeyGenPool.Close()
	}
	if s.uuidV7KeyGenPool != nil {
		s.uuidV7KeyGenPool.Close()
	}
}

// func (s *JwkGenService) GenerateJweJwkFunction(enc *joseJwa.ContentEncryptionAlgorithm, alg *joseJwa.KeyEncryptionAlgorithm) func() (*googleUuid.UUID, joseJwk.Key, []byte, error) {
// 	return func() (*googleUuid.UUID, joseJwk.Key, []byte, error) { return s.GenerateJweJwk(enc, alg) }
// }

func (s *JwkGenService) GenerateJweJwk(enc *joseJwa.ContentEncryptionAlgorithm, alg *joseJwa.KeyEncryptionAlgorithm) (*googleUuid.UUID, joseJwk.Key, []byte, error) {
	switch *alg {
	case AlgDir:
		switch *enc {
		case EncA256GCM:
			return GenerateJweJwkFromKeyPool(enc, alg, s.uuidV7KeyGenPool, s.aes256KeyGenPool)
		case EncA192GCM:
			return GenerateJweJwkFromKeyPool(enc, alg, s.uuidV7KeyGenPool, s.aes192KeyGenPool)
		case EncA128GCM:
			return GenerateJweJwkFromKeyPool(enc, alg, s.uuidV7KeyGenPool, s.aes128KeyGenPool)
		case EncA256CBC_HS512:
			return GenerateJweJwkFromKeyPool(enc, alg, s.uuidV7KeyGenPool, s.aes256HS512KeyGenPool)
		case EncA192CBC_HS384:
			return GenerateJweJwkFromKeyPool(enc, alg, s.uuidV7KeyGenPool, s.aes192HS384KeyGenPool)
		case EncA128CBC_HS256:
			return GenerateJweJwkFromKeyPool(enc, alg, s.uuidV7KeyGenPool, s.aes128HS256KeyGenPool)
		default:
			return nil, nil, nil, fmt.Errorf("unsupported JWE JWK enc %s", *enc)
		}

	case AlgA256KW, AlgA256GCMKW:
		return GenerateJweJwkFromKeyPool(enc, alg, s.uuidV7KeyGenPool, s.aes256KeyGenPool)
	case AlgA192KW, AlgA192GCMKW:
		return GenerateJweJwkFromKeyPool(enc, alg, s.uuidV7KeyGenPool, s.aes192KeyGenPool)
	case AlgA128KW, AlgA128GCMKW:
		return GenerateJweJwkFromKeyPool(enc, alg, s.uuidV7KeyGenPool, s.aes128KeyGenPool)

	case AlgRSAOAEP512:
		return GenerateJweJwkFromKeyPool(enc, alg, s.uuidV7KeyGenPool, s.rsa4096KeyGenPool)
	case AlgRSAOAEP384:
		return GenerateJweJwkFromKeyPool(enc, alg, s.uuidV7KeyGenPool, s.rsa3072KeyGenPool)
	case AlgRSAOAEP256, AlgRSA15, AlgRSAOAEP:
		return GenerateJweJwkFromKeyPool(enc, alg, s.uuidV7KeyGenPool, s.rsa2048KeyGenPool)

	case AlgECDHES, AlgECDHESA256KW:
		return GenerateJweJwkFromKeyPool(enc, alg, s.uuidV7KeyGenPool, s.ecdhP521KeyGenPool)
	case AlgECDHESA192KW:
		return GenerateJweJwkFromKeyPool(enc, alg, s.uuidV7KeyGenPool, s.ecdhP384KeyGenPool)
	case AlgECDHESA128KW:
		return GenerateJweJwkFromKeyPool(enc, alg, s.uuidV7KeyGenPool, s.ecdhP256KeyGenPool)

	default:
		return nil, nil, nil, fmt.Errorf("unsupported JWE JWK alg %s", *alg)
	}
}

func (s *JwkGenService) GenerateJwsJwk(alg *joseJwa.SignatureAlgorithm) (*googleUuid.UUID, joseJwk.Key, []byte, error) {
	switch *alg {
	case AlgRS512, AlgPS512:
		return GenerateJwsJwkFromKeyPool(alg, s.uuidV7KeyGenPool, s.rsa4096KeyGenPool)
	case AlgRS384, AlgPS384:
		return GenerateJwsJwkFromKeyPool(alg, s.uuidV7KeyGenPool, s.rsa3072KeyGenPool)
	case AlgRS256, AlgPS256:
		return GenerateJwsJwkFromKeyPool(alg, s.uuidV7KeyGenPool, s.rsa2048KeyGenPool)
	case AlgES512:
		return GenerateJwsJwkFromKeyPool(alg, s.uuidV7KeyGenPool, s.ecdsaP521KeyGenPool)
	case AlgES384:
		return GenerateJwsJwkFromKeyPool(alg, s.uuidV7KeyGenPool, s.ecdsaP384KeyGenPool)
	case AlgES256:
		return GenerateJwsJwkFromKeyPool(alg, s.uuidV7KeyGenPool, s.ecdsaP256KeyGenPool)
	case AlgEdDSA:
		return GenerateJwsJwkFromKeyPool(alg, s.uuidV7KeyGenPool, s.ed25519KeyGenPool)
	case AlgHS512:
		return GenerateJwsJwkFromKeyPool(alg, s.uuidV7KeyGenPool, s.hmac512KeyGenPool)
	case AlgHS384:
		return GenerateJwsJwkFromKeyPool(alg, s.uuidV7KeyGenPool, s.hmac384KeyGenPool)
	case AlgHS256:
		return GenerateJwsJwkFromKeyPool(alg, s.uuidV7KeyGenPool, s.hmac256KeyGenPool)
	default:
		return nil, nil, nil, fmt.Errorf("unsupported JWS JWK alg: %s", alg)
	}
}
