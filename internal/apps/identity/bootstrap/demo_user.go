// Copyright (c) 2025 Justin Cranford
//
//

package bootstrap

import (
	"context"
	"errors"
	"fmt"

	googleUuid "github.com/google/uuid"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedCryptoHash "cryptoutil/internal/shared/crypto/hash"
)

// CreateDemoUser creates the demo user for testing OAuth flows if it doesn't exist.
// Returns the user's sub claim and plaintext password (only on creation).
func CreateDemoUser(
	ctx context.Context,
	repoFactory *cryptoutilIdentityRepository.RepositoryFactory,
) (sub string, plainPassword string, created bool, err error) {
	userRepo := repoFactory.UserRepository()

	// Check if demo user already exists.
	const demoUserSub = "demo-user"

	existingUser, err := userRepo.GetBySub(ctx, demoUserSub)
	if err != nil && !errors.Is(err, cryptoutilIdentityAppErr.ErrUserNotFound) {
		return "", "", false, fmt.Errorf("failed to check for existing demo user: %w", err)
	}

	if existingUser != nil {
		// User exists, return without password.
		return demoUserSub, "", false, nil
	}

	// Generate demo user password hash.
	plainPassword = "demo-password"

	passwordHash, err := cryptoutilSharedCryptoHash.HashLowEntropyNonDeterministic(plainPassword)
	if err != nil {
		return "", "", false, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrPasswordHashFailed,
			fmt.Errorf("failed to hash demo user password: %w", err),
		)
	}

	// Create demo user with standard OIDC claims.
	demoUser := &cryptoutilIdentityDomain.User{
		ID:                googleUuid.Must(googleUuid.NewV7()),
		Sub:               demoUserSub,
		Name:              "Demo User",
		GivenName:         "Demo",
		FamilyName:        "User",
		PreferredUsername: "demo",
		Email:             "demo@example.com",
		EmailVerified:     true,
		PhoneNumber:       "+1-555-000-0000",
		PhoneVerified:     false,
		PasswordHash:      passwordHash,
		Enabled:           true,
		Locked:            false,
	}

	if err := userRepo.Create(ctx, demoUser); err != nil {
		return "", "", false, fmt.Errorf("failed to create demo user: %w", err)
	}

	return demoUserSub, plainPassword, true, nil
}

// BootstrapUsers creates all bootstrap users for the identity server.
func BootstrapUsers(
	ctx context.Context,
	repoFactory *cryptoutilIdentityRepository.RepositoryFactory,
) error {
	// Create demo user.
	sub, password, created, err := CreateDemoUser(ctx, repoFactory)
	if err != nil {
		return fmt.Errorf("failed to bootstrap demo user: %w", err)
	}

	if created {
		fmt.Printf("✅ Created bootstrap user: %s (password: %s)\n", sub, password)
		fmt.Printf("   Email: demo@example.com\n")
		fmt.Printf("   Username: demo\n")
		fmt.Printf("   ⚠️  SAVE THIS PASSWORD - it will not be shown again!\n")
	} else {
		fmt.Printf("ℹ️  Bootstrap user already exists: %s\n", sub)
	}

	return nil
}
