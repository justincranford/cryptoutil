// Copyright 2025 Cisco Systems, Inc. and its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite" // CGO-free SQLite driver

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedPwdgen "cryptoutil/internal/shared/pwdgen"
)

// setupRealmTestDB creates an in-memory SQLite database for testing realm service.
func setupRealmTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Create unique database name to avoid sharing between tests.
	dbName := fmt.Sprintf("file:test_%s.db?mode=memory&cache=private", strings.ReplaceAll(t.Name(), "/", "_"))
	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dbName)
	require.NoError(t, err)

	// Enable WAL mode for better concurrency.
	_, err = sqlDB.ExecContext(context.Background(), "PRAGMA journal_mode=WAL;")
	require.NoError(t, err)

	// Set busy timeout for concurrent writes.
	_, err = sqlDB.ExecContext(context.Background(), "PRAGMA busy_timeout = 30000;")
	require.NoError(t, err)

	// Pass to GORM with auto-transactions disabled.
	dialector := sqlite.Dialector{Conn: sqlDB}
	db, err := gorm.Open(dialector, &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	// Configure connection pool.
	sqlDB, err = db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)
	sqlDB.SetMaxIdleConns(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)
	sqlDB.SetConnMaxLifetime(0)

	// Auto-migrate all required tables.
	err = db.AutoMigrate(
		&cryptoutilAppsTemplateServiceServerRepository.Tenant{},
		&cryptoutilAppsTemplateServiceServerRepository.TenantRealm{},
	)
	require.NoError(t, err)

	return db
}

// setupRealmService creates a RealmService with all dependencies for testing.
func setupRealmService(t *testing.T) (RealmService, *gorm.DB) {
	t.Helper()

	db := setupRealmTestDB(t)
	realmRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRealmRepository(db)
	svc := NewRealmService(realmRepo)

	return svc, db
}

// createRealmTestTenant creates a tenant for testing realms.
func createRealmTestTenant(t *testing.T, db *gorm.DB, tenantName string) *cryptoutilAppsTemplateServiceServerRepository.Tenant {
	t.Helper()

	tenant := &cryptoutilAppsTemplateServiceServerRepository.Tenant{
		ID:          googleUuid.New(),
		Name:        tenantName,
		Description: "Test tenant for realm testing",
		Active:      1,
	}
	require.NoError(t, db.Create(tenant).Error)

	return tenant
}

// TestRealmService_CreateRealm_UsernamePassword tests creating a username/password realm.
func TestRealmService_CreateRealm_UsernamePassword(t *testing.T) {
	t.Parallel()

	svc, db := setupRealmService(t)
	ctx := context.Background()

	tenant := createRealmTestTenant(t, db, "realm-username-"+googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength])

	config := &UsernamePasswordConfig{
		MinPasswordLength: cryptoutilSharedMagic.IMMinPasswordLength,
		RequireUppercase:  true,
		RequireLowercase:  true,
		RequireDigit:      true,
		RequireSpecial:    false,
	}

	realm, err := svc.CreateRealm(ctx, tenant.ID, string(RealmTypeUsernamePassword), config)
	require.NoError(t, err)
	require.NotNil(t, realm)
	require.Equal(t, tenant.ID, realm.TenantID)
	require.Equal(t, string(RealmTypeUsernamePassword), realm.Type)
	require.True(t, realm.Active)
	require.Equal(t, "db", realm.Source)
	require.NotEmpty(t, realm.Config)
}

// TestRealmService_CreateRealm_LDAP tests creating an LDAP realm.
func TestRealmService_CreateRealm_LDAP(t *testing.T) {
	t.Parallel()

	svc, db := setupRealmService(t)
	ctx := context.Background()

	tenant := createRealmTestTenant(t, db, "realm-ldap-"+googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength])

	pwdGen, err := cryptoutilSharedPwdgen.NewPasswordGenerator(cryptoutilSharedPwdgen.BasicPolicy)
	require.NoError(t, err)
	bindPassword, err := pwdGen.Generate()
	require.NoError(t, err)

	config := &LDAPConfig{
		URL:           "ldap://ldap.example.com:389",
		BindDN:        "cn=admin,dc=example,dc=com",
		BindPassword:  bindPassword,
		BaseDN:        "dc=example,dc=com",
		UserFilter:    "(uid=%s)",
		GroupFilter:   "(member=%s)",
		UseTLS:        true,
		SkipTLSVerify: false,
	}

	realm, err := svc.CreateRealm(ctx, tenant.ID, string(RealmTypeLDAP), config)
	require.NoError(t, err)
	require.NotNil(t, realm)
	require.Equal(t, string(RealmTypeLDAP), realm.Type)
}

// TestRealmService_CreateRealm_OAuth2 tests creating an OAuth2 realm.
func TestRealmService_CreateRealm_OAuth2(t *testing.T) {
	t.Parallel()

	svc, db := setupRealmService(t)
	ctx := context.Background()

	tenant := createRealmTestTenant(t, db, "realm-oauth2-"+googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength])

	pwdGen, err := cryptoutilSharedPwdgen.NewPasswordGenerator(cryptoutilSharedPwdgen.StrongPolicy)
	require.NoError(t, err)
	clientSecret, err := pwdGen.Generate()
	require.NoError(t, err)

	config := &OAuth2Config{
		ProviderURL:  "https://auth.example.com",
		ClientID:     "my-client-id",
		ClientSecret: clientSecret,
		Scopes:       []string{cryptoutilSharedMagic.ScopeOpenID, cryptoutilSharedMagic.ClaimProfile, cryptoutilSharedMagic.ClaimEmail},
		RedirectURI:  "https://myapp.example.com/callback",
		UseDiscovery: true,
	}

	realm, err := svc.CreateRealm(ctx, tenant.ID, string(RealmTypeOAuth2), config)
	require.NoError(t, err)
	require.NotNil(t, realm)
	require.Equal(t, string(RealmTypeOAuth2), realm.Type)
}

// TestRealmService_CreateRealm_SAML tests creating a SAML realm.
func TestRealmService_CreateRealm_SAML(t *testing.T) {
	t.Parallel()

	svc, db := setupRealmService(t)
	ctx := context.Background()

	tenant := createRealmTestTenant(t, db, "realm-saml-"+googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength])

	config := &SAMLConfig{
		MetadataURL:  "https://idp.example.com/metadata",
		EntityID:     "https://myapp.example.com",
		AssertionURL: "https://myapp.example.com/saml/acs",
		SignRequests: true,
	}

	realm, err := svc.CreateRealm(ctx, tenant.ID, string(RealmTypeSAML), config)
	require.NoError(t, err)
	require.NotNil(t, realm)
	require.Equal(t, string(RealmTypeSAML), realm.Type)
}

// TestRealmService_CreateRealm_InvalidType tests creating a realm with invalid type.
func TestRealmService_CreateRealm_InvalidType(t *testing.T) {
	t.Parallel()

	svc, db := setupRealmService(t)
	ctx := context.Background()

	tenant := createRealmTestTenant(t, db, "realm-invalid-"+googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength])

	_, err := svc.CreateRealm(ctx, tenant.ID, "invalid_type", nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported realm type")
}

// TestRealmService_CreateRealm_InvalidConfig tests creating a realm with invalid configuration.
func TestRealmService_CreateRealm_InvalidConfig(t *testing.T) {
	t.Parallel()

	svc, db := setupRealmService(t)
	ctx := context.Background()

	tenant := createRealmTestTenant(t, db, "realm-invalid-config-"+googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength])

	// LDAP config without required URL.
	config := &LDAPConfig{
		BaseDN: "dc=example,dc=com",
		// Missing required URL.
	}

	_, err := svc.CreateRealm(ctx, tenant.ID, string(RealmTypeLDAP), config)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid realm configuration")
}

// TestRealmService_GetRealm tests retrieving a realm by ID.
func TestRealmService_GetRealm(t *testing.T) {
	t.Parallel()

	svc, db := setupRealmService(t)
	ctx := context.Background()

	tenant := createRealmTestTenant(t, db, "realm-get-"+googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength])

	config := &UsernamePasswordConfig{MinPasswordLength: cryptoutilSharedMagic.IMMinPasswordLength}
	created, err := svc.CreateRealm(ctx, tenant.ID, string(RealmTypeUsernamePassword), config)
	require.NoError(t, err)

	// Get realm.
	retrieved, err := svc.GetRealm(ctx, tenant.ID, created.RealmID)
	require.NoError(t, err)
	require.Equal(t, created.ID, retrieved.ID)
	require.Equal(t, created.RealmID, retrieved.RealmID)
	require.Equal(t, created.Type, retrieved.Type)
}

// TestRealmService_GetRealm_WrongTenant tests getting realm with wrong tenant.
func TestRealmService_GetRealm_WrongTenant(t *testing.T) {
	t.Parallel()

	svc, db := setupRealmService(t)
	ctx := context.Background()

	tenant1 := createRealmTestTenant(t, db, "realm-tenant1-"+googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength])
	tenant2 := createRealmTestTenant(t, db, "realm-tenant2-"+googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength])

	config := &UsernamePasswordConfig{MinPasswordLength: cryptoutilSharedMagic.IMMinPasswordLength}
	created, err := svc.CreateRealm(ctx, tenant1.ID, string(RealmTypeUsernamePassword), config)
	require.NoError(t, err)

	// Try to get realm from wrong tenant.
	_, err = svc.GetRealm(ctx, tenant2.ID, created.RealmID)
	require.Error(t, err)
}

// TestRealmService_ListRealms tests listing realms for a tenant.
func TestRealmService_ListRealms(t *testing.T) {
	t.Parallel()

	svc, db := setupRealmService(t)
	ctx := context.Background()

	tenant := createRealmTestTenant(t, db, "realm-list-"+googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength])

	// Create multiple realms.
	config1 := &UsernamePasswordConfig{MinPasswordLength: cryptoutilSharedMagic.IMMinPasswordLength}
	_, err := svc.CreateRealm(ctx, tenant.ID, string(RealmTypeUsernamePassword), config1)
	require.NoError(t, err)

	config2 := &LDAPConfig{URL: "ldap://ldap.example.com", BaseDN: "dc=example,dc=com"}
	_, err = svc.CreateRealm(ctx, tenant.ID, string(RealmTypeLDAP), config2)
	require.NoError(t, err)

	// List all realms.
	realms, err := svc.ListRealms(ctx, tenant.ID, false)
	require.NoError(t, err)
	require.Len(t, realms, 2)
}

// TestRealmService_ListRealms_ActiveOnly tests listing only active realms.
func TestRealmService_ListRealms_ActiveOnly(t *testing.T) {
	t.Parallel()

	svc, db := setupRealmService(t)
	ctx := context.Background()

	tenant := createRealmTestTenant(t, db, "realm-active-"+googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength])

	// Create active realm.
	config1 := &UsernamePasswordConfig{MinPasswordLength: cryptoutilSharedMagic.IMMinPasswordLength}
	_, err := svc.CreateRealm(ctx, tenant.ID, string(RealmTypeUsernamePassword), config1)
	require.NoError(t, err)

	// Create and deactivate realm.
	config2 := &LDAPConfig{URL: "ldap://ldap.example.com", BaseDN: "dc=example,dc=com"}
	inactive, err := svc.CreateRealm(ctx, tenant.ID, string(RealmTypeLDAP), config2)
	require.NoError(t, err)

	err = svc.DeleteRealm(ctx, tenant.ID, inactive.RealmID) // Soft delete.
	require.NoError(t, err)

	// List only active realms.
	realms, err := svc.ListRealms(ctx, tenant.ID, true)
	require.NoError(t, err)
	require.Len(t, realms, 1)
	require.Equal(t, string(RealmTypeUsernamePassword), realms[0].Type)
}

// TestRealmService_UpdateRealm tests updating a realm's configuration.
func TestRealmService_UpdateRealm(t *testing.T) {
	t.Parallel()

	svc, db := setupRealmService(t)
	ctx := context.Background()

	tenant := createRealmTestTenant(t, db, "realm-update-"+googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength])

	config := &UsernamePasswordConfig{MinPasswordLength: cryptoutilSharedMagic.IMMinPasswordLength}
	created, err := svc.CreateRealm(ctx, tenant.ID, string(RealmTypeUsernamePassword), config)
	require.NoError(t, err)

	// Update configuration.
	newConfig := &UsernamePasswordConfig{
		MinPasswordLength: cryptoutilSharedMagic.HashPrefixLength,
		RequireUppercase:  true,
	}

	updated, err := svc.UpdateRealm(ctx, tenant.ID, created.RealmID, newConfig, nil)
	require.NoError(t, err)
	require.NotNil(t, updated)

	// Verify config was updated.
	parsedConfig, err := svc.GetRealmConfig(ctx, tenant.ID, created.RealmID)
	require.NoError(t, err)

	pwConfig, ok := parsedConfig.(*UsernamePasswordConfig)
	require.True(t, ok)
	require.Equal(t, cryptoutilSharedMagic.HashPrefixLength, pwConfig.MinPasswordLength)
	require.True(t, pwConfig.RequireUppercase)
}

// TestRealmService_UpdateRealm_ActiveFlag tests updating realm active status.
func TestRealmService_UpdateRealm_ActiveFlag(t *testing.T) {
	t.Parallel()

	svc, db := setupRealmService(t)
	ctx := context.Background()

	tenant := createRealmTestTenant(t, db, "realm-active-flag-"+googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength])

	config := &UsernamePasswordConfig{MinPasswordLength: cryptoutilSharedMagic.IMMinPasswordLength}
	created, err := svc.CreateRealm(ctx, tenant.ID, string(RealmTypeUsernamePassword), config)
	require.NoError(t, err)
	require.True(t, created.Active)

	// Deactivate realm.
	active := false
	updated, err := svc.UpdateRealm(ctx, tenant.ID, created.RealmID, nil, &active)
	require.NoError(t, err)
	require.False(t, updated.Active)
}

// TestRealmService_DeleteRealm tests soft deleting a realm.
func TestRealmService_DeleteRealm(t *testing.T) {
	t.Parallel()

	svc, db := setupRealmService(t)
	ctx := context.Background()

	tenant := createRealmTestTenant(t, db, "realm-delete-"+googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength])

	config := &UsernamePasswordConfig{MinPasswordLength: cryptoutilSharedMagic.IMMinPasswordLength}
	created, err := svc.CreateRealm(ctx, tenant.ID, string(RealmTypeUsernamePassword), config)
	require.NoError(t, err)

	// Delete realm.
	err = svc.DeleteRealm(ctx, tenant.ID, created.RealmID)
	require.NoError(t, err)

	// Verify realm is inactive but still exists.
	retrieved, err := svc.GetRealm(ctx, tenant.ID, created.RealmID)
	require.NoError(t, err)
	require.False(t, retrieved.Active)
}

// TestRealmService_GetRealmConfig tests parsing realm configuration.
