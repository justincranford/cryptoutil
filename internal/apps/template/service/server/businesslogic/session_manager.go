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
	"crypto/elliptic"
	json "encoding/json"
	"errors"
	"fmt"
	"time"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"gorm.io/gorm"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
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
	barrier *cryptoutilAppsTemplateServiceServerBarrier.Service
	config  *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings

	// Runtime state for browser sessions
	browserAlgorithm cryptoutilSharedMagic.SessionAlgorithmType
	browserJWKID     *googleUuid.UUID // Active JWK ID for JWS/JWE, nil for OPAQUE

	// Runtime state for service sessions
	serviceAlgorithm cryptoutilSharedMagic.SessionAlgorithmType
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
func NewSessionManager(db *gorm.DB, barrier *cryptoutilAppsTemplateServiceServerBarrier.Service, config *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) *SessionManager {
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
	browserAlg := cryptoutilSharedMagic.SessionAlgorithmType(sm.config.BrowserSessionAlgorithm)
	if browserAlg == "" {
		browserAlg = cryptoutilSharedMagic.DefaultBrowserSessionAlgorithm
	}

	sm.browserAlgorithm = browserAlg

	// Parse service session algorithm
	serviceAlg := cryptoutilSharedMagic.SessionAlgorithmType(sm.config.ServiceSessionAlgorithm)
	if serviceAlg == "" {
		serviceAlg = cryptoutilSharedMagic.DefaultServiceSessionAlgorithm
	}

	sm.serviceAlgorithm = serviceAlg

	// Initialize browser session JWKs if using JWS/JWE
	if browserAlg == cryptoutilSharedMagic.SessionAlgorithmJWS || browserAlg == cryptoutilSharedMagic.SessionAlgorithmJWE {
		jwkID, err := sm.initializeSessionJWK(ctx, true, browserAlg)
		if err != nil {
			return fmt.Errorf("failed to initialize browser session JWK: %w", err)
		}

		sm.browserJWKID = &jwkID
	}

	// Initialize service session JWKs if using JWS/JWE
	if serviceAlg == cryptoutilSharedMagic.SessionAlgorithmJWS || serviceAlg == cryptoutilSharedMagic.SessionAlgorithmJWE {
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
func (sm *SessionManager) initializeSessionJWK(ctx context.Context, isBrowser bool, algorithm cryptoutilSharedMagic.SessionAlgorithmType) (googleUuid.UUID, error) {
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
	case cryptoutilSharedMagic.SessionAlgorithmJWS:
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
				hmacBits = cryptoutilSharedMagic.HMACKeySize256
				algValue = joseJwa.HS256()
			case algIdentifierHS384:
				hmacBits = cryptoutilSharedMagic.HMACKeySize384
				algValue = joseJwa.HS384()
			case algIdentifierHS512:
				hmacBits = cryptoutilSharedMagic.HMACKeySize512
				algValue = joseJwa.HS512()
			}

			jwk, genErr = cryptoutilSharedCryptoJose.GenerateHMACJWK(hmacBits)
			if genErr == nil {
				genErr = jwk.Set(joseJwk.AlgorithmKey, algValue)
				if genErr == nil {
					genErr = jwk.Set(joseJwk.KeyUsageKey, joseJwk.ForSignature)
				}
			}
		case algIdentifierRS256, algIdentifierRS384, algIdentifierRS512:
			jwk, genErr = cryptoutilSharedCryptoJose.GenerateRSAJWK(cryptoutilSharedMagic.RSAKeySize2048)
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
			jwk, genErr = cryptoutilSharedCryptoJose.GenerateECDSAJWK(elliptic.P256())
			if genErr == nil {
				genErr = jwk.Set(joseJwk.AlgorithmKey, joseJwa.ES256())
				if genErr == nil {
					genErr = jwk.Set(joseJwk.KeyUsageKey, joseJwk.ForSignature)
				}
			}
		case algIdentifierES384:
			jwk, genErr = cryptoutilSharedCryptoJose.GenerateECDSAJWK(elliptic.P384())
			if genErr == nil {
				genErr = jwk.Set(joseJwk.AlgorithmKey, joseJwa.ES384())
				if genErr == nil {
					genErr = jwk.Set(joseJwk.KeyUsageKey, joseJwk.ForSignature)
				}
			}
		case algIdentifierES512:
			jwk, genErr = cryptoutilSharedCryptoJose.GenerateECDSAJWK(elliptic.P521())
			if genErr == nil {
				genErr = jwk.Set(joseJwk.AlgorithmKey, joseJwa.ES512())
				if genErr == nil {
					genErr = jwk.Set(joseJwk.KeyUsageKey, joseJwk.ForSignature)
				}
			}
		case algIdentifierEdDSA:
			jwk, genErr = cryptoutilSharedCryptoJose.GenerateEDDSAJWK("Ed25519")
			if genErr == nil {
				genErr = jwk.Set(joseJwk.AlgorithmKey, joseJwa.EdDSA())
				if genErr == nil {
					genErr = jwk.Set(joseJwk.KeyUsageKey, joseJwk.ForSignature)
				}
			}
		default:
			return googleUuid.UUID{}, fmt.Errorf("unsupported JWS algorithm: %s", algIdentifier)
		}
	case cryptoutilSharedMagic.SessionAlgorithmJWE:
		// Generate encryption JWK based on algorithm
		switch algIdentifier {
		case cryptoutilSharedMagic.SessionJWEAlgorithmDirA256GCM:
			jwk, genErr = cryptoutilSharedCryptoJose.GenerateAESJWK(cryptoutilSharedMagic.AESKeySize256)
			if genErr == nil {
				// Set 'enc' and 'alg' attributes for encryption
				genErr = jwk.Set("enc", joseJwa.A256GCM())
				if genErr == nil {
					genErr = jwk.Set(joseJwk.AlgorithmKey, joseJwa.DIRECT())
				}
			}
		case cryptoutilSharedMagic.SessionJWEAlgorithmA256GCMKWA256GCM:
			jwk, genErr = cryptoutilSharedCryptoJose.GenerateAESJWK(cryptoutilSharedMagic.AESKeySize256)
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
		CreatedAt:    time.Now().UTC(),
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
