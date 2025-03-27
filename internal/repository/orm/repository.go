package orm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"

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

type RepositoryOrm struct {
	gormDB              *gorm.DB
	sqlDB               *sql.DB
	shutdownDBContainer func()
}

func NewRepositoryOrm(ctx context.Context, dbType DBType, databaseUrl string, containerMode ContainerMode, applyMigrations bool) (*RepositoryOrm, error) {
	sqlDB, shutdownDBContainer, err := CreateSqlDB(ctx, dbType, databaseUrl, containerMode)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SQL DB: %w", err)
	}

	gormDB, err := CreateGormDB(dbType, sqlDB)
	if err != nil {
		shutdownDBContainer()
		return nil, fmt.Errorf("failed to connect with gormDB: %w", err)
	}

	if applyMigrations {
		log.Printf("Applying %s migrations", string(dbType))
		err = gormDB.AutoMigrate(ormTableStructs...)
		if err != nil {
			shutdownDBContainer()
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}
	} else {
		log.Printf("Skipping %s migrations", string(dbType))
	}

	return &RepositoryOrm{sqlDB: sqlDB, gormDB: gormDB, shutdownDBContainer: shutdownDBContainer}, nil
}

func (s *RepositoryOrm) Shutdown() {
	if err := s.sqlDB.Close(); err != nil {
		log.Printf("failed to close DB: %v", err)
	}
	if s.shutdownDBContainer != nil {
		s.shutdownDBContainer()
	}
}

func (s *RepositoryOrm) AddKeyPool(keyPool *KeyPool) error {
	if keyPool.KeyPoolID != uuidZero {
		return ErrKeyPoolIDMustBeZeroUUID
	}
	err := s.gormDB.Create(keyPool).Error
	if err != nil {
		return fmt.Errorf("failed to insert Key Pool: %w", err)
	}
	return nil
}

func (s *RepositoryOrm) AddKey(key *Key) error {
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

func (s *RepositoryOrm) FindKeyPools() ([]KeyPool, error) {
	var keyPools []KeyPool
	err := s.gormDB.Find(&keyPools).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find Key Pools: %w", err)
	}
	return keyPools, nil
}

func (s *RepositoryOrm) GetKeyPoolByID(keyPoolID uuid.UUID) (*KeyPool, error) {
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

func (s *RepositoryOrm) ListKeysByKeyPoolID(keyPoolID uuid.UUID) ([]Key, error) {
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

func (s *RepositoryOrm) ListMaxKeyIDByKeyPoolID(keyPoolID uuid.UUID) (int, error) {
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
