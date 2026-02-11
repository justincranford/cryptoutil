// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	"context"

	"gorm.io/gorm"
)

// contextKey is the type for context keys to avoid collisions.
type contextKey string

// txKey is the context key for storing the transaction DB.
const txKey contextKey = "gorm_tx"

// getDB returns the transaction DB from context if present, otherwise returns the base DB.
// This function should be used by all repository implementations to ensure transaction awareness.
func getDB(ctx context.Context, baseDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey).(*gorm.DB); ok {
		return tx
	}

	return baseDB
}
