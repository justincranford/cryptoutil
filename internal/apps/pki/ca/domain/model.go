// Copyright (c) 2025 Justin Cranford

// Package domain provides domain models for the pki-ca service.
// These models represent a minimal CA item for demonstration purposes.
package domain

import (
	"time"

	googleUuid "github.com/google/uuid"
)

// CAItem represents a minimal domain entity for the pki-ca service.
// This model demonstrates best-practice GORM tagging with cross-DB compatibility.
// CRITICAL: TenantID for data scoping only - realms are authentication-only, NOT data scope.
type CAItem struct {
	ID        googleUuid.UUID `gorm:"type:text;primaryKey"`
	TenantID  googleUuid.UUID `gorm:"type:text;not null;index:idx_ca_items_tenant"`
	CreatedAt time.Time       `gorm:"not null;autoCreateTime"`
}

// TableName specifies the database table name for CAItem.
func (CAItem) TableName() string {
	return "ca_items"
}
