# JOSE (JSON Object Signing and Encryption) Product

This directory contains the JOSE product implementation for cryptoutil.

## Current Status

**Note**: JOSE functionality is currently implemented in `internal/common/crypto/jose/`.
This directory is prepared for the migration to the 4-products architecture where JOSE
becomes a standalone product with its own API service.

## Target Structure

```text
internal/jose/
├── application/     # Server application lifecycle
├── config/          # JOSE-specific configuration
├── handler/         # HTTP handlers
├── jwk/             # JWK operations
├── jwe/             # JWE operations
├── jws/             # JWS operations
├── jwt/             # JWT operations
├── middleware/      # HTTP middleware
└── service/         # JOSE service (was jwkgen_service.go)
```

## Key Concepts

### JOSE Standards

- **JWK** (RFC 7517) - JSON Web Key
- **JWKS** - JSON Web Key Set
- **JWE** (RFC 7516) - JSON Web Encryption
- **JWS** (RFC 7515) - JSON Web Signature
- **JWT** (RFC 7519) - JSON Web Token

### Supported Algorithms (FIPS 140-3 Compliant)

| Type | Algorithms |
|------|------------|
| Signing | PS256, PS384, PS512, RS256, RS384, RS512, ES256, ES384, ES512, EdDSA |
| Key Wrapping | RSA-OAEP, RSA-OAEP-256, A128KW, A192KW, A256KW |
| Content Encryption | A128GCM, A192GCM, A256GCM, A128CBC-HS256, A192CBC-HS384, A256CBC-HS512 |
| Key Agreement | ECDH-ES, ECDH-ES+A128KW, ECDH-ES+A192KW, ECDH-ES+A256KW |

## Architecture

JOSE is the foundational product (P1) that is embedded in all other products:

- **P2 Identity** uses JOSE for JWT tokens, JWKS endpoints
- **P3 KMS** uses JOSE for key representation and crypto operations
- **P4 CA** will use JOSE for certificate-related cryptographic operations

## Migration

Files will be migrated from `internal/common/crypto/jose/` to `internal/jose/`:

| Source | Target |
|--------|--------|
| `common/crypto/jose/jwkgen.go` | `jose/jwk/generator.go` |
| `common/crypto/jose/jwkgen_service.go` | `jose/service/service.go` |
| `common/crypto/jose/jwe_*.go` | `jose/jwe/*.go` |
| `common/crypto/jose/jws_*.go` | `jose/jws/*.go` |
| `common/crypto/jose/jwk_*.go` | `jose/jwk/*.go` |
