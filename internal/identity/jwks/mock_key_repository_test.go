// Copyright (c) 2025 Justin Cranford

package jwks

import (
	"context"

	identityDomain "cryptoutil/internal/identity/domain"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockKeyRepository is a mock implementation of KeyRepository for testing.
type MockKeyRepository struct {
	mock.Mock
}

// FindByUsage mocks the FindByUsage method.
func (m *MockKeyRepository) FindByUsage(ctx context.Context, usage string, active bool) ([]*identityDomain.Key, error) {
	args := m.Called(ctx, usage, active)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*identityDomain.Key), args.Error(1)
}

// Create mocks the Create method.
func (m *MockKeyRepository) Create(ctx context.Context, key *identityDomain.Key) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

// GetByID mocks the GetByID method.
func (m *MockKeyRepository) GetByID(ctx context.Context, id googleUuid.UUID) (*identityDomain.Key, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*identityDomain.Key), args.Error(1)
}

// Update mocks the Update method.
func (m *MockKeyRepository) Update(ctx context.Context, key *identityDomain.Key) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

// Delete mocks the Delete method.
func (m *MockKeyRepository) Delete(ctx context.Context, id googleUuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// List mocks the List method.
func (m *MockKeyRepository) List(ctx context.Context, offset, limit int) ([]*identityDomain.Key, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*identityDomain.Key), args.Error(1)
}

// Count mocks the Count method.
func (m *MockKeyRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}
