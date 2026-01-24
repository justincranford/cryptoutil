// Copyright (c) 2025 Justin Cranford
//

// Package server provides the JOSE Authority Server HTTP service.
package server

import (
	"context"
	json "encoding/json"
	"strings"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"

	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	cryptoutilJoseService "cryptoutil/internal/jose/service"
)

// Default tenant and realm IDs for JWKS endpoint.
// These match the jose-ja server defaults for now.
// TODO: Extract from authentication context in the future.
var (
	defaultJWKSTenantID = googleUuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479")
	defaultJWKSRealmID  = googleUuid.MustParse("7c9e6679-7425-40de-944b-e07fc1f90ae7")
)

// joseHandlerAdapter provides JOSE-specific route handlers using the existing KeyStore.
// This adapter wraps the handler logic to work with both the legacy Server and new JoseServer.
type joseHandlerAdapter struct {
	telemetryService  *cryptoutilSharedTelemetry.TelemetryService
	jwkGenService     *cryptoutilSharedCryptoJose.JWKGenService
	elasticJWKService *cryptoutilJoseService.ElasticJWKService
	keyStore          *KeyStore
}

// ============================================================================
// JWK Handlers
// ============================================================================

func (h *joseHandlerAdapter) handleJWKGenerate(c *fiber.Ctx) error {
	var req JWKGenerateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	alg := cryptoutilOpenapiModel.GenerateAlgorithm(req.Algorithm)
	if !isValidAlgorithm(alg) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid algorithm",
			"allowed": []string{"RSA/4096", "RSA/3072", "RSA/2048", "EC/P521", "EC/P384", "EC/P256", "OKP/Ed25519", "oct/512", "oct/384", "oct/256"},
		})
	}

	var (
		kid        *googleUuid.UUID
		privateJWK joseJwk.Key
		publicJWK  joseJwk.Key
		err        error
	)

	if req.Use == cryptoutilSharedMagic.JoseKeyUseEnc {
		enc, keyAlg := mapToEncryptionAlgorithms(alg)

		kid, privateJWK, publicJWK, _, _, err = h.jwkGenService.GenerateJWEJWK(&enc, &keyAlg)
		if err != nil {
			h.telemetryService.Slogger.Error("Failed to generate JWE JWK", "error", err)

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to generate encryption key",
			})
		}
	} else {
		sigAlg := mapToSignatureAlgorithm(alg)

		kid, privateJWK, publicJWK, _, _, err = h.jwkGenService.GenerateJWSJWK(sigAlg)
		if err != nil {
			h.telemetryService.Slogger.Error("Failed to generate JWS JWK", "error", err)

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to generate signing key",
			})
		}
	}

	// Store key using new StoredKey structure.
	alg2, _ := privateJWK.Algorithm()
	kty := privateJWK.KeyType()
	storedKey := &StoredKey{
		KID:        *kid,
		PrivateJWK: privateJWK,
		PublicJWK:  publicJWK,
		KeyType:    kty.String(),
		Algorithm:  alg2.String(),
		Use:        req.Use,
		CreatedAt:  time.Now().Unix(),
	}

	if err := h.keyStore.Store(storedKey); err != nil {
		h.telemetryService.Slogger.Error("Failed to store JWK", "error", err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to store key",
		})
	}

	var publicJWKJSON []byte
	if publicJWK != nil {
		publicJWKJSON, _ = json.Marshal(publicJWK)
	}

	return c.Status(fiber.StatusCreated).JSON(JWKGenerateResponse{
		KID:       kid.String(),
		Algorithm: req.Algorithm,
		Use:       req.Use,
		KeyType:   kty.String(),
		PublicJWK: publicJWKJSON,
		CreatedAt: time.Now().Unix(),
	})
}

func (h *joseHandlerAdapter) handleJWKGet(c *fiber.Ctx) error {
	kid := c.Params("kid")
	if kid == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Key ID is required",
		})
	}

	storedKey, exists := h.keyStore.Get(kid)
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Key not found",
		})
	}

	// Return public key only.
	var jwkJSON []byte
	if storedKey.PublicJWK != nil {
		jwkJSON, _ = json.Marshal(storedKey.PublicJWK)
	} else if storedKey.PrivateJWK != nil {
		publicKey, err := storedKey.PrivateJWK.PublicKey()
		if err != nil {
			jwkJSON, _ = json.Marshal(storedKey.PrivateJWK)
		} else {
			jwkJSON, _ = json.Marshal(publicKey)
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"kid": kid,
		"jwk": json.RawMessage(jwkJSON),
	})
}

func (h *joseHandlerAdapter) handleJWKDelete(c *fiber.Ctx) error {
	kid := c.Params("kid")
	if kid == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Key ID is required",
		})
	}

	if !h.keyStore.Delete(kid) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Key not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Key deleted",
		"kid":     kid,
	})
}

func (h *joseHandlerAdapter) handleJWKList(c *fiber.Ctx) error {
	keys := h.keyStore.List()
	result := make([]fiber.Map, 0, len(keys))

	for _, storedKey := range keys {
		result = append(result, fiber.Map{
			"kid": storedKey.KID.String(),
			"kty": storedKey.KeyType,
			"alg": storedKey.Algorithm,
			"use": storedKey.Use,
		})
	}

	return c.JSON(fiber.Map{
		"keys":  result,
		"count": len(result),
	})
}

func (h *joseHandlerAdapter) handleJWKS(c *fiber.Ctx) error {
	keys := h.keyStore.List()
	jwks := joseJwk.NewSet()

	for _, storedKey := range keys {
		var publicKey joseJwk.Key

		if storedKey.PublicJWK != nil {
			publicKey = storedKey.PublicJWK
		} else if storedKey.PrivateJWK != nil {
			pk, err := storedKey.PrivateJWK.PublicKey()
			if err != nil {
				continue
			}

			publicKey = pk
		}

		if publicKey != nil {
			if err := jwks.AddKey(publicKey); err != nil {
				h.telemetryService.Slogger.Error("Failed to add key to JWKS", "error", err)
			}
		}
	}

	jwksJSON, err := json.Marshal(jwks)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to serialize JWKS",
		})
	}

	c.Set("Content-Type", "application/json")

	return c.Send(jwksJSON)
}

// ============================================================================
// JWS Handlers
// ============================================================================

func (h *joseHandlerAdapter) handleJWSSign(c *fiber.Ctx) error {
	var req struct {
		KID     string `json:"kid"`
		Payload string `json:"payload"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.KID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Key ID is required",
		})
	}

	storedKey, exists := h.keyStore.Get(req.KID)
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Key not found",
		})
	}

	_, signed, err := cryptoutilSharedCryptoJose.SignBytes([]joseJwk.Key{storedKey.PrivateJWK}, []byte(req.Payload))
	if err != nil {
		h.telemetryService.Slogger.Error("Failed to sign JWS", "error", err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to sign payload",
		})
	}

	return c.JSON(fiber.Map{
		"jws": string(signed),
	})
}

func (h *joseHandlerAdapter) handleJWSVerify(c *fiber.Ctx) error {
	var req struct {
		JWS string `json:"jws"`
		KID string `json:"kid"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	storedKey, exists := h.keyStore.Get(req.KID)
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Key not found",
		})
	}

	var verifyKey joseJwk.Key
	if storedKey.PublicJWK != nil {
		verifyKey = storedKey.PublicJWK
	} else if storedKey.PrivateJWK != nil {
		pk, err := storedKey.PrivateJWK.PublicKey()
		if err != nil {
			verifyKey = storedKey.PrivateJWK
		} else {
			verifyKey = pk
		}
	}

	payload, err := cryptoutilSharedCryptoJose.VerifyBytes([]joseJwk.Key{verifyKey}, []byte(req.JWS))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Signature verification failed",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"valid":   true,
		"payload": string(payload),
	})
}

// ============================================================================
// JWE Handlers
// ============================================================================

func (h *joseHandlerAdapter) handleJWEEncrypt(c *fiber.Ctx) error {
	var req struct {
		KID       string `json:"kid"`
		Plaintext string `json:"plaintext"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.KID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Key ID is required",
		})
	}

	storedKey, exists := h.keyStore.Get(req.KID)
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Key not found",
		})
	}

	var encryptKey joseJwk.Key
	if storedKey.PublicJWK != nil {
		encryptKey = storedKey.PublicJWK
	} else if storedKey.PrivateJWK != nil {
		pk, err := storedKey.PrivateJWK.PublicKey()
		if err != nil {
			encryptKey = storedKey.PrivateJWK
		} else {
			encryptKey = pk
		}
	}

	_, encrypted, err := cryptoutilSharedCryptoJose.EncryptBytes([]joseJwk.Key{encryptKey}, []byte(req.Plaintext))
	if err != nil {
		h.telemetryService.Slogger.Error("Failed to encrypt JWE", "error", err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to encrypt payload",
		})
	}

	return c.JSON(fiber.Map{
		"jwe": string(encrypted),
	})
}

func (h *joseHandlerAdapter) handleJWEDecrypt(c *fiber.Ctx) error {
	var req struct {
		JWE string `json:"jwe"`
		KID string `json:"kid"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	storedKey, exists := h.keyStore.Get(req.KID)
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Key not found",
		})
	}

	decrypted, err := cryptoutilSharedCryptoJose.DecryptBytes([]joseJwk.Key{storedKey.PrivateJWK}, []byte(req.JWE))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Decryption failed",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"plaintext": string(decrypted),
	})
}

// ============================================================================
// JWT Handlers
// ============================================================================

func (h *joseHandlerAdapter) handleJWTSign(c *fiber.Ctx) error {
	var req struct {
		KID    string         `json:"kid"`
		Claims map[string]any `json:"claims"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.KID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Key ID is required",
		})
	}

	storedKey, exists := h.keyStore.Get(req.KID)
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Key not found",
		})
	}

	claimsJSON, _ := json.Marshal(req.Claims)

	_, signed, err := cryptoutilSharedCryptoJose.SignBytes([]joseJwk.Key{storedKey.PrivateJWK}, claimsJSON)
	if err != nil {
		h.telemetryService.Slogger.Error("Failed to sign JWT", "error", err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to sign JWT",
		})
	}

	return c.JSON(fiber.Map{
		"jwt": string(signed),
	})
}

func (h *joseHandlerAdapter) handleJWTVerify(c *fiber.Ctx) error {
	var req struct {
		JWT string `json:"jwt"`
		KID string `json:"kid"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	storedKey, exists := h.keyStore.Get(req.KID)
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Key not found",
		})
	}

	var verifyKey joseJwk.Key
	if storedKey.PublicJWK != nil {
		verifyKey = storedKey.PublicJWK
	} else if storedKey.PrivateJWK != nil {
		pk, err := storedKey.PrivateJWK.PublicKey()
		if err != nil {
			verifyKey = storedKey.PrivateJWK
		} else {
			verifyKey = pk
		}
	}

	payload, err := cryptoutilSharedCryptoJose.VerifyBytes([]joseJwk.Key{verifyKey}, []byte(req.JWT))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "JWT verification failed",
			"details": err.Error(),
		})
	}

	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return c.JSON(fiber.Map{
			"valid":   true,
			"payload": string(payload),
		})
	}

	return c.JSON(fiber.Map{
		"valid":  true,
		"claims": claims,
	})
}

// ============================================================================
// Elastic JWK JWKS Handler
// ============================================================================

// handleElasticJWKS handles GET /service/api/v1/jose/elastic-jwks/:kid/.well-known/jwks.json
// Returns public keys for an elastic JWK in JWKS format for external verification/encryption.
// Symmetric keys (AES, HMAC) return 404 since they cannot be published.
// Sets Cache-Control: max-age=300 for caching.
func (h *joseHandlerAdapter) handleElasticJWKS(c *fiber.Ctx) error {
	// Get elastic key ID from path parameter.
	kidStr := c.Params("kid")
	if kidStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing kid parameter",
		})
	}

	// Parse the KID.
	kid, err := googleUuid.Parse(kidStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid kid format",
		})
	}

	// Check if elasticJWKService is available.
	if h.elasticJWKService == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Elastic JWK service not configured",
		})
	}

	// Get public keys from elastic JWK service.
	// Note: tenant isolation is handled by the service layer using context.
	ctx := context.Background()

	publicJWKs, err := h.elasticJWKService.GetDecryptedPublicJWKs(ctx, defaultJWKSTenantID, defaultJWKSRealmID, kid)
	if err != nil {
		h.telemetryService.Slogger.Error("Failed to get elastic JWK public keys", "kid", kidStr, "error", err)

		// Check for "not found" errors.
		if isNotFoundError(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Elastic key not found",
			})
		}

		// Check for symmetric key errors (cannot publish symmetric keys).
		if isSymmetricKeyError(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "JWKS not available for symmetric keys",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve public keys",
		})
	}

	// Symmetric keys have nil public JWKs - return 404.
	if len(publicJWKs) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "JWKS not available for symmetric keys",
		})
	}

	// Build JWKS response.
	keys := make([]json.RawMessage, 0, len(publicJWKs))

	for _, jwk := range publicJWKs {
		if jwk == nil {
			continue
		}

		keyJSON, err := json.Marshal(jwk)
		if err != nil {
			h.telemetryService.Slogger.Error("Failed to marshal public JWK", "error", err)

			continue
		}

		keys = append(keys, keyJSON)
	}

	// Return 404 if no valid public keys (shouldn't happen for asymmetric keys).
	if len(keys) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No public keys available",
		})
	}

	// Set Cache-Control header (5 minutes).
	c.Set("Cache-Control", "max-age=300")
	c.Set("Content-Type", "application/json")

	// Return JWKS format.
	return c.JSON(fiber.Map{
		"keys": keys,
	})
}

// isNotFoundError checks if the error indicates a not-found condition.
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	return strings.Contains(errStr, "not found") || strings.Contains(errStr, "does not exist")
}

// isSymmetricKeyError checks if the error indicates a symmetric key (no public key).
func isSymmetricKeyError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	return strings.Contains(errStr, "symmetric") || strings.Contains(errStr, "no public key")
}
