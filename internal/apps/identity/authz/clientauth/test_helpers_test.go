// Copyright (c) 2025 Justin Cranford

package clientauth

import (
	"context"

	googleUuid "github.com/google/uuid"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

// boolPtr converts bool to *bool for struct literals requiring pointer fields.
func boolPtr(b bool) *bool {
	return &b
}

// mockClientRepo is a shared mock implementation used across multiple test files.
// It implements cryptoutilIdentityRepository.ClientRepository interface.
type mockClientRepo struct {
	clients map[string]*cryptoutilIdentityDomain.Client
}

func (m *mockClientRepo) Create(_ context.Context, _ *cryptoutilIdentityDomain.Client) error {
	return nil
}

func (m *mockClientRepo) GetByID(_ context.Context, _ googleUuid.UUID) (*cryptoutilIdentityDomain.Client, error) {
	return nil, nil
}

func (m *mockClientRepo) GetByClientID(_ context.Context, clientID string) (*cryptoutilIdentityDomain.Client, error) {
	if m.clients != nil {
		if client, ok := m.clients[clientID]; ok {
			return client, nil
		}
	}

	return nil, nil
}

func (m *mockClientRepo) Update(_ context.Context, _ *cryptoutilIdentityDomain.Client) error {
	return nil
}

func (m *mockClientRepo) Delete(_ context.Context, _ googleUuid.UUID) error {
	return nil
}

func (m *mockClientRepo) List(_ context.Context, _ int, _ int) ([]*cryptoutilIdentityDomain.Client, error) {
	return nil, nil
}

func (m *mockClientRepo) Count(_ context.Context) (int64, error) {
	return 0, nil
}

func (m *mockClientRepo) RotateSecret(_ context.Context, _ googleUuid.UUID, _ string, _ string, _ string) error {
	return nil
}

func (m *mockClientRepo) GetSecretHistory(_ context.Context, _ googleUuid.UUID) ([]cryptoutilIdentityDomain.ClientSecretHistory, error) {
	return nil, nil
}

func (m *mockClientRepo) GetAll(_ context.Context) ([]*cryptoutilIdentityDomain.Client, error) {
	return nil, nil
}
