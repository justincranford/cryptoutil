// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

package model

import (
	"time"

	googleUuid "github.com/google/uuid"
)

// ResourceRecord is a minimal model placeholder for server/model package structure.
type ResourceRecord struct {
	ID         googleUuid.UUID `gorm:"type:text;primaryKey"`
	TenantID   googleUuid.UUID `gorm:"type:text;not null;index:idx_rs_resources_tenant"`
	ResourceID string          `gorm:"type:text;not null"`
	CreatedAt  time.Time       `gorm:"not null;autoCreateTime"`
}

// TableName returns the persistence table name.
func (ResourceRecord) TableName() string {
	return "identity_rs_resources"
}
