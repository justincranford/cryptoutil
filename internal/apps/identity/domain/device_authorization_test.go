// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestDeviceAuthorization_TableName validates the table name.
func TestDeviceAuthorization_TableName(t *testing.T) {
	t.Parallel()

	auth := &DeviceAuthorization{}
	require.Equal(t, "device_authorizations", auth.TableName())
}

// TestDeviceAuthorization_IsExpired validates expiration check.
func TestDeviceAuthorization_IsExpired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "not expired - future expiration",
			expiresAt: time.Now().UTC().Add(cryptoutilSharedMagic.JoseJADefaultMaxMaterials * time.Minute),
			want:      false,
		},
		{
			name:      "expired - past expiration",
			expiresAt: time.Now().UTC().Add(-cryptoutilSharedMagic.JoseJADefaultMaxMaterials * time.Minute),
			want:      true,
		},
		{
			name:      "expired - just expired",
			expiresAt: time.Now().UTC().Add(-1 * time.Second),
			want:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			auth := &DeviceAuthorization{
				ExpiresAt: tc.expiresAt,
			}

			got := auth.IsExpired()
			require.Equal(t, tc.want, got, "IsExpired() mismatch")
		})
	}
}

// TestDeviceAuthorization_StatusChecks validates status check methods.
func TestDeviceAuthorization_StatusChecks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		status         string
		wantPending    bool
		wantAuthorized bool
		wantDenied     bool
		wantUsed       bool
	}{
		{
			name:           "pending status",
			status:         DeviceAuthStatusPending,
			wantPending:    true,
			wantAuthorized: false,
			wantDenied:     false,
			wantUsed:       false,
		},
		{
			name:           "authorized status",
			status:         DeviceAuthStatusAuthorized,
			wantPending:    false,
			wantAuthorized: true,
			wantDenied:     false,
			wantUsed:       false,
		},
		{
			name:           "denied status",
			status:         DeviceAuthStatusDenied,
			wantPending:    false,
			wantAuthorized: false,
			wantDenied:     true,
			wantUsed:       false,
		},
		{
			name:           "used status",
			status:         DeviceAuthStatusUsed,
			wantPending:    false,
			wantAuthorized: false,
			wantDenied:     false,
			wantUsed:       true,
		},
		{
			name:           "unknown status",
			status:         "unknown",
			wantPending:    false,
			wantAuthorized: false,
			wantDenied:     false,
			wantUsed:       false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			auth := &DeviceAuthorization{
				Status: tc.status,
			}

			require.Equal(t, tc.wantPending, auth.IsPending(), "IsPending() mismatch")
			require.Equal(t, tc.wantAuthorized, auth.IsAuthorized(), "IsAuthorized() mismatch")
			require.Equal(t, tc.wantDenied, auth.IsDenied(), "IsDenied() mismatch")
			require.Equal(t, tc.wantUsed, auth.IsUsed(), "IsUsed() mismatch")
		})
	}
}

// TestDeviceAuthorization_FullLifecycle validates the complete authorization lifecycle.
func TestDeviceAuthorization_FullLifecycle(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	authID := googleUuid.Must(googleUuid.NewV7())

	auth := &DeviceAuthorization{
		ID:         authID,
		ClientID:   "test-client",
		DeviceCode: "device-code-123",
		UserCode:   "WDJB-MJHT",
		Scope:      "openid profile",
		Status:     DeviceAuthStatusPending,
		CreatedAt:  now,
		ExpiresAt:  now.Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days * time.Minute),
	}

	// Initial state: pending, not expired.
	require.True(t, auth.IsPending(), "Should be pending initially")
	require.False(t, auth.IsExpired(), "Should not be expired")
	require.False(t, auth.IsAuthorized(), "Should not be authorized yet")
	require.False(t, auth.IsDenied(), "Should not be denied")
	require.False(t, auth.IsUsed(), "Should not be used yet")

	// User authorizes.
	auth.Status = DeviceAuthStatusAuthorized
	userID := googleUuid.Must(googleUuid.NewV7())
	auth.UserID = NullableUUID{UUID: userID, Valid: true}

	require.False(t, auth.IsPending(), "Should not be pending after authorization")
	require.True(t, auth.IsAuthorized(), "Should be authorized")
	require.False(t, auth.IsDenied(), "Should not be denied")
	require.False(t, auth.IsUsed(), "Should not be used yet")

	// Device exchanges code for tokens.
	auth.Status = DeviceAuthStatusUsed
	usedTime := time.Now().UTC()
	auth.UsedAt = &usedTime

	require.False(t, auth.IsPending(), "Should not be pending after token exchange")
	require.False(t, auth.IsAuthorized(), "Should not be authorized after use")
	require.False(t, auth.IsDenied(), "Should not be denied")
	require.True(t, auth.IsUsed(), "Should be used after token exchange")
}

// TestDeviceAuthorization_PollingMetadata validates polling timestamp tracking.
func TestDeviceAuthorization_PollingMetadata(t *testing.T) {
	t.Parallel()

	auth := &DeviceAuthorization{
		ID:         googleUuid.Must(googleUuid.NewV7()),
		ClientID:   "test-client",
		DeviceCode: "device-code-123",
		UserCode:   "ABCD-1234",
		Status:     DeviceAuthStatusPending,
		CreatedAt:  time.Now().UTC(),
		ExpiresAt:  time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days * time.Minute),
	}

	// Initially no polling timestamp.
	require.Nil(t, auth.LastPolledAt, "LastPolledAt should be nil initially")

	// First poll.
	firstPoll := time.Now().UTC()
	auth.LastPolledAt = &firstPoll

	require.NotNil(t, auth.LastPolledAt, "LastPolledAt should be set after first poll")
	require.WithinDuration(t, firstPoll, *auth.LastPolledAt, 1*time.Second, "LastPolledAt should match first poll time")

	// Second poll (5 seconds later).
	time.Sleep(cryptoutilSharedMagic.JoseJAMaxMaterials * time.Millisecond) // Simulate small delay for test.

	secondPoll := time.Now().UTC()
	auth.LastPolledAt = &secondPoll

	require.True(t, secondPoll.After(firstPoll), "Second poll should be after first poll")
	require.WithinDuration(t, secondPoll, *auth.LastPolledAt, 1*time.Second, "LastPolledAt should match second poll time")
}
