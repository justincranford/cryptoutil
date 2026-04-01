// Copyright (c) 2025 Justin Cranford
//
// TEMPLATE: Copy and rename 'skeleton' -> your-service-name before use.

// Package repository provides unit tests for the skeleton-template item repository.
package repository

import (
	"context"
	"database/sql"
	"testing"

	cryptoutilAppsSkeletonTemplateServerModel "cryptoutil/internal/apps/skeleton-template/server/model"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// newRepoTestDB opens a per-test in-memory SQLite DB and runs AutoMigrate.
// MaxOpenConns=1 ensures all GORM operations share the same single connection,
// which is required when using cache=private so that AutoMigrate tables are
// visible to subsequent queries.
func newRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbID, err := googleUuid.NewV7()
	require.NoError(t, err)

	dsn := "file:" + dbID.String() + "?mode=memory&cache=private"

	rawDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dsn)
	require.NoError(t, err)

	_, err = rawDB.ExecContext(context.Background(), "PRAGMA journal_mode=WAL;")
	require.NoError(t, err)

	_, err = rawDB.ExecContext(context.Background(), "PRAGMA busy_timeout = 30000;")
	require.NoError(t, err)

	rawDB.SetMaxOpenConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	rawDB.SetMaxIdleConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	rawDB.SetConnMaxLifetime(0)

	db, err := gorm.Open(sqlite.Dialector{Conn: rawDB}, &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	require.NoError(t, db.AutoMigrate(&cryptoutilAppsSkeletonTemplateServerModel.TemplateItem{}))

	t.Cleanup(func() { _ = rawDB.Close() })

	return db
}

// newRepoBrokenDB opens an in-memory DB WITHOUT creating the template_items
// table so CRUD calls fail with a DB error, exercising the 500/error paths.
func newRepoBrokenDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbID, err := googleUuid.NewV7()
	require.NoError(t, err)

	dsn := "file:" + dbID.String() + "?mode=memory&cache=private"

	rawDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dsn)
	require.NoError(t, err)

	rawDB.SetMaxOpenConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	rawDB.SetMaxIdleConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	rawDB.SetConnMaxLifetime(0)

	db, err := gorm.Open(sqlite.Dialector{Conn: rawDB}, &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	t.Cleanup(func() { _ = rawDB.Close() })

	return db
}

func TestNewItemRepository(t *testing.T) {
	t.Parallel()

	require.NotNil(t, NewItemRepository(newRepoTestDB(t)))
}

func TestItemRepository_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		item    *cryptoutilAppsSkeletonTemplateServerModel.TemplateItem
		useBad  bool
		wantErr bool
	}{
		{
			name: "success",
			item: &cryptoutilAppsSkeletonTemplateServerModel.TemplateItem{
				ID:          googleUuid.Must(googleUuid.NewV7()),
				TenantID:    googleUuid.Must(googleUuid.NewV7()),
				Name:        "test item",
				Description: "desc",
			},
			wantErr: false,
		},
		{
			name: "db_error_no_table",
			item: &cryptoutilAppsSkeletonTemplateServerModel.TemplateItem{
				ID:       googleUuid.Must(googleUuid.NewV7()),
				TenantID: googleUuid.Must(googleUuid.NewV7()),
				Name:     "fail",
			},
			useBad:  true,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var repo *ItemRepository
			if tc.useBad {
				repo = NewItemRepository(newRepoBrokenDB(t))
			} else {
				repo = NewItemRepository(newRepoTestDB(t))
			}

			err := repo.Create(context.Background(), tc.item)
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to create item")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestItemRepository_GetByID(t *testing.T) {
	t.Parallel()

	type setupResult struct{ tenantID, itemID googleUuid.UUID }

	tests := []struct {
		name    string
		setup   func(t *testing.T, repo *ItemRepository) setupResult
		useBad  bool
		wantErr bool
		errMsg  string
	}{
		{
			name: "success",
			setup: func(t *testing.T, repo *ItemRepository) setupResult {
				t.Helper()

				tenantID := googleUuid.Must(googleUuid.NewV7())
				item := &cryptoutilAppsSkeletonTemplateServerModel.TemplateItem{
					ID:       googleUuid.Must(googleUuid.NewV7()),
					TenantID: tenantID,
					Name:     "found",
				}
				require.NoError(t, repo.Create(context.Background(), item))

				return setupResult{tenantID: tenantID, itemID: item.ID}
			},
			wantErr: false,
		},
		{
			name: "not_found",
			setup: func(_ *testing.T, _ *ItemRepository) setupResult {
				return setupResult{
					tenantID: googleUuid.Must(googleUuid.NewV7()),
					itemID:   googleUuid.Must(googleUuid.NewV7()),
				}
			},
			wantErr: true,
			errMsg:  "failed to get item",
		},
		{
			name: "db_error",
			setup: func(_ *testing.T, _ *ItemRepository) setupResult {
				return setupResult{
					tenantID: googleUuid.Must(googleUuid.NewV7()),
					itemID:   googleUuid.Must(googleUuid.NewV7()),
				}
			},
			useBad:  true,
			wantErr: true,
			errMsg:  "failed to get item",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var repo *ItemRepository
			if tc.useBad {
				repo = NewItemRepository(newRepoBrokenDB(t))
			} else {
				repo = NewItemRepository(newRepoTestDB(t))
			}

			ids := tc.setup(t, repo)

			result, err := repo.GetByID(context.Background(), ids.tenantID, ids.itemID)
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, ids.itemID, result.ID)
			}
		})
	}
}

func TestItemRepository_List(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		seedCount int
		page      int
		size      int
		useBad    bool
		wantErr   bool
		wantCount int
	}{
		{name: "empty_list", seedCount: 0, page: 1, size: cryptoutilSharedMagic.SuiteServiceCount, wantCount: 0},
		{name: "with_items", seedCount: 3, page: 1, size: cryptoutilSharedMagic.SuiteServiceCount, wantCount: 3},
		{name: "pagination_page2", seedCount: cryptoutilSharedMagic.DBMaxPingAttempts, page: 2, size: 3, wantCount: 2},
		{name: "db_error", useBad: true, page: 1, size: cryptoutilSharedMagic.SuiteServiceCount, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var repo *ItemRepository
			if tc.useBad {
				repo = NewItemRepository(newRepoBrokenDB(t))
			} else {
				repo = NewItemRepository(newRepoTestDB(t))
			}

			tenantID := googleUuid.Must(googleUuid.NewV7())

			for range tc.seedCount {
				item := &cryptoutilAppsSkeletonTemplateServerModel.TemplateItem{
					ID:       googleUuid.Must(googleUuid.NewV7()),
					TenantID: tenantID,
					Name:     "item",
				}
				require.NoError(t, repo.Create(context.Background(), item))
			}

			items, total, err := repo.List(context.Background(), tenantID, tc.page, tc.size)
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to")
			} else {
				require.NoError(t, err)
				require.Len(t, items, tc.wantCount)
				require.Equal(t, int64(tc.seedCount), total)
			}
		})
	}
}

func TestItemRepository_Update(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(t *testing.T, repo *ItemRepository) *cryptoutilAppsSkeletonTemplateServerModel.TemplateItem
		useBad  bool
		wantErr bool
		errMsg  string
	}{
		{
			name: "success",
			setup: func(t *testing.T, repo *ItemRepository) *cryptoutilAppsSkeletonTemplateServerModel.TemplateItem {
				t.Helper()

				tenantID := googleUuid.Must(googleUuid.NewV7())
				item := &cryptoutilAppsSkeletonTemplateServerModel.TemplateItem{
					ID:       googleUuid.Must(googleUuid.NewV7()),
					TenantID: tenantID,
					Name:     "original",
				}
				require.NoError(t, repo.Create(context.Background(), item))
				item.Name = "updated"

				return item
			},
			wantErr: false,
		},
		{
			// NOTE: With GORM v1.31.1 + SQLite, Save() performs UPDATE then
			// INSERT ON CONFLICT (upsert). RowsAffected is always >=1 when there
			// is no DB error, so the RowsAffected==0 branch in Update is unreachable
			// in SQLite. That path is only exercisable with PostgreSQL in E2E tests.
			name: "db_error",
			setup: func(_ *testing.T, _ *ItemRepository) *cryptoutilAppsSkeletonTemplateServerModel.TemplateItem {
				return &cryptoutilAppsSkeletonTemplateServerModel.TemplateItem{
					ID:       googleUuid.Must(googleUuid.NewV7()),
					TenantID: googleUuid.Must(googleUuid.NewV7()),
					Name:     "fail",
				}
			},
			useBad:  true,
			wantErr: true,
			errMsg:  "failed to update item",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var repo *ItemRepository
			if tc.useBad {
				repo = NewItemRepository(newRepoBrokenDB(t))
			} else {
				repo = NewItemRepository(newRepoTestDB(t))
			}

			item := tc.setup(t, repo)
			err := repo.Update(context.Background(), item)

			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestItemRepository_Delete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(t *testing.T, repo *ItemRepository) (googleUuid.UUID, googleUuid.UUID)
		useBad  bool
		wantErr bool
		errMsg  string
	}{
		{
			name: "success",
			setup: func(t *testing.T, repo *ItemRepository) (googleUuid.UUID, googleUuid.UUID) {
				t.Helper()

				tenantID := googleUuid.Must(googleUuid.NewV7())
				item := &cryptoutilAppsSkeletonTemplateServerModel.TemplateItem{
					ID:       googleUuid.Must(googleUuid.NewV7()),
					TenantID: tenantID,
					Name:     "delete-me",
				}
				require.NoError(t, repo.Create(context.Background(), item))

				return tenantID, item.ID
			},
			wantErr: false,
		},
		{
			name: "not_found",
			setup: func(_ *testing.T, _ *ItemRepository) (googleUuid.UUID, googleUuid.UUID) {
				return googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7())
			},
			wantErr: true,
			errMsg:  "record not found",
		},
		{
			name: "db_error",
			setup: func(_ *testing.T, _ *ItemRepository) (googleUuid.UUID, googleUuid.UUID) {
				return googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7())
			},
			useBad:  true,
			wantErr: true,
			errMsg:  "failed to delete item",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var repo *ItemRepository
			if tc.useBad {
				repo = NewItemRepository(newRepoBrokenDB(t))
			} else {
				repo = NewItemRepository(newRepoTestDB(t))
			}

			tenantID, itemID := tc.setup(t, repo)
			err := repo.Delete(context.Background(), tenantID, itemID)

			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
