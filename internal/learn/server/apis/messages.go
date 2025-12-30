// Copyright (c) 2025 Justin Cranford

// Package apis provides business API handlers for learn-im.
package apis

import (
	"github.com/gofiber/fiber/v2"

	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilLearnRepo "cryptoutil/internal/learn/repository"
)

// MessageHandler handles message-related operations (send, receive, delete).
type MessageHandler struct {
	messageRepo             *cryptoutilLearnRepo.MessageRepository
	messageRecipientJWKRepo *cryptoutilLearnRepo.MessageRecipientJWKRepository
	jwkGenService           *cryptoutilJose.JWKGenService
}

// NewMessageHandler creates a new message handler.
func NewMessageHandler(
	messageRepo *cryptoutilLearnRepo.MessageRepository,
	messageRecipientJWKRepo *cryptoutilLearnRepo.MessageRecipientJWKRepository,
	jwkGenService *cryptoutilJose.JWKGenService,
) *MessageHandler {
	return &MessageHandler{
		messageRepo:             messageRepo,
		messageRecipientJWKRepo: messageRecipientJWKRepo,
		jwkGenService:           jwkGenService,
	}
}

// HandleSendMessage returns a Fiber handler for sending encrypted messages.
// TODO: Implement full message encryption logic from deleted message_handlers.go.
func (h *MessageHandler) HandleSendMessage() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
			"error": "Message send not yet implemented (Phase 0.3 stub)",
		})
	}
}

// HandleReceiveMessages returns a Fiber handler for receiving messages.
// TODO: Implement message retrieval logic from deleted message_handlers.go.
func (h *MessageHandler) HandleReceiveMessages() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
			"error": "Message receive not yet implemented (Phase 0.3 stub)",
		})
	}
}

// HandleDeleteMessage returns a Fiber handler for deleting messages.
// TODO: Implement message deletion logic from deleted message_handlers.go.
func (h *MessageHandler) HandleDeleteMessage() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
			"error": "Message delete not yet implemented (Phase 0.3 stub)",
		})
	}
}
