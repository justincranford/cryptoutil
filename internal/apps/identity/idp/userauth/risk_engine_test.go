// Copyright (c) 2025 Justin Cranford

package userauth

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBehavioralRiskEngine_LoadWeights(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create temporary policy file.
	tempDir := t.TempDir()
	policyFile := filepath.Join(tempDir, "risk_scoring.yml")

	policyContent := `version: "1.0"
risk_factors:
  location:
    weight: 0.30
    description: "Geographic location anomalies"
  device:
    weight: 0.25
    description: "Device fingerprint anomalies"
  time:
    weight: 0.10
    description: "Time-based anomalies"
  behavior:
    weight: 0.15
    description: "User behavior anomalies"
  network:
    weight: 0.10
    description: "VPN/proxy/Tor detection"
  velocity:
    weight: 0.10
    description: "Rapid attempt patterns"
risk_thresholds:
  low:
    min: 0.0
    max: 0.1
    auth_requirements: ["basic"]
    max_session_duration: "24h"
    description: ""
confidence_weights:
  factor_count: 0.5
  baseline_data: 0.15
  behavior_profile: 0.10
  description: ""
network_risks: {}
geographic_risks:
  high_risk_countries:
    countries: []
    score: 0.6
    description: ""
  embargoed_countries:
    countries: []
    score: 0.8
    description: ""
velocity_limits: {}
time_risks: {}
behavior_risks: {}
`

	err := os.WriteFile(policyFile, []byte(policyContent), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	// Create policy loader.
	loader := NewYAMLPolicyLoader(policyFile, "", "")

	// Create mock stores.
	userHistory := &mockUserBehaviorStore{}
	geoIP := &mockGeoIPService{}
	deviceDB := &mockDeviceFingerprintDB{}

	// Create risk engine with policy loader.
	engine := NewBehavioralRiskEngine(loader, userHistory, geoIP, deviceDB)

	// Verify weights are initially zero.
	require.Zero(t, engine.locationWeight)
	require.Zero(t, engine.deviceWeight)

	// Load weights.
	err = engine.loadWeights(ctx)
	require.NoError(t, err)

	// Verify weights loaded from policy.
	require.InDelta(t, 0.30, engine.locationWeight, 0.001)
	require.InDelta(t, cryptoutilSharedMagic.TestProbQuarter, engine.deviceWeight, 0.001)
	require.InDelta(t, cryptoutilSharedMagic.ConfidenceWeightBehavior, engine.timeWeight, 0.001)
	require.InDelta(t, cryptoutilSharedMagic.ConfidenceWeightBaseline, engine.behaviorWeight, 0.001)
	require.InDelta(t, cryptoutilSharedMagic.ConfidenceWeightBehavior, engine.networkWeight, 0.001)
	require.InDelta(t, cryptoutilSharedMagic.ConfidenceWeightBehavior, engine.velocityWeight, 0.001)

	// Load weights again (should use cached values).
	err = engine.loadWeights(ctx)
	require.NoError(t, err)

	// Verify weights unchanged.
	require.InDelta(t, 0.30, engine.locationWeight, 0.001)
}

func TestBehavioralRiskEngine_AssessRisk(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create temporary policy file with standard weights.
	tempDir := t.TempDir()
	policyFile := filepath.Join(tempDir, "risk_scoring.yml")

	policyContent := `version: "1.0"
risk_factors:
  location:
    weight: 0.25
    description: ""
  device:
    weight: 0.20
    description: ""
  time:
    weight: 0.15
    description: ""
  behavior:
    weight: 0.20
    description: ""
  network:
    weight: 0.10
    description: ""
  velocity:
    weight: 0.10
    description: ""
risk_thresholds:
  low:
    min: 0.0
    max: 0.1
    auth_requirements: ["basic"]
    max_session_duration: "24h"
    description: ""
confidence_weights:
  factor_count: 0.5
  baseline_data: 0.15
  behavior_profile: 0.10
  description: ""
network_risks: {}
geographic_risks:
  high_risk_countries:
    countries: []
    score: 0.6
    description: ""
  embargoed_countries:
    countries: []
    score: 0.8
    description: ""
velocity_limits: {}
time_risks: {}
behavior_risks: {}
`

	err := os.WriteFile(policyFile, []byte(policyContent), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	loader := NewYAMLPolicyLoader(policyFile, "", "")

	// Create mock baseline with known location and device.
	baseline := &UserBaseline{
		KnownLocations: []GeoLocation{
			{Country: "US"},
		},
		KnownDevices:  []string{"device-fingerprint-1"},
		TypicalHours:  []int{9, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, 11, 14, 15, cryptoutilSharedMagic.RealmMinTokenLengthBytes},
		KnownNetworks: []string{"192.168.1.1"},
		LastAuthTime:  time.Now().UTC().Add(-1 * time.Hour),
	}

	userHistory := &mockUserBehaviorStore{
		baseline: baseline,
	}

	geoIP := &mockGeoIPService{}
	deviceDB := &mockDeviceFingerprintDB{}

	engine := NewBehavioralRiskEngine(loader, userHistory, geoIP, deviceDB)

	tests := []struct {
		name          string
		authContext   *AuthContext
		wantRiskLevel RiskLevel
		wantLowScore  float64
		wantHighScore float64
	}{
		{
			name: "low-risk known context",
			authContext: &AuthContext{
				Location: &GeoLocation{Country: "US"},
				Device:   &DeviceFingerprint{ID: "device-fingerprint-1"},
				Network:  &NetworkInfo{IPAddress: "192.168.1.1", IsVPN: false, IsProxy: false},
				Time:     time.Date(2025, 1, 15, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, 0, 0, 0, time.UTC), // Wednesday, 10 AM.
			},
			wantRiskLevel: RiskLevelLow,
			wantLowScore:  cryptoutilSharedMagic.BaselineContributionZero,
			wantHighScore: cryptoutilSharedMagic.TestProbQuarter,
		},
		{
			name: "medium-risk new location",
			authContext: &AuthContext{
				Location: &GeoLocation{Country: "CA"}, // New country.
				Device:   &DeviceFingerprint{ID: "device-fingerprint-1"},
				Network:  &NetworkInfo{IPAddress: "192.168.1.1", IsVPN: false, IsProxy: false},
				Time:     time.Date(2025, 1, 15, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, 0, 0, 0, time.UTC),
			},
			wantRiskLevel: RiskLevelMedium,
			wantLowScore:  cryptoutilSharedMagic.TestProbQuarter,
			wantHighScore: 0.50,
		},
		{
			name: "high-risk VPN usage",
			authContext: &AuthContext{
				Location: &GeoLocation{Country: "US"},
				Device:   &DeviceFingerprint{ID: "device-fingerprint-1"},
				Network:  &NetworkInfo{IPAddress: "10.0.0.1", IsVPN: true, IsProxy: false}, // VPN detected.
				Time:     time.Date(2025, 1, 15, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, 0, 0, 0, time.UTC),
			},
			wantRiskLevel: RiskLevelMedium,
			wantLowScore:  cryptoutilSharedMagic.TestProbQuarter,
			wantHighScore: 0.50,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			score, err := engine.AssessRisk(ctx, "user-123", tc.authContext)
			require.NoError(t, err)
			require.NotNil(t, score)
			require.Equal(t, tc.wantRiskLevel, score.Level)
			require.GreaterOrEqual(t, score.Score, tc.wantLowScore)
			require.LessOrEqual(t, score.Score, tc.wantHighScore)
			require.Greater(t, score.Confidence, cryptoutilSharedMagic.BaselineContributionZero)
			require.NotEmpty(t, score.Factors)
		})
	}
}

// TestBehavioralRiskEngine_AssessRisk_VelocityThresholds tests velocity risk scoring.
