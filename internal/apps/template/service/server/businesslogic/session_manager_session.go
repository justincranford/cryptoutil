// Copyright (c) 2025 Justin Cranford
//
//

// TODO(cipher-im-migration): This SessionManager implementation is work-in-progress.
// Current status:
// - ✅ OPAQUE session issuance and validation (uses hash package directly)
// - ✅ JWS/JWE session issuance and validation (complete implementation)
// - ✅ JWK encryption with barrier service (fully implemented)
// - ✅ Comprehensive unit tests (all 25 tests passing)
//
// Next steps:
// - Write integration tests
// - Verify quality gates (coverage ≥95%, mutations ≥85%)
// - Integrate with cipher-im service

// Package businesslogic provides business logic services for the template service.
package businesslogic

import (
	"context"
	"crypto"
	"crypto/elliptic"
	"errors"
	"fmt"
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilSharedCryptoHash "cryptoutil/internal/shared/crypto/hash"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func (sm *SessionManager) generateSessionJWK(isBrowser bool, algorithm cryptoutilSharedMagic.SessionAlgorithmType) (crypto.PrivateKey, error) {
	var algIdentifier string
	if isBrowser {
		algIdentifier = sm.getAlgorithmIdentifier(isBrowser, algorithm)
	} else {
		algIdentifier = sm.getAlgorithmIdentifier(isBrowser, algorithm)
	}

	switch algorithm {
	case cryptoutilSharedMagic.SessionAlgorithmJWS:
		// Generate asymmetric key for JWS signature
		return sm.generateJWSKey(algIdentifier)
	case cryptoutilSharedMagic.SessionAlgorithmJWE:
		// Generate symmetric key for JWE encryption
		return sm.generateJWEKey(algIdentifier)
	default:
		return nil, fmt.Errorf("unsupported session algorithm: %s", algorithm)
	}
}

// generateJWSKey generates an asymmetric signing key for JWS tokens.
func (sm *SessionManager) generateJWSKey(algorithm string) (crypto.PrivateKey, error) {
	switch algorithm {
	case cryptoutilSharedMagic.SessionJWSAlgorithmRS256,
		cryptoutilSharedMagic.SessionJWSAlgorithmRS384,
		cryptoutilSharedMagic.SessionJWSAlgorithmRS512:
		// RSA key generation
		keyPair, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(cryptoutilSharedMagic.RSAKeySize2048)
		if err != nil {
			return nil, fmt.Errorf("failed to generate RSA key pair: %w", err)
		}

		return keyPair.Private, nil
	case cryptoutilSharedMagic.SessionJWSAlgorithmES256:
		// ECDSA P-256
		keyPair, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P256())
		if err != nil {
			return nil, fmt.Errorf("failed to generate ECDSA P-256 key pair: %w", err)
		}

		return keyPair.Private, nil
	case cryptoutilSharedMagic.SessionJWSAlgorithmES384:
		// ECDSA P-384
		keyPair, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P384())
		if err != nil {
			return nil, fmt.Errorf("failed to generate ECDSA P-384 key pair: %w", err)
		}

		return keyPair.Private, nil
	case cryptoutilSharedMagic.SessionJWSAlgorithmES512:
		// ECDSA P-521
		keyPair, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P521())
		if err != nil {
			return nil, fmt.Errorf("failed to generate ECDSA P-521 key pair: %w", err)
		}

		return keyPair.Private, nil
	case cryptoutilSharedMagic.SessionJWSAlgorithmEdDSA:
		// Ed25519
		keyPair, err := cryptoutilSharedCryptoKeygen.GenerateEDDSAKeyPair(cryptoutilSharedCryptoKeygen.EdCurveEd25519)
		if err != nil {
			return nil, fmt.Errorf("failed to generate EdDSA key pair: %w", err)
		}

		return keyPair.Private, nil
	default:
		return nil, fmt.Errorf("unsupported JWS algorithm: %s", algorithm)
	}
}

// generateJWEKey generates a symmetric encryption key for JWE tokens.
func (sm *SessionManager) generateJWEKey(algorithm string) (crypto.PrivateKey, error) {
	switch algorithm {
	case cryptoutilSharedMagic.SessionJWEAlgorithmDirA256GCM,
		cryptoutilSharedMagic.SessionJWEAlgorithmA256GCMKWA256GCM:
		// AES-256 key generation (32 bytes)
		key, err := cryptoutilSharedCryptoKeygen.GenerateAESKey(cryptoutilSharedMagic.AESKeySize256)
		if err != nil {
			return nil, fmt.Errorf("failed to generate AES key: %w", err)
		}

		return key, nil
	default:
		return nil, fmt.Errorf("unsupported JWE algorithm: %s", algorithm)
	}
}

// getAlgorithmIdentifier returns the specific algorithm identifier for session tokens.
func (sm *SessionManager) getAlgorithmIdentifier(isBrowser bool, sessionAlgorithm cryptoutilSharedMagic.SessionAlgorithmType) string {
	switch sessionAlgorithm {
	case cryptoutilSharedMagic.SessionAlgorithmJWS:
		if isBrowser {
			return sm.config.BrowserSessionJWSAlgorithm
		}

		return sm.config.ServiceSessionJWSAlgorithm
	case cryptoutilSharedMagic.SessionAlgorithmJWE:
		if isBrowser {
			return sm.config.BrowserSessionJWEAlgorithm
		}

		return sm.config.ServiceSessionJWEAlgorithm
	default:
		return string(sessionAlgorithm)
	}
}

// IssueBrowserSession issues a new browser session token.
//
// For OPAQUE: Generates UUIDv7, hashes it, stores hash in database
// For JWS/JWE: Generates JWT with jti claim, optionally stores jti in database
//
// Parameters:
//   - ctx: Context for database operations
//   - userID: User identifier (optional, can be empty string)
//   - tenantID: Tenant identifier for multi-tenancy isolation
//   - realmID: Realm identifier within tenant
//
// Returns session token string for client.
func (sm *SessionManager) IssueBrowserSession(ctx context.Context, userID string, tenantID, realmID googleUuid.UUID) (string, error) {
	switch sm.browserAlgorithm {
	case cryptoutilSharedMagic.SessionAlgorithmOPAQUE:
		return sm.issueOPAQUESession(ctx, true, userID, tenantID, realmID)
	case cryptoutilSharedMagic.SessionAlgorithmJWS:
		return sm.issueJWSSession(ctx, true, userID, tenantID, realmID)
	case cryptoutilSharedMagic.SessionAlgorithmJWE:
		return sm.issueJWESession(ctx, true, userID, tenantID, realmID)
	default:
		return "", fmt.Errorf("unsupported browser session algorithm: %s", sm.browserAlgorithm)
	}
}

// ValidateBrowserSession validates a browser session token.
//
// For OPAQUE: Hashes token, looks up hash in database
// For JWS/JWE: Verifies JWT signature/encryption, checks expiration
//
// Parameters:
//   - ctx: Context for database operations
//   - token: Session token to validate
//
// Returns session metadata if valid, error otherwise.
func (sm *SessionManager) ValidateBrowserSession(ctx context.Context, token string) (*cryptoutilAppsTemplateServiceServerRepository.BrowserSession, error) {
	var (
		result any
		err    error
	)

	switch sm.browserAlgorithm {
	case cryptoutilSharedMagic.SessionAlgorithmOPAQUE:
		result, err = sm.validateOPAQUESession(ctx, true, token)
	case cryptoutilSharedMagic.SessionAlgorithmJWS:
		result, err = sm.validateJWSSession(ctx, true, token)
	case cryptoutilSharedMagic.SessionAlgorithmJWE:
		result, err = sm.validateJWESession(ctx, true, token)
	default:
		return nil, fmt.Errorf("unsupported browser session algorithm: %s", sm.browserAlgorithm)
	}

	if err != nil {
		return nil, err
	}

	// Type assert to BrowserSession
	browserSession, ok := result.(*cryptoutilAppsTemplateServiceServerRepository.BrowserSession)
	if !ok {
		return nil, fmt.Errorf("invalid session type returned")
	}

	return browserSession, nil
}

// IssueServiceSession issues a new service session token.
// Similar to IssueBrowserSession but for service-to-service authentication.
func (sm *SessionManager) IssueServiceSession(ctx context.Context, clientID string, tenantID, realmID googleUuid.UUID) (string, error) {
	switch sm.serviceAlgorithm {
	case cryptoutilSharedMagic.SessionAlgorithmOPAQUE:
		return sm.issueOPAQUESession(ctx, false, clientID, tenantID, realmID)
	case cryptoutilSharedMagic.SessionAlgorithmJWS:
		return sm.issueJWSSession(ctx, false, clientID, tenantID, realmID)
	case cryptoutilSharedMagic.SessionAlgorithmJWE:
		return sm.issueJWESession(ctx, false, clientID, tenantID, realmID)
	default:
		return "", fmt.Errorf("unsupported service session algorithm: %s", sm.serviceAlgorithm)
	}
}

// ValidateServiceSession validates a service session token.
// Similar to ValidateBrowserSession but for service-to-service authentication.
func (sm *SessionManager) ValidateServiceSession(ctx context.Context, token string) (*cryptoutilAppsTemplateServiceServerRepository.ServiceSession, error) {
	var (
		result any
		err    error
	)

	switch sm.serviceAlgorithm {
	case cryptoutilSharedMagic.SessionAlgorithmOPAQUE:
		result, err = sm.validateOPAQUESession(ctx, false, token)
	case cryptoutilSharedMagic.SessionAlgorithmJWS:
		result, err = sm.validateJWSSession(ctx, false, token)
	case cryptoutilSharedMagic.SessionAlgorithmJWE:
		result, err = sm.validateJWESession(ctx, false, token)
	default:
		return nil, fmt.Errorf("unsupported service session algorithm: %s", sm.serviceAlgorithm)
	}

	if err != nil {
		return nil, err
	}

	// Type assert to ServiceSession
	serviceSession, ok := result.(*cryptoutilAppsTemplateServiceServerRepository.ServiceSession)
	if !ok {
		return nil, fmt.Errorf("invalid session type returned")
	}

	return serviceSession, nil
}

// CleanupExpiredSessions removes expired sessions from the database.
//
// Deletes sessions where:
//   - Expiration timestamp is in the past
//   - Last activity timestamp exceeds idle timeout
//
// Should be called periodically (e.g., every hour) to prevent database bloat.
func (sm *SessionManager) CleanupExpiredSessions(ctx context.Context) error {
	now := time.Now().UTC()
	idleThreshold := now.Add(-sm.config.SessionIdleTimeout)

	// Cleanup browser sessions
	err := sm.db.WithContext(ctx).
		Where("expiration < ? OR (last_activity IS NOT NULL AND last_activity < ?)", now, idleThreshold).
		Delete(&cryptoutilAppsTemplateServiceServerRepository.BrowserSession{}).
		Error
	if err != nil {
		return fmt.Errorf("failed to cleanup browser sessions: %w", err)
	}

	// Cleanup service sessions
	err = sm.db.WithContext(ctx).
		Where("expiration < ? OR (last_activity IS NOT NULL AND last_activity < ?)", now, idleThreshold).
		Delete(&cryptoutilAppsTemplateServiceServerRepository.ServiceSession{}).
		Error
	if err != nil {
		return fmt.Errorf("failed to cleanup service sessions: %w", err)
	}

	return nil
}

// StartCleanupTask starts a background goroutine that periodically cleans up expired sessions.
//
// The cleanup task runs at the interval specified in config.SessionCleanupInterval.
// The goroutine will stop when the context is cancelled.
//
// Example usage:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	sm.StartCleanupTask(ctx)
func (sm *SessionManager) StartCleanupTask(ctx context.Context) {
	ticker := time.NewTicker(sm.config.SessionCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := sm.CleanupExpiredSessions(ctx); err != nil {
				// Log error but continue cleanup task
				// TODO: Use proper logger from context
				fmt.Printf("Session cleanup error: %v\n", err)
			}
		}
	}
}

// issueOPAQUESession issues an OPAQUE session token (hashed UUIDv7).
func (sm *SessionManager) issueOPAQUESession(ctx context.Context, isBrowser bool, principalID string, tenantID, realmID googleUuid.UUID) (string, error) {
	// Generate UUIDv7 token
	tokenID := googleUuid.Must(googleUuid.NewV7())
	token := tokenID.String()

	// Hash token for database storage using HighEntropyDeterministic
	tokenHash, err := cryptoutilSharedCryptoHash.HashHighEntropyDeterministic(token)
	if err != nil {
		return "", fmt.Errorf("failed to hash session token: %w", err)
	}

	// Calculate expiration
	var expiration time.Time
	if isBrowser {
		expiration = time.Now().UTC().Add(sm.config.BrowserSessionExpiration)
	} else {
		expiration = time.Now().UTC().Add(sm.config.ServiceSessionExpiration)
	}

	// Create session record
	now := time.Now().UTC()
	session := cryptoutilAppsTemplateServiceServerRepository.Session{
		ID:           tokenID,
		TenantID:     tenantID,
		RealmID:      realmID,
		TokenHash:    &tokenHash,
		Expiration:   expiration,
		CreatedAt:    now,
		LastActivity: &now,
	}

	// Store session in database
	var createErr error

	if isBrowser {
		browserSession := cryptoutilAppsTemplateServiceServerRepository.BrowserSession{
			Session: session,
			UserID:  &principalID,
		}
		createErr = sm.db.WithContext(ctx).Create(&browserSession).Error
	} else {
		serviceSession := cryptoutilAppsTemplateServiceServerRepository.ServiceSession{
			Session:  session,
			ClientID: &principalID,
		}
		createErr = sm.db.WithContext(ctx).Create(&serviceSession).Error
	}

	if createErr != nil {
		return "", fmt.Errorf("failed to store session: %w", createErr)
	}

	return token, nil
}

// validateOPAQUESession validates an OPAQUE session token.
func (sm *SessionManager) validateOPAQUESession(ctx context.Context, isBrowser bool, token string) (any, error) {
	// Hash token for database lookup
	tokenHash, err := cryptoutilSharedCryptoHash.HashHighEntropyDeterministic(token)
	if err != nil {
		return nil, fmt.Errorf("failed to hash session token: %w", err)
	}

	// Look up session by token hash
	now := time.Now().UTC()

	var (
		session any
		findErr error
	)

	if isBrowser {
		browserSession := &cryptoutilAppsTemplateServiceServerRepository.BrowserSession{}
		findErr = sm.db.WithContext(ctx).
			Where("token_hash = ? AND expiration > ?", tokenHash, now).
			First(browserSession).
			Error
		session = browserSession
	} else {
		serviceSession := &cryptoutilAppsTemplateServiceServerRepository.ServiceSession{}
		findErr = sm.db.WithContext(ctx).
			Where("token_hash = ? AND expiration > ?", tokenHash, now).
			First(serviceSession).
			Error
		session = serviceSession
	}

	if findErr != nil {
		if errors.Is(findErr, gorm.ErrRecordNotFound) {
			summary := "Invalid or expired session token"

			return nil, cryptoutilSharedApperr.NewHTTP401Unauthorized(&summary, findErr)
		}

		return nil, fmt.Errorf("failed to query session: %w", findErr)
	}

	// Update last activity timestamp
	updateErr := sm.db.WithContext(ctx).Model(session).
		Update("last_activity", now).
		Error
	if updateErr != nil {
		// Log warning but don't fail validation
		fmt.Printf("Failed to update session activity: %v\n", updateErr)
	}

	return session, nil
}

// issueJWSSession issues a JWS session token (signed JWT).
