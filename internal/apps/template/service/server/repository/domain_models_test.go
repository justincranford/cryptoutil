// Copyright (c) 2025 Justin Cranford
//
//

package repository_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

// TestClient_TableName tests Client table name.
func TestClient_TableName(t *testing.T) {
	t.Parallel()

	client := cryptoutilAppsTemplateServiceServerRepository.Client{}
	require.Equal(t, "clients", client.TableName())
}

// TestClient_IsActive tests IsActive method.
func TestClient_IsActive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		active   int
		expected bool
	}{
		{"Active client", 1, true},
		{"Inactive client", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := &cryptoutilAppsTemplateServiceServerRepository.Client{Active: tt.active}
			require.Equal(t, tt.expected, client.IsActive())
		})
	}
}

// TestClient_SetActive tests SetActive method.
func TestClient_SetActive(t *testing.T) {
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

			client := &cryptoutilAppsTemplateServiceServerRepository.Client{}
			client.SetActive(tt.active)
			require.Equal(t, tt.expected, client.Active)
		})
	}
}

// TestUnverifiedUser_TableName tests UnverifiedUser table name.
func TestUnverifiedUser_TableName(t *testing.T) {
	t.Parallel()

	user := cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser{}
	require.Equal(t, "unverified_users", user.TableName())
}

// TestUnverifiedUser_IsExpired tests IsExpired method.
func TestUnverifiedUser_IsExpired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		expiresAt time.Time
		expected  bool
	}{
		{"Expired user", time.Now().UTC().Add(-1 * time.Hour), true},
		{"Not expired user", time.Now().UTC().Add(1 * time.Hour), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			user := &cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser{ExpiresAt: tt.expiresAt}
			require.Equal(t, tt.expected, user.IsExpired())
		})
	}
}

// TestUnverifiedClient_TableName tests UnverifiedClient table name.
func TestUnverifiedClient_TableName(t *testing.T) {
	t.Parallel()

	client := cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient{}
	require.Equal(t, "unverified_clients", client.TableName())
}

// TestUnverifiedClient_IsExpired tests IsExpired method.
func TestUnverifiedClient_IsExpired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		expiresAt time.Time
		expected  bool
	}{
		{"Expired client", time.Now().UTC().Add(-1 * time.Hour), true},
		{"Not expired client", time.Now().UTC().Add(1 * time.Hour), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := &cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient{ExpiresAt: tt.expiresAt}
			require.Equal(t, tt.expected, client.IsExpired())
		})
	}
}

// TestRole_TableName tests Role table name.
func TestRole_TableName(t *testing.T) {
	t.Parallel()

	role := cryptoutilAppsTemplateServiceServerRepository.Role{}
	require.Equal(t, "roles", role.TableName())
}

// TestUserRole_TableName tests UserRole table name.
func TestUserRole_TableName(t *testing.T) {
	t.Parallel()

	userRole := cryptoutilAppsTemplateServiceServerRepository.UserRole{}
	require.Equal(t, "user_roles", userRole.TableName())
}

// TestClientRole_TableName tests ClientRole table name.
func TestClientRole_TableName(t *testing.T) {
	t.Parallel()

	clientRole := cryptoutilAppsTemplateServiceServerRepository.ClientRole{}
	require.Equal(t, "client_roles", clientRole.TableName())
}

// TestTenantRealm_TableName tests TenantRealm table name.
func TestTenantRealm_TableName(t *testing.T) {
	t.Parallel()

	tenantRealm := cryptoutilAppsTemplateServiceServerRepository.TenantRealm{}
	require.Equal(t, "tenant_realms", tenantRealm.TableName())
}
