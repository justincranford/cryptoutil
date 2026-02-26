// Copyright (c) 2025 Justin Cranford
//

// Package apis provides HTTP API handlers for jose-ja service.
package apis

import (
	json "encoding/json"
	"time"

	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilAppsJoseJaRepository "cryptoutil/internal/apps/jose/ja/repository"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

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
	barrierService  *cryptoutilAppsTemplateServiceServerBarrier.Service
}

// NewJWKHandler creates a new JWK handler.
func NewJWKHandler(
	elasticJWKRepo cryptoutilAppsJoseJaRepository.ElasticJWKRepository,
	materialJWKRepo cryptoutilAppsJoseJaRepository.MaterialJWKRepository,
	auditConfigRepo cryptoutilAppsJoseJaRepository.AuditConfigRepository,
	auditLogRepo cryptoutilAppsJoseJaRepository.AuditLogRepository,
	jwkGenService *cryptoutilSharedCryptoJose.JWKGenService,
	barrierService *cryptoutilAppsTemplateServiceServerBarrier.Service,
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

// CreateElasticJWKRequest represents the request to create an elastic JWK.
type CreateElasticJWKRequest struct {
	Algorithm    string `json:"algorithm"`               // Algorithm (e.g., "RSA/2048", "EC/P256").
	Use          string `json:"use"`                     // Key use: "sig" or "enc".
	MaxMaterials int    `json:"max_materials,omitempty"` // Max material keys (default: 10).
}

// ElasticJWKResponse represents the response for an elastic JWK.
// CRITICAL: No realm_id - realms are authn-only, NOT data scope.
type ElasticJWKResponse struct {
	KID                  string `json:"kid"`
	TenantID             string `json:"tenant_id"`
	KeyType              string `json:"kty"`
	Algorithm            string `json:"alg"`
	Use                  string `json:"use"`
	MaxMaterials         int    `json:"max_materials"`
	CurrentMaterialCount int    `json:"current_material_count"`
	CreatedAt            int64  `json:"created_at"`
}

// MaterialJWKResponse represents the response for a material JWK.
type MaterialJWKResponse struct {
	MaterialKID    string          `json:"material_kid"`
	ElasticJWKID   string          `json:"elastic_jwk_id"`
	PublicJWK      json.RawMessage `json:"public_jwk,omitempty"`
	Active         bool            `json:"active"`
	BarrierVersion int             `json:"barrier_version"`
	CreatedAt      int64           `json:"created_at"`
	RetiredAt      *int64          `json:"retired_at,omitempty"`
}

// ListResponse represents a paginated list response.
type ListResponse struct {
	Items any   `json:"items"`
	Total int64 `json:"total"`
}

// HandleCreateElasticJWK creates a new elastic JWK container.
// CRITICAL: tenant_id for data scoping only - realms are authn-only, NOT data scope.
func (h *JWKHandler) HandleCreateElasticJWK() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req CreateElasticJWKRequest
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
		maxMaterials := req.MaxMaterials
		if maxMaterials <= 0 {
			maxMaterials = cryptoutilSharedMagic.JoseJADefaultMaxMaterials
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
		elasticJWK := &cryptoutilAppsJoseJaDomain.ElasticJWK{
			ID:                   kid,
			TenantID:             tenantUUID,
			KID:                  kid.String(),
			KeyType:              keyType,
			Algorithm:            req.Algorithm,
			Use:                  req.Use,
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

		return c.Status(fiber.StatusCreated).JSON(ElasticJWKResponse{
			KID:                  elasticJWK.KID,
			TenantID:             elasticJWK.TenantID.String(),
			KeyType:              elasticJWK.KeyType,
			Algorithm:            elasticJWK.Algorithm,
			Use:                  elasticJWK.Use,
			MaxMaterials:         elasticJWK.MaxMaterials,
			CurrentMaterialCount: elasticJWK.CurrentMaterialCount,
			CreatedAt:            elasticJWK.CreatedAt.Unix(),
		})
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

		return c.JSON(ElasticJWKResponse{
			KID:                  elasticJWK.KID,
			TenantID:             elasticJWK.TenantID.String(),
			KeyType:              elasticJWK.KeyType,
			Algorithm:            elasticJWK.Algorithm,
			Use:                  elasticJWK.Use,
			MaxMaterials:         elasticJWK.MaxMaterials,
			CurrentMaterialCount: elasticJWK.CurrentMaterialCount,
			CreatedAt:            elasticJWK.CreatedAt.Unix(),
		})
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

		responses := make([]ElasticJWKResponse, len(elasticJWKs))
		for i, jwk := range elasticJWKs {
			responses[i] = ElasticJWKResponse{
				KID:                  jwk.KID,
				TenantID:             jwk.TenantID.String(),
				KeyType:              jwk.KeyType,
				Algorithm:            jwk.Algorithm,
				Use:                  jwk.Use,
				MaxMaterials:         jwk.MaxMaterials,
				CurrentMaterialCount: jwk.CurrentMaterialCount,
				CreatedAt:            jwk.CreatedAt.Unix(),
			}
		}

		return c.JSON(ListResponse{
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

		// TODO: Implement actual JWK generation based on elastic JWK algorithm.
		// For now, create placeholder material with string JWE placeholders.
		material := &cryptoutilAppsJoseJaDomain.MaterialJWK{
			ID:             materialKID,
			ElasticJWKID:   elasticJWK.ID,
			MaterialKID:    materialKID.String(),
			PrivateJWKJWE:  "encrypted-private-key-placeholder",
			PublicJWKJWE:   "encrypted-public-key-placeholder",
			Active:         true,
			CreatedAt:      time.Now().UTC(),
			BarrierVersion: 1, // TODO: Get from barrier service.
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

		return c.Status(fiber.StatusCreated).JSON(MaterialJWKResponse{
			MaterialKID:    material.MaterialKID,
			ElasticJWKID:   material.ElasticJWKID.String(),
			Active:         material.Active,
			BarrierVersion: material.BarrierVersion,
			CreatedAt:      material.CreatedAt.Unix(),
		})
	}
}

// HandleListMaterialJWKs lists all material JWKs for an elastic JWK.
