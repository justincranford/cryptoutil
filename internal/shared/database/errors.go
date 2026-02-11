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

// ErrShardNotFound indicates the requested shard does not exist.
ErrShardNotFound = errors.New("shard not found")

// ErrShardUnavailable indicates the shard is temporarily unavailable.
ErrShardUnavailable = errors.New("shard unavailable")

// ErrInvalidShardKey indicates the shard key is invalid.
ErrInvalidShardKey = errors.New("invalid shard key")

// ErrCrossShardOperation indicates an operation cannot span multiple shards.
ErrCrossShardOperation = errors.New("cross-shard operation not supported")

// ErrSchemaNotFound indicates the tenant schema does not exist.
ErrSchemaNotFound = errors.New("tenant schema not found")

// ErrSchemaAlreadyExists indicates the tenant schema already exists.
ErrSchemaAlreadyExists = errors.New("tenant schema already exists")
)
