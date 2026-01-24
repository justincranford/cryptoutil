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
	"encoding/json"
	"errors"
	"fmt"
	"time"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"gorm.io/gorm"

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilHash "cryptoutil/internal/shared/crypto/hash"
	cryptoutilJOSE "cryptoutil/internal/shared/crypto/jose"
	cryptoutilKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// Session validation error messages.
const (
	errMsgMissingInvalidExpClaim = "Missing or invalid exp claim"
	errMsgMissingInvalidJTIClaim = "Missing or invalid jti claim"
	errMsgSessionRevokedNotFound = "Session revoked or not found"
	errMsgInvalidSessionToken    = "Invalid session token"
	algIdentifierHS256           = "HS256"
	algIdentifierHS384           = "HS384"
	algIdentifierHS512           = "HS512"
	algIdentifierRS256           = "RS256"
	algIdentifierRS384           = "RS384"
	algIdentifierRS512           = "RS512"
	algIdentifierES256           = "ES256"
	algIdentifierES384           = "ES384"
	algIdentifierES512           = "ES512"
	algIdentifierEdDSA           = "EdDSA"
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
	// Determine table name
	var tableName string
	if isBrowser {
		tableName = "browser_session_jwks"
	} else {
		tableName = "service_session_jwks"
	}

	// Check for existing active JWK (deterministic selection using max timestamp)
	var existingJWK cryptoutilAppsTemplateServiceServerRepository.SessionJWK

	err := sm.db.WithContext(ctx).
		Table(tableName).
		Where("active = ?", true).
		Order("created_at DESC").
		First(&existingJWK).
		Error
	if err == nil {
		// Found existing active JWK with latest timestamp, use it
		return existingJWK.ID, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return googleUuid.UUID{}, fmt.Errorf("failed to query active JWK: %w", err)
	}

	// No active JWK found, generate new one
	var (
		jwk    joseJwk.Key
		genErr error
	)

	// Determine which JWK generation function to use based on algorithm
	algIdentifier := sm.getAlgorithmIdentifier(isBrowser, algorithm)

	switch algorithm {
	case cryptoutilMagic.SessionAlgorithmJWS:
		// Generate signing JWK based on algorithm
		switch algIdentifier {
		case algIdentifierHS256, algIdentifierHS384, algIdentifierHS512:
			// HMAC symmetric key algorithms
			var (
				hmacBits int
				algValue joseJwa.SignatureAlgorithm
			)

			switch algIdentifier {
			case algIdentifierHS256:
				hmacBits = cryptoutilMagic.HMACKeySize256
				algValue = joseJwa.HS256()
			case algIdentifierHS384:
				hmacBits = cryptoutilMagic.HMACKeySize384
				algValue = joseJwa.HS384()
			case algIdentifierHS512:
				hmacBits = cryptoutilMagic.HMACKeySize512
				algValue = joseJwa.HS512()
			}

			jwk, genErr = cryptoutilJOSE.GenerateHMACJWK(hmacBits)
			if genErr == nil {
				genErr = jwk.Set(joseJwk.AlgorithmKey, algValue)
				if genErr == nil {
					genErr = jwk.Set(joseJwk.KeyUsageKey, joseJwk.ForSignature)
				}
			}
		case algIdentifierRS256, algIdentifierRS384, algIdentifierRS512:
			jwk, genErr = cryptoutilJOSE.GenerateRSAJWK(cryptoutilMagic.RSAKeySize2048)
			if genErr == nil {
				// Set 'alg' attribute for signing
				var algValue joseJwa.SignatureAlgorithm

				switch algIdentifier {
				case algIdentifierRS256:
					algValue = joseJwa.RS256()
				case algIdentifierRS384:
					algValue = joseJwa.RS384()
				case algIdentifierRS512:
					algValue = joseJwa.RS512()
				}

				genErr = jwk.Set(joseJwk.AlgorithmKey, algValue)
				if genErr == nil {
					genErr = jwk.Set(joseJwk.KeyUsageKey, joseJwk.ForSignature)
				}
			}
		case algIdentifierES256:
			jwk, genErr = cryptoutilJOSE.GenerateECDSAJWK(elliptic.P256())
			if genErr == nil {
				genErr = jwk.Set(joseJwk.AlgorithmKey, joseJwa.ES256())
				if genErr == nil {
					genErr = jwk.Set(joseJwk.KeyUsageKey, joseJwk.ForSignature)
				}
			}
		case algIdentifierES384:
			jwk, genErr = cryptoutilJOSE.GenerateECDSAJWK(elliptic.P384())
			if genErr == nil {
				genErr = jwk.Set(joseJwk.AlgorithmKey, joseJwa.ES384())
				if genErr == nil {
					genErr = jwk.Set(joseJwk.KeyUsageKey, joseJwk.ForSignature)
				}
			}
		case algIdentifierES512:
			jwk, genErr = cryptoutilJOSE.GenerateECDSAJWK(elliptic.P521())
			if genErr == nil {
				genErr = jwk.Set(joseJwk.AlgorithmKey, joseJwa.ES512())
				if genErr == nil {
					genErr = jwk.Set(joseJwk.KeyUsageKey, joseJwk.ForSignature)
				}
			}
		case algIdentifierEdDSA:
			jwk, genErr = cryptoutilJOSE.GenerateEDDSAJWK("Ed25519")
			if genErr == nil {
				genErr = jwk.Set(joseJwk.AlgorithmKey, joseJwa.EdDSA())
				if genErr == nil {
					genErr = jwk.Set(joseJwk.KeyUsageKey, joseJwk.ForSignature)
				}
			}
		default:
			return googleUuid.UUID{}, fmt.Errorf("unsupported JWS algorithm: %s", algIdentifier)
		}
	case cryptoutilMagic.SessionAlgorithmJWE:
		// Generate encryption JWK based on algorithm
		switch algIdentifier {
		case cryptoutilMagic.SessionJWEAlgorithmDirA256GCM:
			jwk, genErr = cryptoutilJOSE.GenerateAESJWK(cryptoutilMagic.AESKeySize256)
			if genErr == nil {
				// Set 'enc' and 'alg' attributes for encryption
				genErr = jwk.Set("enc", joseJwa.A256GCM())
				if genErr == nil {
					genErr = jwk.Set(joseJwk.AlgorithmKey, joseJwa.DIRECT())
				}
			}
		case cryptoutilMagic.SessionJWEAlgorithmA256GCMKWA256GCM:
			jwk, genErr = cryptoutilJOSE.GenerateAESJWK(cryptoutilMagic.AESKeySize256)
			if genErr == nil {
				genErr = jwk.Set("enc", joseJwa.A256GCM())
				if genErr == nil {
					genErr = jwk.Set(joseJwk.AlgorithmKey, joseJwa.A256GCMKW())
				}
			}
		default:
			return googleUuid.UUID{}, fmt.Errorf("unsupported JWE algorithm: %s", algIdentifier)
		}
	default:
		return googleUuid.UUID{}, fmt.Errorf("unsupported session algorithm: %s", algorithm)
	}

	if genErr != nil {
		return googleUuid.UUID{}, fmt.Errorf("failed to generate JWK: %w", genErr)
	}

	// Marshal JWK to JSON bytes
	jwkBytes, marshalErr := json.Marshal(jwk)
	if marshalErr != nil {
		return googleUuid.UUID{}, fmt.Errorf("failed to marshal JWK: %w", marshalErr)
	}

	// Encrypt JWK with barrier service (skip encryption if no barrier service for tests)
	var (
		encryptedJWK []byte
		encryptErr   error
	)

	if sm.barrier != nil {
		encryptedJWK, encryptErr = sm.barrier.EncryptBytesWithContext(ctx, jwkBytes)
		if encryptErr != nil {
			return googleUuid.UUID{}, fmt.Errorf("failed to encrypt JWK: %w", encryptErr)
		}
	} else {
		// No barrier service (test mode) - store as plain text
		encryptedJWK = jwkBytes
	}

	// Store JWK in database (encrypted)
	jwkID := googleUuid.Must(googleUuid.NewV7())
	newJWK := cryptoutilAppsTemplateServiceServerRepository.SessionJWK{
		ID:           jwkID,
		EncryptedJWK: string(encryptedJWK),
		CreatedAt:    time.Now(),
		Algorithm:    algIdentifier,
		Active:       true, // Mark as active key for signing.
	}

	var createErr error

	if isBrowser {
		browserJWK := cryptoutilAppsTemplateServiceServerRepository.BrowserSessionJWK{SessionJWK: newJWK}
		createErr = sm.db.WithContext(ctx).Create(&browserJWK).Error
	} else {
		serviceJWK := cryptoutilAppsTemplateServiceServerRepository.ServiceSessionJWK{SessionJWK: newJWK}
		createErr = sm.db.WithContext(ctx).Create(&serviceJWK).Error
	}

	if createErr != nil {
		return googleUuid.UUID{}, fmt.Errorf("failed to store JWK: %w", createErr)
	}

	// After creating our JWK, re-query to get the canonical active JWK.
	// In multi-instance deployments, another instance might have created a JWK
	// at nearly the same time. By re-querying with ORDER BY created_at DESC,
	// all instances will converge on the same JWK (the one with latest timestamp).
	var canonicalJWK cryptoutilAppsTemplateServiceServerRepository.SessionJWK

	canonicalErr := sm.db.WithContext(ctx).
		Table(tableName).
		Where("active = ?", true).
		Order("created_at DESC").
		First(&canonicalJWK).
		Error
	if canonicalErr != nil {
		return googleUuid.UUID{}, fmt.Errorf("failed to query canonical JWK after creation: %w", canonicalErr)
	}

	return canonicalJWK.ID, nil
}

// generateSessionJWK generates a new private key for session tokens.
//
// For JWS: Generates asymmetric signing key (RSA, ECDSA, or EdDSA)
// For JWE: Generates symmetric encryption key (AES)
//
// Returns crypto.PrivateKey that can be converted to JWK.
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
		keyPair, err := cryptoutilKeygen.GenerateRSAKeyPair(cryptoutilMagic.RSAKeySize2048)
		if err != nil {
			return nil, fmt.Errorf("failed to generate RSA key pair: %w", err)
		}

		return keyPair.Private, nil
	case cryptoutilMagic.SessionJWSAlgorithmES256:
		// ECDSA P-256
		keyPair, err := cryptoutilKeygen.GenerateECDSAKeyPair(elliptic.P256())
		if err != nil {
			return nil, fmt.Errorf("failed to generate ECDSA P-256 key pair: %w", err)
		}

		return keyPair.Private, nil
	case cryptoutilMagic.SessionJWSAlgorithmES384:
		// ECDSA P-384
		keyPair, err := cryptoutilKeygen.GenerateECDSAKeyPair(elliptic.P384())
		if err != nil {
			return nil, fmt.Errorf("failed to generate ECDSA P-384 key pair: %w", err)
		}

		return keyPair.Private, nil
	case cryptoutilMagic.SessionJWSAlgorithmES512:
		// ECDSA P-521
		keyPair, err := cryptoutilKeygen.GenerateECDSAKeyPair(elliptic.P521())
		if err != nil {
			return nil, fmt.Errorf("failed to generate ECDSA P-521 key pair: %w", err)
		}

		return keyPair.Private, nil
	case cryptoutilMagic.SessionJWSAlgorithmEdDSA:
		// Ed25519
		keyPair, err := cryptoutilKeygen.GenerateEDDSAKeyPair(cryptoutilKeygen.EdCurveEd25519)
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
	case cryptoutilMagic.SessionJWEAlgorithmDirA256GCM,
		cryptoutilMagic.SessionJWEAlgorithmA256GCMKWA256GCM:
		// AES-256 key generation (32 bytes)
		key, err := cryptoutilKeygen.GenerateAESKey(cryptoutilMagic.AESKeySize256)
		if err != nil {
			return nil, fmt.Errorf("failed to generate AES key: %w", err)
		}

		return key, nil
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
//   - tenantID: Tenant identifier for multi-tenancy isolation
//   - realmID: Realm identifier within tenant
//
// Returns session token string for client.
func (sm *SessionManager) IssueBrowserSession(ctx context.Context, userID string, tenantID, realmID googleUuid.UUID) (string, error) {
	switch sm.browserAlgorithm {
	case cryptoutilMagic.SessionAlgorithmOPAQUE:
		return sm.issueOPAQUESession(ctx, true, userID, tenantID, realmID)
	case cryptoutilMagic.SessionAlgorithmJWS:
		return sm.issueJWSSession(ctx, true, userID, tenantID, realmID)
	case cryptoutilMagic.SessionAlgorithmJWE:
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
	case cryptoutilMagic.SessionAlgorithmOPAQUE:
		result, err = sm.validateOPAQUESession(ctx, true, token)
	case cryptoutilMagic.SessionAlgorithmJWS:
		result, err = sm.validateJWSSession(ctx, true, token)
	case cryptoutilMagic.SessionAlgorithmJWE:
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
	case cryptoutilMagic.SessionAlgorithmOPAQUE:
		return sm.issueOPAQUESession(ctx, false, clientID, tenantID, realmID)
	case cryptoutilMagic.SessionAlgorithmJWS:
		return sm.issueJWSSession(ctx, false, clientID, tenantID, realmID)
	case cryptoutilMagic.SessionAlgorithmJWE:
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
	case cryptoutilMagic.SessionAlgorithmOPAQUE:
		result, err = sm.validateOPAQUESession(ctx, false, token)
	case cryptoutilMagic.SessionAlgorithmJWS:
		result, err = sm.validateJWSSession(ctx, false, token)
	case cryptoutilMagic.SessionAlgorithmJWE:
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
	now := time.Now()
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
	tokenHash, err := cryptoutilHash.HashHighEntropyDeterministic(token)
	if err != nil {
		return "", fmt.Errorf("failed to hash session token: %w", err)
	}

	// Calculate expiration
	var expiration time.Time
	if isBrowser {
		expiration = time.Now().Add(sm.config.BrowserSessionExpiration)
	} else {
		expiration = time.Now().Add(sm.config.ServiceSessionExpiration)
	}

	// Create session record
	now := time.Now()
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
	tokenHash, err := cryptoutilHash.HashHighEntropyDeterministic(token)
	if err != nil {
		return nil, fmt.Errorf("failed to hash session token: %w", err)
	}

	// Look up session by token hash
	now := time.Now()

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
func (sm *SessionManager) issueJWSSession(ctx context.Context, isBrowser bool, principalID string, tenantID, realmID googleUuid.UUID) (string, error) {
	// Load JWK from database
	var jwkID googleUuid.UUID
	if isBrowser {
		jwkID = *sm.browserJWKID
	} else {
		jwkID = *sm.serviceJWKID
	}

	var (
		jwkBytes []byte
		loadErr  error
	)

	if isBrowser {
		var browserJWK cryptoutilAppsTemplateServiceServerRepository.BrowserSessionJWK

		loadErr = sm.db.WithContext(ctx).Where("id = ?", jwkID).First(&browserJWK).Error
		if loadErr == nil {
			jwkBytes = []byte(browserJWK.EncryptedJWK)
		}
	} else {
		var serviceJWK cryptoutilAppsTemplateServiceServerRepository.ServiceSessionJWK

		loadErr = sm.db.WithContext(ctx).Where("id = ?", jwkID).First(&serviceJWK).Error
		if loadErr == nil {
			jwkBytes = []byte(serviceJWK.EncryptedJWK)
		}
	}

	if loadErr != nil {
		return "", fmt.Errorf("failed to load session JWK: %w", loadErr)
	}

	// Decrypt JWK bytes with barrier service (skip decryption if no barrier service for tests)
	var decryptErr error
	if sm.barrier != nil {
		jwkBytes, decryptErr = sm.barrier.DecryptBytesWithContext(ctx, []byte(jwkBytes))
		if decryptErr != nil {
			return "", fmt.Errorf("failed to decrypt JWK: %w", decryptErr)
		}
	}
	// If no barrier service (test mode), jwkBytes are already plain text

	// Parse JWK from JSON and ensure 'alg' is properly typed for signing
	jwk, err := joseJwk.ParseKey(jwkBytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse JWK: %w", err)
	}
	// Algorithm type normalization no longer needed; extraction handles typing.

	// Create JWT claims
	jti := googleUuid.Must(googleUuid.NewV7())
	now := time.Now()

	var exp time.Time
	if isBrowser {
		exp = now.Add(sm.config.BrowserSessionExpiration)
	} else {
		exp = now.Add(sm.config.ServiceSessionExpiration)
	}

	claims := map[string]any{
		"jti":       jti.String(),
		"iat":       now.Unix(),
		"exp":       exp.Unix(),
		"sub":       principalID,
		"tenant_id": tenantID.String(),
		"realm_id":  realmID.String(),
	}

	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JWT claims: %w", err)
	}

	// Sign JWT
	_, jwsBytes, err := cryptoutilJOSE.SignBytes([]joseJwk.Key{jwk}, claimsBytes)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	// Store session metadata in database (for revocation)
	tokenHash, err := cryptoutilHash.HashHighEntropyDeterministic(jti.String())
	if err != nil {
		return "", fmt.Errorf("failed to hash jti: %w", err)
	}

	session := cryptoutilAppsTemplateServiceServerRepository.Session{
		ID:           jti,
		TenantID:     tenantID,
		RealmID:      realmID,
		TokenHash:    &tokenHash,
		Expiration:   exp,
		CreatedAt:    now,
		LastActivity: &now,
	}

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

	return string(jwsBytes), nil
}

// validateJWSSession validates a JWS session token.
func (sm *SessionManager) validateJWSSession(ctx context.Context, isBrowser bool, token string) (any, error) {
	// Load JWK from database
	var jwkID googleUuid.UUID
	if isBrowser {
		jwkID = *sm.browserJWKID
	} else {
		jwkID = *sm.serviceJWKID
	}

	var (
		jwkBytes []byte
		loadErr  error
	)

	if isBrowser {
		var browserJWK cryptoutilAppsTemplateServiceServerRepository.BrowserSessionJWK

		loadErr = sm.db.WithContext(ctx).Where("id = ?", jwkID).First(&browserJWK).Error
		if loadErr == nil {
			jwkBytes = []byte(browserJWK.EncryptedJWK)
		}
	} else {
		var serviceJWK cryptoutilAppsTemplateServiceServerRepository.ServiceSessionJWK

		loadErr = sm.db.WithContext(ctx).Where("id = ?", jwkID).First(&serviceJWK).Error
		if loadErr == nil {
			jwkBytes = []byte(serviceJWK.EncryptedJWK)
		}
	}

	if loadErr != nil {
		summary := "Failed to load session JWK"

		return nil, cryptoutilSharedApperr.NewHTTP401Unauthorized(&summary, loadErr)
	}

	// Decrypt JWK bytes with barrier service (skip decryption if no barrier service for tests)
	var (
		decryptedJWKBytes []byte
		decryptErr        error
	)

	if sm.barrier != nil {
		decryptedJWKBytes, decryptErr = sm.barrier.DecryptBytesWithContext(ctx, jwkBytes)
		if decryptErr != nil {
			summary := "Failed to decrypt JWK"

			return nil, cryptoutilSharedApperr.NewHTTP401Unauthorized(&summary, decryptErr)
		}
	} else {
		// No barrier service (test mode) - jwkBytes are already plain text
		decryptedJWKBytes = jwkBytes
	}

	// Parse JWK from JSON
	privateJWK, err := joseJwk.ParseKey(decryptedJWKBytes)
	if err != nil {
		summary := "Failed to parse session JWK"

		return nil, cryptoutilSharedApperr.NewHTTP401Unauthorized(&summary, err)
	}

	// Extract public key from private JWK for verification
	publicJWK, err := privateJWK.PublicKey()
	if err != nil {
		summary := "Failed to extract public key from JWK"

		return nil, cryptoutilSharedApperr.NewHTTP401Unauthorized(&summary, err)
	}
	// No normalization required; verification utilities will validate algorithm type.

	// Verify JWT signature
	claimsBytes, err := cryptoutilJOSE.VerifyBytes([]joseJwk.Key{publicJWK}, []byte(token))
	if err != nil {
		summary := "Invalid JWT signature"

		return nil, cryptoutilSharedApperr.NewHTTP401Unauthorized(&summary, err)
	}

	// Parse and validate claims
	var claims map[string]any
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		summary := "Failed to parse JWT claims"

		return nil, cryptoutilSharedApperr.NewHTTP401Unauthorized(&summary, err)
	}

	// Validate expiration
	expFloat, ok := claims["exp"].(float64)
	if !ok {
		summary := errMsgMissingInvalidExpClaim

		return nil, cryptoutilSharedApperr.NewHTTP401Unauthorized(&summary, fmt.Errorf("exp claim not found"))
	}

	exp := time.Unix(int64(expFloat), 0)
	if time.Now().After(exp) {
		summary := "JWT expired"

		return nil, cryptoutilSharedApperr.NewHTTP401Unauthorized(&summary, fmt.Errorf("token expired at %v", exp))
	}

	// Extract jti and validate against database
	jtiStr, ok := claims["jti"].(string)
	if !ok {
		summary := errMsgMissingInvalidJTIClaim

		return nil, cryptoutilSharedApperr.NewHTTP401Unauthorized(&summary, fmt.Errorf("jti claim not found"))
	}

	jti, err := googleUuid.Parse(jtiStr)
	if err != nil {
		summary := "Invalid jti format"

		return nil, cryptoutilSharedApperr.NewHTTP401Unauthorized(&summary, err)
	}

	// Hash jti for database lookup
	tokenHash, err := cryptoutilHash.HashHighEntropyDeterministic(jti.String())
	if err != nil {
		return nil, fmt.Errorf("failed to hash jti: %w", err)
	}

	// Look up session by token hash
	now := time.Now()

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
			summary := errMsgSessionRevokedNotFound

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

// issueJWESession issues a JWE session token (encrypted JWT).
func (sm *SessionManager) issueJWESession(ctx context.Context, isBrowser bool, principalID string, tenantID, realmID googleUuid.UUID) (string, error) {
	// Load JWK from database
	var jwkID googleUuid.UUID
	if isBrowser {
		jwkID = *sm.browserJWKID
	} else {
		jwkID = *sm.serviceJWKID
	}

	var (
		jwkBytes []byte
		loadErr  error
	)

	if isBrowser {
		var browserJWK cryptoutilAppsTemplateServiceServerRepository.BrowserSessionJWK

		loadErr = sm.db.WithContext(ctx).Where("id = ?", jwkID).First(&browserJWK).Error
		if loadErr == nil {
			jwkBytes = []byte(browserJWK.EncryptedJWK)
		}
	} else {
		var serviceJWK cryptoutilAppsTemplateServiceServerRepository.ServiceSessionJWK

		loadErr = sm.db.WithContext(ctx).Where("id = ?", jwkID).First(&serviceJWK).Error
		if loadErr == nil {
			jwkBytes = []byte(serviceJWK.EncryptedJWK)
		}
	}

	if loadErr != nil {
		return "", fmt.Errorf("failed to load session JWK: %w", loadErr)
	}

	// Decrypt JWK with barrier service (skip decryption if no barrier service for tests)
	var (
		decryptedJWKBytes []byte
		decryptErr        error
	)

	if sm.barrier != nil {
		decryptedJWKBytes, decryptErr = sm.barrier.DecryptBytesWithContext(ctx, jwkBytes)
		if decryptErr != nil {
			return "", fmt.Errorf("failed to decrypt session JWK: %w", decryptErr)
		}
	} else {
		// No barrier service (test mode) - jwkBytes are already plain text
		decryptedJWKBytes = jwkBytes
	}

	// Parse JWK from JSON bytes
	jwk, parseErr := joseJwk.ParseKey(decryptedJWKBytes)
	if parseErr != nil {
		return "", fmt.Errorf("failed to parse JWK: %w", parseErr)
	}

	// Create JWT claims
	now := time.Now()

	var exp time.Time
	if isBrowser {
		exp = now.Add(sm.config.BrowserSessionExpiration)
	} else {
		exp = now.Add(sm.config.ServiceSessionExpiration)
	}

	jti := googleUuid.Must(googleUuid.NewV7())

	claims := map[string]any{
		"jti":       jti.String(),
		"iat":       now.Unix(),
		"exp":       exp.Unix(),
		"sub":       principalID,
		"tenant_id": tenantID.String(),
		"realm_id":  realmID.String(),
	}

	claimsBytes, marshalErr := json.Marshal(claims)
	if marshalErr != nil {
		return "", fmt.Errorf("failed to marshal JWT claims: %w", marshalErr)
	}

	// Encrypt JWT claims with JWK
	_, jweBytes, encryptErr := cryptoutilJOSE.EncryptBytes([]joseJwk.Key{jwk}, claimsBytes)
	if encryptErr != nil {
		return "", fmt.Errorf("failed to encrypt JWT: %w", encryptErr)
	}

	// Hash jti for database storage (enables revocation)
	tokenHash, hashErr := cryptoutilHash.HashHighEntropyDeterministic(jti.String())
	if hashErr != nil {
		return "", fmt.Errorf("failed to hash jti: %w", hashErr)
	}

	// Store session metadata in database
	session := cryptoutilAppsTemplateServiceServerRepository.Session{
		ID:           jti,
		TenantID:     tenantID,
		RealmID:      realmID,
		TokenHash:    &tokenHash,
		Expiration:   exp,
		CreatedAt:    now,
		LastActivity: &now,
	}

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
		return "", fmt.Errorf("failed to create session: %w", createErr)
	}

	return string(jweBytes), nil
}

// validateJWESession validates a JWE session token.
func (sm *SessionManager) validateJWESession(ctx context.Context, isBrowser bool, token string) (any, error) {
	// Load JWK from database
	var jwkID googleUuid.UUID
	if isBrowser {
		jwkID = *sm.browserJWKID
	} else {
		jwkID = *sm.serviceJWKID
	}

	var (
		jwkBytes []byte
		loadErr  error
	)

	if isBrowser {
		var browserJWK cryptoutilAppsTemplateServiceServerRepository.BrowserSessionJWK

		loadErr = sm.db.WithContext(ctx).Where("id = ?", jwkID).First(&browserJWK).Error
		if loadErr == nil {
			jwkBytes = []byte(browserJWK.EncryptedJWK)
		}
	} else {
		var serviceJWK cryptoutilAppsTemplateServiceServerRepository.ServiceSessionJWK

		loadErr = sm.db.WithContext(ctx).Where("id = ?", jwkID).First(&serviceJWK).Error
		if loadErr == nil {
			jwkBytes = []byte(serviceJWK.EncryptedJWK)
		}
	}

	if loadErr != nil {
		summary := errMsgInvalidSessionToken

		return nil, cryptoutilSharedApperr.NewHTTP401Unauthorized(&summary, loadErr)
	}

	// Decrypt JWK with barrier service (skip decryption if no barrier service for tests)
	var (
		decryptedJWKBytes []byte
		decryptErr        error
	)

	if sm.barrier != nil {
		decryptedJWKBytes, decryptErr = sm.barrier.DecryptBytesWithContext(ctx, jwkBytes)
		if decryptErr != nil {
			summary := errMsgInvalidSessionToken

			return nil, cryptoutilSharedApperr.NewHTTP401Unauthorized(&summary, decryptErr)
		}
	} else {
		// No barrier service (test mode) - jwkBytes are already plain text
		decryptedJWKBytes = jwkBytes
	}

	// Parse JWK from JSON bytes
	jwk, parseErr := joseJwk.ParseKey(decryptedJWKBytes)
	if parseErr != nil {
		summary := errMsgInvalidSessionToken

		return nil, cryptoutilSharedApperr.NewHTTP401Unauthorized(&summary, parseErr)
	}

	// Decrypt and verify JWT
	claimsBytes, verifyErr := cryptoutilJOSE.DecryptBytes([]joseJwk.Key{jwk}, []byte(token))
	if verifyErr != nil {
		summary := errMsgInvalidSessionToken

		return nil, cryptoutilSharedApperr.NewHTTP401Unauthorized(&summary, verifyErr)
	}

	// Parse JWT claims
	var claims map[string]any

	unmarshalErr := json.Unmarshal(claimsBytes, &claims)
	if unmarshalErr != nil {
		summary := errMsgInvalidSessionToken

		return nil, cryptoutilSharedApperr.NewHTTP401Unauthorized(&summary, unmarshalErr)
	}

	// Validate expiration claim
	expFloat, expOk := claims["exp"].(float64)
	if !expOk {
		summary := errMsgMissingInvalidExpClaim

		return nil, cryptoutilSharedApperr.NewHTTP401Unauthorized(&summary, fmt.Errorf("exp claim not found or invalid type"))
	}

	exp := time.Unix(int64(expFloat), 0)

	now := time.Now()
	if now.After(exp) {
		summary := "Session expired"

		return nil, cryptoutilSharedApperr.NewHTTP401Unauthorized(&summary, fmt.Errorf("token expired at %v", exp))
	}

	// Extract jti (token ID) and hash it for database lookup
	jtiStr, jtiOk := claims["jti"].(string)
	if !jtiOk {
		summary := errMsgMissingInvalidJTIClaim

		return nil, cryptoutilSharedApperr.NewHTTP401Unauthorized(&summary, fmt.Errorf("jti claim not found or invalid type"))
	}

	jti, parseJtiErr := googleUuid.Parse(jtiStr)
	if parseJtiErr != nil {
		summary := "Invalid jti claim format"

		return nil, cryptoutilSharedApperr.NewHTTP401Unauthorized(&summary, parseJtiErr)
	}

	// Look up session in database by jti (enables revocation)
	var (
		session any
		findErr error
	)

	if isBrowser {
		browserSession := &cryptoutilAppsTemplateServiceServerRepository.BrowserSession{}
		findErr = sm.db.WithContext(ctx).
			Where("id = ?", jti).
			First(browserSession).
			Error
		session = browserSession
	} else {
		serviceSession := &cryptoutilAppsTemplateServiceServerRepository.ServiceSession{}
		findErr = sm.db.WithContext(ctx).
			Where("id = ?", jti).
			First(serviceSession).
			Error
		session = serviceSession
	}

	if findErr != nil {
		if errors.Is(findErr, gorm.ErrRecordNotFound) {
			summary := errMsgSessionRevokedNotFound

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
