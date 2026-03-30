# Quizme v1 — Parameterization Opportunities

**Created**: 2026-03-29
**Purpose**: Clarify open decisions before implementation begins.

---

## Question 1: Registry YAML Location

**Question**: Where should the canonical `registry.yaml` file live?

**Context**: The Go [standard project layout](https://github.com/golang-standards/project-layout/tree/master/api)
puts API definitions and configs in `/api/`. The current Go entity registry lives at
`internal/apps/tools/cicd_lint/lint_fitness/registry/registry.go`. A new YAML file that
serves as the single source of truth for suite, product, and product-service metadata
needs a canonical home. The `#01` item in PARAMETERIZATION-OPPORTUNITIES.md references
this question.

**A)** `api/cryptosuite-registry/registry.yaml` — follows Go standard project layout;
`api/` is for API definitions, schemas, and config artifacts. A `generate.go` file alongside
invokes `./cmd/tools/cicd-registry` to produce `registry.go` from the YAML.

**B)** `api/cryptosuite-registry/registry.yaml` — same location, but NO code generation.
`cicd-lint lint-fitness` reads the YAML at runtime, builds a parameterized model in-memory,
and diffs it against the actual Go registry and codebase to emit errors for deviations.

**C)** `internal/apps/tools/cicd_lint/lint_fitness/registry/registry.yaml` — alongside the
existing `registry.go`. Stays internal; consumed by fitness linters at compile time via
`//go:embed`. Simpler, but keeps the schema hidden from external tooling.

**D)** Repository root `registry.yaml` — maximum discoverability. All consumers (Go, CI/CD,
docs) reference it from the root. Risk: root-level file proliferation.

**E)**

**Answer**:

**Rationale**: This decision affects import paths, build pipeline, and which consumers can
access the registry. Option A adds a build step but follows Go conventions. Option B avoids
code generation but requires runtime YAML parsing. Option C is simplest but limits access.

---

## Question 2: Code Generation vs Pure Validation

**Question**: Should the YAML registry generate Go code, or should fitness linters validate
YAML↔Go consistency at lint time?

**Context**: The existing `registry.go` has hardcoded Go structs that consumers import. Two
approaches: (a) generate `registry.go` from YAML so the YAML is always authoritative, or
(b) keep maintaining `registry.go` manually and use a fitness linter to catch drift.

**A)** **Generate-and-validate** — YAML → `go generate` → `registry.go`. The generated file
replaces the manual one. Fitness linters also read the YAML at runtime for cross-validation.
Advantage: eliminates manual struct maintenance. Risk: adds build step dependency; generated
code must be committed (not gitignored) for IDE support.

**B)** **Validate-only** — YAML is the schema; fitness linter loads it at runtime and compares
against the manually-maintained `registry.go`. Advantage: no new `cmd/` binary, no code
generation step, no build pipeline change. Risk: YAML and Go registry can silently drift if
`lint-fitness` is not run.

**C)** **Hybrid** — YAML is source of truth; `registry.go` is generated via `go generate`;
fitness linters import the generated Go code (not the YAML). Advantage: strong compile-time
guarantees and standard Go tooling. Risk: more complex build pipeline.

**D)** **YAML replaces Go entirely** — no `registry.go` at all; all consumers parse YAML at
startup via `//go:embed`. Advantage: single source, zero drift risk. Risk: YAML parsing at
every fitness linter invocation; breaks existing `AllProducts()`/`AllProductServices()`
import patterns used by 57+ sub-linters.

**E)**

**Answer**:

**Rationale**: The generate approach eliminates drift by construction but adds pipeline
complexity. The validate approach preserves existing imports but requires discipline. The
hybrid approach is strongest but most complex. The YAML-only approach risks breaking 57+
existing consumers.

---

## Question 3: Standalone Items Parallelization

**Question**: Should the 5 standalone items (#03, #15, #18, #19, #20) be implemented in
parallel with Phase 1, or sequentially after Phase 1?

**Context**: These 5 items have NO dependency on the entity registry (#01). They can be
implemented at any time. Doing them in parallel with Phase 1 maximizes throughput but
increases context-switching overhead.

**A)** **Parallel with Phase 1** — interleave standalone items with registry work. Higher
throughput; each standalone item is small (4h). Risk: context-switching overhead.

**B)** **Sequential after Phase 1** — complete Foundation first, then tackle standalones.
Cleaner mental model; Phase 1 insights might inform standalone implementations. Risk:
delays standalone items unnecessarily.

**C)** **Standalones FIRST, then Phase 1** — quick wins build momentum and validate the
fitness linter testing patterns before tackling the complex registry work.

**D)** **Interleaved by complexity** — alternate between standalone items and Phase 1 tasks
to maintain variety and reduce fatigue. No strict ordering.

**E)**

**Answer**:

**Rationale**: The recommended order in PARAMETERIZATION-OPPORTUNITIES.md is
`#01 → #03 → #04 → ...`, suggesting #01 first then #03. However, standalones have zero
dependency, so reordering is safe.
