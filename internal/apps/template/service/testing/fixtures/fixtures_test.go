// Copyright (c) 2025 Justin Cranford
//

package fixtures_test

import (
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilTestingFixtures "cryptoutil/internal/apps/template/service/testing/fixtures"
	cryptoutilTestingTestdb "cryptoutil/internal/apps/template/service/testing/testdb"
)

func TestCreateTestTenant(t *testing.T) {
	t.Parallel()

	db := cryptoutilTestingTestdb.RequireNewInMemorySQLiteDB(t, &cryptoutilAppsTemplateServiceServerRepository.Tenant{})

	tenant := cryptoutilTestingFixtures.CreateTestTenant(t, db)

	require.NotNil(t, tenant)
	assert.NotEqual(t, googleUuid.Nil, tenant.ID, "ID should not be zero UUID")
	assert.NotEmpty(t, tenant.Name, "Name should not be empty")
	assert.Equal(t, 1, tenant.Active, "tenant should be active")

	var found cryptoutilAppsTemplateServiceServerRepository.Tenant
	require.NoError(t, db.First(&found, "id = ?", tenant.ID).Error)
	assert.Equal(t, tenant.Name, found.Name)
}

func TestCreateTestTenant_UniquePerCall(t *testing.T) {
	t.Parallel()

	db := cryptoutilTestingTestdb.RequireNewInMemorySQLiteDB(t, &cryptoutilAppsTemplateServiceServerRepository.Tenant{})

	tenant1 := cryptoutilTestingFixtures.CreateTestTenant(t, db)
	tenant2 := cryptoutilTestingFixtures.CreateTestTenant(t, db)

	assert.NotEqual(t, tenant1.ID, tenant2.ID, "each call should produce a unique ID")
	assert.NotEqual(t, tenant1.Name, tenant2.Name, "each call should produce a unique name")
}

func TestCreateTestRealm(t *testing.T) {
	t.Parallel()

	db := cryptoutilTestingTestdb.RequireNewInMemorySQLiteDB(t, &cryptoutilAppsTemplateServiceServerRepository.Tenant{}, &cryptoutilAppsTemplateServiceServerRepository.TenantRealm{})

	tenant := cryptoutilTestingFixtures.CreateTestTenant(t, db)
	realm := cryptoutilTestingFixtures.CreateTestRealm(t, db, tenant.ID)

	require.NotNil(t, realm)
	assert.NotEqual(t, googleUuid.Nil, realm.ID, "realm ID should not be zero UUID")
	assert.Equal(t, tenant.ID, realm.TenantID)
	assert.NotEqual(t, googleUuid.Nil, realm.RealmID, "realm RealmID should not be zero UUID")
	assert.True(t, realm.Active, "realm should be active")

	var found cryptoutilAppsTemplateServiceServerRepository.TenantRealm
	require.NoError(t, db.First(&found, "id = ?", realm.ID).Error)
	assert.Equal(t, realm.TenantID, found.TenantID)
}

func TestCreateTestUser(t *testing.T) {
	t.Parallel()

	db := cryptoutilTestingTestdb.RequireNewInMemorySQLiteDB(t, &cryptoutilAppsTemplateServiceServerRepository.Tenant{}, &cryptoutilAppsTemplateServiceServerRepository.User{})

	tenant := cryptoutilTestingFixtures.CreateTestTenant(t, db)
	user := cryptoutilTestingFixtures.CreateTestUser(t, db, tenant.ID)

	require.NotNil(t, user)
	assert.NotEqual(t, googleUuid.Nil, user.ID, "user ID should not be zero UUID")
	assert.Equal(t, tenant.ID, user.TenantID)
	assert.NotEmpty(t, user.Username, "username should not be empty")
	assert.Equal(t, 1, user.Active, "user should be active")

	var found cryptoutilAppsTemplateServiceServerRepository.User
	require.NoError(t, db.First(&found, "id = ?", user.ID).Error)
	assert.Equal(t, user.Username, found.Username)
}

func TestCreateTestUser_UniqueUsernames(t *testing.T) {
	t.Parallel()

	db := cryptoutilTestingTestdb.RequireNewInMemorySQLiteDB(t, &cryptoutilAppsTemplateServiceServerRepository.Tenant{}, &cryptoutilAppsTemplateServiceServerRepository.User{})

	tenant := cryptoutilTestingFixtures.CreateTestTenant(t, db)
	user1 := cryptoutilTestingFixtures.CreateTestUser(t, db, tenant.ID)
	user2 := cryptoutilTestingFixtures.CreateTestUser(t, db, tenant.ID)

	assert.NotEqual(t, user1.ID, user2.ID, "each call should produce a unique ID")
	assert.NotEqual(t, user1.Username, user2.Username, "each call should produce a unique username")
}
