// Copyright (c) 2025 Justin Cranford
//
//

package authz

import (
	"context"
	"fmt"

	"cryptoutil/internal/identity/authz/clientauth"

	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityIssuer "cryptoutil/internal/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
	cryptoutilIdentityRotation "cryptoutil/internal/identity/rotation"
)

// Service provides OAuth 2.1 authorization server functionality.
type Service struct {
	config       *cryptoutilIdentityConfig.Config
	repoFactory  *cryptoutilIdentityRepository.RepositoryFactory
	tokenSvc     *cryptoutilIdentityIssuer.TokenService
	clientAuth   *clientauth.Registry
	authReqStore AuthorizationRequestStore
}

// NewService creates a new authorization server service.
func NewService(
	config *cryptoutilIdentityConfig.Config,
	repoFactory *cryptoutilIdentityRepository.RepositoryFactory,
	tokenSvc *cryptoutilIdentityIssuer.TokenService,
) *Service {
	// Create rotation service for multi-secret authentication
	rotationService := cryptoutilIdentityRotation.NewSecretRotationService(repoFactory.DB())

	return &Service{
		config:       config,
		repoFactory:  repoFactory,
		tokenSvc:     tokenSvc,
		clientAuth:   clientauth.NewRegistry(repoFactory, config, rotationService),
		authReqStore: NewInMemoryAuthorizationRequestStore(),
	}
}

// Start starts the authorization server.
func (s *Service) Start(ctx context.Context) error {
	// Validate database connectivity at startup.
	db := s.repoFactory.DB()

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection for startup validation: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database during startup: %w", err)
	}

	return nil
}

// Stop stops the authorization server.
func (s *Service) Stop(ctx context.Context) error {
	// Clean up expired tokens.
	tokenRepo := s.repoFactory.TokenRepository()

	if err := tokenRepo.DeleteExpired(ctx); err != nil {
		return fmt.Errorf("failed to delete expired tokens during shutdown: %w", err)
	}

	// Close database connections gracefully.
	db := s.repoFactory.DB()

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection for shutdown: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	return nil
}
