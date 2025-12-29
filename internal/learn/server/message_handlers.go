// Copyright (c) 2025 Justin Cranford
//

package server

import (
	"encoding/json"
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

	// Create message with JWE ciphertext.
	message := &cryptoutilDomain.Message{
		ID:       googleUuid.New(),
		SenderID: senderID,
		JWE:      string(jweCompactBytes),
	}

	// Save message.
	if err := s.messageRepo.Create(c.Context(), message); err != nil {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to save message",
		})
	}

	// Store per-recipient decryption keys in messages_recipient_jwks table.
	// This enables multi-recipient support where each recipient has their own JWK copy.
	// Future: Phase 5b will encrypt JWK with barrier service before storing.
	for _, recipientIDStr := range req.ReceiverIDs {
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

		// Store encrypted JWK for this recipient.
		// NOTE: Current implementation combines cekPubBytes + cekPrivBytes into JSON format.
		// Phase 5b will use barrier service to encrypt JWK before storing.
		// For now, concatenate pub+priv keys as single JSON object: {"pub":"...", "priv":"..."}
		jwkJSON := fmt.Sprintf(`{"keyID":"%s","pub":%s,"priv":%s}`, keyID, cekPubBytes, cekPrivBytes)

		messageRecipientJWK := &cryptoutilDomain.MessageRecipientJWK{
			ID:          googleUuid.New(),
			MessageID:   message.ID,
			RecipientID: recipientID,
			JWK:         jwkJSON,
		}

		if err := s.messageRecipientJWKRepo.Create(c.Context(), messageRecipientJWK); err != nil {
			//nolint:wrapcheck // Fiber framework error, wrapping not needed.
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to save recipient JWK",
			})
		}
	}

	// NOTE: Removed in-memory cache (s.messageKeysCache.Store) - now using database storage.
	_ = jweMessage // JWE message structure (contains headers, useful for Phase 5b audit logs).

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
		// Load recipient's JWK from messages_recipient_jwks table.
		recipientJWKRecord, err := s.messageRecipientJWKRepo.FindByRecipientAndMessage(c.Context(), recipientID, msg.ID)
		if err != nil {
			// No JWK found for this recipient (message not addressed to them, or data corrupted).
			continue
		}

		// Parse JWK JSON to extract private key bytes.
		// NOTE: Current format: {"keyID":"...", "pub":{...}, "priv":{...}}
		// Phase 5b will decrypt JWK using barrier service before parsing.
		var jwkData map[string]any
		if err := json.Unmarshal([]byte(recipientJWKRecord.JWK), &jwkData); err != nil {
			// Invalid JWK JSON format.
			continue
		}

		// Extract private key JWK (used for decryption).
		privKeyData, ok := jwkData["priv"]
		if !ok {
			// Missing private key in JWK.
			continue
		}

		privKeyBytes, err := json.Marshal(privKeyData)
		if err != nil {
			// Failed to serialize private key.
			continue
		}

		// Parse private key into JWK object.
		cekJWK, err := joseJwk.ParseKey(privKeyBytes)
		if err != nil {
			// Failed to parse JWK.
			continue
		}

		// Decrypt JWE to get plaintext message.
		jwks := []joseJwk.Key{cekJWK}

		plaintext, err := cryptoutilJose.DecryptBytes(jwks, []byte(msg.JWE))
		if err != nil {
			// Decryption failed (corrupted message or wrong key).
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

	// Delete recipient JWKs first (cascade delete for orphaned keys).
	if err := s.messageRecipientJWKRepo.DeleteByMessageID(c.Context(), messageID); err != nil {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete recipient JWKs",
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
