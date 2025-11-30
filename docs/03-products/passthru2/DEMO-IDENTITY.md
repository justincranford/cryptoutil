# DEMO-IDENTITY: Identity-Only Working Demo (Passthru2)

**Purpose**: Stabilize Identity demo and provide feature parity for KMS demo experiences
**Priority**: HIGH
**Timeline**: Day 2-5

---

## Differences vs Passthru1

- Prioritize demo seeding and a `demo` mode for Identity
- Fix missing flows (authorize, PKCE, refresh rotation)
- Seed clients and users with the same standard used in KMS demo

---

## Key Tasks

- Add `--demo` mode to identity server for auto seeding demo data
- Add robust PKCE handling and validations
- Implement and test token revocation and refresh token rotation
- Add identity demo script, similar to KMS, with sample calls and Compose profile

---

## Success Criteria

- [ ] Docker compose `demo` starts Identity with seeded users and clients
- [ ] Discovery endpoint returns valid config and JWKS
- [ ] Authorization code + PKCE + token exchange flow works end-to-end
- [ ] Token introspection and revocation validated with demo scripts

---

**Status**: WIP
