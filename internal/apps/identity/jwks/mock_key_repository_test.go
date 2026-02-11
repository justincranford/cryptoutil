// Copyright (c) 2025 Justin Cranford

package jwks

import (
	"context"
	"fmt"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockKeyRepository is a mock implementation of KeyRepository for testing.
type MockKeyRepository struct {
	mock.Mock
}

// FindByUsage mocks the FindByUsage method.
func (m *MockKeyRepository) FindByUsage(ctx context.Context, usage string, active bool) ([]*cryptoutilIdentityDomain.Key, error) {
	args := m.Called(ctx, usage, active)
	if args.Get(0) == nil {
		if err := args.Error(1); err != nil {
			return nil, fmt.Errorf("FindByUsage mock error: %w", err)
		}

		return nil, nil
	}

	keys, ok := args.Get(0).([]*cryptoutilIdentityDomain.Key)
	if !ok {
		return nil, fmt.Errorf("FindByUsage type assertion failed")
	}

	if err := args.Error(1); err != nil {
		return keys, fmt.Errorf("FindByUsage mock error: %w", err)
	}

	return keys, nil
}

// Create mocks the Create method.
func (m *MockKeyRepository) Create(ctx context.Context, key *cryptoutilIdentityDomain.Key) error {
	args := m.Called(ctx, key)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("Create mock error: %w", err)
	}

	return nil
}

// FindByKID mocks the FindByKID method.
func (m *MockKeyRepository) FindByKID(ctx context.Context, kid string) (*cryptoutilIdentityDomain.Key, error) {
	args := m.Called(ctx, kid)
	if args.Get(0) == nil {
		if err := args.Error(1); err != nil {
			return nil, fmt.Errorf("FindByKID mock error: %w", err)
		}

		return nil, nil
	}

	key, ok := args.Get(0).(*cryptoutilIdentityDomain.Key)
	if !ok {
		return nil, fmt.Errorf("FindByKID type assertion failed")
	}

	if err := args.Error(1); err != nil {
		return key, fmt.Errorf("FindByKID mock error: %w", err)
	}

	return key, nil
}

// FindByID mocks the FindByID method.
func (m *MockKeyRepository) FindByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.Key, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		if err := args.Error(1); err != nil {
			return nil, fmt.Errorf("FindByID mock error: %w", err)
		}

		return nil, nil
	}

	key, ok := args.Get(0).(*cryptoutilIdentityDomain.Key)
	if !ok {
		return nil, fmt.Errorf("FindByID type assertion failed")
	}

	if err := args.Error(1); err != nil {
		return key, fmt.Errorf("FindByID mock error: %w", err)
	}

	return key, nil
}

// Update mocks the Update method.
func (m *MockKeyRepository) Update(ctx context.Context, key *cryptoutilIdentityDomain.Key) error {
	args := m.Called(ctx, key)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("Update mock error: %w", err)
	}

	return nil
}

// Delete mocks the Delete method.
func (m *MockKeyRepository) Delete(ctx context.Context, id googleUuid.UUID) error {
	args := m.Called(ctx, id)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("Delete mock error: %w", err)
	}

	return nil
}

// List mocks the List method.
func (m *MockKeyRepository) List(ctx context.Context, limit, offset int) ([]*cryptoutilIdentityDomain.Key, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		if err := args.Error(1); err != nil {
			return nil, fmt.Errorf("List mock error: %w", err)
		}

		return nil, nil
	}

	keys, ok := args.Get(0).([]*cryptoutilIdentityDomain.Key)
	if !ok {
		return nil, fmt.Errorf("List type assertion failed")
	}

	if err := args.Error(1); err != nil {
		return keys, fmt.Errorf("List mock error: %w", err)
	}

	return keys, nil
}

// Count mocks the Count method.
func (m *MockKeyRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)

	count, ok := args.Get(0).(int64)
	if !ok {
		return 0, fmt.Errorf("Count type assertion failed")
	}

	if err := args.Error(1); err != nil {
		return count, fmt.Errorf("Count mock error: %w", err)
	}

	return count, nil
}
