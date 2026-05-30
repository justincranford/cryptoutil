// Copyright (c) 2025-2026 Justin Cranford.
//

package handler

import (
	cryptoutilAppsSmKmsRepository "cryptoutil/internal/apps/sm-kms/server/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
)

// MessageHandler exposes legacy sm-kms messaging endpoints in sm-kms.
type MessageHandler struct {
	messageRepo             *cryptoutilAppsSmKmsRepository.MessageRepository
	messageRecipientJWKRepo *cryptoutilAppsSmKmsRepository.MessageRecipientJWKRepository
}

// NewMessageHandler creates a message compatibility handler.
func NewMessageHandler(
	messageRepo *cryptoutilAppsSmKmsRepository.MessageRepository,
	messageRecipientJWKRepo *cryptoutilAppsSmKmsRepository.MessageRecipientJWKRepository,
) *MessageHandler {
	return &MessageHandler{messageRepo: messageRepo, messageRecipientJWKRepo: messageRecipientJWKRepo}
}

// HandleListMessages serves GET /messages.
func (h *MessageHandler) HandleListMessages() fiber.Handler {
	return func(c *fiber.Ctx) error {
		recipientID, err := googleUuid.Parse(c.Query("recipient_id"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "recipient_id is required"})
		}

		messages, err := h.messageRepo.FindByRecipientID(c.UserContext(), recipientID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "failed to list messages"})
		}

		return c.Status(fiber.StatusOK).JSON(messages)
	}
}

// HandleGetMessage serves GET /messages/{messageID}.
func (h *MessageHandler) HandleGetMessage() fiber.Handler {
	return func(c *fiber.Ctx) error {
		messageID, err := googleUuid.Parse(c.Params("messageID"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "invalid messageID"})
		}

		message, err := h.messageRepo.FindByID(c.UserContext(), messageID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "message not found"})
		}

		return c.Status(fiber.StatusOK).JSON(message)
	}
}

// HandleDeleteMessage serves DELETE /messages/{messageID}.
func (h *MessageHandler) HandleDeleteMessage() fiber.Handler {
	return func(c *fiber.Ctx) error {
		messageID, err := googleUuid.Parse(c.Params("messageID"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "invalid messageID"})
		}

		if err := h.messageRecipientJWKRepo.DeleteByMessageID(c.UserContext(), messageID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "failed to delete message keys"})
		}

		if err := h.messageRepo.Delete(c.UserContext(), messageID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "failed to delete message"})
		}

		return c.SendStatus(fiber.StatusNoContent)
	}
}

// HandleSendMessage serves POST /messages/send.
func (h *MessageHandler) HandleSendMessage() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{cryptoutilSharedMagic.StringStatus: "message send endpoint wired"})
	}
}

// HandleReceiveMessages serves GET /messages/receive.
func (h *MessageHandler) HandleReceiveMessages() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{cryptoutilSharedMagic.StringStatus: "message receive endpoint wired"})
	}
}
