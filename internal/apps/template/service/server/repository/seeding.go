// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"
	"fmt"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

// EnsureDefaultTenant creates default tenant and realm if they don't exist.
// Idempotent - safe to call multiple times without duplicating data.
//
// Parameters:
//   - ctx: Context for database operations
//   - db: GORM database connection
//   - tenantID: UUIDv7 for default tenant (service-specific magic constant)
//   - realmID: UUIDv7 for default realm (service-specific magic constant)
//
// Returns error if tenant/realm creation fails.
//
// Usage:
//
//	import cryptoutilMagic "cryptoutil/internal/shared/magic"
//
//	err := EnsureDefaultTenant(
//	    ctx,
//	    db,
//	    cryptoutilMagic.CipherIMDefaultTenantID,
//	    cryptoutilMagic.CipherIMDefaultRealmID,
//	)
func EnsureDefaultTenant(ctx context.Context, db *gorm.DB, tenantID, realmID googleUuid.UUID) error {
	// Check if default tenant already exists (idempotent).
	var existingTenant Tenant

	err := db.WithContext(ctx).Where("id = ?", tenantID).First(&existingTenant).Error
	if err == nil {
		// Tenant exists, verify realm exists too.
		var existingRealm TenantRealm

		err := db.WithContext(ctx).Where("id = ? AND tenant_id = ?", realmID, tenantID).First(&existingRealm).Error
		if err == nil {
			return nil // Both tenant and realm exist, nothing to do.
		}

		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("failed to query realm: %w", err)
		}

		// Tenant exists but realm missing - create realm only.
		realm := TenantRealm{
			ID:       realmID,
			TenantID: tenantID,
			RealmID:  realmID, // RealmID same as ID for default realm.
			Type:     "username_password",
			Config:   "{}",
			Active:   true,
			Source:   "db",
		}

		if err := db.WithContext(ctx).Create(&realm).Error; err != nil {
			return fmt.Errorf("failed to create default realm: %w", err)
		}

		return nil
	}

	if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to query tenant: %w", err)
	}

	// Tenant doesn't exist - create both tenant and realm in transaction.
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create default tenant.
		tenant := Tenant{
			ID:          tenantID,
			Name:        "default",
			Description: "Default tenant for single-tenant deployment",
		}
		tenant.SetActive(true)

		if err := tx.Create(&tenant).Error; err != nil {
			return fmt.Errorf("failed to create default tenant: %w", err)
		}

		// Create default realm for tenant.
		realm := TenantRealm{
			ID:       realmID,
			TenantID: tenantID,
			RealmID:  realmID, // RealmID same as ID for default realm.
			Type:     "username_password",
			Config:   "{}",
			Active:   true,
			Source:   "db",
		}

		if err := tx.Create(&realm).Error; err != nil {
			return fmt.Errorf("failed to create default realm: %w", err)
		}

		return nil
	})
}
