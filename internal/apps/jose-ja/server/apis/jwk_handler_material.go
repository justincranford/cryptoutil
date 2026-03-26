// Copyright (c) 2025 Justin Cranford
//

// Package apis provides HTTP API handlers for jose-ja service.
package apis

import (
	"time"

	cryptoutilAppsJoseJaModel "cryptoutil/internal/apps/jose-ja/model"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilJoseModels "cryptoutil/api/jose-ja/models"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
)

// JWKHandler handles JWK-related HTTP requests.
func (h *JWKHandler) HandleListMaterialJWKs() fiber.Handler {
	return func(c *fiber.Ctx) error {
		kid := c.Params("kid")
		if kid == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "Missing elastic key ID",
			})
		}

		// Get tenant from session context.
		tenantID := c.Locals("tenant_id")

		if tenantID == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "Missing tenant context",
			})
		}

		tenantUUID, ok := tenantID.(googleUuid.UUID)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "Invalid tenant ID format",
			})
		}

		// Parse pagination parameters.
		offset := c.QueryInt("offset", 0)
		limit := c.QueryInt("limit", cryptoutilSharedMagic.DefaultAPIListLimit)

		ctx := c.Context()

		// Verify elastic JWK exists and belongs to tenant.
		elasticJWK, err := h.elasticJWKRepo.Get(ctx, tenantUUID, kid)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "Elastic JWK not found",
			})
		}

		materials, total, err := h.materialJWKRepo.ListByElasticJWK(ctx, elasticJWK.ID, offset, limit)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "Failed to list material JWKs",
			})
		}

		responses := make([]cryptoutilJoseModels.MaterialJWKResponse, len(materials))
		for i := range materials {
			responses[i] = toMaterialJWKResponse(materials[i])
		}

		return c.JSON(cryptoutilJoseModels.MaterialJWKListResponse{
			Items: responses,
			Total: total,
		})
	}
}

// HandleGetActiveMaterialJWK gets the active material JWK for an elastic JWK.
func (h *JWKHandler) HandleGetActiveMaterialJWK() fiber.Handler {
	return func(c *fiber.Ctx) error {
		kid := c.Params("kid")
		if kid == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "Missing elastic key ID",
			})
		}

		// Get tenant from session context.
		tenantID := c.Locals("tenant_id")

		if tenantID == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "Missing tenant context",
			})
		}

		tenantUUID, ok := tenantID.(googleUuid.UUID)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "Invalid tenant ID format",
			})
		}

		ctx := c.Context()

		// Verify elastic JWK exists and belongs to tenant.
		elasticJWK, err := h.elasticJWKRepo.Get(ctx, tenantUUID, kid)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "Elastic JWK not found",
			})
		}

		material, err := h.materialJWKRepo.GetActiveMaterial(ctx, elasticJWK.ID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "No active material JWK found",
			})
		}

		return c.JSON(toMaterialJWKResponse(material))
	}
}

// HandleRotateMaterialJWK rotates the active material JWK.
func (h *JWKHandler) HandleRotateMaterialJWK() fiber.Handler {
	return func(c *fiber.Ctx) error {
		kid := c.Params("kid")
		if kid == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "Missing elastic key ID",
			})
		}

		// Get tenant from session context.
		tenantID := c.Locals("tenant_id")

		if tenantID == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "Missing tenant context",
			})
		}

		tenantUUID, ok := tenantID.(googleUuid.UUID)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "Invalid tenant ID format",
			})
		}

		ctx := c.Context()

		// Verify elastic JWK exists and belongs to tenant.
		elasticJWK, err := h.elasticJWKRepo.Get(ctx, tenantUUID, kid)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "Elastic JWK not found",
			})
		}

		// Check material limit.
		if elasticJWK.CurrentMaterialCount >= elasticJWK.MaxMaterials {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "Maximum material keys reached, cannot rotate",
			})
		}

		// Create new material key.
		newMaterialKID := googleUuid.New()
		newMaterial := &cryptoutilAppsJoseJaModel.MaterialJWK{
			ID:             newMaterialKID,
			ElasticJWKID:   elasticJWK.ID,
			MaterialKID:    newMaterialKID.String(),
			PrivateJWKJWE:  "encrypted-private-key-placeholder",
			PublicJWKJWE:   "encrypted-public-key-placeholder",
			Active:         true,
			CreatedAt:      time.Now().UTC(),
			BarrierVersion: 1,
		}

		// Rotate using repository transaction.
		if err := h.materialJWKRepo.RotateMaterial(ctx, elasticJWK.ID, newMaterial); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "Failed to rotate material JWK",
			})
		}

		// Increment material count.
		if err := h.elasticJWKRepo.IncrementMaterialCount(ctx, elasticJWK.ID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "Failed to update material count",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(toMaterialJWKResponse(newMaterial))
	}
}

// HandleGetJWKS returns the public JWKS for verification.
func (h *JWKHandler) HandleGetJWKS() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// JWKS endpoint is typically public - return all active public keys.
		// In production, this would filter by tenant or return only public keys.
		return c.JSON(fiber.Map{
			"keys": []any{},
		})
	}
}

// HandleSign signs data with a specific key.
func (h *JWKHandler) HandleSign() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Implement signing operation.
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: "Sign operation not yet implemented",
		})
	}
}

// HandleVerify verifies a signature.
func (h *JWKHandler) HandleVerify() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Implement verification operation.
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: "Verify operation not yet implemented",
		})
	}
}

// HandleEncrypt encrypts data with a specific key.
func (h *JWKHandler) HandleEncrypt() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Implement encryption operation.
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: "Encrypt operation not yet implemented",
		})
	}
}

// HandleDecrypt decrypts data with a specific key.
func (h *JWKHandler) HandleDecrypt() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Implement decryption operation.
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: "Decrypt operation not yet implemented",
		})
	}
}

// mapAlgorithmToKeyType maps an algorithm string to a JWK key type.
func mapAlgorithmToKeyType(algorithm string) string {
	switch algorithm {
	case cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilSharedMagic.JoseKeyTypeRSA3072, cryptoutilSharedMagic.JoseKeyTypeRSA4096:
		return cryptoutilAppsJoseJaModel.KeyTypeRSA
	case cryptoutilSharedMagic.JoseKeyTypeECP256, cryptoutilSharedMagic.JoseKeyTypeECP384, cryptoutilSharedMagic.JoseKeyTypeECP521:
		return cryptoutilAppsJoseJaModel.KeyTypeEC
	case cryptoutilSharedMagic.JoseKeyTypeOKPEd25519, "OKP/Ed448":
		return cryptoutilAppsJoseJaModel.KeyTypeOKP
	case cryptoutilSharedMagic.JoseKeyTypeOct256, cryptoutilSharedMagic.JoseKeyTypeOct384, cryptoutilSharedMagic.JoseKeyTypeOct512:
		return cryptoutilAppsJoseJaModel.KeyTypeOct
	default:
		return ""
	}
}
