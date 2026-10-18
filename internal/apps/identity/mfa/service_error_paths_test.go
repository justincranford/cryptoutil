// Copyright (c) 2025-2026 Justin Cranford.
package mfa_test

import (
	"context"
	"errors"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityMfa "cryptoutil/internal/apps/identity/mfa"
	cryptoutilSharedCryptoPassword "cryptoutil/internal/shared/crypto/password"
)

const testLookupFailureCaseName = "lookup failure"

type errorEmailOTPRepository struct {
	createErr error
	getErr    error
	updateErr error
	otp       *cryptoutilIdentityDomain.EmailOTP
}

func (r *errorEmailOTPRepository) Create(_ context.Context, otp *cryptoutilIdentityDomain.EmailOTP) error {
	r.otp = otp

	return r.createErr
}

func (r *errorEmailOTPRepository) GetByUserID(_ context.Context, _ googleUuid.UUID) (*cryptoutilIdentityDomain.EmailOTP, error) {
	return r.otp, r.getErr
}

func (r *errorEmailOTPRepository) Update(_ context.Context, otp *cryptoutilIdentityDomain.EmailOTP) error {
	r.otp = otp

	return r.updateErr
}

type errorEmailService struct {
	err error
}

func (s errorEmailService) SendEmail(context.Context, string, string, string) error {
	return s.err
}

type errorRecoveryCodeRepository struct {
	createErr error
	getErr    error
	updateErr error
	deleteErr error
	countErr  error
	codes     []*cryptoutilIdentityDomain.RecoveryCode
}

func (r *errorRecoveryCodeRepository) Create(_ context.Context, code *cryptoutilIdentityDomain.RecoveryCode) error {
	r.codes = append(r.codes, code)

	return r.createErr
}

func (r *errorRecoveryCodeRepository) CreateBatch(_ context.Context, codes []*cryptoutilIdentityDomain.RecoveryCode) error {
	r.codes = codes

	return r.createErr
}

func (r *errorRecoveryCodeRepository) GetByUserID(context.Context, googleUuid.UUID) ([]*cryptoutilIdentityDomain.RecoveryCode, error) {
	return r.codes, r.getErr
}

func (r *errorRecoveryCodeRepository) Update(_ context.Context, code *cryptoutilIdentityDomain.RecoveryCode) error {
	r.codes = []*cryptoutilIdentityDomain.RecoveryCode{code}

	return r.updateErr
}

func (r *errorRecoveryCodeRepository) DeleteByUserID(context.Context, googleUuid.UUID) error {
	return r.deleteErr
}

func (r *errorRecoveryCodeRepository) CountUnused(context.Context, googleUuid.UUID) (int64, error) {
	return 0, r.countErr
}

func TestEmailOTPService_ErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		repo     *errorEmailOTPRepository
		emailErr error
		run      func(*testing.T, *cryptoutilIdentityMfa.EmailOTPService, *errorEmailOTPRepository, googleUuid.UUID)
	}{
		{
			name: "create failure",
			repo: &errorEmailOTPRepository{createErr: errors.New("create failed")},
			run: func(t *testing.T, service *cryptoutilIdentityMfa.EmailOTPService, _ *errorEmailOTPRepository, userID googleUuid.UUID) {
				t.Helper()
				require.ErrorContains(t, service.SendOTP(context.Background(), userID, testUserEmail), "failed to create OTP record")
			},
		},
		{
			name:     "email delivery failure",
			repo:     &errorEmailOTPRepository{},
			emailErr: errors.New("delivery failed"),
			run: func(t *testing.T, service *cryptoutilIdentityMfa.EmailOTPService, _ *errorEmailOTPRepository, userID googleUuid.UUID) {
				t.Helper()
				require.ErrorContains(t, service.SendOTP(context.Background(), userID, testUserEmail), "failed to send email")
			},
		},
		{
			name: testLookupFailureCaseName,
			repo: &errorEmailOTPRepository{getErr: errors.New("lookup failed")},
			run: func(t *testing.T, service *cryptoutilIdentityMfa.EmailOTPService, _ *errorEmailOTPRepository, userID googleUuid.UUID) {
				t.Helper()
				require.ErrorContains(t, service.VerifyOTP(context.Background(), userID, "123456"), "lookup failed")
			},
		},
		{
			name: "invalid stored hash",
			repo: &errorEmailOTPRepository{otp: &cryptoutilIdentityDomain.EmailOTP{
				ExpiresAt: time.Now().UTC().Add(time.Hour),
				CodeHash:  "invalid hash",
			}},
			run: func(t *testing.T, service *cryptoutilIdentityMfa.EmailOTPService, _ *errorEmailOTPRepository, userID googleUuid.UUID) {
				t.Helper()
				require.ErrorContains(t, service.VerifyOTP(context.Background(), userID, "123456"), "verification failed")
			},
		},
		{
			name: "update failure",
			repo: &errorEmailOTPRepository{
				updateErr: errors.New("update failed"),
				otp: &cryptoutilIdentityDomain.EmailOTP{
					ExpiresAt: time.Now().UTC().Add(time.Hour),
				},
			},
			run: func(t *testing.T, service *cryptoutilIdentityMfa.EmailOTPService, repo *errorEmailOTPRepository, userID googleUuid.UUID) {
				t.Helper()

				hash, err := cryptoutilSharedCryptoPassword.HashPassword("123456")
				require.NoError(t, err)

				repo.otp.CodeHash = hash

				require.ErrorContains(t, service.VerifyOTP(context.Background(), userID, "123456"), "failed to mark OTP as used")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			service := cryptoutilIdentityMfa.NewEmailOTPService(tc.repo, errorEmailService{err: tc.emailErr})
			tc.run(t, service, tc.repo, googleUuid.Must(googleUuid.NewV7()))
		})
	}
}

func TestRecoveryCodeService_ErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		repo *errorRecoveryCodeRepository
		run  func(*testing.T, *cryptoutilIdentityMfa.RecoveryCodeService, googleUuid.UUID)
	}{
		{
			name: "batch create failure",
			repo: &errorRecoveryCodeRepository{createErr: errors.New("create failed")},
			run: func(t *testing.T, service *cryptoutilIdentityMfa.RecoveryCodeService, userID googleUuid.UUID) {
				t.Helper()

				codes, err := service.GenerateForUser(context.Background(), userID, 1)
				require.ErrorContains(t, err, "failed to store recovery codes")
				require.Nil(t, codes)
			},
		},
		{
			name: testLookupFailureCaseName,
			repo: &errorRecoveryCodeRepository{getErr: errors.New("lookup failed")},
			run: func(t *testing.T, service *cryptoutilIdentityMfa.RecoveryCodeService, userID googleUuid.UUID) {
				t.Helper()
				require.ErrorContains(t, service.Verify(context.Background(), userID, "unused"), "failed to get recovery codes")
			},
		},
		{
			name: "mark used failure",
			repo: &errorRecoveryCodeRepository{updateErr: errors.New("update failed")},
			run: func(t *testing.T, service *cryptoutilIdentityMfa.RecoveryCodeService, userID googleUuid.UUID) {
				t.Helper()

				hash, err := cryptoutilSharedCryptoPassword.HashPassword("recovery-code")
				require.NoError(t, err)

				serviceRepo := &errorRecoveryCodeRepository{
					updateErr: errors.New("update failed"),
					codes: []*cryptoutilIdentityDomain.RecoveryCode{{
						UserID:    userID,
						CodeHash:  hash,
						ExpiresAt: time.Now().UTC().Add(time.Hour),
					}},
				}
				recoveryService := cryptoutilIdentityMfa.NewRecoveryCodeService(serviceRepo)
				require.ErrorContains(t, recoveryService.Verify(context.Background(), userID, "recovery-code"), "failed to mark recovery code as used")
			},
		},
		{
			name: "delete failure",
			repo: &errorRecoveryCodeRepository{deleteErr: errors.New("delete failed")},
			run: func(t *testing.T, service *cryptoutilIdentityMfa.RecoveryCodeService, userID googleUuid.UUID) {
				t.Helper()

				codes, err := service.RegenerateForUser(context.Background(), userID, 1)
				require.ErrorContains(t, err, "failed to delete old recovery codes")
				require.Nil(t, codes)
			},
		},
		{
			name: "count failure",
			repo: &errorRecoveryCodeRepository{countErr: errors.New("count failed")},
			run: func(t *testing.T, service *cryptoutilIdentityMfa.RecoveryCodeService, userID googleUuid.UUID) {
				t.Helper()

				count, err := service.GetRemainingCount(context.Background(), userID)
				require.ErrorContains(t, err, "failed to count unused recovery codes")
				require.Zero(t, count)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			service := cryptoutilIdentityMfa.NewRecoveryCodeService(tc.repo)
			tc.run(t, service, googleUuid.Must(googleUuid.NewV7()))
		})
	}
}
