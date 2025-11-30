# Passthru2 Evidence & Completion Checklist

Use this checklist to validate that tasks are complete and that acceptance criteria are met.

## Demo & DX
- [ ] `make demo-kms` or `docker compose -f deployments/telemetry/compose.yml -f deployments/kms/compose.demo.yml up` boots KMS with seeded accounts
- [ ] `make demo-identity` or `docker compose -f deployments/telemetry/compose.yml -f deployments/identity/compose.demo.yml up` boots Identity with seeded users
- [ ] Both demos run with `demo` profile and health checks pass

## CI & Tests
- [ ] Per-product `go test` runs with coverage ≥ target (KMS & Identity 85%)
- [ ] infra/test coverage ≥ 95%
- [ ] Lint and formatting checks pass (`golangci-lint run --fix` and gofumpt applied)

## Security
- [ ] Password & client secret hashing uses PBKDF2
- [ ] PKCE enforced for public clients
- [ ] TLS endpoints use valid certs or `localhost` debug certs
- [ ] Docker secrets in place for DB credentials

## Migration
- [ ] Telemetry extracted into `deployments/telemetry/compose.yml` and used by both products
- [ ] `deployments/<product>/config/` standardized across products
- [ ] No duplicate infra packages remain (or a plan to migrate them)

## Documentation
- [ ] `docs/03-products/passthru2/README.md` includes step-by-step quickstart
- [ ] `docs/03-products/passthru2/TASK-LIST.md` updated
- [ ] `grooming/GROOMING-QUESTIONS.md` answered and committed

**Sign-off**: commit reviewer, tests green, and coverage pass
