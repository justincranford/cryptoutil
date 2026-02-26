// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

func TestClientRepository_Create(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewClientRepository(testDB.db)
	ctx := context.Background()

	client := &cryptoutilIdentityDomain.Client{
		ID:                      googleUuid.Must(googleUuid.NewV7()),
		ClientID:                "test-client-create",
		ClientSecret:            googleUuid.Must(googleUuid.NewV7()).String(),
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		Name:                    "Test Client",
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
		AllowedGrantTypes:       []string{cryptoutilSharedMagic.GrantTypeAuthorizationCode},
		AllowedResponseTypes:    []string{cryptoutilSharedMagic.ResponseTypeCode},
		AllowedScopes:           []string{cryptoutilSharedMagic.ScopeOpenID},
		RedirectURIs:            []string{cryptoutilSharedMagic.DemoRedirectURI},
		RequirePKCE:             boolPtr(true),
		AccessTokenLifetime:     cryptoutilSharedMagic.IMDefaultSessionTimeout,
		RefreshTokenLifetime:    cryptoutilSharedMagic.IMDefaultSessionAbsoluteMax,
		IDTokenLifetime:         cryptoutilSharedMagic.IMDefaultSessionTimeout,
		Enabled:                 boolPtr(true),
	}

	err := repo.Create(ctx, client)
	require.NoError(t, err)

	// Verify client was created.
	retrieved, err := repo.GetByID(ctx, client.ID)
	require.NoError(t, err)
	require.Equal(t, client.ClientID, retrieved.ClientID)

	// Verify ClientSecretVersion was created (version 1, active).
	var secretVersions []cryptoutilIdentityDomain.ClientSecretVersion

	err = testDB.db.Where("client_id = ?", client.ID).Find(&secretVersions).Error
	require.NoError(t, err)
	require.Len(t, secretVersions, 1, "Expected exactly 1 initial secret version")
	require.Equal(t, 1, secretVersions[0].Version, "Expected version 1 for initial secret")
	require.Equal(t, cryptoutilIdentityDomain.SecretStatusActive, secretVersions[0].Status, "Expected active status")
	require.Nil(t, secretVersions[0].ExpiresAt, "Expected no expiration for initial secret")
	require.NotEmpty(t, secretVersions[0].SecretHash, "Expected non-empty secret hash")

	// Verify KeyRotationEvent was created.
	var events []cryptoutilIdentityDomain.KeyRotationEvent

	err = testDB.db.Where("key_id = ?", client.ID.String()).Find(&events).Error
	require.NoError(t, err)
	require.Len(t, events, 1, "Expected exactly 1 audit event for client creation")
	require.Equal(t, "secret_created", events[0].EventType, "Expected secret_created event type")
	require.Equal(t, cryptoutilSharedMagic.ParamClientSecret, events[0].KeyType, "Expected client_secret key type")
	require.Equal(t, client.ID.String(), events[0].KeyID, "Expected client ID in event")
	require.Equal(t, cryptoutilSharedMagic.SystemInitiatorName, events[0].Initiator, "Expected system initiator")
	require.NotNil(t, events[0].OldKeyVersion, "Expected OldKeyVersion to be set")
	require.Equal(t, 0, *events[0].OldKeyVersion, "Expected OldKeyVersion = 0")
	require.NotNil(t, events[0].NewKeyVersion, "Expected NewKeyVersion to be set")
	require.Equal(t, 1, *events[0].NewKeyVersion, "Expected NewKeyVersion = 1")
	require.NotNil(t, events[0].Success, "Expected Success to be set")
	require.True(t, *events[0].Success, "Expected successful event")
}

func TestClientRepository_GetByID(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewClientRepository(testDB.db)
	ctx := context.Background()

	tests := []struct {
		name    string
		id      googleUuid.UUID
		wantErr error
	}{
		{
			name:    "client not found",
			id:      googleUuid.Must(googleUuid.NewV7()),
			wantErr: cryptoutilIdentityAppErr.ErrClientNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client, err := repo.GetByID(ctx, tc.id)
			require.ErrorIs(t, err, tc.wantErr)
			require.Nil(t, client)
		})
	}
}

func TestClientRepository_GetByClientID(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewClientRepository(testDB.db)
	ctx := context.Background()

	testClient := &cryptoutilIdentityDomain.Client{
		ID:                      googleUuid.Must(googleUuid.NewV7()),
		ClientID:                "test-client-get",
		ClientSecret:            googleUuid.Must(googleUuid.NewV7()).String(),
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		Name:                    "Test Client",
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
		AllowedGrantTypes:       []string{cryptoutilSharedMagic.GrantTypeAuthorizationCode},
		AllowedResponseTypes:    []string{cryptoutilSharedMagic.ResponseTypeCode},
		AllowedScopes:           []string{cryptoutilSharedMagic.ScopeOpenID},
		RedirectURIs:            []string{cryptoutilSharedMagic.DemoRedirectURI},
		RequirePKCE:             boolPtr(true),
		AccessTokenLifetime:     cryptoutilSharedMagic.IMDefaultSessionTimeout,
		RefreshTokenLifetime:    cryptoutilSharedMagic.IMDefaultSessionAbsoluteMax,
		IDTokenLifetime:         cryptoutilSharedMagic.IMDefaultSessionTimeout,
		Enabled:                 boolPtr(true),
	}

	err := repo.Create(ctx, testClient)
	require.NoError(t, err)

	tests := []struct {
		name     string
		clientID string
		wantErr  error
	}{
		{
			name:     "client found",
			clientID: "test-client-get",
			wantErr:  nil,
		},
		{
			name:     "client not found",
			clientID: "nonexistent",
			wantErr:  cryptoutilIdentityAppErr.ErrClientNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client, err := repo.GetByClientID(ctx, tc.clientID)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				require.Nil(t, client)
			} else {
				require.NoError(t, err)
				require.NotNil(t, client)
				require.Equal(t, tc.clientID, client.ClientID)
			}
		})
	}
}

func TestClientRepository_Update(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewClientRepository(testDB.db)
	ctx := context.Background()

	client := &cryptoutilIdentityDomain.Client{
		ID:                      googleUuid.Must(googleUuid.NewV7()),
		ClientID:                "test-client-update",
		ClientSecret:            googleUuid.Must(googleUuid.NewV7()).String(),
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		Name:                    "Test Client",
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
		AllowedGrantTypes:       []string{cryptoutilSharedMagic.GrantTypeAuthorizationCode},
		AllowedResponseTypes:    []string{cryptoutilSharedMagic.ResponseTypeCode},
		AllowedScopes:           []string{cryptoutilSharedMagic.ScopeOpenID},
		RedirectURIs:            []string{cryptoutilSharedMagic.DemoRedirectURI},
		RequirePKCE:             boolPtr(true),
		AccessTokenLifetime:     cryptoutilSharedMagic.IMDefaultSessionTimeout,
		RefreshTokenLifetime:    cryptoutilSharedMagic.IMDefaultSessionAbsoluteMax,
		IDTokenLifetime:         cryptoutilSharedMagic.IMDefaultSessionTimeout,
		Enabled:                 boolPtr(true),
	}

	err := repo.Create(ctx, client)
	require.NoError(t, err)

	client.Name = "Updated Client"
	err = repo.Update(ctx, client)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, client.ID)
	require.NoError(t, err)
	require.Equal(t, "Updated Client", retrieved.Name)
}

func TestClientRepository_Delete(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewClientRepository(testDB.db)
	ctx := context.Background()

	client := &cryptoutilIdentityDomain.Client{
		ID:                      googleUuid.Must(googleUuid.NewV7()),
		ClientID:                "test-client-delete",
		ClientSecret:            googleUuid.Must(googleUuid.NewV7()).String(),
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		Name:                    "Test Client",
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
		AllowedGrantTypes:       []string{cryptoutilSharedMagic.GrantTypeAuthorizationCode},
		AllowedResponseTypes:    []string{cryptoutilSharedMagic.ResponseTypeCode},
		AllowedScopes:           []string{cryptoutilSharedMagic.ScopeOpenID},
		RedirectURIs:            []string{cryptoutilSharedMagic.DemoRedirectURI},
		RequirePKCE:             boolPtr(true),
		AccessTokenLifetime:     cryptoutilSharedMagic.IMDefaultSessionTimeout,
		RefreshTokenLifetime:    cryptoutilSharedMagic.IMDefaultSessionAbsoluteMax,
		IDTokenLifetime:         cryptoutilSharedMagic.IMDefaultSessionTimeout,
		Enabled:                 boolPtr(true),
	}

	err := repo.Create(ctx, client)
	require.NoError(t, err)

	err = repo.Delete(ctx, client.ID)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, client.ID)
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrClientNotFound)
	require.Nil(t, retrieved)
}

func TestClientRepository_List(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewClientRepository(testDB.db)
	ctx := context.Background()

	for i := range cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries {
		client := &cryptoutilIdentityDomain.Client{
			ID:                      googleUuid.Must(googleUuid.NewV7()),
			ClientID:                "test-client-" + string(rune('0'+i)),
			ClientSecret:            googleUuid.Must(googleUuid.NewV7()).String(),
			ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
			Name:                    "Test Client",
			TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
			AllowedGrantTypes:       []string{cryptoutilSharedMagic.GrantTypeAuthorizationCode},
			AllowedResponseTypes:    []string{cryptoutilSharedMagic.ResponseTypeCode},
			AllowedScopes:           []string{cryptoutilSharedMagic.ScopeOpenID},
			RedirectURIs:            []string{cryptoutilSharedMagic.DemoRedirectURI},
			RequirePKCE:             boolPtr(true),
			AccessTokenLifetime:     cryptoutilSharedMagic.IMDefaultSessionTimeout,
			RefreshTokenLifetime:    cryptoutilSharedMagic.IMDefaultSessionAbsoluteMax,
			IDTokenLifetime:         cryptoutilSharedMagic.IMDefaultSessionTimeout,
			Enabled:                 boolPtr(true),
		}
		err := repo.Create(ctx, client)
		require.NoError(t, err)
	}

	clients, err := repo.List(ctx, 0, 3)
	require.NoError(t, err)
	require.Len(t, clients, 3)

	clients, err = repo.List(ctx, 3, 3)
	require.NoError(t, err)
	require.Len(t, clients, 2)
}

func TestClientRepository_Count(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewClientRepository(testDB.db)
	ctx := context.Background()

	count, err := repo.Count(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(0), count)

	for i := range cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries {
		client := &cryptoutilIdentityDomain.Client{
			ID:                      googleUuid.Must(googleUuid.NewV7()),
			ClientID:                "test-client-count-" + string(rune('0'+i)),
			ClientSecret:            googleUuid.Must(googleUuid.NewV7()).String(),
			ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
			Name:                    "Test Client",
			TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
			AllowedGrantTypes:       []string{cryptoutilSharedMagic.GrantTypeAuthorizationCode},
			AllowedResponseTypes:    []string{cryptoutilSharedMagic.ResponseTypeCode},
			AllowedScopes:           []string{cryptoutilSharedMagic.ScopeOpenID},
			RedirectURIs:            []string{cryptoutilSharedMagic.DemoRedirectURI},
			RequirePKCE:             boolPtr(true),
			AccessTokenLifetime:     cryptoutilSharedMagic.IMDefaultSessionTimeout,
			RefreshTokenLifetime:    cryptoutilSharedMagic.IMDefaultSessionAbsoluteMax,
			IDTokenLifetime:         cryptoutilSharedMagic.IMDefaultSessionTimeout,
			Enabled:                 boolPtr(true),
		}
		err := repo.Create(ctx, client)
		require.NoError(t, err)
	}

	count, err = repo.Count(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries), count)
}

func TestClientRepository_GetAll(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewClientRepository(testDB.db)
	ctx := context.Background()

	for i := range cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries {
		client := &cryptoutilIdentityDomain.Client{
			ID:                      googleUuid.Must(googleUuid.NewV7()),
			ClientID:                "test-client-getall-" + string(rune('0'+i)),
			ClientSecret:            googleUuid.Must(googleUuid.NewV7()).String(),
			ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
			Name:                    "Test Client",
			TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
			AllowedGrantTypes:       []string{cryptoutilSharedMagic.GrantTypeAuthorizationCode},
			AllowedResponseTypes:    []string{cryptoutilSharedMagic.ResponseTypeCode},
			AllowedScopes:           []string{cryptoutilSharedMagic.ScopeOpenID},
			RedirectURIs:            []string{cryptoutilSharedMagic.DemoRedirectURI},
			RequirePKCE:             boolPtr(true),
			AccessTokenLifetime:     cryptoutilSharedMagic.IMDefaultSessionTimeout,
			RefreshTokenLifetime:    cryptoutilSharedMagic.IMDefaultSessionAbsoluteMax,
			IDTokenLifetime:         cryptoutilSharedMagic.IMDefaultSessionTimeout,
			Enabled:                 boolPtr(true),
		}
		err := repo.Create(ctx, client)
		require.NoError(t, err)
	}

	clients, err := repo.GetAll(ctx)
	require.NoError(t, err)
	require.Len(t, clients, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)
}
