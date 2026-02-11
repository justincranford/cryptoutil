// Copyright (c) 2025 Justin Cranford
//
//

package middleware

import (
	"context"
	"errors"
	"fmt"
	"strings"

	fiber "github.com/gofiber/fiber/v2"
)

// ScopeConfig configures scope validation behavior.
type ScopeConfig struct {
	// CoarseScopes are high-level scopes (e.g., kms:admin, kms:read, kms:write).
	CoarseScopes []string

	// FineScopes are operation-specific scopes (e.g., kms:encrypt, kms:decrypt, kms:sign).
	FineScopes []string

	// ScopeHierarchy defines scope inheritance (coarse scope -> fine scopes).
	ScopeHierarchy map[string][]string

	// ErrorDetailLevel controls error verbosity.
	ErrorDetailLevel string
}

// DefaultScopeConfig returns a default scope configuration for KMS.
func DefaultScopeConfig() ScopeConfig {
	return ScopeConfig{
		CoarseScopes: []string{
			"kms:admin",
			"kms:read",
			"kms:write",
		},
		FineScopes: []string{
			"kms:encrypt",
			"kms:decrypt",
			"kms:sign",
			"kms:verify",
			"kms:wrap",
			"kms:unwrap",
			"kms:generate",
			"kms:derive",
			"kms:pool:create",
			"kms:pool:read",
			"kms:pool:update",
			"kms:pool:delete",
			"kms:key:create",
			"kms:key:read",
			"kms:key:rotate",
			"kms:key:delete",
		},
		ScopeHierarchy: map[string][]string{
			// Admin includes all operations.
			"kms:admin": {
				"kms:read",
				"kms:write",
				"kms:encrypt",
				"kms:decrypt",
				"kms:sign",
				"kms:verify",
				"kms:wrap",
				"kms:unwrap",
				"kms:generate",
				"kms:derive",
				"kms:pool:create",
				"kms:pool:read",
				"kms:pool:update",
				"kms:pool:delete",
				"kms:key:create",
				"kms:key:read",
				"kms:key:rotate",
				"kms:key:delete",
			},
			// Read includes read-only operations.
			"kms:read": {
				"kms:pool:read",
				"kms:key:read",
				"kms:verify", // Verify is read-like.
			},
			// Write includes write operations.
			"kms:write": {
				"kms:encrypt",
				"kms:decrypt",
				"kms:sign",
				"kms:wrap",
				"kms:unwrap",
				"kms:generate",
				"kms:derive",
				"kms:pool:create",
				"kms:pool:update",
				"kms:pool:delete",
				"kms:key:create",
				"kms:key:rotate",
				"kms:key:delete",
			},
		},
		ErrorDetailLevel: errorDetailLevelMin,
	}
}

// ScopeValidator validates and enforces scopes.
type ScopeValidator struct {
	config ScopeConfig
}

// NewScopeValidator creates a new scope validator.
func NewScopeValidator(config ScopeConfig) *ScopeValidator {
	return &ScopeValidator{
		config: config,
	}
}

// ExpandScopes expands coarse scopes to include all implied fine scopes.
func (v *ScopeValidator) ExpandScopes(scopes []string) []string {
	expanded := make(map[string]bool)

	for _, scope := range scopes {
		expanded[scope] = true

		// Check if this scope has children in the hierarchy.
		if children, ok := v.config.ScopeHierarchy[scope]; ok {
			for _, child := range children {
				expanded[child] = true
			}
		}
	}

	// Convert map to slice.
	result := make([]string, 0, len(expanded))
	for scope := range expanded {
		result = append(result, scope)
	}

	return result
}

// HasScope checks if the scopes include a required scope (with hierarchy expansion).
func (v *ScopeValidator) HasScope(scopes []string, required string) bool {
	expanded := v.ExpandScopes(scopes)

	for _, scope := range expanded {
		if scope == required {
			return true
		}
	}

	return false
}

// HasAnyScope checks if scopes include any of the required scopes.
func (v *ScopeValidator) HasAnyScope(scopes []string, required []string) bool {
	for _, req := range required {
		if v.HasScope(scopes, req) {
			return true
		}
	}

	return false
}

// HasAllScopes checks if scopes include all required scopes.
func (v *ScopeValidator) HasAllScopes(scopes []string, required []string) bool {
	for _, req := range required {
		if !v.HasScope(scopes, req) {
			return false
		}
	}

	return true
}

// ParseScopeString parses a space-separated scope string into a slice.
func ParseScopeString(scopeString string) []string {
	if scopeString == "" {
		return []string{}
	}

	// Split by space (OAuth2 standard) or comma.
	scopes := strings.FieldsFunc(scopeString, func(r rune) bool {
		return r == ' ' || r == ','
	})

	// Remove empty strings.
	result := make([]string, 0, len(scopes))

	for _, s := range scopes {
		trimmed := strings.TrimSpace(s)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// ValidateScope validates a scope string against known scopes.
func (v *ScopeValidator) ValidateScope(scope string) error {
	// Check if it's a known coarse scope.
	for _, coarse := range v.config.CoarseScopes {
		if scope == coarse {
			return nil
		}
	}

	// Check if it's a known fine scope.
	for _, fine := range v.config.FineScopes {
		if scope == fine {
			return nil
		}
	}

	return fmt.Errorf("unknown scope: %s", scope)
}

// ValidateScopes validates multiple scopes.
func (v *ScopeValidator) ValidateScopes(scopes []string) error {
	var errs []string

	for _, scope := range scopes {
		if err := v.ValidateScope(scope); err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}

// ScopeContextKey is the context key for validated scopes.
type ScopeContextKey struct{}

// GetScopes extracts scopes from request context.
func GetScopes(ctx context.Context) []string {
	if scopes, ok := ctx.Value(ScopeContextKey{}).([]string); ok {
		return scopes
	}

	// Try to get from JWT claims.
	if claims, ok := ctx.Value(JWTContextKey{}).(*JWTClaims); ok {
		return claims.Scopes
	}

	// Try to get from service auth info.
	if info, ok := ctx.Value(ServiceAuthContextKey{}).(*ServiceAuthInfo); ok {
		return info.Scopes
	}

	return nil
}

// RequireScope middleware enforces a required scope.
func RequireScope(validator *ScopeValidator, requiredScope string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		scopes := GetScopes(c.UserContext())
		if scopes == nil {
			return insufficientScopeError(c, validator.config.ErrorDetailLevel, requiredScope, nil)
		}

		if !validator.HasScope(scopes, requiredScope) {
			return insufficientScopeError(c, validator.config.ErrorDetailLevel, requiredScope, scopes)
		}

		return c.Next()
	}
}

// RequireAnyScope middleware enforces at least one of the required scopes.
func RequireAnyScope(validator *ScopeValidator, requiredScopes ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		scopes := GetScopes(c.UserContext())
		if scopes == nil {
			return insufficientScopeError(c, validator.config.ErrorDetailLevel, strings.Join(requiredScopes, " OR "), nil)
		}

		if !validator.HasAnyScope(scopes, requiredScopes) {
			return insufficientScopeError(c, validator.config.ErrorDetailLevel, strings.Join(requiredScopes, " OR "), scopes)
		}

		return c.Next()
	}
}

// RequireAllScopes middleware enforces all required scopes.
func RequireAllScopes(validator *ScopeValidator, requiredScopes ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		scopes := GetScopes(c.UserContext())
		if scopes == nil {
			return insufficientScopeError(c, validator.config.ErrorDetailLevel, strings.Join(requiredScopes, " AND "), nil)
		}

		if !validator.HasAllScopes(scopes, requiredScopes) {
			return insufficientScopeError(c, validator.config.ErrorDetailLevel, strings.Join(requiredScopes, " AND "), scopes)
		}

		return c.Next()
	}
}

// insufficientScopeError returns a 403 error for insufficient scopes.
func insufficientScopeError(c *fiber.Ctx, detailLevel, required string, provided []string) error {
	response := fiber.Map{
		"error": "insufficient_scope",
	}

	if detailLevel == errorDetailLevelStd || detailLevel == errorDetailLevelDebug {
		response["error_description"] = "The request requires higher privileges than provided by the access token"
		response["required_scope"] = required
	}

	if detailLevel == errorDetailLevelDebug && provided != nil {
		response["provided_scopes"] = provided
	}

	if err := c.Status(fiber.StatusForbidden).JSON(response); err != nil {
		return fmt.Errorf("failed to send forbidden response: %w", err)
	}

	return nil
}
