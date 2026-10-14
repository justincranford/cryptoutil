# Lessons â€” Framework v25: 8-to-7 PS-ID Consolidation

**Created**: 2026-05-25
**Purpose**: Lessons learned from each phase. Filled by `implementation-execution` agent
after each phase's quality gates pass.

---

## Executive Summary

1. [Phase 1](#phase-1-port-pki-ca-packages-to-sm-kms): The PKI packages, migration, and certificate profiles are present under sm-kms and pass the final build and test suite.
2. [Phase 2](#phase-2-extend-sm-kms-api-with-pki-endpoints): All 18 PKI routes are registered under both public API prefixes; OpenAPI and handler tests pass.
3. [Phase 3](#phase-3-delete-pki-ca-service-artifacts): The standalone pki-ca service artifacts are absent and normal/tagged builds pass.
4. [Phase 4](#phase-4-delete-pki-product-and-update-registry): The canonical registry, runtime registry, and fitness validation reflect 3 products and 7 PS-IDs.
5. [Phase 5](#phase-5-quality-gates): Full builds, shuffled tests, Go linters, cicd-lint checks, PKI coverage, and sm-kms race tests pass.
6. [Phase 6](#phase-6-knowledge-propagation): README and handbook topology documentation are reconciled, lessons are recorded, and docs propagation validation passes.

---

## Actions

1. PKI persistence follow-up resolved: certificates, revocations, and CA authority state use the shared sm-kms SQLite/PostgreSQL database; the CA private key is barrier-encrypted, and CRL/OCSP responses read tenant-scoped durable status.
2. Keep the retired PKI port 8300, PostgreSQL host port 54323, and migration range 5001-5999 unassigned until an explicit canonical registry change authorizes reuse.
3. For future migrations, create and update each phase's evidence directory and task statuses as its quality gates complete instead of reconstructing status from the final tree.

---

> **Mandatory per-phase structure** (each `## Phase N:` section MUST use these 4 sub-headings):
>
> ### What Worked
>
> ### What Didn't Work
>
> ### Root Causes
>
> ### Patterns for Future Phases

---

## Phase 1: Port pki-ca Packages to sm-kms

### What Worked

The package tree under `internal/apps/sm-kms/server/pki/` keeps PKI implementation within the owning service. The full shuffled test suite and package coverage profile verify the ported packages in their current location.

### What Didn't Work

No Phase 1-specific build or test logs were retained, so the original execution sequence and intermediate failures cannot be reconstructed.

### Root Causes

The phase-level evidence archive was not maintained alongside the individual task statuses. The existing migration and code are present, but historical per-phase validation evidence is missing.

### Patterns for Future Phases

Run package-scoped build/tests while porting, preserve the logs under the phase directory, and update each task status immediately after its acceptance criteria pass. The original in-memory port is retained only as a test store; production persistence belongs to the shared sm-kms database.

---

## Phase 2: Extend sm-kms API with PKI Endpoints

### What Worked

The dual-prefix route test enumerates all 18 PKI paths under both `/service/` and `/browser/` prefixes and checks that the profiles endpoint responds successfully at both mounts. OpenAPI generation and handler package tests provide additional validation.

### What Didn't Work

No Phase 2-specific logs were retained. The mounted-route test checks registration for every route, while endpoint request/response behavior is primarily exercised by the PKI handler package tests.

### Root Causes

The evidence archive did not distinguish route-registration checks from handler-level behavior checks, and the phase logs were not persisted.

### Patterns for Future Phases

Keep OpenAPI validation, generated-code builds, route registration, and handler behavior tests as separate evidence entries. Use in-memory `app.Test()` requests for HTTP behavior tests and avoid treating route-table inspection alone as an endpoint behavior test.

---

## Phase 3: Delete pki-ca Service Artifacts

### What Worked

The current tree has no standalone pki-ca API, service, command, config, or deployment directories. Full normal/tagged builds and deployment validation pass after removal.

### What Didn't Work

No Phase 3-specific deletion or build logs were retained.

### Root Causes

The execution tracker did not preserve per-phase evidence when the deletion work was completed.

### Patterns for Future Phases

Pair large artifact removals with a bounded existence check, repository-wide build, and deployment validator; save the exact command output in the phase evidence directory.

---

## Phase 4: Delete pki Product and Update Registry

### What Worked

The machine-readable registry, magic tier constants, and lint-fitness entity registry now agree on the 3-product / 7-PS-ID topology. Fitness and deployment validation pass.

### What Didn't Work

No Phase 4-specific registry or validator logs were retained. The retired pki-ca sentinel constants remain referenced by regression tests by design, so broad text searches need to distinguish retained retirement guards from live registrations.

### Root Causes

The original checks did not document the intentional retired-ID exception, and the phase-specific evidence directory was not kept.

### Patterns for Future Phases

Treat the machine-readable registry as the source of truth, validate all derived registries with fitness checks, and retain retired identifiers only when regression guards require them. State those exceptions explicitly in completion criteria.

---

## Phase 5: Quality Gates

### What Worked

The standard and tagged builds, full shuffled test suite, normal and tagged golangci-lint runs, eight cicd-lint checks, PKI coverage, and sm-kms race tests all passed. Every PKI package with executable statements exceeded 95% coverage. The current sm-kms aggregate coverage is 79.9%, above the retained 78.3% baseline; businesslogic improved from 95.2% to 95.3%, and handler coverage remains 100%.

The final coverage review added focused tests for issuer/date-bound storage filters, decrementing an uninitialized gauge, and JSON encoder failure propagation. Those paths now reach 100% function coverage; storage is 98.9%, observability is 100%, and the PKI handler package is 95.3%. Thirty functions remain below 95% individually, but no PKI package is below its package target. Their uncovered statements are mostly alternate validation, crypto/persistence failures, and response-write failures. The `SendStatus` error branches in `ESTCSRAttrs` and `ocspErrorResponse` cannot be forced through Fiber's in-memory `app.Test()` writer, so no tests were added solely to inflate function percentages.

### What Didn't Work

The Windows host has no GCC compiler, so the local race command could not run. An initial inline Windows coverage-profile argument was parsed as an extra package.

### Root Causes

Go's race detector requires CGO and a C compiler; the host has neither. The inline coverage argument form was incompatible with path parsing in this shell invocation.

### Patterns for Future Phases

Run race tests in a Go container with CGO enabled when the Windows host lacks GCC. Pass the coverage-profile path as a quoted, separate argument and validate it with `go tool cover` before retaining the artifact. Treat function-level coverage as a diagnostic: inspect uncovered behavior, test meaningful branches, and document structural or external-boundary limits without weakening package targets.

---

## Phase 6: Knowledge Propagation

### What Worked

The README and ENG-HANDBOOK now describe three products and seven services, position PKI as an sm-kms capability, and use corrected directory-count formulas. `lint-docs` and `lint-text` pass.

### What Didn't Work

The plan and task trackers still reflected stale task-level labels, and the older phase evidence folders were absent.

### Root Causes

Tracker and evidence updates had not been part of the phase completion workflow.

### Patterns for Future Phases

Reconcile phase status from current implementation and executable evidence; do not interpret stale task labels as proof that work was not done. Preserve missing historical evidence as an explicit limitation instead of fabricating it.
