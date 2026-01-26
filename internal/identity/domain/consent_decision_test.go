// Copyright (c) 2025 Justin Cranford
//
//

package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
)

func TestConsentDecision_TableName(t *testing.T) {
	t.Parallel()

	consent := cryptoutilIdentityDomain.ConsentDecision{}
	require.Equal(t, "consent_decisions", consent.TableName())
}

func TestConsentDecision_IsExpired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		expiryFn func() time.Time
		want     bool
	}{
		{
			name:     "expired consent",
			expiryFn: func() time.Time { return time.Now().UTC().Add(-1 * time.Hour) },
			want:     true,
		},
		{
			name:     "valid consent",
			expiryFn: func() time.Time { return time.Now().UTC().Add(1 * time.Hour) },
			want:     false,
		},
		{
			name:     "consent expiring now",
			expiryFn: func() time.Time { return time.Now().UTC().Add(1 * time.Millisecond) },
			want:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			consent := &cryptoutilIdentityDomain.ConsentDecision{
				ExpiresAt: tc.expiryFn(),
			}

			got := consent.IsExpired()
			require.Equal(t, tc.want, got)
		})
	}
}

func TestConsentDecision_IsRevoked(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		revokedFn func() *time.Time
		want      bool
	}{
		{
			name: "not revoked",
			revokedFn: func() *time.Time {
				return nil
			},
			want: false,
		},
		{
			name: "revoked",
			revokedFn: func() *time.Time {
				t := time.Now().UTC()

				return &t
			},
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			consent := &cryptoutilIdentityDomain.ConsentDecision{
				RevokedAt: tc.revokedFn(),
			}

			got := consent.IsRevoked()
			require.Equal(t, tc.want, got)
		})
	}
}
