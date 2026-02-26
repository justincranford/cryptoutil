// Copyright (c) 2025 Justin Cranford

// Package apis provides business logic handlers for message operations.
package apis

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"

	cryptoutilAppsSmImDomain "cryptoutil/internal/apps/sm/im/domain"
	cryptoutilAppsSmImRepository "cryptoutil/internal/apps/sm/im/repository"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilAppsTemplateServiceServerMiddleware "cryptoutil/internal/apps/template/service/server/middleware"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
)

// MessageHandler handles message operations (send, receive, delete).
type MessageHandler struct {
	messageRepo             *cryptoutilAppsSmImRepository.MessageRepository
	messageRecipientJWKRepo *cryptoutilAppsSmImRepository.MessageRecipientJWKRepository
	jwkGenService           *cryptoutilSharedCryptoJose.JWKGenService
	barrierService          *cryptoutilAppsTemplateServiceServerBarrier.Service
}

// NewMessageHandler creates a new MessageHandler with injected dependencies.
func NewMessageHandler(
	messageRepo *cryptoutilAppsSmImRepository.MessageRepository,
	messageRecipientJWKRepo *cryptoutilAppsSmImRepository.MessageRecipientJWKRepository,
	jwkGenService *cryptoutilSharedCryptoJose.JWKGenService,
	barrierService *cryptoutilAppsTemplateServiceServerBarrier.Service,
) *MessageHandler {
	return &MessageHandler{
		messageRepo:             messageRepo,
		messageRecipientJWKRepo: messageRecipientJWKRepo,
		jwkGenService:           jwkGenService,
		barrierService:          barrierService,
	}
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

// HandleSendMessage handles PUT /messages/tx.
func (h *MessageHandler) HandleSendMessage() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req SendMessageRequest
		if err := c.BodyParser(&req); err != nil {
			//nolint:wrapcheck // Fiber framework error, wrapping not needed.
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "invalid request body",
			})
		}

		// Validate request.
		if len(req.ReceiverIDs) == 0 {
			//nolint:wrapcheck // Fiber framework error, wrapping not needed.
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "receiver_ids cannot be empty",
			})
		}

		if req.Message == "" {
			//nolint:wrapcheck // Fiber framework error, wrapping not needed.
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "message cannot be empty",
			})
		}

		// Extract sender ID from authentication context.
		senderID, ok := c.Locals(cryptoutilAppsTemplateServiceServerMiddleware.ContextKeyUserID).(googleUuid.UUID)
		if !ok {
			//nolint:wrapcheck // Fiber framework error, wrapping not needed.
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "authentication required",
			})
		}

		// Generate JWE JWK for message encryption using dir + A256GCM.
		_, cekJWK, _, cekJWKBytes, _, err := h.jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgDir)
		if err != nil {
			//nolint:wrapcheck // Fiber framework error, wrapping not needed.
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "failed to generate message encryption key",
			})
		}

		// Validate JWK bytes are not empty.
		if len(cekJWKBytes) == 0 {
			//nolint:wrapcheck // Fiber framework error, wrapping not needed.
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "generated JWK is empty",
			})
		}

		// Encrypt message using JWE Compact Serialization.
		jwks := []joseJwk.Key{cekJWK}

		jweMessage, jweCompactBytes, err := cryptoutilSharedCryptoJose.EncryptBytesWithContext(jwks, []byte(req.Message), nil)
		if err != nil {
			//nolint:wrapcheck // Fiber framework error, wrapping not needed.
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "failed to encrypt message",
			})
		}

		// Create message with JWE ciphertext.
		message := &cryptoutilAppsSmImDomain.Message{
			ID:       googleUuid.New(),
			SenderID: senderID,
			JWE:      string(jweCompactBytes),
		}

		// Save message.
		if err := h.messageRepo.Create(c.Context(), message); err != nil {
			//nolint:wrapcheck // Fiber framework error, wrapping not needed.
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "failed to save message",
			})
		}

		// Store per-recipient decryption keys.
		for _, recipientIDStr := range req.ReceiverIDs {
			recipientID, err := googleUuid.Parse(recipientIDStr)
			if err != nil {
				//nolint:wrapcheck // Fiber framework error, wrapping not needed.
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					cryptoutilSharedMagic.StringError: fmt.Sprintf("invalid recipient ID: %s", recipientIDStr),
				})
			}

			// Encrypt JWK for this recipient using barrier service.
			encryptedJWKBytes, err := h.barrierService.EncryptContentWithContext(c.Context(), cekJWKBytes)
			if err != nil {
				//nolint:wrapcheck // Fiber framework error, wrapping not needed.
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					cryptoutilSharedMagic.StringError: "failed to encrypt recipient JWK",
				})
			}

			// Store encrypted JWK for this recipient.
			messageRecipientJWK := &cryptoutilAppsSmImDomain.MessageRecipientJWK{
				ID:           googleUuid.New(),
				MessageID:    message.ID,
				RecipientID:  recipientID,
				EncryptedJWK: string(encryptedJWKBytes),
			}

			if err := h.messageRecipientJWKRepo.Create(c.Context(), messageRecipientJWK); err != nil {
				//nolint:wrapcheck // Fiber framework error, wrapping not needed.
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					cryptoutilSharedMagic.StringError: "failed to save recipient JWK",
				})
			}
		}

		_ = jweMessage // JWE message structure (contains headers).
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusCreated).JSON(SendMessageResponse{
			MessageID: message.ID.String(),
		})
	}
}

// HandleReceiveMessages handles GET /messages/rx.
func (h *MessageHandler) HandleReceiveMessages() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract recipient ID from authentication context.
		recipientID, ok := c.Locals(cryptoutilAppsTemplateServiceServerMiddleware.ContextKeyUserID).(googleUuid.UUID)
		if !ok {
			//nolint:wrapcheck // Fiber framework error, wrapping not needed.
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "authentication required",
			})
		}

		// Retrieve messages for recipient.
		messages, err := h.messageRepo.FindByRecipientID(c.Context(), recipientID)
		if err != nil {
			//nolint:wrapcheck // Fiber framework error, wrapping not needed.
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "failed to retrieve messages",
			})
		}

		// Mark messages as read.
		for _, msg := range messages {
			if err := h.messageRepo.MarkAsRead(c.Context(), msg.ID); err != nil {
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
			recipientJWKRecord, err := h.messageRecipientJWKRepo.FindByRecipientAndMessage(c.Context(), recipientID, msg.ID)
			if err != nil {
				// No JWK found for this recipient.
				continue
			}

			// Decrypt JWK for this recipient using barrier service.
			decryptedJWKBytes, err := h.barrierService.DecryptContentWithContext(c.Context(), []byte(recipientJWKRecord.EncryptedJWK))
			if err != nil {
				// Failed to decrypt JWK.
				continue
			}

			// Parse JWK JSON.
			cekJWK, err := joseJwk.ParseKey(decryptedJWKBytes)
			if err != nil {
				// Failed to parse JWK.
				continue
			}

			// Decrypt JWE to get plaintext message.
			jwks := []joseJwk.Key{cekJWK}

			plaintext, err := cryptoutilSharedCryptoJose.DecryptBytesWithContext(jwks, []byte(msg.JWE), nil)
			if err != nil {
				// Decryption failed.
				continue
			}

			response.Messages = append(response.Messages, MessageResponse{
				MessageID:        msg.ID.String(),
				SenderPubKey:     msg.Sender.Username,
				EncryptedContent: string(plaintext),
				Nonce:            "",
				CreatedAt:        msg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			})
		}

		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusOK).JSON(response)
	}
}

// HandleDeleteMessage handles DELETE /messages/:id.
func (h *MessageHandler) HandleDeleteMessage() fiber.Handler {
	return func(c *fiber.Ctx) error {
		messageIDStr := c.Params("id")
		if messageIDStr == "" {
			//nolint:wrapcheck // Fiber framework error, wrapping not needed.
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "message ID is required",
			})
		}

		messageID, err := googleUuid.Parse(messageIDStr)
		if err != nil {
			//nolint:wrapcheck // Fiber framework error, wrapping not needed.
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "invalid message ID",
			})
		}

		// Retrieve message to verify ownership.
		message, err := h.messageRepo.FindByID(c.Context(), messageID)
		if err != nil {
			//nolint:wrapcheck // Fiber framework error, wrapping not needed.
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "message not found",
			})
		}

		// Extract authenticated user ID from context.
		userID, ok := c.Locals(cryptoutilAppsTemplateServiceServerMiddleware.ContextKeyUserID).(googleUuid.UUID)
		if !ok {
			//nolint:wrapcheck // Fiber framework error, wrapping not needed.
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "authentication required",
			})
		}

		// Verify ownership (only sender can delete message).
		if message.SenderID != userID {
			//nolint:wrapcheck // Fiber framework error, wrapping not needed.
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "only the sender can delete this message",
			})
		}

		// Delete recipient JWKs first.
		if err := h.messageRecipientJWKRepo.DeleteByMessageID(c.Context(), messageID); err != nil {
			//nolint:wrapcheck // Fiber framework error, wrapping not needed.
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "failed to delete recipient JWKs",
			})
		}

		// Delete message.
		if err := h.messageRepo.Delete(c.Context(), messageID); err != nil {
			//nolint:wrapcheck // Fiber framework error, wrapping not needed.
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "failed to delete message",
			})
		}

		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.SendStatus(fiber.StatusNoContent)
	}
}
