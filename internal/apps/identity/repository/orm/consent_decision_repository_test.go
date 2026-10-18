// Copyright (c) 2025-2026 Justin Cranford.
//
//

package orm

import (
	"context"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

const scopeOpenIDProfile = "openid profile"

func TestConsentDecisionRepository_Create(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewConsentDecisionRepository(testDB.db)
	ctx := context.Background()

	userID := googleUuid.Must(googleUuid.NewV7())
	clientID := "client-id-123"
	scope := testOpenidProfileEmail
	grantedAt := time.Now().UTC()
	expiresAt := grantedAt.Add(cryptoutilSharedMagic.HoursPerDay * time.Hour)

	consent := &cryptoutilIdentityDomain.ConsentDecision{
		ID:        googleUuid.Must(googleUuid.NewV7()),
		UserID:    userID,
		ClientID:  clientID,
		Scope:     scope,
		GrantedAt: grantedAt,
		ExpiresAt: expiresAt,
	}

	err := repo.Create(ctx, consent)
	require.NoError(t, err, "Create should succeed")

	retrieved, err := repo.GetByID(ctx, consent.ID)
	require.NoError(t, err, "GetByID should find created consent")
	require.Equal(t, consent.ID, retrieved.ID)
	require.Equal(t, userID, retrieved.UserID)
	require.Equal(t, clientID, retrieved.ClientID)
	require.Equal(t, scope, retrieved.Scope)
}

func TestConsentDecisionRepository_GetByID(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewConsentDecisionRepository(testDB.db)
	ctx := context.Background()

	nonExistentID := googleUuid.Must(googleUuid.NewV7())

	_, err := repo.GetByID(ctx, nonExistentID)
	require.Error(t, err, "GetByID should return error for non-existent consent")
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrConsentNotFound)
}

func TestConsentDecisionRepository_GetByUserClientScope(t *testing.T) {
	t.Parallel()

	userID := googleUuid.Must(googleUuid.NewV7())
	clientID := "client-id-456"
	scope := scopeOpenIDProfile

	tests := []struct {
		name            string
		setupConsent    bool
		revoked         bool
		expired         bool
		activeDuplicate bool
		wantErr         bool
		wantNotFound    bool
	}{
		{
			name:         "consent_found",
			setupConsent: true,
			revoked:      false,
			expired:      false,
			wantErr:      false,
			wantNotFound: false,
		},
		{
			name:         "consent_not_found",
			setupConsent: false,
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:         "consent_revoked",
			setupConsent: true,
			revoked:      true,
			expired:      false,
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:         testConsentExpired,
			setupConsent: true,
			revoked:      false,
			expired:      true,
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:            "active_consent_with_expired_duplicate",
			setupConsent:    true,
			expired:         true,
			activeDuplicate: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			testDB := setupTestDB(t)
			repo := NewConsentDecisionRepository(testDB.db)
			ctx := context.Background()

			if tc.setupConsent {
				grantedAt := time.Now().UTC().Add(-cryptoutilSharedMagic.JoseJADefaultMaxMaterials * time.Second)

				expiresAt := grantedAt.Add(cryptoutilSharedMagic.HoursPerDay * time.Hour)
				if tc.expired {
					expiresAt = time.Now().UTC().Add(-cryptoutilSharedMagic.JoseJADefaultMaxMaterials * time.Minute) // Clearly expired (10 minutes ago)
				}

				consent := &cryptoutilIdentityDomain.ConsentDecision{
					ID:        googleUuid.Must(googleUuid.NewV7()),
					UserID:    userID,
					ClientID:  clientID,
					Scope:     scope,
					GrantedAt: grantedAt,
					ExpiresAt: expiresAt,
				}

				if tc.revoked {
					revokedAt := time.Now().UTC()
					consent.RevokedAt = &revokedAt
				}

				err := repo.Create(ctx, consent)
				require.NoError(t, err, "Setup consent creation should succeed")

				if tc.activeDuplicate {
					activeConsent := &cryptoutilIdentityDomain.ConsentDecision{
						ID:        googleUuid.Must(googleUuid.NewV7()),
						UserID:    userID,
						ClientID:  clientID,
						Scope:     scope,
						GrantedAt: time.Now().UTC(),
						ExpiresAt: time.Now().UTC().Add(cryptoutilSharedMagic.HoursPerDay * time.Hour),
					}
					require.NoError(t, repo.Create(ctx, activeConsent))
				}
			}

			retrieved, err := repo.GetByUserClientScope(ctx, userID, clientID, scope)

			if tc.wantErr {
				require.Error(t, err)

				if tc.wantNotFound {
					require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrConsentNotFound)
				}

				require.Nil(t, retrieved, "Retrieved consent should be nil on error")
			} else {
				require.NoError(t, err)
				require.NotNil(t, retrieved)
				require.Equal(t, userID, retrieved.UserID)
				require.Equal(t, clientID, retrieved.ClientID)
				require.Equal(t, scope, retrieved.Scope)
			}
		})
	}
}

func TestConsentDecisionRepository_Update(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewConsentDecisionRepository(testDB.db)
	ctx := context.Background()

	userID := googleUuid.Must(googleUuid.NewV7())
	clientID := "client-id-789"
	scope := "openid email"
	grantedAt := time.Now().UTC()
	expiresAt := grantedAt.Add(cryptoutilSharedMagic.HoursPerDay * time.Hour)

	consent := &cryptoutilIdentityDomain.ConsentDecision{
		ID:        googleUuid.Must(googleUuid.NewV7()),
		UserID:    userID,
		ClientID:  clientID,
		Scope:     scope,
		GrantedAt: grantedAt,
		ExpiresAt: expiresAt,
	}

	err := repo.Create(ctx, consent)
	require.NoError(t, err, "Create should succeed")

	newExpiresAt := expiresAt.Add(cryptoutilSharedMagic.HMACSHA384KeySize * time.Hour)
	consent.ExpiresAt = newExpiresAt

	err = repo.Update(ctx, consent)
	require.NoError(t, err, "Update should succeed")

	updated, err := repo.GetByID(ctx, consent.ID)
	require.NoError(t, err, "GetByID should find updated consent")
	require.WithinDuration(t, newExpiresAt, updated.ExpiresAt, time.Second)
}

func TestConsentDecisionRepository_Delete(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewConsentDecisionRepository(testDB.db)
	ctx := context.Background()

	userID := googleUuid.Must(googleUuid.NewV7())
	clientID := "client-id-delete"
	scope := cryptoutilSharedMagic.ScopeOpenID
	grantedAt := time.Now().UTC()
	expiresAt := grantedAt.Add(cryptoutilSharedMagic.HoursPerDay * time.Hour)

	consent := &cryptoutilIdentityDomain.ConsentDecision{
		ID:        googleUuid.Must(googleUuid.NewV7()),
		UserID:    userID,
		ClientID:  clientID,
		Scope:     scope,
		GrantedAt: grantedAt,
		ExpiresAt: expiresAt,
	}

	err := repo.Create(ctx, consent)
	require.NoError(t, err, "Create should succeed")

	err = repo.Delete(ctx, consent.ID)
	require.NoError(t, err, "Delete should succeed")

	_, err = repo.GetByID(ctx, consent.ID)
	require.Error(t, err, "GetByID should fail after delete")
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrConsentNotFound)
}

func TestConsentDecisionRepository_RevokeByID(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewConsentDecisionRepository(testDB.db)
	ctx := context.Background()

	userID := googleUuid.Must(googleUuid.NewV7())
	clientID := "client-id-revoke"
	scope := "openid profile"
	grantedAt := time.Now().UTC()
	expiresAt := grantedAt.Add(cryptoutilSharedMagic.HoursPerDay * time.Hour)

	consent := &cryptoutilIdentityDomain.ConsentDecision{
		ID:        googleUuid.Must(googleUuid.NewV7()),
		UserID:    userID,
		ClientID:  clientID,
		Scope:     scope,
		GrantedAt: grantedAt,
		ExpiresAt: expiresAt,
	}

	err := repo.Create(ctx, consent)
	require.NoError(t, err, "Create should succeed")

	err = repo.RevokeByID(ctx, consent.ID)
	require.NoError(t, err, "RevokeByID should succeed")

	revoked, err := repo.GetByID(ctx, consent.ID)
	require.NoError(t, err, "GetByID should still find revoked consent")
	require.NotNil(t, revoked.RevokedAt, "RevokedAt should be set")
	require.WithinDuration(t, time.Now().UTC(), *revoked.RevokedAt, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
}

func TestConsentDecisionRepository_DeleteExpired(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewConsentDecisionRepository(testDB.db)
	ctx := context.Background()

	userID := googleUuid.Must(googleUuid.NewV7())
	clientID := "client-id-expired"
	now := time.Now().UTC()

	validConsents := []*cryptoutilIdentityDomain.ConsentDecision{
		{
			ID:        googleUuid.Must(googleUuid.NewV7()),
			UserID:    userID,
			ClientID:  clientID,
			Scope:     cryptoutilSharedMagic.ClaimEmail,
			GrantedAt: now,
			ExpiresAt: now.Add(cryptoutilSharedMagic.HoursPerDay * time.Hour),
		},
	}

	for _, consent := range validConsents {
		err := repo.Create(ctx, consent)
		require.NoError(t, err, "Create valid consent should succeed")
	}

	count, err := repo.DeleteExpired(ctx)
	require.NoError(t, err, "DeleteExpired should succeed")
	require.Equal(t, int64(0), count, "Should delete 0 expired consents (none exist)")

	for _, consent := range validConsents {
		retrieved, err := repo.GetByID(ctx, consent.ID)
		require.NoError(t, err, "GetByID should find valid consent")
		require.Equal(t, consent.ID, retrieved.ID)
	}
}
