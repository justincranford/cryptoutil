// Copyright (c) 2025 Justin Cranford

package userauth

import (
	"context"
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

	err := os.WriteFile(policyFile, []byte(policyContent), 0o600)
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
	require.InDelta(t, 0.25, engine.deviceWeight, 0.001)
	require.InDelta(t, 0.10, engine.timeWeight, 0.001)
	require.InDelta(t, 0.15, engine.behaviorWeight, 0.001)
	require.InDelta(t, 0.10, engine.networkWeight, 0.001)
	require.InDelta(t, 0.10, engine.velocityWeight, 0.001)

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

	err := os.WriteFile(policyFile, []byte(policyContent), 0o600)
	require.NoError(t, err)

	loader := NewYAMLPolicyLoader(policyFile, "", "")

	// Create mock baseline with known location and device.
	baseline := &UserBaseline{
		KnownLocations: []GeoLocation{
			{Country: "US"},
		},
		KnownDevices:  []string{"device-fingerprint-1"},
		TypicalHours:  []int{9, 10, 11, 14, 15, 16},
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
				Time:     time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), // Wednesday, 10 AM.
			},
			wantRiskLevel: RiskLevelLow,
			wantLowScore:  0.0,
			wantHighScore: 0.25,
		},
		{
			name: "medium-risk new location",
			authContext: &AuthContext{
				Location: &GeoLocation{Country: "CA"}, // New country.
				Device:   &DeviceFingerprint{ID: "device-fingerprint-1"},
				Network:  &NetworkInfo{IPAddress: "192.168.1.1", IsVPN: false, IsProxy: false},
				Time:     time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
			},
			wantRiskLevel: RiskLevelMedium,
			wantLowScore:  0.25,
			wantHighScore: 0.50,
		},
		{
			name: "high-risk VPN usage",
			authContext: &AuthContext{
				Location: &GeoLocation{Country: "US"},
				Device:   &DeviceFingerprint{ID: "device-fingerprint-1"},
				Network:  &NetworkInfo{IPAddress: "10.0.0.1", IsVPN: true, IsProxy: false}, // VPN detected.
				Time:     time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
			},
			wantRiskLevel: RiskLevelMedium,
			wantLowScore:  0.25,
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
			require.Greater(t, score.Confidence, 0.0)
			require.NotEmpty(t, score.Factors)
		})
	}
}

// TestBehavioralRiskEngine_AssessRisk_VelocityThresholds tests velocity risk scoring.
func TestBehavioralRiskEngine_AssessRisk_VelocityThresholds(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create temporary policy file.
	tempDir := t.TempDir()
	policyFile := filepath.Join(tempDir, "risk_scoring.yml")

	policyContent := `version: "1.0"
risk_factors:
  velocity:
    weight: 1.0
    description: "Velocity risk testing"
risk_thresholds:
  low:
    min: 0.0
    max: 0.25
  medium:
    min: 0.25
    max: 0.50
  high:
    min: 0.50
    max: 0.75
  extreme:
    min: 0.75
    max: 1.0
confidence_weights:
  factor_count: 1.0
network_risks: {}
geographic_risks:
  high_risk_countries:
    countries: []
  embargoed_countries:
    countries: []
velocity_limits: {}
time_risks: {}
behavior_risks: {}
`

	err := os.WriteFile(policyFile, []byte(policyContent), 0o600)
	require.NoError(t, err)

	loader := NewYAMLPolicyLoader(policyFile, "", "")
	geoIP := &mockGeoIPService{}
	deviceDB := &mockDeviceFingerprintDB{}

	tests := []struct {
		name              string
		lastAuthTime      time.Time
		currentAuthTime   time.Time
		wantRiskLevel     RiskLevel
		wantMinScore      float64
		wantMaxScore      float64
		velocityRiskScore float64
	}{
		{
			name:              "no previous auth (first login)",
			lastAuthTime:      time.Time{}, // Zero time.
			currentAuthTime:   time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
			wantRiskLevel:     RiskLevelLow,
			wantMinScore:      0.0,
			wantMaxScore:      0.25,
			velocityRiskScore: 0.1, // Low risk.
		},
		{
			name:              "very fast auth (<5s) - critical risk",
			lastAuthTime:      time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
			currentAuthTime:   time.Date(2025, 1, 15, 10, 0, 3, 0, time.UTC), // 3 seconds later.
			wantRiskLevel:     RiskLevelCritical,
			wantMinScore:      0.85,
			wantMaxScore:      1.0,
			velocityRiskScore: 1.0, // Extreme risk.
		},
		{
			name:              "fast auth (<1min) - high risk",
			lastAuthTime:      time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
			currentAuthTime:   time.Date(2025, 1, 15, 10, 0, 30, 0, time.UTC), // 30 seconds later.
			wantRiskLevel:     RiskLevelHigh,
			wantMinScore:      0.5,
			wantMaxScore:      0.9,
			velocityRiskScore: 0.8, // High risk.
		},
		{
			name:              "normal auth (>1min) - low risk",
			lastAuthTime:      time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
			currentAuthTime:   time.Date(2025, 1, 15, 10, 5, 0, 0, time.UTC), // 5 minutes later.
			wantRiskLevel:     RiskLevelLow,
			wantMinScore:      0.0,
			wantMaxScore:      0.25,
			velocityRiskScore: 0.2, // Normal/low risk.
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create baseline with specific last auth time.
			baseline := &UserBaseline{
				KnownLocations: []GeoLocation{{Country: "US"}},
				KnownDevices:   []string{"device-1"},
				KnownNetworks:  []string{"192.168.1.1"},
				TypicalHours:   []int{9, 10, 11},
				LastAuthTime:   tc.lastAuthTime,
			}

			userHistory := &mockUserBehaviorStore{baseline: baseline}
			engine := NewBehavioralRiskEngine(loader, userHistory, geoIP, deviceDB)

			authContext := &AuthContext{
				Location: &GeoLocation{Country: "US"},
				Device:   &DeviceFingerprint{ID: "device-1"},
				Network:  &NetworkInfo{IPAddress: "192.168.1.1"},
				Time:     tc.currentAuthTime,
			}

			score, err := engine.AssessRisk(ctx, "user-velocity-test", authContext)
			require.NoError(t, err)
			require.NotNil(t, score)
			require.Equal(t, tc.wantRiskLevel, score.Level, "Risk level should match expected")
			require.GreaterOrEqual(t, score.Score, tc.wantMinScore, "Score should be >= min")
			require.LessOrEqual(t, score.Score, tc.wantMaxScore, "Score should be <= max")
		})
	}
}

// Mock implementations for testing.

type mockUserBehaviorStore struct {
	baseline *UserBaseline
}

func (m *mockUserBehaviorStore) GetBaseline(_ context.Context, _ string) (*UserBaseline, error) {
	if m.baseline != nil {
		return m.baseline, nil
	}

	return &UserBaseline{}, nil
}

func (m *mockUserBehaviorStore) UpdateBaseline(_ context.Context, _ string, _ *AuthContext) error {
	return nil
}

func (m *mockUserBehaviorStore) RecordAuthentication(_ context.Context, _ string, _ bool, _ *AuthContext) error {
	return nil
}

type mockGeoIPService struct{}

func (m *mockGeoIPService) Lookup(_ context.Context, _ string) (*GeoLocation, error) {
	return &GeoLocation{Country: "US"}, nil
}

type mockDeviceFingerprintDB struct{}

func (m *mockDeviceFingerprintDB) GetFingerprint(_ context.Context, _ string, _ map[string]string) (*DeviceFingerprint, error) {
	return &DeviceFingerprint{ID: "test-device"}, nil
}

func (m *mockDeviceFingerprintDB) StoreFingerprint(_ context.Context, _ *DeviceFingerprint) error {
	return nil
}

// TestBehavioralRiskEngine_assessBehaviorRisk tests the assessBehaviorRisk method.
func TestBehavioralRiskEngine_assessBehaviorRisk(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		behavior  *UserBehavior
		baseline  *UserBaseline
		wantScore float64
	}{
		{
			name:     "nil behavior profile returns moderate risk",
			behavior: &UserBehavior{},
			baseline: &UserBaseline{
				BehaviorProfile: nil,
			},
			wantScore: 0.40 + 0.10, // RiskScoreMedium + RiskScoreLow.
		},
		{
			name:     "with behavior profile returns low risk",
			behavior: &UserBehavior{},
			baseline: &UserBaseline{
				BehaviorProfile: &BehaviorProfile{
					AverageSessionLength: 3600,
				},
			},
			wantScore: 0.10, // RiskScoreLow.
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			engine := &BehavioralRiskEngine{}

			score := engine.assessBehaviorRisk(tc.behavior, tc.baseline)
			require.InDelta(t, tc.wantScore, score, 0.001)
		})
	}
}

// TestRiskBasedAuthenticator_Authenticate tests the Authenticate method.
func TestRiskBasedAuthenticator_Authenticate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name             string
		riskLevel        RiskLevel
		wantRequiresAuth bool
		wantRequirements int
	}{
		{
			name:             "low risk does not require additional auth",
			riskLevel:        RiskLevelLow,
			wantRequiresAuth: false,
			wantRequirements: 1,
		},
		{
			name:             "medium risk requires MFA",
			riskLevel:        RiskLevelMedium,
			wantRequiresAuth: true,
			wantRequirements: 2,
		},
		{
			name:             "high risk requires strong auth",
			riskLevel:        RiskLevelHigh,
			wantRequiresAuth: true,
			wantRequirements: 2,
		},
		{
			name:             "critical risk requires strong auth",
			riskLevel:        RiskLevelCritical,
			wantRequiresAuth: true,
			wantRequirements: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create mock risk engine that returns configured risk level.
			riskEngine := &mockRiskEngine{
				riskScore: &RiskScore{
					Level:      tc.riskLevel,
					Score:      0.5,
					Confidence: 0.8,
					Factors: []RiskFactor{
						{Type: "test", Score: 0.5, Weight: 1.0, Reason: "Test factor"},
					},
				},
			}

			// Create mock context analyzer.
			contextAnalyzer := &mockContextAnalyzer{
				authContext: &AuthContext{
					Location: &GeoLocation{Country: "US"},
					Device:   &DeviceFingerprint{ID: "test-device"},
					Network:  &NetworkInfo{IPAddress: "192.168.1.1"},
					Time:     time.Now().UTC(),
				},
			}

			// Create mock user behavior store.
			userBehaviorStore := &mockUserBehaviorStore{}

			// Create authenticator.
			thresholds := DefaultRiskThresholds()
			auth := NewRiskBasedAuthenticator(riskEngine, contextAnalyzer, nil, thresholds, userBehaviorStore)

			// Create auth request.
			authRequest := &AuthRequest{
				IPAddress: "192.168.1.1",
				UserAgent: "Test Browser",
			}

			// Call Authenticate.
			decision, err := auth.Authenticate(ctx, "test-user", authRequest)
			require.NoError(t, err, "Authenticate should succeed")
			require.NotNil(t, decision, "Decision should not be nil")
			require.Equal(t, tc.wantRequiresAuth, decision.RequiresAuth, "RequiresAuth should match")
			require.NotNil(t, decision.Requirements, "Requirements should not be nil")
			require.Equal(t, tc.wantRequirements, decision.Requirements.MinFactors, "MinFactors should match")
			require.NotNil(t, decision.RiskScore, "RiskScore should not be nil")
			require.Equal(t, tc.riskLevel, decision.RiskScore.Level, "RiskLevel should match")
		})
	}
}

// mockRiskEngine implements RiskEngine for testing.
type mockRiskEngine struct {
	riskScore *RiskScore
	err       error
}

func (m *mockRiskEngine) AssessRisk(_ context.Context, _ string, _ *AuthContext) (*RiskScore, error) {
	if m.err != nil {
		return nil, m.err
	}

	return m.riskScore, nil
}

func (m *mockRiskEngine) CalculateRiskFactors(_ *AuthContext, _ *UserBaseline) []RiskFactor {
	return m.riskScore.Factors
}

// mockContextAnalyzer implements ContextAnalyzer for testing.
type mockContextAnalyzer struct {
	authContext *AuthContext
	err         error
}

func (m *mockContextAnalyzer) AnalyzeContext(_ context.Context, _ *AuthRequest) (*AuthContext, error) {
	if m.err != nil {
		return nil, m.err
	}

	return m.authContext, nil
}

func (m *mockContextAnalyzer) DetectAnomalies(_ context.Context, _ *AuthContext, _ *UserBaseline) ([]Anomaly, error) {
	return nil, nil
}
