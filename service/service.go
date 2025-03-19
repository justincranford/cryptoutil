package service

import (
	"context"
	"crypto/rand"
	"cryptoutil/api/openapi"
	"cryptoutil/orm"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type KEKPoolService struct {
	// dbService *database.Service
	dbService *orm.Service
}

func NewService(dbService *orm.Service) *KEKPoolService {
	return &KEKPoolService{dbService: dbService}
}

func (service *KEKPoolService) PostKekpool(ctx context.Context, request openapi.PostKekpoolRequestObject) (openapi.PostKekpoolResponseObject, error) {
	kekPoolStatus := "pending_generate"
	if *request.Body.IsImportAllowed {
		kekPoolStatus = "pending_import"
	}
	gormKekPool := orm.KEKPool{
		KEKPoolName:                request.Body.Name,
		KEKPoolDescription:         request.Body.Description,
		KEKPoolAlgorithm:           string(*request.Body.Algorithm),
		KEKPoolProvider:            string(*request.Body.Provider),
		KEKPoolIsVersioningAllowed: *request.Body.IsVersioningAllowed,
		KEKPoolIsImportAllowed:     *request.Body.IsImportAllowed,
		KEKPoolIsExportAllowed:     *request.Body.IsExportAllowed,
		KEKPoolStatus:              kekPoolStatus,
	}

	result := service.dbService.GormDB.Create(&gormKekPool)
	if result.Error != nil {
		return nil, result.Error
	}

	// Map the GORM model to the OpenAPI model (API response)
	kekPoolID := gormKekPool.KEKPoolID.String()
	kekPoolAlgorithm := openapi.KEKPoolAlgorithm(gormKekPool.KEKPoolAlgorithm)
	kekPoolProvider := openapi.KEKPoolProvider(gormKekPool.KEKPoolProvider)
	kekPoolStatus2 := openapi.KEKPoolStatus(gormKekPool.KEKPoolStatus)
	openapiKekPool := openapi.PostKekpool200JSONResponse{
		Id:                  &kekPoolID,
		Name:                &gormKekPool.KEKPoolName,
		Description:         &gormKekPool.KEKPoolDescription,
		Algorithm:           &kekPoolAlgorithm,
		Provider:            &kekPoolProvider,
		IsVersioningAllowed: &gormKekPool.KEKPoolIsVersioningAllowed,
		IsImportAllowed:     &gormKekPool.KEKPoolIsImportAllowed,
		IsExportAllowed:     &gormKekPool.KEKPoolIsExportAllowed,
		Status:              &kekPoolStatus2,
	}

	return &openapiKekPool, nil
}

func (service *KEKPoolService) GetKEKPool(ctx context.Context, request openapi.GetKekpoolRequestObject) (openapi.GetKekpoolResponseObject, error) {
	var gormKekPools []orm.KEKPool
	result := service.dbService.GormDB.Find(&gormKekPools)
	if result.Error != nil {
		return nil, result.Error
	}

	kekPools := make([]openapi.KEKPool, len(gormKekPools))
	for i, gormKekPool := range gormKekPools {
		algorithm := openapi.KEKPoolAlgorithm(gormKekPool.KEKPoolAlgorithm)
		provider := openapi.KEKPoolProvider(gormKekPool.KEKPoolProvider)
		status := openapi.KEKPoolStatus(gormKekPool.KEKPoolStatus)

		kekPoolID := gormKekPool.KEKPoolID.String()
		kekPools[i] = openapi.KEKPool{
			Id:                  &kekPoolID,
			Name:                &gormKekPool.KEKPoolName,
			Description:         &gormKekPool.KEKPoolDescription,
			Algorithm:           &algorithm,
			Provider:            &provider,
			IsVersioningAllowed: &gormKekPool.KEKPoolIsVersioningAllowed,
			IsImportAllowed:     &gormKekPool.KEKPoolIsImportAllowed,
			IsExportAllowed:     &gormKekPool.KEKPoolIsExportAllowed,
			Status:              &status,
		}
	}

	response := openapi.GetKekpool200JSONResponse(kekPools)
	return &response, nil
}

func (service *KEKPoolService) PostKekpoolKekPoolIDKek(ctx context.Context, request openapi.PostKekpoolKekPoolIDKekRequestObject) (openapi.PostKekpoolKekPoolIDKekResponseObject, error) {
	kekPoolID, err := uuid.Parse(request.KekPoolID)
	if err != nil {
		return &openapi.PostKekpoolKekPoolIDKek400JSONResponse{HTTP400JSONResponse: openapi.HTTP400JSONResponse{Error: stringPtr("KEK Pool ID")}}, nil
	}

	var kekPool orm.KEKPool
	result := service.dbService.GormDB.First(&kekPool, "kek_pool_id = ?", request.KekPoolID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return &openapi.PostKekpoolKekPoolIDKek404JSONResponse{HTTP404JSONResponse: openapi.HTTP404JSONResponse{Error: stringPtr("KEK Pool not found")}}, nil
		}
		return nil, result.Error
	}

	if kekPool.KEKPoolStatus != string(openapi.Active) && kekPool.KEKPoolStatus != string(openapi.PendingGenerate) {
		return &openapi.PostKekpoolKekPoolIDKek400JSONResponse{HTTP400JSONResponse: openapi.HTTP400JSONResponse{Error: stringPtr("KEK Pool invalid state")}}, nil
	}

	var maxID int
	service.dbService.GormDB.Model(&orm.KEK{}).Where("kek_pool_id = ?", request.KekPoolID).Select("COALESCE(MAX(kek_id), 0)").Scan(&maxID)
	nextKekId := maxID + 1
	generateDate := time.Now().UTC()

	var keyMaterial []byte
	switch kekPool.KEKPoolAlgorithm {
	case string(openapi.AES256):
		keyMaterial = make([]byte, 32)
	case string(openapi.AES192):
		keyMaterial = make([]byte, 24)
	case string(openapi.AES128):
		keyMaterial = make([]byte, 16)
	default:
		return &openapi.PostKekpoolKekPoolIDKek500JSONResponse{HTTP500JSONResponse: openapi.HTTP500JSONResponse{Error: stringPtr("KEK Pool invalid algorithm")}}, nil
	}
	_, err = rand.Read(keyMaterial)
	if err != nil {
		return &openapi.PostKekpoolKekPoolIDKek500JSONResponse{HTTP500JSONResponse: openapi.HTTP500JSONResponse{Error: stringPtr("Failed to generate key material")}}, nil
	}

	newVar := generateDate.Format("2006-01-02T15:04:05Z")
	gormKek := orm.KEK{
		KEKPoolID:       kekPoolID,
		KEKID:           nextKekId,
		KEKMaterial:     keyMaterial,
		KEKGenerateDate: &newVar,
	}

	result = service.dbService.GormDB.Create(&gormKek)
	if result.Error != nil {
		return nil, result.Error
	}

	kekResponse := openapi.PostKekpoolKekPoolIDKek200JSONResponse{
		KekId:        &gormKek.KEKID,
		KekPoolId:    &request.KekPoolID,
		GenerateDate: &generateDate,
	}

	return &kekResponse, nil
}

func (service *KEKPoolService) GetKekpoolKekPoolIDKek(ctx context.Context, request openapi.GetKekpoolKekPoolIDKekRequestObject) (openapi.GetKekpoolKekPoolIDKekResponseObject, error) {
	_, err := uuid.Parse(request.KekPoolID)
	if err != nil {
		return &openapi.GetKekpoolKekPoolIDKek400JSONResponse{HTTP400JSONResponse: openapi.HTTP400JSONResponse{Error: stringPtr("KEK Pool ID")}}, nil
	}

	var kekPool orm.KEKPool
	result := service.dbService.GormDB.First(&kekPool, "kek_pool_id = ?", request.KekPoolID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return &openapi.GetKekpoolKekPoolIDKek404JSONResponse{HTTP404JSONResponse: openapi.HTTP404JSONResponse{Error: stringPtr("KEK Pool not found")}}, nil
		}
		return nil, result.Error
	}
	var gormKeks []orm.KEK
	query := service.dbService.GormDB.Where("kek_pool_id = ?", request.KekPoolID)
	result = query.Find(&gormKeks)
	if result.Error != nil {
		return nil, result.Error
	}

	keks := make([]openapi.KEK, len(gormKeks))
	for i, gormKek := range gormKeks {
		var openapiKEKGenerateDate *time.Time
		if gormKek.KEKGenerateDate != nil {
			*openapiKEKGenerateDate, err = time.Parse("2006-01-02T15:04:05Z", *gormKek.KEKGenerateDate)
			if err != nil {
				return &openapi.GetKekpoolKekPoolIDKek500JSONResponse{HTTP500JSONResponse: openapi.HTTP500JSONResponse{Error: stringPtr("KEK generate date parse error")}}, nil
			}
		}
		kek := openapi.KEK{
			KekId:        &gormKek.KEKID,
			KekPoolId:    &request.KekPoolID,
			GenerateDate: openapiKEKGenerateDate,
		}
		keks[i] = kek
	}

	keksResponse := openapi.GetKekpoolKekPoolIDKek200JSONResponse(keks)
	return &keksResponse, nil
}

// Helper function to get string pointer
func stringPtr(s string) *string {
	return &s
}
