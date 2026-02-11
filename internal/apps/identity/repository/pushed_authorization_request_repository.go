// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"

	googleUuid "github.com/google/uuid"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

// PushedAuthorizationRequestRepository manages pushed authorization requests (RFC 9126).
// PAR allows OAuth clients to push authorization request parameters directly to the
// authorization server, providing request integrity, confidentiality, and protection
// against parameter tampering and phishing attacks.
type PushedAuthorizationRequestRepository interface {
	// Create stores a new pushed authorization request.
	Create(ctx context.Context, req *cryptoutilIdentityDomain.PushedAuthorizationRequest) error

	// GetByRequestURI retrieves a PAR by its request_uri value.
	// Returns ErrPushedAuthorizationRequestNotFound if not found.
	GetByRequestURI(ctx context.Context, requestURI string) (*cryptoutilIdentityDomain.PushedAuthorizationRequest, error)

	// GetByID retrieves a PAR by its primary key ID.
	// Returns ErrPushedAuthorizationRequestNotFound if not found.
	GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.PushedAuthorizationRequest, error)

	// Update modifies an existing PAR (typically to mark as used).
	Update(ctx context.Context, req *cryptoutilIdentityDomain.PushedAuthorizationRequest) error

	// DeleteExpired removes all expired PAR entries from the database.
	// Returns the number of deleted entries.
	DeleteExpired(ctx context.Context) (int64, error)
}
