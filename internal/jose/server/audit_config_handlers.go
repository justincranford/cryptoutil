// Copyright (c) 2025 Justin Cranford
//

// Package server provides the JOSE Authority Server HTTP service.
package server

import (
	cryptoutilJoseService "cryptoutil/internal/jose/service"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
)

// AuditConfigRequest represents the request body for setting audit config.
type AuditConfigRequest struct {
	Operation    string  `json:"operation"`
	Enabled      bool    `json:"enabled"`
	SamplingRate float64 `json:"sampling_rate"`
}

// AuditConfigResponse represents the response for audit config.
type AuditConfigResponse struct {
	TenantID     string  `json:"tenant_id"`
	Operation    string  `json:"operation"`
	Enabled      bool    `json:"enabled"`
	SamplingRate float64 `json:"sampling_rate"`
}

// AuditConfigListResponse represents the response for listing all audit configs.
type AuditConfigListResponse struct {
	Configs []AuditConfigResponse `json:"configs"`
}

// auditConfigHandlers provides audit config route handlers.
type auditConfigHandlers struct {
	auditConfigService *cryptoutilJoseService.AuditConfigService
}

// newAuditConfigHandlers creates a new audit config handlers instance.
func newAuditConfigHandlers(auditConfigService *cryptoutilJoseService.AuditConfigService) *auditConfigHandlers {
	return &auditConfigHandlers{
		auditConfigService: auditConfigService,
	}
}

// handleGetAuditConfig handles GET /browser/api/v1/admin/audit-config.
// Returns all audit configurations for the tenant.
func (h *auditConfigHandlers) handleGetAuditConfig(c *fiber.Ctx) error {
	// TODO: Get tenant ID from authentication context.
	// For now, use a default tenant ID.
	tenantID := googleUuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479")

	// Get all configs for the tenant.
	configs, err := h.auditConfigService.GetAllConfigs(c.Context(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get audit configurations",
		})
	}

	// Convert to response format.
	response := AuditConfigListResponse{
		Configs: make([]AuditConfigResponse, 0, len(configs)),
	}

	for _, cfg := range configs {
		response.Configs = append(response.Configs, AuditConfigResponse{
			TenantID:     cfg.TenantID.String(),
			Operation:    cfg.Operation,
			Enabled:      cfg.Enabled,
			SamplingRate: cfg.SamplingRate,
		})
	}

	return c.JSON(response)
}

// handleGetAuditConfigByOperation handles GET /browser/api/v1/admin/audit-config/:operation.
// Returns audit configuration for a specific operation.
func (h *auditConfigHandlers) handleGetAuditConfigByOperation(c *fiber.Ctx) error {
	operation := c.Params("operation")
	if operation == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Operation is required",
		})
	}

	// TODO: Get tenant ID from authentication context.
	tenantID := googleUuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479")

	// Get config for the operation.
	config, err := h.auditConfigService.GetConfig(c.Context(), tenantID, operation)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(AuditConfigResponse{
		TenantID:     config.TenantID.String(),
		Operation:    config.Operation,
		Enabled:      config.Enabled,
		SamplingRate: config.SamplingRate,
	})
}

// handleSetAuditConfig handles PUT /browser/api/v1/admin/audit-config.
// Sets audit configuration for an operation.
func (h *auditConfigHandlers) handleSetAuditConfig(c *fiber.Ctx) error {
	var req AuditConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate operation.
	if req.Operation == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Operation is required",
		})
	}

	// Validate sampling rate.
	if req.SamplingRate < 0 || req.SamplingRate > 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Sampling rate must be between 0.0 and 1.0",
		})
	}

	// TODO: Get tenant ID from authentication context.
	tenantID := googleUuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479")

	// Set the config.
	err := h.auditConfigService.SetConfig(c.Context(), tenantID, req.Operation, req.Enabled, req.SamplingRate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Return the updated config.
	return c.JSON(AuditConfigResponse{
		TenantID:     tenantID.String(),
		Operation:    req.Operation,
		Enabled:      req.Enabled,
		SamplingRate: req.SamplingRate,
	})
}
