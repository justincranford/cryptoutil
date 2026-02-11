// Copyright (c) 2025 Justin Cranford
//
//

package issuer

import (
	"context"
	"fmt"

	googleUuid "github.com/google/uuid"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
)

// UUIDIssuer issues opaque UUID-based tokens.
type UUIDIssuer struct{}

// NewUUIDIssuer creates a new UUID token issuer.
func NewUUIDIssuer() *UUIDIssuer {
	return &UUIDIssuer{}
}

// IssueToken issues a new opaque UUID token.
func (i *UUIDIssuer) IssueToken(_ context.Context) (string, error) {
	token := googleUuid.NewString()
	if token == "" {
		return "", cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrTokenIssuanceFailed,
			fmt.Errorf("failed to generate UUID token"),
		)
	}

	return token, nil
}

// ValidateToken validates a UUID token format (basic check).
func (i *UUIDIssuer) ValidateToken(_ context.Context, token string) error {
	if _, err := googleUuid.Parse(token); err != nil {
		return cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrInvalidToken,
			fmt.Errorf("invalid UUID format: %w", err),
		)
	}

	return nil
}
