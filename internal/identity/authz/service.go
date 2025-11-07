package authz

import (
	"context"

	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityIssuer "cryptoutil/internal/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

// Service provides OAuth 2.1 authorization server functionality.
type Service struct {
	config      *cryptoutilIdentityConfig.Config
	repoFactory *cryptoutilIdentityRepository.RepositoryFactory
	tokenSvc    *cryptoutilIdentityIssuer.TokenService
}

// NewService creates a new authorization server service.
func NewService(
	config *cryptoutilIdentityConfig.Config,
	repoFactory *cryptoutilIdentityRepository.RepositoryFactory,
	tokenSvc *cryptoutilIdentityIssuer.TokenService,
) *Service {
	return &Service{
		config:      config,
		repoFactory: repoFactory,
		tokenSvc:    tokenSvc,
	}
}

// Start starts the authorization server.
func (s *Service) Start(ctx context.Context) error {
	// TODO: Implement server startup logic.
	return nil
}

// Stop stops the authorization server.
func (s *Service) Stop(ctx context.Context) error {
	// TODO: Implement server shutdown logic.
	return nil
}
