// Copyright (c) 2025-2026 Justin Cranford.
package orm

import (
	"log/slog"

	cryptoutilKmsServer "cryptoutil/api/sm-kms/server"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"

	"gorm.io/gorm"

	googleUuid "github.com/google/uuid"
)

// Error messages for elastic key and material key operations.
var (
	ErrFailedToAddElasticKey                                = "failed to add Elastic Key"
	ErrFailedToGetElasticKeyByElasticKeyID                  = "failed to get Elastic Key by Elastic Key ID"
	ErrFailedToUpdateElasticKeyByElasticKeyID               = "failed to update Elastic Key by Elastic Key ID"
	ErrFailedToUpdateElasticKeyStatusByElasticKeyID         = "failed to update Elastic Key Status by Elastic Key ID"
	ErrFailedToGetElasticKeys                               = "failed to get Elastic Keys"
	ErrFailedToAddMaterialKey                               = "failed to add Material Key"
	ErrFailedToGetMaterialKeysByElasticKeyID                = "failed to get Keys by Elastic Key ID"
	ErrInvalidElasticKeyID                                  = "invalid Elastic Key ID"
	ErrInvalidMaterialKeyID                                 = "invalid Material Key ID"
	ErrFailedToGetMaterialKeys                              = "failed to get Material Keys"
	ErrFailedToGetMaterialKeyByElasticKeyIDAndMaterialKeyID = "failed to get Material Key by Elastic Key ID and Material Key ID"
	ErrFailedToGetLatestMaterialKeyByElasticKeyID           = "failed to get latest Material Key by Elastic Key ID"
	ErrFailedToUpdateMaterialKey                            = "failed to update Material Key"
)

// AddElasticKey adds a new elastic key to the database.
func AddElasticKey(gormTx *gorm.DB, slogger *slog.Logger, elasticKey *ElasticKey) error {
	if err := cryptoutilSharedUtilRandom.ValidateUUID(&elasticKey.ElasticKeyID, ErrInvalidElasticKeyID); err != nil {
		return toAppErr(slogger, &ErrFailedToAddElasticKey, err)
	}

	if err := gormTx.Create(elasticKey).Error; err != nil {
		return toAppErr(slogger, &ErrFailedToAddElasticKey, err)
	}

	return nil
}

// GetElasticKey retrieves an elastic key by ID from the database, filtered by tenant.
func GetElasticKey(gormTx *gorm.DB, slogger *slog.Logger, tenantID googleUuid.UUID, elasticKeyID *googleUuid.UUID) (*ElasticKey, error) {
	if err := cryptoutilSharedUtilRandom.ValidateUUID(elasticKeyID, ErrInvalidElasticKeyID); err != nil {
		return nil, toAppErr(slogger, &ErrFailedToGetElasticKeyByElasticKeyID, err)
	}

	var elasticKey ElasticKey

	if err := gormTx.First(&elasticKey, "tenant_id=? AND elastic_key_id=?", tenantID, elasticKeyID).Error; err != nil {
		return nil, toAppErr(slogger, &ErrFailedToGetElasticKeyByElasticKeyID, err)
	}

	return &elasticKey, nil
}

// UpdateElasticKey updates an existing elastic key in the database.
func UpdateElasticKey(gormTx *gorm.DB, slogger *slog.Logger, elasticKey *ElasticKey) error {
	if err := cryptoutilSharedUtilRandom.ValidateUUID(&elasticKey.ElasticKeyID, ErrInvalidElasticKeyID); err != nil {
		return toAppErr(slogger, &ErrFailedToUpdateElasticKeyByElasticKeyID, err)
	}

	if err := gormTx.UpdateColumns(elasticKey).Error; err != nil {
		return toAppErr(slogger, &ErrFailedToUpdateElasticKeyByElasticKeyID, err)
	}

	return nil
}

// UpdateElasticKeyStatus updates the status of an elastic key in the database.
func UpdateElasticKeyStatus(gormTx *gorm.DB, slogger *slog.Logger, elasticKeyID googleUuid.UUID, elasticKeyStatus cryptoutilKmsServer.ElasticKeyStatus) error {
	if err := cryptoutilSharedUtilRandom.ValidateUUID(&elasticKeyID, ErrInvalidElasticKeyID); err != nil {
		return toAppErr(slogger, &ErrFailedToUpdateElasticKeyStatusByElasticKeyID, err)
	}

	if err := gormTx.Model(&ElasticKey{}).Where("elastic_key_id=?", elasticKeyID).Update("elastic_key_status", elasticKeyStatus).Error; err != nil {
		return toAppErr(slogger, &ErrFailedToUpdateElasticKeyStatusByElasticKeyID, err)
	}

	return nil
}

// GetElasticKeys retrieves elastic keys with optional filters from the database.
func GetElasticKeys(gormTx *gorm.DB, slogger *slog.Logger, getElasticKeysFilters *GetElasticKeysFilters) ([]ElasticKey, error) {
	var elasticKeys []ElasticKey

	if err := applyGetElasticKeysFilters(gormTx, getElasticKeysFilters).Find(&elasticKeys).Error; err != nil {
		return nil, toAppErr(slogger, &ErrFailedToGetElasticKeys, err)
	}

	return elasticKeys, nil
}

// AddElasticKeyMaterialKey adds a new material key for an elastic key to the database.
func AddElasticKeyMaterialKey(gormTx *gorm.DB, slogger *slog.Logger, key *MaterialKey) error {
	if err := cryptoutilSharedUtilRandom.ValidateUUID(&key.ElasticKeyID, ErrInvalidElasticKeyID); err != nil {
		return toAppErr(slogger, &ErrFailedToAddMaterialKey, err)
	} else if err := cryptoutilSharedUtilRandom.ValidateUUID(&key.MaterialKeyID, ErrInvalidMaterialKeyID); err != nil {
		return toAppErr(slogger, &ErrFailedToAddMaterialKey, err)
	}

	if err := gormTx.Create(key).Error; err != nil {
		return toAppErr(slogger, &ErrFailedToAddMaterialKey, err)
	}

	return nil
}

// GetMaterialKeysForElasticKey retrieves material keys for an elastic key with optional filters.
func GetMaterialKeysForElasticKey(gormTx *gorm.DB, slogger *slog.Logger, elasticKeyID *googleUuid.UUID, getElasticKeyKeysFilters *GetElasticKeyMaterialKeysFilters) ([]MaterialKey, error) {
	if err := cryptoutilSharedUtilRandom.ValidateUUID(elasticKeyID, ErrFailedToGetMaterialKeysByElasticKeyID); err != nil {
		return nil, toAppErr(slogger, &ErrFailedToGetMaterialKeysByElasticKeyID, err)
	}

	var keys []MaterialKey

	query := gormTx.Where("elastic_key_id=?", elasticKeyID)

	if err := applyGetElasticKeyKeysFilters(query, getElasticKeyKeysFilters).Find(&keys).Error; err != nil {
		return nil, toAppErr(slogger, &ErrFailedToGetMaterialKeysByElasticKeyID, err)
	}

	return keys, nil
}

// GetMaterialKeys retrieves material keys based on the provided filter criteria.
func GetMaterialKeys(gormTx *gorm.DB, slogger *slog.Logger, getKeysFilters *GetMaterialKeysFilters) ([]MaterialKey, error) {
	var keys []MaterialKey

	if err := applyKeyFilters(gormTx, getKeysFilters).Find(&keys).Error; err != nil {
		return nil, toAppErr(slogger, &ErrFailedToGetMaterialKeys, err)
	}

	return keys, nil
}

// GetElasticKeyMaterialKeyVersion retrieves a specific version of a material key by elastic key ID and material key ID.
func GetElasticKeyMaterialKeyVersion(gormTx *gorm.DB, slogger *slog.Logger, elasticKeyID, materialKeyID *googleUuid.UUID) (*MaterialKey, error) {
	if err := cryptoutilSharedUtilRandom.ValidateUUID(elasticKeyID, ErrInvalidElasticKeyID); err != nil {
		return nil, toAppErr(slogger, &ErrFailedToGetMaterialKeyByElasticKeyIDAndMaterialKeyID, err)
	} else if err := cryptoutilSharedUtilRandom.ValidateUUID(materialKeyID, ErrInvalidMaterialKeyID); err != nil {
		return nil, toAppErr(slogger, &ErrFailedToGetMaterialKeyByElasticKeyIDAndMaterialKeyID, err)
	}

	var key MaterialKey

	if err := gormTx.First(&key, "elastic_key_id=? AND material_key_id=?", elasticKeyID, materialKeyID).Error; err != nil {
		return nil, toAppErr(slogger, &ErrFailedToGetMaterialKeyByElasticKeyIDAndMaterialKeyID, err)
	}

	return &key, nil
}

// GetElasticKeyMaterialKeyLatest retrieves the latest material key for the given elastic key ID.
func GetElasticKeyMaterialKeyLatest(gormTx *gorm.DB, slogger *slog.Logger, elasticKeyID googleUuid.UUID) (*MaterialKey, error) {
	if err := cryptoutilSharedUtilRandom.ValidateUUID(&elasticKeyID, ErrInvalidElasticKeyID); err != nil {
		return nil, toAppErr(slogger, &ErrFailedToGetLatestMaterialKeyByElasticKeyID, err)
	}

	var key MaterialKey

	if err := gormTx.Order("material_key_id DESC").First(&key, "elastic_key_id=?", elasticKeyID).Error; err != nil {
		return nil, toAppErr(slogger, &ErrFailedToGetLatestMaterialKeyByElasticKeyID, err)
	}

	return &key, nil
}

// UpdateElasticKeyMaterialKeyRevoke updates the revocation date for a material key.
func UpdateElasticKeyMaterialKeyRevoke(gormTx *gorm.DB, slogger *slog.Logger, materialKey *MaterialKey) error {
	if err := cryptoutilSharedUtilRandom.ValidateUUID(&materialKey.ElasticKeyID, ErrInvalidElasticKeyID); err != nil {
		return toAppErr(slogger, &ErrFailedToUpdateMaterialKey, err)
	} else if err := cryptoutilSharedUtilRandom.ValidateUUID(&materialKey.MaterialKeyID, ErrInvalidMaterialKeyID); err != nil {
		return toAppErr(slogger, &ErrFailedToUpdateMaterialKey, err)
	}

	err := gormTx.Model(&MaterialKey{}).
		Where("elastic_key_id=? AND material_key_id=?", materialKey.ElasticKeyID, materialKey.MaterialKeyID).
		Update("material_key_revocation_date", materialKey.MaterialKeyRevocationDate).Error
	if err != nil {
		return toAppErr(slogger, &ErrFailedToUpdateMaterialKey, err)
	}

	return nil
}

func toAppErr(slogger *slog.Logger, msg *string, err error) error {
	return cryptoutilSharedApperr.MapGormError(slogger, msg, err)
}

func applyGetElasticKeysFilters(db *gorm.DB, filters *GetElasticKeysFilters) *gorm.DB {
	if filters == nil {
		return db
	}

	db = db.Where("tenant_id=?", filters.TenantID)

	if len(filters.ElasticKeyID) > 0 {
		db = db.Where("elastic_key_id IN ?", filters.ElasticKeyID)
	}

	if len(filters.Name) > 0 {
		db = db.Where("elastic_key_name IN ?", filters.Name)
	}

	if len(filters.Algorithm) > 0 {
		db = db.Where("elastic_key_algorithm IN ?", filters.Algorithm)
	}

	if filters.VersioningAllowed != nil {
		db = db.Where("elastic_key_versioning_allowed=?", *filters.VersioningAllowed)
	}

	if filters.ImportAllowed != nil {
		db = db.Where("elastic_key_import_allowed=?", *filters.ImportAllowed)
	}

	if filters.ExportAllowed != nil {
		db = db.Where("elastic_key_export_allowed=?", *filters.ExportAllowed)
	}

	if len(filters.Sort) > 0 {
		for _, sort := range filters.Sort {
			db = db.Order(sort)
		}
	}

	if filters.PageSize > 0 {
		db = db.Offset(filters.PageNumber * filters.PageSize)
		db = db.Limit(filters.PageSize)
	}

	return db
}

func applyKeyFilters(db *gorm.DB, filters *GetMaterialKeysFilters) *gorm.DB {
	if filters == nil {
		return db
	}

	if len(filters.MaterialKeyID) > 0 {
		db = db.Where("material_key_id IN ?", filters.MaterialKeyID)
	}

	if len(filters.ElasticKeyID) > 0 {
		db = db.Where("elastic_key_id IN ?", filters.ElasticKeyID)
	}

	if filters.MinimumGenerateDate != nil {
		db = db.Where("material_key_generate_date>=?", *filters.MinimumGenerateDate)
	}

	if filters.MaximumGenerateDate != nil {
		db = db.Where("material_key_generate_date<=?", *filters.MaximumGenerateDate)
	}

	if len(filters.Sort) > 0 {
		for _, sort := range filters.Sort {
			db = db.Order(sort)
		}
	}

	if filters.PageSize > 0 {
		db = db.Offset(filters.PageNumber * filters.PageSize)
		db = db.Limit(filters.PageSize)
	}

	return db
}

func applyGetElasticKeyKeysFilters(db *gorm.DB, filters *GetElasticKeyMaterialKeysFilters) *gorm.DB {
	if filters == nil {
		return db
	}

	if len(filters.ElasticKeyID) > 0 {
		db = db.Where("elastic_key_id IN ?", filters.ElasticKeyID)
	}

	if filters.MinimumGenerateDate != nil {
		db = db.Where("material_key_generate_date>=?", *filters.MinimumGenerateDate)
	}

	if filters.MaximumGenerateDate != nil {
		db = db.Where("material_key_generate_date<=?", *filters.MaximumGenerateDate)
	}

	if len(filters.Sort) > 0 {
		for _, sort := range filters.Sort {
			db = db.Order(sort)
		}
	}

	if filters.PageSize > 0 {
		db = db.Offset(filters.PageNumber * filters.PageSize)
		db = db.Limit(filters.PageSize)
	}

	return db
}
