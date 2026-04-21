# Quizme V1 - Framework V15

**Purpose**: Clarify strategic decisions before implementation begins. Fill in the `Answer:` field
with A, B, C, D, or E (custom) for each question, then re-invoke the planning agent to merge
answers into plan.md/tasks.md.

---

## Question 1: Phase 1 Test Generator Seam Location

**Question**: The pki-init generator tests (Tasks 1.1–1.3) require a seam-injected (stub crypto)
generator for speed. Where should the test seam (`ExportedNewTestGenerator`) be defined?

**A)** Add `ExportedNewTestGenerator` to a new `export_test.go` file in
`internal/apps/framework/tls/` that creates a real generator with a stub crypto backend
(fast for unit tests, no disk I/O in the hot path).

**B)** Add a `TestMode bool` field directly to the `Generator` struct in `generator.go`, and
check it in `generate*` functions to skip real crypto. (Simpler but couples test logic to
production code.)

**C)** Create a separate `generator_stub.go` file (build tag `//go:build !production`) that
provides a faster `GenerateStub()` function used only in tests. (Requires build tag gymnastics.)

**D)** Do not use a stub — run all 16-tier tests with real RSA/ECDSA crypto. Accept that Phase 1
tests will be slower but fully representative. (Straightforward but ~60s per test run.)

**E)**

**Answer**:

**Rationale**: Option A follows the project's established `export_test.go` seam pattern.
Option D is safe if the test runtime is acceptable. The decision determines how Task 1.1–1.3
use the `ExportedNewTestGenerator` seam and whether `generator_integration_test.go` is the
ONLY file using real crypto.

---

## Question 2: Go E2E Test Package Location

**Question**: Phases 4, 7, and 11 require committed Go E2E tests that orchestrate Docker Compose
and verify TLS via `crypto/tls.Dial` (per ENG-HANDBOOK.md §10.4.4). Where should these tests live?

**A)** `test/e2e/` — the existing E2E test directory (if present). One file per phase:
`framework_tls_otel_test.go`, `framework_tls_grafana_test.go`, `framework_tls_full_test.go`.
All share the same Docker Compose orchestration helpers.

**B)** `internal/apps/framework/tls/` — adjacent to the generator code. Use build tag
`//go:build e2e`. Tests live near the code they validate; no separate test package needed.

**C)** A new package `internal/apps/framework/tls/e2etests/` dedicated to TLS E2E tests.
Cleanly separated from unit/integration tests; single `TestMain` owns the compose lifecycle.

**D)** `test/e2e/framework/tls/` — a nested directory under the existing `test/e2e/` tree,
organized by product area. All framework TLS tests co-located and easily discoverable.

**E)**

**Answer**:

**Rationale**: This determines the `go test` invocation path for CI/CD and how the
`ComposeManager` + `TLSDialer` helpers are structured and shared. Option A reuses existing
E2E infrastructure. Option D adds one nesting level for discoverability. The choice affects
Phase 4 Tasks 4.1–4.3, Phase 7 Tasks 7.1–7.3, and Phase 11 Tasks 11.2–11.3.

---

## Question 3: D6 — Grafana OTLP Ingest mTLS Strategy

**Question**: Phase 5 Task 5.2 needs to configure Grafana OTLP ingest TLS (ports :14317/:14318).
The `grafana/otel-lgtm` image bundles its own OTel Collector and Grafana. Its support for
explicit OTLP ingest mTLS (separate from the UI cert) is not fully documented. What strategy
should Task 5.2 follow?

**A)** Assume D6=A: `grafana/otel-lgtm` supports OTLP ingest mTLS via grafana.ini or bundled OTel
config. Attempt to configure it and proceed. If it fails during Phase 5 verification, create a
fix task (D6=C pivot) immediately.

**B)** Skip OTLP ingest mTLS for `grafana/otel-lgtm` entirely. The App→OTel leg (Phase 3) and
OTel→Grafana UI (Phase 7) are sufficient. Document the decision: Grafana OTLP ingest uses TLS
but NOT mTLS (no client cert required for OTel→Grafana OTLP path).

**C)** Pre-validate empirically FIRST (Phase 5 Task 5.2 is split into 5.2a verify + 5.2b apply):
spin up the image locally, attempt OTLP ingest with a client cert, document the result, THEN
apply the appropriate config. Do not assume D6=A or D6=B without evidence.

**D)** Pivot immediately to D6=C (OTel sidecar architecture): deploy a separate
`otel-collector-contrib` sidecar that accepts OTLP from the bundled Grafana collector and
applies mTLS on the Grafana ingest path. Adds compose complexity but guarantees mTLS.

**E)**

**Answer**:

**Rationale**: This is the biggest unknown in V15. If `grafana/otel-lgtm` does NOT support OTLP
ingest mTLS natively, Phase 5 scope changes significantly. Option C (empirical validation first)
prevents wasted implementation effort. Option A is faster but risks a mid-phase pivot. The choice
determines Task 5.2 scope and whether Phase 5 needs a sub-phase for verification.

---

## Question 4: Deployment Template Update Timing

**Question**: Phases 2, 3, 5, 6, 8 each modify `deployments/` compose files. The canonical
templates live in `api/cryptosuite-registry/templates/`. When should the canonical templates be
updated relative to the per-service deployment files?

**A)** Within the SAME task that modifies the per-service file: update both the actual file in
`deployments/` AND the canonical template in `api/cryptosuite-registry/templates/` atomically.
The `template-compliance` fitness linter then validates both match.

**B)** Update actual deployment files first (Phases 2, 3, 5, 6, 8); update canonical templates
in Phase 9 (Deployment Templates) as a dedicated phase. This is the current plan structure.
Phase 9 is the synchronization point.

**C)** Add a new acceptance criterion to each compose-file task: "canonical template updated
in the same commit as the per-service deployment file." Phase 9 becomes a verification-only
phase (run template-compliance linter, no writing). This combines A and B.

**D)** Templates are updated only once at the end of all compose changes (Phase 9 as-is),
and the `template-compliance` linter is explicitly skipped during Phases 2–8 since templates
are intentionally out of sync. A lint-disable comment or temporary exclusion is added.

**E)**

**Answer**:

**Rationale**: The `template-compliance` fitness linter in `lint-fitness` enforces template
drift. If actual deployment files diverge from templates between phases, the linter fails CI.
Option A prevents CI drift but increases per-task scope. Option B risks CI failures during
Phases 2–8 if the linter runs in CI. Option C is the highest-discipline approach but requires
more commits per task. Option D trades immediate CI compliance for deferred synchronization
(higher risk if Phase 9 is deferred or skipped).
