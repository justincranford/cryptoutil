// Copyright (c) 2025-2026 Justin Cranford.
//
// SPDX-License-Identifier: AGPL-3.0-only
package model

import (
	"time"

	googleUuid "github.com/google/uuid"
)

// AuthorizationDecision is a minimal model placeholder for server/model package structure.
type AuthorizationDecision struct {
	ID        googleUuid.UUID `gorm:"type:text;primaryKey"`
	TenantID  googleUuid.UUID `gorm:"type:text;not null;index:idx_authz_decisions_tenant"`
	Subject   string          `gorm:"type:text;not null"`
	Decision  string          `gorm:"type:text;not null"`
	CreatedAt time.Time       `gorm:"not null;autoCreateTime"`
}

// TableName returns the persistence table name.
func (AuthorizationDecision) TableName() string {
	return "authorization_decisions"
}
