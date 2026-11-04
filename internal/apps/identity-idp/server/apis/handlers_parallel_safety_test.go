// Copyright (c) 2025-2026 Justin Cranford.
package apis

import (
	"context"
	"sync"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm/logger"

	cryptoutilTestHelpDB "cryptoutil/internal/apps-framework/service/test_help_db"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityORM "cryptoutil/internal/apps/identity/repository/orm"
)

// TestParallelTestSafety validates integration tests run safely in parallel without race
// conditions or database conflicts. Uses GORM's AutoMigrate with in-memory SQLite for
// faster test execution. Satisfies R03-05: Integration tests run in parallel safely.
func TestParallelTestSafety(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		parallelOps    int
		entityType     string
		validateUnique bool
	}{
		{
			name:           "parallel_user_creation",
			parallelOps:    cryptoutilSharedMagic.MaxErrorDisplay,
			entityType:     testUserValue04F8996D,
			validateUnique: true,
		},
		{
			name:           "parallel_client_creation",
			parallelOps:    cryptoutilSharedMagic.MaxErrorDisplay,
			entityType:     testClientValue948FE603,
			validateUnique: true,
		},
		{
			name:           "parallel_mixed_operations",
			parallelOps:    cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days,
			entityType:     testMixedOps,
			validateUnique: true,
		},
	}

	for _, tc := range tests {
		// Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			db := cryptoutilTestHelpDB.NewInMemorySQLiteDB(t)
			db.Logger = logger.Default.LogMode(logger.Silent)

			// Auto-migrate test schemas.
			require.NoError(t, db.AutoMigrate(
				&cryptoutilIdentityDomain.User{},
				&cryptoutilIdentityDomain.Client{},
				&cryptoutilIdentityDomain.ClientSecretVersion{}, // Required for client creation
				&cryptoutilIdentityDomain.KeyRotationEvent{},    // Required for secret rotation audit
				&cryptoutilIdentityDomain.Token{},
			))

			// Run parallel operations.
			var wg sync.WaitGroup

			errors := make(chan error, tc.parallelOps)
			createdIDs := make(chan string, tc.parallelOps)

			switch tc.entityType {
			case testUserValue04F8996D:
				userRepo := cryptoutilIdentityORM.NewUserRepository(db)

				for i := 0; i < tc.parallelOps; i++ {
					wg.Add(1)

					go func(_ int) {
						defer wg.Done()

						uniqueID := googleUuid.Must(googleUuid.NewV7()).String()
						user := &cryptoutilIdentityDomain.User{
							ID:                googleUuid.Must(googleUuid.NewV7()),
							Sub:               uniqueID,
							PreferredUsername: "testuser_parallel_" + uniqueID,
							Email:             "parallel_" + uniqueID + "@example.com",
							EmailVerified:     false,
							PasswordHash:      testHashedpassword,
						}

						if err := userRepo.Create(ctx, user); err != nil {
							errors <- err

							return
						}

						createdIDs <- user.Sub
					}(i)
				}

			case testClientValue948FE603:
				clientRepo := cryptoutilIdentityORM.NewClientRepository(db)

				for i := 0; i < tc.parallelOps; i++ {
					wg.Add(1)

					go func(_ int) {
						defer wg.Done()

						uniqueID := googleUuid.Must(googleUuid.NewV7()).String()
						client := &cryptoutilIdentityDomain.Client{
							ID:           googleUuid.Must(googleUuid.NewV7()),
							ClientID:     "client_parallel_" + uniqueID,
							ClientSecret: "secret_" + uniqueID,
							Name:         "Parallel Test Client " + uniqueID,
							RedirectURIs: []string{cryptoutilSharedMagic.TestRedirectURI},
						}

						if err := clientRepo.Create(ctx, client); err != nil {
							errors <- err

							return
						}

						createdIDs <- client.ClientID
					}(i)
				}

			case testMixedOps:
				userRepo := cryptoutilIdentityORM.NewUserRepository(db)
				clientRepo := cryptoutilIdentityORM.NewClientRepository(db)

				for i := 0; i < tc.parallelOps; i++ {
					wg.Add(1)

					go func(index int) {
						defer wg.Done()

						uniqueID := googleUuid.Must(googleUuid.NewV7()).String()

						// Alternate between creating users and clients.
						if index%2 == 0 {
							user := &cryptoutilIdentityDomain.User{
								ID:                googleUuid.Must(googleUuid.NewV7()),
								Sub:               uniqueID,
								PreferredUsername: "testuser_mixed_" + uniqueID,
								Email:             "mixed_" + uniqueID + "@example.com",
								EmailVerified:     false,
								PasswordHash:      testHashedpassword,
							}

							if err := userRepo.Create(ctx, user); err != nil {
								errors <- err

								return
							}

							createdIDs <- user.Sub
						} else {
							client := &cryptoutilIdentityDomain.Client{
								ID:           googleUuid.Must(googleUuid.NewV7()),
								ClientID:     "client_mixed_" + uniqueID,
								ClientSecret: "secret_" + uniqueID,
								Name:         "Mixed Test Client " + uniqueID,
								RedirectURIs: []string{cryptoutilSharedMagic.TestRedirectURI},
							}

							if err := clientRepo.Create(ctx, client); err != nil {
								errors <- err

								return
							}

							createdIDs <- client.ClientID
						}
					}(i)
				}
			}

			// Wait for all goroutines to complete.
			wg.Wait()
			close(errors)
			close(createdIDs)

			// Check for errors during parallel operations.
			var errList []error
			for err := range errors {
				errList = append(errList, err)
			}

			require.Empty(t, errList, "parallel operations should not produce errors: %v", errList)

			// Validate unique IDs created (no duplicates).
			if tc.validateUnique {
				uniqueMap := make(map[string]bool)
				duplicates := 0

				for id := range createdIDs {
					if uniqueMap[id] {
						duplicates++
					}

					uniqueMap[id] = true
				}

				require.Equal(t, 0, duplicates, "parallel operations should create unique entities")
				require.Equal(t, tc.parallelOps, len(uniqueMap), "should create expected number of entities")
			}
		})
	}
}
