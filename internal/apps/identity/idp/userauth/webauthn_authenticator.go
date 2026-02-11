// Copyright (c) 2025 Justin Cranford
//
//

package userauth

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	googleUuid "github.com/google/uuid"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"
)

// CredentialType represents the type of WebAuthn credential.
type CredentialType string

// Credential type constants.
const (
	// CredentialTypePasskey indicates a passkey credential.
	CredentialTypePasskey CredentialType = "passkey"
)

// Credential represents a WebAuthn/FIDO2 credential.
type Credential struct {
	ID              string
	UserID          string
	Type            CredentialType
	PublicKey       []byte
	AttestationType string
	AAGUID          []byte
	SignCount       uint32
	CreatedAt       time.Time
	LastUsedAt      time.Time
	Metadata        map[string]any
}

// CredentialStore manages WebAuthn/FIDO2 credentials.
type CredentialStore interface {
	StoreCredential(ctx context.Context, credential *Credential) error
	GetCredential(ctx context.Context, credentialID string) (*Credential, error)
	GetUserCredentials(ctx context.Context, userID string) ([]*Credential, error)
	DeleteCredential(ctx context.Context, credentialID string) error
}

// WebAuthnUser adapts cryptoutilIdentityDomain.User to webauthn.User interface.
type WebAuthnUser struct {
	user        *cryptoutilIdentityDomain.User
	credentials []webauthn.Credential
}

// WebAuthnID returns user ID as byte slice for WebAuthn protocol.
func (u *WebAuthnUser) WebAuthnID() []byte {
	return []byte(u.user.ID.String())
}

// WebAuthnName returns user's preferred username.
func (u *WebAuthnUser) WebAuthnName() string {
	return u.user.PreferredUsername
}

// WebAuthnDisplayName returns user's display name.
func (u *WebAuthnUser) WebAuthnDisplayName() string {
	if u.user.Name != "" {
		return u.user.Name
	}

	return u.user.PreferredUsername
}

// WebAuthnIcon returns user's avatar icon URL (optional).
func (u *WebAuthnUser) WebAuthnIcon() string {
	return "" // Optional: implement avatar URLs later.
}

// WebAuthnCredentials returns user's registered WebAuthn credentials.
func (u *WebAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	return u.credentials
}

// WebAuthnConfig holds WebAuthn configuration.
type WebAuthnConfig struct {
	RPID          string
	RPDisplayName string
	RPOrigins     []string
	Timeout       time.Duration
}

// WebAuthnAuthenticator implements WebAuthn/FIDO2 authentication using go-webauthn library.
type WebAuthnAuthenticator struct {
	webauthn        *webauthn.WebAuthn
	credentialStore CredentialStore
	challengeStore  ChallengeStore
	config          *WebAuthnConfig
}

// NewWebAuthnAuthenticator creates a new WebAuthn authenticator.
func NewWebAuthnAuthenticator(
	config *WebAuthnConfig,
	credentialStore CredentialStore,
	challengeStore ChallengeStore,
) (*WebAuthnAuthenticator, error) {
	if config == nil {
		return nil, fmt.Errorf("webAuthn config cannot be nil")
	}

	if credentialStore == nil {
		return nil, fmt.Errorf("credential store cannot be nil")
	}

	if challengeStore == nil {
		return nil, fmt.Errorf("challenge store cannot be nil")
	}

	if config.Timeout == 0 {
		config.Timeout = cryptoutilIdentityMagic.DefaultOTPLifetime
	}

	// Create WebAuthn instance.
	wconfig := &webauthn.Config{
		RPID:                  config.RPID,
		RPDisplayName:         config.RPDisplayName,
		RPOrigins:             config.RPOrigins,
		AttestationPreference: protocol.PreferNoAttestation,
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			RequireResidentKey: protocol.ResidentKeyNotRequired(),
			ResidentKey:        protocol.ResidentKeyRequirementDiscouraged,
			UserVerification:   protocol.VerificationPreferred,
		},
	}

	w, err := webauthn.New(wconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create WebAuthn instance: %w", err)
	}

	return &WebAuthnAuthenticator{
		webauthn:        w,
		credentialStore: credentialStore,
		challengeStore:  challengeStore,
		config:          config,
	}, nil
}

// Method returns the authentication method name.
func (w *WebAuthnAuthenticator) Method() string {
	return "passkey_webauthn"
}

// BeginRegistration begins WebAuthn credential registration ceremony.
func (w *WebAuthnAuthenticator) BeginRegistration(ctx context.Context, user *cryptoutilIdentityDomain.User) (*protocol.CredentialCreation, error) {
	if user == nil {
		return nil, fmt.Errorf("user cannot be nil")
	}

	// Load existing credentials for this user.
	storedCreds, err := w.credentialStore.GetUserCredentials(ctx, user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get user credentials: %w", err)
	}

	// Convert stored credentials to webauthn.Credential format.
	webauthnCreds := make([]webauthn.Credential, 0, len(storedCreds))

	for _, cred := range storedCreds {
		webauthnCreds = append(webauthnCreds, webauthn.Credential{
			ID:              []byte(cred.ID),
			PublicKey:       cred.PublicKey,
			AttestationType: cred.AttestationType,
			Authenticator: webauthn.Authenticator{
				AAGUID:    cred.AAGUID,
				SignCount: cred.SignCount,
			},
		})
	}

	// Create WebAuthnUser adapter.
	webauthnUser := &WebAuthnUser{
		user:        user,
		credentials: webauthnCreds,
	}

	// Begin registration using go-webauthn library.
	creation, session, err := w.webauthn.BeginRegistration(webauthnUser)
	if err != nil {
		return nil, fmt.Errorf("failed to begin WebAuthn registration: %w", err)
	}

	// Store session as challenge.
	challengeID, err := googleUuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate challenge ID: %w", err)
	}

	challenge := &AuthChallenge{
		ID:        challengeID,
		UserID:    user.ID.String(),
		Method:    "passkey_webauthn",
		ExpiresAt: time.Now().UTC().Add(w.config.Timeout),
		Metadata: map[string]any{
			"session_data": session.Challenge,
			"operation":    "registration",
			"user_handle":  session.UserID,
		},
	}

	if err := w.challengeStore.Store(ctx, challenge, challengeID.String()); err != nil {
		return nil, fmt.Errorf("failed to store WebAuthn challenge: %w", err)
	}

	return creation, nil
}

// FinishRegistration completes WebAuthn credential registration ceremony.
func (w *WebAuthnAuthenticator) FinishRegistration(
	ctx context.Context,
	user *cryptoutilIdentityDomain.User,
	challengeID string,
	credentialCreationResponse *protocol.ParsedCredentialCreationData,
) (*Credential, error) {
	if user == nil {
		return nil, fmt.Errorf("user cannot be nil")
	}

	if credentialCreationResponse == nil {
		return nil, fmt.Errorf("credential creation response cannot be nil")
	}

	// Parse challenge ID.
	id, err := googleUuid.Parse(challengeID)
	if err != nil {
		return nil, fmt.Errorf("invalid challenge ID: %w", err)
	}

	// Retrieve challenge.
	challenge, _, err := w.challengeStore.Retrieve(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("challenge not found: %w", err)
	}

	// Check expiration.
	if time.Now().UTC().After(challenge.ExpiresAt) {
		// Best-effort cleanup of expired challenge.
		if err := w.challengeStore.Delete(ctx, id); err != nil {
			fmt.Printf("warning: failed to delete expired challenge: %v\n", err)
		}

		return nil, fmt.Errorf("webAuthn challenge expired")
	}

	// Reconstruct session from stored challenge.
	sessionDataStr, ok := challenge.Metadata["session_data"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid session data in challenge")
	}

	userHandle, ok := challenge.Metadata["user_handle"].([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid user handle in challenge")
	}

	session := &webauthn.SessionData{
		Challenge:            sessionDataStr,
		UserID:               userHandle,
		AllowedCredentialIDs: [][]byte{},
	}

	// Load existing credentials for this user.
	storedCreds, err := w.credentialStore.GetUserCredentials(ctx, user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get user credentials: %w", err)
	}

	// Convert stored credentials to webauthn.Credential format.
	webauthnCreds := make([]webauthn.Credential, 0, len(storedCreds))

	for _, cred := range storedCreds {
		webauthnCreds = append(webauthnCreds, webauthn.Credential{
			ID:              []byte(cred.ID),
			PublicKey:       cred.PublicKey,
			AttestationType: cred.AttestationType,
			Authenticator: webauthn.Authenticator{
				AAGUID:    cred.AAGUID,
				SignCount: cred.SignCount,
			},
		})
	}

	// Create WebAuthnUser adapter.
	webauthnUser := &WebAuthnUser{
		user:        user,
		credentials: webauthnCreds,
	}

	// Finish registration using go-webauthn library.
	credential, err := w.webauthn.CreateCredential(webauthnUser, *session, credentialCreationResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to create WebAuthn credential: %w", err)
	}

	// Convert to internal Credential format.
	credID := base64.RawURLEncoding.EncodeToString(credential.ID)

	newCred := &Credential{
		ID:              credID,
		UserID:          user.ID.String(),
		Type:            CredentialTypePasskey,
		PublicKey:       credential.PublicKey,
		AttestationType: credential.AttestationType,
		AAGUID:          credential.Authenticator.AAGUID,
		SignCount:       credential.Authenticator.SignCount,
		CreatedAt:       time.Now().UTC(),
		LastUsedAt:      time.Now().UTC(),
		Metadata:        map[string]any{},
	}

	// Store credential.
	if err := w.credentialStore.StoreCredential(ctx, newCred); err != nil {
		return nil, fmt.Errorf("failed to store credential: %w", err)
	}

	// Delete challenge.
	if err := w.challengeStore.Delete(ctx, id); err != nil {
		fmt.Printf("warning: failed to delete challenge: %v\n", err)
	}

	return newCred, nil
}

// InitiateAuth initiates WebAuthn authentication ceremony (implements UserAuthenticator).
func (w *WebAuthnAuthenticator) InitiateAuth(ctx context.Context, userID string) (*AuthChallenge, error) {
	// Load user credentials for assertion options.
	storedCreds, err := w.credentialStore.GetUserCredentials(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user credentials: %w", err)
	}

	if len(storedCreds) == 0 {
		return nil, fmt.Errorf("no WebAuthn credentials registered for user")
	}

	// Convert to webauthn.Credential format.
	webauthnCreds := make([]webauthn.Credential, 0, len(storedCreds))

	for _, cred := range storedCreds {
		webauthnCreds = append(webauthnCreds, webauthn.Credential{
			ID:              []byte(cred.ID),
			PublicKey:       cred.PublicKey,
			AttestationType: cred.AttestationType,
			Authenticator: webauthn.Authenticator{
				AAGUID:    cred.AAGUID,
				SignCount: cred.SignCount,
			},
		})
	}

	// Create WebAuthnUser adapter (user details from stored credentials).
	webauthnUser := &WebAuthnUser{
		user: &cryptoutilIdentityDomain.User{
			ID: googleUuid.MustParse(userID),
		},
		credentials: webauthnCreds,
	}

	// Begin login using go-webauthn library.
	assertion, session, err := w.webauthn.BeginLogin(webauthnUser)
	if err != nil {
		return nil, fmt.Errorf("failed to begin WebAuthn login: %w", err)
	}

	// Store session as challenge.
	challengeID, err := googleUuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate challenge ID: %w", err)
	}

	challenge := &AuthChallenge{
		ID:        challengeID,
		UserID:    userID,
		Method:    "passkey_webauthn",
		ExpiresAt: time.Now().UTC().Add(w.config.Timeout),
		Metadata: map[string]any{
			"session_data":           session.Challenge,
			"operation":              "authentication",
			"user_handle":            session.UserID,
			"allowed_credential_ids": session.AllowedCredentialIDs,
			"assertion":              assertion,
		},
	}

	if err := w.challengeStore.Store(ctx, challenge, challengeID.String()); err != nil {
		return nil, fmt.Errorf("failed to store WebAuthn challenge: %w", err)
	}

	return challenge, nil
}

// VerifyAuth verifies WebAuthn authentication ceremony (implements UserAuthenticator).
func (w *WebAuthnAuthenticator) VerifyAuth(
	ctx context.Context,
	challengeID string,
	credentialAssertionResponse *protocol.ParsedCredentialAssertionData,
) (*cryptoutilIdentityDomain.User, error) {
	if credentialAssertionResponse == nil {
		return nil, fmt.Errorf("credential assertion response cannot be nil")
	}

	// Parse challenge ID.
	id, err := googleUuid.Parse(challengeID)
	if err != nil {
		return nil, fmt.Errorf("invalid challenge ID: %w", err)
	}

	// Retrieve challenge.
	challenge, _, err := w.challengeStore.Retrieve(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("challenge not found: %w", err)
	}

	// Check expiration.
	if time.Now().UTC().After(challenge.ExpiresAt) {
		// Best-effort cleanup of expired challenge.
		if err := w.challengeStore.Delete(ctx, id); err != nil {
			fmt.Printf("warning: failed to delete expired challenge: %v\n", err)
		}

		return nil, fmt.Errorf("webAuthn challenge expired")
	}

	// Reconstruct session from stored challenge.
	sessionDataStr, ok := challenge.Metadata["session_data"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid session data in challenge")
	}

	userHandle, ok := challenge.Metadata["user_handle"].([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid user handle in challenge")
	}

	allowedCredentialIDs, ok := challenge.Metadata["allowed_credential_ids"].([][]byte)
	if !ok {
		return nil, fmt.Errorf("invalid allowed credential IDs in challenge")
	}

	session := &webauthn.SessionData{
		Challenge:            sessionDataStr,
		UserID:               userHandle,
		AllowedCredentialIDs: allowedCredentialIDs,
	}

	// Load user credentials.
	storedCreds, err := w.credentialStore.GetUserCredentials(ctx, challenge.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user credentials: %w", err)
	}

	// Convert to webauthn.Credential format.
	webauthnCreds := make([]webauthn.Credential, 0, len(storedCreds))

	for _, cred := range storedCreds {
		webauthnCreds = append(webauthnCreds, webauthn.Credential{
			ID:              []byte(cred.ID),
			PublicKey:       cred.PublicKey,
			AttestationType: cred.AttestationType,
			Authenticator: webauthn.Authenticator{
				AAGUID:    cred.AAGUID,
				SignCount: cred.SignCount,
			},
		})
	}

	// Create WebAuthnUser adapter.
	webauthnUser := &WebAuthnUser{
		user: &cryptoutilIdentityDomain.User{
			ID: googleUuid.MustParse(challenge.UserID),
		},
		credentials: webauthnCreds,
	}

	// Validate login using go-webauthn library.
	credential, err := w.webauthn.ValidateLogin(webauthnUser, *session, credentialAssertionResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to validate WebAuthn login: %w", err)
	}

	// Update credential sign count (anti-cloning protection).
	credID := base64.RawURLEncoding.EncodeToString(credential.ID)

	storedCred, err := w.credentialStore.GetCredential(ctx, credID)
	if err != nil {
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}

	storedCred.SignCount = credential.Authenticator.SignCount
	storedCred.LastUsedAt = time.Now().UTC()

	if err := w.credentialStore.StoreCredential(ctx, storedCred); err != nil {
		fmt.Printf("warning: failed to update credential sign count: %v\n", err)
	}

	// Delete challenge.
	if err := w.challengeStore.Delete(ctx, id); err != nil {
		fmt.Printf("warning: failed to delete challenge: %v\n", err)
	}

	// Return authenticated user.
	userID, err := googleUuid.Parse(challenge.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	return &cryptoutilIdentityDomain.User{
		ID: userID,
	}, nil
}
