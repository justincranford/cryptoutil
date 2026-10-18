// Copyright (c) 2025-2026 Justin Cranford.
//

package elastickeytest

import (
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestRequireNewGenElasticKeyForTest_NilTelemetry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		call func()
	}{
		{name: "RSA-4096", call: func() { RequireNewRSA4096GenElasticKeyForTest(nil) }},
		{name: "RSA-3072", call: func() { RequireNewRSA3072GenElasticKeyForTest(nil) }},
		{name: "RSA-2048", call: func() { RequireNewRSA2048GenElasticKeyForTest(nil) }},
		{name: "ECDSA-P521", call: func() { RequireNewECDSAP521GenElasticKeyForTest(nil) }},
		{name: "ECDSA-P384", call: func() { RequireNewECDSAP384GenElasticKeyForTest(nil) }},
		{name: "ECDSA-P256", call: func() { RequireNewECDSAP256GenElasticKeyForTest(nil) }},
		{name: "ECDH-P521", call: func() { RequireNewECDHP521GenElasticKeyForTest(nil) }},
		{name: "ECDH-P384", call: func() { RequireNewECDHP384GenElasticKeyForTest(nil) }},
		{name: "ECDH-P256", call: func() { RequireNewECDHP256GenElasticKeyForTest(nil) }},
		{name: cryptoutilSharedMagic.EdCurveEd25519, call: func() { RequireNewEd25519GenElasticKeyForTest(nil) }},
		{name: "AES-256-GCM", call: func() { RequireNewAES256GcmGenElasticKeyForTest(nil) }},
		{name: "AES-192-GCM", call: func() { RequireNewA192GcmGenElasticKeyForTest(nil) }},
		{name: "AES-128-GCM", call: func() { RequireNewA128GcmGenElasticKeyForTest(nil) }},
		{name: "AES-256-CBC-HS512", call: func() { RequireNewAES256CbcHs512GcmGenElasticKeyForTest(nil) }},
		{name: "AES-192-CBC-HS384", call: func() { RequireNewA192CbcHs384GenElasticKeyForTest(nil) }},
		{name: "AES-128-CBC-HS256", call: func() { RequireNewA128CbcHs256GenElasticKeyForTest(nil) }},
		{name: "UUIDv7", call: func() { RequireNewUUIDV7GenElasticKeyForTest(nil) }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Panics(t, tc.call)
		})
	}
}

func TestRequireNewUUIDV7GenElasticKeyForTest_CreatesPool(t *testing.T) {
	t.Parallel()

	pool := RequireNewUUIDV7GenElasticKeyForTest(testTelemetryService)
	require.NotNil(t, pool)
	require.Equal(t, "Test UUIDv7", pool.Name())
	pool.Cancel()
}
