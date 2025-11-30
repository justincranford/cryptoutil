# DEMO-KMS: KMS-Only Working Demo (Passthru2)

**Purpose**: KMS demo parity and improved DX
**Priority**: HIGHEST
**Timeline**: Day 1-2

---

## Differences vs Passthru1

- Add demo accounts, pre-seeding and a `demo` compose profile
- Ensure Swagger UI uses seeded identities for "Try it out" flows
- Add browser and service demo scripts to match Identity experience

---

## Key Tasks

- Create pre-seeded demo accounts and keys as part of `deployments/kms/demo` profile
- Add a `demo` flag to the KMS server that seeds everything on first run
- Add `make demo`, `make demo-kms` CLI convenience commands
- Add end-to-end interactive script (same UX as Identity) and more sample requests

---

## Success Criteria

- [ ] Docker compose `demo` starts KMS with pre-seeded accounts
- [ ] Demo script executes full flow (login → create pool → create key → encrypt → decrypt) automatically
- [ ] Swagger UI interactive calls work with demo accounts
- [ ] Documentation is updated with quickstart and troubleshooting

---

**Status**: WIP
