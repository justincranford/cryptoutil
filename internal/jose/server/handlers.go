// Copyright (c) 2025 Justin Cranford
//
//

package server

import (
	json "encoding/json"
	"time"

	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

// Health check handlers.

func (s *Server) handleHealth(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "healthy",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) handleLivez(c *fiber.Ctx) error {
	return c.SendString("OK")
}

func (s *Server) handleReadyz(c *fiber.Ctx) error {
	return c.SendString("OK")
}

// JWK handlers.

// JWKGenerateRequest represents the request body for JWK generation.
type JWKGenerateRequest struct {
	Algorithm string `json:"algorithm"` // Algorithm: RSA/4096, RSA/3072, RSA/2048, EC/P521, EC/P384, EC/P256, OKP/Ed25519, oct/512, oct/384, oct/256.
	Use       string `json:"use"`       // Key use: sig (signing) or enc (encryption).
}

// JWKGenerateResponse represents the response for JWK generation.
type JWKGenerateResponse struct {
	KID       string          `json:"kid"`
	Algorithm string          `json:"algorithm"`
	Use       string          `json:"use"`
	KeyType   string          `json:"kty"`
	PublicJWK json.RawMessage `json:"public_jwk,omitempty"` // Public key JSON.
	CreatedAt int64           `json:"created_at"`
}

func (s *Server) handleJWKGenerate(c *fiber.Ctx) error {
	var req JWKGenerateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate algorithm.
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

	// Generate JWK based on use (signing vs encryption).
	if req.Use == cryptoutilSharedMagic.JoseKeyUseEnc {
		// Map GenerateAlgorithm to JWE encryption parameters.
		enc, keyAlg := mapToEncryptionAlgorithms(alg)

		// Generate the JWK using the JWE-specific method for proper headers.
		kid, privateJWK, publicJWK, _, _, err = s.jwkGenService.GenerateJWEJWK(&enc, &keyAlg)
		if err != nil {
			s.telemetryService.Slogger.Error("Failed to generate JWE JWK", "error", err)

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to generate encryption key",
			})
		}
	} else {
		// Map GenerateAlgorithm to JWS signature algorithm for signing keys.
		sigAlg := mapToSignatureAlgorithm(alg)

		// Generate the JWK using the JWS-specific method for proper headers.
		kid, privateJWK, publicJWK, _, _, err = s.jwkGenService.GenerateJWSJWK(sigAlg)
		if err != nil {
			s.telemetryService.Slogger.Error("Failed to generate JWS JWK", "error", err)

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to generate signing key",
			})
		}
	}

	// Determine key type from the generated key.
	kty := getKeyType(privateJWK)

	// Store the key.
	storedKey := &StoredKey{
		KID:        *kid,
		PrivateJWK: privateJWK,
		PublicJWK:  publicJWK,
		KeyType:    kty,
		Algorithm:  req.Algorithm,
		Use:        req.Use,
		CreatedAt:  time.Now().Unix(),
	}

	if err := s.keyStore.Store(storedKey); err != nil {
		s.telemetryService.Slogger.Error("Failed to store JWK", "error", err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to store key",
		})
	}

	// Serialize public key for response.
	var publicJWKJSON json.RawMessage

	if publicJWK != nil {
		publicJWKJSON, _ = json.Marshal(publicJWK)
	}

	s.telemetryService.Slogger.Info("Generated JWK", "kid", kid.String(), "algorithm", req.Algorithm)

	return c.Status(fiber.StatusCreated).JSON(JWKGenerateResponse{
		KID:       kid.String(),
		Algorithm: req.Algorithm,
		Use:       req.Use,
		KeyType:   kty,
		PublicJWK: publicJWKJSON,
		CreatedAt: storedKey.CreatedAt,
	})
}

func (s *Server) handleJWKGet(c *fiber.Ctx) error {
	kid := c.Params("kid")
	if kid == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing key ID",
		})
	}

	key, exists := s.keyStore.Get(kid)
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Key not found",
		})
	}

	var publicJWKJSON json.RawMessage

	if key.PublicJWK != nil {
		publicJWKJSON, _ = json.Marshal(key.PublicJWK)
	}

	return c.JSON(JWKGenerateResponse{
		KID:       key.KID.String(),
		Algorithm: key.Algorithm,
		Use:       key.Use,
		KeyType:   key.KeyType,
		PublicJWK: publicJWKJSON,
		CreatedAt: key.CreatedAt,
	})
}

func (s *Server) handleJWKDelete(c *fiber.Ctx) error {
	kid := c.Params("kid")
	if kid == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing key ID",
		})
	}

	if !s.keyStore.Delete(kid) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Key not found",
		})
	}

	s.telemetryService.Slogger.Info("Deleted JWK", "kid", kid)

	return c.SendStatus(fiber.StatusNoContent)
}

func (s *Server) handleJWKList(c *fiber.Ctx) error {
	keys := s.keyStore.List()
	response := make([]JWKGenerateResponse, 0, len(keys))

	for _, key := range keys {
		var publicJWKJSON json.RawMessage

		if key.PublicJWK != nil {
			publicJWKJSON, _ = json.Marshal(key.PublicJWK)
		}

		response = append(response, JWKGenerateResponse{
			KID:       key.KID.String(),
			Algorithm: key.Algorithm,
			Use:       key.Use,
			KeyType:   key.KeyType,
			PublicJWK: publicJWKJSON,
			CreatedAt: key.CreatedAt,
		})
	}

	return c.JSON(fiber.Map{
		"keys":  response,
		"count": len(response),
	})
}

func (s *Server) handleJWKS(c *fiber.Ctx) error {
	jwks := s.keyStore.GetJWKS()

	// Serialize JWKS to JSON.
	jwksJSON, err := json.Marshal(jwks)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to serialize JWKS",
		})
	}

	c.Set("Content-Type", "application/json")

	return c.Send(jwksJSON)
}

// JWS handlers.

// JWSSignRequest represents the request body for JWS signing.
type JWSSignRequest struct {
	KID     string `json:"kid"`     // Key ID to use for signing.
	Payload string `json:"payload"` // Base64URL-encoded payload to sign.
}

// JWSSignResponse represents the response for JWS signing.
type JWSSignResponse struct {
	JWS string `json:"jws"` // Compact JWS serialization.
}

func (s *Server) handleJWSSign(c *fiber.Ctx) error {
	var req JWSSignRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.KID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing key ID",
		})
	}

	if req.Payload == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing payload",
		})
	}

	key, exists := s.keyStore.Get(req.KID)
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Key not found",
		})
	}

	// Sign the payload.
	_, jwsBytes, err := cryptoutilSharedCryptoJose.SignBytes([]joseJwk.Key{key.PrivateJWK}, []byte(req.Payload))
	if err != nil {
		s.telemetryService.Slogger.Error("Failed to sign payload", "error", err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to sign payload",
		})
	}

	return c.JSON(JWSSignResponse{
		JWS: string(jwsBytes),
	})
}

// JWSVerifyRequest represents the request body for JWS verification.
type JWSVerifyRequest struct {
	JWS string `json:"jws"` // Compact JWS to verify.
	KID string `json:"kid"` // Optional: specific key ID to use for verification.
}

// JWSVerifyResponse represents the response for JWS verification.
type JWSVerifyResponse struct {
	Valid   bool   `json:"valid"`
	Payload string `json:"payload,omitempty"` // Decoded payload if valid.
	KID     string `json:"kid,omitempty"`     // Key ID used for verification.
	Error   string `json:"error,omitempty"`   // Error message if invalid.
}

func (s *Server) handleJWSVerify(c *fiber.Ctx) error {
	var req JWSVerifyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.JWS == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing JWS",
		})
	}

	// Get the verification key.
	var verifyKey joseJwk.Key

	var usedKID string

	if req.KID != "" {
		key, exists := s.keyStore.Get(req.KID)
		if !exists {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Key not found",
			})
		}

		verifyKey = key.PublicJWK
		if verifyKey == nil {
			verifyKey = key.PrivateJWK // Symmetric keys.
		}

		usedKID = req.KID
	} else {
		// Try to verify with any key in the key store.
		keys := s.keyStore.List()
		if len(keys) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "No keys available for verification",
			})
		}

		// Try each key until one works.
		for _, storedKey := range keys {
			tryKey := storedKey.PublicJWK
			if tryKey == nil {
				tryKey = storedKey.PrivateJWK
			}

			payload, err := cryptoutilSharedCryptoJose.VerifyBytes([]joseJwk.Key{tryKey}, []byte(req.JWS))
			if err == nil {
				return c.JSON(JWSVerifyResponse{
					Valid:   true,
					Payload: string(payload),
					KID:     storedKey.KID.String(),
				})
			}
		}

		return c.JSON(JWSVerifyResponse{
			Valid: false,
			Error: "No key could verify the signature",
		})
	}

	// Verify with specific key.
	payload, err := cryptoutilSharedCryptoJose.VerifyBytes([]joseJwk.Key{verifyKey}, []byte(req.JWS))
	if err != nil {
		return c.JSON(JWSVerifyResponse{
			Valid: false,
			Error: err.Error(),
		})
	}

	return c.JSON(JWSVerifyResponse{
		Valid:   true,
		Payload: string(payload),
		KID:     usedKID,
	})
}

// JWE handlers.

// JWEEncryptRequest represents the request body for JWE encryption.
type JWEEncryptRequest struct {
	KID       string `json:"kid"`       // Key ID to use for encryption.
	Plaintext string `json:"plaintext"` // Plaintext to encrypt.
}

// JWEEncryptResponse represents the response for JWE encryption.
type JWEEncryptResponse struct {
	JWE string `json:"jwe"` // Compact JWE serialization.
}

func (s *Server) handleJWEEncrypt(c *fiber.Ctx) error {
	var req JWEEncryptRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.KID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing key ID",
		})
	}

	if req.Plaintext == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing plaintext",
		})
	}

	key, exists := s.keyStore.Get(req.KID)
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Key not found",
		})
	}

	// Use public key for encryption (or symmetric key).
	encryptKey := key.PublicJWK
	if encryptKey == nil {
		encryptKey = key.PrivateJWK
	}

	// Encrypt the plaintext.
	_, jweBytes, err := cryptoutilSharedCryptoJose.EncryptBytes([]joseJwk.Key{encryptKey}, []byte(req.Plaintext))
	if err != nil {
		s.telemetryService.Slogger.Error("Failed to encrypt plaintext", "error", err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to encrypt plaintext",
		})
	}

	return c.JSON(JWEEncryptResponse{
		JWE: string(jweBytes),
	})
}

// JWEDecryptRequest represents the request body for JWE decryption.
type JWEDecryptRequest struct {
	JWE string `json:"jwe"` // Compact JWE to decrypt.
	KID string `json:"kid"` // Optional: specific key ID to use for decryption.
}

// JWEDecryptResponse represents the response for JWE decryption.
type JWEDecryptResponse struct {
	Plaintext string `json:"plaintext,omitempty"` // Decrypted plaintext.
	KID       string `json:"kid,omitempty"`       // Key ID used for decryption.
	Error     string `json:"error,omitempty"`     // Error message if decryption failed.
}

func (s *Server) handleJWEDecrypt(c *fiber.Ctx) error {
	var req JWEDecryptRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.JWE == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing JWE",
		})
	}

	if req.KID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing key ID",
		})
	}

	key, exists := s.keyStore.Get(req.KID)
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Key not found",
		})
	}

	// Use private key for decryption.
	plaintext, err := cryptoutilSharedCryptoJose.DecryptBytes([]joseJwk.Key{key.PrivateJWK}, []byte(req.JWE))
	if err != nil {
		return c.JSON(JWEDecryptResponse{
			Error: err.Error(),
		})
	}

	return c.JSON(JWEDecryptResponse{
		Plaintext: string(plaintext),
		KID:       req.KID,
	})
}

// JWT handlers.

// JWTCreateRequest represents the request body for JWT creation.
type JWTCreateRequest struct {
	KID    string         `json:"kid"`    // Key ID to use for signing.
	Claims map[string]any `json:"claims"` // JWT claims.
}

// JWTCreateResponse represents the response for JWT creation.
type JWTCreateResponse struct {
	JWT string `json:"jwt"` // Compact JWT serialization.
}

func (s *Server) handleJWTCreate(c *fiber.Ctx) error {
	var req JWTCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.KID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing key ID",
		})
	}

	if req.Claims == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing claims",
		})
	}

	key, exists := s.keyStore.Get(req.KID)
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Key not found",
		})
	}

	// Serialize claims to JSON for signing.
	claimsJSON, err := json.Marshal(req.Claims)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid claims",
		})
	}

	// Sign the claims as a JWT (JWS with JSON payload).
	_, jwtBytes, err := cryptoutilSharedCryptoJose.SignBytes([]joseJwk.Key{key.PrivateJWK}, claimsJSON)
	if err != nil {
		s.telemetryService.Slogger.Error("Failed to create JWT", "error", err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create JWT",
		})
	}

	return c.JSON(JWTCreateResponse{
		JWT: string(jwtBytes),
	})
}

// JWTVerifyRequest represents the request body for JWT verification.
type JWTVerifyRequest struct {
	JWT string `json:"jwt"` // Compact JWT to verify.
	KID string `json:"kid"` // Optional: specific key ID to use for verification.
}

// JWTVerifyResponse represents the response for JWT verification.
type JWTVerifyResponse struct {
	Valid  bool           `json:"valid"`
	Claims map[string]any `json:"claims,omitempty"` // Decoded claims if valid.
	KID    string         `json:"kid,omitempty"`    // Key ID used for verification.
	Error  string         `json:"error,omitempty"`  // Error message if invalid.
}

func (s *Server) handleJWTVerify(c *fiber.Ctx) error {
	var req JWTVerifyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.JWT == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing JWT",
		})
	}

	// Get the verification key.
	var verifyKey joseJwk.Key

	var usedKID string

	if req.KID != "" {
		key, exists := s.keyStore.Get(req.KID)
		if !exists {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Key not found",
			})
		}

		verifyKey = key.PublicJWK
		if verifyKey == nil {
			verifyKey = key.PrivateJWK
		}

		usedKID = req.KID
	} else {
		// Try to verify with any key in the key store.
		keys := s.keyStore.List()
		if len(keys) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "No keys available for verification",
			})
		}

		// Try each key until one works.
		for _, storedKey := range keys {
			tryKey := storedKey.PublicJWK
			if tryKey == nil {
				tryKey = storedKey.PrivateJWK
			}

			payload, err := cryptoutilSharedCryptoJose.VerifyBytes([]joseJwk.Key{tryKey}, []byte(req.JWT))
			if err == nil {
				var claims map[string]any
				if jsonErr := json.Unmarshal(payload, &claims); jsonErr != nil {
					return c.JSON(JWTVerifyResponse{
						Valid: false,
						Error: "Failed to parse claims",
					})
				}

				return c.JSON(JWTVerifyResponse{
					Valid:  true,
					Claims: claims,
					KID:    storedKey.KID.String(),
				})
			}
		}

		return c.JSON(JWTVerifyResponse{
			Valid: false,
			Error: "No key could verify the signature",
		})
	}

	// Verify with specific key.
	payload, err := cryptoutilSharedCryptoJose.VerifyBytes([]joseJwk.Key{verifyKey}, []byte(req.JWT))
	if err != nil {
		return c.JSON(JWTVerifyResponse{
			Valid: false,
			Error: err.Error(),
		})
	}

	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return c.JSON(JWTVerifyResponse{
			Valid: false,
			Error: "Failed to parse claims",
		})
	}

	return c.JSON(JWTVerifyResponse{
		Valid:  true,
		Claims: claims,
		KID:    usedKID,
	})
}

// Helper functions.

func isValidAlgorithm(alg cryptoutilOpenapiModel.GenerateAlgorithm) bool {
	validAlgorithms := []cryptoutilOpenapiModel.GenerateAlgorithm{
		cryptoutilOpenapiModel.RSA4096,
		cryptoutilOpenapiModel.RSA3072,
		cryptoutilOpenapiModel.RSA2048,
		cryptoutilOpenapiModel.ECP521,
		cryptoutilOpenapiModel.ECP384,
		cryptoutilOpenapiModel.ECP256,
		cryptoutilOpenapiModel.OKPEd25519,
		cryptoutilOpenapiModel.Oct512,
		cryptoutilOpenapiModel.Oct384,
		cryptoutilOpenapiModel.Oct256,
		cryptoutilOpenapiModel.Oct192,
		cryptoutilOpenapiModel.Oct128,
	}

	for _, valid := range validAlgorithms {
		if alg == valid {
			return true
		}
	}

	return false
}

func getKeyType(jwk joseJwk.Key) string {
	if jwk == nil {
		return ""
	}

	kty := jwk.KeyType()

	return kty.String()
}

// mapToSignatureAlgorithm maps GenerateAlgorithm to the corresponding JWA signature algorithm.
func mapToSignatureAlgorithm(alg cryptoutilOpenapiModel.GenerateAlgorithm) joseJwa.SignatureAlgorithm {
	switch alg {
	case cryptoutilOpenapiModel.RSA4096:
		return joseJwa.PS512()
	case cryptoutilOpenapiModel.RSA3072:
		return joseJwa.PS384()
	case cryptoutilOpenapiModel.RSA2048:
		return joseJwa.PS256()
	case cryptoutilOpenapiModel.ECP521:
		return joseJwa.ES512()
	case cryptoutilOpenapiModel.ECP384:
		return joseJwa.ES384()
	case cryptoutilOpenapiModel.ECP256:
		return joseJwa.ES256()
	case cryptoutilOpenapiModel.OKPEd25519:
		return joseJwa.EdDSA()
	case cryptoutilOpenapiModel.Oct512:
		return joseJwa.HS512()
	case cryptoutilOpenapiModel.Oct384:
		return joseJwa.HS384()
	case cryptoutilOpenapiModel.Oct256, cryptoutilOpenapiModel.Oct192, cryptoutilOpenapiModel.Oct128:
		return joseJwa.HS256()
	default:
		return joseJwa.ES256()
	}
}

// mapToEncryptionAlgorithms maps GenerateAlgorithm to the corresponding JWE enc and alg parameters.
func mapToEncryptionAlgorithms(alg cryptoutilOpenapiModel.GenerateAlgorithm) (joseJwa.ContentEncryptionAlgorithm, joseJwa.KeyEncryptionAlgorithm) {
	switch alg {
	case cryptoutilOpenapiModel.RSA4096:
		return joseJwa.A256GCM(), joseJwa.RSA_OAEP_512()
	case cryptoutilOpenapiModel.RSA3072:
		return joseJwa.A256GCM(), joseJwa.RSA_OAEP_384()
	case cryptoutilOpenapiModel.RSA2048:
		return joseJwa.A256GCM(), joseJwa.RSA_OAEP_256()
	case cryptoutilOpenapiModel.ECP521:
		return joseJwa.A256GCM(), joseJwa.ECDH_ES_A256KW()
	case cryptoutilOpenapiModel.ECP384:
		return joseJwa.A256GCM(), joseJwa.ECDH_ES_A192KW()
	case cryptoutilOpenapiModel.ECP256:
		return joseJwa.A256GCM(), joseJwa.ECDH_ES_A128KW()
	case cryptoutilOpenapiModel.Oct512:
		return joseJwa.A256GCM(), joseJwa.A256GCMKW()
	case cryptoutilOpenapiModel.Oct384:
		return joseJwa.A192GCM(), joseJwa.A192GCMKW()
	case cryptoutilOpenapiModel.Oct256:
		return joseJwa.A256GCM(), joseJwa.DIRECT()
	case cryptoutilOpenapiModel.Oct192:
		return joseJwa.A192GCM(), joseJwa.DIRECT()
	case cryptoutilOpenapiModel.Oct128:
		return joseJwa.A128GCM(), joseJwa.DIRECT()
	default:
		// Default to A256GCM with ECDH-ES+A256KW for unknown algorithms.
		return joseJwa.A256GCM(), joseJwa.ECDH_ES_A256KW()
	}
}
