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
	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// PublicServer implements the template.PublicServer interface for learn-im.
type PublicServer struct {
	port        int
	userRepo    *repository.UserRepository
	messageRepo *repository.MessageRepository

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
	tlsCfg *cryptoutilTLSGenerator.TLSGeneratedSettings,
) (*PublicServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}

	if userRepo == nil {
		return nil, fmt.Errorf("user repository cannot be nil")
	}

	if messageRepo == nil {
		return nil, fmt.Errorf("message repository cannot be nil")
	}

	if tlsCfg == nil {
		return nil, fmt.Errorf("TLS configuration cannot be nil")
	}

	// Generate TLS material using centralized infrastructure.
	tlsMaterial, err := cryptoutilTLSGenerator.GenerateTLSMaterial(tlsCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate TLS material: %w", err)
	}

	s := &PublicServer{
		port:        port,
		userRepo:    userRepo,
		messageRepo: messageRepo,
		app:         fiber.New(fiber.Config{DisableStartupMessage: true}),
		tlsMaterial: tlsMaterial,
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
	s.app.Put("/service/api/v1/messages/tx", JWTMiddleware(JWTSecret), s.handleSendMessage)
	s.app.Get("/service/api/v1/messages/rx", JWTMiddleware(JWTSecret), s.handleReceiveMessages)
	s.app.Delete("/service/api/v1/messages/:id", JWTMiddleware(JWTSecret), s.handleDeleteMessage)

	s.app.Put("/browser/api/v1/messages/tx", JWTMiddleware(JWTSecret), s.handleSendMessage)
	s.app.Get("/browser/api/v1/messages/rx", JWTMiddleware(JWTSecret), s.handleReceiveMessages)
	s.app.Delete("/browser/api/v1/messages/:id", JWTMiddleware(JWTSecret), s.handleDeleteMessage)
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
	// NOTE: Storing private key on server is ONLY acceptable for educational demo purposes.
	// Production systems should use client-side key management.
	user := &cryptoutilDomain.User{
		ID:           googleUuid.New(),
		Username:     req.Username,
		PasswordHash: passwordHashHex,
		PublicKey:    publicKeyBytes,
		PrivateKey:   privateKeyBytes,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

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
	token, expiresAt, err := GenerateJWT(user.ID, user.Username, JWTSecret)
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

	// Create message with receivers.
	message := &cryptoutilDomain.Message{
		ID:       googleUuid.New(),
		SenderID: senderID,
	}

	plaintext := []byte(req.Message)

	// Encrypt message for each receiver.
	for _, receiverIDStr := range req.ReceiverIDs {
		receiverID, err := googleUuid.Parse(receiverIDStr)
		if err != nil {
			//nolint:wrapcheck // Fiber framework error, wrapping not needed.
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("invalid receiver ID: %s", receiverIDStr),
			})
		}

		// Lookup receiver's public key.
		receiver, err := s.userRepo.FindByID(c.Context(), receiverID)
		if err != nil {
			//nolint:wrapcheck // Fiber framework error, wrapping not needed.
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": fmt.Sprintf("receiver not found: %s", receiverIDStr),
			})
		}

		// Parse receiver's ECDH public key.
		receiverPublicKey, err := cryptoutilCrypto.ParseECDHPublicKey(receiver.PublicKey)
		if err != nil {
			//nolint:wrapcheck // Fiber framework error, wrapping not needed.
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to parse receiver public key",
			})
		}

		// Encrypt message for this specific receiver.
		// Each receiver gets their own encrypted copy because ECDH produces
		// a different shared secret for each receiver's public key.
		ephemeralPubKey, ciphertext, nonce, err := cryptoutilCrypto.EncryptMessage(plaintext, receiverPublicKey)
		if err != nil {
			//nolint:wrapcheck // Fiber framework error, wrapping not needed.
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to encrypt message",
			})
		}

		// Create message receiver entry with this receiver's encrypted copy.
		messageReceiver := cryptoutilDomain.MessageReceiver{
			ID:               googleUuid.New(),
			MessageID:        message.ID,
			ReceiverID:       receiverID,
			SenderPubKey:     ephemeralPubKey,
			EncryptedContent: ciphertext,
			Nonce:            nonce,
		}

		message.Receivers = append(message.Receivers, messageReceiver)
	}

	// Save message with receivers.
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
	// Extract receiver ID from authentication context.
	receiverID, ok := c.Locals(ContextKeyUserID).(googleUuid.UUID)
	if !ok {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authentication required",
		})
	}

	// Retrieve messages for receiver.
	messages, err := s.messageRepo.FindByReceiverID(c.Context(), receiverID)
	if err != nil {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve messages",
		})
	}

	// Mark messages as received.
	for _, msg := range messages {
		if err := s.messageRepo.MarkAsReceived(c.Context(), msg.ID, receiverID); err != nil {
			// Log error but continue processing other messages.
			continue
		}
	}

	// Build response.
	response := ReceiveMessagesResponse{
		Messages: make([]MessageResponse, 0, len(messages)),
	}

	for _, msg := range messages {
		// Find receiver entry for this receiver.
		var receiverEntry *cryptoutilDomain.MessageReceiver

		for i := range msg.Receivers {
			if msg.Receivers[i].ReceiverID == receiverID {
				receiverEntry = &msg.Receivers[i]

				break
			}
		}

		if receiverEntry == nil {
			// Skip if receiver entry not found (shouldn't happen).
			continue
		}

		response.Messages = append(response.Messages, MessageResponse{
			MessageID:        msg.ID.String(),
			SenderPubKey:     fmt.Sprintf("%x", receiverEntry.SenderPubKey),     // Hex encoding for simplicity.
			EncryptedContent: fmt.Sprintf("%x", receiverEntry.EncryptedContent), // Receiver's encrypted copy.
			Nonce:            fmt.Sprintf("%x", receiverEntry.Nonce),            // Receiver's nonce.
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
