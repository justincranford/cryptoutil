// Copyright (c) 2025 Justin Cranford.
// SPDX-License-Identifier: Apache-2.0.

// Package database provides shared database utilities for multi-tenancy and sharding.
package database

import (
	"context"

	googleUuid "github.com/google/uuid"
)

// TenantContextKey is the context key for tenant ID.
type TenantContextKey struct{}

// TenantContext holds tenant isolation information for database operations.
type TenantContext struct {
	TenantID googleUuid.UUID
	RealmID  googleUuid.UUID // Optional realm within tenant.
	UserID   googleUuid.UUID // Optional user making the request.
}

// WithTenantContext adds tenant context to the given context.
func WithTenantContext(ctx context.Context, tc *TenantContext) context.Context {
	return context.WithValue(ctx, TenantContextKey{}, tc)
}

// GetTenantContext retrieves tenant context from the context.
// Returns nil if no tenant context is set.
func GetTenantContext(ctx context.Context) *TenantContext {
	tc, ok := ctx.Value(TenantContextKey{}).(*TenantContext)
	if !ok {
		return nil
	}
	return tc
}

// GetTenantID retrieves the tenant ID from context.
// Returns zero UUID if no tenant context is set.
func GetTenantID(ctx context.Context) googleUuid.UUID {
	tc := GetTenantContext(ctx)
	if tc == nil {
		return googleUuid.Nil
	}
	return tc.TenantID
}

// MustGetTenantID retrieves the tenant ID from context, panicking if not set.
// Use this only when tenant isolation is absolutely required.
func MustGetTenantID(ctx context.Context) googleUuid.UUID {
	tc := GetTenantContext(ctx)
	if tc == nil || tc.TenantID == googleUuid.Nil {
		panic("tenant ID not set in context")
	}
	return tc.TenantID
}

// RequireTenantContext validates that tenant context is present and returns error if not.
func RequireTenantContext(ctx context.Context) (*TenantContext, error) {
	tc := GetTenantContext(ctx)
	if tc == nil {
		return nil, ErrNoTenantContext
	}
	if tc.TenantID == googleUuid.Nil {
		return nil, ErrInvalidTenantID
	}
	return tc, nil
}
