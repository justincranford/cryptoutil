// Copyright (c) 2025 Justin Cranford
//
//

package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"

	cryptoutilCrypto "cryptoutil/internal/learn/crypto"
	cryptoutilDomain "cryptoutil/internal/learn/domain"
	"cryptoutil/internal/learn/repository"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilTemplateServer "cryptoutil/internal/template/server"
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
	tlsMaterial *cryptoutilTemplateServer.TLSMaterial
}

// NewPublicServer creates a new learn-im public server.
func NewPublicServer(
	ctx context.Context,
	port int,
	userRepo *repository.UserRepository,
	messageRepo *repository.MessageRepository,
	tlsCfg *cryptoutilTemplateServer.TLSConfig,
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
	tlsMaterial, err := cryptoutilTemplateServer.GenerateTLSMaterial(tlsCfg)
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

	// Business logic endpoints.
	s.app.Put("/service/api/v1/messages/tx", s.handleSendMessage)
	s.app.Get("/service/api/v1/messages/rx", s.handleReceiveMessages)
	s.app.Delete("/service/api/v1/messages/:id", s.handleDeleteMessage)

	s.app.Put("/browser/api/v1/messages/tx", s.handleSendMessage)
	s.app.Get("/browser/api/v1/messages/rx", s.handleReceiveMessages)
	s.app.Delete("/browser/api/v1/messages/:id", s.handleDeleteMessage)
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

	// For now, hardcode sender ID (will be replaced with authentication).
	// TODO: Extract sender ID from authentication context.
	senderID := googleUuid.New()

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

		// Encrypt message for this receiver.
		ephemeralPubKey, ciphertext, nonce, err := cryptoutilCrypto.EncryptMessage(plaintext, receiverPublicKey)
		if err != nil {
			//nolint:wrapcheck // Fiber framework error, wrapping not needed.
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to encrypt message",
			})
		}

		// Store encrypted content and nonce (same for all receivers).
		message.EncryptedContent = ciphertext
		message.Nonce = nonce

		// Create message receiver entry.
		messageReceiver := cryptoutilDomain.MessageReceiver{
			MessageID:    message.ID,
			ReceiverID:   receiverID,
			SenderPubKey: ephemeralPubKey,
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
	// For now, hardcode receiver ID (will be replaced with authentication).
	// TODO: Extract receiver ID from authentication context.
	receiverID := googleUuid.New()

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
			SenderPubKey:     fmt.Sprintf("%x", receiverEntry.SenderPubKey), // Hex encoding for simplicity.
			EncryptedContent: fmt.Sprintf("%x", msg.EncryptedContent),
			Nonce:            fmt.Sprintf("%x", msg.Nonce),
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

	// For now, skip ownership verification (will be added with authentication).
	// TODO: Verify that authenticated user is the sender.
	_ = message

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
