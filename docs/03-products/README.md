# Products and Infrastructure Vision

**Purpose**: Organize cryptoutil repository into reusable infrastructure components and deployable products
**Created**: November 26, 2025
**Status**: PLANNING

---

## Active Project: Passthru1

**See [passthru1/README.md](passthru1/README.md) for current implementation work.**

Passthru1 focuses on getting 3 working demos:

1. **KMS Demo** - Protect existing manual implementation
2. **Identity Demo** - Fix LLM-generated code to working state
3. **Integration Demo** - KMS authenticated by Identity

Timeline: 1-2 weeks aggressive

---

## Vision Overview

### Current Reality

The cryptoutil repository contains valuable code organized by technical layer (internal/client, internal/server, internal/identity) rather than by reusable infrastructure components and deployable products. This makes it difficult to:

- Reuse common infrastructure across multiple products
- Deploy products independently with different combinations of components
- Scale the codebase as new products are added
- Understand what can be deployed vs what is shared infrastructure

### Future Vision

Reorganize the repository into two clear categories:

1. **Infrastructure Components** (`internal/infra/*`): Reusable building blocks used across multiple products
2. **Products** (`internal/product/*`): Deployable services/applications built from infrastructure components

### Benefits

- **Reusability**: Infrastructure components can be shared across products
- **Maintainability**: Clear boundaries between infrastructure and products
- **Deployment Flexibility**: Products can be deployed standalone or coordinated
- **Development Velocity**: New products can leverage existing infrastructure
- **Testing Clarity**: Infrastructure tests vs product integration tests

---

## Infrastructure Components

### Existing Infrastructure (to be refactored/organized)

| Component | Current Location | Target Location | Description |
|-----------|------------------|-----------------|-------------|
| **I1. Configuration** | internal/*/config/* | internal/infra/configuration | Config files, env vars, secrets, feature flags, validation |
| **I2. Networking** | internal/server/* | internal/infra/networking | HTTP, HTTPS, gRPC, REST, load balancing, firewalls |
| **I3. Testing** | internal/*/test/* | internal/infra/testing | Unit, integration, E2E, mocking, coverage |
| **I4. Performance** | test/load/* | internal/infra/performance | Load testing, profiling, caching, benchmarking |
| **I5. Telemetry** | internal/common/telemetry/* | internal/infra/telemetry | Logging, metrics, tracing, monitoring, Otel, Grafana |
| **I6. Crypto** | internal/common/crypto/* | internal/infra/crypto | Key gen, pools, encrypt/decrypt, sign/verify, HSM |
| **I7. Database** | internal/*/repository/* | internal/infra/database | SQLite, PostgreSQL, migrations, ORM, caching |
| **I8. Containers** | deployments/* | internal/infra/containers | PostgreSQL, Otel, Grafana, Redis, MongoDB configs |
| **I9. Deployment** | .github/workflows/* | internal/infra/deployment | CI/CD, containerization, orchestration, IaC |

### Future Infrastructure (not yet implemented)

| Component | Target Location | Description | Priority |
|-----------|-----------------|-------------|----------|
| **I10. Messaging** | internal/infra/messaging | Message queues, pub/sub, event streaming | MEDIUM |
| **I11. Auditing** | internal/infra/auditing | Immutable logs, log signing, compliance | HIGH |
| **I12. Documentation** | internal/infra/documentation | API docs, guides, tutorials, changelogs | MEDIUM |
| **I13. Internationalization** | internal/infra/internationalization | Localization, translation, formatting | LOW |
| **I14. Search** | internal/infra/search | Full-text search, indexing, faceted search | LOW |
| **I15. Dev Tools** | internal/infra/devtools | Debugging, profiling, code analysis, build tools | MEDIUM |
| **I16. Security** | internal/infra/security | Vulnerability scanning, threat modeling, audits | HIGH |

---

## Products

### Existing Products (to be organized)

| Product | Current Location | Target Location | Description |
|---------|------------------|-----------------|-------------|
| **P1. JOSE** | internal/common/crypto/jose/* | internal/product/jose | JWK, JWKS, JWE, JWS, JWT, OAuth2.1, OIDC1.0, SSO |
| **P2. Identity** | internal/identity/* | internal/product/identity | Users, groups, roles, auth, authz, MFA, FIDO2, WebAuthn |
| **P3. KMS** | internal/server/*, internal/client/* | internal/product/kms | Key management, rotation, policies, access control, HSM |

### Future Products (new capabilities)

| Product | Target Location | Description | Dependencies | Priority |
|---------|-----------------|-------------|--------------|----------|
| **P4. Certificates** | internal/product/certificates | X.509, CSR, issuance, OCSP, CRL, TLS, PKI, ACME | I6, I7, I11 | HIGH |

### Product Architecture Pattern

Each product should follow this structure:

```
internal/product/<product-name>/
├── cmd/                    # CLI entry points (main.go for server/client)
│   ├── server/            # Product server command
│   └── client/            # Product client command (if applicable)
├── api/                    # OpenAPI specs, generated clients/servers
│   ├── openapi.yaml
│   ├── client/            # Generated client code
│   └── server/            # Generated server code
├── domain/                 # Domain models, business logic
│   ├── models.go
│   └── services.go
├── repository/             # Data access layer
│   ├── orm/               # GORM repositories
│   └── sql/               # Direct SQL repositories
├── handlers/               # HTTP/gRPC handlers
│   ├── http/
│   └── grpc/              # Optional
├── config/                 # Product-specific configuration
│   └── config.go
├── embedded/               # Embedded service libraries (for use in other products)
│   ├── service.go
│   └── client.go
└── README.md               # Product documentation
```

---

## Migration Strategy

### Phase 1: Infrastructure Extraction (Weeks 1-4)

**Goal**: Move common infrastructure to `internal/infra/*` without breaking existing code

**Steps**:

1. Create `internal/infra/` directory structure
2. Move shared components (telemetry, crypto, database) to infra packages
3. Update import paths in existing code to use infra packages
4. Verify all tests still pass after refactoring
5. Update documentation to reflect new structure

**Risk Mitigation**:

- Work incrementally (one infrastructure component at a time)
- Run full test suite after each component move
- Keep git history clean with descriptive commit messages

### Phase 2: Product Organization (Weeks 5-8)

**Goal**: Reorganize existing products into `internal/product/*` structure

**Steps**:

1. Create `internal/product/jose/` from scattered JOSE code
2. Consolidate `internal/identity/` → `internal/product/identity/`
3. Consolidate `internal/server/*` + `internal/client/*` → `internal/product/kms/`
4. Create embedded service libraries for cross-product use
5. Update `cmd/` entry points to use product packages

**Risk Mitigation**:

- Test each product independently after reorganization
- Maintain backward compatibility for external APIs
- Document breaking changes in migration guide

### Phase 3: New Product Development (Weeks 9-12)

**Goal**: Build P4 (Certificates) as proof-of-concept for new architecture

**Steps**:

1. Design certificates product architecture (following product pattern)
2. Leverage existing I6 (crypto), I7 (database), I11 (auditing) infrastructure
3. Implement X.509 certificate operations (CSR, issuance, revocation)
4. Add ACME protocol support for automated certificate management
5. Create embedded service for certificate operations in other products

**Success Metrics**:

- P4 built using ONLY infrastructure components (no duplicate code)
- P4 deployable standalone or embedded in other products
- P4 demonstrates reusability benefits of new architecture

### Phase 4: Continuous Improvement (Weeks 13+)

**Goal**: Refine architecture based on learnings, add missing infrastructure

**Ongoing Activities**:

- Extract additional infrastructure components as patterns emerge
- Build new products (P5, P6, etc.) using established patterns
- Improve cross-product integration and testing
- Enhance documentation and examples
- Gather feedback and iterate on architecture

---

## Documentation Organization

### Planning Documents Structure

```
docs/03-products/
├── README.md                           # This file
├── VISION.md                           # Long-term vision and roadmap
├── MIGRATION-GUIDE.md                  # Step-by-step migration guide
├── ARCHITECTURE.md                     # Architecture decision records
├── infrastructure/                     # Infrastructure component docs
│   ├── I01-configuration.md
│   ├── I02-networking.md
│   ├── I03-testing.md
│   ├── I04-performance.md
│   ├── I05-telemetry.md
│   ├── I06-crypto.md
│   ├── I07-database.md
│   ├── I08-containers.md
│   ├── I09-deployment.md
│   ├── I10-messaging.md              # Future
│   ├── I11-auditing.md               # Future
│   ├── I12-documentation.md          # Future
│   ├── I13-internationalization.md   # Future
│   ├── I14-search.md                 # Future
│   ├── I15-devtools.md               # Future
│   └── I16-security.md               # Future
└── products/                           # Product planning docs
    ├── P01-jose.md
    ├── P02-identity.md
    ├── P03-kms.md
    └── P04-certificates.md           # Future
```

### Documentation Standards

**Infrastructure Component Docs** (I##-*.md):

- Component overview and purpose
- API surface (packages, interfaces, functions)
- Usage examples and patterns
- Configuration options
- Testing strategies
- Integration with other components
- Performance considerations

**Product Planning Docs** (P##-*.md):

- Product overview and use cases
- Architecture (what infrastructure components are used)
- API design (endpoints, schemas)
- Deployment options (standalone, embedded, coordinated)
- Dependencies on other products
- Migration path (if refactoring existing code)
- Testing and quality assurance

---

## LLM Agent Guidance

### When to Use This Structure

**For Infrastructure Work**:

- Adding new infrastructure components (I10-I16)
- Refactoring existing code into infrastructure packages
- Improving shared utilities and libraries
- Enhancing cross-cutting concerns (logging, metrics, security)

**For Product Work**:

- Building new products (P4 Certificates, P5, etc.)
- Enhancing existing products (P1 JOSE, P2 Identity, P3 KMS)
- Creating embedded service libraries
- Product-specific features and APIs

### How to Analyze and Plan

**Step 1: Understand Current State**

- Read relevant infrastructure component docs (I##-*.md)
- Review existing product docs (P##-*.md)
- Examine current code structure
- Identify gaps and opportunities

**Step 2: Plan Changes**

- Determine if change is infrastructure or product work
- Identify affected infrastructure components
- Plan minimal set of changes to achieve goal
- Consider impact on other products/components

**Step 3: Execute Incrementally**

- Work on one infrastructure component or product at a time
- Run tests after each incremental change
- Update documentation as you go
- Create ADRs for significant architectural decisions

**Step 4: Validate and Document**

- Verify all tests passing
- Update README and component/product docs
- Create migration guide if breaking changes
- Document lessons learned

### Best Practices for LLM Agents

1. **Read Before Writing**: Always read component/product docs before making changes
2. **Test Incrementally**: Run tests after each logical unit of work
3. **Document Decisions**: Create ADRs for non-trivial architectural choices
4. **Maintain Backwards Compatibility**: Avoid breaking changes when possible
5. **Update Docs**: Keep README and component docs in sync with code
6. **Ask for Clarification**: If requirements unclear, document assumptions and proceed

---

## Success Metrics

### Infrastructure Component Quality

- **Reusability**: Used by 2+ products
- **Test Coverage**: ≥95% for infrastructure components (higher than products)
- **Documentation**: Complete API docs, usage examples, integration guides
- **Performance**: Benchmarks show acceptable performance
- **Maintainability**: Low cyclomatic complexity, clear interfaces

### Product Quality

- **Deployability**: Can deploy standalone or embedded
- **Test Coverage**: ≥85% for product code (integration + unit)
- **Documentation**: Complete user guides, API docs, runbooks
- **Dependencies**: Uses ONLY infrastructure components (no duplicate code)
- **Integration**: Works correctly when coordinated with other products

### Overall Repository Health

- **Structure Clarity**: Clear separation between infra and products
- **Code Duplication**: <5% duplication across codebase
- **Build Time**: <5 minutes for full build and test suite
- **Onboarding**: New developers productive within 1 week
- **Maintenance**: Issues resolved within 1 week on average

---

## Next Steps

1. **Review and Feedback**: Get stakeholder input on vision and structure
2. **Create Detailed Plans**: Develop detailed migration plans for Phase 1-4
3. **Prioritize Infrastructure**: Identify highest-value infrastructure components to extract first
4. **Prototype P4**: Build certificates product as architecture proof-of-concept
5. **Iterate and Improve**: Refine based on learnings and feedback

---

**Status**: PLANNING
**Next Review**: After infrastructure component docs (I01-I09) are created
**Stakeholders**: Development team, product owners, operations team
