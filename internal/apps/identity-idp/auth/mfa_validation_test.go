// Copyright (c) 2025-2026 Justin Cranford.
package auth

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	mfaValidationTestSecret                    = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	mfaValidationInvalidBase32Secret           = "invalid-base32-secret!"
	mfaValidationTOTPInvalidCodeCaseName       = "totp invalid code"
	mfaValidationMissingOTPCodeCaseName        = "missing otp code"
	mfaValidationMissingOTPCodeError           = "missing otp_code"
	mfaValidationOTPCredentialsField           = "otp_code"
	mfaValidationUnknownFactorTypeCaseAndError = "unknown factor type"
)

type mfaValidationTestSecretStore struct {
	totpSecret  string
	emailSecret string
	smsSecret   string
	err         error
}

func (s mfaValidationTestSecretStore) GetTOTPSecret(context.Context, string) (string, error) {
	return s.totpSecret, s.err
}

func (s mfaValidationTestSecretStore) GetEmailOTPSecret(context.Context, string) (string, error) {
	return s.emailSecret, s.err
}

func (s mfaValidationTestSecretStore) GetSMSOTPSecret(context.Context, string) (string, error) {
	return s.smsSecret, s.err
}

func TestTOTPValidator_ValidateMethods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		secret     string
		period     uint
		algorithm  otp.Algorithm
		secretKind string
		validate   func(*TOTPValidator, string) (bool, error)
		wantValid  bool
		wantErr    string
		storeError bool
	}{
		{
			name:       "totp valid",
			secret:     mfaValidationTestSecret,
			period:     totpPeriod30Seconds,
			algorithm:  otp.AlgorithmSHA1,
			secretKind: cryptoutilSharedMagic.MFATypeTOTP,
			validate: func(v *TOTPValidator, code string) (bool, error) {
				return v.ValidateTOTP(context.Background(), "user", code)
			},
			wantValid: true,
		},
		{
			name:       mfaValidationTOTPInvalidCodeCaseName,
			secret:     mfaValidationTestSecret,
			period:     totpPeriod30Seconds,
			algorithm:  otp.AlgorithmSHA1,
			secretKind: cryptoutilSharedMagic.MFATypeTOTP,
			validate: func(v *TOTPValidator, code string) (bool, error) {
				return v.ValidateTOTP(context.Background(), "user", code)
			},
		},
		{
			name:       "totp custom window valid",
			secret:     mfaValidationTestSecret,
			period:     totpPeriod30Seconds,
			algorithm:  otp.AlgorithmSHA1,
			secretKind: cryptoutilSharedMagic.MFATypeTOTP,
			validate: func(v *TOTPValidator, code string) (bool, error) {
				return v.ValidateTOTPWithWindow(context.Background(), "user", code, 1)
			},
			wantValid: true,
		},
		{
			name:       "totp custom window invalid secret",
			secret:     mfaValidationInvalidBase32Secret,
			period:     totpPeriod30Seconds,
			algorithm:  otp.AlgorithmSHA1,
			secretKind: cryptoutilSharedMagic.MFATypeTOTP,
			validate: func(v *TOTPValidator, code string) (bool, error) {
				return v.ValidateTOTPWithWindow(context.Background(), "user", code, 1)
			},
			wantErr: "TOTP validation failed",
		},
		{
			name:       "email otp valid",
			secret:     mfaValidationTestSecret,
			period:     totpPeriod300Seconds,
			algorithm:  otp.AlgorithmSHA256,
			secretKind: cryptoutilSharedMagic.ScopeEmail,
			validate: func(v *TOTPValidator, code string) (bool, error) {
				return v.ValidateEmailOTP(context.Background(), "user", code)
			},
			wantValid: true,
		},
		{
			name:       "email otp invalid secret",
			secret:     mfaValidationInvalidBase32Secret,
			period:     totpPeriod300Seconds,
			algorithm:  otp.AlgorithmSHA256,
			secretKind: cryptoutilSharedMagic.ScopeEmail,
			validate: func(v *TOTPValidator, code string) (bool, error) {
				return v.ValidateEmailOTP(context.Background(), "user", code)
			},
			wantErr: "email OTP validation failed",
		},
		{
			name:       "sms otp valid",
			secret:     mfaValidationTestSecret,
			period:     totpPeriod600Seconds,
			algorithm:  otp.AlgorithmSHA256,
			secretKind: cryptoutilSharedMagic.MFATypeSMS,
			validate: func(v *TOTPValidator, code string) (bool, error) {
				return v.ValidateSMSOTP(context.Background(), "user", code)
			},
			wantValid: true,
		},
		{
			name:       "sms otp invalid secret",
			secret:     mfaValidationInvalidBase32Secret,
			period:     totpPeriod600Seconds,
			algorithm:  otp.AlgorithmSHA256,
			secretKind: cryptoutilSharedMagic.MFATypeSMS,
			validate: func(v *TOTPValidator, code string) (bool, error) {
				return v.ValidateSMSOTP(context.Background(), "user", code)
			},
			wantErr: "SMS OTP validation failed",
		},
		{
			name:       "secret store error",
			storeError: true,
			secretKind: cryptoutilSharedMagic.MFATypeTOTP,
			validate: func(v *TOTPValidator, code string) (bool, error) {
				return v.ValidateTOTP(context.Background(), "user", code)
			},
			wantErr: "failed to retrieve TOTP secret",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := mfaValidationTestSecretStore{
				totpSecret:  tc.secret,
				emailSecret: tc.secret,
				smsSecret:   tc.secret,
			}
			if tc.storeError {
				store.err = errors.New("injected secret store failure")
			}

			code := "123456"

			if tc.secret == mfaValidationTestSecret {
				var err error

				code, err = totp.GenerateCodeCustom(tc.secret, time.Now().UTC(), totp.ValidateOpts{
					Period:    tc.period,
					Skew:      1,
					Digits:    otp.DigitsSix,
					Algorithm: tc.algorithm,
				})
				require.NoError(t, err)

				if tc.name == mfaValidationTOTPInvalidCodeCaseName {
					codeValue, parseErr := strconv.Atoi(code)
					require.NoError(t, parseErr)

					code = fmt.Sprintf("%06d", (codeValue+1)%1000000)
				}
			}

			switch tc.secretKind {
			case cryptoutilSharedMagic.MFATypeTOTP:
				store.totpSecret = tc.secret
			case cryptoutilSharedMagic.ScopeEmail:
				store.emailSecret = tc.secret
			case cryptoutilSharedMagic.MFATypeSMS:
				store.smsSecret = tc.secret
			}

			valid, err := tc.validate(NewTOTPValidator(store), code)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantValid, valid)
		})
	}
}

func TestMFAOrchestrator_IntegrateTOTPValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		factorType  cryptoutilIdentityDomain.MFAFactorType
		period      uint
		algorithm   otp.Algorithm
		secret      string
		credentials map[string]string
		invalidCode bool
		wantErr     string
		wantValid   bool
	}{
		{
			name:        mfaValidationMissingOTPCodeCaseName,
			factorType:  cryptoutilIdentityDomain.MFAFactorTypeTOTP,
			credentials: map[string]string{},
			wantErr:     mfaValidationMissingOTPCodeError,
		},
		{
			name:        "totp validator missing",
			factorType:  cryptoutilIdentityDomain.MFAFactorTypeTOTP,
			credentials: map[string]string{mfaValidationOTPCredentialsField: "123456"},
			wantErr:     "TOTP validator not configured",
		},
		{
			name:        "password factor is unsupported",
			factorType:  cryptoutilIdentityDomain.MFAFactorTypePassword,
			secret:      mfaValidationTestSecret,
			credentials: map[string]string{mfaValidationOTPCredentialsField: "123456"},
			wantErr:     "unsupported OTP factor type",
		},
		{
			name:        mfaValidationUnknownFactorTypeCaseAndError,
			factorType:  cryptoutilIdentityDomain.MFAFactorType(cryptoutilSharedMagic.STRATEGY_UNKNOWN),
			secret:      mfaValidationTestSecret,
			credentials: map[string]string{mfaValidationOTPCredentialsField: "123456"},
			wantErr:     mfaValidationUnknownFactorTypeCaseAndError,
		},
		{
			name:        "totp valid code",
			factorType:  cryptoutilIdentityDomain.MFAFactorTypeTOTP,
			period:      totpPeriod30Seconds,
			algorithm:   otp.AlgorithmSHA1,
			secret:      mfaValidationTestSecret,
			credentials: map[string]string{},
			wantValid:   true,
		},
		{
			name:        "email otp valid code",
			factorType:  cryptoutilIdentityDomain.MFAFactorTypeEmailOTP,
			period:      totpPeriod300Seconds,
			algorithm:   otp.AlgorithmSHA256,
			secret:      mfaValidationTestSecret,
			credentials: map[string]string{},
			wantValid:   true,
		},
		{
			name:        "sms otp valid code",
			factorType:  cryptoutilIdentityDomain.MFAFactorTypeSMSOTP,
			period:      totpPeriod600Seconds,
			algorithm:   otp.AlgorithmSHA256,
			secret:      mfaValidationTestSecret,
			credentials: map[string]string{},
			wantValid:   true,
		},
		{
			name:        mfaValidationTOTPInvalidCodeCaseName,
			factorType:  cryptoutilIdentityDomain.MFAFactorTypeTOTP,
			period:      totpPeriod30Seconds,
			algorithm:   otp.AlgorithmSHA1,
			secret:      mfaValidationTestSecret,
			credentials: map[string]string{},
			invalidCode: true,
		},
		{
			name:        "email otp invalid code",
			factorType:  cryptoutilIdentityDomain.MFAFactorTypeEmailOTP,
			period:      totpPeriod300Seconds,
			algorithm:   otp.AlgorithmSHA256,
			secret:      mfaValidationTestSecret,
			credentials: map[string]string{},
			invalidCode: true,
		},
		{
			name:        "sms otp invalid code",
			factorType:  cryptoutilIdentityDomain.MFAFactorTypeSMSOTP,
			period:      totpPeriod600Seconds,
			algorithm:   otp.AlgorithmSHA256,
			secret:      mfaValidationTestSecret,
			credentials: map[string]string{},
			invalidCode: true,
		},
		{
			name:        "totp validation error",
			factorType:  cryptoutilIdentityDomain.MFAFactorTypeTOTP,
			secret:      mfaValidationInvalidBase32Secret,
			credentials: map[string]string{mfaValidationOTPCredentialsField: "123456"},
			wantErr:     "TOTP validation error",
		},
		{
			name:        "email otp validation error",
			factorType:  cryptoutilIdentityDomain.MFAFactorTypeEmailOTP,
			secret:      mfaValidationInvalidBase32Secret,
			credentials: map[string]string{mfaValidationOTPCredentialsField: "123456"},
			wantErr:     "email OTP validation error",
		},
		{
			name:        "sms otp validation error",
			factorType:  cryptoutilIdentityDomain.MFAFactorTypeSMSOTP,
			secret:      mfaValidationInvalidBase32Secret,
			credentials: map[string]string{mfaValidationOTPCredentialsField: "123456"},
			wantErr:     "SMS OTP validation error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			orchestrator := NewMFAOrchestrator(nil, nil, nil, newTestTelemetry(t), nil)
			if tc.secret != "" {
				orchestrator.totpValidator = NewTOTPValidator(mfaValidationTestSecretStore{
					totpSecret:  tc.secret,
					emailSecret: tc.secret,
					smsSecret:   tc.secret,
				})
			}

			credentials := tc.credentials
			if tc.wantValid || tc.invalidCode {
				code, err := totp.GenerateCodeCustom(tc.secret, time.Now().UTC(), totp.ValidateOpts{
					Period:    tc.period,
					Skew:      1,
					Digits:    otp.DigitsSix,
					Algorithm: tc.algorithm,
				})
				require.NoError(t, err)

				if tc.invalidCode {
					codeValue, parseErr := strconv.Atoi(code)
					require.NoError(t, parseErr)

					code = fmt.Sprintf("%06d", (codeValue+1)%1000000)
				}

				credentials = map[string]string{mfaValidationOTPCredentialsField: code}
			}

			factor := &cryptoutilIdentityDomain.MFAFactor{
				FactorType:    tc.factorType,
				AuthProfileID: googleUuid.Must(googleUuid.NewV7()),
			}

			valid, err := orchestrator.IntegrateTOTPValidation(context.Background(), factor, credentials)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantValid, valid)
		})
	}
}

func TestMFAOrchestrator_ValidateFactorWithOTP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		factorType  cryptoutilIdentityDomain.MFAFactorType
		period      uint
		algorithm   otp.Algorithm
		secret      string
		missingCode bool
		updateError bool
		invalidCode bool
		wantErr     string
	}{
		{
			name:       "totp success",
			factorType: cryptoutilIdentityDomain.MFAFactorTypeTOTP,
			period:     totpPeriod30Seconds,
			algorithm:  otp.AlgorithmSHA1,
			secret:     mfaValidationTestSecret,
		},
		{
			name:       "email otp success",
			factorType: cryptoutilIdentityDomain.MFAFactorTypeEmailOTP,
			period:     totpPeriod300Seconds,
			algorithm:  otp.AlgorithmSHA256,
			secret:     mfaValidationTestSecret,
		},
		{
			name:       "sms otp success",
			factorType: cryptoutilIdentityDomain.MFAFactorTypeSMSOTP,
			period:     totpPeriod600Seconds,
			algorithm:  otp.AlgorithmSHA256,
			secret:     mfaValidationTestSecret,
		},
		{
			name:        mfaValidationMissingOTPCodeCaseName,
			factorType:  cryptoutilIdentityDomain.MFAFactorTypeTOTP,
			secret:      mfaValidationTestSecret,
			missingCode: true,
			wantErr:     mfaValidationMissingOTPCodeError,
		},
		{
			name:        "repository update error",
			factorType:  cryptoutilIdentityDomain.MFAFactorTypeTOTP,
			period:      totpPeriod30Seconds,
			algorithm:   otp.AlgorithmSHA1,
			secret:      mfaValidationTestSecret,
			updateError: true,
			wantErr:     "failed to mark nonce as used",
		},
		{
			name:        "invalid totp code",
			factorType:  cryptoutilIdentityDomain.MFAFactorTypeTOTP,
			period:      totpPeriod30Seconds,
			algorithm:   otp.AlgorithmSHA1,
			secret:      mfaValidationTestSecret,
			invalidCode: true,
			wantErr:     "invalid TOTP code",
		},
		{
			name:        "invalid email otp code",
			factorType:  cryptoutilIdentityDomain.MFAFactorTypeEmailOTP,
			period:      totpPeriod300Seconds,
			algorithm:   otp.AlgorithmSHA256,
			secret:      mfaValidationTestSecret,
			invalidCode: true,
			wantErr:     "invalid email OTP code",
		},
		{
			name:        "invalid sms otp code",
			factorType:  cryptoutilIdentityDomain.MFAFactorTypeSMSOTP,
			period:      totpPeriod600Seconds,
			algorithm:   otp.AlgorithmSHA256,
			secret:      mfaValidationTestSecret,
			invalidCode: true,
			wantErr:     "invalid SMS OTP code",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			authProfileID := googleUuid.Must(googleUuid.NewV7())
			nonceExpiry := time.Now().UTC().Add(time.Minute)
			factor := &cryptoutilIdentityDomain.MFAFactor{
				ID:             googleUuid.Must(googleUuid.NewV7()),
				AuthProfileID:  authProfileID,
				FactorType:     tc.factorType,
				Nonce:          googleUuid.Must(googleUuid.NewV7()).String(),
				NonceExpiresAt: &nonceExpiry,
			}

			repo := &mockMFAFactorRepo{factors: []*cryptoutilIdentityDomain.MFAFactor{factor}}
			if tc.updateError {
				repo.updateFunc = func(context.Context, *cryptoutilIdentityDomain.MFAFactor) error {
					return errors.New("injected update failure")
				}
			}

			store := mfaValidationTestSecretStore{
				totpSecret:  tc.secret,
				emailSecret: tc.secret,
				smsSecret:   tc.secret,
			}
			orchestrator := NewMFAOrchestrator(repo, nil, nil, newTestTelemetry(t), NewTOTPValidator(store))

			credentials := map[string]string{}

			if !tc.missingCode {
				code, err := totp.GenerateCodeCustom(tc.secret, time.Now().UTC(), totp.ValidateOpts{
					Period:    tc.period,
					Skew:      1,
					Digits:    otp.DigitsSix,
					Algorithm: tc.algorithm,
				})
				require.NoError(t, err)

				if tc.invalidCode {
					codeValue, parseErr := strconv.Atoi(code)
					require.NoError(t, parseErr)

					code = fmt.Sprintf("%06d", (codeValue+1)%1000000)
				}

				credentials[mfaValidationOTPCredentialsField] = code
			}

			err := orchestrator.ValidateFactor(context.Background(), authProfileID, string(tc.factorType), credentials)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, factor.NonceUsedAt)
		})
	}
}

func TestMFAOrchestrator_RequiresMFARepositoryError(t *testing.T) {
	t.Parallel()

	orchestrator := NewMFAOrchestrator(
		&mockMFAFactorRepo{getByAuthProfileErr: errors.New("injected lookup failure")},
		nil,
		nil,
		newTestTelemetry(t),
		nil,
	)

	requiresMFA, err := orchestrator.RequiresMFA(context.Background(), googleUuid.Must(googleUuid.NewV7()))
	require.ErrorContains(t, err, "failed to fetch MFA factors")
	require.False(t, requiresMFA)
}
