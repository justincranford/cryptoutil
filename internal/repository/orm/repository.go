package orm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"

	cryptoutilSqlProvider "cryptoutil/internal/repository/sqlprovider"

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

type Repository struct {
	gormDB *gorm.DB
	sqlDB  *sql.DB
}

func NewRepositoryOrm(ctx context.Context, dbType cryptoutilSqlProvider.SupportedSqlDB, sqlDB *sql.DB, applyMigrations bool) (*Repository, error) {
	gormDB, err := cryptoutilSqlProvider.CreateGormDB(dbType, sqlDB)
	if err != nil {
		return nil, fmt.Errorf("failed to connect with gormDB: %w", err)
	}

	if applyMigrations {
		log.Printf("Applying migrations")
		err = gormDB.AutoMigrate(ormTableStructs...)
		if err != nil {
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}
	} else {
		log.Printf("Skipping migrations")
	}

	return &Repository{sqlDB: sqlDB, gormDB: gormDB}, nil
}

func (s *Repository) Shutdown() {
	if err := s.sqlDB.Close(); err != nil {
		log.Printf("failed to close DB: %v", err)
	}
}

func (s *Repository) AddKeyPool(keyPool *KeyPool) error {
	if keyPool.KeyPoolID != uuidZero {
		return ErrKeyPoolIDMustBeZeroUUID
	}
	err := s.gormDB.Create(keyPool).Error
	if err != nil {
		return fmt.Errorf("failed to insert Key Pool: %w", err)
	}
	return nil
}

func (s *Repository) UpdateKeyPool(keyPool *KeyPool) error {
	if keyPool.KeyPoolID == uuidZero {
		return ErrKeyPoolIDMustBeNonZeroUUID
	}
	err := s.gormDB.UpdateColumns(keyPool).Error
	if err != nil {
		return fmt.Errorf("failed to update Key Pool: %w", err)
	}
	return nil
}

func (s *Repository) UpdateKeyPoolStatus(keyPoolID uuid.UUID, keyPoolStatus KeyPoolStatusEnum) error {
	if keyPoolID == uuidZero {
		return ErrKeyPoolIDMustBeNonZeroUUID
	}
	err := s.gormDB.Model(&KeyPool{}).Where("key_pool_id = ?", keyPoolID).Update("key_pool_status", keyPoolStatus).Error
	if err != nil {
		return fmt.Errorf("failed to update Key Pool Status: %w", err)
	}
	return nil
}

func (s *Repository) AddKey(key *Key) error {
	if key.KeyPoolID == uuidZero {
		return ErrKeyPoolIDMustBeNonZeroUUID
	} else if key.KeyID == 0 {
		return fmt.Errorf("Key ID must be non-zero, but got 0")
	}
	err := s.gormDB.Create(key).Error
	if err != nil {
		return fmt.Errorf("failed to insert Key: %w", err)
	}
	return nil
}

func (s *Repository) FindKeyPools() ([]KeyPool, error) {
	var keyPools []KeyPool
	err := s.gormDB.Find(&keyPools).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find Key Pools: %w", err)
	}
	return keyPools, nil
}

func (s *Repository) GetKeyPoolByID(keyPoolID uuid.UUID) (*KeyPool, error) {
	if keyPoolID == uuidZero {
		return nil, ErrKeyPoolIDMustBeNonZeroUUID
	}
	var keyPool KeyPool
	err := s.gormDB.First(&keyPool, "key_pool_id=?", keyPoolID).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find Key Pool by Key Pool ID: %w", err)
	}
	return &keyPool, nil
}

func (s *Repository) ListKeysByKeyPoolID(keyPoolID uuid.UUID) ([]Key, error) {
	if keyPoolID == uuidZero {
		return nil, ErrKeyPoolIDMustBeNonZeroUUID
	}
	var keys []Key
	err := s.gormDB.Where("key_pool_id=?", keyPoolID).Find(&keys).Error
	if err != nil {
		return keys, fmt.Errorf("failed to find Keys by Key Pool ID: %w", err)
	}
	return keys, nil
}

func (s *Repository) ListMaxKeyIDByKeyPoolID(keyPoolID uuid.UUID) (int, error) {
	if keyPoolID == uuidZero {
		return math.MinInt, ErrKeyPoolIDMustBeNonZeroUUID
	}
	var maxKeyID int
	err := s.gormDB.Model(&Key{}).Where("key_pool_id=?", keyPoolID).Select("COALESCE(MAX(key_id), 0)").Scan(&maxKeyID).Error
	if err != nil {
		return math.MinInt, fmt.Errorf("failed to get max Key ID by Key Pool ID: %w", err)
	}
	return maxKeyID, nil
}
