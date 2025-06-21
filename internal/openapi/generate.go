package openapi

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1-0.20250618140738-aae687ce8fe9 --config=./openapi_gen_model.yaml  ./openapi_spec_components.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1-0.20250618140738-aae687ce8fe9 --config=./openapi_gen_server.yaml ./openapi_spec_paths.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1-0.20250618140738-aae687ce8fe9 --config=./openapi_gen_client.yaml ./openapi_spec_paths.yaml
