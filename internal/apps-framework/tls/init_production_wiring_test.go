// Copyright (c) 2025-2026 Justin Cranford.
//

package tls_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
)

// TestProductionNewTelemetryService_Success verifies that productionNewTelemetryService
// creates a real TelemetryService without error using a valid context.
// OTLP connections are established lazily so no live collector is required.
func TestProductionNewTelemetryService_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ts, err := cryptoutilAppsFrameworkTls.ExportedProductionNewTelemetryService(ctx)
	require.NoError(t, err)
	require.NotNil(t, ts)

	t.Cleanup(ts.Shutdown)
}

// TestProductionNewGenerator_Success verifies that productionNewGenerator creates a
// real Generator (including a live ECDSA P-384 key pool) without error.
func TestProductionNewGenerator_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ts, err := cryptoutilAppsFrameworkTls.ExportedProductionNewTelemetryService(ctx)
	require.NoError(t, err)
	require.NotNil(t, ts)

	t.Cleanup(ts.Shutdown)

	gen, err := cryptoutilAppsFrameworkTls.ExportedProductionNewGenerator(ctx, ts)
	require.NoError(t, err)
	require.NotNil(t, gen)

	t.Cleanup(gen.Shutdown)
}

// TestProductionGenerator_WriteClosures exercises the encodePKCS12Fn and
// encodeTrustPKCS12Fn closure bodies inside NewGenerator's return struct.
// Those closures are only reachable via a Generator created by NewGenerator
// (not ExportedNewTestGenerator which injects stubs), and they are only invoked
// when writeKeystore / writeTruststore are called on the production Generator.
// A stub P-256 subject (fast to create) is used — pkcs12.Modern.Encode accepts any key type.
func TestProductionGenerator_WriteClosures(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ts, err := cryptoutilAppsFrameworkTls.ExportedProductionNewTelemetryService(ctx)
	require.NoError(t, err)
	t.Cleanup(ts.Shutdown)

	gen, err := cryptoutilAppsFrameworkTls.ExportedProductionNewGenerator(ctx, ts)
	require.NoError(t, err)
	t.Cleanup(gen.Shutdown)

	// makeStubSubject is defined in generator_test.go (same package tls_test).
	// It creates a real P-256 self-signed subject fast enough for unit tests.
	subject := makeStubSubject(t)
	tmpDir := t.TempDir()

	// ExportedWriteKeystore calls encodePKCS12Fn — exercising that closure body.
	require.NoError(t, gen.ExportedWriteKeystore(tmpDir, "test-ks", subject))

	// ExportedWriteTruststore calls encodeTrustPKCS12Fn — exercising that closure body.
	require.NoError(t, gen.ExportedWriteTruststore(tmpDir, "test-ts", subject))
}
