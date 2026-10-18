// Copyright (c) 2025-2026 Justin Cranford.
package mfa

import (
	"context"
	"database/sql"
	"encoding/base32"
	"fmt"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
)

const (
	testInvalidTOTPSecret               = "%"
	testUnsupportedTOTPAlgorithmMessage = "unsupported algorithm"
)

func createTOTPTestService(t *testing.T) (*TOTPService, *gorm.DB) {
	t.Helper()

	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, "file::memory:?cache=private")
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })

	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&TOTPSecret{}, &BackupCode{}))

	return NewTOTPService(db), db
}

func TestTOTPService_GenerateTOTP(t *testing.T) {
	t.Parallel()

	encodeSecret := func(secret string) string {
		return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte(secret))
	}
	sha1Secret := encodeSecret("12345678901234567890")
	sha256Secret := encodeSecret("12345678901234567890123456789012")
	sha512Secret := encodeSecret("1234567890123456789012345678901234567890123456789012345678901234")

	tests := []struct {
		name      string
		secret    string
		algorithm string
		digits    int
		wantCode  string
		wantError string
	}{
		{name: "RFC 6238 SHA-1 vector", secret: sha1Secret, algorithm: cryptoutilSharedMagic.DefaultTOTPAlgorithm, digits: 6, wantCode: "287082"},
		{name: "RFC 6238 SHA-256 vector", secret: sha256Secret, algorithm: cryptoutilSharedMagic.SHA256, digits: 8, wantCode: "46119246"},
		{name: "RFC 6238 SHA-512 vector", secret: sha512Secret, algorithm: cryptoutilSharedMagic.SHA512, digits: 8, wantCode: "90693936"},
		{name: "invalid base32", secret: testInvalidTOTPSecret, algorithm: DefaultTOTPAlgorithm, digits: DefaultTOTPDigits, wantError: "failed to decode secret"},
		{name: testUnsupportedTOTPAlgorithmMessage, secret: "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ", algorithm: "unsupported", digits: DefaultTOTPDigits, wantError: testUnsupportedTOTPAlgorithmMessage},
	}

	service := NewTOTPService(nil)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			code, err := service.generateTOTP(tc.secret, 1, tc.algorithm, tc.digits)
			if tc.wantError != "" {
				require.ErrorContains(t, err, tc.wantError)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantCode, code)
		})
	}
}

func TestTOTPService_VerifyCode(t *testing.T) {
	t.Parallel()

	service := NewTOTPService(nil)
	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte("totp test secret"))
	totpSecret := &TOTPSecret{Secret: secret, Algorithm: DefaultTOTPAlgorithm, Digits: DefaultTOTPDigits, Period: DefaultTOTPPeriod}
	currentStep := time.Now().UTC().Unix() / int64(totpSecret.Period)
	validCode, err := service.generateTOTP(secret, currentStep, totpSecret.Algorithm, totpSecret.Digits)
	require.NoError(t, err)
	windowCode, err := service.generateTOTP(secret, currentStep+1, totpSecret.Algorithm, totpSecret.Digits)
	require.NoError(t, err)

	valid, err := service.verifyCode(totpSecret, validCode, 1)
	require.NoError(t, err)
	require.True(t, valid)

	valid, err = service.verifyCode(totpSecret, windowCode, 1)
	require.NoError(t, err)
	require.True(t, valid)

	valid, err = service.verifyCode(totpSecret, cryptoutilSharedMagic.DevSetupTestInvalidName, 1)
	require.NoError(t, err)
	require.False(t, valid)

	totpSecret.Secret = testInvalidTOTPSecret
	_, err = service.verifyCode(totpSecret, validCode, 1)
	require.ErrorContains(t, err, "failed to decode secret")
}

func TestTOTPService_GenerateOTPAuthURI(t *testing.T) {
	t.Parallel()

	service := NewTOTPService(nil)
	uri := service.generateOTPAuthURI("Example", "user@example.com", "SECRET")
	require.Equal(
		t,
		fmt.Sprintf("otpauth://totp/Example:user@example.com?secret=SECRET&issuer=Example&algorithm=%s&digits=%d&period=%d", DefaultTOTPAlgorithm, DefaultTOTPDigits, DefaultTOTPPeriod),
		uri,
	)
}

func TestTOTPService_EnrollVerifyAndBackupCodes(t *testing.T) {
	t.Parallel()

	service, db := createTOTPTestService(t)
	ctx := context.Background()
	userID := googleUuid.Must(googleUuid.NewV7())

	secret, uri, backupCodes, err := service.EnrollTOTP(ctx, userID, "Example", "user@example.com")
	require.NoError(t, err)
	require.NotNil(t, secret)
	require.Contains(t, uri, "otpauth://totp/Example:user@example.com")
	require.Len(t, backupCodes, BackupCodeCount)

	code, err := service.generateTOTP(secret.Secret, time.Now().UTC().Unix()/int64(secret.Period), secret.Algorithm, secret.Digits)
	require.NoError(t, err)
	require.NoError(t, service.VerifyTOTP(ctx, userID, code))

	var stored TOTPSecret
	require.NoError(t, db.Where("user_id = ?", userID).First(&stored).Error)
	require.True(t, stored.Verified)
	require.Zero(t, stored.FailedAttempts)
	require.False(t, stored.LastUsedAt.IsZero())

	require.NoError(t, service.VerifyBackupCode(ctx, userID, backupCodes[0]))
	require.ErrorContains(t, service.VerifyBackupCode(ctx, userID, backupCodes[0]), "invalid backup code")

	replacementCodes, err := service.GenerateBackupCodes(ctx, userID)
	require.NoError(t, err)
	require.Len(t, replacementCodes, BackupCodeCount)

	var backupCount int64
	require.NoError(t, db.Model(&BackupCode{}).Where("user_id = ?", userID).Count(&backupCount).Error)
	require.EqualValues(t, BackupCodeCount, backupCount)
}

func TestTOTPService_FailuresAndStepUp(t *testing.T) {
	t.Parallel()

	service, db := createTOTPTestService(t)
	ctx := context.Background()
	userID := googleUuid.Must(googleUuid.NewV7())

	required, err := service.RequiresMFAStepUp(ctx, userID)
	require.NoError(t, err)
	require.False(t, required)
	require.ErrorContains(t, service.VerifyTOTP(ctx, userID, "123456"), "not enrolled")

	secret, _, _, err := service.EnrollTOTP(ctx, userID, "Example", "user@example.com")
	require.NoError(t, err)
	required, err = service.RequiresMFAStepUp(ctx, userID)
	require.NoError(t, err)
	require.True(t, required)

	for range MaxFailedAttempts {
		require.ErrorContains(t, service.VerifyTOTP(ctx, userID, cryptoutilSharedMagic.DevSetupTestInvalidName), "invalid TOTP code")
	}

	require.ErrorIs(t, service.VerifyTOTP(ctx, userID, cryptoutilSharedMagic.DevSetupTestInvalidName), cryptoutilIdentityAppErr.ErrTOTPAccountLocked)

	require.NoError(t, db.Model(&TOTPSecret{}).Where("user_id = ?", userID).Update("locked_until", time.Time{}).Error)

	code, err := service.generateTOTP(secret.Secret, time.Now().UTC().Unix()/int64(secret.Period), secret.Algorithm, secret.Digits)
	require.NoError(t, err)
	require.NoError(t, service.VerifyTOTP(ctx, userID, code))

	required, err = service.RequiresMFAStepUp(ctx, userID)
	require.NoError(t, err)
	require.False(t, required)

	stale := time.Now().UTC().Add(-MFAStepUpDuration - time.Minute)
	require.NoError(t, db.Model(&TOTPSecret{}).Where("user_id = ?", userID).Update("last_used_at", stale).Error)
	required, err = service.RequiresMFAStepUp(ctx, userID)
	require.NoError(t, err)
	require.True(t, required)
}

func TestTOTPService_DatabaseErrors(t *testing.T) {
	t.Parallel()

	service, db := createTOTPTestService(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	ctx := context.Background()
	userID := googleUuid.Must(googleUuid.NewV7())
	_, _, _, err = service.EnrollTOTP(ctx, userID, "Example", "user@example.com")
	require.ErrorContains(t, err, "failed to save TOTP secret")
	require.ErrorContains(t, service.VerifyTOTP(ctx, userID, "123456"), "failed to get TOTP secret")
	_, err = service.RequiresMFAStepUp(ctx, userID)
	require.ErrorContains(t, err, "failed to get TOTP secret")
	_, err = service.GenerateBackupCodes(ctx, userID)
	require.ErrorContains(t, err, "failed to delete old backup codes")
	require.ErrorContains(t, service.VerifyBackupCode(ctx, userID, "unused"), "failed to get backup codes")
}

func TestTOTPModels_GenerateIDsWhenMissing(t *testing.T) {
	t.Parallel()

	_, db := createTOTPTestService(t)
	userID := googleUuid.Must(googleUuid.NewV7())
	now := time.Now().UTC()

	totpSecret := &TOTPSecret{
		UserID:    userID,
		Secret:    googleUuid.Must(googleUuid.NewV7()).String(),
		CreatedAt: now,
		UpdatedAt: now,
	}
	require.NoError(t, db.Create(totpSecret).Error)
	require.NotEqual(t, googleUuid.Nil, totpSecret.ID)

	backupCode := &BackupCode{
		UserID:    userID,
		CodeHash:  "stored-hash",
		CreatedAt: now,
		UpdatedAt: now,
	}
	require.NoError(t, db.Create(backupCode).Error)
	require.NotEqual(t, googleUuid.Nil, backupCode.ID)
}

func TestTOTPService_StorageErrors(t *testing.T) {
	t.Parallel()

	service, db := createTOTPTestService(t)
	ctx := context.Background()
	userID := googleUuid.Must(googleUuid.NewV7())
	secret, _, backupCodes, err := service.EnrollTOTP(ctx, userID, "Example", "user@example.com")
	require.NoError(t, err)

	require.NoError(t, db.Exec(`CREATE TRIGGER fail_totp_update BEFORE UPDATE ON totp_secrets BEGIN SELECT RAISE(ABORT, 'totp update rejected'); END`).Error)
	require.NoError(t, db.Exec(`CREATE TRIGGER fail_backup_insert BEFORE INSERT ON backup_codes BEGIN SELECT RAISE(ABORT, 'backup insert rejected'); END`).Error)
	require.NoError(t, db.Exec(`CREATE TRIGGER fail_backup_update BEFORE UPDATE ON backup_codes WHEN NEW.deleted_at IS NULL BEGIN SELECT RAISE(ABORT, 'backup update rejected'); END`).Error)

	require.ErrorContains(t, service.VerifyTOTP(ctx, userID, cryptoutilSharedMagic.DevSetupTestInvalidName), "failed to update failed attempts")
	validCode, err := service.generateTOTP(secret.Secret, time.Now().UTC().Unix()/int64(secret.Period), secret.Algorithm, secret.Digits)
	require.NoError(t, err)
	require.ErrorContains(t, service.VerifyTOTP(ctx, userID, validCode), "failed to update TOTP secret")

	require.ErrorContains(t, service.VerifyBackupCode(ctx, userID, backupCodes[0]), "failed to mark backup code as used")
	_, err = service.GenerateBackupCodes(ctx, userID)
	require.ErrorContains(t, err, "failed to save backup codes")

	_, _, _, err = service.EnrollTOTP(ctx, googleUuid.Must(googleUuid.NewV7()), "Example", "second@example.com")
	require.ErrorContains(t, err, "failed to generate backup codes")
}
