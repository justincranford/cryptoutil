// Copyright (c) 2025 Justin Cranford
//
//

//go:build e2e

package e2e

import (
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestOIDCCoreEndpoints(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create in-memory SQLite database configuration.
	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: cryptoutilSharedMagic.TestDatabaseSQLite,
		DSN:  cryptoutilSharedMagic.SQLiteInMemoryDSN,
	}

	// Initialize repository factory.
	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err, "Failed to create repository factory")

	// Create test user with OIDC claims.
	userRepo := repoFactory.UserRepository()
	testUser := &cryptoutilIdentityDomain.User{
		ID:                googleUuid.Must(googleUuid.NewV7()),
		Sub:               "testuser_" + googleUuid.New().String(),
		PreferredUsername: "testuser",
		PasswordHash:      "$2a$10$abc123...", // Placeholder hash.
		Name:              "Test User",
		GivenName:         "Test",
		FamilyName:        "User",
		Email:             "test@example.com",
		EmailVerified:     true,
		PhoneNumber:       "+1234567890",
		PhoneVerified:     false,
		Address: &cryptoutilIdentityDomain.Address{
			Formatted:     "123 Test St, Test City, TS 12345, USA",
			StreetAddress: "123 Test St",
			Locality:      "Test City",
			Region:        "TS",
			PostalCode:    "12345",
			Country:       "USA",
		},
		Enabled:   true,
		Locked:    false,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	err = userRepo.Create(ctx, testUser)
	require.NoError(t, err, "Failed to create test user")

	// Create test OAuth client.
	clientRepo := repoFactory.ClientRepository()
	testClient := &cryptoutilIdentityDomain.Client{
		ID:       googleUuid.Must(googleUuid.NewV7()),
		Name:     "Test Client",
		ClientID: "test_client_" + googleUuid.New().String(),
		RedirectURIs: []string{
			"https://localhost:8080/callback",
		},
		AllowedGrantTypes: []string{
			cryptoutilSharedMagic.GrantTypeAuthorizationCode,
			cryptoutilSharedMagic.GrantTypeRefreshToken,
		},
		AllowedResponseTypes: []string{
			cryptoutilSharedMagic.ResponseTypeCode,
		},
		AllowedScopes: []string{
			cryptoutilSharedMagic.ScopeOpenID,
			cryptoutilSharedMagic.ClaimProfile,
			cryptoutilSharedMagic.ClaimEmail,
			cryptoutilSharedMagic.ClaimAddress,
			cryptoutilSharedMagic.ScopePhone,
		},
		TokenEndpointAuthMethod: cryptoutilSharedMagic.ClientAuthMethodSecretBasic,
		Enabled:                 boolPtr(true),
		CreatedAt:               time.Now().UTC(),
		UpdatedAt:               time.Now().UTC(),
	}

	err = clientRepo.Create(ctx, testClient)
	require.NoError(t, err, "Failed to create test client")

	// Create authorization request with user ID (simulating completed login).
	authRequestRepo := repoFactory.AuthorizationRequestRepository()
	authRequest := &cryptoutilIdentityDomain.AuthorizationRequest{
		ID:                  googleUuid.Must(googleUuid.NewV7()),
		ClientID:            testClient.ClientID,
		RedirectURI:         testClient.RedirectURIs[0],
		Scope:               "openid profile email address phone",
		State:               "test_state",
		ResponseType:        cryptoutilSharedMagic.ResponseTypeCode,
		CodeChallenge:       "test_code_challenge",
		CodeChallengeMethod: cryptoutilSharedMagic.PKCEMethodS256,
		UserID: cryptoutilIdentityDomain.NullableUUID{
			UUID:  testUser.ID,
			Valid: true,
		},
		CreatedAt:      time.Now().UTC(),
		ExpiresAt:      time.Now().UTC().Add(cryptoutilSharedMagic.DefaultCodeLifetime),
		ConsentGranted: false,
	}

	err = authRequestRepo.Create(ctx, authRequest)
	require.NoError(t, err, "Failed to create authorization request")

	// Create session (simulating successful login).
	sessionRepo := repoFactory.SessionRepository()
	session := &cryptoutilIdentityDomain.Session{
		UserID:                testUser.ID,
		SessionID:             "test_session_" + googleUuid.New().String(),
		IPAddress:             cryptoutilSharedMagic.IPv4Loopback,
		UserAgent:             "test-agent",
		IssuedAt:              time.Now().UTC(),
		ExpiresAt:             time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days * time.Minute),
		LastSeenAt:            time.Now().UTC(),
		Active:                boolPtr(true),
		AuthenticationMethods: []string{cryptoutilSharedMagic.AuthMethodUsernamePassword},
		AuthenticationTime:    time.Now().UTC(),
	}

	err = sessionRepo.Create(ctx, session)
	require.NoError(t, err, "Failed to create session")

	// Test consent decision creation.
	consentRepo := repoFactory.ConsentDecisionRepository()
	consentDecision := &cryptoutilIdentityDomain.ConsentDecision{
		ID:        googleUuid.Must(googleUuid.NewV7()),
		UserID:    testUser.ID,
		ClientID:  testClient.ClientID,
		Scope:     authRequest.Scope,
		GrantedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(cryptoutilSharedMagic.DefaultRefreshTokenLifetime),
	}

	err = consentRepo.Create(ctx, consentDecision)
	require.NoError(t, err, "Failed to create consent decision")

	// Verify consent can be retrieved.
	retrievedConsent, err := consentRepo.GetByUserClientScope(ctx, testUser.ID, testClient.ClientID, authRequest.Scope)
	require.NoError(t, err, "Failed to retrieve consent decision")
	require.NotNil(t, retrievedConsent, "Consent decision should exist")
	require.Equal(t, testUser.ID, retrievedConsent.UserID, "Consent user ID mismatch")
	require.Equal(t, testClient.ClientID, retrievedConsent.ClientID, "Consent client ID mismatch")
	require.False(t, retrievedConsent.IsRevoked(), "Consent should not be revoked")
	require.False(t, retrievedConsent.IsExpired(), "Consent should not be expired")

	// Update authorization request with authorization code (simulating consent approval).
	authRequest.Code = "test_auth_code_" + googleUuid.New().String()
	authRequest.ConsentGranted = true

	err = authRequestRepo.Update(ctx, authRequest)
	require.NoError(t, err, "Failed to update authorization request with code")

	// Verify authorization code retrieval.
	retrievedAuthRequest, err := authRequestRepo.GetByCode(ctx, authRequest.Code)
	require.NoError(t, err, "Failed to retrieve authorization request by code")
	require.NotNil(t, retrievedAuthRequest, "Authorization request should exist")
	require.Equal(t, authRequest.Code, retrievedAuthRequest.Code, "Authorization code mismatch")
	require.True(t, retrievedAuthRequest.ConsentGranted.Bool(), "Consent should be granted")
	require.False(t, retrievedAuthRequest.IsExpired(), "Authorization request should not be expired")
	require.False(t, retrievedAuthRequest.IsUsed(), "Authorization request should not be used yet")

	// Mark authorization code as used (simulating token exchange).
	now := time.Now().UTC()
	retrievedAuthRequest.Used = true
	retrievedAuthRequest.UsedAt = &now

	err = authRequestRepo.Update(ctx, retrievedAuthRequest)
	require.NoError(t, err, "Failed to mark authorization code as used")

	// Verify single-use enforcement.
	usedAuthRequest, err := authRequestRepo.GetByCode(ctx, authRequest.Code)
	require.NoError(t, err, "Failed to retrieve used authorization request")
	require.True(t, usedAuthRequest.IsUsed(), "Authorization code should be marked as used")

	// Test session deletion (simulating logout).
	err = sessionRepo.Delete(ctx, session.ID)
	require.NoError(t, err, "Failed to delete session")

	// Verify session no longer exists.
	_, err = sessionRepo.GetBySessionID(ctx, session.SessionID)
	require.Error(t, err, "Session should not exist after deletion")
}
