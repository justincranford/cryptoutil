// Copyright (c) 2025 Justin Cranford
//
//

package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

func TestAuthorizationRequest_TableName(t *testing.T) {
	t.Parallel()

	authzReq := cryptoutilIdentityDomain.AuthorizationRequest{}
	require.Equal(t, "authorization_requests", authzReq.TableName())
}

func TestAuthorizationRequest_IsExpired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		expiryFn func() time.Time
		want     bool
	}{
		{
			name:     "expired request",
			expiryFn: func() time.Time { return time.Now().UTC().Add(-1 * time.Hour) },
			want:     true,
		},
		{
			name:     "valid request",
			expiryFn: func() time.Time { return time.Now().UTC().Add(1 * time.Hour) },
			want:     false,
		},
		{
			name:     "request expiring now",
			expiryFn: func() time.Time { return time.Now().UTC().Add(1 * time.Millisecond) },
			want:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			authzReq := &cryptoutilIdentityDomain.AuthorizationRequest{
				ExpiresAt: tc.expiryFn(),
			}

			got := authzReq.IsExpired()
			require.Equal(t, tc.want, got)
		})
	}
}

func TestAuthorizationRequest_IsUsed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		used cryptoutilIdentityDomain.IntBool
		want bool
	}{
		{
			name: "unused code",
			used: false,
			want: false,
		},
		{
			name: "used code",
			used: true,
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			authzReq := &cryptoutilIdentityDomain.AuthorizationRequest{
				Used: tc.used,
			}

			got := authzReq.IsUsed()
			require.Equal(t, tc.want, got)
		})
	}
}
