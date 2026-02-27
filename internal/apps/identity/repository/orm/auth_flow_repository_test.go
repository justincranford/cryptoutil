// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

const updatedDescription = "Updated description"

func TestAuthFlowRepository_Create(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewAuthFlowRepository(testDB.db)

	tests := []struct {
		name      string
		setup     func() *cryptoutilIdentityDomain.AuthFlow
		wantError bool
	}{
		{
			name: "successful_creation",
			setup: func() *cryptoutilIdentityDomain.AuthFlow {
				return &cryptoutilIdentityDomain.AuthFlow{
					Name:        "test_flow",
					FlowType:    cryptoutilIdentityDomain.AuthFlowTypeAuthorizationCode,
					RequirePKCE: true,
					Enabled:     true,
				}
			},
			wantError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			flow := tc.setup()
			err := repo.Create(context.Background(), flow)

			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEqual(t, googleUuid.Nil, flow.ID)

				retrieved, err := repo.GetByID(context.Background(), flow.ID)
				require.NoError(t, err)
				require.Equal(t, flow.ID, retrieved.ID)
				require.Equal(t, flow.Name, retrieved.Name)
			}
		})
	}
}

func TestAuthFlowRepository_GetByID(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewAuthFlowRepository(testDB.db)

	tests := []struct {
		name      string
		setup     func() googleUuid.UUID
		wantError error
	}{
		{
			name: "auth_flow_found",
			setup: func() googleUuid.UUID {
				flow := &cryptoutilIdentityDomain.AuthFlow{
					Name:        "test_flow",
					FlowType:    cryptoutilIdentityDomain.AuthFlowTypeAuthorizationCode,
					RequirePKCE: true,
					Enabled:     true,
				}
				require.NoError(t, repo.Create(context.Background(), flow))

				return flow.ID
			},
			wantError: nil,
		},
		{
			name: "auth_flow_not_found",
			setup: func() googleUuid.UUID {
				return googleUuid.Must(googleUuid.NewV7())
			},
			wantError: cryptoutilIdentityAppErr.ErrAuthFlowNotFound,
		},
		{
			name: "database_error_invalid_uuid",
			setup: func() googleUuid.UUID {
				return googleUuid.Nil
			},
			wantError: cryptoutilIdentityAppErr.ErrAuthFlowNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			flowID := tc.setup()
			flow, err := repo.GetByID(context.Background(), flowID)

			if tc.wantError != nil {
				require.ErrorIs(t, err, tc.wantError)
				require.Nil(t, flow)
			} else {
				require.NoError(t, err)
				require.NotNil(t, flow)
				require.Equal(t, flowID, flow.ID)
			}
		})
	}
}

func TestAuthFlowRepository_GetByName(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewAuthFlowRepository(testDB.db)

	tests := []struct {
		name    string
		setup   func() string
		wantErr error
	}{
		{
			name: "auth_flow_found",
			setup: func() string {
				flow := &cryptoutilIdentityDomain.AuthFlow{
					Name:        "test_flow",
					FlowType:    cryptoutilIdentityDomain.AuthFlowTypeAuthorizationCode,
					RequirePKCE: true,
					Enabled:     true,
				}
				require.NoError(t, repo.Create(context.Background(), flow))

				return flow.Name
			},
			wantErr: nil,
		},
		{
			name: "auth_flow_not_found",
			setup: func() string {
				return "nonexistent_flow"
			},
			wantErr: cryptoutilIdentityAppErr.ErrAuthFlowNotFound,
		},
		{
			name: "database_error_empty_name",
			setup: func() string {
				return ""
			},
			wantErr: cryptoutilIdentityAppErr.ErrAuthFlowNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			flowName := tc.setup()
			flow, err := repo.GetByName(context.Background(), flowName)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				require.Nil(t, flow)
			} else {
				require.NoError(t, err)
				require.NotNil(t, flow)
				require.Equal(t, flowName, flow.Name)
			}
		})
	}
}

func TestAuthFlowRepository_Update(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewAuthFlowRepository(testDB.db)

	tests := []struct {
		name      string
		setup     func() *cryptoutilIdentityDomain.AuthFlow
		modify    func(*cryptoutilIdentityDomain.AuthFlow)
		verify    func(*cryptoutilIdentityDomain.AuthFlow)
		wantError bool
	}{
		{
			name: "successful update",
			setup: func() *cryptoutilIdentityDomain.AuthFlow {
				flow := &cryptoutilIdentityDomain.AuthFlow{
					Name:        "update_test_flow",
					FlowType:    cryptoutilIdentityDomain.AuthFlowTypeAuthorizationCode,
					RequirePKCE: true,
					Enabled:     true,
				}
				require.NoError(t, repo.Create(context.Background(), flow))

				return flow
			},
			modify: func(flow *cryptoutilIdentityDomain.AuthFlow) {
				flow.Description = updatedDescription
				flow.RequireConsent = true
			},
			verify: func(flow *cryptoutilIdentityDomain.AuthFlow) {
				require.Equal(t, updatedDescription, flow.Description)
				require.True(t, flow.RequireConsent)
			},
			wantError: false,
		},
		{
			name: "update_with_invalid_id",
			setup: func() *cryptoutilIdentityDomain.AuthFlow {
				return &cryptoutilIdentityDomain.AuthFlow{
					ID:          googleUuid.Nil,
					Name:        "invalid_flow",
					FlowType:    cryptoutilIdentityDomain.AuthFlowTypeAuthorizationCode,
					RequirePKCE: true,
					Enabled:     true,
				}
			},
			modify: func(flow *cryptoutilIdentityDomain.AuthFlow) {
				flow.Description = "Should not persist"
			},
			verify:    func(_ *cryptoutilIdentityDomain.AuthFlow) {},
			wantError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			flow := tc.setup()
			tc.modify(flow)

			err := repo.Update(context.Background(), flow)

			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				if flow.ID != googleUuid.Nil {
					retrieved, err := repo.GetByID(context.Background(), flow.ID)
					require.NoError(t, err)
					tc.verify(retrieved)
				}
			}
		})
	}
}

func TestAuthFlowRepository_Delete(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewAuthFlowRepository(testDB.db)

	tests := []struct {
		name      string
		setup     func() googleUuid.UUID
		wantError bool
	}{
		{
			name: "successful deletion",
			setup: func() googleUuid.UUID {
				flow := &cryptoutilIdentityDomain.AuthFlow{
					Name:        "delete_test_flow",
					FlowType:    cryptoutilIdentityDomain.AuthFlowTypeAuthorizationCode,
					RequirePKCE: true,
					Enabled:     true,
				}
				require.NoError(t, repo.Create(context.Background(), flow))

				return flow.ID
			},
			wantError: false,
		},
		{
			name: "delete_nonexistent_flow",
			setup: func() googleUuid.UUID {
				return googleUuid.Must(googleUuid.NewV7())
			},
			wantError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			flowID := tc.setup()
			err := repo.Delete(context.Background(), flowID)

			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				_, err := repo.GetByID(context.Background(), flowID)
				require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrAuthFlowNotFound)
			}
		})
	}
}

func TestAuthFlowRepository_List(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewAuthFlowRepository(testDB.db)

	tests := []struct {
		name       string
		setupCount int
		offset     int
		limit      int
		expectMin  int
		wantError  bool
	}{
		{
			name:       "list first page",
			setupCount: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries,
			offset:     0,
			limit:      3,
			expectMin:  3,
			wantError:  false,
		},
		{
			name:       "list with offset",
			setupCount: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries,
			offset:     3,
			limit:      3,
			expectMin:  2,
			wantError:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			for i := 0; i < tc.setupCount; i++ {
				flow := &cryptoutilIdentityDomain.AuthFlow{
					Name:        tc.name + "_flow_" + string(rune('a'+i)),
					FlowType:    cryptoutilIdentityDomain.AuthFlowTypeAuthorizationCode,
					RequirePKCE: true,
					Enabled:     true,
				}
				require.NoError(t, repo.Create(context.Background(), flow))
			}

			flows, err := repo.List(context.Background(), tc.offset, tc.limit)

			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.GreaterOrEqual(t, len(flows), tc.expectMin)
			}
		})
	}
}

func TestAuthFlowRepository_Count(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewAuthFlowRepository(testDB.db)

	tests := []struct {
		name       string
		setupCount int
		expectMin  int64
		wantError  bool
	}{
		{
			name:       "zero count",
			setupCount: 0,
			expectMin:  0,
			wantError:  false,
		},
		{
			name:       "multiple items",
			setupCount: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries,
			expectMin:  cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries,
			wantError:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			for i := 0; i < tc.setupCount; i++ {
				flow := &cryptoutilIdentityDomain.AuthFlow{
					Name:        tc.name + "_flow_" + string(rune('a'+i)),
					FlowType:    cryptoutilIdentityDomain.AuthFlowTypeAuthorizationCode,
					RequirePKCE: true,
					Enabled:     true,
				}
				require.NoError(t, repo.Create(context.Background(), flow))
			}

			count, err := repo.Count(context.Background())

			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.GreaterOrEqual(t, count, tc.expectMin)
			}
		})
	}
}
