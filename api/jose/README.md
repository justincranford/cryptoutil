# JOSE OpenAPI Specifications

This directory contains the OpenAPI specifications for the JOSE (JSON Object Signing and Encryption) product.

## Planned Files

- `openapi_spec_jose.yaml` - JOSE API paths and operations
- `openapi_spec_components.yaml` - Shared schemas and components
- `openapi-gen_config_*.yaml` - oapi-codegen configuration files

## Current Status

**Note**: JOSE functionality is currently implemented as a library in `internal/common/crypto/jose/`.
This directory is a placeholder for the future JOSE Authority API service.

## Planned Endpoints

- `POST /jwk/generate` - Generate JWK
- `GET /jwks` - Get JWKS (JSON Web Key Set)
- `POST /jwe/encrypt` - Encrypt payload with JWE
- `POST /jwe/decrypt` - Decrypt JWE payload
- `POST /jws/sign` - Sign payload with JWS
- `POST /jws/verify` - Verify JWS signature
- `POST /jwt/create` - Create JWT
- `POST /jwt/validate` - Validate JWT

## Generation

```bash
cd api/jose
go generate ./...
```
