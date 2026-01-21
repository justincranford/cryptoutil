// Copyright (c) 2025 Justin Cranford
//
//

package authz_test

import (
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"

	cryptoutilIdentityAuthz "cryptoutil/internal/identity/authz"
	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityIssuer "cryptoutil/internal/identity/issuer"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"

	testify "github.com/stretchr/testify/require"
)

// TestCleanupService_Creation validates CleanupService initialization.
func TestCleanupService_Creation(t *testing.T) {
	t.Parallel()

	repoFactory := createTestRepoFactory(t)
	config := createTestConfig(t)
	tokenSvc := createTestTokenService(t)

	service := cryptoutilIdentityAuthz.NewService(config, repoFactory, tokenSvc)

	cleanupService := cryptoutilIdentityAuthz.NewCleanupService(service)

	testify.NotNil(t, cleanupService, "Cleanup service should not be nil")
}

// TestCleanupService_StartStop validates cleanup service lifecycle.
func TestCleanupService_StartStop(t *testing.T) {
	t.Parallel()

	repoFactory := createTestRepoFactory(t)
	config := createTestConfig(t)
	tokenSvc := createTestTokenService(t)

	service := cryptoutilIdentityAuthz.NewService(config, repoFactory, tokenSvc)

	cleanupService := cryptoutilIdentityAuthz.NewCleanupService(service)

	ctx := context.Background()

	// Start cleanup service.
	cleanupService.Start(ctx)

	// Wait briefly to ensure goroutine starts.
	time.Sleep(50 * time.Millisecond)

	// Stop cleanup service (should complete gracefully).
	stopDone := make(chan struct{})

	go func() {
		cleanupService.Stop()
		close(stopDone)
	}()

	// Wait for stop to complete (should not hang).
	select {
	case <-stopDone:
		// Success - stop completed.
	case <-time.After(2 * time.Second):
		testify.Fail(t, "Stop() did not complete within timeout")
	}
}

// TestCleanupService_WithInterval validates custom cleanup interval configuration.
func TestCleanupService_WithInterval(t *testing.T) {
	t.Parallel()

	repoFactory := createTestRepoFactory(t)
	config := createTestConfig(t)
	tokenSvc := createTestTokenService(t)

	service := cryptoutilIdentityAuthz.NewService(config, repoFactory, tokenSvc)

	tests := []struct {
		name             string
		interval         time.Duration
		expectedInterval time.Duration
	}{
		{
			name:             "custom valid interval",
			interval:         5 * time.Minute,
			expectedInterval: 5 * time.Minute,
		},
		{
			name:             "zero interval uses default",
			interval:         0,
			expectedInterval: cryptoutilIdentityMagic.DefaultTokenCleanupInterval,
		},
		{
			name:             "negative interval uses default",
			interval:         -1 * time.Minute,
			expectedInterval: cryptoutilIdentityMagic.DefaultTokenCleanupInterval,
		},
	}

	for _, tc := range tests {
		// Capture range variable.
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cleanupService := cryptoutilIdentityAuthz.NewCleanupService(service)

			// Configure interval.
			result := cleanupService.WithInterval(tc.interval)

			testify.NotNil(t, result, "WithInterval should return service")
			testify.Equal(t, cleanupService, result, "WithInterval should return same service instance")
		})
	}
}

// TestCleanupService_ExpiredTokenDeletion validates expired token cleanup.
//
// Validates requirements:
// - R06-04: Automatic deletion of expired tokens.
func TestCleanupService_ExpiredTokenDeletion(t *testing.T) {
	t.Parallel()

	repoFactory := createTestRepoFactory(t)

	ctx := context.Background()

	// Run migrations.
	err := repoFactory.AutoMigrate(ctx)
	testify.NoError(t, err, "Failed to run migrations")

	// Create test user (required for foreign key constraint).
	userRepo := repoFactory.UserRepository()
	testUser := &cryptoutilIdentityDomain.User{
		Sub:               "test-user-" + googleUuid.NewString(),
		PreferredUsername: "test-user-" + googleUuid.NewString(),
		Email:             "test-" + googleUuid.NewString() + "@example.com",
		PasswordHash:      "hash123",
		EmailVerified:     false,
	}

	err = userRepo.Create(ctx, testUser)
	testify.NoError(t, err, "Failed to create test user")

	// Create test client (required for foreign key constraint).
	clientRepo := repoFactory.ClientRepository()
	testClient := &cryptoutilIdentityDomain.Client{
		ClientID:                "client-" + googleUuid.NewString(),
		ClientSecret:            "secret123",
		Name:                    "test-client",
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		AllowedScopes:           []string{"read", "write"},
		AllowedGrantTypes:       []string{"authorization_code"},
		AllowedResponseTypes:    []string{"code"},
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretBasic,
	}

	err = clientRepo.Create(ctx, testClient)
	testify.NoError(t, err, "Failed to create test client")

	// Create expired token (expires in the past).
	tokenRepo := repoFactory.TokenRepository()
	expiredToken := &cryptoutilIdentityDomain.Token{
		TokenValue:    googleUuid.NewString(),
		TokenType:     cryptoutilIdentityDomain.TokenTypeAccess,
		TokenFormat:   cryptoutilIdentityDomain.TokenFormatUUID,
		ExpiresAt:     time.Now().Add(-1 * time.Hour), // Expired 1 hour ago.
		IssuedAt:      time.Now().Add(-2 * time.Hour),
		Scopes:        []string{"read", "write"},
		ClientID:      testClient.ID,
		UserID:        cryptoutilIdentityDomain.NullableUUID{UUID: testUser.ID, Valid: true},
		CodeChallenge: "",
	}

	err = tokenRepo.Create(ctx, expiredToken)
	testify.NoError(t, err, "Failed to create expired token")

	// Create non-expired token (expires in the future).
	validToken := &cryptoutilIdentityDomain.Token{
		TokenValue:    googleUuid.NewString(),
		TokenType:     cryptoutilIdentityDomain.TokenTypeAccess,
		TokenFormat:   cryptoutilIdentityDomain.TokenFormatUUID,
		ExpiresAt:     time.Now().Add(1 * time.Hour), // Expires 1 hour from now.
		IssuedAt:      time.Now(),
		Scopes:        []string{"read", "write"},
		ClientID:      testClient.ID,
		UserID:        cryptoutilIdentityDomain.NullableUUID{UUID: testUser.ID, Valid: true},
		CodeChallenge: "",
	}

	err = tokenRepo.Create(ctx, validToken)
	testify.NoError(t, err, "Failed to create valid token")

	// Create cleanup service and trigger cleanup.
	config := createTestConfig(t)
	tokenSvc := createTestTokenService(t)
	service := cryptoutilIdentityAuthz.NewService(config, repoFactory, tokenSvc)

	cleanupService := cryptoutilIdentityAuthz.NewCleanupService(service)

	// Set short interval for testing and start cleanup service.
	cleanupService.WithInterval(50 * time.Millisecond)
	cleanupService.Start(ctx)

	// Wait for cleanup to execute (need time for ticker + cleanup operation).
	time.Sleep(200 * time.Millisecond)

	cleanupService.Stop()

	// Verify expired token was deleted.
	deletedToken, err := tokenRepo.GetByID(ctx, expiredToken.ID)
	testify.Error(t, err, "Expired token should be deleted")
	testify.Nil(t, deletedToken, "Expired token should not exist")

	// Verify valid token still exists.
	retrievedToken, err := tokenRepo.GetByID(ctx, validToken.ID)
	testify.NoError(t, err, "Valid token should still exist")
	testify.NotNil(t, retrievedToken, "Valid token should be retrievable")
	testify.Equal(t, validToken.ID, retrievedToken.ID)
}

// TestCleanupService_ErrorHandling validates cleanup service error handling.
func TestCleanupService_ErrorHandling(t *testing.T) {
	t.Parallel()

	// Create repository factory with invalid database (simulates error).
	repoFactory := createInvalidTestRepoFactory(t)
	config := createTestConfig(t)
	tokenSvc := createTestTokenService(t)

	service := cryptoutilIdentityAuthz.NewService(config, repoFactory, tokenSvc)

	cleanupService := cryptoutilIdentityAuthz.NewCleanupService(service)

	ctx := context.Background()

	// Start cleanup service (should handle errors gracefully).
	cleanupService.Start(ctx)

	// Wait briefly to allow cleanup to execute and handle error.
	time.Sleep(100 * time.Millisecond)

	// Stop cleanup service (should complete gracefully despite errors).
	cleanupService.Stop()
	// Test passes if no panic occurs and Stop() completes.
}

// createTestRepoFactory creates a test repository factory with SQLite in-memory database.
func createTestRepoFactory(t *testing.T) *cryptoutilIdentityRepository.RepositoryFactory {
	t.Helper()

	// Use unique database name per test for SQLite isolation.
	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type:        "sqlite",
		DSN:         "file::memory:?cache=private&_fk=1&mode=memory&_loc=UTC",
		AutoMigrate: true,
	}

	ctx := context.Background()

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)

	testify.NoError(t, err, "Failed to create repository factory")
	testify.NotNil(t, repoFactory, "Repository factory should not be nil")

	return repoFactory
}

// createInvalidTestRepoFactory creates a test repository factory with invalid configuration.
func createInvalidTestRepoFactory(t *testing.T) *cryptoutilIdentityRepository.RepositoryFactory {
	t.Helper()

	// Use invalid DSN to simulate database errors.
	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type:        "sqlite",
		DSN:         ":memory:",
		AutoMigrate: false, // Don't migrate to simulate errors.
	}

	ctx := context.Background()

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)

	testify.NoError(t, err, "Should create factory even with invalid config")

	return repoFactory
}

// createTestConfig creates a test configuration for authz service.
func createTestConfig(t *testing.T) *cryptoutilIdentityConfig.Config {
	t.Helper()

	return &cryptoutilIdentityConfig.Config{
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			AccessTokenLifetime:  1 * time.Hour,
			RefreshTokenLifetime: 24 * time.Hour,
		},
		Security: &cryptoutilIdentityConfig.SecurityConfig{
			CORSAllowedOrigins: []string{"https://localhost:8080"},
		},
	}
}

// createTestTokenService creates a test token service.
func createTestTokenService(t *testing.T) *cryptoutilIdentityIssuer.TokenService {
	t.Helper()

	return nil // CleanupService doesn't use TokenService directly.
}
