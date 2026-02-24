// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

package container

import (
	"context"
	"fmt"
	"testing"

	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

// mockContainer implements testcontainers.Container for testing GetContainerHostAndMappedPort.
// Only Host and MappedPort are called by the function under test; all other methods
// are promoted from the embedded nil interface and will panic if ever called.
type mockContainer struct {
	testcontainers.Container
	hostFn       func(ctx context.Context) (string, error)
	mappedPortFn func(ctx context.Context, port nat.Port) (nat.Port, error)
}

func (m *mockContainer) Host(ctx context.Context) (string, error) {
	return m.hostFn(ctx)
}

func (m *mockContainer) MappedPort(ctx context.Context, port nat.Port) (nat.Port, error) {
	return m.mappedPortFn(ctx, port)
}

func setupTestTelemetry(t *testing.T) *cryptoutilSharedTelemetry.TelemetryService {
	t.Helper()

	ctx := context.Background()
	settings := cryptoutilSharedTelemetry.NewTestTelemetrySettings("container_test")

	svc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, settings)
	require.NoError(t, err)

	t.Cleanup(svc.Shutdown)

	return svc
}

// TestGetContainerHostAndMappedPort covers success and error paths using a mock container.
func TestGetContainerHostAndMappedPort(t *testing.T) {
	t.Parallel()

	telemetrySvc := setupTestTelemetry(t)

	tests := []struct {
		name       string
		hostFn     func(ctx context.Context) (string, error)
		mappedFn   func(ctx context.Context, port nat.Port) (nat.Port, error)
		wantHost   string
		wantPort   string
		wantErrMsg string
	}{
		{
			name: "success",
			hostFn: func(_ context.Context) (string, error) {
				return cryptoutilSharedMagic.IPv4Loopback, nil
			},
			mappedFn: func(_ context.Context, _ nat.Port) (nat.Port, error) {
				return nat.Port("54321/tcp"), nil
			},
			wantHost: cryptoutilSharedMagic.IPv4Loopback,
			wantPort: "54321",
		},
		{
			name: "host error",
			hostFn: func(_ context.Context) (string, error) {
				return "", fmt.Errorf("host lookup failed")
			},
			mappedFn: func(_ context.Context, _ nat.Port) (nat.Port, error) {
				return "", nil
			},
			wantErrMsg: "failed to get container host",
		},
		{
			name: "mapped port error",
			hostFn: func(_ context.Context) (string, error) {
				return cryptoutilSharedMagic.IPv4Loopback, nil
			},
			mappedFn: func(_ context.Context, _ nat.Port) (nat.Port, error) {
				return "", fmt.Errorf("port mapping not found")
			},
			wantErrMsg: "failed to get container mapped port",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mock := &mockContainer{
				hostFn:       tc.hostFn,
				mappedPortFn: tc.mappedFn,
			}

			host, port, err := GetContainerHostAndMappedPort(context.Background(), telemetrySvc, mock, "5432")

			if tc.wantErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErrMsg)
				require.Empty(t, host)
				require.Empty(t, port)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantHost, host)
				require.Equal(t, tc.wantPort, port)
			}
		})
	}
}

// TestVerifyPostgresConnection_InvalidDSN verifies error handling when the connection string
// points to a non-existent PostgreSQL server.
func TestVerifyPostgresConnection_InvalidDSN(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		connStr string
	}{
		{
			name:    "unreachable host",
			connStr: "postgres://user:pass@127.0.0.1:1/nonexistent?sslmode=disable&connect_timeout=1",
		},
		{
			name:    "malformed DSN",
			connStr: "not-a-valid-dsn",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := VerifyPostgresConnection(tc.connStr)
			require.Error(t, err)
		})
	}
}
