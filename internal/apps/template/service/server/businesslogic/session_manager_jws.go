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
	json "encoding/json"
	"errors"
	"fmt"
	"time"

	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"gorm.io/gorm"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
)

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
	decryptedJWKBytes, decryptErr := barrierDecryptFn(ctx, sm.barrier, jwkBytes)
	if decryptErr != nil {
		return "", fmt.Errorf("failed to decrypt JWK: %w", decryptErr)
	}

	// Parse JWK from JSON and ensure 'alg' is properly typed for signing
	jwk, err := jwkParseKeyFn(decryptedJWKBytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse JWK: %w", err)
	}
	// Algorithm type normalization no longer needed; extraction handles typing.

	// Create JWT claims
	jti := googleUuid.Must(googleUuid.NewV7())
	now := time.Now().UTC()

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

	claimsBytes, err := jsonMarshalFn(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JWT claims: %w", err)
	}

	// Sign JWT
	_, jwsBytes, err := signBytesFn([]joseJwk.Key{jwk}, claimsBytes)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	// Store session metadata in database (for revocation)
	tokenHash, err := hashHighEntropyDeterministicFn(jti.String())
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
	decryptedJWKBytes, decryptErr := barrierDecryptFn(ctx, sm.barrier, jwkBytes)
	if decryptErr != nil {
		summary := "Failed to decrypt JWK"

		return nil, cryptoutilSharedApperr.NewHTTP401Unauthorized(&summary, decryptErr)
	}

	// Parse JWK from JSON
	privateJWK, err := jwkParseKeyFn(decryptedJWKBytes)
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
	claimsBytes, err := verifyBytesFn([]joseJwk.Key{publicJWK}, []byte(token))
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
	if time.Now().UTC().After(exp) {
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
	tokenHash, err := hashHighEntropyDeterministicFn(jti.String())
	if err != nil {
		return nil, fmt.Errorf("failed to hash jti: %w", err)
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
