// Copyright (c) 2025 Justin Cranford
//
//
// SPDX-License-Identifier: MIT

package model

import (
	"time"

	googleUuid "github.com/google/uuid"
)

// SessionRecord is a minimal model placeholder for server/model package structure.
type SessionRecord struct {
	ID        googleUuid.UUID `gorm:"type:text;primaryKey"`
	TenantID  googleUuid.UUID `gorm:"type:text;not null;index:idx_idp_sessions_tenant"`
	Subject   string          `gorm:"type:text;not null"`
	Status    string          `gorm:"type:text;not null"`
	CreatedAt time.Time       `gorm:"not null;autoCreateTime"`
}

// TableName returns the persistence table name.
func (SessionRecord) TableName() string {
	return "identity_idp_sessions"
}
