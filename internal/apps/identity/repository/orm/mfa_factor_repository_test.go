// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

func TestMFAFactorRepository_Create(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewMFAFactorRepository(testDB.db)
	ctx := context.Background()

	authProfileID := googleUuid.Must(googleUuid.NewV7())

	factor := &cryptoutilIdentityDomain.MFAFactor{
		Name:          "totp_factor",
		Description:   "TOTP authentication factor",
		FactorType:    cryptoutilIdentityDomain.MFAFactorTypeTOTP,
		Order:         1,
		Required:      true,
		TOTPAlgorithm: "SHA256",
		TOTPDigits:    6,
		TOTPPeriod:    30,
		AuthProfileID: authProfileID,
		Enabled:       true,
	}

	err := repo.Create(ctx, factor)
	require.NoError(t, err, "Create should succeed")

	retrieved, err := repo.GetByID(ctx, factor.ID)
	require.NoError(t, err, "GetByID should find created factor")
	require.Equal(t, factor.ID, retrieved.ID)
	require.Equal(t, "totp_factor", retrieved.Name)
	require.Equal(t, cryptoutilIdentityDomain.MFAFactorTypeTOTP, retrieved.FactorType)
}

func TestMFAFactorRepository_GetByID(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewMFAFactorRepository(testDB.db)
	ctx := context.Background()

	nonExistentID := googleUuid.Must(googleUuid.NewV7())

	_, err := repo.GetByID(ctx, nonExistentID)
	require.Error(t, err, "GetByID should return error for non-existent factor")
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrMFAFactorNotFound)
}

func TestMFAFactorRepository_GetByAuthProfileID(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewMFAFactorRepository(testDB.db)
	ctx := context.Background()

	authProfileID := googleUuid.Must(googleUuid.NewV7())

	factors := []*cryptoutilIdentityDomain.MFAFactor{
		{
			Name:          "totp_factor",
			FactorType:    cryptoutilIdentityDomain.MFAFactorTypeTOTP,
			Order:         1,
			AuthProfileID: authProfileID,
			Enabled:       true,
		},
		{
			Name:          "sms_factor",
			FactorType:    cryptoutilIdentityDomain.MFAFactorTypeSMSOTP,
			Order:         2,
			AuthProfileID: authProfileID,
			Enabled:       true,
		},
	}

	for _, factor := range factors {
		err := repo.Create(ctx, factor)
		require.NoError(t, err)
	}

	retrieved, err := repo.GetByAuthProfileID(ctx, authProfileID)
	require.NoError(t, err, "GetByAuthProfileID should succeed")
	require.Len(t, retrieved, 2, "Should return 2 factors")
	require.Equal(t, 1, retrieved[0].Order, "First factor should have order 1")
	require.Equal(t, 2, retrieved[1].Order, "Second factor should have order 2")
}

func TestMFAFactorRepository_Update(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewMFAFactorRepository(testDB.db)
	ctx := context.Background()

	authProfileID := googleUuid.Must(googleUuid.NewV7())

	factor := &cryptoutilIdentityDomain.MFAFactor{
		Name:          "totp_factor",
		Description:   "Original description",
		FactorType:    cryptoutilIdentityDomain.MFAFactorTypeTOTP,
		Order:         1,
		AuthProfileID: authProfileID,
		Enabled:       true,
	}

	err := repo.Create(ctx, factor)
	require.NoError(t, err)

	factor.Description = "Updated description"
	factor.Required = true

	err = repo.Update(ctx, factor)
	require.NoError(t, err, "Update should succeed")

	updated, err := repo.GetByID(ctx, factor.ID)
	require.NoError(t, err, "GetByID should find updated factor")
	require.Equal(t, "Updated description", updated.Description)
	require.True(t, updated.Required.Bool())
}

func TestMFAFactorRepository_Delete(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewMFAFactorRepository(testDB.db)
	ctx := context.Background()

	authProfileID := googleUuid.Must(googleUuid.NewV7())

	factor := &cryptoutilIdentityDomain.MFAFactor{
		Name:          "totp_factor",
		FactorType:    cryptoutilIdentityDomain.MFAFactorTypeTOTP,
		Order:         1,
		AuthProfileID: authProfileID,
		Enabled:       true,
	}

	err := repo.Create(ctx, factor)
	require.NoError(t, err)

	err = repo.Delete(ctx, factor.ID)
	require.NoError(t, err, "Delete should succeed")

	_, err = repo.GetByID(ctx, factor.ID)
	require.Error(t, err, "GetByID should fail after soft delete")
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrMFAFactorNotFound)
}

func TestMFAFactorRepository_List(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewMFAFactorRepository(testDB.db)
	ctx := context.Background()

	authProfileID := googleUuid.Must(googleUuid.NewV7())

	for i := 0; i < 5; i++ {
		factor := &cryptoutilIdentityDomain.MFAFactor{
			Name:          googleUuid.Must(googleUuid.NewV7()).String(),
			FactorType:    cryptoutilIdentityDomain.MFAFactorTypeTOTP,
			Order:         i + 1,
			AuthProfileID: authProfileID,
			Enabled:       true,
		}
		err := repo.Create(ctx, factor)
		require.NoError(t, err)
	}

	factors, err := repo.List(ctx, 0, 3)
	require.NoError(t, err, "List should succeed")
	require.Len(t, factors, 3, "Should return 3 factors with limit 3")

	factors, err = repo.List(ctx, 3, 3)
	require.NoError(t, err, "List should succeed")
	require.Len(t, factors, 2, "Should return 2 factors with offset 3")
}

func TestMFAFactorRepository_Count(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewMFAFactorRepository(testDB.db)
	ctx := context.Background()

	count, err := repo.Count(ctx)
	require.NoError(t, err, "Count should succeed")
	require.Equal(t, int64(0), count, "Count should be 0 initially")

	authProfileID := googleUuid.Must(googleUuid.NewV7())

	for i := 0; i < 5; i++ {
		factor := &cryptoutilIdentityDomain.MFAFactor{
			Name:          googleUuid.Must(googleUuid.NewV7()).String(),
			FactorType:    cryptoutilIdentityDomain.MFAFactorTypeTOTP,
			Order:         i + 1,
			AuthProfileID: authProfileID,
			Enabled:       true,
		}
		err := repo.Create(ctx, factor)
		require.NoError(t, err)
	}

	count, err = repo.Count(ctx)
	require.NoError(t, err, "Count should succeed")
	require.Equal(t, int64(5), count, "Count should be 5 after creating 5 factors")
}
