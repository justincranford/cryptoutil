// Copyright (c) 2025-2026 Justin Cranford.
package orm

import (
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestCoreRepositories_DatabaseErrors(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	sqlDB, err := testDB.db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	ctx := context.Background()
	id := googleUuid.Must(googleUuid.NewV7())
	authFlow := NewAuthFlowRepository(testDB.db)
	authProfile := NewAuthProfileRepository(testDB.db)
	clientProfile := NewClientProfileRepository(testDB.db)
	authorizationRequest := NewAuthorizationRequestRepository(testDB.db)

	tests := []struct {
		name      string
		operation func() error
	}{
		{"auth_flow_create", func() error { return authFlow.Create(ctx, &cryptoutilIdentityDomain.AuthFlow{}) }},
		{"auth_flow_get_by_id", func() error {
			_, err := authFlow.GetByID(ctx, id)

			return err
		}},
		{"auth_flow_get_by_name", func() error {
			_, err := authFlow.GetByName(ctx, "flow")

			return err
		}},
		{"auth_flow_update", func() error { return authFlow.Update(ctx, &cryptoutilIdentityDomain.AuthFlow{}) }},
		{"auth_flow_delete", func() error { return authFlow.Delete(ctx, id) }},
		{"auth_flow_list", func() error {
			_, err := authFlow.List(ctx, 0, 1)

			return err
		}},
		{"auth_flow_count", func() error {
			_, err := authFlow.Count(ctx)

			return err
		}},
		{"auth_profile_create", func() error { return authProfile.Create(ctx, &cryptoutilIdentityDomain.AuthProfile{}) }},
		{"auth_profile_get_by_id", func() error {
			_, err := authProfile.GetByID(ctx, id)

			return err
		}},
		{"auth_profile_get_by_name", func() error {
			_, err := authProfile.GetByName(ctx, cryptoutilSharedMagic.ScopeProfile)

			return err
		}},
		{"auth_profile_update", func() error { return authProfile.Update(ctx, &cryptoutilIdentityDomain.AuthProfile{}) }},
		{"auth_profile_delete", func() error { return authProfile.Delete(ctx, id) }},
		{"auth_profile_list", func() error {
			_, err := authProfile.List(ctx, 0, 1)

			return err
		}},
		{"auth_profile_count", func() error {
			_, err := authProfile.Count(ctx)

			return err
		}},
		{"client_profile_create", func() error { return clientProfile.Create(ctx, &cryptoutilIdentityDomain.ClientProfile{}) }},
		{"client_profile_get_by_id", func() error {
			_, err := clientProfile.GetByID(ctx, id)

			return err
		}},
		{"client_profile_get_by_name", func() error {
			_, err := clientProfile.GetByName(ctx, cryptoutilSharedMagic.ScopeProfile)

			return err
		}},
		{"client_profile_update", func() error { return clientProfile.Update(ctx, &cryptoutilIdentityDomain.ClientProfile{}) }},
		{"client_profile_delete", func() error { return clientProfile.Delete(ctx, id) }},
		{"client_profile_list", func() error {
			_, err := clientProfile.List(ctx, 0, 1)

			return err
		}},
		{"client_profile_count", func() error {
			_, err := clientProfile.Count(ctx)

			return err
		}},
		{"authorization_request_create", func() error {
			return authorizationRequest.Create(ctx, &cryptoutilIdentityDomain.AuthorizationRequest{})
		}},
		{"authorization_request_get_by_id", func() error {
			_, err := authorizationRequest.GetByID(ctx, id)

			return err
		}},
		{"authorization_request_get_by_code", func() error {
			_, err := authorizationRequest.GetByCode(ctx, cryptoutilSharedMagic.ResponseTypeCode)

			return err
		}},
		{"authorization_request_update", func() error {
			return authorizationRequest.Update(ctx, &cryptoutilIdentityDomain.AuthorizationRequest{})
		}},
		{"authorization_request_delete", func() error { return authorizationRequest.Delete(ctx, id) }},
		{"authorization_request_delete_expired", func() error {
			_, err := authorizationRequest.DeleteExpired(ctx)

			return err
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Error(t, tc.operation())
		})
	}
}
