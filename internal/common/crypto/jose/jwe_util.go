package jose

import (
	"encoding/json"
	"fmt"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

var (
	KtyOct           = joseJwa.OctetSeq()                                                   // KeyType
	EncA256GCM       = joseJwa.A256GCM()                                                    // ContentEncryptionAlgorithm
	EncA192GCM       = joseJwa.A192GCM()                                                    // ContentEncryptionAlgorithm
	EncA128GCM       = joseJwa.A128GCM()                                                    // ContentEncryptionAlgorithm
	EncA256CBC_HS512 = joseJwa.A256CBC_HS512()                                              // ContentEncryptionAlgorithm
	EncA192CBC_HS384 = joseJwa.A192CBC_HS384()                                              // ContentEncryptionAlgorithm
	EncA128CBC_HS256 = joseJwa.A128CBC_HS256()                                              // ContentEncryptionAlgorithm
	AlgA256KW        = joseJwa.A256KW()                                                     // KeyEncryptionAlgorithm
	AlgA192KW        = joseJwa.A192KW()                                                     // KeyEncryptionAlgorithm
	AlgA128KW        = joseJwa.A128KW()                                                     // KeyEncryptionAlgorithm
	AlgA256GCMKW     = joseJwa.A256GCMKW()                                                  // KeyEncryptionAlgorithm
	AlgA192GCMKW     = joseJwa.A192GCMKW()                                                  // KeyEncryptionAlgorithm
	AlgA128GCMKW     = joseJwa.A128GCMKW()                                                  // KeyEncryptionAlgorithm
	AlgDir           = joseJwa.DIRECT()                                                     // KeyEncryptionAlgorithm
	OpsEncDec        = joseJwk.KeyOperationList{joseJwk.KeyOpEncrypt, joseJwk.KeyOpDecrypt} // []KeyOperation
)

func EncryptBytes(jwks []joseJwk.Key, clearBytes []byte) (*joseJwe.Message, []byte, error) {
	if jwks == nil {
		return nil, nil, fmt.Errorf("invalid JWKs: %w", cryptoutilAppErr.ErrCantBeNil)
	} else if len(jwks) == 0 {
		return nil, nil, fmt.Errorf("invalid JWKs: %w", cryptoutilAppErr.ErrCantBeEmpty)
	}

	encs := make(map[joseJwa.ContentEncryptionAlgorithm]struct{})
	jweEncryptOptions := make([]joseJwe.EncryptOption, 0, len(jwks))
	if len(jwks) > 1 {
		jweEncryptOptions = append(jweEncryptOptions, joseJwe.WithJSON()) // if more than one JWK, must use JSON encoding instead of default Compact encoding
	}
	for i, jwk := range jwks {
		enc, alg, err := getJwkAlgAndEnc(jwk, i)
		if err != nil {
			return nil, nil, fmt.Errorf("JWK %d invalid: %w", i, err)
		}
		if len(encs) == 0 {
			jweEncryptOptions = append(jweEncryptOptions, joseJwe.WithContentEncryption(*enc)) // only add CEK alg once
		}
		encs[*enc] = struct{}{} // ensure CEK alg is the same for all JWKs
		if len(encs) != 1 {
			return nil, nil, fmt.Errorf("can't use JWK %d 'enc' attributes; only one unique 'enc' attribute is allowed", i)
		}
		jweEncryptOptions = append(jweEncryptOptions, joseJwe.WithKey(*alg, jwk)) // add ALG+JWK tuple for each JWK
	}

	encodedJweMessage, err := joseJwe.Encrypt(clearBytes, jweEncryptOptions...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt clearBytes: %w", err)
	}

	jweMessage, err := joseJwe.Parse(encodedJweMessage)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse encrypted JWE message bytes: %w", err)
	}

	return jweMessage, encodedJweMessage, nil
}

func DecryptBytes(jwks []joseJwk.Key, jweMessageBytes []byte) ([]byte, error) {
	if jwks == nil {
		return nil, fmt.Errorf("invalid JWKs: %w", cryptoutilAppErr.ErrCantBeNil)
	} else if len(jwks) == 0 {
		return nil, fmt.Errorf("invalid JWKs: %w", cryptoutilAppErr.ErrCantBeEmpty)
	}

	jweMessage, err := joseJwe.Parse(jweMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse encrypted JWE message bytes: %w", err)
	}

	encs := make(map[joseJwa.ContentEncryptionAlgorithm]struct{})
	jweDecryptOptions := make([]joseJwe.DecryptOption, 0, len(jwks))
	for i, jwk := range jwks {
		enc, alg, err := getJwkAlgAndEnc(jwk, i)
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

func JSONHeadersString(jweMessage *joseJwe.Message) (string, error) {
	jweHeadersString, err := json.Marshal(jweMessage.ProtectedHeaders())
	if err != nil {
		return "", fmt.Errorf("failed to marshall JWE headers: %w", err)
	}
	return string(jweHeadersString), err
}

func getJwkAlgAndEnc(jwk joseJwk.Key, i int) (*joseJwa.ContentEncryptionAlgorithm, *joseJwa.KeyAlgorithm, error) {
	if jwk == nil {
		return nil, nil, fmt.Errorf("JWK %d invalid: %w", i, cryptoutilAppErr.ErrCantBeNil)
	}

	var enc joseJwa.ContentEncryptionAlgorithm
	err := jwk.Get("enc", &enc) // Example: A256GCM, A192GCM, A128GCM, A256CBC-HS512, A192CBC-HS384, A128CBC-HS256
	if err != nil {
		// Workaround: If JWK was serialized (for encryption) and parsed (after decryption), 'enc' header incorrect gets parsed as string, so try getting as string converting it to joseJwa.ContentEncryptionAlgorithm
		var encString string
		err = jwk.Get("enc", &encString)
		if err != nil {
			return nil, nil, fmt.Errorf("can't get JWK %d 'enc' attribute: %w", i, err)
		}
		enc = joseJwa.NewContentEncryptionAlgorithm(encString)
	}

	var alg joseJwa.KeyAlgorithm
	err = jwk.Get(joseJwk.AlgorithmKey, &alg) // Example: A256KW, A192KW, A128KW, A256GCMKW, A192GCMKW, A128GCMKW, dir
	if err != nil {
		return nil, nil, fmt.Errorf("can't get JWK %d 'alg' attribute: %w", i, err)
	}

	return &enc, &alg, nil
}
