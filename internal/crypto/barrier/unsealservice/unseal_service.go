package unsealservice

import (
	cryptoutilUnsealRepository "cryptoutil/internal/crypto/barrier/unsealrepository"
	cryptoutilJose "cryptoutil/internal/crypto/jose"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"
	"errors"
	"fmt"

	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

type UnsealService struct {
	unsealedRootJwksMap    *map[googleUuid.UUID]*joseJwk.Key
	unsealedRootJwksLatest *joseJwk.Key
}

func (u *UnsealService) Get(uuid googleUuid.UUID) *joseJwk.Key {
	return (*u.unsealedRootJwksMap)[uuid]
}

func (u *UnsealService) GetLatest() *joseJwk.Key {
	return u.unsealedRootJwksLatest
}

func NewUnsealService(telemetryService *cryptoutilTelemetry.Service, ormRepository *cryptoutilOrmRepository.RepositoryProvider, unsealKeyRepository *cryptoutilUnsealRepository.UnsealKeyRepository) (*UnsealService, error) {
	unsealJwks := unsealKeyRepository.UnsealJwks() // unseal keys from unseal repository
	if len(unsealJwks) == 0 {
		return nil, fmt.Errorf("no unseal JWKs")
	}

	encryptedRootJwks, err := ormRepository.GetRootKeys() // encrypted root JWKs from DB
	if err != nil {
		return nil, fmt.Errorf("failed to get encrypted root JWKs from DB")
	}

	var unsealService *UnsealService
	if len(encryptedRootJwks) == 0 {
		unsealService, err = createFirstRootJwk(telemetryService, ormRepository, unsealJwks)
	} else {
		unsealService, err = decryptExistingRootJwks(telemetryService, ormRepository, unsealJwks, encryptedRootJwks)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to initialize unseal service: %w", err)
	}
	return unsealService, nil
}

func createFirstRootJwk(telemetryService *cryptoutilTelemetry.Service, ormRepository *cryptoutilOrmRepository.RepositoryProvider, unsealJwks []joseJwk.Key) (*UnsealService, error) {
	unsealedRootJwksLatest, _, err := cryptoutilJose.GenerateAesJWK(cryptoutilJose.AlgDIRECT)
	if err != nil {
		return nil, fmt.Errorf("failed to generate root JWK: %w", err)
	}

	unsealedRootJwksLatestKidUuid, err := cryptoutilJose.ExtractKidUuid(unsealedRootJwksLatest)
	if err != nil {
		return nil, fmt.Errorf("failed to get root JWK kid uuid: %w", err)
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
	err = ormRepository.AddRootKey(&cryptoutilOrmRepository.RootKey{UUID: unsealedRootJwksLatestKidUuid, Serialized: string(jweMessageBytes), KEKUUID: sealJwkKidUuid})
	if err != nil {
		return nil, fmt.Errorf("failed to store root JWK: %w", err)
	}

	// put new, clear root JWK in-memory
	unsealedRootJwksMap := make(map[googleUuid.UUID]*joseJwk.Key)
	unsealedRootJwksMap[unsealedRootJwksLatestKidUuid] = &unsealedRootJwksLatest

	return &UnsealService{unsealedRootJwksMap: &unsealedRootJwksMap, unsealedRootJwksLatest: &unsealedRootJwksLatest}, nil
}

func decryptExistingRootJwks(telemetryService *cryptoutilTelemetry.Service, ormRepository *cryptoutilOrmRepository.RepositoryProvider, unsealJwks []joseJwk.Key, encryptedRootJwks []cryptoutilOrmRepository.RootKey) (*UnsealService, error) {
	encryptedRootJwkLatest, err := ormRepository.GetRootKeyLatest() // First row using ORDER BY uuid DESC
	if err != nil {
		return nil, fmt.Errorf("failed to get root JWK latest from database")
	}
	encryptedRootJwkLatestKidUuid := encryptedRootJwkLatest.GetUUID() // during decrypt loop, look for latest by UUID to grab a copy of the decrypted JWK

	// loop through encryptedRootJwks from DB, try using all of the unsealJwks to decrypt them
	unsealedRootJwksMap := make(map[googleUuid.UUID]*joseJwk.Key)
	var unsealedRootJwksLatest joseJwk.Key
	var errs []error
	for encryptedRootJwkIndex, encryptedRootJwk := range encryptedRootJwks {
		encryptedRootJwkKidUuid := encryptedRootJwk.GetUUID()
		var unsealedRootJwk joseJwk.Key
		for unsealJwkIndex, unsealJwk := range unsealJwks {
			unsealedRootJwkBytes, err := cryptoutilJose.DecryptBytes([]joseJwk.Key{unsealJwk}, []byte(encryptedRootJwk.GetSerialized()))
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to decrypt root JWK %d with unseal JWK %d: %w", encryptedRootJwkIndex, unsealJwkIndex, err))
				continue
			}
			unsealedRootJwk, err = joseJwk.ParseKey(unsealedRootJwkBytes)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to parse decrypted root JWK %d after using unseal JWK %d: %w", encryptedRootJwkIndex, unsealJwkIndex, err))
			}
			unsealedRootJwksMap[encryptedRootJwkKidUuid] = &unsealedRootJwk // decrypt success, store it in-memory
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
	return &UnsealService{unsealedRootJwksMap: &unsealedRootJwksMap, unsealedRootJwksLatest: &unsealedRootJwksLatest}, nil
}
