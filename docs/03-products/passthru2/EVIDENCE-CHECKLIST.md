# Passthru2 Evidence & Completion Checklist

**Updated**: 2025-11-30 (aligned with Grooming Sessions 1 & 2 decisions)

Use this checklist to validate that tasks are complete and that acceptance criteria are met.

---

## CRITICAL: TLS/HTTPS Fix (from Q20)

- [ ] Identity reuses KMS cert utility functions
- [ ] CA-chained certs used (never self-signed leaf certs)
- [ ] Consistent HTTPS across all services
- [ ] Config options for cert chain lengths and TLS parameters

---

## Phase 0: Developer Experience Foundation

- [ ] `deployments/telemetry/compose.yml` created and working (Q6)
- [ ] `deployments/<product>/config/` structure standardized (Q7)
- [ ] All secrets converted to Docker secrets (Q7, Q10)
- [ ] Compose profiles `dev`, `demo`, `ci` created per product (Q8)
- [ ] `--reset-demo` flag implemented for data cleanup (Q15)

---

## Demo & DX (from Q25)

- [ ] **A**: `docker compose` starts KMS with seeded accounts and data
- [ ] **A**: `docker compose` starts Identity with seeded users and clients
- [ ] **B**: KMS Swagger UI "Try it out" works with demo credentials
- [ ] **B**: Identity discovery endpoints functional
- [ ] Both demos run with `demo` profile and health checks pass
- [ ] Real functional keys created in demo (not placeholders) (Q11)
- [ ] Profile-based persistence: dev=persist, ci=ephemeral (Q12)
- [ ] Predictable passwords documented with warnings (Q13)
- [ ] Client secrets in Docker secrets even for demo (Q14)

---

## Token Validation (from Q6-10)

- [ ] In-memory JWKS caching with configurable TTL (Q6)
- [ ] Revocation check frequency configurable (Q7)
- [ ] 401 for auth issues, 403 for scope issues (Q8)
- [ ] Configurable error detail level (Q8)
- [ ] Service-to-service auth supports client-creds/mTLS/API-key (Q9)
- [ ] All OIDC + custom claims extracted (Q10)

---

## Integration (from Q25)

- [ ] **C**: Integration demo starts both services with shared telemetry
- [ ] **C**: KMS accepts tokens from Identity
- [ ] **C**: KMS enforces scopes correctly (hybrid model - Q18)
- [ ] Token validation uses configurable approach (local + introspection)
- [ ] Full dependency chain health checks (DB + Identity + Telemetry) (Q16)
- [ ] Per-product + shared telemetry networks (Q17)

---

## KMS Realm (from Q1-5)

- [ ] Separate `realms.yml` config file (Q1)
- [ ] PBKDF2 + plaintext password support (Q2)
- [ ] `kms_realm_users` table for DB realm (Q3)
- [ ] Configurable realm priority order (Q4)
- [ ] Database-level tenant isolation (Q5)

---

## CI & Tests (from Q21-25)

- [ ] **D**: Per-product `go test` runs with coverage ≥ 80% minimum
- [ ] Coverage threshold enforced in CI
- [ ] Demo profile CI jobs added for KMS and Identity
- [ ] SQLite and PostgreSQL matrix runs in CI (Q20)
- [ ] Lint and formatting checks pass (`golangci-lint run --fix`)
- [ ] All tests use UUIDv7 unique prefixes (Q23 - CRITICAL)
- [ ] Basic benchmarks for critical paths (Q24)
- [ ] Test case descriptions in code (Q25)
- [ ] Integration tests: startup + CRUD + full flow (Q21)
- [ ] Compose for local/CI, Testcontainers for unit/integration (Q22)

---

## Security (from Q16)

- [ ] Password & client secret hashing uses PBKDF2 (NO bcrypt!)
- [ ] PKCE enforced for public clients
- [ ] TLS endpoints use CA-chained certs (not self-signed)
- [ ] Docker secrets in place for all credentials

---

## Infrastructure (from Q6, Q7, Q10)

- [ ] **E**: Telemetry extracted into `deployments/telemetry/compose.yml`
- [ ] **E**: Secrets standardized to Docker secrets across all products
- [ ] Config locations standardized under `deployments/<product>/config/`
- [ ] Empty directories removed
- [ ] Named volumes for all persistent data (Q18)
- [ ] Product-specific ports: KMS=8081, Identity=8082 (Q19)

---

## Documentation

- [ ] `docs/03-products/passthru2/README.md` includes decision summary
- [ ] `docs/03-products/passthru2/TASK-LIST.md` updated with phases
- [ ] `grooming/GROOMING-SESSION-1.md` answered and committed
- [ ] `grooming/GROOMING-SESSION-2.md` answered and committed
- [ ] `grooming/GROOMING-SESSION-3.md` answered

---

## Final Sign-off Criteria (Q25 - ALL must be true)

- [ ] **A**: KMS and Identity demos both start with `docker compose` and include seeded data
- [ ] **B**: KMS and Identity both have interactive demo scripts and Swagger UI usable with demo creds
- [ ] **C**: Integration demo runs and validates token-based auth and scopes
- [ ] **D**: All product tests pass with coverage targets achieved (80%+)
- [ ] **E**: Telemetry extracted to shared compose and secrets standardized
- [ ] **F**: TLS/HTTPS pattern fixed - Identity uses KMS CA-chained cert utilities (CRITICAL)

**Sign-off**: Commit reviewer, tests green, and coverage pass
