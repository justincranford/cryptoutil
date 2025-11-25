// Copyright (c) 2025 Justin Cranford
//
//

package idp

import (
	"context"
	"embed"
	"html/template"

	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityAuth "cryptoutil/internal/identity/idp/auth"
	cryptoutilIdentityIssuer "cryptoutil/internal/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

// Embed HTML templates at compile time.
//
//go:embed templates/*.html
var templatesFS embed.FS

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
	// Parse HTML templates from embedded filesystem.
	templates := template.Must(template.ParseFS(templatesFS, "templates/*.html"))

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
	// Clean up expired sessions.
	sessionRepo := s.repoFactory.SessionRepository()
	if err := sessionRepo.DeleteExpired(ctx); err != nil {
		return err
	}

	return nil
}

// initializeAuthProfiles sets up the available authentication profiles.
func (s *Service) initializeAuthProfiles() {
	// Register username/password authentication.
	usernamePasswordProfile := cryptoutilIdentityAuth.NewUsernamePasswordProfile(s.repoFactory.UserRepository())
	s.authProfiles.Register(usernamePasswordProfile)

	// Future authentication profiles to register:
	// - Email + OTP: cryptoutilIdentityAuth.NewEmailOTPProfile(...)
	// - TOTP: cryptoutilIdentityAuth.NewTOTPProfile(...)
	// - Passkey (WebAuthn): cryptoutilIdentityAuth.NewPasskeyProfile(...)
}
