# cryptoutil
Golang util for crypto

Requires Go 1.24+

# Setup Project

## Initialize
```sh
go mod init github.com/justincranford/cryptoutil
```

## Install Utilities
```sh
go get -tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install mvdan.cc/gofumpt@latest
```

## Generate Go Fiber Code From OpenAPI Spec
```sh
go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest --config=openapi_gen_model.yaml  openapi_spec_components.yaml
go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest --config=openapi_gen_server.yaml openapi_spec_paths.yaml

go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest --config=openapi_gen_client.yaml openapi_spec_paths.yaml
```

## Run Automated Tests

```sh
go test ./... -coverprofile=coverage.out          && \
go tool cover -html=coverage.out -o coverage.html && \
start coverage.html
```

## Open Swagger UI for Manual Tests

```sh
(go run main.go &) && \
start http://localhost:8080/swagger
```

## Run Linters

```sh
golangci-lint run
gofumpt -l -w .
```
