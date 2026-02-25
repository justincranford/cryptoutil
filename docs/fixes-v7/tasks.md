# Remaining Tasks - fixes-v7 Carryover

**Source**: Archived fixes-v7 (218/220 complete, 2 blocked by OTel Docker socket)

## E2E OTel Collector Fix

- [ ] Fix OTel collector config: remove `docker` from `resourcedetection` detectors
  - File: `deployments/shared-telemetry/otel/otel-collector-config.yaml` line 79
  - Change: `detectors: [env, docker, system]` → `detectors: [env, system]`
- [ ] Verify E2E: `go test -tags=e2e -timeout=30m ./internal/apps/sm/im/e2e/...` passes
- [ ] Verify E2E: sm-im E2E passes end-to-end
- [ ] Verify: deployment validators still pass after OTel config change

## Completion Criteria

- [ ] All E2E tests pass without OTel Docker socket dependency
- [ ] Deployment validators: all pass
- [ ] Build clean: `go build ./...` and `go build -tags e2e,integration ./...`
- [ ] Lint clean: `golangci-lint run` and `golangci-lint run --build-tags e2e,integration`
