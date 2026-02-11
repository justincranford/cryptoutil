// Copyright (c) 2025 Justin Cranford
//
//

package barrier

import (
	"fmt"
	"time"

	fiber "github.com/gofiber/fiber/v2"
)

const (
	// MinRotationReasonLength is the minimum length for rotation reason field.
	MinRotationReasonLength = 10
	// MaxRotationReasonLength is the maximum length for rotation reason field.
	MaxRotationReasonLength = 500
)

// RotateKeyRequest is the request payload for key rotation endpoints.
type RotateKeyRequest struct {
	Reason string `json:"reason" validate:"required,min=10,max=500"`
}

// RotateRootKeyResponse is the response for root key rotation.
type RotateRootKeyResponse struct {
	OldKeyUUID string `json:"old_key_uuid"`
	NewKeyUUID string `json:"new_key_uuid"`
	Reason     string `json:"reason"`
	RotatedAt  int64  `json:"rotated_at"` // Unix epoch milliseconds
}

// RotateIntermediateKeyResponse is the response for intermediate key rotation.
type RotateIntermediateKeyResponse struct {
	OldKeyUUID string `json:"old_key_uuid"`
	NewKeyUUID string `json:"new_key_uuid"`
	Reason     string `json:"reason"`
	RotatedAt  int64  `json:"rotated_at"` // Unix epoch milliseconds
}

// RotateContentKeyResponse is the response for content key rotation.
type RotateContentKeyResponse struct {
	NewKeyUUID string `json:"new_key_uuid"`
	Reason     string `json:"reason"`
	RotatedAt  int64  `json:"rotated_at"` // Unix epoch milliseconds
}

// HandleRotateRootKey handles POST /admin/api/v1/barrier/rotate/root requests.
func HandleRotateRootKey(rotationService *RotationService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req RotateKeyRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "invalid_request_body",
				"message": "Failed to parse request body",
			})
		}

		// Validate reason length
		if len(req.Reason) < MinRotationReasonLength {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "validation_error",
				"message": "Reason must be at least 10 characters",
			})
		}

		if len(req.Reason) > MaxRotationReasonLength {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "validation_error",
				"message": "Reason must be at most 500 characters",
			})
		}

		result, err := rotationService.RotateRootKey(c.Context(), req.Reason)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "rotation_failed",
				"message": fmt.Sprintf("Failed to rotate root key: %v", err),
			})
		}

		return c.Status(fiber.StatusOK).JSON(&RotateRootKeyResponse{
			OldKeyUUID: result.OldKeyUUID.String(),
			NewKeyUUID: result.NewKeyUUID.String(),
			Reason:     result.Reason,
			RotatedAt:  time.Now().UTC().UnixMilli(),
		})
	}
}

// HandleRotateIntermediateKey handles POST /admin/api/v1/barrier/rotate/intermediate requests.
func HandleRotateIntermediateKey(rotationService *RotationService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req RotateKeyRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "invalid_request_body",
				"message": "Failed to parse request body",
			})
		}

		// Validate reason length
		if len(req.Reason) < MinRotationReasonLength {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "validation_error",
				"message": "Reason must be at least 10 characters",
			})
		}

		if len(req.Reason) > MaxRotationReasonLength {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "validation_error",
				"message": "Reason must be at most 500 characters",
			})
		}

		result, err := rotationService.RotateIntermediateKey(c.Context(), req.Reason)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "rotation_failed",
				"message": fmt.Sprintf("Failed to rotate intermediate key: %v", err),
			})
		}

		return c.Status(fiber.StatusOK).JSON(&RotateIntermediateKeyResponse{
			OldKeyUUID: result.OldKeyUUID.String(),
			NewKeyUUID: result.NewKeyUUID.String(),
			Reason:     result.Reason,
			RotatedAt:  time.Now().UTC().UnixMilli(),
		})
	}
}

// HandleRotateContentKey handles POST /admin/api/v1/barrier/rotate/content requests.
func HandleRotateContentKey(rotationService *RotationService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req RotateKeyRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "invalid_request_body",
				"message": "Failed to parse request body",
			})
		}

		// Validate reason length
		if len(req.Reason) < MinRotationReasonLength {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "validation_error",
				"message": "Reason must be at least 10 characters",
			})
		}

		if len(req.Reason) > MaxRotationReasonLength {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "validation_error",
				"message": "Reason must be at most 500 characters",
			})
		}

		result, err := rotationService.RotateContentKey(c.Context(), req.Reason)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "rotation_failed",
				"message": fmt.Sprintf("Failed to rotate content key: %v", err),
			})
		}

		return c.Status(fiber.StatusOK).JSON(&RotateContentKeyResponse{
			NewKeyUUID: result.NewKeyUUID.String(),
			Reason:     result.Reason,
			RotatedAt:  time.Now().UTC().UnixMilli(),
		})
	}
}

// RegisterRotationRoutes registers rotation endpoints on the admin server.
// Routes:
//   - POST /admin/api/v1/barrier/rotate/root
//   - POST /admin/api/v1/barrier/rotate/intermediate
//   - POST /admin/api/v1/barrier/rotate/content
func RegisterRotationRoutes(adminServer *fiber.App, rotationService *RotationService) {
	if adminServer == nil {
		panic("adminServer must be non-nil")
	}

	if rotationService == nil {
		panic("rotationService must be non-nil")
	}

	adminV1 := adminServer.Group("/admin/api/v1/barrier/rotate")

	adminV1.Post("/root", HandleRotateRootKey(rotationService))
	adminV1.Post("/intermediate", HandleRotateIntermediateKey(rotationService))
	adminV1.Post("/content", HandleRotateContentKey(rotationService))
}
