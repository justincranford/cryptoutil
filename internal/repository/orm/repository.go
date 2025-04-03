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
	_ "github.com/lib/pq"
	"modernc.org/sqlite"
	_ "modernc.org/sqlite"
)

var (
	ErrKeyPoolIDMustBeNonZeroUUID = errors.New("Key Pool ID must not be 00000000-0000-0000-0000-000000000000")
	ErrKeyIDMustBeNonZeroUUID     = errors.New("Key ID must not be 00000000-0000-0000-0000-000000000000")
)

type RepositoryProvider struct {
	telemetryService *cryptoutilTelemetry.Service
	sqlProvider      *cryptoutilSqlProvider.SqlProvider
	uuidV7Pool       *cryptoutilKeygen.KeyPool
	gormDB           *gorm.DB
	applyMigrations  bool
}

func NewRepositoryOrm(ctx context.Context, telemetryService *cryptoutilTelemetry.Service, sqlProvider *cryptoutilSqlProvider.SqlProvider, applyMigrations bool) (*RepositoryProvider, error) {
	uuidV7Pool := cryptoutilKeygen.NewKeyPool(ctx, telemetryService, "Orm UUIDv7", 3, 2, cryptoutilKeygen.MaxKeys, cryptoutilKeygen.MaxTime, cryptoutilKeygen.GenerateUUIDv7Function())

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

func (s *RepositoryTransaction) GetKeyPoolByKeyPoolID(keyPoolID uuid.UUID) (*KeyPool, error) {
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

func (s *RepositoryTransaction) UpdateKeyPoolByKeyPoolID(keyPool *KeyPool) error {
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
	err := s.state.gormTx.Model(&KeyPool{}).Where("key_pool_id = ?", keyPoolID).Update("key_pool_status", keyPoolStatus).Error
	if err != nil {
		return s.toAppErr("failed to update Key Pool Status", err)
	}
	return nil
}

func (s *RepositoryTransaction) GetKeyPools(ormKeyPoolsQueryParams *GetKeyPoolsFilters) ([]KeyPool, error) {
	var keyPools []KeyPool
	err := s.state.gormTx.Find(&keyPools).Error
	// order := fmt.Sprintf("%s %s", sortBy, sortOrder)
	// err := s.state.gormTx.Where(filter).Order(order).Offset(page * pageSize).Limit(pageSize).Find(&keyPools).Error
	if err != nil {
		return nil, s.toAppErr("failed to get Key Pools", err)
	}
	return keyPools, nil
}

func (s *RepositoryTransaction) AddKey(key *Key) error {
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

func (s *RepositoryTransaction) FindKeysByKeyPoolID(keyPoolID uuid.UUID, ormKeyPoolKeysQueryParams *GetKeyPoolKeysFilters) ([]Key, error) {
	if keyPoolID == cryptoutilUtil.ZeroUUID {
		return nil, s.toAppErr("failed to find Keys by Key Pool ID", ErrKeyPoolIDMustBeNonZeroUUID)
	}
	var keys []Key
	err := s.state.gormTx.Where("key_pool_id=?", keyPoolID).Find(&keys).Error
	if err != nil {
		return nil, s.toAppErr("failed to find Keys by Key Pool ID", err)
	}
	return keys, nil
}

func (s *RepositoryTransaction) GetKeys(ormKeysQueryParams *GetKeysFilters) ([]Key, error) {
	var keys []Key
	err := s.state.gormTx.Find(&keys).Error
	if err != nil {
		return nil, s.toAppErr("failed to find Keys", err)
	}
	return keys, nil
}

func (s *RepositoryTransaction) GetKeyByKeyPoolIDAndKeyID(keyPoolID uuid.UUID, keyID uuid.UUID) (*Key, error) {
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
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperr.NewHTTP404NotFound(msg, fmt.Errorf("%s: %w", msg, err))
	} else if errors.Is(err, gorm.ErrDuplicatedKey) {
		return apperr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", msg, err))
	} else if errors.Is(err, gorm.ErrForeignKeyViolated) {
		return apperr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", msg, err))
	} else if errors.Is(err, gorm.ErrCheckConstraintViolated) {
		return apperr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", msg, err))
	} else if errors.Is(err, gorm.ErrInvalidData) {
		return apperr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", msg, err))
	} else if errors.Is(err, gorm.ErrInvalidValueOfLength) {
		return apperr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", msg, err))
	} else if errors.Is(err, gorm.ErrNotImplemented) {
		return apperr.NewHTTP501StatusLineAndCodeNotImplemented(msg, fmt.Errorf("%s: %w", msg, err))
	}

	// SQLite errors
	var sqliteErr *sqlite.Error
	if errors.As(err, &sqliteErr) && sqliteErr.Code() == 2067 {
		return apperr.NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", msg, err))
	}

	return apperr.NewHTTP500InternalServerError(msg, fmt.Errorf("%s: %w", msg, err))
}
