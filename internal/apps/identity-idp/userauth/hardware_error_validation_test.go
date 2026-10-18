// Copyright (c) 2025-2026 Justin Cranford.
//
//

package userauth

import (
	"context"
	"errors"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
)

// TestNewHardwareErrorValidator tests validator creation.
func TestNewHardwareErrorValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		maxPINRetries      int
		authTimeout        time.Duration
		devicePollInterval time.Duration
		wantErr            bool
	}{
		{
			name:               "valid configuration",
			maxPINRetries:      value3,
			authTimeout:        cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days * time.Second,
			devicePollInterval: 1 * time.Second,
			wantErr:            false,
		},
		{
			name:               "invalid maxPINRetries (zero)",
			maxPINRetries:      0,
			authTimeout:        cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days * time.Second,
			devicePollInterval: 1 * time.Second,
			wantErr:            true,
		},
		{
			name:               "invalid authTimeout (zero)",
			maxPINRetries:      value3,
			authTimeout:        0,
			devicePollInterval: 1 * time.Second,
			wantErr:            true,
		},
		{
			name:               "invalid devicePollInterval (zero)",
			maxPINRetries:      value3,
			authTimeout:        cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days * time.Second,
			devicePollInterval: 0,
			wantErr:            true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			validator, err := NewHardwareErrorValidator(tc.maxPINRetries, tc.authTimeout, tc.devicePollInterval)

			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, validator)
			} else {
				require.NoError(t, err)
				require.NotNil(t, validator)
				require.Equal(t, tc.maxPINRetries, validator.maxPINRetries)
				require.Equal(t, tc.authTimeout, validator.authTimeout)
				require.Equal(t, tc.devicePollInterval, validator.devicePollInterval)
			}
		})
	}
}

// TestValidateAuthentication tests authentication validation with timeout.
func TestValidateAuthentication(t *testing.T) {
	t.Parallel()

	validator, err := NewHardwareErrorValidator(value3, cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days*time.Second, cryptoutilSharedMagic.JoseJAMaxMaterials*time.Millisecond)
	require.NoError(t, err)

	tests := []struct {
		name            string
		authFunc        func(context.Context) error
		wantErrContains string
	}{
		{
			name: "successful authentication",
			authFunc: func(_ context.Context) error {
				return nil
			},
			wantErrContains: "",
		},
		{
			name: testDeviceRemovedError,
			authFunc: func(_ context.Context) error {
				return ErrDeviceRemoved
			},
			wantErrContains: testHardwareDeviceRemoved,
		},
		{
			name: testPinRetryExhausted,
			authFunc: func(_ context.Context) error {
				return ErrPINRetryExhausted
			},
			wantErrContains: valuePinRetryLimitExhausted,
		},
		{
			name: testAuthenticationTimeout,
			authFunc: func(_ context.Context) error {
				time.Sleep(cryptoutilSharedMagic.IMMaxUsernameLength * time.Millisecond)

				return ErrAuthenticationTimeout
			},
			wantErrContains: testAuthenticationTimeout,
		},
		{
			name: "device unresponsive",
			authFunc: func(_ context.Context) error {
				return ErrDeviceUnresponsive
			},
			wantErrContains: "hardware device unresponsive",
		},
		{
			name: testInvalidPin,
			authFunc: func(_ context.Context) error {
				return ErrInvalidPIN
			},
			wantErrContains: testInvalidPin,
		},
		{
			name: "device locked",
			authFunc: func(_ context.Context) error {
				return ErrDeviceLocked
			},
			wantErrContains: valueHardwareDeviceLocked,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			err := validator.ValidateAuthentication(ctx, tc.authFunc)

			if tc.wantErrContains == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErrContains)
			}
		})
	}
}

func TestValidateAuthentication_Timeout(t *testing.T) {
	t.Parallel()

	validator, err := NewHardwareErrorValidator(
		value3,
		cryptoutilSharedMagic.JoseJAMaxMaterials*time.Millisecond,
		time.Second,
	)
	require.NoError(t, err)

	operationDone := make(chan struct{})
	err = validator.ValidateAuthentication(context.Background(), func(context.Context) error {
		<-operationDone

		return nil
	})

	close(operationDone)

	require.ErrorContains(t, err, testAuthenticationTimeout)
}

// TestRetryWithBackoff tests retry logic with exponential backoff.
func TestRetryWithBackoff(t *testing.T) {
	t.Parallel()

	validator, err := NewHardwareErrorValidator(value3, cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days*time.Second, cryptoutilSharedMagic.JoseJADefaultMaxMaterials*time.Millisecond)
	require.NoError(t, err)

	tests := []struct {
		name            string
		maxRetries      int
		operation       func(context.Context) error
		wantErrContains string
	}{
		{
			name:       "successful on first attempt",
			maxRetries: value3,
			operation: func(_ context.Context) error {
				return nil
			},
			wantErrContains: "",
		},
		{
			name:       "non-retriable error (PIN exhausted)",
			maxRetries: value3,
			operation: func(_ context.Context) error {
				return ErrPINRetryExhausted
			},
			wantErrContains: valuePinRetryLimitExhausted,
		},
		{
			name:       "non-retriable error (device locked)",
			maxRetries: value3,
			operation: func(_ context.Context) error {
				return ErrDeviceLocked
			},
			wantErrContains: valueHardwareDeviceLocked,
		},
		{
			name:       testMaxRetriesExhausted,
			maxRetries: value3,
			operation: func(_ context.Context) error {
				return ErrDeviceUnresponsive
			},
			wantErrContains: testMaxRetriesExhausted,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			err := validator.RetryWithBackoff(ctx, tc.maxRetries, tc.operation)

			if tc.wantErrContains == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErrContains)
			}
		})
	}
}

func TestRetryWithBackoff_CancelledContext(t *testing.T) {
	t.Parallel()

	validator, err := NewHardwareErrorValidator(value3, time.Second, time.Second)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = validator.RetryWithBackoff(ctx, value3, func(context.Context) error {
		return ErrDeviceUnresponsive
	})

	require.ErrorContains(t, err, "retry cancelled")
}

// TestMonitorDevicePresence tests device presence monitoring.
func TestMonitorDevicePresence(t *testing.T) {
	t.Parallel()

	validator, err := NewHardwareErrorValidator(value3, cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days*time.Second, cryptoutilSharedMagic.IMMaxUsernameLength*time.Millisecond)
	require.NoError(t, err)

	tests := []struct {
		name            string
		checkFunc       func(context.Context) error
		contextCancel   bool
		wantErrContains string
	}{
		{
			name: "device present",
			checkFunc: func(_ context.Context) error {
				return nil
			},
			contextCancel:   false,
			wantErrContains: "",
		},
		{
			name: "device removed",
			checkFunc: func(_ context.Context) error {
				return ErrDeviceRemoved
			},
			contextCancel:   false,
			wantErrContains: testHardwareDeviceRemoved,
		},
		{
			name: "context cancelled",
			checkFunc: func(_ context.Context) error {
				return nil
			},
			contextCancel:   true,
			wantErrContains: "device monitoring cancelled",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			if tc.contextCancel {
				cancelCtx, cancel := context.WithCancel(ctx)
				cancel()

				ctx = cancelCtx
			}

			err := validator.MonitorDevicePresence(ctx, tc.checkFunc)

			if tc.wantErrContains == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErrContains)
			}
		})
	}
}

func TestMonitorDevicePresence_TransientError(t *testing.T) {
	t.Parallel()

	validator, err := NewHardwareErrorValidator(value3, time.Second, time.Millisecond)
	require.NoError(t, err)

	attempts := 0
	err = validator.MonitorDevicePresence(context.Background(), func(context.Context) error {
		attempts++
		if attempts == 1 {
			return errors.New("temporary device query failure")
		}

		return nil
	})

	require.NoError(t, err)
	require.Equal(t, 2, attempts)
}

// TestClassifyError tests error classification.
func TestClassifyError(t *testing.T) {
	t.Parallel()

	validator, err := NewHardwareErrorValidator(value3, cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days*time.Second, cryptoutilSharedMagic.JoseJAMaxMaterials*time.Millisecond)
	require.NoError(t, err)

	tests := []struct {
		name            string
		inputErr        error
		wantErrContains string
	}{
		{
			name:            "nil error",
			inputErr:        nil,
			wantErrContains: "",
		},
		{
			name:            testDeviceRemovedError,
			inputErr:        ErrDeviceRemoved,
			wantErrContains: testHardwareDeviceRemoved,
		},
		{
			name:            testPinRetryExhausted,
			inputErr:        ErrPINRetryExhausted,
			wantErrContains: valuePinRetryLimitExhausted,
		},
		{
			name:            "unknown hardware error",
			inputErr:        errors.New("unknown hardware failure"),
			wantErrContains: "hardware authentication error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := validator.classifyError(tc.inputErr)

			if tc.wantErrContains == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErrContains)
				require.True(t, errors.Is(err, cryptoutilIdentityAppErr.ErrAuthenticationFailed), "Should wrap ErrAuthenticationFailed")
			}
		})
	}
}
