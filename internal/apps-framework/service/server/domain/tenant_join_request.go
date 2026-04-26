// Copyright (c) 2025 Justin Cranford.
// SPDX-License-Identifier: Apache-2.0.

// Package domain provides domain models and business logic for the service template.
package domain

import (
	"time"

	googleUuid "github.com/google/uuid"
)

// TenantJoinRequest represents a user or client requesting to join an existing tenant.
type TenantJoinRequest struct {
	ID          googleUuid.UUID  `gorm:"type:text;primaryKey"`
	UserID      *googleUuid.UUID `gorm:"type:text;index"` // Nullable - mutually exclusive with ClientID.
	ClientID    *googleUuid.UUID `gorm:"type:text;index"` // Nullable - mutually exclusive with UserID.
	TenantID    googleUuid.UUID  `gorm:"type:text;not null;index"`
	Status      string           `gorm:"type:text;not null;index"` // pending, approved, rejected.
	RequestedAt time.Time        `gorm:"not null;default:CURRENT_TIMESTAMP"`
	ProcessedAt *time.Time       `gorm:""`
	ProcessedBy *googleUuid.UUID `gorm:"type:text"` // User ID of admin who processed the request.
}

// TableName overrides the default table name.
func (TenantJoinRequest) TableName() string {
	return "tenant_join_requests"
}

// JoinRequestStatus represents the possible states of a join request.
const (
	JoinRequestStatusPending  = "pending"
	JoinRequestStatusApproved = "approved"
	JoinRequestStatusRejected = "rejected"
)
