// Copyright (c) 2025 Justin Cranford
//
//

// Package repository provides database repositories and migrations for the KMS server.
package repository

import (
	"embed"
)

// MigrationsFS contains the embedded KMS domain migrations (2001+).
// Template migrations (1001-1004) are provided by ServerBuilder.
// These are merged by ServerBuilder.WithDomainMigrations() to create a unified migration stream.
//
//go:embed migrations/*.sql
var MigrationsFS embed.FS
