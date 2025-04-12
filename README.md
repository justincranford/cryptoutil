# cryptoutil
Golang util for crypto

Requires Go 1.24+

# Setup Project

## Initialize
```sh
go mod init github.com/justincranford/cryptoutil
```

## Generate OpenAPI Handlers & Models

### Install

```sh
go get -tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
```

## Generate
```sh
go generate ./...
```

Look at [internal/openapi/generate.go](https://github.com/justincranford/cryptoutil/blob/main/internal/openapi/generate.go) for the commands used to generate internal/openapi/{model|server|client}/*.go code from OpenAPI spec files.

## Testing

### Automated Tests
```sh
go test ./... -coverprofile=coverage.out          && \
go tool cover -html=coverage.out -o coverage.html && \
start coverage.html
```

### Manual Tests with Swagger UI
```sh
go run main.go &
start http://localhost:8080/swagger
fg
```

## Linters

### Install

```sh
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install mvdan.cc/gofumpt@latest
```

### Run

```sh
golangci-lint run
gofumpt -l -w .
```
