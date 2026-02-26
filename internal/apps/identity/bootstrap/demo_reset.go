// Copyright (c) 2025 Justin Cranford
//
//

package bootstrap

import (
	"context"
	"errors"
	"fmt"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// ResetDemoData deletes all demo data from the repository.
// This includes the demo-client and demo-user created during bootstrap.
// Reference: Session 3 Q15 - data cleanup for demo mode.
func ResetDemoData(
	ctx context.Context,
	repoFactory *cryptoutilIdentityRepository.RepositoryFactory,
) error {
	// Delete demo client.
	if err := deleteDemoClient(ctx, repoFactory); err != nil {
		return fmt.Errorf("failed to delete demo client: %w", err)
	}

	// Delete demo user.
	if err := deleteDemoUser(ctx, repoFactory); err != nil {
		return fmt.Errorf("failed to delete demo user: %w", err)
	}

	fmt.Println("✅ Demo data reset complete")

	return nil
}

// deleteDemoClient removes the demo-client from the repository.
func deleteDemoClient(
	ctx context.Context,
	repoFactory *cryptoutilIdentityRepository.RepositoryFactory,
) error {
	clientRepo := repoFactory.ClientRepository()


	existingClient, err := clientRepo.GetByClientID(ctx, cryptoutilSharedMagic.DemoClientID)
	if err != nil {
		if errors.Is(err, cryptoutilIdentityAppErr.ErrClientNotFound) {
			fmt.Println("ℹ️  Demo client already deleted or never existed")

			return nil
		}

		return fmt.Errorf("failed to check for demo client: %w", err)
	}

	if err := clientRepo.Delete(ctx, existingClient.ID); err != nil {
		return fmt.Errorf("failed to delete demo client: %w", err)
	}

	fmt.Println("✅ Deleted demo client")

	return nil
}

// deleteDemoUser removes the demo-user from the repository.
func deleteDemoUser(
	ctx context.Context,
	repoFactory *cryptoutilIdentityRepository.RepositoryFactory,
) error {
	userRepo := repoFactory.UserRepository()

	const demoUserSub = "demo-user"

	existingUser, err := userRepo.GetBySub(ctx, demoUserSub)
	if err != nil {
		if errors.Is(err, cryptoutilIdentityAppErr.ErrUserNotFound) {
			fmt.Println("ℹ️  Demo user already deleted or never existed")

			return nil
		}

		return fmt.Errorf("failed to check for demo user: %w", err)
	}

	if err := userRepo.Delete(ctx, existingUser.ID); err != nil {
		return fmt.Errorf("failed to delete demo user: %w", err)
	}

	fmt.Println("✅ Deleted demo user")

	return nil
}

// ResetAndReseedDemo deletes all demo data and recreates it.
// This is useful for resetting to a clean demo state.
func ResetAndReseedDemo(
	ctx context.Context,
	repoFactory *cryptoutilIdentityRepository.RepositoryFactory,
) error {
	fmt.Println("🔄 Resetting demo data...")

	// Delete existing demo data.
	if err := ResetDemoData(ctx, repoFactory); err != nil {
		return fmt.Errorf("failed to reset demo data: %w", err)
	}

	// Reseed demo data.
	if err := BootstrapUsers(ctx, repoFactory); err != nil {
		return fmt.Errorf("failed to reseed demo users: %w", err)
	}

	// Note: BootstrapClients is typically called separately with config.
	// For now, we only reset and reseed users here.
	// Client reseeding requires config which should be done at a higher level.

	fmt.Println("✅ Demo data reset and reseeded successfully")

	return nil
}
