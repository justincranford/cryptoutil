// Copyright (c) 2025 Justin Cranford
//
//
// SPDX-License-Identifier: MIT

package model

import (
	"time"

	googleUuid "github.com/google/uuid"
)

// BrowserSession is a minimal model placeholder for server/model package structure.
type BrowserSession struct {
	ID        googleUuid.UUID `gorm:"type:text;primaryKey"`
	TenantID  googleUuid.UUID `gorm:"type:text;not null;index:idx_spa_sessions_tenant"`
	Subject   string          `gorm:"type:text;not null"`
	CreatedAt time.Time       `gorm:"not null;autoCreateTime"`
}

// TableName returns the persistence table name.
func (BrowserSession) TableName() string {
	return "identity_spa_sessions"
}
