// Copyright (c) 2025-2026 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package apis

import (
	crand "crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	http "net/http"
	"strings"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"

	cryptoutilAppsFrameworkServiceServerRepository "cryptoutil/internal/apps-framework/service/server/repository"
	cryptoutilAppsFrameworkServiceServerService "cryptoutil/internal/apps-framework/service/server/service"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilSharedCryptoHash "cryptoutil/internal/shared/crypto/hash"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	clientAuthErrorInvalidCredentials = "Invalid client credentials"
	clientAuthErrorRegistrationFailed = "Client registration failed"
	clientAuthErrorSessionFailed      = "Service session issuance failed"
)

type clientAuthenticationHandlers struct {
	registrationService cryptoutilAppsFrameworkServiceServerService.RegistrationService
	clientRepository    cryptoutilAppsFrameworkServiceServerRepository.ClientRepository
	tenantRepository    cryptoutilAppsFrameworkServiceServerRepository.TenantRepository
	sessionManager      SessionManager
	realmService        cryptoutilAppsFrameworkServiceServerService.RealmService

	generateCredentialsFn func() (string, string, error)
	hashSecretFn          func(string) (string, error)
	verifySecretFn        func(string, string) (bool, error)
}

// RegisterClientAuthenticationRoutes adds rate-limited headless client registration and
// HTTP Basic client-credential exchange routes to the public service API.
func RegisterClientAuthenticationRoutes(
	app *fiber.App,
	registrationService cryptoutilAppsFrameworkServiceServerService.RegistrationService,
	clientRepository cryptoutilAppsFrameworkServiceServerRepository.ClientRepository,
	tenantRepository cryptoutilAppsFrameworkServiceServerRepository.TenantRepository,
	sessionManager SessionManager,
	realmService cryptoutilAppsFrameworkServiceServerService.RealmService,
	requestsPerMin int,
) {
	handlers := &clientAuthenticationHandlers{
		registrationService:   registrationService,
		clientRepository:      clientRepository,
		tenantRepository:      tenantRepository,
		sessionManager:        sessionManager,
		realmService:          realmService,
		generateCredentialsFn: generateServiceClientCredentials,
		hashSecretFn:          cryptoutilSharedCryptoHash.HashHighEntropyNonDeterministic,
		verifySecretFn:        cryptoutilSharedCryptoHash.VerifySecretHKDFRandom,
	}
	rateLimitMiddleware := newAuthenticationRateLimitMiddleware(requestsPerMin)

	app.Post(
		cryptoutilSharedMagic.DefaultPublicServiceAuthClientRegisterPath,
		rateLimitMiddleware,
		handlers.handleRegisterClient,
	)
	app.Post(
		cryptoutilSharedMagic.DefaultPublicServiceAuthTokenPath,
		rateLimitMiddleware,
		handlers.handleIssueServiceToken,
	)
}

type registerClientRequest struct {
	TenantName string `json:"tenant_name"`
}

type registerClientResponse struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	TenantID     string `json:"tenant_id"`
}

func (h *clientAuthenticationHandlers) handleRegisterClient(c *fiber.Ctx) error {
	var req registerClientRequest
	if err := c.BodyParser(&req); err != nil {
		return clientAuthJSON(c, http.StatusBadRequest, fiber.Map{
			cryptoutilSharedMagic.StringError: "Invalid request body",
		})
	}

	req.TenantName = strings.TrimSpace(req.TenantName)
	if len(req.TenantName) < cryptoutilSharedMagic.DefaultTenantNameMinLength ||
		len(req.TenantName) > cryptoutilSharedMagic.DefaultTenantNameMaxLength {
		return clientAuthJSON(c, http.StatusBadRequest, fiber.Map{
			cryptoutilSharedMagic.StringError: "tenant_name length is invalid",
		})
	}

	clientID, clientSecret, err := h.generateCredentialsFn()
	if err != nil {
		return clientAuthJSON(c, http.StatusInternalServerError, fiber.Map{
			cryptoutilSharedMagic.StringError: clientAuthErrorRegistrationFailed,
		})
	}

	clientSecretHash, err := h.hashSecretFn(clientSecret)
	if err != nil {
		return clientAuthJSON(c, http.StatusInternalServerError, fiber.Map{
			cryptoutilSharedMagic.StringError: clientAuthErrorRegistrationFailed,
		})
	}

	registrationResult, err := h.registrationService.RegisterClient(
		c.Context(),
		clientID,
		clientSecretHash,
		&cryptoutilAppsFrameworkServiceServerService.NewTenantInfo{Name: req.TenantName},
		nil,
	)
	if err != nil || registrationResult == nil {
		return clientAuthJSON(c, http.StatusInternalServerError, fiber.Map{
			cryptoutilSharedMagic.StringError: clientAuthErrorRegistrationFailed,
		})
	}

	return clientAuthJSON(c, http.StatusCreated, registerClientResponse{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TenantID:     registrationResult.TenantID.String(),
	})
}

func (h *clientAuthenticationHandlers) handleIssueServiceToken(c *fiber.Ctx) error {
	clientID, clientSecret, ok := parseBasicClientCredentials(c.Get("Authorization"))
	if !ok {
		return clientAuthUnauthorized(c)
	}

	client, err := h.clientRepository.GetByClientID(c.Context(), clientID)
	if err != nil {
		if isNotFoundAppError(err) {
			return clientAuthUnauthorized(c)
		}

		return clientAuthJSON(c, http.StatusInternalServerError, fiber.Map{
			cryptoutilSharedMagic.StringError: clientAuthErrorSessionFailed,
		})
	}

	if !client.IsActive() {
		return clientAuthUnauthorized(c)
	}

	valid, err := h.verifySecretFn(client.ClientSecretHash, clientSecret)
	if err != nil || !valid {
		return clientAuthUnauthorized(c)
	}

	tenant, err := h.tenantRepository.GetByID(c.Context(), client.TenantID)
	if err != nil {
		if isNotFoundAppError(err) {
			return clientAuthUnauthorized(c)
		}

		return clientAuthJSON(c, http.StatusInternalServerError, fiber.Map{
			cryptoutilSharedMagic.StringError: clientAuthErrorSessionFailed,
		})
	}

	if !tenant.IsActive() {
		return clientAuthUnauthorized(c)
	}

	var realmID googleUuid.UUID

	if h.realmService != nil {
		realm, realmErr := h.realmService.GetFirstActiveRealm(c.Context(), tenant.ID)
		if realmErr == nil && realm != nil {
			realmID = realm.RealmID
		}
	}

	token, err := h.sessionManager.IssueServiceSessionWithTenant(
		c.Context(),
		client.ClientID,
		tenant.ID,
		realmID,
	)
	if err != nil {
		return clientAuthJSON(c, http.StatusInternalServerError, fiber.Map{
			cryptoutilSharedMagic.StringError: clientAuthErrorSessionFailed,
		})
	}

	return clientAuthJSON(c, http.StatusOK, fiber.Map{
		cryptoutilSharedMagic.ParamToken:     token,
		cryptoutilSharedMagic.ParamTokenType: cryptoutilSharedMagic.AuthorizationBearer,
		"expires_at":                         time.Now().UTC().Add(h.sessionManager.ServiceSessionExpiration()).Format(time.RFC3339),
	})
}

func parseBasicClientCredentials(authHeader string) (string, string, bool) {
	scheme, encoded, found := strings.Cut(authHeader, " ")
	if !found || !strings.EqualFold(scheme, "Basic") || strings.TrimSpace(encoded) == "" {
		return "", "", false
	}

	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		return "", "", false
	}

	clientID, clientSecret, found := strings.Cut(string(decoded), ":")
	if !found || clientID == "" || clientSecret == "" {
		return "", "", false
	}

	return clientID, clientSecret, true
}

func clientAuthUnauthorized(c *fiber.Ctx) error {
	return clientAuthJSON(c, http.StatusUnauthorized, fiber.Map{
		cryptoutilSharedMagic.StringError: clientAuthErrorInvalidCredentials,
	})
}

func clientAuthJSON(c *fiber.Ctx, status int, body any) error {
	if err := c.Status(status).JSON(body); err != nil {
		return fmt.Errorf("write client authentication response: %w", err)
	}

	return nil
}

func isNotFoundAppError(err error) bool {
	var appErr *cryptoutilSharedApperr.Error
	if !errors.As(err, &appErr) {
		return false
	}

	return appErr.HTTPStatusLineAndCode.StatusLine.StatusCode == http.StatusNotFound
}

func generateServiceClientCredentials() (string, string, error) {
	return generateServiceClientCredentialsWith(googleUuid.NewV7, crand.Read)
}

func generateServiceClientCredentialsWith(
	generateUUIDFn func() (googleUuid.UUID, error),
	readRandomFn func([]byte) (int, error),
) (string, string, error) {
	clientUUID, err := generateUUIDFn()
	if err != nil {
		return "", "", fmt.Errorf("generate client ID: %w", err)
	}

	secretBytes := make([]byte, cryptoutilSharedMagic.DefaultServiceClientSecretLengthBytes)
	if _, err := readRandomFn(secretBytes); err != nil {
		return "", "", fmt.Errorf("generate client secret: %w", err)
	}

	return clientUUID.String(), base64.RawURLEncoding.EncodeToString(secretBytes), nil
}
