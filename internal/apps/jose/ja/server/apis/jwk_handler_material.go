// Copyright (c) 2025 Justin Cranford
//

// Package apis provides HTTP API handlers for jose-ja service.
package apis

import (
	"time"

	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
)

// JWKHandler handles JWK-related HTTP requests.
func (h *JWKHandler) HandleListMaterialJWKs() fiber.Handler {
	return func(c *fiber.Ctx) error {
		kid := c.Params("kid")
		if kid == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Missing elastic key ID",
			})
		}

		// Get tenant from session context.
		tenantID := c.Locals("tenant_id")

		if tenantID == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing tenant context",
			})
		}

		tenantUUID, ok := tenantID.(googleUuid.UUID)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Invalid tenant ID format",
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
				"error": "Elastic JWK not found",
			})
		}

		materials, total, err := h.materialJWKRepo.ListByElasticJWK(ctx, elasticJWK.ID, offset, limit)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to list material JWKs",
			})
		}

		responses := make([]MaterialJWKResponse, len(materials))
		for i, mat := range materials {
			responses[i] = MaterialJWKResponse{
				MaterialKID:    mat.MaterialKID,
				ElasticJWKID:   mat.ElasticJWKID.String(),
				Active:         mat.Active,
				BarrierVersion: mat.BarrierVersion,
				CreatedAt:      mat.CreatedAt.Unix(),
			}

			if mat.RetiredAt != nil {
				retiredUnix := mat.RetiredAt.Unix()
				responses[i].RetiredAt = &retiredUnix
			}
		}

		return c.JSON(ListResponse{
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
				"error": "Missing elastic key ID",
			})
		}

		// Get tenant from session context.
		tenantID := c.Locals("tenant_id")

		if tenantID == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing tenant context",
			})
		}

		tenantUUID, ok := tenantID.(googleUuid.UUID)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Invalid tenant ID format",
			})
		}

		ctx := c.Context()

		// Verify elastic JWK exists and belongs to tenant.
		elasticJWK, err := h.elasticJWKRepo.Get(ctx, tenantUUID, kid)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Elastic JWK not found",
			})
		}

		material, err := h.materialJWKRepo.GetActiveMaterial(ctx, elasticJWK.ID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "No active material JWK found",
			})
		}

		return c.JSON(MaterialJWKResponse{
			MaterialKID:    material.MaterialKID,
			ElasticJWKID:   material.ElasticJWKID.String(),
			Active:         material.Active,
			BarrierVersion: material.BarrierVersion,
			CreatedAt:      material.CreatedAt.Unix(),
		})
	}
}

// HandleRotateMaterialJWK rotates the active material JWK.
func (h *JWKHandler) HandleRotateMaterialJWK() fiber.Handler {
	return func(c *fiber.Ctx) error {
		kid := c.Params("kid")
		if kid == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Missing elastic key ID",
			})
		}

		// Get tenant from session context.
		tenantID := c.Locals("tenant_id")

		if tenantID == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing tenant context",
			})
		}

		tenantUUID, ok := tenantID.(googleUuid.UUID)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Invalid tenant ID format",
			})
		}

		ctx := c.Context()

		// Verify elastic JWK exists and belongs to tenant.
		elasticJWK, err := h.elasticJWKRepo.Get(ctx, tenantUUID, kid)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Elastic JWK not found",
			})
		}

		// Check material limit.
		if elasticJWK.CurrentMaterialCount >= elasticJWK.MaxMaterials {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Maximum material keys reached, cannot rotate",
			})
		}

		// Create new material key.
		newMaterialKID := googleUuid.New()
		newMaterial := &cryptoutilAppsJoseJaDomain.MaterialJWK{
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
				"error": "Failed to rotate material JWK",
			})
		}

		// Increment material count.
		if err := h.elasticJWKRepo.IncrementMaterialCount(ctx, elasticJWK.ID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update material count",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(MaterialJWKResponse{
			MaterialKID:    newMaterial.MaterialKID,
			ElasticJWKID:   newMaterial.ElasticJWKID.String(),
			Active:         newMaterial.Active,
			BarrierVersion: newMaterial.BarrierVersion,
			CreatedAt:      newMaterial.CreatedAt.Unix(),
		})
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
			"error": "Sign operation not yet implemented",
		})
	}
}

// HandleVerify verifies a signature.
func (h *JWKHandler) HandleVerify() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Implement verification operation.
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
			"error": "Verify operation not yet implemented",
		})
	}
}

// HandleEncrypt encrypts data with a specific key.
func (h *JWKHandler) HandleEncrypt() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Implement encryption operation.
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
			"error": "Encrypt operation not yet implemented",
		})
	}
}

// HandleDecrypt decrypts data with a specific key.
func (h *JWKHandler) HandleDecrypt() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Implement decryption operation.
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
			"error": "Decrypt operation not yet implemented",
		})
	}
}

// mapAlgorithmToKeyType maps an algorithm string to a JWK key type.
func mapAlgorithmToKeyType(algorithm string) string {
	switch algorithm {
	case "RSA/2048", "RSA/3072", "RSA/4096":
		return cryptoutilAppsJoseJaDomain.KeyTypeRSA
	case "EC/P256", "EC/P384", "EC/P521":
		return cryptoutilAppsJoseJaDomain.KeyTypeEC
	case "OKP/Ed25519", "OKP/Ed448":
		return cryptoutilAppsJoseJaDomain.KeyTypeOKP
	case "oct/256", "oct/384", "oct/512":
		return cryptoutilAppsJoseJaDomain.KeyTypeOct
	default:
		return ""
	}
}
