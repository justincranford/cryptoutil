// Copyright (c) 2025 Justin Cranford

package domain_test

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

func TestRecoveryCode_IsExpired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "not expired",
			expiresAt: time.Now().UTC().Add(1 * time.Hour),
			want:      false,
		},
		{
			name:      "expired",
			expiresAt: time.Now().UTC().Add(-1 * time.Hour),
			want:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			code := &cryptoutilIdentityDomain.RecoveryCode{
				ID:        googleUuid.New(),
				UserID:    googleUuid.New(),
				CodeHash:  "hash",
				ExpiresAt: tc.expiresAt,
				CreatedAt: time.Now().UTC(),
			}

			got := code.IsExpired()
			require.Equal(t, tc.want, got)
		})
	}
}

func TestRecoveryCode_IsUsed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		used bool
		want bool
	}{
		{
			name: "used",
			used: true,
			want: true,
		},
		{
			name: "not used",
			used: false,
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			code := &cryptoutilIdentityDomain.RecoveryCode{
				ID:        googleUuid.New(),
				UserID:    googleUuid.New(),
				CodeHash:  "hash",
				Used:      tc.used,
				ExpiresAt: time.Now().UTC().Add(1 * time.Hour),
				CreatedAt: time.Now().UTC(),
			}

			got := code.IsUsed()
			require.Equal(t, tc.want, got)
		})
	}
}

func TestRecoveryCode_MarkAsUsed(t *testing.T) {
	t.Parallel()

	code := &cryptoutilIdentityDomain.RecoveryCode{
		ID:        googleUuid.New(),
		UserID:    googleUuid.New(),
		CodeHash:  "hash",
		Used:      false,
		UsedAt:    nil,
		ExpiresAt: time.Now().UTC().Add(1 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}

	require.False(t, code.IsUsed())
	require.Nil(t, code.UsedAt)

	code.MarkAsUsed()

	require.True(t, code.IsUsed())
	require.NotNil(t, code.UsedAt)
	require.WithinDuration(t, time.Now().UTC(), *code.UsedAt, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
}
