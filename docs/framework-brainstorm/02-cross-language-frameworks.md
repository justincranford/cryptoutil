# Cross-Language Framework Lessons

What do mature frameworks across languages do that cryptoutil can learn from?
Organized by language, then pattern, then applicability to cryptoutil.

---

## Java

### Spring Boot — The Gold Standard for App Frameworks

Spring Boot is the most influential application framework in modern software.
Its innovations are directly applicable to cryptoutil:

#### Auto-Configuration
Mechanism: @ConditionalOnClass, @ConditionalOnBean, @ConditionalOnProperty
Result: Add spring-security to classpath -> security is auto-wired. No boilerplate.
cryptoutil equivalent: ServiceManifest.barriers = true -> BarrierService auto-wired.
ServiceManifest.openapi = true -> StrictServer auto-wired.

#### Spring Boot Starters
A starter is a curated dependency bundle + auto-configuration.
spring-boot-starter-web -> Tomcat + Spring MVC + Jackson, pre-configured.
spring-boot-starter-data-jpa -> Hibernate + DataSource + TransactionManager.
cryptoutil equivalent: A framework.BarrierModule could bundle:
- BarrierService init, unseal key service, key rotation, status routes
- All wired together, service just enables it via manifest

#### Spring Initializr (start.spring.io)
Web form: select Spring Web, Spring Data JPA, Spring Security -> download ZIP.
The ZIP is a WORKING project with all selected modules wired.
cryptoutil equivalent: cicd new-service --product sm --name myservice --with barrier,openapi,sessions
Outputs: a fully compiling service with all the right files, ready for domain logic.

#### ApplicationRunner / CommandLineRunner
Interface a bean implements to run code after context is fully initialized.
Analogous to: WithPostStartHook(func(res *ServiceResources) error)
cryptoutil does not have this — services cannot easily run post-start without hacks.

#### Spring Boot Actuator
Exposes /actuator/health, /actuator/metrics, /actuator/info automatically.
Pluggable health indicators: implement HealthIndicator -> appears in /health.
cryptoutil has this (livez/readyz) but DB health is not pluggable per-service.

#### @Profile("test", "dev", "prod")
Different beans activated per environment profile.
cryptoutil equivalent: build tags (test vs production), but no runtime profile mechanism.
Could be powerful: Profile("dev") -> in-memory SQLite; Profile("prod") -> PostgreSQL.

---

### Dropwizard — The Bundle Pattern

Dropwizard uses a Bundle system for extending applications:

`java
bootstrap.addBundle(new AssetsBundle("/assets/", "/static/", "index.html"));
bootstrap.addBundle(new DBIBundle(database));
bootstrap.addBundle(new AuthBundle<>(authenticator));
`

Each bundle:
- Has initialize() called during bootstrap (registration phase)
- Has run() called after environment is ready (wiring phase)
- Can add routes, healthchecks, managed objects, commands

This is the CLOSEST analog to what cryptoutil's ServerBuilder should become.
Each bundle is independently testable and reusable.
cryptoutil could model: framework.Bundle interface with Init() and Run().

---

### Micronaut — Compile-Time Everything

Micronaut resolves DI at compile time, not runtime:
- No reflection — GraalVM native images work out of the box
- @Factory methods create beans at compile time
- Interfaces resolved at compile time via annotation processors

cryptoutil equivalent: Google Wire (see 03-go-framework-patterns.md).
Compile-time DI would eliminate ServiceResources nil-checks and runtime panics.
Less relevant for Go since Go has no annotation processors, but Wire provides it.

---

### Quarkus — Developer Mode + Live Reload

Quarkus dev mode: save a file -> server reloads in <1s (JVM warm path).
cryptoutil equivalent: air (https://github.com/air-verse/air):
- air watches for file changes -> go build -> restart server
- This is immediately applicable, requires zero framework changes
- Would significantly speed up the inner development loop

Quarkus also has a RESTEasy Reactive model where all middleware is composable Chain.
cryptoutil has this via Fiber middleware composition, but does not enforce consistency.

---

## Python

### Django — Batteries Included + Reusable Apps

Django's key insight: a project is a collection of apps. Each app is:
- A directory with models.py, views.py, urls.py, admin.py, migrations/
- Self-contained and reusable across projects
- Plugged in via settings.INSTALLED_APPS

`python
INSTALLED_APPS = [
    'django.contrib.auth',      # Auth framework = template equivalent
    'django.contrib.admin',     # Admin UI
    'myproject.billing',        # Domain service A
    'myproject.products',       # Domain service B
]
`

cryptoutil equivalent: each service is a "Django app" that implements an interface,
and the suite registers them:
`go
suite := framework.NewSuite(
    tempalte.AuthModule,        // Framework: sessions, auth, realms
    template.BarrierModule,     // Framework: barrier, key management
    smim.IMModule,              // Domain: instant messaging
    joseja.JAModule,            // Domain: JWK authority
)
`

#### Django management commands
python manage.py makemigrations — generates SQL from model changes.
python manage.py migrate — applies migrations.
python manage.py createsuperuser — creates admin.
python manage.py startapp billing — scaffolds a new app.

cryptoutil equivalent: the cicd command already exists! The missing pieces are:
- cicd new-service --product x --service y
- cicd make-migration --service jose-ja --name add_revocation
- cicd diff-skeleton --service pki-ca

#### Django Signal System
Decoupled observer pattern: code in billing app listens to auth events.
cryptoutil equivalent: post-startup hooks, barrier rotation events, etc.
Could be useful for cross-service event notification within the suite.

---

### FastAPI — OpenAPI-First with Full Code Generation

FastAPI is essentially what cryptoutil does, but for Python:
- Routes declared as Python functions with type annotations
- OpenAPI spec generated FROM code (inverse of cryptoutil's approach)
- Pydantic models = type-validated request/response schemas

The key difference: FastAPI generates OpenAPI FROM code; cryptoutil generates code FROM OpenAPI.
Both approaches work. OpenAPI-first (cryptoutil) is better for contract-first design.
FastAPI is better for fast iteration. Consider: could cryptoutil bridge both?
- Write an OpenAPI spec, generate the handler stubs,
  developer implements stubs, spec stays authoritative.

FastAPI's dependency injection via type annotations is elegant:
`python
async def create_key(
    body: CreateKeyRequest,          # auto-validated from OpenAPI
    db: Session = Depends(get_db),   # injected
    current_user: User = Depends(get_current_user),  # injected
):
`

cryptoutil equivalent: oapi-codegen strict server already does this for requests.
But DB and auth are injected via closure (public_server.go captures them).
No explicit DI annotation system needed in Go — closures work well.

---

### Flask — Blueprint Composability

Flask Blueprints = composable route groups:
`python
api_v1 = Blueprint('api_v1', __name__, url_prefix='/service/api/v1')

@api_v1.route('/keys', methods=['GET'])
def list_keys(): ...

app.register_blueprint(api_v1)
app.register_blueprint(admin_blueprint)
`

This is VERY similar to how cryptoutil registers routes via Fiber app.Group().
The difference: Flask blueprints are self-contained modules with their own
routes, error handlers, template filters, etc.

cryptoutil could define a DomainModule interface:
`go
type DomainModule interface {
    RegisterPublicRoutes(router fiber.Router, res *ServiceResources)
    RegisterAdminRoutes(router fiber.Router, res *ServiceResources)
    HealthChecks() []HealthCheck
    Migrations() fs.FS
}
`

---

## JavaScript / TypeScript

### NestJS — Most Spring-Boot-Like in JS

NestJS uses decorators and modules, mapping almost 1:1 to Spring Boot:
- @Module({imports, controllers, providers}) = Spring @Configuration
- @Injectable() = Spring @Service / @Component
- @Controller('/keys') = Spring @RestController
- Dependency injection in constructors

NestJS modules are the most explicit module system in any JS framework:
`	ypescript
@Module({
  imports: [DatabaseModule, AuthModule],
  controllers: [KeysController],
  providers: [KeysService, KeysRepository],
  exports: [KeysService],
})
export class KeysModule {}
`

KEY INSIGHT: explicitness of dependency graph. In cryptoutil, ServiceResources
contains everything — services take what they need via parameter unpacking.
NestJS forces explicit declaration of what each module needs and provides.

---

### Fastify — Plugin System with fp

Fastify's plugin system (fastify-plugin fp) is the most sophisticated HTTP plugin
system in any framework:
`javascript
const myPlugin = fp(async (fastify, options) => {
  fastify.decorate('db', createDB(options.dsn))
  fastify.addHook('onRequest', authenticate)
  fastify.addHook('preHandler', authorize)
}, {
  name: 'my-plugin',
  dependencies: ['fastify-jwt'],
})

fastify.register(myPlugin, { dsn: 'postgres://...' })
fastify.register(require('./routes/keys'))
fastify.register(require('./routes/certs'))
`

KEY INSIGHT: plugins can decorate the server instance with new capabilities.
They can add hooks. They can declare dependencies on other plugins.
This is more powerful than what cryptoutil has: route registration is just routes.

Fastify encapsulation: plugins registered inside fastify.register() are scoped.
Routes in plugin A cannot see routes in plugin B unless explicitly shared.
cryptoutil currently has no encapsulation — all routes are on the same Fiber app.

---

### Express — Middleware Pipeline

Express established the middleware pattern that Fiber follows:
`javascript
app.use(cors())
app.use(rateLimiter)
app.use('/service', authenticate)
app.use('/browser', sessionAuth)
`

cryptoutil already uses this pattern via Fiber. The improvement would be:
consistent middleware stacks defined in ONE place (framework) not per-service.

---

## Rust

### Axum — Extractors and Typed Routing

Axum's extractor system is the most elegant dependency injection in any HTTP framework:
`ust
async fn create_key(
    State(db): State<Database>,       // extracted from app state
    Json(body): Json<CreateKeyRequest>, // extracted and validated from body
    Extension(user): Extension<User>, // extracted from request extensions
) -> impl IntoResponse { ... }
`

KEY INSIGHT: the function SIGNATURE declares its dependencies.
No explicit DI wiring needed — Axum figures it out.
This is more ergonomic than Go closures but comes at the cost of complex traits.

In Go, the closest is oapi-codegen strict server where the handler signature
is generated from the OpenAPI spec. The types enforce correctness.

### Tower — Service Trait and Layers

Tower defines Service<Request> as the core abstraction:
`ust
trait Service<Request> {
    type Response;
    type Error;
    type Future;
    fn call(&mut self, req: Request) -> Self::Future;
}
`

Every middleware is a Service that wraps another Service.
Layers compose middleware: ServiceBuilder::new().layer(LoggingLayer).layer(AuthLayer).

cryptoutil's Fiber middleware is similar. The Tower model is more composable
because each layer is independently testable and reusable.

Go analogy: net/http.Handler interface + middleware pattern (already in use).

---

## Go

### Buffalo — Closest to Rails in Go

Buffalo (https://gobuffalo.io) attempts to be Rails for Go:
- buffalo new myapp -> fully scaffolded project
- buffalo generate resource user name:text email:text -> CRUD + migrations + tests
- buffalo dev -> live reload
- buffalo build -> optimized production binary

KEY LESSON FROM BUFFALO'S FATE: Buffalo was ambitious but struggled because:
1. Rails conventions are opinionated in ways that do not transfer to Go idioms
2. Code generation creates maintainability concerns when output is modified
3. The framework grew too large and maintenance suffered (one lead developer)

APPLICABLE TO CRYPTOUTIL: be careful not to over-generate. Generate stubs, not
implementations. The developer fills in the domain logic; the framework provides
the structure.

### Beego — MVC with Code Generation

Beego's bee tool:
- bee new myproject — scaffold a new project
- bee generate controller user — generate controller
- bee generate model user name:string email:string — generate model + migration
- bee run — live reload

The bee generate approach is valuable. Note that Beego has declined in popularity
partly due to framework lock-in: once on Beego, hard to leave.
cryptoutil should generate into a well-known structure (its own skeleton),
not into a proprietary format.

### go-kit — Microservice Toolkit

go-kit (https://gokit.io) takes a different approach: not a framework but a toolkit.
Each service is composed from:
- Transport layer (HTTP, gRPC, AMQP)
- Endpoint layer (request/response)
- Service layer (business logic)
- Middleware layer (logging, tracing, circuit breaker)

KEY INSIGHT: go-kit separates transport from logic with a explicit boundary.
cryptoutil has this: Fiber handlers (transport) are separate from service layer.
The difference is go-kit generates the transport binding from an interface.

go-kit with go generate:
`
//go:generate gokit_gen --src=service.go --transport=http
`

Generates HTTP binding, client, logging middleware, tracing middleware,
rate limiting middleware — all from the interface.

This is a powerful pattern. Extend to cryptoutil:
`
//go:generate cryptoutil-gen --spec=openapi.yaml --service=KeysService
`

Generates: handler stubs, client, test stubs, mock.

---

## Key Cross-Language Synthesis

### Pattern 1: Module/Bundle System (Universal)
Every mature framework has modules: Spring Starters, Django apps, NestJS modules,
Fastify plugins, Dropwizard bundles. These share:
- Clear interface: what a module provides and requires
- Lifecycle: init, start, stop
- Isolation: modules do not leak into each other
- Discoverability: framework knows all registered modules

Current cryptoutil: single ServerBuilder with options. Should become: modules.

### Pattern 2: Convention Over Configuration (Universal)
Rails, Django, Spring Boot all: if it is named the right way in the right place,
it works automatically. No wiring needed.
cryptoutil opportunity: if a service has migrations/2001_*.sql, auto-register.
If it has api/openapi_spec.yaml, auto-configure strict server.

### Pattern 3: Scaffolding / Code Generation (Universal)
Rails scaffold, Django startapp, Spring Initializr, buffalo generate, bee generate.
A solo developer with 10 services NEEDS: cicd new-service.
It is not optional. Without it, every service is a full manual effort.

### Pattern 4: Fitness Functions / Architecture Tests (Emerging)
ArchUnit (Java), Dependency Cruiser (JS), Conform (Go, see next doc).
Automated tests that verify architectural constraints do not regress.
Cross-service: every service has a health endpoint, no cross-service imports, etc.

### Pattern 5: Live Reload / Fast Iteration (Universal)
air (Go), buffalo dev, quarkus dev, nodemon (Node), flask --debug.
Inner loop speed matters enormously for a solo developer. 5s restart vs 30s restart
over 8 hours = significant productivity difference.

### Pattern 6: OpenAPI-First + Full Generation (Modern Trend)
Not just stubs: generate models, handlers, clients, mocks, tests.
buf (protobuf), oapi-codegen, openapi-generator.
More generation = less manual code = less divergence.