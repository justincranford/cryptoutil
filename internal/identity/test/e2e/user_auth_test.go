package e2e

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// performUsernamePasswordAuth performs username/password authentication.
func (s *E2ETestSuite) performUsernamePasswordAuth(ctx context.Context) error {
	loginURL := fmt.Sprintf("%s/login", s.IDPURL)

	formData := url.Values{}
	formData.Set("username", "testuser@example.com")
	formData.Set("password", "SecureP@ssw0rd123")
	formData.Set("auth_method", string(UserAuthUsernamePassword))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform login: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) //nolint:errcheck // Error message is for logging only

		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// performEmailOTPAuth performs email OTP authentication.
func (s *E2ETestSuite) performEmailOTPAuth(ctx context.Context) error {
	// Step 1: Request OTP via email.
	loginURL := fmt.Sprintf("%s/login", s.IDPURL)

	formData := url.Values{}
	formData.Set("email", "testuser@example.com")
	formData.Set("auth_method", string(UserAuthEmailOTP))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create OTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to request OTP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) //nolint:errcheck // Error message is for logging only

		return fmt.Errorf("OTP request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Step 2: Verify OTP (in test environment, use fixed OTP).
	otpCode := "123456" // Test OTP code.

	formData = url.Values{}
	formData.Set("email", "testuser@example.com")
	formData.Set("otp_code", otpCode)
	formData.Set("auth_method", string(UserAuthEmailOTP))

	req, err = http.NewRequestWithContext(ctx, http.MethodPost, loginURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create OTP verification request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err = s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to verify OTP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) //nolint:errcheck // Error message is for logging only

		return fmt.Errorf("OTP verification failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// performSMSOTPAuth performs SMS OTP authentication.
func (s *E2ETestSuite) performSMSOTPAuth(ctx context.Context) error {
	// Similar to email OTP, but uses phone number.
	loginURL := fmt.Sprintf("%s/login", s.IDPURL)

	// Step 1: Request OTP via SMS.
	formData := url.Values{}
	formData.Set("phone_number", "+15555551234")
	formData.Set("auth_method", string(UserAuthSMSOTP))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create SMS OTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to request SMS OTP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) //nolint:errcheck // Error message is for logging only

		return fmt.Errorf("SMS OTP request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Step 2: Verify OTP.
	otpCode := "123456" // Test OTP code.

	formData = url.Values{}
	formData.Set("phone_number", "+15555551234")
	formData.Set("otp_code", otpCode)
	formData.Set("auth_method", string(UserAuthSMSOTP))

	req, err = http.NewRequestWithContext(ctx, http.MethodPost, loginURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create SMS OTP verification request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err = s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to verify SMS OTP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) //nolint:errcheck // Error message is for logging only

		return fmt.Errorf("SMS OTP verification failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// performTOTPAuth performs TOTP authentication.
func (s *E2ETestSuite) performTOTPAuth(ctx context.Context) error {
	loginURL := fmt.Sprintf("%s/login", s.IDPURL)

	// In test environment, use fixed TOTP code.
	totpCode := "123456"

	formData := url.Values{}
	formData.Set("username", "testuser@example.com")
	formData.Set("totp_code", totpCode)
	formData.Set("auth_method", string(UserAuthTOTP))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create TOTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform TOTP auth: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) //nolint:errcheck // Error message is for logging only

		return fmt.Errorf("TOTP auth failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// performHOTPAuth performs HOTP authentication.
func (s *E2ETestSuite) performHOTPAuth(ctx context.Context) error {
	loginURL := fmt.Sprintf("%s/login", s.IDPURL)

	// In test environment, use fixed HOTP code and counter.
	hotpCode := "654321"
	counter := "1"

	formData := url.Values{}
	formData.Set("username", "testuser@example.com")
	formData.Set("hotp_code", hotpCode)
	formData.Set("counter", counter)
	formData.Set("auth_method", string(UserAuthHOTP))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create HOTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform HOTP auth: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) //nolint:errcheck // Error message is for logging only

		return fmt.Errorf("HOTP auth failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// performMagicLinkAuth performs magic link authentication.
func (s *E2ETestSuite) performMagicLinkAuth(ctx context.Context) error {
	loginURL := fmt.Sprintf("%s/login", s.IDPURL)

	// Step 1: Request magic link.
	formData := url.Values{}
	formData.Set("email", "testuser@example.com")
	formData.Set("auth_method", string(UserAuthMagicLink))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create magic link request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to request magic link: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) //nolint:errcheck // Error message is for logging only

		return fmt.Errorf("magic link request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Step 2: Simulate clicking magic link (in test, use fixed token).
	magicToken := "test_magic_token_12345"

	verifyURL := fmt.Sprintf("%s/login?magic_token=%s", s.IDPURL, url.QueryEscape(magicToken))

	req, err = http.NewRequestWithContext(ctx, http.MethodGet, verifyURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create magic link verification request: %w", err)
	}

	resp, err = s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to verify magic link: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) //nolint:errcheck // Error message is for logging only

		return fmt.Errorf("magic link verification failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// performPasskeyAuth performs passkey (WebAuthn) authentication.
func (s *E2ETestSuite) performPasskeyAuth(ctx context.Context) error {
	loginURL := fmt.Sprintf("%s/login", s.IDPURL)

	// Step 1: Get authentication challenge.
	formData := url.Values{}
	formData.Set("username", "testuser@example.com")
	formData.Set("auth_method", string(UserAuthPasskey))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create passkey challenge request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get passkey challenge: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) //nolint:errcheck // Error message is for logging only

		return fmt.Errorf("passkey challenge failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Step 2: Respond to challenge (in test, use mock credential).
	// In real scenario, this would involve WebAuthn API.
	mockCredentialID := base64.StdEncoding.EncodeToString([]byte("test_credential_id"))
	mockAuthenticatorData := base64.StdEncoding.EncodeToString([]byte("mock_authenticator_data"))
	mockSignature := base64.StdEncoding.EncodeToString([]byte("mock_signature"))

	formData = url.Values{}
	formData.Set("credential_id", mockCredentialID)
	formData.Set("authenticator_data", mockAuthenticatorData)
	formData.Set("signature", mockSignature)
	formData.Set("auth_method", string(UserAuthPasskey))

	req, err = http.NewRequestWithContext(ctx, http.MethodPost, loginURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create passkey response request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err = s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to verify passkey: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) //nolint:errcheck // Error message is for logging only

		return fmt.Errorf("passkey verification failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// performBiometricAuth performs biometric authentication.
func (s *E2ETestSuite) performBiometricAuth(ctx context.Context) error {
	loginURL := fmt.Sprintf("%s/login", s.IDPURL)

	// Simulate biometric authentication with mock biometric data.
	biometricData := make([]byte, 32)
	if _, err := rand.Read(biometricData); err != nil {
		return fmt.Errorf("failed to generate mock biometric data: %w", err)
	}

	formData := url.Values{}
	formData.Set("username", "testuser@example.com")
	formData.Set("biometric_data", base64.StdEncoding.EncodeToString(biometricData))
	formData.Set("biometric_type", "fingerprint")
	formData.Set("auth_method", string(UserAuthBiometric))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create biometric request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform biometric auth: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) //nolint:errcheck // Error message is for logging only

		return fmt.Errorf("biometric auth failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// performHardwareKeyAuth performs hardware key authentication.
func (s *E2ETestSuite) performHardwareKeyAuth(ctx context.Context) error {
	loginURL := fmt.Sprintf("%s/login", s.IDPURL)

	// Step 1: Get authentication challenge.
	formData := url.Values{}
	formData.Set("username", "testuser@example.com")
	formData.Set("auth_method", string(UserAuthHardwareKey))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create hardware key challenge request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get hardware key challenge: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) //nolint:errcheck // Error message is for logging only

		return fmt.Errorf("hardware key challenge failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Step 2: Respond with hardware key signature (mock U2F/FIDO2 response).
	mockKeyHandle := base64.StdEncoding.EncodeToString([]byte("test_key_handle"))
	mockSignature := base64.StdEncoding.EncodeToString([]byte("mock_u2f_signature"))

	formData = url.Values{}
	formData.Set("key_handle", mockKeyHandle)
	formData.Set("signature", mockSignature)
	formData.Set("auth_method", string(UserAuthHardwareKey))

	req, err = http.NewRequestWithContext(ctx, http.MethodPost, loginURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create hardware key response request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err = s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to verify hardware key: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) //nolint:errcheck // Error message is for logging only

		return fmt.Errorf("hardware key verification failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// TestUserAuthentication tests each user authentication method individually.
func TestUserAuthentication(t *testing.T) {
	suite := NewE2ETestSuite()
	ctx := context.Background()

	t.Run("UsernamePassword", func(t *testing.T) {
		err := suite.performUserAuth(ctx, UserAuthUsernamePassword)
		require.NoError(t, err, "Username/password auth should succeed")
	})

	t.Run("EmailOTP", func(t *testing.T) {
		err := suite.performUserAuth(ctx, UserAuthEmailOTP)
		require.NoError(t, err, "Email OTP auth should succeed")
	})

	t.Run("SMSOTP", func(t *testing.T) {
		err := suite.performUserAuth(ctx, UserAuthSMSOTP)
		require.NoError(t, err, "SMS OTP auth should succeed")
	})

	t.Run("TOTP", func(t *testing.T) {
		err := suite.performUserAuth(ctx, UserAuthTOTP)
		require.NoError(t, err, "TOTP auth should succeed")
	})

	t.Run("HOTP", func(t *testing.T) {
		err := suite.performUserAuth(ctx, UserAuthHOTP)
		require.NoError(t, err, "HOTP auth should succeed")
	})

	t.Run("MagicLink", func(t *testing.T) {
		err := suite.performUserAuth(ctx, UserAuthMagicLink)
		require.NoError(t, err, "Magic link auth should succeed")
	})

	t.Run("Passkey", func(t *testing.T) {
		err := suite.performUserAuth(ctx, UserAuthPasskey)
		require.NoError(t, err, "Passkey auth should succeed")
	})

	t.Run("Biometric", func(t *testing.T) {
		err := suite.performUserAuth(ctx, UserAuthBiometric)
		require.NoError(t, err, "Biometric auth should succeed")
	})

	t.Run("HardwareKey", func(t *testing.T) {
		err := suite.performUserAuth(ctx, UserAuthHardwareKey)
		require.NoError(t, err, "Hardware key auth should succeed")
	})
}
