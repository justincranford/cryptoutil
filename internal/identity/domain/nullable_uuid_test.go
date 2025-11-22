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

	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
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
