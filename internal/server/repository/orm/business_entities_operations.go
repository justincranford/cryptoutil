package orm

import (
	"errors"
	"fmt"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilBusinessModel "cryptoutil/internal/common/businessmodel"
	cryptoutilUtil "cryptoutil/internal/common/util"

	"gorm.io/gorm"

	googleUuid "github.com/google/uuid"
	"github.com/lib/pq"
	"modernc.org/sqlite"
)

// Service-Repository calls

func (tx *OrmTransaction) AddElasticKey(elasticKey *ElasticKey) error {
	if err := cryptoutilUtil.ValidateUUID(&elasticKey.ElasticKeyID, "invalid Elastic Key ID"); err != nil {
		return tx.toAppErr("failed to add Elastic Key", err)
	}
	err := tx.state.gormTx.Create(elasticKey).Error
	if err != nil {
		return tx.toAppErr("failed to add Elastic Key", err)
	}
	return nil
}

func (tx *OrmTransaction) GetElasticKey(elasticKeyID googleUuid.UUID) (*ElasticKey, error) {
	if err := cryptoutilUtil.ValidateUUID(&elasticKeyID, "invalid Elastic Key ID"); err != nil {
		return nil, tx.toAppErr("failed to get Elastic Key by Elastic Key ID", err)
	}
	var elasticKey ElasticKey
	err := tx.state.gormTx.First(&elasticKey, "elastic_key_id=?", elasticKeyID).Error
	if err != nil {
		return nil, tx.toAppErr("failed to get Elastic Key by Elastic Key ID", err)
	}
	return &elasticKey, nil
}

func (tx *OrmTransaction) UpdateElasticKey(elasticKey *ElasticKey) error {
	if err := cryptoutilUtil.ValidateUUID(&elasticKey.ElasticKeyID, "invalid Elastic Key ID"); err != nil {
		return tx.toAppErr("failed to update Elastic Key by Elastic Key ID", err)
	}
	err := tx.state.gormTx.UpdateColumns(elasticKey).Error
	if err != nil {
		return tx.toAppErr("failed to update Elastic Key", err)
	}
	return nil
}

func (tx *OrmTransaction) UpdateElasticKeyStatus(elasticKeyID googleUuid.UUID, elasticKeyStatus cryptoutilBusinessModel.ElasticKeyStatus) error {
	if err := cryptoutilUtil.ValidateUUID(&elasticKeyID, "invalid Elastic Key ID"); err != nil {
		return tx.toAppErr("failed to update Elastic Key Status by Elastic Key ID", err)
	}
	err := tx.state.gormTx.Model(&ElasticKey{}).Where("elastic_key_id=?", elasticKeyID).Update("elastic_key_status", elasticKeyStatus).Error
	if err != nil {
		return tx.toAppErr("failed to update Elastic Key Status", err)
	}
	return nil
}

func (tx *OrmTransaction) GetElasticKeys(getElasticKeysFilters *GetElasticKeysFilters) ([]ElasticKey, error) {
	var elasticKeys []ElasticKey
	query := tx.state.gormTx
	err := applyGetElasticKeysFilters(query, getElasticKeysFilters).Find(&elasticKeys).Error
	if err != nil {
		return nil, tx.toAppErr("failed to get Elastic Keys", err)
	}
	return elasticKeys, nil
}

func (tx *OrmTransaction) AddElasticKeyKey(key *MaterialKey) error {
	if err := cryptoutilUtil.ValidateUUID(&key.ElasticKeyID, "invalid Elastic Key ID"); err != nil {
		return tx.toAppErr("failed to add Key", err)
	} else if err := cryptoutilUtil.ValidateUUID(&key.MaterialKeyID, "invalid Key ID"); err != nil {
		return tx.toAppErr("failed to add Key", err)
	}
	err := tx.state.gormTx.Create(key).Error
	if err != nil {
		return tx.toAppErr("failed to add Key", err)
	}
	return nil
}

func (tx *OrmTransaction) GetMaterialKeysForElasticKey(elasticKeyID googleUuid.UUID, getElasticKeyKeysFilters *GetElasticKeyMaterialKeysFilters) ([]MaterialKey, error) {
	if err := cryptoutilUtil.ValidateUUID(&elasticKeyID, "failed to get Keys by Elastic Key ID"); err != nil {
		return nil, tx.toAppErr("invalid Elastic Key ID", err)
	}
	var keys []MaterialKey
	query := tx.state.gormTx.Where("elastic_key_id=?", elasticKeyID)
	err := applyGetElasticKeyKeysFilters(query, getElasticKeyKeysFilters).Find(&keys).Error
	if err != nil {
		return nil, tx.toAppErr("failed to get Keys by Elastic Key ID", err)
	}
	return keys, nil
}

func (tx *OrmTransaction) GetMaterialKeys(getKeysFilters *GetMaterialKeysFilters) ([]MaterialKey, error) {
	var keys []MaterialKey
	query := tx.state.gormTx
	err := applyKeyFilters(query, getKeysFilters).Find(&keys).Error
	if err != nil {
		return nil, tx.toAppErr("failed to get Keys", err)
	}
	return keys, nil
}

func (tx *OrmTransaction) GetElasticKeyMaterialKeyVersion(elasticKeyID googleUuid.UUID, materialKeyID googleUuid.UUID) (*MaterialKey, error) {
	if err := cryptoutilUtil.ValidateUUID(&elasticKeyID, "invalid Elastic Key ID"); err != nil {
		return nil, tx.toAppErr("failed to get Key by Elastic Key ID and Key ID", err)
	} else if err := cryptoutilUtil.ValidateUUID(&materialKeyID, "invalid Key ID"); err != nil {
		return nil, tx.toAppErr("failed to get Key by Elastic Key ID and Key ID", err)
	}
	var key MaterialKey
	err := tx.state.gormTx.First(&key, "elastic_key_id=? AND material_key_id=?", elasticKeyID, materialKeyID).Error
	if err != nil {
		return nil, tx.toAppErr("failed to get Key by Elastic Key ID and Key ID", err)
	}
	return &key, nil
}

func (tx *OrmTransaction) GetElasticKeyMaterialKeyLatest(elasticKeyID googleUuid.UUID) (*MaterialKey, error) {
	if err := cryptoutilUtil.ValidateUUID(&elasticKeyID, "invalid Elastic Key ID"); err != nil {
		return nil, tx.toAppErr("failed to get latest Key by Elastic Key ID", err)
	}
	var key MaterialKey
	err := tx.state.gormTx.Order("material_key_id DESC").First(&key, "elastic_key_id=?", elasticKeyID).Error
	if err != nil {
		return nil, tx.toAppErr("failed to get latest Key by Elastic Key ID", err)
	}
	return &key, nil
}

func (tx *OrmTransaction) toAppErr(msg string, err error) error {
	tx.ormRepository.telemetryService.Slogger.Error(msg, "error", err)

	// custom errors
	if cryptoutilAppErr.IsAppErr(err) {
		return cryptoutilAppErr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", msg, err))
	}

	// gorm errors
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return cryptoutilAppErr.NewHTTP404NotFound(msg, fmt.Errorf("%s: %w", msg, err))
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return cryptoutilAppErr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", msg, err))
	case errors.Is(err, gorm.ErrForeignKeyViolated):
		return cryptoutilAppErr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", msg, err))
	case errors.Is(err, gorm.ErrCheckConstraintViolated):
		return cryptoutilAppErr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", msg, err))
	case errors.Is(err, gorm.ErrInvalidData):
		return cryptoutilAppErr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", msg, err))
	case errors.Is(err, gorm.ErrInvalidValueOfLength):
		return cryptoutilAppErr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", msg, err))
	case errors.Is(err, gorm.ErrNotImplemented):
		return cryptoutilAppErr.NewHTTP501StatusLineAndCodeNotImplemented(msg, fmt.Errorf("%s: %w", msg, err))
	}

	// SQLite errors
	var sqliteErr *sqlite.Error
	if errors.As(err, &sqliteErr) {
		switch sqliteErr.Code() {
		case 2067: // UNIQUE constraint failed
			return cryptoutilAppErr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", msg, err))
		case 787: // FOREIGN KEY constraint failed
			return cryptoutilAppErr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", msg, err))
		case 1299: // CHECK constraint failed
			return cryptoutilAppErr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", msg, err))
		}
	}

	// PostgreSQL errors
	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return cryptoutilAppErr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", msg, err))
		case "23503": // foreign_key_violation
			return cryptoutilAppErr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", msg, err))
		case "23514": // check_violation
			return cryptoutilAppErr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", msg, err))
		case "22001": // string_data_right_truncation
			return cryptoutilAppErr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", msg, err))
		}
	}

	return cryptoutilAppErr.NewHTTP500InternalServerError(msg, fmt.Errorf("%s: %w", msg, err))
}

func applyGetElasticKeysFilters(db *gorm.DB, filters *GetElasticKeysFilters) *gorm.DB {
	if filters == nil {
		return db
	}
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
	db = db.Offset(filters.PageNumber * filters.PageSize)
	db = db.Limit(filters.PageSize)
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
	db = db.Offset(filters.PageNumber * filters.PageSize)
	db = db.Limit(filters.PageSize)
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
	db = db.Offset(filters.PageNumber * filters.PageSize)
	db = db.Limit(filters.PageSize)
	return db
}
