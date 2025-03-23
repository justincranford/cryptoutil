package service

import (
	"cryptoutil/api/openapi"
	"cryptoutil/orm"
	"cryptoutil/util"
)

type OpenapiOrmMapper struct{}

func NewMapper() *OpenapiOrmMapper {
	return &OpenapiOrmMapper{}
}

func (m *OpenapiOrmMapper) toOpenapiKEKPools(gormKEKPools []orm.KEKPool) []openapi.KEKPool {
	openapiKEKPools := make([]openapi.KEKPool, len(gormKEKPools))
	for i, gormKekPool := range gormKEKPools {
		openapiKEKPools[i] = m.toOpenapiKEKPool(gormKekPool)
	}
	return openapiKEKPools
}

func (*OpenapiOrmMapper) toOpenapiKEKPool(gormKekPool orm.KEKPool) openapi.KEKPool {
	openapiKekPool := openapi.KEKPool{
		Id:                  (*openapi.KEKPoolId)(util.StringPtr(gormKekPool.KEKPoolID.String())),
		Name:                &gormKekPool.KEKPoolName,
		Description:         &gormKekPool.KEKPoolDescription,
		Algorithm:           (*openapi.KEKPoolAlgorithm)(&gormKekPool.KEKPoolAlgorithm),
		Provider:            (*openapi.KEKPoolProvider)(&gormKekPool.KEKPoolProvider),
		IsVersioningAllowed: &gormKekPool.KEKPoolIsVersioningAllowed,
		IsImportAllowed:     &gormKekPool.KEKPoolIsImportAllowed,
		IsExportAllowed:     &gormKekPool.KEKPoolIsExportAllowed,
		Status:              (*openapi.KEKPoolStatus)(&gormKekPool.KEKPoolStatus),
	}
	return openapiKekPool
}

func (m *OpenapiOrmMapper) toOpenapiKEKs(gormKEKs []orm.KEK) []openapi.KEK {
	openapiKEKs := make([]openapi.KEK, len(gormKEKs))
	for i, gormKEK := range gormKEKs {
		openapiKEKs[i] = m.toOpenapiKEK(gormKEK)
	}
	return openapiKEKs
}

func (*OpenapiOrmMapper) toOpenapiKEK(gormKEK orm.KEK) openapi.KEK {
	openapiKEKResponse := openapi.KEK{
		KekId:        &gormKEK.KEKID,
		KekPoolId:    (*openapi.KEKPoolId)(util.StringPtr(gormKEK.KEKPoolID.String())),
		GenerateDate: (*openapi.KEKGenerateDate)(gormKEK.KEKGenerateDate),
	}
	return openapiKEKResponse
}

func (*OpenapiOrmMapper) toKEKPoolProviderEnum(openapiKEKPoolProvider openapi.KEKPoolProvider) orm.KEKPoolProviderEnum {
	return orm.KEKPoolProviderEnum(openapiKEKPoolProvider)
}

func (*OpenapiOrmMapper) toKKEKPoolAlgorithmEnum(openapiKEKPoolProvider openapi.KEKPoolAlgorithm) orm.KEKPoolAlgorithmEnum {
	return orm.KEKPoolAlgorithmEnum(openapiKEKPoolProvider)
}

func (*OpenapiOrmMapper) toKEKPoolStatusImportOrGenerate(openapiKEKPoolIsImportAllowed openapi.KEKPoolIsImportAllowed) orm.KEKPoolStatusEnum {
	if openapiKEKPoolIsImportAllowed {
		return orm.KEKPoolStatusEnum("pending_import")
	}
	return orm.KEKPoolStatusEnum("pending_generate")
}
