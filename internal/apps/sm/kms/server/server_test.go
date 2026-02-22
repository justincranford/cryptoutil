package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewKMSServer_NilChecks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ctx     context.Context
		wantErr string
	}{
		{
			name:    "nil context",
			ctx:     nil,
			wantErr: "context cannot be nil",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server, err := NewKMSServer(tc.ctx, nil) //nolint:staticcheck // SA1012: Intentionally testing nil context handling
			require.Error(t, err)
			require.Nil(t, server)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestNewKMSServer_NilSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	server, err := NewKMSServer(ctx, nil)
	require.Error(t, err)
	require.Nil(t, server)
	require.Contains(t, err.Error(), "settings cannot be nil")
}

func TestKMSServer_StartNotInitialized(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		server *KMSServer
	}{
		{
			name:   "nil resources",
			server: &KMSServer{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.server.Start()
			require.Error(t, err)
			require.Contains(t, err.Error(), "server not initialized")
		})
	}
}

func TestKMSServer_ShutdownNilFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		server *KMSServer
	}{
		{
			name:   "all nil fields",
			server: &KMSServer{},
		},
		{
			name: "nil kmsCore only",
			server: &KMSServer{
				kmsCore: nil,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Should not panic with nil fields.
			require.NotPanics(t, func() {
				tc.server.Shutdown()
			})
		})
	}
}

func TestKMSServer_Accessors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		server *KMSServer
	}{
		{
			name:   "zero value server",
			server: &KMSServer{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.False(t, tc.server.IsReady())
			require.Equal(t, 0, tc.server.PublicPort())
			require.Equal(t, 0, tc.server.AdminPort())
			require.Empty(t, tc.server.PublicBaseURL())
			require.Empty(t, tc.server.AdminBaseURL())
			require.Nil(t, tc.server.Resources())
			require.Nil(t, tc.server.KMSCore())
			require.Nil(t, tc.server.Settings())
		})
	}
}
