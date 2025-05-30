package contentkeysservice

import (
	"fmt"

	cryptoutilJose "cryptoutil/internal/common/crypto/jose"
	cryptoutilKeygen "cryptoutil/internal/common/crypto/keygen"
	cryptoutilPool "cryptoutil/internal/common/pool"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilIntermediateKeysService "cryptoutil/internal/server/barrier/intermediatekeysservice"
	cryptoutilOrmRepository "cryptoutil/internal/server/repository/orm"

	googleUuid "github.com/google/uuid"

	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

type ContentKeysService struct {
	telemetryService        *cryptoutilTelemetry.TelemetryService
	ormRepository           *cryptoutilOrmRepository.OrmRepository
	intermediateKeysService *cryptoutilIntermediateKeysService.IntermediateKeysService
	uuidV7KeyGenPool        *cryptoutilPool.ValueGenPool[*googleUuid.UUID]
	aes256KeyGenPool        *cryptoutilPool.ValueGenPool[cryptoutilKeygen.SecretKey]
}

func NewContentKeysService(telemetryService *cryptoutilTelemetry.TelemetryService, ormRepository *cryptoutilOrmRepository.OrmRepository, intermediateKeysService *cryptoutilIntermediateKeysService.IntermediateKeysService, uuidV7KeyGenPool *cryptoutilPool.ValueGenPool[*googleUuid.UUID], aes256KeyGenPool *cryptoutilPool.ValueGenPool[cryptoutilKeygen.SecretKey]) (*ContentKeysService, error) {
	if telemetryService == nil {
		return nil, fmt.Errorf("telemetryService must be non-nil")
	} else if ormRepository == nil {
		return nil, fmt.Errorf("ormRepository must be non-nil")
	} else if intermediateKeysService == nil {
		return nil, fmt.Errorf("intermediateKeysService must be non-nil")
	} else if uuidV7KeyGenPool == nil {
		return nil, fmt.Errorf("uuidV7KeyGenPool must be non-nil")
	} else if aes256KeyGenPool == nil {
		return nil, fmt.Errorf("aes256KeyGenPool must be non-nil")
	}
	return &ContentKeysService{telemetryService: telemetryService, ormRepository: ormRepository, intermediateKeysService: intermediateKeysService, uuidV7KeyGenPool: uuidV7KeyGenPool, aes256KeyGenPool: aes256KeyGenPool}, nil
}

func (s *ContentKeysService) EncryptContent(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, clearContentBytes []byte) ([]byte, *googleUuid.UUID, error) {
	contentKeyKidUuid, clearContentKey, _, err := cryptoutilJose.GenerateJweJwkFromSecretKeyPool(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgDir, s.uuidV7KeyGenPool, s.aes256KeyGenPool)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate content JWK: %w", err)
	}
	_, encryptedContentJweMessageBytes, err := cryptoutilJose.EncryptBytes([]joseJwk.Key{clearContentKey}, clearContentBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt content with JWK: %w", err)
	}
	encryptedContentKeyJweMessageBytes, intermediateKeyKidUuid, err := s.intermediateKeysService.EncryptKey(sqlTransaction, clearContentKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt content JWK with intermediate JWK: %w", err)
	}
	err = sqlTransaction.AddContentKey(&cryptoutilOrmRepository.BarrierContentKey{UUID: *contentKeyKidUuid, Encrypted: string(encryptedContentKeyJweMessageBytes), KEKUUID: *intermediateKeyKidUuid})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to add content key to DB: %w", err)
	}
	return encryptedContentJweMessageBytes, contentKeyKidUuid, nil
}

func (s *ContentKeysService) DecryptContent(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, encryptedContentJweMessageBytes []byte) ([]byte, error) {
	encryptedContentJweMessage, err := joseJwe.Parse(encryptedContentJweMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWE message: %w", err)
	}
	var encryptedContentKeyKidString string
	err = encryptedContentJweMessage.ProtectedHeaders().Get(joseJwk.KeyIDKey, &encryptedContentKeyKidString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWE message kid: %w", err)
	}
	encryptedContentKeyKidUuid, err := googleUuid.Parse(encryptedContentKeyKidString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kid as uuid: %w", err)
	}
	encryptedContentKey, err := sqlTransaction.GetContentKey(encryptedContentKeyKidUuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get encrypted content key: %w", err)
	}
	decryptedContentKey, err := s.intermediateKeysService.DecryptKey(sqlTransaction, []byte(encryptedContentKey.GetEncrypted()))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt content key: %w", err)
	}
	decryptedBytes, err := cryptoutilJose.DecryptBytes([]joseJwk.Key{decryptedContentKey}, encryptedContentJweMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt content with content key: %w", err)
	}
	return decryptedBytes, nil
}

func (s *ContentKeysService) Shutdown() {
	s.aes256KeyGenPool = nil
	s.uuidV7KeyGenPool = nil
	s.telemetryService = nil
	s.ormRepository = nil
	s.intermediateKeysService = nil
}
