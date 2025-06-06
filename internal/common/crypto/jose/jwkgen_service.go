package jose

import (
	"context"
	"crypto/ecdh"
	"crypto/elliptic"
	"errors"
	"fmt"

	cryptoutilKeygen "cryptoutil/internal/common/crypto/keygen"
	cryptoutilPool "cryptoutil/internal/common/pool"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilUtil "cryptoutil/internal/common/util"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

type JwkGenService struct {
	telemetryService      *cryptoutilTelemetry.TelemetryService
	rsa4096KeyGenPool     *cryptoutilPool.ValueGenPool[*cryptoutilKeygen.KeyPair]  // 512-bytes
	rsa3072KeyGenPool     *cryptoutilPool.ValueGenPool[*cryptoutilKeygen.KeyPair]  // 384-bytes
	rsa2048KeyGenPool     *cryptoutilPool.ValueGenPool[*cryptoutilKeygen.KeyPair]  // 256-bytes
	ecdsaP521KeyGenPool   *cryptoutilPool.ValueGenPool[*cryptoutilKeygen.KeyPair]  // 65.125-bytes
	ecdsaP384KeyGenPool   *cryptoutilPool.ValueGenPool[*cryptoutilKeygen.KeyPair]  // 48-bytes
	ecdsaP256KeyGenPool   *cryptoutilPool.ValueGenPool[*cryptoutilKeygen.KeyPair]  // 32-bytes
	ecdhP521KeyGenPool    *cryptoutilPool.ValueGenPool[*cryptoutilKeygen.KeyPair]  // 65.125-bytes
	ecdhP384KeyGenPool    *cryptoutilPool.ValueGenPool[*cryptoutilKeygen.KeyPair]  // 48-bytes
	ecdhP256KeyGenPool    *cryptoutilPool.ValueGenPool[*cryptoutilKeygen.KeyPair]  // 32-bytes
	ed25519KeyGenPool     *cryptoutilPool.ValueGenPool[*cryptoutilKeygen.KeyPair]  // 32-bytes
	aes256KeyGenPool      *cryptoutilPool.ValueGenPool[cryptoutilKeygen.SecretKey] // 32-bytes A256GCM, A256KW, A256GCMKW
	aes192KeyGenPool      *cryptoutilPool.ValueGenPool[cryptoutilKeygen.SecretKey] // 24-bytes A192GCM, A192KW, A192GCMKW
	aes128KeyGenPool      *cryptoutilPool.ValueGenPool[cryptoutilKeygen.SecretKey] // 16-bytes A128GCM, A128KW, A128GCMKW
	aes256HS512KeyGenPool *cryptoutilPool.ValueGenPool[cryptoutilKeygen.SecretKey] // 32-bytes A256CBC + 32-bytes HS512 (half of 64-bytes)
	aes192HS384KeyGenPool *cryptoutilPool.ValueGenPool[cryptoutilKeygen.SecretKey] // 24-bytes A192CBC + 24-bytes HS384 (half of 48-bytes)
	aes128HS256KeyGenPool *cryptoutilPool.ValueGenPool[cryptoutilKeygen.SecretKey] // 16-bytes A128CBC + 16-bytes HS256 (half of 32-bytes)
	hmac512KeyGenPool     *cryptoutilPool.ValueGenPool[cryptoutilKeygen.SecretKey] // 64-bytes HS512
	hmac384KeyGenPool     *cryptoutilPool.ValueGenPool[cryptoutilKeygen.SecretKey] // 48-bytes HS384
	hmac256KeyGenPool     *cryptoutilPool.ValueGenPool[cryptoutilKeygen.SecretKey] // 32-bytes HS256
	uuidV7KeyGenPool      *cryptoutilPool.ValueGenPool[*googleUuid.UUID]
}

func NewJwkGenService(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService) (*JwkGenService, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context must be non-nil")
	} else if telemetryService == nil {
		return nil, fmt.Errorf("telemetry service must be non-nil")
	}
	rsa4096KeyGenPool, err1 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JwkGenService RSA 4096", 9, 9, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateRSAKeyPairFunction(4096)))
	rsa3072KeyGenPool, err2 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JwkGenService RSA 3072", 6, 6, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateRSAKeyPairFunction(3072)))
	rsa2048KeyGenPool, err3 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JwkGenService RSA 2048", 3, 3, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateRSAKeyPairFunction(2048)))
	ecdsaP521KeyGenPool, err4 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JwkGenService ECDSA-P521", 3, 9, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateECDSAKeyPairFunction(elliptic.P521())))
	ecdsaP384KeyGenPool, err5 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JwkGenService ECDSA-P384", 2, 6, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateECDSAKeyPairFunction(elliptic.P384())))
	ecdsaP256KeyGenPool, err6 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JwkGenService ECDSA-P256", 1, 3, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateECDSAKeyPairFunction(elliptic.P256())))
	ecdhP521KeyGenPool, err7 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JwkGenService ECDH-P521", 3, 9, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateECDHKeyPairFunction(ecdh.P521())))
	ecdhP384KeyGenPool, err8 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JwkGenService ECSH-P384", 2, 6, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateECDHKeyPairFunction(ecdh.P384())))
	ecdhP256KeyGenPool, err9 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JwkGenService ECDH-P256", 1, 3, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateECDHKeyPairFunction(ecdh.P256())))
	ed25519KeyGenPool, err10 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JwkGenService Ed25519", 1, 2, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateEDDSAKeyPairFunction("Ed25519")))
	aes256KeyGenPool, err11 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JwkGenService AES-256-GCM", 3, 9, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESKeyFunction(256)))
	aes192KeyGenPool, err12 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JwkGenService AES-192-GCM", 2, 6, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESKeyFunction(192)))
	aes128KeyGenPool, err13 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JwkGenService AES-128-GCM", 1, 3, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESKeyFunction(128)))
	aes256HS512KeyGenPool, err14 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JwkGenService AES-256-CBC HS-512", 3, 9, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESHSKeyFunction(512)))
	aes192HS384KeyGenPool, err15 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JwkGenService AES-192-CBC HS-384", 2, 6, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESHSKeyFunction(384)))
	aes128HS256KeyGenPool, err16 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JwkGenService AES-128-CBC HS-256", 1, 3, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESHSKeyFunction(256)))
	hmac512KeyGenPool, err17 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JwkGenService HMAC-512", 3, 9, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateHMACKeyFunction(512)))
	hmac384KeyGenPool, err18 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JwkGenService HMAC-384", 2, 6, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateHMACKeyFunction(384)))
	hmac256KeyGenPool, err19 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JwkGenService HMAC-256", 1, 3, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateHMACKeyFunction(256)))
	uuidV7KeyGenPool, err20 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "JwkGenService UUIDv7", 2, 20, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilUtil.GenerateUUIDv7Function()))
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
		s.rsa4096KeyGenPool.Cancel()
	}
	if s.rsa3072KeyGenPool != nil {
		s.rsa3072KeyGenPool.Cancel()
	}
	if s.rsa2048KeyGenPool != nil {
		s.rsa2048KeyGenPool.Cancel()
	}
	if s.ecdsaP521KeyGenPool != nil {
		s.ecdsaP521KeyGenPool.Cancel()
	}
	if s.ecdsaP384KeyGenPool != nil {
		s.ecdsaP384KeyGenPool.Cancel()
	}
	if s.ecdsaP256KeyGenPool != nil {
		s.ecdsaP256KeyGenPool.Cancel()
	}
	if s.ecdhP521KeyGenPool != nil {
		s.ecdhP521KeyGenPool.Cancel()
	}
	if s.ecdhP384KeyGenPool != nil {
		s.ecdhP384KeyGenPool.Cancel()
	}
	if s.ecdhP256KeyGenPool != nil {
		s.ecdhP256KeyGenPool.Cancel()
	}
	if s.ed25519KeyGenPool != nil {
		s.ed25519KeyGenPool.Cancel()
	}
	if s.aes256KeyGenPool != nil {
		s.aes256KeyGenPool.Cancel()
	}
	if s.aes192KeyGenPool != nil {
		s.aes192KeyGenPool.Cancel()
	}
	if s.aes128KeyGenPool != nil {
		s.aes128KeyGenPool.Cancel()
	}
	if s.aes256HS512KeyGenPool != nil {
		s.aes256HS512KeyGenPool.Cancel()
	}
	if s.aes192HS384KeyGenPool != nil {
		s.aes192HS384KeyGenPool.Cancel()
	}
	if s.aes128HS256KeyGenPool != nil {
		s.aes128HS256KeyGenPool.Cancel()
	}
	if s.hmac512KeyGenPool != nil {
		s.hmac512KeyGenPool.Cancel()
	}
	if s.hmac384KeyGenPool != nil {
		s.hmac384KeyGenPool.Cancel()
	}
	if s.hmac256KeyGenPool != nil {
		s.hmac256KeyGenPool.Cancel()
	}
	if s.uuidV7KeyGenPool != nil {
		s.uuidV7KeyGenPool.Cancel()
	}
}

func (s *JwkGenService) GenerateJweJwk(enc *joseJwa.ContentEncryptionAlgorithm, alg *joseJwa.KeyEncryptionAlgorithm) (*googleUuid.UUID, joseJwk.Key, []byte, error) {
	switch *alg {
	case AlgDir:
		switch *enc {
		case EncA256GCM:
			return CreateJweJwkFromKey(s.uuidV7KeyGenPool.Get(), enc, alg, s.aes256KeyGenPool.Get())
		case EncA192GCM:
			return CreateJweJwkFromKey(s.uuidV7KeyGenPool.Get(), enc, alg, s.aes192KeyGenPool.Get())
		case EncA128GCM:
			return CreateJweJwkFromKey(s.uuidV7KeyGenPool.Get(), enc, alg, s.aes128KeyGenPool.Get())
		case EncA256CBC_HS512:
			return CreateJweJwkFromKey(s.uuidV7KeyGenPool.Get(), enc, alg, s.aes256HS512KeyGenPool.Get())
		case EncA192CBC_HS384:
			return CreateJweJwkFromKey(s.uuidV7KeyGenPool.Get(), enc, alg, s.aes192HS384KeyGenPool.Get())
		case EncA128CBC_HS256:
			return CreateJweJwkFromKey(s.uuidV7KeyGenPool.Get(), enc, alg, s.aes128HS256KeyGenPool.Get())
		default:
			return nil, nil, nil, fmt.Errorf("unsupported JWE JWK enc %s", *enc)
		}

	case AlgA256KW, AlgA256GCMKW:
		return CreateJweJwkFromKey(s.uuidV7KeyGenPool.Get(), enc, alg, s.aes256KeyGenPool.Get())
	case AlgA192KW, AlgA192GCMKW:
		return CreateJweJwkFromKey(s.uuidV7KeyGenPool.Get(), enc, alg, s.aes192KeyGenPool.Get())
	case AlgA128KW, AlgA128GCMKW:
		return CreateJweJwkFromKey(s.uuidV7KeyGenPool.Get(), enc, alg, s.aes128KeyGenPool.Get())

	case AlgRSAOAEP512:
		return CreateJweJwkFromKey(s.uuidV7KeyGenPool.Get(), enc, alg, s.rsa4096KeyGenPool.Get())
	case AlgRSAOAEP384:
		return CreateJweJwkFromKey(s.uuidV7KeyGenPool.Get(), enc, alg, s.rsa3072KeyGenPool.Get())
	case AlgRSAOAEP256:
		return CreateJweJwkFromKey(s.uuidV7KeyGenPool.Get(), enc, alg, s.rsa2048KeyGenPool.Get())
	case AlgRSAOAEP:
		return CreateJweJwkFromKey(s.uuidV7KeyGenPool.Get(), enc, alg, s.rsa2048KeyGenPool.Get())
	case AlgRSA15:
		return CreateJweJwkFromKey(s.uuidV7KeyGenPool.Get(), enc, alg, s.rsa2048KeyGenPool.Get())

	case AlgECDHES, AlgECDHESA256KW:
		return CreateJweJwkFromKey(s.uuidV7KeyGenPool.Get(), enc, alg, s.ecdhP521KeyGenPool.Get())
	case AlgECDHESA192KW:
		return CreateJweJwkFromKey(s.uuidV7KeyGenPool.Get(), enc, alg, s.ecdhP384KeyGenPool.Get())
	case AlgECDHESA128KW:
		return CreateJweJwkFromKey(s.uuidV7KeyGenPool.Get(), enc, alg, s.ecdhP256KeyGenPool.Get())

	default:
		return nil, nil, nil, fmt.Errorf("unsupported JWE JWK alg %s", *alg)
	}
}

func (s *JwkGenService) GenerateJwsJwk(alg *joseJwa.SignatureAlgorithm) (*googleUuid.UUID, joseJwk.Key, []byte, error) {
	switch *alg {
	case AlgRS512, AlgPS512:
		return CreateJwsJwkFromKey(s.uuidV7KeyGenPool.Get(), alg, s.rsa4096KeyGenPool.Get())
	case AlgRS384, AlgPS384:
		return CreateJwsJwkFromKey(s.uuidV7KeyGenPool.Get(), alg, s.rsa3072KeyGenPool.Get())
	case AlgRS256, AlgPS256:
		return CreateJwsJwkFromKey(s.uuidV7KeyGenPool.Get(), alg, s.rsa2048KeyGenPool.Get())
	case AlgES512:
		return CreateJwsJwkFromKey(s.uuidV7KeyGenPool.Get(), alg, s.ecdsaP521KeyGenPool.Get())
	case AlgES384:
		return CreateJwsJwkFromKey(s.uuidV7KeyGenPool.Get(), alg, s.ecdsaP384KeyGenPool.Get())
	case AlgES256:
		return CreateJwsJwkFromKey(s.uuidV7KeyGenPool.Get(), alg, s.ecdsaP256KeyGenPool.Get())
	case AlgEdDSA:
		return CreateJwsJwkFromKey(s.uuidV7KeyGenPool.Get(), alg, s.ed25519KeyGenPool.Get())
	case AlgHS512:
		return CreateJwsJwkFromKey(s.uuidV7KeyGenPool.Get(), alg, s.hmac512KeyGenPool.Get())
	case AlgHS384:
		return CreateJwsJwkFromKey(s.uuidV7KeyGenPool.Get(), alg, s.hmac384KeyGenPool.Get())
	case AlgHS256:
		return CreateJwsJwkFromKey(s.uuidV7KeyGenPool.Get(), alg, s.hmac256KeyGenPool.Get())
	default:
		return nil, nil, nil, fmt.Errorf("unsupported JWS JWK alg: %s", alg)
	}
}

func (s *JwkGenService) GenerateUUIDv7() *googleUuid.UUID {
	return s.uuidV7KeyGenPool.Get()
}

func (s *JwkGenService) GetAes256KeyGenPool() *cryptoutilPool.ValueGenPool[cryptoutilKeygen.SecretKey] {
	return s.aes256KeyGenPool
}

func (s *JwkGenService) GetUUIDv7KeyGenPool() *cryptoutilPool.ValueGenPool[*googleUuid.UUID] {
	return s.uuidV7KeyGenPool
}
