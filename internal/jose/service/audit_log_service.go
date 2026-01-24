// Copyright (c) 2025 Justin Cranford
//
//

package service

import (
	"context"
	json "encoding/json"

	cryptoutilJoseDomain "cryptoutil/internal/jose/domain"
	cryptoutilJoseRepository "cryptoutil/internal/jose/repository"

	googleUuid "github.com/google/uuid"
)

// Context keys for audit logging.
type (
	userContextKey    struct{}
	sessionContextKey struct{}
)

// ContextWithUser adds a user ID to the context for audit logging.
func ContextWithUser(ctx context.Context, userID googleUuid.UUID) context.Context {
	return context.WithValue(ctx, userContextKey{}, userID)
}

// UserFromContext retrieves the user ID from context, if present.
func UserFromContext(ctx context.Context) *googleUuid.UUID {
	userID, ok := ctx.Value(userContextKey{}).(googleUuid.UUID)
	if !ok {
		return nil
	}

	return &userID
}

// ContextWithSession adds a session ID to the context for audit logging.
func ContextWithSession(ctx context.Context, sessionID googleUuid.UUID) context.Context {
	return context.WithValue(ctx, sessionContextKey{}, sessionID)
}

// SessionFromContext retrieves the session ID from context, if present.
func SessionFromContext(ctx context.Context) *googleUuid.UUID {
	sessionID, ok := ctx.Value(sessionContextKey{}).(googleUuid.UUID)
	if !ok {
		return nil
	}

	return &sessionID
}

// AuditLogService handles audit logging for cryptographic operations.
type AuditLogService struct {
	configService *AuditConfigService
	logRepo       cryptoutilJoseRepository.AuditLogRepository
}

// NewAuditLogService creates a new audit log service.
func NewAuditLogService(configService *AuditConfigService, logRepo cryptoutilJoseRepository.AuditLogRepository) *AuditLogService {
	return &AuditLogService{
		configService: configService,
		logRepo:       logRepo,
	}
}

// AuditLogParams contains parameters for creating an audit log entry.
type AuditLogParams struct {
	TenantID     googleUuid.UUID
	RealmID      googleUuid.UUID
	UserID       *googleUuid.UUID // Optional: from session context.
	SessionID    *googleUuid.UUID // Optional: from session context.
	Operation    string
	ResourceType string // "elastic_jwk" or "material_jwk".
	ResourceID   string // KID of the elastic or material JWK.
	Success      bool
	ErrorMessage *string
	Metadata     map[string]any // Additional operation-specific details.
}

// Log creates an audit log entry if enabled and sampling passes.
// Returns true if the entry was logged, false if skipped.
func (s *AuditLogService) Log(ctx context.Context, params AuditLogParams) (bool, error) {
	// Check if audit logging is enabled for this operation.
	enabled, samplingRate, err := s.configService.IsEnabled(ctx, params.TenantID, params.Operation)
	if err != nil {
		// If we can't check config, skip logging but don't fail the operation.
		return false, nil //nolint:nilerr // Audit logging errors should not fail crypto operations.
	}

	if !enabled {
		return false, nil
	}

	// Build the audit log entry.
	entry := &cryptoutilJoseDomain.AuditLogEntry{
		ID:           googleUuid.New(),
		TenantID:     params.TenantID,
		RealmID:      params.RealmID,
		UserID:       params.UserID,
		SessionID:    params.SessionID,
		Operation:    params.Operation,
		ResourceType: params.ResourceType,
		ResourceID:   params.ResourceID,
		Success:      params.Success,
		ErrorMessage: params.ErrorMessage,
	}

	// Serialize metadata if provided.
	if params.Metadata != nil {
		metadataJSON, marshalErr := json.Marshal(params.Metadata)
		if marshalErr == nil {
			metadataStr := string(metadataJSON)
			entry.Metadata = &metadataStr
		}
	}

	// Create with sampling - only logs if random check passes.
	logged, createErr := s.logRepo.CreateWithSampling(ctx, entry, samplingRate)
	if createErr != nil {
		// Audit logging errors should not fail the operation.
		return false, nil //nolint:nilerr // Audit logging errors should not fail crypto operations.
	}

	return logged, nil
}

// LogSuccess is a convenience method for logging successful operations.
// Automatically extracts user_id and session_id from context if present.
func (s *AuditLogService) LogSuccess(ctx context.Context, tenantID, realmID googleUuid.UUID, operation, resourceType, resourceID string, metadata map[string]any) (bool, error) {
	return s.Log(ctx, AuditLogParams{
		TenantID:     tenantID,
		RealmID:      realmID,
		UserID:       UserFromContext(ctx),
		SessionID:    SessionFromContext(ctx),
		Operation:    operation,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Success:      true,
		Metadata:     metadata,
	})
}

// LogFailure is a convenience method for logging failed operations.
// Automatically extracts user_id and session_id from context if present.
func (s *AuditLogService) LogFailure(ctx context.Context, tenantID, realmID googleUuid.UUID, operation, resourceType, resourceID string, err error, metadata map[string]any) (bool, error) {
	errMsg := err.Error()

	return s.Log(ctx, AuditLogParams{
		TenantID:     tenantID,
		RealmID:      realmID,
		UserID:       UserFromContext(ctx),
		SessionID:    SessionFromContext(ctx),
		Operation:    operation,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Success:      false,
		ErrorMessage: &errMsg,
		Metadata:     metadata,
	})
}
