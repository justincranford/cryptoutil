// Copyright (c) 2025 Justin Cranford
//
//

// Package businesslogic provides business logic services for the template service.
package businesslogic

import (
	"context"
	"crypto"
	"errors"
	"fmt"
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilCrypto "cryptoutil/internal/shared/crypto"
	cryptoutilHash "cryptoutil/internal/shared/crypto/hash"
	cryptoutilJOSE "cryptoutil/internal/shared/crypto/jose"
	cryptoutilKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// SessionManager manages session tokens for browser and service clients.
//
// Supports three session token algorithms:
//   - OPAQUE: Hashed UUIDv7 tokens stored in database (server-side validation only)
//   - JWS: Signed JWT tokens with asymmetric signature verification (stateless validation)
//   - JWE: Encrypted JWT tokens with symmetric encryption (stateless validation)
//
// The SessionManager:
//   - Generates and stores session JWKs encrypted with barrier layer
//   - Issues session tokens with configurable expiration
//   - Validates session tokens (database lookup for OPAQUE, JWT verification for JWS/JWE)
//   - Tracks session activity for idle timeout enforcement
//   - Periodically cleans up expired sessions
//
// Usage:
//
//	manager := NewSessionManager(db, barrierService, config)
//	err := manager.Initialize(ctx)
//	
//	// Issue browser session token
//	token, err := manager.IssueBrowserSession(ctx, userID, realm)
//
//	// Validate browser session token
//	session, err := manager.ValidateBrowserSession(ctx, token)
//
//	// Cleanup expired sessions
//	err := manager.CleanupExpiredSessions(ctx)
type SessionManager struct {
	db      *gorm.DB
	barrier *cryptoutilBarrier.BarrierService
	config  *cryptoutilConfig.ServiceTemplateServerSettings

	// Runtime state for browser sessions
	browserAlgorithm cryptoutilMagic.SessionAlgorithmType
	browserJWKID     *googleUuid.UUID // Active JWK ID for JWS/JWE, nil for OPAQUE

	// Runtime state for service sessions
	serviceAlgorithm cryptoutilMagic.SessionAlgorithmType
	serviceJWKID     *googleUuid.UUID // Active JWK ID for JWS/JWE, nil for OPAQUE

	// Hash service for OPAQUE tokens
	hashService *cryptoutilHash.HashService
}

// NewSessionManager creates a new SessionManager instance.
//
// Parameters:
//   - db: GORM database connection for session storage
//   - barrier: Barrier service for encrypting JWKs
//   - config: Service configuration with session settings
//
// Returns configured SessionManager (call Initialize before use).
func NewSessionManager(db *gorm.DB, barrier *cryptoutilBarrier.BarrierService, config *cryptoutilConfig.ServiceTemplateServerSettings) *SessionManager {
	return &SessionManager{
		db:      db,
		barrier: barrier,
		config:  config,
	}
}

// Initialize prepares the SessionManager for use.
//
// For JWS/JWE algorithms:
//   - Generates new JWKs if none exist
//   - Loads existing JWKs from database
//   - Decrypts JWKs using barrier service
//   - Stores active JWK ID for session issuance
//
// For OPAQUE algorithm:
//   - Initializes hash service for token hashing
//   - No JWKs generated or stored
//
// MUST be called before issuing or validating sessions.
func (sm *SessionManager) Initialize(ctx context.Context) error {
	// Parse browser session algorithm
	browserAlg := cryptoutilMagic.SessionAlgorithmType(sm.config.BrowserSessionAlgorithm)
	if browserAlg == "" {
		browserAlg = cryptoutilMagic.DefaultBrowserSessionAlgorithm
	}
	sm.browserAlgorithm = browserAlg

	// Parse service session algorithm
	serviceAlg := cryptoutilMagic.SessionAlgorithmType(sm.config.ServiceSessionAlgorithm)
	if serviceAlg == "" {
		serviceAlg = cryptoutilMagic.DefaultServiceSessionAlgorithm
	}
	sm.serviceAlgorithm = serviceAlg

	// Initialize browser session JWKs if using JWS/JWE
	if browserAlg == cryptoutilMagic.SessionAlgorithmJWS || browserAlg == cryptoutilMagic.SessionAlgorithmJWE {
		jwkID, err := sm.initializeSessionJWK(ctx, true, browserAlg)
		if err != nil {
			return fmt.Errorf("failed to initialize browser session JWK: %w", err)
		}
		sm.browserJWKID = &jwkID
	}

	// Initialize service session JWKs if using JWS/JWE
	if serviceAlg == cryptoutilMagic.SessionAlgorithmJWS || serviceAlg == cryptoutilMagic.SessionAlgorithmJWE {
		jwkID, err := sm.initializeSessionJWK(ctx, false, serviceAlg)
		if err != nil {
			return fmt.Errorf("failed to initialize service session JWK: %w", err)
		}
		sm.serviceJWKID = &jwkID
	}

	// Initialize hash service for OPAQUE tokens
	if browserAlg == cryptoutilMagic.SessionAlgorithmOPAQUE || serviceAlg == cryptoutilMagic.SessionAlgorithmOPAQUE {
		// Hash service will be initialized on first use (lazy initialization)
		// This avoids requiring unseal keys during manager construction
		sm.hashService = cryptoutilHash.NewHashService()
	}

	return nil
}

// initializeSessionJWK generates or loads session JWK for JWS/JWE algorithms.
//
// Parameters:
//   - ctx: Context for database operations
//   - isBrowser: true for browser sessions, false for service sessions
//   - algorithm: Session algorithm type (JWS or JWE)
//
// Returns active JWK ID for session issuance.
func (sm *SessionManager) initializeSessionJWK(ctx context.Context, isBrowser bool, algorithm cryptoutilMagic.SessionAlgorithmType) (googleUuid.UUID, error) {
	// Determine table model
	var jwkModel interface{}
	var tableName string
	if isBrowser {
		jwkModel = &cryptoutilRepository.BrowserSessionJWK{}
		tableName = "browser_session_jwks"
	} else {
		jwkModel = &cryptoutilRepository.ServiceSessionJWK{}
		tableName = "service_session_jwks"
	}

	// Check for existing active JWK
	var existingJWK cryptoutilRepository.SessionJWK
	err := sm.db.WithContext(ctx).
		Table(tableName).
		Where("active = ?", true).
		Order("created_at DESC").
		First(&existingJWK).
		Error

	if err == nil {
		// Found existing active JWK, use it
		return existingJWK.ID, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return googleUuid.UUID{}, fmt.Errorf("failed to query active JWK: %w", err)
	}

	// No active JWK found, generate new one
	jwk, err := sm.generateSessionJWK(isBrowser, algorithm)
	if err != nil {
		return googleUuid.UUID{}, fmt.Errorf("failed to generate session JWK: %w", err)
	}

	// Encrypt JWK with barrier service
	jwkBytes, err := cryptoutilJOSE.MarshalJWK(jwk)
	if err != nil {
		return googleUuid.UUID{}, fmt.Errorf("failed to marshal JWK: %w", err)
	}

	encryptedJWK, err := sm.barrier.EncryptContent(ctx, jwkBytes)
	if err != nil {
		return googleUuid.UUID{}, fmt.Errorf("failed to encrypt JWK with barrier: %w", err)
	}

	// Store encrypted JWK in database
	jwkID := googleUuid.Must(googleUuid.NewV7())
	newJWK := cryptoutilRepository.SessionJWK{
		ID:           jwkID,
		EncryptedJWK: string(encryptedJWK),
		CreatedAt:    time.Now(),
		Algorithm:    sm.getAlgorithmIdentifier(isBrowser, algorithm),
		Active:       true,
	}

	var createErr error
	if isBrowser {
		browserJWK := cryptoutilRepository.BrowserSessionJWK{SessionJWK: newJWK}
		createErr = sm.db.WithContext(ctx).Create(&browserJWK).Error
	} else {
		serviceJWK := cryptoutilRepository.ServiceSessionJWK{SessionJWK: newJWK}
		createErr = sm.db.WithContext(ctx).Create(&serviceJWK).Error
	}

	if createErr != nil {
		return googleUuid.UUID{}, fmt.Errorf("failed to store encrypted JWK: %w", createErr)
	}

	return jwkID, nil
}

// generateSessionJWK generates a new JWK for session tokens.
//
// For JWS: Generates asymmetric signing key (RSA, ECDSA, or EdDSA)
// For JWE: Generates symmetric encryption key (AES)
func (sm *SessionManager) generateSessionJWK(isBrowser bool, algorithm cryptoutilMagic.SessionAlgorithmType) (crypto.PrivateKey, error) {
	var algIdentifier string
	if isBrowser {
		algIdentifier = sm.getAlgorithmIdentifier(isBrowser, algorithm)
	} else {
		algIdentifier = sm.getAlgorithmIdentifier(isBrowser, algorithm)
	}

	switch algorithm {
	case cryptoutilMagic.SessionAlgorithmJWS:
		// Generate asymmetric key for JWS signature
		return sm.generateJWSKey(algIdentifier)
	case cryptoutilMagic.SessionAlgorithmJWE:
		// Generate symmetric key for JWE encryption
		return sm.generateJWEKey(algIdentifier)
	default:
		return nil, fmt.Errorf("unsupported session algorithm: %s", algorithm)
	}
}

// generateJWSKey generates an asymmetric signing key for JWS tokens.
func (sm *SessionManager) generateJWSKey(algorithm string) (crypto.PrivateKey, error) {
	switch algorithm {
	case cryptoutilMagic.SessionJWSAlgorithmRS256,
		cryptoutilMagic.SessionJWSAlgorithmRS384,
		cryptoutilMagic.SessionJWSAlgorithmRS512:
		// RSA key generation
		return cryptoutilKeygen.GenerateRSAKeyPair(2048)
	case cryptoutilMagic.SessionJWSAlgorithmES256:
		// ECDSA P-256
		return cryptoutilKeygen.GenerateECDSAKeyPair("P-256")
	case cryptoutilMagic.SessionJWSAlgorithmES384:
		// ECDSA P-384
		return cryptoutilKeygen.GenerateECDSAKeyPair("P-384")
	case cryptoutilMagic.SessionJWSAlgorithmES512:
		// ECDSA P-521
		return cryptoutilKeygen.GenerateECDSAKeyPair("P-521")
	case cryptoutilMagic.SessionJWSAlgorithmEdDSA:
		// Ed25519
		return cryptoutilKeygen.GenerateEd25519KeyPair()
	default:
		return nil, fmt.Errorf("unsupported JWS algorithm: %s", algorithm)
	}
}

// generateJWEKey generates a symmetric encryption key for JWE tokens.
func (sm *SessionManager) generateJWEKey(algorithm string) (crypto.PrivateKey, error) {
	switch algorithm {
	case cryptoutilMagic.SessionJWEAlgorithmDirA256GCM,
		cryptoutilMagic.SessionJWEAlgorithmA256GCMKWA256GCM:
		// AES-256 key generation (32 bytes)
		return cryptoutilKeygen.GenerateAESKey(256)
	default:
		return nil, fmt.Errorf("unsupported JWE algorithm: %s", algorithm)
	}
}

// getAlgorithmIdentifier returns the specific algorithm identifier for session tokens.
func (sm *SessionManager) getAlgorithmIdentifier(isBrowser bool, sessionAlgorithm cryptoutilMagic.SessionAlgorithmType) string {
	switch sessionAlgorithm {
	case cryptoutilMagic.SessionAlgorithmJWS:
		if isBrowser {
			return sm.config.BrowserSessionJWSAlgorithm
		}
		return sm.config.ServiceSessionJWSAlgorithm
	case cryptoutilMagic.SessionAlgorithmJWE:
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
//   - realm: Realm identifier (optional, can be empty string)
//
// Returns session token string for client.
func (sm *SessionManager) IssueBrowserSession(ctx context.Context, userID, realm string) (string, error) {
	// Implementation will be added in next commit
	return "", errors.New("not implemented yet")
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
func (sm *SessionManager) ValidateBrowserSession(ctx context.Context, token string) (*cryptoutilRepository.BrowserSession, error) {
	// Implementation will be added in next commit
	return nil, errors.New("not implemented yet")
}

// IssueServiceSession issues a new service session token.
// Similar to IssueBrowserSession but for service-to-service authentication.
func (sm *SessionManager) IssueServiceSession(ctx context.Context, clientID, realm string) (string, error) {
	// Implementation will be added in next commit
	return "", errors.New("not implemented yet")
}

// ValidateServiceSession validates a service session token.
// Similar to ValidateBrowserSession but for service-to-service authentication.
func (sm *SessionManager) ValidateServiceSession(ctx context.Context, token string) (*cryptoutilRepository.ServiceSession, error) {
	// Implementation will be added in next commit
	return nil, errors.New("not implemented yet")
}

// CleanupExpiredSessions removes expired sessions from the database.
//
// Deletes sessions where:
//   - Expiration timestamp is in the past
//   - Last activity timestamp exceeds idle timeout
//
// Should be called periodically (e.g., every hour) to prevent database bloat.
func (sm *SessionManager) CleanupExpiredSessions(ctx context.Context) error {
	now := time.Now()
	idleThreshold := now.Add(-sm.config.SessionIdleTimeout)

	// Cleanup browser sessions
	err := sm.db.WithContext(ctx).
		Where("expiration < ? OR (last_activity IS NOT NULL AND last_activity < ?)", now, idleThreshold).
		Delete(&cryptoutilRepository.BrowserSession{}).
		Error
	if err != nil {
		return fmt.Errorf("failed to cleanup browser sessions: %w", err)
	}

	// Cleanup service sessions
	err = sm.db.WithContext(ctx).
		Where("expiration < ? OR (last_activity IS NOT NULL AND last_activity < ?)", now, idleThreshold).
		Delete(&cryptoutilRepository.ServiceSession{}).
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
