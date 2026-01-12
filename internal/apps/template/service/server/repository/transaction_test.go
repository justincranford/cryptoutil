// Copyright (c) 2025 Justin Cranford
//
//

package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"cryptoutil/internal/apps/template/service/server/repository"
)

// TestWithTransaction tests transaction context storage.
func TestWithTransaction(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockTx := &gorm.DB{}

	txCtx := repository.WithTransaction(ctx, mockTx)

	require.NotNil(t, txCtx)
	require.NotEqual(t, ctx, txCtx)
}

// TestGetDB_WithTransaction tests GetDB with transaction in context.
func TestGetDB_WithTransaction(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	baseDB := &gorm.DB{}
	mockTx := &gorm.DB{Config: &gorm.Config{}}

	txCtx := repository.WithTransaction(ctx, mockTx)
	db := repository.GetDB(txCtx, baseDB)

	require.Same(t, mockTx, db)
}

// TestGetDB_WithoutTransaction tests GetDB without transaction in context.
func TestGetDB_WithoutTransaction(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	baseDB := &gorm.DB{}

	db := repository.GetDB(ctx, baseDB)

	require.Equal(t, baseDB, db)
}

// TestGetDB_NilTransaction tests GetDB with nil transaction in context.
func TestGetDB_NilTransaction(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	baseDB := &gorm.DB{}

	txCtx := repository.WithTransaction(ctx, nil)
	db := repository.GetDB(txCtx, baseDB)

	require.Equal(t, baseDB, db)
}
