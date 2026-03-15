// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestClientSecretVersion_IsValid(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	tests := []struct {
		name     string
		secret   *ClientSecretVersion
		checkAt  time.Time
		expected bool
	}{
		{
			name: "active secret with no expiration",
			secret: &ClientSecretVersion{
				Status:    SecretStatusActive,
				ExpiresAt: nil,
			},
			checkAt:  now,
			expected: true,
		},
		{
			name: "active secret with future expiration",
			secret: &ClientSecretVersion{
				Status:    SecretStatusActive,
				ExpiresAt: &future,
			},
			checkAt:  now,
			expected: true,
		},
		{
			name: "active secret with past expiration",
			secret: &ClientSecretVersion{
				Status:    SecretStatusActive,
				ExpiresAt: &past,
			},
			checkAt:  now,
			expected: false,
		},
		{
			name: "expired secret",
			secret: &ClientSecretVersion{
				Status:    SecretStatusExpired,
				ExpiresAt: &past,
			},
			checkAt:  now,
			expected: false,
		},
		{
			name: "revoked secret",
			secret: &ClientSecretVersion{
				Status:    SecretStatusRevoked,
				ExpiresAt: nil,
			},
			checkAt:  now,
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := tc.secret.IsValid(tc.checkAt)
			require.Equal(t, tc.expected, result, "IsValid should return expected result")
		})
	}
}

func TestClientSecretVersion_MarkExpired(t *testing.T) {
	t.Parallel()

	secret := &ClientSecretVersion{
		Status: SecretStatusActive,
	}

	secret.MarkExpired()

	require.Equal(t, SecretStatusExpired, secret.Status, "Status should be expired")
}

func TestClientSecretVersion_MarkRevoked(t *testing.T) {
	t.Parallel()

	secret := &ClientSecretVersion{
		Status: SecretStatusActive,
	}

	revoker := "test-user-id"

	secret.MarkRevoked(revoker)

	require.Equal(t, SecretStatusRevoked, secret.Status, "Status should be revoked")
	require.NotNil(t, secret.RevokedAt, "RevokedAt should be set")
	require.Equal(t, revoker, secret.RevokedBy, "RevokedBy should be set")
}

func TestClientSecretVersion_BeforeCreate(t *testing.T) {
	t.Parallel()

	secret := &ClientSecretVersion{}

	err := secret.BeforeCreate(nil)
	require.NoError(t, err, "BeforeCreate should not return error")
	require.NotEqual(t, googleUuid.Nil, secret.ID, "ID should be generated")
}

func TestClientSecretVersion_TableName(t *testing.T) {
	t.Parallel()

	secret := &ClientSecretVersion{}

	tableName := secret.TableName()
	require.Equal(t, "client_secret_versions", tableName, "TableName should match expected")
}
