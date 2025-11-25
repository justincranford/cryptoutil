// Copyright (c) 2025 Justin Cranford
//
//

package authz

import (
	"context"

	"cryptoutil/internal/identity/authz/clientauth"

	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityIssuer "cryptoutil/internal/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
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
	return &Service{
		config:       config,
		repoFactory:  repoFactory,
		tokenSvc:     tokenSvc,
		clientAuth:   clientauth.NewRegistry(repoFactory, config),
		authReqStore: NewInMemoryAuthorizationRequestStore(),
	}
}

// Start starts the authorization server.
func (s *Service) Start(ctx context.Context) error {
	// Validate database connectivity at startup.
	db := s.repoFactory.DB()
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return err
	}

	return nil
}

// Stop stops the authorization server.
func (s *Service) Stop(ctx context.Context) error {
	// Clean up expired tokens.
	tokenRepo := s.repoFactory.TokenRepository()
	if err := tokenRepo.DeleteExpired(ctx); err != nil {
		return err
	}

	// Close database connections gracefully.
	db := s.repoFactory.DB()
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}
