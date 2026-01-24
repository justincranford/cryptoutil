// Copyright (c) 2025 Justin Cranford
//
//

package authz

import (
	"context"
	"fmt"

	cryptoutilIdentityClientAuth "cryptoutil/internal/identity/authz/clientauth"

	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityEmail "cryptoutil/internal/identity/email"
	cryptoutilIdentityIssuer "cryptoutil/internal/identity/issuer"
	cryptoutilIdentityMfa "cryptoutil/internal/identity/mfa"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
	cryptoutilIdentityRotation "cryptoutil/internal/identity/rotation"
)

// Service provides OAuth 2.1 authorization server functionality.
type Service struct {
	config          *cryptoutilIdentityConfig.Config
	repoFactory     *cryptoutilIdentityRepository.RepositoryFactory
	tokenSvc        *cryptoutilIdentityIssuer.TokenService
	clientAuth      *cryptoutilIdentityClientAuth.Registry
	authReqStore    AuthorizationRequestStore
	emailOTPService *cryptoutilIdentityMfa.EmailOTPService
}

// NewService creates a new authorization server service.
func NewService(
	config *cryptoutilIdentityConfig.Config,
	repoFactory *cryptoutilIdentityRepository.RepositoryFactory,
	tokenSvc *cryptoutilIdentityIssuer.TokenService,
) *Service {
	// Create rotation service for multi-secret authentication
	rotationService := cryptoutilIdentityRotation.NewSecretRotationService(repoFactory.DB())

	// Create email service (mock for now, will be configured via SMTP later).
	emailService := cryptoutilIdentityEmail.NewMockEmailService()

	// Create email OTP service.
	emailOTPService := cryptoutilIdentityMfa.NewEmailOTPService(
		repoFactory.EmailOTPRepository(),
		emailService,
	)

	return &Service{
		config:          config,
		repoFactory:     repoFactory,
		tokenSvc:        tokenSvc,
		clientAuth:      cryptoutilIdentityClientAuth.NewRegistry(repoFactory, config, rotationService),
		authReqStore:    NewInMemoryAuthorizationRequestStore(),
		emailOTPService: emailOTPService,
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
// Note: Database connections are managed by RepositoryFactory, not by this service.
// Closing the database here would break parallel test execution where multiple
// services share the same in-memory database connection pool.
func (s *Service) Stop(ctx context.Context) error {
	// Clean up expired tokens.
	tokenRepo := s.repoFactory.TokenRepository()

	if err := tokenRepo.DeleteExpired(ctx); err != nil {
		return fmt.Errorf("failed to delete expired tokens during shutdown: %w", err)
	}

	// Database connections are NOT closed here. The RepositoryFactory owns the
	// connection lifecycle. For production deployments, call RepositoryFactory.Close()
	// separately after all services have stopped.
	return nil
}
