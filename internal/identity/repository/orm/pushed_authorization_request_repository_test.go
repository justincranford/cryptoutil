// Copyright (c) 2025 Justin Cranford
//
//

package orm_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityORM "cryptoutil/internal/identity/repository/orm"
)

// TestPushedAuthorizationRequestRepository_Create tests all Create paths.
func TestPushedAuthorizationRequestRepository_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupPAR  func() *cryptoutilIdentityDomain.PushedAuthorizationRequest
		expectErr bool
	}{
		{
			name: "SuccessfulCreate",
			setupPAR: func() *cryptoutilIdentityDomain.PushedAuthorizationRequest {
				return &cryptoutilIdentityDomain.PushedAuthorizationRequest{
					ID:                  googleUuid.Must(googleUuid.NewV7()),
					RequestURI:          "urn:ietf:params:oauth:request_uri:abc123",
					ClientID:            googleUuid.Must(googleUuid.NewV7()),
					ResponseType:        "code",
					RedirectURI:         "https://example.com/callback",
					Scope:               "openid profile",
					State:               "state123",
					CodeChallenge:       "challenge",
					CodeChallengeMethod: "S256",
					ExpiresAt:           time.Now().Add(90 * time.Second),
					CreatedAt:           time.Now(),
				}
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := setupPARTestDB(t)
			repo := cryptoutilIdentityORM.NewPushedAuthorizationRequestRepository(db)

			par := tt.setupPAR()
			err := repo.Create(ctx, par)

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestPushedAuthorizationRequestRepository_GetByRequestURI tests all GetByRequestURI paths.
func TestPushedAuthorizationRequestRepository_GetByRequestURI(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupPARs   func() []*cryptoutilIdentityDomain.PushedAuthorizationRequest
		queryURI    string
		expectErr   error
		expectFound bool
	}{
		{
			name: "Found",
			setupPARs: func() []*cryptoutilIdentityDomain.PushedAuthorizationRequest {
				return []*cryptoutilIdentityDomain.PushedAuthorizationRequest{
					{
						ID:                  googleUuid.Must(googleUuid.NewV7()),
						RequestURI:          "urn:ietf:params:oauth:request_uri:found",
						ClientID:            googleUuid.Must(googleUuid.NewV7()),
						ResponseType:        "code",
						RedirectURI:         "https://example.com/callback",
						CodeChallenge:       "challenge",
						CodeChallengeMethod: "S256",
						ExpiresAt:           time.Now().Add(90 * time.Second),
						CreatedAt:           time.Now(),
					},
				}
			},
			queryURI:    "urn:ietf:params:oauth:request_uri:found",
			expectErr:   nil,
			expectFound: true,
		},
		{
			name: "NotFound",
			setupPARs: func() []*cryptoutilIdentityDomain.PushedAuthorizationRequest {
				return []*cryptoutilIdentityDomain.PushedAuthorizationRequest{
					{
						ID:                  googleUuid.Must(googleUuid.NewV7()),
						RequestURI:          "urn:ietf:params:oauth:request_uri:different",
						ClientID:            googleUuid.Must(googleUuid.NewV7()),
						ResponseType:        "code",
						RedirectURI:         "https://example.com/callback",
						CodeChallenge:       "challenge",
						CodeChallengeMethod: "S256",
						ExpiresAt:           time.Now().Add(90 * time.Second),
						CreatedAt:           time.Now(),
					},
				}
			},
			queryURI:    "urn:ietf:params:oauth:request_uri:notfound",
			expectErr:   cryptoutilIdentityAppErr.ErrPushedAuthorizationRequestNotFound,
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := setupPARTestDB(t)
			repo := cryptoutilIdentityORM.NewPushedAuthorizationRequestRepository(db)

			pars := tt.setupPARs()
			for _, par := range pars {
				require.NoError(t, repo.Create(ctx, par))
			}

			result, err := repo.GetByRequestURI(ctx, tt.queryURI)

			if tt.expectErr != nil {
				require.ErrorIs(t, err, tt.expectErr)
			} else {
				require.NoError(t, err)
			}

			if tt.expectFound {
				require.NotNil(t, result)
				require.Equal(t, tt.queryURI, result.RequestURI)
			} else {
				require.Nil(t, result)
			}
		})
	}
}

// TestPushedAuthorizationRequestRepository_GetByID tests all GetByID paths.
func TestPushedAuthorizationRequestRepository_GetByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupPAR  func() *cryptoutilIdentityDomain.PushedAuthorizationRequest
		queryID   func(*cryptoutilIdentityDomain.PushedAuthorizationRequest) googleUuid.UUID
		expectErr error
	}{
		{
			name: "Found",
			setupPAR: func() *cryptoutilIdentityDomain.PushedAuthorizationRequest {
				return &cryptoutilIdentityDomain.PushedAuthorizationRequest{
					ID:                  googleUuid.Must(googleUuid.NewV7()),
					RequestURI:          "urn:ietf:params:oauth:request_uri:getbyid",
					ClientID:            googleUuid.Must(googleUuid.NewV7()),
					ResponseType:        "code",
					RedirectURI:         "https://example.com/callback",
					CodeChallenge:       "challenge",
					CodeChallengeMethod: "S256",
					ExpiresAt:           time.Now().Add(90 * time.Second),
					CreatedAt:           time.Now(),
				}
			},
			queryID: func(par *cryptoutilIdentityDomain.PushedAuthorizationRequest) googleUuid.UUID {
				return par.ID
			},
			expectErr: nil,
		},
		{
			name: "NotFound",
			setupPAR: func() *cryptoutilIdentityDomain.PushedAuthorizationRequest {
				return &cryptoutilIdentityDomain.PushedAuthorizationRequest{
					ID:                  googleUuid.Must(googleUuid.NewV7()),
					RequestURI:          "urn:ietf:params:oauth:request_uri:notfound",
					ClientID:            googleUuid.Must(googleUuid.NewV7()),
					ResponseType:        "code",
					RedirectURI:         "https://example.com/callback",
					CodeChallenge:       "challenge",
					CodeChallengeMethod: "S256",
					ExpiresAt:           time.Now().Add(90 * time.Second),
					CreatedAt:           time.Now(),
				}
			},
			queryID: func(_ *cryptoutilIdentityDomain.PushedAuthorizationRequest) googleUuid.UUID {
				return googleUuid.Must(googleUuid.NewV7())
			},
			expectErr: cryptoutilIdentityAppErr.ErrPushedAuthorizationRequestNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := setupPARTestDB(t)
			repo := cryptoutilIdentityORM.NewPushedAuthorizationRequestRepository(db)

			par := tt.setupPAR()
			require.NoError(t, repo.Create(ctx, par))

			result, err := repo.GetByID(ctx, tt.queryID(par))

			if tt.expectErr != nil {
				require.ErrorIs(t, err, tt.expectErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, par.ID, result.ID)
			}
		})
	}
}

// TestPushedAuthorizationRequestRepository_Update tests Update paths.
func TestPushedAuthorizationRequestRepository_Update(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupPAR  func() *cryptoutilIdentityDomain.PushedAuthorizationRequest
		modifyPAR func(*cryptoutilIdentityDomain.PushedAuthorizationRequest)
	}{
		{
			name: "MarkAsUsed",
			setupPAR: func() *cryptoutilIdentityDomain.PushedAuthorizationRequest {
				return &cryptoutilIdentityDomain.PushedAuthorizationRequest{
					ID:                  googleUuid.Must(googleUuid.NewV7()),
					RequestURI:          "urn:ietf:params:oauth:request_uri:update",
					ClientID:            googleUuid.Must(googleUuid.NewV7()),
					ResponseType:        "code",
					RedirectURI:         "https://example.com/callback",
					CodeChallenge:       "challenge",
					CodeChallengeMethod: "S256",
					Used:                false,
					ExpiresAt:           time.Now().Add(90 * time.Second),
					CreatedAt:           time.Now(),
				}
			},
			modifyPAR: func(par *cryptoutilIdentityDomain.PushedAuthorizationRequest) {
				par.MarkAsUsed()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := setupPARTestDB(t)
			repo := cryptoutilIdentityORM.NewPushedAuthorizationRequestRepository(db)

			par := tt.setupPAR()
			require.NoError(t, repo.Create(ctx, par))
			require.False(t, par.Used)

			tt.modifyPAR(par)
			require.NoError(t, repo.Update(ctx, par))

			retrieved, err := repo.GetByID(ctx, par.ID)
			require.NoError(t, err)
			require.True(t, retrieved.Used)
			require.NotNil(t, retrieved.UsedAt)
		})
	}
}

// TestPushedAuthorizationRequestRepository_DeleteExpired tests DeleteExpired paths.
func TestPushedAuthorizationRequestRepository_DeleteExpired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupPARs      func() []*cryptoutilIdentityDomain.PushedAuthorizationRequest
		expectedCount  int64
		remainingCount int
	}{
		{
			name: "DeleteExpiredPARs",
			setupPARs: func() []*cryptoutilIdentityDomain.PushedAuthorizationRequest {
				now := time.Now().UTC()
				expired1 := &cryptoutilIdentityDomain.PushedAuthorizationRequest{
					ID:                  googleUuid.Must(googleUuid.NewV7()),
					RequestURI:          "urn:ietf:params:oauth:request_uri:expired1",
					ClientID:            googleUuid.Must(googleUuid.NewV7()),
					ResponseType:        "code",
					RedirectURI:         "https://example.com/callback",
					CodeChallenge:       "challenge",
					CodeChallengeMethod: "S256",
					ExpiresAt:           now.Add(-10 * time.Minute),
					CreatedAt:           now.Add(-20 * time.Minute),
				}
				expired2 := &cryptoutilIdentityDomain.PushedAuthorizationRequest{
					ID:                  googleUuid.Must(googleUuid.NewV7()),
					RequestURI:          "urn:ietf:params:oauth:request_uri:expired2",
					ClientID:            googleUuid.Must(googleUuid.NewV7()),
					ResponseType:        "code",
					RedirectURI:         "https://example.com/callback",
					CodeChallenge:       "challenge",
					CodeChallengeMethod: "S256",
					ExpiresAt:           now.Add(-5 * time.Minute),
					CreatedAt:           now.Add(-15 * time.Minute),
				}
				valid := &cryptoutilIdentityDomain.PushedAuthorizationRequest{
					ID:                  googleUuid.Must(googleUuid.NewV7()),
					RequestURI:          "urn:ietf:params:oauth:request_uri:valid",
					ClientID:            googleUuid.Must(googleUuid.NewV7()),
					ResponseType:        "code",
					RedirectURI:         "https://example.com/callback",
					CodeChallenge:       "challenge",
					CodeChallengeMethod: "S256",
					ExpiresAt:           now.Add(90 * time.Second),
					CreatedAt:           now,
				}

				return []*cryptoutilIdentityDomain.PushedAuthorizationRequest{expired1, expired2, valid}
			},
			expectedCount:  2,
			remainingCount: 1,
		},
		{
			name: "NoExpiredPARs",
			setupPARs: func() []*cryptoutilIdentityDomain.PushedAuthorizationRequest {
				now := time.Now().UTC()
				valid := &cryptoutilIdentityDomain.PushedAuthorizationRequest{
					ID:                  googleUuid.Must(googleUuid.NewV7()),
					RequestURI:          "urn:ietf:params:oauth:request_uri:valid2",
					ClientID:            googleUuid.Must(googleUuid.NewV7()),
					ResponseType:        "code",
					RedirectURI:         "https://example.com/callback",
					CodeChallenge:       "challenge",
					CodeChallengeMethod: "S256",
					ExpiresAt:           now.Add(90 * time.Second),
					CreatedAt:           now,
				}

				return []*cryptoutilIdentityDomain.PushedAuthorizationRequest{valid}
			},
			expectedCount:  0,
			remainingCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := setupPARTestDB(t)
			repo := cryptoutilIdentityORM.NewPushedAuthorizationRequestRepository(db)

			pars := tt.setupPARs()
			for _, par := range pars {
				require.NoError(t, repo.Create(ctx, par))
			}

			count, err := repo.DeleteExpired(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.expectedCount, count)

			// Verify remaining count.
			var remaining []cryptoutilIdentityDomain.PushedAuthorizationRequest

			require.NoError(t, db.Find(&remaining).Error)
			require.Len(t, remaining, tt.remainingCount)
		})
	}
}

// setupPARTestDB creates an in-memory SQLite database for testing (CGO-free).
func setupPARTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	ctx := context.Background()

	// Use testDSNInMemory for isolated per-test database (no shared cache for parallel test safety).
	dsn := testDSNInMemory
	sqlDB, err := sql.Open("sqlite", dsn)
	require.NoError(t, err)

	_, err = sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
	require.NoError(t, err)

	_, err = sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;")
	require.NoError(t, err)

	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(0)

	dialector := sqlite.Dialector{Conn: sqlDB}
	db, err := gorm.Open(dialector, &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	err = db.AutoMigrate(&cryptoutilIdentityDomain.PushedAuthorizationRequest{})
	require.NoError(t, err)

	return db
}
