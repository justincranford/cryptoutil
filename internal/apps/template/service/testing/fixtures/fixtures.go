// Copyright (c) 2025 Justin Cranford
//

// Package fixtures provides test entity factories for cryptoutil service tests.
// Each factory creates a unique, persisted entity using UUIDv7 for deterministic test isolation.
package fixtures

import (
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

// CreateTestTenant creates and persists a unique Tenant entity for testing.
func CreateTestTenant(t testing.TB, db *gorm.DB) *cryptoutilAppsTemplateServiceServerRepository.Tenant {
	t.Helper()

	id := googleUuid.Must(googleUuid.NewV7())
	now := time.Now().UTC()
	tenant := &cryptoutilAppsTemplateServiceServerRepository.Tenant{
		ID:          id,
		Name:        "test-tenant-" + id.String(),
		Description: "Test tenant for automated testing.",
		Active:      1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	require.NoError(t, db.Create(tenant).Error)

	return tenant
}

// CreateTestRealm creates and persists a unique TenantRealm entity for testing.
func CreateTestRealm(t testing.TB, db *gorm.DB, tenantID googleUuid.UUID) *cryptoutilAppsTemplateServiceServerRepository.TenantRealm {
	t.Helper()

	now := time.Now().UTC()
	realm := &cryptoutilAppsTemplateServiceServerRepository.TenantRealm{
		ID:        googleUuid.Must(googleUuid.NewV7()),
		TenantID:  tenantID,
		RealmID:   googleUuid.Must(googleUuid.NewV7()),
		Type:      "username_password",
		Active:    true,
		Source:    "db",
		CreatedAt: now,
		UpdatedAt: now,
	}

	require.NoError(t, db.Create(realm).Error)

	return realm
}

// CreateTestUser creates and persists a unique User entity for testing.
func CreateTestUser(t testing.TB, db *gorm.DB, tenantID googleUuid.UUID) *cryptoutilAppsTemplateServiceServerRepository.User {
	t.Helper()

	id := googleUuid.Must(googleUuid.NewV7())
	now := time.Now().UTC()
	user := &cryptoutilAppsTemplateServiceServerRepository.User{
		ID:           id,
		TenantID:     tenantID,
		Username:     "testuser-" + id.String(),
		PasswordHash: "test-hash-placeholder-for-testing",
		Email:        "test-" + id.String() + "@example.com",
		Active:       1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	require.NoError(t, db.Create(user).Error)

	return user
}
