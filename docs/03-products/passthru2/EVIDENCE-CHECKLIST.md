# Passthru2 Evidence & Completion Checklist

**Updated**: 2025-11-30 (aligned with Grooming Session 1 decisions)

Use this checklist to validate that tasks are complete and that acceptance criteria are met.

---

## Phase 0: Developer Experience Foundation

- [ ] `deployments/telemetry/compose.yml` created and working (Q6)
- [ ] `deployments/<product>/config/` structure standardized (Q7)
- [ ] All secrets converted to Docker secrets (Q7, Q10)
- [ ] Compose profiles `dev`, `demo`, `ci` created per product (Q8)

---

## Demo & DX (from Q25)

- [ ] **A**: `docker compose` starts KMS with seeded accounts and data
- [ ] **A**: `docker compose` starts Identity with seeded users and clients
- [ ] **B**: KMS Swagger UI "Try it out" works with demo credentials
- [ ] **B**: Identity discovery endpoints functional
- [ ] Both demos run with `demo` profile and health checks pass

---

## Integration (from Q25)

- [ ] **C**: Integration demo starts both services with shared telemetry
- [ ] **C**: KMS accepts tokens from Identity
- [ ] **C**: KMS enforces scopes correctly (hybrid model - Q18)
- [ ] Token validation uses mixed approach (local + introspection - Q17)

---

## CI & Tests (from Q21, Q24)

- [ ] **D**: Per-product `go test` runs with coverage ≥ 80% minimum
- [ ] Coverage threshold enforced in CI
- [ ] Demo profile CI jobs added for KMS and Identity
- [ ] SQLite and PostgreSQL matrix runs in CI (Q20)
- [ ] Lint and formatting checks pass (`golangci-lint run --fix`)

---

## Security (from Q16)

- [ ] Password & client secret hashing uses PBKDF2 (NO bcrypt!)
- [ ] PKCE enforced for public clients
- [ ] TLS endpoints use valid certs or `localhost` debug certs
- [ ] Docker secrets in place for all credentials

---

## Infrastructure (from Q6, Q7, Q10)

- [ ] **E**: Telemetry extracted into `deployments/telemetry/compose.yml`
- [ ] **E**: Secrets standardized to Docker secrets across all products
- [ ] Config locations standardized under `deployments/<product>/config/`
- [ ] Empty directories removed

---

## Documentation

- [ ] `docs/03-products/passthru2/README.md` includes decision summary
- [ ] `docs/03-products/passthru2/TASK-LIST.md` updated with phases
- [ ] `grooming/GROOMING-SESSION-1.md` answered and committed
- [ ] `grooming/GROOMING-SESSION-2.md` prepared

---

## Final Sign-off Criteria (Q25 - ALL must be true)

- [ ] **A**: KMS and Identity demos both start with `docker compose` and include seeded data
- [ ] **B**: KMS and Identity both have interactive demo scripts and Swagger UI usable with demo creds
- [ ] **C**: Integration demo runs and validates token-based auth and scopes
- [ ] **D**: All product tests pass with coverage targets achieved (80%+)
- [ ] **E**: Telemetry extracted to shared compose and secrets standardized

**Sign-off**: Commit reviewer, tests green, and coverage pass

