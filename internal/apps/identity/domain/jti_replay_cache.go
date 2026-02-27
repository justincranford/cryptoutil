// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
	"time"

	googleUuid "github.com/google/uuid"
)

// JTIReplayCache represents a cached JWT ID (jti) to prevent token replay attacks.
// JTI values are stored with expiration times to allow cleanup of old entries.
type JTIReplayCache struct {
	JTI       string          `gorm:"type:text;primaryKey" json:"jti"`     // Unique JWT identifier from jti claim.
	ClientID  googleUuid.UUID `gorm:"type:text;not null" json:"client_id"` // Client that used this jti.
	ExpiresAt time.Time       `gorm:"not null;index" json:"expires_at"`    // Expiration time from JWT exp claim.
	CreatedAt time.Time       `gorm:"autoCreateTime" json:"created_at"`    // When jti was first seen.
}

// TableName specifies the database table name.
func (JTIReplayCache) TableName() string {
	return "jti_replay_cache"
}
