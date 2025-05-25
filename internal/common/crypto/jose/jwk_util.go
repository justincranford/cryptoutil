package jose

import (
	"fmt"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilUtil "cryptoutil/internal/common/util"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

var (
	KtyOCT = joseJwa.OctetSeq() // KeyType
	KtyRSA = joseJwa.RSA()      // KeyType
	KtyEC  = joseJwa.EC()       // KeyType
	KtyOKP = joseJwa.OKP()      // KeyType

	EncA256GCM       = joseJwa.A256GCM()                                // ContentEncryptionAlgorithm
	EncA192GCM       = joseJwa.A192GCM()                                // ContentEncryptionAlgorithm
	EncA128GCM       = joseJwa.A128GCM()                                // ContentEncryptionAlgorithm
	EncA256CBC_HS512 = joseJwa.A256CBC_HS512()                          // ContentEncryptionAlgorithm
	EncA192CBC_HS384 = joseJwa.A192CBC_HS384()                          // ContentEncryptionAlgorithm
	EncA128CBC_HS256 = joseJwa.A128CBC_HS256()                          // ContentEncryptionAlgorithm
	EncInvalid       = joseJwa.NewContentEncryptionAlgorithm("invalid") // ContentEncryptionAlgorithm

	AlgA256KW       = joseJwa.A256KW()                             // KeyEncryptionAlgorithm
	AlgA192KW       = joseJwa.A192KW()                             // KeyEncryptionAlgorithm
	AlgA128KW       = joseJwa.A128KW()                             // KeyEncryptionAlgorithm
	AlgA256GCMKW    = joseJwa.A256GCMKW()                          // KeyEncryptionAlgorithm
	AlgA192GCMKW    = joseJwa.A192GCMKW()                          // KeyEncryptionAlgorithm
	AlgA128GCMKW    = joseJwa.A128GCMKW()                          // KeyEncryptionAlgorithm
	AlgRSA15        = joseJwa.RSA1_5()                             // KeyEncryptionAlgorithm
	AlgRSAOAEP      = joseJwa.RSA_OAEP()                           // KeyEncryptionAlgorithm
	AlgRSAOAEP256   = joseJwa.RSA_OAEP_256()                       // KeyEncryptionAlgorithm
	AlgRSAOAEP384   = joseJwa.RSA_OAEP_384()                       // KeyEncryptionAlgorithm
	AlgRSAOAEP512   = joseJwa.RSA_OAEP_512()                       // KeyEncryptionAlgorithm
	AlgECDHES       = joseJwa.ECDH_ES()                            // KeyEncryptionAlgorithm
	AlgECDHESA128KW = joseJwa.ECDH_ES_A128KW()                     // KeyEncryptionAlgorithm
	AlgECDHESA192KW = joseJwa.ECDH_ES_A192KW()                     // KeyEncryptionAlgorithm
	AlgECDHESA256KW = joseJwa.ECDH_ES_A256KW()                     // KeyEncryptionAlgorithm
	AlgDir          = joseJwa.DIRECT()                             // KeyEncryptionAlgorithm
	AlgEncInvalid   = joseJwa.NewKeyEncryptionAlgorithm("invalid") // KeyEncryptionAlgorithm

	AlgRS256      = joseJwa.RS256()                          // SignatureAlgorithm
	AlgRS384      = joseJwa.RS384()                          // SignatureAlgorithm
	AlgRS512      = joseJwa.RS512()                          // SignatureAlgorithm
	AlgPS256      = joseJwa.PS256()                          // SignatureAlgorithm
	AlgPS384      = joseJwa.PS384()                          // SignatureAlgorithm
	AlgPS512      = joseJwa.PS512()                          // SignatureAlgorithm
	AlgES256      = joseJwa.ES256()                          // SignatureAlgorithm
	AlgES384      = joseJwa.ES384()                          // SignatureAlgorithm
	AlgES512      = joseJwa.ES512()                          // SignatureAlgorithm
	AlgHS256      = joseJwa.HS256()                          // SignatureAlgorithm
	AlgHS384      = joseJwa.HS384()                          // SignatureAlgorithm
	AlgHS512      = joseJwa.HS512()                          // SignatureAlgorithm
	AlgEdDSA      = joseJwa.EdDSA()                          // SignatureAlgorithm
	AlgSigInvalid = joseJwa.NewSignatureAlgorithm("invalid") // SignatureAlgorithm

	OpsEncDec = joseJwk.KeyOperationList{joseJwk.KeyOpEncrypt, joseJwk.KeyOpDecrypt} // []KeyOperation
	OpsSigVer = joseJwk.KeyOperationList{joseJwk.KeyOpSign, joseJwk.KeyOpVerify}     // []KeyOperation
)

func ExtractKidUuid(jwk joseJwk.Key) (*googleUuid.UUID, error) {
	if jwk == nil {
		return nil, fmt.Errorf("invalid jwk: %w", cryptoutilAppErr.ErrCantBeNil)
	}
	var err error
	var kidString string
	if err = jwk.Get(joseJwk.KeyIDKey, &kidString); err != nil {
		return nil, fmt.Errorf("failed to get kid header: %w", err)
	}
	var kidUuid googleUuid.UUID
	if kidUuid, err = googleUuid.Parse(kidString); err != nil {
		return nil, fmt.Errorf("failed to parse kid as UUID: %w", err)
	}
	if err = cryptoutilUtil.ValidateUUID(&kidUuid, "invalid kid"); err != nil {
		return nil, err
	}
	return &kidUuid, nil
}
