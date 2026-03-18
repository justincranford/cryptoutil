# Current State Analysis

## What Works Well

### 1. ServerBuilder Fluent API

The builder pattern in internal/apps/framework/service/server/builder/ is solid:
- NewServerBuilder() -> WithDomainMigrations() -> WithPublicRouteRegistration() -> Build()
- Error accumulation (railway-oriented) prevents partial construction
- Single Build() produces a fully-wired ServiceResources bundle
- Eliminates ~48,000 lines of boilerplate per service

This is the right pattern. The problem is not the pattern — it is what surrounds it.

### 2. ServiceResources Dependency Injection

The ServiceResources struct provides clean constructor injection:
- DB (GORM), TelemetryService, JWKGenService, BarrierService
- UnsealKeysService, SessionManager, RealmService, RealmRepository
- Application, ShutdownCore(), ShutdownContainer()
Services receive all dependencies ready-to-use without knowing how they were created.

### 3. Infrastructure Quality

- Dual HTTPS (public :8080 + admin :9090) separation is production-grade
- Health checks (livez/readyz/shutdown) are automatic
- TLS auto-generation for tests; real certs in production
- Migration merging (template 1001-1004 + domain 2001+) works

### 4. Security-First Design

- FIPS 140-3 compliance built into the framework, not added later
- Barrier (hierarchical encryption at rest) is mandatory, not optional
- Multi-tenancy via tenant_id filters baked into the pattern
- JWT/session authentication modes are configurable

### 5. OpenAPI-First

- oapi-codegen strict server ensures type safety
- Client generation available
- Consistent API patterns across services

---

## What Hurts

### Pain Point 1: Skeleton Is a Stub (21 files)

The skeleton-template exists but has almost no substance. It demonstrates that
the builder CAN be used, but does not show HOW a full domain service should be
structured. When starting a new service, a developer must:
1. Read through 316+ template files to understand the framework
2. Read an existing migrated service (e.g., sm-im) to see patterns
3. Mentally map those patterns to their new service
4. Manually write everything from scratch

A real stereotype would SHORT-CIRCUIT steps 1-4.

### Pain Point 2: No Compile-Time Contract Enforcement

There is no Go interface that a service MUST implement to be a valid framework
service. A service can forget to:
- Register health check routes (caught at runtime, not compile time)
- Register migrations (caught at runtime)
- Implement dual admin/public paths (not checked at all)
- Follow session/auth patterns (not checked at all)

Compare with Spring Boot: if you do not implement ApplicationRunner, you simply
do not get that lifecycle hook. Missing it is a choice, not an accident.

### Pain Point 3: Framework Evolution Requires Manual Updates

When the template framework changes (e.g., a new WithStrictServer() method),
every service that uses the old pattern must be manually updated. There is no:
- Automated diff tool to show what changed
- Deprecation mechanism that compiles with a warning
- Automated migration tool from old pattern to new

### Pain Point 4: Each Service Migration Is a Full Deep-Dive

Migrating a service to the template requires understanding ALL of:
- TLS configuration options
- BarrierConfig vs default barrier
- MigrationMode (TemplateWithDomain vs DomainOnly)
- JWTAuthConfig vs session auth
- StrictServerConfig for OpenAPI
This knowledge is not encoded in code — it is in ARCHITECTURE.md and comments.

### Pain Point 5: No Cross-Service Conformance Checking

Services drift apart silently:
- sm-im may handle sessions differently than jose-ja
- jose-ja may have a different CORS policy than pki-ca
- Error formats may diverge between services
There is no automated check that says "all 10 services have consistent X".

### Pain Point 6: No Scaffolding

A solo developer with 10 services to build needs:
- cicd new-service --product sm --service newservice
- Outputs a fully-wired, compiling, passing-tests skeleton
- Developer adds only domain logic

---

## Root Cause Analysis

The fundamental issue is a FRAMEWORK/LIBRARY CONFUSION:

A LIBRARY (what cryptoutil has): You call it. It gives you building blocks.
You still have to wire everything together. Deep knowledge required.

A FRAMEWORK (what cryptoutil needs): It calls you. You fill in the blanks.
Convention over configuration. Minimal knowledge required to be productive.

The ServerBuilder IS a good framework start (it calls your registration function).
But the surrounding ecosystem — skeleton, contracts, scaffolding, conformance —
is still in library mode.

### The Spring Boot Lesson

Spring Boot's genius was not the DI container (that was Spring 2.x).
The genius was AUTO-CONFIGURATION: "if you have a DataSource bean, I will
auto-configure a JdbcTemplate. If you have spring-security on the classpath,
I will auto-configure security with sensible defaults."

The cryptoutil equivalent: "if your ServiceManifest declares BarrierEnabled=true,
I will auto-configure BarrierService, key generation, unseal key management."
Service code should not select options — the manifest declares capabilities,
the framework configures them.

### The Identity Services Challenge

Identity has 5 services (authz, idp, rp, rs, spa) that are the hardest to migrate:
- They share auth flows (OAuth 2.1, OIDC)  
- They have complex cross-service dependencies
- They have the most security surface area
- They require federation patterns

This is not just a migration problem — it is a framework modularity problem.
The identity services need modules that sm/jose/pki do not need. The framework
needs a MODULE SYSTEM for this.

---

## The Divergence Math Problem

With 10 services each with 50-100 files, any shared pattern change multiplies:
- 1 pattern change *10 services* 2 hours/service = 20 hours to propagate
- 10 pattern changes per month = 200 hours/month maintenance overhead

This is unsustainable for a solo developer. The solution is not to work harder —
it is to make shared patterns live in exactly ONE place that auto-propagates.

In Go, the best mechanism for this is: INTERFACES + CODE GENERATION.
