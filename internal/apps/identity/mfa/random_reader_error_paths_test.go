// Copyright (c) 2025-2026 Justin Cranford.
//

package mfa_test

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityMfa "cryptoutil/internal/apps/identity/mfa"
)

// Sequential: replaces crypto/rand.Reader to exercise cryptographic random-source failures.
func TestEmailOTP_RandomReaderError(t *testing.T) {
	originalReader := crand.Reader
	defer func() {
		crand.Reader = originalReader
	}()

	crand.Reader = bytes.NewReader(nil)

	_, err := cryptoutilIdentityMfa.GenerateEmailOTP()
	require.ErrorContains(t, err, "failed to generate random number")

	service := cryptoutilIdentityMfa.NewEmailOTPService(&errorEmailOTPRepository{}, errorEmailService{})
	err = service.SendOTP(context.Background(), googleUuid.Must(googleUuid.NewV7()), testUserEmail)
	require.ErrorContains(t, err, "failed to generate OTP")
}

// Sequential: replaces crypto/rand.Reader to exercise MFA random-source failures.
func TestMFA_RandomReaderErrors(t *testing.T) {
	originalReader := crand.Reader
	defer func() {
		crand.Reader = originalReader
	}()

	crand.Reader = bytes.NewReader(nil)
	ctx := context.Background()
	userID := googleUuid.Must(googleUuid.NewV7())

	_, err := cryptoutilIdentityMfa.GenerateRecoveryCode()
	require.ErrorContains(t, err, "failed to generate random bytes")

	_, err = cryptoutilIdentityMfa.GenerateRecoveryCodes(1)
	require.ErrorContains(t, err, "failed to generate random bytes")

	recoveryService := cryptoutilIdentityMfa.NewRecoveryCodeService(nil)
	_, err = recoveryService.GenerateForUser(ctx, userID, 1)
	require.ErrorContains(t, err, "failed to generate recovery codes")

	db := createTestDB(t)
	require.NoError(t, db.AutoMigrate(&cryptoutilIdentityMfa.TOTPSecret{}, &cryptoutilIdentityMfa.BackupCode{}))
	totpService := cryptoutilIdentityMfa.NewTOTPService(db)
	_, _, _, err = totpService.EnrollTOTP(ctx, userID, "Example", "user@example.com")
	require.ErrorContains(t, err, "failed to generate TOTP secret")

	_, err = totpService.GenerateBackupCodes(ctx, userID)
	require.ErrorContains(t, err, "failed to generate backup code")
}

// Sequential: replaces crypto/rand.Reader with fixed bytes to force a duplicate recovery code.
func TestGenerateRecoveryCodes_RetriesDuplicate(t *testing.T) {
	originalReader := crand.Reader
	defer func() {
		crand.Reader = originalReader
	}()

	duplicate := bytes.Repeat([]byte{0}, 16)
	firstUnique := bytes.Repeat([]byte{1}, 16)
	secondUnique := bytes.Repeat([]byte{2}, 16)
	crand.Reader = bytes.NewReader(bytes.Join([][]byte{duplicate, firstUnique, duplicate, secondUnique}, nil))

	codes, err := cryptoutilIdentityMfa.GenerateRecoveryCodes(3)
	require.NoError(t, err)
	require.Len(t, codes, 3)
	require.NotEqual(t, codes[0], codes[1])
	require.NotEqual(t, codes[0], codes[2])
	require.NotEqual(t, codes[1], codes[2])
}
