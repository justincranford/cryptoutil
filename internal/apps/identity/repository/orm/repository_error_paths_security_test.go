// Copyright (c) 2025-2026 Justin Cranford.
package orm

import (
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestSecurityRepositories_DatabaseErrors(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	sqlDB, err := testDB.db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	ctx := context.Background()
	id := googleUuid.Must(googleUuid.NewV7())
	consent := NewConsentDecisionRepository(testDB.db)
	device := NewDeviceAuthorizationRepository(testDB.db)
	emailOTP := NewEmailOTPRepository(testDB.db)
	key := NewKeyRepository(testDB.db)
	mfaFactor := NewMFAFactorRepository(testDB.db)
	pushedRequest := NewPushedAuthorizationRequestRepository(testDB.db)
	recoveryCode := NewRecoveryCodeRepository(testDB.db)

	tests := []struct {
		name      string
		operation func() error
	}{
		{"consent_create", func() error { return consent.Create(ctx, &cryptoutilIdentityDomain.ConsentDecision{}) }},
		{"consent_get_by_id", func() error {
			_, err := consent.GetByID(ctx, id)

			return err
		}},
		{"consent_get_by_user_client_scope", func() error {
			_, err := consent.GetByUserClientScope(ctx, id, id.String(), cryptoutilSharedMagic.ParamScope)

			return err
		}},
		{"consent_update", func() error { return consent.Update(ctx, &cryptoutilIdentityDomain.ConsentDecision{}) }},
		{"consent_delete", func() error { return consent.Delete(ctx, id) }},
		{"consent_revoke_by_id", func() error { return consent.RevokeByID(ctx, id) }},
		{"consent_delete_expired", func() error {
			_, err := consent.DeleteExpired(ctx)

			return err
		}},
		{"device_create", func() error { return device.Create(ctx, &cryptoutilIdentityDomain.DeviceAuthorization{}) }},
		{"device_get_by_device_code", func() error {
			_, err := device.GetByDeviceCode(ctx, "device-code")

			return err
		}},
		{"device_get_by_user_code", func() error {
			_, err := device.GetByUserCode(ctx, "user-code")

			return err
		}},
		{"device_get_by_id", func() error {
			_, err := device.GetByID(ctx, id)

			return err
		}},
		{"device_update", func() error { return device.Update(ctx, &cryptoutilIdentityDomain.DeviceAuthorization{}) }},
		{"device_delete_expired", func() error { return device.DeleteExpired(ctx) }},
		{"email_otp_create", func() error { return emailOTP.Create(ctx, &cryptoutilIdentityDomain.EmailOTP{}) }},
		{"email_otp_get_by_user_id", func() error {
			_, err := emailOTP.GetByUserID(ctx, id)

			return err
		}},
		{"email_otp_get_by_id", func() error {
			_, err := emailOTP.GetByID(ctx, id)

			return err
		}},
		{"email_otp_update", func() error { return emailOTP.Update(ctx, &cryptoutilIdentityDomain.EmailOTP{}) }},
		{"email_otp_delete_by_user_id", func() error { return emailOTP.DeleteByUserID(ctx, id) }},
		{"email_otp_delete_expired", func() error {
			_, err := emailOTP.DeleteExpired(ctx)

			return err
		}},
		{"key_create", func() error { return key.Create(ctx, &cryptoutilIdentityDomain.Key{}) }},
		{"key_find_by_id", func() error {
			_, err := key.FindByID(ctx, id)

			return err
		}},
		{"key_find_by_usage", func() error {
			_, err := key.FindByUsage(ctx, cryptoutilSharedMagic.KeyUsageSigning, true)

			return err
		}},
		{"key_update", func() error { return key.Update(ctx, &cryptoutilIdentityDomain.Key{}) }},
		{"key_delete", func() error { return key.Delete(ctx, id) }},
		{"key_list", func() error {
			_, err := key.List(ctx, 0, 1)

			return err
		}},
		{"key_count", func() error {
			_, err := key.Count(ctx)

			return err
		}},
		{"mfa_factor_create", func() error { return mfaFactor.Create(ctx, &cryptoutilIdentityDomain.MFAFactor{}) }},
		{"mfa_factor_get_by_id", func() error {
			_, err := mfaFactor.GetByID(ctx, id)

			return err
		}},
		{"mfa_factor_get_by_auth_profile_id", func() error {
			_, err := mfaFactor.GetByAuthProfileID(ctx, id)

			return err
		}},
		{"mfa_factor_update", func() error { return mfaFactor.Update(ctx, &cryptoutilIdentityDomain.MFAFactor{}) }},
		{"mfa_factor_delete", func() error { return mfaFactor.Delete(ctx, id) }},
		{"mfa_factor_list", func() error {
			_, err := mfaFactor.List(ctx, 0, 1)

			return err
		}},
		{"mfa_factor_count", func() error {
			_, err := mfaFactor.Count(ctx)

			return err
		}},
		{"pushed_request_create", func() error { return pushedRequest.Create(ctx, &cryptoutilIdentityDomain.PushedAuthorizationRequest{}) }},
		{"pushed_request_get_by_uri", func() error {
			_, err := pushedRequest.GetByRequestURI(ctx, "request-uri")

			return err
		}},
		{"pushed_request_get_by_id", func() error {
			_, err := pushedRequest.GetByID(ctx, id)

			return err
		}},
		{"pushed_request_update", func() error { return pushedRequest.Update(ctx, &cryptoutilIdentityDomain.PushedAuthorizationRequest{}) }},
		{"pushed_request_delete_expired", func() error {
			_, err := pushedRequest.DeleteExpired(ctx)

			return err
		}},
		{"recovery_code_create", func() error { return recoveryCode.Create(ctx, &cryptoutilIdentityDomain.RecoveryCode{}) }},
		{"recovery_code_create_batch", func() error { return recoveryCode.CreateBatch(ctx, []*cryptoutilIdentityDomain.RecoveryCode{{}}) }},
		{"recovery_code_get_by_user_id", func() error {
			_, err := recoveryCode.GetByUserID(ctx, id)

			return err
		}},
		{"recovery_code_get_by_id", func() error {
			_, err := recoveryCode.GetByID(ctx, id)

			return err
		}},
		{"recovery_code_update", func() error { return recoveryCode.Update(ctx, &cryptoutilIdentityDomain.RecoveryCode{}) }},
		{"recovery_code_delete_by_user_id", func() error { return recoveryCode.DeleteByUserID(ctx, id) }},
		{"recovery_code_delete_expired", func() error {
			_, err := recoveryCode.DeleteExpired(ctx)

			return err
		}},
		{"recovery_code_count_unused", func() error {
			_, err := recoveryCode.CountUnused(ctx, id)

			return err
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Error(t, tc.operation())
		})
	}
}
