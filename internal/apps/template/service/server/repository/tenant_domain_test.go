// Copyright (c) 2025 Justin Cranford
//
//

package repository_test

import (
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

// TestTenant_TableName tests Tenant table name.
func TestTenant_TableName(t *testing.T) {
	t.Parallel()

	tenant := cryptoutilAppsTemplateServiceServerRepository.Tenant{}
	require.Equal(t, "tenants", tenant.TableName())
}

// TestTenant_IsActive tests IsActive method.
func TestTenant_IsActive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		active   int
		expected bool
	}{
		{"Active tenant", 1, true},
		{"Inactive tenant", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tenant := &cryptoutilAppsTemplateServiceServerRepository.Tenant{Active: tt.active}
			require.Equal(t, tt.expected, tenant.IsActive())
		})
	}
}

// TestTenant_SetActive tests SetActive method.
func TestTenant_SetActive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		active   bool
		expected int
	}{
		{"Set active to true", true, 1},
		{"Set active to false", false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tenant := &cryptoutilAppsTemplateServiceServerRepository.Tenant{}
			tenant.SetActive(tt.active)
			require.Equal(t, tt.expected, tenant.Active)
		})
	}
}

// TestUser_TableName tests User table name.
func TestUser_TableName(t *testing.T) {
	t.Parallel()

	user := cryptoutilAppsTemplateServiceServerRepository.User{}
	require.Equal(t, "users", user.TableName())
}

// TestUser_GetID tests GetID method.
func TestUser_GetID(t *testing.T) {
	t.Parallel()

	id := googleUuid.New()
	user := &cryptoutilAppsTemplateServiceServerRepository.User{ID: id}
	require.Equal(t, id, user.GetID())
}

// TestUser_GetUsername tests GetUsername method.
func TestUser_GetUsername(t *testing.T) {
	t.Parallel()

	user := &cryptoutilAppsTemplateServiceServerRepository.User{Username: "testuser"}
	require.Equal(t, "testuser", user.GetUsername())
}

// TestUser_GetPasswordHash tests GetPasswordHash method.
func TestUser_GetPasswordHash(t *testing.T) {
	t.Parallel()

	user := &cryptoutilAppsTemplateServiceServerRepository.User{PasswordHash: "hash123"}
	require.Equal(t, "hash123", user.GetPasswordHash())
}

// TestUser_SetID tests SetID method.
func TestUser_SetID(t *testing.T) {
	t.Parallel()

	id := googleUuid.New()
	user := &cryptoutilAppsTemplateServiceServerRepository.User{}
	user.SetID(id)
	require.Equal(t, id, user.ID)
}

// TestUser_SetUsername tests SetUsername method.
func TestUser_SetUsername(t *testing.T) {
	t.Parallel()

	user := &cryptoutilAppsTemplateServiceServerRepository.User{}
	user.SetUsername("newuser")
	require.Equal(t, "newuser", user.Username)
}

// TestUser_SetPasswordHash tests SetPasswordHash method.
func TestUser_SetPasswordHash(t *testing.T) {
	t.Parallel()

	user := &cryptoutilAppsTemplateServiceServerRepository.User{}
	user.SetPasswordHash("newhash")
	require.Equal(t, "newhash", user.PasswordHash)
}

// TestUser_GetTenantID tests GetTenantID method.
func TestUser_GetTenantID(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	user := &cryptoutilAppsTemplateServiceServerRepository.User{TenantID: tenantID}
	require.Equal(t, tenantID, user.GetTenantID())
}

// TestUser_IsActive tests IsActive method.
func TestUser_IsActive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		active   int
		expected bool
	}{
		{"Active user", 1, true},
		{"Inactive user", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			user := &cryptoutilAppsTemplateServiceServerRepository.User{Active: tt.active}
			require.Equal(t, tt.expected, user.IsActive())
		})
	}
}

// TestUser_SetActive tests SetActive method.
func TestUser_SetActive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		active   bool
		expected int
	}{
		{"Set active to true", true, 1},
		{"Set active to false", false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			user := &cryptoutilAppsTemplateServiceServerRepository.User{}
			user.SetActive(tt.active)
			require.Equal(t, tt.expected, user.Active)
		})
	}
}

// TestTenantRealm_GetRealmID tests GetRealmID method.
func TestTenantRealm_GetRealmID(t *testing.T) {
	t.Parallel()

	realmID := googleUuid.New()
	tr := &cryptoutilAppsTemplateServiceServerRepository.TenantRealm{
		RealmID: realmID,
	}

	require.Equal(t, realmID, tr.GetRealmID())
}
