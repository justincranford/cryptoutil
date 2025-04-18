package rootkeysservice

import (
	"context"
	"errors"
	"fmt"

	cryptoutilUnsealRepository "cryptoutil/internal/crypto/barrier/unsealrepository"
	cryptoutilJose "cryptoutil/internal/crypto/jose"
	cryptoutilKeygen "cryptoutil/internal/crypto/keygen"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"

	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

type RootKeysService struct {
	aes256KeyGenPool      *cryptoutilKeygen.KeyGenPool
	unsealedRootJwksList  []joseJwk.Key
	unsealedRootJwksMap   *map[googleUuid.UUID]joseJwk.Key
	unsealedRootJwkLatest joseJwk.Key
}

func (u *RootKeysService) GetAll() []joseJwk.Key {
	return u.unsealedRootJwksList
}

func (u *RootKeysService) Get(uuid googleUuid.UUID) joseJwk.Key {
	return (*u.unsealedRootJwksMap)[uuid]
}

func (u *RootKeysService) GetLatest() joseJwk.Key {
	return u.unsealedRootJwkLatest
}

func (u *RootKeysService) Shutdown() {
	u.unsealedRootJwksList = nil
	u.unsealedRootJwksMap = nil
	u.unsealedRootJwkLatest = nil
}

func NewRootKeysService(telemetryService *cryptoutilTelemetry.TelemetryService, ormRepository *cryptoutilOrmRepository.OrmRepository, unsealRepository cryptoutilUnsealRepository.UnsealRepository, aes256KeyGenPool *cryptoutilKeygen.KeyGenPool) (*RootKeysService, error) {
	if telemetryService == nil {
		return nil, fmt.Errorf("telemetryService must be non-nil")
	} else if ormRepository == nil {
		return nil, fmt.Errorf("ormRepository must be non-nil")
	} else if unsealRepository == nil {
		return nil, fmt.Errorf("unsealRepository must be non-nil")
	} else if aes256KeyGenPool == nil {
		return nil, fmt.Errorf("aes256KeyGenPool must be non-nil")
	}

	unsealJwks := unsealRepository.UnsealJwks() // unseal keys from unseal repository
	if len(unsealJwks) == 0 {
		return nil, fmt.Errorf("no unseal JWKs")
	}

	var encryptedRootJwks []cryptoutilOrmRepository.BarrierRootKey
	err := ormRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		encryptedRootJwks, err = sqlTransaction.GetRootKeys() // encrypted root JWKs from DB
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get encrypted root JWKs from DB")
	}

	var rootKeysService *RootKeysService
	if len(encryptedRootJwks) == 0 {
		rootKeysService, err = createFirstRootJwk(telemetryService, ormRepository, aes256KeyGenPool, unsealJwks)
	} else {
		rootKeysService, err = decryptExistingRootJwks(telemetryService, ormRepository, aes256KeyGenPool, unsealJwks, encryptedRootJwks)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to initialize unseal service: %w", err)
	}

	return rootKeysService, nil
}

func createFirstRootJwk(telemetryService *cryptoutilTelemetry.TelemetryService, ormRepository *cryptoutilOrmRepository.OrmRepository, aes256KeyGenPool *cryptoutilKeygen.KeyGenPool, unsealJwks []joseJwk.Key) (*RootKeysService, error) {
	unsealedRootJwksLatest, _, unsealedRootJwksLatestKidUuid, err := cryptoutilJose.GenerateAesJWK(cryptoutilJose.AlgDIRECT)
	if err != nil {
		return nil, fmt.Errorf("failed to generate root JWK: %w", err)
	}

	jweMessage, jweMessageBytes, err := cryptoutilJose.EncryptKey(unsealJwks, unsealedRootJwksLatest)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt root JWK: %w", err)
	}
	jweHeaders, err := cryptoutilJose.JSONHeadersString(jweMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to get JWE message headers for encrypt root JWK: %w", err)
	}
	telemetryService.Slogger.Info("Encrypted Root JWK with Unseal JWK", "JWE Headers", jweHeaders)

	sealJwkKidUuid, err := cryptoutilJose.ExtractKidUuid(unsealJwks[0])
	if err != nil {
		return nil, fmt.Errorf("failed to get seal JWK kid uuid: %w", err)
	}

	// put new, encrypted root JWK in DB
	err = ormRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		return sqlTransaction.AddRootKey(&cryptoutilOrmRepository.BarrierRootKey{UUID: unsealedRootJwksLatestKidUuid, Encrypted: string(jweMessageBytes), KEKUUID: *sealJwkKidUuid})
	})
	if err != nil {
		return nil, fmt.Errorf("failed to store root JWK: %w", err)
	}

	// put new, clear root JWK in-memory
	unsealedRootKeys := []joseJwk.Key{unsealedRootJwksLatest}
	unsealedRootJwksMap := make(map[googleUuid.UUID]joseJwk.Key)
	unsealedRootJwksMap[unsealedRootJwksLatestKidUuid] = unsealedRootJwksLatest

	return &RootKeysService{unsealedRootJwksList: unsealedRootKeys, unsealedRootJwksMap: &unsealedRootJwksMap, aes256KeyGenPool: aes256KeyGenPool, unsealedRootJwkLatest: unsealedRootJwksLatest}, nil
}

func decryptExistingRootJwks(telemetryService *cryptoutilTelemetry.TelemetryService, ormRepository *cryptoutilOrmRepository.OrmRepository, aes256KeyGenPool *cryptoutilKeygen.KeyGenPool, unsealJwks []joseJwk.Key, encryptedRootJwks []cryptoutilOrmRepository.BarrierRootKey) (*RootKeysService, error) {
	var encryptedRootJwkLatest *cryptoutilOrmRepository.BarrierRootKey
	err := ormRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		encryptedRootJwkLatest, err = sqlTransaction.GetRootKeyLatest() // First row using ORDER BY uuid DESC
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get root JWK latest from database")
	}
	encryptedRootJwkLatestKidUuid := encryptedRootJwkLatest.GetUUID() // during decrypt loop, look for latest by UUID to grab a copy of the decrypted JWK

	// loop through encryptedRootJwks from DB, try using all of the unsealJwks to decrypt them
	unsealedRootKeys := make([]joseJwk.Key, 0, len(encryptedRootJwks))
	unsealedRootJwksMap := make(map[googleUuid.UUID]joseJwk.Key)
	var unsealedRootJwksLatest joseJwk.Key
	var errs []error
	for encryptedRootJwkIndex, encryptedRootJwk := range encryptedRootJwks {
		encryptedRootJwkKidUuid := encryptedRootJwk.GetUUID()
		var unsealedRootJwk joseJwk.Key
		for unsealJwkIndex, unsealJwk := range unsealJwks {
			unsealedRootJwkBytes, err := cryptoutilJose.DecryptBytes([]joseJwk.Key{unsealJwk}, []byte(encryptedRootJwk.GetEncrypted()))
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to decrypt root JWK %d with unseal JWK %d: %w", encryptedRootJwkIndex, unsealJwkIndex, err))
				continue
			}
			unsealedRootJwk, err = joseJwk.ParseKey(unsealedRootJwkBytes)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to parse decrypted root JWK %d after using unseal JWK %d: %w", encryptedRootJwkIndex, unsealJwkIndex, err))
			}
			unsealedRootKeys = append(unsealedRootKeys, unsealedRootJwk)
			unsealedRootJwksMap[encryptedRootJwkKidUuid] = unsealedRootJwk // decrypt success, store it in-memory
			if unsealedRootJwksLatest == nil && encryptedRootJwkKidUuid == encryptedRootJwkLatestKidUuid {
				unsealedRootJwksLatest = unsealedRootJwk // latest UUID matched, grab a copy
			}
		}
		// verify a copy of the latest was found (by UUID) during the decrypt loop
		if unsealedRootJwk == nil {
			errs = append(errs, fmt.Errorf("failed to decrypt root JWK %d: %w", encryptedRootJwkIndex, err))
		}
	}
	if len(unsealedRootJwksMap) == 0 {
		return nil, fmt.Errorf("failed to unseal all root JWKs: %w", errors.Join(errs...)) // FATAL: all decrypt attempts failed
	} else if unsealedRootJwksLatest == nil {
		return nil, fmt.Errorf("failed to unseal latest root JWK: %w", errors.Join(errs...)) // FATAL: encrypted latest not decrypted and found during the the decrypt loop
	}
	// older root keys might not decrypt anymore, but as long as the latest decrypted then the service can start
	if len(errs) == 0 {
		telemetryService.Slogger.Debug("unsealed all root JWKs", "unsealed", len(unsealedRootJwksMap))
	} else {
		telemetryService.Slogger.Warn("unsealed some root JWKs", "unsealed", len(unsealedRootJwksMap), "errors", len(errs), "error", errors.Join(errs...))
	}
	return &RootKeysService{unsealedRootJwksList: unsealedRootKeys, unsealedRootJwksMap: &unsealedRootJwksMap, aes256KeyGenPool: aes256KeyGenPool, unsealedRootJwkLatest: unsealedRootJwksLatest}, nil
}
