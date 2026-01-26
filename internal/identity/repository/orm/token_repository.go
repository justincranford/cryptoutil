// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	"context"
	"errors"
	"fmt"
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
)

// TokenRepositoryGORM implements the TokenRepository interface using GORM.
type TokenRepositoryGORM struct {
	db *gorm.DB
}

// NewTokenRepository creates a new TokenRepositoryGORM.
func NewTokenRepository(db *gorm.DB) *TokenRepositoryGORM {
	return &TokenRepositoryGORM{db: db}
}

// Create creates a new token.
func (r *TokenRepositoryGORM) Create(ctx context.Context, token *cryptoutilIdentityDomain.Token) error {
	if err := getDB(ctx, r.db).WithContext(ctx).Create(token).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to create token: %w", err))
	}

	return nil
}

// GetByID retrieves a token by ID.
func (r *TokenRepositoryGORM) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.Token, error) {
	var token cryptoutilIdentityDomain.Token
	if err := getDB(ctx, r.db).WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cryptoutilIdentityAppErr.ErrTokenNotFound
		}

		return nil, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to get token by ID: %w", err))
	}

	return &token, nil
}

// GetByTokenValue retrieves a token by its token value.
func (r *TokenRepositoryGORM) GetByTokenValue(ctx context.Context, tokenValue string) (*cryptoutilIdentityDomain.Token, error) {
	var token cryptoutilIdentityDomain.Token
	if err := getDB(ctx, r.db).WithContext(ctx).Where("token_value = ? AND deleted_at IS NULL", tokenValue).First(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cryptoutilIdentityAppErr.ErrTokenNotFound
		}

		return nil, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to get token by value: %w", err))
	}

	return &token, nil
}

// Update updates an existing token.
func (r *TokenRepositoryGORM) Update(ctx context.Context, token *cryptoutilIdentityDomain.Token) error {
	token.UpdatedAt = time.Now().UTC()
	if err := getDB(ctx, r.db).WithContext(ctx).Save(token).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to update token: %w", err))
	}

	return nil
}

// Delete deletes a token by ID (soft delete).
func (r *TokenRepositoryGORM) Delete(ctx context.Context, id googleUuid.UUID) error {
	if err := getDB(ctx, r.db).WithContext(ctx).Where("id = ?", id).Delete(&cryptoutilIdentityDomain.Token{}).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to delete token: %w", err))
	}

	return nil
}

// RevokeByID revokes a token by ID.
func (r *TokenRepositoryGORM) RevokeByID(ctx context.Context, id googleUuid.UUID) error {
	result := getDB(ctx, r.db).WithContext(ctx).Model(&cryptoutilIdentityDomain.Token{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("revoked", true)

	if result.Error != nil {
		return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to revoke token by ID: %w", result.Error))
	}

	if result.RowsAffected == 0 {
		return cryptoutilIdentityAppErr.ErrTokenNotFound
	}

	return nil
}

// RevokeByTokenValue revokes a token by token value.
func (r *TokenRepositoryGORM) RevokeByTokenValue(ctx context.Context, tokenValue string) error {
	now := time.Now().UTC()
	result := getDB(ctx, r.db).WithContext(ctx).Model(&cryptoutilIdentityDomain.Token{}).
		Where("token_value = ? AND deleted_at IS NULL", tokenValue).
		Updates(map[string]any{
			"revoked":    true,
			"revoked_at": now,
		})

	if result.Error != nil {
		return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to revoke token by value: %w", result.Error))
	}

	if result.RowsAffected == 0 {
		return cryptoutilIdentityAppErr.ErrTokenNotFound
	}

	return nil
}

// DeleteExpired deletes expired tokens (hard delete).
func (r *TokenRepositoryGORM) DeleteExpired(ctx context.Context) error {
	if err := getDB(ctx, r.db).WithContext(ctx).Unscoped().
		Where("expires_at < ?", time.Now().UTC()).
		Delete(&cryptoutilIdentityDomain.Token{}).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to delete expired tokens: %w", err))
	}

	return nil
}

// DeleteExpiredBefore deletes all tokens expired before the given time (hard delete).
func (r *TokenRepositoryGORM) DeleteExpiredBefore(ctx context.Context, beforeTime time.Time) (int, error) {
	result := getDB(ctx, r.db).WithContext(ctx).Unscoped().
		Where("expires_at < ?", beforeTime).
		Delete(&cryptoutilIdentityDomain.Token{})

	if result.Error != nil {
		return 0, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to delete expired tokens before %s: %w", beforeTime, result.Error))
	}

	return int(result.RowsAffected), nil
}

// List lists tokens with pagination.
func (r *TokenRepositoryGORM) List(ctx context.Context, offset, limit int) ([]*cryptoutilIdentityDomain.Token, error) {
	var tokens []*cryptoutilIdentityDomain.Token
	if err := getDB(ctx, r.db).WithContext(ctx).Where("deleted_at IS NULL").Offset(offset).Limit(limit).Find(&tokens).Error; err != nil {
		return nil, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to list tokens: %w", err))
	}

	return tokens, nil
}

// Count returns the total number of tokens.
func (r *TokenRepositoryGORM) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := getDB(ctx, r.db).WithContext(ctx).Model(&cryptoutilIdentityDomain.Token{}).Where("deleted_at IS NULL").Count(&count).Error; err != nil {
		return 0, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to count tokens: %w", err))
	}

	return count, nil
}
