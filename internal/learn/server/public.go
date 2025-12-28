// Copyright (c) 2025 Justin Cranford
//
//

package server

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"

	cryptoutilCrypto "cryptoutil/internal/learn/crypto"
	cryptoutilDomain "cryptoutil/internal/learn/domain"
	"cryptoutil/internal/learn/repository"
	cryptoutilConfig "cryptoutil/internal/shared/config"
	cryptoutilTLSGenerator "cryptoutil/internal/shared/config/tls_generator"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilMagic "cryptoutil/internal/shared/magic"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

// PublicServer implements the template.PublicServer interface for learn-im.
type PublicServer struct {
	port          int
	userRepo      *repository.UserRepository
	messageRepo   *repository.MessageRepository
	jwkGenService *cryptoutilJose.JWKGenService // JWK generation for message encryption
	jwtSecret     string                        // JWT signing secret for authentication

	// In-memory key cache for Phase 5a (no barrier service yet).
	// TODO(Phase 5b): Replace with barrier-encrypted database storage.
	messageKeysCache sync.Map // map[string]joseJwk.Key (keyID -> decryption JWK)

	app         *fiber.App
	mu          sync.RWMutex
	shutdown    bool
	actualPort  int
	tlsMaterial *cryptoutilConfig.TLSMaterial
}

// NewPublicServer creates a new learn-im public server.
func NewPublicServer(
	ctx context.Context,
	port int,
	userRepo *repository.UserRepository,
	messageRepo *repository.MessageRepository,
	jwkGenService *cryptoutilJose.JWKGenService,
	jwtSecret string,
	tlsCfg *cryptoutilTLSGenerator.TLSGeneratedSettings,
) (*PublicServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	} else if userRepo == nil {
		return nil, fmt.Errorf("user repository cannot be nil")
	} else if messageRepo == nil {
		return nil, fmt.Errorf("message repository cannot be nil")
	} else if jwkGenService == nil {
		return nil, fmt.Errorf("JWK generation service cannot be nil")
	} else if tlsCfg == nil {
		return nil, fmt.Errorf("TLS configuration cannot be nil")
	}

	// Generate TLS material using centralized infrastructure.
	tlsMaterial, err := cryptoutilTLSGenerator.GenerateTLSMaterial(tlsCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate TLS material: %w", err)
	}

	s := &PublicServer{
		port:          port,
		userRepo:      userRepo,
		messageRepo:   messageRepo,
		jwkGenService: jwkGenService,
		jwtSecret:     jwtSecret,
		app:           fiber.New(fiber.Config{DisableStartupMessage: true}),
		tlsMaterial:   tlsMaterial,
	}

	s.registerRoutes()

	return s, nil
}

// registerRoutes sets up the API endpoints.
func (s *PublicServer) registerRoutes() {
	// Health endpoints (required by template pattern).
	s.app.Get("/service/api/v1/health", s.handleServiceHealth)
	s.app.Get("/browser/api/v1/health", s.handleBrowserHealth)

	// User management endpoints (authentication - no JWT required).
	s.app.Post("/service/api/v1/users/register", s.handleRegisterUser)
	s.app.Post("/service/api/v1/users/login", s.handleLoginUser)
	s.app.Post("/browser/api/v1/users/register", s.handleRegisterUser)
	s.app.Post("/browser/api/v1/users/login", s.handleLoginUser)

	// Business logic endpoints (message operations - JWT required).
	s.app.Put("/service/api/v1/messages/tx", JWTMiddleware(s.jwtSecret), s.handleSendMessage)
	s.app.Get("/service/api/v1/messages/rx", JWTMiddleware(s.jwtSecret), s.handleReceiveMessages)
	s.app.Delete("/service/api/v1/messages/:id", JWTMiddleware(s.jwtSecret), s.handleDeleteMessage)

	s.app.Put("/browser/api/v1/messages/tx", JWTMiddleware(s.jwtSecret), s.handleSendMessage)
	s.app.Get("/browser/api/v1/messages/rx", JWTMiddleware(s.jwtSecret), s.handleReceiveMessages)
	s.app.Delete("/browser/api/v1/messages/:id", JWTMiddleware(s.jwtSecret), s.handleDeleteMessage)
}

// handleServiceHealth returns health status for service-to-service clients.
func (s *PublicServer) handleServiceHealth(c *fiber.Ctx) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.shutdown {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status": "shutting down",
		})
	}

	//nolint:wrapcheck // Fiber framework error, wrapping not needed.
	return c.JSON(fiber.Map{
		"status": "healthy",
	})
}

// handleBrowserHealth returns health status for browser clients.
func (s *PublicServer) handleBrowserHealth(c *fiber.Ctx) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.shutdown {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status": "shutting down",
		})
	}

	//nolint:wrapcheck // Fiber framework error, wrapping not needed.
	return c.JSON(fiber.Map{
		"status": "healthy",
	})
}

// RegisterUserRequest represents the request to register a new user.
type RegisterUserRequest struct {
	Username string `json:"username"` // Username (3-50 characters).
	Password string `json:"password"` // Password (minimum 8 characters).
}

// RegisterUserResponse represents the response after successful registration.
type RegisterUserResponse struct {
	UserID     string `json:"user_id"`     // Created user ID.
	PublicKey  string `json:"public_key"`  // User's ECDH public key (hex-encoded).
	PrivateKey string `json:"private_key"` // User's ECDH private key (hex-encoded, for testing only).
}

// LoginUserRequest represents the request to login.
type LoginUserRequest struct {
	Username string `json:"username"` // Username.
	Password string `json:"password"` // Password.
}

// LoginUserResponse represents the response after successful login.
type LoginUserResponse struct {
	Token     string `json:"token"`      // JWT authentication token.
	ExpiresAt string `json:"expires_at"` // Token expiration (RFC3339).
}

// handleRegisterUser implements POST /users/register.
func (s *PublicServer) handleRegisterUser(c *fiber.Ctx) error {
	// Parse request.
	var req RegisterUserRequest
	if err := c.BodyParser(&req); err != nil {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validate request.
	const (
		minUsernameLength = 3
		maxUsernameLength = 50
		minPasswordLength = 8
	)

	if len(req.Username) < minUsernameLength || len(req.Username) > maxUsernameLength {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "username must be 3-50 characters",
		})
	}

	if len(req.Password) < minPasswordLength {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "password must be at least 8 characters",
		})
	}

	// Check username uniqueness.
	existing, err := s.userRepo.FindByUsername(c.Context(), req.Username)
	if err == nil && existing != nil {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "username already exists",
		})
	}

	// Generate ECDH key pair for message encryption.
	privateKey, publicKeyBytes, err := cryptoutilCrypto.GenerateECDHKeyPair()
	if err != nil {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate encryption keys",
		})
	}

	// Extract private key bytes.
	privateKeyBytes := privateKey.Bytes()

	// Hash password using PBKDF2-HMAC-SHA256.
	passwordHash, err := cryptoutilCrypto.HashPassword(req.Password)
	if err != nil {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to hash password",
		})
	}

	// Encode password hash as hex for string storage.
	passwordHashHex := hex.EncodeToString(passwordHash)

	// Create user.
	// TODO(Phase 5): Remove ECDH key generation - keys are ephemeral per-message in 3-table design.
	user := &cryptoutilDomain.User{
		ID:           googleUuid.New(),
		Username:     req.Username,
		PasswordHash: passwordHashHex,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// TODO(Phase 5): Remove temporary ECDH key generation (not used in 3-table design).
	_ = publicKeyBytes  // Will be removed in Phase 5.
	_ = privateKeyBytes // Will be removed in Phase 5.

	if err := s.userRepo.Create(c.Context(), user); err != nil {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create user",
		})
	}

	// Return response.
	//nolint:wrapcheck // Fiber framework error, wrapping not needed.
	return c.Status(fiber.StatusCreated).JSON(RegisterUserResponse{
		UserID:     user.ID.String(),
		PublicKey:  hex.EncodeToString(publicKeyBytes),
		PrivateKey: hex.EncodeToString(privateKeyBytes),
	})
}

// handleLoginUser implements POST /users/login.
func (s *PublicServer) handleLoginUser(c *fiber.Ctx) error {
	// Parse request.
	var req LoginUserRequest
	if err := c.BodyParser(&req); err != nil {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validate request.
	if req.Username == "" || req.Password == "" {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "username and password are required",
		})
	}

	// Find user by username.
	user, err := s.userRepo.FindByUsername(c.Context(), req.Username)
	if err != nil || user == nil {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid credentials",
		})
	}

	// Decode hex-encoded password hash from database.
	storedPasswordHash, err := hex.DecodeString(user.PasswordHash)
	if err != nil {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to decode password hash",
		})
	}

	// Verify password.
	verified, err := cryptoutilCrypto.VerifyPassword(req.Password, storedPasswordHash)
	if err != nil || !verified {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid credentials",
		})
	}

	// Generate JWT token.
	token, expiresAt, err := GenerateJWT(user.ID, user.Username, s.jwtSecret)
	if err != nil {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate authentication token",
		})
	}

	//nolint:wrapcheck // Fiber framework error, wrapping not needed.
	return c.Status(fiber.StatusOK).JSON(LoginUserResponse{
		Token:     token,
		ExpiresAt: expiresAt.Format(time.RFC3339),
	})
}

// SendMessageRequest represents the request to send a message.
type SendMessageRequest struct {
	ReceiverIDs []string `json:"receiver_ids"` // Receiver user IDs (UUIDs).
	Message     string   `json:"message"`      // Plaintext message.
}

// SendMessageResponse represents the response after sending a message.
type SendMessageResponse struct {
	MessageID string `json:"message_id"` // Created message ID.
}

// handleSendMessage handles PUT /messages/tx.
func (s *PublicServer) handleSendMessage(c *fiber.Ctx) error {
	var req SendMessageRequest
	if err := c.BodyParser(&req); err != nil {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validate request.
	if len(req.ReceiverIDs) == 0 {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "receiver_ids cannot be empty",
		})
	}

	if req.Message == "" {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "message cannot be empty",
		})
	}

	// Extract sender ID from authentication context.
	senderID, ok := c.Locals(ContextKeyUserID).(googleUuid.UUID)
	if !ok {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authentication required",
		})
	}

	// TODO(Phase 5): Implement full JWE Compact Serialization encryption.
	// For now, store plaintext to unblock Phase 3 completion.
	// This is a temporary implementation that will be replaced in Phase 5.

	// Parse first receiver ID (simplified single-recipient model).
	recipientIDStr := req.ReceiverIDs[0]

	recipientID, err := googleUuid.Parse(recipientIDStr)
	if err != nil {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("invalid recipient ID: %s", recipientIDStr),
		})
	}

	// Verify recipient exists.
	_, err = s.userRepo.FindByID(c.Context(), recipientID)
	if err != nil {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fmt.Sprintf("recipient not found: %s", recipientIDStr),
		})
	}

	// Generate JWE JWK for message encryption using dir + A256GCM (direct key agreement with AES-256-GCM).
	keyID, cekJWK, _, cekPubBytes, cekPrivBytes, err := s.jwkGenService.GenerateJWEJWK(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgDir)
	if err != nil {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate message encryption key",
		})
	}

	// Encrypt message using JWE Compact Serialization.
	// Format: BASE64URL(header).BASE64URL(encrypted_key).BASE64URL(iv).BASE64URL(ciphertext).BASE64URL(tag)
	jwks := []joseJwk.Key{cekJWK}

	jweMessage, jweCompactBytes, err := cryptoutilJose.EncryptBytes(jwks, []byte(req.Message))
	if err != nil {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to encrypt message",
		})
	}

	// TODO(Phase 5): Replace with multi-recipient JWE JSON encryption.
	// Current: Single-recipient JWE Compact (temporary for compilation).
	// Future: EncryptBytesWithContext(plaintext, []RecipientJWK) → JWE JSON with N keys.
	message := &cryptoutilDomain.Message{
		ID:       googleUuid.New(),
		SenderID: senderID,
		JWE:      string(jweCompactBytes), // TODO: Replace with JWE JSON format (multi-recipient).
	}

	// TODO(Phase 5): Store encrypted recipient JWKs in messages_recipient_jwks table.
	_ = recipientID // Will be used in messages_recipient_jwks table.

	// TODO(Phase 5): Store encrypted recipient JWKs in messages_recipient_jwks table.
	_ = recipientID // Will be used in messages_recipient_jwks table.

	// Store decryption key in cache using message ID for Phase 5a (in-memory, acceptable for demo).
	// TODO(Phase 5b): Store encrypted JWK in messages_recipient_jwks table with barrier service.
	s.messageKeysCache.Store(message.ID.String(), cekJWK)

	// TODO(Phase 5b): Store encrypted JWK (cekPubBytes, cekPrivBytes) in messages_recipient_jwks table
	// using barrier service encryption instead of in-memory cache.
	_ = keyID        // Will be removed in Phase 5.
	_ = cekPubBytes  // Will be used in Phase 5b for encrypted storage.
	_ = cekPrivBytes // Will be used in Phase 5b for encrypted storage.
	_ = jweMessage   // JWE message structure (contains headers, useful for Phase 5b audit logs).

	// Save message.
	if err := s.messageRepo.Create(c.Context(), message); err != nil {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to save message",
		})
	}

	//nolint:wrapcheck // Fiber framework error, wrapping not needed.
	return c.Status(fiber.StatusCreated).JSON(SendMessageResponse{
		MessageID: message.ID.String(),
	})
}

// MessageResponse represents a message in the response.
type MessageResponse struct {
	MessageID        string `json:"message_id"`        // Message ID.
	SenderPubKey     string `json:"sender_pub_key"`    // Ephemeral sender public key (base64).
	EncryptedContent string `json:"encrypted_content"` // Encrypted message content (base64).
	Nonce            string `json:"nonce"`             // GCM nonce (base64).
	CreatedAt        string `json:"created_at"`        // Message timestamp.
}

// ReceiveMessagesResponse represents the response for receiving messages.
type ReceiveMessagesResponse struct {
	Messages []MessageResponse `json:"messages"`
}

// handleReceiveMessages handles GET /messages/rx.
func (s *PublicServer) handleReceiveMessages(c *fiber.Ctx) error {
	// Extract recipient ID from authentication context.
	recipientID, ok := c.Locals(ContextKeyUserID).(googleUuid.UUID)
	if !ok {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authentication required",
		})
	}

	// Retrieve messages for recipient.
	messages, err := s.messageRepo.FindByRecipientID(c.Context(), recipientID)
	if err != nil {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve messages",
		})
	}

	// Mark messages as read.
	for _, msg := range messages {
		if err := s.messageRepo.MarkAsRead(c.Context(), msg.ID); err != nil {
			// Log error but continue processing other messages.
			continue
		}
	}

	// Build response.
	response := ReceiveMessagesResponse{
		Messages: make([]MessageResponse, 0, len(messages)),
	}

	for _, msg := range messages {
		// TODO(Phase 5): Replace with DecryptBytesWithContext(msg.JWE, recipientJWK).
		// Current: Using temporary in-memory cache for single-recipient JWE Compact.
		// Future: Load encrypted recipientJWK from messages_recipient_jwks table.
		keyInterface, found := s.messageKeysCache.Load(msg.ID.String()) // TODO: Use RecipientID lookup
		if !found {
			// Key not found in cache (server restarted or key expired).
			// For Phase 5, will load from messages_recipient_jwks table.
			continue
		}

		cekJWK, ok := keyInterface.(joseJwk.Key)
		if !ok {
			// Invalid key type in cache (should never happen).
			continue
		}

		// Decrypt JWE to get plaintext message.
		jwks := []joseJwk.Key{cekJWK}

		plaintext, err := cryptoutilJose.DecryptBytes(jwks, []byte(msg.JWE))
		if err != nil {
			// Decryption failed (corrupted message or wrong key).
			// For Phase 5a, skip this message. Phase 5b will include audit logging.
			continue
		}

		response.Messages = append(response.Messages, MessageResponse{
			MessageID:        msg.ID.String(),
			SenderPubKey:     "",                // Not used with JWE Compact (symmetric encryption).
			EncryptedContent: string(plaintext), // Decrypted plaintext message.
			Nonce:            "",                // Not used with JWE Compact (nonce embedded in format).
			CreatedAt:        msg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	//nolint:wrapcheck // Fiber framework error, wrapping not needed.
	return c.Status(fiber.StatusOK).JSON(response)
}

// handleDeleteMessage handles DELETE /messages/:id.
func (s *PublicServer) handleDeleteMessage(c *fiber.Ctx) error {
	messageIDStr := c.Params("id")
	if messageIDStr == "" {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "message ID is required",
		})
	}

	messageID, err := googleUuid.Parse(messageIDStr)
	if err != nil {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid message ID",
		})
	}

	// Retrieve message to verify ownership.
	message, err := s.messageRepo.FindByID(c.Context(), messageID)
	if err != nil {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "message not found",
		})
	}

	// Extract authenticated user ID from context.
	userID, ok := c.Locals(ContextKeyUserID).(googleUuid.UUID)
	if !ok {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authentication required",
		})
	}

	// Verify ownership (only sender can delete message).
	if message.SenderID != userID {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "only the sender can delete this message",
		})
	}

	// Delete message.
	if err := s.messageRepo.Delete(c.Context(), messageID); err != nil {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete message",
		})
	}

	//nolint:wrapcheck // Fiber framework error, wrapping not needed.
	return c.SendStatus(fiber.StatusNoContent)
}

// Start starts the HTTPS server (implements template.PublicServer).
func (s *PublicServer) Start(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}

	// Create TCP listener.
	listenConfig := &net.ListenConfig{}

	listener, err := listenConfig.Listen(ctx, "tcp", fmt.Sprintf("%s:%d", cryptoutilMagic.IPv4Loopback, s.port))
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	s.mu.Lock()

	tcpAddr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		s.mu.Unlock()

		return fmt.Errorf("listener address is not *net.TCPAddr")
	}

	s.actualPort = tcpAddr.Port
	s.mu.Unlock()

	// Create TLS listener using centralized TLS material.
	tlsListener := tls.NewListener(listener, s.tlsMaterial.Config)

	// Start server in goroutine.
	errChan := make(chan error, 1)

	go func() {
		if err := s.app.Listener(tlsListener); err != nil {
			errChan <- fmt.Errorf("public server error: %w", err)
		} else {
			errChan <- nil
		}
	}()

	// Wait for either context cancellation or server error.
	select {
	case <-ctx.Done():
		// Context cancelled - trigger graceful shutdown.
		const shutdownTimeout = 5

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout*time.Second)
		defer cancel()

		_ = s.Shutdown(shutdownCtx)

		return fmt.Errorf("public server stopped: %w", ctx.Err())
	case err := <-errChan:
		return err
	}
}

// Shutdown gracefully shuts down the server (implements template.PublicServer).
func (s *PublicServer) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.shutdown {
		return fmt.Errorf("public server already shutdown")
	}

	s.shutdown = true

	if s.app != nil {
		if err := s.app.Shutdown(); err != nil {
			return fmt.Errorf("failed to shutdown fiber app: %w", err)
		}
	}

	return nil
}

// ActualPort returns the actual port the server is listening on (implements template.PublicServer).
func (s *PublicServer) ActualPort() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.actualPort
}
