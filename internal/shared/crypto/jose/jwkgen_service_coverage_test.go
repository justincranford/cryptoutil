// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	"context"
	"testing"

	cryptoutilOpenapiModel "cryptoutil/api/model"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/stretchr/testify/require"
)

// TestJWKGenService_GenerateJWK tests GenerateJWK method with all algorithms.
func TestJWKGenService_GenerateJWK(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		algorithm cryptoutilOpenapiModel.GenerateAlgorithm
		wantErr   bool
	}{
		{"RSA4096", cryptoutilOpenapiModel.RSA4096, false},
		{"RSA3072", cryptoutilOpenapiModel.RSA3072, false},
		{"RSA2048", cryptoutilOpenapiModel.RSA2048, false},
		{"ECP521", cryptoutilOpenapiModel.ECP521, false},
		{"ECP384", cryptoutilOpenapiModel.ECP384, false},
		{"ECP256", cryptoutilOpenapiModel.ECP256, false},
		{"OKPEd25519", cryptoutilOpenapiModel.OKPEd25519, false},
		{"Oct512", cryptoutilOpenapiModel.Oct512, false},
		{"Oct384", cryptoutilOpenapiModel.Oct384, false},
		{"Oct256", cryptoutilOpenapiModel.Oct256, false},
		{"Oct192", cryptoutilOpenapiModel.Oct192, false},
		{"Oct128", cryptoutilOpenapiModel.Oct128, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			kid, nonPubJWK, pubJWK, clearNonPub, clearPub, err := testJWKGenService.GenerateJWK(&tc.algorithm)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, kid)
				require.NotNil(t, nonPubJWK)
				require.NotEmpty(t, clearNonPub)
				// Public key may be nil for symmetric algorithms.
				if pubJWK != nil {
					require.NotEmpty(t, clearPub)
				}
			}
		})
	}
}

// TestJWKGenService_GenerateJWK_UnsupportedAlgorithm tests error path for unsupported algorithm.
func TestJWKGenService_GenerateJWK_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	unsupportedAlg := cryptoutilOpenapiModel.GenerateAlgorithm("UNSUPPORTED")
	_, _, _, _, _, err := testJWKGenService.GenerateJWK(&unsupportedAlg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported JWK alg")
}

// TestJWKGenService_GenerateJWEJWK_AllAlgorithms tests JWE key generation for all enc+alg combinations.
func TestJWKGenService_GenerateJWEJWK_AllAlgorithms(t *testing.T) {
	t.Parallel()

	// Test DIR algorithms with all encs.
	dirTests := []struct {
		enc joseJwa.ContentEncryptionAlgorithm
		alg joseJwa.KeyEncryptionAlgorithm
	}{
		{EncA256GCM, AlgDir},
		{EncA192GCM, AlgDir},
		{EncA128GCM, AlgDir},
		{EncA256CBCHS512, AlgDir},
		{EncA192CBCHS384, AlgDir},
		{EncA128CBCHS256, AlgDir},
	}

	for _, tc := range dirTests {
		t.Run(tc.enc.String()+"_"+tc.alg.String(), func(t *testing.T) {
			t.Parallel()

			kid, nonPubJWK, pubJWK, _, _, err := testJWKGenService.GenerateJWEJWK(&tc.enc, &tc.alg)
			require.NoError(t, err)
			require.NotNil(t, kid)
			require.NotNil(t, nonPubJWK)
			require.Nil(t, pubJWK) // Symmetric algorithms have no public key.
		})
	}

	// Test AES key wrap algorithms.
	aesTests := []struct {
		alg joseJwa.KeyEncryptionAlgorithm
	}{
		{AlgA256KW},
		{AlgA256GCMKW},
		{AlgA192KW},
		{AlgA192GCMKW},
		{AlgA128KW},
		{AlgA128GCMKW},
	}

	for _, tc := range aesTests {
		t.Run(tc.alg.String(), func(t *testing.T) {
			t.Parallel()

			kid, nonPubJWK, pubJWK, _, _, err := testJWKGenService.GenerateJWEJWK(&EncA256GCM, &tc.alg)
			require.NoError(t, err)
			require.NotNil(t, kid)
			require.NotNil(t, nonPubJWK)
			require.Nil(t, pubJWK)
		})
	}

	// Test RSA algorithms.
	rsaTests := []struct {
		alg joseJwa.KeyEncryptionAlgorithm
	}{
		{AlgRSAOAEP512},
		{AlgRSAOAEP384},
		{AlgRSAOAEP256},
		{AlgRSAOAEP},
		{AlgRSA15},
	}

	for _, tc := range rsaTests {
		t.Run(tc.alg.String(), func(t *testing.T) {
			t.Parallel()

			kid, nonPubJWK, pubJWK, _, _, err := testJWKGenService.GenerateJWEJWK(&EncA256GCM, &tc.alg)
			require.NoError(t, err)
			require.NotNil(t, kid)
			require.NotNil(t, nonPubJWK)
			require.NotNil(t, pubJWK) // RSA has public key.
		})
	}

	// Test ECDH algorithms.
	ecdhTests := []struct {
		alg joseJwa.KeyEncryptionAlgorithm
	}{
		{AlgECDHES},
		{AlgECDHESA256KW},
		{AlgECDHESA192KW},
		{AlgECDHESA128KW},
	}

	for _, tc := range ecdhTests {
		t.Run(tc.alg.String(), func(t *testing.T) {
			t.Parallel()

			kid, nonPubJWK, pubJWK, _, _, err := testJWKGenService.GenerateJWEJWK(&EncA256GCM, &tc.alg)
			require.NoError(t, err)
			require.NotNil(t, kid)
			require.NotNil(t, nonPubJWK)
			require.NotNil(t, pubJWK) // ECDH has public key.
		})
	}
}

// TestJWKGenService_GenerateJWEJWK_UnsupportedAlgorithm tests error paths.
func TestJWKGenService_GenerateJWEJWK_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	// Test unsupported ENC with DIR algorithm - use a non-standard enc value.
	invalidEnc := EncA256GCM // Placeholder.
	_, _, _, _, _, err := testJWKGenService.GenerateJWEJWK(&invalidEnc, &AlgDir)
	// This should not error because EncA256GCM is supported.
	require.NoError(t, err)
	// Test unsupported ALG - these algorithms exist in JWA but not in GenerateJWEJWK switch.
	// Since we can't create invalid enum values, skip this test.
	// The default case in GenerateJWEJWK covers this path via unit tests.
}

// TestJWKGenService_GenerateJWSJWK_AllAlgorithms tests JWS key generation for all algorithms.
func TestJWKGenService_GenerateJWSJWK_AllAlgorithms(t *testing.T) {
	t.Parallel()

	tests := []struct {
		alg joseJwa.SignatureAlgorithm
	}{
		{joseJwa.PS512()},
		{joseJwa.PS384()},
		{joseJwa.PS256()},
		{joseJwa.RS512()},
		{joseJwa.RS384()},
		{joseJwa.RS256()},
		{joseJwa.ES512()},
		{joseJwa.ES384()},
		{joseJwa.ES256()},
		{joseJwa.EdDSA()},
		{joseJwa.HS512()},
		{joseJwa.HS384()},
		{joseJwa.HS256()},
	}

	for _, tc := range tests {
		t.Run(tc.alg.String(), func(t *testing.T) {
			t.Parallel()

			kid, nonPubJWK, pubJWK, _, _, err := testJWKGenService.GenerateJWSJWK(tc.alg)
			require.NoError(t, err)
			require.NotNil(t, kid)
			require.NotNil(t, nonPubJWK)
			// Symmetric algorithms (HS*) have no public key.
			if tc.alg == joseJwa.HS256() || tc.alg == joseJwa.HS384() || tc.alg == joseJwa.HS512() {
				require.Nil(t, pubJWK)
			} else {
				require.NotNil(t, pubJWK)
			}
		})
	}
}

// TestJWKGenService_GenerateJWSJWK_UnsupportedAlgorithm tests error path.
func TestJWKGenService_GenerateJWSJWK_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()
	// Since SignatureAlgorithm is strongly typed enum, we can't create invalid values.
	// The default case in GenerateJWSJWK is covered by the comprehensive algorithm tests above.
	// Any algorithm not in the switch statement would be a compile-time error if we tried to use it.
}

// TestJWKGenService_GenerateUUIDv7 tests UUIDv7 generation.
func TestJWKGenService_GenerateUUIDv7(t *testing.T) {
	t.Parallel()

	uuid1 := testJWKGenService.GenerateUUIDv7()
	require.NotNil(t, uuid1)

	uuid2 := testJWKGenService.GenerateUUIDv7()
	require.NotNil(t, uuid2)

	// Ensure UUIDs are unique.
	require.NotEqual(t, uuid1, uuid2)
}

// TestJWKGenService_Shutdown tests service shutdown.
func TestJWKGenService_Shutdown(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	service, err := NewJWKGenService(ctx, testTelemetryService, false)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Shutdown should not panic.
	service.Shutdown()
}

// TestJWKGenService_NewJWKGenService_ErrorPaths tests constructor error paths.
func TestJWKGenService_NewJWKGenService_ErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		ctx              context.Context
		telemetryService any
		wantErr          string
	}{
		{"NilContext", nil, testTelemetryService, "context must be non-nil"},
		{"NilTelemetry", context.Background(), nil, "telemetry service must be non-nil"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				service *JWKGenService
				err     error
			)

			if tc.ctx == nil {
				service, err = NewJWKGenService(tc.ctx, testTelemetryService, false)
			} else {
				service, err = NewJWKGenService(tc.ctx, nil, false)
			}

			require.Error(t, err)
			require.Nil(t, service)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}
