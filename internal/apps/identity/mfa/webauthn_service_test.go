// Copyright (c) 2025 Justin Cranford

package mfa_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"

	cryptoutilIdentityMfa "cryptoutil/internal/apps/identity/mfa"
)

func createWebAuthnTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	sqlDB, err := sql.Open("sqlite", "file::memory:?cache=private")
	require.NoError(t, err)

	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	err = db.AutoMigrate(&cryptoutilIdentityMfa.WebAuthnCredential{}, &cryptoutilIdentityMfa.WebAuthnSession{})
	require.NoError(t, err)

	return db
}

func createWebAuthnConfig() cryptoutilIdentityMfa.WebAuthnConfig {
	return cryptoutilIdentityMfa.WebAuthnConfig{
		RPDisplayName: "Test Application",
		RPID:          "localhost",
		RPOrigins:     []string{"https://localhost:8080"},
	}
}

func TestNewWebAuthnService_Success(t *testing.T) {
	t.Parallel()

	db := createWebAuthnTestDB(t)
	config := createWebAuthnConfig()

	service, err := cryptoutilIdentityMfa.NewWebAuthnService(db, config)
	require.NoError(t, err)
	require.NotNil(t, service)
}

func TestNewWebAuthnService_InvalidConfig(t *testing.T) {
	t.Parallel()

	db := createWebAuthnTestDB(t)
	config := cryptoutilIdentityMfa.WebAuthnConfig{
		RPDisplayName: "",
		RPID:          "",
		RPOrigins:     []string{},
	}

	service, err := cryptoutilIdentityMfa.NewWebAuthnService(db, config)
	require.Error(t, err)
	require.Nil(t, service)
}

func TestWebAuthnUser_Interface(t *testing.T) {
	t.Parallel()

	userID := googleUuid.Must(googleUuid.NewV7())
	user := &cryptoutilIdentityMfa.WebAuthnUser{
		ID:          userID,
		Name:        "testuser",
		DisplayName: "Test User",
		Credentials: []cryptoutilIdentityMfa.WebAuthnCredential{},
	}

	require.Equal(t, userID[:], user.WebAuthnID())
	require.Equal(t, "testuser", user.WebAuthnName())
	require.Equal(t, "Test User", user.WebAuthnDisplayName())
	require.Empty(t, user.WebAuthnCredentials())
}

func TestWebAuthnService_BeginRegistration(t *testing.T) {
	t.Parallel()

	db := createWebAuthnTestDB(t)
	config := createWebAuthnConfig()

	service, err := cryptoutilIdentityMfa.NewWebAuthnService(db, config)
	require.NoError(t, err)

	userID := googleUuid.Must(googleUuid.NewV7())
	user := &cryptoutilIdentityMfa.WebAuthnUser{
		ID:          userID,
		Name:        "testuser",
		DisplayName: "Test User",
	}

	options, sessionID, err := service.BeginRegistration(context.Background(), user)
	require.NoError(t, err)
	require.NotNil(t, options)
	require.NotEqual(t, googleUuid.Nil, sessionID)

	require.NotNil(t, options.Response)
	require.NotEmpty(t, options.Response.Challenge)
}

func TestWebAuthnService_GetCredentials_Empty(t *testing.T) {
	t.Parallel()

	db := createWebAuthnTestDB(t)
	config := createWebAuthnConfig()

	service, err := cryptoutilIdentityMfa.NewWebAuthnService(db, config)
	require.NoError(t, err)

	userID := googleUuid.Must(googleUuid.NewV7())
	credentials, err := service.GetCredentials(context.Background(), userID)
	require.NoError(t, err)
	require.Empty(t, credentials)
}

func TestWebAuthnService_DeleteCredential_NotFound(t *testing.T) {
	t.Parallel()

	db := createWebAuthnTestDB(t)
	config := createWebAuthnConfig()

	service, err := cryptoutilIdentityMfa.NewWebAuthnService(db, config)
	require.NoError(t, err)

	userID := googleUuid.Must(googleUuid.NewV7())
	credentialID := googleUuid.Must(googleUuid.NewV7())

	err = service.DeleteCredential(context.Background(), userID, credentialID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestWebAuthnService_CleanupExpiredSessions(t *testing.T) {
	t.Parallel()

	db := createWebAuthnTestDB(t)
	config := createWebAuthnConfig()

	service, err := cryptoutilIdentityMfa.NewWebAuthnService(db, config)
	require.NoError(t, err)

	err = service.CleanupExpiredSessions(context.Background())
	require.NoError(t, err)
}

func TestWebAuthnCredential_BeforeCreate(t *testing.T) {
	t.Parallel()

	db := createWebAuthnTestDB(t)

	credential := &cryptoutilIdentityMfa.WebAuthnCredential{
		UserID:          googleUuid.Must(googleUuid.NewV7()),
		CredentialID:    []byte("test-credential-id"),
		PublicKey:       []byte("test-public-key"),
		AttestationType: "none",
		DisplayName:     "Test Key",
	}

	err := db.Create(credential).Error
	require.NoError(t, err)
	require.NotEqual(t, googleUuid.Nil, credential.ID)
	require.False(t, credential.CreatedAt.IsZero())
	require.False(t, credential.UpdatedAt.IsZero())
}

func TestWebAuthnSession_BeforeCreate(t *testing.T) {
	t.Parallel()

	db := createWebAuthnTestDB(t)

	session := &cryptoutilIdentityMfa.WebAuthnSession{
		UserID:       googleUuid.Must(googleUuid.NewV7()),
		SessionData:  []byte("{}"),
		CeremonyType: string(cryptoutilIdentityMfa.WebAuthnCeremonyRegistration),
		ExpiresAt:    time.Now().UTC().Add(5 * time.Minute),
	}

	err := db.Create(session).Error
	require.NoError(t, err)
	require.NotEqual(t, googleUuid.Nil, session.ID)
	require.False(t, session.CreatedAt.IsZero())
}

func TestWebAuthnSession_IsExpired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		expiresAt time.Duration
		expected  bool
	}{
		{
			name:      "not expired - future",
			expiresAt: time.Hour,
			expected:  false,
		},
		{
			name:      "expired - past",
			expiresAt: -time.Hour,
			expected:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			session := &cryptoutilIdentityMfa.WebAuthnSession{
				UserID:       googleUuid.Must(googleUuid.NewV7()),
				SessionData:  []byte("{}"),
				CeremonyType: string(cryptoutilIdentityMfa.WebAuthnCeremonyRegistration),
				ExpiresAt:    time.Now().UTC().Add(tc.expiresAt),
			}

			require.Equal(t, tc.expected, session.IsExpired())
		})
	}
}

func TestWebAuthnCeremonyType_Constants(t *testing.T) {
	t.Parallel()

	require.Equal(t, cryptoutilIdentityMfa.WebAuthnCeremonyType("registration"), cryptoutilIdentityMfa.WebAuthnCeremonyRegistration)
	require.Equal(t, cryptoutilIdentityMfa.WebAuthnCeremonyType("authentication"), cryptoutilIdentityMfa.WebAuthnCeremonyAuthentication)
}
