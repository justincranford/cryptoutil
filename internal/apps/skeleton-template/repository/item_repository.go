// Copyright (c) 2025 Justin Cranford
//
// TEMPLATE: Copy and rename 'skeleton' -> your-service-name before use.

// Package repository provides data access for the skeleton-template service.
package repository

import (
	"context"
	"fmt"

	cryptoutilAppsSkeletonTemplateDomain "cryptoutil/internal/apps/skeleton-template/domain"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

// ItemRepository handles CRUD operations for TemplateItem.
type ItemRepository struct {
	db *gorm.DB
}

// NewItemRepository creates a new ItemRepository.
func NewItemRepository(db *gorm.DB) *ItemRepository {
	return &ItemRepository{db: db}
}

// Create inserts a new TemplateItem into the database.
func (r *ItemRepository) Create(ctx context.Context, item *cryptoutilAppsSkeletonTemplateDomain.TemplateItem) error {
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		return fmt.Errorf("failed to create item: %w", err)
	}

	return nil
}

// GetByID retrieves a TemplateItem by tenant and item ID.
func (r *ItemRepository) GetByID(ctx context.Context, tenantID, itemID googleUuid.UUID) (*cryptoutilAppsSkeletonTemplateDomain.TemplateItem, error) {
	var item cryptoutilAppsSkeletonTemplateDomain.TemplateItem

	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID.String(), itemID.String()).First(&item).Error; err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	return &item, nil
}

// List retrieves TemplateItems for a tenant with pagination.
func (r *ItemRepository) List(ctx context.Context, tenantID googleUuid.UUID, page, size int) ([]cryptoutilAppsSkeletonTemplateDomain.TemplateItem, int64, error) {
	var items []cryptoutilAppsSkeletonTemplateDomain.TemplateItem

	var total int64

	query := r.db.WithContext(ctx).Model(&cryptoutilAppsSkeletonTemplateDomain.TemplateItem{}).Where("tenant_id = ?", tenantID.String())

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count items: %w", err)
	}

	offset := (page - 1) * size

	if err := query.Offset(offset).Limit(size).Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list items: %w", err)
	}

	return items, total, nil
}

// Update modifies an existing TemplateItem.
func (r *ItemRepository) Update(ctx context.Context, item *cryptoutilAppsSkeletonTemplateDomain.TemplateItem) error {
	result := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", item.TenantID.String(), item.ID.String()).Save(item)
	if result.Error != nil {
		return fmt.Errorf("failed to update item: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// Delete removes a TemplateItem by tenant and item ID.
func (r *ItemRepository) Delete(ctx context.Context, tenantID, itemID googleUuid.UUID) error {
	result := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID.String(), itemID.String()).Delete(&cryptoutilAppsSkeletonTemplateDomain.TemplateItem{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete item: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
