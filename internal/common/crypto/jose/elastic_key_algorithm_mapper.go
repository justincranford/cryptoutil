package jose

import (
	cryptoutilOpenapiModel "cryptoutil/internal/openapi/model"

	"fmt"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
)

var (
	elasticKeyAlgorithmToJoseEncAndAlg = map[cryptoutilOpenapiModel.ElasticKeyAlgorithm]struct {
		enc *joseJwa.ContentEncryptionAlgorithm
		alg *joseJwa.KeyEncryptionAlgorithm
	}{
		cryptoutilOpenapiModel.A256GCMA256KW:    {enc: &EncA256GCM, alg: &AlgA256KW},
		cryptoutilOpenapiModel.A192GCMA256KW:    {enc: &EncA192GCM, alg: &AlgA256KW},
		cryptoutilOpenapiModel.A128GCMA256KW:    {enc: &EncA128GCM, alg: &AlgA256KW},
		cryptoutilOpenapiModel.A256GCMA192KW:    {enc: &EncA256GCM, alg: &AlgA192KW},
		cryptoutilOpenapiModel.A192GCMA192KW:    {enc: &EncA192GCM, alg: &AlgA192KW},
		cryptoutilOpenapiModel.A128GCMA192KW:    {enc: &EncA128GCM, alg: &AlgA192KW},
		cryptoutilOpenapiModel.A256GCMA128KW:    {enc: &EncA256GCM, alg: &AlgA128KW},
		cryptoutilOpenapiModel.A192GCMA128KW:    {enc: &EncA192GCM, alg: &AlgA128KW},
		cryptoutilOpenapiModel.A128GCMA128KW:    {enc: &EncA128GCM, alg: &AlgA128KW},
		cryptoutilOpenapiModel.A256GCMA256GCMKW: {enc: &EncA256GCM, alg: &AlgA256GCMKW},
		cryptoutilOpenapiModel.A192GCMA256GCMKW: {enc: &EncA192GCM, alg: &AlgA256GCMKW},
		cryptoutilOpenapiModel.A128GCMA256GCMKW: {enc: &EncA128GCM, alg: &AlgA256GCMKW},
		cryptoutilOpenapiModel.A256GCMA192GCMKW: {enc: &EncA256GCM, alg: &AlgA192GCMKW},
		cryptoutilOpenapiModel.A192GCMA192GCMKW: {enc: &EncA192GCM, alg: &AlgA192GCMKW},
		cryptoutilOpenapiModel.A128GCMA192GCMKW: {enc: &EncA128GCM, alg: &AlgA192GCMKW},
		cryptoutilOpenapiModel.A256GCMA128GCMKW: {enc: &EncA256GCM, alg: &AlgA128GCMKW},
		cryptoutilOpenapiModel.A192GCMA128GCMKW: {enc: &EncA192GCM, alg: &AlgA128GCMKW},
		cryptoutilOpenapiModel.A128GCMA128GCMKW: {enc: &EncA128GCM, alg: &AlgA128GCMKW},
		cryptoutilOpenapiModel.A256GCMDir:       {enc: &EncA256GCM, alg: &AlgDir},
		cryptoutilOpenapiModel.A192GCMDir:       {enc: &EncA192GCM, alg: &AlgDir},
		cryptoutilOpenapiModel.A128GCMDir:       {enc: &EncA128GCM, alg: &AlgDir},

		cryptoutilOpenapiModel.A256GCMRSAOAEP512: {enc: &EncA256GCM, alg: &AlgRSAOAEP512},
		cryptoutilOpenapiModel.A192GCMRSAOAEP512: {enc: &EncA192GCM, alg: &AlgRSAOAEP512},
		cryptoutilOpenapiModel.A128GCMRSAOAEP512: {enc: &EncA128GCM, alg: &AlgRSAOAEP512},
		cryptoutilOpenapiModel.A256GCMRSAOAEP384: {enc: &EncA256GCM, alg: &AlgRSAOAEP384},
		cryptoutilOpenapiModel.A192GCMRSAOAEP384: {enc: &EncA192GCM, alg: &AlgRSAOAEP384},
		cryptoutilOpenapiModel.A128GCMRSAOAEP384: {enc: &EncA128GCM, alg: &AlgRSAOAEP384},
		cryptoutilOpenapiModel.A256GCMRSAOAEP256: {enc: &EncA256GCM, alg: &AlgRSAOAEP256},
		cryptoutilOpenapiModel.A192GCMRSAOAEP256: {enc: &EncA192GCM, alg: &AlgRSAOAEP256},
		cryptoutilOpenapiModel.A128GCMRSAOAEP256: {enc: &EncA128GCM, alg: &AlgRSAOAEP256},
		cryptoutilOpenapiModel.A256GCMRSAOAEP:    {enc: &EncA256GCM, alg: &AlgRSAOAEP},
		cryptoutilOpenapiModel.A192GCMRSAOAEP:    {enc: &EncA192GCM, alg: &AlgRSAOAEP},
		cryptoutilOpenapiModel.A128GCMRSAOAEP:    {enc: &EncA128GCM, alg: &AlgRSAOAEP},
		cryptoutilOpenapiModel.A256GCMRSA15:      {enc: &EncA256GCM, alg: &AlgRSA15},
		cryptoutilOpenapiModel.A192GCMRSA15:      {enc: &EncA192GCM, alg: &AlgRSA15},
		cryptoutilOpenapiModel.A128GCMRSA15:      {enc: &EncA128GCM, alg: &AlgRSA15},

		cryptoutilOpenapiModel.A256GCMECDHESA256KW: {enc: &EncA256GCM, alg: &AlgECDHESA256KW},
		cryptoutilOpenapiModel.A192GCMECDHESA256KW: {enc: &EncA192GCM, alg: &AlgECDHESA256KW},
		cryptoutilOpenapiModel.A128GCMECDHESA256KW: {enc: &EncA128GCM, alg: &AlgECDHESA256KW},
		cryptoutilOpenapiModel.A256GCMECDHESA192KW: {enc: &EncA256GCM, alg: &AlgECDHESA192KW},
		cryptoutilOpenapiModel.A192GCMECDHESA192KW: {enc: &EncA192GCM, alg: &AlgECDHESA192KW},
		cryptoutilOpenapiModel.A128GCMECDHESA192KW: {enc: &EncA128GCM, alg: &AlgECDHESA192KW},
		cryptoutilOpenapiModel.A256GCMECDHESA128KW: {enc: &EncA256GCM, alg: &AlgECDHESA128KW},
		cryptoutilOpenapiModel.A192GCMECDHESA128KW: {enc: &EncA192GCM, alg: &AlgECDHESA128KW},
		cryptoutilOpenapiModel.A128GCMECDHESA128KW: {enc: &EncA128GCM, alg: &AlgECDHESA128KW},
		cryptoutilOpenapiModel.A256GCMECDHES:       {enc: &EncA256GCM, alg: &AlgECDHES},
		cryptoutilOpenapiModel.A192GCMECDHES:       {enc: &EncA192GCM, alg: &AlgECDHES},
		cryptoutilOpenapiModel.A128GCMECDHES:       {enc: &EncA128GCM, alg: &AlgECDHES},

		cryptoutilOpenapiModel.A256CBCHS512A256KW:    {enc: &EncA256CBCHS512, alg: &AlgA256KW},
		cryptoutilOpenapiModel.A192CBCHS384A256KW:    {enc: &EncA192CBCHS384, alg: &AlgA256KW},
		cryptoutilOpenapiModel.A128CBCHS256A256KW:    {enc: &EncA128CBCHS256, alg: &AlgA256KW},
		cryptoutilOpenapiModel.A256CBCHS512A192KW:    {enc: &EncA256CBCHS512, alg: &AlgA192KW},
		cryptoutilOpenapiModel.A192CBCHS384A192KW:    {enc: &EncA192CBCHS384, alg: &AlgA192KW},
		cryptoutilOpenapiModel.A128CBCHS256A192KW:    {enc: &EncA128CBCHS256, alg: &AlgA192KW},
		cryptoutilOpenapiModel.A256CBCHS512A128KW:    {enc: &EncA256CBCHS512, alg: &AlgA128KW},
		cryptoutilOpenapiModel.A192CBCHS384A128KW:    {enc: &EncA192CBCHS384, alg: &AlgA128KW},
		cryptoutilOpenapiModel.A128CBCHS256A128KW:    {enc: &EncA128CBCHS256, alg: &AlgA128KW},
		cryptoutilOpenapiModel.A256CBCHS512A256GCMKW: {enc: &EncA256CBCHS512, alg: &AlgA256GCMKW},
		cryptoutilOpenapiModel.A192CBCHS384A256GCMKW: {enc: &EncA192CBCHS384, alg: &AlgA256GCMKW},
		cryptoutilOpenapiModel.A128CBCHS256A256GCMKW: {enc: &EncA128CBCHS256, alg: &AlgA256GCMKW},
		cryptoutilOpenapiModel.A256CBCHS512A192GCMKW: {enc: &EncA256CBCHS512, alg: &AlgA192GCMKW},
		cryptoutilOpenapiModel.A192CBCHS384A192GCMKW: {enc: &EncA192CBCHS384, alg: &AlgA192GCMKW},
		cryptoutilOpenapiModel.A128CBCHS256A192GCMKW: {enc: &EncA128CBCHS256, alg: &AlgA192GCMKW},
		cryptoutilOpenapiModel.A256CBCHS512A128GCMKW: {enc: &EncA256CBCHS512, alg: &AlgA128GCMKW},
		cryptoutilOpenapiModel.A192CBCHS384A128GCMKW: {enc: &EncA192CBCHS384, alg: &AlgA128GCMKW},
		cryptoutilOpenapiModel.A128CBCHS256A128GCMKW: {enc: &EncA128CBCHS256, alg: &AlgA128GCMKW},
		cryptoutilOpenapiModel.A256CBCHS512Dir:       {enc: &EncA256CBCHS512, alg: &AlgDir},
		cryptoutilOpenapiModel.A192CBCHS384Dir:       {enc: &EncA192CBCHS384, alg: &AlgDir},
		cryptoutilOpenapiModel.A128CBCHS256Dir:       {enc: &EncA128CBCHS256, alg: &AlgDir},

		cryptoutilOpenapiModel.A256CBCHS512RSAOAEP512: {enc: &EncA256CBCHS512, alg: &AlgRSAOAEP512},
		cryptoutilOpenapiModel.A192CBCHS384RSAOAEP512: {enc: &EncA192CBCHS384, alg: &AlgRSAOAEP512},
		cryptoutilOpenapiModel.A128CBCHS256RSAOAEP512: {enc: &EncA128CBCHS256, alg: &AlgRSAOAEP512},
		cryptoutilOpenapiModel.A256CBCHS512RSAOAEP384: {enc: &EncA256CBCHS512, alg: &AlgRSAOAEP384},
		cryptoutilOpenapiModel.A192CBCHS384RSAOAEP384: {enc: &EncA192CBCHS384, alg: &AlgRSAOAEP384},
		cryptoutilOpenapiModel.A128CBCHS256RSAOAEP384: {enc: &EncA128CBCHS256, alg: &AlgRSAOAEP384},
		cryptoutilOpenapiModel.A256CBCHS512RSAOAEP256: {enc: &EncA256CBCHS512, alg: &AlgRSAOAEP256},
		cryptoutilOpenapiModel.A192CBCHS384RSAOAEP256: {enc: &EncA192CBCHS384, alg: &AlgRSAOAEP256},
		cryptoutilOpenapiModel.A128CBCHS256RSAOAEP256: {enc: &EncA128CBCHS256, alg: &AlgRSAOAEP256},
		cryptoutilOpenapiModel.A256CBCHS512RSAOAEP:    {enc: &EncA256CBCHS512, alg: &AlgRSAOAEP},
		cryptoutilOpenapiModel.A192CBCHS384RSAOAEP:    {enc: &EncA192CBCHS384, alg: &AlgRSAOAEP},
		cryptoutilOpenapiModel.A128CBCHS256RSAOAEP:    {enc: &EncA128CBCHS256, alg: &AlgRSAOAEP},
		cryptoutilOpenapiModel.A256CBCHS512RSA15:      {enc: &EncA256CBCHS512, alg: &AlgRSA15},
		cryptoutilOpenapiModel.A192CBCHS384RSA15:      {enc: &EncA192CBCHS384, alg: &AlgRSA15},
		cryptoutilOpenapiModel.A128CBCHS256RSA15:      {enc: &EncA128CBCHS256, alg: &AlgRSA15},

		cryptoutilOpenapiModel.A256CBCHS512ECDHESA256KW: {enc: &EncA256CBCHS512, alg: &AlgECDHESA256KW},
		cryptoutilOpenapiModel.A192CBCHS384ECDHESA256KW: {enc: &EncA192CBCHS384, alg: &AlgECDHESA256KW},
		cryptoutilOpenapiModel.A128CBCHS256ECDHESA256KW: {enc: &EncA128CBCHS256, alg: &AlgECDHESA256KW},
		cryptoutilOpenapiModel.A192CBCHS384ECDHESA192KW: {enc: &EncA192CBCHS384, alg: &AlgECDHESA192KW},
		cryptoutilOpenapiModel.A128CBCHS256ECDHESA192KW: {enc: &EncA128CBCHS256, alg: &AlgECDHESA192KW},
		cryptoutilOpenapiModel.A128CBCHS256ECDHESA128KW: {enc: &EncA128CBCHS256, alg: &AlgECDHESA128KW},
		cryptoutilOpenapiModel.A256CBCHS512ECDHES:       {enc: &EncA256CBCHS512, alg: &AlgECDHES},
		cryptoutilOpenapiModel.A192CBCHS384ECDHES:       {enc: &EncA192CBCHS384, alg: &AlgECDHES},
		cryptoutilOpenapiModel.A128CBCHS256ECDHES:       {enc: &EncA128CBCHS256, alg: &AlgECDHES},
	}

	elasticKeyAlgorithmToJoseAlg = map[cryptoutilOpenapiModel.ElasticKeyAlgorithm]*joseJwa.SignatureAlgorithm{
		cryptoutilOpenapiModel.RS512: &AlgRS512,
		cryptoutilOpenapiModel.RS384: &AlgRS384,
		cryptoutilOpenapiModel.RS256: &AlgRS256,
		cryptoutilOpenapiModel.PS512: &AlgPS512,
		cryptoutilOpenapiModel.PS384: &AlgPS384,
		cryptoutilOpenapiModel.PS256: &AlgPS256,
		cryptoutilOpenapiModel.ES512: &AlgES512,
		cryptoutilOpenapiModel.ES384: &AlgES384,
		cryptoutilOpenapiModel.ES256: &AlgES256,
		cryptoutilOpenapiModel.HS512: &AlgHS512,
		cryptoutilOpenapiModel.HS384: &AlgHS384,
		cryptoutilOpenapiModel.HS256: &AlgHS256,
		cryptoutilOpenapiModel.EdDSA: &AlgEdDSA,
	}
)

func ToJweEncAndAlg(elasticKeyAlgorithm *cryptoutilOpenapiModel.ElasticKeyAlgorithm) (*joseJwa.ContentEncryptionAlgorithm, *joseJwa.KeyEncryptionAlgorithm, error) {
	if encAndAlg, ok := elasticKeyAlgorithmToJoseEncAndAlg[*elasticKeyAlgorithm]; ok {
		return encAndAlg.enc, encAndAlg.alg, nil
	}
	return nil, nil, fmt.Errorf("unsupported JWE ElasticKeyAlgorithm '%s'", *elasticKeyAlgorithm)
}

func ToJwsAlg(elasticKeyAlgorithm *cryptoutilOpenapiModel.ElasticKeyAlgorithm) (*joseJwa.SignatureAlgorithm, error) {
	if alg, ok := elasticKeyAlgorithmToJoseAlg[*elasticKeyAlgorithm]; ok {
		return alg, nil
	}
	return nil, fmt.Errorf("unsupported JWS ElasticKeyAlgorithm '%s'", *elasticKeyAlgorithm)
}

func IsJwe(elasticKeyAlgorithm *cryptoutilOpenapiModel.ElasticKeyAlgorithm) bool {
	_, ok := elasticKeyAlgorithmToJoseEncAndAlg[*elasticKeyAlgorithm]
	return ok
}

func IsJws(elasticKeyAlgorithm *cryptoutilOpenapiModel.ElasticKeyAlgorithm) bool {
	_, ok := elasticKeyAlgorithmToJoseAlg[*elasticKeyAlgorithm]
	return ok
}
