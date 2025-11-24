package orm

import (
	"context"
	"fmt"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	identityDomain "cryptoutil/internal/identity/domain"
)

// KeyRepositoryGORM implements KeyRepository using GORM ORM.
type KeyRepositoryGORM struct {
	db *gorm.DB
}

// NewKeyRepository creates a new GORM-based key repository.
func NewKeyRepository(db *gorm.DB) *KeyRepositoryGORM {
	return &KeyRepositoryGORM{db: db}
}

// Create inserts a new cryptographic key into the database.
func (r *KeyRepositoryGORM) Create(ctx context.Context, key *identityDomain.Key) error {
	if key == nil {
		return fmt.Errorf("key cannot be nil")
	}

	return getDB(ctx, r.db).WithContext(ctx).Create(key).Error
}

// FindByID retrieves a key by its unique identifier.
func (r *KeyRepositoryGORM) FindByID(ctx context.Context, id googleUuid.UUID) (*identityDomain.Key, error) {
	var key identityDomain.Key
	err := getDB(ctx, r.db).WithContext(ctx).Where("id = ?", id).First(&key).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

// FindByUsage retrieves all keys matching the specified usage and active status.
func (r *KeyRepositoryGORM) FindByUsage(ctx context.Context, usage string, active bool) ([]*identityDomain.Key, error) {
	var keys []*identityDomain.Key
	query := getDB(ctx, r.db).WithContext(ctx).Where("usage = ?", usage)

	if active {
		query = query.Where("active = ?", true)
	}

	err := query.Order("created_at DESC").Find(&keys).Error
	if err != nil {
		return nil, err
	}

	return keys, nil
}

// Update modifies an existing key in the database.
func (r *KeyRepositoryGORM) Update(ctx context.Context, key *identityDomain.Key) error {
	if key == nil {
		return fmt.Errorf("key cannot be nil")
	}

	return getDB(ctx, r.db).WithContext(ctx).Save(key).Error
}

// Delete removes a key from the database (soft delete).
func (r *KeyRepositoryGORM) Delete(ctx context.Context, id googleUuid.UUID) error {
	return getDB(ctx, r.db).WithContext(ctx).Delete(&identityDomain.Key{}, "id = ?", id).Error
}

// List retrieves all keys with optional pagination.
func (r *KeyRepositoryGORM) List(ctx context.Context, limit, offset int) ([]*identityDomain.Key, error) {
	var keys []*identityDomain.Key
	query := getDB(ctx, r.db).WithContext(ctx).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&keys).Error
	if err != nil {
		return nil, err
	}

	return keys, nil
}

// Count returns the total number of keys in the database.
func (r *KeyRepositoryGORM) Count(ctx context.Context) (int64, error) {
	var count int64
	err := getDB(ctx, r.db).WithContext(ctx).Model(&identityDomain.Key{}).Count(&count).Error
	return count, err
}
