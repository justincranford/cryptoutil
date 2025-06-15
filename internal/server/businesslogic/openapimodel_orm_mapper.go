package businesslogic

import (
	"errors"
	"fmt"
	"time"

	cryptoutilJose "cryptoutil/internal/common/crypto/jose"
	cryptoutilUtil "cryptoutil/internal/common/util"
	cryptoutilBusinessLogicModel "cryptoutil/internal/openapi/model"
	cryptoutilOrmRepository "cryptoutil/internal/server/repository/orm"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
)

var (
	ormKeyPoolAlgorithmToJoseEncAndAlg = map[cryptoutilOrmRepository.KeyPoolAlgorithm]struct {
		enc *joseJwa.ContentEncryptionAlgorithm
		alg *joseJwa.KeyEncryptionAlgorithm
	}{
		cryptoutilOrmRepository.A256GCM_A256KW:    {enc: &cryptoutilJose.EncA256GCM, alg: &cryptoutilJose.AlgA256KW},
		cryptoutilOrmRepository.A192GCM_A256KW:    {enc: &cryptoutilJose.EncA192GCM, alg: &cryptoutilJose.AlgA256KW},
		cryptoutilOrmRepository.A128GCM_A256KW:    {enc: &cryptoutilJose.EncA128GCM, alg: &cryptoutilJose.AlgA256KW},
		cryptoutilOrmRepository.A256GCM_A192KW:    {enc: &cryptoutilJose.EncA256GCM, alg: &cryptoutilJose.AlgA192KW},
		cryptoutilOrmRepository.A192GCM_A192KW:    {enc: &cryptoutilJose.EncA192GCM, alg: &cryptoutilJose.AlgA192KW},
		cryptoutilOrmRepository.A128GCM_A192KW:    {enc: &cryptoutilJose.EncA128GCM, alg: &cryptoutilJose.AlgA192KW},
		cryptoutilOrmRepository.A256GCM_A128KW:    {enc: &cryptoutilJose.EncA256GCM, alg: &cryptoutilJose.AlgA128KW},
		cryptoutilOrmRepository.A192GCM_A128KW:    {enc: &cryptoutilJose.EncA192GCM, alg: &cryptoutilJose.AlgA128KW},
		cryptoutilOrmRepository.A128GCM_A128KW:    {enc: &cryptoutilJose.EncA128GCM, alg: &cryptoutilJose.AlgA128KW},
		cryptoutilOrmRepository.A256GCM_A256GCMKW: {enc: &cryptoutilJose.EncA256GCM, alg: &cryptoutilJose.AlgA256GCMKW},
		cryptoutilOrmRepository.A192GCM_A256GCMKW: {enc: &cryptoutilJose.EncA192GCM, alg: &cryptoutilJose.AlgA256GCMKW},
		cryptoutilOrmRepository.A128GCM_A256GCMKW: {enc: &cryptoutilJose.EncA128GCM, alg: &cryptoutilJose.AlgA256GCMKW},
		cryptoutilOrmRepository.A256GCM_A192GCMKW: {enc: &cryptoutilJose.EncA256GCM, alg: &cryptoutilJose.AlgA192GCMKW},
		cryptoutilOrmRepository.A192GCM_A192GCMKW: {enc: &cryptoutilJose.EncA192GCM, alg: &cryptoutilJose.AlgA192GCMKW},
		cryptoutilOrmRepository.A128GCM_A192GCMKW: {enc: &cryptoutilJose.EncA128GCM, alg: &cryptoutilJose.AlgA192GCMKW},
		cryptoutilOrmRepository.A256GCM_A128GCMKW: {enc: &cryptoutilJose.EncA256GCM, alg: &cryptoutilJose.AlgA128GCMKW},
		cryptoutilOrmRepository.A192GCM_A128GCMKW: {enc: &cryptoutilJose.EncA192GCM, alg: &cryptoutilJose.AlgA128GCMKW},
		cryptoutilOrmRepository.A128GCM_A128GCMKW: {enc: &cryptoutilJose.EncA128GCM, alg: &cryptoutilJose.AlgA128GCMKW},
		cryptoutilOrmRepository.A256GCM_dir:       {enc: &cryptoutilJose.EncA256GCM, alg: &cryptoutilJose.AlgDir},
		cryptoutilOrmRepository.A192GCM_dir:       {enc: &cryptoutilJose.EncA192GCM, alg: &cryptoutilJose.AlgDir},
		cryptoutilOrmRepository.A128GCM_dir:       {enc: &cryptoutilJose.EncA128GCM, alg: &cryptoutilJose.AlgDir},

		cryptoutilOrmRepository.A256GCM_RSAOAEP512: {enc: &cryptoutilJose.EncA256GCM, alg: &cryptoutilJose.AlgRSAOAEP512},
		cryptoutilOrmRepository.A192GCM_RSAOAEP512: {enc: &cryptoutilJose.EncA192GCM, alg: &cryptoutilJose.AlgRSAOAEP512},
		cryptoutilOrmRepository.A128GCM_RSAOAEP512: {enc: &cryptoutilJose.EncA128GCM, alg: &cryptoutilJose.AlgRSAOAEP512},
		cryptoutilOrmRepository.A256GCM_RSAOAEP384: {enc: &cryptoutilJose.EncA256GCM, alg: &cryptoutilJose.AlgRSAOAEP384},
		cryptoutilOrmRepository.A192GCM_RSAOAEP384: {enc: &cryptoutilJose.EncA192GCM, alg: &cryptoutilJose.AlgRSAOAEP384},
		cryptoutilOrmRepository.A128GCM_RSAOAEP384: {enc: &cryptoutilJose.EncA128GCM, alg: &cryptoutilJose.AlgRSAOAEP384},
		cryptoutilOrmRepository.A256GCM_RSAOAEP256: {enc: &cryptoutilJose.EncA256GCM, alg: &cryptoutilJose.AlgRSAOAEP256},
		cryptoutilOrmRepository.A192GCM_RSAOAEP256: {enc: &cryptoutilJose.EncA192GCM, alg: &cryptoutilJose.AlgRSAOAEP256},
		cryptoutilOrmRepository.A128GCM_RSAOAEP256: {enc: &cryptoutilJose.EncA128GCM, alg: &cryptoutilJose.AlgRSAOAEP256},
		cryptoutilOrmRepository.A256GCM_RSAOAEP:    {enc: &cryptoutilJose.EncA256GCM, alg: &cryptoutilJose.AlgRSAOAEP},
		cryptoutilOrmRepository.A192GCM_RSAOAEP:    {enc: &cryptoutilJose.EncA192GCM, alg: &cryptoutilJose.AlgRSAOAEP},
		cryptoutilOrmRepository.A128GCM_RSAOAEP:    {enc: &cryptoutilJose.EncA128GCM, alg: &cryptoutilJose.AlgRSAOAEP},
		cryptoutilOrmRepository.A256GCM_RSA15:      {enc: &cryptoutilJose.EncA256GCM, alg: &cryptoutilJose.AlgRSA15},
		cryptoutilOrmRepository.A192GCM_RSA15:      {enc: &cryptoutilJose.EncA192GCM, alg: &cryptoutilJose.AlgRSA15},
		cryptoutilOrmRepository.A128GCM_RSA15:      {enc: &cryptoutilJose.EncA128GCM, alg: &cryptoutilJose.AlgRSA15},

		cryptoutilOrmRepository.A256GCM_ECDHESA256KW: {enc: &cryptoutilJose.EncA256GCM, alg: &cryptoutilJose.AlgECDHESA256KW},
		cryptoutilOrmRepository.A192GCM_ECDHESA256KW: {enc: &cryptoutilJose.EncA192GCM, alg: &cryptoutilJose.AlgECDHESA256KW},
		cryptoutilOrmRepository.A128GCM_ECDHESA256KW: {enc: &cryptoutilJose.EncA128GCM, alg: &cryptoutilJose.AlgECDHESA256KW},
		cryptoutilOrmRepository.A256GCM_ECDHESA192KW: {enc: &cryptoutilJose.EncA256GCM, alg: &cryptoutilJose.AlgECDHESA192KW},
		cryptoutilOrmRepository.A192GCM_ECDHESA192KW: {enc: &cryptoutilJose.EncA192GCM, alg: &cryptoutilJose.AlgECDHESA192KW},
		cryptoutilOrmRepository.A128GCM_ECDHESA192KW: {enc: &cryptoutilJose.EncA128GCM, alg: &cryptoutilJose.AlgECDHESA192KW},
		cryptoutilOrmRepository.A256GCM_ECDHESA128KW: {enc: &cryptoutilJose.EncA256GCM, alg: &cryptoutilJose.AlgECDHESA128KW},
		cryptoutilOrmRepository.A192GCM_ECDHESA128KW: {enc: &cryptoutilJose.EncA192GCM, alg: &cryptoutilJose.AlgECDHESA128KW},
		cryptoutilOrmRepository.A128GCM_ECDHESA128KW: {enc: &cryptoutilJose.EncA128GCM, alg: &cryptoutilJose.AlgECDHESA128KW},
		cryptoutilOrmRepository.A256GCM_ECDHES:       {enc: &cryptoutilJose.EncA256GCM, alg: &cryptoutilJose.AlgECDHES},
		cryptoutilOrmRepository.A192GCM_ECDHES:       {enc: &cryptoutilJose.EncA192GCM, alg: &cryptoutilJose.AlgECDHES},
		cryptoutilOrmRepository.A128GCM_ECDHES:       {enc: &cryptoutilJose.EncA128GCM, alg: &cryptoutilJose.AlgECDHES},

		cryptoutilOrmRepository.A256CBCHS512_A256KW:    {enc: &cryptoutilJose.EncA256CBC_HS512, alg: &cryptoutilJose.AlgA256KW},
		cryptoutilOrmRepository.A192CBCHS384_A256KW:    {enc: &cryptoutilJose.EncA192CBC_HS384, alg: &cryptoutilJose.AlgA256KW},
		cryptoutilOrmRepository.A128CBCHS256_A256KW:    {enc: &cryptoutilJose.EncA128CBC_HS256, alg: &cryptoutilJose.AlgA256KW},
		cryptoutilOrmRepository.A256CBCHS512_A192KW:    {enc: &cryptoutilJose.EncA256CBC_HS512, alg: &cryptoutilJose.AlgA192KW},
		cryptoutilOrmRepository.A192CBCHS384_A192KW:    {enc: &cryptoutilJose.EncA192CBC_HS384, alg: &cryptoutilJose.AlgA192KW},
		cryptoutilOrmRepository.A128CBCHS256_A192KW:    {enc: &cryptoutilJose.EncA128CBC_HS256, alg: &cryptoutilJose.AlgA192KW},
		cryptoutilOrmRepository.A256CBCHS512_A128KW:    {enc: &cryptoutilJose.EncA256CBC_HS512, alg: &cryptoutilJose.AlgA128KW},
		cryptoutilOrmRepository.A192CBCHS384_A128KW:    {enc: &cryptoutilJose.EncA192CBC_HS384, alg: &cryptoutilJose.AlgA128KW},
		cryptoutilOrmRepository.A128CBCHS256_A128KW:    {enc: &cryptoutilJose.EncA128CBC_HS256, alg: &cryptoutilJose.AlgA128KW},
		cryptoutilOrmRepository.A256CBCHS512_A256GCMKW: {enc: &cryptoutilJose.EncA256CBC_HS512, alg: &cryptoutilJose.AlgA256GCMKW},
		cryptoutilOrmRepository.A192CBCHS384_A256GCMKW: {enc: &cryptoutilJose.EncA192CBC_HS384, alg: &cryptoutilJose.AlgA256GCMKW},
		cryptoutilOrmRepository.A128CBCHS256_A256GCMKW: {enc: &cryptoutilJose.EncA128CBC_HS256, alg: &cryptoutilJose.AlgA256GCMKW},
		cryptoutilOrmRepository.A256CBCHS512_A192GCMKW: {enc: &cryptoutilJose.EncA256CBC_HS512, alg: &cryptoutilJose.AlgA192GCMKW},
		cryptoutilOrmRepository.A192CBCHS384_A192GCMKW: {enc: &cryptoutilJose.EncA192CBC_HS384, alg: &cryptoutilJose.AlgA192GCMKW},
		cryptoutilOrmRepository.A128CBCHS256_A192GCMKW: {enc: &cryptoutilJose.EncA128CBC_HS256, alg: &cryptoutilJose.AlgA192GCMKW},
		cryptoutilOrmRepository.A256CBCHS512_A128GCMKW: {enc: &cryptoutilJose.EncA256CBC_HS512, alg: &cryptoutilJose.AlgA128GCMKW},
		cryptoutilOrmRepository.A192CBCHS384_A128GCMKW: {enc: &cryptoutilJose.EncA192CBC_HS384, alg: &cryptoutilJose.AlgA128GCMKW},
		cryptoutilOrmRepository.A128CBCHS256_A128GCMKW: {enc: &cryptoutilJose.EncA128CBC_HS256, alg: &cryptoutilJose.AlgA128GCMKW},
		cryptoutilOrmRepository.A256CBCHS512_dir:       {enc: &cryptoutilJose.EncA256CBC_HS512, alg: &cryptoutilJose.AlgDir},
		cryptoutilOrmRepository.A192CBCHS384_dir:       {enc: &cryptoutilJose.EncA192CBC_HS384, alg: &cryptoutilJose.AlgDir},
		cryptoutilOrmRepository.A128CBCHS256_dir:       {enc: &cryptoutilJose.EncA128CBC_HS256, alg: &cryptoutilJose.AlgDir},

		cryptoutilOrmRepository.A256CBC_HS512_RSAOAEP512: {enc: &cryptoutilJose.EncA256CBC_HS512, alg: &cryptoutilJose.AlgRSAOAEP512},
		cryptoutilOrmRepository.A192CBC_HS384_RSAOAEP512: {enc: &cryptoutilJose.EncA192CBC_HS384, alg: &cryptoutilJose.AlgRSAOAEP512},
		cryptoutilOrmRepository.A128CBC_HS256_RSAOAEP512: {enc: &cryptoutilJose.EncA128CBC_HS256, alg: &cryptoutilJose.AlgRSAOAEP512},
		cryptoutilOrmRepository.A256CBC_HS512_RSAOAEP384: {enc: &cryptoutilJose.EncA256CBC_HS512, alg: &cryptoutilJose.AlgRSAOAEP384},
		cryptoutilOrmRepository.A192CBC_HS384_RSAOAEP384: {enc: &cryptoutilJose.EncA192CBC_HS384, alg: &cryptoutilJose.AlgRSAOAEP384},
		cryptoutilOrmRepository.A128CBC_HS256_RSAOAEP384: {enc: &cryptoutilJose.EncA128CBC_HS256, alg: &cryptoutilJose.AlgRSAOAEP384},
		cryptoutilOrmRepository.A256CBC_HS512_RSAOAEP256: {enc: &cryptoutilJose.EncA256CBC_HS512, alg: &cryptoutilJose.AlgRSAOAEP256},
		cryptoutilOrmRepository.A192CBC_HS384_RSAOAEP256: {enc: &cryptoutilJose.EncA192CBC_HS384, alg: &cryptoutilJose.AlgRSAOAEP256},
		cryptoutilOrmRepository.A128CBC_HS256_RSAOAEP256: {enc: &cryptoutilJose.EncA128CBC_HS256, alg: &cryptoutilJose.AlgRSAOAEP256},
		cryptoutilOrmRepository.A256CBC_HS512_RSAOAEP:    {enc: &cryptoutilJose.EncA256CBC_HS512, alg: &cryptoutilJose.AlgRSAOAEP},
		cryptoutilOrmRepository.A192CBC_HS384_RSAOAEP:    {enc: &cryptoutilJose.EncA192CBC_HS384, alg: &cryptoutilJose.AlgRSAOAEP},
		cryptoutilOrmRepository.A128CBC_HS256_RSAOAEP:    {enc: &cryptoutilJose.EncA128CBC_HS256, alg: &cryptoutilJose.AlgRSAOAEP},
		cryptoutilOrmRepository.A256CBC_HS512_RSA15:      {enc: &cryptoutilJose.EncA256CBC_HS512, alg: &cryptoutilJose.AlgRSA15},
		cryptoutilOrmRepository.A192CBC_HS384_RSA15:      {enc: &cryptoutilJose.EncA192CBC_HS384, alg: &cryptoutilJose.AlgRSA15},
		cryptoutilOrmRepository.A128CBC_HS256_RSA15:      {enc: &cryptoutilJose.EncA128CBC_HS256, alg: &cryptoutilJose.AlgRSA15},

		cryptoutilOrmRepository.A256CBC_HS512_ECDHESA256KW: {enc: &cryptoutilJose.EncA256CBC_HS512, alg: &cryptoutilJose.AlgECDHESA256KW},
		cryptoutilOrmRepository.A192CBC_HS384_ECDHESA256KW: {enc: &cryptoutilJose.EncA192CBC_HS384, alg: &cryptoutilJose.AlgECDHESA256KW},
		cryptoutilOrmRepository.A128CBC_HS256_ECDHESA256KW: {enc: &cryptoutilJose.EncA128CBC_HS256, alg: &cryptoutilJose.AlgECDHESA256KW},
		cryptoutilOrmRepository.A192CBC_HS384_ECDHESA192KW: {enc: &cryptoutilJose.EncA192CBC_HS384, alg: &cryptoutilJose.AlgECDHESA192KW},
		cryptoutilOrmRepository.A128CBC_HS256_ECDHESA192KW: {enc: &cryptoutilJose.EncA128CBC_HS256, alg: &cryptoutilJose.AlgECDHESA192KW},
		cryptoutilOrmRepository.A128CBC_HS256_ECDHESA128KW: {enc: &cryptoutilJose.EncA128CBC_HS256, alg: &cryptoutilJose.AlgECDHESA128KW},
		cryptoutilOrmRepository.A256CBC_HS512_ECDHES:       {enc: &cryptoutilJose.EncA256CBC_HS512, alg: &cryptoutilJose.AlgECDHES},
		cryptoutilOrmRepository.A192CBC_HS384_ECDHES:       {enc: &cryptoutilJose.EncA192CBC_HS384, alg: &cryptoutilJose.AlgECDHES},
		cryptoutilOrmRepository.A128CBC_HS256_ECDHES:       {enc: &cryptoutilJose.EncA128CBC_HS256, alg: &cryptoutilJose.AlgECDHES},
	}

	ormKeyPoolAlgorithmToJoseAlg = map[cryptoutilOrmRepository.KeyPoolAlgorithm]*joseJwa.SignatureAlgorithm{
		cryptoutilOrmRepository.RS512: &cryptoutilJose.AlgRS512,
		cryptoutilOrmRepository.RS384: &cryptoutilJose.AlgRS384,
		cryptoutilOrmRepository.RS256: &cryptoutilJose.AlgRS256,
		cryptoutilOrmRepository.PS512: &cryptoutilJose.AlgPS512,
		cryptoutilOrmRepository.PS384: &cryptoutilJose.AlgPS384,
		cryptoutilOrmRepository.PS256: &cryptoutilJose.AlgPS256,
		cryptoutilOrmRepository.ES512: &cryptoutilJose.AlgES512,
		cryptoutilOrmRepository.ES384: &cryptoutilJose.AlgES384,
		cryptoutilOrmRepository.ES256: &cryptoutilJose.AlgES256,
		cryptoutilOrmRepository.HS512: &cryptoutilJose.AlgHS512,
		cryptoutilOrmRepository.HS384: &cryptoutilJose.AlgHS384,
		cryptoutilOrmRepository.HS256: &cryptoutilJose.AlgHS256,
		cryptoutilOrmRepository.EdDSA: &cryptoutilJose.AlgEdDSA,
	}
)

type serviceOrmMapper struct{}

func NewMapper() *serviceOrmMapper {
	return &serviceOrmMapper{}
}

// service => orm

func (m *serviceOrmMapper) toOrmAddKeyPool(keyPoolID googleUuid.UUID, serviceKeyPoolCreate *cryptoutilBusinessLogicModel.KeyPoolCreate) *cryptoutilOrmRepository.KeyPool {
	return &cryptoutilOrmRepository.KeyPool{
		KeyPoolID:                keyPoolID,
		KeyPoolName:              serviceKeyPoolCreate.Name,
		KeyPoolDescription:       serviceKeyPoolCreate.Description,
		KeyPoolProvider:          *m.toOrmKeyPoolProvider(serviceKeyPoolCreate.Provider),
		KeyPoolAlgorithm:         *m.toOrmKeyPoolAlgorithm(serviceKeyPoolCreate.Algorithm),
		KeyPoolVersioningAllowed: *serviceKeyPoolCreate.VersioningAllowed,
		KeyPoolImportAllowed:     *serviceKeyPoolCreate.ImportAllowed,
		KeyPoolExportAllowed:     *serviceKeyPoolCreate.ExportAllowed,
		KeyPoolStatus:            *m.toKeyPoolInitialStatus(serviceKeyPoolCreate.ImportAllowed),
	}
}

func (m *serviceOrmMapper) toOrmKeyPoolProvider(serviceKeyPoolProvider *cryptoutilBusinessLogicModel.KeyPoolProvider) *cryptoutilOrmRepository.KeyPoolProvider {
	ormKeyPoolProvider := cryptoutilOrmRepository.KeyPoolProvider(*serviceKeyPoolProvider)
	return &ormKeyPoolProvider
}

func (m *serviceOrmMapper) toOrmKeyPoolAlgorithm(serviceKeyPoolProvider *cryptoutilBusinessLogicModel.KeyPoolAlgorithm) *cryptoutilOrmRepository.KeyPoolAlgorithm {
	ormKeyPoolAlgorithm := cryptoutilOrmRepository.KeyPoolAlgorithm(*serviceKeyPoolProvider)
	return &ormKeyPoolAlgorithm
}

func (m *serviceOrmMapper) toKeyPoolInitialStatus(serviceKeyPoolImportAllowed *cryptoutilBusinessLogicModel.KeyPoolImportAllowed) *cryptoutilOrmRepository.KeyPoolStatus {
	var ormKeyPoolStatus cryptoutilOrmRepository.KeyPoolStatus
	if *serviceKeyPoolImportAllowed {
		ormKeyPoolStatus = cryptoutilOrmRepository.KeyPoolStatus("pending_import")
	} else {
		ormKeyPoolStatus = cryptoutilOrmRepository.KeyPoolStatus("pending_generate")
	}
	return &ormKeyPoolStatus
}

// orm => service

func (m *serviceOrmMapper) toServiceKeyPools(ormKeyPools []cryptoutilOrmRepository.KeyPool) []cryptoutilBusinessLogicModel.KeyPool {
	serviceKeyPools := make([]cryptoutilBusinessLogicModel.KeyPool, len(ormKeyPools))
	for i, ormKeyPool := range ormKeyPools {
		serviceKeyPools[i] = *m.toServiceKeyPool(&ormKeyPool)
	}
	return serviceKeyPools
}

func (s *serviceOrmMapper) toServiceKeyPool(ormKeyPool *cryptoutilOrmRepository.KeyPool) *cryptoutilBusinessLogicModel.KeyPool {
	return &cryptoutilBusinessLogicModel.KeyPool{
		Id:                (*cryptoutilBusinessLogicModel.KeyPoolId)(&ormKeyPool.KeyPoolID),
		Name:              &ormKeyPool.KeyPoolName,
		Description:       &ormKeyPool.KeyPoolDescription,
		Algorithm:         s.toServiceKeyPoolAlgorithm(&ormKeyPool.KeyPoolAlgorithm),
		Provider:          s.toServiceKeyPoolProvider(&ormKeyPool.KeyPoolProvider),
		VersioningAllowed: &ormKeyPool.KeyPoolVersioningAllowed,
		ImportAllowed:     &ormKeyPool.KeyPoolImportAllowed,
		ExportAllowed:     &ormKeyPool.KeyPoolExportAllowed,
		Status:            s.toServiceKeyPoolStatus(&ormKeyPool.KeyPoolStatus),
	}
}

func (m *serviceOrmMapper) toServiceKeyPoolAlgorithm(ormKeyPoolAlgorithm *cryptoutilOrmRepository.KeyPoolAlgorithm) *cryptoutilBusinessLogicModel.KeyPoolAlgorithm {
	serviceKeyPoolAlgorithm := cryptoutilBusinessLogicModel.KeyPoolAlgorithm(*ormKeyPoolAlgorithm)
	return &serviceKeyPoolAlgorithm
}

func (m *serviceOrmMapper) toServiceKeyPoolProvider(ormKeyPoolProvider *cryptoutilOrmRepository.KeyPoolProvider) *cryptoutilBusinessLogicModel.KeyPoolProvider {
	serviceKeyPoolProvider := cryptoutilBusinessLogicModel.KeyPoolProvider(*ormKeyPoolProvider)
	return &serviceKeyPoolProvider
}

func (m *serviceOrmMapper) toServiceKeyPoolStatus(ormKeyPoolStatus *cryptoutilOrmRepository.KeyPoolStatus) *cryptoutilBusinessLogicModel.KeyPoolStatus {
	serviceKeyPoolStatus := cryptoutilBusinessLogicModel.KeyPoolStatus(*ormKeyPoolStatus)
	return &serviceKeyPoolStatus
}

func (m *serviceOrmMapper) toServiceKeys(ormKeys []cryptoutilOrmRepository.Key, repositoryKeyMaterials []*keyExportableMaterial) ([]cryptoutilBusinessLogicModel.Key, error) {
	serviceKeys := make([]cryptoutilBusinessLogicModel.Key, len(ormKeys))
	var serviceKey *cryptoutilBusinessLogicModel.Key
	var err error
	for i, ormKey := range ormKeys {
		serviceKey, err = m.toServiceKey(&ormKey, repositoryKeyMaterials[i])
		if err != nil {
			return nil, fmt.Errorf("failed to get service key: %w", err)
		}
		serviceKeys[i] = *serviceKey
	}
	return serviceKeys, nil
}

func (m *serviceOrmMapper) toServiceKey(ormKey *cryptoutilOrmRepository.Key, repositoryKeyMaterial *keyExportableMaterial) (*cryptoutilBusinessLogicModel.Key, error) {
	return &cryptoutilBusinessLogicModel.Key{
		Pool:           cryptoutilBusinessLogicModel.KeyPoolId(ormKey.KeyPoolID),
		Id:             ormKey.KeyID,
		GenerateDate:   (*cryptoutilBusinessLogicModel.KeyGenerateDate)(ormKey.KeyGenerateDate),
		ImportDate:     (*cryptoutilBusinessLogicModel.KeyGenerateDate)(ormKey.KeyImportDate),
		ExpirationDate: (*cryptoutilBusinessLogicModel.KeyGenerateDate)(ormKey.KeyExpirationDate),
		RevocationDate: (*cryptoutilBusinessLogicModel.KeyGenerateDate)(ormKey.KeyRevocationDate),
		Public:         repositoryKeyMaterial.public,
		Decrypted:      repositoryKeyMaterial.decrypted,
	}, nil
}

func (m *serviceOrmMapper) toOrmGetKeyPoolsQueryParams(params *cryptoutilBusinessLogicModel.KeyPoolsQueryParams) (*cryptoutilOrmRepository.GetKeyPoolsFilters, error) {
	if params == nil {
		return nil, nil
	}
	var errs []error
	keyPoolIDs, err := m.toOptionalOrmUUIDs(params.Id)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Key Pool ID: %w", err))
	}
	names, err := m.toOptionalOrmStrings(params.Name)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Key Pool Name: %w", err))
	}
	algorithms, err := m.toOrmAlgorithms(params.Algorithm)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Key Pool Algorithm: %w", err))
	}
	sorts, err := m.toOrmKeyPoolSorts(params.Sort)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Key Pool Sort: %w", err))
	}
	pageNumber, err := m.toOrmPageNumber(params.Page)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Page Number: %w", err))
	}
	pageSize, err := m.toOrmPageSize(params.Size)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Page Size: %w", err))
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("invalid Get Key Pools parameters: %w", errors.Join(errs...))
	}

	return &cryptoutilOrmRepository.GetKeyPoolsFilters{
		ID:                keyPoolIDs,
		Name:              names,
		Algorithm:         algorithms,
		VersioningAllowed: params.VersioningAllowed,
		ImportAllowed:     params.ImportAllowed,
		ExportAllowed:     params.ExportAllowed,
		Sort:              sorts,
		PageNumber:        pageNumber,
		PageSize:          pageSize,
	}, nil
}

func (m *serviceOrmMapper) toOrmGetKeyPoolKeysQueryParams(params *cryptoutilBusinessLogicModel.KeyPoolKeysQueryParams) (*cryptoutilOrmRepository.GetKeyPoolKeysFilters, error) {
	if params == nil {
		return nil, nil
	}
	var errs []error
	keyIDs, err := m.toOptionalOrmUUIDs(params.Id)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid KeyID: %w", err))
	}
	minGenerateDate, maxGenerateDate, err := m.toOrmDateRange(params.MinGenerateDate, params.MaxGenerateDate)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Generate Date range: %w", err))
	}
	sorts, err := m.toOrmKeySorts(params.Sort)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Key Sort: %w", err))
	}
	pageNumber, err := m.toOrmPageNumber(params.Page)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Page Number: %w", err))
	}
	pageSize, err := m.toOrmPageSize(params.Size)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Page Size: %w", err))
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("invalid Get Key Pool Keys parameters: %w", errors.Join(errs...))
	}
	return &cryptoutilOrmRepository.GetKeyPoolKeysFilters{
		ID:                  keyIDs,
		MinimumGenerateDate: minGenerateDate,
		MaximumGenerateDate: maxGenerateDate,
		Sort:                sorts,
		PageNumber:          pageNumber,
		PageSize:            pageSize,
	}, nil
}

func (m *serviceOrmMapper) toOrmGetKeysQueryParams(params *cryptoutilBusinessLogicModel.KeysQueryParams) (*cryptoutilOrmRepository.GetKeysFilters, error) {
	if params == nil {
		return nil, nil
	}
	var errs []error
	keyPoolIDs, err := m.toOptionalOrmUUIDs(params.Pool)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid KeyPoolID: %w", err))
	}
	keyIDs, err := m.toOptionalOrmUUIDs(params.Id)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid KeyID: %w", err))
	}
	minGenerateDate, maxGenerateDate, err := m.toOrmDateRange(params.MinGenerateDate, params.MaxGenerateDate)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Generate Date range: %w", err))
	}
	sorts, err := m.toOrmKeySorts(params.Sort)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Key Sort: %w", err))
	}
	pageNumber, err := m.toOrmPageNumber(params.Page)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Page Number: %w", err))
	}
	pageSize, err := m.toOrmPageSize(params.Size)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Page Size: %w", err))
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("invalid Get Keys parameters: %w", errors.Join(errs...))
	}

	return &cryptoutilOrmRepository.GetKeysFilters{
		Pool:                keyPoolIDs,
		ID:                  keyIDs,
		MinimumGenerateDate: minGenerateDate,
		MaximumGenerateDate: maxGenerateDate,
		Sort:                sorts,
		PageNumber:          pageNumber,
		PageSize:            pageSize,
	}, nil
}

// Helper methods

func (*serviceOrmMapper) toOptionalOrmUUIDs(uuids *[]googleUuid.UUID) ([]googleUuid.UUID, error) {
	if uuids == nil || len(*uuids) == 0 {
		return nil, nil
	}
	if err := cryptoutilUtil.ValidateUUIDs(*uuids, "invalid UUIDs"); err != nil {
		return nil, err
	}
	return *uuids, nil
}

func (*serviceOrmMapper) toOptionalOrmStrings(strings *[]string) ([]string, error) {
	if strings == nil || len(*strings) == 0 {
		return nil, nil
	}
	for _, value := range *strings {
		if len(value) == 0 {
			return nil, fmt.Errorf("value must not be empty string")
		}
	}
	return *strings, nil
}

func (*serviceOrmMapper) toOrmDateRange(minDate *time.Time, maxDate *time.Time) (*time.Time, *time.Time, error) {
	var errs []error
	nonNullMinDate := minDate != nil
	nonNullMaxDate := maxDate != nil
	if nonNullMinDate || nonNullMaxDate {
		now := time.Now().UTC()
		if nonNullMinDate && minDate.Compare(now) > 0 {
			errs = append(errs, fmt.Errorf("Min Date can't be in the future"))
		}
		if nonNullMaxDate {
			// if maxDate.Compare(now) > 0 {
			// 	errs = append(errs, fmt.Errorf("Max Date can't be in the future"))
			// }
			if nonNullMinDate && minDate.Compare(*maxDate) > 0 {
				errs = append(errs, fmt.Errorf("Min Date must be before Max Date"))
			}
		}
	}
	return minDate, maxDate, errors.Join(errs...)
}

func (m *serviceOrmMapper) toOrmAlgorithms(algorithms *[]cryptoutilBusinessLogicModel.KeyPoolAlgorithm) ([]string, error) {
	newVar := toStrings(algorithms, func(algorithm cryptoutilBusinessLogicModel.KeyPoolAlgorithm) string {
		return string(algorithm)
	})
	return newVar, nil
}

func (m *serviceOrmMapper) toOrmKeyPoolSorts(keyPoolSorts *[]cryptoutilBusinessLogicModel.KeyPoolSort) ([]string, error) {
	newVar := toStrings(keyPoolSorts, func(keyPoolSort cryptoutilBusinessLogicModel.KeyPoolSort) string { return string(keyPoolSort) })
	return newVar, nil
}

func (m *serviceOrmMapper) toOrmKeySorts(keySorts *[]cryptoutilBusinessLogicModel.KeySort) ([]string, error) {
	newVar := toStrings(keySorts, func(keySort cryptoutilBusinessLogicModel.KeySort) string { return string(keySort) })
	return newVar, nil
}

func (*serviceOrmMapper) toOrmPageNumber(pageNumber *cryptoutilBusinessLogicModel.PageNumber) (int, error) {
	if pageNumber == nil {
		return 0, nil
	} else if *pageNumber >= 0 {
		return *pageNumber, nil
	}
	return 0, fmt.Errorf("Page Number must be zero or higher")
}

func (*serviceOrmMapper) toOrmPageSize(pageSize *cryptoutilBusinessLogicModel.PageSize) (int, error) {
	if pageSize == nil {
		return 25, nil
	} else if *pageSize >= 1 {
		return *pageSize, nil
	}
	return 0, fmt.Errorf("Page Size must be one or higher")
}

func toStrings[T any](items *[]T, toString func(T) string) []string {
	if items == nil || len(*items) == 0 {
		return nil
	}
	converted := make([]string, 0, len(*items))
	for _, item := range *items {
		converted = append(converted, toString(item))
	}
	return converted
}

func (m *serviceOrmMapper) isJwe(ormKeyPoolAlgorithm *cryptoutilOrmRepository.KeyPoolAlgorithm) bool {
	_, ok := ormKeyPoolAlgorithmToJoseEncAndAlg[*ormKeyPoolAlgorithm]
	return ok
}

func (m *serviceOrmMapper) toJweEncAndAlg(ormKeyPoolAlgorithm *cryptoutilOrmRepository.KeyPoolAlgorithm) (*joseJwa.ContentEncryptionAlgorithm, *joseJwa.KeyEncryptionAlgorithm, error) {
	if encAndAlg, ok := ormKeyPoolAlgorithmToJoseEncAndAlg[*ormKeyPoolAlgorithm]; ok {
		return encAndAlg.enc, encAndAlg.alg, nil
	}
	return nil, nil, fmt.Errorf("unsupported JWE KeyPoolAlgorithm '%s'", *ormKeyPoolAlgorithm)
}

func (m *serviceOrmMapper) isJws(ormKeyPoolAlgorithm *cryptoutilOrmRepository.KeyPoolAlgorithm) bool {
	_, ok := ormKeyPoolAlgorithmToJoseAlg[*ormKeyPoolAlgorithm]
	return ok
}

func (m *serviceOrmMapper) toJwsAlg(ormKeyPoolAlgorithm *cryptoutilOrmRepository.KeyPoolAlgorithm) (*joseJwa.SignatureAlgorithm, error) {
	if alg, ok := ormKeyPoolAlgorithmToJoseAlg[*ormKeyPoolAlgorithm]; ok {
		return alg, nil
	}
	return nil, fmt.Errorf("unsupported JWS KeyPoolAlgorithm '%s'", *ormKeyPoolAlgorithm)
}
