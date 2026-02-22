// Copyright (c) 2025 Justin Cranford
//
//

package bootstrap_test

import (
	"context"
	"errors"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityBootstrap "cryptoutil/internal/apps/identity/bootstrap"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

// --- mock implementations ---

// mockClientRepo implements repository.ClientRepository for testing.
type mockClientRepo struct {
	getByClientIDFn func(ctx context.Context, clientID string) (*cryptoutilIdentityDomain.Client, error)
	createFn        func(ctx context.Context, client *cryptoutilIdentityDomain.Client) error
	deleteFn        func(ctx context.Context, id googleUuid.UUID) error
}

func (m *mockClientRepo) GetByClientID(ctx context.Context, clientID string) (*cryptoutilIdentityDomain.Client, error) {
	if m.getByClientIDFn != nil {
		return m.getByClientIDFn(ctx, clientID)
	}

	return nil, cryptoutilIdentityAppErr.ErrClientNotFound
}

func (m *mockClientRepo) Create(ctx context.Context, client *cryptoutilIdentityDomain.Client) error {
	if m.createFn != nil {
		return m.createFn(ctx, client)
	}

	return nil
}

func (m *mockClientRepo) Delete(ctx context.Context, id googleUuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}

	return nil
}

func (m *mockClientRepo) GetByID(_ context.Context, _ googleUuid.UUID) (*cryptoutilIdentityDomain.Client, error) {
	return nil, nil
}

func (m *mockClientRepo) GetAll(_ context.Context) ([]*cryptoutilIdentityDomain.Client, error) {
	return nil, nil
}

func (m *mockClientRepo) Update(_ context.Context, _ *cryptoutilIdentityDomain.Client) error {
	return nil
}

func (m *mockClientRepo) List(_ context.Context, _, _ int) ([]*cryptoutilIdentityDomain.Client, error) {
	return nil, nil
}
func (m *mockClientRepo) Count(_ context.Context) (int64, error) { return 0, nil }
func (m *mockClientRepo) RotateSecret(_ context.Context, _ googleUuid.UUID, _, _, _ string) error {
	return nil
}

func (m *mockClientRepo) GetSecretHistory(_ context.Context, _ googleUuid.UUID) ([]cryptoutilIdentityDomain.ClientSecretHistory, error) {
	return nil, nil
}

// mockUserRepo implements repository.UserRepository for testing.
type mockUserRepo struct {
	getBySubFn func(ctx context.Context, sub string) (*cryptoutilIdentityDomain.User, error)
	createFn   func(ctx context.Context, user *cryptoutilIdentityDomain.User) error
	deleteFn   func(ctx context.Context, id googleUuid.UUID) error
}

func (m *mockUserRepo) GetBySub(ctx context.Context, sub string) (*cryptoutilIdentityDomain.User, error) {
	if m.getBySubFn != nil {
		return m.getBySubFn(ctx, sub)
	}

	return nil, cryptoutilIdentityAppErr.ErrUserNotFound
}

func (m *mockUserRepo) Create(ctx context.Context, user *cryptoutilIdentityDomain.User) error {
	if m.createFn != nil {
		return m.createFn(ctx, user)
	}

	return nil
}

func (m *mockUserRepo) Delete(ctx context.Context, id googleUuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}

	return nil
}

func (m *mockUserRepo) GetByID(_ context.Context, _ googleUuid.UUID) (*cryptoutilIdentityDomain.User, error) {
	return nil, nil
}

func (m *mockUserRepo) GetByUsername(_ context.Context, _ string) (*cryptoutilIdentityDomain.User, error) {
	return nil, nil
}

func (m *mockUserRepo) GetByEmail(_ context.Context, _ string) (*cryptoutilIdentityDomain.User, error) {
	return nil, nil
}
func (m *mockUserRepo) Update(_ context.Context, _ *cryptoutilIdentityDomain.User) error { return nil }
func (m *mockUserRepo) List(_ context.Context, _, _ int) ([]*cryptoutilIdentityDomain.User, error) {
	return nil, nil
}
func (m *mockUserRepo) Count(_ context.Context) (int64, error) { return 0, nil }

// --- helper ---

func makeFactory(clientRepo cryptoutilIdentityRepository.ClientRepository, userRepo cryptoutilIdentityRepository.UserRepository) *cryptoutilIdentityRepository.RepositoryFactory {
	return cryptoutilIdentityRepository.NewRepositoryFactoryForTesting(userRepo, clientRepo)
}

// --- CreateDemoClient error path tests ---

func TestCreateDemoClient_GetByClientIDError(t *testing.T) {
	t.Parallel()

	dbErr := errors.New("db connection error")
	client := &mockClientRepo{
		getByClientIDFn: func(_ context.Context, _ string) (*cryptoutilIdentityDomain.Client, error) {
			return nil, dbErr
		},
	}
	factory := makeFactory(client, &mockUserRepo{})
	_, _, _, err := cryptoutilIdentityBootstrap.CreateDemoClient(context.Background(), factory)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to check for existing demo-client")
}

func TestCreateDemoClient_CreateError(t *testing.T) {
	t.Parallel()

	createErr := errors.New("insert failed")
	client := &mockClientRepo{
		getByClientIDFn: func(_ context.Context, _ string) (*cryptoutilIdentityDomain.Client, error) {
			return nil, cryptoutilIdentityAppErr.ErrClientNotFound
		},
		createFn: func(_ context.Context, _ *cryptoutilIdentityDomain.Client) error {
			return createErr
		},
	}
	factory := makeFactory(client, &mockUserRepo{})
	_, _, _, err := cryptoutilIdentityBootstrap.CreateDemoClient(context.Background(), factory)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create demo-client")
}

// --- CreateDemoUser error path tests ---

func TestCreateDemoUser_GetBySubError(t *testing.T) {
	t.Parallel()

	dbErr := errors.New("db error")
	user := &mockUserRepo{
		getBySubFn: func(_ context.Context, _ string) (*cryptoutilIdentityDomain.User, error) {
			return nil, dbErr
		},
	}
	factory := makeFactory(&mockClientRepo{}, user)
	_, _, _, err := cryptoutilIdentityBootstrap.CreateDemoUser(context.Background(), factory)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to check for existing demo user")
}

func TestCreateDemoUser_CreateError(t *testing.T) {
	t.Parallel()

	createErr := errors.New("insert failed")
	user := &mockUserRepo{
		getBySubFn: func(_ context.Context, _ string) (*cryptoutilIdentityDomain.User, error) {
			return nil, cryptoutilIdentityAppErr.ErrUserNotFound
		},
		createFn: func(_ context.Context, _ *cryptoutilIdentityDomain.User) error {
			return createErr
		},
	}
	factory := makeFactory(&mockClientRepo{}, user)
	_, _, _, err := cryptoutilIdentityBootstrap.CreateDemoUser(context.Background(), factory)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create demo user")
}

// --- deleteDemoClient / deleteDemoUser error path tests ---

func TestDeleteDemoClient_DeleteError(t *testing.T) {
	t.Parallel()

	existingClient := &cryptoutilIdentityDomain.Client{
		ID:       googleUuid.New(),
		ClientID: "demo-client",
	}
	deleteErr := errors.New("delete failed")
	client := &mockClientRepo{
		getByClientIDFn: func(_ context.Context, _ string) (*cryptoutilIdentityDomain.Client, error) {
			return existingClient, nil
		},
		deleteFn: func(_ context.Context, _ googleUuid.UUID) error {
			return deleteErr
		},
	}
	factory := makeFactory(client, &mockUserRepo{})
	err := cryptoutilIdentityBootstrap.ResetDemoData(context.Background(), factory)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to delete demo client")
}

func TestDeleteDemoUser_DeleteError(t *testing.T) {
	t.Parallel()

	existingUser := &cryptoutilIdentityDomain.User{
		ID:  googleUuid.New(),
		Sub: "demo-user",
	}
	deleteErr := errors.New("delete failed")
	// clientRepo returns ErrClientNotFound (client already deleted) so deleteDemoClient succeeds
	client := &mockClientRepo{
		getByClientIDFn: func(_ context.Context, _ string) (*cryptoutilIdentityDomain.Client, error) {
			return nil, cryptoutilIdentityAppErr.ErrClientNotFound
		},
	}
	user := &mockUserRepo{
		getBySubFn: func(_ context.Context, _ string) (*cryptoutilIdentityDomain.User, error) {
			return existingUser, nil
		},
		deleteFn: func(_ context.Context, _ googleUuid.UUID) error {
			return deleteErr
		},
	}
	factory := makeFactory(client, user)
	err := cryptoutilIdentityBootstrap.ResetDemoData(context.Background(), factory)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to delete demo user")
}

// --- ResetDemoData error propagation ---

func TestResetDemoData_DeleteClientCheckError(t *testing.T) {
	t.Parallel()

	dbErr := errors.New("db unreachable")
	client := &mockClientRepo{
		getByClientIDFn: func(_ context.Context, _ string) (*cryptoutilIdentityDomain.Client, error) {
			return nil, dbErr
		},
	}
	factory := makeFactory(client, &mockUserRepo{})
	err := cryptoutilIdentityBootstrap.ResetDemoData(context.Background(), factory)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to delete demo client")
}

// --- ResetAndReseedDemo error paths ---

func TestResetAndReseedDemo_ResetError(t *testing.T) {
	t.Parallel()

	dbErr := errors.New("db error")
	client := &mockClientRepo{
		getByClientIDFn: func(_ context.Context, _ string) (*cryptoutilIdentityDomain.Client, error) {
			return nil, dbErr
		},
	}
	factory := makeFactory(client, &mockUserRepo{})
	err := cryptoutilIdentityBootstrap.ResetAndReseedDemo(context.Background(), factory)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to reset demo data")
}

func TestResetAndReseedDemo_BootstrapUsersError(t *testing.T) {
	t.Parallel()

	bootstrapErr := errors.New("create user failed")
	// Reset succeeds (both repos return not-found), but BootstrapUsers calls CreateDemoUser which calls GetBySub
	// then Create — make Create fail after Reset succeeds
	createCalled := false
	user := &mockUserRepo{
		getBySubFn: func(_ context.Context, _ string) (*cryptoutilIdentityDomain.User, error) {
			return nil, cryptoutilIdentityAppErr.ErrUserNotFound
		},
		createFn: func(_ context.Context, _ *cryptoutilIdentityDomain.User) error {
			if !createCalled {
				createCalled = true

				return bootstrapErr
			}

			return nil
		},
	}
	client := &mockClientRepo{
		getByClientIDFn: func(_ context.Context, _ string) (*cryptoutilIdentityDomain.Client, error) {
			return nil, cryptoutilIdentityAppErr.ErrClientNotFound
		},
	}
	factory := makeFactory(client, user)
	err := cryptoutilIdentityBootstrap.ResetAndReseedDemo(context.Background(), factory)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to reseed demo users")
}

// --- BootstrapClients additional coverage ---

func TestBootstrapClients_CreateFails(t *testing.T) {
	t.Parallel()

	createErr := errors.New("create failed")
	client := &mockClientRepo{
		getByClientIDFn: func(_ context.Context, _ string) (*cryptoutilIdentityDomain.Client, error) {
			return nil, cryptoutilIdentityAppErr.ErrClientNotFound
		},
		createFn: func(_ context.Context, _ *cryptoutilIdentityDomain.Client) error {
			return createErr
		},
	}
	factory := makeFactory(client, &mockUserRepo{})
	err := cryptoutilIdentityBootstrap.BootstrapClients(context.Background(), nil, factory)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to bootstrap demo-client")
}

func TestBootstrapClients_AlreadyExists(t *testing.T) {
	t.Parallel()

	existingClient := &cryptoutilIdentityDomain.Client{
		ID:       googleUuid.New(),
		ClientID: "demo-client",
	}
	client := &mockClientRepo{
		getByClientIDFn: func(_ context.Context, _ string) (*cryptoutilIdentityDomain.Client, error) {
			return existingClient, nil
		},
	}
	factory := makeFactory(client, &mockUserRepo{})
	err := cryptoutilIdentityBootstrap.BootstrapClients(context.Background(), nil, factory)
	require.NoError(t, err)
}

// --- deleteDemoUser check error ---

func TestDeleteDemoUser_CheckError(t *testing.T) {
	t.Parallel()

	dbErr := errors.New("db unreachable")
	// deleteDemoClient succeeds (client not found), deleteDemoUser GetBySub returns non-ErrUserNotFound error
	client := &mockClientRepo{
		getByClientIDFn: func(_ context.Context, _ string) (*cryptoutilIdentityDomain.Client, error) {
			return nil, cryptoutilIdentityAppErr.ErrClientNotFound
		},
	}
	user := &mockUserRepo{
		getBySubFn: func(_ context.Context, _ string) (*cryptoutilIdentityDomain.User, error) {
			return nil, dbErr
		},
	}
	factory := makeFactory(client, user)
	err := cryptoutilIdentityBootstrap.ResetDemoData(context.Background(), factory)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to delete demo user")
}
