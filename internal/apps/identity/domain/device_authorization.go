// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
	"time"

	googleUuid "github.com/google/uuid"
)

// Device Authorization status values (RFC 8628).
const (
	DeviceAuthStatusPending    = "pending"    // User has not yet authorized.
	DeviceAuthStatusAuthorized = "authorized" // User authorized the device.
	DeviceAuthStatusDenied     = "denied"     // User denied authorization.
	DeviceAuthStatusUsed       = "used"       // Device code exchanged for tokens.
)

// DeviceAuthorization represents a pending device authorization request (RFC 8628).
// This domain model supports the OAuth 2.0 Device Authorization Grant flow for
// devices with limited input capabilities (smart TVs, IoT devices, CLI tools).
type DeviceAuthorization struct {
	// Primary key.
	ID googleUuid.UUID `gorm:"type:text;primaryKey" json:"id"`

	// Client information.
	ClientID string `gorm:"type:text;not null;index" json:"client_id"`

	// Device codes.
	DeviceCode string `gorm:"type:text;not null;uniqueIndex" json:"device_code"` // Used by device for polling.
	UserCode   string `gorm:"type:text;not null;uniqueIndex" json:"user_code"`   // Displayed to user for verification.

	// Request parameters.
	Scope string `gorm:"type:text" json:"scope"`

	// User information (populated after user authorizes on secondary device).
	UserID NullableUUID `gorm:"type:text;index" json:"user_id"`

	// Authorization status: pending, authorized, denied, used.
	Status string `gorm:"type:text;not null;index;default:'pending'" json:"status"`

	// Polling control (RFC 8628 Section 3.5 - rate limiting).
	LastPolledAt *time.Time `gorm:"index" json:"last_polled_at,omitempty"`

	// Request metadata.
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`

	// Token issuance metadata (populated when grant_type=device_code succeeds).
	UsedAt *time.Time `gorm:"index" json:"used_at,omitempty"`
}

// TableName returns the database table name for DeviceAuthorization.
func (DeviceAuthorization) TableName() string {
	return "device_authorizations"
}

// IsExpired checks if the device code has expired.
func (d *DeviceAuthorization) IsExpired() bool {
	return time.Now().UTC().After(d.ExpiresAt)
}

// IsPending checks if authorization is pending user action.
func (d *DeviceAuthorization) IsPending() bool {
	return d.Status == DeviceAuthStatusPending
}

// IsAuthorized checks if user has authorized the device.
func (d *DeviceAuthorization) IsAuthorized() bool {
	return d.Status == DeviceAuthStatusAuthorized
}

// IsDenied checks if user denied the authorization.
func (d *DeviceAuthorization) IsDenied() bool {
	return d.Status == DeviceAuthStatusDenied
}

// IsUsed checks if device code has been exchanged for tokens.
func (d *DeviceAuthorization) IsUsed() bool {
	return d.Status == DeviceAuthStatusUsed
}
