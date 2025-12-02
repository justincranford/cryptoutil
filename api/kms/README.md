# KMS OpenAPI Specifications

This directory contains the OpenAPI specifications for the Key Management Service (KMS) product.

## Files

- `openapi_spec_kms.yaml` - KMS API paths and operations
- `openapi_spec_components.yaml` - Shared schemas and components
- `openapi-gen_config_*.yaml` - oapi-codegen configuration files

## Current Status

**Note**: KMS OpenAPI specs are currently in the parent `api/` directory:

- `api/openapi_spec_paths.yaml` - Contains KMS paths (to be moved)
- `api/openapi_spec_components.yaml` - Contains KMS components (to be moved)

This directory structure is being prepared for the 4-products architecture alignment.

## Generation

```bash
cd api/kms
go generate ./...
```
