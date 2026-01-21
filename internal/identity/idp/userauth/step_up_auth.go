// Copyright (c) 2025 Justin Cranford
//
//

package userauth

import (
	"context"
	"fmt"
	"time"

	googleUuid "github.com/google/uuid"

	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// AuthenticationLevel represents the strength/level of authentication.
type AuthenticationLevel int

// Authentication level constants.
const (
	// AuthLevelNone indicates no authentication.
	AuthLevelNone AuthenticationLevel = 0
	// AuthLevelBasic indicates password-only authentication.
	AuthLevelBasic AuthenticationLevel = 1
	// AuthLevelMFA indicates multi-factor authentication.
	AuthLevelMFA AuthenticationLevel = 2
	// AuthLevelStepUp indicates step-up authentication completed.
	AuthLevelStepUp AuthenticationLevel = 3
	// AuthLevelStrongMFA indicates strong MFA (hardware key, biometric).
	AuthLevelStrongMFA AuthenticationLevel = 4
)

const (
	authLevelStringNone      = "none"
	authLevelStringBasic     = "basic"
	authLevelStringMFA       = "mfa"
	authLevelStringStepUp    = "step_up"
	authLevelStringStrongMFA = "strong_mfa"
)

// StepUpChallenge represents a step-up authentication challenge.
type StepUpChallenge struct {
	ChallengeID     string
	UserID          string
	Operation       string
	RequiredLevel   AuthenticationLevel
	CurrentLevel    AuthenticationLevel
	ExpiresAt       time.Time
	ChallengeMethod string
	Metadata        map[string]any
}

// StepUpPolicy defines when step-up authentication is required.
type StepUpPolicy struct {
	OperationPattern string              // Pattern matching operation name.
	RequiredLevel    AuthenticationLevel // Required authentication level.
	AllowedMethods   []string            // Allowed step-up methods.
	MaxAge           time.Duration       // Maximum age of existing authentication.
}

// StepUpAuthenticator implements step-up authentication with policy-driven configuration.
type StepUpAuthenticator struct {
	policyLoader    PolicyLoader
	riskEngine      RiskEngine
	contextAnalyzer ContextAnalyzer
	challengeStore  ChallengeStore
	authenticators  map[string]UserAuthenticator

	// Policies loaded from YAML (cached after first load).
	policies map[string]*StepUpPolicy
}

// NewStepUpAuthenticator creates a new step-up authenticator with policy-driven configuration.
func NewStepUpAuthenticator(
	policyLoader PolicyLoader,
	riskEngine RiskEngine,
	contextAnalyzer ContextAnalyzer,
	challengeStore ChallengeStore,
	authenticators map[string]UserAuthenticator,
) *StepUpAuthenticator {
	return &StepUpAuthenticator{
		policyLoader:    policyLoader,
		riskEngine:      riskEngine,
		contextAnalyzer: contextAnalyzer,
		challengeStore:  challengeStore,
		authenticators:  authenticators,
	}
}

// loadPolicies loads step-up policies from YAML if not already loaded.
func (s *StepUpAuthenticator) loadPolicies(ctx context.Context) error {
	// Check if policies already loaded.
	if s.policies != nil {
		return nil
	}

	// Load step-up policies.
	policyConfig, err := s.policyLoader.LoadStepUpPolicies(ctx)
	if err != nil {
		return fmt.Errorf("failed to load step-up policies: %w", err)
	}

	// Convert YAML policies to internal policy structure.
	s.policies = make(map[string]*StepUpPolicy)

	for op, policyYAML := range policyConfig.Policies {
		maxAge, err := time.ParseDuration(policyYAML.MaxAge)
		if err != nil {
			return fmt.Errorf("invalid max_age for operation %s: %w", op, err)
		}

		requiredLevel := s.parseAuthLevel(policyYAML.RequiredLevel)

		s.policies[op] = &StepUpPolicy{
			OperationPattern: policyYAML.OperationPattern,
			RequiredLevel:    requiredLevel,
			AllowedMethods:   policyYAML.AllowedMethods,
			MaxAge:           maxAge,
		}
	}

	// Add default policy.
	if policyConfig.DefaultPolicy.RequiredLevel != "" {
		defaultMaxAge, err := time.ParseDuration(policyConfig.DefaultPolicy.MaxAge)
		if err != nil {
			return fmt.Errorf("invalid default max_age: %w", err)
		}

		defaultLevel := s.parseAuthLevel(policyConfig.DefaultPolicy.RequiredLevel)

		s.policies["default"] = &StepUpPolicy{
			OperationPattern: "*",
			RequiredLevel:    defaultLevel,
			AllowedMethods:   policyConfig.DefaultPolicy.AllowedMethods,
			MaxAge:           defaultMaxAge,
		}
	}

	return nil
}

// parseAuthLevel converts string auth level to AuthenticationLevel enum.
func (s *StepUpAuthenticator) parseAuthLevel(level string) AuthenticationLevel {
	switch level {
	case authLevelStringNone:
		return AuthLevelNone
	case authLevelStringBasic:
		return AuthLevelBasic
	case authLevelStringMFA:
		return AuthLevelMFA
	case authLevelStringStepUp:
		return AuthLevelStepUp
	case authLevelStringStrongMFA:
		return AuthLevelStrongMFA
	default:
		return AuthLevelBasic // Default to basic if unknown.
	}
}

// Method returns the authentication method name.
func (s *StepUpAuthenticator) Method() string {
	return "step_up"
}

// EvaluateStepUp determines if step-up authentication is required for an operation using policy-driven rules.
func (s *StepUpAuthenticator) EvaluateStepUp(
	ctx context.Context,
	userID string,
	operation string,
	currentLevel AuthenticationLevel,
	authTime time.Time,
) (*StepUpChallenge, error) {
	// Load policies from YAML if not already loaded.
	if err := s.loadPolicies(ctx); err != nil {
		return nil, fmt.Errorf("failed to load policies: %w", err)
	}

	// Find applicable policy.
	policy, found := s.policies[operation]
	if !found {
		// Try default policy.
		policy, found = s.policies["default"]
		if !found {
			// No policy for this operation, no step-up required.
			return nil, nil
		}
	}

	// Check if current authentication level is sufficient.
	if currentLevel >= policy.RequiredLevel {
		// Check if authentication is recent enough.
		if time.Since(authTime) <= policy.MaxAge {
			// No step-up required.
			return nil, nil
		}
	}

	// Step-up required - create challenge.
	challengeID, err := googleUuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate challenge ID: %w", err)
	}

	expiresAt := time.Now().Add(cryptoutilIdentityMagic.DefaultOTPLifetime)

	// Select step-up method (first allowed method for simplicity).
	challengeMethod := cryptoutilIdentityMagic.AuthMethodTOTP
	if len(policy.AllowedMethods) > 0 {
		challengeMethod = policy.AllowedMethods[0]
	}

	challenge := &StepUpChallenge{
		ChallengeID:     challengeID.String(),
		UserID:          userID,
		Operation:       operation,
		RequiredLevel:   policy.RequiredLevel,
		CurrentLevel:    currentLevel,
		ExpiresAt:       expiresAt,
		ChallengeMethod: challengeMethod,
		Metadata: map[string]any{
			"policy": operation,
		},
	}

	// Nil check for challenge store before storing.
	if s.challengeStore == nil {
		return challenge, nil // Return challenge without storing if no store configured.
	}

	// Store challenge.
	authChallenge := &AuthChallenge{
		ID:        challengeID,
		UserID:    userID,
		Method:    "step_up",
		ExpiresAt: expiresAt,
		Metadata: map[string]any{
			"operation":        operation,
			"required_level":   policy.RequiredLevel,
			"current_level":    currentLevel,
			"challenge_method": challengeMethod,
		},
	}

	if err := s.challengeStore.Store(ctx, authChallenge, challenge.ChallengeID); err != nil {
		return nil, fmt.Errorf("failed to store step-up challenge: %w", err)
	}

	return challenge, nil
}

// VerifyStepUp verifies a step-up authentication attempt.
func (s *StepUpAuthenticator) VerifyStepUp(
	ctx context.Context,
	challengeID string,
	response string,
) (*cryptoutilIdentityDomain.User, AuthenticationLevel, error) {
	// Parse challenge ID.
	id, err := googleUuid.Parse(challengeID)
	if err != nil {
		return nil, AuthLevelNone, fmt.Errorf("invalid challenge ID: %w", err)
	}

	// Retrieve challenge.
	authChallenge, challengeData, err := s.challengeStore.Retrieve(ctx, id)
	if err != nil {
		return nil, AuthLevelNone, fmt.Errorf("challenge not found: %w", err)
	}

	// Check expiration.
	if time.Now().After(authChallenge.ExpiresAt) {
		// Best-effort cleanup of expired challenge.
		if err := s.challengeStore.Delete(ctx, id); err != nil {
			fmt.Printf("warning: failed to delete expired challenge: %v\n", err)
		}

		return nil, AuthLevelNone, fmt.Errorf("step-up challenge expired")
	}

	// Get challenge method.
	challengeMethod, ok := authChallenge.Metadata["challenge_method"].(string)
	if !ok {
		return nil, AuthLevelNone, fmt.Errorf("invalid challenge metadata")
	}

	// Get authenticator for the challenge method.
	authenticator, found := s.authenticators[challengeMethod]
	if !found {
		return nil, AuthLevelNone, fmt.Errorf("authenticator not found for method: %s", challengeMethod)
	}

	// Verify using the appropriate authenticator.
	user, err := authenticator.VerifyAuth(ctx, challengeData, response)
	if err != nil {
		return nil, AuthLevelNone, fmt.Errorf("step-up verification failed: %w", err)
	}

	// Get required level from metadata.
	requiredLevel, ok := authChallenge.Metadata["required_level"].(AuthenticationLevel)
	if !ok {
		requiredLevel = AuthLevelStepUp
	}

	// Delete challenge (single-use).
	if err := s.challengeStore.Delete(ctx, id); err != nil {
		fmt.Printf("warning: failed to delete challenge: %v\n", err)
	}

	return user, requiredLevel, nil
}

// InitiateAuth initiates step-up authentication (implements UserAuthenticator).
func (s *StepUpAuthenticator) InitiateAuth(ctx context.Context, userID string) (*AuthChallenge, error) {
	// Step-up authentication is operation-specific, so this method creates a generic challenge.
	challengeID, err := googleUuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate challenge ID: %w", err)
	}

	expiresAt := time.Now().Add(cryptoutilIdentityMagic.DefaultOTPLifetime)

	challenge := &AuthChallenge{
		ID:        challengeID,
		UserID:    userID,
		Method:    "step_up",
		ExpiresAt: expiresAt,
		Metadata: map[string]any{
			"operation": "generic",
		},
	}

	if err := s.challengeStore.Store(ctx, challenge, challengeID.String()); err != nil {
		return nil, fmt.Errorf("failed to store step-up challenge: %w", err)
	}

	return challenge, nil
}

// VerifyAuth verifies step-up authentication (implements UserAuthenticator).
func (s *StepUpAuthenticator) VerifyAuth(ctx context.Context, challengeID, response string) (*cryptoutilIdentityDomain.User, error) {
	user, _, err := s.VerifyStepUp(ctx, challengeID, response)

	return user, err
}

// DefaultStepUpPolicies returns default step-up policies for common operations.
func DefaultStepUpPolicies() map[string]*StepUpPolicy {
	const (
		policyTransferFunds  = "transfer_funds"
		policyChangePassword = "change_password"
		policyAddPayee       = "add_payee"
		policyDeleteAccount  = "delete_account"
		policyViewPII        = "view_pii"
	)

	return map[string]*StepUpPolicy{
		policyTransferFunds: {
			OperationPattern: policyTransferFunds,
			RequiredLevel:    AuthLevelMFA,
			AllowedMethods:   []string{cryptoutilIdentityMagic.AuthMethodTOTP, cryptoutilIdentityMagic.AuthMethodSMSOTP, cryptoutilIdentityMagic.AuthMethodHardwareKey},
			MaxAge:           cryptoutilIdentityMagic.StepUpTransferFundsMaxAge,
		},
		policyChangePassword: {
			OperationPattern: policyChangePassword,
			RequiredLevel:    AuthLevelStepUp,
			AllowedMethods:   []string{cryptoutilIdentityMagic.AuthMethodTOTP, cryptoutilIdentityMagic.AuthMethodSMSOTP},
			MaxAge:           cryptoutilIdentityMagic.StepUpChangePasswordMaxAge,
		},
		policyAddPayee: {
			OperationPattern: policyAddPayee,
			RequiredLevel:    AuthLevelMFA,
			AllowedMethods:   []string{cryptoutilIdentityMagic.AuthMethodTOTP, cryptoutilIdentityMagic.AuthMethodSMSOTP, cryptoutilIdentityMagic.AuthMethodHardwareKey},
			MaxAge:           cryptoutilIdentityMagic.StepUpAddPayeeMaxAge,
		},
		policyDeleteAccount: {
			OperationPattern: policyDeleteAccount,
			RequiredLevel:    AuthLevelStrongMFA,
			AllowedMethods:   []string{cryptoutilIdentityMagic.AuthMethodHardwareKey, cryptoutilIdentityMagic.AuthMethodBiometric},
			MaxAge:           cryptoutilIdentityMagic.StepUpDeleteAccountMaxAge,
		},
		policyViewPII: {
			OperationPattern: policyViewPII,
			RequiredLevel:    AuthLevelMFA,
			AllowedMethods:   []string{cryptoutilIdentityMagic.AuthMethodTOTP, cryptoutilIdentityMagic.AuthMethodSMSOTP},
			MaxAge:           cryptoutilIdentityMagic.StepUpViewPIIMaxAge,
		},
	}
}
