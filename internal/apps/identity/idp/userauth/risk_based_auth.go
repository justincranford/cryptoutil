// Copyright (c) 2025 Justin Cranford
//
//

package userauth

import (
	"context"
	"fmt"
	"time"

	googleUuid "github.com/google/uuid"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// AuthRequirements specifies authentication requirements based on risk level.
type AuthRequirements struct {
	MinFactors        int           // Minimum number of authentication factors required.
	AllowedMethods    []string      // Allowed authentication methods.
	RequiresBiometric bool          // Whether biometric authentication is required.
	RequiresHardware  bool          // Whether hardware key is required.
	MaxAge            time.Duration // Maximum age of existing authentication.
}

// AuthDecision represents the decision from risk-based authentication.
type AuthDecision struct {
	Allowed      bool
	RequiresAuth bool
	Requirements *AuthRequirements
	RiskScore    *RiskScore
	Reason       string
}

// RiskBasedAuthenticator implements risk-based authentication.
type RiskBasedAuthenticator struct {
	riskEngine        RiskEngine
	contextAnalyzer   ContextAnalyzer
	challengeStore    ChallengeStore
	thresholds        map[RiskLevel]*AuthRequirements
	userBehaviorStore UserBehaviorStore
}

// NewRiskBasedAuthenticator creates a new risk-based authenticator.
func NewRiskBasedAuthenticator(
	riskEngine RiskEngine,
	contextAnalyzer ContextAnalyzer,
	challengeStore ChallengeStore,
	thresholds map[RiskLevel]*AuthRequirements,
	userBehaviorStore UserBehaviorStore,
) *RiskBasedAuthenticator {
	if thresholds == nil {
		thresholds = DefaultRiskThresholds()
	}

	return &RiskBasedAuthenticator{
		riskEngine:        riskEngine,
		contextAnalyzer:   contextAnalyzer,
		challengeStore:    challengeStore,
		thresholds:        thresholds,
		userBehaviorStore: userBehaviorStore,
	}
}

// Method returns the authentication method name.
func (r *RiskBasedAuthenticator) Method() string {
	return "risk_based"
}

// Authenticate performs risk-based authentication.
func (r *RiskBasedAuthenticator) Authenticate(
	ctx context.Context,
	userID string,
	authRequest *AuthRequest,
) (*AuthDecision, error) {
	// Analyze authentication context.
	authContext, err := r.contextAnalyzer.AnalyzeContext(ctx, authRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze context: %w", err)
	}

	// Assess risk.
	riskScore, err := r.riskEngine.AssessRisk(ctx, userID, authContext)
	if err != nil {
		return nil, fmt.Errorf("failed to assess risk: %w", err)
	}

	// Get authentication requirements based on risk level.
	requirements, found := r.thresholds[riskScore.Level]
	if !found {
		requirements = r.thresholds[RiskLevelMedium] // Default to medium.
	}

	// Make authentication decision.
	decision := &AuthDecision{
		Allowed:      true,
		RequiresAuth: false,
		Requirements: requirements,
		RiskScore:    riskScore,
		Reason:       fmt.Sprintf("Risk level: %s (score: %.2f)", riskScore.Level, riskScore.Score),
	}

	// Determine if additional authentication is required.
	switch riskScore.Level {
	case RiskLevelLow:
		// Low risk - allow with basic authentication.
		decision.RequiresAuth = false

	case RiskLevelMedium:
		// Medium risk - require MFA.
		decision.RequiresAuth = true

	case RiskLevelHigh, RiskLevelCritical:
		// High/Critical risk - require strong authentication.
		decision.RequiresAuth = true
	}

	// Record authentication attempt.
	if err := r.userBehaviorStore.RecordAuthentication(ctx, userID, true, authContext); err != nil {
		// Log but don't fail.
		fmt.Printf("warning: failed to record authentication: %v\n", err)
	}

	return decision, nil
}

// InitiateAuth initiates risk-based authentication challenge (implements UserAuthenticator).
func (r *RiskBasedAuthenticator) InitiateAuth(ctx context.Context, userID string) (*AuthChallenge, error) {
	// Create a generic risk-based authentication challenge.
	challengeID, err := googleUuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate challenge ID: %w", err)
	}

	expiresAt := time.Now().UTC().Add(cryptoutilSharedMagic.DefaultOTPLifetime)

	challenge := &AuthChallenge{
		ID:        challengeID,
		UserID:    userID,
		Method:    "risk_based",
		ExpiresAt: expiresAt,
		Metadata: map[string]any{
			"risk_assessed": false,
		},
	}

	if err := r.challengeStore.Store(ctx, challenge, challengeID.String()); err != nil {
		return nil, fmt.Errorf("failed to store risk-based challenge: %w", err)
	}

	return challenge, nil
}

// VerifyAuth verifies risk-based authentication (implements UserAuthenticator).
func (r *RiskBasedAuthenticator) VerifyAuth(ctx context.Context, challengeID, _ string) (*cryptoutilIdentityDomain.User, error) {
	// Parse challenge ID.
	id, err := googleUuid.Parse(challengeID)
	if err != nil {
		return nil, fmt.Errorf("invalid challenge ID: %w", err)
	}

	// Retrieve challenge.
	challenge, _, err := r.challengeStore.Retrieve(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("challenge not found: %w", err)
	}

	// Check expiration.
	if time.Now().UTC().After(challenge.ExpiresAt) {
		// Best-effort cleanup of expired challenge.
		if err := r.challengeStore.Delete(ctx, id); err != nil {
			fmt.Printf("warning: failed to delete expired challenge: %v\n", err)
		}

		return nil, fmt.Errorf("risk-based challenge expired")
	}

	// Risk-based authentication verification is handled by the Authenticate method.
	// This method is a placeholder for interface compliance.

	// Delete challenge (single-use).
	if err := r.challengeStore.Delete(ctx, id); err != nil {
		fmt.Printf("warning: failed to delete challenge: %v\n", err)
	}

	return nil, fmt.Errorf("risk-based authentication requires context-specific verification")
}

// DefaultRiskThresholds returns default authentication requirements for each risk level.
func DefaultRiskThresholds() map[RiskLevel]*AuthRequirements {
	const (
		factorsOne   = 1
		factorsTwo   = 2
		factorsThree = 3
	)

	return map[RiskLevel]*AuthRequirements{
		RiskLevelLow: {
			MinFactors:        factorsOne,
			AllowedMethods:    []string{"password", "magic_link"},
			RequiresBiometric: false,
			RequiresHardware:  false,
			MaxAge:            cryptoutilSharedMagic.LowRiskAuthMaxAge,
		},
		RiskLevelMedium: {
			MinFactors:        factorsTwo,
			AllowedMethods:    []string{"password", cryptoutilSharedMagic.MFATypeTOTP, cryptoutilSharedMagic.AuthMethodSMSOTP},
			RequiresBiometric: false,
			RequiresHardware:  false,
			MaxAge:            cryptoutilSharedMagic.MediumRiskAuthMaxAge,
		},
		RiskLevelHigh: {
			MinFactors:        factorsTwo,
			AllowedMethods:    []string{cryptoutilSharedMagic.MFATypeTOTP, cryptoutilSharedMagic.AuthMethodHardwareKey, cryptoutilSharedMagic.AuthMethodBiometric},
			RequiresBiometric: false,
			RequiresHardware:  true,
			MaxAge:            cryptoutilSharedMagic.HighRiskAuthMaxAge,
		},
		RiskLevelCritical: {
			MinFactors:        factorsThree,
			AllowedMethods:    []string{cryptoutilSharedMagic.AuthMethodHardwareKey, cryptoutilSharedMagic.AuthMethodBiometric},
			RequiresBiometric: true,
			RequiresHardware:  true,
			MaxAge:            cryptoutilSharedMagic.CriticalRiskAuthMaxAge,
		},
	}
}
