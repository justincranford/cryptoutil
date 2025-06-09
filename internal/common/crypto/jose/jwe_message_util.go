package jose

import (
	"encoding/json"
	"fmt"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

func EncryptBytes(jwks []joseJwk.Key, clearBytes []byte) (*joseJwe.Message, []byte, error) {
	if jwks == nil {
		return nil, nil, fmt.Errorf("invalid JWKs: %w", cryptoutilAppErr.ErrCantBeNil)
	} else if len(jwks) == 0 {
		return nil, nil, fmt.Errorf("invalid JWKs: %w", cryptoutilAppErr.ErrCantBeEmpty)
	} else if clearBytes == nil {
		return nil, nil, fmt.Errorf("invalid clearBytes: %w", cryptoutilAppErr.ErrCantBeNil)
	} else if len(clearBytes) == 0 {
		return nil, nil, fmt.Errorf("invalid clearBytes: %w", cryptoutilAppErr.ErrCantBeEmpty)
	}

	encs := make(map[joseJwa.ContentEncryptionAlgorithm]struct{})
	jweEncryptOptions := make([]joseJwe.EncryptOption, 0, len(jwks))
	if len(jwks) > 1 { // more than one JWK requires using JSON encoding, instead of default Compact encoding
		jweEncryptOptions = append(jweEncryptOptions, joseJwe.WithJSON())
	}
	for i, jwk := range jwks {
		enc, alg, err := ExtractAlgEncFromJweJwk(&jwk, i)
		if err != nil {
			return nil, nil, fmt.Errorf("JWK %d invalid: %w", i, err)
		}
		if len(encs) == 0 {
			jweEncryptOptions = append(jweEncryptOptions, joseJwe.WithContentEncryption(*enc)) // only add CEK alg once
		}
		encs[*enc] = struct{}{} // track ContentEncryptionAlgorithm counts
		if len(encs) != 1 {     // validate that one-and-only-one ContentEncryptionAlgorithm is used across all JWKs
			return nil, nil, fmt.Errorf("can't use JWK %d 'enc' attributes; only one unique 'enc' attribute is allowed", i)
		}
		jweEncryptOptions = append(jweEncryptOptions, joseJwe.WithKey(*alg, jwk)) // add ALG+JWK tuple for each JWK
	}

	jweMessageBytes, err := joseJwe.Encrypt(clearBytes, jweEncryptOptions...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt clearBytes: %w", err)
	}

	jweMessage, err := joseJwe.Parse(jweMessageBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse JWE message bytes: %w", err)
	}

	return jweMessage, jweMessageBytes, nil
}

func DecryptBytes(jwks []joseJwk.Key, jweMessageBytes []byte) ([]byte, error) {
	if jwks == nil {
		return nil, fmt.Errorf("invalid JWKs: %w", cryptoutilAppErr.ErrCantBeNil)
	} else if len(jwks) == 0 {
		return nil, fmt.Errorf("invalid JWKs: %w", cryptoutilAppErr.ErrCantBeEmpty)
	} else if jweMessageBytes == nil {
		return nil, fmt.Errorf("invalid jweMessageBytes: %w", cryptoutilAppErr.ErrCantBeNil)
	} else if len(jweMessageBytes) == 0 {
		return nil, fmt.Errorf("invalid jweMessageBytes: %w", cryptoutilAppErr.ErrCantBeEmpty)
	}

	jweMessage, err := joseJwe.Parse(jweMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWE message bytes: %w", err)
	}

	encs := make(map[joseJwa.ContentEncryptionAlgorithm]struct{})
	jweDecryptOptions := make([]joseJwe.DecryptOption, 0, len(jwks))
	for i, jwk := range jwks {
		enc, alg, err := ExtractAlgEncFromJweJwk(&jwk, i)
		if err != nil {
			return nil, fmt.Errorf("JWK %d invalid: %w", i, err)
		}
		encs[*enc] = struct{}{}
		if len(encs) != 1 {
			return nil, fmt.Errorf("can't use JWK %d 'enc' attributes; only one unique 'enc' attribute is allowed", i)
		}
		jweDecryptOptions = append(jweDecryptOptions, joseJwe.WithKey(*alg, jwk))
	}
	jweDecryptOptions = append(jweDecryptOptions, joseJwe.WithMessage(jweMessage))

	decryptedBytes, err := joseJwe.Decrypt(jweMessageBytes, jweDecryptOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt JWE message bytes: %w", err)
	}

	return decryptedBytes, nil
}

func EncryptKey(keks []joseJwk.Key, clearCek joseJwk.Key) (*joseJwe.Message, []byte, error) {
	clearCekBytes, err := json.Marshal(clearCek)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode CEK: %w", err)
	}
	return EncryptBytes(keks, clearCekBytes)
}

func DecryptKey(kdks []joseJwk.Key, encryptedCdkBytes []byte) (joseJwk.Key, error) {
	decryptedCdkBytes, err := DecryptBytes(kdks, encryptedCdkBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt CDK bytes: %w", err)
	}
	decryptedCdk, err := joseJwk.ParseKey(decryptedCdkBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to derypt CDK: %w", err)
	}
	return decryptedCdk, nil
}

func JweHeadersString(jweMessage *joseJwe.Message) (string, error) {
	jweHeadersString, err := json.Marshal(jweMessage.ProtectedHeaders())
	if err != nil {
		return "", fmt.Errorf("failed to marshall JWE headers: %w", err)
	}
	return string(jweHeadersString), err
}

func ExtractKidEncAlgFromJweMessage(jweMessage *joseJwe.Message) (*googleUuid.UUID, *joseJwa.ContentEncryptionAlgorithm, *joseJwa.KeyEncryptionAlgorithm, error) {
	var kidUuidString string
	err := jweMessage.ProtectedHeaders().Get(joseJwk.KeyIDKey, &kidUuidString)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get kid UUID: %w", err)
	}
	kidUuid, err := googleUuid.Parse(kidUuidString)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse kid UUID: %w", err)
	}

	var enc joseJwa.ContentEncryptionAlgorithm
	err = jweMessage.ProtectedHeaders().Get("enc", &enc)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get enc: %w", err)
	}
	var alg joseJwa.KeyEncryptionAlgorithm
	err = jweMessage.ProtectedHeaders().Get(joseJwk.AlgorithmKey, &alg)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get alg: %w", err)
	}
	return &kidUuid, &enc, &alg, nil
}
