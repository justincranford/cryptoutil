// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	"fmt"

	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
)

func ToJWEEncAndAlg(elasticKeyAlgorithm *cryptoutilOpenapiModel.ElasticKeyAlgorithm) (*joseJwa.ContentEncryptionAlgorithm, *joseJwa.KeyEncryptionAlgorithm, error) {
	if encAndAlg, ok := elasticKeyAlgorithmToJoseEncAndAlg[*elasticKeyAlgorithm]; ok {
		return encAndAlg.enc, encAndAlg.alg, nil
	}

	return nil, nil, fmt.Errorf("unsupported JWE ElasticKeyAlgorithm '%s'", *elasticKeyAlgorithm)
}

// ToJWSAlg converts an ElasticKeyAlgorithm to a JWS signature algorithm.
func ToJWSAlg(elasticKeyAlgorithm *cryptoutilOpenapiModel.ElasticKeyAlgorithm) (*joseJwa.SignatureAlgorithm, error) {
	if alg, ok := elasticKeyAlgorithmToJoseAlg[*elasticKeyAlgorithm]; ok {
		return alg, nil
	}

	return nil, fmt.Errorf("unsupported JWS ElasticKeyAlgorithm '%s'", *elasticKeyAlgorithm)
}

// IsJWE returns true if the algorithm is a JWE encryption algorithm.
func IsJWE(elasticKeyAlgorithm *cryptoutilOpenapiModel.ElasticKeyAlgorithm) bool {
	_, ok := elasticKeyAlgorithmToJoseEncAndAlg[*elasticKeyAlgorithm]

	return ok
}

// IsJWS returns true if the algorithm is a JWS signature algorithm.
func IsJWS(elasticKeyAlgorithm *cryptoutilOpenapiModel.ElasticKeyAlgorithm) bool {
	_, ok := elasticKeyAlgorithmToJoseAlg[*elasticKeyAlgorithm]

	return ok
}

// IsSymmetric returns true if the algorithm uses symmetric keys.
func IsSymmetric(elasticKeyAlgorithm *cryptoutilOpenapiModel.ElasticKeyAlgorithm) (bool, error) {
	isSymmetric, ok := symmetricElasticKeyAlgorithm[*elasticKeyAlgorithm]
	if ok {
		return isSymmetric, nil
	}

	return false, fmt.Errorf("unsupported ElasticKeyAlgorithm '%s'", *elasticKeyAlgorithm)
}

// IsAsymmetric returns true if the algorithm uses asymmetric keys.
func IsAsymmetric(elasticKeyAlgorithm *cryptoutilOpenapiModel.ElasticKeyAlgorithm) (bool, error) {
	isAsymmetric, ok := asymmetricElasticKeyAlgorithm[*elasticKeyAlgorithm]
	if ok {
		return isAsymmetric, nil
	}

	return false, fmt.Errorf("unsupported ElasticKeyAlgorithm '%s'", *elasticKeyAlgorithm)
}

// ToElasticKeyAlgorithm converts a string to an ElasticKeyAlgorithm.
func ToElasticKeyAlgorithm(algorithm *string) (*cryptoutilOpenapiModel.ElasticKeyAlgorithm, error) {
	if alg, exists := elasticKeyAlgorithms[*algorithm]; exists {
		return &alg, nil
	}

	return nil, fmt.Errorf("invalid elastic Key algorithm: %v", algorithm)
}

// ToGenerateAlgorithm converts a string to a GenerateAlgorithm.
func ToGenerateAlgorithm(algorithm *string) (*cryptoutilOpenapiModel.GenerateAlgorithm, error) {
	if alg, exists := generateAlgorithms[*algorithm]; exists {
		return &alg, nil
	}

	return nil, fmt.Errorf("invalid generate algorithm: %v", algorithm)
}

// GetGenerateAlgorithmTestProbability returns the execution probability for table-driven tests.
// Different key sizes of the same algorithm type can use lower probabilities to reduce test time.
// Base algorithms (e.g., RSA2048, ECP256, Oct256) use TestProbAlways for comprehensive coverage.
// Larger variants (e.g., RSA4096, ECP521, Oct512) use TestProbThird for sampling coverage.
func GetGenerateAlgorithmTestProbability(alg cryptoutilOpenapiModel.GenerateAlgorithm) float64 {
	switch alg {
	// Base RSA size - always test.
	case cryptoutilOpenapiModel.RSA2048:
		return cryptoutilSharedMagic.TestProbAlways
	// Larger RSA sizes - sample testing.
	case cryptoutilOpenapiModel.RSA3072, cryptoutilOpenapiModel.RSA4096:
		return cryptoutilSharedMagic.TestProbThird
	// Base EC size - always test.
	case cryptoutilOpenapiModel.ECP256:
		return cryptoutilSharedMagic.TestProbAlways
	// Larger EC sizes - sample testing.
	case cryptoutilOpenapiModel.ECP384, cryptoutilOpenapiModel.ECP521:
		return cryptoutilSharedMagic.TestProbThird
	// EdDSA - always test (only one size).
	case cryptoutilOpenapiModel.OKPEd25519:
		return cryptoutilSharedMagic.TestProbAlways
	// Base symmetric key size - always test.
	case cryptoutilOpenapiModel.Oct256:
		return cryptoutilSharedMagic.TestProbAlways
	// Other symmetric sizes - sample testing.
	case cryptoutilOpenapiModel.Oct128, cryptoutilOpenapiModel.Oct192, cryptoutilOpenapiModel.Oct384, cryptoutilOpenapiModel.Oct512:
		return cryptoutilSharedMagic.TestProbThird
	default:
		return cryptoutilSharedMagic.TestProbAlways
	}
}

// GetElasticKeyAlgorithmTestProbability returns the execution probability for table-driven tests.
// Encryption/signing algorithms with multiple key sizes use lower probabilities for variants.
// Base algorithms (e.g., A256GCM, RS256, ES256) use TestProbAlways for comprehensive coverage.
// Variants with different key sizes use TestProbQuarter for sampling coverage.
func GetElasticKeyAlgorithmTestProbability(alg cryptoutilOpenapiModel.ElasticKeyAlgorithm) float64 {
	switch alg {
	// Base AES-GCM + Key Wrap combinations - always test 256-bit.
	case cryptoutilOpenapiModel.A256GCMA256KW, cryptoutilOpenapiModel.A256GCMA256GCMKW, cryptoutilOpenapiModel.A256GCMDir:
		return cryptoutilSharedMagic.TestProbAlways
	// Other AES-GCM + Key Wrap - sample testing.
	case cryptoutilOpenapiModel.A192GCMA256KW, cryptoutilOpenapiModel.A128GCMA256KW,
		cryptoutilOpenapiModel.A256GCMA192KW, cryptoutilOpenapiModel.A192GCMA192KW, cryptoutilOpenapiModel.A128GCMA192KW,
		cryptoutilOpenapiModel.A256GCMA128KW, cryptoutilOpenapiModel.A192GCMA128KW, cryptoutilOpenapiModel.A128GCMA128KW,
		cryptoutilOpenapiModel.A192GCMA256GCMKW, cryptoutilOpenapiModel.A128GCMA256GCMKW,
		cryptoutilOpenapiModel.A256GCMA192GCMKW, cryptoutilOpenapiModel.A192GCMA192GCMKW, cryptoutilOpenapiModel.A128GCMA192GCMKW,
		cryptoutilOpenapiModel.A256GCMA128GCMKW, cryptoutilOpenapiModel.A192GCMA128GCMKW, cryptoutilOpenapiModel.A128GCMA128GCMKW,
		cryptoutilOpenapiModel.A192GCMDir, cryptoutilOpenapiModel.A128GCMDir:
		return cryptoutilSharedMagic.TestProbQuarter
	// Base RSA OAEP - always test.
	case cryptoutilOpenapiModel.A256GCMRSAOAEP256, cryptoutilOpenapiModel.A256CBCHS512RSAOAEP256:
		return cryptoutilSharedMagic.TestProbAlways
	// Other RSA OAEP variants - sample testing.
	case cryptoutilOpenapiModel.A192GCMRSAOAEP512, cryptoutilOpenapiModel.A128GCMRSAOAEP512,
		cryptoutilOpenapiModel.A256GCMRSAOAEP384, cryptoutilOpenapiModel.A192GCMRSAOAEP384, cryptoutilOpenapiModel.A128GCMRSAOAEP384,
		cryptoutilOpenapiModel.A192GCMRSAOAEP256, cryptoutilOpenapiModel.A128GCMRSAOAEP256,
		cryptoutilOpenapiModel.A256GCMRSAOAEP, cryptoutilOpenapiModel.A192GCMRSAOAEP, cryptoutilOpenapiModel.A128GCMRSAOAEP,
		cryptoutilOpenapiModel.A256GCMRSA15, cryptoutilOpenapiModel.A192GCMRSA15, cryptoutilOpenapiModel.A128GCMRSA15,
		cryptoutilOpenapiModel.A192CBCHS384RSAOAEP512, cryptoutilOpenapiModel.A128CBCHS256RSAOAEP512,
		cryptoutilOpenapiModel.A256CBCHS512RSAOAEP384, cryptoutilOpenapiModel.A192CBCHS384RSAOAEP384, cryptoutilOpenapiModel.A128CBCHS256RSAOAEP384,
		cryptoutilOpenapiModel.A192CBCHS384RSAOAEP256, cryptoutilOpenapiModel.A128CBCHS256RSAOAEP256,
		cryptoutilOpenapiModel.A256CBCHS512RSAOAEP, cryptoutilOpenapiModel.A192CBCHS384RSAOAEP, cryptoutilOpenapiModel.A128CBCHS256RSAOAEP,
		cryptoutilOpenapiModel.A256CBCHS512RSA15, cryptoutilOpenapiModel.A192CBCHS384RSA15, cryptoutilOpenapiModel.A128CBCHS256RSA15:
		return cryptoutilSharedMagic.TestProbQuarter
	// Base ECDH-ES - always test.
	case cryptoutilOpenapiModel.A256GCMECDHESA256KW, cryptoutilOpenapiModel.A256CBCHS512ECDHESA256KW:
		return cryptoutilSharedMagic.TestProbAlways
	// Other ECDH-ES variants - sample testing.
	case cryptoutilOpenapiModel.A192GCMECDHESA256KW, cryptoutilOpenapiModel.A128GCMECDHESA256KW,
		cryptoutilOpenapiModel.A256GCMECDHESA192KW, cryptoutilOpenapiModel.A192GCMECDHESA192KW, cryptoutilOpenapiModel.A128GCMECDHESA192KW,
		cryptoutilOpenapiModel.A256GCMECDHESA128KW, cryptoutilOpenapiModel.A192GCMECDHESA128KW, cryptoutilOpenapiModel.A128GCMECDHESA128KW,
		cryptoutilOpenapiModel.A256GCMECDHES, cryptoutilOpenapiModel.A192GCMECDHES, cryptoutilOpenapiModel.A128GCMECDHES,
		cryptoutilOpenapiModel.A192CBCHS384ECDHESA256KW, cryptoutilOpenapiModel.A128CBCHS256ECDHESA256KW,
		cryptoutilOpenapiModel.A256CBCHS512ECDHESA192KW, cryptoutilOpenapiModel.A192CBCHS384ECDHESA192KW, cryptoutilOpenapiModel.A128CBCHS256ECDHESA192KW,
		cryptoutilOpenapiModel.A128CBCHS256ECDHESA128KW,
		cryptoutilOpenapiModel.A256CBCHS512ECDHES, cryptoutilOpenapiModel.A192CBCHS384ECDHES, cryptoutilOpenapiModel.A128CBCHS256ECDHES:
		return cryptoutilSharedMagic.TestProbQuarter
	// Base AES-CBC-HMAC + Key Wrap - always test 256-bit.
	case cryptoutilOpenapiModel.A256CBCHS512A256KW, cryptoutilOpenapiModel.A256CBCHS512A256GCMKW, cryptoutilOpenapiModel.A256CBCHS512Dir:
		return cryptoutilSharedMagic.TestProbAlways
	// Other AES-CBC-HMAC + Key Wrap - sample testing.
	case cryptoutilOpenapiModel.A192CBCHS384A256KW, cryptoutilOpenapiModel.A128CBCHS256A256KW,
		cryptoutilOpenapiModel.A256CBCHS512A192KW, cryptoutilOpenapiModel.A192CBCHS384A192KW, cryptoutilOpenapiModel.A128CBCHS256A192KW,
		cryptoutilOpenapiModel.A256CBCHS512A128KW, cryptoutilOpenapiModel.A192CBCHS384A128KW, cryptoutilOpenapiModel.A128CBCHS256A128KW,
		cryptoutilOpenapiModel.A192CBCHS384A256GCMKW, cryptoutilOpenapiModel.A128CBCHS256A256GCMKW,
		cryptoutilOpenapiModel.A256CBCHS512A192GCMKW, cryptoutilOpenapiModel.A192CBCHS384A192GCMKW, cryptoutilOpenapiModel.A128CBCHS256A192GCMKW,
		cryptoutilOpenapiModel.A256CBCHS512A128GCMKW, cryptoutilOpenapiModel.A192CBCHS384A128GCMKW, cryptoutilOpenapiModel.A128CBCHS256A128GCMKW,
		cryptoutilOpenapiModel.A192CBCHS384Dir, cryptoutilOpenapiModel.A128CBCHS256Dir:
		return cryptoutilSharedMagic.TestProbQuarter
	// Base signature algorithms - always test.
	case cryptoutilOpenapiModel.RS256, cryptoutilOpenapiModel.PS256, cryptoutilOpenapiModel.ES256, cryptoutilOpenapiModel.HS256, cryptoutilOpenapiModel.EdDSA:
		return cryptoutilSharedMagic.TestProbAlways
	// Other signature algorithm sizes - sample testing.
	case cryptoutilOpenapiModel.RS384, cryptoutilOpenapiModel.RS512,
		cryptoutilOpenapiModel.PS384, cryptoutilOpenapiModel.PS512,
		cryptoutilOpenapiModel.ES384, cryptoutilOpenapiModel.ES512,
		cryptoutilOpenapiModel.HS384, cryptoutilOpenapiModel.HS512:
		return cryptoutilSharedMagic.TestProbThird
	default:
		return cryptoutilSharedMagic.TestProbAlways
	}
}
