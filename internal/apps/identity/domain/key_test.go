// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestKey_TableName(t *testing.T) {
	t.Parallel()

	key := &Key{}
	require.Equal(t, "keys", key.TableName())
}

func TestKey_BeforeCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		key     *Key
		wantNil bool
	}{
		{
			name:    "id_not_set",
			key:     &Key{ID: googleUuid.Nil},
			wantNil: false,
		},
		{
			name: "id_already_set",
			key: &Key{
				ID: googleUuid.Must(googleUuid.NewV7()),
			},
			wantNil: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			originalID := tc.key.ID
			err := tc.key.BeforeCreate(&gorm.DB{})

			require.NoError(t, err)

			if originalID == googleUuid.Nil {
				require.NotEqual(t, googleUuid.Nil, tc.key.ID, "ID should be set")
			} else {
				require.Equal(t, originalID, tc.key.ID, "ID should remain unchanged")
			}
		})
	}
}

func TestKey_IsExpired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "expired_key",
			expiresAt: time.Now().UTC().Add(-1 * time.Hour),
			want:      true,
		},
		{
			name:      "not_expired_key",
			expiresAt: time.Now().UTC().Add(1 * time.Hour),
			want:      false,
		},
		{
			name:      "just_expired",
			expiresAt: time.Now().UTC().Add(-1 * time.Second),
			want:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			key := &Key{ExpiresAt: tc.expiresAt}
			result := key.IsExpired()

			require.Equal(t, tc.want, result)
		})
	}
}
