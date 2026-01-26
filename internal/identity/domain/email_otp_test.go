// Copyright (c) 2025 Justin Cranford

package domain_test

import (
	"testing"
	"time"

	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestEmailOTP_IsExpired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "not_expired",
			expiresAt: time.Now().UTC().Add(5 * time.Minute),
			want:      false,
		},
		{
			name:      "expired",
			expiresAt: time.Now().UTC().Add(-1 * time.Minute),
			want:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			otp := &cryptoutilIdentityDomain.EmailOTP{
				ID:        googleUuid.New(),
				UserID:    googleUuid.New(),
				ExpiresAt: tc.expiresAt,
			}

			got := otp.IsExpired()
			require.Equal(t, tc.want, got)
		})
	}
}

func TestEmailOTP_IsUsed(t *testing.T) {
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
			name: "not_used",
			used: false,
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			otp := &cryptoutilIdentityDomain.EmailOTP{
				ID:     googleUuid.New(),
				UserID: googleUuid.New(),
				Used:   tc.used,
			}

			got := otp.IsUsed()
			require.Equal(t, tc.want, got)
		})
	}
}

func TestEmailOTP_MarkAsUsed(t *testing.T) {
	t.Parallel()

	otp := &cryptoutilIdentityDomain.EmailOTP{
		ID:     googleUuid.New(),
		UserID: googleUuid.New(),
		Used:   false,
		UsedAt: nil,
	}

	beforeMark := time.Now().UTC()

	otp.MarkAsUsed()

	afterMark := time.Now().UTC()

	require.True(t, otp.Used, "OTP should be marked as used")
	require.NotNil(t, otp.UsedAt, "UsedAt should be set")
	require.True(t, otp.UsedAt.After(beforeMark) || otp.UsedAt.Equal(beforeMark), "UsedAt should be after or equal to beforeMark")
	require.True(t, otp.UsedAt.Before(afterMark) || otp.UsedAt.Equal(afterMark), "UsedAt should be before or equal to afterMark")
}
