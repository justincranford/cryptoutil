// Copyright (c) 2025 Justin Cranford

// Package magic provides domain-specific constants and configuration values.
package magic

// JOSE-JA Service Configuration.
const (
	// OTLPServiceJoseJA is the OTLP service name for jose-ja telemetry.
	OTLPServiceJoseJA = "jose-ja"

	// JoseJAServicePort is the default public API port for jose-ja service.
	JoseJAServicePort = 9443

	// JoseJAAdminPort is the admin API port (same for all services).
	JoseJAAdminPort = 9090
)

// JOSE Key Use Constants.
const (
	// JoseKeyUseSig indicates a key is used for signing (JWS).
	JoseKeyUseSig = "sig"

	// JoseKeyUseEnc indicates a key is used for encryption (JWE).
	JoseKeyUseEnc = "enc"
)

// JOSE-JA Elastic Key Configuration.
const (
	// JoseJADefaultMaxMaterials is the default maximum number of material keys per elastic key.
	JoseJADefaultMaxMaterials = 10

	// JoseJAMinMaterials is the minimum number of material keys per elastic key.
	JoseJAMinMaterials = 1

	// JoseJAMaxMaterials is the maximum number of material keys per elastic key.
	JoseJAMaxMaterials = 100
)

// JOSE-JA Audit Configuration.
const (
	// JoseJAAuditDefaultEnabled is the default audit enabled state.
	JoseJAAuditDefaultEnabled = true

	// JoseJAAuditDefaultSamplingRate is the default audit sampling rate (100% = log all).
	JoseJAAuditDefaultSamplingRate = 100

	// JoseJAAuditFallbackSamplingRate is the fallback sampling rate when no config exists (1%).
	JoseJAAuditFallbackSamplingRate = 0.01

	// JoseJAAuditMinSamplingRate is the minimum audit sampling rate.
	JoseJAAuditMinSamplingRate = 0

	// JoseJAAuditMaxSamplingRate is the maximum audit sampling rate.
	JoseJAAuditMaxSamplingRate = 100
)

// JOSE-JA E2E Test Configuration.
const (
	// JoseJAE2ESQLitePublicPort is the public port for SQLite E2E tests.
	JoseJAE2ESQLitePublicPort = 9443

	// JoseJAE2EPostgreSQL1PublicPort is the public port for first PostgreSQL E2E tests.
	JoseJAE2EPostgreSQL1PublicPort = 9444

	// JoseJAE2EPostgreSQL2PublicPort is the public port for second PostgreSQL E2E tests.
	JoseJAE2EPostgreSQL2PublicPort = 9445

	// JoseJAE2EGrafanaPort is the Grafana port for E2E tests.
	JoseJAE2EGrafanaPort = 3000

	// JoseJAE2EOtelCollectorGRPCPort is the OpenTelemetry collector gRPC port for E2E tests.
	JoseJAE2EOtelCollectorGRPCPort = 4317

	// JoseJAE2EOtelCollectorHTTPPort is the OpenTelemetry collector HTTP port for E2E tests.
	JoseJAE2EOtelCollectorHTTPPort = 4318
)

// JOSE-JA Service Pagination Constants.
const (
	// JoseJADefaultListLimit is the default limit for list operations.
	JoseJADefaultListLimit = 1000
)

// JOSE Algorithm Constants for Key Generation and Use.
const (
	// JoseAlgRS256 is RSA PKCS#1 signature with SHA-256.
	JoseAlgRS256 = "RS256"
	// JoseAlgRS384 is RSA PKCS#1 signature with SHA-384.
	JoseAlgRS384 = "RS384"
	// JoseAlgRS512 is RSA PKCS#1 signature with SHA-512.
	JoseAlgRS512 = "RS512"
	// JoseAlgPS256 is RSA PSS signature with SHA-256.
	JoseAlgPS256 = "PS256"
	// JoseAlgPS384 is RSA PSS signature with SHA-384.
	JoseAlgPS384 = "PS384"
	// JoseAlgPS512 is RSA PSS signature with SHA-512.
	JoseAlgPS512 = "PS512"
	// JoseAlgES256 is ECDSA signature with P-256 and SHA-256.
	JoseAlgES256 = "ES256"
	// JoseAlgES384 is ECDSA signature with P-384 and SHA-384.
	JoseAlgES384 = "ES384"
	// JoseAlgES512 is ECDSA signature with P-521 and SHA-512.
	JoseAlgES512 = "ES512"
	// JoseAlgEdDSA is EdDSA signature using Ed25519.
	JoseAlgEdDSA = "EdDSA"

	// JoseKeyTypeRSA2048 is RSA with 2048-bit key.
	JoseKeyTypeRSA2048 = "RSA/2048"
	// JoseKeyTypeRSA3072 is RSA with 3072-bit key.
	JoseKeyTypeRSA3072 = "RSA/3072"
	// JoseKeyTypeRSA4096 is RSA with 4096-bit key.
	JoseKeyTypeRSA4096 = "RSA/4096"

	// JoseKeyTypeECP256 is ECDSA with P-256 curve.
	JoseKeyTypeECP256 = "EC/P256"
	// JoseKeyTypeECP384 is ECDSA with P-384 curve.
	JoseKeyTypeECP384 = "EC/P384"
	// JoseKeyTypeECP521 is ECDSA with P-521 curve.
	JoseKeyTypeECP521 = "EC/P521"

	// JoseKeyTypeOKPEd25519 is OKP with Ed25519 curve.
	JoseKeyTypeOKPEd25519 = "OKP/Ed25519"

	// JoseKeyTypeOct128 is symmetric 128-bit key.
	JoseKeyTypeOct128 = "oct/128"
	// JoseKeyTypeOct192 is symmetric 192-bit key.
	JoseKeyTypeOct192 = "oct/192"
	// JoseKeyTypeOct256 is symmetric 256-bit key.
	JoseKeyTypeOct256 = "oct/256"
	// JoseKeyTypeOct384 is symmetric 384-bit key.
	JoseKeyTypeOct384 = "oct/384"
	// JoseKeyTypeOct512 is symmetric 512-bit key.
	JoseKeyTypeOct512 = "oct/512"

	// JoseEncA128GCM is AES 128-bit GCM encryption.
	JoseEncA128GCM = "A128GCM"
	// JoseEncA192GCM is AES 192-bit GCM encryption.
	JoseEncA192GCM = "A192GCM"
	// JoseEncA256GCM is AES 256-bit GCM encryption.
	JoseEncA256GCM = "A256GCM"
	// JoseEncA128CBCHS256 is AES 128-bit CBC with HMAC SHA-256.
	JoseEncA128CBCHS256 = "A128CBC-HS256"
	// JoseEncA192CBCHS384 is AES 192-bit CBC with HMAC SHA-384.
	JoseEncA192CBCHS384 = "A192CBC-HS384"
	// JoseEncA256CBCHS512 is AES 256-bit CBC with HMAC SHA-512.
	JoseEncA256CBCHS512 = "A256CBC-HS512"

	// JoseAlgRSAOAEP is RSA-OAEP key encryption.
	JoseAlgRSAOAEP = "RSA-OAEP"
	// JoseAlgRSAOAEP256 is RSA-OAEP-256 key encryption.
	JoseAlgRSAOAEP256 = "RSA-OAEP-256"
	// JoseAlgECDHES is ECDH-ES key encryption.
	JoseAlgECDHES = "ECDH-ES"
	// JoseAlgDir is direct encryption.
	JoseAlgDir = "dir"
)
