# quizme-v3 — Decisions D12 and D14 Need Your Confirmation

**Context**: quizme-v2 Q1 and Q3 answers were "E — not enough context." I provided executive summaries and
made recommendations, but recorded them as decisions without your confirmation. That was wrong. This file
gives you the full picture and asks for your final call. **D12 and D14 in plan.md are tentative until you answer these.**

**Instructions**: Fill in `**Answer**:` with A, B, C, D, or E for each question.

---

## Question 1: skeleton-template Role (D12 tentative)

**What it is today**: 19 Go files, 61KB. Does exactly one job: provides a clean starting point for
`/new-service`. It has no domain logic, no CRUD, no real HTTP routes — just builder wiring + health checks +
contract tests.

**What lint-fitness does with it**: `discoverServices()` **explicitly excludes** skeleton. It has its own
sub-linter `check-skeleton-placeholders` that validates placeholder strings exist (so generated services
must replace them). Skeleton is NOT validated as a real service — it's validated as a template source.

**The real CRUD reference is sm-im**: 61 files, 357KB, proven through full Phase 1. When someone needs to
understand "how does a real domain service work," sm-im is the answer. skeleton-template is only "how do I
start a new service from scratch."

**Problem statement**: When you create a new service with `/new-service`, does skeleton-template give you
enough as a starting point? Currently it gives you the builder pattern, contract tests, health checks, and
placeholder names to replace. It does NOT show you how to add a domain table, a repository, or a handler.

| Option | What you get | Work to implement | Risk |
|--------|-------------|-------------------|------|
| **A) Keep as-is** | Skeleton stays 19 files. /new-service gives you a running skeleton with wiring only. You look at sm-im for how to add domain code. | 0h — just document the relationship | Low — nothing changes |
| **B) Add minimal CRUD example** | Add 1 example table + repository + handler to skeleton showing the pattern. ~30 lines. `/new-service` generates it, dev deletes what they don't need. | ~4h | Low — adds concrete guidance |
| **C) Add code generation** | skeleton generates domain code from a config (table name, fields, operations). Like Rails scaffold. | ~40h, highly complex | High — over-engineered for 10 services |
| **D) Deprecate skeleton, use sm-im as template source** | `/new-service` strips sm-im down to skeleton, dev adds back what they need | ~8h, fragile — sm-im evolves constantly | High — sm-im changes break new-service |
| **E)** | | | |

**My judgment**: **A or B**. The question is whether a developer creating a new service needs to see a CRUD
example inside the skeleton or can be pointed at sm-im. If you think "I'd always open sm-im side-by-side
anyway," pick A. If you think "the skeleton should show me one complete domain example so I can copy-paste,"
pick B.

**Answer**:

---

## Question 2: InsecureSkipVerify Scope (D14 tentative)

**The 47 files broken down precisely**:

| Category | Count | Fix approach |
|----------|-------|-------------|
| Integration/contract test HTTP clients | 38 | Replace with `TLSClientConfig(t)` — mechanical, low risk |
| lint-fitness (the DETECTOR, not user) | 2 | No change needed — it detects the pattern, doesn't use it |
| Demo/script files | 2 | Leave as-is (acceptable in demos) |
| Archived/legacy files | 2 | Leave as-is (archived) |
| **1 production file** (identity-rp) | 1 | Real bug — production code skips cert verification |
| E2E/Docker test helpers | 2 | Needs real CA chain in Docker — much more work |

**How the fix works for Phase 2A**:
The service auto-generates a 3-tier CA chain at startup (Root → Intermediate → Leaf via `tls_generator.go`).
Phase 2A adds `TLSBundle()` to `ServiceServer` so test code can get the CA cert, then builds a
`*tls.Config` that trusts it. Replace `InsecureSkipVerify: true` with that config. ~38 files, all
mechanical search-and-replace.

**The production bug**: `identity/rp/server/public_server.go` uses `InsecureSkipVerify: true` in
production code — actual data in flight is not verified. This is a real security defect. However,
identity-rp will be fully extracted and replaced with a clean skeleton in Phase 7, which eliminates this
file entirely. Fix it now or let Phase 7 erase it?

| Option | Scope | Duration | What's deferred |
|--------|-------|----------|----------------|
| **A) Phase 2A only** | 38 integration test files | ~2-3 days | Production bug (erased by Phase 7), E2E TLS (2B), mTLS (2C), PostgreSQL TLS (2D) |
| **B) Phase 2A + production bug fix** | 38 integration test files + 1 production file | ~2-3 days + 2h | E2E TLS (2B), mTLS (2C), PostgreSQL TLS (2D) |
| **C) Phase 2A + 2B (include E2E)** | 38 integration + E2E Docker TLS | ~2 weeks | mTLS (2C), PostgreSQL TLS (2D) |
| **D) Defer all until after Phase 7** | Nothing now | 0h now | Everything — gosec keeps flagging until then |
| **E)** | | | |

**My judgment**: **B** slightly over A. The production file is a real security defect and the fix is
trivial (`identity/rp/server/public_server.go` replaces `InsecureSkipVerify: true` with the CA-trusting
config). It disappears in Phase 7 anyway, but it's a 2-hour fix that removes a genuine security smell
from code review. If you're comfortable knowing a production file has it until Phase 7, pick A. If you
want zero production `InsecureSkipVerify` regardless, pick B.

C is disproportionate (E2E TLS requires real CA certs in Docker Compose, separate infrastructure work).
D creates ongoing noise in every PR.

**Answer**:

---

## After You Answer

Fill in both `**Answer**:` fields above, then say `continue` or `update`. I will:
1. Confirm/update D12 and D14 in plan.md based on your answers
2. Adjust Phase 2 tasks if needed (especially if B — adds Task 2.X: fix production InsecureSkipVerify)
3. Delete this file
