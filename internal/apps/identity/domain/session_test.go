// Copyright (c) 2025 Justin Cranford
//
//

package domain_test

import (
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

func TestSession_BeforeCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		setupSession      func() *cryptoutilIdentityDomain.Session
		validateID        bool
		validateSessionID bool
	}{
		{
			name: "generates ID when empty",
			setupSession: func() *cryptoutilIdentityDomain.Session {
				return &cryptoutilIdentityDomain.Session{}
			},
			validateID:        true,
			validateSessionID: true,
		},
		{
			name: "preserves existing ID",
			setupSession: func() *cryptoutilIdentityDomain.Session {
				existingID := googleUuid.Must(googleUuid.NewV7())

				return &cryptoutilIdentityDomain.Session{ID: existingID}
			},
			validateID:        false,
			validateSessionID: true,
		},
		{
			name: "generates SessionID when empty",
			setupSession: func() *cryptoutilIdentityDomain.Session {
				existingID := googleUuid.Must(googleUuid.NewV7())

				return &cryptoutilIdentityDomain.Session{ID: existingID}
			},
			validateID:        false,
			validateSessionID: true,
		},
		{
			name: "preserves existing SessionID",
			setupSession: func() *cryptoutilIdentityDomain.Session {
				existingID := googleUuid.Must(googleUuid.NewV7())
				existingSessionID := "existing-session-id"

				return &cryptoutilIdentityDomain.Session{
					ID:        existingID,
					SessionID: existingSessionID,
				}
			},
			validateID:        false,
			validateSessionID: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			session := tc.setupSession()
			originalID := session.ID
			originalSessionID := session.SessionID

			err := session.BeforeCreate(nil)
			require.NoError(t, err)

			if tc.validateID {
				require.NotEqual(t, googleUuid.Nil, session.ID, "ID should be generated")
			} else {
				require.Equal(t, originalID, session.ID, "ID should be preserved")
			}

			if tc.validateSessionID {
				require.NotEmpty(t, session.SessionID, "SessionID should be generated")
			} else {
				require.Equal(t, originalSessionID, session.SessionID, "SessionID should be preserved")
			}
		})
	}
}

func TestSession_TableName(t *testing.T) {
	t.Parallel()

	session := cryptoutilIdentityDomain.Session{}
	require.Equal(t, "sessions", session.TableName())
}

func TestSession_IsExpired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		expiryFn func() time.Time
		want     bool
	}{
		{
			name:     "expired session",
			expiryFn: func() time.Time { return time.Now().UTC().Add(-1 * time.Hour) },
			want:     true,
		},
		{
			name:     "valid session",
			expiryFn: func() time.Time { return time.Now().UTC().Add(1 * time.Hour) },
			want:     false,
		},
		{
			name:     "session expiring now",
			expiryFn: func() time.Time { return time.Now().UTC().Add(1 * time.Millisecond) },
			want:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			session := &cryptoutilIdentityDomain.Session{
				ExpiresAt: tc.expiryFn(),
			}

			got := session.IsExpired()
			require.Equal(t, tc.want, got)
		})
	}
}

func TestSession_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		expiryFn func() time.Time
		active   *bool
		want     bool
	}{
		{
			name:     "valid active session",
			expiryFn: func() time.Time { return time.Now().UTC().Add(1 * time.Hour) },
			active:   boolPtr(true),
			want:     true,
		},
		{
			name:     "expired but active",
			expiryFn: func() time.Time { return time.Now().UTC().Add(-1 * time.Hour) },
			active:   boolPtr(true),
			want:     false,
		},
		{
			name:     "valid but inactive",
			expiryFn: func() time.Time { return time.Now().UTC().Add(1 * time.Hour) },
			active:   boolPtr(false),
			want:     false,
		},
		{
			name:     "expired and inactive",
			expiryFn: func() time.Time { return time.Now().UTC().Add(-1 * time.Hour) },
			active:   boolPtr(false),
			want:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			session := &cryptoutilIdentityDomain.Session{
				ExpiresAt: tc.expiryFn(),
				Active:    tc.active,
			}

			got := session.IsValid()
			require.Equal(t, tc.want, got)
		})
	}
}
