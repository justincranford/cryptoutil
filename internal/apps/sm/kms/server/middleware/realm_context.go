// Copyright (c) 2025 Justin Cranford
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package middleware provides HTTP middleware for the KMS server.
package middleware

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
)

// RealmContextKey is the context key for realm context.
type RealmContextKey struct{}

// RealmContext holds tenant and realm information extracted from authentication.
type RealmContext struct {
	TenantID googleUuid.UUID
	RealmID  googleUuid.UUID
	UserID   googleUuid.UUID
	ClientID googleUuid.UUID
	Scopes   []string
	Source   string // "jwt", "oidc", "header", "mtls"
}

// RealmContextMiddleware extracts realm context from authenticated requests.
// It should run AFTER JWT middleware or session middleware has validated credentials.
//
// Priority order for tenant extraction:
// 1. JWT claims (from JWTMiddleware) - custom claims tenant_id, realm_id
// 2. OIDC claims (from OIDCMiddleware) - OIDCClaims.TenantID
// 3. X-Tenant-ID header (from TenantMiddleware) - GetTenantID()
//
// This middleware provides backward compatibility by also setting TenantContextKey.
func RealmContextMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		realmCtx := &RealmContext{}

		// Try JWT claims first (highest priority for service-to-service)
		if jwtClaims := GetJWTClaims(c.UserContext()); jwtClaims != nil {
			realmCtx.Source = "jwt"
			realmCtx.Scopes = jwtClaims.Scopes

			// Extract tenant_id from custom claims
			if jwtClaims.Custom != nil {
				if tenantIDStr, ok := jwtClaims.Custom["tenant_id"].(string); ok {
					if tid, err := googleUuid.Parse(tenantIDStr); err == nil {
						realmCtx.TenantID = tid
					}
				}

				if realmIDStr, ok := jwtClaims.Custom["realm_id"].(string); ok {
					if rid, err := googleUuid.Parse(realmIDStr); err == nil {
						realmCtx.RealmID = rid
					}
				}

				if userIDStr, ok := jwtClaims.Custom["user_id"].(string); ok {
					if uid, err := googleUuid.Parse(userIDStr); err == nil {
						realmCtx.UserID = uid
					}
				}

				if clientIDStr, ok := jwtClaims.Custom[cryptoutilSharedMagic.ClaimClientID].(string); ok {
					if cid, err := googleUuid.Parse(clientIDStr); err == nil {
						realmCtx.ClientID = cid
					}
				}
			}

			// Use subject as fallback for user/client ID
			if realmCtx.UserID == googleUuid.Nil && jwtClaims.Subject != "" {
				if uid, err := googleUuid.Parse(jwtClaims.Subject); err == nil {
					realmCtx.UserID = uid
				}
			}
		}

		// Try OIDC claims if no tenant from JWT
		if realmCtx.TenantID == googleUuid.Nil {
			if oidcClaims := GetOIDCClaims(c.UserContext()); oidcClaims != nil {
				realmCtx.Source = "oidc"

				if oidcClaims.TenantID != "" {
					if tid, err := googleUuid.Parse(oidcClaims.TenantID); err == nil {
						realmCtx.TenantID = tid
					}
				}
				// OIDC may have multiple tenants, use first if single tenant needed
				if realmCtx.TenantID == googleUuid.Nil && len(oidcClaims.TenantIDs) > 0 {
					if tid, err := googleUuid.Parse(oidcClaims.TenantIDs[0]); err == nil {
						realmCtx.TenantID = tid
					}
				}
			}
		}

		// Try X-Tenant-ID header as last resort
		if realmCtx.TenantID == googleUuid.Nil {
			if tenantIDStr := GetTenantID(c.UserContext()); tenantIDStr != "" {
				if tid, err := googleUuid.Parse(tenantIDStr); err == nil {
					realmCtx.TenantID = tid
					realmCtx.Source = "header"
				}
			}
		}

		// Store realm context
		ctx := context.WithValue(c.UserContext(), RealmContextKey{}, realmCtx)

		// Backward compatibility: also store tenant ID in TenantContextKey
		if realmCtx.TenantID != googleUuid.Nil {
			ctx = context.WithValue(ctx, TenantContextKey{}, realmCtx.TenantID)
		}

		c.SetUserContext(ctx)

		return c.Next()
	}
}

// GetRealmContext retrieves the realm context from the request context.
func GetRealmContext(ctx context.Context) *RealmContext {
	if ctx == nil {
		return nil
	}

	if rc, ok := ctx.Value(RealmContextKey{}).(*RealmContext); ok {
		return rc
	}

	return nil
}

// RequireRealmMiddleware ensures that a valid tenant context exists.
// Returns 401 Unauthorized if no tenant is available.
func RequireRealmMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		realmCtx := GetRealmContext(c.UserContext())
		if realmCtx == nil || realmCtx.TenantID == googleUuid.Nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError:   "unauthorized",
				"message": "tenant context required",
			})
		}

		return c.Next()
	}
}
