// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestMFAFactor_BeforeCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		factor *MFAFactor
		check  func(*testing.T, *MFAFactor)
	}{
		{
			name: "generates_id_when_nil",
			factor: &MFAFactor{
				Name:          "test_factor",
				FactorType:    MFAFactorTypeTOTP,
				Order:         1,
				AuthProfileID: googleUuid.Must(googleUuid.NewV7()),
			},
			check: func(t *testing.T, mf *MFAFactor) {
				t.Helper()
				require.NotEqual(t, googleUuid.Nil, mf.ID)
			},
		},
		{
			name: "preserves_existing_id",
			factor: &MFAFactor{
				ID:            googleUuid.Must(googleUuid.NewV7()),
				Name:          "test_factor",
				FactorType:    MFAFactorTypeTOTP,
				Order:         1,
				AuthProfileID: googleUuid.Must(googleUuid.NewV7()),
			},
			check: func(t *testing.T, mf *MFAFactor) {
				t.Helper()
				require.NotEqual(t, googleUuid.Nil, mf.ID)
			},
		},
		{
			name: "generates_nonce_when_empty",
			factor: &MFAFactor{
				Name:          "test_factor",
				FactorType:    MFAFactorTypeTOTP,
				Order:         1,
				AuthProfileID: googleUuid.Must(googleUuid.NewV7()),
			},
			check: func(t *testing.T, mf *MFAFactor) {
				t.Helper()
				require.NotEmpty(t, mf.Nonce)
				_, err := googleUuid.Parse(mf.Nonce)
				require.NoError(t, err)
			},
		},
		{
			name: "preserves_existing_nonce",
			factor: &MFAFactor{
				Name:          "test_factor",
				FactorType:    MFAFactorTypeTOTP,
				Order:         1,
				AuthProfileID: googleUuid.Must(googleUuid.NewV7()),
				Nonce:         "existing_nonce_value",
			},
			check: func(t *testing.T, mf *MFAFactor) {
				t.Helper()
				require.Equal(t, "existing_nonce_value", mf.Nonce)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.factor.BeforeCreate(&gorm.DB{})
			require.NoError(t, err)
			tc.check(t, tc.factor)
		})
	}
}

func TestMFAFactor_TableName(t *testing.T) {
	t.Parallel()

	factor := &MFAFactor{}
	require.Equal(t, "mfa_factors", factor.TableName())
}

func TestMFAFactor_IsNonceValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		factor    *MFAFactor
		wantValid bool
	}{
		{
			name: "valid_nonce_no_expiry",
			factor: &MFAFactor{
				Nonce: "valid_nonce",
			},
			wantValid: true,
		},
		{
			name: "valid_nonce_not_expired",
			factor: &MFAFactor{
				Nonce: "valid_nonce",
				NonceExpiresAt: func() *time.Time {
					t := time.Now().UTC().Add(1 * time.Hour)

					return &t
				}(),
			},
			wantValid: true,
		},
		{
			name: "expired_nonce",
			factor: &MFAFactor{
				Nonce: "expired_nonce",
				NonceExpiresAt: func() *time.Time {
					t := time.Now().UTC().Add(-1 * time.Hour)

					return &t
				}(),
			},
			wantValid: false,
		},
		{
			name: "used_nonce",
			factor: &MFAFactor{
				Nonce: "used_nonce",
				NonceUsedAt: func() *time.Time {
					t := time.Now().UTC()

					return &t
				}(),
			},
			wantValid: false,
		},
		{
			name: "used_and_expired_nonce",
			factor: &MFAFactor{
				Nonce: "used_expired_nonce",
				NonceExpiresAt: func() *time.Time {
					t := time.Now().UTC().Add(-1 * time.Hour)

					return &t
				}(),
				NonceUsedAt: func() *time.Time {
					t := time.Now().UTC().Add(-30 * time.Minute)

					return &t
				}(),
			},
			wantValid: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			isValid := tc.factor.IsNonceValid()
			require.Equal(t, tc.wantValid, isValid)
		})
	}
}

func TestMFAFactor_MarkNonceAsUsed(t *testing.T) {
	t.Parallel()

	factor := &MFAFactor{
		Nonce: "test_nonce",
	}

	require.Nil(t, factor.NonceUsedAt)

	factor.MarkNonceAsUsed()

	require.NotNil(t, factor.NonceUsedAt)
	require.WithinDuration(t, time.Now().UTC(), *factor.NonceUsedAt, 1*time.Second)
}
