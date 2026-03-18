// Copyright (c) 2025 Justin Cranford
//
// TEMPLATE: Copy and rename 'skeleton' -> your-service-name before use.

// Package handler provides unit tests for the skeleton-template OpenAPI strict server.
package handler

import (
	"context"
	"database/sql"
	"testing"

	cryptoutilSkeletonTemplateServer "cryptoutil/api/skeleton-template/server"
	cryptoutilAppsSkeletonTemplateDomain "cryptoutil/internal/apps/skeleton/template/domain"
	cryptoutilAppsSkeletonTemplateRepository "cryptoutil/internal/apps/skeleton/template/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // CGO-free SQLite driver.
)

// newHandlerTestDB creates a per-test in-memory SQLite DB with migrations applied.
// Uses MaxOpenConns=1 so all queries share the same connection (required for
// in-memory SQLite with cache=private mode).
func newHandlerTestDB(t *testing.T) *gorm.DB {
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

	// MaxOpenConns=1 forces all GORM operations to share a single connection,
	// which is required for in-memory SQLite with cache=private so that all
	// queries see the same in-memory database (the one where AutoMigrate ran).
	rawDB.SetMaxOpenConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	rawDB.SetMaxIdleConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	rawDB.SetConnMaxLifetime(0)

	db, err := gorm.Open(sqlite.Dialector{Conn: rawDB}, &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	require.NoError(t, db.AutoMigrate(&cryptoutilAppsSkeletonTemplateDomain.TemplateItem{}))

	t.Cleanup(func() { _ = rawDB.Close() })

	return db
}

// newHandlerNoTableDB creates a DB with no migrations — queries will fail with
// "no such table: template_items", triggering 500 error paths in the handler.
func newHandlerNoTableDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbID, err := googleUuid.NewV7()
	require.NoError(t, err)

	dsn := "file:" + dbID.String() + "?mode=memory&cache=private"

	rawDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dsn)
	require.NoError(t, err)

	t.Cleanup(func() { _ = rawDB.Close() })

	rawDB.SetMaxOpenConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	rawDB.SetMaxIdleConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	rawDB.SetConnMaxLifetime(0)

	// No AutoMigrate — template_items table does not exist.
	db, err := gorm.Open(sqlite.Dialector{Conn: rawDB}, &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	return db
}

func newHandlerTestServer(t *testing.T) *StrictServer {
	t.Helper()

	return NewStrictServer(cryptoutilAppsSkeletonTemplateRepository.NewItemRepository(newHandlerTestDB(t)))
}

func newHandlerErrorServer(t *testing.T) *StrictServer {
	t.Helper()

	return NewStrictServer(cryptoutilAppsSkeletonTemplateRepository.NewItemRepository(newHandlerNoTableDB(t)))
}

func handlerStrPtr(s string) *string { return &s }
func handlerIntPtr(n int) *int       { return &n }

func TestNewStrictServer(t *testing.T) {
	t.Parallel()

	require.NotNil(t, newHandlerTestServer(t))
}

func TestStrictServer_CreateItem(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		body     *cryptoutilSkeletonTemplateServer.ItemCreate
		wantCode int
	}{
		{
			name:     "success",
			body:     &cryptoutilSkeletonTemplateServer.ItemCreate{Name: "Test Item", Description: handlerStrPtr("desc")},
			wantCode: 201,
		},
		{
			name:     "nil_body",
			body:     nil,
			wantCode: 400,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			resp, err := newHandlerTestServer(t).CreateItem(context.Background(),
				cryptoutilSkeletonTemplateServer.CreateItemRequestObject{Body: tc.body},
			)
			require.NoError(t, err)

			switch resp.(type) {
			case cryptoutilSkeletonTemplateServer.CreateItem201JSONResponse:
				require.Equal(t, 201, tc.wantCode)
			case cryptoutilSkeletonTemplateServer.CreateItem400JSONResponse:
				require.Equal(t, 400, tc.wantCode)
			default:
				require.Failf(t, "unexpected response type", "%T", resp)
			}
		})
	}
}

func TestStrictServer_CreateItem_DBError(t *testing.T) {
	t.Parallel()

	resp, err := newHandlerErrorServer(t).CreateItem(context.Background(),
		cryptoutilSkeletonTemplateServer.CreateItemRequestObject{
			Body: &cryptoutilSkeletonTemplateServer.ItemCreate{Name: "item"},
		},
	)
	require.NoError(t, err)
	require.IsType(t, cryptoutilSkeletonTemplateServer.CreateItem500JSONResponse{}, resp)
}

func TestStrictServer_ListItems(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		page *int
		size *int
	}{
		{name: "default_params"},
		{name: "with_pagination", page: handlerIntPtr(1), size: handlerIntPtr(cryptoutilSharedMagic.SuiteServiceCount)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			resp, err := newHandlerTestServer(t).ListItems(context.Background(),
				cryptoutilSkeletonTemplateServer.ListItemsRequestObject{
					Params: cryptoutilSkeletonTemplateServer.ListItemsParams{Page: tc.page, Size: tc.size},
				},
			)
			require.NoError(t, err)

			r, ok := resp.(cryptoutilSkeletonTemplateServer.ListItems200JSONResponse)
			require.True(t, ok, "expected ListItems200JSONResponse, got %T", resp)
			require.NotNil(t, r.Items)
		})
	}
}

func TestStrictServer_ListItems_DBError(t *testing.T) {
	t.Parallel()

	resp, err := newHandlerErrorServer(t).ListItems(context.Background(),
		cryptoutilSkeletonTemplateServer.ListItemsRequestObject{
			Params: cryptoutilSkeletonTemplateServer.ListItemsParams{},
		},
	)
	require.NoError(t, err)
	require.IsType(t, cryptoutilSkeletonTemplateServer.ListItems500JSONResponse{}, resp)
}

func TestStrictServer_GetItem_Success(t *testing.T) {
	t.Parallel()

	srv := newHandlerTestServer(t)
	ctx := context.Background()

	createResp, err := srv.CreateItem(ctx, cryptoutilSkeletonTemplateServer.CreateItemRequestObject{
		Body: &cryptoutilSkeletonTemplateServer.ItemCreate{Name: "Get Success"},
	})
	require.NoError(t, err)

	created, ok := createResp.(cryptoutilSkeletonTemplateServer.CreateItem201JSONResponse)
	require.True(t, ok)

	resp, err := srv.GetItem(ctx, cryptoutilSkeletonTemplateServer.GetItemRequestObject{ItemID: created.ID})
	require.NoError(t, err)

	r, ok := resp.(cryptoutilSkeletonTemplateServer.GetItem200JSONResponse)
	require.True(t, ok, "expected GetItem200JSONResponse, got %T", resp)
	require.Equal(t, created.ID, r.ID)
}

func TestStrictServer_GetItem_NotFound(t *testing.T) {
	t.Parallel()

	resp, err := newHandlerTestServer(t).GetItem(context.Background(),
		cryptoutilSkeletonTemplateServer.GetItemRequestObject{ItemID: googleUuid.Must(googleUuid.NewV7())},
	)
	require.NoError(t, err)
	require.IsType(t, cryptoutilSkeletonTemplateServer.GetItem404JSONResponse{}, resp)
}

func TestStrictServer_GetItem_DBError(t *testing.T) {
	t.Parallel()

	resp, err := newHandlerErrorServer(t).GetItem(context.Background(),
		cryptoutilSkeletonTemplateServer.GetItemRequestObject{ItemID: googleUuid.Must(googleUuid.NewV7())},
	)
	require.NoError(t, err)
	require.IsType(t, cryptoutilSkeletonTemplateServer.GetItem500JSONResponse{}, resp)
}

func TestStrictServer_UpdateItem_Success(t *testing.T) {
	t.Parallel()

	srv := newHandlerTestServer(t)
	ctx := context.Background()

	createResp, err := srv.CreateItem(ctx, cryptoutilSkeletonTemplateServer.CreateItemRequestObject{
		Body: &cryptoutilSkeletonTemplateServer.ItemCreate{Name: "Update Test"},
	})
	require.NoError(t, err)

	created, ok := createResp.(cryptoutilSkeletonTemplateServer.CreateItem201JSONResponse)
	require.True(t, ok)

	resp, err := srv.UpdateItem(ctx, cryptoutilSkeletonTemplateServer.UpdateItemRequestObject{
		ItemID: created.ID,
		Body:   &cryptoutilSkeletonTemplateServer.ItemUpdate{Name: "Updated", Description: handlerStrPtr("new desc")},
	})
	require.NoError(t, err)

	r, ok := resp.(cryptoutilSkeletonTemplateServer.UpdateItem200JSONResponse)
	require.True(t, ok, "expected UpdateItem200JSONResponse, got %T", resp)
	require.Equal(t, "Updated", r.Name)
}

func TestStrictServer_UpdateItem_NilBody(t *testing.T) {
	t.Parallel()

	srv := newHandlerTestServer(t)
	ctx := context.Background()

	createResp, err := srv.CreateItem(ctx, cryptoutilSkeletonTemplateServer.CreateItemRequestObject{
		Body: &cryptoutilSkeletonTemplateServer.ItemCreate{Name: "Nil Body Target"},
	})
	require.NoError(t, err)

	created, ok := createResp.(cryptoutilSkeletonTemplateServer.CreateItem201JSONResponse)
	require.True(t, ok)

	resp, err := srv.UpdateItem(ctx, cryptoutilSkeletonTemplateServer.UpdateItemRequestObject{
		ItemID: created.ID,
		Body:   nil,
	})
	require.NoError(t, err)
	require.IsType(t, cryptoutilSkeletonTemplateServer.UpdateItem400JSONResponse{}, resp)
}

func TestStrictServer_UpdateItem_NotFound(t *testing.T) {
	t.Parallel()

	resp, err := newHandlerTestServer(t).UpdateItem(context.Background(),
		cryptoutilSkeletonTemplateServer.UpdateItemRequestObject{
			ItemID: googleUuid.Must(googleUuid.NewV7()),
			Body:   &cryptoutilSkeletonTemplateServer.ItemUpdate{Name: "X"},
		},
	)
	require.NoError(t, err)
	require.IsType(t, cryptoutilSkeletonTemplateServer.UpdateItem404JSONResponse{}, resp)
}

func TestStrictServer_UpdateItem_DBError(t *testing.T) {
	t.Parallel()

	resp, err := newHandlerErrorServer(t).UpdateItem(context.Background(),
		cryptoutilSkeletonTemplateServer.UpdateItemRequestObject{
			ItemID: googleUuid.Must(googleUuid.NewV7()),
			Body:   &cryptoutilSkeletonTemplateServer.ItemUpdate{Name: "X"},
		},
	)
	require.NoError(t, err)
	require.IsType(t, cryptoutilSkeletonTemplateServer.UpdateItem500JSONResponse{}, resp)
}

func TestStrictServer_DeleteItem_Success(t *testing.T) {
	t.Parallel()

	srv := newHandlerTestServer(t)
	ctx := context.Background()

	createResp, err := srv.CreateItem(ctx, cryptoutilSkeletonTemplateServer.CreateItemRequestObject{
		Body: &cryptoutilSkeletonTemplateServer.ItemCreate{Name: "Delete Target"},
	})
	require.NoError(t, err)

	created, ok := createResp.(cryptoutilSkeletonTemplateServer.CreateItem201JSONResponse)
	require.True(t, ok)

	resp, err := srv.DeleteItem(ctx, cryptoutilSkeletonTemplateServer.DeleteItemRequestObject{ItemID: created.ID})
	require.NoError(t, err)
	require.IsType(t, cryptoutilSkeletonTemplateServer.DeleteItem204Response{}, resp)
}

func TestStrictServer_DeleteItem_NotFound(t *testing.T) {
	t.Parallel()

	resp, err := newHandlerTestServer(t).DeleteItem(context.Background(),
		cryptoutilSkeletonTemplateServer.DeleteItemRequestObject{ItemID: googleUuid.Must(googleUuid.NewV7())},
	)
	require.NoError(t, err)
	require.IsType(t, cryptoutilSkeletonTemplateServer.DeleteItem404JSONResponse{}, resp)
}

func TestStrictServer_DeleteItem_DBError(t *testing.T) {
	t.Parallel()

	resp, err := newHandlerErrorServer(t).DeleteItem(context.Background(),
		cryptoutilSkeletonTemplateServer.DeleteItemRequestObject{ItemID: googleUuid.Must(googleUuid.NewV7())},
	)
	require.NoError(t, err)
	require.IsType(t, cryptoutilSkeletonTemplateServer.DeleteItem500JSONResponse{}, resp)
}

func TestDerefString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input *string
		want  string
	}{
		{name: "nil_pointer", input: nil, want: ""},
		{name: "empty_string", input: handlerStrPtr(""), want: ""},
		{name: "non_empty", input: handlerStrPtr("hello"), want: "hello"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, derefString(tc.input))
		})
	}
}
