package jose

import (
	"encoding/json"
	"fmt"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"

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
		alg, err := ExtractJwsJwkAlg(&jwk, i)
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
		alg, err := ExtractJwsJwkAlg(&jwk, i)
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

func LogJwsInfo(jwsMessage *joseJws.Message) error {
	if jwsMessage == nil {
		return fmt.Errorf("jwsMessage is nil")
	} else if len(jwsMessage.Signatures()) == 0 {
		return fmt.Errorf("jwsMessage has no signatures")
	}

	for i, jwsSignature := range jwsMessage.Signatures() {
		logMessageSigHeader := fmt.Sprintf("JWS Header[%d]:", i)

		protectedHeaders := jwsSignature.ProtectedHeaders()
		if alg, ok := protectedHeaders.Algorithm(); ok {
			logMessageSigHeader += fmt.Sprintf(" alg=%s\n", alg)
		}
		if kid, ok := protectedHeaders.KeyID(); ok {
			logMessageSigHeader += fmt.Sprintf(" kid=%s\n", kid)
		}
		if typ, ok := protectedHeaders.Type(); ok {
			logMessageSigHeader += fmt.Sprintf(" typ=%s\n", typ)
		}
		if cty, ok := protectedHeaders.ContentType(); ok {
			logMessageSigHeader += fmt.Sprintf(" cty=%s\n", cty)
		}
		if jku, ok := protectedHeaders.JWKSetURL(); ok {
			logMessageSigHeader += fmt.Sprintf(" jku=%s\n", jku)
		}
		if x5u, ok := protectedHeaders.X509URL(); ok {
			logMessageSigHeader += fmt.Sprintf(" x5u=%s\n", x5u)
		}
		if x5c, ok := protectedHeaders.X509CertChain(); ok {
			logMessageSigHeader += fmt.Sprintf(" x5c=%v\n", x5c)
		}
		if x5t, ok := protectedHeaders.X509CertThumbprint(); ok {
			logMessageSigHeader += fmt.Sprintf(" x5t=%s\n", x5t)
		}
		if x5tS256, ok := protectedHeaders.X509CertThumbprintS256(); ok {
			logMessageSigHeader += fmt.Sprintf(" x5t#S256=%s\n", x5tS256)
		}
		if crit, ok := protectedHeaders.Critical(); ok {
			logMessageSigHeader += fmt.Sprintf(" crit=%v\n", crit)
		}

		publicHeaders := jwsSignature.PublicHeaders()
		for _, key := range publicHeaders.Keys() {
			var value any
			err := publicHeaders.Get(key, &value)
			if err != nil {
				logMessageSigHeader += fmt.Sprintf(" %s=%v\n", key, value)
			}
		}

		fmt.Print(logMessageSigHeader)
	}

	fmt.Printf("JWS Payload: %s\n", string(jwsMessage.Payload()))

	for i, jwsSignature := range jwsMessage.Signatures() {
		fmt.Printf("JWS Signature[%d]: %s\n", i, jwsSignature.Signature())
	}

	return nil
}
