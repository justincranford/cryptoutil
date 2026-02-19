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
)

// CredentialType represents the type of WebAuthn credential.
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
