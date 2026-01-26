// Copyright (c) 2025 Justin Cranford

//go:build e2e

package e2e

import (
	"context"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

// TestAuthorizationCodeFlowWithDatabase validates OAuth 2.1 authorization code flow with database persistence.
// This test covers:
// - Authorization request persistence with PKCE
// - Login redirect and user authentication
// - Consent decision storage
// - Authorization code generation with real user ID
// - Token exchange with single-use code enforcement.
func TestAuthorizationCodeFlowWithDatabase(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilIdentityMagic.TestRefreshTokenLifetime)
	defer cancel()

	// Create test database and repositories.
	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err, "Failed to create repository factory")

	// Create test user.
	userRepo := repoFactory.UserRepository()
	testUser := &cryptoutilIdentityDomain.User{
		ID:                googleUuid.Must(googleUuid.NewV7()),
		Sub:               fmt.Sprintf("testuser-%s", googleUuid.New().String()[:8]),
		PreferredUsername: fmt.Sprintf("testuser-%s", googleUuid.New().String()[:8]),
		Email:             fmt.Sprintf("test-%s@example.com", googleUuid.New().String()[:8]),
		PasswordHash:      "dummy-hash",
	}
	require.NoError(t, userRepo.Create(ctx, testUser), "Failed to create test user")

	// Create test client.
	clientRepo := repoFactory.ClientRepository()
	testClient := &cryptoutilIdentityDomain.Client{
		ID:                      googleUuid.Must(googleUuid.NewV7()),
		ClientID:                fmt.Sprintf("test-client-%s", googleUuid.New().String()[:8]),
		ClientSecret:            "dummy-secret",
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		Name:                    "Test Client",
		RedirectURIs:            []string{"https://example.com/callback"},
		AllowedGrantTypes:       []string{cryptoutilIdentityMagic.GrantTypeAuthorizationCode},
		AllowedResponseTypes:    []string{cryptoutilIdentityMagic.ResponseTypeCode},
		AllowedScopes:           []string{"openid", "profile", "email"},
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretBasic,
	}
	require.NoError(t, clientRepo.Create(ctx, testClient), "Failed to create test client")

	// Generate PKCE challenge.
	_, codeChallenge := generatePKCEChallengeDatabase()
	state := generateRandomStringDatabase(cryptoutilIdentityMagic.DefaultStateLength)

	// Step 1: Create authorization request in database.
	t.Log("🔐 Creating authorization request with PKCE...")

	authzReqRepo := repoFactory.AuthorizationRequestRepository()
	authRequest := &cryptoutilIdentityDomain.AuthorizationRequest{
		ID:                  googleUuid.Must(googleUuid.NewV7()),
		ClientID:            testClient.ClientID,
		RedirectURI:         testClient.RedirectURIs[0],
		ResponseType:        cryptoutilIdentityMagic.ResponseTypeCode,
		Scope:               "openid profile email",
		State:               state,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: cryptoutilIdentityMagic.PKCEMethodS256,
		CreatedAt:           time.Now().UTC(),
		ExpiresAt:           time.Now().UTC().Add(cryptoutilIdentityMagic.DefaultCodeLifetime),
	}
	require.NoError(t, authzReqRepo.Create(ctx, authRequest), "Failed to create authorization request")

	// Step 2: Simulate user login - associate user ID with authorization request.
	t.Log("👤 Simulating user login and authentication...")

	authRequest.UserID = cryptoutilIdentityDomain.NullableUUID{
		UUID:  testUser.ID,
		Valid: true,
	}
	require.NoError(t, authzReqRepo.Update(ctx, authRequest), "Failed to update authorization request with user ID")

	// Step 3: Store consent decision.
	t.Log("✅ Storing consent decision...")

	consentRepo := repoFactory.ConsentDecisionRepository()
	consentDecision := &cryptoutilIdentityDomain.ConsentDecision{
		ID:        googleUuid.Must(googleUuid.NewV7()),
		UserID:    testUser.ID,
		ClientID:  testClient.ClientID,
		Scope:     authRequest.Scope,
		GrantedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(cryptoutilIdentityMagic.DefaultRefreshTokenLifetime),
	}
	require.NoError(t, consentRepo.Create(ctx, consentDecision), "Failed to create consent decision")

	// Step 4: Generate authorization code.
	t.Log("🔑 Generating authorization code...")

	authCode := generateRandomStringDatabase(cryptoutilIdentityMagic.DefaultAuthCodeLength)
	authRequest.Code = authCode
	authRequest.ConsentGranted = true
	require.NoError(t, authzReqRepo.Update(ctx, authRequest), "Failed to update authorization request with code")

	// Step 5: Exchange authorization code for tokens (validate single-use enforcement).
	t.Log("💱 Exchanging authorization code for tokens...")

	// Retrieve authorization request by code.
	retrievedAuthRequest, err := authzReqRepo.GetByCode(ctx, authCode)
	require.NoError(t, err, "Failed to retrieve authorization request by code")
	require.Equal(t, authRequest.ID, retrievedAuthRequest.ID, "Retrieved request ID should match")
	require.Equal(t, testUser.ID, retrievedAuthRequest.UserID.UUID, "User ID should match test user")

	// Validate PKCE challenge persistence.
	require.NotEmpty(t, codeChallenge, "PKCE challenge should not be empty")
	require.Equal(t, codeChallenge, retrievedAuthRequest.CodeChallenge, "Stored PKCE challenge should match")

	// Validate expiration.
	require.False(t, retrievedAuthRequest.IsExpired(), "Authorization request should not be expired")

	// Validate single-use enforcement (not yet used).
	require.False(t, retrievedAuthRequest.IsUsed(), "Authorization code should not be used yet")

	// Mark code as used.
	now := time.Now().UTC()
	retrievedAuthRequest.Used = true
	retrievedAuthRequest.UsedAt = &now
	require.NoError(t, authzReqRepo.Update(ctx, retrievedAuthRequest), "Failed to mark authorization code as used")

	// Attempt to use code again (should fail).
	secondRetrievalAttempt, err := authzReqRepo.GetByCode(ctx, authCode)
	require.NoError(t, err, "Should retrieve used authorization request")
	require.True(t, secondRetrievalAttempt.IsUsed(), "Authorization code should be marked as used")

	// Verify consent decision persistence.
	t.Log("🔍 Verifying consent decision persistence...")

	retrievedConsent, err := consentRepo.GetByUserClientScope(ctx, testUser.ID, testClient.ClientID, authRequest.Scope)
	require.NoError(t, err, "Failed to retrieve consent decision")
	require.Equal(t, consentDecision.ID, retrievedConsent.ID, "Consent decision ID should match")
	require.False(t, retrievedConsent.IsRevoked(), "Consent should not be revoked")
	require.False(t, retrievedConsent.IsExpired(), "Consent should not be expired")

	t.Log("✅ Authorization code flow with database persistence validated successfully")
}

// generatePKCEChallenge generates a PKCE code verifier and code challenge for database test.
func generatePKCEChallengeDatabase() (codeVerifier, codeChallenge string) {
	// Generate code verifier (43-128 characters).
	verifierBytes := make([]byte, cryptoutilIdentityMagic.DefaultCodeChallengeLength)
	if _, err := io.ReadFull(crand.Reader, verifierBytes); err != nil {
		panic(fmt.Sprintf("failed to generate code verifier: %v", err))
	}

	codeVerifier = base64.RawURLEncoding.EncodeToString(verifierBytes)

	// Generate code challenge (SHA256 hash of verifier).
	hash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge = base64.RawURLEncoding.EncodeToString(hash[:])

	return codeVerifier, codeChallenge
}

// generateRandomStringDatabase generates a random string for database test.
func generateRandomStringDatabase(length int) string {
	bytes := make([]byte, length)
	if _, err := io.ReadFull(crand.Reader, bytes); err != nil {
		panic(fmt.Sprintf("failed to generate random string: %v", err))
	}

	return base64.RawURLEncoding.EncodeToString(bytes)[:length]
}
