# Framework v4 - Task Checklist

---

## Quality Mandate

**Every task must satisfy ALL gates before it can be checked off:**

- [ ] `go build ./...` exits 0
- [ ] `go build -tags e2e,integration ./...` exits 0
- [ ] `golangci-lint run --fix` exits 0
- [ ] `golangci-lint run --build-tags e2e,integration` exits 0
- [ ] `go test ./...` exits 0 â€” 100% pass, zero skips
- [ ] `go run ./cmd/cicd lint-fitness` exits 0 â€” all checks, including newly added ones
- [ ] `go run ./cmd/cicd lint-deployments` exits 0 â€” 68 deployment validators
- [ ] Coverage maintained or improved
- [ ] Conventional commit created
- [ ] `git status --porcelain` returns empty

**No task is done merely because "it works." Evidence of all gates passing is required.**

---

## Phase 1: Fix Legacy sm-kms-pg- Naming and Add OTLP Service Name Check

**Status**: âœ… COMPLETE (commit `dc5970d47`, lessons commit `e9be1a7d3`)

- [x] 1.1 Fix `configs/sm/kms/config-pg-1.yml`: update comment header and `otlp-service: "sm-kms-pg-1"` â†’ `"sm-kms-postgres-1"`
- [x] 1.2 Fix `configs/sm/kms/config-pg-2.yml`: update comment header and `otlp-service: "sm-kms-pg-2"` â†’ `"sm-kms-postgres-2"`
- [x] 1.3 Verify `configs/sm/im/config-pg-1.yml` uses `"sm-im-postgres-1"` (already done in prior session, confirm unchanged)
- [x] 1.4 Verify `configs/sm/im/config-pg-2.yml` uses `"sm-im-postgres-2"` (already done in prior session, confirm unchanged)
- [x] 1.5 Create `internal/apps/cicd/lint_fitness/otlp_service_name_pattern/` directory and implementation
- [x] 1.6 Implement `Check(logger)` function: for each `configs/{PRODUCT}/{SERVICE}/config-*.yml`, parse `otlp-service` key, verify matches `{PS-ID}-sqlite-1`, `{PS-ID}-postgres-1`, or `{PS-ID}-postgres-2` pattern
- [x] 1.7 Register `otlp-service-name-pattern` in `internal/apps/cicd/lint_fitness/lint_fitness.go`
- [x] 1.8 Add unit tests for the check with positive and negative cases using table-driven tests
- [x] 1.9 Run `go run ./cmd/cicd lint-fitness` â€” verify `otlp-service-name-pattern` passes
- [x] 1.10 Commit Phase 1 changes
- [x] 1.11 Update lessons.md with Phase 1 post-mortem

## Phase 2: Registry-Driven Foundation and Entity Registry Check

- [x] 2.1 Design Go struct schema for entity registry (Product, ProductService, Suite)
- [x] 2.2 Create `internal/apps/cicd/lint_fitness/registry/registry.go` with canonical entity registry for all 5 products, 10 product-services, 1 suite
- [x] 2.3 Add unit tests for registry package (validate all 16 entities are present, all fields non-empty)
- [x] 2.4 Create `internal/apps/cicd/lint_fitness/entity_registry_completeness/` directory and implementation
- [x] 2.5 Implement `Check(logger)` function: for each entity in registry, verify deployment dir, config dir, and magic file exist on disk
- [x] 2.6 Register `entity-registry-completeness` in `lint_fitness.go`
- [x] 2.7 Add unit tests (table-driven: present entity passes, missing entity fails)
- [x] 2.8 Migrate existing hardcoded PS-ID lists in other checks to use registry package (reduce duplication)
- [x] 2.9 Run `go run ./cmd/cicd lint-fitness` â€” verify `entity-registry-completeness` passes
- [x] 2.10 Commit Phase 2 changes
- [x] 2.11 Update lessons.md with Phase 2 post-mortem

## Phase 3: Banned Name Detection

- [x] 3.1 Finalize banned phrase list (exact strings to ban â€” not substrings of `cipher`)
- [x] 3.2 Create `internal/apps/cicd/lint_fitness/banned_product_names/` directory and implementation
- [x] 3.3 Implement `Check(logger)`: scan `.go`, `.yml`, `.yaml`, `.sql`, `.md` files for banned phrases (exact match), report file+line
- [x] 3.4 Implement exclusion rule: crypto terminology like `cipher.Block`, `ciphertext` allowed (not the exact banned phrases)
- [x] 3.5 Add unit tests: at minimum 1 test per banned phrase showing detection, 1 test showing exclusion
- [x] 3.6 Register `banned-product-names` in `lint_fitness.go`
- [x] 3.7 Create `internal/apps/cicd/lint_fitness/legacy_dir_detection/` directory and implementation
- [x] 3.8 Implement `Check(logger)`: verify `internal/apps/cipher/` does not exist; verify no `*-cipher-*` directories in `deployments/`, `configs/`, `cmd/`
- [x] 3.9 Register `legacy-dir-detection` in `lint_fitness.go`
- [x] 3.10 Run `go run ./cmd/cicd lint-fitness` â€” verify both checks pass
- [x] 3.11 Commit Phase 3 changes
- [x] 3.12 Update lessons.md with Phase 3 post-mortem

## Phase 4: Deployment Directory Completeness

- [x] 4.1 Create `internal/apps/cicd/lint_fitness/deployment_dir_completeness/` directory and implementation
- [x] 4.2 Implement `Check(logger)`: for each PS in registry, verify Dockerfile, compose.yml, secrets/, config/ exist under `deployments/{PS-ID}/`
- [x] 4.3 Verify config subdir contains: `{PS-ID}-app-common.yml`, `{PS-ID}-app-sqlite-1.yml`, `{PS-ID}-app-postgresql-1.yml`, `{PS-ID}-app-postgresql-2.yml`
- [x] 4.4 Report missing files clearly: `deployments/sm-im/config/sm-im-app-postgresql-2.yml: missing`
- [x] 4.5 Add unit tests (table-driven: all files present passes; each missing file type fails independently)
- [x] 4.6 Register `deployment-dir-completeness` in `lint_fitness.go`
- [x] 4.7 Fix any missing deployment config files discovered during check development
- [x] 4.8 Run `go run ./cmd/cicd lint-fitness` â€” verify `deployment-dir-completeness` passes for all 10 PS
- [x] 4.9 Commit Phase 4 changes
- [x] 4.10 Update lessons.md with Phase 4 post-mortem

## Phase 5: Compose File Header and Service Name Validation

- [x] 5.1 Create `internal/apps/cicd/lint_fitness/compose_header_format/` directory and implementation
- [x] 5.2 Implement `Check(logger)`: for each PS in registry, read first 7 lines of `deployments/{PS-ID}/compose.yml`, verify lines 3 and 5 match expected patterns
- [x] 5.3 Add unit tests (generates minimal compose.yml with correct/incorrect headers in temp dir)
- [x] 5.4 Register `compose-header-format` in `lint_fitness.go`
- [x] 5.5 Fix any compose files with non-conforming headers discovered during check development
- [x] 5.6 Create `internal/apps/cicd/lint_fitness/compose_service_names/` directory and implementation
- [x] 5.7 Implement `Check(logger)`: parse each `deployments/{PS-ID}/compose.yml` with yaml library, verify top-level service keys include all required names
- [x] 5.8 Register `compose-service-names` in `lint_fitness.go`
- [x] 5.9 Create `internal/apps/cicd/lint_fitness/compose_db_naming/` directory and implementation
- [x] 5.10 Implement `Check(logger)`: parse compose.yml, verify `{PS-ID}-db-postgres-1` service has `container_name: {PS-ID}-postgres` and `hostname: {PS-ID}-postgres`
- [x] 5.11 Register `compose-db-naming` in `lint_fitness.go`
- [x] 5.12 Add unit tests for both new checks
- [x] 5.13 Run `go run ./cmd/cicd lint-fitness` â€” verify all three Phase 5 checks pass
- [x] 5.14 Commit Phase 5 changes
- [x] 5.15 Update lessons.md with Phase 5 post-mortem

## Phase 6: Magic Constants Cross-Reference Validation

- [x] 6.1 Create `internal/apps/cicd/lint_fitness/magic_e2e_container_names/` directory and implementation
- [x] 6.2 Implement `Check(logger)`: for each PS in registry, parse `internal/shared/magic/magic_*.go` for `*E2ESQLiteContainer`, `*E2EPostgreSQL1Container`, `*E2EPostgreSQL2Container` constant values using Go AST or regex
- [x] 6.3 Cross-reference: parsed container name constants must match expected compose service names (`{PS-ID}-app-sqlite-1`, etc.)
- [x] 6.4 Add unit tests (in-memory Go source: correct constant passes, wrong constant fails)
- [x] 6.5 Register `magic-e2e-container-names` in `lint_fitness.go`
- [x] 6.6 Create `internal/apps/cicd/lint_fitness/magic_e2e_compose_path/` directory and implementation
- [x] 6.7 Implement `Check(logger)`: verify `*E2EComposeFile` constant value points to an existing file relative to the magic file's location
- [x] 6.8 Register `magic-e2e-compose-path` in `lint_fitness.go`
- [x] 6.9 Add unit tests for compose path check
- [x] 6.10 Run `go run ./cmd/cicd lint-fitness` â€” verify both checks pass
- [x] 6.11 Commit Phase 6 changes
- [x] 6.12 Update lessons.md with Phase 6 post-mortem

## Phase 7: Standalone Config File Presence and Naming

- [ ] 7.1 Establish allowlist of PS IDs that have standardized standalone configs: `sm-im`, `sm-kms`
- [ ] 7.2 Create `internal/apps/cicd/lint_fitness/standalone_config_presence/` directory and implementation
- [ ] 7.3 Implement `Check(logger)`: for each PS in standalone allowlist, verify `config-sqlite.yml`, `config-pg-1.yml`, `config-pg-2.yml` exist under `configs/{PRODUCT}/{SERVICE}/`
- [ ] 7.4 Register `standalone-config-presence` in `lint_fitness.go`
- [ ] 7.5 Create `internal/apps/cicd/lint_fitness/standalone_config_otlp_names/` directory and implementation
- [ ] 7.6 Implement `Check(logger)`: parse YAML, extract `otlp-service` value, verify against expected pattern
- [ ] 7.7 Register `standalone-config-otlp-names` in `lint_fitness.go`
- [ ] 7.8 Add unit tests (table-driven for both checks with positive and negative cases)
- [ ] 7.9 Run `go run ./cmd/cicd lint-fitness` â€” verify both checks pass after Phase 1 fixes
- [ ] 7.10 Commit Phase 7 changes
- [ ] 7.11 Update lessons.md with Phase 7 post-mortem

## Phase 8: Migration Comment Header Validation

- [ ] 8.1 Create `internal/apps/cicd/lint_fitness/migration_comment_headers/` directory and implementation
- [ ] 8.2 Implement `Check(logger)`: for each PS in registry that has `internal/apps/{PRODUCT}/{SERVICE}/repository/migrations/`, scan `*.up.sql` â€” first comment line must contain `{Display Name} database schema`
- [ ] 8.3 Also check `*.down.sql` â€” first comment line must contain `{Display Name} database schema rollback`
- [ ] 8.4 Skip framework migration files (1001-1999 range) â€” only domain migrations (2001+) are validated
- [ ] 8.5 Add unit tests (in-memory SQL files with correct/incorrect headers)
- [ ] 8.6 Register `migration-comment-headers` in `lint_fitness.go`
- [ ] 8.7 Fix any migration files with non-conforming headers discovered
- [ ] 8.8 Run `go run ./cmd/cicd lint-fitness` â€” verify `migration-comment-headers` passes
- [ ] 8.9 Commit Phase 8 changes
- [ ] 8.10 Update lessons.md with Phase 8 post-mortem

## Phase 9: ARCHITECTURE.md Updates and CICD Tool Catalog

- [ ] 9.1 Count total fitness checks after all phases complete
- [ ] 9.2 Update ARCHITECTURE.md Section 9.11 count from "23 total" (or current stale value) to new total
- [ ] 9.3 Add all new fitness checks to the sub-linter catalog table in ARCHITECTURE.md Section 9.11.1
- [ ] 9.4 Add "Entity Registry" sub-section to ARCHITECTURE.md Section 9.11: location, structure, update procedure
- [ ] 9.5 Add "Naming Convention Catalog" reference to ARCHITECTURE.md with pointer to `plan.md` tables
- [ ] 9.6 Update `cicd-lint-fitness` workflow description (if separate workflow file exists) to mention expanded scope
- [ ] 9.7 Run `go run ./cmd/cicd lint-docs` â€” verify ARCHITECTURE.md propagation passes
- [ ] 9.8 Run `go run ./cmd/cicd lint-fitness` â€” ALL checks pass (final full suite run)
- [ ] 9.9 Commit Phase 9 changes
- [ ] 9.10 Update lessons.md with Phase 9 post-mortem

## Phase 10: Knowledge Propagation

- [ ] 10.1 Update `docs/framework-v3/plan.md` Status to COMPLETE (if not already)
- [ ] 10.2 Propagate entity registry pattern and banned name list to `.github/instructions/02-01.architecture.instructions.md`
- [ ] 10.3 Update `fitness-function-gen` skill to document registry-driven check pattern for new contributors
- [ ] 10.4 Verify all `@source` propagation blocks in instruction files match ARCHITECTURE.md after updates
- [ ] 10.5 Run full quality gate suite one final time:
  - `go build ./...`
  - `go build -tags e2e,integration ./...`
  - `golangci-lint run`
  - `golangci-lint run --build-tags e2e,integration`
  - `go test ./... -shuffle=on`
  - `go run ./cmd/cicd lint-fitness`
  - `go run ./cmd/cicd lint-deployments`
  - `go run ./cmd/cicd lint-docs`
- [ ] 10.6 Commit Phase 10 changes
- [ ] 10.7 Update lessons.md with Phase 10 post-mortem
- [ ] 10.8 Mark all phases COMPLETE in plan.md top-level status
