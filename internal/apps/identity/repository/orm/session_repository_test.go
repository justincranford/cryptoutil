// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

func TestSessionRepository_Create(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewSessionRepository(testDB.db)

	userID := googleUuid.Must(googleUuid.NewV7())
	clientID := googleUuid.Must(googleUuid.NewV7())
	active := true

	session := &cryptoutilIdentityDomain.Session{
		SessionID:             "session-12345",
		UserID:                userID,
		ClientID:              cryptoutilIdentityDomain.NullableUUID{UUID: clientID, Valid: true},
		IPAddress:             "192.168.1.100",
		UserAgent:             "Mozilla/5.0",
		IssuedAt:              time.Now().UTC(),
		ExpiresAt:             time.Now().UTC().Add(cryptoutilSharedMagic.HoursPerDay * time.Hour),
		LastSeenAt:            time.Now().UTC(),
		Active:                &active,
		AuthenticationMethods: []string{"password", cryptoutilSharedMagic.MFATypeTOTP},
		AuthenticationTime:    time.Now().UTC(),
		GrantedScopes:         []string{cryptoutilSharedMagic.ScopeOpenID, cryptoutilSharedMagic.ClaimProfile},
	}

	err := repo.Create(context.Background(), session)
	require.NoError(t, err)
	require.NotEqual(t, googleUuid.Nil, session.ID)

	retrieved, err := repo.GetByID(context.Background(), session.ID)
	require.NoError(t, err)
	require.Equal(t, session.SessionID, retrieved.SessionID)
	require.Equal(t, session.UserID, retrieved.UserID)
	require.NotNil(t, retrieved.Active)
	require.True(t, *retrieved.Active)
	require.Len(t, retrieved.AuthenticationMethods, 2)
	require.Len(t, retrieved.GrantedScopes, 2)
}

func TestSessionRepository_GetByID(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewSessionRepository(testDB.db)

	nonExistentID := googleUuid.Must(googleUuid.NewV7())
	_, err := repo.GetByID(context.Background(), nonExistentID)
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrSessionNotFound)
}

func TestSessionRepository_GetBySessionID(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewSessionRepository(testDB.db)

	tests := []struct {
		name    string
		setup   func() string
		wantErr error
	}{
		{
			name: "session_found",
			setup: func() string {
				userID := googleUuid.Must(googleUuid.NewV7())
				session := &cryptoutilIdentityDomain.Session{
					SessionID:  "test-session-123",
					UserID:     userID,
					IPAddress:  cryptoutilSharedMagic.IPv4Loopback,
					IssuedAt:   time.Now().UTC(),
					ExpiresAt:  time.Now().UTC().Add(cryptoutilSharedMagic.HoursPerDay * time.Hour),
					LastSeenAt: time.Now().UTC(),
					Active:     boolPtr(true),
				}
				err := repo.Create(context.Background(), session)
				require.NoError(t, err)

				return session.SessionID
			},
			wantErr: nil,
		},
		{
			name: "session_not_found",
			setup: func() string {
				return "nonexistent-session-id"
			},
			wantErr: cryptoutilIdentityAppErr.ErrSessionNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sessionID := tc.setup()
			_, err := repo.GetBySessionID(context.Background(), sessionID)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSessionRepository_Update(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewSessionRepository(testDB.db)

	userID := googleUuid.Must(googleUuid.NewV7())

	session := &cryptoutilIdentityDomain.Session{
		SessionID:  "update-session",
		UserID:     userID,
		IPAddress:  cryptoutilSharedMagic.IPv4Loopback,
		IssuedAt:   time.Now().UTC(),
		ExpiresAt:  time.Now().UTC().Add(cryptoutilSharedMagic.HoursPerDay * time.Hour),
		LastSeenAt: time.Now().UTC(),
		Active:     boolPtr(true),
	}
	err := repo.Create(context.Background(), session)
	require.NoError(t, err)

	session.LastSeenAt = time.Now().UTC().Add(1 * time.Hour)
	session.IPAddress = "192.168.1.200"
	err = repo.Update(context.Background(), session)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(context.Background(), session.ID)
	require.NoError(t, err)
	require.Equal(t, "192.168.1.200", retrieved.IPAddress)
}

func TestSessionRepository_Delete(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewSessionRepository(testDB.db)

	userID := googleUuid.Must(googleUuid.NewV7())

	session := &cryptoutilIdentityDomain.Session{
		SessionID:  "delete-session",
		UserID:     userID,
		IPAddress:  cryptoutilSharedMagic.IPv4Loopback,
		IssuedAt:   time.Now().UTC(),
		ExpiresAt:  time.Now().UTC().Add(cryptoutilSharedMagic.HoursPerDay * time.Hour),
		LastSeenAt: time.Now().UTC(),
		Active:     boolPtr(true),
	}
	err := repo.Create(context.Background(), session)
	require.NoError(t, err)

	err = repo.Delete(context.Background(), session.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(context.Background(), session.ID)
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrSessionNotFound)
}

func TestSessionRepository_TerminateByID(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewSessionRepository(testDB.db)

	userID := googleUuid.Must(googleUuid.NewV7())

	session := &cryptoutilIdentityDomain.Session{
		SessionID:  "terminate-by-id-session",
		UserID:     userID,
		IPAddress:  cryptoutilSharedMagic.IPv4Loopback,
		IssuedAt:   time.Now().UTC(),
		ExpiresAt:  time.Now().UTC().Add(cryptoutilSharedMagic.HoursPerDay * time.Hour),
		LastSeenAt: time.Now().UTC(),
		Active:     boolPtr(true),
	}
	err := repo.Create(context.Background(), session)
	require.NoError(t, err)
	require.NotNil(t, session.Active)
	require.True(t, *session.Active)

	err = repo.TerminateByID(context.Background(), session.ID)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(context.Background(), session.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved.Active)
	require.False(t, *retrieved.Active)
}

func TestSessionRepository_TerminateBySessionID(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewSessionRepository(testDB.db)

	userID := googleUuid.Must(googleUuid.NewV7())

	session := &cryptoutilIdentityDomain.Session{
		SessionID:  "terminate-by-session-id",
		UserID:     userID,
		IPAddress:  cryptoutilSharedMagic.IPv4Loopback,
		IssuedAt:   time.Now().UTC(),
		ExpiresAt:  time.Now().UTC().Add(cryptoutilSharedMagic.HoursPerDay * time.Hour),
		LastSeenAt: time.Now().UTC(),
		Active:     boolPtr(true),
	}
	err := repo.Create(context.Background(), session)
	require.NoError(t, err)
	require.NotNil(t, session.Active)
	require.True(t, *session.Active)

	err = repo.TerminateBySessionID(context.Background(), session.SessionID)
	require.NoError(t, err)

	retrieved, err := repo.GetBySessionID(context.Background(), session.SessionID)
	require.NoError(t, err)
	require.NotNil(t, retrieved.Active)
	require.False(t, *retrieved.Active)
}

func TestSessionRepository_DeleteExpired(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewSessionRepository(testDB.db)

	userID := googleUuid.Must(googleUuid.NewV7())

	// Create expired session.
	expiredSession := &cryptoutilIdentityDomain.Session{
		SessionID:  "expired-session",
		UserID:     userID,
		IPAddress:  cryptoutilSharedMagic.IPv4Loopback,
		IssuedAt:   time.Now().UTC().Add(-cryptoutilSharedMagic.HMACSHA384KeySize * time.Hour),
		ExpiresAt:  time.Now().UTC().Add(-cryptoutilSharedMagic.HoursPerDay * time.Hour), // Expired.
		LastSeenAt: time.Now().UTC().Add(-cryptoutilSharedMagic.HoursPerDay * time.Hour),
		Active:     boolPtr(true),
	}
	err := repo.Create(context.Background(), expiredSession)
	require.NoError(t, err)

	// Create valid session.
	validSession := &cryptoutilIdentityDomain.Session{
		SessionID:  "valid-session",
		UserID:     userID,
		IPAddress:  cryptoutilSharedMagic.IPv4Loopback,
		IssuedAt:   time.Now().UTC(),
		ExpiresAt:  time.Now().UTC().Add(cryptoutilSharedMagic.HoursPerDay * time.Hour), // Not expired.
		LastSeenAt: time.Now().UTC(),
		Active:     boolPtr(true),
	}
	err = repo.Create(context.Background(), validSession)
	require.NoError(t, err)

	err = repo.DeleteExpired(context.Background())
	require.NoError(t, err)

	_, err = repo.GetByID(context.Background(), expiredSession.ID)
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrSessionNotFound)

	_, err = repo.GetByID(context.Background(), validSession.ID)
	require.NoError(t, err)
}

func TestSessionRepository_DeleteExpiredBefore(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewSessionRepository(testDB.db)

	userID := googleUuid.Must(googleUuid.NewV7())

	// Create sessions expired at different times.
	session1 := &cryptoutilIdentityDomain.Session{
		SessionID:  "session-1",
		UserID:     userID,
		IPAddress:  cryptoutilSharedMagic.IPv4Loopback,
		IssuedAt:   time.Now().UTC().Add(-72 * time.Hour),
		ExpiresAt:  time.Now().UTC().Add(-cryptoutilSharedMagic.HMACSHA384KeySize * time.Hour), // Expired 48h ago.
		LastSeenAt: time.Now().UTC().Add(-cryptoutilSharedMagic.HMACSHA384KeySize * time.Hour),
		Active:     boolPtr(true),
	}
	err := repo.Create(context.Background(), session1)
	require.NoError(t, err)

	session2 := &cryptoutilIdentityDomain.Session{
		SessionID:  "session-2",
		UserID:     userID,
		IPAddress:  cryptoutilSharedMagic.IPv4Loopback,
		IssuedAt:   time.Now().UTC().Add(-cryptoutilSharedMagic.UUIDStringLength * time.Hour),
		ExpiresAt:  time.Now().UTC().Add(-cryptoutilSharedMagic.HashPrefixLength * time.Hour), // Expired 12h ago.
		LastSeenAt: time.Now().UTC().Add(-cryptoutilSharedMagic.HashPrefixLength * time.Hour),
		Active:     boolPtr(true),
	}
	err = repo.Create(context.Background(), session2)
	require.NoError(t, err)

	cutoffTime := time.Now().UTC().Add(-cryptoutilSharedMagic.HoursPerDay * time.Hour)
	deletedCount, err := repo.DeleteExpiredBefore(context.Background(), cutoffTime)
	require.NoError(t, err)
	require.Equal(t, 1, deletedCount) // Only session1 deleted (expired 48h ago).

	_, err = repo.GetByID(context.Background(), session1.ID)
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrSessionNotFound)

	_, err = repo.GetByID(context.Background(), session2.ID)
	require.NoError(t, err) // session2 still exists (expired 12h ago).
}

func TestSessionRepository_List(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewSessionRepository(testDB.db)

	userID := googleUuid.Must(googleUuid.NewV7())

	// Create 5 sessions.
	for i := 0; i < cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries; i++ {
		session := &cryptoutilIdentityDomain.Session{
			SessionID:  googleUuid.Must(googleUuid.NewV7()).String(),
			UserID:     userID,
			IPAddress:  cryptoutilSharedMagic.IPv4Loopback,
			IssuedAt:   time.Now().UTC(),
			ExpiresAt:  time.Now().UTC().Add(cryptoutilSharedMagic.HoursPerDay * time.Hour),
			LastSeenAt: time.Now().UTC(),
			Active:     boolPtr(true),
		}
		err := repo.Create(context.Background(), session)
		require.NoError(t, err)
	}

	// Test pagination.
	sessions, err := repo.List(context.Background(), 0, 3)
	require.NoError(t, err)
	require.Len(t, sessions, 3)

	sessions, err = repo.List(context.Background(), 3, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)
	require.NoError(t, err)
	require.Len(t, sessions, 2)
}

func TestSessionRepository_Count(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewSessionRepository(testDB.db)

	// Initial count should be 0.
	count, err := repo.Count(context.Background())
	require.NoError(t, err)
	require.Equal(t, int64(0), count)

	userID := googleUuid.Must(googleUuid.NewV7())

	// Create 5 sessions.
	for i := 0; i < cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries; i++ {
		session := &cryptoutilIdentityDomain.Session{
			SessionID:  googleUuid.Must(googleUuid.NewV7()).String(),
			UserID:     userID,
			IPAddress:  cryptoutilSharedMagic.IPv4Loopback,
			IssuedAt:   time.Now().UTC(),
			ExpiresAt:  time.Now().UTC().Add(cryptoutilSharedMagic.HoursPerDay * time.Hour),
			LastSeenAt: time.Now().UTC(),
			Active:     boolPtr(true),
		}
		err := repo.Create(context.Background(), session)
		require.NoError(t, err)
	}

	// Count should be 5.
	count, err = repo.Count(context.Background())
	require.NoError(t, err)
	require.Equal(t, int64(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries), count)
}
