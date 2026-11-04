// Copyright (c) 2025-2026 Justin Cranford.
//
// TEMPLATE: Copy and rename 'skeleton' -> your-service-name before use.

// Package apis provides unit tests for the skeleton-template OpenAPI strict server.
package apis

import (
	"context"
	"testing"

	cryptoutilSkeletonTemplateServer "cryptoutil/api/skeleton-template/server"
	cryptoutilTestHelpDB "cryptoutil/internal/apps-framework/service/test_help_db"
	cryptoutilAppsSkeletonTemplateServerModel "cryptoutil/internal/apps/skeleton-template/server/model"
	cryptoutilAppsSkeletonTemplateServerRepository "cryptoutil/internal/apps/skeleton-template/server/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// newHandlerTestDB creates a uniquely named in-memory SQLite DB with migrations applied.
func newHandlerTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db := cryptoutilTestHelpDB.NewInMemorySQLiteDB(t)
	require.NoError(t, db.AutoMigrate(&cryptoutilAppsSkeletonTemplateServerModel.TemplateItem{}))

	return db
}

// newHandlerNoTableDB creates a DB with no migrations — queries will fail with
// "no such table: template_items", triggering 500 error paths in the handler.
func newHandlerNoTableDB(t *testing.T) *gorm.DB {
	t.Helper()

	// No AutoMigrate — template_items table does not exist.
	return cryptoutilTestHelpDB.NewInMemorySQLiteDB(t)
}

func newHandlerTestServer(t *testing.T) *StrictServer {
	t.Helper()

	return NewStrictServer(cryptoutilAppsSkeletonTemplateServerRepository.NewItemRepository(newHandlerTestDB(t)))
}

func newHandlerErrorServer(t *testing.T) *StrictServer {
	t.Helper()

	return NewStrictServer(cryptoutilAppsSkeletonTemplateServerRepository.NewItemRepository(newHandlerNoTableDB(t)))
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

			resp, err := newHandlerTestServer(t).CreateItem(
				context.Background(),
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

	resp, err := newHandlerErrorServer(t).CreateItem(
		context.Background(),
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

			resp, err := newHandlerTestServer(t).ListItems(
				context.Background(),
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

	resp, err := newHandlerErrorServer(t).ListItems(
		context.Background(),
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

	resp, err := newHandlerTestServer(t).GetItem(
		context.Background(),
		cryptoutilSkeletonTemplateServer.GetItemRequestObject{ItemID: googleUuid.Must(googleUuid.NewV7())},
	)
	require.NoError(t, err)
	require.IsType(t, cryptoutilSkeletonTemplateServer.GetItem404JSONResponse{}, resp)
}

func TestStrictServer_GetItem_DBError(t *testing.T) {
	t.Parallel()

	resp, err := newHandlerErrorServer(t).GetItem(
		context.Background(),
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

	resp, err := newHandlerTestServer(t).UpdateItem(
		context.Background(),
		cryptoutilSkeletonTemplateServer.UpdateItemRequestObject{
			ItemID: googleUuid.Must(googleUuid.NewV7()),
			Body:   &cryptoutilSkeletonTemplateServer.ItemUpdate{Name: testX},
		},
	)
	require.NoError(t, err)
	require.IsType(t, cryptoutilSkeletonTemplateServer.UpdateItem404JSONResponse{}, resp)
}

func TestStrictServer_UpdateItem_DBError(t *testing.T) {
	t.Parallel()

	resp, err := newHandlerErrorServer(t).UpdateItem(
		context.Background(),
		cryptoutilSkeletonTemplateServer.UpdateItemRequestObject{
			ItemID: googleUuid.Must(googleUuid.NewV7()),
			Body:   &cryptoutilSkeletonTemplateServer.ItemUpdate{Name: testX},
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

	resp, err := newHandlerTestServer(t).DeleteItem(
		context.Background(),
		cryptoutilSkeletonTemplateServer.DeleteItemRequestObject{ItemID: googleUuid.Must(googleUuid.NewV7())},
	)
	require.NoError(t, err)
	require.IsType(t, cryptoutilSkeletonTemplateServer.DeleteItem404JSONResponse{}, resp)
}

func TestStrictServer_DeleteItem_DBError(t *testing.T) {
	t.Parallel()

	resp, err := newHandlerErrorServer(t).DeleteItem(
		context.Background(),
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
		{name: "non_empty", input: handlerStrPtr(cryptoutilSharedMagic.TEST_NEEDLE_HELLO), want: cryptoutilSharedMagic.TEST_NEEDLE_HELLO},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, derefString(tc.input))
		})
	}
}
