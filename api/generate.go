// Copyright (c) 2025 Justin Cranford
//
//

// Package openapi provides generated OpenAPI client and server code for the cryptoutil API.
package openapi

// TODO Move to golang command pattern directory
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1-0.20250618140738-aae687ce8fe9 --config=./openapi-gen_config_model.yaml  ./openapi_spec_components.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1-0.20250618140738-aae687ce8fe9 --config=./openapi-gen_config_server.yaml ./openapi_spec_paths.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1-0.20250618140738-aae687ce8fe9 --config=./openapi-gen_config_client.yaml ./openapi_spec_paths.yaml
//go:generate go run -tags=fixexternalref ./fix_external_ref
