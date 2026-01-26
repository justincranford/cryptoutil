// Copyright (c) 2025 Justin Cranford
//
//

//go:build e2e

package e2e

import (
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuth "cryptoutil/internal/identity/idp/auth"
)

// MockOTPSecretStore implements OTPSecretStore for testing.
type MockOTPSecretStore struct {
	totpSecrets  map[string]string
	emailSecrets map[string]string
	smsSecrets   map[string]string
}

// NewMockOTPSecretStore creates a new mock OTP secret store.
func NewMockOTPSecretStore() *MockOTPSecretStore {
	return &MockOTPSecretStore{
		totpSecrets:  make(map[string]string),
		emailSecrets: make(map[string]string),
		smsSecrets:   make(map[string]string),
	}
}

// GetTOTPSecret retrieves TOTP secret for a user.
func (m *MockOTPSecretStore) GetTOTPSecret(_ context.Context, userID string) (string, error) {
	if secret, ok := m.totpSecrets[userID]; ok {
		return secret, nil
	}

	return "", nil
}

// GetEmailOTPSecret retrieves email OTP secret for a user.
func (m *MockOTPSecretStore) GetEmailOTPSecret(_ context.Context, userID string) (string, error) {
	if secret, ok := m.emailSecrets[userID]; ok {
		return secret, nil
	}

	return "", nil
}

// GetSMSOTPSecret retrieves SMS OTP secret for a user.
func (m *MockOTPSecretStore) GetSMSOTPSecret(_ context.Context, userID string) (string, error) {
	if secret, ok := m.smsSecrets[userID]; ok {
		return secret, nil
	}

	return "", nil
}

// TestTOTPValidation tests TOTP code validation.
func TestTOTPValidation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	secretStore := NewMockOTPSecretStore()
	validator := cryptoutilIdentityAuth.NewTOTPValidator(secretStore)

	userID := googleUuid.Must(googleUuid.NewV7()).String()

	// Generate TOTP secret.
	// cspell:disable-next-line
	secret := "JBSWY3DPEHPK3PXP" // Base32-encoded secret.
	secretStore.totpSecrets[userID] = secret

	t.Run("Valid_TOTP_Code", func(t *testing.T) {
		t.Parallel()
		// Generate current TOTP code.
		code, err := totp.GenerateCode(secret, time.Now().UTC())
		require.NoError(t, err)

		// Validate code.
		valid, err := validator.ValidateTOTP(ctx, userID, code)
		require.NoError(t, err)
		require.True(t, valid, "Valid TOTP code should pass validation")
	})

	t.Run("Invalid_TOTP_Code", func(t *testing.T) {
		t.Parallel()
		// Use invalid code.
		invalidCode := "000000"

		valid, err := validator.ValidateTOTP(ctx, userID, invalidCode)
		require.NoError(t, err)
		require.False(t, valid, "Invalid TOTP code should fail validation")
	})

	t.Run("TOTP_With_Time_Window", func(t *testing.T) {
		t.Parallel()
		// Generate code for 30 seconds ago (outside standard window).
		pastTime := time.Now().UTC().Add(-30 * time.Second)
		pastCode, err := totp.GenerateCode(secret, pastTime)
		require.NoError(t, err)

		// Validate with windowSize=1 (allows 30s before/after).
		valid, err := validator.ValidateTOTPWithWindow(ctx, userID, pastCode, 1)
		require.NoError(t, err)
		require.True(t, valid, "Past code should be valid with time window")
	})
}

// TestEmailOTPValidation tests email OTP validation.
func TestEmailOTPValidation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	secretStore := NewMockOTPSecretStore()
	validator := cryptoutilIdentityAuth.NewTOTPValidator(secretStore)

	userID := googleUuid.Must(googleUuid.NewV7()).String()

	// Generate email OTP secret.
	// cspell:disable-next-line
	secret := "KBSWY3DPEHPK3PXQ"
	secretStore.emailSecrets[userID] = secret

	t.Run("Valid_Email_OTP", func(t *testing.T) {
		t.Parallel()
		// Generate current email OTP code (5-minute period).
		code, err := totp.GenerateCode(secret, time.Now().UTC())
		require.NoError(t, err)

		valid, err := validator.ValidateEmailOTP(ctx, userID, code)
		require.NoError(t, err)
		require.True(t, valid, "Valid email OTP should pass validation")
	})

	t.Run("Invalid_Email_OTP", func(t *testing.T) {
		t.Parallel()

		invalidCode := "111111"

		valid, err := validator.ValidateEmailOTP(ctx, userID, invalidCode)
		require.NoError(t, err)
		require.False(t, valid, "Invalid email OTP should fail validation")
	})
}

// TestSMSOTPValidation tests SMS OTP validation.
func TestSMSOTPValidation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	secretStore := NewMockOTPSecretStore()
	validator := cryptoutilIdentityAuth.NewTOTPValidator(secretStore)

	userID := googleUuid.Must(googleUuid.NewV7()).String()

	// Generate SMS OTP secret.
	// cspell:disable-next-line
	secret := "LBSWY3DPEHPK3PXR"
	secretStore.smsSecrets[userID] = secret

	t.Run("Valid_SMS_OTP", func(t *testing.T) {
		t.Parallel()
		// Generate current SMS OTP code (10-minute period).
		code, err := totp.GenerateCode(secret, time.Now().UTC())
		require.NoError(t, err)

		valid, err := validator.ValidateSMSOTP(ctx, userID, code)
		require.NoError(t, err)
		require.True(t, valid, "Valid SMS OTP should pass validation")
	})

	t.Run("Invalid_SMS_OTP", func(t *testing.T) {
		t.Parallel()

		invalidCode := "222222"

		valid, err := validator.ValidateSMSOTP(ctx, userID, invalidCode)
		require.NoError(t, err)
		require.False(t, valid, "Invalid SMS OTP should fail validation")
	})
}

// TestOTPConcurrency tests OTP validation under concurrent load.
func TestOTPConcurrency(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	secretStore := NewMockOTPSecretStore()
	validator := cryptoutilIdentityAuth.NewTOTPValidator(secretStore)

	const parallelValidations = 10

	t.Run("Concurrent_TOTP_Validation", func(t *testing.T) {
		t.Parallel()

		// cspell:disable-next-line
		secret := "MBSWY3DPEHPK3PXS"

		// Create 10 users with same TOTP secret.
		for i := 0; i < parallelValidations; i++ {
			userID := googleUuid.Must(googleUuid.NewV7()).String()
			secretStore.totpSecrets[userID] = secret
		}

		// Generate current code.
		code, err := totp.GenerateCode(secret, time.Now().UTC())
		require.NoError(t, err)

		// Validate concurrently from 10 goroutines.
		results := make(chan bool, parallelValidations)

		for i := 0; i < parallelValidations; i++ {
			go func(idx int) {
				userID := googleUuid.Must(googleUuid.NewV7()).String()
				secretStore.totpSecrets[userID] = secret

				valid, validationErr := validator.ValidateTOTP(ctx, userID, code)
				if validationErr == nil && valid {
					results <- true
				} else {
					results <- false
				}
			}(i)
		}

		// Collect results.
		successCount := 0

		for i := 0; i < parallelValidations; i++ {
			if <-results {
				successCount++
			}
		}

		require.Equal(t, parallelValidations, successCount, "All concurrent TOTP validations should succeed")
	})
}

// TestOTPExpiration tests OTP code expiration.
func TestOTPExpiration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	secretStore := NewMockOTPSecretStore()
	validator := cryptoutilIdentityAuth.NewTOTPValidator(secretStore)

	userID := googleUuid.Must(googleUuid.NewV7()).String()

	// cspell:disable-next-line
	secret := "NBSWY3DPEHPK3PXT"
	secretStore.totpSecrets[userID] = secret

	t.Run("Expired_TOTP_Code", func(t *testing.T) {
		t.Parallel()
		// Generate code for 2 minutes ago (outside all windows).
		expiredTime := time.Now().UTC().Add(-2 * time.Minute)
		expiredCode, err := totp.GenerateCode(secret, expiredTime)
		require.NoError(t, err)

		// Validate without time window.
		valid, err := validator.ValidateTOTP(ctx, userID, expiredCode)
		require.NoError(t, err)
		require.False(t, valid, "Expired TOTP code should fail validation")
	})

	t.Run("Expired_Code_With_Window", func(t *testing.T) {
		t.Parallel()
		// Generate code for 90 seconds ago.
		expiredTime := time.Now().UTC().Add(-90 * time.Second)
		expiredCode, err := totp.GenerateCode(secret, expiredTime)
		require.NoError(t, err)

		// Validate with windowSize=1 (allows 30s before/after = max 60s skew).
		valid, err := validator.ValidateTOTPWithWindow(ctx, userID, expiredCode, 1)
		require.NoError(t, err)
		require.False(t, valid, "Code beyond window should fail validation")
	})
}
