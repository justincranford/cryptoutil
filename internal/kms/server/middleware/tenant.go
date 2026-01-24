// Copyright (c) 2025 Justin Cranford
//
//

// Package middleware provides HTTP middleware for the KMS server.
package middleware

import (
	"context"
	"strings"

	fiber "github.com/gofiber/fiber/v2"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TenantContextKey is the context key for tenant ID.
type TenantContextKey struct{}

// TenantIDHeader is the HTTP header for tenant ID.
// Reference: Session 4 Q10 - tenant ID always via header, never path/query.
const TenantIDHeader = "X-Tenant-ID"

// TenantMiddleware extracts tenant ID from Authorization header.
// Reference: Session 4 Q10 - tenant ID always via Authorization header.
func TenantMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract tenant ID from dedicated header.
		tenantID := c.Get(TenantIDHeader)
		if tenantID != "" {
			tenantID = strings.TrimSpace(tenantID)

			// Validate UUID format.
			if !isValidUUID(tenantID) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   "invalid_tenant_id",
					"message": "Tenant ID must be a valid UUID",
				})
			}

			// Store tenant ID in context.
			ctx := context.WithValue(c.UserContext(), TenantContextKey{}, tenantID)
			c.SetUserContext(ctx)
		}

		return c.Next()
	}
}

// GetTenantID extracts tenant ID from request context.
func GetTenantID(ctx context.Context) string {
	if tenantID, ok := ctx.Value(TenantContextKey{}).(string); ok {
		return tenantID
	}

	return ""
}

// RequireTenantMiddleware ensures tenant ID is present.
func RequireTenantMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tenantID := GetTenantID(c.UserContext())
		if tenantID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "missing_tenant_id",
				"message": "X-Tenant-ID header is required",
			})
		}

		return c.Next()
	}
}

// isValidUUID validates UUID format (8-4-4-4-12 with hyphens).
func isValidUUID(s string) bool {
	if len(s) != cryptoutilSharedMagic.UUIDStringLength {
		return false
	}

	// Check hyphen positions: 8, 13, 18, 23.
	hyphenPositions := []int{8, 13, 18, 23}
	for _, pos := range hyphenPositions {
		if s[pos] != '-' {
			return false
		}
	}

	// Check all other characters are hex.
	for i, char := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			continue
		}

		if !isHexChar(byte(char)) {
			return false
		}
	}

	return true
}

// isHexChar checks if a byte is a valid hex character.
func isHexChar(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}
