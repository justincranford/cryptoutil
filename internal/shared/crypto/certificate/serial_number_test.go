// Copyright (c) 2025 Justin Cranford
//
//

package certificate

import (
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestGenerateSerialNumber(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "generate serial number"},
		{name: "generate multiple unique serial numbers"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			serial, err := GenerateSerialNumber()
			require.NoError(t, err)
			require.NotNil(t, serial)

			// Verify serial is >= 2^64 (minSerialNumber).
			require.True(t, serial.Cmp(minSerialNumber) >= 0, "serial number should be >= 2^64")

			// Verify serial is < 2^159 (maxSerialNumber).
			require.True(t, serial.Cmp(maxSerialNumber) < 0, "serial number should be < 2^159")

			// Generate second serial and verify uniqueness (very high probability).
			if tc.name == "generate multiple unique serial numbers" {
				serial2, err2 := GenerateSerialNumber()
				require.NoError(t, err2)
				require.NotEqual(t, serial, serial2, "serial numbers should be unique")
			}
		})
	}
}

func TestRandomizedNotBeforeNotAfterCA_HappyPath(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	requestedDuration := 24 * time.Hour
	minSubtract := 5 * time.Minute
	maxSubtract := 10 * time.Minute

	notBefore, notAfter, err := randomizedNotBeforeNotAfterCA(now, requestedDuration, minSubtract, maxSubtract)
	require.NoError(t, err)

	// Verify notBefore is before requested start time.
	require.True(t, notBefore.Before(now) || notBefore.Equal(now))

	// Verify duration is approximately requestedDuration (with randomization).
	actualDuration := notAfter.Sub(notBefore)
	require.True(t, actualDuration >= requestedDuration-maxSubtract)
	require.True(t, actualDuration <= requestedDuration+maxSubtract)

	// Verify does not exceed max CA cert duration.
	require.True(t, actualDuration <= cryptoutilSharedMagic.TLSMaxCACertDuration)
}

func TestRandomizedNotBeforeNotAfterCA_ExceedsMaxDuration(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	requestedDuration := cryptoutilSharedMagic.TLSMaxCACertDuration + time.Hour // Exceeds max.
	minSubtract := 5 * time.Minute
	maxSubtract := 10 * time.Minute

	_, _, err := randomizedNotBeforeNotAfterCA(now, requestedDuration, minSubtract, maxSubtract)
	require.Error(t, err)
	require.Contains(t, err.Error(), "requestedDuration exceeds maxCACertDuration")
}

func TestRandomizedNotBeforeNotAfterEndEntity_HappyPath(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	requestedDuration := 24 * time.Hour
	minSubtract := 5 * time.Minute
	maxSubtract := 10 * time.Minute

	notBefore, notAfter, err := randomizedNotBeforeNotAfterEndEntity(now, requestedDuration, minSubtract, maxSubtract)
	require.NoError(t, err)

	// Verify notBefore is before requested start time.
	require.True(t, notBefore.Before(now) || notBefore.Equal(now))

	// Verify duration is approximately requestedDuration (with randomization).
	actualDuration := notAfter.Sub(notBefore)
	require.True(t, actualDuration >= requestedDuration-maxSubtract)
	require.True(t, actualDuration <= requestedDuration+maxSubtract)

	// Verify does not exceed max subscriber cert duration.
	require.True(t, actualDuration <= cryptoutilSharedMagic.TLSDefaultSubscriberCertDuration)
}

func TestRandomizedNotBeforeNotAfterEndEntity_ExceedsMaxDuration(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	requestedDuration := cryptoutilSharedMagic.TLSDefaultSubscriberCertDuration + time.Hour // Exceeds max.
	minSubtract := 5 * time.Minute
	maxSubtract := 10 * time.Minute

	_, _, err := randomizedNotBeforeNotAfterEndEntity(now, requestedDuration, minSubtract, maxSubtract)
	require.Error(t, err)
	require.Contains(t, err.Error(), "requestedDuration exceeds maxSubscriberCertDuration")
}

func TestRandomizedNotBeforeNotAfterCA_ExactMaxDurationSucceeds(t *testing.T) {
	t.Parallel()

	// Use exactly TLSDefaultMaxCACertDuration (== not >) with standard randomization.
	// The post-check uses requestedDuration+maxSubtract, so this succeeds.
	notBefore, notAfter, err := randomizedNotBeforeNotAfterCA(
		time.Now().UTC(),
		cryptoutilSharedMagic.TLSDefaultMaxCACertDuration,
		time.Minute,
		cryptoutilSharedMagic.CertificateRandomizationNotBeforeMinutes*time.Minute,
	)
	require.NoError(t, err)
	require.True(t, notBefore.Before(notAfter))
}

func TestRandomizedNotBeforeNotAfterEndEntity_ExactMaxDurationSucceeds(t *testing.T) {
	t.Parallel()

	// Use exactly TLSDefaultSubscriberCertDuration (== not >) with standard randomization.
	// The post-check uses requestedDuration+maxSubtract, so this succeeds.
	notBefore, notAfter, err := randomizedNotBeforeNotAfterEndEntity(
		time.Now().UTC(),
		cryptoutilSharedMagic.TLSDefaultSubscriberCertDuration,
		time.Minute,
		cryptoutilSharedMagic.CertificateRandomizationNotBeforeMinutes*time.Minute,
	)
	require.NoError(t, err)
	require.True(t, notBefore.Before(notAfter))
}

func TestGenerateNotBeforeNotAfter(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	requestedDuration := 48 * time.Hour
	minSubtract := 10 * time.Minute
	maxSubtract := 20 * time.Minute

	notBefore, notAfter, err := generateNotBeforeNotAfter(now, requestedDuration, minSubtract, maxSubtract)
	require.NoError(t, err)

	// Verify notBefore is before requested start time.
	require.True(t, notBefore.Before(now) || notBefore.Equal(now))

	// Verify duration is approximately requestedDuration (with randomization).
	actualDuration := notAfter.Sub(notBefore)
	require.True(t, actualDuration >= requestedDuration-maxSubtract)
	require.True(t, actualDuration <= requestedDuration+maxSubtract)
}
