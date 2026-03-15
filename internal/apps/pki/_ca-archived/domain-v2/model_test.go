// Copyright (c) 2025 Justin Cranford
//

package domain

import (
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type tableNamer interface {
	TableName() string
}

func TestTableNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		model     tableNamer
		wantTable string
	}{
		{name: "CAItem", model: &CAItem{}, wantTable: "ca_items"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.wantTable, tc.model.TableName())
		})
	}
}

func TestCAItem_Fields(t *testing.T) {
	t.Parallel()

	id := googleUuid.New()
	tenantID := googleUuid.New()
	now := time.Now().UTC()

	item := CAItem{
		ID:        id,
		TenantID:  tenantID,
		CreatedAt: now,
	}

	require.Equal(t, id, item.ID)
	require.Equal(t, tenantID, item.TenantID)
	require.Equal(t, now, item.CreatedAt)
}

func TestCAItem_ZeroValue(t *testing.T) {
	t.Parallel()

	var item CAItem
	require.Equal(t, googleUuid.UUID{}, item.ID)
	require.Equal(t, googleUuid.UUID{}, item.TenantID)
	require.True(t, item.CreatedAt.IsZero())
}
