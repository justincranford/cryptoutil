// Copyright (c) 2025 Justin Cranford
//
//

// Package repository provides reusable GORM transaction patterns for all cryptoutil services.
package repository

import (
	"context"

	"gorm.io/gorm"
)

// txKey is the context key for database transactions.
//
// This private type prevents collisions with context keys from other packages.
type txKey struct{}

// WithTransaction stores a GORM transaction in the context for repository transparency.
//
// Usage pattern (service layer):
//
//	tx := db.Begin()
//	txCtx := repository.WithTransaction(ctx, tx)
//	if err := userRepo.Create(txCtx, user); err != nil {
//	    tx.Rollback()
//	    return err
//	}
//	if err := profileRepo.Create(txCtx, profile); err != nil {
//	    tx.Rollback()
//	    return err
//	}
//	return tx.Commit().Error
//
// Repositories automatically use the transaction from context when present.
func WithTransaction(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// GetDB returns the transaction from context if present, otherwise returns the base DB.
//
// MANDATORY: All repository methods MUST use this pattern for transaction transparency:
//
//	func (r *Repository) Create(ctx context.Context, entity *Entity) error {
//	    return GetDB(ctx, r.db).WithContext(ctx).Create(entity).Error
//	}
//
// This enables:
// - Automatic transaction participation when ctx contains transaction
// - Normal operation when ctx has no transaction
// - Zero changes needed when adding/removing transactions in service layer
//
// Pattern Reference: See internal/sm/im/repository/user_repository.go for example usage.
func GetDB(ctx context.Context, baseDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok && tx != nil {
		return tx
	}

	return baseDB
}
