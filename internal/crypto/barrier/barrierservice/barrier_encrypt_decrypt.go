package barrierservice

import (
	"errors"
	"fmt"

	cryptoutilBarrierRepository "cryptoutil/internal/crypto/barrier/barrierrepository"
	cryptoutilJose "cryptoutil/internal/crypto/jose"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

func (d *BarrierService) EncryptContent(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, clearBytes []byte) ([]byte, error) {
	if d.closed {
		return nil, fmt.Errorf("barrier service is closed")
	}
	rawKey, ok := d.aes256KeyGenPool.Get().Private.([]byte)
	if !ok {
		return nil, fmt.Errorf("failed to cast AES-256 pool key to []byte")
	}
	cek, _, _, err := cryptoutilJose.CreateAesJWK(cryptoutilJose.AlgDIRECT, rawKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content JWK: %w", err)
	}
	err = d.contentKeyRepository.Put(sqlTransaction, cek)
	if err != nil {
		return nil, fmt.Errorf("failed to put content JWK in cache: %w", err)
	}
	jweMessage, encodedJweMessage, err := cryptoutilJose.EncryptBytes([]joseJwk.Key{cek}, clearBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt clear bytes: %w", err)
	}
	jweHeaders, err := cryptoutilJose.JSONHeadersString(jweMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to get JWE message headers: %w", err)
	}
	d.telemetryService.Slogger.Info("Encrypted Bytes", "JWE Headers", jweHeaders)

	return encodedJweMessage, nil
}

func (d *BarrierService) DecryptContent(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, encodedJweMessage []byte) ([]byte, error) {
	if d.closed {
		return nil, fmt.Errorf("barrier service is closed")
	}
	jweMessage, err := joseJwe.Parse(encodedJweMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWE message: %w", err)
	}
	var kid string
	err = jweMessage.ProtectedHeaders().Get(joseJwk.KeyIDKey, &kid)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWE message kid: %w", err)
	}
	kidUuid, err := googleUuid.Parse(kid)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kid as uuid: %w", err)
	}
	jwk, err := d.contentKeyRepository.Get(sqlTransaction, kidUuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get key by kid as uuid: %w", err)
	}
	decryptedBytes, err := cryptoutilJose.DecryptBytes([]joseJwk.Key{jwk}, encodedJweMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt with JWK %s: %w", kid, err)
	}
	return decryptedBytes, nil
}

// Helpers

func encrypt(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, jwk joseJwk.Key, kekRepository *cryptoutilBarrierRepository.BarrierRepository, telemetryService *cryptoutilTelemetry.TelemetryService) (googleUuid.UUID, googleUuid.UUID, []byte, error) {
	jwkKidUuid, err := cryptoutilJose.ExtractKidUuid(jwk)
	if err != nil {
		return googleUuid.UUID{}, googleUuid.UUID{}, nil, fmt.Errorf("failed to get jwk kid uuid: %w", err)
	}

	kek, err := kekRepository.GetLatest(sqlTransaction)
	if err != nil {
		return googleUuid.UUID{}, googleUuid.UUID{}, nil, fmt.Errorf("failed to get latest kek jwk kid uuid: %w", err)
	}
	kekKidUuid, err := cryptoutilJose.ExtractKidUuid(kek)
	if err != nil {
		return googleUuid.UUID{}, googleUuid.UUID{}, nil, fmt.Errorf("failed to get latest kek kid uuid: %w", err)
	}

	jweMessage, jweMessageBytes, err := cryptoutilJose.EncryptKey([]joseJwk.Key{kek}, jwk)
	if err != nil {
		return googleUuid.UUID{}, googleUuid.UUID{}, nil, fmt.Errorf("failed to serialize jwk: %w", err)
	}
	jweHeaders, err := cryptoutilJose.JSONHeadersString(jweMessage)
	if err != nil {
		return googleUuid.UUID{}, googleUuid.UUID{}, nil, fmt.Errorf("failed to get jwe message headers: %w", err)
	}
	telemetryService.Slogger.Info("Encrypted Intermediate JWK", "JWE Headers", jweHeaders)

	return *jwkKidUuid, *kekKidUuid, jweMessageBytes, nil
}

func decrypt(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, kekRepository *cryptoutilBarrierRepository.BarrierRepository, barrierKey cryptoutilOrmRepository.BarrierKey, err error) (joseJwk.Key, error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to load Key from database: %w", err)
	}
	kekJwk, err := kekRepository.Get(sqlTransaction, barrierKey.GetKEKUUID())
	if err != nil {
		return nil, fmt.Errorf("failed to parse kek kid from database: %w", err)
	}
	jwk, err := cryptoutilJose.DecryptKey([]joseJwk.Key{kekJwk}, []byte((barrierKey).GetEncrypted()))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt JWK from database: %w", err)
	}
	return jwk, nil
}
