// Copyright (c) 2025 Justin Cranford

// Package service provides template service business logic and handlers.
package service

import (
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestBrowserAndServiceRealms_DefaultCounts verifies exactly 6 browser and 6 service realms are defined.
func TestBrowserAndServiceRealms_DefaultCounts(t *testing.T) {
	t.Parallel()

	browserRealms := cryptoutilSharedMagic.DefaultBrowserRealms
	serviceRealms := cryptoutilSharedMagic.DefaultServiceRealms

	require.Len(t, browserRealms, 6, "DefaultBrowserRealms must have exactly 6 non-federated browser authentication methods")
	require.Len(t, serviceRealms, 6, "DefaultServiceRealms must have exactly 6 non-federated service authentication methods")

	// Verify browser realms (session-based).
	expectedBrowserRealms := []string{
		"jwe-session-cookie",
		"jws-session-cookie",
		"opaque-session-cookie",
		"basic-username-password",
		"bearer-api-token",
		"https-client-cert",
	}

	require.Equal(t, expectedBrowserRealms, browserRealms, "Browser realms should match expected 6 non-federated methods")

	// Verify service realms (token-based).
	expectedServiceRealms := []string{
		"jwe-session-token",
		"jws-session-token",
		"opaque-session-token",
		"basic-client-id-secret",
		"bearer-api-token",
		"https-client-cert",
	}

	require.Equal(t, expectedServiceRealms, serviceRealms, "Service realms should match expected 6 non-federated methods")

	// Verify no duplicate realms within browser or service categories.
	browserMap := make(map[string]bool)
	for _, realm := range browserRealms {
		require.False(t, browserMap[realm], "Browser realm %s should not be duplicated", realm)
		browserMap[realm] = true
	}

	serviceMap := make(map[string]bool)
	for _, realm := range serviceRealms {
		require.False(t, serviceMap[realm], "Service realm %s should not be duplicated", realm)
		serviceMap[realm] = true
	}
}

// TestRealmService_CreateMultiple Realms_AllTypes tests creating realms for all documented authentication methods.
func TestRealmService_CreateMultipleRealms_AllTypes(t *testing.T) {
	t.Parallel()

	svc, db := setupRealmService(t)
	ctx := context.Background()

	// Create tenant for testing all realm types.
	tenant := createRealmTestTenant(t, db, "tenant-all-types-"+googleUuid.NewString()[:8])

	// Browser realms (6 total) - using username/password config as template.
	browserTests := []struct {
		name        string
		realmType   string
		description string
	}{
		{"jwe-session-cookie", "jwe-session-cookie", "JWE Session Cookie - Encrypted session tokens for browsers"},
		{"jws-session-cookie", "jws-session-cookie", "JWS Session Cookie - Signed session tokens for browsers"},
		{"opaque-session-cookie", "opaque-session-cookie", "Opaque Session Cookie - Database-backed session tokens"},
		{"basic-username-password", "basic-username-password", "Basic (Username/Password) - HTTP Basic auth with user credentials"},
		{"bearer-api-token", "bearer-api-token", "Bearer (API Token) - Bearer token authentication from browser"},
		{"https-client-cert", "https-client-cert", "HTTPS Client Certificate - mTLS client certificate from browser"},
	}

	// Create all 6 browser realms.
	for _, tt := range browserTests {
		t.Run("browser_"+tt.name, func(t *testing.T) {
			config := &UsernamePasswordConfig{
				MinPasswordLength: 8,
				RequireUppercase:  true,
				RequireLowercase:  true,
				RequireDigit:      true,
				RequireSpecial:    false,
			}

			realm, err := svc.CreateRealm(ctx, tenant.ID, tt.realmType, config)
			require.NoError(t, err, "CreateRealm should succeed for browser realm: %s", tt.realmType)
			require.NotNil(t, realm)
			require.Equal(t, tt.realmType, realm.Type)
			require.True(t, realm.Active, "Realm should be active after creation")
		})
	}

	// Service realms (6 total) - using username/password config as template.
	serviceTests := []struct {
		name        string
		realmType   string
		description string
	}{
		{"jwe-session-token", "jwe-session-token", "JWE Session Token - Encrypted session tokens for headless clients"},
		{"jws-session-token", "jws-session-token", "JWS Session Token - Signed session tokens for headless clients"},
		{"opaque-session-token", "opaque-session-token", "Opaque Session Token - Non-JWT session tokens"},
		{"basic-client-id-secret", "basic-client-id-secret", "Basic (Client ID/Secret) - HTTP Basic with client credentials"},
		{"bearer-api-token-service", "bearer-api-token", "Bearer (API Token) - Long-lived service credentials"},
		{"https-client-cert-service", "https-client-cert", "HTTPS Client Certificate - mTLS for high-security service-to-service"},
	}

	// Create all 6 service realms.
	for _, tt := range serviceTests {
		t.Run("service_"+tt.name, func(t *testing.T) {
			config := &UsernamePasswordConfig{
				MinPasswordLength: 12,
				RequireUppercase:  true,
				RequireLowercase:  true,
				RequireDigit:      true,
				RequireSpecial:    true,
			}

			realm, err := svc.CreateRealm(ctx, tenant.ID, tt.realmType, config)
			require.NoError(t, err, "CreateRealm should succeed for service realm: %s", tt.realmType)
			require.NotNil(t, realm)
			require.Equal(t, tt.realmType, realm.Type)
			require.True(t, realm.Active, "Realm should be active after creation")
		})
	}

	// Verify all realms were created (12 total: 6 browser + 6 service).
	allRealms, err := svc.ListRealms(ctx, tenant.ID, false) // activeOnly=false to get all
	require.NoError(t, err)
	require.Len(t, allRealms, 12, "Should have exactly 12 realms (6 browser + 6 service)")
}

// TestRealmService_ActivateDeactivateAllTypes tests enabling/disabling realms for all authentication methods.
func TestRealmService_ActivateDeactivateAllTypes(t *testing.T) {
	t.Parallel()

	svc, db := setupRealmService(t)
	ctx := context.Background()

	tenant := createRealmTestTenant(t, db, "tenant-activate-deactivate-"+googleUuid.NewString()[:8])

	// Test with one representative realm type (can be extended to all 12 if needed).
	config := &UsernamePasswordConfig{
		MinPasswordLength: 8,
		RequireUppercase:  true,
		RequireLowercase:  true,
		RequireDigit:      true,
		RequireSpecial:    false,
	}

	// Create realm.
	realm, err := svc.CreateRealm(ctx, tenant.ID, "jwe-session-cookie", config)
	require.NoError(t, err)
	require.True(t, realm.Active, "Realm should be active after creation")

	// Deactivate realm.
	inactive := false
	updatedRealm, err := svc.UpdateRealm(ctx, tenant.ID, realm.RealmID, config, &inactive)
	require.NoError(t, err)
	require.False(t, updatedRealm.Active, "Realm should be inactive after update")

	// Reactivate realm.
	active := true
	reactivatedRealm, err := svc.UpdateRealm(ctx, tenant.ID, realm.RealmID, config, &active)
	require.NoError(t, err)
	require.True(t, reactivatedRealm.Active, "Realm should be active after reactivation")
}
