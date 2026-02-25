# Remaining Work Plan - fixes-v7 Carryover

**Status**: Active
**Created**: 2026-02-25
**Source**: Archived from fixes-v7 (218/220 tasks complete, 2 blocked)

## Background

The fixes-v7 plan completed 218 of 220 tasks across 11 phases (8 original + 3 added).
Two tasks remain blocked by a pre-existing OTel collector Docker socket issue.

## Blocked Work: E2E OTel Collector

**Root Cause**: `deployments/shared-telemetry/otel/otel-collector-config.yaml` line 79
uses `resourcedetection` processor with `detectors: [env, docker, system]`. The `docker`
detector requires `/var/run/docker.sock` mounted inside the OTel collector container.
Without it, the collector fails to start, blocking the entire E2E service chain.

**Fix Options** (mutually exclusive, pick one):

1. **Remove `docker` detector** — Change `detectors: [env, docker, system]` to
   `detectors: [env, system]`. Loses Docker metadata enrichment but unblocks E2E.
2. **Mount Docker socket** — Add `volumes: ["/var/run/docker.sock:/var/run/docker.sock:ro"]`
   to the OTel collector service in compose files. Preserves metadata but requires
   socket access in CI/CD runners.
3. **Conditional config** — Use separate OTel configs for dev (no docker) vs prod (docker).

**Recommendation**: Option 1 (remove `docker` detector). Docker metadata enrichment
is low-value for this project; system and env detectors provide sufficient context.

## Tasks

1. Fix OTel collector config (remove `docker` detector or mount socket)
2. Verify E2E tests pass: `go test -tags=e2e -timeout=30m ./internal/apps/sm/im/e2e/...`
3. Verify sm-im E2E passes end-to-end
4. Update archived tasks.md completion criteria
