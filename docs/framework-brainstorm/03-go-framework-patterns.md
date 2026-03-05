# Go-Specific Framework Patterns

Patterns and libraries available in the Go ecosystem that are directly applicable
to cryptoutil's framework evolution goals.

---

## Dependency Injection in Go

### The Problem

Go does not have annotations or decorators. Frameworks like Spring rely on
reflection + annotations for DI. Go's philosophy is explicit over implicit.
Two viable approaches: compile-time DI (Wire) and runtime DI (fx).

---

### Google Wire — Compile-Time Dependency Injection

Wire (<https://github.com/google/wire>) generates wiring code at compile time.

How it works:
1. You declare provider functions: NewDB(cfg Config) (*DB, error)
2. You declare injector specifications: //go:build wireinject
3. wire binary runs: reads providers, generates wire_gen.go
4. Generated code wires everything at compile time

`go
// providers.go
func NewDB(cfg *Config) (*gorm.DB, error) { ... }
func NewBarrierService(db *gorm.DB, keys UnsealKeysService) (*BarrierService, error) { ... }
func NewKeyHandler(barrier *BarrierService)*KeyHandler { ... }

// wire_injector.go (build tag: wireinject)
func InitializeApp(cfg *Config) (*App, error) {
    wire.Build(NewDB, NewBarrierService, NewKeyHandler, NewApp)
    return nil, nil // Wire fills this in
}

// wire_gen.go (GENERATED, do not edit)
func InitializeApp(cfg *Config) (*App, error) {
    db, err := NewDB(cfg)
    if err != nil { return nil, err }
    barrier, err := NewBarrierService(db, ...)
    ...
}
`

Benefits for cryptoutil:
- No more ServiceResources god-object; each service gets exactly what it needs
- Wiring errors caught at compile time, not runtime
- Each service's dependency graph is visible in one file
- Changes to dependencies show up as compile errors immediately

Downsides:
- Requires running wire binary (a go generate step)
- Generated code can be verbose
- Learning curve for the provider/injector pattern

Verdict: HIGHLY RECOMMENDED for cryptoutil if you want compile-time DI safety.
Especially valuable for the identity services which have complex cross-dependencies.

---

### Uber fx — Runtime Dependency Injection

fx (<https://github.com/uber-go/fx>) uses reflection-based DI similar to Spring,
but with explicit registration:

`go
app := fx.New(
    fx.Provide(NewDB),
    fx.Provide(NewBarrierService),
    fx.Provide(NewKeyHandler),
    fx.Invoke(func(h *KeyHandler, server *fiber.App) {
        server.Get("/service/api/v1/keys", h.ListKeys)
    }),
)
app.Run()
`

Benefits:
- No code generation step needed
- fx.Module() for grouping related providers
- Lifecycle hooks: fx.Hook{OnStart, OnStop}
- Excellent for large services with many dependencies

Downsides:
- Runtime reflection (not compile-time safety)
- Dependency errors discovered at startup, not compile time
- Black-box wiring can be hard to debug

fx.Module system maps directly to the module concept from 04-framework-design.md:
`go
var BarrierModule = fx.Module("barrier",
    fx.Provide(NewBarrierService),
    fx.Provide(NewUnsealKeysService),
    fx.Provide(NewKeyRotationService),
)

var SessionModule = fx.Module("sessions",
    fx.Provide(NewSessionManager),
    fx.Provide(NewRealmService),
)

// Service assembly:
fx.New(
    BarrierModule,
    SessionModule,
    fx.Module("sm-im", ...),
)
`

Verdict: fx.Module system is the best analog to Spring Boot starters in Go.
Excellent for composing services from reusable modules.

---

### Functional Options Pattern — Already Used

cryptoutil already uses functional options in some places. This pattern scales:
`go
type ServerOption func(*serverConfig)

func WithBarrier(config BarrierConfig) ServerOption {
    return func(c *serverConfig) { c.barrier = &config }
}

func WithSessions(config SessionConfig) ServerOption {
    return func(c *serverConfig) { c.sessions = &config }
}

func NewServer(options ...ServerOption) *Server {
    cfg := defaultServerConfig()
    for _, opt := range options {
        opt(cfg)
    }
    return buildServer(cfg)
}
`

This is simpler than Wire/fx and works well for up to ~10 options.
For cryptoutil's 6+ modules, the module system via fx.Module is cleaner.

---

## Code Generation

### go generate — Standard Mechanism

//go:generate command is the standard Go code generation hook.
cryptoutil already uses it for oapi-codegen.

Extension opportunities:
`go
// In skeleton-template service.go:
//go:generate cryptoutil-gen handler --spec=api/openapi_spec_paths.yaml --output=server/apis/generated_handlers.go
//go:generate cryptoutil-gen client --spec=api/openapi_spec_paths.yaml --output=client/generated_client.go
//go:generate cryptoutil-gen mock --interface=KeyRepository --output=testing/mock_repository.go
//go:generate cryptoutil-gen test-stubs --handler=server/apis/handlers.go --output=server/apis/handlers_test.go
`

The cryptoutil-gen tool does not exist yet. It should be in cmd/cicd or cmd/codegen.

---

### oapi-codegen — Already In Use, Extend It

Currently: generates server, models, client from OpenAPI spec.
Could also generate:
- Mock implementations of the StrictServerInterface (for testing)
- OpenAPI spec validation middleware
- API documentation site (Redoc, Swagger UI already done)
- E2E test stubs that call generated client against live server

---

### Buf (protobuf/gRPC generation) — Future Consideration

If cryptoutil ever needs gRPC (internal service communication), buf provides:
- buf generate: generates Go code from .proto files
- buf lint: enforces style conventions on .proto files  
- buf breaking: detects breaking API changes between versions
- buf push: pushes proto definitions to BSR (Buf Schema Registry)

The buf plugin system + protobuf could replace or complement OpenAPI for
internal service-to-service communication (currently uses mTLS HTTP).

---

### Text/Template and Go Templates

For the cicd new-service scaffolding tool, Go's text/template is sufficient:
`go
tmpl := template.Must(template.ParseFS(skeletonFS, "skeleton/**/*.tmpl"))
tmpl.Execute(outFile, ServiceManifest{
    Product: "sm",
    Name: "myservice",
    WithBarrier: true,
    WithSessions: true,
    WithOpenAPI: true,
})
`

The templates live in the skeleton-template directory as *.tmpl files.
Running cicd new-service renders them into the new service directory.

---

## CLI Architecture

### Cobra — Already Used in cryptoutil

cryptoutil uses Cobra for CLI commands. Key patterns to add:
1. Command groups per service: cryptoutil sm im server already works
2. Completion: cryptoutil completion bash for shell completion
3. Version command: cryptoutil version showing all 10 service versions

---

### Viper — Already Used

Viper handles config loading (YAML + secrets + CLI flags).
Additional Viper patterns to consider:
- viper.WatchConfig() for hot config reload in development
- Config validation on startup (already done via validate_schema.go)
- Config documentation generation from struct tags

---

## Testing Patterns

### Testify Suite — For Contract Testing

testify.Suite allows setup/teardown around a test group.
For cross-service contract tests:
`go
type ServiceContractSuite struct {
    suite.Suite
    client ServiceClient  // any service that implements the contract
}

func (s *ServiceContractSuite) TestHealthEndpointAvailable() {
    resp, err := s.client.Health()
    s.NoError(err)
    s.Equal(200, resp.StatusCode)
}

func (s *ServiceContractSuite) TestReadyzReturnsOKWhenReady() { ... }
func (s*ServiceContractSuite) TestShutdownEndpointGraceful() { ... }

// Run the same suite against all services:
suite.Run(t, &ServiceContractSuite{client: smIMClient})
suite.Run(t, &ServiceContractSuite{client: joseJAClient})
suite.Run(t, &ServiceContractSuite{client: pkiCAClient})
`

This is EXACTLY how protocol testing works in the Go standard library.
io.Reader has a standard test suite; any implementation can run it.

---

### gomock / moq — Mock Generation

`
//go:generate moq -out testing/mock_key_repository.go . KeyRepository
`

moq generates a mock struct from an interface. Avoids hand-written mocks.
Every service interface (Repository, Service, Handler) should have a generated mock.

---

### testcontainers-go — Already Used

cryptoutil already uses testcontainers for PostgreSQL.
Consider extending for:
- Redis (if ever needed)
- OTel collector (for verifying spans are emitted correctly)
- Service-under-test containers (for E2E tests in CI)

---

## Architecture Enforcement in Go

### go-arch-lint — Architecture Invariants

go-arch-lint enforces import rules:
`yaml

# .arch-lint.yml

components:
  framework:   { files: ['internal/apps/template/**'] }
  domain-sm:   { files: ['internal/apps/sm/**'] }
  domain-jose: { files: ['internal/apps/jose/**'] }

rules:
- component: domain-sm
    allow: [framework]     # sm can import template
    deny: [domain-jose]    # sm cannot import jose

- component: domain-jose
    allow: [framework]
    deny: [domain-sm]
`

Already enforced via cicd go-check-identity-imports, but that is custom code.
go-arch-lint would be more general and more powerful.

---

### ArchUnit for Go (Custom)

There is no official ArchUnit port for Go, but you can write architecture tests
using go/packages and go/types:

`go
// internal/arch_test.go
func TestNoServiceImportsAnotherService(t *testing.T) {
    pkgs := loadAllPackages("cryptoutil/internal/apps/...")
    for _, pkg := range pkgs {
        for _, imp := range pkg.Imports {
            if isServicePackage(imp) && isDifferentServicePackage(pkg, imp) {
                t.Errorf("%s imports %s (cross-service import forbidden)", pkg.Path, imp)
            }
        }
    }
}
`

This catches violations at test time in CI. Not at compile time, but close.

---

## Monitoring Framework Code Quality

### Fitness Functions via go test

Architecture fitness functions encoded as tests:
- All services have health endpoint: parse routes at startup
- All services follow naming conventions: ast inspection
- No cross-service imports: import graph analysis
- Coverage thresholds: parse coverprofile.out
- Migration version ranges: inspect migration FS at compile time

go-critic, staticcheck already check many patterns.
Custom fitness function tests fill the gaps.

---

## Live Reload for Development

### air — Live Reload for Go

air (<https://github.com/air-verse/air>) watches for file changes and restarts:
` oml

# .air.toml

[build]
  cmd = "go build -o ./tmp/main ./cmd/sm-im/"
  bin = "./tmp/main"
  include_ext = ["go", "yaml"]
  exclude_dir = ["vendor", "tmp", "testdata"]

[run]
  args = ["server", "--config", "configs/sm-im/development.yml"]
`

air would dramatically improve the inner development loop for cryptoutil.
Currently: edit -> wait 30s for rebuild -> test. With air: edit -> <3s -> test.

---

## Summary: Go Ecosystem Picks

| Need | Tool | Priority |
|------|------|----------|
| Compile-time DI | Google Wire | High |
| Runtime DI / modules | Uber fx | High |
| Code generation | text/template + go generate | Critical |
| Mock generation | moq | Medium |
| Live reload | air | High (dev experience) |
| Architecture tests | custom + go/packages | High |
| Import enforcement | go-arch-lint | Medium |
| gRPC (future) | buf | Low |
