// Copyright (c) 2025 Justin Cranford
//
//

package magic

import "time"

// Adaptive authentication timeouts.
const (
	LowRiskAuthMaxAge      = 24 * time.Hour  // Maximum age for low-risk authentication.
	MediumRiskAuthMaxAge   = 1 * time.Hour   // Maximum age for medium-risk authentication.
	HighRiskAuthMaxAge     = 5 * time.Minute // Maximum age for high-risk authentication.
	CriticalRiskAuthMaxAge = 1 * time.Minute // Maximum age for critical-risk authentication.
)

// Step-up policy timeouts.
const (
	StepUpTransferFundsMaxAge  = 5 * time.Minute  // Maximum age for transfer funds authentication.
	StepUpChangePasswordMaxAge = 2 * time.Minute  // Maximum age for change password authentication.
	StepUpAddPayeeMaxAge       = 5 * time.Minute  // Maximum age for add payee authentication.
	StepUpDeleteAccountMaxAge  = 1 * time.Minute  // Maximum age for delete account authentication.
	StepUpViewPIIMaxAge        = 10 * time.Minute // Maximum age for view PII authentication.
)

// Risk score thresholds and weights.
const (
	RiskScoreLow      = 0.1 // Low risk score value.
	RiskScoreMedium   = 0.4 // Medium risk score value.
	RiskScoreHigh     = 0.6 // High risk score value.
	RiskScoreCritical = 0.7 // Critical risk score value.
	RiskScoreVeryHigh = 0.8 // Very high risk score value.
	RiskScoreExtreme  = 0.9 // Extreme risk score value.

	// Confidence score weights.
	ConfidenceWeightFactors  = 0.5  // Confidence contribution from factor count.
	ConfidenceWeightBaseline = 0.15 // Confidence contribution from baseline data.
	ConfidenceWeightBehavior = 0.10 // Confidence contribution from behavior profile.
)

// VPN/Proxy risk scores.
const (
	VPNRiskScore   = 0.5 // Risk score for VPN usage.
	ProxyRiskScore = 0.6 // Risk score for proxy usage.
)

// Baseline contribution values.
const (
	BaselineContributionZero = 0.0 // Zero baseline contribution for initial calculation.
)
