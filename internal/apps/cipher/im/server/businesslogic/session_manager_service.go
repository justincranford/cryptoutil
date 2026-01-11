// Copyright (c) 2025 Justin Cranford
//

package businesslogic

import (
	"context"
	"fmt"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilTemplateBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilTemplateBusinessLogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilTemplateRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
)

// SessionManagerService provides session management functionality for cipher-im.
// It wraps the template SessionManager with cipher-im specific configuration.
type SessionManagerService struct {
	sessionManager *cryptoutilTemplateBusinessLogic.SessionManager
}

// NewSessionManagerService creates a new SessionManagerService instance.
func NewSessionManagerService(
	ctx context.Context,
	db *gorm.DB,
	telemetryService *cryptoutilTelemetry.TelemetryService,
	jwkGenService *cryptoutilJose.JWKGenService,
	barrierService *cryptoutilTemplateBarrier.BarrierService,
	config *cryptoutilConfig.ServiceTemplateServerSettings,
) (*SessionManagerService, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}

	if db == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}

	if telemetryService == nil {
		return nil, fmt.Errorf("telemetry service cannot be nil")
	}

	if jwkGenService == nil {
		return nil, fmt.Errorf("JWK generation service cannot be nil")
	}

	if barrierService == nil {
		return nil, fmt.Errorf("barrier service cannot be nil")
	}

	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Create SessionManager with cipher-im specific configuration.
	sessionManager := cryptoutilTemplateBusinessLogic.NewSessionManager(
		db,
		barrierService,
		config,
	)

	// Initialize the SessionManager.
	if err := sessionManager.Initialize(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize session manager: %w", err)
	}

	return &SessionManagerService{
		sessionManager: sessionManager,
	}, nil
}

// IssueBrowserSessionWithTenant creates a new browser session token (multi-tenant version).
func (s *SessionManagerService) IssueBrowserSessionWithTenant(
	ctx context.Context,
	userID string,
	tenantID googleUuid.UUID,
	realmID googleUuid.UUID,
) (string, error) {
	if ctx == nil {
		return "", fmt.Errorf("context cannot be nil")
	}

	if userID == "" {
		return "", fmt.Errorf("user ID cannot be empty")
	}

	return s.sessionManager.IssueBrowserSession(ctx, userID, tenantID, realmID)
}

// ValidateBrowserSession validates a browser session token and returns the session.
func (s *SessionManagerService) ValidateBrowserSession(
	ctx context.Context,
	token string,
) (*cryptoutilTemplateRepository.BrowserSession, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}

	if token == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	return s.sessionManager.ValidateBrowserSession(ctx, token)
}

// IssueServiceSessionWithTenant creates a new service session token (multi-tenant version).
func (s *SessionManagerService) IssueServiceSessionWithTenant(
	ctx context.Context,
	clientID string,
	tenantID googleUuid.UUID,
	realmID googleUuid.UUID,
) (string, error) {
	if ctx == nil {
		return "", fmt.Errorf("context cannot be nil")
	}

	if clientID == "" {
		return "", fmt.Errorf("client ID cannot be empty")
	}

	return s.sessionManager.IssueServiceSession(ctx, clientID, tenantID, realmID)
}

// ValidateServiceSession validates a service session token and returns the session.
func (s *SessionManagerService) ValidateServiceSession(
	ctx context.Context,
	token string,
) (*cryptoutilTemplateRepository.ServiceSession, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}

	if token == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	return s.sessionManager.ValidateServiceSession(ctx, token)
}

// CleanupExpiredSessions removes expired sessions from the database.
func (s *SessionManagerService) CleanupExpiredSessions(
	ctx context.Context,
) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}

	return s.sessionManager.CleanupExpiredSessions(ctx)
}

// IssueBrowserSession (simplified signature for single-tenant applications).
// Wrapper method for compatibility with template realms handler.
func (s *SessionManagerService) IssueBrowserSession(ctx context.Context, userID string, realm string) (string, error) {
	// For single-tenant cipher-im, use nil UUIDs for tenant and realm.
	var nilUUID googleUuid.UUID

	return s.sessionManager.IssueBrowserSession(ctx, userID, nilUUID, nilUUID)
}

// IssueServiceSession (simplified signature for single-tenant applications).
// Wrapper method for compatibility with template realms handler.
func (s *SessionManagerService) IssueServiceSession(ctx context.Context, userID string, realm string) (string, error) {
	// For single-tenant cipher-im, use nil UUIDs for tenant and realm.
	var nilUUID googleUuid.UUID

	return s.sessionManager.IssueServiceSession(ctx, userID, nilUUID, nilUUID)
}
