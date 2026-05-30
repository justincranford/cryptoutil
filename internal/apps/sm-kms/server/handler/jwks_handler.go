// Copyright (c) 2025-2026 Justin Cranford.
//

package handler

import (
	cryptoutilAppsSmKmsRepository "cryptoutil/internal/apps/sm-kms/server/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
)

// JWKSCompatHandler serves the public JWKS projection from active key materials.
type JWKSCompatHandler struct {
	elasticRepo  cryptoutilAppsSmKmsRepository.ElasticJWKRepository
	materialRepo cryptoutilAppsSmKmsRepository.MaterialJWKRepository
}

// NewJWKSCompatHandler creates a new JWKS compatibility handler.
func NewJWKSCompatHandler(
	elasticRepo cryptoutilAppsSmKmsRepository.ElasticJWKRepository,
	materialRepo cryptoutilAppsSmKmsRepository.MaterialJWKRepository,
) *JWKSCompatHandler {
	return &JWKSCompatHandler{elasticRepo: elasticRepo, materialRepo: materialRepo}
}

// HandleGetJWKS serves GET /jwks.
func (h *JWKSCompatHandler) HandleGetJWKS() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tenantID, err := googleUuid.Parse(c.Query("tenant_id"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "tenant_id is required"})
		}

		elasticKeys, _, err := h.elasticRepo.List(c.UserContext(), tenantID, 0, cryptoutilSharedMagic.JoseJADefaultListLimit)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "failed to list elastic keys"})
		}

		keys := make([]fiber.Map, 0, len(elasticKeys))
		for _, elastic := range elasticKeys {
			mat, err := h.materialRepo.GetActiveMaterial(c.UserContext(), elastic.ID)
			if err != nil {
				continue
			}

			keys = append(keys, fiber.Map{"kid": mat.MaterialKID, "kty": cryptoutilSharedMagic.KeyTypeOct, "use": cryptoutilSharedMagic.JoseKeyUseSig})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{"keys": keys})
	}
}
