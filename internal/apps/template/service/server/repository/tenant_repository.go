// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
)

// Error summary messages.
const (
	errSummaryRecordNotFound        = "record not found"
	errSummaryDuplicateKeyViolation = "duplicate key violation"
	errSummaryDatabaseError         = "database error"
)

// TenantRepository provides CRUD operations for tenants.
type TenantRepository interface {
	Create(ctx context.Context, tenant *Tenant) error
	GetByID(ctx context.Context, id googleUuid.UUID) (*Tenant, error)
	GetByName(ctx context.Context, name string) (*Tenant, error)
	List(ctx context.Context, activeOnly bool) ([]*Tenant, error)
	Update(ctx context.Context, tenant *Tenant) error
	Delete(ctx context.Context, id googleUuid.UUID) error
	CountUsersAndClients(ctx context.Context, tenantID googleUuid.UUID) (users int64, clients int64, err error)
}

// TenantRepositoryImpl implements TenantRepository using GORM.
type TenantRepositoryImpl struct {
	db *gorm.DB
}

// NewTenantRepository creates a new TenantRepository.
func NewTenantRepository(db *gorm.DB) TenantRepository {
	return &TenantRepositoryImpl{db: db}
}

// Create creates a new tenant.
func (r *TenantRepositoryImpl) Create(ctx context.Context, tenant *Tenant) error {
	if err := r.db.WithContext(ctx).Create(tenant).Error; err != nil {
		return toAppErr(err)
	}

	return nil
}

// GetByID retrieves a tenant by ID.
func (r *TenantRepositoryImpl) GetByID(ctx context.Context, id googleUuid.UUID) (*Tenant, error) {
	var tenant Tenant

	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&tenant).Error; err != nil {
		return nil, toAppErr(err)
	}

	return &tenant, nil
}

// GetByName retrieves a tenant by name.
func (r *TenantRepositoryImpl) GetByName(ctx context.Context, name string) (*Tenant, error) {
	var tenant Tenant

	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&tenant).Error; err != nil {
		return nil, toAppErr(err)
	}

	return &tenant, nil
}

// List retrieves all tenants, optionally filtering by active status.
func (r *TenantRepositoryImpl) List(ctx context.Context, activeOnly bool) ([]*Tenant, error) {
	var tenants []*Tenant

	query := r.db.WithContext(ctx)

	if activeOnly {
		query = query.Where("active = ?", true)
	}

	if err := query.Order("created_at DESC").Find(&tenants).Error; err != nil {
		return nil, toAppErr(err)
	}

	return tenants, nil
}

// Update updates a tenant.
func (r *TenantRepositoryImpl) Update(ctx context.Context, tenant *Tenant) error {
	tenant.UpdatedAt = time.Now().UTC()

	if err := r.db.WithContext(ctx).Save(tenant).Error; err != nil {
		return toAppErr(err)
	}

	return nil
}

// Delete deletes a tenant by ID.
// Returns error if tenant still has users or clients.
func (r *TenantRepositoryImpl) Delete(ctx context.Context, id googleUuid.UUID) error {
	// Check if tenant has users or clients.
	users, clients, err := r.CountUsersAndClients(ctx, id)
	if err != nil {
		return err
	}

	if users > 0 || clients > 0 {
		summary := fmt.Sprintf("cannot delete tenant: has %d users and %d clients", users, clients)

		return cryptoutilSharedApperr.NewHTTP409Conflict(&summary, nil)
	}

	if err := r.db.WithContext(ctx).Delete(&Tenant{}, "id = ?", id).Error; err != nil {
		return toAppErr(err)
	}

	return nil
}

// CountUsersAndClients counts the number of users and clients for a tenant.
func (r *TenantRepositoryImpl) CountUsersAndClients(ctx context.Context, tenantID googleUuid.UUID) (users int64, clients int64, err error) {
	if err := r.db.WithContext(ctx).Model(&User{}).Where("tenant_id = ?", tenantID).Count(&users).Error; err != nil {
		return 0, 0, toAppErr(err)
	}

	if err := r.db.WithContext(ctx).Model(&Client{}).Where("tenant_id = ?", tenantID).Count(&clients).Error; err != nil {
		return 0, 0, toAppErr(err)
	}

	return users, clients, nil
}

// toAppErr maps GORM errors to application errors.
func toAppErr(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		summary := errSummaryRecordNotFound

		return cryptoutilSharedApperr.NewHTTP404NotFound(&summary, err)
	}

	if errors.Is(err, gorm.ErrDuplicatedKey) {
		summary := errSummaryDuplicateKeyViolation

		return cryptoutilSharedApperr.NewHTTP409Conflict(&summary, err)
	}

	// Check for SQLite UNIQUE constraint violations (error code 2067 or message contains "UNIQUE constraint")
	errMsg := err.Error()
	if strings.Contains(errMsg, "UNIQUE constraint") || strings.Contains(errMsg, "(2067)") {
		summary := errSummaryDuplicateKeyViolation

		return cryptoutilSharedApperr.NewHTTP409Conflict(&summary, err)
	}

	summary := errSummaryDatabaseError

	return cryptoutilSharedApperr.NewHTTP500InternalServerError(&summary, err)
}
