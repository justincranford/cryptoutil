// Copyright (c) 2025-2026 Justin Cranford.
// SPDX-License-Identifier: AGPL-3.0-only
package database

import "errors"

// Database package errors for multi-tenancy and sharding.
var (
	// ErrNoTenantContext indicates the tenant context is not set in the context.
	ErrNoTenantContext = errors.New("tenant context not set")

	// ErrInvalidTenantID indicates the tenant ID is invalid (zero UUID).
	ErrInvalidTenantID = errors.New("invalid tenant ID")
)
