# Server Builder Instructions QUIZME - Gaps & Improvements

**Last Updated**: 2026-01-16
**Purpose**: Identify gaps and improvements for 03-08.server-builder.instructions.md after analyzing actual implementation
**Context**: Comparing instructions file against actual service-template/server/builder/server_builder.go and cipher-im usage

## Instructions

**This document contains QUESTIONS about gaps or improvements needed in 03-08.server-builder.instructions.md.**

After answering, update 03-08.server-builder.instructions.md to address gaps.

**ANSWER FORMAT**: For each question, write your choice (A, B, C, D, or E) on the **YOUR ANSWER: __** line.
- If choosing E (write-in), also provide your custom answer below the question.

---

## Section 1: Missing Documentation of Builder Methods

### Q1: WithDomainMigrations Validation Rules

**Gap Identified**: Instructions show usage but don't document validation rules.

**Implementation Code**:
```go
func (b *ServerBuilder) WithDomainMigrations(migrationFS fs.FS, migrationsPath string) *ServerBuilder {
    if migrationFS == nil {
        b.err = fmt.Errorf("migration FS cannot be nil")
        return b
    }
    if migrationsPath == "" {
        b.err = fmt.Errorf("migrations path cannot be empty")
        return b
    }
    // ...
}
```

**Question**: Should instructions document validation rules for WithDomainMigrations?

**A)** YES - Add validation section documenting nil FS and empty path checks
- **Pros**: Helps developers avoid common errors
- **Cons**: Instructions become longer

**B)** NO - Validation is self-explanatory from signature
- **Pros**: Keeps instructions concise
- **Cons**: Developers may pass invalid values without understanding why errors occur

**C)** PARTIAL - Add troubleshooting section with common errors only
- **Pros**: Focused on solving problems, not preventing them
- **Cons**: Reactive rather than proactive

**D)** DEFER - Add validation docs when first reported issue occurs
- **Pros**: YAGNI principle
- **Cons**: May cause avoidable issues

**E)** Write-in (describe approach):

**YOUR ANSWER: __** B

---

### Q2: WithDefaultTenant Behavior Documentation

**Gap Identified**: Instructions don't explain what happens when defaultTenantID/defaultRealmID are Nil.

**Implementation Code**:
```go
// Ensure default tenant exists if configured.
if b.defaultTenantID != googleUuid.Nil && b.defaultRealmID != googleUuid.Nil {
    if err := b.ensureDefaultTenant(core.DB); err != nil {
        core.Shutdown()
        return nil, err
    }
}
```

**Question**: Should instructions document Nil UUID behavior?

**A)** YES - Add note: "Nil UUIDs skip default tenant creation (multi-tenant mode)"
- **Pros**: Clarifies behavior, shows how to disable default tenant
- **Cons**: Adds implementation detail

**B)** NO - Behavior obvious from "WithDefaultTenant" method name
- **Pros**: Concise instructions
- **Cons**: Not obvious that Nil values skip creation

**C)** YES - Add examples showing both single-tenant and multi-tenant patterns
- **Pros**: Clear usage patterns for both modes
- **Cons**: More verbose examples

**D)** PARTIAL - Document only in multi-tenant migration guide
- **Pros**: Focused documentation
- **Cons**: Developers may not find it

**E)** Write-in (describe approach): There is no such thing as default tenant. If it mentions default tenant, remove it, and replace it with guidance that default tenant is not allowed. New tenants must only ever be created via registering a browser user || service client, and specifying option to associate them with a new tenant.

**YOUR ANSWER: __** E

---

### Q3: WithPublicRouteRegistration Callback Pattern

**Gap Identified**: Instructions show callback but don't explain typical implementation patterns.

**Cipher-IM Example**:
```go
builder.WithPublicRouteRegistration(func(base *PublicServerBase, res *ServiceResources) error {
    messageRepo := repository.NewMessageRepository(res.DB)
    messageService := service.NewMessageService(messageRepo, res.BarrierService)
    publicServer := server.NewPublicServer(base, messageService, res.SessionManager)
    return publicServer.RegisterRoutes()
})
```

**Question**: Should instructions include full callback implementation example?

**A)** YES - Add complete cipher-im pattern as reference example
- **Pros**: Clear pattern to follow, reduces cognitive load
- **Cons**: Long example, may be service-specific

**B)** YES - Add generic pseudo-code pattern without service specifics
- **Pros**: Shows structure without tying to specific service
- **Cons**: Less concrete than real example

**C)** NO - Current minimal example is sufficient
- **Pros**: Concise, developers can infer from type signatures
- **Cons**: May not be clear how to create repos/services/servers

**D)** PARTIAL - Add link to cipher-im as reference implementation
- **Pros**: Provides real example without bloating instructions
- **Cons**: Requires navigating to external file

**E)** Write-in (describe approach):

**YOUR ANSWER: __** C

---

## Section 2: ServiceResources Struct Documentation

### Q4: ServiceResources Field Usage Guidance

**Gap Identified**: Instructions list ServiceResources fields but don't explain when to use each.

**Question**: Should instructions add "When to Use" guidance for each ServiceResources field?

**A)** YES - Add table with field name, purpose, and typical usage scenarios
- **Pros**: Developers know which resources to use for specific tasks
- **Cons**: Adds significant length to instructions

**B)** YES - Add usage examples for most common resources only (DB, SessionManager, BarrierService)
- **Pros**: Covers 80% of use cases without being exhaustive
- **Cons**: Less common resources not documented

**C)** NO - Field names and types are self-documenting
- **Pros**: Concise instructions
- **Cons**: May not be obvious how to use BarrierService vs JWKGenService

**D)** PARTIAL - Add cross-references to other instruction files
- **Pros**: Avoids duplication, leverages existing docs
- **Cons**: Requires navigating multiple files

**E)** Write-in (describe approach):

**YOUR ANSWER: __** B; but not too verbose, keep instructions concise

---

### Q5: ShutdownCore vs ShutdownContainer

**Gap Identified**: Instructions don't explain difference between ShutdownCore and ShutdownContainer.

**Implementation Context**:
- `ShutdownCore`: Closes database connection, stops telemetry
- `ShutdownContainer`: Stops Docker test containers (only exists in tests)

**Question**: Should instructions document shutdown function differences?

**A)** YES - Add note explaining ShutdownCore (prod + test) vs ShutdownContainer (test only)
- **Pros**: Clarifies when to call each
- **Cons**: Implementation detail

**B)** NO - Usage obvious from context (both are shutdown functions)
- **Pros**: Concise
- **Cons**: May call wrong function or both unnecessarily

**C)** YES - Add shutdown pattern example (defer both in correct order)
- **Pros**: Shows correct usage
- **Cons**: Adds complexity

**D)** PARTIAL - Document in testing instructions only
- **Pros**: Focused on test-specific concerns
- **Cons**: Production code developers may not see it

**E)** Write-in (describe approach): Globally rename ShutdownCore to ShutdownCoreServices. Globally rename ShutdownContainer to ShutdownTestContainers. Then the names will be more obvious.

**YOUR ANSWER: __**  E

---

## Section 3: Migration Versioning Strategy

### Q6: Migration Numbering Collision Prevention

**Gap Identified**: Instructions say "Template 1001-1004, domain 2001+" but don't explain what happens if domain needs >999 migrations.

**Question**: Should instructions document migration number ranges and collision prevention?

**A)** YES - Add explicit number ranges (template 1001-1999, domain 2001-9999) with rationale
- **Pros**: Prevents future collisions
- **Cons**: Arbitrary limits

**B)** YES - Document that template migrations are FROZEN (no new template migrations added)
- **Pros**: Clarifies template is stable baseline
- **Cons**: May not be obvious to new developers

**C)** NO - Current guidance sufficient (1001-1004 vs 2001+ is clear)
- **Pros**: Simple documentation
- **Cons**: Doesn't prevent domain using 1005+

**D)** PARTIAL - Add validation in ApplyMigrations to reject domain migrations <2001
- **Pros**: Enforces numbering programmatically
- **Cons**: Adds runtime validation overhead

**E)** Write-in (describe approach):  Add explicit number ranges (template 1001-1999, domain 2001-2999) with rationale

**YOUR ANSWER: __** E

---

### Q7: Migration Naming Conventions

**Gap Identified**: Instructions don't show migration file naming pattern.

**Actual Pattern** (from template migrations):
- `1001_create_sessions_tables.up.sql`
- `1001_create_sessions_tables.down.sql`
- `1002_create_barrier_tables.up.sql`
- etc.

**Question**: Should instructions document migration file naming conventions?

**A)** YES - Add section with naming pattern and examples
- **Pros**: Clear guidance for creating new migrations
- **Cons**: Adds length

**B)** YES - Add link to golang-migrate documentation
- **Pros**: Leverages external docs
- **Cons**: Requires leaving instructions file

**C)** NO - golang-migrate conventions are standard
- **Pros**: Concise
- **Cons**: Developers may use wrong patterns

**D)** PARTIAL - Add example in Quick Reference only
- **Pros**: Visible without full section
- **Cons**: No detailed explanation

**E)** Write-in (describe approach):

**YOUR ANSWER: __** C

---

## Section 4: TLS Configuration Modes

### Q8: TLS Mode Documentation Completeness

**Gap Identified**: Instructions mention TLS but don't document three TLS modes (static, mixed, auto).

**Implementation Modes**:
- **static**: Pre-provided cert/key PEM (production)
- **mixed**: Generate from CA cert/key PEM (E2E tests)
- **auto**: Fully auto-generate (unit/integration tests)

**Question**: Should instructions document TLS modes?

**A)** YES - Add TLS modes section with use case matrix
- **Pros**: Developers understand when to use each mode
- **Cons**: Overlaps with 02-03.https-ports instructions

**B)** YES - Add cross-reference to 02-03.https-ports.instructions.md for TLS details
- **Pros**: Avoids duplication
- **Cons**: Requires navigating to other file

**C)** NO - TLS configuration is advanced topic, not needed in builder instructions
- **Pros**: Keeps focus on builder pattern
- **Cons**: Developers may not understand config fields

**D)** PARTIAL - Add Quick Reference one-liner only
- **Pros**: Minimal documentation overhead
- **Cons**: Not detailed enough for implementation

**E)** Write-in (describe approach):

**YOUR ANSWER: __** B; 02-03.https-ports.instructions.md should already cover those 3 options for TLS details

---

## Section 5: Error Handling in Builder Chain

### Q9: Fluent API Error Accumulation

**Gap Identified**: Instructions don't explain how builder accumulates errors during fluent chain.

**Implementation Pattern**:
```go
type ServerBuilder struct {
    err error  // Accumulates errors during fluent chain
}

func (b *ServerBuilder) WithDomainMigrations(...) *ServerBuilder {
    if b.err != nil {
        return b  // Short-circuit if previous error
    }
    // ...
}
```

**Question**: Should instructions document error accumulation pattern?

**A)** YES - Add section explaining short-circuit behavior and when errors surface (at Build())
- **Pros**: Developers understand error flow
- **Cons**: Implementation detail

**B)** YES - Add troubleshooting tip: "Errors from With*() methods only surface at Build()"
- **Pros**: Helps debugging
- **Cons**: May not be obvious this is by design

**C)** NO - Standard fluent builder pattern, no need to explain
- **Pros**: Concise
- **Cons**: Developers may expect errors immediately

**D)** PARTIAL - Add to Key Takeaways as bullet point
- **Pros**: Visible summary
- **Cons**: No detailed explanation

**E)** Write-in (describe approach):

**YOUR ANSWER: __** C

---

## Section 6: Test Compatibility Accessor Methods

### Q10: Accessor Method Completeness

**Gap Identified**: Instructions list some accessor methods but don't explain WHY they're needed or WHEN to add more.

**Context**: Tests need access to private fields (jwkGenService, telemetryService) for assertions.

**Question**: Should instructions add guidance on accessor method requirements?

**A)** YES - Add section: "Tests must access private services → add accessors for all private fields"
- **Pros**: Clear rule for when to add accessors
- **Cons**: May lead to over-exposure of internals

**B)** YES - Add rule: "Only expose what tests actually need, add more as needed"
- **Pros**: Minimal exposure principle
- **Cons**: May require multiple refactoring rounds

**C)** NO - Cipher-IM example is sufficient reference
- **Pros**: Developers can copy pattern
- **Cons**: May not understand why pattern exists

**D)** PARTIAL - Add link to 03-02.testing.instructions.md for test patterns
- **Pros**: Avoids duplication
- **Cons**: Cross-file navigation

**E)** Write-in (describe approach):

**YOUR ANSWER: __** B

---

### Q11: Accessor Method Naming Convention

**Gap Identified**: Instructions show accessor methods but don't explain naming pattern.

**Actual Pattern**:
```go
func (s *Server) JWKGen() *JWKGenService        // Drop "Service" suffix
func (s *Server) Telemetry() *TelemetryService  // Drop "Service" suffix
func (s *Server) PublicPort() int               // Match field exactly
```

**Question**: Should instructions document accessor naming conventions?

**A)** YES - Add rule: "Drop 'Service' suffix for accessors (JWKGenService → JWKGen)"
- **Pros**: Consistent naming across services
- **Cons**: Adds minor detail

**B)** NO - Naming obvious from examples
- **Pros**: Concise
- **Cons**: May lead to inconsistent naming

**C)** PARTIAL - Add to Quick Reference only
- **Pros**: Visible without full section
- **Cons**: No explanation of why

**D)** DEFER - Add when second service shows inconsistent naming
- **Pros**: YAGNI
- **Cons**: Preventable inconsistency

**E)** Write-in (describe approach):

**YOUR ANSWER: __** C

---

## Section 7: Merged Migrations Implementation Details

### Q12: mergedMigrations fs.FS Interface Methods

**Gap Identified**: Instructions mention fs.FS interface but don't explain implementation nuances.

**Implementation Detail**: mergedMigrations tries domain FS first (higher version numbers), then falls back to template FS.

**Question**: Should instructions document domain-first fallback pattern?

**A)** YES - Add note: "Domain migrations tried first (2001+ > 1001-1004)"
- **Pros**: Clarifies search order
- **Cons**: Implementation detail

**B)** YES - Add full mergedMigrations code example
- **Pros**: Complete reference
- **Cons**: Very long example (80+ lines)

**C)** NO - fs.FS interface is standard, no need to explain
- **Pros**: Concise
- **Cons**: Domain-first pattern not obvious

**D)** PARTIAL - Add link to server_builder.go implementation
- **Pros**: Reference without duplication
- **Cons**: External navigation

**E)** Write-in (describe approach): My assumption is service-template migrations are always applied first, followed by domain-specific migrations. Add note: "Service-template migrations (1001-1999) are applied before domain-specific migrations (2001-2999) to ensure base functionality is in place before applying domain-specific changes."

**YOUR ANSWER: __** E

---

## Section 8: Builder Usage Anti-Patterns

### Q13: Common Builder Mistakes

**Gap Identified**: Instructions don't document anti-patterns or common mistakes.

**Potential Anti-Patterns**:
- Calling Build() multiple times (creates multiple apps)
- Not checking Build() error
- Modifying ServiceResources after Build()
- Not calling Shutdown functions
- Mixing builder pattern with manual initialization

**Question**: Should instructions add anti-patterns section?

**A)** YES - Add "Common Mistakes to Avoid" section with examples
- **Pros**: Prevents known issues
- **Cons**: Adds significant length

**B)** YES - Add to troubleshooting section only
- **Pros**: Focused on problem-solving
- **Cons**: Reactive not proactive

**C)** NO - Mistakes will surface during code review
- **Pros**: Concise instructions
- **Cons**: Preventable issues

**D)** DEFER - Add after first real mistake occurs
- **Pros**: YAGNI
- **Cons**: May repeat known issues

**E)** Write-in (describe approach):

**YOUR ANSWER: __** B; keep concise, not verbose, to avoid wasting LLM tokens

---

## Section 9: Builder Extension Points

### Q14: Custom Initialization Hooks

**Gap Identified**: Instructions don't explain how to extend builder for service-specific initialization beyond public routes.

**Use Case**: Service needs custom initialization (background jobs, cron tasks, external service connections) not covered by public routes.

**Question**: Should instructions document extension patterns?

**A)** YES - Add "Advanced: Custom Initialization" section with callback examples
- **Pros**: Supports advanced use cases
- **Cons**: May encourage over-customization

**B)** YES - Add note: "For custom init, modify domain server's NewFromConfig() instead of builder"
- **Pros**: Clear separation of concerns
- **Cons**: May not be obvious where custom init belongs

**C)** NO - Builder handles common cases, custom cases are service-specific
- **Pros**: Focused builder scope
- **Cons**: No guidance for edge cases

**D)** PARTIAL - Add cross-reference to service-template.instructions.md for custom patterns
- **Pros**: Avoids duplication
- **Cons**: May not exist yet

**E)** Write-in (describe approach):

**YOUR ANSWER: __** D

---

## Section 10: Performance Characteristics

### Q15: Builder Initialization Performance

**Gap Identified**: Instructions don't mention builder initialization is SLOW (database, migrations, TLS generation).

**Context**: Builder startup takes 5-30 seconds depending on database type and migration count.

**Question**: Should instructions document performance characteristics?

**A)** YES - Add note: "Builder initialization 5-30s (database, migrations, TLS) - use TestMain for test suites"
- **Pros**: Sets expectations, guides test patterns
- **Cons**: Specific numbers may change

**B)** YES - Add cross-reference to 03-02.testing.instructions.md TestMain pattern
- **Pros**: Leverages existing guidance
- **Cons**: Doesn't explain why TestMain needed

**C)** NO - Performance is implementation detail
- **Pros**: Concise
- **Cons**: Developers may create slow tests unknowingly

**D)** PARTIAL - Add to troubleshooting for "tests are slow"
- **Pros**: Focused on problem-solving
- **Cons**: Reactive not proactive

**E)** Write-in (describe approach):

**YOUR ANSWER: __** B; also, make sure 03-02.testing.instructions.md includes note: "use TestMain for test suites, because service initialization is slow (5-30 seconds depending on database type and migration count)"

---

## Section 11: Database Connection Pool Configuration

### Q16: Connection Pool Defaults from Service-Template

**Gap Identified**: Instructions don't document that builder uses service-template connection pool defaults (NOT cipher-im).

**Context**: This was clarified in JOSE-JA-QUIZME Round 1 Q16 - connection pool settings come from service-template initialization, not cipher-im.

**Question**: Should instructions document connection pool configuration source?

**A)** YES - Add note: "Connection pool configured by service-template StartApplicationCore(), not domain services"
- **Pros**: Clarifies configuration source
- **Cons**: Implementation detail

**B)** YES - Add cross-reference to 03-05.sqlite-gorm.instructions.md for pool settings
- **Pros**: Leverages existing detailed docs
- **Cons**: Cross-file navigation

**C)** NO - Connection pool is database concern, not builder concern
- **Pros**: Focused builder scope
- **Cons**: May be unclear where pool is configured

**D)** PARTIAL - Add to ServiceResources.DB field description only
- **Pros**: Contextual documentation
- **Cons**: May be missed

**E)** Write-in (describe approach):

**YOUR ANSWER: __** B

---

## Section 12: Multi-Tenancy Migration Path

### Q17: Single-Tenant to Multi-Tenant Migration

**Gap Identified**: Instructions show both patterns but don't document migration path from single-tenant to multi-tenant.

**Context**:
- Single-tenant: Call `WithDefaultTenant(magic.DefaultTenantID, magic.DefaultRealmID)`
- Multi-tenant: Don't call WithDefaultTenant (Nil UUIDs skip creation)

**Question**: Should instructions document single→multi tenant migration path?

**A)** YES - Add section: "Migrating from single-tenant to multi-tenant requires data migration"
- **Pros**: Warns about breaking change
- **Cons**: May be premature (no service has migrated yet)

**B)** YES - Add note: "Single-tenant services should plan for future multi-tenant from start"
- **Pros**: Proactive architecture guidance
- **Cons**: May be YAGNI for some services

**C)** NO - Migration path is service-specific, not builder concern
- **Pros**: Focused builder scope
- **Cons**: Common question not addressed

**D)** DEFER - Add when first service actually migrates
- **Pros**: Real-world example
- **Cons**: Preventable issues for early services

**E)** Write-in (describe approach): All product-service are multi-tenant. There is no such thing as a single-tenant product-service. Fix the instructions to remove references to single-tenant product-service.

**YOUR ANSWER: __** E

---

## Section 13: Barrier Service Integration

### Q18: Barrier Service Usage Guidance

**Gap Identified**: Instructions mention BarrierService in ServiceResources but don't explain when/how to use it.

**Context**: BarrierService encrypts/decrypts private keys for storage (used by cipher-im for message private keys, jose-ja for JWK private||symmetric keys).

**Question**: Should instructions add BarrierService usage guidance?

**A)** YES - Add section: "Use BarrierService for all sensitive content storage" with encrypt/decrypt example
- **Pros**: Clear security pattern
- **Cons**: May overlap with cryptography instructions

**B)** YES - Add cross-reference to barrier service documentation
- **Pros**: Avoids duplication
- **Cons**: May not exist yet

**C)** NO - BarrierService usage is security concern, documented elsewhere
- **Pros**: Focused builder scope
- **Cons**: Developers may not know to use barrier

**D)** PARTIAL - Add one-liner in ServiceResources.BarrierService description
- **Pros**: Contextual hint
- **Cons**: No detailed guidance

**E)** Write-in (describe approach):

**YOUR ANSWER: __** A

---

## Section 14: SessionManager Integration

### Q19: SessionManager Usage Patterns

**Gap Identified**: Instructions mention SessionManager but don't show typical usage patterns.

**Cipher-IM Example**: SessionManager used for browser /browser/** paths, NOT for /service/** paths.

**Question**: Should instructions document SessionManager usage patterns?

**A)** YES - Add section: "SessionManager for /browser/** paths, token validation for /service/**"
- **Pros**: Clarifies path-based auth pattern
- **Cons**: May overlap with 02-03.https-ports instructions

**B)** YES - Add example showing session creation/validation in route handler
- **Pros**: Concrete usage pattern
- **Cons**: Adds significant length

**C)** NO - SessionManager is authentication concern, not builder concern
- **Pros**: Focused builder scope
- **Cons**: Developers may not understand how to use resource

**D)** PARTIAL - Add cross-reference to 02-10.authn.instructions.md
- **Pros**: Leverages existing auth docs
- **Cons**: Cross-file navigation

**E)** Write-in (describe approach): SessionManager is for all browser user AND service client sessions!!! FIX! Check if service-template and cipher-im and following this pattern. If not, it needs to be included in the plan and tasks for docs\jose-ja, because it is a blocker for all products-services!

**YOUR ANSWER: __** E

---

## Section 15: Comparison with Manual Initialization

### Q20: Before/After Builder Pattern

**Gap Identified**: Instructions claim "260+ lines eliminated" but don't show side-by-side comparison.

**Question**: Should instructions add before/after code comparison?

**A)** YES - Add appendix showing manual initialization vs builder (full 260-line comparison)
- **Pros**: Demonstrates value clearly
- **Cons**: Very long appendix

**B)** YES - Add simplified comparison (50 lines manual vs 10 lines builder)
- **Pros**: Shows pattern without full detail
- **Cons**: Not representative of real savings

**C)** NO - Line count claim sufficient, comparison not needed
- **Pros**: Concise
- **Cons**: Claim not proven

**D)** PARTIAL - Add link to cipher-im commit that adopted builder
- **Pros**: Real-world evidence
- **Cons**: External navigation, may not exist

**E)** Write-in (describe approach): side-by-side comparison is too verbose and not desirable to include in copilot instructions

**YOUR ANSWER: __**

---

## Summary

**Total Questions**: 20 across 15 sections
**Answer Format**: Write A, B, C, D, or E in **YOUR ANSWER: __** field
**Write-In Instructions**: If choosing E, provide detailed custom answer below question

**After Answering**:
1. Review answers for consistency
2. Update 03-08.server-builder.instructions.md based on decisions
3. Validate updates don't conflict with other instruction files

---

## Cross-References

- **Target File**: [03-08.server-builder.instructions.md](03-08.server-builder.instructions.md)
- **Implementation**: [server_builder.go](../../internal/apps/template/service/server/builder/server_builder.go)
- **Reference Service**: [cipher-im](../../internal/apps/cipher/im/)
- **Related Instructions**: [02-03.https-ports](02-03.https-ports.instructions.md), [03-02.testing](03-02.testing.instructions.md), [03-05.sqlite-gorm](03-05.sqlite-gorm.instructions.md)
