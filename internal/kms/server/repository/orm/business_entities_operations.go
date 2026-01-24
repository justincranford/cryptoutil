// Copyright (c) 2025 Justin Cranford

package orm

import (
	"errors"
	"fmt"

	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"

	"gorm.io/gorm"

	googleUuid "github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"modernc.org/sqlite"
)

// Database error code constants (SQLite numeric codes and Postgres SQLSTATE strings).
const (
	sqliteErrUniqueConstraint = cryptoutilSharedMagic.SQLiteErrUniqueConstraint
	sqliteErrForeignKey       = cryptoutilSharedMagic.SQLiteErrForeignKey
	sqliteErrCheckConstraint  = cryptoutilSharedMagic.SQLiteErrCheckConstraint

	pgCodeUniqueViolation      = cryptoutilSharedMagic.PGCodeUniqueViolation
	pgCodeForeignKeyViolation  = cryptoutilSharedMagic.PGCodeForeignKeyViolation
	pgCodeCheckViolation       = cryptoutilSharedMagic.PGCodeCheckViolation
	pgCodeStringDataTruncation = cryptoutilSharedMagic.PGCodeStringDataTruncation
)

// Service-Repository calls

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
func (tx *OrmTransaction) AddElasticKey(elasticKey *ElasticKey) error {
	if err := cryptoutilSharedUtilRandom.ValidateUUID(&elasticKey.ElasticKeyID, &ErrInvalidElasticKeyID); err != nil {
		return tx.toAppErr(&ErrFailedToAddElasticKey, err)
	}

	err := tx.state.gormTx.Create(elasticKey).Error
	if err != nil {
		return tx.toAppErr(&ErrFailedToAddElasticKey, err)
	}

	return nil
}

// GetElasticKey retrieves an elastic key by ID from the database.
func (tx *OrmTransaction) GetElasticKey(elasticKeyID *googleUuid.UUID) (*ElasticKey, error) {
	if err := cryptoutilSharedUtilRandom.ValidateUUID(elasticKeyID, &ErrInvalidElasticKeyID); err != nil {
		return nil, tx.toAppErr(&ErrFailedToGetElasticKeyByElasticKeyID, err)
	}

	var elasticKey ElasticKey

	err := tx.state.gormTx.First(&elasticKey, "elastic_key_id=?", elasticKeyID).Error
	if err != nil {
		return nil, tx.toAppErr(&ErrFailedToGetElasticKeyByElasticKeyID, err)
	}

	return &elasticKey, nil
}

// UpdateElasticKey updates an existing elastic key in the database.
func (tx *OrmTransaction) UpdateElasticKey(elasticKey *ElasticKey) error {
	if err := cryptoutilSharedUtilRandom.ValidateUUID(&elasticKey.ElasticKeyID, &ErrInvalidElasticKeyID); err != nil {
		return tx.toAppErr(&ErrFailedToUpdateElasticKeyByElasticKeyID, err)
	}

	err := tx.state.gormTx.UpdateColumns(elasticKey).Error
	if err != nil {
		return tx.toAppErr(&ErrFailedToUpdateElasticKeyByElasticKeyID, err)
	}

	return nil
}

// UpdateElasticKeyStatus updates the status of an elastic key in the database.
func (tx *OrmTransaction) UpdateElasticKeyStatus(elasticKeyID googleUuid.UUID, elasticKeyStatus cryptoutilOpenapiModel.ElasticKeyStatus) error {
	if err := cryptoutilSharedUtilRandom.ValidateUUID(&elasticKeyID, &ErrInvalidElasticKeyID); err != nil {
		return tx.toAppErr(&ErrFailedToUpdateElasticKeyStatusByElasticKeyID, err)
	}

	err := tx.state.gormTx.Model(&ElasticKey{}).Where("elastic_key_id=?", elasticKeyID).Update("elastic_key_status", elasticKeyStatus).Error
	if err != nil {
		return tx.toAppErr(&ErrFailedToUpdateElasticKeyStatusByElasticKeyID, err)
	}

	return nil
}

// GetElasticKeys retrieves elastic keys with optional filters from the database.
func (tx *OrmTransaction) GetElasticKeys(getElasticKeysFilters *GetElasticKeysFilters) ([]ElasticKey, error) {
	var elasticKeys []ElasticKey

	query := tx.state.gormTx

	err := applyGetElasticKeysFilters(query, getElasticKeysFilters).Find(&elasticKeys).Error
	if err != nil {
		return nil, tx.toAppErr(&ErrFailedToGetElasticKeys, err)
	}

	return elasticKeys, nil
}

// AddElasticKeyMaterialKey adds a new material key for an elastic key to the database.
func (tx *OrmTransaction) AddElasticKeyMaterialKey(key *MaterialKey) error {
	if err := cryptoutilSharedUtilRandom.ValidateUUID(&key.ElasticKeyID, &ErrInvalidElasticKeyID); err != nil {
		return tx.toAppErr(&ErrFailedToAddMaterialKey, err)
	} else if err := cryptoutilSharedUtilRandom.ValidateUUID(&key.MaterialKeyID, &ErrInvalidMaterialKeyID); err != nil {
		return tx.toAppErr(&ErrFailedToAddMaterialKey, err)
	}

	err := tx.state.gormTx.Create(key).Error
	if err != nil {
		return tx.toAppErr(&ErrFailedToAddMaterialKey, err)
	}

	return nil
}

// GetMaterialKeysForElasticKey retrieves material keys for an elastic key with optional filters.
func (tx *OrmTransaction) GetMaterialKeysForElasticKey(elasticKeyID *googleUuid.UUID, getElasticKeyKeysFilters *GetElasticKeyMaterialKeysFilters) ([]MaterialKey, error) {
	if err := cryptoutilSharedUtilRandom.ValidateUUID(elasticKeyID, &ErrFailedToGetMaterialKeysByElasticKeyID); err != nil {
		return nil, tx.toAppErr(&ErrFailedToGetMaterialKeysByElasticKeyID, err)
	}

	var keys []MaterialKey

	query := tx.state.gormTx.Where("elastic_key_id=?", elasticKeyID)

	err := applyGetElasticKeyKeysFilters(query, getElasticKeyKeysFilters).Find(&keys).Error
	if err != nil {
		return nil, tx.toAppErr(&ErrFailedToGetMaterialKeysByElasticKeyID, err)
	}

	return keys, nil
}

// GetMaterialKeys retrieves material keys based on the provided filter criteria.
func (tx *OrmTransaction) GetMaterialKeys(getKeysFilters *GetMaterialKeysFilters) ([]MaterialKey, error) {
	var keys []MaterialKey

	query := tx.state.gormTx

	err := applyKeyFilters(query, getKeysFilters).Find(&keys).Error
	if err != nil {
		return nil, tx.toAppErr(&ErrFailedToGetMaterialKeys, err)
	}

	return keys, nil
}

// GetElasticKeyMaterialKeyVersion retrieves a specific version of a material key by elastic key ID and material key ID.
func (tx *OrmTransaction) GetElasticKeyMaterialKeyVersion(elasticKeyID, materialKeyID *googleUuid.UUID) (*MaterialKey, error) {
	if err := cryptoutilSharedUtilRandom.ValidateUUID(elasticKeyID, &ErrInvalidElasticKeyID); err != nil {
		return nil, tx.toAppErr(&ErrFailedToGetMaterialKeyByElasticKeyIDAndMaterialKeyID, err)
	} else if err := cryptoutilSharedUtilRandom.ValidateUUID(materialKeyID, &ErrInvalidMaterialKeyID); err != nil {
		return nil, tx.toAppErr(&ErrFailedToGetMaterialKeyByElasticKeyIDAndMaterialKeyID, err)
	}

	var key MaterialKey

	err := tx.state.gormTx.First(&key, "elastic_key_id=? AND material_key_id=?", elasticKeyID, materialKeyID).Error
	if err != nil {
		return nil, tx.toAppErr(&ErrFailedToGetMaterialKeyByElasticKeyIDAndMaterialKeyID, err)
	}

	return &key, nil
}

// GetElasticKeyMaterialKeyLatest retrieves the latest material key for the given elastic key ID.
func (tx *OrmTransaction) GetElasticKeyMaterialKeyLatest(elasticKeyID googleUuid.UUID) (*MaterialKey, error) {
	if err := cryptoutilSharedUtilRandom.ValidateUUID(&elasticKeyID, &ErrInvalidElasticKeyID); err != nil {
		return nil, tx.toAppErr(&ErrFailedToGetLatestMaterialKeyByElasticKeyID, err)
	}

	var key MaterialKey

	err := tx.state.gormTx.Order("material_key_id DESC").First(&key, "elastic_key_id=?", elasticKeyID).Error
	if err != nil {
		return nil, tx.toAppErr(&ErrFailedToGetLatestMaterialKeyByElasticKeyID, err)
	}

	return &key, nil
}

// UpdateElasticKeyMaterialKeyRevoke updates the revocation date for a material key.
func (tx *OrmTransaction) UpdateElasticKeyMaterialKeyRevoke(materialKey *MaterialKey) error {
	if err := cryptoutilSharedUtilRandom.ValidateUUID(&materialKey.ElasticKeyID, &ErrInvalidElasticKeyID); err != nil {
		return tx.toAppErr(&ErrFailedToUpdateMaterialKey, err)
	} else if err := cryptoutilSharedUtilRandom.ValidateUUID(&materialKey.MaterialKeyID, &ErrInvalidMaterialKeyID); err != nil {
		return tx.toAppErr(&ErrFailedToUpdateMaterialKey, err)
	}

	err := tx.state.gormTx.Model(&MaterialKey{}).
		Where("elastic_key_id=? AND material_key_id=?", materialKey.ElasticKeyID, materialKey.MaterialKeyID).
		Update("material_key_revocation_date", materialKey.MaterialKeyRevocationDate).Error
	if err != nil {
		return tx.toAppErr(&ErrFailedToUpdateMaterialKey, err)
	}

	return nil
}

func (tx *OrmTransaction) toAppErr(msg *string, err error) error {
	tx.ormRepository.telemetryService.Slogger.Error(*msg, "error", err)

	// custom errors
	if cryptoutilSharedApperr.IsAppErr(err) {
		return cryptoutilSharedApperr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", *msg, err))
	}

	// gorm errors
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return cryptoutilSharedApperr.NewHTTP404NotFound(msg, fmt.Errorf("%s: %w", *msg, err))
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return cryptoutilSharedApperr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", *msg, err))
	case errors.Is(err, gorm.ErrForeignKeyViolated):
		return cryptoutilSharedApperr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", *msg, err))
	case errors.Is(err, gorm.ErrCheckConstraintViolated):
		return cryptoutilSharedApperr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", *msg, err))
	case errors.Is(err, gorm.ErrInvalidData):
		return cryptoutilSharedApperr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", *msg, err))
	case errors.Is(err, gorm.ErrInvalidValueOfLength):
		return cryptoutilSharedApperr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", *msg, err))
	case errors.Is(err, gorm.ErrNotImplemented):
		return cryptoutilSharedApperr.NewHTTP501StatusLineAndCodeNotImplemented(msg, fmt.Errorf("%s: %w", *msg, err))
	}

	// SQLite errors
	var sqliteErr *sqlite.Error
	if errors.As(err, &sqliteErr) {
		switch sqliteErr.Code() {
		case sqliteErrUniqueConstraint: // UNIQUE constraint failed
			return cryptoutilSharedApperr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", *msg, err))
		case sqliteErrForeignKey: // FOREIGN KEY constraint failed
			return cryptoutilSharedApperr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", *msg, err))
		case sqliteErrCheckConstraint: // CHECK constraint failed
			return cryptoutilSharedApperr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", *msg, err))
		}
	}

	// PostgreSQL errors
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgCodeUniqueViolation: // unique_violation
			return cryptoutilSharedApperr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", *msg, err))
		case pgCodeForeignKeyViolation: // foreign_key_violation
			return cryptoutilSharedApperr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", *msg, err))
		case pgCodeCheckViolation: // check_violation
			return cryptoutilSharedApperr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", *msg, err))
		case pgCodeStringDataTruncation: // string_data_right_truncation
			return cryptoutilSharedApperr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", *msg, err))
		}
	}

	return cryptoutilSharedApperr.NewHTTP500InternalServerError(msg, fmt.Errorf("%s: %w", *msg, err))
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

	// Only apply pagination if PageSize is set (> 0).
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

	// Only apply pagination if PageSize is set (> 0).
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

	// Only apply pagination if PageSize is set (> 0).
	if filters.PageSize > 0 {
		db = db.Offset(filters.PageNumber * filters.PageSize)
		db = db.Limit(filters.PageSize)
	}

	return db
}
