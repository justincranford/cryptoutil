// Copyright (c) 2025-2026 Justin Cranford.
//

package mfa_test

import (
	"bytes"
	"context"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	sha256 "crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	json "encoding/json"
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityMfa "cryptoutil/internal/apps/identity/mfa"
)

const (
	testWebAuthnCredentialIDJSONField    = "id"
	testWebAuthnRawCredentialIDJSONField = "rawId"
	testWebAuthnPublicKeyFixture         = "public-key"
	testWebAuthnResponseJSONField        = "response"
	testWebAuthnClientDataJSONField      = "clientDataJSON"
	testWebAuthnRPID                     = "example.org"
	testWebAuthnRPOrigin                 = "https://example.org"
	testWebAuthnMissingSessionCaseName   = "missing session"
	testWebAuthnExpiredSessionCaseName   = "expired session"
	testWebAuthnSessionUnmarshalError    = "failed to unmarshal session"
	testWebAuthnCredentialTypeJSONField  = "type"
	testWebAuthnSessionNotFoundError     = "session not found"
	testWebAuthnSessionExpiredError      = "session expired"
	testWebAuthnInvalidSessionDataCase   = "invalid session data"
)

func parseWebAuthnRegistrationFixture(t *testing.T) (*protocol.ParsedCredentialCreationData, string, []byte) {
	t.Helper()

	const (
		attestationObjectHex = "a363666d74646e6f6e656761747453746d74a068617574684461746158a4bfabc37432958b063360d3ad6461c9c4735ae7f8edd46592a5e0f01452b2e4b559000000008446ccb9ab1db374750b2367ff6f3a1f0020f91f391db4c9b2fde0ea70189cba3fb63f579ba6122b33ad94ff3ec330084be4a5010203262001215820afefa16f97ca9b2d23eb86ccb64098d20db90856062eb249c33a9b672f26df61225820930a56b87a2fca66334b03458abf879717c12cc68ed73290af2e2664796b9220"
		clientDataJSONHex    = "7b2274797065223a22776562617574686e2e637265617465222c226368616c6c656e6765223a22414d4d507434557878475453746e63647134313759447742466938767049612d7077386f4f755657345441222c226f726967696e223a2268747470733a2f2f6578616d706c652e6f7267222c2263726f73734f726967696e223a66616c73652c22657874726144617461223a22636c69656e74446174614a534f4e206d617920626520657874656e6465642077697468206164646974696f6e616c206669656c647320696e20746865206675747572652c207375636820617320746869733a20426b5165446a646354427258426941774a544c453551227d"
		credentialIDHex      = "f91f391db4c9b2fde0ea70189cba3fb63f579ba6122b33ad94ff3ec330084be4"
		challengeHex         = "00c30fb78531c464d2b6771dab8d7b603c01162f2fa486bea70f283ae556e130"
	)

	attestationObject, err := hex.DecodeString(attestationObjectHex)
	require.NoError(t, err)
	clientDataJSON, err := hex.DecodeString(clientDataJSONHex)
	require.NoError(t, err)
	credentialID, err := hex.DecodeString(credentialIDHex)
	require.NoError(t, err)
	challengeBytes, err := hex.DecodeString(challengeHex)
	require.NoError(t, err)

	credentialIDEncoded := base64.RawURLEncoding.EncodeToString(credentialID)
	body, err := json.Marshal(map[string]any{
		testWebAuthnCredentialIDJSONField:    credentialIDEncoded,
		testWebAuthnRawCredentialIDJSONField: credentialIDEncoded,
		testWebAuthnCredentialTypeJSONField:  testWebAuthnPublicKeyFixture,
		testWebAuthnResponseJSONField: map[string]string{
			"attestationObject":             base64.RawURLEncoding.EncodeToString(attestationObject),
			testWebAuthnClientDataJSONField: base64.RawURLEncoding.EncodeToString(clientDataJSON),
		},
	})
	require.NoError(t, err)

	response, err := protocol.ParseCredentialCreationResponseBody(bytes.NewReader(body))
	require.NoError(t, err)

	return response, base64.RawURLEncoding.EncodeToString(challengeBytes), credentialID
}

func TestWebAuthnService_FinishRegistrationSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := createWebAuthnTestDB(t)
	config := cryptoutilIdentityMfa.WebAuthnConfig{
		RPDisplayName: testWebAuthnRPDisplayName,
		RPID:          testWebAuthnRPID,
		RPOrigins:     []string{testWebAuthnRPOrigin},
	}
	service, err := cryptoutilIdentityMfa.NewWebAuthnService(db, config)
	require.NoError(t, err)

	user := &cryptoutilIdentityMfa.WebAuthnUser{
		ID:          googleUuid.Must(googleUuid.NewV7()),
		Name:        testTestuser,
		DisplayName: testTestUser,
	}
	_, sessionID, err := service.BeginRegistration(ctx, user)
	require.NoError(t, err)

	response, challenge, credentialID := parseWebAuthnRegistrationFixture(t)
	sessionData, err := json.Marshal(webauthn.SessionData{
		Challenge:      challenge,
		RelyingPartyID: config.RPID,
		UserID:         user.WebAuthnID(),
		CredParams:     []protocol.CredentialParameter{{Type: protocol.PublicKeyCredentialType, Algorithm: -7}},
	})
	require.NoError(t, err)
	require.NoError(t, db.Model(&cryptoutilIdentityMfa.WebAuthnSession{}).Where("id = ?", sessionID).Update("session_data", sessionData).Error)

	credential, err := service.FinishRegistration(ctx, user, sessionID, response, "test key")
	require.NoError(t, err)
	require.Equal(t, credentialID, credential.CredentialID)
	require.Equal(t, "test key", credential.DisplayName)

	var storedCredential cryptoutilIdentityMfa.WebAuthnCredential
	require.NoError(t, db.Where("id = ?", credential.ID).First(&storedCredential).Error)
	require.Equal(t, credentialID, storedCredential.CredentialID)

	var sessionCount int64
	require.NoError(t, db.Model(&cryptoutilIdentityMfa.WebAuthnSession{}).Where("id = ?", sessionID).Count(&sessionCount).Error)
	require.Zero(t, sessionCount)
}

func TestWebAuthnService_FinishRegistrationCredentialStorageError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := createWebAuthnTestDB(t)
	config := cryptoutilIdentityMfa.WebAuthnConfig{
		RPDisplayName: testWebAuthnRPDisplayName,
		RPID:          testWebAuthnRPID,
		RPOrigins:     []string{testWebAuthnRPOrigin},
	}
	service, err := cryptoutilIdentityMfa.NewWebAuthnService(db, config)
	require.NoError(t, err)

	user := &cryptoutilIdentityMfa.WebAuthnUser{
		ID:          googleUuid.Must(googleUuid.NewV7()),
		Name:        testTestuser,
		DisplayName: testTestUser,
	}
	_, sessionID, err := service.BeginRegistration(ctx, user)
	require.NoError(t, err)

	response, challenge, _ := parseWebAuthnRegistrationFixture(t)
	sessionData, err := json.Marshal(webauthn.SessionData{
		Challenge:      challenge,
		RelyingPartyID: config.RPID,
		UserID:         user.WebAuthnID(),
		CredParams:     []protocol.CredentialParameter{{Type: protocol.PublicKeyCredentialType, Algorithm: -7}},
	})
	require.NoError(t, err)
	require.NoError(t, db.Model(&cryptoutilIdentityMfa.WebAuthnSession{}).Where("id = ?", sessionID).Update("session_data", sessionData).Error)
	require.NoError(t, db.Exec(`CREATE TRIGGER fail_webauthn_credential_insert BEFORE INSERT ON webauthn_credentials BEGIN SELECT RAISE(ABORT, 'credential storage rejected'); END`).Error)

	credential, err := service.FinishRegistration(ctx, user, sessionID, response, "test key")
	require.ErrorContains(t, err, "failed to save credential")
	require.Nil(t, credential)
}

func TestWebAuthnService_FinishAuthenticationSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := createWebAuthnTestDB(t)
	config := cryptoutilIdentityMfa.WebAuthnConfig{
		RPDisplayName: testWebAuthnRPDisplayName,
		RPID:          testWebAuthnRPID,
		RPOrigins:     []string{testWebAuthnRPOrigin},
	}
	service, err := cryptoutilIdentityMfa.NewWebAuthnService(db, config)
	require.NoError(t, err)

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)
	ecdhPrivateKey, err := privateKey.ECDH()
	require.NoError(t, err)

	publicKeyBytes := ecdhPrivateKey.PublicKey().Bytes()
	publicKey, err := cbor.Marshal(map[int]any{
		1:  2,
		3:  -7,
		-1: 1,
		-2: publicKeyBytes[1 : 1+sha256.Size],
		-3: publicKeyBytes[1+sha256.Size:],
	})
	require.NoError(t, err)

	credentialUUID := googleUuid.Must(googleUuid.NewV7())
	credentialID := credentialUUID[:]
	userID := googleUuid.Must(googleUuid.NewV7())
	user := &cryptoutilIdentityMfa.WebAuthnUser{
		ID:          userID,
		Name:        testTestuser,
		DisplayName: testTestUser,
		Credentials: []cryptoutilIdentityMfa.WebAuthnCredential{{
			UserID:       userID,
			CredentialID: credentialID,
			PublicKey:    publicKey,
			Transports:   "[]",
		}},
	}
	require.NoError(t, db.Create(&user.Credentials[0]).Error)

	_, sessionID, err := service.BeginAuthentication(ctx, user)
	require.NoError(t, err)

	var storedSession cryptoutilIdentityMfa.WebAuthnSession
	require.NoError(t, db.Where("id = ?", sessionID).First(&storedSession).Error)

	var sessionData webauthn.SessionData
	require.NoError(t, json.Unmarshal(storedSession.SessionData, &sessionData))

	rpIDHash := sha256.Sum256([]byte(config.RPID))

	const authenticatorDataMetadataLength = 1 + 4

	authenticatorDataLength := sha256.Size + authenticatorDataMetadataLength
	authenticatorData := make([]byte, authenticatorDataLength, authenticatorDataLength+sha256.Size)
	copy(authenticatorData, rpIDHash[:])
	authenticatorData[sha256.Size] = 1
	binary.BigEndian.PutUint32(authenticatorData[sha256.Size+1:], 1)

	clientDataJSON, err := json.Marshal(map[string]any{
		testWebAuthnCredentialTypeJSONField: "webauthn.get",
		"challenge":                         sessionData.Challenge,
		"origin":                            config.RPOrigins[0],
		"crossOrigin":                       false,
	})
	require.NoError(t, err)

	clientDataHash := sha256.Sum256(clientDataJSON)
	signedData := make([]byte, 0, len(authenticatorData)+len(clientDataHash))
	signedData = append(signedData, authenticatorData...)
	signedData = append(signedData, clientDataHash[:]...)
	signedHash := sha256.Sum256(signedData)
	signature, err := ecdsa.SignASN1(crand.Reader, privateKey, signedHash[:])
	require.NoError(t, err)

	credentialIDEncoded := base64.RawURLEncoding.EncodeToString(credentialID)
	body, err := json.Marshal(map[string]any{
		testWebAuthnCredentialIDJSONField:    credentialIDEncoded,
		testWebAuthnRawCredentialIDJSONField: credentialIDEncoded,
		testWebAuthnCredentialTypeJSONField:  testWebAuthnPublicKeyFixture,
		testWebAuthnResponseJSONField: map[string]string{
			"authenticatorData":             base64.RawURLEncoding.EncodeToString(authenticatorData),
			testWebAuthnClientDataJSONField: base64.RawURLEncoding.EncodeToString(clientDataJSON),
			"signature":                     base64.RawURLEncoding.EncodeToString(signature),
			"userHandle":                    base64.RawURLEncoding.EncodeToString(user.WebAuthnID()),
		},
	})
	require.NoError(t, err)
	response, err := protocol.ParseCredentialRequestResponseBytes(body)
	require.NoError(t, err)

	updatedCredential, err := service.FinishAuthentication(ctx, user, sessionID, response)
	require.NoError(t, err)
	require.Equal(t, credentialID, updatedCredential.CredentialID)
	require.NotNil(t, updatedCredential.LastUsedAt)

	var sessionCount int64
	require.NoError(t, db.Model(&cryptoutilIdentityMfa.WebAuthnSession{}).Where("id = ?", sessionID).Count(&sessionCount).Error)
	require.Zero(t, sessionCount)
}

func TestWebAuthnService_FinishRegistrationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		sessionData       []byte
		expiresAt         time.Time
		wantErr           string
		wantIdentityError error
	}{
		{name: testWebAuthnMissingSessionCaseName, wantErr: testWebAuthnSessionNotFoundError, wantIdentityError: cryptoutilIdentityAppErr.ErrWebAuthnSessionNotFound},
		{name: testWebAuthnExpiredSessionCaseName, sessionData: []byte("{}"), expiresAt: time.Now().UTC().Add(-time.Minute), wantErr: testWebAuthnSessionExpiredError, wantIdentityError: cryptoutilIdentityAppErr.ErrWebAuthnSessionExpired},
		{name: testWebAuthnInvalidSessionDataCase, sessionData: []byte("{"), expiresAt: time.Now().UTC().Add(time.Minute), wantErr: testWebAuthnSessionUnmarshalError},
		{name: "invalid registration response", sessionData: []byte("{}"), expiresAt: time.Now().UTC().Add(time.Minute), wantErr: "failed to create credential", wantIdentityError: cryptoutilIdentityAppErr.ErrWebAuthnVerificationFailed},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := createWebAuthnTestDB(t)
			service, err := cryptoutilIdentityMfa.NewWebAuthnService(db, createWebAuthnConfig())
			require.NoError(t, err)

			user := &cryptoutilIdentityMfa.WebAuthnUser{ID: googleUuid.Must(googleUuid.NewV7()), Name: testTestuser, DisplayName: testTestUser}

			sessionID := googleUuid.Must(googleUuid.NewV7())
			if tc.name != testWebAuthnMissingSessionCaseName {
				session := &cryptoutilIdentityMfa.WebAuthnSession{
					ID:           sessionID,
					UserID:       user.ID,
					SessionData:  tc.sessionData,
					CeremonyType: string(cryptoutilIdentityMfa.WebAuthnCeremonyRegistration),
					ExpiresAt:    tc.expiresAt,
				}
				require.NoError(t, db.Create(session).Error)
			}

			credential, err := service.FinishRegistration(ctx, user, sessionID, &protocol.ParsedCredentialCreationData{}, "test key")
			require.ErrorContains(t, err, tc.wantErr)

			if tc.wantIdentityError != nil {
				require.ErrorIs(t, err, tc.wantIdentityError)
			}

			require.Nil(t, credential)
		})
	}
}

func TestWebAuthnService_FinishAuthenticationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		sessionData       []byte
		expiresAt         time.Time
		wantErr           string
		wantIdentityError error
	}{
		{name: testWebAuthnMissingSessionCaseName, wantErr: testWebAuthnSessionNotFoundError, wantIdentityError: cryptoutilIdentityAppErr.ErrWebAuthnSessionNotFound},
		{name: testWebAuthnExpiredSessionCaseName, sessionData: []byte("{}"), expiresAt: time.Now().UTC().Add(-time.Minute), wantErr: testWebAuthnSessionExpiredError, wantIdentityError: cryptoutilIdentityAppErr.ErrWebAuthnSessionExpired},
		{name: testWebAuthnInvalidSessionDataCase, sessionData: []byte("{"), expiresAt: time.Now().UTC().Add(time.Minute), wantErr: testWebAuthnSessionUnmarshalError},
		{name: "invalid assertion response", sessionData: []byte("{}"), expiresAt: time.Now().UTC().Add(time.Minute), wantErr: "failed to validate login", wantIdentityError: cryptoutilIdentityAppErr.ErrWebAuthnVerificationFailed},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := createWebAuthnTestDB(t)
			service, err := cryptoutilIdentityMfa.NewWebAuthnService(db, createWebAuthnConfig())
			require.NoError(t, err)

			user := &cryptoutilIdentityMfa.WebAuthnUser{ID: googleUuid.Must(googleUuid.NewV7()), Name: testTestuser, DisplayName: testTestUser}

			sessionID := googleUuid.Must(googleUuid.NewV7())
			if tc.name != testWebAuthnMissingSessionCaseName {
				session := &cryptoutilIdentityMfa.WebAuthnSession{
					ID:           sessionID,
					UserID:       user.ID,
					SessionData:  tc.sessionData,
					CeremonyType: string(cryptoutilIdentityMfa.WebAuthnCeremonyAuthentication),
					ExpiresAt:    tc.expiresAt,
				}
				require.NoError(t, db.Create(session).Error)
			}

			credential, err := service.FinishAuthentication(ctx, user, sessionID, &protocol.ParsedCredentialAssertionData{})
			require.ErrorContains(t, err, tc.wantErr)

			if tc.wantIdentityError != nil {
				require.ErrorIs(t, err, tc.wantIdentityError)
			}

			require.Nil(t, credential)
		})
	}
}

func TestWebAuthnService_BeginAuthentication(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	user := &cryptoutilIdentityMfa.WebAuthnUser{
		ID:          googleUuid.Must(googleUuid.NewV7()),
		Name:        testTestuser,
		DisplayName: testTestUser,
		Credentials: []cryptoutilIdentityMfa.WebAuthnCredential{
			{CredentialID: []byte("credential-id"), PublicKey: []byte(testWebAuthnPublicKeyFixture)},
		},
	}

	db := createWebAuthnTestDB(t)
	service, err := cryptoutilIdentityMfa.NewWebAuthnService(db, createWebAuthnConfig())
	require.NoError(t, err)

	options, sessionID, err := service.BeginAuthentication(ctx, user)
	require.NoError(t, err)
	require.NotNil(t, options)
	require.NotEqual(t, googleUuid.Nil, sessionID)

	closedDB, err := createClosedWebAuthnDB(t)
	require.NoError(t, err)
	closedService, err := cryptoutilIdentityMfa.NewWebAuthnService(closedDB, createWebAuthnConfig())
	require.NoError(t, err)
	options, sessionID, err = closedService.BeginAuthentication(ctx, user)
	require.ErrorContains(t, err, "failed to save session")
	require.Nil(t, options)
	require.Equal(t, googleUuid.Nil, sessionID)
}

func TestWebAuthnService_DatabaseErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	closedDB, err := createClosedWebAuthnDB(t)
	require.NoError(t, err)
	service, err := cryptoutilIdentityMfa.NewWebAuthnService(closedDB, createWebAuthnConfig())
	require.NoError(t, err)

	user := &cryptoutilIdentityMfa.WebAuthnUser{ID: googleUuid.Must(googleUuid.NewV7()), Name: testTestuser, DisplayName: testTestUser}
	options, sessionID, err := service.BeginRegistration(ctx, user)
	require.ErrorContains(t, err, "failed to save session")
	require.Nil(t, options)
	require.Equal(t, googleUuid.Nil, sessionID)

	credential, err := service.FinishRegistration(ctx, user, googleUuid.Must(googleUuid.NewV7()), &protocol.ParsedCredentialCreationData{}, "key")
	require.ErrorContains(t, err, "failed to get registration session")
	require.Nil(t, credential)

	credential, err = service.FinishAuthentication(ctx, user, googleUuid.Must(googleUuid.NewV7()), &protocol.ParsedCredentialAssertionData{})
	require.ErrorContains(t, err, "failed to get authentication session")
	require.Nil(t, credential)

	require.ErrorContains(t, service.DeleteCredential(ctx, user.ID, googleUuid.Must(googleUuid.NewV7())), "failed to delete credential")
}

func TestWebAuthnUser_CredentialConversion(t *testing.T) {
	t.Parallel()

	user := &cryptoutilIdentityMfa.WebAuthnUser{
		Credentials: []cryptoutilIdentityMfa.WebAuthnCredential{
			{
				CredentialID:    []byte("credential-one"),
				PublicKey:       []byte(testWebAuthnPublicKeyFixture),
				AttestationType: cryptoutilSharedMagic.ClientAuthMethodNone,
				Transports:      `["usb","internal"]`,
				AAGUID:          []byte("authenticator"),
				SignCount:       7,
				CloneWarning:    true,
			},
			{CredentialID: []byte("credential-two"), Transports: "invalid JSON"},
		},
	}

	credentials := user.WebAuthnCredentials()
	require.Len(t, credentials, 2)
	require.Equal(t, []byte("credential-one"), credentials[0].ID)
	require.Equal(t, []byte(testWebAuthnPublicKeyFixture), credentials[0].PublicKey)
	require.Equal(t, []protocol.AuthenticatorTransport{"usb", "internal"}, credentials[0].Transport)
	require.Equal(t, []byte("credential-two"), credentials[1].ID)
	require.Empty(t, credentials[1].Transport)
}

func TestWebAuthnService_CleanupAndDeleteErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := createWebAuthnTestDB(t)
	service, err := cryptoutilIdentityMfa.NewWebAuthnService(db, createWebAuthnConfig())
	require.NoError(t, err)

	userID := googleUuid.Must(googleUuid.NewV7())
	expiredSession := &cryptoutilIdentityMfa.WebAuthnSession{
		UserID:       userID,
		SessionData:  []byte("{}"),
		CeremonyType: string(cryptoutilIdentityMfa.WebAuthnCeremonyRegistration),
		ExpiresAt:    time.Now().UTC().Add(-time.Minute),
	}
	currentSession := &cryptoutilIdentityMfa.WebAuthnSession{
		UserID:       userID,
		SessionData:  []byte("{}"),
		CeremonyType: string(cryptoutilIdentityMfa.WebAuthnCeremonyAuthentication),
		ExpiresAt:    time.Now().UTC().Add(time.Hour),
	}

	require.NoError(t, db.Create(expiredSession).Error)
	require.NoError(t, db.Create(currentSession).Error)
	require.NoError(t, service.CleanupExpiredSessions(ctx))

	var sessions []cryptoutilIdentityMfa.WebAuthnSession
	require.NoError(t, db.Find(&sessions).Error)
	require.Len(t, sessions, 1)
	require.Equal(t, currentSession.ID, sessions[0].ID)

	credential := &cryptoutilIdentityMfa.WebAuthnCredential{
		UserID:          userID,
		CredentialID:    []byte("credential-for-delete"),
		PublicKey:       []byte(testWebAuthnPublicKeyFixture),
		AttestationType: cryptoutilSharedMagic.PromptNone,
		DisplayName:     "delete me",
	}
	require.NoError(t, db.Create(credential).Error)
	require.NoError(t, service.DeleteCredential(ctx, userID, credential.ID))

	var deleted int64
	require.NoError(t, db.Model(&cryptoutilIdentityMfa.WebAuthnCredential{}).Where("id = ?", credential.ID).Count(&deleted).Error)
	require.Zero(t, deleted)

	invalidDB, err := createClosedWebAuthnDB(t)
	require.NoError(t, err)
	invalidService, err := cryptoutilIdentityMfa.NewWebAuthnService(invalidDB, createWebAuthnConfig())
	require.NoError(t, err)
	require.ErrorContains(t, invalidService.CleanupExpiredSessions(ctx), "failed to cleanup sessions")
	credentials, err := invalidService.GetCredentials(ctx, userID)
	require.ErrorContains(t, err, "failed to get credentials")
	require.Nil(t, credentials)
}

func createClosedWebAuthnDB(t *testing.T) (*gorm.DB, error) {
	t.Helper()

	db := createWebAuthnTestDB(t)

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if err := sqlDB.Close(); err != nil {
		return nil, err
	}

	return db, nil
}

func TestWebAuthnCredential_BeforeUpdate(t *testing.T) {
	t.Parallel()

	db := createWebAuthnTestDB(t)
	credential := &cryptoutilIdentityMfa.WebAuthnCredential{
		UserID:          googleUuid.Must(googleUuid.NewV7()),
		CredentialID:    []byte("credential-for-update"),
		PublicKey:       []byte(testWebAuthnPublicKeyFixture),
		AttestationType: cryptoutilSharedMagic.PromptNone,
		DisplayName:     "before update",
	}
	require.NoError(t, db.Create(credential).Error)

	originalUpdatedAt := time.Now().UTC().Add(-time.Hour)
	require.NoError(t, db.Model(credential).UpdateColumn(cryptoutilSharedMagic.ClaimUpdatedAt, originalUpdatedAt).Error)

	var persisted cryptoutilIdentityMfa.WebAuthnCredential
	require.NoError(t, db.First(&persisted, "id = ?", credential.ID).Error)
	updatedAt := persisted.UpdatedAt

	require.NoError(t, db.Model(credential).Update("display_name", "after update").Error)

	require.NoError(t, db.First(&persisted, "id = ?", credential.ID).Error)
	require.Equal(t, "after update", persisted.DisplayName)
	require.True(t, persisted.UpdatedAt.After(updatedAt))
}
