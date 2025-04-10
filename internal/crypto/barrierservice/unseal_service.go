package barrierservice

import (
	cryptoutilCombinations "cryptoutil/internal/combinations"
	cryptoutilDigests "cryptoutil/internal/crypto/digests"
	cryptoutilJose "cryptoutil/internal/crypto/jose"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilSysinfo "cryptoutil/internal/sysinfo"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"
	"errors"
	"fmt"

	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

const fingerprintLeeway = 1

type unsealService struct {
	unsealedRootJwksMap    *map[googleUuid.UUID]joseJwk.Key
	unsealedRootJwkslatest *joseJwk.Key
}

func (u *unsealService) get(uuid googleUuid.UUID) *joseJwk.Key {
	x := (*u.unsealedRootJwksMap)[uuid]
	return &x
}

func (u *unsealService) getLatest() *joseJwk.Key {
	return u.unsealedRootJwkslatest
}

func newUnsealService(telemetryService *cryptoutilTelemetry.Service, ormRepository *cryptoutilOrmRepository.RepositoryProvider) (*unsealService, error) {
	// TODO Support other sources of unseal JWKs (e.g. HSM, 3rd-party KMS, secret key sharing, etc)
	unsealJwks, unsealJwksErr := sysFingerprintUnsealJwks()
	if unsealJwksErr != nil {
		return nil, fmt.Errorf("failed to get unseal JWKs: %w", unsealJwksErr)
	}

	encryptedRootJwks, err := ormRepository.GetRootKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to get encrypted root JWKs from DB")
	}

	unsealedRootJwksMap := make(map[googleUuid.UUID]joseJwk.Key)
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
		unsealedRootJwksMap[unsealedRootJwkslatestKidUuid] = unsealedRootJwkslatest // generate success, store it in-memory

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

		err = ormRepository.AddRootKey(&cryptoutilOrmRepository.RootKey{UUID: unsealedRootJwkslatestKidUuid, Serialized: string(jweMessageBytes), KEKUUID: sealJwkKidUuid})
		if err != nil {
			return nil, fmt.Errorf("failed to store root JWK: %w", err)
		}

		telemetryService.Slogger.Info("Encrypted Root JWK", "JWE Headers", jweHeaders)
	} else {
		// at least one root JWK was created and stored in the DB, try to decrypt them using provided unsealJwkSet
		encryptedRootJwkLatest, err := ormRepository.GetRootKeyLatest()
		if err != nil {
			return nil, fmt.Errorf("failed to get root JWK latest from database")
		}
		encryptedRootJwkLatestKidUuid := encryptedRootJwkLatest.GetUUID()

		// loop through encryptedRootJwks, use provided unsealJwkSet to attempt decryption of all root JWKs from DB
		var errs []error
		for unsealJwkIndex, encryptedRootJwk := range encryptedRootJwks {
			encryptedRootJwkKidUuid := encryptedRootJwk.GetUUID()
			var unsealedRootJwk joseJwk.Key
			for rootJwkIndex, unsealJwk := range unsealJwks {
				unsealedRootJwkBytes, err := cryptoutilJose.DecryptBytes([]joseJwk.Key{unsealJwk}, []byte(encryptedRootJwk.GetSerialized()))
				if err != nil {
					errs = append(errs, fmt.Errorf("failed to decrypt root JWK %d with unseal JWK %d: %w", unsealJwkIndex, rootJwkIndex, err))
					continue
				}
				unsealedRootJwk, err = joseJwk.ParseKey(unsealedRootJwkBytes)
				if err != nil {
					errs = append(errs, fmt.Errorf("failed to parse decrypted root JWK %d after using unseal JWK %d: %w", unsealJwkIndex, rootJwkIndex, err))
				}
				unsealedRootJwksMap[encryptedRootJwkKidUuid] = unsealedRootJwk // decrypt success, store it in-memory
				if unsealedRootJwkslatest == nil && encryptedRootJwkKidUuid == encryptedRootJwkLatestKidUuid {
					unsealedRootJwkslatest = unsealedRootJwk
				}
			}
			if unsealedRootJwk == nil {
				errs = append(errs, fmt.Errorf("failed to decrypt root JWK %d: %w", unsealJwkIndex, err)) // non-fatal until we have tried all unsealJwks and no root keys were unsealed
			}
		}
		if len(unsealedRootJwksMap) == 0 {
			return nil, fmt.Errorf("failed to unseal all root JWKs: %w", errors.Join(errs...)) // no encrypted root JWKs from DB weren't decrypted via the unseal JWKs
		} else if unsealedRootJwkslatest == nil {
			return nil, fmt.Errorf("failed to unseal latest root JWK: %w", errors.Join(errs...)) // latest encrypted root JWK from DB wasn't decrypted via the unseal JWKs
		} else if len(errs) == 0 {
			telemetryService.Slogger.Debug("unsealed all root JWKs", "unsealed", len(unsealedRootJwksMap))
		} else {
			telemetryService.Slogger.Warn("unsealed some root JWKs", "unsealed", len(unsealedRootJwksMap), "errors", len(errs), "error", errors.Join(errs...))
		}
	}
	return &unsealService{unsealedRootJwksMap: &unsealedRootJwksMap, unsealedRootJwkslatest: &unsealedRootJwkslatest}, nil
}

func sysFingerprintUnsealJwks() ([]joseJwk.Key, error) {
	sysinfos, err := cryptoutilSysinfo.GetAllInfo(&cryptoutilSysinfo.DefaultSysInfoProvider{})
	if err != nil {
		return nil, fmt.Errorf("failed to get sysinfo: %w", err)
	}

	numSysinfos := len(sysinfos)
	if numSysinfos == 0 {
		return nil, fmt.Errorf("empty sysinfos not supported")
	}

	var chooseN int
	if numSysinfos == 1 {
		chooseN = numSysinfos // use it as-is
	} else {
		chooseN = numSysinfos - fingerprintLeeway // use combinations of M choose M-1
	}

	unsealJwks := make([]joseJwk.Key, 0, numSysinfos)                                  // could be more if leeway is more than 1
	combinations, err := cryptoutilCombinations.ComputeCombinations(sysinfos, chooseN) // M choose N combinationss
	if err != nil {
		return nil, fmt.Errorf("failed to compute %d of %d combinations of sysinfo: %w", numSysinfos, numSysinfos-1, err)
	} else if len(combinations) == 0 {
		return nil, fmt.Errorf("no combinations")
	}
	for _, combination := range combinations {
		var sysinfoCombinationBytes []byte
		for _, value := range combination {
			sysinfoCombinationBytes = append(sysinfoCombinationBytes, value...)
		}

		derivedSecretBytes := cryptoutilDigests.SHA512(fmt.Append(sysinfoCombinationBytes, []byte("secret")))
		derivedSaltBytes := cryptoutilDigests.SHA512(fmt.Append(sysinfoCombinationBytes, []byte("salt")))
		derivedUnsealKeyBytes, err := cryptoutilDigests.HKDFwithSHA256(derivedSecretBytes, derivedSaltBytes, []byte("derive unsealed JWKs algorithm v1"), 32)
		if err != nil {
			return nil, fmt.Errorf("failed to create JWK: %w", err)
		}

		unsealJwk, _, err := cryptoutilJose.CreateAesJWK(cryptoutilJose.AlgA256GCMKW, derivedUnsealKeyBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to create JWK: %w", err)
		}
		unsealJwks = append(unsealJwks, unsealJwk)
	}

	return unsealJwks, nil
}
