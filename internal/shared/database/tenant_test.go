// Copyright (c) 2025 Justin Cranford.
// SPDX-License-Identifier: Apache-2.0.

package database

import (
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestTenantContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		tenantID googleUuid.UUID
		realmID  googleUuid.UUID
		userID   googleUuid.UUID
	}{
		{
			name:     "valid context",
			tenantID: googleUuid.New(),
			realmID:  googleUuid.New(),
			userID:   googleUuid.New(),
		},
		{
			name:     "tenant only",
			tenantID: googleUuid.New(),
			realmID:  googleUuid.Nil,
			userID:   googleUuid.Nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tc := &TenantContext{
				TenantID: tt.tenantID,
				RealmID:  tt.realmID,
				UserID:   tt.userID,
			}

			ctx := WithTenantContext(context.Background(), tc)
			retrieved := GetTenantContext(ctx)
			require.NotNil(t, retrieved)
			require.Equal(t, tt.tenantID, retrieved.TenantID)
			require.Equal(t, tt.realmID, retrieved.RealmID)
			require.Equal(t, tt.userID, retrieved.UserID)
		})
	}
}

func TestGetTenantContext_Nil(t *testing.T) {
	t.Parallel()

	tc := GetTenantContext(context.Background())
	require.Nil(t, tc)
}

func TestGetTenantID(t *testing.T) {
	t.Parallel()

	testTenantID := googleUuid.Must(googleUuid.NewV7())
	tests := []struct {
		name     string
		ctx      context.Context
		expected googleUuid.UUID
	}{
		{
			name:     "no context",
			ctx:      context.Background(),
			expected: googleUuid.Nil,
		},
		{
			name:     "with context",
			ctx:      WithTenantContext(context.Background(), &TenantContext{TenantID: testTenantID}),
			expected: testTenantID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := GetTenantID(tt.ctx)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestMustGetTenantID_Panic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ctx  context.Context
	}{
		{
			name: "no context",
			ctx:  context.Background(),
		},
		{
			name: "nil tenant id",
			ctx:  WithTenantContext(context.Background(), &TenantContext{TenantID: googleUuid.Nil}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Panics(t, func() {
				MustGetTenantID(tt.ctx)
			})
		})
	}
}

func TestMustGetTenantID_Success(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	ctx := WithTenantContext(context.Background(), &TenantContext{TenantID: tenantID})
	result := MustGetTenantID(ctx)
	require.Equal(t, tenantID, result)
}

func TestRequireTenantContext(t *testing.T) {
	t.Parallel()

	testTenantID := googleUuid.Must(googleUuid.NewV7())
	tests := []struct {
		name       string
		ctx        context.Context
		wantErr    error
		wantTenant googleUuid.UUID
	}{
		{
			name:    "no context",
			ctx:     context.Background(),
			wantErr: ErrNoTenantContext,
		},
		{
			name:    "nil tenant id",
			ctx:     WithTenantContext(context.Background(), &TenantContext{TenantID: googleUuid.Nil}),
			wantErr: ErrInvalidTenantID,
		},
		{
			name:       "valid context",
			ctx:        WithTenantContext(context.Background(), &TenantContext{TenantID: testTenantID}),
			wantErr:    nil,
			wantTenant: testTenantID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tc, err := RequireTenantContext(tt.ctx)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, tc)
			} else {
				require.NoError(t, err)
				require.NotNil(t, tc)
				require.Equal(t, tt.wantTenant, tc.TenantID)
			}
		})
	}
}
