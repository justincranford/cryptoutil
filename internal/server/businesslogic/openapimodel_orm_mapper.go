package businesslogic

import (
	"errors"
	"fmt"
	"time"

	"cryptoutil/internal/common/constant"
	cryptoutilJose "cryptoutil/internal/common/crypto/jose"
	cryptoutilUtil "cryptoutil/internal/common/util"
	cryptoutilBusinessLogicModel "cryptoutil/internal/openapi/model"
	cryptoutilOrmRepository "cryptoutil/internal/server/repository/orm"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
)

var (
	ormElasticKeyAlgorithmToJoseEncAndAlg = map[constant.ElasticKeyAlgorithm]struct {
		enc *joseJwa.ContentEncryptionAlgorithm
		alg *joseJwa.KeyEncryptionAlgorithm
	}{
		constant.A256GCM_A256KW:    {enc: &cryptoutilJose.EncA256GCM, alg: &cryptoutilJose.AlgA256KW},
		constant.A192GCM_A256KW:    {enc: &cryptoutilJose.EncA192GCM, alg: &cryptoutilJose.AlgA256KW},
		constant.A128GCM_A256KW:    {enc: &cryptoutilJose.EncA128GCM, alg: &cryptoutilJose.AlgA256KW},
		constant.A256GCM_A192KW:    {enc: &cryptoutilJose.EncA256GCM, alg: &cryptoutilJose.AlgA192KW},
		constant.A192GCM_A192KW:    {enc: &cryptoutilJose.EncA192GCM, alg: &cryptoutilJose.AlgA192KW},
		constant.A128GCM_A192KW:    {enc: &cryptoutilJose.EncA128GCM, alg: &cryptoutilJose.AlgA192KW},
		constant.A256GCM_A128KW:    {enc: &cryptoutilJose.EncA256GCM, alg: &cryptoutilJose.AlgA128KW},
		constant.A192GCM_A128KW:    {enc: &cryptoutilJose.EncA192GCM, alg: &cryptoutilJose.AlgA128KW},
		constant.A128GCM_A128KW:    {enc: &cryptoutilJose.EncA128GCM, alg: &cryptoutilJose.AlgA128KW},
		constant.A256GCM_A256GCMKW: {enc: &cryptoutilJose.EncA256GCM, alg: &cryptoutilJose.AlgA256GCMKW},
		constant.A192GCM_A256GCMKW: {enc: &cryptoutilJose.EncA192GCM, alg: &cryptoutilJose.AlgA256GCMKW},
		constant.A128GCM_A256GCMKW: {enc: &cryptoutilJose.EncA128GCM, alg: &cryptoutilJose.AlgA256GCMKW},
		constant.A256GCM_A192GCMKW: {enc: &cryptoutilJose.EncA256GCM, alg: &cryptoutilJose.AlgA192GCMKW},
		constant.A192GCM_A192GCMKW: {enc: &cryptoutilJose.EncA192GCM, alg: &cryptoutilJose.AlgA192GCMKW},
		constant.A128GCM_A192GCMKW: {enc: &cryptoutilJose.EncA128GCM, alg: &cryptoutilJose.AlgA192GCMKW},
		constant.A256GCM_A128GCMKW: {enc: &cryptoutilJose.EncA256GCM, alg: &cryptoutilJose.AlgA128GCMKW},
		constant.A192GCM_A128GCMKW: {enc: &cryptoutilJose.EncA192GCM, alg: &cryptoutilJose.AlgA128GCMKW},
		constant.A128GCM_A128GCMKW: {enc: &cryptoutilJose.EncA128GCM, alg: &cryptoutilJose.AlgA128GCMKW},
		constant.A256GCM_dir:       {enc: &cryptoutilJose.EncA256GCM, alg: &cryptoutilJose.AlgDir},
		constant.A192GCM_dir:       {enc: &cryptoutilJose.EncA192GCM, alg: &cryptoutilJose.AlgDir},
		constant.A128GCM_dir:       {enc: &cryptoutilJose.EncA128GCM, alg: &cryptoutilJose.AlgDir},

		constant.A256GCM_RSAOAEP512: {enc: &cryptoutilJose.EncA256GCM, alg: &cryptoutilJose.AlgRSAOAEP512},
		constant.A192GCM_RSAOAEP512: {enc: &cryptoutilJose.EncA192GCM, alg: &cryptoutilJose.AlgRSAOAEP512},
		constant.A128GCM_RSAOAEP512: {enc: &cryptoutilJose.EncA128GCM, alg: &cryptoutilJose.AlgRSAOAEP512},
		constant.A256GCM_RSAOAEP384: {enc: &cryptoutilJose.EncA256GCM, alg: &cryptoutilJose.AlgRSAOAEP384},
		constant.A192GCM_RSAOAEP384: {enc: &cryptoutilJose.EncA192GCM, alg: &cryptoutilJose.AlgRSAOAEP384},
		constant.A128GCM_RSAOAEP384: {enc: &cryptoutilJose.EncA128GCM, alg: &cryptoutilJose.AlgRSAOAEP384},
		constant.A256GCM_RSAOAEP256: {enc: &cryptoutilJose.EncA256GCM, alg: &cryptoutilJose.AlgRSAOAEP256},
		constant.A192GCM_RSAOAEP256: {enc: &cryptoutilJose.EncA192GCM, alg: &cryptoutilJose.AlgRSAOAEP256},
		constant.A128GCM_RSAOAEP256: {enc: &cryptoutilJose.EncA128GCM, alg: &cryptoutilJose.AlgRSAOAEP256},
		constant.A256GCM_RSAOAEP:    {enc: &cryptoutilJose.EncA256GCM, alg: &cryptoutilJose.AlgRSAOAEP},
		constant.A192GCM_RSAOAEP:    {enc: &cryptoutilJose.EncA192GCM, alg: &cryptoutilJose.AlgRSAOAEP},
		constant.A128GCM_RSAOAEP:    {enc: &cryptoutilJose.EncA128GCM, alg: &cryptoutilJose.AlgRSAOAEP},
		constant.A256GCM_RSA15:      {enc: &cryptoutilJose.EncA256GCM, alg: &cryptoutilJose.AlgRSA15},
		constant.A192GCM_RSA15:      {enc: &cryptoutilJose.EncA192GCM, alg: &cryptoutilJose.AlgRSA15},
		constant.A128GCM_RSA15:      {enc: &cryptoutilJose.EncA128GCM, alg: &cryptoutilJose.AlgRSA15},

		constant.A256GCM_ECDHESA256KW: {enc: &cryptoutilJose.EncA256GCM, alg: &cryptoutilJose.AlgECDHESA256KW},
		constant.A192GCM_ECDHESA256KW: {enc: &cryptoutilJose.EncA192GCM, alg: &cryptoutilJose.AlgECDHESA256KW},
		constant.A128GCM_ECDHESA256KW: {enc: &cryptoutilJose.EncA128GCM, alg: &cryptoutilJose.AlgECDHESA256KW},
		constant.A256GCM_ECDHESA192KW: {enc: &cryptoutilJose.EncA256GCM, alg: &cryptoutilJose.AlgECDHESA192KW},
		constant.A192GCM_ECDHESA192KW: {enc: &cryptoutilJose.EncA192GCM, alg: &cryptoutilJose.AlgECDHESA192KW},
		constant.A128GCM_ECDHESA192KW: {enc: &cryptoutilJose.EncA128GCM, alg: &cryptoutilJose.AlgECDHESA192KW},
		constant.A256GCM_ECDHESA128KW: {enc: &cryptoutilJose.EncA256GCM, alg: &cryptoutilJose.AlgECDHESA128KW},
		constant.A192GCM_ECDHESA128KW: {enc: &cryptoutilJose.EncA192GCM, alg: &cryptoutilJose.AlgECDHESA128KW},
		constant.A128GCM_ECDHESA128KW: {enc: &cryptoutilJose.EncA128GCM, alg: &cryptoutilJose.AlgECDHESA128KW},
		constant.A256GCM_ECDHES:       {enc: &cryptoutilJose.EncA256GCM, alg: &cryptoutilJose.AlgECDHES},
		constant.A192GCM_ECDHES:       {enc: &cryptoutilJose.EncA192GCM, alg: &cryptoutilJose.AlgECDHES},
		constant.A128GCM_ECDHES:       {enc: &cryptoutilJose.EncA128GCM, alg: &cryptoutilJose.AlgECDHES},

		constant.A256CBCHS512_A256KW:    {enc: &cryptoutilJose.EncA256CBC_HS512, alg: &cryptoutilJose.AlgA256KW},
		constant.A192CBCHS384_A256KW:    {enc: &cryptoutilJose.EncA192CBC_HS384, alg: &cryptoutilJose.AlgA256KW},
		constant.A128CBCHS256_A256KW:    {enc: &cryptoutilJose.EncA128CBC_HS256, alg: &cryptoutilJose.AlgA256KW},
		constant.A256CBCHS512_A192KW:    {enc: &cryptoutilJose.EncA256CBC_HS512, alg: &cryptoutilJose.AlgA192KW},
		constant.A192CBCHS384_A192KW:    {enc: &cryptoutilJose.EncA192CBC_HS384, alg: &cryptoutilJose.AlgA192KW},
		constant.A128CBCHS256_A192KW:    {enc: &cryptoutilJose.EncA128CBC_HS256, alg: &cryptoutilJose.AlgA192KW},
		constant.A256CBCHS512_A128KW:    {enc: &cryptoutilJose.EncA256CBC_HS512, alg: &cryptoutilJose.AlgA128KW},
		constant.A192CBCHS384_A128KW:    {enc: &cryptoutilJose.EncA192CBC_HS384, alg: &cryptoutilJose.AlgA128KW},
		constant.A128CBCHS256_A128KW:    {enc: &cryptoutilJose.EncA128CBC_HS256, alg: &cryptoutilJose.AlgA128KW},
		constant.A256CBCHS512_A256GCMKW: {enc: &cryptoutilJose.EncA256CBC_HS512, alg: &cryptoutilJose.AlgA256GCMKW},
		constant.A192CBCHS384_A256GCMKW: {enc: &cryptoutilJose.EncA192CBC_HS384, alg: &cryptoutilJose.AlgA256GCMKW},
		constant.A128CBCHS256_A256GCMKW: {enc: &cryptoutilJose.EncA128CBC_HS256, alg: &cryptoutilJose.AlgA256GCMKW},
		constant.A256CBCHS512_A192GCMKW: {enc: &cryptoutilJose.EncA256CBC_HS512, alg: &cryptoutilJose.AlgA192GCMKW},
		constant.A192CBCHS384_A192GCMKW: {enc: &cryptoutilJose.EncA192CBC_HS384, alg: &cryptoutilJose.AlgA192GCMKW},
		constant.A128CBCHS256_A192GCMKW: {enc: &cryptoutilJose.EncA128CBC_HS256, alg: &cryptoutilJose.AlgA192GCMKW},
		constant.A256CBCHS512_A128GCMKW: {enc: &cryptoutilJose.EncA256CBC_HS512, alg: &cryptoutilJose.AlgA128GCMKW},
		constant.A192CBCHS384_A128GCMKW: {enc: &cryptoutilJose.EncA192CBC_HS384, alg: &cryptoutilJose.AlgA128GCMKW},
		constant.A128CBCHS256_A128GCMKW: {enc: &cryptoutilJose.EncA128CBC_HS256, alg: &cryptoutilJose.AlgA128GCMKW},
		constant.A256CBCHS512_dir:       {enc: &cryptoutilJose.EncA256CBC_HS512, alg: &cryptoutilJose.AlgDir},
		constant.A192CBCHS384_dir:       {enc: &cryptoutilJose.EncA192CBC_HS384, alg: &cryptoutilJose.AlgDir},
		constant.A128CBCHS256_dir:       {enc: &cryptoutilJose.EncA128CBC_HS256, alg: &cryptoutilJose.AlgDir},

		constant.A256CBC_HS512_RSAOAEP512: {enc: &cryptoutilJose.EncA256CBC_HS512, alg: &cryptoutilJose.AlgRSAOAEP512},
		constant.A192CBC_HS384_RSAOAEP512: {enc: &cryptoutilJose.EncA192CBC_HS384, alg: &cryptoutilJose.AlgRSAOAEP512},
		constant.A128CBC_HS256_RSAOAEP512: {enc: &cryptoutilJose.EncA128CBC_HS256, alg: &cryptoutilJose.AlgRSAOAEP512},
		constant.A256CBC_HS512_RSAOAEP384: {enc: &cryptoutilJose.EncA256CBC_HS512, alg: &cryptoutilJose.AlgRSAOAEP384},
		constant.A192CBC_HS384_RSAOAEP384: {enc: &cryptoutilJose.EncA192CBC_HS384, alg: &cryptoutilJose.AlgRSAOAEP384},
		constant.A128CBC_HS256_RSAOAEP384: {enc: &cryptoutilJose.EncA128CBC_HS256, alg: &cryptoutilJose.AlgRSAOAEP384},
		constant.A256CBC_HS512_RSAOAEP256: {enc: &cryptoutilJose.EncA256CBC_HS512, alg: &cryptoutilJose.AlgRSAOAEP256},
		constant.A192CBC_HS384_RSAOAEP256: {enc: &cryptoutilJose.EncA192CBC_HS384, alg: &cryptoutilJose.AlgRSAOAEP256},
		constant.A128CBC_HS256_RSAOAEP256: {enc: &cryptoutilJose.EncA128CBC_HS256, alg: &cryptoutilJose.AlgRSAOAEP256},
		constant.A256CBC_HS512_RSAOAEP:    {enc: &cryptoutilJose.EncA256CBC_HS512, alg: &cryptoutilJose.AlgRSAOAEP},
		constant.A192CBC_HS384_RSAOAEP:    {enc: &cryptoutilJose.EncA192CBC_HS384, alg: &cryptoutilJose.AlgRSAOAEP},
		constant.A128CBC_HS256_RSAOAEP:    {enc: &cryptoutilJose.EncA128CBC_HS256, alg: &cryptoutilJose.AlgRSAOAEP},
		constant.A256CBC_HS512_RSA15:      {enc: &cryptoutilJose.EncA256CBC_HS512, alg: &cryptoutilJose.AlgRSA15},
		constant.A192CBC_HS384_RSA15:      {enc: &cryptoutilJose.EncA192CBC_HS384, alg: &cryptoutilJose.AlgRSA15},
		constant.A128CBC_HS256_RSA15:      {enc: &cryptoutilJose.EncA128CBC_HS256, alg: &cryptoutilJose.AlgRSA15},

		constant.A256CBC_HS512_ECDHESA256KW: {enc: &cryptoutilJose.EncA256CBC_HS512, alg: &cryptoutilJose.AlgECDHESA256KW},
		constant.A192CBC_HS384_ECDHESA256KW: {enc: &cryptoutilJose.EncA192CBC_HS384, alg: &cryptoutilJose.AlgECDHESA256KW},
		constant.A128CBC_HS256_ECDHESA256KW: {enc: &cryptoutilJose.EncA128CBC_HS256, alg: &cryptoutilJose.AlgECDHESA256KW},
		constant.A192CBC_HS384_ECDHESA192KW: {enc: &cryptoutilJose.EncA192CBC_HS384, alg: &cryptoutilJose.AlgECDHESA192KW},
		constant.A128CBC_HS256_ECDHESA192KW: {enc: &cryptoutilJose.EncA128CBC_HS256, alg: &cryptoutilJose.AlgECDHESA192KW},
		constant.A128CBC_HS256_ECDHESA128KW: {enc: &cryptoutilJose.EncA128CBC_HS256, alg: &cryptoutilJose.AlgECDHESA128KW},
		constant.A256CBC_HS512_ECDHES:       {enc: &cryptoutilJose.EncA256CBC_HS512, alg: &cryptoutilJose.AlgECDHES},
		constant.A192CBC_HS384_ECDHES:       {enc: &cryptoutilJose.EncA192CBC_HS384, alg: &cryptoutilJose.AlgECDHES},
		constant.A128CBC_HS256_ECDHES:       {enc: &cryptoutilJose.EncA128CBC_HS256, alg: &cryptoutilJose.AlgECDHES},
	}

	ormElasticKeyAlgorithmToJoseAlg = map[constant.ElasticKeyAlgorithm]*joseJwa.SignatureAlgorithm{
		constant.RS512: &cryptoutilJose.AlgRS512,
		constant.RS384: &cryptoutilJose.AlgRS384,
		constant.RS256: &cryptoutilJose.AlgRS256,
		constant.PS512: &cryptoutilJose.AlgPS512,
		constant.PS384: &cryptoutilJose.AlgPS384,
		constant.PS256: &cryptoutilJose.AlgPS256,
		constant.ES512: &cryptoutilJose.AlgES512,
		constant.ES384: &cryptoutilJose.AlgES384,
		constant.ES256: &cryptoutilJose.AlgES256,
		constant.HS512: &cryptoutilJose.AlgHS512,
		constant.HS384: &cryptoutilJose.AlgHS384,
		constant.HS256: &cryptoutilJose.AlgHS256,
		constant.EdDSA: &cryptoutilJose.AlgEdDSA,
	}

	isSymmetric = map[constant.ElasticKeyAlgorithm]bool{
		constant.A256GCM_A256KW:    true,
		constant.A192GCM_A256KW:    true,
		constant.A128GCM_A256KW:    true,
		constant.A256GCM_A192KW:    true,
		constant.A192GCM_A192KW:    true,
		constant.A128GCM_A192KW:    true,
		constant.A256GCM_A128KW:    true,
		constant.A192GCM_A128KW:    true,
		constant.A128GCM_A128KW:    true,
		constant.A256GCM_A256GCMKW: true,
		constant.A192GCM_A256GCMKW: true,
		constant.A128GCM_A256GCMKW: true,
		constant.A256GCM_A192GCMKW: true,
		constant.A192GCM_A192GCMKW: true,
		constant.A128GCM_A192GCMKW: true,
		constant.A256GCM_A128GCMKW: true,
		constant.A192GCM_A128GCMKW: true,
		constant.A128GCM_A128GCMKW: true,
		constant.A256GCM_dir:       true,
		constant.A192GCM_dir:       true,
		constant.A128GCM_dir:       true,

		constant.A256GCM_RSAOAEP512: false,
		constant.A192GCM_RSAOAEP512: false,
		constant.A128GCM_RSAOAEP512: false,
		constant.A256GCM_RSAOAEP384: false,
		constant.A192GCM_RSAOAEP384: false,
		constant.A128GCM_RSAOAEP384: false,
		constant.A256GCM_RSAOAEP256: false,
		constant.A192GCM_RSAOAEP256: false,
		constant.A128GCM_RSAOAEP256: false,
		constant.A256GCM_RSAOAEP:    false,
		constant.A192GCM_RSAOAEP:    false,
		constant.A128GCM_RSAOAEP:    false,
		constant.A256GCM_RSA15:      false,
		constant.A192GCM_RSA15:      false,
		constant.A128GCM_RSA15:      false,

		constant.A256GCM_ECDHESA256KW: false,
		constant.A192GCM_ECDHESA256KW: false,
		constant.A128GCM_ECDHESA256KW: false,
		constant.A256GCM_ECDHESA192KW: false,
		constant.A192GCM_ECDHESA192KW: false,
		constant.A128GCM_ECDHESA192KW: false,
		constant.A256GCM_ECDHESA128KW: false,
		constant.A192GCM_ECDHESA128KW: false,
		constant.A128GCM_ECDHESA128KW: false,
		constant.A256GCM_ECDHES:       false,
		constant.A192GCM_ECDHES:       false,
		constant.A128GCM_ECDHES:       false,

		constant.A256CBCHS512_A256KW:    true,
		constant.A192CBCHS384_A256KW:    true,
		constant.A128CBCHS256_A256KW:    true,
		constant.A256CBCHS512_A192KW:    true,
		constant.A192CBCHS384_A192KW:    true,
		constant.A128CBCHS256_A192KW:    true,
		constant.A256CBCHS512_A128KW:    true,
		constant.A192CBCHS384_A128KW:    true,
		constant.A128CBCHS256_A128KW:    true,
		constant.A256CBCHS512_A256GCMKW: true,
		constant.A192CBCHS384_A256GCMKW: true,
		constant.A128CBCHS256_A256GCMKW: true,
		constant.A256CBCHS512_A192GCMKW: true,
		constant.A192CBCHS384_A192GCMKW: true,
		constant.A128CBCHS256_A192GCMKW: true,
		constant.A256CBCHS512_A128GCMKW: true,
		constant.A192CBCHS384_A128GCMKW: true,
		constant.A128CBCHS256_A128GCMKW: true,
		constant.A256CBCHS512_dir:       true,
		constant.A192CBCHS384_dir:       true,
		constant.A128CBCHS256_dir:       true,

		constant.A256CBC_HS512_RSAOAEP512: false,
		constant.A192CBC_HS384_RSAOAEP512: false,
		constant.A128CBC_HS256_RSAOAEP512: false,
		constant.A256CBC_HS512_RSAOAEP384: false,
		constant.A192CBC_HS384_RSAOAEP384: false,
		constant.A128CBC_HS256_RSAOAEP384: false,
		constant.A256CBC_HS512_RSAOAEP256: false,
		constant.A192CBC_HS384_RSAOAEP256: false,
		constant.A128CBC_HS256_RSAOAEP256: false,
		constant.A256CBC_HS512_RSAOAEP:    false,
		constant.A192CBC_HS384_RSAOAEP:    false,
		constant.A128CBC_HS256_RSAOAEP:    false,
		constant.A256CBC_HS512_RSA15:      false,
		constant.A192CBC_HS384_RSA15:      false,
		constant.A128CBC_HS256_RSA15:      false,

		constant.A256CBC_HS512_ECDHESA256KW: false,
		constant.A192CBC_HS384_ECDHESA256KW: false,
		constant.A128CBC_HS256_ECDHESA256KW: false,
		constant.A192CBC_HS384_ECDHESA192KW: false,
		constant.A128CBC_HS256_ECDHESA192KW: false,
		constant.A128CBC_HS256_ECDHESA128KW: false,
		constant.A256CBC_HS512_ECDHES:       false,
		constant.A192CBC_HS384_ECDHES:       false,
		constant.A128CBC_HS256_ECDHES:       false,
	}
)

type serviceOrmMapper struct{}

func NewMapper() *serviceOrmMapper {
	return &serviceOrmMapper{}
}

// service => orm

func (m *serviceOrmMapper) toOrmAddElasticKey(elasticKeyID googleUuid.UUID, serviceElasticKeyCreate *cryptoutilBusinessLogicModel.ElasticKeyCreate) *cryptoutilOrmRepository.ElasticKey {
	return &cryptoutilOrmRepository.ElasticKey{
		ElasticKeyID:                elasticKeyID,
		ElasticKeyName:              serviceElasticKeyCreate.Name,
		ElasticKeyDescription:       serviceElasticKeyCreate.Description,
		ElasticKeyProvider:          *m.toOrmElasticKeyProvider(serviceElasticKeyCreate.Provider),
		ElasticKeyAlgorithm:         *m.toOrmElasticKeyAlgorithm(serviceElasticKeyCreate.Algorithm),
		ElasticKeyVersioningAllowed: *serviceElasticKeyCreate.VersioningAllowed,
		ElasticKeyImportAllowed:     *serviceElasticKeyCreate.ImportAllowed,
		ElasticKeyExportAllowed:     *serviceElasticKeyCreate.ExportAllowed,
		ElasticKeyStatus:            *m.toElasticKeyInitialStatus(serviceElasticKeyCreate.ImportAllowed),
	}
}

func (m *serviceOrmMapper) toOrmElasticKeyProvider(serviceElasticKeyProvider *cryptoutilBusinessLogicModel.ElasticKeyProvider) *constant.ElasticKeyProvider {
	ormElasticKeyProvider := constant.ElasticKeyProvider(*serviceElasticKeyProvider)
	return &ormElasticKeyProvider
}

func (m *serviceOrmMapper) toOrmElasticKeyAlgorithm(serviceElasticKeyProvider *cryptoutilBusinessLogicModel.ElasticKeyAlgorithm) *constant.ElasticKeyAlgorithm {
	ormElasticKeyAlgorithm := constant.ElasticKeyAlgorithm(*serviceElasticKeyProvider)
	return &ormElasticKeyAlgorithm
}

func (m *serviceOrmMapper) toElasticKeyInitialStatus(serviceElasticKeyImportAllowed *cryptoutilBusinessLogicModel.ElasticKeyImportAllowed) *constant.ElasticKeyStatus {
	var ormElasticKeyStatus constant.ElasticKeyStatus
	if *serviceElasticKeyImportAllowed {
		ormElasticKeyStatus = constant.ElasticKeyStatus("pending_import")
	} else {
		ormElasticKeyStatus = constant.ElasticKeyStatus("pending_generate")
	}
	return &ormElasticKeyStatus
}

// orm => service

func (m *serviceOrmMapper) toServiceElasticKeys(ormElasticKeys []cryptoutilOrmRepository.ElasticKey) []cryptoutilBusinessLogicModel.ElasticKey {
	serviceElasticKeys := make([]cryptoutilBusinessLogicModel.ElasticKey, len(ormElasticKeys))
	for i, ormElasticKey := range ormElasticKeys {
		serviceElasticKeys[i] = *m.toServiceElasticKey(&ormElasticKey)
	}
	return serviceElasticKeys
}

func (s *serviceOrmMapper) toServiceElasticKey(ormElasticKey *cryptoutilOrmRepository.ElasticKey) *cryptoutilBusinessLogicModel.ElasticKey {
	return &cryptoutilBusinessLogicModel.ElasticKey{
		ElasticKeyID:      (*cryptoutilBusinessLogicModel.ElasticKeyID)(&ormElasticKey.ElasticKeyID),
		Name:              &ormElasticKey.ElasticKeyName,
		Description:       &ormElasticKey.ElasticKeyDescription,
		Algorithm:         s.toServiceElasticKeyAlgorithm(&ormElasticKey.ElasticKeyAlgorithm),
		Provider:          s.toServiceElasticKeyProvider(&ormElasticKey.ElasticKeyProvider),
		VersioningAllowed: &ormElasticKey.ElasticKeyVersioningAllowed,
		ImportAllowed:     &ormElasticKey.ElasticKeyImportAllowed,
		ExportAllowed:     &ormElasticKey.ElasticKeyExportAllowed,
		Status:            s.toServiceElasticKeyStatus(&ormElasticKey.ElasticKeyStatus),
	}
}

func (m *serviceOrmMapper) toServiceElasticKeyAlgorithm(ormElasticKeyAlgorithm *constant.ElasticKeyAlgorithm) *cryptoutilBusinessLogicModel.ElasticKeyAlgorithm {
	serviceElasticKeyAlgorithm := cryptoutilBusinessLogicModel.ElasticKeyAlgorithm(*ormElasticKeyAlgorithm)
	return &serviceElasticKeyAlgorithm
}

func (m *serviceOrmMapper) toServiceElasticKeyProvider(ormElasticKeyProvider *constant.ElasticKeyProvider) *cryptoutilBusinessLogicModel.ElasticKeyProvider {
	serviceElasticKeyProvider := cryptoutilBusinessLogicModel.ElasticKeyProvider(*ormElasticKeyProvider)
	return &serviceElasticKeyProvider
}

func (m *serviceOrmMapper) toServiceElasticKeyStatus(ormElasticKeyStatus *constant.ElasticKeyStatus) *cryptoutilBusinessLogicModel.ElasticKeyStatus {
	serviceElasticKeyStatus := cryptoutilBusinessLogicModel.ElasticKeyStatus(*ormElasticKeyStatus)
	return &serviceElasticKeyStatus
}

func (m *serviceOrmMapper) toServiceKeys(ormKeys []cryptoutilOrmRepository.MaterialKey, repositoryKeyMaterials []*materialKeyExport) ([]cryptoutilBusinessLogicModel.MaterialKey, error) {
	serviceKeys := make([]cryptoutilBusinessLogicModel.MaterialKey, len(ormKeys))
	var serviceKey *cryptoutilBusinessLogicModel.MaterialKey
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

func (m *serviceOrmMapper) toServiceKey(ormKey *cryptoutilOrmRepository.MaterialKey, repositoryKeyMaterial *materialKeyExport) (*cryptoutilBusinessLogicModel.MaterialKey, error) {
	return &cryptoutilBusinessLogicModel.MaterialKey{
		ElasticKeyID:   cryptoutilBusinessLogicModel.ElasticKeyID(ormKey.ElasticKeyID),
		MaterialKeyID:  ormKey.MaterialKeyID,
		GenerateDate:   (*cryptoutilBusinessLogicModel.MaterialKeyGenerateDate)(ormKey.MaterialKeyGenerateDate),
		ImportDate:     (*cryptoutilBusinessLogicModel.MaterialKeyGenerateDate)(ormKey.MaterialKeyImportDate),
		ExpirationDate: (*cryptoutilBusinessLogicModel.MaterialKeyGenerateDate)(ormKey.MaterialKeyExpirationDate),
		RevocationDate: (*cryptoutilBusinessLogicModel.MaterialKeyGenerateDate)(ormKey.MaterialKeyRevocationDate),
		Public:         repositoryKeyMaterial.clearPublic,
		Decrypted:      repositoryKeyMaterial.clearNonPublic,
	}, nil
}

func (m *serviceOrmMapper) toOrmGetElasticKeysQueryParams(params *cryptoutilBusinessLogicModel.ElasticKeysQueryParams) (*cryptoutilOrmRepository.GetElasticKeysFilters, error) {
	if params == nil {
		return nil, nil
	}
	var errs []error
	elasticKeyIDs, err := m.toOptionalOrmUUIDs(params.ElasticKeyID)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Elastic Key ID: %w", err))
	}
	names, err := m.toOptionalOrmStrings(params.Name)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Elastic Key Name: %w", err))
	}
	algorithms, err := m.toOrmAlgorithms(params.Algorithm)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Elastic Key Algorithm: %w", err))
	}
	sorts, err := m.toOrmElasticKeySorts(params.Sort)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Elastic Key Sort: %w", err))
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
		return nil, fmt.Errorf("invalid Get Elastic Keys parameters: %w", errors.Join(errs...))
	}

	return &cryptoutilOrmRepository.GetElasticKeysFilters{
		ElasticKeyID:      elasticKeyIDs,
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

func (m *serviceOrmMapper) toOrmGetMaterialKeysForElasticKeyQueryParams(params *cryptoutilBusinessLogicModel.ElasticKeyMaterialKeysQueryParams) (*cryptoutilOrmRepository.GetElasticKeyMaterialKeysFilters, error) {
	if params == nil {
		return nil, nil
	}
	var errs []error
	materialKeyIDs, err := m.toOptionalOrmUUIDs(params.MaterialKeyID)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid MaterialKeyID: %w", err))
	}
	minGenerateDate, maxGenerateDate, err := m.toOrmDateRange(params.MinGenerateDate, params.MaxGenerateDate)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Generate Date range: %w", err))
	}
	sorts, err := m.toOrmMaterialKeySorts(params.Sort)
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
		return nil, fmt.Errorf("invalid Get Elastic Key Keys parameters: %w", errors.Join(errs...))
	}
	return &cryptoutilOrmRepository.GetElasticKeyMaterialKeysFilters{
		ElasticKeyID:        materialKeyIDs,
		MinimumGenerateDate: minGenerateDate,
		MaximumGenerateDate: maxGenerateDate,
		Sort:                sorts,
		PageNumber:          pageNumber,
		PageSize:            pageSize,
	}, nil
}

func (m *serviceOrmMapper) toOrmGetMaterialKeysQueryParams(params *cryptoutilBusinessLogicModel.MaterialKeysQueryParams) (*cryptoutilOrmRepository.GetMaterialKeysFilters, error) {
	if params == nil {
		return nil, nil
	}
	var errs []error
	elasticKeyIDs, err := m.toOptionalOrmUUIDs(params.ElasticKeyID)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid ElasticKeyID: %w", err))
	}
	materialKeyIDs, err := m.toOptionalOrmUUIDs(params.MaterialKeyID)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid MaterialKeyID: %w", err))
	}
	minGenerateDate, maxGenerateDate, err := m.toOrmDateRange(params.MinGenerateDate, params.MaxGenerateDate)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Generate Date range: %w", err))
	}
	sorts, err := m.toOrmMaterialKeySorts(params.Sort)
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

	return &cryptoutilOrmRepository.GetMaterialKeysFilters{
		ElasticKeyID:        elasticKeyIDs,
		MaterialKeyID:       materialKeyIDs,
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

func (m *serviceOrmMapper) toOrmAlgorithms(algorithms *[]cryptoutilBusinessLogicModel.ElasticKeyAlgorithm) ([]string, error) {
	newVar := toStrings(algorithms, func(algorithm cryptoutilBusinessLogicModel.ElasticKeyAlgorithm) string {
		return string(algorithm)
	})
	return newVar, nil
}

func (m *serviceOrmMapper) toOrmElasticKeySorts(elasticMaterialKeySorts *[]cryptoutilBusinessLogicModel.ElasticKeySort) ([]string, error) {
	newVar := toStrings(elasticMaterialKeySorts, func(elasticMaterialKeySort cryptoutilBusinessLogicModel.ElasticKeySort) string {
		return string(elasticMaterialKeySort)
	})
	return newVar, nil
}

func (m *serviceOrmMapper) toOrmMaterialKeySorts(keySorts *[]cryptoutilBusinessLogicModel.MaterialKeySort) ([]string, error) {
	newVar := toStrings(keySorts, func(keySort cryptoutilBusinessLogicModel.MaterialKeySort) string { return string(keySort) })
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

func (m *serviceOrmMapper) isJwe(ormElasticKeyAlgorithm *constant.ElasticKeyAlgorithm) bool {
	_, ok := ormElasticKeyAlgorithmToJoseEncAndAlg[*ormElasticKeyAlgorithm]
	return ok
}

func (m *serviceOrmMapper) toJweEncAndAlg(ormElasticKeyAlgorithm *constant.ElasticKeyAlgorithm) (*joseJwa.ContentEncryptionAlgorithm, *joseJwa.KeyEncryptionAlgorithm, error) {
	if encAndAlg, ok := ormElasticKeyAlgorithmToJoseEncAndAlg[*ormElasticKeyAlgorithm]; ok {
		return encAndAlg.enc, encAndAlg.alg, nil
	}
	return nil, nil, fmt.Errorf("unsupported JWE ElasticKeyAlgorithm '%s'", *ormElasticKeyAlgorithm)
}

func (m *serviceOrmMapper) isJws(ormElasticKeyAlgorithm *constant.ElasticKeyAlgorithm) bool {
	_, ok := ormElasticKeyAlgorithmToJoseAlg[*ormElasticKeyAlgorithm]
	return ok
}

func (m *serviceOrmMapper) toJwsAlg(ormElasticKeyAlgorithm *constant.ElasticKeyAlgorithm) (*joseJwa.SignatureAlgorithm, error) {
	if alg, ok := ormElasticKeyAlgorithmToJoseAlg[*ormElasticKeyAlgorithm]; ok {
		return alg, nil
	}
	return nil, fmt.Errorf("unsupported JWS ElasticKeyAlgorithm '%s'", *ormElasticKeyAlgorithm)
}

func (m *serviceOrmMapper) isSymmetric(ormElasticKeyAlgorithm *constant.ElasticKeyAlgorithm) (bool, error) {
	isSymmetric, ok := isSymmetric[*ormElasticKeyAlgorithm]
	if ok {
		return isSymmetric, nil
	}
	return false, fmt.Errorf("unsupported ElasticKeyAlgorithm '%s'", *ormElasticKeyAlgorithm)
}

func (m *serviceOrmMapper) isAsymmetric(ormElasticKeyAlgorithm *constant.ElasticKeyAlgorithm) (bool, error) {
	isSymmetric, ok := isSymmetric[*ormElasticKeyAlgorithm]
	if ok {
		return !isSymmetric, nil
	}
	return false, fmt.Errorf("unsupported ElasticKeyAlgorithm '%s'", *ormElasticKeyAlgorithm)
}
