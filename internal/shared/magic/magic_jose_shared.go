// Copyright (c) 2025-2026 Justin Cranford.
//

package magic

// JOSE key use constants shared by multiple services.
const (
	JoseKeyUseSig = "sig"
	JoseKeyUseEnc = "enc"
)

// Shared JOSE/JWK limits used across services.
const (
	JoseJADefaultMaxMaterials       = 10
	JoseJAMinMaterials              = 1
	JoseJAMaxMaterials              = 100
	JoseJADefaultListLimit          = 1000
	JoseJAAuditFallbackSamplingRate = 0.01
)

// JOSE algorithm constants.
const (
	JoseAlgRS256      = "RS256"
	JoseAlgRS384      = "RS384"
	JoseAlgRS512      = "RS512"
	JoseAlgPS256      = "PS256"
	JoseAlgPS384      = "PS384"
	JoseAlgPS512      = "PS512"
	JoseAlgES256      = "ES256"
	JoseAlgES384      = "ES384"
	JoseAlgES512      = "ES512"
	JoseAlgEdDSA      = "EdDSA"
	JoseAlgHS256      = "HS256"
	JoseAlgHS384      = "HS384"
	JoseAlgHS512      = "HS512"
	JoseAlgRSAOAEP    = "RSA-OAEP"
	JoseAlgRSAOAEP256 = "RSA-OAEP-256"
	JoseAlgECDHES     = "ECDH-ES"
	JoseAlgDir        = "dir"
)

// JOSE key type constants.
const (
	JoseKeyTypeRSA2048    = "RSA/2048"
	JoseKeyTypeRSA3072    = "RSA/3072"
	JoseKeyTypeRSA4096    = "RSA/4096"
	JoseKeyTypeECP256     = "EC/P256"
	JoseKeyTypeECP384     = "EC/P384"
	JoseKeyTypeECP521     = "EC/P521"
	JoseKeyTypeOKPEd25519 = "OKP/Ed25519"
	JoseKeyTypeOct128     = "oct/128"
	JoseKeyTypeOct192     = "oct/192"
	JoseKeyTypeOct256     = "oct/256"
	JoseKeyTypeOct384     = "oct/384"
	JoseKeyTypeOct512     = "oct/512"
)

// JOSE encryption constants.
const (
	JoseEncA128GCM      = "A128GCM"
	JoseEncA192GCM      = "A192GCM"
	JoseEncA256GCM      = "A256GCM"
	JoseEncA128CBCHS256 = "A128CBC-HS256"
	JoseEncA192CBCHS384 = "A192CBC-HS384"
	JoseEncA256CBCHS512 = "A256CBC-HS512"
)
