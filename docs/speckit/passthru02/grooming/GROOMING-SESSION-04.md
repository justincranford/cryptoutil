# Speckit Passthru02 Grooming Session 04: Infrastructure & Deployment Patterns

**Purpose**: Structured questions to refine infrastructure components, deployment strategies, CI/CD patterns, and operational concerns.
**Created**: 2025-12-02
**Status**: AWAITING ANSWERS

---

## Instructions

Select your answer by changing `[ ]` to `[x]` for each question. Add comments in the "Notes" field if needed. Multiple selections allowed where indicated.

---

## Section 1: Database Infrastructure (Q1-8)

### Q1. Database Backend Strategy

Current: PostgreSQL for production, SQLite for dev/test. Is this sufficient?

- [ ] A. Yes - dual backend is appropriate
- [ ] B. Remove SQLite - focus on PostgreSQL only
- [ ] C. SQLite only - simplify to embedded database
- [ ] D. Add more - CockroachDB, MySQL for enterprise customers

**Notes**:

---

### Q2. Database Per Product

Should each product have its own database?

- [ ] A. Single database for all products
- [ ] B. Per-product databases (jose_db, identity_db, kms_db, ca_db)
- [ ] C. Per-product schemas in single database
- [ ] D. Configurable - single or separate based on deployment

**Notes**:

---

### Q3. Migration Strategy

Current: golang-migrate with embedded SQL. Is this appropriate?

- [ ] A. Yes - keep golang-migrate
- [ ] B. Switch to GORM AutoMigrate for simplicity
- [ ] C. Use both - AutoMigrate for dev, golang-migrate for prod
- [ ] D. Atlas or other modern migration tool

**Notes**:

---

### Q4. Connection Pooling

What connection pool settings are appropriate?

- [ ] A. Defaults - let GORM/database/sql handle it
- [ ] B. Conservative (5-10 connections) - resource efficiency
- [ ] C. Aggressive (50-100 connections) - performance
- [ ] D. Per-product pools - different needs per product

**Notes**:

---

### Q5. Transaction Isolation

What transaction isolation level should be default?

- [ ] A. Read Committed - PostgreSQL default
- [ ] B. Repeatable Read - stronger consistency
- [ ] C. Serializable - strongest guarantees
- [ ] D. Per-operation - different isolation for different operations

**Notes**:

---

### Q6. Audit Trail Storage

Where should audit logs be stored?

- [ ] A. Same database as application data
- [ ] B. Separate audit database
- [ ] C. External logging system (Loki, ELK)
- [ ] D. Hybrid - summary in DB, details in logs

**Notes**:

---

### Q7. Backup Strategy

What backup strategy for databases?

- [ ] A. Database-native backups (pg_dump)
- [ ] B. Continuous replication
- [ ] C. Point-in-time recovery (WAL archiving)
- [ ] D. All of the above, deployment-dependent

**Notes**:

---

### Q8. SQLite Production Use

Should SQLite be supported in production for small deployments?

- [ ] A. Yes - single-node deployments
- [ ] B. No - PostgreSQL only in production
- [ ] C. Yes but with warnings and limitations documented
- [ ] D. Defer decision - not priority now

**Notes**:

---

## Section 2: Telemetry Infrastructure (Q9-16)

### Q9. Telemetry Stack Completeness

Current: OTLP → Collector → Grafana (Loki/Tempo/Prometheus). Is this complete?

- [ ] A. Yes - current stack is sufficient
- [ ] B. Add Prometheus direct export for some environments
- [ ] C. Add more dashboards/alerts
- [ ] D. Simplify - direct export to Grafana, skip collector

**Notes**:

---

### Q10. Trace Sampling Strategy

What trace sampling strategy?

- [ ] A. Sample everything (100%)
- [ ] B. Head-based sampling (configurable percentage)
- [ ] C. Tail-based sampling (keep interesting traces)
- [ ] D. Adaptive - based on load

**Notes**:

---

### Q11. Metrics Cardinality

How to handle metrics cardinality?

- [ ] A. No limits - collect everything
- [ ] B. Conservative labels - minimize cardinality
- [ ] C. Aggregation - reduce cardinality in collector
- [ ] D. Tiered - high cardinality in dev, low in prod

**Notes**:

---

### Q12. Log Levels

What log level strategy?

- [ ] A. Debug in dev, Info in prod
- [ ] B. Info everywhere, Debug via config
- [ ] C. Warn in prod to reduce volume
- [ ] D. Configurable per-component

**Notes**:

---

### Q13. Sensitive Data in Telemetry

How to handle sensitive data in logs/traces?

- [ ] A. Never log sensitive data
- [ ] B. Redact automatically (middleware)
- [ ] C. Hash/mask sensitive fields
- [ ] D. Separate sensitive and non-sensitive log streams

**Notes**:

---

### Q14. Alerting Strategy

What alerting infrastructure?

- [ ] A. Grafana alerts only
- [ ] B. Prometheus AlertManager
- [ ] C. External (PagerDuty, OpsGenie)
- [ ] D. Defer - not implementing alerts now

**Notes**:

---

### Q15. SLO/SLI Tracking

Should SLOs be tracked?

- [ ] A. Yes - define and track SLOs
- [ ] B. No - premature optimization
- [ ] C. Basic availability only
- [ ] D. After production deployment

**Notes**:

---

### Q16. Distributed Tracing Depth

How deep should distributed tracing go?

- [ ] A. HTTP handlers only
- [ ] B. HTTP + database operations
- [ ] C. HTTP + database + crypto operations
- [ ] D. Full stack including internal function calls

**Notes**:

---

## Section 3: Networking & Security (Q17-24)

### Q17. TLS Configuration

TLS 1.3+ is required. Any additional requirements?

- [ ] A. TLS 1.3 only - current requirement is sufficient
- [ ] B. Add cipher suite restrictions
- [ ] C. Add certificate pinning support
- [ ] D. Add mutual TLS (mTLS) everywhere

**Notes**:

---

### Q18. Rate Limiting Implementation

Where should rate limiting be implemented?

- [ ] A. Application level (Fiber middleware)
- [ ] B. API gateway (external)
- [ ] C. Both - defense in depth
- [ ] D. Database level (row-level)

**Notes**:

---

### Q19. Rate Limit Storage

Where should rate limit counters be stored?

- [ ] A. In-memory (per-instance)
- [ ] B. Redis (distributed)
- [ ] C. Database
- [ ] D. Configurable - memory for single instance, Redis for cluster

**Notes**:

---

### Q20. CORS Configuration

How should CORS be configured?

- [ ] A. Strict - explicit allowed origins only
- [ ] B. Permissive in dev, strict in prod
- [ ] C. Per-product CORS rules
- [ ] D. Configurable via config file

**Notes**:

---

### Q21. API Gateway

Should an API gateway be used?

- [ ] A. No - direct access to services
- [ ] B. Yes - for all external traffic
- [ ] C. Optional - support both patterns
- [ ] D. Only for multi-product deployments

**Notes**:

---

### Q22. Service Mesh

Should a service mesh be supported?

- [ ] A. No - overkill for this project
- [ ] B. Yes - Istio support
- [ ] C. Yes - Linkerd support (simpler)
- [ ] D. Defer - not priority now

**Notes**:

---

### Q23. Network Segmentation

How should network segmentation work?

- [ ] A. Single network for all services
- [ ] B. Separate networks per product
- [ ] C. Frontend/backend network split
- [ ] D. Full microsegmentation

**Notes**:

---

### Q24. Health Check Exposure

Which health endpoints should be exposed?

- [ ] A. Liveness only (/livez)
- [ ] B. Liveness + Readiness (/livez, /readyz)
- [ ] C. Above + Deep health (/healthz with deps)
- [ ] D. All above + metrics endpoint

**Notes**:

---

## Section 4: Container & Deployment (Q25-32)

### Q25. Container Base Image

What base image for containers?

- [ ] A. Alpine - smallest
- [ ] B. Distroless - more secure
- [ ] C. Ubuntu - easier debugging
- [ ] D. Scratch - smallest possible (Go static binary)

**Notes**:

---

### Q26. Container Security

What container security measures?

- [ ] A. Non-root user only
- [ ] B. Above + read-only filesystem
- [ ] C. Above + dropped capabilities
- [ ] D. Above + seccomp profiles

**Notes**:

---

### Q27. Resource Limits

Should resource limits be set?

- [ ] A. No limits - let orchestrator handle
- [ ] B. Conservative limits (256MB RAM, 0.5 CPU)
- [ ] C. Production limits (512MB-1GB RAM, 1-2 CPU)
- [ ] D. Per-product limits based on profiling

**Notes**:

---

### Q28. Docker Compose vs Kubernetes

Which orchestration should be primary?

- [ ] A. Docker Compose only
- [ ] B. Kubernetes only
- [ ] C. Both equally supported
- [ ] D. Docker Compose for dev, Kubernetes for prod

**Notes**:

---

### Q29. Helm Charts

Should Helm charts be provided?

- [ ] A. Yes - full Helm chart
- [ ] B. No - plain Kubernetes manifests
- [ ] C. Both Helm and plain manifests
- [ ] D. Defer - not priority now

**Notes**:

---

### Q30. Multi-Architecture

Should multi-arch images be built?

- [ ] A. amd64 only
- [ ] B. amd64 + arm64
- [ ] C. All supported Go architectures
- [ ] D. On-demand based on user requests

**Notes**:

---

### Q31. Image Registry

Where should images be published?

- [ ] A. GitHub Container Registry (ghcr.io) only
- [ ] B. Docker Hub only
- [ ] C. Both GHCR and Docker Hub
- [ ] D. Self-hosted registry option

**Notes**:

---

### Q32. Deployment Strategies

What deployment strategies should be documented?

- [ ] A. Single node only
- [ ] B. Single node + HA cluster
- [ ] C. Above + multi-region
- [ ] D. All of the above

**Notes**:

---

## Section 5: CI/CD & Operations (Q33-40)

### Q33. CI Workflow Completeness

Current workflows: quality, coverage, benchmark, fuzz, race, sast, gitleaks, dast, e2e, load. Is this complete?

- [ ] A. Yes - comprehensive CI
- [ ] B. Add dependency scanning
- [ ] C. Add container scanning
- [ ] D. Simplify - too many workflows

**Notes**:

---

### Q34. Release Strategy

How should releases be handled?

- [ ] A. Manual releases with semantic versioning
- [ ] B. Automated releases on tag push
- [ ] C. Continuous deployment to staging
- [ ] D. GitOps with ArgoCD/Flux

**Notes**:

---

### Q35. Changelog Generation

How should changelog be maintained?

- [ ] A. Manual CHANGELOG.md
- [ ] B. Auto-generated from conventional commits
- [ ] C. GitHub releases only
- [ ] D. Release-please or similar automation

**Notes**:

---

### Q36. Security Scanning Frequency

How often should security scans run?

- [ ] A. On every PR
- [ ] B. Daily scheduled
- [ ] C. Weekly scheduled
- [ ] D. Multiple: PR + daily + before release

**Notes**:

---

### Q37. Dependency Updates

How should dependencies be updated?

- [ ] A. Manual updates only
- [ ] B. Dependabot with auto-merge for minor
- [ ] C. Renovate with grouping
- [ ] D. Manual review of all updates

**Notes**:

---

### Q38. Documentation Deployment

Should documentation be auto-deployed?

- [ ] A. No - README.md in repo is sufficient
- [ ] B. Yes - GitHub Pages
- [ ] C. Yes - dedicated docs site
- [ ] D. Defer - not priority now

**Notes**:

---

### Q39. Runbook Automation

Should runbooks be automated?

- [ ] A. Manual runbooks only
- [ ] B. Scripted procedures
- [ ] C. Full automation (Ansible, Terraform)
- [ ] D. Hybrid - documented with automation helpers

**Notes**:

---

### Q40. Incident Response

What incident response tooling?

- [ ] A. None - manual response
- [ ] B. PagerDuty/OpsGenie integration
- [ ] C. Slack/Teams notifications
- [ ] D. Defer - not priority now

**Notes**:

---

## Summary & Next Steps

After completing this grooming session:

1. Review answers for consistency
2. Identify infrastructure priorities
3. Update deployment documentation
4. Plan infrastructure improvements
5. Share answers with Copilot for implementation guidance

---

*Session Created: 2025-12-02*
*Expected Completion: [DATE]*
