# Tasks — Framework v9: Quality & Consistency

**Status**: 0 of 31 tasks complete (0%)
**Created**: 2026-04-08
**Updated**: 2026-04-10

---

## Phase 1: Dockerfile & Deployment Fixes (Items 1, 3, 4)

### Task 1.1: Standardize EXPOSE in all Dockerfiles

- [ ] Update `deployments/sm-im/Dockerfile` EXPOSE to `8080 9090`
- [ ] Update `deployments/sm-kms/Dockerfile` EXPOSE to `8080 9090`
- [ ] Update `deployments/jose-ja/Dockerfile` EXPOSE to `8080 9090`
- [ ] Update `deployments/pki-ca/Dockerfile` EXPOSE to `8080 9090`
- [ ] Update `deployments/skeleton-template/Dockerfile` EXPOSE to `8080 9090`
- [ ] Update `deployments/identity-authz/Dockerfile` EXPOSE to `8080 9090`
- [ ] Update `deployments/identity-idp/Dockerfile` EXPOSE to `8080 9090`
- [ ] Update `deployments/identity-rp/Dockerfile` EXPOSE to `8080 9090`
- [ ] Update `deployments/identity-rs/Dockerfile` EXPOSE to `8080 9090`
- [ ] Update `deployments/identity-spa/Dockerfile` EXPOSE to `8080 9090`

### Task 1.2: Pin Docker image versions in compose files

- [ ] Replace `alpine:latest` with `alpine:3.19` in all 11 healthcheck-secrets services
- [ ] Pin `otel/opentelemetry-collector-contrib` to specific version in shared-telemetry
- [ ] Pin `grafana/otel-lgtm` to specific version in shared-telemetry

### Task 1.3: Standardize Dockerfile healthchecks

- [ ] Set `--start-period=30s` in all Dockerfiles
- [ ] Replace sm-kms custom healthcheck with standard `wget` pattern
- [ ] Verify all healthchecks use `https://127.0.0.1:9090/admin/api/v1/livez`

### Task 1.4: Run E2E validation

- [ ] Run `go run ./cmd/cicd-lint lint-deployments` — all validators pass
- [ ] Build all Docker images and verify startup with `docker compose up`

## Phase 2: Config Key Naming Migration (Item 2)

### Task 2.1: Audit config parsers

- [ ] Identify Go code that parses snake_case config keys for each affected service
- [ ] Determine if services use framework config parser or custom parsers
- [ ] Document required code changes per service

### Task 2.2: Migrate configs/sm-kms/ to kebab-case

- [ ] Update `configs/sm-kms/sm-kms.yml` keys to kebab-case
- [ ] Update `deployments/sm-kms/config/` overlay files to kebab-case
- [ ] Update Go parser code in `internal/apps/sm-kms/`
- [ ] Verify service starts and tests pass

### Task 2.3: Migrate configs/sm-im/ to kebab-case

- [ ] Update configs, deployment overlays, and Go parser code
- [ ] Verify service starts and tests pass

### Task 2.4: Migrate identity service configs to kebab-case

- [ ] Update configs for identity-authz, identity-idp, identity-rp, identity-rs, identity-spa
- [ ] Update deployment overlays for all 5 identity services
- [ ] Update Go parser code in `internal/apps/identity-*/`
- [ ] Verify all identity services start and tests pass

## Phase 3: Linter Configuration (Items 5, 6, 13)

### Task 3.1: Resolve testpackage linter

- [ ] Audit which packages can use external test packages
- [ ] If migration feasible: narrow skip-regexp and migrate tests
- [ ] If migration too large: remove testpackage from enabled linters
- [ ] Update §11.3.1 documentation with resolution

### Task 3.2: Monitor goheader golangci-lint v2.8+

- [ ] Check golangci-lint releases for v2.8+ with goheader fix
- [ ] If available: test on branch, re-enable if fixed
- [ ] Update §11.3.1 documentation

### Task 3.3: Add unpinned image tag fitness linter

- [ ] Implement sub-linter in `lint_fitness/` that scans compose files
- [ ] Detect `:latest` tags and missing version pins
- [ ] Exclude `image: cryptoutil:local` (local build)
- [ ] Add tests with >=98% coverage
- [ ] Update §9.11 fitness linter catalog

## Phase 4: Test Quality (Items 7, 8, 9, 10)

### Task 4.1: Implement jose-ja P2.4 skipped tests

- [ ] Implement FK constraint tests in `material_jwk_repository_error_test.go`
- [ ] Implement FK constraint tests in `elastic_jwk_repository_error_test.go`
- [ ] Implement mocked database tests for error scenarios
- [ ] Implement cascade deletion tests
- [ ] Remove all `t.Skip("TODO P2.4: ...")` calls

### Task 4.2: Resolve Phase W migration TODOs

- [ ] Evaluate `StartCoreWithServices` migration handling
- [ ] Implement migration handling in startup OR document design decision
- [ ] Remove TODO comments from `application_listener_test.go`

### Task 4.3: Integrate rate limiter in identity-idp

- [ ] Wire framework `RateLimiter` into identity-idp handler chain
- [ ] Configure per §8.5.2 two-layer rate limiting
- [ ] Add tests for rate limiting behavior
- [ ] Remove deferred TODO from `handlers_security_validation_rate_test.go`

### Task 4.4: Refactor oversized test files

- [ ] Split `validate_chunks_test.go` (544 lines) into <=500 line files
- [ ] Split `jose_seam_injection_test.go` (509 lines) into <=500 line files
- [ ] Split `issuer_operations_test.go` (501 lines) into <=500 line files

## Phase 5: Low-Priority Improvements (Items 11, 12, 14)

### Task 5.1: Extend Gatling load tests

- [ ] Add product-level simulation classes (5 products)
- [ ] Add suite-level simulation class
- [ ] Update `pom.xml` with new entry points

### Task 5.2: Increase ENG-HANDBOOK propagation coverage

- [ ] Identify 10 most-referenced orphaned sections
- [ ] Add `@propagate`/`@source` blocks for selected sections
- [ ] Run `lint-docs` to verify

### Task 5.3: Replace context.TODO() in migrations

- [ ] Pass startup context through to migration functions
- [ ] Remove `context.TODO()` from `identity/repository/migrations.go`
