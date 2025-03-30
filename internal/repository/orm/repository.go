package orm

import (
	"context"
	"errors"
	"fmt"
	"math"

	cryptoutilSqlProvider "cryptoutil/internal/repository/sqlprovider"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"

	"gorm.io/gorm"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

var (
	uuidZero                      = uuid.UUID{}
	ErrKeyPoolIDMustBeZeroUUID    = errors.New("Key Pool ID must be 00000000-0000-0000-0000-000000000000 for add Key Pool")
	ErrKeyPoolIDMustBeNonZeroUUID = errors.New("Key Pool ID must not be 00000000-0000-0000-0000-000000000000 existing Key Pool")
)

type RepositoryProvider struct {
	telemetryService *cryptoutilTelemetry.Service
	sqlProvider      *cryptoutilSqlProvider.SqlProvider
	gormDB           *gorm.DB
	applyMigrations  bool
}

func NewRepositoryOrm(ctx context.Context, telemetryService *cryptoutilTelemetry.Service, sqlProvider *cryptoutilSqlProvider.SqlProvider, applyMigrations bool) (*RepositoryProvider, error) {
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

	return &RepositoryProvider{telemetryService: telemetryService, sqlProvider: sqlProvider, gormDB: gormDB, applyMigrations: applyMigrations}, nil
}

func (s *RepositoryProvider) Shutdown() {
	s.telemetryService.Slogger.Debug("stopping ORM repository")
	s.sqlProvider.Shutdown()
	s.telemetryService.Slogger.Debug("stopped ORM repository")
}

func (s *RepositoryTransaction) AddKeyPool(keyPool *KeyPool) error {
	if keyPool.KeyPoolID != uuidZero {
		return ErrKeyPoolIDMustBeZeroUUID
	}
	err := s.state.gormTx.Create(keyPool).Error
	if err != nil {
		s.repositoryProvider.telemetryService.Slogger.Error("failed to insert Key Pool", "error", err)
		return fmt.Errorf("failed to insert Key Pool: %w", err)
	}
	return nil
}

func (s *RepositoryTransaction) UpdateKeyPool(keyPool *KeyPool) error {
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

func (s *RepositoryTransaction) AddKey(key *Key) error {
	if key.KeyPoolID == uuidZero {
		return ErrKeyPoolIDMustBeNonZeroUUID
	} else if key.KeyID == 0 {
		s.repositoryProvider.telemetryService.Slogger.Error("Key ID must be non-zero, but got 0")
		return fmt.Errorf("Key ID must be non-zero, but got 0")
	}
	err := s.state.gormTx.Create(key).Error
	if err != nil {
		s.repositoryProvider.telemetryService.Slogger.Error("failed to insert Key", "error", err)
		return fmt.Errorf("failed to insert Key: %w", err)
	}
	return nil
}

func (s *RepositoryTransaction) FindKeyPools() ([]KeyPool, error) {
	var keyPools []KeyPool
	err := s.state.gormTx.Find(&keyPools).Error
	if err != nil {
		s.repositoryProvider.telemetryService.Slogger.Error("failed to find Key Pools", "error", err)
		return nil, fmt.Errorf("failed to find Key Pools: %w", err)
	}
	return keyPools, nil
}

func (s *RepositoryTransaction) GetKeyPoolByID(keyPoolID uuid.UUID) (*KeyPool, error) {
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

func (s *RepositoryTransaction) ListKeysByKeyPoolID(keyPoolID uuid.UUID) ([]Key, error) {
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

func (s *RepositoryTransaction) ListMaxKeyIDByKeyPoolID(keyPoolID uuid.UUID) (int, error) {
	if keyPoolID == uuidZero {
		return math.MinInt, ErrKeyPoolIDMustBeNonZeroUUID
	}
	var maxKeyID int
	err := s.state.gormTx.Model(&Key{}).Where("key_pool_id=?", keyPoolID).Select("COALESCE(MAX(key_id), 0)").Scan(&maxKeyID).Error
	if err != nil {
		s.repositoryProvider.telemetryService.Slogger.Error("failed to get max Key ID by Key Pool ID", "error", err)
		return math.MinInt, fmt.Errorf("failed to get max Key ID by Key Pool ID: %w", err)
	}
	return maxKeyID, nil
}
