// Copyright (c) 2025 Justin Cranford

// Package mfa provides multi-factor authentication utilities.
package mfa

import (
	crand "crypto/rand"
	"fmt"
	"math/big"

	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// GenerateEmailOTP generates a 6-digit numeric OTP for email delivery.
func GenerateEmailOTP() (string, error) {
	otp := ""
	charset := cryptoutilIdentityMagic.EmailOTPCharset
	charsetLen := big.NewInt(int64(len(charset)))

	for i := 0; i < cryptoutilIdentityMagic.DefaultEmailOTPLength; i++ {
		randomIndex, err := crand.Int(crand.Reader, charsetLen)
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}

		otp += string(charset[randomIndex.Int64()])
	}

	return otp, nil
}
