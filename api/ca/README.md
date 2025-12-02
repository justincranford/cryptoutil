# CA OpenAPI Specifications

This directory contains the OpenAPI specifications for the Certificate Authority (CA) product.

## Planned Files

- `openapi_spec_ca.yaml` - CA API paths and operations
- `openapi_spec_components.yaml` - Shared schemas and components
- `openapi-gen_config_*.yaml` - oapi-codegen configuration files

## Current Status

**Note**: CA functionality is planned but not yet implemented.
This directory is a placeholder for the future CA API service.

## Planned Endpoints

### Certificate Lifecycle

- `POST /certificates` - Request new certificate
- `GET /certificates/{id}` - Get certificate by ID
- `GET /certificates` - List certificates
- `POST /certificates/{id}/revoke` - Revoke certificate

### CSR Operations

- `POST /csr/submit` - Submit Certificate Signing Request
- `GET /csr/{id}` - Get CSR status

### CRL/OCSP

- `GET /crl` - Get Certificate Revocation List
- `POST /ocsp` - OCSP request/response

### CA Management

- `GET /ca/chain` - Get CA certificate chain
- `GET /ca/root` - Get root CA certificate

## Compliance

- RFC 5280 - X.509 PKI Certificate and CRL Profile
- CA/Browser Forum Baseline Requirements
- ACME Protocol (RFC 8555)

## Generation

```bash
cd api/ca
go generate ./...
```
