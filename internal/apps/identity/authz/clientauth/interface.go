// Copyright (c) 2025 Justin Cranford
//
//

package clientauth

import (
	"context"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

// ClientAuthenticator defines the interface for client authentication methods.
type ClientAuthenticator interface {
	// Method returns the authentication method name.
	Method() string

	// Authenticate authenticates a client and returns the client domain object.
	Authenticate(ctx context.Context, clientID, credential string) (*cryptoutilIdentityDomain.Client, error)
}
