// Copyright (c) 2025-2026 Justin Cranford.
package userauth

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const riskEngineTestBaselineFailure = "injected baseline lookup failure"

func TestBehavioralRiskEngine_AssessRiskErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		writePolicy   bool
		baselineError error
		wantError     string
	}{
		{
			name:      "policy loading fails",
			wantError: "failed to load policy weights",
		},
		{
			name:          "baseline lookup fails",
			writePolicy:   true,
			baselineError: errors.New(riskEngineTestBaselineFailure),
			wantError:     "failed to get user baseline",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			policyPath := filepath.Join(t.TempDir(), "risk_scoring.yml")
			if tc.writePolicy {
				err := os.WriteFile(policyPath, []byte(testRiskScoringYAML), cryptoutilSharedMagic.CacheFilePermissions)
				require.NoError(t, err)
			}

			engine := NewBehavioralRiskEngine(
				NewYAMLPolicyLoader(policyPath, "", ""),
				&mockUserBehaviorStore{baselineError: tc.baselineError},
				&mockGeoIPService{},
				&mockDeviceFingerprintDB{},
			)

			score, err := engine.AssessRisk(
				context.Background(),
				testUser123,
				&AuthContext{Time: time.Now().UTC()},
			)

			require.Nil(t, score)
			require.ErrorContains(t, err, tc.wantError)
		})
	}
}

func TestBehavioralRiskEngine_OptionalBehaviorBranches(t *testing.T) {
	t.Parallel()

	engine := &BehavioralRiskEngine{}
	authTime := time.Date(2025, 1, 15, 3, 0, 0, 0, time.UTC)
	baseline := &UserBaseline{}

	factors := engine.CalculateRiskFactors(
		&AuthContext{Behavior: &UserBehavior{}, Time: authTime},
		baseline,
	)

	require.Len(t, factors, 3)
	require.Equal(t, RiskFactorTime, factors[0].Type)
	require.Equal(t, RiskFactorBehavior, factors[1].Type)
	require.Equal(t, RiskFactorVelocity, factors[2].Type)
	require.Equal(t, cryptoutilSharedMagic.RiskScoreMedium, engine.assessTimeRisk(authTime, baseline))

	confidence := engine.calculateConfidence(
		&UserBaseline{BehaviorProfile: &BehaviorProfile{}},
		0,
	)
	require.Equal(t, cryptoutilSharedMagic.ConfidenceWeightBehavior, confidence)
}
