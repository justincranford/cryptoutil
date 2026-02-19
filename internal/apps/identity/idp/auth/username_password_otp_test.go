// Copyright (c) 2025 Justin Cranford
//
//

package auth_test

import (
	"context"
	"fmt"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityAuth "cryptoutil/internal/apps/identity/idp/auth"
)

func TestNewOTPService(t *testing.T) {
	t.Parallel()

	service := cryptoutilIdentityAuth.NewOTPService()
	require.NotNil(t, service, "NewOTPService should return non-nil service")
}

// TestOTPService_GenerateOTP tests GenerateOTP (returns not implemented error).
func TestOTPService_GenerateOTP(t *testing.T) {
	t.Parallel()

	service := cryptoutilIdentityAuth.NewOTPService()
	user := &cryptoutilIdentityDomain.User{
		ID:                googleUuid.New(),
		Sub:               "testuser",
		PreferredUsername: "testuser",
		Email:             "test@example.com",
	}

	// Test with email method.
	otp, err := service.GenerateOTP(context.Background(), user, cryptoutilIdentityAuth.OTPMethodEmail)
	require.Error(t, err, "GenerateOTP should return error (not implemented)")
	require.Empty(t, otp, "OTP should be empty on error")

	// Test with SMS method.
	otp, err = service.GenerateOTP(context.Background(), user, cryptoutilIdentityAuth.OTPMethodSMS)
	require.Error(t, err, "GenerateOTP should return error (not implemented)")
	require.Empty(t, otp, "OTP should be empty on error")
}

// TestOTPService_ValidateOTP tests ValidateOTP (returns not implemented error).
func TestOTPService_ValidateOTP(t *testing.T) {
	t.Parallel()

	service := cryptoutilIdentityAuth.NewOTPService()
	user := &cryptoutilIdentityDomain.User{
		ID:                googleUuid.New(),
		Sub:               "testuser",
		PreferredUsername: "testuser",
		Email:             "test@example.com",
	}

	// Test with email method.
	err := service.ValidateOTP(context.Background(), user, "123456", cryptoutilIdentityAuth.OTPMethodEmail)
	require.Error(t, err, "ValidateOTP should return error (not implemented)")

	// Test with SMS method.
	err = service.ValidateOTP(context.Background(), user, "654321", cryptoutilIdentityAuth.OTPMethodSMS)
	require.Error(t, err, "ValidateOTP should return error (not implemented)")
}

// ---------------------- TOTPValidator Tests ----------------------

// mockOTPSecretStore is a mock implementation of OTPSecretStore.
type mockOTPSecretStore struct {
	totpSecret  string
	emailSecret string
	smsSecret   string
	err         error
}

func (m *mockOTPSecretStore) GetTOTPSecret(_ context.Context, _ string) (string, error) {
	if m.err != nil {
		return "", m.err
	}

	return m.totpSecret, nil
}

func (m *mockOTPSecretStore) GetEmailOTPSecret(_ context.Context, _ string) (string, error) {
	if m.err != nil {
		return "", m.err
	}

	return m.emailSecret, nil
}

func (m *mockOTPSecretStore) GetSMSOTPSecret(_ context.Context, _ string) (string, error) {
	if m.err != nil {
		return "", m.err
	}

	return m.smsSecret, nil
}

// TestNewTOTPValidator tests TOTPValidator creation.
func TestNewTOTPValidator(t *testing.T) {
	t.Parallel()

	store := &mockOTPSecretStore{}
	validator := cryptoutilIdentityAuth.NewTOTPValidator(store)
	require.NotNil(t, validator, "NewTOTPValidator should return non-nil validator")
}

// TestTOTPValidator_ValidateTOTP tests TOTP validation.
func TestTOTPValidator_ValidateTOTP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		secret    string
		code      string
		storeErr  error
		wantValid bool
		wantErr   bool
	}{
		{
			name:      "invalid code with valid secret",
			secret:    "JBSWY3DPEHPK3PXP", // Base32-encoded secret
			code:      "000000",
			wantValid: false,
			wantErr:   false,
		},
		{
			name:     "store error",
			secret:   "",
			code:     "123456",
			storeErr: fmt.Errorf("store error"),
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := &mockOTPSecretStore{
				totpSecret: tc.secret,
				err:        tc.storeErr,
			}
			validator := cryptoutilIdentityAuth.NewTOTPValidator(store)

			valid, err := validator.ValidateTOTP(context.Background(), "user123", tc.code)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantValid, valid)
			}
		})
	}
}

// TestTOTPValidator_ValidateTOTPWithWindow tests TOTP validation with window.
func TestTOTPValidator_ValidateTOTPWithWindow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		secret     string
		code       string
		windowSize uint
		storeErr   error
		wantErr    bool
	}{
		{
			name:       "invalid code with valid secret",
			secret:     "JBSWY3DPEHPK3PXP",
			code:       "000000",
			windowSize: 1,
			wantErr:    false,
		},
		{
			name:       "store error",
			secret:     "",
			code:       "123456",
			windowSize: 1,
			storeErr:   fmt.Errorf("store error"),
			wantErr:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := &mockOTPSecretStore{
				totpSecret: tc.secret,
				err:        tc.storeErr,
			}
			validator := cryptoutilIdentityAuth.NewTOTPValidator(store)

			_, err := validator.ValidateTOTPWithWindow(context.Background(), "user123", tc.code, tc.windowSize)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestTOTPValidator_ValidateEmailOTP tests email OTP validation.
func TestTOTPValidator_ValidateEmailOTP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		secret   string
		code     string
		storeErr error
		wantErr  bool
	}{
		{
			name:    "invalid code with valid secret",
			secret:  "JBSWY3DPEHPK3PXP",
			code:    "000000",
			wantErr: false,
		},
		{
			name:     "store error",
			secret:   "",
			code:     "123456",
			storeErr: fmt.Errorf("store error"),
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := &mockOTPSecretStore{
				emailSecret: tc.secret,
				err:         tc.storeErr,
			}
			validator := cryptoutilIdentityAuth.NewTOTPValidator(store)

			_, err := validator.ValidateEmailOTP(context.Background(), "user123", tc.code)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestTOTPValidator_ValidateSMSOTP tests SMS OTP validation.
func TestTOTPValidator_ValidateSMSOTP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		secret   string
		code     string
		storeErr error
		wantErr  bool
	}{
		{
			name:    "invalid code with valid secret",
			secret:  "JBSWY3DPEHPK3PXP",
			code:    "000000",
			wantErr: false,
		},
		{
			name:     "store error",
			secret:   "",
			code:     "123456",
			storeErr: fmt.Errorf("store error"),
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := &mockOTPSecretStore{
				smsSecret: tc.secret,
				err:       tc.storeErr,
			}
			validator := cryptoutilIdentityAuth.NewTOTPValidator(store)

			_, err := validator.ValidateSMSOTP(context.Background(), "user123", tc.code)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestMFAOrchestrator_NewMFAOrchestrator tests NewMFAOrchestrator.
func TestMFAOrchestrator_NewMFAOrchestrator(t *testing.T) {
	t.Parallel()

	orchestrator := cryptoutilIdentityAuth.NewMFAOrchestrator(nil, nil, nil, nil, nil)
	require.NotNil(t, orchestrator, "NewMFAOrchestrator should return non-nil orchestrator")
}
