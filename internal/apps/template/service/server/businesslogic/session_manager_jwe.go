// Copyright (c) 2025 Justin Cranford
//
//

// TODO(sm-im-migration): This SessionManager implementation is work-in-progress.
// Current status:
// - ✅ OPAQUE session issuance and validation (uses hash package directly)
// - ✅ JWS/JWE session issuance and validation (complete implementation)
// - ✅ JWK encryption with barrier service (fully implemented)
// - ✅ Comprehensive unit tests (all 25 tests passing)
//
// Next steps:
// - Write integration tests
// - Verify quality gates (coverage ≥95%, mutations ≥85%)
// - Integrate with sm-im service

// Package businesslogic provides business logic services for the template service.
package businesslogic

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
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
	decryptedJWKBytes, decryptErr := barrierDecryptFn(ctx, sm.barrier, jwkBytes)
	if decryptErr != nil {
		return "", fmt.Errorf("failed to decrypt session JWK: %w", decryptErr)
	}

	// Parse JWK from JSON bytes
	jwk, parseErr := jwkParseKeyFn(decryptedJWKBytes)
	if parseErr != nil {
		return "", fmt.Errorf("failed to parse JWK: %w", parseErr)
	}

	// Create JWT claims
	now := time.Now().UTC()

	var exp time.Time
	if isBrowser {
		exp = now.Add(sm.config.BrowserSessionExpiration)
	} else {
		exp = now.Add(sm.config.ServiceSessionExpiration)
	}

	jti := googleUuid.Must(googleUuid.NewV7())

	claims := map[string]any{
		cryptoutilSharedMagic.ClaimJti: jti.String(),
		cryptoutilSharedMagic.ClaimIat: now.Unix(),
		cryptoutilSharedMagic.ClaimExp: exp.Unix(),
		cryptoutilSharedMagic.ClaimSub: principalID,
		"tenant_id":                    tenantID.String(),
		"realm_id":                     realmID.String(),
	}

	claimsBytes, marshalErr := jsonMarshalFn(claims)
	if marshalErr != nil {
		return "", fmt.Errorf("failed to marshal JWT claims: %w", marshalErr)
	}

	// Encrypt JWT claims with JWK
	_, jweBytes, encryptErr := encryptBytesFn([]joseJwk.Key{jwk}, claimsBytes)
	if encryptErr != nil {
		return "", fmt.Errorf("failed to encrypt JWT: %w", encryptErr)
	}

	// Hash jti for database storage (enables revocation)
	tokenHash, hashErr := hashHighEntropyDeterministicFn(jti.String())
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
	decryptedJWKBytes, decryptErr := barrierDecryptFn(ctx, sm.barrier, jwkBytes)
	if decryptErr != nil {
		summary := errMsgInvalidSessionToken

		return nil, cryptoutilSharedApperr.NewHTTP401Unauthorized(&summary, decryptErr)
	}

	// Parse JWK from JSON bytes
	jwk, parseErr := jwkParseKeyFn(decryptedJWKBytes)
	if parseErr != nil {
		summary := errMsgInvalidSessionToken

		return nil, cryptoutilSharedApperr.NewHTTP401Unauthorized(&summary, parseErr)
	}

	// Decrypt and verify JWT
	claimsBytes, verifyErr := decryptBytesFn([]joseJwk.Key{jwk}, []byte(token))
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
	expFloat, expOk := claims[cryptoutilSharedMagic.ClaimExp].(float64)
	if !expOk {
		summary := errMsgMissingInvalidExpClaim

		return nil, cryptoutilSharedApperr.NewHTTP401Unauthorized(&summary, fmt.Errorf("exp claim not found or invalid type"))
	}

	exp := time.Unix(int64(expFloat), 0)

	now := time.Now().UTC()
	if now.After(exp) {
		summary := "Session expired"

		return nil, cryptoutilSharedApperr.NewHTTP401Unauthorized(&summary, fmt.Errorf("token expired at %v", exp))
	}

	// Extract jti (token ID) and hash it for database lookup
	jtiStr, jtiOk := claims[cryptoutilSharedMagic.ClaimJti].(string)
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
