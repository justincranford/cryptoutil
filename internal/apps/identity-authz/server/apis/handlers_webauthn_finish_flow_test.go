// Copyright (c) 2025-2026 Justin Cranford.
//

package apis_test

import (
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	sha256 "crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	json "encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/go-webauthn/webauthn/webauthn"
	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity-authz/server/apis"
	cryptoutilIdentityMfa "cryptoutil/internal/apps/identity/mfa"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	webAuthnRegistrationAttestationFixture  = "a363666d74646e6f6e656761747453746d74a068617574684461746158a4bfabc37432958b063360d3ad6461c9c4735ae7f8edd46592a5e0f01452b2e4b559000000008446ccb9ab1db374750b2367ff6f3a1f0020f91f391db4c9b2fde0ea70189cba3fb63f579ba6122b33ad94ff3ec330084be4a5010203262001215820afefa16f97ca9b2d23eb86ccb64098d20db90856062eb249c33a9b672f26df61225820930a56b87a2fca66334b03458abf879717c12cc68ed73290af2e2664796b9220"
	webAuthnRegistrationClientDataFixture   = "7b2274797065223a22776562617574686e2e637265617465222c226368616c6c656e6765223a22414d4d507434557878475453746e63647134313759447742466938767049612d7077386f4f755657345441222c226f726967696e223a2268747470733a2f2f6578616d706c652e6f7267222c2263726f73734f726967696e223a66616c73652c22657874726144617461223a22636c69656e74446174614a534f4e206d617920626520657874656e6465642077697468206164646974696f6e616c206669656c647320696e20746865206675747572652c207375636820617320746869733a20426b5165446a646354427258426941774a544c453551227d"
	webAuthnRegistrationCredentialIDFixture = "f91f391db4c9b2fde0ea70189cba3fb63f579ba6122b33ad94ff3ec330084be4"
	webAuthnRegistrationChallengeFixture    = "00c30fb78531c464d2b6771dab8d7b603c01162f2fa486bea70f283ae556e130"
	webAuthnRegistrationCredentialType      = "public-key"
	webAuthnAuthenticationCredentialType    = "public-key"
)

func TestFinishWebAuthnRegistration_CredentialLifecycle(t *testing.T) {
	t.Parallel()

	const webAuthnIssuer = "https://example.org"

	userID := googleUuid.Must(googleUuid.NewV7())

	config := *totpTestConfig
	tokens := *totpTestConfig.Tokens
	tokens.Issuer = webAuthnIssuer
	config.Tokens = &tokens

	webAuthnService, err := cryptoutilIdentityMfa.NewWebAuthnService(totpTestRepoFactory.DB(), cryptoutilIdentityMfa.WebAuthnConfig{
		RPDisplayName: webAuthnRPDisplayName,
		RPID:          webAuthnRPID,
		RPOrigins:     []string{webAuthnIssuer},
	})
	require.NoError(t, err)

	_, sessionID, err := webAuthnService.BeginRegistration(t.Context(), &cryptoutilIdentityMfa.WebAuthnUser{
		ID:          userID,
		Name:        webAuthnFinishUsername,
		DisplayName: webAuthnFinishDisplayName,
	})
	require.NoError(t, err)

	var session cryptoutilIdentityMfa.WebAuthnSession
	require.NoError(t, totpTestRepoFactory.DB().Where("id = ?", sessionID).First(&session).Error)

	var sessionData webauthn.SessionData
	require.NoError(t, json.Unmarshal(session.SessionData, &sessionData))

	challenge, err := hex.DecodeString(webAuthnRegistrationChallengeFixture)
	require.NoError(t, err)

	sessionData.Challenge = base64.RawURLEncoding.EncodeToString(challenge)
	session.SessionData, err = json.Marshal(sessionData)
	require.NoError(t, err)
	require.NoError(t, totpTestRepoFactory.DB().Model(&session).Update("session_data", session.SessionData).Error)

	attestationObject, err := hex.DecodeString(webAuthnRegistrationAttestationFixture)
	require.NoError(t, err)
	clientDataJSON, err := hex.DecodeString(webAuthnRegistrationClientDataFixture)
	require.NoError(t, err)
	credentialID, err := hex.DecodeString(webAuthnRegistrationCredentialIDFixture)
	require.NoError(t, err)

	encodedCredentialID := base64.RawURLEncoding.EncodeToString(credentialID)
	responseBody, err := json.Marshal(map[string]any{
		webAuthnCredentialIDJSONField:    encodedCredentialID,
		webAuthnRawCredentialIDJSONField: encodedCredentialID,
		webAuthnCredentialTypeJSONField:  webAuthnRegistrationCredentialType,
		webAuthnCredentialResponseJSONField: map[string]string{
			testWebAuthnAttestationObjectField: base64.RawURLEncoding.EncodeToString(attestationObject),
			webAuthnClientDataJSONField:        base64.RawURLEncoding.EncodeToString(clientDataJSON),
		},
	})
	require.NoError(t, err)

	service := cryptoutilIdentityAuthz.NewService(&config, totpTestRepoFactory, nil)
	app := fiber.New()
	service.RegisterRoutes(app)

	requestBody, err := json.Marshal(cryptoutilIdentityAuthz.FinishWebAuthnRegistrationRequest{
		UserID:      userID.String(),
		Username:    webAuthnFinishUsername,
		DisplayName: webAuthnFinishDisplayName,
		SessionID:   sessionID.String(),
		Response:    base64.StdEncoding.EncodeToString(responseBody),
	})
	require.NoError(t, err)

	request := httptest.NewRequest("POST", webAuthnRegistrationFinishPath, strings.NewReader(string(requestBody)))
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request, -1)

	require.NoError(t, err)
	defer func() { require.NoError(t, response.Body.Close()) }()

	require.Equal(t, fiber.StatusCreated, response.StatusCode)

	var result cryptoutilIdentityAuthz.FinishWebAuthnRegistrationResponse
	require.NoError(t, json.NewDecoder(response.Body).Decode(&result))
	require.True(t, result.Success)
	require.Equal(t, base64.StdEncoding.EncodeToString(credentialID), result.CredentialID)
}

func TestFinishWebAuthnAuthentication_CredentialLifecycle(t *testing.T) {
	t.Parallel()

	const (
		webAuthnIssuer               = "https://example.org"
		authenticatorDataFlagsOffset = 32
		authenticatorDataCounterSize = 4
	)

	userID := googleUuid.Must(googleUuid.NewV7())
	credentialUUID := googleUuid.Must(googleUuid.NewV7())
	credentialID := credentialUUID[:]

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

	config := *totpTestConfig
	tokens := *totpTestConfig.Tokens
	tokens.Issuer = webAuthnIssuer
	config.Tokens = &tokens
	webAuthnConfig := cryptoutilIdentityMfa.WebAuthnConfig{
		RPDisplayName: "Cryptoutil Identity",
		RPID:          "example.org",
		RPOrigins:     []string{webAuthnIssuer},
	}
	credential := cryptoutilIdentityMfa.WebAuthnCredential{
		UserID:          userID,
		CredentialID:    credentialID,
		PublicKey:       publicKey,
		AttestationType: cryptoutilSharedMagic.WebAuthnAttestationTypeNone,
		Transports:      testEmptyJSONArray,
		DisplayName:     "WebAuthn test credential",
	}
	require.NoError(t, totpTestRepoFactory.DB().Create(&credential).Error)

	const username = "webauthn-authentication-user"

	service := cryptoutilIdentityAuthz.NewService(&config, totpTestRepoFactory, nil)
	app := fiber.New()
	service.RegisterRoutes(app)

	beginRequestBody, err := json.Marshal(cryptoutilIdentityAuthz.BeginWebAuthnAuthenticationRequest{
		UserID:   userID.String(),
		Username: username,
	})
	require.NoError(t, err)

	beginRequest := httptest.NewRequest("POST", webAuthnAuthenticationBeginPath, strings.NewReader(string(beginRequestBody)))
	beginRequest.Header.Set("Content-Type", "application/json")
	beginResponse, err := app.Test(beginRequest, -1)

	require.NoError(t, err)
	defer func() { require.NoError(t, beginResponse.Body.Close()) }()

	require.Equal(t, fiber.StatusOK, beginResponse.StatusCode)

	var beginResult cryptoutilIdentityAuthz.BeginWebAuthnAuthenticationResponse
	require.NoError(t, json.NewDecoder(beginResponse.Body).Decode(&beginResult))
	sessionID, err := googleUuid.Parse(beginResult.SessionID)
	require.NoError(t, err)

	var storedSession cryptoutilIdentityMfa.WebAuthnSession
	require.NoError(t, totpTestRepoFactory.DB().Where("id = ?", sessionID).First(&storedSession).Error)

	var sessionData webauthn.SessionData
	require.NoError(t, json.Unmarshal(storedSession.SessionData, &sessionData))

	rpIDHash := sha256.Sum256([]byte(webAuthnConfig.RPID))
	authenticatorData := make([]byte, sha256.Size+1+authenticatorDataCounterSize)
	copy(authenticatorData, rpIDHash[:])
	authenticatorData[authenticatorDataFlagsOffset] = 1
	binary.BigEndian.PutUint32(authenticatorData[authenticatorDataFlagsOffset+1:], 1)

	clientDataJSON, err := json.Marshal(map[string]any{
		webAuthnCredentialTypeJSONField: "webauthn.get",
		"challenge":                     sessionData.Challenge,
		"origin":                        webAuthnConfig.RPOrigins[0],
		"crossOrigin":                   false,
	})
	require.NoError(t, err)

	clientDataHash := sha256.Sum256(clientDataJSON)
	signedData := make([]byte, len(authenticatorData)+len(clientDataHash))
	copy(signedData, authenticatorData)
	copy(signedData[len(authenticatorData):], clientDataHash[:])
	signedHash := sha256.Sum256(signedData)
	signature, err := ecdsa.SignASN1(crand.Reader, privateKey, signedHash[:])
	require.NoError(t, err)

	encodedCredentialID := base64.RawURLEncoding.EncodeToString(credentialID)
	responseBody, err := json.Marshal(map[string]any{
		webAuthnCredentialIDJSONField:    encodedCredentialID,
		webAuthnRawCredentialIDJSONField: encodedCredentialID,
		webAuthnCredentialTypeJSONField:  webAuthnAuthenticationCredentialType,
		webAuthnCredentialResponseJSONField: map[string]string{
			testWebAuthnAuthenticatorDataField: base64.RawURLEncoding.EncodeToString(authenticatorData),
			webAuthnClientDataJSONField:        base64.RawURLEncoding.EncodeToString(clientDataJSON),
			testWebAuthnSignatureField:         base64.RawURLEncoding.EncodeToString(signature),
			testWebAuthnUserHandleField:        base64.RawURLEncoding.EncodeToString(userID[:]),
		},
	})
	require.NoError(t, err)

	requestBody, err := json.Marshal(cryptoutilIdentityAuthz.FinishWebAuthnAuthenticationRequest{
		UserID:    userID.String(),
		Username:  username,
		SessionID: sessionID.String(),
		Response:  base64.StdEncoding.EncodeToString(responseBody),
	})
	require.NoError(t, err)

	request := httptest.NewRequest("POST", webAuthnAuthenticationFinishPath, strings.NewReader(string(requestBody)))
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request, -1)

	require.NoError(t, err)
	defer func() { require.NoError(t, response.Body.Close()) }()

	require.Equal(t, fiber.StatusOK, response.StatusCode)

	var result cryptoutilIdentityAuthz.FinishWebAuthnAuthenticationResponse
	require.NoError(t, json.NewDecoder(response.Body).Decode(&result))
	require.True(t, result.Success)
	require.Equal(t, base64.StdEncoding.EncodeToString(credentialID), result.CredentialID)
	require.NotEmpty(t, result.LastUsedAt)
}
