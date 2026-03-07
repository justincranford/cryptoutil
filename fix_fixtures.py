#!/usr/bin/env python3
"""Write fixtures package files without BOM."""
import os

base = os.path.join("internal", "apps", "template", "service", "testing", "fixtures")
os.makedirs(base, exist_ok=True)

fixtures_go = (
    "// Copyright (c) 2025 Justin Cranford\n"
    "//\n"
    "\n"
    "// Package fixtures provides test entity factories for cryptoutil service tests.\n"
    "// Each factory creates a unique, persisted entity using UUIDv7 for deterministic test isolation.\n"
    "package fixtures\n"
    "\n"
    "import (\n"
    '\t"testing"\n'
    '\t"time"\n'
    "\n"
    '\tgoogleUuid "github.com/google/uuid"\n'
    '\t"github.com/stretchr/testify/require"\n'
    '\t"gorm.io/gorm"\n'
    "\n"
    '\tcryptoutilRepo "cryptoutil/internal/apps/template/service/server/repository"\n'
    ")\n"
    "\n"
    "// CreateTestTenant creates and persists a unique Tenant entity for testing.\n"
    "func CreateTestTenant(t testing.TB, db *gorm.DB) *cryptoutilRepo.Tenant {\n"
    "\tt.Helper()\n"
    "\n"
    "\tnow := time.Now().UTC()\n"
    "\ttenant := &cryptoutilRepo.Tenant{\n"
    "\t\tID:          googleUuid.Must(googleUuid.NewV7()),\n"
    '\t\tName:        "test-tenant-" + googleUuid.Must(googleUuid.NewV7()).String(),\n'
    '\t\tDescription: "Test tenant for automated testing.",\n'
    "\t\tActive:      1,\n"
    "\t\tCreatedAt:   now,\n"
    "\t\tUpdatedAt:   now,\n"
    "\t}\n"
    "\n"
    "\trequire.NoError(t, db.Create(tenant).Error)\n"
    "\n"
    "\treturn tenant\n"
    "}\n"
    "\n"
    "// CreateTestRealm creates and persists a unique TenantRealm entity for testing.\n"
    "func CreateTestRealm(t testing.TB, db *gorm.DB, tenantID googleUuid.UUID) *cryptoutilRepo.TenantRealm {\n"
    "\tt.Helper()\n"
    "\n"
    "\tnow := time.Now().UTC()\n"
    "\trealm := &cryptoutilRepo.TenantRealm{\n"
    "\t\tID:        googleUuid.Must(googleUuid.NewV7()),\n"
    "\t\tTenantID:  tenantID,\n"
    "\t\tRealmID:   googleUuid.Must(googleUuid.NewV7()),\n"
    '\t\tType:      "username_password",\n'
    "\t\tActive:    true,\n"
    '\t\tSource:    "db",\n'
    "\t\tCreatedAt: now,\n"
    "\t\tUpdatedAt: now,\n"
    "\t}\n"
    "\n"
    "\trequire.NoError(t, db.Create(realm).Error)\n"
    "\n"
    "\treturn realm\n"
    "}\n"
    "\n"
    "// CreateTestUser creates and persists a unique User entity for testing.\n"
    "func CreateTestUser(t testing.TB, db *gorm.DB, tenantID googleUuid.UUID) *cryptoutilRepo.User {\n"
    "\tt.Helper()\n"
    "\n"
    "\tsuffix := googleUuid.Must(googleUuid.NewV7()).String()[:8]\n"
    "\tnow := time.Now().UTC()\n"
    "\tuser := &cryptoutilRepo.User{\n"
    "\t\tID:           googleUuid.Must(googleUuid.NewV7()),\n"
    "\t\tTenantID:     tenantID,\n"
    '\t\tUsername:     "testuser-" + suffix,\n'
    '\t\tPasswordHash: "test-hash-placeholder-for-testing",\n'
    '\t\tEmail:        "test-" + suffix + "@example.com",\n'
    "\t\tActive:       1,\n"
    "\t\tCreatedAt:    now,\n"
    "\t\tUpdatedAt:    now,\n"
    "\t}\n"
    "\n"
    "\trequire.NoError(t, db.Create(user).Error)\n"
    "\n"
    "\treturn user\n"
    "}\n"
)

fixtures_test_go = (
    "// Copyright (c) 2025 Justin Cranford\n"
    "//\n"
    "\n"
    "package fixtures_test\n"
    "\n"
    "import (\n"
    '\t"testing"\n'
    "\n"
    '\tgoogleUuid "github.com/google/uuid"\n'
    '\t"github.com/stretchr/testify/assert"\n'
    '\t"github.com/stretchr/testify/require"\n'
    "\n"
    '\tcryptoutilRepo "cryptoutil/internal/apps/template/service/server/repository"\n'
    '\tcryptoutilTestingFixtures "cryptoutil/internal/apps/template/service/testing/fixtures"\n'
    '\tcryptoutilTestingTestdb "cryptoutil/internal/apps/template/service/testing/testdb"\n'
    ")\n"
    "\n"
    "func TestCreateTestTenant(t *testing.T) {\n"
    "\tt.Parallel()\n"
    "\n"
    "\tdb := cryptoutilTestingTestdb.RequireNewInMemorySQLiteDB(t, &cryptoutilRepo.Tenant{})\n"
    "\n"
    "\ttenant := cryptoutilTestingFixtures.CreateTestTenant(t, db)\n"
    "\n"
    "\trequire.NotNil(t, tenant)\n"
    '\tassert.NotEqual(t, googleUuid.Nil, tenant.ID, "ID should not be zero UUID")\n'
    '\tassert.NotEmpty(t, tenant.Name, "Name should not be empty")\n'
    '\tassert.Equal(t, 1, tenant.Active, "tenant should be active")\n'
    "\n"
    "\tvar found cryptoutilRepo.Tenant\n"
    '\trequire.NoError(t, db.First(&found, "id = ?", tenant.ID).Error)\n'
    "\tassert.Equal(t, tenant.Name, found.Name)\n"
    "}\n"
    "\n"
    "func TestCreateTestTenant_UniquePerCall(t *testing.T) {\n"
    "\tt.Parallel()\n"
    "\n"
    "\tdb := cryptoutilTestingTestdb.RequireNewInMemorySQLiteDB(t, &cryptoutilRepo.Tenant{})\n"
    "\n"
    "\ttenant1 := cryptoutilTestingFixtures.CreateTestTenant(t, db)\n"
    "\ttenant2 := cryptoutilTestingFixtures.CreateTestTenant(t, db)\n"
    "\n"
    '\tassert.NotEqual(t, tenant1.ID, tenant2.ID, "each call should produce a unique ID")\n'
    '\tassert.NotEqual(t, tenant1.Name, tenant2.Name, "each call should produce a unique name")\n'
    "}\n"
    "\n"
    "func TestCreateTestRealm(t *testing.T) {\n"
    "\tt.Parallel()\n"
    "\n"
    "\tdb := cryptoutilTestingTestdb.RequireNewInMemorySQLiteDB(t, &cryptoutilRepo.Tenant{}, &cryptoutilRepo.TenantRealm{})\n"
    "\n"
    "\ttenant := cryptoutilTestingFixtures.CreateTestTenant(t, db)\n"
    "\trealm := cryptoutilTestingFixtures.CreateTestRealm(t, db, tenant.ID)\n"
    "\n"
    "\trequire.NotNil(t, realm)\n"
    '\tassert.NotEqual(t, googleUuid.Nil, realm.ID, "realm ID should not be zero UUID")\n'
    "\tassert.Equal(t, tenant.ID, realm.TenantID)\n"
    '\tassert.NotEqual(t, googleUuid.Nil, realm.RealmID, "realm RealmID should not be zero UUID")\n'
    '\tassert.True(t, realm.Active, "realm should be active")\n'
    "\n"
    "\tvar found cryptoutilRepo.TenantRealm\n"
    '\trequire.NoError(t, db.First(&found, "id = ?", realm.ID).Error)\n'
    "\tassert.Equal(t, realm.TenantID, found.TenantID)\n"
    "}\n"
    "\n"
    "func TestCreateTestUser(t *testing.T) {\n"
    "\tt.Parallel()\n"
    "\n"
    "\tdb := cryptoutilTestingTestdb.RequireNewInMemorySQLiteDB(t, &cryptoutilRepo.Tenant{}, &cryptoutilRepo.User{})\n"
    "\n"
    "\ttenant := cryptoutilTestingFixtures.CreateTestTenant(t, db)\n"
    "\tuser := cryptoutilTestingFixtures.CreateTestUser(t, db, tenant.ID)\n"
    "\n"
    "\trequire.NotNil(t, user)\n"
    '\tassert.NotEqual(t, googleUuid.Nil, user.ID, "user ID should not be zero UUID")\n'
    "\tassert.Equal(t, tenant.ID, user.TenantID)\n"
    '\tassert.NotEmpty(t, user.Username, "username should not be empty")\n'
    '\tassert.Equal(t, 1, user.Active, "user should be active")\n'
    "\n"
    "\tvar found cryptoutilRepo.User\n"
    '\trequire.NoError(t, db.First(&found, "id = ?", user.ID).Error)\n'
    "\tassert.Equal(t, user.Username, found.Username)\n"
    "}\n"
    "\n"
    "func TestCreateTestUser_UniqueUsernames(t *testing.T) {\n"
    "\tt.Parallel()\n"
    "\n"
    "\tdb := cryptoutilTestingTestdb.RequireNewInMemorySQLiteDB(t, &cryptoutilRepo.Tenant{}, &cryptoutilRepo.User{})\n"
    "\n"
    "\ttenant := cryptoutilTestingFixtures.CreateTestTenant(t, db)\n"
    "\tuser1 := cryptoutilTestingFixtures.CreateTestUser(t, db, tenant.ID)\n"
    "\tuser2 := cryptoutilTestingFixtures.CreateTestUser(t, db, tenant.ID)\n"
    "\n"
    '\tassert.NotEqual(t, user1.ID, user2.ID, "each call should produce a unique ID")\n'
    '\tassert.NotEqual(t, user1.Username, user2.Username, "each call should produce a unique username")\n'
    "}\n"
)

for fname, content in [("fixtures.go", fixtures_go), ("fixtures_test.go", fixtures_test_go)]:
    path = os.path.join(base, fname)
    with open(path, "w", encoding="utf-8", newline="\n") as f:
        f.write(content)
    print(f"Written: {path}")
