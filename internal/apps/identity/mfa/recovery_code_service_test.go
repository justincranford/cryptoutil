// Copyright (c) 2025 Justin Cranford

package mfa_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite" // CGO-free SQLite driver.

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityMfa "cryptoutil/internal/apps/identity/mfa"
	cryptoutilIdentityORM "cryptoutil/internal/apps/identity/repository/orm"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// createTestDB creates an in-memory SQLite database for testing.
func createTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Use modernc.org/sqlite driver (CGO-free).
	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, "file::memory:?cache=private")
	require.NoError(t, err)

	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	// Auto-migrate schema.
	err = db.AutoMigrate(&cryptoutilIdentityDomain.RecoveryCode{})
	require.NoError(t, err)

	return db
}

func TestRecoveryCodeService_GenerateForUser(t *testing.T) {
	t.Parallel()

	db := createTestDB(t)
	repo := cryptoutilIdentityORM.NewRecoveryCodeRepository(db)
	service := cryptoutilIdentityMfa.NewRecoveryCodeService(repo)

	userID := googleUuid.New()
	count := cryptoutilSharedMagic.JoseJADefaultMaxMaterials

	codes, err := service.GenerateForUser(context.Background(), userID, count)
	require.NoError(t, err)
	require.Len(t, codes, count, "should generate %d codes", count)

	// Verify all codes are unique.
	seen := make(map[string]bool, count)
	for _, code := range codes {
		require.False(t, seen[code], "duplicate code detected: %q", code)
		seen[code] = true
	}

	// Verify codes stored in database.
	storedCodes, err := repo.GetByUserID(context.Background(), userID)
	require.NoError(t, err)
	require.Len(t, storedCodes, count, "should store %d codes in database", count)

	// Verify codes are hashed (not plaintext).
	for _, storedCode := range storedCodes {
		require.NotEmpty(t, storedCode.CodeHash)
		require.NotContains(t, codes, storedCode.CodeHash, "code should be hashed, not plaintext")
		require.False(t, storedCode.IsUsed())
		require.False(t, storedCode.IsExpired())
	}
}

func TestRecoveryCodeService_Verify_Success(t *testing.T) {
	t.Parallel()

	db := createTestDB(t)
	repo := cryptoutilIdentityORM.NewRecoveryCodeRepository(db)
	service := cryptoutilIdentityMfa.NewRecoveryCodeService(repo)

	userID := googleUuid.New()

	// Generate codes.
	codes, err := service.GenerateForUser(context.Background(), userID, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)
	require.NoError(t, err)
	require.Len(t, codes, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)

	// Verify first code.
	err = service.Verify(context.Background(), userID, codes[0])
	require.NoError(t, err)

	// Verify code is marked as used.
	storedCodes, err := repo.GetByUserID(context.Background(), userID)
	require.NoError(t, err)

	usedCount := 0

	for _, storedCode := range storedCodes {
		if storedCode.IsUsed() {
			usedCount++

			require.NotNil(t, storedCode.UsedAt)
		}
	}

	require.Equal(t, 1, usedCount, "exactly 1 code should be marked as used")
}

func TestRecoveryCodeService_Verify_InvalidCode(t *testing.T) {
	t.Parallel()

	db := createTestDB(t)
	repo := cryptoutilIdentityORM.NewRecoveryCodeRepository(db)
	service := cryptoutilIdentityMfa.NewRecoveryCodeService(repo)

	userID := googleUuid.New()

	// Generate codes.
	_, err := service.GenerateForUser(context.Background(), userID, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)
	require.NoError(t, err)

	// Try to verify invalid code.
	err = service.Verify(context.Background(), userID, "XXXX-XXXX-XXXX-XXXX")
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrRecoveryCodeNotFound)
}

func TestRecoveryCodeService_Verify_AlreadyUsed(t *testing.T) {
	t.Parallel()

	db := createTestDB(t)
	repo := cryptoutilIdentityORM.NewRecoveryCodeRepository(db)
	service := cryptoutilIdentityMfa.NewRecoveryCodeService(repo)

	userID := googleUuid.New()

	// Generate codes.
	codes, err := service.GenerateForUser(context.Background(), userID, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)
	require.NoError(t, err)

	// Use first code.
	err = service.Verify(context.Background(), userID, codes[0])
	require.NoError(t, err)

	// Try to use same code again.
	err = service.Verify(context.Background(), userID, codes[0])
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrRecoveryCodeNotFound)
}

func TestRecoveryCodeService_Verify_Expired(t *testing.T) {
	t.Parallel()

	db := createTestDB(t)
	repo := cryptoutilIdentityORM.NewRecoveryCodeRepository(db)
	service := cryptoutilIdentityMfa.NewRecoveryCodeService(repo)

	userID := googleUuid.New()

	// Create expired code manually.
	plaintext := "TEST-CODE-XXXX-XXXX"
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	require.NoError(t, err)

	expiredCode := &cryptoutilIdentityDomain.RecoveryCode{
		ID:        googleUuid.New(),
		UserID:    userID,
		CodeHash:  string(hash),
		Used:      false,
		UsedAt:    nil,
		CreatedAt: time.Now().UTC().Add(-cryptoutilSharedMagic.DefaultRecoveryCodeLifetime - 1*time.Hour),
		ExpiresAt: time.Now().UTC().Add(-1 * time.Hour), // Expired 1 hour ago.
	}

	err = repo.Create(context.Background(), expiredCode)
	require.NoError(t, err)

	// Try to verify expired code.
	err = service.Verify(context.Background(), userID, plaintext)
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrRecoveryCodeNotFound)
}

func TestRecoveryCodeService_RegenerateForUser(t *testing.T) {
	t.Parallel()

	db := createTestDB(t)
	repo := cryptoutilIdentityORM.NewRecoveryCodeRepository(db)
	service := cryptoutilIdentityMfa.NewRecoveryCodeService(repo)

	userID := googleUuid.New()

	// Generate initial codes.
	oldCodes, err := service.GenerateForUser(context.Background(), userID, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)
	require.NoError(t, err)
	require.Len(t, oldCodes, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)

	// Regenerate codes.
	newCodes, err := service.RegenerateForUser(context.Background(), userID, cryptoutilSharedMagic.IMMinPasswordLength)
	require.NoError(t, err)
	require.Len(t, newCodes, cryptoutilSharedMagic.IMMinPasswordLength)

	// Verify old codes are deleted.
	for _, oldCode := range oldCodes {
		err = service.Verify(context.Background(), userID, oldCode)
		require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrRecoveryCodeNotFound)
	}

	// Verify new codes work.
	err = service.Verify(context.Background(), userID, newCodes[0])
	require.NoError(t, err)
}

func TestRecoveryCodeService_GetRemainingCount(t *testing.T) {
	t.Parallel()

	db := createTestDB(t)
	repo := cryptoutilIdentityORM.NewRecoveryCodeRepository(db)
	service := cryptoutilIdentityMfa.NewRecoveryCodeService(repo)

	userID := googleUuid.New()

	// Generate codes.
	codes, err := service.GenerateForUser(context.Background(), userID, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	// Check initial count.
	count, err := service.GetRemainingCount(context.Background(), userID)
	require.NoError(t, err)
	require.Equal(t, int64(cryptoutilSharedMagic.JoseJADefaultMaxMaterials), count, "should have 10 unused codes")

	// Use 3 codes.
	for i := range 3 {
		err = service.Verify(context.Background(), userID, codes[i])
		require.NoError(t, err)
	}

	// Check remaining count.
	count, err = service.GetRemainingCount(context.Background(), userID)
	require.NoError(t, err)
	require.Equal(t, int64(cryptoutilSharedMagic.GitRecentActivityDays), count, "should have 7 unused codes remaining")
}
