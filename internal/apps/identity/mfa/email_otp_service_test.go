// Copyright (c) 2025 Justin Cranford

package mfa_test

import (
	"context"
	"testing"
	"time"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityEmail "cryptoutil/internal/apps/identity/email"
	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"
	cryptoutilIdentityMfa "cryptoutil/internal/apps/identity/mfa"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const testUserEmail = "user@example.com"

// mockEmailOTPRepository is a simple in-memory implementation for testing.
type mockEmailOTPRepository struct {
	otps map[string]*cryptoutilIdentityDomain.EmailOTP
}

func newMockEmailOTPRepository() *mockEmailOTPRepository {
	return &mockEmailOTPRepository{
		otps: make(map[string]*cryptoutilIdentityDomain.EmailOTP),
	}
}

func (m *mockEmailOTPRepository) Create(_ context.Context, otp *cryptoutilIdentityDomain.EmailOTP) error {
	m.otps[otp.UserID.String()] = otp

	return nil
}

func (m *mockEmailOTPRepository) GetByUserID(_ context.Context, userID googleUuid.UUID) (*cryptoutilIdentityDomain.EmailOTP, error) {
	otp, exists := m.otps[userID.String()]
	if !exists {
		return nil, cryptoutilIdentityAppErr.ErrEmailOTPNotFound
	}

	return otp, nil
}

func (m *mockEmailOTPRepository) Update(_ context.Context, otp *cryptoutilIdentityDomain.EmailOTP) error {
	m.otps[otp.UserID.String()] = otp

	return nil
}

func TestEmailOTPService_SendOTP(t *testing.T) {
	t.Parallel()

	repo := newMockEmailOTPRepository()
	mockEmail := cryptoutilIdentityEmail.NewMockEmailService()
	service := cryptoutilIdentityMfa.NewEmailOTPService(repo, mockEmail)

	ctx := context.Background()
	userID := googleUuid.New()
	email := testUserEmail

	// Send OTP.
	err := service.SendOTP(ctx, userID, email)
	require.NoError(t, err)

	// Verify email was sent.
	require.Len(t, mockEmail.SentEmails, 1)
	lastEmail := mockEmail.GetLastEmail()
	require.NotNil(t, lastEmail)
	require.Equal(t, email, lastEmail.To)
	require.Contains(t, lastEmail.Subject, "One-Time Password")

	// Extract OTP from email.
	otpCode, found := mockEmail.ContainsOTP(lastEmail)
	require.True(t, found, "Email should contain OTP")
	require.Len(t, otpCode, cryptoutilIdentityMagic.DefaultEmailOTPLength)

	// Verify OTP record was created in database.
	otp, err := repo.GetByUserID(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, userID, otp.UserID)
	require.False(t, otp.Used)
	require.False(t, otp.IsExpired())
}

func TestEmailOTPService_VerifyOTP_Success(t *testing.T) {
	t.Parallel()

	repo := newMockEmailOTPRepository()
	mockEmail := cryptoutilIdentityEmail.NewMockEmailService()
	service := cryptoutilIdentityMfa.NewEmailOTPService(repo, mockEmail)

	ctx := context.Background()
	userID := googleUuid.New()
	email := testUserEmail

	// Send OTP.
	err := service.SendOTP(ctx, userID, email)
	require.NoError(t, err)

	// Extract OTP from email.
	lastEmail := mockEmail.GetLastEmail()
	otpCode, found := mockEmail.ContainsOTP(lastEmail)
	require.True(t, found)

	// Verify OTP.
	err = service.VerifyOTP(ctx, userID, otpCode)
	require.NoError(t, err)

	// Verify OTP is marked as used.
	otp, err := repo.GetByUserID(ctx, userID)
	require.NoError(t, err)
	require.True(t, otp.Used)
	require.NotNil(t, otp.UsedAt)
}

func TestEmailOTPService_VerifyOTP_InvalidCode(t *testing.T) {
	t.Parallel()

	repo := newMockEmailOTPRepository()
	mockEmail := cryptoutilIdentityEmail.NewMockEmailService()
	service := cryptoutilIdentityMfa.NewEmailOTPService(repo, mockEmail)

	ctx := context.Background()
	userID := googleUuid.New()
	email := testUserEmail

	// Send OTP.
	err := service.SendOTP(ctx, userID, email)
	require.NoError(t, err)

	// Verify with wrong OTP.
	err = service.VerifyOTP(ctx, userID, "000000")
	require.Error(t, err)
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrInvalidOTP)
}

func TestEmailOTPService_VerifyOTP_AlreadyUsed(t *testing.T) {
	t.Parallel()

	repo := newMockEmailOTPRepository()
	mockEmail := cryptoutilIdentityEmail.NewMockEmailService()
	service := cryptoutilIdentityMfa.NewEmailOTPService(repo, mockEmail)

	ctx := context.Background()
	userID := googleUuid.New()
	email := testUserEmail

	// Send OTP.
	err := service.SendOTP(ctx, userID, email)
	require.NoError(t, err)

	// Extract OTP.
	lastEmail := mockEmail.GetLastEmail()
	otpCode, found := mockEmail.ContainsOTP(lastEmail)
	require.True(t, found)

	// Verify OTP first time (should succeed).
	err = service.VerifyOTP(ctx, userID, otpCode)
	require.NoError(t, err)

	// Verify OTP second time (should fail).
	err = service.VerifyOTP(ctx, userID, otpCode)
	require.Error(t, err)
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrOTPAlreadyUsed)
}

func TestEmailOTPService_VerifyOTP_Expired(t *testing.T) {
	t.Parallel()

	repo := newMockEmailOTPRepository()
	mockEmail := cryptoutilIdentityEmail.NewMockEmailService()
	service := cryptoutilIdentityMfa.NewEmailOTPService(repo, mockEmail)

	ctx := context.Background()
	userID := googleUuid.New()
	email := testUserEmail

	// Send OTP.
	err := service.SendOTP(ctx, userID, email)
	require.NoError(t, err)

	// Extract OTP.
	lastEmail := mockEmail.GetLastEmail()
	otpCode, found := mockEmail.ContainsOTP(lastEmail)
	require.True(t, found)

	// Manually expire the OTP.
	otp, err := repo.GetByUserID(ctx, userID)
	require.NoError(t, err)

	otp.ExpiresAt = time.Now().UTC().Add(-1 * time.Minute)
	err = repo.Update(ctx, otp)
	require.NoError(t, err)

	// Verify OTP (should fail due to expiration).
	err = service.VerifyOTP(ctx, userID, otpCode)
	require.Error(t, err)
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrExpiredOTP)
}

func TestEmailOTPService_RateLimit(t *testing.T) {
	t.Parallel()

	repo := newMockEmailOTPRepository()
	mockEmail := cryptoutilIdentityEmail.NewMockEmailService()
	service := cryptoutilIdentityMfa.NewEmailOTPService(repo, mockEmail)

	ctx := context.Background()
	userID := googleUuid.New()
	email := testUserEmail

	// Send OTP 5 times (rate limit).
	for i := 0; i < cryptoutilIdentityMagic.DefaultEmailOTPRateLimit; i++ {
		err := service.SendOTP(ctx, userID, email)
		require.NoError(t, err, "Request %d should succeed", i+1)
	}

	// 6th request should fail (rate limit exceeded).
	err := service.SendOTP(ctx, userID, email)
	require.Error(t, err)
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrRateLimitExceeded)
}
