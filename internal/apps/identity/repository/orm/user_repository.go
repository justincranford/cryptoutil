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

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

// UserRepositoryGORM implements UserRepository using GORM.
type UserRepositoryGORM struct {
	db *gorm.DB
}

// NewUserRepository creates a new GORM user repository.
func NewUserRepository(db *gorm.DB) *UserRepositoryGORM {
	return &UserRepositoryGORM{db: db}
}

// Create creates a new user.
func (r *UserRepositoryGORM) Create(ctx context.Context, user *cryptoutilIdentityDomain.User) error {
	if err := getDB(ctx, r.db).WithContext(ctx).Create(user).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to create user: %w", err))
	}

	return nil
}

// GetByID retrieves a user by ID.
func (r *UserRepositoryGORM) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.User, error) {
	var user cryptoutilIdentityDomain.User
	if err := getDB(ctx, r.db).WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cryptoutilIdentityAppErr.ErrUserNotFound
		}

		return nil, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to get user by ID: %w", err))
	}

	return &user, nil
}

// GetBySub retrieves a user by subject identifier.
func (r *UserRepositoryGORM) GetBySub(ctx context.Context, sub string) (*cryptoutilIdentityDomain.User, error) {
	var user cryptoutilIdentityDomain.User
	// Enable debug mode to see SQL queries.
	if err := getDB(ctx, r.db).Debug().WithContext(ctx).Where("sub = ? AND deleted_at IS NULL", sub).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cryptoutilIdentityAppErr.ErrUserNotFound
		}

		return nil, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to get user by subject: %w", err))
	}

	return &user, nil
}

// GetByUsername retrieves a user by preferred username.
func (r *UserRepositoryGORM) GetByUsername(ctx context.Context, username string) (*cryptoutilIdentityDomain.User, error) {
	var user cryptoutilIdentityDomain.User
	if err := getDB(ctx, r.db).WithContext(ctx).Where("preferred_username = ? AND deleted_at IS NULL", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cryptoutilIdentityAppErr.ErrUserNotFound
		}

		return nil, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to get user by username: %w", err))
	}

	return &user, nil
}

// GetByEmail retrieves a user by email address.
func (r *UserRepositoryGORM) GetByEmail(ctx context.Context, email string) (*cryptoutilIdentityDomain.User, error) {
	var user cryptoutilIdentityDomain.User
	if err := getDB(ctx, r.db).WithContext(ctx).Where("email = ? AND deleted_at IS NULL", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cryptoutilIdentityAppErr.ErrUserNotFound
		}

		return nil, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to get user by email: %w", err))
	}

	return &user, nil
}

// Update updates an existing user.
func (r *UserRepositoryGORM) Update(ctx context.Context, user *cryptoutilIdentityDomain.User) error {
	user.UpdatedAt = time.Now().UTC()
	if err := getDB(ctx, r.db).WithContext(ctx).Save(user).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to update user: %w", err))
	}

	return nil
}

// Delete deletes a user by ID (soft delete).
func (r *UserRepositoryGORM) Delete(ctx context.Context, id googleUuid.UUID) error {
	if err := getDB(ctx, r.db).WithContext(ctx).Where("id = ?", id).Delete(&cryptoutilIdentityDomain.User{}).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to delete user: %w", err))
	}

	return nil
}

// List lists users with pagination.
func (r *UserRepositoryGORM) List(ctx context.Context, offset, limit int) ([]*cryptoutilIdentityDomain.User, error) {
	var users []*cryptoutilIdentityDomain.User
	if err := getDB(ctx, r.db).WithContext(ctx).Where("deleted_at IS NULL").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to list users: %w", err))
	}

	return users, nil
}

// Count returns the total number of users.
func (r *UserRepositoryGORM) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := getDB(ctx, r.db).WithContext(ctx).Model(&cryptoutilIdentityDomain.User{}).Where("deleted_at IS NULL").Count(&count).Error; err != nil {
		return 0, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to count users: %w", err))
	}

	return count, nil
}
