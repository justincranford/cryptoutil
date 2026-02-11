// Copyright (c) 2025 Justin Cranford
//
//

package idp

import (
	"context"
	"embed"
	"fmt"
	"html/template"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityAuth "cryptoutil/internal/apps/identity/idp/auth"
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
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
func (s *Service) Start(_ context.Context) error {
	// Initialize authentication profiles
	s.initializeAuthProfiles()

	return nil
}

// Stop stops the identity provider service.
func (s *Service) Stop(ctx context.Context) error {
	// Clean up expired sessions.
	sessionRepo := s.repoFactory.SessionRepository()

	if err := sessionRepo.DeleteExpired(ctx); err != nil {
		return fmt.Errorf("failed to delete expired sessions during shutdown: %w", err)
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

// RotateClientSecret rotates a client secret and returns the new plaintext secret.
// The new secret is returned ONCE - caller MUST save it.
// Old secret is archived in history and immediately invalidated.
func (s *Service) RotateClientSecret(ctx context.Context, clientID string, rotatedBy string, reason string) (string, error) {
	clientRepo := s.repoFactory.ClientRepository()

	// 1. Find client by OAuth client_id.
	client, err := clientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		return "", fmt.Errorf("failed to find client: %w", err)
	}

	// 2. Generate new secret.
	newPlaintext, newHashed, err := GenerateClientSecret()
	if err != nil {
		return "", fmt.Errorf("failed to generate new secret: %w", err)
	}

	// 3. Rotate in repository (archives old secret, updates client).
	if err := clientRepo.RotateSecret(ctx, client.ID, newHashed, rotatedBy, reason); err != nil {
		return "", fmt.Errorf("failed to rotate secret in repository: %w", err)
	}

	return newPlaintext, nil
}
