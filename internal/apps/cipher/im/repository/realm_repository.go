// Copyright (c) 2025 Justin Cranford

package repository

import (
	"context"
	"fmt"
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

// RealmRepository provides CRUD operations for cipher-im authentication realms.
type RealmRepository interface {
	Create(ctx context.Context, realm *Realm) error
	GetByID(ctx context.Context, id googleUuid.UUID) (*Realm, error)
	GetByRealmID(ctx context.Context, realmID googleUuid.UUID) (*Realm, error)
	GetByName(ctx context.Context, name string) (*Realm, error)
	GetByType(ctx context.Context, realmType RealmType) ([]*Realm, error)
	ListAll(ctx context.Context, activeOnly bool) ([]*Realm, error)
	Update(ctx context.Context, realm *Realm) error
	Delete(ctx context.Context, id googleUuid.UUID) error
	GetActiveByPriority(ctx context.Context) ([]*Realm, error)
}

// RealmRepositoryImpl implements RealmRepository using GORM.
type RealmRepositoryImpl struct {
	db *gorm.DB
}

// NewRealmRepository creates a new RealmRepository.
func NewRealmRepository(db *gorm.DB) RealmRepository {
	return &RealmRepositoryImpl{db: db}
}

// Create creates a new realm.
func (r *RealmRepositoryImpl) Create(ctx context.Context, realm *Realm) error {
	if err := r.db.WithContext(ctx).Create(realm).Error; err != nil {
		return fmt.Errorf("failed to create realm: %w", err)
	}

	return nil
}

// GetByID retrieves a realm by primary key ID.
func (r *RealmRepositoryImpl) GetByID(ctx context.Context, id googleUuid.UUID) (*Realm, error) {
	var realm Realm

	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&realm).Error; err != nil {
		return nil, fmt.Errorf("failed to get realm by ID: %w", err)
	}

	return &realm, nil
}

// GetByRealmID retrieves a realm by realm_id (unique identifier).
func (r *RealmRepositoryImpl) GetByRealmID(ctx context.Context, realmID googleUuid.UUID) (*Realm, error) {
	var realm Realm

	if err := r.db.WithContext(ctx).Where("realm_id = ?", realmID).First(&realm).Error; err != nil {
		return nil, fmt.Errorf("failed to get realm by realm ID: %w", err)
	}

	return &realm, nil
}

// GetByName retrieves a realm by name.
func (r *RealmRepositoryImpl) GetByName(ctx context.Context, name string) (*Realm, error) {
	var realm Realm

	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&realm).Error; err != nil {
		return nil, fmt.Errorf("failed to get realm by name: %w", err)
	}

	return &realm, nil
}

// GetByType retrieves all realms of a specific type.
func (r *RealmRepositoryImpl) GetByType(ctx context.Context, realmType RealmType) ([]*Realm, error) {
	var realms []*Realm

	if err := r.db.WithContext(ctx).Where("type = ?", string(realmType)).Find(&realms).Error; err != nil {
		return nil, fmt.Errorf("failed to get realms by type: %w", err)
	}

	return realms, nil
}

// ListAll retrieves all realms.
func (r *RealmRepositoryImpl) ListAll(ctx context.Context, activeOnly bool) ([]*Realm, error) {
	var realms []*Realm

	query := r.db.WithContext(ctx)

	if activeOnly {
		query = query.Where("active = ?", true)
	}

	if err := query.Order("priority DESC, created_at DESC").Find(&realms).Error; err != nil {
		return nil, fmt.Errorf("failed to list realms: %w", err)
	}

	return realms, nil
}

// GetActiveByPriority retrieves active realms ordered by priority (highest first).
func (r *RealmRepositoryImpl) GetActiveByPriority(ctx context.Context) ([]*Realm, error) {
	var realms []*Realm

	if err := r.db.WithContext(ctx).
		Where("active = ?", true).
		Order("priority DESC, created_at ASC").
		Find(&realms).Error; err != nil {
		return nil, fmt.Errorf("failed to get active realms by priority: %w", err)
	}

	return realms, nil
}

// Update updates a realm.
func (r *RealmRepositoryImpl) Update(ctx context.Context, realm *Realm) error {
	realm.UpdatedAt = time.Now()

	if err := r.db.WithContext(ctx).Save(realm).Error; err != nil {
		return fmt.Errorf("failed to update realm: %w", err)
	}

	return nil
}

// Delete deletes a realm by ID.
func (r *RealmRepositoryImpl) Delete(ctx context.Context, id googleUuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&Realm{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete realm: %w", err)
	}

	return nil
}
