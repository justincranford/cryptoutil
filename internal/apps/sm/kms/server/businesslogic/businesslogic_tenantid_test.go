package businesslogic

import (
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilKmsMiddleware "cryptoutil/internal/apps/sm/kms/server/middleware"
)

func TestGetTenantID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupCtx   func() context.Context
		wantErr    bool
		wantErrMsg string
		wantID     googleUuid.UUID
	}{
		{
			name: "valid tenant ID in realm context",
			setupCtx: func() context.Context {
				tenantID := googleUuid.New()
				rc := &cryptoutilKmsMiddleware.RealmContext{
					TenantID: tenantID,
				}

				return context.WithValue(context.Background(), cryptoutilKmsMiddleware.RealmContextKey{}, rc)
			},
			wantErr: false,
		},
		{
			name: "no realm context in context",
			setupCtx: func() context.Context {
				return context.Background()
			},
			wantErr:    true,
			wantErrMsg: "tenant context required",
		},
		{
			name: "nil tenant ID in realm context",
			setupCtx: func() context.Context {
				rc := &cryptoutilKmsMiddleware.RealmContext{
					TenantID: googleUuid.Nil,
				}

				return context.WithValue(context.Background(), cryptoutilKmsMiddleware.RealmContextKey{}, rc)
			},
			wantErr:    true,
			wantErrMsg: "tenant context required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := tc.setupCtx()
			tenantID, err := getTenantID(ctx)

			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErrMsg)
				require.Equal(t, googleUuid.Nil, tenantID)
			} else {
				require.NoError(t, err)
				require.NotEqual(t, googleUuid.Nil, tenantID)
			}
		})
	}
}
