// Copyright (c) 2025-2026 Justin Cranford.

package repository

import (
	"context"
	"errors"
	"strings"
	"testing"

	cryptoutilAppsSkeletonTemplateServerModel "cryptoutil/internal/apps/skeleton-template/server/model"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestItemRepository_ListFetchError(t *testing.T) {
	t.Parallel()

	db := newRepoTestDB(t)
	require.NoError(t, db.Callback().Query().After("gorm:query").Register("test:fail_list_fetch", func(tx *gorm.DB) {
		if strings.HasPrefix(strings.TrimSpace(tx.Statement.SQL.String()), "SELECT *") {
			_ = tx.AddError(errors.New("injected list fetch failure"))
		}
	}))

	items, total, err := NewItemRepository(db).List(context.Background(), googleUuid.Must(googleUuid.NewV7()), 1, 10)
	require.Nil(t, items)
	require.Zero(t, total)
	require.ErrorContains(t, err, "failed to list items")
}

func TestItemRepository_UpdateRowsAffectedZero(t *testing.T) {
	t.Parallel()

	db := newRepoTestDB(t)
	require.NoError(t, db.Exec(`
CREATE TRIGGER ignore_template_item_insert
BEFORE INSERT ON template_items
BEGIN
	SELECT RAISE(IGNORE);
END;`).Error)

	item := &cryptoutilAppsSkeletonTemplateServerModel.TemplateItem{
		ID:       googleUuid.Must(googleUuid.NewV7()),
		TenantID: googleUuid.Must(googleUuid.NewV7()),
		Name:     "ignored insert",
	}

	require.ErrorIs(t, NewItemRepository(db).Update(context.Background(), item), gorm.ErrRecordNotFound)
}
