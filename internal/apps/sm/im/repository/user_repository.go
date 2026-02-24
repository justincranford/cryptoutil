// Copyright (c) 2025 Justin Cranford
//
//

// Package repository provides database access for sm-im domain models.
package repository

import (
	"context"
	"fmt"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

// UserRepository handles database operations for User entities.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user into the database.
func (r *UserRepository) Create(ctx context.Context, user *cryptoutilAppsTemplateServiceServerRepository.User) error {
	if err := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db).WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// FindByID retrieves a user by ID.
func (r *UserRepository) FindByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.User, error) {
	var user cryptoutilAppsTemplateServiceServerRepository.User
	if err := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db).WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return &user, nil
}

// FindByUsername retrieves a user by username.
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*cryptoutilAppsTemplateServiceServerRepository.User, error) {
	var user cryptoutilAppsTemplateServiceServerRepository.User
	if err := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db).WithContext(ctx).First(&user, "username = ?", username).Error; err != nil {
		return nil, fmt.Errorf("failed to find user by username: %w", err)
	}

	return &user, nil
}

// Update updates an existing user.
func (r *UserRepository) Update(ctx context.Context, user *cryptoutilAppsTemplateServiceServerRepository.User) error {
	if err := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db).WithContext(ctx).Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// Delete removes a user from the database.
func (r *UserRepository) Delete(ctx context.Context, id googleUuid.UUID) error {
	if err := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db).WithContext(ctx).Delete(&cryptoutilAppsTemplateServiceServerRepository.User{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}
