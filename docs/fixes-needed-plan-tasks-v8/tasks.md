# Tasks - Complete KMS Migration (V8)

**Status**: 1 of 71 tasks complete (1.4%)
**Last Updated**: 2026-02-14
**Purpose**: Complete the ACTUAL remaining work from V7

## Testing Strategy (from Executive Decisions)

**Phase-Level Testing**: Unit + integration + E2E tests in EVERY phase
**Mutation Testing**: Phase 5 at END (NOT deferred - strategically ordered)
**Documentation**: Incremental updates to actually-wrong instructions only

---

## ~~Phase 1: Barrier Integration~~ **SUPERSEDED BY PHASE 13**

**NOTE**: Phase 1 tasks 1.1-1.5 used INCORRECT adapter approach. 
**SEE PHASE 13** for the CORRECT direct migration (like cipher-im, NO adapter).

### ~~Task 1.1-1.5: SUPERSEDED~~
- **Status**: ⏭️ SUPERSEDED by Phase 13
- **Reason**: Phase 1 incorrectly recommended using `orm_barrier_adapter.go` to bridge interfaces
- **Correct Approach**: Use template barrier DIRECTLY like cipher-im (Phase 13 tasks)

---

## Phase 2: Testing & Verification

### Task 2.1: KMS Unit Tests Pass
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Phase 1 complete
- **Description**: Run all KMS unit tests
- **Acceptance Criteria**:
  - [ ] `go test ./internal/kms/... -count=1` passes
  - [ ] No test regressions
- **Evidence**: test output

### Task 2.2: cipher-im Regression Tests Pass
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Phase 1 complete
- **Description**: Verify cipher-im not broken by any shared changes
- **Acceptance Criteria**:
  - [ ] `go test ./internal/apps/cipher/... -count=1` passes
- **Evidence**: test output

### Task 2.3: jose-ja Regression Tests Pass
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Phase 1 complete
- **Description**: Verify jose-ja not broken by any shared changes
- **Acceptance Criteria**:
  - [ ] `go test ./internal/apps/jose/... -count=1` passes
- **Evidence**: test output

### Task 2.4: Full Build Verification
- **Status**: ❌ Not Started
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: Tasks 2.1-2.3
- **Description**: Verify full project builds
- **Acceptance Criteria**:
  - [ ] `go build ./...` passes
  - [ ] No errors or warnings
- **Evidence**: build output

---

## Phase 3: Documentation & Cleanup

### Task 3.1: Update server-builder.instructions.md
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Phase 2 complete
- **Description**: Update ServerBuilder documentation
- **Acceptance Criteria**:
  - [ ] Remove outdated V6 optional mode references
  - [ ] Document unified architecture
- **Files**:
  - `.github/instructions/03-08.server-builder.instructions.md`

### Task 3.2: Remove V7 builder_adapter.go if Still Exists
- **Status**: ❌ Not Started
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: Phase 2 complete
- **Description**: Clean up V7 Task 1.5 (remove builder_adapter.go)
- **Acceptance Criteria**:
  - [ ] builder_adapter.go deleted (if exists)
  - [ ] Build still passes
- **Files**:
  - `internal/kms/server/builder_adapter.go` (DELETE if exists)

### Task 3.3: Archive V7 Documentation
- **Status**: ❌ Not Started
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: Task 3.2
- **Description**: Move V7 docs to archive
- **Acceptance Criteria**:
  - [ ] V7 files moved to `docs/fixes-needed-plan-tasks-v8/archive/v7/`
  - [ ] Clean documentation structure

---

## Phase 3.5: Realm Design Verification

**Background**: Realms define authentication METHOD only, NOT data scoping.
- `tenant_id` = data isolation (ALL data queries filter by tenant_id)
- `realm_id` = authentication method (HOW users authenticate)

### Task 3.5.1: Verify cipher-im Realm Usage
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: Phase 2 complete
- **Description**: Verify cipher-im uses realms correctly (authentication method only, not data scoping)
- **Acceptance Criteria**:
  - [ ] Data queries filter by tenant_id (NOT realm_id)
  - [ ] Realm only determines authentication method
  - [ ] No realm-based data isolation logic
- **Evidence**: Code audit showing tenant_id in data queries

### Task 3.5.2: Verify jose-ja Realm Usage
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: Phase 2 complete
- **Description**: Verify jose-ja uses realms correctly
- **Acceptance Criteria**:
  - [ ] Data queries filter by tenant_id (NOT realm_id)
  - [ ] Realm only determines authentication method
  - [ ] No realm-based data isolation logic
- **Evidence**: Code audit showing tenant_id in data queries

### Task 3.5.3: Verify sm-kms Realm Usage
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: Phase 2 complete
- **Description**: Verify sm-kms uses realms correctly (after barrier migration)
- **Acceptance Criteria**:
  - [ ] Data queries filter by tenant_id (NOT realm_id)
  - [ ] Realm only determines authentication method
  - [ ] No realm-based data isolation logic
- **Evidence**: Code audit showing tenant_id in data queries

### Task 3.5.4: Verify Template RealmService
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: None
- **Description**: Verify template RealmService exposes all 16 realm types correctly
- **Acceptance Criteria**:
  - [ ] 4 federated types: username_password, ldap, oauth2, saml
  - [ ] 6 browser types: jwe-session-cookie, jws-session-cookie, opaque-session-cookie, basic-username-password, bearer-api-token, https-client-cert
  - [ ] 6 service types: jwe-session-token, jws-session-token, opaque-session-token, basic-client-id-secret, (shared: bearer-api-token, https-client-cert)
  - [ ] RealmConfig has password, session, MFA, rate limiting settings
- **Evidence**: Code review of realm_service.go and realm_config.go

### Task 3.5.5: Update Realm Documentation in Instructions
- **Status**: ✅ Complete
- **Estimated**: 0.5h
- **Actual**: 0.5h
- **Dependencies**: None
- **Description**: Update instruction files with correct realm definitions
- **Acceptance Criteria**:
  - [x] ARCHITECTURE.md expanded with 16 realm types
  - [x] SERVICE-TEMPLATE.md has Realm Pattern section
  - [x] analysis-overview.md Section 12 added
  - [x] analysis-thorough.md Section 12 added
- **Evidence**: Documentation commits

---

## Phase 4: Delete shared/barrier

### Task 4.1: Delete internal/shared/barrier/
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Phase 1 complete, Phase 2 verified
- **Description**: Remove the deprecated shared/barrier implementation
- **Acceptance Criteria**:
  - [ ] `rm -rf internal/shared/barrier/` executed
  - [ ] No compile errors
  - [ ] All tests pass
- **Evidence**: `ls internal/shared/barrier/` returns "No such file or directory"

### Task 4.2: Verify No Remaining shared/barrier Imports
- **Status**: ❌ Not Started
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: Task 4.1
- **Description**: Ensure no orphaned imports anywhere in codebase
- **Acceptance Criteria**:
  - [ ] `grep -r "shared/barrier" .` returns only this tasks.md mention
  - [ ] Build clean
  - [ ] All tests pass
- **Evidence**: grep command output

---

## Phase 5: Mutation Testing (Grouped at End - NOT DEFERRED)

⚠️ **STRATEGIC ORDERING**: Mutations at end by design, NOT deferred/skipped

### Task 5.1: KMS Mutation Testing
- **Status**: ❌ Not Started
- **Estimated**: 2h
- **Actual**:
- **Dependencies**: Phases 1-4 complete
- **Description**: Run gremlins on KMS after barrier migration complete
- **Acceptance Criteria**:
  - [ ] `gremlins unleash ./internal/kms/...` completes
  - [ ] ≥85% mutation score (production code)
  - [ ] Results documented in test-output/mutation-results/
- **Evidence**: gremlins output

### Task 5.2: Template Barrier Mutation Testing
- **Status**: ❌ Not Started
- **Estimated**: 1.5h
- **Actual**:
- **Dependencies**: Task 5.1
- **Description**: Verify template barrier tests catch mutations
- **Acceptance Criteria**:
  - [ ] `gremlins unleash ./internal/apps/template/service/server/barrier/...` completes
  - [ ] ≥98% mutation score (infrastructure code)
- **Evidence**: gremlins output

---

## Phase 6: sm-kms Structure Migration

### Task 6.1: Create cmd/sm-kms Entry Point
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: Phase 1-5 complete
- **Description**: Create proper cmd entry point following pattern
- **Acceptance Criteria**:
  - [ ] `cmd/sm-kms/main.go` exists
  - [ ] Delegates to `internal/apps/sm/kms/`
  - [ ] `go build ./cmd/sm-kms/` succeeds
- **Evidence**: File exists, builds clean

### Task 6.2: Create internal/apps/sm/kms Directory
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Task 6.1
- **Description**: Create directory structure for sm/kms
- **Acceptance Criteria**:
  - [ ] `internal/apps/sm/kms/` directory exists
  - [ ] Matches cipher-im/jose-ja structure
- **Evidence**: Directory listing

### Task 6.3: Migrate internal/kms to internal/apps/sm/kms
- **Status**: ❌ Not Started
- **Estimated**: 4h
- **Actual**:
- **Dependencies**: Task 6.2
- **Description**: Move all code from internal/kms/ to internal/apps/sm/kms/
- **Acceptance Criteria**:
  - [ ] All .go files moved
  - [ ] Package names updated (kms → smkms where needed)
  - [ ] Build succeeds
- **Evidence**: `go build ./...`

### Task 6.4: Update All Imports
- **Status**: ❌ Not Started
- **Estimated**: 2h
- **Actual**:
- **Dependencies**: Task 6.3
- **Description**: Update all imports from internal/kms to internal/apps/sm/kms
- **Acceptance Criteria**:
  - [ ] `grep -r "internal/kms" .` returns only migration docs
  - [ ] All tests pass
- **Evidence**: grep output, test results

### Task 6.5: Update Deployment Files
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: Task 6.4
- **Description**: Update compose.yml, Dockerfile for new structure
- **Acceptance Criteria**:
  - [ ] Docker build succeeds
  - [ ] Docker compose up works
- **Evidence**: docker compose logs

### Task 6.6: Delete internal/kms
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Tasks 6.1-6.5 complete, all tests pass
- **Description**: Remove old directory after migration verified
- **Acceptance Criteria**:
  - [ ] `rm -rf internal/kms/` executed
  - [ ] All tests still pass
  - [ ] All builds still work
- **Evidence**: Directory gone, tests pass

---

## Phase 7: jose-ja Consolidation

### Task 7.1: Rename cmd/jose-server to cmd/jose-ja
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Phase 6 complete
- **Description**: Rename cmd entry to follow pattern
- **Acceptance Criteria**:
  - [ ] `cmd/jose-ja/main.go` exists
  - [ ] `cmd/jose-server/` deleted
  - [ ] Build succeeds
- **Evidence**: Directory listing, build clean

### Task 7.2: Analyze jose Implementations
- **Status**: ❌ Not Started
- **Estimated**: 2h
- **Actual**:
- **Dependencies**: Task 7.1
- **Description**: Document differences between internal/jose/ and internal/apps/jose/ja/
- **Acceptance Criteria**:
  - [ ] Analysis document created
  - [ ] Consolidation plan defined
- **Evidence**: Analysis in analysis-thorough.md

### Task 7.3: Consolidate to internal/apps/jose/ja
- **Status**: ❌ Not Started
- **Estimated**: 4h
- **Actual**:
- **Dependencies**: Task 7.2
- **Description**: Merge implementations into conformant structure
- **Acceptance Criteria**:
  - [ ] All functionality in internal/apps/jose/ja/
  - [ ] Tests pass
  - [ ] Build succeeds
- **Evidence**: test results

### Task 7.4: Delete internal/jose
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Task 7.3 verified
- **Description**: Remove old directory
- **Acceptance Criteria**:
  - [ ] `rm -rf internal/jose/` executed
  - [ ] All tests pass
- **Evidence**: Directory gone, tests pass

---

## Phase 8: pki-ca Renaming

### Task 8.1: Rename cmd/ca-server to cmd/pki-ca
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Phase 7 complete
- **Description**: Rename cmd entry
- **Acceptance Criteria**:
  - [ ] `cmd/pki-ca/main.go` exists
  - [ ] `cmd/ca-server/` deleted
- **Evidence**: Directory listing

### Task 8.2: Move internal/apps/ca to internal/apps/pki/ca
- **Status**: ❌ Not Started
- **Estimated**: 2h
- **Actual**:
- **Dependencies**: Task 8.1
- **Description**: Move to correct product directory
- **Acceptance Criteria**:
  - [ ] `internal/apps/pki/ca/` exists
  - [ ] `internal/apps/ca/` deleted
  - [ ] All imports updated
  - [ ] Tests pass
- **Evidence**: Directory listing, test results

### Task 8.3: Update Deployment Files
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: Task 8.2
- **Description**: Update compose.yml, Dockerfile
- **Acceptance Criteria**:
  - [ ] Docker build succeeds
  - [ ] E2E tests pass
- **Evidence**: docker logs

---

## Summary Statistics

| Phase | Tasks | Completed | Percentage |
|-------|-------|-----------|------------|
| Phase 1: Barrier Integration | 5 | 0 | 0% |
| Phase 2: Testing | 4 | 0 | 0% |
| Phase 3: Documentation | 3 | 0 | 0% |
| Phase 3.5: Realm Verification | 5 | 1 | 20% |
| Phase 4: Delete shared/barrier | 2 | 0 | 0% |
| Phase 5: Mutation Testing | 2 | 0 | 0% |
| Phase 6: sm-kms Structure | 6 | 0 | 0% |
| Phase 7: jose-ja Consolidation | 4 | 0 | 0% |
| Phase 8: pki-ca Renaming | 3 | 0 | 0% |
| **Total** | **34** | **1** | **3%** |

---

## V7 Tasks Carried Over (Not Actually Done)

| V7 Task | V8 Task | Reason |
|---------|---------|--------|
| 5.3: Integrate Template Barrier with KMS | 1.1-1.4 | Marked "Not Started" in V7 |
| 5.4: Remove shared/barrier Usage from KMS | 1.5 | Marked "Not Started" in V7 |
| 1.5: Remove KMS builder_adapter.go | 3.2 | Marked "Not Started" in V7 |
| 1.6: Update ServerBuilder Documentation | 3.1 | Marked "Not Started" in V7 |
| 6.1-6.7: All Phase 6 Testing | 2.1-2.4 | All marked "Not Started" in V7 |
| 7.1-7.4: All Phase 7 Documentation | 3.1-3.3 | All marked "Not Started" in V7 |

---

## Phase 9: pki-ca Health Path Standardization

### Task 9.1: Update CA Server Admin Routes
- **Status**: ❌ Not Started
- **Estimated**: 2h
- **Actual**:
- **Dependencies**: Phase 2 complete
- **Description**: Update CA server to expose admin endpoints at `/admin/api/v1/*`
- **Acceptance Criteria**:
  - [ ] `/admin/api/v1/livez` endpoint exists
  - [ ] `/admin/api/v1/readyz` endpoint exists
  - [ ] Old `/livez` path removed or redirects
  - [ ] CA server uses service-template admin pattern
- **Files**:
  - `internal/apps/ca/server/*.go`
  - `internal/apps/pki/ca/server/*.go` (if moved)

### Task 9.2: Update CA Compose Healthchecks
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Task 9.1
- **Description**: Update all CA compose files to use standard health path
- **Acceptance Criteria**:
  - [ ] `deployments/ca/compose.yml` uses `/admin/api/v1/livez`
  - [ ] `deployments/ca/compose.simple.yml` uses `/admin/api/v1/livez`
  - [ ] All CA container healthchecks updated
- **Files**:
  - `deployments/ca/compose.yml`
  - `deployments/ca/compose.simple.yml`

### Task 9.3: Update CA Configuration Files
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Task 9.1
- **Description**: Update CA config files for new admin paths
- **Acceptance Criteria**:
  - [ ] All CA YAML configs updated
  - [ ] Admin server settings correct
- **Files**:
  - `deployments/ca/config/*.yml`
  - `configs/ca/*.yml`

### Task 9.4: Verify CA Tests Pass
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: Tasks 9.1-9.3
- **Description**: Run CA tests with new health paths
- **Acceptance Criteria**:
  - [ ] `go test ./internal/apps/ca/... -count=1` passes
  - [ ] E2E tests with compose work
- **Evidence**: test output

---

## Phase 10: jose-ja Admin Port Standardization

### Task 10.1: Update JOSE Server Admin Port
- **Status**: ❌ Not Started
- **Estimated**: 1.5h
- **Actual**:
- **Dependencies**: Phase 2 complete
- **Description**: Update JOSE server to use admin port 9090 instead of 9092
- **Acceptance Criteria**:
  - [ ] Admin server binds to port 9090
  - [ ] Config defaults updated
  - [ ] Magic constants updated if any
- **Files**:
  - `internal/apps/jose/ja/server/*.go`
  - `internal/jose/server/*.go` (if exists)

### Task 10.2: Update JOSE Compose Files
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Task 10.1
- **Description**: Update JOSE compose with correct port 9090
- **Acceptance Criteria**:
  - [ ] Admin port mapping is 9090:9090
  - [ ] Healthcheck uses port 9090
  - [ ] Public port mapping is 8060:8060 (new)
- **Files**:
  - `deployments/jose/compose.yml`

### Task 10.3: Update JOSE Configuration
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Task 10.1
- **Description**: Update JOSE config files
- **Acceptance Criteria**:
  - [ ] Config files use port 9090 for admin
  - [ ] Config files use port 8060 for public (new)
- **Files**:
  - `deployments/jose/config/*.yml`
  - `configs/jose/*.yml`

### Task 10.4: Verify JOSE Tests Pass
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: Tasks 10.1-10.3
- **Description**: Run JOSE tests with new ports
- **Acceptance Criteria**:
  - [ ] `go test ./internal/apps/jose/... -count=1` passes
  - [ ] `go test ./internal/jose/... -count=1` passes (if exists)
- **Evidence**: test output

---

## Phase 11: Port Range Standardization (All Services)

### Task 11.1: Update cipher-im Port (8888 → 8070)
- **Status**: ❌ Not Started
- **Estimated**: 1.5h
- **Actual**:
- **Dependencies**: Phase 10 complete
- **Description**: Change cipher-im from 8888 to 8070
- **Acceptance Criteria**:
  - [ ] Server code uses 8070
  - [ ] Config files use 8070
  - [ ] Compose uses 8070:8070, 8071:8070, 8072:8070
  - [ ] All tests pass
- **Files**:
  - `internal/apps/cipher/im/server/*.go`
  - `configs/cipher/im/*.yml`
  - `cmd/cipher-im/docker-compose.yml`

### Task 11.2: Update jose-ja Public Port (8092 → 8060)
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: Task 10.4
- **Description**: Change jose-ja public port from 8092 to 8060
- **Acceptance Criteria**:
  - [ ] Server code uses 8060
  - [ ] Config files use 8060
  - [ ] Compose uses 8060:8060
- **Files**:
  - `internal/apps/jose/ja/server/*.go`
  - `deployments/jose/compose.yml`

### Task 11.3: Update pki-ca Public Port (8443 → 8050)
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: Task 9.4
- **Description**: Change pki-ca public port from 8443 to 8050
- **Acceptance Criteria**:
  - [ ] Server code uses 8050
  - [ ] Config files use 8050
  - [ ] Compose uses 8050:8050, 8051:8050, 8052:8050
- **Files**:
  - `internal/apps/ca/server/*.go`
  - `deployments/ca/compose.yml`

### Task 11.4: Update Architecture Documentation
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: Tasks 11.1-11.3
- **Description**: Update all architecture docs with new ports
- **Acceptance Criteria**:
  - [ ] docs/arch/ARCHITECTURE.md service catalog updated
  - [ ] docs/arch/SERVICE-TEMPLATE.md updated
- **Files**:
  - `docs/arch/ARCHITECTURE.md`
  - `docs/arch/SERVICE-TEMPLATE.md`

### Task 11.5: Update Copilot Instructions
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: Tasks 11.1-11.3
- **Description**: Update instruction files with new port assignments
- **Acceptance Criteria**:
  - [ ] `.github/instructions/02-01.architecture.instructions.md` updated
  - [ ] `.github/instructions/02-03.https-ports.instructions.md` updated
  - [ ] Port table matches new standard
- **Files**:
  - `.github/instructions/02-01.architecture.instructions.md`
  - `.github/instructions/02-03.https-ports.instructions.md`

### Task 11.6: Update V8 Analysis Documents
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Tasks 11.1-11.5
- **Description**: Update analysis docs with new port assignments
- **Acceptance Criteria**:
  - [ ] analysis-overview.md Section 11 updated
  - [ ] analysis-thorough.md updated
- **Files**:
  - `docs/fixes-needed-plan-tasks-v8/analysis-overview.md`
  - `docs/fixes-needed-plan-tasks-v8/analysis-thorough.md`

---

## Phase 12: CICD lint-ports Validation

### Task 12.1: Create lint-ports Command Structure
- **Status**: ❌ Not Started
- **Estimated**: 2h
- **Actual**:
- **Dependencies**: Phase 11 complete
- **Description**: Create cicd lint-ports command skeleton
- **Acceptance Criteria**:
  - [ ] `internal/cmd/cicd/lint_ports/` directory exists
  - [ ] Main command file with cobra integration
  - [ ] Port constants defined (source of truth)
- **Files**:
  - `internal/cmd/cicd/lint_ports/lint_ports.go`
  - `internal/cmd/cicd/lint_ports/constants.go`

### Task 12.2: Implement Code Validation
- **Status**: ❌ Not Started
- **Estimated**: 2h
- **Actual**:
- **Dependencies**: Task 12.1
- **Description**: Validate ports in Go source files
- **Acceptance Criteria**:
  - [ ] Scans `internal/apps/*/` for port references
  - [ ] Validates magic constants
  - [ ] Reports violations
- **Files**:
  - `internal/cmd/cicd/lint_ports/validate_code.go`

### Task 12.3: Implement Config Validation
- **Status**: ❌ Not Started
- **Estimated**: 1.5h
- **Actual**:
- **Dependencies**: Task 12.1
- **Description**: Validate ports in YAML config files
- **Acceptance Criteria**:
  - [ ] Scans `configs/*/` for port references
  - [ ] Validates bind_port settings
  - [ ] Reports violations
- **Files**:
  - `internal/cmd/cicd/lint_ports/validate_config.go`

### Task 12.4: Implement Compose Validation
- **Status**: ❌ Not Started
- **Estimated**: 1.5h
- **Actual**:
- **Dependencies**: Task 12.1
- **Description**: Validate ports in compose files
- **Acceptance Criteria**:
  - [ ] Scans `deployments/*/compose*.yml`
  - [ ] Validates port mappings
  - [ ] Reports violations
- **Files**:
  - `internal/cmd/cicd/lint_ports/validate_compose.go`

### Task 12.5: Implement Documentation Validation
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: Task 12.1
- **Description**: Validate ports in documentation
- **Acceptance Criteria**:
  - [ ] Scans `docs/arch/*.md`
  - [ ] Scans `.github/instructions/*.md`
  - [ ] Reports violations
- **Files**:
  - `internal/cmd/cicd/lint_ports/validate_docs.go`

### Task 12.6: Add lint-ports to Pre-commit
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Tasks 12.2-12.5
- **Description**: Add lint-ports to pre-commit hooks
- **Acceptance Criteria**:
  - [ ] `.pre-commit-config.yaml` includes lint-ports
  - [ ] Hook runs on relevant file changes
- **Files**:
  - `.pre-commit-config.yaml`

### Task 12.7: Add lint-ports to CI/CD
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Task 12.6
- **Description**: Add lint-ports to CI workflow
- **Acceptance Criteria**:
  - [ ] CI workflow runs lint-ports
  - [ ] Fails build on violations
- **Files**:
  - `.github/workflows/ci-quality.yml`

### Task 12.8: lint-ports Unit Tests
- **Status**: ❌ Not Started
- **Estimated**: 1.5h
- **Actual**:
- **Dependencies**: Tasks 12.2-12.5
- **Description**: Write tests for lint-ports command
- **Acceptance Criteria**:
  - [ ] ≥95% coverage for lint_ports package
  - [ ] Tests for each validation type
- **Files**:
  - `internal/cmd/cicd/lint_ports/*_test.go`

---

## Phase 13: KMS Barrier Direct Migration (REVISED)

**Note**: This REPLACES Phase 1 barrier approach. KMS uses template barrier directly like cipher-im.

### Task 13.1: Study cipher-im Barrier Usage Pattern
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: None
- **Description**: Document exactly how cipher-im uses template barrier
- **Acceptance Criteria**:
  - [ ] Import pattern documented
  - [ ] ServerBuilder usage documented
  - [ ] BarrierService access pattern documented
- **Evidence**: Pattern documentation

### Task 13.2: Refactor KMS to Use ServerBuilder
- **Status**: ❌ Not Started
- **Estimated**: 4h
- **Actual**:
- **Dependencies**: Task 13.1
- **Description**: Refactor KMS server.go to use ServerBuilder like cipher-im
- **Acceptance Criteria**:
  - [ ] KMS imports `cryptoutilAppsTemplateServiceServerBuilder`
  - [ ] KMS uses `builder.Build()` to get ServiceResources
  - [ ] KMS uses `res.BarrierService` for barrier operations
- **Files**:
  - `internal/kms/server/server.go`

### Task 13.3: Update KMS businesslogic.go
- **Status**: ❌ Not Started
- **Estimated**: 2h
- **Actual**:
- **Dependencies**: Task 13.2
- **Description**: Update businesslogic to use template barrier
- **Acceptance Criteria**:
  - [ ] Import changed from `shared/barrier` to template barrier
  - [ ] BarrierService parameter type matches template
  - [ ] All encryption/decryption uses template barrier
- **Files**:
  - `internal/kms/server/businesslogic/businesslogic.go`

### Task 13.4: Update KMS application_core.go
- **Status**: ❌ Not Started
- **Estimated**: 1.5h
- **Actual**:
- **Dependencies**: Task 13.3
- **Description**: Update application_core to use template barrier
- **Acceptance Criteria**:
  - [ ] Import changed to template barrier
  - [ ] BarrierService obtained from ServiceResources
- **Files**:
  - `internal/kms/server/application/application_core.go`

### Task 13.5: Update KMS application_basic.go
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: Task 13.4
- **Description**: Update unseal service to use template pattern
- **Acceptance Criteria**:
  - [ ] Import changed from `shared/barrier/unsealkeysservice`
  - [ ] Uses template unseal pattern
- **Files**:
  - `internal/kms/server/application/application_basic.go`

### Task 13.6: Delete orm_barrier_adapter.go
- **Status**: ❌ Not Started
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: Tasks 13.3-13.5
- **Description**: Remove unused adapter file
- **Acceptance Criteria**:
  - [ ] `internal/kms/server/barrier/orm_barrier_adapter.go` deleted
  - [ ] No references remain
- **Files**:
  - `internal/kms/server/barrier/orm_barrier_adapter.go` (DELETE)

### Task 13.7: Verify Zero shared/barrier Imports
- **Status**: ❌ Not Started
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: Task 13.6
- **Description**: Verify no shared/barrier imports remain
- **Acceptance Criteria**:
  - [ ] `grep -r "shared/barrier" internal/kms/` returns empty
  - [ ] Build passes
- **Evidence**: grep output

### Task 13.8: Run KMS Tests
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: Task 13.7
- **Description**: Verify all KMS tests pass after migration
- **Acceptance Criteria**:
  - [ ] `go test ./internal/kms/... -count=1` passes
  - [ ] No test regressions
- **Evidence**: test output

### Task 13.9: Delete shared/barrier Directory
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Task 13.8
- **Description**: Remove deprecated shared/barrier
- **Acceptance Criteria**:
  - [ ] `rm -rf internal/shared/barrier/` executed
  - [ ] `go build ./...` passes
  - [ ] All tests pass
- **Evidence**: Directory gone, build clean

---

## Phase 14: Post-Mortem and Documentation Audit

### Task 14.1: Run lint-ports Full Validation
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Phase 12 complete
- **Description**: Run lint-ports across entire codebase
- **Acceptance Criteria**:
  - [ ] `go run ./internal/cmd/cicd lint-ports` passes
  - [ ] Zero violations
- **Evidence**: lint output

### Task 14.2: Verify All Health Paths
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Phases 9-13 complete
- **Description**: Verify all services use `/admin/api/v1/livez`
- **Acceptance Criteria**:
  - [ ] sm-kms: `/admin/api/v1/livez`
  - [ ] cipher-im: `/admin/api/v1/livez`
  - [ ] jose-ja: `/admin/api/v1/livez`
  - [ ] pki-ca: `/admin/api/v1/livez`
- **Evidence**: grep output from compose files

### Task 14.3: Update analysis-overview.md Final State
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: Tasks 14.1-14.2
- **Description**: Update analysis with verified final state
- **Acceptance Criteria**:
  - [ ] Section 11 shows new port assignments
  - [ ] Section 14 shows barrier migration complete
  - [ ] All sections accurate
- **Files**:
  - `docs/fixes-needed-plan-tasks-v8/analysis-overview.md`

### Task 14.4: Update analysis-thorough.md
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: Task 14.3
- **Description**: Update thorough analysis with implementation details
- **Acceptance Criteria**:
  - [ ] All sections match implementation
  - [ ] Code samples accurate
- **Files**:
  - `docs/fixes-needed-plan-tasks-v8/analysis-thorough.md`

### Task 14.5: Full Build and Test Verification
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: Tasks 14.3-14.4
- **Description**: Final verification of entire codebase
- **Acceptance Criteria**:
  - [ ] `go build ./...` passes
  - [ ] `go test ./... -count=1` passes
  - [ ] `golangci-lint run` passes
- **Evidence**: command outputs

### Task 14.6: Commit Comprehensive Audit Trail
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Task 14.5
- **Description**: Create final commit with audit summary
- **Acceptance Criteria**:
  - [ ] Conventional commit message
  - [ ] Lists all phases completed
  - [ ] References task IDs
- **Evidence**: git log

---

## Summary Statistics (Updated 2026-02-14)

| Phase | Tasks | Completed | Percentage | Notes |
|-------|-------|-----------|------------|-------|
| ~~Phase 1: Barrier Integration~~ | ~~5~~ | N/A | N/A | SUPERSEDED by Phase 13 |
| Phase 2: Testing | 4 | 0 | 0% | |
| Phase 3: Documentation | 3 | 0 | 0% | |
| Phase 3.5: Realm Verification | 5 | 1 | 20% | |
| Phase 4: Delete shared/barrier | 2 | 0 | 0% | |
| Phase 5: Mutation Testing | 2 | 0 | 0% | |
| Phase 6: sm-kms Structure | 6 | 0 | 0% | |
| Phase 7: jose-ja Consolidation | 4 | 0 | 0% | |
| Phase 8: pki-ca Renaming | 3 | 0 | 0% | |
| Phase 9: pki-ca Health Paths | 4 | 0 | 0% | |
| Phase 10: jose-ja Admin Port | 4 | 0 | 0% | |
| Phase 11: Port Standardization | 6 | 0 | 0% | |
| Phase 12: CICD lint-ports | 8 | 0 | 0% | |
| Phase 13: KMS Direct Migration | 9 | 0 | 0% | Replaces Phase 1 |
| Phase 14: Post-Mortem | 6 | 0 | 0% | |
| **Total** | **66** | **1** | **1.5%** | Phase 1 excluded |

**Execution Order**: Phase 13 (KMS barrier) → Phase 2-5 → Phase 6-8 → Phase 9-12 → Phase 14
