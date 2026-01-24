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
	Items interface{} `json:"items"`
	Total int64       `json:"total"`
}

// defaultLimit is the default pagination limit.
const defaultLimit = 100

// HandleCreateElasticJWK creates a new elastic JWK container.
// CRITICAL: tenant_id for data scoping only - realms are authn-only, NOT data scope.
func (h *JWKHandler) HandleCreateElasticJWK() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req CreateElasticJWKRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
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

		// Set default max materials.
		maxMaterials := req.MaxMaterials
		if maxMaterials <= 0 {
			maxMaterials = 10
		}

		// Map algorithm to key type.
		keyType := mapAlgorithmToKeyType(req.Algorithm)
		if keyType == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid algorithm",
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
			CreatedAt:            time.Now(),
		}

		ctx := c.Context()
		if err := h.elasticJWKRepo.Create(ctx, elasticJWK); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create elastic JWK",
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
				"error": "Missing key ID",
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

		elasticJWK, err := h.elasticJWKRepo.Get(ctx, tenantUUID, kid)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Elastic JWK not found",
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
		limit := c.QueryInt("limit", defaultLimit)

		ctx := c.Context()

		elasticJWKs, total, err := h.elasticJWKRepo.List(ctx, tenantUUID, offset, limit)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to list elastic JWKs",
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
				"error": "Missing key ID",
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

		// First, get the elastic JWK to verify ownership and get its ID.
		elasticJWK, err := h.elasticJWKRepo.Get(ctx, tenantUUID, kid)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Elastic JWK not found",
			})
		}

		// Delete using the ID.
		if err := h.elasticJWKRepo.Delete(ctx, elasticJWK.ID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to delete elastic JWK",
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

		// Get elastic JWK.
		elasticJWK, err := h.elasticJWKRepo.Get(ctx, tenantUUID, kid)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Elastic JWK not found",
			})
		}

		// Check material limit.
		if elasticJWK.CurrentMaterialCount >= elasticJWK.MaxMaterials {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Maximum material keys reached",
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
			CreatedAt:      time.Now(),
			BarrierVersion: 1, // TODO: Get from barrier service.
		}

		if err := h.materialJWKRepo.Create(ctx, material); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create material JWK",
			})
		}

		// Increment material count.
		if err := h.elasticJWKRepo.IncrementMaterialCount(ctx, elasticJWK.ID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update material count",
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
		limit := c.QueryInt("limit", defaultLimit)

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
			CreatedAt:      time.Now(),
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
			"keys": []interface{}{},
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
