// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	"encoding/json"
	"fmt"
	"time"

	cryptoutilAppErr "cryptoutil/internal/shared/apperr"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	joseJws "github.com/lestrrat-go/jwx/v3/jws"
)

// SignBytes signs bytes using the provided JWKs and returns a JWS message.
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

	for _, jwk := range jwks {
		isSignJWK, err := IsSignJWK(jwk)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid JWK: %w", err)
		} else if !isSignJWK {
			return nil, nil, fmt.Errorf("invalid JWK: %w", cryptoutilAppErr.ErrJWKMustBeSignJWK)
		}
	}

	algs := make(map[joseJwa.SignatureAlgorithm]struct{})

	jwsSignOptions := make([]joseJws.SignOption, 0, len(jwks))
	if len(jwks) > 1 {
		jwsSignOptions = append(jwsSignOptions, joseJws.WithJSON()) // if more than one JWK, must use JSON encoding instead of default Compact encoding
	}

	iat := time.Now().UTC().Unix()

	for i, jwk := range jwks {
		kid, err := ExtractKidUUID(jwk)
		if err != nil {
			return nil, nil, fmt.Errorf("JWK %d invalid: %w", i, err)
		}

		alg, err := ExtractAlgFromJWSJWK(jwk, i)
		if err != nil {
			return nil, nil, fmt.Errorf("JWK %d invalid: %w", i, err)
		}

		algs[*alg] = struct{}{} // track SignatureAlgorithm counts
		if len(algs) != 1 {     // validate that one-and-only-one SignatureAlgorithm is used across all JWKs
			return nil, nil, fmt.Errorf("can't use JWK %d 'alg' attribute; only one unique 'alg' attribute is allowed", i)
		}

		jwsProtectedHeaders := joseJws.NewHeaders()
		if err := jwsProtectedHeaders.Set(`iat`, iat); err != nil {
			return nil, nil, fmt.Errorf("failed to set iat header: %w", err)
		}

		if err := jwsProtectedHeaders.Set(joseJwk.KeyIDKey, kid.String()); err != nil {
			return nil, nil, fmt.Errorf("failed to set kid header: %w", err)
		}

		if err := jwsProtectedHeaders.Set(joseJwk.AlgorithmKey, *alg); err != nil {
			return nil, nil, fmt.Errorf("failed to set alg header: %w", err)
		}

		jwsSignOptions = append(jwsSignOptions, joseJws.WithKey(*alg, jwk, joseJws.WithProtectedHeaders(jwsProtectedHeaders)))
	}

	jwsMessageBytes, err := joseJws.Sign(clearBytes, jwsSignOptions...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to sign clearBytes: %w", err)
	}

	jwsMessage, err := joseJws.Parse(jwsMessageBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse JWS message bytes: %w", err)
	}

	return jwsMessage, jwsMessageBytes, nil
}

// VerifyBytes verifies a JWS message using the provided JWKs.
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

	for _, jwk := range jwks {
		isVerifyJWK, err := IsVerifyJWK(jwk)
		if err != nil {
			return nil, fmt.Errorf("invalid JWK: %w", err)
		} else if !isVerifyJWK {
			return nil, fmt.Errorf("invalid JWK: %w", cryptoutilAppErr.ErrJWKMustBeVerifyJWK)
		}
	}

	jwsMessage, err := joseJws.Parse(jwsMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWS message bytes: %w", err)
	}

	algs := make(map[joseJwa.SignatureAlgorithm]struct{})
	jwsVerifyOptions := make([]joseJws.VerifyOption, 0, len(jwks))

	for i, jwk := range jwks {
		alg, err := ExtractAlgFromJWSJWK(jwk, i)
		if err != nil {
			return nil, fmt.Errorf("JWK %d invalid: %w", i, err)
		}

		algs[*alg] = struct{}{} // track SignatureAlgorithm counts
		// jwsVerifyOptions = append(jwsVerifyOptions, joseJws.WithKey(*alg, jwk))
		if len(algs) != 1 { // validate that one-and-only-one SignatureAlgorithm is used across all JWKs
			return nil, fmt.Errorf("can't use JWK %d 'alg' attribute; only one unique 'alg' attribute is allowed", i)
		}
	}

	jwkSet := joseJwk.NewSet()
	if err := jwkSet.Set("keys", jwks); err != nil {
		return nil, fmt.Errorf("failed to set keys in JWK set: %w", err)
	}

	jwkSetOptions := []joseJws.WithKeySetSuboption{joseJws.WithRequireKid(true)}
	jwsVerifyOptions = append(jwsVerifyOptions, joseJws.WithKeySet(jwkSet, jwkSetOptions...), joseJws.WithMessage(jwsMessage))

	verifiedBytes, err := joseJws.Verify(jwsMessageBytes, jwsVerifyOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to verify JWS message bytes: %w", err)
	}

	return verifiedBytes, nil
}

// JWSHeadersString returns a string representation of JWS message headers.
func JWSHeadersString(jwsMessage *joseJws.Message) (string, error) {
	if jwsMessage == nil {
		return "", fmt.Errorf("jwsMessage is nil")
	}

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

// ExtractKidAlgFromJWSMessage extracts the key ID and algorithm from a JWS message.
func ExtractKidAlgFromJWSMessage(jwsMessage *joseJws.Message) (*googleUuid.UUID, *joseJwa.SignatureAlgorithm, error) {
	if len(jwsMessage.Signatures()) == 0 {
		return nil, nil, fmt.Errorf("JWS message has no signatures")
	}

	// Support multiple signatures by returning the first signature's kid and alg
	jwsMessageSignature := jwsMessage.Signatures()[0]
	jwsMessageProtectedHeaders := jwsMessageSignature.ProtectedHeaders()

	var kidUUIDString string

	err := jwsMessageProtectedHeaders.Get(joseJwk.KeyIDKey, &kidUUIDString)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get kid UUID: %w", err)
	}

	kidUUID, err := googleUuid.Parse(kidUUIDString)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse kid UUID: %w", err)
	}

	var alg joseJwa.SignatureAlgorithm

	err = jwsMessageProtectedHeaders.Get(joseJwk.AlgorithmKey, &alg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get alg: %w", err)
	}

	return &kidUUID, &alg, nil
}

// LogJWSInfo logs information about a JWS message.
func LogJWSInfo(jwsMessage *joseJws.Message) error {
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
		if publicHeaders != nil {
			for _, key := range publicHeaders.Keys() {
				var value any

				err := publicHeaders.Get(key, &value)
				if err != nil {
					logMessageSigHeader += fmt.Sprintf(" %s=%v\n", key, value)
				}
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
