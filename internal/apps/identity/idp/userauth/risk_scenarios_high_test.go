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

func TestRiskScenario_HighRisk(t *testing.T) {
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
			name: "VPN usage, new device, unusual time",
			authContext: &AuthContext{
				DeviceFingerprint: &DeviceFingerprint{
					ID:         "new-device-001",
					UserAgent:  "Mozilla/5.0 (X11; Linux x86_64) Firefox/121.0",
					Platform:   "Linux x86_64",
					Language:   "ru-RU",
					ScreenSize: "1366x768",
					Timezone:   "Europe/Moscow",
				},
				GeoLocation: &GeoLocation{
					IPAddress: "198.51.100.42",
					Country:   "NL",
					City:      "Amsterdam",
					Latitude:  52.3676,
					Longitude: 4.9041,
				},
				NetworkInfo: &NetworkInfo{
					IPAddress: "198.51.100.42",
					IsVPN:     true,
					IsProxy:   false,
					IsTor:     false,
				},
				Timestamp: time.Date(2025, 1, 15, 2, 0, 0, 0, time.UTC), // 2 AM UTC.
			},
			baseline: &UserBaseline{
				UserID: googleUuid.New().String(),
				KnownLocations: []GeoLocation{
					{Country: "US", City: "San Francisco"},
				},
				KnownDevices: []DeviceFingerprint{
					{ID: "known-device-001"},
				},
				TypicalLoginHours: []int{9, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, 11, cryptoutilSharedMagic.HashPrefixLength, 13, 14, 15, cryptoutilSharedMagic.RealmMinTokenLengthBytes, 17, 18},
				LastLoginTime:     time.Now().UTC().Add(-cryptoutilSharedMagic.HMACSHA384KeySize * time.Hour),
				EstablishedAt:     time.Now().UTC().Add(-cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days * cryptoutilSharedMagic.HoursPerDay * time.Hour),
				EventCount:        cryptoutilSharedMagic.IMMaxUsernameLength,
			},
			expectedMinScore: cryptoutilSharedMagic.Tolerance50Percent,
			expectedMaxScore: cryptoutilSharedMagic.RiskScoreVeryHigh,
			expectedLevel:    RiskLevelHigh,
		},
		{
			name: "proxy usage, new location, known device",
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
					IPAddress: "192.0.2.100",
					Country:   "SG",
					City:      "Singapore",
					Latitude:  1.3521,
					Longitude: 103.8198,
				},
				NetworkInfo: &NetworkInfo{
					IPAddress: "192.0.2.100",
					IsVPN:     false,
					IsProxy:   true,
					IsTor:     false,
				},
				Timestamp: time.Date(2025, 1, 15, 14, 0, 0, 0, time.UTC),
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
				LastLoginTime:     time.Now().UTC().Add(-cryptoutilSharedMagic.HashPrefixLength * time.Hour),
				EstablishedAt:     time.Now().UTC().Add(-cryptoutilSharedMagic.StrictCertificateMaxAgeDays * cryptoutilSharedMagic.HoursPerDay * time.Hour),
				EventCount:        150,
			},
			expectedMinScore: cryptoutilSharedMagic.RiskScoreMedium,
			expectedMaxScore: cryptoutilSharedMagic.RiskScoreCritical,
			expectedLevel:    RiskLevelHigh,
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

			require.GreaterOrEqual(t, score, tc.expectedMinScore, "Score should be at least minimum high-risk threshold")
			require.LessOrEqual(t, score, tc.expectedMaxScore, "Score should be within high-risk threshold")

			level := determineLevel(score)

			require.Equal(t, tc.expectedLevel, level)
		})
	}
}

// TestRiskScenario_CriticalRisk tests critical-risk authentication scenarios.
func TestRiskScenario_CriticalRisk(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		authContext      *AuthContext
		baseline         *UserBaseline
		expectedMinScore float64
		expectedLevel    RiskLevel
	}{
		{
			name: "Tor network, new device, high-risk country",
			authContext: &AuthContext{
				DeviceFingerprint: &DeviceFingerprint{
					ID:         "new-device-tor-001",
					UserAgent:  "Mozilla/5.0 (Windows NT 6.1; rv:60.0) Gecko/20100101 Firefox/60.0",
					Platform:   "Win32",
					Language:   "en-US",
					ScreenSize: "1024x768",
					Timezone:   "UTC",
				},
				GeoLocation: &GeoLocation{
					IPAddress: "185.220.101.1",
					Country:   "RU",
					City:      "Moscow",
					Latitude:  55.7558,
					Longitude: 37.6173,
				},
				NetworkInfo: &NetworkInfo{
					IPAddress: "185.220.101.1",
					IsVPN:     false,
					IsProxy:   false,
					IsTor:     true,
				},
				Timestamp: time.Date(2025, 1, 15, 4, 0, 0, 0, time.UTC),
			},
			baseline: &UserBaseline{
				UserID: googleUuid.New().String(),
				KnownLocations: []GeoLocation{
					{Country: "US", City: "Seattle"},
				},
				KnownDevices: []DeviceFingerprint{
					{ID: "known-device-001"},
				},
				TypicalLoginHours: []int{cryptoutilSharedMagic.IMMinPasswordLength, 9, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, 11, cryptoutilSharedMagic.HashPrefixLength, 13, 14, 15, cryptoutilSharedMagic.RealmMinTokenLengthBytes, 17},
				LastLoginTime:     time.Now().UTC().Add(-72 * time.Hour),
				EstablishedAt:     time.Now().UTC().Add(-cryptoutilSharedMagic.CertificateRandomizationNotBeforeMinutes * cryptoutilSharedMagic.HoursPerDay * time.Hour),
				EventCount:        200,
			},
			expectedMinScore: cryptoutilSharedMagic.RiskScoreVeryHigh,
			expectedLevel:    RiskLevelCritical,
		},
		{
			name: "velocity anomaly - 5 locations in 1 hour",
			authContext: &AuthContext{
				DeviceFingerprint: &DeviceFingerprint{
					ID:         "device-velocity-test",
					UserAgent:  "Mozilla/5.0 (iPhone; CPU iPhone OS 17_2 like Mac OS X) Safari/17.2",
					Platform:   "iPhone",
					Language:   "en-US",
					ScreenSize: "390x844",
					Timezone:   "America/New_York",
				},
				GeoLocation: &GeoLocation{
					IPAddress: "192.0.2.200",
					Country:   "JP",
					City:      "Tokyo",
					Latitude:  35.6762,
					Longitude: 139.6503,
				},
				NetworkInfo: &NetworkInfo{
					IPAddress: "192.0.2.200",
					IsVPN:     true,
					IsProxy:   false,
					IsTor:     false,
				},
				Timestamp: time.Date(2025, 1, 15, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, 0, 0, 0, time.UTC),
				Metadata: map[string]any{
					"velocity_anomaly": true,
					"location_count":   cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries,
					"time_window":      "1h",
				},
			},
			baseline: &UserBaseline{
				UserID: googleUuid.New().String(),
				KnownLocations: []GeoLocation{
					{Country: "US", City: "Boston"},
				},
				KnownDevices: []DeviceFingerprint{
					{ID: "device-velocity-test"},
				},
				TypicalLoginHours: []int{cryptoutilSharedMagic.GitRecentActivityDays, cryptoutilSharedMagic.IMMinPasswordLength, 9, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, 11, cryptoutilSharedMagic.HashPrefixLength, 13, 14, 15, cryptoutilSharedMagic.RealmMinTokenLengthBytes},
				LastLoginTime:     time.Now().UTC().Add(-1 * time.Hour),
				EstablishedAt:     time.Now().UTC().Add(-cryptoutilSharedMagic.IdentityDefaultIdleTimeoutSeconds * cryptoutilSharedMagic.HoursPerDay * time.Hour),
				EventCount:        cryptoutilSharedMagic.JoseJAMaxMaterials,
			},
			expectedMinScore: 0.85,
			expectedLevel:    RiskLevelCritical,
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

			require.GreaterOrEqual(t, score, tc.expectedMinScore, "Score should be at least minimum critical-risk threshold")

			level := determineLevel(score)

			require.Equal(t, tc.expectedLevel, level)
		})
	}
}

// Helper functions for risk scoring tests.

func calculateScore(factors []RiskFactor, engine *BehavioralRiskEngine) float64 {
	score := cryptoutilSharedMagic.BaselineContributionZero

	for _, factor := range factors {
		score += factor.Score * factor.Weight
	}

	return math.Min(score, cryptoutilSharedMagic.TestProbAlways)
}

func determineLevel(score float64) RiskLevel {
	switch {
	case score >= cryptoutilIdentityMagic.RiskScoreCriticalThreshold:
		return RiskLevelCritical
	case score >= cryptoutilIdentityMagic.RiskScoreHighThreshold:
		return RiskLevelHigh
	case score >= cryptoutilIdentityMagic.RiskScoreMediumThreshold:
		return RiskLevelMedium
	default:
		return RiskLevelLow
	}
}

func calculateConfidence(baseline *UserBaseline) float64 {
	const (
		minEvents   = 10
		maxEvents   = 100
		maxAgeDays  = 90
		factorCount = 0.4
		baselineAge = 0.3
		eventVolume = 0.3
	)

	// Factor 1: Event count (more events = higher confidence).
	eventScore := math.Min(float64(baseline.EventCount)/float64(maxEvents), cryptoutilSharedMagic.TestProbAlways)

	// Factor 2: Baseline age (older baseline = higher confidence).
	ageDays := time.Since(baseline.EstablishedAt).Hours() / cryptoutilSharedMagic.HoursPerDay //nolint:mnd
	ageScore := math.Min(ageDays/float64(maxAgeDays), cryptoutilSharedMagic.TestProbAlways)

	// Factor 3: Factor count (more known patterns = higher confidence).
	factorCountScore := cryptoutilSharedMagic.Tolerance50Percent
	if len(baseline.KnownLocations) > 0 && len(baseline.KnownDevices) > 0 && len(baseline.TypicalLoginHours) > 0 {
		factorCountScore = cryptoutilSharedMagic.TestProbAlways
	}

	confidence := (eventScore * eventVolume) + (ageScore * baselineAge) + (factorCountScore * factorCount)

	return math.Min(confidence, cryptoutilSharedMagic.TestProbAlways)
}
