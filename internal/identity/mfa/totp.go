// Copyright (c) 2025 Justin Cranford
//
//

package mfa

import (
	"context"
	hmac "crypto/hmac"
	crand "crypto/rand"
	"crypto/sha1"
	sha256 "crypto/sha256"
	sha512 "crypto/sha512"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"math"
	"strings"
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilSharedCryptoPassword "cryptoutil/internal/shared/crypto/password"
)

const (
	DefaultTOTPDigits    = 6
	DefaultTOTPPeriod    = 30
	DefaultTOTPAlgorithm = "SHA1"
	TOTPSecretLength     = 20
	BackupCodeCount      = 10
	BackupCodeLength     = 8
	MaxFailedAttempts    = 5
	LockDuration         = 15 * time.Minute
	MFAStepUpDuration    = 30 * time.Minute
)

// TOTPService provides TOTP MFA operations.
type TOTPService struct {
	db *gorm.DB
}

// NewTOTPService creates a new TOTP service.
func NewTOTPService(db *gorm.DB) *TOTPService {
	return &TOTPService{db: db}
}

// EnrollTOTP enrolls a user in TOTP MFA.
func (s *TOTPService) EnrollTOTP(ctx context.Context, userID googleUuid.UUID, issuer, accountName string) (*TOTPSecret, string, []string, error) {
	// Generate cryptographically secure random secret.
	secretBytes := make([]byte, TOTPSecretLength)
	if _, err := crand.Read(secretBytes); err != nil {
		return nil, "", nil, fmt.Errorf("failed to generate TOTP secret: %w", err)
	}

	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secretBytes)

	// Create TOTP configuration.
	totpSecret := &TOTPSecret{
		ID:              googleUuid.New(),
		UserID:          userID,
		Secret:          secret,
		Algorithm:       DefaultTOTPAlgorithm,
		Digits:          DefaultTOTPDigits,
		Period:          DefaultTOTPPeriod,
		Verified:        false,
		RecoveryEnabled: true,
		FailedAttempts:  0,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	if err := s.db.WithContext(ctx).Create(totpSecret).Error; err != nil {
		return nil, "", nil, fmt.Errorf("failed to save TOTP secret: %w", err)
	}

	// Generate QR code URI.
	qrURI := s.generateOTPAuthURI(issuer, accountName, secret)

	// Generate backup codes.
	backupCodes, err := s.GenerateBackupCodes(ctx, userID)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}

	return totpSecret, qrURI, backupCodes, nil
}

// VerifyTOTP verifies a TOTP code for the user.
func (s *TOTPService) VerifyTOTP(ctx context.Context, userID googleUuid.UUID, code string) error {
	var totpSecret TOTPSecret
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&totpSecret).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("TOTP not enrolled for user")
		}

		return fmt.Errorf("failed to get TOTP secret: %w", err)
	}

	// Check if locked due to too many failed attempts.
	if time.Now().UTC().Before(totpSecret.LockedUntil) {
		return cryptoutilIdentityAppErr.ErrTOTPAccountLocked
	}

	// Verify code with ±1 time window (90s tolerance).
	valid, err := s.verifyCode(&totpSecret, code, 1)
	if err != nil {
		return fmt.Errorf("failed to verify TOTP code: %w", err)
	}

	if !valid {
		// Increment failed attempts.
		totpSecret.FailedAttempts++
		if totpSecret.FailedAttempts >= MaxFailedAttempts {
			totpSecret.LockedUntil = time.Now().UTC().Add(LockDuration)
		}

		totpSecret.UpdatedAt = time.Now().UTC()
		if err := s.db.WithContext(ctx).Save(&totpSecret).Error; err != nil {
			return fmt.Errorf("failed to update failed attempts: %w", err)
		}

		return fmt.Errorf("invalid TOTP code")
	}

	// Reset failed attempts on successful verification.
	totpSecret.FailedAttempts = 0
	totpSecret.Verified = true
	totpSecret.LastUsedAt = time.Now().UTC()
	totpSecret.UpdatedAt = time.Now().UTC()

	if err := s.db.WithContext(ctx).Save(&totpSecret).Error; err != nil {
		return fmt.Errorf("failed to update TOTP secret: %w", err)
	}

	return nil
}

// RequiresMFAStepUp checks if MFA step-up is required.
func (s *TOTPService) RequiresMFAStepUp(ctx context.Context, userID googleUuid.UUID) (bool, error) {
	var totpSecret TOTPSecret
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&totpSecret).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}

		return false, fmt.Errorf("failed to get TOTP secret: %w", err)
	}

	if totpSecret.LastUsedAt.IsZero() {
		return true, nil
	}

	return time.Now().UTC().Sub(totpSecret.LastUsedAt) > MFAStepUpDuration, nil
}

// GenerateBackupCodes generates backup codes for account recovery.
func (s *TOTPService) GenerateBackupCodes(ctx context.Context, userID googleUuid.UUID) ([]string, error) {
	// Delete existing backup codes.
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&BackupCode{}).Error; err != nil {
		return nil, fmt.Errorf("failed to delete old backup codes: %w", err)
	}

	plaintextCodes := make([]string, BackupCodeCount)
	backupCodes := make([]*BackupCode, BackupCodeCount)

	for i := 0; i < BackupCodeCount; i++ {
		// Generate random backup code.
		codeBytes := make([]byte, BackupCodeLength)
		if _, err := crand.Read(codeBytes); err != nil {
			return nil, fmt.Errorf("failed to generate backup code: %w", err)
		}

		plaintext := base64.RawURLEncoding.EncodeToString(codeBytes)
		plaintextCodes[i] = plaintext

		// Hash code with SHA-256.
		hashed, err := cryptoutilSharedCryptoPassword.HashPassword(plaintext)
		if err != nil {
			return nil, fmt.Errorf("failed to hash backup code: %w", err)
		}

		backupCodes[i] = &BackupCode{
			ID:        googleUuid.New(),
			UserID:    userID,
			CodeHash:  hashed,
			Used:      false,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
	}

	if err := s.db.WithContext(ctx).Create(&backupCodes).Error; err != nil {
		return nil, fmt.Errorf("failed to save backup codes: %w", err)
	}

	return plaintextCodes, nil
}

// VerifyBackupCode verifies a backup code and marks it as used.
func (s *TOTPService) VerifyBackupCode(ctx context.Context, userID googleUuid.UUID, code string) error {
	var backupCodes []BackupCode
	if err := s.db.WithContext(ctx).Where("user_id = ? AND used = ?", userID, false).Find(&backupCodes).Error; err != nil {
		return fmt.Errorf("failed to get backup codes: %w", err)
	}

	for i := range backupCodes {
		match, _, err := cryptoutilSharedCryptoPassword.VerifyPassword(code, backupCodes[i].CodeHash)
		if err == nil && match {
			now := time.Now().UTC()
			backupCodes[i].Used = true
			backupCodes[i].UsedAt = &now
			backupCodes[i].UpdatedAt = now

			if err := s.db.WithContext(ctx).Save(&backupCodes[i]).Error; err != nil {
				return fmt.Errorf("failed to mark backup code as used: %w", err)
			}

			return nil
		}
	}

	return fmt.Errorf("invalid backup code")
}

// generateOTPAuthURI generates the otpauth:// URI for QR code generation.
func (s *TOTPService) generateOTPAuthURI(issuer, accountName, secret string) string {
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=%s&digits=%d&period=%d",
		issuer, accountName, secret, issuer, DefaultTOTPAlgorithm, DefaultTOTPDigits, DefaultTOTPPeriod)
}

// verifyCode verifies a TOTP code with time window tolerance.
func (s *TOTPService) verifyCode(secret *TOTPSecret, code string, window int) (bool, error) {
	currentTime := time.Now().UTC().Unix()

	for offset := -window; offset <= window; offset++ {
		timeStep := currentTime/int64(secret.Period) + int64(offset)

		expectedCode, err := s.generateTOTP(secret.Secret, timeStep, secret.Algorithm, secret.Digits)
		if err != nil {
			return false, err
		}

		if code == expectedCode {
			return true, nil
		}
	}

	return false, nil
}

// generateTOTP generates a TOTP code per RFC 6238.
func (s *TOTPService) generateTOTP(secret string, timeStep int64, algorithm string, digits int) (string, error) {
	// Decode base32 secret.
	secretBytes, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		return "", fmt.Errorf("failed to decode secret: %w", err)
	}

	// Convert time step to byte array (big endian).
	timeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timeBytes, uint64(timeStep))

	// Generate HMAC based on algorithm.
	var mac hash.Hash

	switch strings.ToUpper(algorithm) {
	case "SHA1":
		mac = hmac.New(sha1.New, secretBytes)
	case "SHA256":
		mac = hmac.New(sha256.New, secretBytes)
	case "SHA512":
		mac = hmac.New(sha512.New, secretBytes)
	default:
		return "", fmt.Errorf("unsupported algorithm: %s", algorithm)
	}

	mac.Write(timeBytes)
	hmacResult := mac.Sum(nil)

	// Dynamic truncation per RFC 4226 (HOTP).
	offset := hmacResult[len(hmacResult)-1] & 0x0F
	truncatedHash := binary.BigEndian.Uint32(hmacResult[offset:offset+4]) & 0x7FFFFFFF

	// Generate code with specified digits.
	code := truncatedHash % uint32(math.Pow10(digits))

	return fmt.Sprintf("%0*d", digits, code), nil
}
