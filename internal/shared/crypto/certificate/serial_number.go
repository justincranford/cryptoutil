// Copyright (c) 2025 Justin Cranford
//
//

package certificate

import (
	crand "crypto/rand"
	"fmt"
	"math/big"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	minSerialNumber   = new(big.Int).Lsh(big.NewInt(1), cryptoutilSharedMagic.MinSerialNumberBits) // 2^64
	maxSerialNumber   = new(big.Int).Lsh(big.NewInt(1), cryptoutilSharedMagic.MaxSerialNumberBits) // 2^159
	rangeSerialNumber = new(big.Int).Sub(maxSerialNumber, minSerialNumber)                         // Range size: 2^159 - 2^64
)

// GenerateSerialNumber generates a cryptographically random serial number in the range [2^64, 2^159) per CA/Browser Forum requirements.
func GenerateSerialNumber() (*big.Int, error) {
	randomOffsetFromMin, err := crand.Int(crand.Reader, rangeSerialNumber) // Range [0, rangeSerialNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random serial number offset: %w", err)
	}

	return new(big.Int).Add(randomOffsetFromMin, minSerialNumber), nil // Valid range [2^64, 2^159)
}

func randomizedNotBeforeNotAfterCA(requestedStart time.Time, requestedDuration, minSubtract, maxSubtract time.Duration) (time.Time, time.Time, error) {
	if requestedDuration > cryptoutilSharedMagic.TLSDefaultMaxCACertDuration {
		return time.Time{}, time.Time{}, fmt.Errorf("requestedDuration exceeds maxCACertDuration")
	}

	notBefore, notAfter, err := generateNotBeforeNotAfter(requestedStart, requestedDuration, minSubtract, maxSubtract)
	if err != nil {
		return notBefore, notAfter, fmt.Errorf("failed to generate notBefore/notAfter: %w", err)
	} else if notAfter.Sub(notBefore) > cryptoutilSharedMagic.TLSDefaultMaxCACertDuration {
		return notBefore, notAfter, fmt.Errorf("actual duration exceeds maxCACertDuration")
	}

	return notBefore, notAfter, nil
}

func randomizedNotBeforeNotAfterEndEntity(requestedStart time.Time, requestedDuration, minSubtract, maxSubtract time.Duration) (time.Time, time.Time, error) {
	if requestedDuration > cryptoutilSharedMagic.TLSDefaultSubscriberCertDuration {
		return time.Time{}, time.Time{}, fmt.Errorf("requestedDuration exceeds maxSubscriberCertDuration")
	}

	notBefore, notAfter, err := generateNotBeforeNotAfter(requestedStart, requestedDuration, minSubtract, maxSubtract)
	if err != nil {
		return notBefore, notAfter, fmt.Errorf("failed to generate notBefore/notAfter: %w", err)
	} else if notAfter.Sub(notBefore) > cryptoutilSharedMagic.TLSDefaultSubscriberCertDuration {
		return notBefore, notAfter, fmt.Errorf("actual duration exceeds maxSubscriberCertDuration")
	}

	return notBefore, notAfter, nil
}

func generateNotBeforeNotAfter(requestedStart time.Time, requestedDuration, minSubtract, maxSubtract time.Duration) (time.Time, time.Time, error) {
	maxRangeOffset := int64(maxSubtract - minSubtract)

	if requestedDuration <= 0 {
		return time.Time{}, time.Time{}, fmt.Errorf("requestedDuration must be positive, got: %v", requestedDuration)
	} else if minSubtract <= 0 {
		return time.Time{}, time.Time{}, fmt.Errorf("minSubtract must be positive, got: %v", minSubtract)
	} else if maxSubtract <= 0 {
		return time.Time{}, time.Time{}, fmt.Errorf("maxSubtract must be positive, got: %v", maxSubtract)
	} else if maxRangeOffset <= 0 {
		return time.Time{}, time.Time{}, fmt.Errorf("maxRangeOffset must be positive, got (%v)", maxRangeOffset)
	}

	rangeOffset, err := crand.Int(crand.Reader, big.NewInt(maxRangeOffset)) // [0, maxRangeOffset)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed to generate random range offset: %w", err)
	}

	notBefore := requestedStart.Add(-minSubtract - time.Duration(rangeOffset.Int64())) // [-minSubtract - maxRangeOffset, -minSubtract)
	notAfter := requestedStart.Add(requestedDuration)

	return notBefore, notAfter, nil
}
