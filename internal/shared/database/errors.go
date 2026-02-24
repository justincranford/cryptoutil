// Copyright (c) 2025 Justin Cranford.
// SPDX-License-Identifier: Apache-2.0.

package database

import "errors"

// Database package errors for multi-tenancy and sharding.
var (
	// ErrNoTenantContext indicates the tenant context is not set in the context.
	ErrNoTenantContext = errors.New("tenant context not set")

	// ErrInvalidTenantID indicates the tenant ID is invalid (zero UUID).
	ErrInvalidTenantID = errors.New("invalid tenant ID")
)
