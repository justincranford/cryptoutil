package businessmodel

import (
	cryptoutilOpenapiModel "cryptoutil/internal/openapi/model"

	"fmt"
)

func ToElasticKeyAlgorithm(algorithm string) (*cryptoutilOpenapiModel.ElasticKeyAlgorithm, error) {
	if alg, exists := elasticKeyAlgorithms[algorithm]; exists {
		return &alg, nil
	}
	return nil, fmt.Errorf("invalid elastic Key algorithm: %s", algorithm)
}

func IsSymmetric(elasticKeyAlgorithm *cryptoutilOpenapiModel.ElasticKeyAlgorithm) (bool, error) {
	isSymmetric, ok := symmetricElasticKeyAlgorithm[*elasticKeyAlgorithm]
	if ok {
		return isSymmetric, nil
	}
	return false, fmt.Errorf("unsupported ElasticKeyAlgorithm '%s'", *elasticKeyAlgorithm)
}

func IsAsymmetric(elasticKeyAlgorithm *cryptoutilOpenapiModel.ElasticKeyAlgorithm) (bool, error) {
	isAsymmetric, ok := asymmetricElasticKeyAlgorithm[*elasticKeyAlgorithm]
	if ok {
		return isAsymmetric, nil
	}
	return false, fmt.Errorf("unsupported ElasticKeyAlgorithm '%s'", *elasticKeyAlgorithm)
}
