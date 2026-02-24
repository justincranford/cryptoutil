// Copyright (c) 2025 Justin Cranford
//
//

package userauth

import (
	"context"
	"fmt"
	"math"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// RiskLevel represents the assessed risk level.
type RiskLevel string

// Risk level constants.
const (
	// RiskLevelLow indicates a low risk assessment.
	RiskLevelLow RiskLevel = "low"
	// RiskLevelMedium indicates a medium risk assessment.
	RiskLevelMedium RiskLevel = "medium"
	// RiskLevelHigh indicates a high risk assessment.
	RiskLevelHigh RiskLevel = "high"
	// RiskLevelCritical indicates a critical risk assessment.
	RiskLevelCritical RiskLevel = "critical"
)

// String returns the string representation of the risk level.
func (r RiskLevel) String() string {
	return string(r)
}

// RiskFactorType represents the type of risk factor.
type RiskFactorType string

// Risk factor type constants.
const (
	// RiskFactorLocation indicates a location-based risk factor.
	RiskFactorLocation RiskFactorType = "location"
	// RiskFactorDevice indicates a device-based risk factor.
	RiskFactorDevice RiskFactorType = "device"
	// RiskFactorTime indicates a time-based risk factor.
	RiskFactorTime RiskFactorType = "time"
	// RiskFactorBehavior indicates a behavior-based risk factor.
	RiskFactorBehavior RiskFactorType = "behavior"
	// RiskFactorNetwork indicates a network-based risk factor.
	RiskFactorNetwork RiskFactorType = "network"
	// RiskFactorVelocity indicates a velocity-based risk factor.
	RiskFactorVelocity       RiskFactorType = "velocity"
	RiskFactorAuthentication RiskFactorType = "authentication"
)

// RiskFactor represents an individual risk factor contributing to the overall risk score.
type RiskFactor struct {
	Type     RiskFactorType
	Score    float64
	Weight   float64
	Reason   string
	Metadata map[string]any
}

// RiskScore represents the overall risk assessment result.
type RiskScore struct {
	Score      float64      // Overall risk score (0.0-1.0).
	Level      RiskLevel    // Risk level categorization.
	Factors    []RiskFactor // Individual risk factors.
	Confidence float64      // Confidence in the assessment (0.0-1.0).
	AssessedAt time.Time    // When the assessment was performed.
}

// RiskEngine assesses authentication risk based on context.
type RiskEngine interface {
	AssessRisk(ctx context.Context, userID string, authContext *AuthContext) (*RiskScore, error)
	CalculateRiskFactors(authContext *AuthContext, baseline *UserBaseline) []RiskFactor
}

// BehavioralRiskEngine assesses authentication risk based on behavioral patterns using configurable policies.
type BehavioralRiskEngine struct {
	policyLoader PolicyLoader
	userHistory  UserBehaviorStore
	geoIP        GeoIPService
	deviceDB     DeviceFingerprintDB

	// Weights loaded from policy (cached after first load).
	locationWeight float64
	deviceWeight   float64
	timeWeight     float64
	behaviorWeight float64
	networkWeight  float64
	velocityWeight float64
}

// NewBehavioralRiskEngine creates a new behavioral risk engine with policy-driven configuration.
func NewBehavioralRiskEngine(
	policyLoader PolicyLoader,
	userHistory UserBehaviorStore,
	geoIP GeoIPService,
	deviceDB DeviceFingerprintDB,
) *BehavioralRiskEngine {
	return &BehavioralRiskEngine{
		policyLoader: policyLoader,
		userHistory:  userHistory,
		geoIP:        geoIP,
		deviceDB:     deviceDB,
	}
}

// loadWeights loads risk factor weights from policy if not already loaded.
func (e *BehavioralRiskEngine) loadWeights(ctx context.Context) error {
	// Check if weights already loaded.
	if e.locationWeight > 0 {
		return nil
	}

	// Load risk scoring policy.
	policy, err := e.policyLoader.LoadRiskScoringPolicy(ctx)
	if err != nil {
		return fmt.Errorf("failed to load risk scoring policy: %w", err)
	}

	// Extract weights from policy.
	if location, ok := policy.RiskFactors["location"]; ok {
		e.locationWeight = location.Weight
	}

	if device, ok := policy.RiskFactors["device"]; ok {
		e.deviceWeight = device.Weight
	}

	if timeRisk, ok := policy.RiskFactors["time"]; ok {
		e.timeWeight = timeRisk.Weight
	}

	if behavior, ok := policy.RiskFactors["behavior"]; ok {
		e.behaviorWeight = behavior.Weight
	}

	if network, ok := policy.RiskFactors["network"]; ok {
		e.networkWeight = network.Weight
	}

	if velocity, ok := policy.RiskFactors["velocity"]; ok {
		e.velocityWeight = velocity.Weight
	}

	return nil
}

// AssessRisk evaluates the overall risk for an authentication attempt using policy-driven weights.
func (e *BehavioralRiskEngine) AssessRisk(ctx context.Context, userID string, authContext *AuthContext) (*RiskScore, error) {
	// Load weights from policy if not already loaded.
	if err := e.loadWeights(ctx); err != nil {
		return nil, fmt.Errorf("failed to load policy weights: %w", err)
	}

	// Get user baseline.
	baseline, err := e.userHistory.GetBaseline(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user baseline: %w", err)
	}

	// Calculate individual risk factors.
	factors := e.CalculateRiskFactors(authContext, baseline)

	// Calculate weighted risk score.
	var totalScore float64

	var totalWeight float64

	for _, factor := range factors {
		totalScore += factor.Score * factor.Weight
		totalWeight += factor.Weight
	}

	// Normalize score to 0.0-1.0 range.
	score := totalScore / totalWeight

	// Determine risk level.
	level := e.determineRiskLevel(score)

	// Calculate confidence based on baseline data availability.
	confidence := e.calculateConfidence(baseline, len(factors))

	return &RiskScore{
		Score:      score,
		Level:      level,
		Factors:    factors,
		Confidence: confidence,
		AssessedAt: time.Now().UTC(),
	}, nil
}

// CalculateRiskFactors computes individual risk factors.
func (e *BehavioralRiskEngine) CalculateRiskFactors(authContext *AuthContext, baseline *UserBaseline) []RiskFactor {
	factors := make([]RiskFactor, 0)

	// Location risk.
	if authContext.Location != nil {
		locationRisk := e.assessLocationRisk(authContext.Location, baseline)
		factors = append(factors, RiskFactor{
			Type:   RiskFactorLocation,
			Score:  locationRisk,
			Weight: e.locationWeight,
			Reason: "Location-based risk assessment",
			Metadata: map[string]any{
				"country": authContext.Location.Country,
				"city":    authContext.Location.City,
			},
		})
	}

	// Device risk.
	if authContext.Device != nil {
		deviceRisk := e.assessDeviceRisk(authContext.Device, baseline)
		factors = append(factors, RiskFactor{
			Type:   RiskFactorDevice,
			Score:  deviceRisk,
			Weight: e.deviceWeight,
			Reason: "Device fingerprint risk assessment",
			Metadata: map[string]any{
				"device_id": authContext.Device.ID,
			},
		})
	}

	// Time-based risk.
	timeRisk := e.assessTimeRisk(authContext.Time, baseline)
	factors = append(factors, RiskFactor{
		Type:   RiskFactorTime,
		Score:  timeRisk,
		Weight: e.timeWeight,
		Reason: "Time-based risk assessment",
		Metadata: map[string]any{
			"hour": authContext.Time.Hour(),
		},
	})

	// Behavioral risk.
	if authContext.Behavior != nil {
		behaviorRisk := e.assessBehaviorRisk(authContext.Behavior, baseline)
		factors = append(factors, RiskFactor{
			Type:   RiskFactorBehavior,
			Score:  behaviorRisk,
			Weight: e.behaviorWeight,
			Reason: "Behavioral pattern risk assessment",
		})
	}

	// Network risk.
	if authContext.Network != nil {
		networkRisk := e.assessNetworkRisk(authContext.Network, baseline)
		factors = append(factors, RiskFactor{
			Type:   RiskFactorNetwork,
			Score:  networkRisk,
			Weight: e.networkWeight,
			Reason: "Network-based risk assessment",
			Metadata: map[string]any{
				"ip": authContext.Network.IPAddress,
			},
		})
	}

	// Velocity risk (how quickly authentication attempts occur).
	velocityRisk := e.assessVelocityRisk(authContext, baseline)
	factors = append(factors, RiskFactor{
		Type:   RiskFactorVelocity,
		Score:  velocityRisk,
		Weight: e.velocityWeight,
		Reason: "Authentication velocity risk assessment",
	})

	return factors
}

// Helper methods for individual risk assessments.

func (e *BehavioralRiskEngine) assessLocationRisk(location *GeoLocation, baseline *UserBaseline) float64 {
	// Check if location is in user's known locations.
	for _, knownLoc := range baseline.KnownLocations {
		if knownLoc.Country == location.Country && knownLoc.City == location.City {
			return cryptoutilSharedMagic.RiskScoreLow // Low risk for known location.
		}
	}

	// Unknown location - higher risk.
	return cryptoutilSharedMagic.RiskScoreCritical
}

func (e *BehavioralRiskEngine) assessDeviceRisk(device *DeviceFingerprint, baseline *UserBaseline) float64 {
	// Check if device is recognized.
	for _, knownDevice := range baseline.KnownDevices {
		if knownDevice == device.ID {
			return cryptoutilSharedMagic.RiskScoreLow // Low risk for known device.
		}
	}

	// Unknown device - higher risk.
	return cryptoutilSharedMagic.RiskScoreHigh
}

func (e *BehavioralRiskEngine) assessTimeRisk(authTime time.Time, baseline *UserBaseline) float64 {
	hour := authTime.Hour()

	// Check if time matches user's typical authentication hours.
	for _, typicalHour := range baseline.TypicalHours {
		if hour == typicalHour {
			return cryptoutilSharedMagic.RiskScoreLow // Low risk for typical time.
		}
	}

	// Unusual time - moderate risk.
	return cryptoutilSharedMagic.RiskScoreMedium
}

func (e *BehavioralRiskEngine) assessBehaviorRisk(_ *UserBehavior, baseline *UserBaseline) float64 {
	// Compare behavior patterns with baseline.
	// This is simplified - real implementation would use ML/statistical analysis.
	if baseline.BehaviorProfile == nil {
		return cryptoutilSharedMagic.RiskScoreMedium + cryptoutilSharedMagic.RiskScoreLow // No baseline, moderate risk.
	}

	// Simple similarity check (real implementation would be more sophisticated).
	return cryptoutilSharedMagic.RiskScoreLow
}

func (e *BehavioralRiskEngine) assessNetworkRisk(network *NetworkInfo, baseline *UserBaseline) float64 {
	// Check if network is recognized.
	for _, knownNet := range baseline.KnownNetworks {
		if knownNet == network.IPAddress {
			return cryptoutilSharedMagic.RiskScoreLow // Low risk for known network.
		}
	}

	// Check if using VPN/Proxy.
	if network.IsVPN {
		return cryptoutilSharedMagic.VPNRiskScore
	}

	if network.IsProxy {
		return cryptoutilSharedMagic.ProxyRiskScore
	}

	// Unknown network - moderate risk.
	return cryptoutilSharedMagic.RiskScoreMedium
}

func (e *BehavioralRiskEngine) assessVelocityRisk(authContext *AuthContext, baseline *UserBaseline) float64 {
	if baseline.LastAuthTime.IsZero() {
		return cryptoutilSharedMagic.RiskScoreLow // No previous authentication, low risk.
	}

	// Calculate time since last authentication.
	timeSinceLastAuth := authContext.Time.Sub(baseline.LastAuthTime)

	const (
		veryFastThreshold = 5 * time.Second
		fastThreshold     = 1 * time.Minute
	)

	if timeSinceLastAuth < veryFastThreshold {
		return cryptoutilSharedMagic.RiskScoreExtreme
	}

	if timeSinceLastAuth < fastThreshold {
		return cryptoutilSharedMagic.RiskScoreHigh
	}

	const velocityRiskNormal = 0.2

	return velocityRiskNormal
}

func (e *BehavioralRiskEngine) determineRiskLevel(score float64) RiskLevel {
	const (
		lowThreshold    = 0.25
		mediumThreshold = 0.50
		highThreshold   = 0.75
	)

	switch {
	case score < lowThreshold:
		return RiskLevelLow
	case score < mediumThreshold:
		return RiskLevelMedium
	case score < highThreshold:
		return RiskLevelHigh
	default:
		return RiskLevelCritical
	}
}

func (e *BehavioralRiskEngine) calculateConfidence(baseline *UserBaseline, factorCount int) float64 {
	// Confidence based on:
	// 1. Number of available risk factors.
	// 2. Quality of baseline data.
	// Factor count contribution (max 50%).
	const maxFactors = 6.0

	factorContribution := math.Min(float64(factorCount)/maxFactors, 1.0) * cryptoutilSharedMagic.ConfidenceWeightFactors

	// Baseline quality contribution (max 50%).
	baselineContribution := cryptoutilSharedMagic.BaselineContributionZero
	if len(baseline.KnownLocations) > 0 {
		baselineContribution += cryptoutilSharedMagic.ConfidenceWeightBaseline
	}

	if len(baseline.KnownDevices) > 0 {
		baselineContribution += cryptoutilSharedMagic.ConfidenceWeightBaseline
	}

	if len(baseline.TypicalHours) > 0 {
		baselineContribution += cryptoutilSharedMagic.ConfidenceWeightBehavior
	}

	if baseline.BehaviorProfile != nil {
		baselineContribution += cryptoutilSharedMagic.ConfidenceWeightBehavior
	}

	return factorContribution + baselineContribution
}
