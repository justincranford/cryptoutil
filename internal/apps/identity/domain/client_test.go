// Copyright (c) 2025 Justin Cranford

package domain

import (
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestClient_BeforeCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		clientID googleUuid.UUID
		wantNil  bool
	}{
		{
			name:     "nil_uuid_generates_new",
			clientID: googleUuid.Nil,
			wantNil:  false,
		},
		{
			name:     "existing_uuid_preserved",
			clientID: googleUuid.Must(googleUuid.NewV7()),
			wantNil:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := &Client{
				ID: tc.clientID,
			}

			originalID := client.ID

			err := client.BeforeCreate(&gorm.DB{})
			require.NoError(t, err, "BeforeCreate should not return error")

			if tc.clientID == googleUuid.Nil {
				require.NotEqual(t, googleUuid.Nil, client.ID, "Nil UUID should be replaced with generated UUID")
			} else {
				require.Equal(t, originalID, client.ID, "Existing UUID should be preserved")
			}
		})
	}
}

func TestClient_TableName(t *testing.T) {
	t.Parallel()

	client := Client{}
	tableName := client.TableName()

	require.Equal(t, "clients", tableName, "TableName should return 'clients'")
}
