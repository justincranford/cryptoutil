// Copyright (c) 2025 Justin Cranford
//
//
// SPDX-License-Identifier: MIT

package model

import (
	"time"

	googleUuid "github.com/google/uuid"
)

// ClientSession is a minimal model placeholder for server/model package structure.
type ClientSession struct {
	ID        googleUuid.UUID `gorm:"type:text;primaryKey"`
	TenantID  googleUuid.UUID `gorm:"type:text;not null;index:idx_rp_sessions_tenant"`
	Subject   string          `gorm:"type:text;not null"`
	CreatedAt time.Time       `gorm:"not null;autoCreateTime"`
}

// TableName returns the persistence table name.
func (ClientSession) TableName() string {
	return "identity_rp_sessions"
}
