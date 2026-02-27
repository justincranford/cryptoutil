// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"

	googleUuid "github.com/google/uuid"

	cryptoutilAppsTemplateServiceServerRealms "cryptoutil/internal/apps/template/service/server/realms"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

// UserRepositoryAdapter adapts UserRepository to realms.UserRepository interface.
// This allows sm-im to use the template realms service while keeping
// the existing UserRepository implementation unchanged.
type UserRepositoryAdapter struct {
	repo *UserRepository
}

// NewUserRepositoryAdapter creates a new UserRepositoryAdapter.
func NewUserRepositoryAdapter(repo *UserRepository) *UserRepositoryAdapter {
	return &UserRepositoryAdapter{repo: repo}
}

// Create creates a new user in the database.
// Adapts realms.UserModel interface to concrete template repository.User.
func (a *UserRepositoryAdapter) Create(ctx context.Context, user cryptoutilAppsTemplateServiceServerRealms.UserModel) error {
	// Type assertion: UserModel -> *repository.User
	concreteUser, ok := user.(*cryptoutilAppsTemplateServiceServerRepository.User)
	if !ok {
		// This should never happen if used correctly
		panic("UserRepositoryAdapter.Create: expected *repository.User")
	}

	return a.repo.Create(ctx, concreteUser)
}

// FindByUsername finds a user by username.
// Adapts realms.UserModel interface to concrete template repository.User.
func (a *UserRepositoryAdapter) FindByUsername(ctx context.Context, username string) (cryptoutilAppsTemplateServiceServerRealms.UserModel, error) {
	user, err := a.repo.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	return user, nil // *repository.User implements UserModel
}

// FindByID finds a user by ID.
// Adapts realms.UserModel interface to concrete template repository.User.
func (a *UserRepositoryAdapter) FindByID(ctx context.Context, id googleUuid.UUID) (cryptoutilAppsTemplateServiceServerRealms.UserModel, error) {
	user, err := a.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return user, nil // *repository.User implements UserModel
}

// Compile-time check that UserRepositoryAdapter implements realms.UserRepository interface.
var _ cryptoutilAppsTemplateServiceServerRealms.UserRepository = (*UserRepositoryAdapter)(nil)
