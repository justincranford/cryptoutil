# DEMO-INTEGRATION: KMS + Identity Integrated Demo (Passthru2)

**Purpose**: Ensure KMS and Identity integrate cleanly and have parity in demo UX
**Priority**: HIGH
**Timeline**: Day 5-7

---

## Differences vs Passthru1

- Extract telemetry to shared compose to avoid coupling
- Implement per-product `demo` mode with curated seed data
- Embed Identity option still supported but encourage external Identity by default

---

## Key Tasks

- Implement token validation middleware with caching for KMS
- Implement comprehensive scope enforcement and tests
- Provide embedded and external Identity modes with clear documented CLI flags
- Add integration demo script with step-by-step checks

---

## Success Criteria

- [ ] `docker compose up` for integration demo starts both services and telemetry
- [ ] KMS accepts tokens from Identity and enforces scopes correctly
- [ ] Integration demo script validates all flows (admin token, service token, user token)
- [ ] Documentation updated and E2E tests implemented per product

---

**Status**: WIP
