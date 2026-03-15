// Copyright (c) 2025 Justin Cranford
//
//

package mfa

import (
	"context"
	json "encoding/json"
	"fmt"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// WebAuthnService handles WebAuthn registration and authentication ceremonies.
type WebAuthnService struct {
	db       *gorm.DB
	webauthn *webauthn.WebAuthn
}

// WebAuthnConfig holds configuration for WebAuthn.
type WebAuthnConfig struct {
	RPDisplayName string
	RPID          string
	RPOrigins     []string
}

// NewWebAuthnService creates a new WebAuthn service.
func NewWebAuthnService(db *gorm.DB, config WebAuthnConfig) (*WebAuthnService, error) {
	wconfig := &webauthn.Config{
		RPDisplayName: config.RPDisplayName,
		RPID:          config.RPID,
		RPOrigins:     config.RPOrigins,
	}

	w, err := webauthn.New(wconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create webauthn: %w", err)
	}

	return &WebAuthnService{
		db:       db,
		webauthn: w,
	}, nil
}

// WebAuthnUser implements webauthn.User interface.
type WebAuthnUser struct {
	ID          googleUuid.UUID
	Name        string
	DisplayName string
	Credentials []WebAuthnCredential
}

// WebAuthnID returns the user ID as bytes.
func (u *WebAuthnUser) WebAuthnID() []byte {
	return u.ID[:]
}

// WebAuthnName returns the username.
func (u *WebAuthnUser) WebAuthnName() string {
	return u.Name
}

// WebAuthnDisplayName returns the display name.
func (u *WebAuthnUser) WebAuthnDisplayName() string {
	return u.DisplayName
}

// WebAuthnCredentials returns the user credentials.
func (u *WebAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	creds := make([]webauthn.Credential, len(u.Credentials))
	for i, c := range u.Credentials {
		creds[i] = credentialToWebAuthn(c)
	}

	return creds
}

// credentialToWebAuthn converts a stored credential to webauthn.Credential.
func credentialToWebAuthn(c WebAuthnCredential) webauthn.Credential {
	var transports []protocol.AuthenticatorTransport
	if c.Transports != "" {
		_ = json.Unmarshal([]byte(c.Transports), &transports)
	}

	return webauthn.Credential{
		ID:              c.CredentialID,
		PublicKey:       c.PublicKey,
		AttestationType: c.AttestationType,
		Transport:       transports,
		Authenticator: webauthn.Authenticator{
			AAGUID:       c.AAGUID,
			SignCount:    c.SignCount,
			CloneWarning: c.CloneWarning,
		},
	}
}

// BeginRegistration starts a WebAuthn registration ceremony.
func (s *WebAuthnService) BeginRegistration(ctx context.Context, user *WebAuthnUser) (*protocol.CredentialCreation, googleUuid.UUID, error) {
	options, session, err := s.webauthn.BeginRegistration(user)
	if err != nil {
		return nil, googleUuid.Nil, fmt.Errorf("failed to begin registration: %w", err)
	}

	sessionData, err := json.Marshal(session)
	if err != nil {
		return nil, googleUuid.Nil, fmt.Errorf("failed to marshal session: %w", err)
	}

	webauthnSession := &WebAuthnSession{
		UserID:       user.ID,
		SessionData:  sessionData,
		CeremonyType: string(WebAuthnCeremonyRegistration),
		ExpiresAt:    time.Now().UTC().Add(cryptoutilSharedMagic.DefaultWebAuthnSessionExpiry),
	}

	if err := s.db.WithContext(ctx).Create(webauthnSession).Error; err != nil {
		return nil, googleUuid.Nil, fmt.Errorf("failed to save session: %w", err)
	}

	return options, webauthnSession.ID, nil
}

// FinishRegistration completes a WebAuthn registration ceremony.
func (s *WebAuthnService) FinishRegistration(ctx context.Context, user *WebAuthnUser, sessionID googleUuid.UUID, response *protocol.ParsedCredentialCreationData, displayName string) (*WebAuthnCredential, error) {
	var session WebAuthnSession
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ? AND ceremony_type = ?", sessionID, user.ID, WebAuthnCeremonyRegistration).First(&session).Error; err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if session.IsExpired() {
		_ = s.db.WithContext(ctx).Delete(&session)

		return nil, fmt.Errorf("session expired")
	}

	var webauthnSession webauthn.SessionData
	if err := json.Unmarshal(session.SessionData, &webauthnSession); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	credential, err := s.webauthn.CreateCredential(user, webauthnSession, response)
	if err != nil {
		return nil, fmt.Errorf("failed to create credential: %w", err)
	}

	transportsJSON, _ := json.Marshal(credential.Transport)

	webauthnCred := &WebAuthnCredential{
		UserID:          user.ID,
		CredentialID:    credential.ID,
		PublicKey:       credential.PublicKey,
		AttestationType: credential.AttestationType,
		Transports:      string(transportsJSON),
		SignCount:       credential.Authenticator.SignCount,
		AAGUID:          credential.Authenticator.AAGUID,
		CloneWarning:    credential.Authenticator.CloneWarning,
		DisplayName:     displayName,
	}

	if err := s.db.WithContext(ctx).Create(webauthnCred).Error; err != nil {
		return nil, fmt.Errorf("failed to save credential: %w", err)
	}

	_ = s.db.WithContext(ctx).Delete(&session)

	return webauthnCred, nil
}

// BeginAuthentication starts a WebAuthn authentication ceremony.
func (s *WebAuthnService) BeginAuthentication(ctx context.Context, user *WebAuthnUser) (*protocol.CredentialAssertion, googleUuid.UUID, error) {
	options, session, err := s.webauthn.BeginLogin(user)
	if err != nil {
		return nil, googleUuid.Nil, fmt.Errorf("failed to begin login: %w", err)
	}

	sessionData, err := json.Marshal(session)
	if err != nil {
		return nil, googleUuid.Nil, fmt.Errorf("failed to marshal session: %w", err)
	}

	webauthnSession := &WebAuthnSession{
		UserID:       user.ID,
		SessionData:  sessionData,
		CeremonyType: string(WebAuthnCeremonyAuthentication),
		ExpiresAt:    time.Now().UTC().Add(cryptoutilSharedMagic.DefaultWebAuthnSessionExpiry),
	}

	if err := s.db.WithContext(ctx).Create(webauthnSession).Error; err != nil {
		return nil, googleUuid.Nil, fmt.Errorf("failed to save session: %w", err)
	}

	return options, webauthnSession.ID, nil
}

// FinishAuthentication completes a WebAuthn authentication ceremony.
func (s *WebAuthnService) FinishAuthentication(ctx context.Context, user *WebAuthnUser, sessionID googleUuid.UUID, response *protocol.ParsedCredentialAssertionData) (*WebAuthnCredential, error) {
	var session WebAuthnSession
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ? AND ceremony_type = ?", sessionID, user.ID, WebAuthnCeremonyAuthentication).First(&session).Error; err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if session.IsExpired() {
		_ = s.db.WithContext(ctx).Delete(&session)

		return nil, fmt.Errorf("session expired")
	}

	var webauthnSession webauthn.SessionData
	if err := json.Unmarshal(session.SessionData, &webauthnSession); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	credential, err := s.webauthn.ValidateLogin(user, webauthnSession, response)
	if err != nil {
		return nil, fmt.Errorf("failed to validate login: %w", err)
	}

	var storedCred WebAuthnCredential
	if err := s.db.WithContext(ctx).Where("credential_id = ? AND user_id = ?", credential.ID, user.ID).First(&storedCred).Error; err != nil {
		return nil, fmt.Errorf("credential not found: %w", err)
	}

	now := time.Now().UTC()
	storedCred.SignCount = credential.Authenticator.SignCount
	storedCred.CloneWarning = credential.Authenticator.CloneWarning
	storedCred.LastUsedAt = &now

	if err := s.db.WithContext(ctx).Save(&storedCred).Error; err != nil {
		return nil, fmt.Errorf("failed to update credential: %w", err)
	}

	_ = s.db.WithContext(ctx).Delete(&session)

	return &storedCred, nil
}

// GetCredentials returns all WebAuthn credentials for a user.
func (s *WebAuthnService) GetCredentials(ctx context.Context, userID googleUuid.UUID) ([]WebAuthnCredential, error) {
	var credentials []WebAuthnCredential
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&credentials).Error; err != nil {
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	return credentials, nil
}

// DeleteCredential removes a WebAuthn credential.
func (s *WebAuthnService) DeleteCredential(ctx context.Context, userID, credentialID googleUuid.UUID) error {
	result := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", credentialID, userID).Delete(&WebAuthnCredential{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete credential: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("credential not found")
	}

	return nil
}

// CleanupExpiredSessions removes expired WebAuthn sessions.
func (s *WebAuthnService) CleanupExpiredSessions(ctx context.Context) error {
	if err := s.db.WithContext(ctx).Where("expires_at < ?", time.Now().UTC()).Delete(&WebAuthnSession{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup sessions: %w", err)
	}

	return nil
}
