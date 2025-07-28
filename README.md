# cryptoutil

cryptoutil is an embedded Key Management System (KMS), and a library for cryptographic operations. It is designed for extensibility, testability, and integration with modern Go frameworks.

## Main Features
- Only use NIST FIPS 140-3 approved algorithms and key sizes (e.g., RSA ≥ 2048 bits, AES ≥ 128 bits, EC NIST curves, EdDSA)
- Key generation for RSA, ECDSA, ECDH, EdDSA, AES, HMAC, UUIDv7
- Generic key pools for efficient concurrent key generation
- OpenAPI-driven API generation and documentation
- Fiber-based HTTP server with strict OpenAPI endpoints
- Integrated telemetry, including OpenTelemetry OTLP for logging, metrics, and tracing

## Extensibility
- OpenAPI specs can be used to add and generate components and handlers
- Generic key pools can be used for concurrent generation of more key types and algorithms

# Setup Project

Requires Go 1.24+

## Initialize
```sh
go mod init github.com/justincranford/cryptoutil
go mod tidy
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

Generate will find and run oapi-codegen commands embedded in [internal/openapi/generate.go](https://github.com/justincranford/cryptoutil/blob/main/internal/openapi/generate.go). The commands use OpenAPI spec files to generate code in internal/openapi/{model|server|client}/*.go.

## Testing

### Running with PostgreSQL
```sh
# Run PostgreSQL in Docker
docker compose up -d postgres

# Run cryptoutil connecting to PostgreSQL
go run main.go --config=./config/postgresql-local-debug.yaml
```

For development in VS Code, use the "cryptoutil postgres-local" launch configuration.

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

# Placeholder for some TO DO tasks

## Performance
- Benchmark pool and key generation under load

## Error Reporting
- Ensure all errors are logged and surfaced to clients appropriately

## Recommendations
- Expand documentation with usage examples and architecture diagrams
- Maintain clear instructions for setup, crypto usage, testing, and error reporting
