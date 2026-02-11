// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"

	googleUuid "github.com/google/uuid"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

// DeviceAuthorizationRepository manages device authorization requests (RFC 8628).
type DeviceAuthorizationRepository interface {
	// Create stores a new device authorization request.
	Create(ctx context.Context, auth *cryptoutilIdentityDomain.DeviceAuthorization) error

	// GetByDeviceCode retrieves an authorization by device code.
	// Returns error if device code not found or expired.
	GetByDeviceCode(ctx context.Context, deviceCode string) (*cryptoutilIdentityDomain.DeviceAuthorization, error)

	// GetByUserCode retrieves an authorization by user code.
	// Used when user visits verification URI and enters user code.
	GetByUserCode(ctx context.Context, userCode string) (*cryptoutilIdentityDomain.DeviceAuthorization, error)

	// GetByID retrieves an authorization by primary key UUID.
	GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.DeviceAuthorization, error)

	// Update modifies an existing device authorization (e.g., status changes, user ID).
	Update(ctx context.Context, auth *cryptoutilIdentityDomain.DeviceAuthorization) error

	// DeleteExpired removes device authorizations past their expiration time.
	// Should be called periodically (e.g., hourly cleanup job).
	DeleteExpired(ctx context.Context) error
}
