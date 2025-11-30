# Passthru1: Aggressive Demo Implementation

**Purpose**: Refactor and stabilize cryptoutil into 3 working self-guided demos
**Created**: November 29, 2025
**Status**: ACTIVE IMPLEMENTATION
**Timeline**: 1-2 weeks (aggressive)

---

## Project Overview

Passthru1 is the first implementation pass focused on delivering **working demos** that can be explored interactively. This project does NOT refactor the internal architecture (no infra/product reorganization yet) - it focuses purely on making demos work.

### Three Demo Targets

| Demo | Status | Source | Goal |
|------|--------|--------|------|
| **KMS Demo** | Existing | Manual implementation | Refactor without breaking |
| **Identity Demo** | Partial | 6 LLM passthrus | Fix to working state |
| **Integration Demo** | New | KMS + Identity | Ultimate integration goal |

---

## Critical Rules

### KMS Demo (Don't Break It)
- **PROTECT** all manually-created KMS code
- **REFACTOR** only to improve demo experience
- **NEVER** remove working functionality
- Test after every change

### Identity Demo (Make It Work)
- **AUDIT** all 6 LLM passthrus for broken code
- **FIX** one component at a time
- **TEST** each component before moving on
- Goal: Working OAuth2.1 demo flow

### Integration Demo (Build Incrementally)
- **ONLY START** after KMS and Identity demos work
- **USE** Identity to authenticate KMS access
- **DEMONSTRATE** embedded vs standalone patterns

---

## Demo Experience Requirements

From Grooming Session 4:

| Requirement | Implementation |
|-------------|----------------|
| Starting Point | `docker compose up -d` → Open http://localhost:8080 |
| Demo Accounts | Admin + Regular User + Service Account |
| Demo Keys | Multiple hierarchies (root, intermediate, leaf) |
| Walkthrough | Login → Logout → View → Generate → Use → Audit |
| Reset | Docker volumes + Admin UI button |
| Duration | 2-3 minutes quick overview |
| Audience | Self (personal portfolio) |

---

## File Organization

```
docs/03-products/passthru1/
├── README.md                    # This file - project overview
├── TASK-LIST.md                 # Aggressive implementation task list
├── DEMO-KMS.md                  # KMS-only demo implementation plan
├── DEMO-IDENTITY.md             # Identity-only demo implementation plan
├── DEMO-INTEGRATION.md          # KMS+Identity integration demo plan
├── grooming/                    # Grooming session archives
│   ├── GROOMING-ANALYSIS.md
│   ├── GROOMING-ANALYSIS-2.md
│   ├── GROOMING-ANALYSIS-3.md
│   ├── GROOMING-QUESTIONS.md
│   ├── GROOMING-SESSION-2.md
│   ├── GROOMING-SESSION-3.md
│   └── GROOMING-SESSION-4.md
└── progress/                    # Progress tracking (created as needed)
```

---

## Success Criteria

### KMS Demo Success (Updated 2025-11-30)

- [x] `docker compose up -d` starts KMS service
- [x] Swagger UI accessible at `/ui/swagger`
- [x] Browser API works (CORS, CSRF)
- [ ] Can create key pools and keys via UI (not yet tested via UI)
- [x] Can encrypt/decrypt via API (verified working)
- [x] Can sign/verify via API (verified working)
- [ ] Demo accounts pre-seeded

### Identity Demo Success
- [ ] Identity server starts without errors
- [ ] OAuth2.1 authorization flow works
- [ ] Token endpoint returns valid tokens
- [ ] Token introspection works
- [ ] Token revocation works
- [ ] User login flow works
- [ ] Session management works

### Integration Demo Success
- [ ] KMS uses Identity for authentication
- [ ] OAuth2 scopes control KMS access
- [ ] Single sign-on works across services
- [ ] Embedded Identity option works
- [ ] Standalone Identity option works

---

## Next Steps

1. **Read TASK-LIST.md** for prioritized implementation order
2. **Start with KMS Demo** (protect existing work)
3. **Then Identity Demo** (fix LLM-generated code)
4. **Finally Integration** (combine both)

---

**Status**: ACTIVE
**Next Review**: After KMS demo verified working
