// Copyright (c) 2025 Justin Cranford
//
//

//go:build integration_placeholder

package userauth

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"math"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityMagic "cryptoutil/internal/shared/magic"
)

// TestRiskScenario_LowRisk tests low-risk authentication scenarios.
func TestRiskScenario_LowRisk(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		authContext      *AuthContext
		baseline         *UserBaseline
		expectedMaxScore float64
		expectedLevel    RiskLevel
		expectedMinConf  float64
	}{
		{
			name: "known device, known location, normal time",
			authContext: &AuthContext{
				DeviceFingerprint: &DeviceFingerprint{
					ID:         "known-device-001",
					UserAgent:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0",
					Platform:   "Win32",
					Language:   "en-US",
					ScreenSize: "1920x1080",
					Timezone:   "America/New_York",
				},
				GeoLocation: &GeoLocation{
					IPAddress: "192.168.1.100",
					Country:   "US",
					City:      "New York",
					Latitude:  40.7128,
					Longitude: -74.0060,
				},
				NetworkInfo: &NetworkInfo{
					IPAddress: "192.168.1.100",
					IsVPN:     false,
					IsProxy:   false,
					IsTor:     false,
				},
				Timestamp: time.Date(2025, 1, 15, 14, 0, 0, 0, time.UTC), // 2 PM UTC (9 AM EST - normal business hours).
			},
			baseline: &UserBaseline{
				UserID: googleUuid.New().String(),
				KnownLocations: []GeoLocation{
					{Country: "US", City: "New York"},
				},
				KnownDevices: []DeviceFingerprint{
					{ID: "known-device-001"},
				},
				TypicalLoginHours: []int{cryptoutilSharedMagic.IMMinPasswordLength, 9, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, 11, cryptoutilSharedMagic.HashPrefixLength, 13, 14, 15, cryptoutilSharedMagic.RealmMinTokenLengthBytes, 17}, // Business hours EST.
				LastLoginTime:     time.Now().UTC().Add(-cryptoutilSharedMagic.HoursPerDay * time.Hour),
				EstablishedAt:     time.Now().UTC().Add(-cryptoutilSharedMagic.StrictCertificateMaxAgeDays * cryptoutilSharedMagic.HoursPerDay * time.Hour), // 90 days of data.
				EventCount:        150,                                  // Well-established baseline.
			},
			expectedMaxScore: 0.2, // Low risk threshold.
			expectedLevel:    RiskLevelLow,
			expectedMinConf:  cryptoutilSharedMagic.RiskScoreCritical, // High confidence due to established baseline.
		},
		{
			name: "trusted location during typical hours",
			authContext: &AuthContext{
				DeviceFingerprint: &DeviceFingerprint{
					ID:         "known-device-002",
					UserAgent:  "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Safari/17.0",
					Platform:   "MacIntel",
					Language:   "en-US",
					ScreenSize: "2560x1440",
					Timezone:   "America/Los_Angeles",
				},
				GeoLocation: &GeoLocation{
					IPAddress: "10.0.0.50",
					Country:   "US",
					City:      "San Francisco",
					Latitude:  37.7749,
					Longitude: -122.4194,
				},
				NetworkInfo: &NetworkInfo{
					IPAddress: "10.0.0.50",
					IsVPN:     false,
					IsProxy:   false,
					IsTor:     false,
				},
				Timestamp: time.Date(2025, 1, 15, 18, 0, 0, 0, time.UTC), // 6 PM UTC (10 AM PST).
			},
			baseline: &UserBaseline{
				UserID: googleUuid.New().String(),
				KnownLocations: []GeoLocation{
					{Country: "US", City: "San Francisco"},
				},
				KnownDevices: []DeviceFingerprint{
					{ID: "known-device-002"},
				},
				TypicalLoginHours: []int{9, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, 11, cryptoutilSharedMagic.HashPrefixLength, 13, 14, 15, cryptoutilSharedMagic.RealmMinTokenLengthBytes, 17, 18},
				LastLoginTime:     time.Now().UTC().Add(-cryptoutilSharedMagic.DefaultEmailOTPLength * time.Hour),
				EstablishedAt:     time.Now().UTC().Add(-cryptoutilSharedMagic.IdentityDefaultIdleTimeoutSeconds * cryptoutilSharedMagic.HoursPerDay * time.Hour),
				EventCount:        cryptoutilSharedMagic.CertificateRandomizationNotBeforeMinutes,
			},
			expectedMaxScore: cryptoutilSharedMagic.ConfidenceWeightBaseline,
			expectedLevel:    RiskLevelLow,
			expectedMinConf:  0.75,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockUserHistory := &mockUserBehaviorStore{
				baseline: tc.baseline,
			}
			mockGeoIP := &mockGeoIPService{}
			mockDeviceDB := &mockDeviceFingerprintDB{}

			engine := NewBehavioralRiskEngine(nil, mockUserHistory, mockGeoIP, mockDeviceDB)

			factors := engine.CalculateRiskFactors(tc.authContext, tc.baseline)

			require.NotEmpty(t, factors)

			// Calculate overall score.
			score := calculateScore(factors, engine)

			require.LessOrEqual(t, score, tc.expectedMaxScore, "Score should be within low-risk threshold")

			// Determine risk level.
			level := determineLevel(score)

			require.Equal(t, tc.expectedLevel, level)

			// Calculate confidence.
			confidence := calculateConfidence(tc.baseline)

			require.GreaterOrEqual(t, confidence, tc.expectedMinConf, "Confidence should be high for established baseline")
		})
	}
}

// TestRiskScenario_MediumRisk tests medium-risk authentication scenarios.
func TestRiskScenario_MediumRisk(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		authContext      *AuthContext
		baseline         *UserBaseline
		expectedMinScore float64
		expectedMaxScore float64
		expectedLevel    RiskLevel
	}{
		{
			name: "new location, known device",
			authContext: &AuthContext{
				DeviceFingerprint: &DeviceFingerprint{
					ID:         "known-device-001",
					UserAgent:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0",
					Platform:   "Win32",
					Language:   "en-US",
					ScreenSize: "1920x1080",
					Timezone:   "Europe/London",
				},
				GeoLocation: &GeoLocation{
					IPAddress: "203.0.113.42",
					Country:   "GB",
					City:      "London",
					Latitude:  51.5074,
					Longitude: -0.1278,
				},
				NetworkInfo: &NetworkInfo{
					IPAddress: "203.0.113.42",
					IsVPN:     false,
					IsProxy:   false,
					IsTor:     false,
				},
				Timestamp: time.Date(2025, 1, 15, cryptoutilSharedMagic.HashPrefixLength, 0, 0, 0, time.UTC),
			},
			baseline: &UserBaseline{
				UserID: googleUuid.New().String(),
				KnownLocations: []GeoLocation{
					{Country: "US", City: "New York"},
				},
				KnownDevices: []DeviceFingerprint{
					{ID: "known-device-001"},
				},
				TypicalLoginHours: []int{cryptoutilSharedMagic.IMMinPasswordLength, 9, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, 11, cryptoutilSharedMagic.HashPrefixLength, 13, 14, 15, cryptoutilSharedMagic.RealmMinTokenLengthBytes, 17},
				LastLoginTime:     time.Now().UTC().Add(-cryptoutilSharedMagic.HoursPerDay * time.Hour),
				EstablishedAt:     time.Now().UTC().Add(-cryptoutilSharedMagic.StrictCertificateMaxAgeDays * cryptoutilSharedMagic.HoursPerDay * time.Hour),
				EventCount:        150,
			},
			expectedMinScore: 0.2,
			expectedMaxScore: cryptoutilSharedMagic.Tolerance50Percent,
			expectedLevel:    RiskLevelMedium,
		},
		{
			name: "unusual login hour, known device and location",
			authContext: &AuthContext{
				DeviceFingerprint: &DeviceFingerprint{
					ID:         "known-device-002",
					UserAgent:  "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Safari/17.0",
					Platform:   "MacIntel",
					Language:   "en-US",
					ScreenSize: "2560x1440",
					Timezone:   "America/New_York",
				},
				GeoLocation: &GeoLocation{
					IPAddress: "192.168.1.100",
					Country:   "US",
					City:      "New York",
					Latitude:  40.7128,
					Longitude: -74.0060,
				},
				NetworkInfo: &NetworkInfo{
					IPAddress: "192.168.1.100",
					IsVPN:     false,
					IsProxy:   false,
					IsTor:     false,
				},
				Timestamp: time.Date(2025, 1, 15, 3, 0, 0, 0, time.UTC), // 3 AM UTC (10 PM EST previous day - unusual).
			},
			baseline: &UserBaseline{
				UserID: googleUuid.New().String(),
				KnownLocations: []GeoLocation{
					{Country: "US", City: "New York"},
				},
				KnownDevices: []DeviceFingerprint{
					{ID: "known-device-002"},
				},
				TypicalLoginHours: []int{cryptoutilSharedMagic.IMMinPasswordLength, 9, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, 11, cryptoutilSharedMagic.HashPrefixLength, 13, 14, 15, cryptoutilSharedMagic.RealmMinTokenLengthBytes, 17},
				LastLoginTime:     time.Now().UTC().Add(-cryptoutilSharedMagic.HashPrefixLength * time.Hour),
				EstablishedAt:     time.Now().UTC().Add(-cryptoutilSharedMagic.IdentityDefaultIdleTimeoutSeconds * cryptoutilSharedMagic.HoursPerDay * time.Hour),
				EventCount:        cryptoutilSharedMagic.JoseJAMaxMaterials,
			},
			expectedMinScore: 0.2,
			expectedMaxScore: cryptoutilSharedMagic.RiskScoreMedium,
			expectedLevel:    RiskLevelMedium,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockUserHistory := &mockUserBehaviorStore{
				baseline: tc.baseline,
			}
			mockGeoIP := &mockGeoIPService{}
			mockDeviceDB := &mockDeviceFingerprintDB{}

			engine := NewBehavioralRiskEngine(nil, mockUserHistory, mockGeoIP, mockDeviceDB)

			factors := engine.CalculateRiskFactors(tc.authContext, tc.baseline)

			require.NotEmpty(t, factors)

			score := calculateScore(factors, engine)

			require.GreaterOrEqual(t, score, tc.expectedMinScore, "Score should be at least minimum medium-risk threshold")
			require.LessOrEqual(t, score, tc.expectedMaxScore, "Score should be within medium-risk threshold")

			level := determineLevel(score)

			require.Equal(t, tc.expectedLevel, level)
		})
	}
}

// TestRiskScenario_HighRisk tests high-risk authentication scenarios.
