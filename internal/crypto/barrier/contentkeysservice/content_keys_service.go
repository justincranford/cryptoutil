package contentkeysservice

import (
	"fmt"

	cryptoutilIntermediateKeysService "cryptoutil/internal/crypto/barrier/intermediatekeysservice"
	cryptoutilJose "cryptoutil/internal/crypto/jose"
	cryptoutilKeygen "cryptoutil/internal/crypto/keygen"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"

	googleUuid "github.com/google/uuid"

	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

type ContentKeysService struct {
	telemetryService        *cryptoutilTelemetry.TelemetryService
	ormRepository           *cryptoutilOrmRepository.OrmRepository
	intermediateKeysService *cryptoutilIntermediateKeysService.IntermediateKeysService
	aes256KeyGenPool        *cryptoutilKeygen.KeyGenPool
}

func NewContentKeysService(telemetryService *cryptoutilTelemetry.TelemetryService, ormRepository *cryptoutilOrmRepository.OrmRepository, intermediateKeysService *cryptoutilIntermediateKeysService.IntermediateKeysService, aes256KeyGenPool *cryptoutilKeygen.KeyGenPool) (*ContentKeysService, error) {
	if telemetryService == nil {
		return nil, fmt.Errorf("telemetryService must be non-nil")
	} else if ormRepository == nil {
		return nil, fmt.Errorf("ormRepository must be non-nil")
	} else if intermediateKeysService == nil {
		return nil, fmt.Errorf("intermediateKeysService must be non-nil")
	} else if aes256KeyGenPool == nil {
		return nil, fmt.Errorf("aes256KeyGenPool must be non-nil")
	}
	return &ContentKeysService{telemetryService: telemetryService, ormRepository: ormRepository, intermediateKeysService: intermediateKeysService, aes256KeyGenPool: aes256KeyGenPool}, nil
}

func (s *ContentKeysService) EncryptContent(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, clearBytes []byte) ([]byte, error) {
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	clearContentKeyBytes, ok := s.aes256KeyGenPool.Get().Private.([]byte)
	if !ok {
		return nil, fmt.Errorf("failed to cast AES-256 pool key to []byte")
	}
	clearContentKey, _, contentKeyKidUUID, err := cryptoutilJose.CreateAesJWK(cryptoutilJose.AlgDIRECT, clearContentKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content JWK: %w", err)
	}
	_, encryptedContentJweMessageBytes, err := cryptoutilJose.EncryptBytes([]joseJwk.Key{clearContentKey}, clearBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt content with JWK")
	}
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	clearIntermediateKeyLatest, err := s.intermediateKeysService.GetLatest(sqlTransaction)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest clear intermediate key")
	}
	clearIntermediateKeyLatestKidUuid, err := cryptoutilJose.ExtractKidUuid(clearIntermediateKeyLatest)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest clear intermediate key kid uuid: %w", err)
	}
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	_, encryptedContentKeyJweMessageBytes, err := cryptoutilJose.EncryptKey([]joseJwk.Key{clearIntermediateKeyLatest}, clearContentKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt content key: %w", err)
	}
	err = sqlTransaction.AddContentKey(&cryptoutilOrmRepository.BarrierContentKey{UUID: contentKeyKidUUID, Encrypted: string(encryptedContentKeyJweMessageBytes), KEKUUID: *clearIntermediateKeyLatestKidUuid})
	if err != nil {
		return nil, fmt.Errorf("failed to add content key to DB: %w", err)
	}
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	return encryptedContentJweMessageBytes, nil
}

func (s *ContentKeysService) DecryptContent(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, encryptedContentJweMessageBytes []byte) ([]byte, error) {
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	encryptedContentJweMessage, err := joseJwe.Parse(encryptedContentJweMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWE message: %w", err)
	}
	var encryptedContentJweMessageKidString string
	err = encryptedContentJweMessage.ProtectedHeaders().Get(joseJwk.KeyIDKey, &encryptedContentJweMessageKidString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWE message kid: %w", err)
	}
	encryptedContentJweMessageKidUuid, err := googleUuid.Parse(encryptedContentJweMessageKidString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kid as uuid: %w", err)
	}
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	encryptedContentKey, err := sqlTransaction.GetContentKey(encryptedContentJweMessageKidUuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get encrypted content key")
	}
	decryptedIntermediateKey, err := s.intermediateKeysService.Get(sqlTransaction, encryptedContentKey.GetKEKUUID())
	if err != nil {
		return nil, fmt.Errorf("failed to get intermediate key")
	}
	decryptedContentKey, err := cryptoutilJose.DecryptKey([]joseJwk.Key{decryptedIntermediateKey}, []byte(encryptedContentKey.GetEncrypted()))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt root key")
	}
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	decryptedBytes, err := cryptoutilJose.DecryptBytes([]joseJwk.Key{decryptedContentKey}, encryptedContentJweMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt content with content key: %w", err)
	}
	return decryptedBytes, nil
}

func (s *ContentKeysService) Shutdown() {
	s.telemetryService = nil
	s.ormRepository = nil
	s.intermediateKeysService = nil
	s.aes256KeyGenPool = nil
}
