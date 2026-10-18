// Copyright (c) 2025-2026 Justin Cranford.
//

package jwkservice

import (
	"testing"

	cryptoutilOpenapiModel "cryptoutil/api/sm-kms/models"
	cryptoutilAppsSmKmsModel "cryptoutil/internal/apps/sm-kms/server/model"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	jose "github.com/go-jose/go-jose/v4"
	"github.com/stretchr/testify/require"
)

var (
	testGenerateRSA2048 = cryptoutilOpenapiModel.RSA2048
	testGenerateRSA3072 = cryptoutilOpenapiModel.RSA3072
	testGenerateRSA4096 = cryptoutilOpenapiModel.RSA4096
	testGenerateECP256  = cryptoutilOpenapiModel.ECP256
	testGenerateECP384  = cryptoutilOpenapiModel.ECP384
	testGenerateECP521  = cryptoutilOpenapiModel.ECP521
	testGenerateOKP     = cryptoutilOpenapiModel.OKPEd25519
	testGenerateOct128  = cryptoutilOpenapiModel.Oct128
	testGenerateOct192  = cryptoutilOpenapiModel.Oct192
	testGenerateOct256  = cryptoutilOpenapiModel.Oct256
	testGenerateOct384  = cryptoutilOpenapiModel.Oct384
	testGenerateOct512  = cryptoutilOpenapiModel.Oct512
)

func TestAlgorithmMappings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		algorithm           string
		keyType             string
		generate            *cryptoutilOpenapiModel.GenerateAlgorithm
		rotationUnsupported bool
		signature           jose.SignatureAlgorithm
		keyAlg              jose.KeyAlgorithm
		contentEnc          jose.ContentEncryption
	}{
		{name: "rsa 2048 signing", algorithm: cryptoutilSharedMagic.JoseAlgRS256, keyType: cryptoutilAppsSmKmsModel.KeyTypeRSA, generate: &testGenerateRSA2048, signature: jose.RS256},
		{name: "rsa 3072 signing", algorithm: cryptoutilSharedMagic.JoseAlgPS256, keyType: cryptoutilAppsSmKmsModel.KeyTypeRSA, generate: &testGenerateRSA3072, signature: jose.PS256},
		{name: "rsa 4096", algorithm: cryptoutilSharedMagic.JoseKeyTypeRSA4096, keyType: cryptoutilAppsSmKmsModel.KeyTypeRSA, generate: &testGenerateRSA4096, signature: jose.RS512, keyAlg: jose.RSA_OAEP_256, contentEnc: jose.A256GCM},
		{name: "rsa 2048 key type", algorithm: cryptoutilSharedMagic.JoseKeyTypeRSA2048, keyType: cryptoutilAppsSmKmsModel.KeyTypeRSA, generate: &testGenerateRSA2048, signature: jose.RS256, keyAlg: jose.RSA_OAEP_256, contentEnc: jose.A256GCM},
		{name: "rsa 3072 key type", algorithm: cryptoutilSharedMagic.JoseKeyTypeRSA3072, keyType: cryptoutilAppsSmKmsModel.KeyTypeRSA, generate: &testGenerateRSA3072, signature: jose.RS384, keyAlg: jose.RSA_OAEP_256, contentEnc: jose.A256GCM},
		{name: "rsa oaep", algorithm: cryptoutilSharedMagic.JoseAlgRSAOAEP, keyAlg: jose.RSA_OAEP_256, contentEnc: jose.A256GCM},
		{name: "rsa oaep 256", algorithm: cryptoutilSharedMagic.JoseAlgRSAOAEP256, keyAlg: jose.RSA_OAEP_256, contentEnc: jose.A256GCM},
		{name: "rsa pss 384", algorithm: cryptoutilSharedMagic.JoseAlgPS384, keyType: cryptoutilAppsSmKmsModel.KeyTypeRSA, generate: &testGenerateRSA3072, signature: jose.PS384},
		{name: "rsa pss 512", algorithm: cryptoutilSharedMagic.JoseAlgPS512, keyType: cryptoutilAppsSmKmsModel.KeyTypeRSA, generate: &testGenerateRSA3072, signature: jose.PS512},
		{name: "ec 256", algorithm: cryptoutilSharedMagic.JoseAlgES256, keyType: cryptoutilAppsSmKmsModel.KeyTypeEC, generate: &testGenerateECP256, signature: jose.ES256},
		{name: "ec 384", algorithm: cryptoutilSharedMagic.JoseAlgES384, keyType: cryptoutilAppsSmKmsModel.KeyTypeEC, generate: &testGenerateECP384, signature: jose.ES384},
		{name: "ec 521", algorithm: cryptoutilSharedMagic.JoseAlgES512, keyType: cryptoutilAppsSmKmsModel.KeyTypeEC, generate: &testGenerateECP521, signature: jose.ES512},
		{name: "ec key type 256", algorithm: cryptoutilSharedMagic.JoseKeyTypeECP256, keyType: cryptoutilAppsSmKmsModel.KeyTypeEC, generate: &testGenerateECP256, signature: jose.ES256, keyAlg: jose.ECDH_ES_A256KW, contentEnc: jose.A256GCM},
		{name: "ec key type 384", algorithm: cryptoutilSharedMagic.JoseKeyTypeECP384, keyType: cryptoutilAppsSmKmsModel.KeyTypeEC, generate: &testGenerateECP384, signature: jose.ES384, keyAlg: jose.ECDH_ES_A256KW, contentEnc: jose.A256GCM},
		{name: "ec key type 521", algorithm: cryptoutilSharedMagic.JoseKeyTypeECP521, keyType: cryptoutilAppsSmKmsModel.KeyTypeEC, generate: &testGenerateECP521, signature: jose.ES512, keyAlg: jose.ECDH_ES_A256KW, contentEnc: jose.A256GCM},
		{name: "ecdh es", algorithm: cryptoutilSharedMagic.JoseAlgECDHES, keyAlg: jose.ECDH_ES_A256KW, contentEnc: jose.A256GCM},
		{name: "eddsa", algorithm: cryptoutilSharedMagic.JoseAlgEdDSA, keyType: cryptoutilAppsSmKmsModel.KeyTypeOKP, generate: &testGenerateOKP, signature: jose.EdDSA},
		{name: "ed25519 key type", algorithm: cryptoutilSharedMagic.JoseKeyTypeOKPEd25519, keyType: cryptoutilAppsSmKmsModel.KeyTypeOKP, generate: &testGenerateOKP, signature: jose.EdDSA},
		{name: "aes 128 gcm", algorithm: cryptoutilSharedMagic.JoseEncA128GCM, keyType: cryptoutilAppsSmKmsModel.KeyTypeOct, generate: &testGenerateOct128},
		{name: "aes 192 gcm", algorithm: cryptoutilSharedMagic.JoseEncA192GCM, keyType: cryptoutilAppsSmKmsModel.KeyTypeOct, generate: &testGenerateOct192},
		{name: "aes 256 gcm", algorithm: cryptoutilSharedMagic.JoseEncA256GCM, keyType: cryptoutilAppsSmKmsModel.KeyTypeOct, generate: &testGenerateOct256},
		{name: "aes 128 cbc hmac", algorithm: cryptoutilSharedMagic.JoseEncA128CBCHS256, keyType: cryptoutilAppsSmKmsModel.KeyTypeOct, generate: &testGenerateOct384},
		{name: "aes 192 cbc hmac", algorithm: cryptoutilSharedMagic.JoseEncA192CBCHS384, keyType: cryptoutilAppsSmKmsModel.KeyTypeOct, generate: &testGenerateOct384, rotationUnsupported: true},
		{name: "aes 256 cbc hmac", algorithm: cryptoutilSharedMagic.JoseEncA256CBCHS512, keyType: cryptoutilAppsSmKmsModel.KeyTypeOct, generate: &testGenerateOct512},
		{name: "oct 128", algorithm: cryptoutilSharedMagic.JoseKeyTypeOct128, keyType: cryptoutilAppsSmKmsModel.KeyTypeOct, generate: &testGenerateOct128, keyAlg: jose.DIRECT, contentEnc: jose.A128GCM},
		{name: "oct 192", algorithm: cryptoutilSharedMagic.JoseKeyTypeOct192, keyType: cryptoutilAppsSmKmsModel.KeyTypeOct, generate: &testGenerateOct192, keyAlg: jose.DIRECT, contentEnc: jose.A192GCM},
		{name: "oct 256", algorithm: cryptoutilSharedMagic.JoseKeyTypeOct256, keyType: cryptoutilAppsSmKmsModel.KeyTypeOct, generate: &testGenerateOct256, signature: jose.HS256, keyAlg: jose.DIRECT, contentEnc: jose.A256GCM},
		{name: "oct 384", algorithm: cryptoutilSharedMagic.JoseKeyTypeOct384, keyType: cryptoutilAppsSmKmsModel.KeyTypeOct, generate: &testGenerateOct384, signature: jose.HS384},
		{name: "oct 512", algorithm: cryptoutilSharedMagic.JoseKeyTypeOct512, keyType: cryptoutilAppsSmKmsModel.KeyTypeOct, generate: &testGenerateOct512, signature: jose.HS512},
		{name: "direct encryption", algorithm: cryptoutilSharedMagic.JoseAlgDir, keyAlg: jose.DIRECT, contentEnc: jose.A128GCM},
		{name: "a128kw", algorithm: string(jose.A128KW), keyAlg: jose.A128KW, contentEnc: jose.A128GCM},
		{name: "a192kw", algorithm: string(jose.A192KW), keyAlg: jose.A192KW, contentEnc: jose.A192GCM},
		{name: "a256kw", algorithm: string(jose.A256KW), keyAlg: jose.A256KW, contentEnc: jose.A256GCM},
		{name: "a128gcmkw", algorithm: string(jose.A128GCMKW), keyAlg: jose.A128GCMKW, contentEnc: jose.A128GCM},
		{name: "a192gcmkw", algorithm: string(jose.A192GCMKW), keyAlg: jose.A192GCMKW, contentEnc: jose.A192GCM},
		{name: "a256gcmkw", algorithm: string(jose.A256GCMKW), keyAlg: jose.A256GCMKW, contentEnc: jose.A256GCM},
		{name: "hmac 256", algorithm: cryptoutilSharedMagic.JoseAlgHS256, signature: jose.HS256},
		{name: "hmac 384", algorithm: cryptoutilSharedMagic.JoseAlgHS384, signature: jose.HS384},
		{name: "hmac 512", algorithm: cryptoutilSharedMagic.JoseAlgHS512, signature: jose.HS512},
		{name: "unknown algorithm", algorithm: cryptoutilSharedMagic.UNKNOWN},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.keyType, mapAlgorithmToKeyType(tc.algorithm))
			require.Equal(t, tc.generate, mapToGenerateAlgorithm(tc.algorithm))

			rotationAlgorithm := mapToGenerateAlgorithmForRotation(tc.algorithm)
			if tc.rotationUnsupported {
				require.Nil(t, rotationAlgorithm)
			} else {
				require.Equal(t, tc.generate, rotationAlgorithm)
			}

			keyAlg, contentEnc := mapToJWEAlgorithms(tc.algorithm)
			require.Equal(t, tc.keyAlg, keyAlg)
			require.Equal(t, tc.contentEnc, contentEnc)
			require.Equal(t, tc.signature, mapToSignatureAlgorithm(tc.algorithm))
		})
	}
}
