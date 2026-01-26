// Copyright (c) 2025 Justin Cranford

package auth

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric/noop"
	tracenoop "go.opentelemetry.io/otel/trace/noop"

	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
)

// TestMFAOrchestrator_GetRequiredFactors tests retrieving required MFA factors.
func TestMFAOrchestrator_GetRequiredFactors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	authProfileID := googleUuid.New()

	tests := []struct {
		name          string
		setupRepo     func() *mockMFAFactorRepo
		expectedCount int
		wantErr       bool
		errContains   string
	}{
		{
			name: "multiple_factors",
			setupRepo: func() *mockMFAFactorRepo {
				repo := &mockMFAFactorRepo{
					factors: []*cryptoutilIdentityDomain.MFAFactor{
						{FactorType: cryptoutilIdentityDomain.MFAFactorTypeTOTP},
						{FactorType: cryptoutilIdentityDomain.MFAFactorTypeEmailOTP},
					},
				}

				return repo
			},
			expectedCount: 2,
			wantErr:       false,
		},
		{
			name: "single_factor",
			setupRepo: func() *mockMFAFactorRepo {
				repo := &mockMFAFactorRepo{
					factors: []*cryptoutilIdentityDomain.MFAFactor{
						{FactorType: cryptoutilIdentityDomain.MFAFactorTypeTOTP},
					},
				}

				return repo
			},
			expectedCount: 1,
			wantErr:       false,
		},
		{
			name: "no_factors",
			setupRepo: func() *mockMFAFactorRepo {
				return &mockMFAFactorRepo{factors: []*cryptoutilIdentityDomain.MFAFactor{}}
			},
			expectedCount: 0,
			wantErr:       false,
		},
		{
			name: "repository_error",
			setupRepo: func() *mockMFAFactorRepo {
				return &mockMFAFactorRepo{
					getByAuthProfileErr: fmt.Errorf("database connection failed"),
				}
			},
			wantErr:     true,
			errContains: "failed to fetch MFA factors",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := tc.setupRepo()
			telemetry := newTestTelemetry(t)
			orchestrator := NewMFAOrchestrator(repo, nil, nil, telemetry, nil)

			factors, err := orchestrator.GetRequiredFactors(ctx, authProfileID)

			if tc.wantErr {
				require.Error(t, err)

				if tc.errContains != "" {
					require.ErrorContains(t, err, tc.errContains)
				}
			} else {
				require.NoError(t, err)
				require.Len(t, factors, tc.expectedCount)
			}
		})
	}
}

// TestMFAOrchestrator_RequiresMFA tests checking if MFA is required.
func TestMFAOrchestrator_RequiresMFA(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	authProfileID := googleUuid.New()

	tests := []struct {
		name         string
		setupRepo    func() *mockMFAFactorRepo
		expectResult bool
		wantErr      bool
	}{
		{
			name: "requires_mfa_with_factors",
			setupRepo: func() *mockMFAFactorRepo {
				return &mockMFAFactorRepo{
					factors: []*cryptoutilIdentityDomain.MFAFactor{
						{FactorType: cryptoutilIdentityDomain.MFAFactorTypeTOTP},
					},
				}
			},
			expectResult: true,
			wantErr:      false,
		},
		{
			name: "no_mfa_required_empty",
			setupRepo: func() *mockMFAFactorRepo {
				return &mockMFAFactorRepo{factors: []*cryptoutilIdentityDomain.MFAFactor{}}
			},
			expectResult: false,
			wantErr:      false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := tc.setupRepo()
			telemetry := newTestTelemetry(t)
			orchestrator := NewMFAOrchestrator(repo, nil, nil, telemetry, nil)

			requiresMFA, err := orchestrator.RequiresMFA(ctx, authProfileID)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectResult, requiresMFA)
			}
		})
	}
}

// TestMFAOrchestrator_ValidateFactor tests validating MFA factors.
func TestMFAOrchestrator_ValidateFactor(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	authProfileID := googleUuid.New()

	tests := []struct {
		name        string
		factorType  string
		setupRepo   func() *mockMFAFactorRepo
		credentials map[string]string
		wantErr     bool
		errContains string
	}{
		{
			name:       "factor_not_found",
			factorType: "unknown_factor",
			setupRepo: func() *mockMFAFactorRepo {
				return &mockMFAFactorRepo{factors: []*cryptoutilIdentityDomain.MFAFactor{}}
			},
			credentials: map[string]string{"code": "123456"},
			wantErr:     true,
			errContains: "MFA factor not found",
		},
		{
			name:       "nonce_expired",
			factorType: string(cryptoutilIdentityDomain.MFAFactorTypeTOTP),
			setupRepo: func() *mockMFAFactorRepo {
				now := time.Now().UTC()
				expiry := now.Add(-5 * time.Minute)
				factor := &cryptoutilIdentityDomain.MFAFactor{
					ID:             googleUuid.New(),
					Name:           "test-totp-expired",
					AuthProfileID:  authProfileID,
					FactorType:     cryptoutilIdentityDomain.MFAFactorTypeTOTP,
					Nonce:          googleUuid.NewString(),
					NonceExpiresAt: &expiry, // Expired.
				}

				return &mockMFAFactorRepo{factors: []*cryptoutilIdentityDomain.MFAFactor{factor}}
			},
			credentials: map[string]string{"code": "123456"},
			wantErr:     true,
			errContains: "nonce already used or expired",
		},
		{
			name:       "repository_fetch_error",
			factorType: string(cryptoutilIdentityDomain.MFAFactorTypeTOTP),
			setupRepo: func() *mockMFAFactorRepo {
				return &mockMFAFactorRepo{
					getByAuthProfileErr: fmt.Errorf("database unavailable"),
				}
			},
			credentials: map[string]string{"code": "123456"},
			wantErr:     true,
			errContains: "failed to fetch MFA factors",
		},
		{
			name:       "unsupported_factor_type",
			factorType: "unknown_mfa_type",
			setupRepo: func() *mockMFAFactorRepo {
				now := time.Now().UTC()
				expiry := now.Add(5 * time.Minute)
				factor := &cryptoutilIdentityDomain.MFAFactor{
					ID:             googleUuid.New(),
					Name:           "test-unknown",
					AuthProfileID:  authProfileID,
					FactorType:     "unknown_mfa_type",
					Nonce:          googleUuid.NewString(),
					NonceExpiresAt: &expiry,
				}

				return &mockMFAFactorRepo{factors: []*cryptoutilIdentityDomain.MFAFactor{factor}}
			},
			credentials: map[string]string{"code": "123456"},
			wantErr:     true,
			errContains: "unsupported MFA factor type",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := tc.setupRepo()
			telemetry := newTestTelemetry(t)
			orchestrator := NewMFAOrchestrator(repo, nil, nil, telemetry, nil)

			err := orchestrator.ValidateFactor(ctx, authProfileID, tc.factorType, tc.credentials)

			if tc.wantErr {
				require.Error(t, err)

				if tc.errContains != "" {
					require.ErrorContains(t, err, tc.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Mock implementations.

type mockMFAFactorRepo struct {
	factors             []*cryptoutilIdentityDomain.MFAFactor
	updateFunc          func(context.Context, *cryptoutilIdentityDomain.MFAFactor) error
	getByAuthProfileErr error
}

func (m *mockMFAFactorRepo) Create(_ context.Context, factor *cryptoutilIdentityDomain.MFAFactor) error {
	m.factors = append(m.factors, factor)

	return nil
}

func (m *mockMFAFactorRepo) GetByID(_ context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.MFAFactor, error) {
	for _, f := range m.factors {
		if f.ID == id {
			return f, nil
		}
	}

	return nil, nil
}

func (m *mockMFAFactorRepo) GetByAuthProfileID(_ context.Context, _ googleUuid.UUID) ([]*cryptoutilIdentityDomain.MFAFactor, error) {
	if m.getByAuthProfileErr != nil {
		return nil, m.getByAuthProfileErr
	}

	return m.factors, nil
}

func (m *mockMFAFactorRepo) Update(ctx context.Context, factor *cryptoutilIdentityDomain.MFAFactor) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, factor)
	}

	return nil
}

func (m *mockMFAFactorRepo) Delete(_ context.Context, _ googleUuid.UUID) error {
	return nil
}

func (m *mockMFAFactorRepo) List(_ context.Context, _, _ int) ([]*cryptoutilIdentityDomain.MFAFactor, error) {
	return m.factors, nil
}

func (m *mockMFAFactorRepo) Count(_ context.Context) (int64, error) {
	return int64(len(m.factors)), nil
}

// newTestTelemetry creates MFA telemetry with noop providers for testing.
func newTestTelemetry(t *testing.T) *MFATelemetry {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	metricsProvider := noop.NewMeterProvider()
	tracesProvider := tracenoop.NewTracerProvider()

	telemetry, err := NewMFATelemetry(logger, metricsProvider, tracesProvider)
	require.NoError(t, err)

	return telemetry
}
