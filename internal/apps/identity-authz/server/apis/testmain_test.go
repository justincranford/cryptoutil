// Copyright (c) 2025-2026 Justin Cranford.
//
//

package apis_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	googleUuid "github.com/google/uuid"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityMfa "cryptoutil/internal/apps/identity/mfa"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	totpTestConfig      *cryptoutilIdentityConfig.Config
	totpTestRepoFactory *cryptoutilIdentityRepository.RepositoryFactory
)

func TestMain(m *testing.M) {
	var err error

	totpTestConfig, totpTestRepoFactory, err = newTOTPIntegrationTestFixture()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize TOTP integration test fixture: %v\n", err)
		os.Exit(1)
	}

	exitCode := m.Run()

	if err := totpTestRepoFactory.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to close TOTP integration test fixture: %v\n", err)

		exitCode = 1
	}

	os.Exit(exitCode)
}

func newTOTPIntegrationTestFixture() (*cryptoutilIdentityConfig.Config, *cryptoutilIdentityRepository.RepositoryFactory, error) {
	databaseID, err := googleUuid.NewV7()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate TOTP test database ID: %w", err)
	}

	config := &cryptoutilIdentityConfig.Config{
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type: cryptoutilSharedMagic.TestDatabaseSQLite,
			DSN:  fmt.Sprintf(cryptoutilSharedMagic.SQLiteInMemoryNamedDSNFormat, databaseID.String()),
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer: cryptoutilSharedMagic.TestIssuer,
		},
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(context.Background(), config.Database)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create TOTP test repository factory: %w", err)
	}

	if err := repoFactory.DB().AutoMigrate(
		&cryptoutilIdentityDomain.User{},
		&cryptoutilIdentityMfa.TOTPSecret{},
		&cryptoutilIdentityMfa.BackupCode{},
	); err != nil {
		return nil, nil, errors.Join(fmt.Errorf("failed to auto-migrate TOTP test database: %w", err), repoFactory.Close())
	}

	return config, repoFactory, nil
}
