// Copyright (c) 2025 Justin Cranford
//
//

package idp

import (
	"context"
	"html/template"

	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityAuth "cryptoutil/internal/identity/idp/auth"
	cryptoutilIdentityIssuer "cryptoutil/internal/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

// Service provides OIDC identity provider functionality.
type Service struct {
	config       *cryptoutilIdentityConfig.Config
	repoFactory  *cryptoutilIdentityRepository.RepositoryFactory
	tokenSvc     *cryptoutilIdentityIssuer.TokenService
	authProfiles *cryptoutilIdentityAuth.ProfileRegistry
	templates    *template.Template
}

// NewService creates a new identity provider service.
func NewService(
	config *cryptoutilIdentityConfig.Config,
	repoFactory *cryptoutilIdentityRepository.RepositoryFactory,
	tokenSvc *cryptoutilIdentityIssuer.TokenService,
) *Service {
	// Parse HTML templates.
	templates := template.Must(template.ParseGlob("internal/identity/idp/templates/*.html"))

	return &Service{
		config:       config,
		repoFactory:  repoFactory,
		tokenSvc:     tokenSvc,
		authProfiles: cryptoutilIdentityAuth.NewProfileRegistry(),
		templates:    templates,
	}
}

// Start starts the identity provider service.
func (s *Service) Start(ctx context.Context) error {
	// Initialize authentication profiles
	s.initializeAuthProfiles()

	return nil
}

// Stop stops the identity provider service.
func (s *Service) Stop(ctx context.Context) error {
	// TODO: Implement cleanup logic for sessions, challenges, etc.
	return nil
}

// initializeAuthProfiles sets up the available authentication profiles.
func (s *Service) initializeAuthProfiles() {
	// Register username/password authentication
	usernamePasswordProfile := cryptoutilIdentityAuth.NewUsernamePasswordProfile(s.repoFactory.UserRepository())
	// TODO: Register additional authentication profiles (email+OTP, TOTP, passkey, etc.)
	s.authProfiles.Register(usernamePasswordProfile)
}
