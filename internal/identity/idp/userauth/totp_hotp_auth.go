// Copyright (c) 2025 Justin Cranford
//
//

package userauth

import (
	"context"
	"crypto/hmac"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	googleUuid "github.com/google/uuid"

	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

// TOTPAuthenticator implements TOTP-based authentication.
type TOTPAuthenticator struct {
	issuer         string
	digits         int
	period         time.Duration
	challengeStore ChallengeStore
	userRepo       cryptoutilIdentityRepository.UserRepository
}

// NewTOTPAuthenticator creates a new TOTP authenticator.
func NewTOTPAuthenticator(
	issuer string,
	challengeStore ChallengeStore,
	userRepo cryptoutilIdentityRepository.UserRepository,
) *TOTPAuthenticator {
	return &TOTPAuthenticator{
		issuer:         issuer,
		digits:         cryptoutilIdentityMagic.DefaultTOTPDigits,
		period:         cryptoutilIdentityMagic.DefaultTOTPPeriod,
		challengeStore: challengeStore,
		userRepo:       userRepo,
	}
}

// Method returns the authentication method name.
func (t *TOTPAuthenticator) Method() string {
	return "totp"
}

// GenerateSecret generates a random TOTP secret.
func (t *TOTPAuthenticator) GenerateSecret(_ context.Context) (string, error) {
	const secretLength = 20

	// Generate random bytes.
	secret := make([]byte, secretLength)
	if _, err := crand.Read(secret); err != nil {
		return "", fmt.Errorf("failed to generate random secret: %w", err)
	}

	// Encode as base32 (standard for TOTP secrets).
	encoded := base32.StdEncoding.EncodeToString(secret)

	// Remove padding.
	encoded = strings.TrimRight(encoded, "=")

	return encoded, nil
}

// GenerateTOTP generates a TOTP code from a secret.
func (t *TOTPAuthenticator) GenerateTOTP(_ context.Context, secret string) (string, error) {
	return t.generateTOTPAtTime(secret, time.Now())
}

// ValidateTOTP validates a TOTP code against a secret.
func (t *TOTPAuthenticator) ValidateTOTP(ctx context.Context, secret, code string) bool {
	// Allow for clock skew - check current time and ±1 period.
	const windowSize = 1

	now := time.Now()

	for i := -windowSize; i <= windowSize; i++ {
		testTime := now.Add(time.Duration(i) * t.period)

		expectedCode, err := t.generateTOTPAtTime(secret, testTime)
		if err != nil {
			return false
		}

		if code == expectedCode {
			return true
		}
	}

	return false
}

// InitiateAuth initiates TOTP authentication.
func (t *TOTPAuthenticator) InitiateAuth(ctx context.Context, userID string) (*AuthChallenge, error) {
	// Create challenge.
	challengeID, err := googleUuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate challenge ID: %w", err)
	}

	expiresAt := time.Now().Add(cryptoutilIdentityMagic.DefaultOTPLifetime)

	challenge := &AuthChallenge{
		ID:        challengeID,
		UserID:    userID,
		Method:    "totp",
		ExpiresAt: expiresAt,
		Metadata: map[string]any{
			"digits": t.digits,
			"period": t.period.Seconds(),
		},
	}

	// Store challenge.
	if err := t.challengeStore.Store(ctx, challenge, challengeID.String()); err != nil {
		return nil, fmt.Errorf("failed to store TOTP challenge: %w", err)
	}

	return challenge, nil
}

// VerifyAuth verifies TOTP authentication.
func (t *TOTPAuthenticator) VerifyAuth(ctx context.Context, challengeID, response string) (*cryptoutilIdentityDomain.User, error) {
	// Parse challenge ID.
	id, err := googleUuid.Parse(challengeID)
	if err != nil {
		return nil, fmt.Errorf("invalid challenge ID: %w", err)
	}

	// Retrieve challenge.
	challenge, storedSecret, err := t.challengeStore.Retrieve(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("challenge not found: %w", err)
	}

	// Check expiration.
	if time.Now().After(challenge.ExpiresAt) {
		// Best-effort cleanup of expired challenge.
		if err := t.challengeStore.Delete(ctx, id); err != nil {
			fmt.Printf("warning: failed to delete expired challenge: %v\n", err)
		}

		return nil, fmt.Errorf("tOTP challenge expired")
	}

	// Validate TOTP code.
	if !t.ValidateTOTP(ctx, storedSecret, response) {
		return nil, fmt.Errorf("invalid TOTP code")
	}

	// Delete challenge (single-use).
	if err := t.challengeStore.Delete(ctx, id); err != nil {
		fmt.Printf("warning: failed to delete challenge: %v\n", err)
	}

	// Get user.
	user, err := t.userRepo.GetBySub(ctx, challenge.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// generateTOTPAtTime generates a TOTP code for a specific time.
func (t *TOTPAuthenticator) generateTOTPAtTime(secret string, testTime time.Time) (string, error) {
	// Decode base32 secret.
	secret = strings.ToUpper(secret)
	// Add padding if needed.
	if l := len(secret) % 8; l != 0 { //nolint:mnd // Base32 requires padding to 8-byte boundary.
		secret += strings.Repeat("=", 8-l) //nolint:mnd // Base32 padding size.
	}

	key, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return "", fmt.Errorf("failed to decode secret: %w", err)
	}

	// Calculate time counter.
	counter := uint64(testTime.Unix()) / uint64(t.period.Seconds()) //nolint:gosec // No overflow in time-to-counter conversion.

	// FIPS 140-2/140-3 compliance: Use HMAC-SHA256 instead of SHA-1.
	// Note: RFC 6238 specifies SHA-1 as default, but SHA-256 is also supported.
	h := hmac.New(sha256.New, key)
	if err := binary.Write(h, binary.BigEndian, counter); err != nil {
		return "", fmt.Errorf("failed to write counter: %w", err)
	}

	hash := h.Sum(nil)

	// Dynamic truncation (RFC 6238 algorithm).
	offset := hash[len(hash)-1] & 0x0f                                           //nolint:mnd // RFC 6238 TOTP algorithm constant.
	truncatedHash := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7fffffff //nolint:mnd // RFC 6238 TOTP algorithm constant.

	// Generate code.
	const base10 = 10

	code := truncatedHash % uint32(pow(base10, t.digits)) //nolint:gosec // Conversion safe within TOTP digit range.

	// Format with leading zeros.
	format := fmt.Sprintf("%%0%dd", t.digits)

	return fmt.Sprintf(format, code), nil
}

// Helper function to calculate power.
func pow(base, exp int) int {
	result := 1
	for i := 0; i < exp; i++ {
		result *= base
	}

	return result
}

// HOTPAuthenticator implements HOTP-based authentication.
type HOTPAuthenticator struct {
	issuer         string
	digits         int
	challengeStore ChallengeStore
	userRepo       cryptoutilIdentityRepository.UserRepository
	counterStore   CounterStore
}

// CounterStore manages HOTP counters.
type CounterStore interface {
	GetCounter(ctx context.Context, userID string) (uint64, error)
	IncrementCounter(ctx context.Context, userID string) (uint64, error)
	SetCounter(ctx context.Context, userID string, counter uint64) error
}

// NewHOTPAuthenticator creates a new HOTP authenticator.
func NewHOTPAuthenticator(
	issuer string,
	challengeStore ChallengeStore,
	userRepo cryptoutilIdentityRepository.UserRepository,
	counterStore CounterStore,
) *HOTPAuthenticator {
	return &HOTPAuthenticator{
		issuer:         issuer,
		digits:         cryptoutilIdentityMagic.DefaultHOTPDigits,
		challengeStore: challengeStore,
		userRepo:       userRepo,
		counterStore:   counterStore,
	}
}

// Method returns the authentication method name.
func (h *HOTPAuthenticator) Method() string {
	return "hotp"
}

// GenerateSecret generates a random HOTP secret (same as TOTP).
func (h *HOTPAuthenticator) GenerateSecret(ctx context.Context) (string, error) {
	const secretLength = 20

	// Generate random bytes.
	secret := make([]byte, secretLength)
	if _, err := crand.Read(secret); err != nil {
		return "", fmt.Errorf("failed to generate random secret: %w", err)
	}

	// Encode as base32.
	encoded := base32.StdEncoding.EncodeToString(secret)

	// Remove padding.
	encoded = strings.TrimRight(encoded, "=")

	return encoded, nil
}

// GenerateHOTP generates an HOTP code from a secret and counter.
func (h *HOTPAuthenticator) GenerateHOTP(ctx context.Context, secret string, counter uint64) (string, error) {
	// Decode base32 secret.
	secret = strings.ToUpper(secret)
	// Add padding if needed.
	if l := len(secret) % 8; l != 0 { //nolint:mnd // Base32 requires padding to 8-byte boundary.
		secret += strings.Repeat("=", 8-l) //nolint:mnd // Base32 requires padding to 8-byte boundary.
	}

	key, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return "", fmt.Errorf("failed to decode secret: %w", err)
	}

	// FIPS 140-2/140-3 compliance: Use HMAC-SHA256 instead of SHA-1.
	// Note: RFC 4226 specifies SHA-1 as default, but SHA-256 is also supported.
	hm := hmac.New(sha256.New, key)
	if err := binary.Write(hm, binary.BigEndian, counter); err != nil {
		return "", fmt.Errorf("failed to write counter: %w", err)
	}

	hash := hm.Sum(nil)

	// Dynamic truncation (RFC 4226 algorithm).
	offset := hash[len(hash)-1] & 0x0f                                           //nolint:mnd // RFC 4226 HOTP algorithm constant.
	truncatedHash := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7fffffff //nolint:mnd // RFC 4226 HOTP algorithm constant.

	// Generate code.
	const base10 = 10

	code := truncatedHash % uint32(pow(base10, h.digits)) //nolint:gosec // Conversion safe within HOTP digit range.

	// Format with leading zeros.
	format := fmt.Sprintf("%%0%dd", h.digits)

	return fmt.Sprintf(format, code), nil
}

// ValidateHOTP validates an HOTP code against a secret.
func (h *HOTPAuthenticator) ValidateHOTP(ctx context.Context, userID, secret, code string) (bool, error) {
	// Get current counter.
	currentCounter, err := h.counterStore.GetCounter(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get counter: %w", err)
	}

	// Allow for lookahead to handle counter desynchronization.
	const lookaheadWindow = 10

	for i := uint64(0); i < lookaheadWindow; i++ {
		testCounter := currentCounter + i

		expectedCode, err := h.GenerateHOTP(ctx, secret, testCounter)
		if err != nil {
			return false, err
		}

		if code == expectedCode {
			// Valid code found - update counter.
			if err := h.counterStore.SetCounter(ctx, userID, testCounter+1); err != nil {
				return false, fmt.Errorf("failed to update counter: %w", err)
			}

			return true, nil
		}
	}

	return false, nil
}

// InitiateAuth initiates HOTP authentication.
func (h *HOTPAuthenticator) InitiateAuth(ctx context.Context, userID string) (*AuthChallenge, error) {
	// Create challenge.
	challengeID, err := googleUuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate challenge ID: %w", err)
	}

	expiresAt := time.Now().Add(cryptoutilIdentityMagic.DefaultOTPLifetime)

	challenge := &AuthChallenge{
		ID:        challengeID,
		UserID:    userID,
		Method:    "hotp",
		ExpiresAt: expiresAt,
		Metadata: map[string]any{
			"digits": h.digits,
		},
	}

	// Store challenge.
	if err := h.challengeStore.Store(ctx, challenge, challengeID.String()); err != nil {
		return nil, fmt.Errorf("failed to store HOTP challenge: %w", err)
	}

	return challenge, nil
}

// VerifyAuth verifies HOTP authentication.
func (h *HOTPAuthenticator) VerifyAuth(ctx context.Context, challengeID, response string) (*cryptoutilIdentityDomain.User, error) {
	// Parse challenge ID.
	id, err := googleUuid.Parse(challengeID)
	if err != nil {
		return nil, fmt.Errorf("invalid challenge ID: %w", err)
	}

	// Retrieve challenge.
	challenge, storedSecret, err := h.challengeStore.Retrieve(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("challenge not found: %w", err)
	}

	// Check expiration.
	if time.Now().After(challenge.ExpiresAt) {
		// Best-effort cleanup of expired challenge.
		if err := h.challengeStore.Delete(ctx, id); err != nil {
			fmt.Printf("warning: failed to delete expired challenge: %v\n", err)
		}

		return nil, fmt.Errorf("hOTP challenge expired")
	}

	// Validate HOTP code.
	valid, err := h.ValidateHOTP(ctx, challenge.UserID, storedSecret, response)
	if err != nil {
		return nil, fmt.Errorf("hOTP validation error: %w", err)
	}

	if !valid {
		return nil, fmt.Errorf("invalid HOTP code")
	}

	// Delete challenge (single-use).
	if err := h.challengeStore.Delete(ctx, id); err != nil {
		fmt.Printf("warning: failed to delete challenge: %v\n", err)
	}

	// Get user.
	user, err := h.userRepo.GetBySub(ctx, challenge.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}
