// Copyright (c) 2025 Justin Cranford
//

package businesslogic

import (
	"context"
	"fmt"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/apps/template/service/telemetry"
)

// SessionManagerService provides session management functionality.
// It wraps the template SessionManager with validation.
type SessionManagerService struct {
	sessionManager *SessionManager
}

// NewSessionManagerService creates a new SessionManagerService instance.
// For multi-tenant applications, use IssueBrowserSessionWithTenant or IssueServiceSessionWithTenant.
func NewSessionManagerService(
	ctx context.Context,
	db *gorm.DB,
	telemetryService *cryptoutilSharedTelemetry.TelemetryService,
	jwkGenService *cryptoutilSharedCryptoJose.JWKGenService,
	barrierService *cryptoutilAppsTemplateServiceServerBarrier.Service,
	config *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings,
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

	// Create SessionManager.
	sessionManager := NewSessionManager(
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
) (*cryptoutilAppsTemplateServiceServerRepository.BrowserSession, error) {
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
) (*cryptoutilAppsTemplateServiceServerRepository.ServiceSession, error) {
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
