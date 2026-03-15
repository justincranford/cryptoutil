// Copyright (c) 2025 Justin Cranford
//
//

package userauth

import (
	"context"
	"errors"
	"fmt"
	"time"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
)

// Hardware authentication error types.
var (
	// ErrDeviceRemoved indicates the hardware device was removed during authentication.
	ErrDeviceRemoved = errors.New("hardware device removed during authentication")

	// ErrPINRetryExhausted indicates PIN retry limit exceeded.
	ErrPINRetryExhausted = errors.New("PIN retry limit exhausted")

	// ErrAuthenticationTimeout indicates hardware authentication timed out.
	ErrAuthenticationTimeout = errors.New("hardware authentication timeout")

	// ErrDeviceUnresponsive indicates the hardware device is not responding.
	ErrDeviceUnresponsive = errors.New("hardware device unresponsive")

	// ErrInvalidPIN indicates the provided PIN is invalid.
	ErrInvalidPIN = errors.New("invalid PIN provided")

	// ErrDeviceLocked indicates the hardware device is locked due to too many failed attempts.
	ErrDeviceLocked = errors.New("hardware device locked")
)

// HardwareErrorValidator validates hardware authentication errors and determines appropriate responses.
type HardwareErrorValidator struct {
	maxPINRetries      int
	authTimeout        time.Duration
	devicePollInterval time.Duration
}

// NewHardwareErrorValidator creates a new hardware error validator.
func NewHardwareErrorValidator(maxPINRetries int, authTimeout time.Duration, devicePollInterval time.Duration) (*HardwareErrorValidator, error) {
	if maxPINRetries <= 0 {
		return nil, fmt.Errorf("maxPINRetries must be positive, got: %d", maxPINRetries)
	}

	if authTimeout <= 0 {
		return nil, fmt.Errorf("authTimeout must be positive, got: %s", authTimeout)
	}

	if devicePollInterval <= 0 {
		return nil, fmt.Errorf("devicePollInterval must be positive, got: %s", devicePollInterval)
	}

	return &HardwareErrorValidator{
		maxPINRetries:      maxPINRetries,
		authTimeout:        authTimeout,
		devicePollInterval: devicePollInterval,
	}, nil
}

// ValidateAuthentication performs hardware authentication with error handling.
func (v *HardwareErrorValidator) ValidateAuthentication(ctx context.Context, authFunc func(context.Context) error) error {
	// Create timeout context.
	timeoutCtx, cancel := context.WithTimeout(ctx, v.authTimeout)
	defer cancel()

	// Execute authentication with timeout.
	errChan := make(chan error, 1)

	go func() {
		errChan <- authFunc(timeoutCtx)
	}()

	select {
	case err := <-errChan:
		// Authentication completed (success or error).
		return v.classifyError(err)
	case <-timeoutCtx.Done():
		// Authentication timed out.
		return cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrAuthenticationFailed,
			ErrAuthenticationTimeout,
		)
	}
}

// classifyError classifies hardware errors into user-facing application errors.
func (v *HardwareErrorValidator) classifyError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, ErrDeviceRemoved):
		wrapped := &cryptoutilIdentityAppErr.IdentityError{
			Code:       cryptoutilIdentityAppErr.ErrAuthenticationFailed.Code,
			Message:    cryptoutilIdentityAppErr.ErrAuthenticationFailed.Message,
			HTTPStatus: cryptoutilIdentityAppErr.ErrAuthenticationFailed.HTTPStatus,
			Internal:   fmt.Errorf("hardware device removed: %w", err),
		}

		return wrapped
	case errors.Is(err, ErrPINRetryExhausted):
		wrapped := &cryptoutilIdentityAppErr.IdentityError{
			Code:       cryptoutilIdentityAppErr.ErrAuthenticationFailed.Code,
			Message:    cryptoutilIdentityAppErr.ErrAuthenticationFailed.Message,
			HTTPStatus: cryptoutilIdentityAppErr.ErrAuthenticationFailed.HTTPStatus,
			Internal:   fmt.Errorf("PIN retry limit exhausted: %w", err),
		}

		return wrapped
	case errors.Is(err, ErrAuthenticationTimeout):
		wrapped := &cryptoutilIdentityAppErr.IdentityError{
			Code:       cryptoutilIdentityAppErr.ErrAuthenticationFailed.Code,
			Message:    cryptoutilIdentityAppErr.ErrAuthenticationFailed.Message,
			HTTPStatus: cryptoutilIdentityAppErr.ErrAuthenticationFailed.HTTPStatus,
			Internal:   fmt.Errorf("authentication timeout: %w", err),
		}

		return wrapped
	case errors.Is(err, ErrDeviceUnresponsive):
		wrapped := &cryptoutilIdentityAppErr.IdentityError{
			Code:       cryptoutilIdentityAppErr.ErrAuthenticationFailed.Code,
			Message:    cryptoutilIdentityAppErr.ErrAuthenticationFailed.Message,
			HTTPStatus: cryptoutilIdentityAppErr.ErrAuthenticationFailed.HTTPStatus,
			Internal:   fmt.Errorf("hardware device unresponsive: %w", err),
		}

		return wrapped
	case errors.Is(err, ErrInvalidPIN):
		wrapped := &cryptoutilIdentityAppErr.IdentityError{
			Code:       cryptoutilIdentityAppErr.ErrAuthenticationFailed.Code,
			Message:    cryptoutilIdentityAppErr.ErrAuthenticationFailed.Message,
			HTTPStatus: cryptoutilIdentityAppErr.ErrAuthenticationFailed.HTTPStatus,
			Internal:   fmt.Errorf("invalid PIN: %w", err),
		}

		return wrapped
	case errors.Is(err, ErrDeviceLocked):
		wrapped := &cryptoutilIdentityAppErr.IdentityError{
			Code:       cryptoutilIdentityAppErr.ErrAuthenticationFailed.Code,
			Message:    cryptoutilIdentityAppErr.ErrAuthenticationFailed.Message,
			HTTPStatus: cryptoutilIdentityAppErr.ErrAuthenticationFailed.HTTPStatus,
			Internal:   fmt.Errorf("hardware device locked: %w", err),
		}

		return wrapped
	default:
		// Unknown hardware error.
		wrapped := &cryptoutilIdentityAppErr.IdentityError{
			Code:       cryptoutilIdentityAppErr.ErrAuthenticationFailed.Code,
			Message:    cryptoutilIdentityAppErr.ErrAuthenticationFailed.Message,
			HTTPStatus: cryptoutilIdentityAppErr.ErrAuthenticationFailed.HTTPStatus,
			Internal:   fmt.Errorf("hardware authentication error: %w", err),
		}

		return wrapped
	}
}

// RetryWithBackoff retries hardware operations with exponential backoff.
func (v *HardwareErrorValidator) RetryWithBackoff(ctx context.Context, maxRetries int, operation func(context.Context) error) error {
	var lastErr error

	backoff := v.devicePollInterval

	for attempt := 0; attempt < maxRetries; attempt++ {
		err := operation(ctx)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check for non-retriable errors.
		if errors.Is(err, ErrPINRetryExhausted) || errors.Is(err, ErrDeviceLocked) {
			return v.classifyError(err)
		}

		// Wait with exponential backoff before retry.
		select {
		case <-ctx.Done():
			return cryptoutilIdentityAppErr.WrapError(
				cryptoutilIdentityAppErr.ErrAuthenticationFailed,
				fmt.Errorf("retry cancelled: %w", ctx.Err()),
			)
		case <-time.After(backoff):
			backoff *= 2
		}
	}

	return cryptoutilIdentityAppErr.WrapError(
		cryptoutilIdentityAppErr.ErrAuthenticationFailed,
		fmt.Errorf("max retries exhausted: %w", lastErr),
	)
}

// MonitorDevicePresence monitors hardware device presence during authentication.
func (v *HardwareErrorValidator) MonitorDevicePresence(ctx context.Context, checkFunc func(context.Context) error) error {
	ticker := time.NewTicker(v.devicePollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return cryptoutilIdentityAppErr.WrapError(
				cryptoutilIdentityAppErr.ErrAuthenticationFailed,
				fmt.Errorf("device monitoring cancelled: %w", ctx.Err()),
			)
		case <-ticker.C:
			err := checkFunc(ctx)
			if err != nil {
				if errors.Is(err, ErrDeviceRemoved) {
					return v.classifyError(err)
				}

				// Continue monitoring for transient errors.
				continue
			}

			// Device present.
			return nil
		}
	}
}
