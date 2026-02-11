// Copyright (c) 2025 Justin Cranford
//
//

package domain_test

import (
	"database/sql"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite" // CGO-free SQLite driver

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

func TestNullableUUID_SQLiteIntegration(t *testing.T) {
	t.Parallel()

	// Open in-memory SQLite database using modernc.org/sqlite (CGO-free).
	sqlDB, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	defer func() {
		if closeErr := sqlDB.Close(); closeErr != nil {
			t.Logf("Failed to close test database: %v", closeErr)
		}
	}()

	// Wrap with GORM.

	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	// AutoMigrate to create table.
	err = db.AutoMigrate(&cryptoutilIdentityDomain.Client{})
	require.NoError(t, err)

	// Create client with nil ClientProfileID.
	testClient := &cryptoutilIdentityDomain.Client{
		ID:                      googleUuid.Must(googleUuid.NewV7()),
		ClientID:                "test-client",
		ClientType:              "confidential",
		Name:                    "Test Client",
		TokenEndpointAuthMethod: "client_secret_post",
	}

	err = db.Create(testClient).Error
	require.NoError(t, err, "Failed to create client with nil ClientProfileID")

	// Verify client was created.
	var count int64

	err = db.Model(&cryptoutilIdentityDomain.Client{}).Count(&count).Error
	require.NoError(t, err)
	require.Equal(t, int64(1), count)

	// Create client WITH ClientProfileID.
	profileID := googleUuid.Must(googleUuid.NewV7())
	testClient2 := &cryptoutilIdentityDomain.Client{
		ID:                      googleUuid.Must(googleUuid.NewV7()),
		ClientID:                "test-client-2",
		ClientType:              "confidential",
		Name:                    "Test Client 2",
		TokenEndpointAuthMethod: "client_secret_post",
		ClientProfileID:         cryptoutilIdentityDomain.NewNullableUUID(&profileID),
	}

	err = db.Create(testClient2).Error
	require.NoError(t, err, "Failed to create client with ClientProfileID")

	// Verify second client was created.
	err = db.Model(&cryptoutilIdentityDomain.Client{}).Count(&count).Error
	require.NoError(t, err)
	require.Equal(t, int64(2), count)
}

func TestNullableUUID_NewNullableUUID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     *googleUuid.UUID
		wantValid bool
	}{
		{
			name: "valid UUID pointer",
			input: func() *googleUuid.UUID {
				id := googleUuid.New()

				return &id
			}(),
			wantValid: true,
		},
		{
			name:      "nil UUID pointer",
			input:     nil,
			wantValid: false,
		},
		{
			name: "Nil UUID value",
			input: func() *googleUuid.UUID {
				id := googleUuid.Nil

				return &id
			}(),
			wantValid: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := cryptoutilIdentityDomain.NewNullableUUID(tc.input)
			require.Equal(t, tc.wantValid, result.Valid)

			if tc.wantValid && tc.input != nil {
				require.Equal(t, *tc.input, result.UUID)
			}
		})
	}
}

func TestNullableUUID_Ptr(t *testing.T) {
	t.Parallel()

	validID := googleUuid.New()

	tests := []struct {
		name     string
		nullable cryptoutilIdentityDomain.NullableUUID
		wantNil  bool
	}{
		{
			name:     "valid UUID returns pointer",
			nullable: cryptoutilIdentityDomain.NullableUUID{UUID: validID, Valid: true},
			wantNil:  false,
		},
		{
			name:     "invalid UUID returns nil",
			nullable: cryptoutilIdentityDomain.NullableUUID{UUID: googleUuid.Nil, Valid: false},
			wantNil:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ptr := tc.nullable.Ptr()

			if tc.wantNil {
				require.Nil(t, ptr)
			} else {
				require.NotNil(t, ptr)
				require.Equal(t, tc.nullable.UUID, *ptr)
			}
		})
	}
}

func TestNullableUUID_Scan(t *testing.T) {
	t.Parallel()

	validUUID := googleUuid.New()
	validUUIDString := validUUID.String()
	validUUIDBytes := []byte(validUUIDString)

	tests := []struct {
		name      string
		value     any
		wantValid bool
		wantUUID  googleUuid.UUID
		wantErr   bool
	}{
		{
			name:      "scan nil",
			value:     nil,
			wantValid: false,
			wantUUID:  googleUuid.Nil,
			wantErr:   false,
		},
		{
			name:      "scan valid UUID string",
			value:     validUUIDString,
			wantValid: true,
			wantUUID:  validUUID,
			wantErr:   false,
		},
		{
			name:      "scan valid UUID bytes",
			value:     validUUIDBytes,
			wantValid: true,
			wantUUID:  validUUID,
			wantErr:   false,
		},
		{
			name:      "scan invalid UUID string",
			value:     "not-a-uuid",
			wantValid: false,
			wantErr:   true,
		},
		{
			name:      "scan invalid UUID bytes",
			value:     []byte("not-a-uuid-in-bytes"),
			wantValid: false,
			wantErr:   true,
		},
		{
			name:      "scan invalid type",
			value:     123,
			wantValid: false,
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var n cryptoutilIdentityDomain.NullableUUID

			err := n.Scan(tc.value)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantValid, n.Valid)

				if tc.wantValid {
					require.Equal(t, tc.wantUUID, n.UUID)
				}
			}
		})
	}
}

func TestNullableUUID_Value(t *testing.T) {
	t.Parallel()

	validUUID := googleUuid.New()

	tests := []struct {
		name      string
		nullable  cryptoutilIdentityDomain.NullableUUID
		wantValue any
		wantErr   bool
	}{
		{
			name:      "valid UUID returns string",
			nullable:  cryptoutilIdentityDomain.NullableUUID{UUID: validUUID, Valid: true},
			wantValue: validUUID.String(),
			wantErr:   false,
		},
		{
			name:      "invalid UUID returns nil",
			nullable:  cryptoutilIdentityDomain.NullableUUID{Valid: false},
			wantValue: nil,
			wantErr:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			value, err := tc.nullable.Value()

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantValue, value)
			}
		})
	}
}
