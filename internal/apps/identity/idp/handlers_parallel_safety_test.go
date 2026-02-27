// Copyright (c) 2025 Justin Cranford

package idp

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"database/sql"
	"fmt"
	"sync"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "modernc.org/sqlite" // Register CGO-free SQLite driver

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
			entityType:     "user",
			validateUnique: true,
		},
		{
			name:           "parallel_client_creation",
			parallelOps:    cryptoutilSharedMagic.MaxErrorDisplay,
			entityType:     "client",
			validateUnique: true,
		},
		{
			name:           "parallel_mixed_operations",
			parallelOps:    cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days,
			entityType:     "mixed",
			validateUnique: true,
		},
	}

	for _, tc := range tests {
		// Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			// Create unique in-memory database per test using UUIDv7.
			dbID, err := googleUuid.NewV7()
			require.NoError(t, err)

			dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", dbID.String())

			// Open database connection using modernc.org/sqlite (CGO-free).
			sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dsn)
			require.NoError(t, err)

			// Apply SQLite PRAGMA settings for WAL mode and busy timeout.
			if _, err := sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;"); err != nil {
				require.FailNowf(t, "failed to enable WAL mode", "%v", err)
			}

			if _, err := sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;"); err != nil {
				require.FailNowf(t, "failed to set busy timeout", "%v", err)
			}

			// Create GORM database with explicit connection.
			dialector := sqlite.Dialector{Conn: sqlDB}

			db, err := gorm.Open(dialector, &gorm.Config{
				Logger:                 logger.Default.LogMode(logger.Silent),
				SkipDefaultTransaction: true, // Disable automatic transactions.
			})
			require.NoError(t, err)

			// Get underlying sql.DB for connection pool configuration.
			gormDB, err := db.DB()
			require.NoError(t, err)

			// Configure connection pool for GORM transaction pattern.
			gormDB.SetMaxOpenConns(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries) // Allows transaction + operations concurrently.
			gormDB.SetMaxIdleConns(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)
			gormDB.SetConnMaxLifetime(0) // In-memory DB: never close connections.
			gormDB.SetConnMaxIdleTime(0)

			// Auto-migrate test schemas.
			err = db.AutoMigrate(
				&cryptoutilIdentityDomain.User{},
				&cryptoutilIdentityDomain.Client{},
				&cryptoutilIdentityDomain.ClientSecretVersion{}, // Required for client creation
				&cryptoutilIdentityDomain.KeyRotationEvent{},    // Required for secret rotation audit
				&cryptoutilIdentityDomain.Token{},
			)
			require.NoError(t, err) // Cleanup function to close database.
			t.Cleanup(func() {
				_ = sqlDB.Close() //nolint:errcheck // Test cleanup - error not critical for test teardown
			})

			// Run parallel operations.
			var wg sync.WaitGroup

			errors := make(chan error, tc.parallelOps)
			createdIDs := make(chan string, tc.parallelOps)

			switch tc.entityType {
			case "user":
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
							PasswordHash:      "hashedpassword",
						}

						if err := userRepo.Create(ctx, user); err != nil {
							errors <- err

							return
						}

						createdIDs <- user.Sub
					}(i)
				}

			case "client":
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
							RedirectURIs: []string{cryptoutilSharedMagic.DemoRedirectURI},
						}

						if err := clientRepo.Create(ctx, client); err != nil {
							errors <- err

							return
						}

						createdIDs <- client.ClientID
					}(i)
				}

			case "mixed":
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
								PasswordHash:      "hashedpassword",
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
								RedirectURIs: []string{cryptoutilSharedMagic.DemoRedirectURI},
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
