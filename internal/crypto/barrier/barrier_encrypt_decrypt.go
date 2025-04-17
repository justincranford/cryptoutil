package barrierservice

import (
	"fmt"

	cryptoutilJose "cryptoutil/internal/crypto/jose"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"

	googleUuid "github.com/google/uuid"

	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

func (d *BarrierService) EncryptContent(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, clearBytes []byte) ([]byte, error) {
	if d.closed {
		return nil, fmt.Errorf("barrier service is closed")
	}

	///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	clearContentKeyBytes, ok := d.aes256KeyGenPool.Get().Private.([]byte)
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
	clearIntermediateKeyLatest, err := d.intermediateKeysService.GetLatest(sqlTransaction)
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

func (d *BarrierService) DecryptContent(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, encryptedContentJweMessageBytes []byte) ([]byte, error) {
	if d.closed {
		return nil, fmt.Errorf("barrier service is closed")
	}

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
	decryptedIntermediateKey, err := d.intermediateKeysService.Get(sqlTransaction, encryptedContentKey.GetKEKUUID())
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
