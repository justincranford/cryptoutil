package orm

import (
	"context"
	"errors"
	"fmt"

	cryptoutilKeygen "cryptoutil/internal/crypto/keygen"
	cryptoutilSqlProvider "cryptoutil/internal/repository/sqlprovider"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"

	"gorm.io/gorm"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

var (
	uuidZero                      = uuid.UUID{}
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

func (s *RepositoryTransaction) AddKeyPool(keyPool *KeyPool) error {
	if keyPool.KeyPoolID == uuidZero {
		return ErrKeyPoolIDMustBeNonZeroUUID
	}
	err := s.state.gormTx.Create(keyPool).Error
	if err != nil {
		s.repositoryProvider.telemetryService.Slogger.Error("failed to insert Key Pool", "error", err)
		return fmt.Errorf("failed to insert Key Pool: %w", err)
	}
	return nil
}

func (s *RepositoryTransaction) GetKeyPoolByKeyPoolID(keyPoolID uuid.UUID) (*KeyPool, error) {
	if keyPoolID == uuidZero {
		return nil, ErrKeyPoolIDMustBeNonZeroUUID
	}
	var keyPool KeyPool
	err := s.state.gormTx.First(&keyPool, "key_pool_id=?", keyPoolID).Error
	if err != nil {
		s.repositoryProvider.telemetryService.Slogger.Error("failed to find Key Pool by Key Pool ID", "error", err)
		return nil, fmt.Errorf("failed to find Key Pool by Key Pool ID: %w", err)
	}
	return &keyPool, nil
}

func (s *RepositoryTransaction) UpdateKeyPoolByKeyPoolID(keyPool *KeyPool) error {
	if keyPool.KeyPoolID == uuidZero {
		return ErrKeyPoolIDMustBeNonZeroUUID
	}
	err := s.state.gormTx.UpdateColumns(keyPool).Error
	if err != nil {
		s.repositoryProvider.telemetryService.Slogger.Error("failed to update Key Pool", "error", err)
		return fmt.Errorf("failed to update Key Pool: %w", err)
	}
	return nil
}

func (s *RepositoryTransaction) UpdateKeyPoolStatus(keyPoolID uuid.UUID, keyPoolStatus KeyPoolStatus) error {
	if keyPoolID == uuidZero {
		return ErrKeyPoolIDMustBeNonZeroUUID
	}
	err := s.state.gormTx.Model(&KeyPool{}).Where("key_pool_id = ?", keyPoolID).Update("key_pool_status", keyPoolStatus).Error
	if err != nil {
		s.repositoryProvider.telemetryService.Slogger.Error("failed to update Key Pool Status", "error", err)
		return fmt.Errorf("failed to update Key Pool Status: %w", err)
	}
	return nil
}

func (s *RepositoryTransaction) GetKeyPools() ([]KeyPool, error) {
	var keyPools []KeyPool
	err := s.state.gormTx.Find(&keyPools).Error
	// order := fmt.Sprintf("%s %s", sortBy, sortOrder)
	// err := s.state.gormTx.Where(filter).Order(order).Offset(page * pageSize).Limit(pageSize).Find(&keyPools).Error
	if err != nil {
		s.repositoryProvider.telemetryService.Slogger.Error("failed to get Key Pools", "error", err)
		return nil, fmt.Errorf("failed to get Key Pools: %w", err)
	}
	return keyPools, nil
}

func (s *RepositoryTransaction) AddKey(key *Key) error {
	if key.KeyPoolID == uuidZero {
		return ErrKeyPoolIDMustBeNonZeroUUID
	} else if key.KeyID == uuidZero {
		return ErrKeyIDMustBeNonZeroUUID
	}
	err := s.state.gormTx.Create(key).Error
	if err != nil {
		s.repositoryProvider.telemetryService.Slogger.Error("failed to insert Key", "error", err)
		return fmt.Errorf("failed to insert Key: %w", err)
	}
	return nil
}

func (s *RepositoryTransaction) FindKeysByKeyPoolID(keyPoolID uuid.UUID) ([]Key, error) {
	if keyPoolID == uuidZero {
		return nil, ErrKeyPoolIDMustBeNonZeroUUID
	}
	var keys []Key
	err := s.state.gormTx.Where("key_pool_id=?", keyPoolID).Find(&keys).Error
	if err != nil {
		s.repositoryProvider.telemetryService.Slogger.Error("failed to find Keys by Key Pool ID", "error", err)
		return keys, fmt.Errorf("failed to find Keys by Key Pool ID: %w", err)
	}
	return keys, nil
}

func (s *RepositoryTransaction) GetKeys() ([]Key, error) {
	var keys []Key
	err := s.state.gormTx.Find(&keys).Error
	if err != nil {
		s.repositoryProvider.telemetryService.Slogger.Error("failed to find Keys", "error", err)
		return keys, fmt.Errorf("failed to find Keys: %w", err)
	}
	return keys, nil
}

func (s *RepositoryTransaction) GetKeyByKeyPoolIDAndKeyID(keyPoolID uuid.UUID, keyID uuid.UUID) (*Key, error) {
	if keyPoolID == uuidZero {
		return nil, ErrKeyPoolIDMustBeNonZeroUUID
	} else if keyID == uuidZero {
		return nil, ErrKeyIDMustBeNonZeroUUID
	}
	var key Key
	err := s.state.gormTx.First(&key, "key_pool_id=? AND key_id=?", keyPoolID, keyID).Error
	if err != nil {
		s.repositoryProvider.telemetryService.Slogger.Error("failed to find Key by Key Pool ID and Key ID", "error", err)
		return nil, fmt.Errorf("failed to find Key by Key Pool ID and Key ID: %w", err)
	}
	return &key, nil
}
