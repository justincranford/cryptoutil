# Tasks: PKI-CA-MERGE0a

**Option**: Move sm-im to SM product as sm-im (standalone service)
**Status**: Research Only
**Created**: 2026-02-23
**Recommendation**: ⭐⭐⭐⭐⭐ (Strongly recommended)
**Prerequisites**: None (can be done immediately, independent of other migration work)

---

## Phase 1: Code Rename (sm-im)

### Task 1.1: Move internal/apps directory
- **Status**: ❌
- **Estimated**: 30min
- **Description**: Move `internal/apps/sm/im/` → `internal/apps/sm/im/`. Update package declarations from `package im` (unchanged) and all internal package comments.
- **Commands**:
  ```bash
  mkdir -p internal/apps/sm/im
  git mv internal/apps/sm/im/ internal/apps/sm/
  # If sm/ has only im/, also:
  git rm -r internal/ap../sm/  # if empty after move
  ```
- **Acceptance Criteria**:
  - [ ] `internal/apps/sm/im/` exists with all sm-im files
  - [ ] `internal/apps/sm/im/` no longer exists
  - [ ] `go build ./internal/apps/sm/im/...` passes (import paths not yet updated)

### Task 1.2: Update all import paths
- **Status**: ❌
- **Estimated**: 30min
- **Description**: Find-replace all `cryptoutil/internal/apps/sm/im` → `cryptoutil/internal/apps/sm/im` across the entire codebase. Use `sed` or IDE refactor.
- **Commands**:
  ```bash
  find . -name "*.go" -exec sed -i 's|cryptoutil/internal/apps/sm/im|cryptoutil/internal/apps/sm/im|g' {} +
  ```
- **Acceptance Criteria**:
  - [ ] `grep -r "apps/sm/im" --include="*.go" .` returns no results
  - [ ] `go build ./...` passes

### Task 1.3: Rename cmd/sm-im/ → cmd/sm-im/
- **Status**: ❌
- **Estimated**: 15min
- **Description**: Rename command directory. Update main.go content to import from new path.
- **Commands**:
  ```bash
  git mv cmd/sm-im/ cmd/sm-im/
  sed -i 's|apps/sm/im|apps/sm/im|g' cmd/sm-im/main.go
  ```
- **Acceptance Criteria**:
  - [ ] `cmd/sm-im/main.go` exists and imports from `internal/apps/sm/im/`
  - [ ] `go build ./cmd/sm-im/...` passes

### Task 1.4: Update c../sm/ → cmd/sm/ routing
- **Status**: ❌
- **Estimated**: 30min
- **Description**: `c../sm/main.go` routes to im sub-service. Update to route to sm/im. Update `cmd/sm/main.go` (currently routes to kms only) to also route to im. Remove SM product CMD or leave stub.
- **Acceptance Criteria**:
  - [ ] `cmd/sm/main.go` can route to both kms and im sub-commands
  - [ ] `go build ./cmd/sm/...` passes

---

## Phase 2: Deployment + Config Rename

### Task 2.1: Move deployments/sm-im/ → deployments/sm-im/
- **Status**: ❌
- **Estimated**: 15min
- **Description**: Rename deployment directory. Update Docker Compose service names from `sm-im` → `sm-im`. Update internal Docker DNS references.
- **Commands**:
  ```bash
  git mv deployments/sm-im/ deployments/sm-im/
  sed -i 's/sm-im/sm-im/g' deployments/sm-im/compose.yml
  ```
- **Acceptance Criteria**:
  - [ ] `deployments/sm-im/compose.yml` exists with service name `sm-im`
  - [ ] `docker compose -f deployments/sm-im/compose.yml config` passes

### Task 2.2: Move deploymen../sm/ → update deployments/sm/
- **Status**: ❌
- **Estimated**: 15min
- **Description**: Update `deploymen../sm/compose.yml` to remove sm-im reference, or move to `deployments/sm/`. Update `deployments/sm/compose.yml` to include sm-im.

### Task 2.3: Move configs/sm/im/ → configs/sm/im/
- **Status**: ❌
- **Estimated**: 15min
- **Description**: Move config files. Update any service name references within YAML.
- **Commands**:
  ```bash
  mkdir -p configs/sm/im
  git mv configs/sm/im/ configs/sm/
  ```
- **Acceptance Criteria**:
  - [ ] `configs/sm/im/config-pg-1.yml`, `config-sqlite.yml`, `im.yml` exist
  - [ ] Deployment validator passes: `go run ./cmd/cicd lint-deployments validate-all`

---

## Phase 3: Documentation + CI Updates

### Task 3.1: Update ARCHITECTURE.md
- **Status**: ❌
- **Estimated**: 1h
- **Description**: Multiple table updates:
  - Service catalog table: Remove sm-im row, add sm-im row under SM
  - SM product section (formerly Cipher): Mark dissolved or reduce to stub
  - SM product section: Add sm-im service
  - Port assignment table: Move 8700-8799 from SM/sm-im to SM/sm-im
  - Directory tree: Update `internal/apps/` layout, `cmd/` layout, `deployments/`, `configs/`
  - Migration order: Update to reflect sm-im → sm-im rename
- **Acceptance Criteria**:
  - [ ] No references to `sm-im` in product catalog sections
  - [ ] SM product shows sm-kms + sm-im

### Task 3.2: Update ci-e2e.yml
- **Status**: ❌
- **Estimated**: 15min
- **Description**: Update sm-im E2E job to reference `sm-im` service name and `deployments/sm-im/` path.
- **Acceptance Criteria**:
  - [ ] `grep -n "sm-im" .github/workflows/ci-e2e.yml` returns no results
  - [ ] E2E job references `deployments/sm-im/compose.yml`

### Task 3.3: Update any remaining references
- **Status**: ❌
- **Estimated**: 15min
- **Description**: Scan for remaining `sm-im` or `sm/im` references in non-code files (README.md, docs/, comments).
- **Commands**:
  ```bash
  grep -r "sm-im\|sm/im" --include="*.md" --include="*.yml" --include="*.yaml" . | grep -v ".git"
  ```

---

## Phase 4 (Optional): jose-ja → sm-jwk

If jose-ja → sm-jwk rename is done simultaneously:

### Task 4.1: Move internal/apps/jose/ja/ → internal/apps/sm/jwk/
- **Estimated**: 30min

### Task 4.2: Update jose-ja import paths to sm/jwk
- **Estimated**: 30min

### Task 4.3: Rename cmd/jose-ja/ → cmd/sm-jwk/
- **Estimated**: 15min

### Task 4.4: Move deployments/jose-ja/ → deployments/sm-jwk/ + configs/jose/ → configs/sm/jwk/
- **Estimated**: 15min

### Task 4.5: ARCHITECTURE.md: dissolve JOSE product, add sm-jwk to SM
- **Estimated**: 30min

---

## Phase 5: Validation

### Task 5.1: Full build and test
- **Status**: ❌
- **Estimated**: 30min
- **Commands**:
  ```bash
  go build ./...
  go build -tags e2e,integration ./...
  go test ./internal/apps/sm/im/...
  golangci-lint run ./internal/apps/sm/im/...
  ```
- **Acceptance Criteria**:
  - [ ] Zero build errors
  - [ ] Zero lint errors
  - [ ] All sm/im unit tests pass

### Task 5.2: Deployment validator
- **Status**: ❌
- **Estimated**: 10min
- **Commands**:
  ```bash
  go run ./cmd/cicd lint-deployments validate-all
  ```

---

## Summary Stats

| Phase | Tasks | Est Effort |
|-------|-------|-----------|
| 1: Code rename | 4 | 1.75h |
| 2: Deployment + config | 3 | 0.75h |
| 3: Docs + CI | 3 | 1.5h |
| 4 (optional): jose-ja rename | 5 | 2h |
| 5: Validation | 2 | 0.67h |
| **Total (sm-im only)** | **12 tasks** | **~4.5h** |
| **Total (sm-im + sm-jwk)** | **17 tasks** | **~6.5h** |

**No prerequisites required.** This can be done before ANY other fixes-v7 work.
