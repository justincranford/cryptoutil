package openapi

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1-0.20250618140738-aae687ce8fe9 --config=./openapi-gen_config_model.yaml  ./openapi_spec_components.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1-0.20250618140738-aae687ce8fe9 --config=./openapi-gen_config_server.yaml ./openapi_spec_paths.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1-0.20250618140738-aae687ce8fe9 --config=./openapi-gen_config_client.yaml ./openapi_spec_paths.yaml
