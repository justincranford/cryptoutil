// Copyright (c) 2025-2026 Justin Cranford.
//

package handler

import (
	cryptoutilAppsSmKmsRepository "cryptoutil/internal/apps/sm-kms/server/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
)

// JWKCompatHandler exposes legacy JOSE JWK management endpoints within sm-kms.
type JWKCompatHandler struct {
	elasticRepo  cryptoutilAppsSmKmsRepository.ElasticJWKRepository
	materialRepo cryptoutilAppsSmKmsRepository.MaterialJWKRepository
}

// NewJWKCompatHandler creates a new compatibility handler.
func NewJWKCompatHandler(
	elasticRepo cryptoutilAppsSmKmsRepository.ElasticJWKRepository,
	materialRepo cryptoutilAppsSmKmsRepository.MaterialJWKRepository,
) *JWKCompatHandler {
	return &JWKCompatHandler{elasticRepo: elasticRepo, materialRepo: materialRepo}
}

// HandleGetActiveMaterialJWK serves GET /elastic-keys/{elasticKeyID}/material-keys/active.
func (h *JWKCompatHandler) HandleGetActiveMaterialJWK() fiber.Handler {
	return func(c *fiber.Ctx) error {
		elasticID, err := googleUuid.Parse(c.Params("elasticKeyID"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "invalid elasticKeyID"})
		}

		mat, err := h.materialRepo.GetActiveMaterial(c.UserContext(), elasticID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "active material not found"})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"id":             mat.ID,
			"material_kid":   mat.MaterialKID,
			"elastic_jwk_id": mat.ElasticJWKID,
			"active":         mat.Active,
			"created_at":     mat.CreatedAt,
		})
	}
}

// HandleRotateMaterialJWK serves POST /elastic-keys/{elasticKeyID}/rotate.
func (h *JWKCompatHandler) HandleRotateMaterialJWK() fiber.Handler {
	return func(c *fiber.Ctx) error {
		elasticID, err := googleUuid.Parse(c.Params("elasticKeyID"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "invalid elasticKeyID"})
		}

		elastic, err := h.elasticRepo.GetByID(c.UserContext(), elasticID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "elastic key not found"})
		}

		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
			cryptoutilSharedMagic.StringStatus: "rotation requested",
			"elasticKeyID":                     elastic.ID,
		})
	}
}
