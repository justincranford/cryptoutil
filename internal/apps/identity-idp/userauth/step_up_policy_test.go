// Copyright (c) 2025-2026 Justin Cranford.
package userauth

import (
	"context"
	"errors"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

type coveragePolicyLoader struct {
	policies *StepUpPolicies
	err      error
}

func (l *coveragePolicyLoader) LoadRiskScoringPolicy(context.Context) (*RiskScoringPolicy, error) {
	return nil, nil
}

func (l *coveragePolicyLoader) LoadStepUpPolicies(context.Context) (*StepUpPolicies, error) {
	return l.policies, l.err
}

func (l *coveragePolicyLoader) LoadAdaptiveAuthPolicy(context.Context) (*AdaptiveAuthPolicy, error) {
	return nil, nil
}

func (l *coveragePolicyLoader) EnableHotReload(context.Context, time.Duration) error {
	return nil
}

func (l *coveragePolicyLoader) DisableHotReload() {}

type coverageUserAuthenticator struct {
	user      *cryptoutilIdentityDomain.User
	verifyErr error
}

func (a *coverageUserAuthenticator) Method() string {
	return valueCoverageTestMethod
}

func (a *coverageUserAuthenticator) InitiateAuth(context.Context, string) (*AuthChallenge, error) {
	return nil, nil
}

func (a *coverageUserAuthenticator) VerifyAuth(context.Context, string, string) (*cryptoutilIdentityDomain.User, error) {
	return a.user, a.verifyErr
}

const (
	valueCoverageTestMethod           = "coverage-test-method"
	valueStepUpTestOperationTransfer  = "transfer"
	errorStepUpTestChallengeStore     = "failed to store step-up challenge"
	testStepUpChallengeStored         = "challenge is stored"
	errorStepUpTestInvalidChallengeID = "invalid challenge ID"
	errorStepUpTestChallengeNotFound  = "challenge not found"
	testStepUpExpiredChallenge        = "expired challenge"
	errorStepUpTestChallengeExpired   = "step-up challenge expired"
	keyStepUpTestChallengeMethod      = "challenge_method"
	testStepUpChallengeStorageFails   = "challenge storage fails"
)

func TestStepUpAuthenticator_EvaluateStepUpCoverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		loader         *coveragePolicyLoader
		challengeStore *coverageChallengeStore
		currentLevel   AuthenticationLevel
		authTime       time.Time
		wantErr        string
		wantChallenge  bool
	}{
		{
			name:    "policy load error",
			loader:  &coveragePolicyLoader{err: errors.New("policy load failed")},
			wantErr: "failed to load policies",
		},
		{
			name: "invalid operation max age",
			loader: &coveragePolicyLoader{policies: &StepUpPolicies{
				Policies: map[string]OperationPolicy{
					valueStepUpTestOperationTransfer: {RequiredLevel: authLevelStringBasic, MaxAge: cryptoutilSharedMagic.DevSetupTestInvalidName},
				},
			}},
			wantErr: "invalid max_age",
		},
		{
			name: "invalid default max age",
			loader: &coveragePolicyLoader{policies: &StepUpPolicies{
				DefaultPolicy: OperationPolicy{RequiredLevel: authLevelStringBasic, MaxAge: cryptoutilSharedMagic.DevSetupTestInvalidName},
			}},
			wantErr: "invalid default max_age",
		},
		{
			name:   "no applicable policy",
			loader: &coveragePolicyLoader{policies: &StepUpPolicies{}},
		},
		{
			name: "recent authentication satisfies policy",
			loader: &coveragePolicyLoader{policies: &StepUpPolicies{
				Policies: map[string]OperationPolicy{
					valueStepUpTestOperationTransfer: {RequiredLevel: authLevelStringBasic, MaxAge: time.Hour.String()},
				},
			}},
			currentLevel: AuthLevelBasic,
			authTime:     time.Now().UTC(),
		},
		{
			name: "challenge returned without configured store",
			loader: &coveragePolicyLoader{policies: &StepUpPolicies{
				Policies: map[string]OperationPolicy{
					valueStepUpTestOperationTransfer: {RequiredLevel: authLevelStringStrongMFA, AllowedMethods: []string{cryptoutilSharedMagic.AuthMethodTOTP}, MaxAge: time.Hour.String()},
				},
			}},
			wantChallenge: true,
		},
		{
			name: "challenge store error",
			loader: &coveragePolicyLoader{policies: &StepUpPolicies{
				Policies: map[string]OperationPolicy{
					valueStepUpTestOperationTransfer: {RequiredLevel: authLevelStringStrongMFA, MaxAge: time.Hour.String()},
				},
			}},
			challengeStore: &coverageChallengeStore{storeErr: errors.New("challenge store failed")},
			wantErr:        errorStepUpTestChallengeStore,
		},
		{
			name: testStepUpChallengeStored,
			loader: &coveragePolicyLoader{policies: &StepUpPolicies{
				Policies: map[string]OperationPolicy{
					valueStepUpTestOperationTransfer: {RequiredLevel: authLevelStringStrongMFA, MaxAge: time.Hour.String()},
				},
			}},
			challengeStore: &coverageChallengeStore{},
			wantChallenge:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var challengeStore ChallengeStore
			if tc.challengeStore != nil {
				challengeStore = tc.challengeStore
			}

			authenticator := NewStepUpAuthenticator(tc.loader, nil, nil, challengeStore, nil)

			challenge, err := authenticator.EvaluateStepUp(context.Background(), "user", valueStepUpTestOperationTransfer, tc.currentLevel, tc.authTime)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				require.Nil(t, challenge)

				return
			}

			require.NoError(t, err)

			if tc.wantChallenge {
				require.NotNil(t, challenge)

				if tc.challengeStore != nil {
					require.NotNil(t, tc.challengeStore.challenge)
				}

				return
			}

			require.Nil(t, challenge)
		})
	}
}

func TestStepUpAuthenticator_VerifyStepUpCoverage(t *testing.T) {
	t.Parallel()

	userID := googleUuid.Must(googleUuid.NewV7())
	challengeID := googleUuid.Must(googleUuid.NewV7())
	tests := []struct {
		name           string
		challengeID    string
		challenge      *AuthChallenge
		retrieveErr    error
		deleteErr      error
		authenticators map[string]UserAuthenticator
		wantErr        string
		wantUser       bool
		wantLevel      AuthenticationLevel
	}{
		{name: "malformed challenge ID", challengeID: "not-a-uuid", wantErr: errorStepUpTestInvalidChallengeID},
		{name: "challenge lookup fails", retrieveErr: errors.New("challenge lookup failed"), wantErr: errorStepUpTestChallengeNotFound},
		{
			name:      testStepUpExpiredChallenge,
			challenge: &AuthChallenge{ExpiresAt: time.Now().UTC().Add(-time.Hour)},
			wantErr:   errorStepUpTestChallengeExpired,
		},
		{
			name:      "expired challenge cleanup fails",
			challenge: &AuthChallenge{ExpiresAt: time.Now().UTC().Add(-time.Hour)},
			deleteErr: errors.New("challenge cleanup failed"),
			wantErr:   errorStepUpTestChallengeExpired,
		},
		{
			name: "invalid method metadata",
			challenge: &AuthChallenge{
				ExpiresAt: time.Now().UTC().Add(time.Hour),
				Metadata:  map[string]any{keyStepUpTestChallengeMethod: 1},
			},
			wantErr: "invalid challenge metadata",
		},
		{
			name: "authenticator missing",
			challenge: &AuthChallenge{
				ExpiresAt: time.Now().UTC().Add(time.Hour),
				Metadata:  map[string]any{keyStepUpTestChallengeMethod: valueCoverageTestMethod},
			},
			wantErr: "authenticator not found",
		},
		{
			name: "authenticator verification fails",
			challenge: &AuthChallenge{
				ExpiresAt: time.Now().UTC().Add(time.Hour),
				Metadata:  map[string]any{keyStepUpTestChallengeMethod: valueCoverageTestMethod},
			},
			authenticators: map[string]UserAuthenticator{
				valueCoverageTestMethod: &coverageUserAuthenticator{verifyErr: errors.New("verification failed")},
			},
			wantErr: "step-up verification failed",
		},
		{
			name: "successful verification defaults level",
			challenge: &AuthChallenge{
				ExpiresAt: time.Now().UTC().Add(time.Hour),
				Metadata:  map[string]any{keyStepUpTestChallengeMethod: valueCoverageTestMethod},
			},
			authenticators: map[string]UserAuthenticator{
				valueCoverageTestMethod: &coverageUserAuthenticator{user: &cryptoutilIdentityDomain.User{ID: userID}},
			},
			wantUser:  true,
			wantLevel: AuthLevelStepUp,
		},
		{
			name: "successful verification keeps configured level",
			challenge: &AuthChallenge{
				ExpiresAt: time.Now().UTC().Add(time.Hour),
				Metadata: map[string]any{
					keyStepUpTestChallengeMethod: valueCoverageTestMethod,
					"required_level":             AuthLevelStrongMFA,
				},
			},
			authenticators: map[string]UserAuthenticator{
				valueCoverageTestMethod: &coverageUserAuthenticator{user: &cryptoutilIdentityDomain.User{ID: userID}},
			},
			wantUser:  true,
			wantLevel: AuthLevelStrongMFA,
		},
		{
			name: "successful verification tolerates cleanup failure",
			challenge: &AuthChallenge{
				ExpiresAt: time.Now().UTC().Add(time.Hour),
				Metadata:  map[string]any{keyStepUpTestChallengeMethod: valueCoverageTestMethod},
			},
			deleteErr: errors.New("challenge cleanup failed"),
			authenticators: map[string]UserAuthenticator{
				valueCoverageTestMethod: &coverageUserAuthenticator{user: &cryptoutilIdentityDomain.User{ID: userID}},
			},
			wantUser:  true,
			wantLevel: AuthLevelStepUp,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			id := tc.challengeID
			if id == "" {
				id = challengeID.String()
			}

			challengeStore := &coverageChallengeStore{challenge: tc.challenge, retrieveErr: tc.retrieveErr, deleteErr: tc.deleteErr}
			if challengeStore.challenge != nil {
				challengeStore.challenge.ID = challengeID
			}

			if tc.challenge == nil && tc.retrieveErr == nil && tc.wantErr == "" {
				require.FailNow(t, "coverage case requires a stored challenge")
			}

			authenticator := NewStepUpAuthenticator(nil, nil, nil, challengeStore, tc.authenticators)

			user, level, err := authenticator.VerifyStepUp(context.Background(), id, "response")
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)

			if tc.wantUser {
				require.NotNil(t, user)
				require.Equal(t, tc.wantLevel, level)
			}

			if tc.deleteErr == nil {
				require.Nil(t, challengeStore.challenge)
			} else {
				require.NotNil(t, challengeStore.challenge)
			}
		})
	}
}

func TestStepUpAuthenticator_InitiateAuthCoverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		store   *coverageChallengeStore
		wantErr string
	}{
		{name: testStepUpChallengeStored, store: &coverageChallengeStore{}},
		{name: testStepUpChallengeStorageFails, store: &coverageChallengeStore{storeErr: errors.New("challenge storage failed")}, wantErr: errorStepUpTestChallengeStore},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			authenticator := NewStepUpAuthenticator(nil, nil, nil, tc.store, nil)

			challenge, err := authenticator.InitiateAuth(context.Background(), "user")
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				require.Nil(t, challenge)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, challenge)
			require.Equal(t, "user", challenge.UserID)
			require.Equal(t, authLevelStringStepUp, challenge.Method)
			require.Equal(t, challenge.ID, tc.store.challenge.ID)
		})
	}
}
