// Copyright (c) 2025 Justin Cranford
//

// Package apis provides HTTP API handlers for jose-ja service.
package apis

import (
	"time"

	cryptoutilAppsFrameworkServiceServerBarrier "cryptoutil/internal/apps/framework/service/server/barrier"
	cryptoutilAppsJoseJaModel "cryptoutil/internal/apps/jose-ja/server/model"
	cryptoutilAppsJoseJaRepository "cryptoutil/internal/apps/jose-ja/server/repository"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilJoseModels "cryptoutil/api/jose-ja/models"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
)

// JWKHandler handles JWK-related HTTP requests.
type JWKHandler struct {
	elasticJWKRepo  cryptoutilAppsJoseJaRepository.ElasticJWKRepository
	materialJWKRepo cryptoutilAppsJoseJaRepository.MaterialJWKRepository
	auditConfigRepo cryptoutilAppsJoseJaRepository.AuditConfigRepository
	auditLogRepo    cryptoutilAppsJoseJaRepository.AuditLogRepository
	jwkGenService   *cryptoutilSharedCryptoJose.JWKGenService
	barrierService  *cryptoutilAppsFrameworkServiceServerBarrier.Service
}

// NewJWKHandler creates a new JWK handler.
func NewJWKHandler(
	elasticJWKRepo cryptoutilAppsJoseJaRepository.ElasticJWKRepository,
	materialJWKRepo cryptoutilAppsJoseJaRepository.MaterialJWKRepository,
	auditConfigRepo cryptoutilAppsJoseJaRepository.AuditConfigRepository,
	auditLogRepo cryptoutilAppsJoseJaRepository.AuditLogRepository,
	jwkGenService *cryptoutilSharedCryptoJose.JWKGenService,
	barrierService *cryptoutilAppsFrameworkServiceServerBarrier.Service,
) *JWKHandler {
	return &JWKHandler{
		elasticJWKRepo:  elasticJWKRepo,
		materialJWKRepo: materialJWKRepo,
		auditConfigRepo: auditConfigRepo,
		auditLogRepo:    auditLogRepo,
		jwkGenService:   jwkGenService,
		barrierService:  barrierService,
	}
}

// toElasticJWKResponse maps a domain ElasticJWK to a generated ElasticJWKResponse.
func toElasticJWKResponse(ejwk *cryptoutilAppsJoseJaModel.ElasticJWK) cryptoutilJoseModels.ElasticJWKResponse {
	return cryptoutilJoseModels.ElasticJWKResponse{
		Kid:                  ejwk.KID,
		TenantID:             ejwk.TenantID.String(),
		Kty:                  ejwk.KeyType,
		Alg:                  ejwk.Algorithm,
		Use:                  cryptoutilJoseModels.ElasticJWKResponseUse(ejwk.Use),
		MaxMaterials:         ejwk.MaxMaterials,
		CurrentMaterialCount: ejwk.CurrentMaterialCount,
		CreatedAt:            ejwk.CreatedAt.Unix(),
	}
}

// toMaterialJWKResponse maps a domain MaterialJWK to a generated MaterialJWKResponse.
func toMaterialJWKResponse(mat *cryptoutilAppsJoseJaModel.MaterialJWK) cryptoutilJoseModels.MaterialJWKResponse {
	resp := cryptoutilJoseModels.MaterialJWKResponse{
		MaterialKid:    mat.MaterialKID,
		ElasticJWKID:   mat.ElasticJWKID.String(),
		Active:         mat.Active,
		BarrierVersion: mat.BarrierVersion,
		CreatedAt:      mat.CreatedAt.Unix(),
	}

	if mat.RetiredAt != nil {
		retiredUnix := mat.RetiredAt.Unix()
		resp.RetiredAt = &retiredUnix
	}

	return resp
}

// HandleCreateElasticJWK creates a new elastic JWK container.
// CRITICAL: tenant_id for data scoping only - realms are authn-only, NOT data scope.
func (h *JWKHandler) HandleCreateElasticJWK() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req cryptoutilJoseModels.ElasticJWKCreateRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "Invalid request body",
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

		// Set default max materials.
		maxMaterials := cryptoutilSharedMagic.JoseJADefaultMaxMaterials
		if req.MaxMaterials != nil && *req.MaxMaterials > 0 {
			maxMaterials = *req.MaxMaterials
		}

		// Extract use as string.
		useStr := ""
		if req.Use != nil {
			useStr = string(*req.Use)
		}

		// Map algorithm to key type.
		keyType := mapAlgorithmToKeyType(req.Algorithm)
		if keyType == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "Invalid algorithm",
			})
		}

		// Generate KID.
		kid := googleUuid.New()

		// Create elastic JWK record.
		elasticJWK := &cryptoutilAppsJoseJaModel.ElasticJWK{
			ID:                   kid,
			TenantID:             tenantUUID,
			KID:                  kid.String(),
			KeyType:              keyType,
			Algorithm:            req.Algorithm,
			Use:                  useStr,
			MaxMaterials:         maxMaterials,
			CurrentMaterialCount: 0,
			CreatedAt:            time.Now().UTC(),
		}

		ctx := c.Context()
		if err := h.elasticJWKRepo.Create(ctx, elasticJWK); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "Failed to create elastic JWK",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(toElasticJWKResponse(elasticJWK))
	}
}

// HandleGetElasticJWK retrieves an elastic JWK by KID.
func (h *JWKHandler) HandleGetElasticJWK() fiber.Handler {
	return func(c *fiber.Ctx) error {
		kid := c.Params("kid")
		if kid == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "Missing key ID",
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

		elasticJWK, err := h.elasticJWKRepo.Get(ctx, tenantUUID, kid)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "Elastic JWK not found",
			})
		}

		return c.JSON(toElasticJWKResponse(elasticJWK))
	}
}

// HandleListElasticJWKs lists all elastic JWKs for a tenant.
func (h *JWKHandler) HandleListElasticJWKs() fiber.Handler {
	return func(c *fiber.Ctx) error {
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

		elasticJWKs, total, err := h.elasticJWKRepo.List(ctx, tenantUUID, offset, limit)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "Failed to list elastic JWKs",
			})
		}

		responses := make([]cryptoutilJoseModels.ElasticJWKResponse, len(elasticJWKs))
		for i := range elasticJWKs {
			responses[i] = toElasticJWKResponse(elasticJWKs[i])
		}

		return c.JSON(cryptoutilJoseModels.ElasticJWKListResponse{
			Items: responses,
			Total: total,
		})
	}
}

// HandleDeleteElasticJWK deletes an elastic JWK.
func (h *JWKHandler) HandleDeleteElasticJWK() fiber.Handler {
	return func(c *fiber.Ctx) error {
		kid := c.Params("kid")
		if kid == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "Missing key ID",
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

		// First, get the elastic JWK to verify ownership and get its ID.
		elasticJWK, err := h.elasticJWKRepo.Get(ctx, tenantUUID, kid)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "Elastic JWK not found",
			})
		}

		// Delete using the ID.
		if err := h.elasticJWKRepo.Delete(ctx, elasticJWK.ID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "Failed to delete elastic JWK",
			})
		}

		return c.SendStatus(fiber.StatusNoContent)
	}
}

// HandleCreateMaterialJWK creates a new material JWK for an elastic JWK.
func (h *JWKHandler) HandleCreateMaterialJWK() fiber.Handler {
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

		// Get elastic JWK.
		elasticJWK, err := h.elasticJWKRepo.Get(ctx, tenantUUID, kid)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "Elastic JWK not found",
			})
		}

		// Check material limit.
		if elasticJWK.CurrentMaterialCount >= elasticJWK.MaxMaterials {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "Maximum material keys reached",
			})
		}

		// Generate material key.
		materialKID := googleUuid.New()

		// Planned: Implement actual JWK generation based on elastic JWK algorithm.
		// For now, create placeholder material with string JWE placeholders.
		material := &cryptoutilAppsJoseJaModel.MaterialJWK{
			ID:             materialKID,
			ElasticJWKID:   elasticJWK.ID,
			MaterialKID:    materialKID.String(),
			PrivateJWKJWE:  "encrypted-private-key-placeholder",
			PublicJWKJWE:   "encrypted-public-key-placeholder",
			Active:         true,
			CreatedAt:      time.Now().UTC(),
			BarrierVersion: 1, // Planned: Get from barrier service.
		}

		if err := h.materialJWKRepo.Create(ctx, material); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "Failed to create material JWK",
			})
		}

		// Increment material count.
		if err := h.elasticJWKRepo.IncrementMaterialCount(ctx, elasticJWK.ID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "Failed to update material count",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(toMaterialJWKResponse(material))
	}
}

// HandleListMaterialJWKs lists all material JWKs for an elastic JWK.
