# Remaining Work Plan - fixes-v7 Carryover

**Status**: Complete
**Created**: 2026-02-25
**Completed**: 2026-02-25
**Source**: Archived from fixes-v7 (220/220 tasks complete)

## Background

The fixes-v7 plan completed all 220 tasks across 11 phases (8 original + 3 added).

## Resolved: E2E OTel Collector (FIXED)

**Root Cause**: `deployments/shared-telemetry/otel/otel-collector-config.yaml` line 79
used `resourcedetection` processor with `detectors: [env, docker, system]`. The `docker`
detector requires `/var/run/docker.sock` mounted inside the OTel collector container.
Without it, the collector fails to start, blocking the entire E2E service chain.

**Resolution**: Option 1 applied — removed `docker` detector. Changed to `detectors: [env, system]`.
Docker metadata enrichment is low-value for this project; system and env detectors provide sufficient context.

**Lesson**: Infrastructure blockers are ALWAYS MANDATORY BLOCKING. NEVER defer as "pre-existing".
See [ARCHITECTURE.md Section 13.7](../ARCHITECTURE.md#137-infrastructure-blocker-escalation).

## ARCH-SUGGESTIONS Applied

All 8 architecture suggestions from ARCH-SUGGESTIONS.md applied to ARCHITECTURE.md
and propagated to instruction/agent files. See ARCH-SUGGESTIONS.md for details.
