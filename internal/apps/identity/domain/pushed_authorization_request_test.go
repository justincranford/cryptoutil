// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"
)

func TestPushedAuthorizationRequest_IsExpired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		expiresAt   time.Time
		wantExpired bool
	}{
		{
			name:        "not expired",
			expiresAt:   time.Now().UTC().Add(cryptoutilIdentityMagic.DefaultPARLifetime),
			wantExpired: false,
		},
		{
			name:        "expired",
			expiresAt:   time.Now().UTC().Add(-1 * time.Second),
			wantExpired: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			par := &PushedAuthorizationRequest{
				ID:         googleUuid.New(),
				RequestURI: "urn:ietf:params:oauth:request_uri:test",
				ClientID:   googleUuid.New(),
				ExpiresAt:  tc.expiresAt,
				CreatedAt:  time.Now().UTC(),
			}

			got := par.IsExpired()
			require.Equal(t, tc.wantExpired, got, "IsExpired() mismatch")
		})
	}
}

func TestPushedAuthorizationRequest_IsUsed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		used     bool
		wantUsed bool
	}{
		{
			name:     "not used",
			used:     false,
			wantUsed: false,
		},
		{
			name:     "used",
			used:     true,
			wantUsed: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			par := &PushedAuthorizationRequest{
				ID:         googleUuid.New(),
				RequestURI: "urn:ietf:params:oauth:request_uri:test",
				ClientID:   googleUuid.New(),
				Used:       tc.used,
				ExpiresAt:  time.Now().UTC().Add(cryptoutilIdentityMagic.DefaultPARLifetime),
				CreatedAt:  time.Now().UTC(),
			}

			got := par.IsUsed()
			require.Equal(t, tc.wantUsed, got, "IsUsed() mismatch")
		})
	}
}

func TestPushedAuthorizationRequest_MarkAsUsed(t *testing.T) {
	t.Parallel()

	par := &PushedAuthorizationRequest{
		ID:         googleUuid.New(),
		RequestURI: "urn:ietf:params:oauth:request_uri:test",
		ClientID:   googleUuid.New(),
		Used:       false,
		UsedAt:     nil,
		ExpiresAt:  time.Now().UTC().Add(cryptoutilIdentityMagic.DefaultPARLifetime),
		CreatedAt:  time.Now().UTC(),
	}

	require.False(t, par.Used, "should not be used initially")
	require.Nil(t, par.UsedAt, "UsedAt should be nil initially")

	par.MarkAsUsed()

	require.True(t, par.Used, "should be marked as used")
	require.NotNil(t, par.UsedAt, "UsedAt should be set")
	require.WithinDuration(t, time.Now().UTC(), *par.UsedAt, 2*time.Second, "UsedAt should be close to now")
}
