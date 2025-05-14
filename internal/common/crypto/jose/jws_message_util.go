package jose

import (
	"encoding/json"
	"fmt"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	joseJws "github.com/lestrrat-go/jwx/v3/jws"
)

func SignBytes(jwks []joseJwk.Key, clearBytes []byte) (*joseJws.Message, []byte, error) {
	if jwks == nil {
		return nil, nil, fmt.Errorf("invalid JWKs: %w", cryptoutilAppErr.ErrCantBeNil)
	} else if len(jwks) == 0 {
		return nil, nil, fmt.Errorf("invalid JWKs: %w", cryptoutilAppErr.ErrCantBeEmpty)
	} else if clearBytes == nil {
		return nil, nil, fmt.Errorf("invalid clearBytes: %w", cryptoutilAppErr.ErrCantBeNil)
	} else if len(clearBytes) == 0 {
		return nil, nil, fmt.Errorf("invalid clearBytes: %w", cryptoutilAppErr.ErrCantBeEmpty)
	}

	jwsSignOptions := make([]joseJws.SignOption, 0, len(jwks))
	if len(jwks) > 1 {
		jwsSignOptions = append(jwsSignOptions, joseJws.WithJSON()) // if more than one JWK, must use JSON encoding instead of default Compact encoding
	}
	for i, jwk := range jwks {
		alg, err := jwsJwkAlg(&jwk, i)
		if err != nil {
			return nil, nil, fmt.Errorf("JWK %d invalid: %w", i, err)
		}
		jwsSignOptions = append(jwsSignOptions, joseJws.WithKey(*alg, jwk)) // add ALG+JWK tuple for each JWK
	}

	jwsMessageBytes, err := joseJws.Sign(clearBytes, jwsSignOptions...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt clearBytes: %w", err)
	}

	jwsMessage, err := joseJws.Parse(jwsMessageBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse JWS message bytes: %w", err)
	}

	return jwsMessage, jwsMessageBytes, nil
}

func VerifyBytes(jwks []joseJwk.Key, jwsMessageBytes []byte) ([]byte, error) {
	if jwks == nil {
		return nil, fmt.Errorf("invalid JWKs: %w", cryptoutilAppErr.ErrCantBeNil)
	} else if len(jwks) == 0 {
		return nil, fmt.Errorf("invalid JWKs: %w", cryptoutilAppErr.ErrCantBeEmpty)
	} else if jwsMessageBytes == nil {
		return nil, fmt.Errorf("invalid jwsMessageBytes: %w", cryptoutilAppErr.ErrCantBeNil)
	} else if len(jwsMessageBytes) == 0 {
		return nil, fmt.Errorf("invalid jwsMessageBytes: %w", cryptoutilAppErr.ErrCantBeEmpty)
	}

	jwsMessage, err := joseJws.Parse(jwsMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWS message bytes: %w", err)
	}

	jwsVerifyOptions := make([]joseJws.VerifyOption, 0, len(jwks))
	for i, jwk := range jwks {
		alg, err := jwsJwkAlg(&jwk, i)
		if err != nil {
			return nil, fmt.Errorf("JWK %d invalid: %w", i, err)
		}
		jwsVerifyOptions = append(jwsVerifyOptions, joseJws.WithKey(*alg, jwk))
	}
	jwsVerifyOptions = append(jwsVerifyOptions, joseJws.WithMessage(jwsMessage))

	decryptedBytes, err := joseJws.Verify(jwsMessageBytes, jwsVerifyOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to verify JWS message bytes: %w", err)
	}

	return decryptedBytes, nil
}

func JwsHeadersString(jwsMessage *joseJws.Message) (string, error) {
	var jwsSignaturesHeadersString string
	for i, jwsMessageSignature := range jwsMessage.Signatures() {
		jwsSignatureHeadersString, err := json.Marshal(jwsMessageSignature.ProtectedHeaders())
		if err != nil {
			return "", fmt.Errorf("failed to marshal JWS headers: %w", err)
		}
		jwsSignaturesHeadersString += string(jwsSignatureHeadersString)
		if i < len(jwsMessage.Signatures())-1 {
			jwsSignaturesHeadersString += "\n"
		}
	}
	return jwsSignaturesHeadersString, nil
}

func jwsJwkAlg(jwk *joseJwk.Key, i int) (*joseJwa.SignatureAlgorithm, error) {
	if jwk == nil {
		return nil, fmt.Errorf("JWK %d invalid: %w", i, cryptoutilAppErr.ErrCantBeNil)
	}

	var alg joseJwa.SignatureAlgorithm
	err := (*jwk).Get(joseJwk.AlgorithmKey, &alg) // Example: RS256, RS384, RS512, ES256, ES384, ES512, PS256, PS384, PS512
	if err != nil {
		return nil, fmt.Errorf("can't get JWK %d 'alg' attribute: %w", i, err)
	}

	return &alg, nil
}
