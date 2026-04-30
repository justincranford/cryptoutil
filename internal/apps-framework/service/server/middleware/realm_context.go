// Copyright (c) 2025-2026 Justin Cranford.
//
//

package middleware

import (
	"context"

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
	Source   string // "session", "jwt", "oidc", "header", "mtls"
}

// GetRealmContext extracts the RealmContext from a Go context.
// Returns nil if no realm context is set.
func GetRealmContext(ctx context.Context) *RealmContext {
	rc, _ := ctx.Value(RealmContextKey{}).(*RealmContext)

	return rc
}
