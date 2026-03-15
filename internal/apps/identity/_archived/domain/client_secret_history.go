// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
	"time"

	googleUuid "github.com/google/uuid"
)

// ClientSecretHistory represents a historical client secret for audit purposes.
type ClientSecretHistory struct {
	ID         googleUuid.UUID `gorm:"type:text;primaryKey" json:"id"`
	ClientID   googleUuid.UUID `gorm:"type:text;index;not null" json:"client_id"`
	SecretHash string          `gorm:"type:text;not null" json:"secret_hash"`
	RotatedAt  time.Time       `gorm:"index;not null;default:CURRENT_TIMESTAMP" json:"rotated_at"`
	RotatedBy  string          `gorm:"type:text" json:"rotated_by,omitempty"`
	Reason     string          `gorm:"type:text" json:"reason,omitempty"`
	ExpiresAt  *time.Time      `gorm:"type:timestamp" json:"expires_at,omitempty"`
	CreatedAt  time.Time       `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  time.Time       `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName specifies the table name for GORM.
func (ClientSecretHistory) TableName() string {
	return "client_secret_history"
}
