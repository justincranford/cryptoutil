# Passthru2: Implementation Task List

**Purpose**: Tasks for applying improvements, demo parity, and developer experience improvements after passthru1
**Created**: 2025-11-30

---

## Phase 0: Developer Experience (Day 1)

- [ ] Create `make demo` or single `docker compose` demo script per product
- [ ] Standardize `deployments/` structure (telemetry extraction already planned)
- [ ] Add demo seed data for KMS and Identity (admin, user, service account)
- [ ] Add `demo` Docker Compose profile for each product

## Phase 1: KMS Demo Parity (Day 2-3)

- [ ] Ensure KMS demo has pre-seeded demo accounts and demo scripts like Identity
- [ ] Confirm KMS Swagger UI supports "Try it out" demo flow with demo accounts
- [ ] Add browser-friendly demo steps and UI simulations
- [ ] Add KMS coverage improvements (handler & businesslogic tests)

## Phase 2: Identity Demo Fixes & Parity (Day 3-5)

- [ ] Implement missing endpoints: /authorize, PKCE, redirect handling
- [ ] Seed demo user accounts and clients
- [ ] Fix refresh token rotation and introspection coverage gaps
- [ ] Add Identity demo scripts & Docker Compose profile

## Phase 3: Integration (Day 5-7)

- [ ] Implement token validation middleware in KMS and KMS scope enforcement
- [ ] Add embedded Identity mode for KMS and demo it
- [ ] Add black box E2E tests for KMS + Identity using Docker Compose

## Phase 4: Improvements & Cleanup (Day 7+)

- [ ] Standardize logging/telemetry in infra
- [ ] Standardize secrets using Docker secrets across all product Compose files
- [ ] Create a `demo.md` quickstart and troubleshooting section
- [ ] Add pre-commit hooks to ensure lint & test before commit

---

**Status**: WIP

