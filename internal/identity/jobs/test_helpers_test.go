// Copyright (c) 2025 Justin Cranford
//
//

package jobs

import (
	"context"
	"testing"

	testify "github.com/stretchr/testify/require"

	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"

	_ "modernc.org/sqlite" // Register CGO-free SQLite driver
)

// createTestRepoFactory creates a repository factory for testing.
func createTestRepoFactory(t *testing.T) *cryptoutilIdentityRepository.RepositoryFactory {
	t.Helper()

	config := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(context.Background(), config)
	testify.NoError(t, err, "Failed to create repository factory")

	return repoFactory
}
