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
	unsealedRootJwkslatest *joseJwk.Key
}

func (u *UnsealService) Get(uuid googleUuid.UUID) *joseJwk.Key {
	return (*u.unsealedRootJwksMap)[uuid]
}

func (u *UnsealService) GetLatest() *joseJwk.Key {
	return u.unsealedRootJwkslatest
}

func NewUnsealService(telemetryService *cryptoutilTelemetry.Service, ormRepository *cryptoutilOrmRepository.RepositoryProvider, unsealKeyRepository *cryptoutilUnsealRepository.UnsealKeyRepository) (*UnsealService, error) {
	unsealJwks := unsealKeyRepository.UnsealJwks()
	if len(unsealJwks) == 0 {
		return nil, fmt.Errorf("no unseal JWKs")
	}

	encryptedRootJwks, err := ormRepository.GetRootKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to get encrypted root JWKs from DB")
	}

	unsealedRootJwksMap := make(map[googleUuid.UUID]*joseJwk.Key)
	var unsealedRootJwkslatest joseJwk.Key
	if len(encryptedRootJwks) == 0 {
		// not root JWKs in the DB, generate one and encrypt it with the first unsealJwkSet, then put it in the DB
		unsealedRootJwkslatest, _, err = cryptoutilJose.GenerateAesJWK(cryptoutilJose.AlgDIRECT)
		if err != nil {
			return nil, fmt.Errorf("failed to generate root JWK: %w", err)
		}

		unsealedRootJwkslatestKidUuid, err := cryptoutilJose.ExtractKidUuid(unsealedRootJwkslatest)
		if err != nil {
			return nil, fmt.Errorf("failed to get root JWK kid uuid: %w", err)
		}

		jweMessage, jweMessageBytes, err := cryptoutilJose.EncryptKey(unsealJwks, unsealedRootJwkslatest)
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

		// put encrypted root JWK in DB (i.e. generated)
		err = ormRepository.AddRootKey(&cryptoutilOrmRepository.RootKey{UUID: unsealedRootJwkslatestKidUuid, Serialized: string(jweMessageBytes), KEKUUID: sealJwkKidUuid})
		if err != nil {
			return nil, fmt.Errorf("failed to store root JWK: %w", err)
		}

		// put clear root JWK in-memory (i.e. generated)
		unsealedRootJwksMap[unsealedRootJwkslatestKidUuid] = &unsealedRootJwkslatest

		telemetryService.Slogger.Info("Encrypted Root JWK", "JWE Headers", jweHeaders)
	} else {
		// one or more encrypted root JWKs was found in DB, try using unsealJwks to decrypt them
		encryptedRootJwkLatest, err := ormRepository.GetRootKeyLatest() // First row using ORDER BY uuid DESC
		if err != nil {
			return nil, fmt.Errorf("failed to get root JWK latest from database")
		}
		encryptedRootJwkLatestKidUuid := encryptedRootJwkLatest.GetUUID() // look for this in the decrypt loop and get a copy

		// loop through encryptedRootJwks from DB, try using unsealJwks to decrypt them
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
				if unsealedRootJwkslatest == nil && encryptedRootJwkKidUuid == encryptedRootJwkLatestKidUuid {
					unsealedRootJwkslatest = unsealedRootJwk // latest UUID matched, grab a copy
				}
			}
			// verify a copy of latest was found during the decrypt loop
			if unsealedRootJwk == nil {
				errs = append(errs, fmt.Errorf("failed to decrypt root JWK %d: %w", encryptedRootJwkIndex, err))
			}
		}
		if len(unsealedRootJwksMap) == 0 {
			return nil, fmt.Errorf("failed to unseal all root JWKs: %w", errors.Join(errs...)) // FATAL: all decrypt attempts failed
		} else if unsealedRootJwkslatest == nil {
			return nil, fmt.Errorf("failed to unseal latest root JWK: %w", errors.Join(errs...)) // FATAL: encrypted latest not decrypted and found during the the decrypt loop
		}
		// older root keys might not decrypt anymore, but as long as the latest decrypted then the service can start
		if len(errs) == 0 {
			telemetryService.Slogger.Debug("unsealed all root JWKs", "unsealed", len(unsealedRootJwksMap))
		} else {
			telemetryService.Slogger.Warn("unsealed some root JWKs", "unsealed", len(unsealedRootJwksMap), "errors", len(errs), "error", errors.Join(errs...))
		}
	}
	return &UnsealService{unsealedRootJwksMap: &unsealedRootJwksMap, unsealedRootJwkslatest: &unsealedRootJwkslatest}, nil
}
