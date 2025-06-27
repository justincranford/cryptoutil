package jose

import (
	cryptoutilBusinessModel "cryptoutil/internal/common/businessmodel"

	"fmt"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
)

var (
	elasticKeyAlgorithmToJoseEncAndAlg = map[cryptoutilBusinessModel.ElasticKeyAlgorithm]struct {
		enc *joseJwa.ContentEncryptionAlgorithm
		alg *joseJwa.KeyEncryptionAlgorithm
	}{
		cryptoutilBusinessModel.A256GCM_A256KW:    {enc: &EncA256GCM, alg: &AlgA256KW},
		cryptoutilBusinessModel.A192GCM_A256KW:    {enc: &EncA192GCM, alg: &AlgA256KW},
		cryptoutilBusinessModel.A128GCM_A256KW:    {enc: &EncA128GCM, alg: &AlgA256KW},
		cryptoutilBusinessModel.A256GCM_A192KW:    {enc: &EncA256GCM, alg: &AlgA192KW},
		cryptoutilBusinessModel.A192GCM_A192KW:    {enc: &EncA192GCM, alg: &AlgA192KW},
		cryptoutilBusinessModel.A128GCM_A192KW:    {enc: &EncA128GCM, alg: &AlgA192KW},
		cryptoutilBusinessModel.A256GCM_A128KW:    {enc: &EncA256GCM, alg: &AlgA128KW},
		cryptoutilBusinessModel.A192GCM_A128KW:    {enc: &EncA192GCM, alg: &AlgA128KW},
		cryptoutilBusinessModel.A128GCM_A128KW:    {enc: &EncA128GCM, alg: &AlgA128KW},
		cryptoutilBusinessModel.A256GCM_A256GCMKW: {enc: &EncA256GCM, alg: &AlgA256GCMKW},
		cryptoutilBusinessModel.A192GCM_A256GCMKW: {enc: &EncA192GCM, alg: &AlgA256GCMKW},
		cryptoutilBusinessModel.A128GCM_A256GCMKW: {enc: &EncA128GCM, alg: &AlgA256GCMKW},
		cryptoutilBusinessModel.A256GCM_A192GCMKW: {enc: &EncA256GCM, alg: &AlgA192GCMKW},
		cryptoutilBusinessModel.A192GCM_A192GCMKW: {enc: &EncA192GCM, alg: &AlgA192GCMKW},
		cryptoutilBusinessModel.A128GCM_A192GCMKW: {enc: &EncA128GCM, alg: &AlgA192GCMKW},
		cryptoutilBusinessModel.A256GCM_A128GCMKW: {enc: &EncA256GCM, alg: &AlgA128GCMKW},
		cryptoutilBusinessModel.A192GCM_A128GCMKW: {enc: &EncA192GCM, alg: &AlgA128GCMKW},
		cryptoutilBusinessModel.A128GCM_A128GCMKW: {enc: &EncA128GCM, alg: &AlgA128GCMKW},
		cryptoutilBusinessModel.A256GCM_dir:       {enc: &EncA256GCM, alg: &AlgDir},
		cryptoutilBusinessModel.A192GCM_dir:       {enc: &EncA192GCM, alg: &AlgDir},
		cryptoutilBusinessModel.A128GCM_dir:       {enc: &EncA128GCM, alg: &AlgDir},

		cryptoutilBusinessModel.A256GCM_RSAOAEP512: {enc: &EncA256GCM, alg: &AlgRSAOAEP512},
		cryptoutilBusinessModel.A192GCM_RSAOAEP512: {enc: &EncA192GCM, alg: &AlgRSAOAEP512},
		cryptoutilBusinessModel.A128GCM_RSAOAEP512: {enc: &EncA128GCM, alg: &AlgRSAOAEP512},
		cryptoutilBusinessModel.A256GCM_RSAOAEP384: {enc: &EncA256GCM, alg: &AlgRSAOAEP384},
		cryptoutilBusinessModel.A192GCM_RSAOAEP384: {enc: &EncA192GCM, alg: &AlgRSAOAEP384},
		cryptoutilBusinessModel.A128GCM_RSAOAEP384: {enc: &EncA128GCM, alg: &AlgRSAOAEP384},
		cryptoutilBusinessModel.A256GCM_RSAOAEP256: {enc: &EncA256GCM, alg: &AlgRSAOAEP256},
		cryptoutilBusinessModel.A192GCM_RSAOAEP256: {enc: &EncA192GCM, alg: &AlgRSAOAEP256},
		cryptoutilBusinessModel.A128GCM_RSAOAEP256: {enc: &EncA128GCM, alg: &AlgRSAOAEP256},
		cryptoutilBusinessModel.A256GCM_RSAOAEP:    {enc: &EncA256GCM, alg: &AlgRSAOAEP},
		cryptoutilBusinessModel.A192GCM_RSAOAEP:    {enc: &EncA192GCM, alg: &AlgRSAOAEP},
		cryptoutilBusinessModel.A128GCM_RSAOAEP:    {enc: &EncA128GCM, alg: &AlgRSAOAEP},
		cryptoutilBusinessModel.A256GCM_RSA15:      {enc: &EncA256GCM, alg: &AlgRSA15},
		cryptoutilBusinessModel.A192GCM_RSA15:      {enc: &EncA192GCM, alg: &AlgRSA15},
		cryptoutilBusinessModel.A128GCM_RSA15:      {enc: &EncA128GCM, alg: &AlgRSA15},

		cryptoutilBusinessModel.A256GCM_ECDHESA256KW: {enc: &EncA256GCM, alg: &AlgECDHESA256KW},
		cryptoutilBusinessModel.A192GCM_ECDHESA256KW: {enc: &EncA192GCM, alg: &AlgECDHESA256KW},
		cryptoutilBusinessModel.A128GCM_ECDHESA256KW: {enc: &EncA128GCM, alg: &AlgECDHESA256KW},
		cryptoutilBusinessModel.A256GCM_ECDHESA192KW: {enc: &EncA256GCM, alg: &AlgECDHESA192KW},
		cryptoutilBusinessModel.A192GCM_ECDHESA192KW: {enc: &EncA192GCM, alg: &AlgECDHESA192KW},
		cryptoutilBusinessModel.A128GCM_ECDHESA192KW: {enc: &EncA128GCM, alg: &AlgECDHESA192KW},
		cryptoutilBusinessModel.A256GCM_ECDHESA128KW: {enc: &EncA256GCM, alg: &AlgECDHESA128KW},
		cryptoutilBusinessModel.A192GCM_ECDHESA128KW: {enc: &EncA192GCM, alg: &AlgECDHESA128KW},
		cryptoutilBusinessModel.A128GCM_ECDHESA128KW: {enc: &EncA128GCM, alg: &AlgECDHESA128KW},
		cryptoutilBusinessModel.A256GCM_ECDHES:       {enc: &EncA256GCM, alg: &AlgECDHES},
		cryptoutilBusinessModel.A192GCM_ECDHES:       {enc: &EncA192GCM, alg: &AlgECDHES},
		cryptoutilBusinessModel.A128GCM_ECDHES:       {enc: &EncA128GCM, alg: &AlgECDHES},

		cryptoutilBusinessModel.A256CBCHS512_A256KW:    {enc: &EncA256CBC_HS512, alg: &AlgA256KW},
		cryptoutilBusinessModel.A192CBCHS384_A256KW:    {enc: &EncA192CBC_HS384, alg: &AlgA256KW},
		cryptoutilBusinessModel.A128CBCHS256_A256KW:    {enc: &EncA128CBC_HS256, alg: &AlgA256KW},
		cryptoutilBusinessModel.A256CBCHS512_A192KW:    {enc: &EncA256CBC_HS512, alg: &AlgA192KW},
		cryptoutilBusinessModel.A192CBCHS384_A192KW:    {enc: &EncA192CBC_HS384, alg: &AlgA192KW},
		cryptoutilBusinessModel.A128CBCHS256_A192KW:    {enc: &EncA128CBC_HS256, alg: &AlgA192KW},
		cryptoutilBusinessModel.A256CBCHS512_A128KW:    {enc: &EncA256CBC_HS512, alg: &AlgA128KW},
		cryptoutilBusinessModel.A192CBCHS384_A128KW:    {enc: &EncA192CBC_HS384, alg: &AlgA128KW},
		cryptoutilBusinessModel.A128CBCHS256_A128KW:    {enc: &EncA128CBC_HS256, alg: &AlgA128KW},
		cryptoutilBusinessModel.A256CBCHS512_A256GCMKW: {enc: &EncA256CBC_HS512, alg: &AlgA256GCMKW},
		cryptoutilBusinessModel.A192CBCHS384_A256GCMKW: {enc: &EncA192CBC_HS384, alg: &AlgA256GCMKW},
		cryptoutilBusinessModel.A128CBCHS256_A256GCMKW: {enc: &EncA128CBC_HS256, alg: &AlgA256GCMKW},
		cryptoutilBusinessModel.A256CBCHS512_A192GCMKW: {enc: &EncA256CBC_HS512, alg: &AlgA192GCMKW},
		cryptoutilBusinessModel.A192CBCHS384_A192GCMKW: {enc: &EncA192CBC_HS384, alg: &AlgA192GCMKW},
		cryptoutilBusinessModel.A128CBCHS256_A192GCMKW: {enc: &EncA128CBC_HS256, alg: &AlgA192GCMKW},
		cryptoutilBusinessModel.A256CBCHS512_A128GCMKW: {enc: &EncA256CBC_HS512, alg: &AlgA128GCMKW},
		cryptoutilBusinessModel.A192CBCHS384_A128GCMKW: {enc: &EncA192CBC_HS384, alg: &AlgA128GCMKW},
		cryptoutilBusinessModel.A128CBCHS256_A128GCMKW: {enc: &EncA128CBC_HS256, alg: &AlgA128GCMKW},
		cryptoutilBusinessModel.A256CBCHS512_dir:       {enc: &EncA256CBC_HS512, alg: &AlgDir},
		cryptoutilBusinessModel.A192CBCHS384_dir:       {enc: &EncA192CBC_HS384, alg: &AlgDir},
		cryptoutilBusinessModel.A128CBCHS256_dir:       {enc: &EncA128CBC_HS256, alg: &AlgDir},

		cryptoutilBusinessModel.A256CBC_HS512_RSAOAEP512: {enc: &EncA256CBC_HS512, alg: &AlgRSAOAEP512},
		cryptoutilBusinessModel.A192CBC_HS384_RSAOAEP512: {enc: &EncA192CBC_HS384, alg: &AlgRSAOAEP512},
		cryptoutilBusinessModel.A128CBC_HS256_RSAOAEP512: {enc: &EncA128CBC_HS256, alg: &AlgRSAOAEP512},
		cryptoutilBusinessModel.A256CBC_HS512_RSAOAEP384: {enc: &EncA256CBC_HS512, alg: &AlgRSAOAEP384},
		cryptoutilBusinessModel.A192CBC_HS384_RSAOAEP384: {enc: &EncA192CBC_HS384, alg: &AlgRSAOAEP384},
		cryptoutilBusinessModel.A128CBC_HS256_RSAOAEP384: {enc: &EncA128CBC_HS256, alg: &AlgRSAOAEP384},
		cryptoutilBusinessModel.A256CBC_HS512_RSAOAEP256: {enc: &EncA256CBC_HS512, alg: &AlgRSAOAEP256},
		cryptoutilBusinessModel.A192CBC_HS384_RSAOAEP256: {enc: &EncA192CBC_HS384, alg: &AlgRSAOAEP256},
		cryptoutilBusinessModel.A128CBC_HS256_RSAOAEP256: {enc: &EncA128CBC_HS256, alg: &AlgRSAOAEP256},
		cryptoutilBusinessModel.A256CBC_HS512_RSAOAEP:    {enc: &EncA256CBC_HS512, alg: &AlgRSAOAEP},
		cryptoutilBusinessModel.A192CBC_HS384_RSAOAEP:    {enc: &EncA192CBC_HS384, alg: &AlgRSAOAEP},
		cryptoutilBusinessModel.A128CBC_HS256_RSAOAEP:    {enc: &EncA128CBC_HS256, alg: &AlgRSAOAEP},
		cryptoutilBusinessModel.A256CBC_HS512_RSA15:      {enc: &EncA256CBC_HS512, alg: &AlgRSA15},
		cryptoutilBusinessModel.A192CBC_HS384_RSA15:      {enc: &EncA192CBC_HS384, alg: &AlgRSA15},
		cryptoutilBusinessModel.A128CBC_HS256_RSA15:      {enc: &EncA128CBC_HS256, alg: &AlgRSA15},

		cryptoutilBusinessModel.A256CBC_HS512_ECDHESA256KW: {enc: &EncA256CBC_HS512, alg: &AlgECDHESA256KW},
		cryptoutilBusinessModel.A192CBC_HS384_ECDHESA256KW: {enc: &EncA192CBC_HS384, alg: &AlgECDHESA256KW},
		cryptoutilBusinessModel.A128CBC_HS256_ECDHESA256KW: {enc: &EncA128CBC_HS256, alg: &AlgECDHESA256KW},
		cryptoutilBusinessModel.A192CBC_HS384_ECDHESA192KW: {enc: &EncA192CBC_HS384, alg: &AlgECDHESA192KW},
		cryptoutilBusinessModel.A128CBC_HS256_ECDHESA192KW: {enc: &EncA128CBC_HS256, alg: &AlgECDHESA192KW},
		cryptoutilBusinessModel.A128CBC_HS256_ECDHESA128KW: {enc: &EncA128CBC_HS256, alg: &AlgECDHESA128KW},
		cryptoutilBusinessModel.A256CBC_HS512_ECDHES:       {enc: &EncA256CBC_HS512, alg: &AlgECDHES},
		cryptoutilBusinessModel.A192CBC_HS384_ECDHES:       {enc: &EncA192CBC_HS384, alg: &AlgECDHES},
		cryptoutilBusinessModel.A128CBC_HS256_ECDHES:       {enc: &EncA128CBC_HS256, alg: &AlgECDHES},
	}

	elasticKeyAlgorithmToJoseAlg = map[cryptoutilBusinessModel.ElasticKeyAlgorithm]*joseJwa.SignatureAlgorithm{
		cryptoutilBusinessModel.RS512: &AlgRS512,
		cryptoutilBusinessModel.RS384: &AlgRS384,
		cryptoutilBusinessModel.RS256: &AlgRS256,
		cryptoutilBusinessModel.PS512: &AlgPS512,
		cryptoutilBusinessModel.PS384: &AlgPS384,
		cryptoutilBusinessModel.PS256: &AlgPS256,
		cryptoutilBusinessModel.ES512: &AlgES512,
		cryptoutilBusinessModel.ES384: &AlgES384,
		cryptoutilBusinessModel.ES256: &AlgES256,
		cryptoutilBusinessModel.HS512: &AlgHS512,
		cryptoutilBusinessModel.HS384: &AlgHS384,
		cryptoutilBusinessModel.HS256: &AlgHS256,
		cryptoutilBusinessModel.EdDSA: &AlgEdDSA,
	}
)

func ToJweEncAndAlg(elasticKeyAlgorithm *cryptoutilBusinessModel.ElasticKeyAlgorithm) (*joseJwa.ContentEncryptionAlgorithm, *joseJwa.KeyEncryptionAlgorithm, error) {
	if encAndAlg, ok := elasticKeyAlgorithmToJoseEncAndAlg[*elasticKeyAlgorithm]; ok {
		return encAndAlg.enc, encAndAlg.alg, nil
	}
	return nil, nil, fmt.Errorf("unsupported JWE ElasticKeyAlgorithm '%s'", *elasticKeyAlgorithm)
}

func ToJwsAlg(elasticKeyAlgorithm *cryptoutilBusinessModel.ElasticKeyAlgorithm) (*joseJwa.SignatureAlgorithm, error) {
	if alg, ok := elasticKeyAlgorithmToJoseAlg[*elasticKeyAlgorithm]; ok {
		return alg, nil
	}
	return nil, fmt.Errorf("unsupported JWS ElasticKeyAlgorithm '%s'", *elasticKeyAlgorithm)
}

func IsJwe(elasticKeyAlgorithm *cryptoutilBusinessModel.ElasticKeyAlgorithm) bool {
	_, ok := elasticKeyAlgorithmToJoseEncAndAlg[*elasticKeyAlgorithm]
	return ok
}

func IsJws(elasticKeyAlgorithm *cryptoutilBusinessModel.ElasticKeyAlgorithm) bool {
	_, ok := elasticKeyAlgorithmToJoseAlg[*elasticKeyAlgorithm]
	return ok
}
