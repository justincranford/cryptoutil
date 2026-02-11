// Copyright (c) 2025 Justin Cranford

package idp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

func TestNewService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  "file::memory:?cache=shared",
	}

	config := &cryptoutilIdentityConfig.Config{}
	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	defer func() {
		_ = repoFactory.Close() //nolint:errcheck // Test cleanup
	}()

	tokenSvc := &cryptoutilIdentityIssuer.TokenService{}

	service := NewService(config, repoFactory, tokenSvc)

	require.NotNil(t, service)
	require.NotNil(t, service.config)
	require.NotNil(t, service.repoFactory)
	require.NotNil(t, service.tokenSvc)
	require.NotNil(t, service.authProfiles)
	require.NotNil(t, service.templates)
}
