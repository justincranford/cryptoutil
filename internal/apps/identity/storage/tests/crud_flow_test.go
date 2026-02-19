// Copyright (c) 2025 Justin Cranford
//
//

package tests

import (
	"context"
	"testing"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite" // Register CGO-free SQLite driver
)

func TestAuthProfileRepositoryCRUD(t *testing.T) {
	t.Parallel()

	if !isCGOAvailable() {
		t.Skip("CGO not available, skipping SQLite tests")
	}

	ctx := context.Background()

	repoFactory := setupTestRepositoryFactory(ctx, t)

	defer func() { _ = repoFactory.Close() }() //nolint:errcheck // Test cleanup //nolint:errcheck // Test cleanup

	profileRepo := repoFactory.AuthProfileRepository()

	// Test Create
	profile := &cryptoutilIdentityDomain.AuthProfile{
		Name:        "test-auth-profile",
		Description: "Test auth profile",
	}

	err := profileRepo.Create(ctx, profile)
	require.NoError(t, err)
	require.NotEmpty(t, profile.ID)

	// Test GetByID
	retrievedProfile, err := profileRepo.GetByID(ctx, profile.ID)
	require.NoError(t, err)
	require.Equal(t, profile.ID, retrievedProfile.ID)

	// Test GetByName
	profileByName, err := profileRepo.GetByName(ctx, profile.Name)
	require.NoError(t, err)
	require.Equal(t, profile.Name, profileByName.Name)

	// Test Update
	updatedProfile := *retrievedProfile
	updatedProfile.Description = "Updated test auth profile"
	err = profileRepo.Update(ctx, &updatedProfile)
	require.NoError(t, err)

	// Test Delete
	err = profileRepo.Delete(ctx, profile.ID)
	require.NoError(t, err)
}

func TestMFAFactorRepositoryCRUD(t *testing.T) {
	t.Parallel()

	if !isCGOAvailable() {
		t.Skip("CGO not available, skipping SQLite tests")
	}

	ctx := context.Background()

	repoFactory := setupTestRepositoryFactory(ctx, t)

	defer func() { _ = repoFactory.Close() }() //nolint:errcheck // Test cleanup //nolint:errcheck // Test cleanup

	factorRepo := repoFactory.MFAFactorRepository()

	// Create an auth profile first
	authProfile := &cryptoutilIdentityDomain.AuthProfile{
		Name: "test-auth-profile",
	}
	authProfileRepo := repoFactory.AuthProfileRepository()
	err := authProfileRepo.Create(ctx, authProfile)
	require.NoError(t, err)

	// Test Create
	factor := &cryptoutilIdentityDomain.MFAFactor{
		Name:          "test-factor",
		FactorType:    cryptoutilIdentityDomain.MFAFactorTypeTOTP,
		Order:         1,
		Required:      true,
		AuthProfileID: authProfile.ID,
	}

	err = factorRepo.Create(ctx, factor)
	require.NoError(t, err)
	require.NotEmpty(t, factor.ID)

	// Test GetByID
	retrievedFactor, err := factorRepo.GetByID(ctx, factor.ID)
	require.NoError(t, err)
	require.Equal(t, factor.ID, retrievedFactor.ID)

	// Test GetByAuthProfileID
	factorsByProfile, err := factorRepo.GetByAuthProfileID(ctx, authProfile.ID)
	require.NoError(t, err)
	require.Len(t, factorsByProfile, 1)
	require.Equal(t, factor.ID, factorsByProfile[0].ID)

	// Test Update
	updatedFactor := *retrievedFactor
	updatedFactor.Required = false
	err = factorRepo.Update(ctx, &updatedFactor)
	require.NoError(t, err)

	// Test Delete
	err = factorRepo.Delete(ctx, factor.ID)
	require.NoError(t, err)
}

// Helper function to set up test repository factory.
func setupTestRepositoryFactory(ctx context.Context, t *testing.T) *cryptoutilIdentityRepository.RepositoryFactory {
	t.Helper()

	// Use unique file-based in-memory database per test to prevent data pollution between parallel tests.
	uuidSuffix := googleUuid.Must(googleUuid.NewV7()).String()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type:            "sqlite",
		DSN:             "file:" + uuidSuffix + ".db?mode=memory&cache=shared",
		MaxOpenConns:    5,
		MaxIdleConns:    5,
		ConnMaxLifetime: 0,
		ConnMaxIdleTime: 0,
		AutoMigrate:     true,
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	// Run migrations
	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err)

	return repoFactory
}
