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

func TestStepUpAuthenticator_LoadPolicies(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create temporary policy file.
	tempDir := t.TempDir()
	stepUpFile := filepath.Join(tempDir, "step_up.yml")

	policyContent := `version: "1.0"
policies:
  transfer_funds:
    operation_pattern: "transfer_funds"
    required_level: "step_up"
    allowed_methods: ["sms_otp", "totp", "webauthn"]
    max_age: "5m"
    description: "High-value financial transfer"
  change_password:
    operation_pattern: "change_password"
    required_level: "step_up"
    allowed_methods: ["email_otp", "totp"]
    max_age: "2m"
    description: "Password change operation"
default_policy:
  required_level: "basic"
  allowed_methods: ["password"]
  max_age: "24h"
auth_levels:
  none: 0
  basic: 1
  mfa: 2
  step_up: 3
  strong_mfa: 4
step_up_methods: {}
session_durations: {}
monitoring:
  step_up_rate: "15%"
  blocked_operations: "5%"
  fallback_methods: "20%"
  description: ""
`

	err := os.WriteFile(stepUpFile, []byte(policyContent), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	loader := NewYAMLPolicyLoader("", stepUpFile, "")

	// Create authenticator.
	auth := NewStepUpAuthenticator(
		loader,
		nil, // riskEngine.
		nil, // contextAnalyzer.
		nil, // challengeStore.
		nil, // authenticators.
	)

	// Verify policies initially nil.
	require.Nil(t, auth.policies)

	// Load policies.
	err = auth.loadPolicies(ctx)
	require.NoError(t, err)

	// Verify policies loaded.
	require.NotNil(t, auth.policies)
	require.Contains(t, auth.policies, "transfer_funds")
	require.Contains(t, auth.policies, "change_password")
	require.Contains(t, auth.policies, "default")

	// Verify transfer_funds policy.
	transferPolicy := auth.policies["transfer_funds"]
	require.Equal(t, "transfer_funds", transferPolicy.OperationPattern)
	require.Equal(t, AuthLevelStepUp, transferPolicy.RequiredLevel)
	require.Equal(t, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Minute, transferPolicy.MaxAge)
	require.ElementsMatch(t, []string{cryptoutilSharedMagic.AuthMethodSMSOTP, cryptoutilSharedMagic.MFATypeTOTP, cryptoutilSharedMagic.MFATypeWebAuthn}, transferPolicy.AllowedMethods)

	// Verify change_password policy.
	passwordPolicy := auth.policies["change_password"]
	require.Equal(t, AuthLevelStepUp, passwordPolicy.RequiredLevel)
	require.Equal(t, 2*time.Minute, passwordPolicy.MaxAge)

	// Verify default policy.
	defaultPolicy := auth.policies["default"]
	require.Equal(t, AuthLevelBasic, defaultPolicy.RequiredLevel)
	require.Equal(t, cryptoutilSharedMagic.HoursPerDay*time.Hour, defaultPolicy.MaxAge)

	// Load policies again (should use cached).
	err = auth.loadPolicies(ctx)
	require.NoError(t, err)
}

func TestStepUpAuthenticator_EvaluateStepUp(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create temporary policy file.
	tempDir := t.TempDir()
	stepUpFile := filepath.Join(tempDir, "step_up.yml")

	policyContent := `version: "1.0"
policies:
  transfer_funds:
    operation_pattern: "transfer_funds"
    required_level: "step_up"
    allowed_methods: ["sms_otp", "totp"]
    max_age: "5m"
  view_profile:
    operation_pattern: "view_profile"
    required_level: "basic"
    allowed_methods: ["password"]
    max_age: "24h"
default_policy:
  required_level: "basic"
  allowed_methods: ["password"]
  max_age: "24h"
auth_levels: {}
step_up_methods: {}
session_durations: {}
monitoring:
  step_up_rate: ""
  blocked_operations: ""
  fallback_methods: ""
  description: ""
`

	err := os.WriteFile(stepUpFile, []byte(policyContent), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	loader := NewYAMLPolicyLoader("", stepUpFile, "")

	auth := NewStepUpAuthenticator(
		loader,
		nil, // riskEngine.
		nil, // contextAnalyzer.
		nil, // challengeStore.
		nil, // authenticators.
	)

	tests := []struct {
		name              string
		operation         string
		currentLevel      AuthenticationLevel
		authTime          time.Time
		wantStepUpNeeded  bool
		wantRequiredLevel AuthenticationLevel
	}{
		{
			name:              "transfer_funds requires step_up - current basic",
			operation:         "transfer_funds",
			currentLevel:      AuthLevelBasic,
			authTime:          time.Now().UTC(),
			wantStepUpNeeded:  true,
			wantRequiredLevel: AuthLevelStepUp,
		},
		{
			name:             "transfer_funds satisfied - recent step_up",
			operation:        "transfer_funds",
			currentLevel:     AuthLevelStepUp,
			authTime:         time.Now().UTC().Add(-2 * time.Minute), // Within 5m max age.
			wantStepUpNeeded: false,
		},
		{
			name:              "transfer_funds expired - old step_up",
			operation:         "transfer_funds",
			currentLevel:      AuthLevelStepUp,
			authTime:          time.Now().UTC().Add(-cryptoutilSharedMagic.JoseJADefaultMaxMaterials * time.Minute), // Beyond 5m max age.
			wantStepUpNeeded:  true,
			wantRequiredLevel: AuthLevelStepUp,
		},
		{
			name:             "view_profile satisfied - basic auth",
			operation:        "view_profile",
			currentLevel:     AuthLevelBasic,
			authTime:         time.Now().UTC().Add(-1 * time.Hour), // Within 24h max age.
			wantStepUpNeeded: false,
		},
		{
			name:             "unknown operation uses default policy",
			operation:        "unknown_operation",
			currentLevel:     AuthLevelBasic,
			authTime:         time.Now().UTC(),
			wantStepUpNeeded: false, // Default policy requires basic.
		},
		{
			name:              "unknown operation with no auth uses default policy",
			operation:         "unknown_operation",
			currentLevel:      AuthLevelNone,
			authTime:          time.Now().UTC(),
			wantStepUpNeeded:  true, // Default policy requires basic.
			wantRequiredLevel: AuthLevelBasic,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			challenge, err := auth.EvaluateStepUp(ctx, "user-123", tc.operation, tc.currentLevel, tc.authTime)
			require.NoError(t, err)

			if tc.wantStepUpNeeded {
				require.NotNil(t, challenge)
				require.Equal(t, tc.wantRequiredLevel, challenge.RequiredLevel)
				require.Equal(t, tc.currentLevel, challenge.CurrentLevel)
			} else {
				require.Nil(t, challenge)
			}
		})
	}
}

func TestStepUpAuthenticator_ParseAuthLevel(t *testing.T) {
	t.Parallel()

	auth := &StepUpAuthenticator{}

	tests := []struct {
		levelString string
		want        AuthenticationLevel
	}{
		{cryptoutilSharedMagic.PromptNone, AuthLevelNone},
		{"basic", AuthLevelBasic},
		{cryptoutilSharedMagic.AMRMultiFactor, AuthLevelMFA},
		{"step_up", AuthLevelStepUp},
		{"strong_mfa", AuthLevelStrongMFA},
		{"unknown", AuthLevelBasic}, // Default to basic for unknown.
	}

	for _, tc := range tests {
		t.Run(tc.levelString, func(t *testing.T) {
			t.Parallel()

			result := auth.parseAuthLevel(tc.levelString)
			require.Equal(t, tc.want, result)
		})
	}
}
