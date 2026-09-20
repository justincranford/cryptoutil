
---

## Appendix A: Decision Records

### A.1 Architectural Decision Records (ADRs)

**ADR Template**:

- Title: ADR-NNNN-descriptive-name
- Status: Proposed, Accepted, Deprecated, Superseded
- Context: Problem statement, constraints, requirements
- Decision: Chosen approach with rationale
- Consequences: Trade-offs, benefits, risks

**Location**: docs/adr/

### A.2 Technology Selection Decisions

**Go 1.26.1**: Static typing, fast compilation, excellent concurrency, CGO-free (portability)
**PostgreSQL + SQLite**: Production (ACID, scalability) + Dev/Test (zero-config, in-memory)
**GORM**: Cross-DB compatibility, migrations, type-safe queries
**Fiber**: Fast HTTP framework, Express-like API, low memory footprint
**OpenTelemetry**: Vendor-neutral observability, OTLP standard, future-proof

### A.3 Pattern Selection Decisions

**Service Template**: Eliminates 48,000+ lines per service, ensures consistency
**Dual HTTPS**: Security (public vs admin), network isolation, health checks
**Multi-Tenancy**: Schema-level isolation (not row-level), compliance, performance
**Hierarchical Keys**: Defense in depth, key rotation, compliance (FIPS 140-3)
