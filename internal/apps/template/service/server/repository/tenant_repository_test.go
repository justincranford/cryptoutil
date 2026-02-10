// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// setupTestDB creates an in-memory SQLite database for testing.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := cryptoutilSharedMagic.SQLiteInMemoryDSN

	sqlDB, err := sql.Open("sqlite", dsn)
	require.NoError(t, err)

	_, err = sqlDB.ExecContext(context.Background(), "PRAGMA journal_mode=WAL;")
	require.NoError(t, err)

	_, err = sqlDB.ExecContext(context.Background(), "PRAGMA busy_timeout = 30000;")
	require.NoError(t, err)

	dialector := sqlite.Dialector{Conn: sqlDB}
	db, err := gorm.Open(dialector, &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	sqlDB, err = db.DB()
	require.NoError(t, err)

	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(0)

	// Run migrations
	err = db.AutoMigrate(&Tenant{}, &User{}, &Client{}, &Role{}, &UserRole{}, &ClientRole{}, &TenantRealm{}, &UnverifiedUser{}, &UnverifiedClient{})
	require.NoError(t, err)

	return db
}

// uniqueTenantName returns a unique tenant name for tests.
func uniqueTenantName(base string) string {
	return base + " " + googleUuid.New().String()[:8]
}

func TestTenantRepository_Create(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := NewTenantRepository(db)
	ctx := context.Background()

	// Create a tenant first for duplicate test
	dupName := uniqueTenantName("DuplicateTest")
	firstTenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        dupName,
		Description: "Test tenant",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}
	err := repo.Create(ctx, firstTenant)
	require.NoError(t, err)

	tests := []struct {
		name      string
		tenant    *Tenant
		wantError bool
	}{
		{
			name: "happy path - valid tenant",
			tenant: &Tenant{
				ID:          googleUuid.New(),
				Name:        uniqueTenantName("Acme"),
				Description: "Test tenant",
				Active:      1,
				CreatedAt:   time.Now().UTC(),
			},
			wantError: false,
		},
		{
			name: "duplicate tenant name",
			tenant: &Tenant{
				ID:          googleUuid.New(),
				Name:        dupName,
				Description: "Duplicate tenant",
				Active:      1,
				CreatedAt:   time.Now().UTC(),
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(ctx, tt.tenant)

			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTenantRepository_GetByID(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := NewTenantRepository(db)
	ctx := context.Background()

	tenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueTenantName("GetByID"),
		Description: "Test tenant",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}

	err := repo.Create(ctx, tenant)
	require.NoError(t, err)

	tests := []struct {
		name      string
		id        googleUuid.UUID
		wantError bool
	}{
		{
			name:      "happy path - existing tenant",
			id:        tenant.ID,
			wantError: false,
		},
		{
			name:      "not found - non-existent tenant",
			id:        googleUuid.New(),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByID(ctx, tt.id)

			if tt.wantError {
				require.Error(t, err)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, tenant.ID, result.ID)
				require.Equal(t, tenant.Name, result.Name)
			}
		})
	}
}

func TestTenantRepository_GetByName(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := NewTenantRepository(db)
	ctx := context.Background()

	tenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueTenantName("GetByName"),
		Description: "Test tenant",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}

	err := repo.Create(ctx, tenant)
	require.NoError(t, err)

	tests := []struct {
		name       string
		tenantName string
		wantError  bool
	}{
		{
			name:       "happy path - existing tenant",
			tenantName: tenant.Name,
			wantError:  false,
		},
		{
			name:       "not found - non-existent tenant",
			tenantName: "Non-Existent Corp",
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByName(ctx, tt.tenantName)

			if tt.wantError {
				require.Error(t, err)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, tenant.ID, result.ID)
				require.Equal(t, tenant.Name, result.Name)
			}
		})
	}
}

func TestTenantRepository_List(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := NewTenantRepository(db)
	ctx := context.Background()

	activeTenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueTenantName("ActiveList"),
		Description: "Active tenant",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}

	inactiveTenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueTenantName("InactiveList"),
		Description: "Inactive tenant",
		Active:      0,
		CreatedAt:   time.Now().UTC(),
	}

	err := repo.Create(ctx, activeTenant)
	require.NoError(t, err)

	err = repo.Create(ctx, inactiveTenant)
	require.NoError(t, err)

	tests := []struct {
		name        string
		activeOnly  bool
		minCount    int
		hasActive   bool
		hasInactive bool
	}{
		{
			name:        "all tenants",
			activeOnly:  false,
			minCount:    2,
			hasActive:   true,
			hasInactive: true,
		},
		{
			name:       "active tenants only",
			activeOnly: true,
			minCount:   1,
			hasActive:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.List(ctx, tt.activeOnly)

			require.NoError(t, err)
			require.GreaterOrEqual(t, len(result), tt.minCount)

			// Verify our test tenants are in the results.
			var foundActive, foundInactive bool

			for _, tenant := range result {
				if tenant.ID == activeTenant.ID {
					foundActive = true
				}

				if tenant.ID == inactiveTenant.ID {
					foundInactive = true
				}

				if tt.activeOnly {
					require.Equal(t, 1, tenant.Active)
				}
			}

			if tt.hasActive {
				require.True(t, foundActive, "Active tenant should be in results")
			}

			if tt.hasInactive {
				require.True(t, foundInactive, "Inactive tenant should be in results")
			}
		})
	}
}

func TestTenantRepository_Update(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := NewTenantRepository(db)
	ctx := context.Background()

	tenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueTenantName("Update"),
		Description: "Test tenant",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}

	err := repo.Create(ctx, tenant)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond) // Ensure UpdatedAt will be different.

	tenant.Description = "Updated description"
	tenant.Active = 0

	err = repo.Update(ctx, tenant)
	require.NoError(t, err)

	result, err := repo.GetByID(ctx, tenant.ID)
	require.NoError(t, err)
	require.Equal(t, "Updated description", result.Description)
	require.Equal(t, 0, result.Active)
	require.True(t, result.UpdatedAt.After(tenant.CreatedAt))
}

func TestTenantRepository_Delete(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	tests := []struct {
		name      string
		setupFunc func(tenantID googleUuid.UUID)
		wantError bool
		errorMsg  string
	}{
		{
			name: "happy path - tenant without users or clients",
			setupFunc: func(_ googleUuid.UUID) {
				// No setup needed
			},
			wantError: false,
		},
		{
			name: "blocked - tenant has users",
			setupFunc: func(tenantID googleUuid.UUID) {
				user := &User{
					ID:        googleUuid.New(),
					TenantID:  tenantID,
					Username:  "testuser-" + googleUuid.New().String()[:8],
					Email:     "test-" + googleUuid.New().String()[:8] + "@example.com",
					Active:    1,
					CreatedAt: time.Now().UTC(),
				}
				err := userRepo.Create(ctx, user)
				require.NoError(t, err)
			},
			wantError: true,
			errorMsg:  "cannot delete tenant: has 1 users and 0 clients",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenant := &Tenant{
				ID:          googleUuid.New(),
				Name:        "Tenant " + googleUuid.NewString(),
				Description: "Test tenant",
				Active:      1,
				CreatedAt:   time.Now().UTC(),
			}

			err := tenantRepo.Create(ctx, tenant)
			require.NoError(t, err)

			tt.setupFunc(tenant.ID)

			err = tenantRepo.Delete(ctx, tenant.ID)

			if tt.wantError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)

				_, err = tenantRepo.GetByID(ctx, tenant.ID)
				require.Error(t, err)
			}
		})
	}
}

func TestTenantRepository_CountUsersAndClients(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	userRepo := NewUserRepository(db)
	clientRepo := NewClientRepository(db)
	ctx := context.Background()

	tenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueTenantName("Count"),
		Description: "Test tenant",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}

	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	user := &User{
		ID:        googleUuid.New(),
		TenantID:  tenant.ID,
		Username:  "testuser-" + googleUuid.New().String()[:8],
		Email:     "test-" + googleUuid.New().String()[:8] + "@example.com",
		Active:    1,
		CreatedAt: time.Now().UTC(),
	}

	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	client := &Client{
		ID:        googleUuid.New(),
		TenantID:  tenant.ID,
		ClientID:  "client-" + googleUuid.New().String()[:8],
		Active:    1,
		CreatedAt: time.Now().UTC(),
	}

	err = clientRepo.Create(ctx, client)
	require.NoError(t, err)

	userCount, clientCount, err := tenantRepo.CountUsersAndClients(ctx, tenant.ID)
	require.NoError(t, err)
	require.Equal(t, int64(1), userCount)
	require.Equal(t, int64(1), clientCount)
}
