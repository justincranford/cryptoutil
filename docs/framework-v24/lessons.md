# Lessons â€” Framework v24: 10-to-8 PS-ID Consolidation

---

## Executive Summary

1. Phase 1 established working jose-ja compatibility inside sm-kms (routes, handlers, repos, migrations, and generated API artifacts).
2. Phase 2 established working sm-im compatibility inside sm-kms (message schema and endpoints), including dual path route registration.
3. Phase 3 removed jose-ja, sm-im, and jose product runtime surfaces from code/config/deployments.
4. Phase 4 reconciled registry, constants, deployment lint rules, and fitness validations to 4 products/8 PS-IDs.
5. Phase 5 quality gates are partially complete: build and lint gates pass, but full repository test migration is still in progress due broad legacy topology assumptions in test suites.

---

## Actions

1. Migrate remaining tests that still encode jose/sm-im expectations to 8-PS-ID topology.
2. Replace lingering jose/sm-im fixture references in lint_fitness and framework/tls test packages with active PS-IDs.
3. Re-run full `go test ./... -shuffle=on` and close Phase 5 once legacy-topology test failures are eliminated.
4. Complete Phase 6 documentation propagation updates after test migration completion.

---

> **Per-phase structure** (fill after each phase's quality gates pass):
>
> ### What Worked
>
> - *(bullet)*
>
> ### What Didn't Work
>
> - *(bullet)*
>
> ### Root Causes
>
> - *(bullet)*
>
> ### Patterns for Future Phases
>
> - *(bullet)*

---

## Phase 1: jose-ja Domain â†’ sm-kms

### What Worked

- Porting jose-ja repository/service/handler logic directly into sm-kms with constructor injection preserved most behavior.
- Adding compatibility routes in both `/browser` and `/service` groups preserved dual-path behavior.

### What Didn't Work

- Initial compatibility route wiring panicked in tests when `ServiceResources` was nil.

### Root Causes

- Route registration assumed DB/barrier resources were always present, which is false in unit route wiring tests.

### Patterns for Future Phases

- Guard optional compatibility wiring behind explicit nil checks when tests intentionally use minimal resource stubs.

---

## Phase 2: sm-im Domain â†’ sm-kms

### What Worked

- Message domain migrations and handlers integrated cleanly into sm-kms once model/repository interfaces were aligned.

### What Didn't Work

- Missing migration comment headers blocked fitness validation despite functional SQL.

### Root Causes

- Copied SQL headers still referenced old service names and violated migration header policy checks.

### Patterns for Future Phases

- Normalize migration headers immediately after copy/rename to avoid late-stage lint churn.

---

## Phase 3: Delete jose-ja, sm-im, jose Product

### What Worked

- Hard deletions across `api/`, `cmd/`, `internal/apps/`, `configs/`, and `deployments/` reduced runtime topology to 8 PS-IDs.

### What Didn't Work

- Removing legacy constants immediately caused widespread test compile failures.

### Root Causes

- Many test and lint packages still referenced old constants/types as fixtures and assertions.

### Patterns for Future Phases

- Use temporary compatibility constants during transition while systematically migrating dependent tests.

---

## Phase 4: Registry, Magic Constants, Fitness Linters

### What Worked

- Registry YAML, tier constants, deployment compose includes, and fitness checks were aligned and `lint-fitness` was restored to green.

### What Didn't Work

- Template-compliance and migration-comment-header checks failed after first pass.

### Root Causes

- One stale SM compose comment and copied migration headers diverged from expected policy text.

### Patterns for Future Phases

- Run `lint-fitness` immediately after structural edits and fix template/header drift before proceeding.

---

## Phase 5: Full Quality Gate Verification

### What Worked

- Compile and lint gates pass: build (tagged and untagged), golangci-lint, lint-fitness, lint-deployments, lint-openapi, lint-docs.

### What Didn't Work

- Full `go test ./... -shuffle=on` remains red in multiple packages.

### Root Causes

- Large portions of the test suite still encode 10-PS-ID assumptions (jose/sm-im existence, expected registry counts, deployment fixtures).

### Patterns for Future Phases

- Treat test-topology migration as a dedicated workstream parallel to runtime topology migration.

---

## Phase 6: Knowledge Propagation

### What Worked

- No regressions in docs linting or propagation mechanics after topology edits.

### What Didn't Work

- Product/service reduction has not yet been fully propagated into handbook-level narrative sections.

### Root Causes

- Priority was given to runtime and lint gate stability before handbook narrative refactor.

### Patterns for Future Phases

- Complete Phase 6 only after Phase 5 full test migration closes to avoid repeated doc churn.
