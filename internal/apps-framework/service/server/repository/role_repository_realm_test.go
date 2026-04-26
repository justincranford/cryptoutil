// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestTenantRealmRepository_ListByTenant(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	realmRepo := NewTenantRealmRepository(db)
	ctx := context.Background()

	tenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueRoleTenantName("Test"),
		Description: "Test tenant",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}

	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	activeRealm := &TenantRealm{
		ID:        googleUuid.New(),
		TenantID:  tenant.ID,
		RealmID:   googleUuid.New(),
		Type:      "DB",
		Active:    true,
		CreatedAt: time.Now().UTC(),
	}

	inactiveRealm := &TenantRealm{
		ID:        googleUuid.New(),
		TenantID:  tenant.ID,
		RealmID:   googleUuid.New(),
		Type:      "FILE",
		Active:    false,
		CreatedAt: time.Now().UTC(),
	}

	err = realmRepo.Create(ctx, activeRealm)
	require.NoError(t, err)

	err = realmRepo.Create(ctx, inactiveRealm)
	require.NoError(t, err)

	tests := []struct {
		name       string
		activeOnly bool
		minCount   int
	}{
		{
			name:       "all realms",
			activeOnly: false,
			minCount:   2,
		},
		{
			name:       "active realms only",
			activeOnly: true,
			minCount:   1,
		},
	}

	var foundActive, foundInactive bool

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := realmRepo.ListByTenant(ctx, tenant.ID, tt.activeOnly)

			require.NoError(t, err)
			require.GreaterOrEqual(t, len(result), tt.minCount)

			if tt.activeOnly {
				for _, realm := range result {
					require.True(t, realm.Active)

					if realm.ID == activeRealm.ID {
						foundActive = true
					}
				}
			} else {
				for _, realm := range result {
					if realm.ID == activeRealm.ID {
						foundActive = true
					}

					if realm.ID == inactiveRealm.ID {
						foundInactive = true
					}
				}
			}
		})
	}

	require.True(t, foundActive, "Active realm should be found")
	require.True(t, foundInactive, "Inactive realm should be found in all realms list")
}
