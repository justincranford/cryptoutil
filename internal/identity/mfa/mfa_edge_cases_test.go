// Copyright (c) 2025 Justin Cranford

package mfa_test

import (
	"context"
	"errors"
	"testing"
	"time"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityEmail "cryptoutil/internal/identity/email"
	cryptoutilIdentityMFA "cryptoutil/internal/identity/mfa"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// failingEmailOTPRepository simulates database failures.
type failingEmailOTPRepository struct {
	createError      error
	getByUserIDError error
	updateError      error
}

func (f *failingEmailOTPRepository) Create(_ context.Context, _ *cryptoutilIdentityDomain.EmailOTP) error {
	if f.createError != nil {
		return f.createError
	}

	return nil
}

func (f *failingEmailOTPRepository) GetByUserID(_ context.Context, _ googleUuid.UUID) (*cryptoutilIdentityDomain.EmailOTP, error) {
	if f.getByUserIDError != nil {
		return nil, f.getByUserIDError
	}

	return nil, cryptoutilIdentityAppErr.ErrEmailOTPNotFound
}

func (f *failingEmailOTPRepository) Update(_ context.Context, _ *cryptoutilIdentityDomain.EmailOTP) error {
	if f.updateError != nil {
		return f.updateError
	}

	return nil
}

// failingEmailService simulates email sending failures.
type failingEmailService struct {
	sendError error
}

func (f *failingEmailService) SendEmail(_ context.Context, _, _, _ string) error {
	if f.sendError != nil {
		return f.sendError
	}

	return nil
}

// TestEmailOTPService_SendOTP_CreateError tests database create failure.
func TestEmailOTPService_SendOTP_CreateError(t *testing.T) {
	t.Parallel()

	repo := &failingEmailOTPRepository{
		createError: errors.New("database connection failed"),
	}
	mockEmail := cryptoutilIdentityEmail.NewMockEmailService()
	service := cryptoutilIdentityMFA.NewEmailOTPService(repo, mockEmail)

	ctx := context.Background()
	userID := googleUuid.New()
	email := testUserEmail

	err := service.SendOTP(ctx, userID, email)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create OTP record")
}

// TestEmailOTPService_SendOTP_EmailSendError tests email sending failure.
func TestEmailOTPService_SendOTP_EmailSendError(t *testing.T) {
	t.Parallel()

	repo := newMockEmailOTPRepository()
	failEmail := &failingEmailService{
		sendError: errors.New("SMTP server unavailable"),
	}
	service := cryptoutilIdentityMFA.NewEmailOTPService(repo, failEmail)

	ctx := context.Background()
	userID := googleUuid.New()
	email := testUserEmail

	err := service.SendOTP(ctx, userID, email)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to send email")
}

// TestEmailOTPService_VerifyOTP_NotFound tests OTP not found scenario.
func TestEmailOTPService_VerifyOTP_NotFound(t *testing.T) {
	t.Parallel()

	repo := newMockEmailOTPRepository()
	mockEmail := cryptoutilIdentityEmail.NewMockEmailService()
	service := cryptoutilIdentityMFA.NewEmailOTPService(repo, mockEmail)

	ctx := context.Background()
	userID := googleUuid.New() // User with no OTP.

	err := service.VerifyOTP(ctx, userID, "123456")
	require.Error(t, err)
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrInvalidOTP)
}

// TestEmailOTPService_VerifyOTP_UpdateError tests database update failure.
func TestEmailOTPService_VerifyOTP_UpdateError(t *testing.T) {
	t.Parallel()

	repo := newMockEmailOTPRepository()
	mockEmail := cryptoutilIdentityEmail.NewMockEmailService()
	service := cryptoutilIdentityMFA.NewEmailOTPService(repo, mockEmail)

	ctx := context.Background()
	userID := googleUuid.New()
	email := testUserEmail

	// Send OTP successfully.
	err := service.SendOTP(ctx, userID, email)
	require.NoError(t, err)

	// Extract OTP.
	lastEmail := mockEmail.GetLastEmail()
	otpCode, found := mockEmail.ContainsOTP(lastEmail)
	require.True(t, found)

	// Replace repository with failing version by creating wrapper.
	otp, err := repo.GetByUserID(ctx, userID)
	require.NoError(t, err)

	// Create wrapper repository that fails on update.
	wrapperRepo := &mockEmailOTPRepositoryWithFailingUpdate{
		underlying: repo,
		otp:        otp,
		updateErr:  errors.New("database update failed"),
	}

	// Create service with wrapper repo.
	serviceWithFailRepo := cryptoutilIdentityMFA.NewEmailOTPService(wrapperRepo, mockEmail)

	// Verify OTP (should fail on update).
	err = serviceWithFailRepo.VerifyOTP(ctx, userID, otpCode)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to mark OTP as used")
}

// mockEmailOTPRepositoryWithFailingUpdate wraps repo with failing update.
type mockEmailOTPRepositoryWithFailingUpdate struct {
	underlying *mockEmailOTPRepository
	otp        *cryptoutilIdentityDomain.EmailOTP
	updateErr  error
}

func (m *mockEmailOTPRepositoryWithFailingUpdate) Create(ctx context.Context, otp *cryptoutilIdentityDomain.EmailOTP) error {
	return m.underlying.Create(ctx, otp)
}

func (m *mockEmailOTPRepositoryWithFailingUpdate) GetByUserID(_ context.Context, _ googleUuid.UUID) (*cryptoutilIdentityDomain.EmailOTP, error) {
	return m.otp, nil
}

func (m *mockEmailOTPRepositoryWithFailingUpdate) Update(ctx context.Context, otp *cryptoutilIdentityDomain.EmailOTP) error {
	if m.updateErr != nil {
		return m.updateErr
	}

	return m.underlying.Update(ctx, otp)
}

// TestRecoveryCodeService_Verify_AllCodesUsed tests all codes used scenario.
func TestRecoveryCodeService_Verify_AllCodesUsed(t *testing.T) {
	t.Parallel()

	repo := newMockRecoveryCodeRepository()
	service := cryptoutilIdentityMFA.NewRecoveryCodeService(repo)

	ctx := context.Background()
	userID := googleUuid.New()

	// Generate codes.
	plaintextCodes, err := service.GenerateForUser(ctx, userID, 3)
	require.NoError(t, err)
	require.Len(t, plaintextCodes, 3)

	// Use all codes.
	for _, code := range plaintextCodes {
		err := service.Verify(ctx, userID, code)
		require.NoError(t, err)
	}

	// Try to verify any code again (all used).
	err = service.Verify(ctx, userID, plaintextCodes[0])
	require.Error(t, err)
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrRecoveryCodeNotFound)
}

// TestRecoveryCodeService_Verify_AllCodesExpired tests all codes expired scenario.
func TestRecoveryCodeService_Verify_AllCodesExpired(t *testing.T) {
	t.Parallel()

	repo := newMockRecoveryCodeRepository()
	service := cryptoutilIdentityMFA.NewRecoveryCodeService(repo)

	ctx := context.Background()
	userID := googleUuid.New()

	// Generate codes.
	plaintextCodes, err := service.GenerateForUser(ctx, userID, 3)
	require.NoError(t, err)

	// Manually expire all codes.
	codes, err := repo.GetByUserID(ctx, userID)
	require.NoError(t, err)

	for _, code := range codes {
		code.ExpiresAt = time.Now().Add(-1 * time.Hour)
		err := repo.Update(ctx, code)
		require.NoError(t, err)
	}

	// Try to verify (all expired).
	err = service.Verify(ctx, userID, plaintextCodes[0])
	require.Error(t, err)
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrRecoveryCodeNotFound)
}

// mockRecoveryCodeRepository for testing recovery code service.
type mockRecoveryCodeRepository struct {
	codes map[string][]*cryptoutilIdentityDomain.RecoveryCode
}

func newMockRecoveryCodeRepository() *mockRecoveryCodeRepository {
	return &mockRecoveryCodeRepository{
		codes: make(map[string][]*cryptoutilIdentityDomain.RecoveryCode),
	}
}

func (m *mockRecoveryCodeRepository) Create(_ context.Context, code *cryptoutilIdentityDomain.RecoveryCode) error {
	userID := code.UserID.String()
	m.codes[userID] = append(m.codes[userID], code)

	return nil
}

func (m *mockRecoveryCodeRepository) CreateBatch(ctx context.Context, codes []*cryptoutilIdentityDomain.RecoveryCode) error {
	for _, code := range codes {
		if err := m.Create(ctx, code); err != nil {
			return err
		}
	}

	return nil
}

func (m *mockRecoveryCodeRepository) GetByUserID(_ context.Context, userID googleUuid.UUID) ([]*cryptoutilIdentityDomain.RecoveryCode, error) {
	codes, exists := m.codes[userID.String()]
	if !exists {
		return nil, cryptoutilIdentityAppErr.ErrRecoveryCodeNotFound
	}

	return codes, nil
}

func (m *mockRecoveryCodeRepository) Update(_ context.Context, code *cryptoutilIdentityDomain.RecoveryCode) error {
	userID := code.UserID.String()
	codes := m.codes[userID]

	for i, c := range codes {
		if c.ID == code.ID {
			codes[i] = code

			return nil
		}
	}

	return errors.New("code not found for update")
}

func (m *mockRecoveryCodeRepository) DeleteByUserID(_ context.Context, userID googleUuid.UUID) error {
	delete(m.codes, userID.String())

	return nil
}

func (m *mockRecoveryCodeRepository) CountUnused(_ context.Context, userID googleUuid.UUID) (int64, error) {
	codes, exists := m.codes[userID.String()]
	if !exists {
		return 0, nil
	}

	count := int64(0)

	for _, code := range codes {
		if !code.IsUsed() && !code.IsExpired() {
			count++
		}
	}

	return count, nil
}
