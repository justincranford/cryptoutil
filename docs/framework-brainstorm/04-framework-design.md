# Framework Design Proposals

Concrete design ideas for evolving internal/apps/template/service/ into a
true application framework equivalent to Spring Boot's @SpringBootApplication.

---

## The Core Insight: Framework vs Library

Current: cryptoutil ServerBuilder is a LIBRARY.
  - You call ServerBuilder
  - You configure it
  - You get back ServiceResources
  - You set up everything else yourself

Target: cryptoutil framework is a FRAMEWORK.
  - Framework calls you via interfaces
  - You declare capabilities via a manifest or interface
  - Framework configures everything
  - You fill in domain logic only

---

## Design Proposal 1: ServiceContract Interface

The simplest improvement: define a Go interface that every service MUST implement.

`go
// Package framework (internal/apps/template/service/framework)

// ServiceContract is the interface every cryptoutil service must satisfy.
// This is enforced at compile time: if a service does not implement this interface,
// it will not compile.
type ServiceContract interface {
    // Metadata
    ServiceID() string   // e.g., "sm-im"
    ProductID() string   // e.g., "sm"
    
    // Domain migrations (2001+). Return nil if no domain tables needed.
    DomainMigrations() fs.FS
    
    // RegisterPublicRoutes adds domain-specific routes to the public server.
    // Called after all framework infrastructure is initialized.
    RegisterPublicRoutes(app *fiber.App, res *ServiceResources) error
    
    // OpenAPISpec returns the embedded OpenAPI spec. Return nil if not using OpenAPI.
    OpenAPISpec() []byte
}

// ExtendedServiceContract adds optional lifecycle hooks.
// Services implement this if they need post-start initialization.
type ExtendedServiceContract interface {
    ServiceContract
    PostStart(res *ServiceResources) error
    PreStop(res *ServiceResources) error
}

// AdminServiceContract adds admin route registration.
type AdminServiceContract interface {
    ServiceContract
    RegisterAdminRoutes(app *fiber.App, res *ServiceResources) error
}
`

Services are then registered like this:
`go
// cmd/sm-im/main.go
func main() {
    svc := smim.NewService(cfg)
    framework.Run(svc)           // framework takes over, calls contract methods
}
`

This is a compile-time guarantee. If smim.Service does not implement ServiceContract,
the build fails. No runtime surprises.

---

## Design Proposal 2: Service Manifest

A manifest is a struct that declares service capabilities.
The framework reads it and auto-configures accordingly.

`go
// ServiceManifest declares what a service needs from the framework.
// Defined by domain services, read by the framework.
type ServiceManifest struct {
    // Identity
    ServiceID string
    ProductID string
    
    // Infrastructure capabilities (auto-configured when true)
    Barrier  bool   // Encryption-at-rest barrier (default: true)
    Sessions bool   // Session management (default: true)
    Realms   bool   // Authentication realm management (default: true)
    
    // Database
    DatabaseMode DatabaseMode   // GORM, RawSQL, or Both (default: GORM)
    
    // Authentication mode
    AuthMode AuthMode   // Session (default), JWT, or Both
    
    // API
    OpenAPI bool   // Register OpenAPI strict server (default: false)
    SwaggerUI bool // Register Swagger UI (default: false)
    
    // Migrations
    MigrationMode MigrationMode  // TemplateWithDomain (default) or DomainOnly
    
    // Paths
    PublicPaths  []string  // custom paths beyond /browser/** and /service/**
}

// Usage in skeleton/sm-im:
var manifest = framework.ServiceManifest{
    ServiceID:     "sm-im",
    ProductID:     "sm",
    Barrier:       true,
    Sessions:      true,
    OpenAPI:       true,
    MigrationMode: framework.TemplateWithDomain,
}
`

The framework reads the manifest during NewServerBuilder() and configures:
- If Barrier=true -> BarrierService fully wired
- If Sessions=true -> SessionManager, RealmService wired
- If OpenAPI=true -> StrictServerConfig wired
- If Barrier=false -> lightweight mode (like KMS which has its own barrier)

This ELIMINATES the need for the developer to call WithJWTAuth(), WithBarrier(),
WithStrictServer() — the manifest declares it all.

---

## Design Proposal 3: Module System (Dropwizard/fx Inspired)

Replace the monolithic ServerBuilder options with composable modules.
Each module is independently testable, versioned, and documented.

`go
// FrameworkModule is implemented by framework modules (built-in).
type FrameworkModule interface {
    Name() string
    Init(ctx context.Context, cfg *Config) error
    Provides() []interface{}   // what types this module provides
    Requires() []interface{}   // what types this module requires
    Shutdown()
}

// Built-in modules (in internal/apps/template/service/modules/):
var (
    DatabaseModule     FrameworkModule  // GORM DB, migrations, connection pool
    TelemetryModule    FrameworkModule  // OTel, structured logging
    BarrierModule      FrameworkModule  // Barrier, key hierarchy, unseal keys
    SessionModule      FrameworkModule  // Sessions, realms, tenant management
    PublicServerModule FrameworkModule  // Fiber app, TLS, CORS, rate limiting
    AdminServerModule  FrameworkModule  // Admin server, health checks
    OpenAPIModule      FrameworkModule  // Strict server, Swagger UI
)

// Service assembly:
server := framework.NewServiceBuilder(ctx, cfg).
    WithModules(
        framework.DatabaseModule,
        framework.TelemetryModule,
        framework.BarrierModule,
        framework.SessionModule,
    ).
    WithDomain(myDomainModule).   // domain-specific module
    Build()
`

Each module handles its own initialization, shutdown, and health checks.
Modules declare dependencies (requires) and make things available (provides).
The framework resolves the dependency graph and initializes in the correct order.

---

## Design Proposal 4: Framework Hooks (Lifecycle-Aware)

Services should be able to participate in the service lifecycle without hacks.

`go
// Current: no way to run code after build, before start
// Proposed: lifecycle hooks

type ServiceHooks struct {
    // OnDBReady: called after DB is connected and migrations applied.
    // Use for: database seeding, pre-warming caches.
    OnDBReady func(db *gorm.DB) error
    
    // OnReady: called after public server is listening.
    // Use for: health check warmup, initial data load.
    OnReady func(res *ServiceResources) error
    
    // OnBarrierSealed: called when barrier becomes sealed (unseal keys unavailable).
    // Use for: graceful degradation, alerts.
    OnBarrierSealed func() error
    
    // OnBarrierUnsealed: called when barrier unseals successfully.
    // Use for: triggering key rotation checks, audit log entries.
    OnBarrierUnsealed func(res *ServiceResources) error
    
    // OnShutdown: called before servers stop accepting connections.
    // Use for: draining in-progress operations, flushing caches.
    OnShutdown func() error
}

// Usage:
builder.WithHooks(framework.ServiceHooks{
    OnReady: func(res *ServiceResources) error {
        return seedTestData(res.DB)
    },
    OnShutdown: func() error {
        return flushPendingMessages()
    },
})
`

---

## Design Proposal 5: Typed Configuration Binding

The current ServerSettings struct is 20+ fields. Navigating it is difficult.
A better pattern: typed sub-configurations per module.

`go
// Framework-level config (in config.go):
type FrameworkConfig struct {
    Server    ServerConfig
    Database  DatabaseConfig
    TLS       TLSConfig
    Telemetry TelemetryConfig
    Barrier   BarrierConfig
    Sessions  SessionConfig
    Auth      AuthConfig
}

// Each module reads only its sub-config. No module sees others' config.
// Domain services add their own config block:
type SMIMConfig struct {
    Framework framework.FrameworkConfig
    SMIM      SMIMDomainConfig
}

type SMIMDomainConfig struct {
    MaxMessageSize int
    RetentionDays  int
    EncryptionAlgo string
}
`

This is how Spring Boot starter properties work:
spring.datasource.url, spring.datasource.username, spring.datasource.password
are read only by the DataSource auto-configuration, not by the whole app.

---

## Design Proposal 6: Admin Route Extension Point

Currently admin routes (barrier rotation, join request management) are registered
inside the builder. Services cannot add custom admin routes without modifying the builder.

`go
builder.WithAdminRouteRegistration(func(
    adminApp *fiber.App,
    res *ServiceResources,
) error {
    // Service-specific admin operations
    adminApp.Get("/admin/api/v1/cache-stats", handleCacheStats)
    adminApp.Post("/admin/api/v1/compact-keys", handleKeyCompaction)
    return nil
})
`

This mirrors WithPublicRouteRegistration but for the admin server.
Consider: all routes under /admin/api/v1/domain/* reserved for domain services;
/admin/api/v1/framework/* reserved for framework infrastructure.

---

## Design Proposal 7: Framework Version and Compatibility

As the framework evolves, services need to know which framework version they use.

`go
// In framework package:
const FrameworkVersion = "1.3.0"
const MinimumSupportedFrameworkVersion = "1.0.0"

// In each service manifest:
var manifest = framework.ServiceManifest{
    ServiceID:        "sm-im",
    FrameworkVersion: framework.FrameworkVersion,  // checked at build time
    // ...
}
`

When the framework changes, a build failure tells you which services need updating.
This is better than silent runtime failures or manual documentation checks.

---

## Design Proposal 8: Framework Validation at Startup

Before starting any service, the framework should validate its own configuration:

`go
// framework/validation.go
func (b *ServerBuilder) Validate() []ValidationError {
    var errors []ValidationError
    
    // Validate service manifest
    if b.manifest.ServiceID == "" {
        errors = append(errors, ValidationError{"manifest.service_id", "required"})
    }
    
    // Validate module dependencies
    if b.manifest.Sessions && !b.manifest.Barrier {
        errors = append(errors, ValidationError{"barrier", "required when sessions=true"})  
    }
    
    // Validate TLS configuration
    if b.config.BindPublicPort > 0 && !b.config.TLSEnabled {
        errors = append(errors, ValidationError{"tls", "required for public port"})
    }
    
    return errors
}
`

Fail fast at startup with ALL validation errors (not just the first one).
This follows the existing validator error aggregation pattern.

---

## Bringing It Together: The Ideal Service Registration

What creating a new service should look like after all proposals are implemented:

`go
// internal/apps/sm/newservice/service.go

package newservice

//go:generate cryptoutil-gen --manifest=manifest.go --spec=api/openapi.yaml

var manifest = framework.ServiceManifest{
    ServiceID:     "sm-newservice",
    ProductID:     "sm",
    Barrier:       true,
    Sessions:      true,
    OpenAPI:       true,
    MigrationMode: framework.TemplateWithDomain,
}

type Service struct {
    res *framework.ServiceResources
}

func New(ctx context.Context, cfg *Config) (*Service, error) {
    builder := framework.NewServerBuilder(ctx, cfg.Framework, manifest)
    builder.WithDomainMigrations(MigrationsFS, "migrations")
    builder.WithPublicRoutes(RegisterRoutes)
    builder.WithAdminRoutes(RegisterAdminRoutes)
    res, err := builder.Build()
    if err != nil { return nil, err }
    return &Service{res: res}, nil
}

// RegisterRoutes is called by the framework after infrastructure is ready.
// This is the ONLY place domain logic is wired.
func RegisterRoutes(app *fiber.App, res *framework.ServiceResources) error {
    repo := NewItemRepository(res.DB)
    svc := NewItemService(repo, res.BarrierService)
    h := NewItemHandler(svc, res.SessionManager)
    
    g := app.Group("/service/api/v1", res.SessionManager.Middleware())
    g.Get("/items", h.List)
    g.Post("/items", h.Create)
    g.Get("/items/:id", h.Get)
    
    return nil
}
`

Compare with today: the developer also has to set up TLS, admin server,
barrier, sessions, telemetry, migrations, all in server.go. That is ~100 lines.
With the framework, RegisterRoutes is the ENTIRE domain wiring: ~15 lines.