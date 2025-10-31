---
description: "Instructions for project configuration: OpenAPI, magic values, linting exclusions"
applyTo: "**"
---
# Project Configuration Instructions

## OpenAPI and Code Generation

- Use [OpenAPI 3.0.3](https://spec.openapis.org/oas/v3.0.3) for all API specs
- Split specs into components.yaml/paths.yaml
- Generate code with [oapi-codegen](https://github.com/deepmap/oapi-codegen) using strict server pattern
- Use request/response validation, status codes, REST conventions, JSON content types, error schemas
- Support pagination with page/size parameters

## Magic Values Management

- Define all magic numbers/values in `cryptoutilMagic` package files (`internal/common/magic/magic_*.go`)
- Group related constants by category: `magic_buffers.go`, `magic_timeouts.go`, `magic_network.go`
- Use descriptive constant names indicating purpose and units
- Magic value files automatically ignored by mnd exclude-files filter
- Update `.golangci.yml` importas configuration to include `cryptoutilMagic` alias
- Remove numbers from mnd ignored-numbers list once defined as constants

**Examples:**
```go
const (
    defaultPort = 8080
    adminPort = 9090
    timeout = 30 * time.Second
    maxRetries = 3
    bufferSize1KB = 1024
)
```

## Linting Exclusions

### Standard Exclusions (Always Apply)

**Generated Code:**
- `_gen.go` - Auto-generated Go files
- `.pb.go` - Protocol buffer files
- `api/` - OpenAPI generated code

**Test Directories:**
- `test/` - Contains Java Gatling tests, not Go code

**Dependencies:**
- `vendor/` - Vendored dependencies

**Build Artifacts:**
- `.exe`, `.dll`, `.so`, `.dylib` - Binaries
- `*.key`, `*.crt`, `*.pem` - Certificates/keys

**IDE Files:**
- `.vscode/` - VS Code settings

### Exclusion Pattern
Use regex: `'_gen\.go$|\.pb\.go$|vendor/|api/|test/'`

### Application Examples

**Pre-commit:**
```yaml
exclude: '_gen\.go$|\.pb\.go$|vendor/|api/|test/'
```

**golangci-lint Config:**
```yaml
issues:
  exclude-dirs: [vendor, api, test]
  exclude-files: [".*\\.pb\\.go$", ".*_gen\\.go$"]
```

**Scripts:**
```bash
golangci-lint run --skip-files='.*_gen\.go$|.*\.pb\.go$' --skip-dirs=vendor,api,test
```
