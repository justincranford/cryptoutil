package orm

import (
	"context"
	"errors"
	"fmt"

	"cryptoutil/internal/apperr"
	cryptoutilKeygen "cryptoutil/internal/crypto/keygen"
	cryptoutilSqlProvider "cryptoutil/internal/repository/sqlprovider"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"
	cryptoutilUtil "cryptoutil/internal/util"

	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"modernc.org/sqlite"
	_ "modernc.org/sqlite"
)

var (
	ormTableStructs               = []any{&KeyPool{}, &Key{}, &RootKey{}, &IntermediateKey{}, &ContentKey{}}
	ErrKeyPoolIDMustBeNonZeroUUID = fmt.Errorf("invalid Key Pool ID: %w", cryptoutilUtil.ErrNonZeroUUID)
	ErrKeyIDMustBeNonZeroUUID     = fmt.Errorf("invalid Key ID: %w", cryptoutilUtil.ErrNonZeroUUID)
)

type RepositoryProvider struct {
	telemetryService *cryptoutilTelemetry.Service
	sqlProvider      *cryptoutilSqlProvider.SqlProvider
	uuidV7Pool       *cryptoutilKeygen.KeyPool
	gormDB           *gorm.DB
	applyMigrations  bool
}

func NewRepositoryOrm(ctx context.Context, telemetryService *cryptoutilTelemetry.Service, sqlProvider *cryptoutilSqlProvider.SqlProvider, applyMigrations bool) (*RepositoryProvider, error) {
	uuidV7Pool, err := cryptoutilKeygen.NewKeyPool(ctx, telemetryService, "Orm UUIDv7", 2, 3, cryptoutilKeygen.MaxLifetimeKeys, cryptoutilKeygen.MaxLifetimeDuration, cryptoutilKeygen.GenerateUUIDv7Function())
	if err != nil {
		return nil, fmt.Errorf("failed to create UUID V7 pool: %w", err)
	}

	gormDB, err := cryptoutilSqlProvider.CreateGormDB(sqlProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to connect with gormDB: %w", err)
	}

	if applyMigrations {
		telemetryService.Slogger.Debug("applying migrations")
		err = gormDB.AutoMigrate(ormTableStructs...)
		if err != nil {
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}
	} else {
		telemetryService.Slogger.Debug("skipping migrations")
	}

	return &RepositoryProvider{telemetryService: telemetryService, sqlProvider: sqlProvider, uuidV7Pool: uuidV7Pool, gormDB: gormDB, applyMigrations: applyMigrations}, nil
}

func (s *RepositoryProvider) Shutdown() {
	s.telemetryService.Slogger.Debug("stopping ORM repository")
	s.sqlProvider.Shutdown()
	s.telemetryService.Slogger.Debug("stopped ORM repository")
}

// Service-Repository calls

func (s *RepositoryTransaction) AddKeyPool(keyPool *KeyPool) error {
	if keyPool.KeyPoolID == cryptoutilUtil.ZeroUUID {
		return s.toAppErr("failed to insert Key Pool", ErrKeyPoolIDMustBeNonZeroUUID)
	}
	err := s.state.gormTx.Create(keyPool).Error
	if err != nil {
		return s.toAppErr("failed to insert Key Pool", err)
	}
	return nil
}

func (s *RepositoryTransaction) GetKeyPool(keyPoolID uuid.UUID) (*KeyPool, error) {
	if keyPoolID == cryptoutilUtil.ZeroUUID {
		return nil, s.toAppErr("failed to find Key Pool by Key Pool ID", ErrKeyPoolIDMustBeNonZeroUUID)
	}
	var keyPool KeyPool
	err := s.state.gormTx.First(&keyPool, "key_pool_id=?", keyPoolID).Error
	if err != nil {
		return nil, s.toAppErr("failed to find Key Pool by Key Pool ID", err)
	}
	return &keyPool, nil
}

func (s *RepositoryTransaction) UpdateKeyPool(keyPool *KeyPool) error {
	if keyPool.KeyPoolID == cryptoutilUtil.ZeroUUID {
		return s.toAppErr("failed to update Key Pool", ErrKeyPoolIDMustBeNonZeroUUID)
	}
	err := s.state.gormTx.UpdateColumns(keyPool).Error
	if err != nil {
		return s.toAppErr("failed to update Key Pool", err)
	}
	return nil
}

func (s *RepositoryTransaction) UpdateKeyPoolStatus(keyPoolID uuid.UUID, keyPoolStatus KeyPoolStatus) error {
	if keyPoolID == cryptoutilUtil.ZeroUUID {
		return s.toAppErr("failed to update Key Pool Status", ErrKeyPoolIDMustBeNonZeroUUID)
	}
	err := s.state.gormTx.Model(&KeyPool{}).Where("key_pool_id=?", keyPoolID).Update("key_pool_status", keyPoolStatus).Error
	if err != nil {
		return s.toAppErr("failed to update Key Pool Status", err)
	}
	return nil
}

func (s *RepositoryTransaction) GetKeyPools(getKeyPoolsFilters *GetKeyPoolsFilters) ([]KeyPool, error) {
	var keyPools []KeyPool
	query := s.state.gormTx
	err := applyGetKeyPoolsFilters(query, getKeyPoolsFilters).Find(&keyPools).Error
	if err != nil {
		return nil, s.toAppErr("failed to get Key Pools", err)
	}
	return keyPools, nil
}

func (s *RepositoryTransaction) AddKeyPoolKey(key *Key) error {
	if key.KeyPoolID == cryptoutilUtil.ZeroUUID {
		return s.toAppErr("failed to insert Key", ErrKeyPoolIDMustBeNonZeroUUID)
	} else if key.KeyID == cryptoutilUtil.ZeroUUID {
		return s.toAppErr("failed to insert Key", ErrKeyIDMustBeNonZeroUUID)
	}
	err := s.state.gormTx.Create(key).Error
	if err != nil {
		return s.toAppErr("failed to insert Key", err)
	}
	return nil
}

func (s *RepositoryTransaction) GetKeyPoolKeys(keyPoolID uuid.UUID, getKeyPoolKeysFilters *GetKeyPoolKeysFilters) ([]Key, error) {
	if keyPoolID == cryptoutilUtil.ZeroUUID {
		return nil, s.toAppErr("failed to find Keys by Key Pool ID", ErrKeyPoolIDMustBeNonZeroUUID)
	}
	var keys []Key
	query := s.state.gormTx.Where("key_pool_id=?", keyPoolID)
	err := applyGetKeyPoolKeysFilters(query, getKeyPoolKeysFilters).Find(&keys).Error
	if err != nil {
		return nil, s.toAppErr("failed to find Keys by Key Pool ID", err)
	}
	return keys, nil
}

func (s *RepositoryTransaction) GetKeys(getKeysFilters *GetKeysFilters) ([]Key, error) {
	var keys []Key
	query := s.state.gormTx
	err := applyKeyFilters(query, getKeysFilters).Find(&keys).Error
	if err != nil {
		return nil, s.toAppErr("failed to find Keys", err)
	}
	return keys, nil
}

func (s *RepositoryTransaction) GetKeyPoolKey(keyPoolID uuid.UUID, keyID uuid.UUID) (*Key, error) {
	if keyPoolID == cryptoutilUtil.ZeroUUID {
		return nil, s.toAppErr("failed to find Key by Key Pool ID and Key ID", ErrKeyPoolIDMustBeNonZeroUUID)
	} else if keyID == cryptoutilUtil.ZeroUUID {
		return nil, s.toAppErr("failed to find Key by Key Pool ID and Key ID", ErrKeyIDMustBeNonZeroUUID)
	}
	var key Key
	err := s.state.gormTx.First(&key, "key_pool_id=? AND key_id=?", keyPoolID, keyID).Error
	if err != nil {
		return nil, s.toAppErr("failed to find Key by Key Pool ID and Key ID", err)
	}
	return &key, nil
}

func (s *RepositoryTransaction) toAppErr(msg string, err error) error {
	s.repositoryProvider.telemetryService.Slogger.Error(msg, "error", err)

	// custom errors
	if errors.Is(err, ErrKeyPoolIDMustBeNonZeroUUID) {
		return apperr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", msg, err))
	} else if errors.Is(err, ErrKeyPoolIDMustBeNonZeroUUID) {
		return apperr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", msg, err))
	}

	// gorm errors
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return apperr.NewHTTP404NotFound(msg, fmt.Errorf("%s: %w", msg, err))
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return apperr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", msg, err))
	case errors.Is(err, gorm.ErrForeignKeyViolated):
		return apperr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", msg, err))
	case errors.Is(err, gorm.ErrCheckConstraintViolated):
		return apperr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", msg, err))
	case errors.Is(err, gorm.ErrInvalidData):
		return apperr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", msg, err))
	case errors.Is(err, gorm.ErrInvalidValueOfLength):
		return apperr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", msg, err))
	case errors.Is(err, gorm.ErrNotImplemented):
		return apperr.NewHTTP501StatusLineAndCodeNotImplemented(msg, fmt.Errorf("%s: %w", msg, err))
	}

	// SQLite errors
	var sqliteErr *sqlite.Error
	if errors.As(err, &sqliteErr) {
		switch sqliteErr.Code() {
		case 2067: // UNIQUE constraint failed
			return apperr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", msg, err))
		case 787: // FOREIGN KEY constraint failed
			return apperr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", msg, err))
		case 1299: // CHECK constraint failed
			return apperr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", msg, err))
		}
	}

	// PostgreSQL errors
	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return apperr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", msg, err))
		case "23503": // foreign_key_violation
			return apperr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", msg, err))
		case "23514": // check_violation
			return apperr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", msg, err))
		case "22001": // string_data_right_truncation
			return apperr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", msg, err))
		}
	}

	return apperr.NewHTTP500InternalServerError(msg, fmt.Errorf("%s: %w", msg, err))
}

func applyGetKeyPoolsFilters(db *gorm.DB, filters *GetKeyPoolsFilters) *gorm.DB {
	if filters == nil {
		return db
	}
	if len(filters.ID) > 0 {
		db = db.Where("key_pool_id IN ?", filters.ID)
	}
	if len(filters.Name) > 0 {
		db = db.Where("key_pool_name IN ?", filters.Name)
	}
	if len(filters.Algorithm) > 0 {
		db = db.Where("key_pool_algorithm IN ?", filters.Algorithm)
	}
	if filters.VersioningAllowed != nil {
		db = db.Where("key_pool_versioning_allowed=?", *filters.VersioningAllowed)
	}
	if filters.ImportAllowed != nil {
		db = db.Where("key_pool_import_allowed=?", *filters.ImportAllowed)
	}
	if filters.ExportAllowed != nil {
		db = db.Where("key_pool_export_allowed=?", *filters.ExportAllowed)
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

func applyKeyFilters(db *gorm.DB, filters *GetKeysFilters) *gorm.DB {
	if filters == nil {
		return db
	}
	if len(filters.ID) > 0 {
		db = db.Where("key_id IN ?", filters.ID)
	}
	if len(filters.Pool) > 0 {
		db = db.Where("key_pool_id IN ?", filters.Pool)
	}
	if filters.MinimumGenerateDate != nil {
		db = db.Where("key_generate_date>=?", *filters.MinimumGenerateDate)
	}
	if filters.MaximumGenerateDate != nil {
		db = db.Where("key_generate_date<=?", *filters.MaximumGenerateDate)
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

func applyGetKeyPoolKeysFilters(db *gorm.DB, filters *GetKeyPoolKeysFilters) *gorm.DB {
	if filters == nil {
		return db
	}
	if len(filters.ID) > 0 {
		db = db.Where("key_id IN ?", filters.ID)
	}
	if filters.MinimumGenerateDate != nil {
		db = db.Where("key_generate_date>=?", *filters.MinimumGenerateDate)
	}
	if filters.MaximumGenerateDate != nil {
		db = db.Where("key_generate_date<=?", *filters.MaximumGenerateDate)
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
