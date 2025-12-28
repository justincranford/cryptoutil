// Copyright (c) 2025 Justin Cranford
//

package server

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"

	cryptoutilDomain "cryptoutil/internal/learn/domain"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

// SendMessageRequest represents the request to send a message.
type SendMessageRequest struct {
	ReceiverIDs []string `json:"receiver_ids"` // Receiver user IDs (UUIDs).
	Message     string   `json:"message"`      // Plaintext message.
}

// SendMessageResponse represents the response after sending a message.
type SendMessageResponse struct {
	MessageID string `json:"message_id"` // Created message ID.
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

	// NOTE: Phase 5 will implement JWE Compact Serialization for multi-recipient encryption.
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

	// NOTE: Current implementation uses single-recipient JWE Compact format.
	// Phase 5 will implement multi-recipient JWE JSON encryption.
	// Future: EncryptBytesWithContext(plaintext, []RecipientJWK) → JWE JSON with N keys.
	message := &cryptoutilDomain.Message{
		ID:       googleUuid.New(),
		SenderID: senderID,
		JWE:      string(jweCompactBytes), // NOTE: JWE JSON format in Phase 5.
	}

	// NOTE: Phase 5 will store encrypted recipient JWKs in messages_recipient_jwks table.
	_ = recipientID // Will be used in messages_recipient_jwks table.

	// Store decryption key in cache using message ID for Phase 5a (in-memory, acceptable for demo).
	// NOTE: Phase 5b will store encrypted JWK in messages_recipient_jwks table with barrier service.
	s.messageKeysCache.Store(message.ID.String(), cekJWK)

	// NOTE: Encrypted JWK (cekPubBytes, cekPrivBytes) stored in messages_recipient_jwks table (Phase 5b)
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
		// NOTE: Phase 5 will use DecryptBytesWithContext(msg.JWE, recipientJWK).
		// Current: Using temporary in-memory cache for single-recipient JWE Compact.
		// Future: Load encrypted recipientJWK from messages_recipient_jwks table.
		keyInterface, found := s.messageKeysCache.Load(msg.ID.String()) // NOTE: Phase 5 uses RecipientID lookup
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
