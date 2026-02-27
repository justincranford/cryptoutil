// Copyright (c) 2025 Justin Cranford
//

// Package domain provides domain models for the skeleton-template service.
// These models represent a minimal template item for demonstration purposes.
package domain

import (
"time"

googleUuid "github.com/google/uuid"
)

// TemplateItem represents a minimal domain entity for the skeleton-template service.
// This model demonstrates best-practice GORM tagging with cross-DB compatibility.
// CRITICAL: TenantID for data scoping only - realms are authentication-only, NOT data scope.
type TemplateItem struct {
ID        googleUuid.UUID `gorm:"type:text;primaryKey"`
TenantID  googleUuid.UUID `gorm:"type:text;not null;index:idx_template_items_tenant"`
CreatedAt time.Time       `gorm:"not null;autoCreateTime"`
}

// TableName specifies the database table name for TemplateItem.
func (TemplateItem) TableName() string {
return "template_items"
}
