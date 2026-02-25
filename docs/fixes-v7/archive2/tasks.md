# Remaining Tasks - fixes-v7 Carryover

**Source**: Archived fixes-v7 (220/220 complete)
**Status**: All tasks complete

## E2E OTel Collector Fix (COMPLETE)

- [x] Fix OTel collector config: remove `docker` from `resourcedetection` detectors
  - File: `deployments/shared-telemetry/otel/otel-collector-config.yaml` line 79
  - Change: `detectors: [env, docker, system]` → `detectors: [env, system]`
- [x] Verify: deployment validators still pass after OTel config change
- [ ] Verify E2E: `go test -tags=e2e -timeout=30m ./internal/apps/sm/im/e2e/...` passes
- [ ] Verify E2E: sm-im E2E passes end-to-end

> **Note**: E2E verification requires Docker Desktop running. OTel config fix is committed.
> E2E tests should be run when Docker Desktop is available.

## ARCH-SUGGESTIONS Applied (COMPLETE)

- [x] Suggestion 1: Coverage ceiling analysis methodology (Section 10.2.3)
- [x] Suggestion 2: Test seam injection pattern (Section 10.2.4)
- [x] Suggestion 3: OTel processor constraints (Section 9.4.1)
- [x] Suggestion 4: Expand propagation mapping (Section 12.7)
- [x] Suggestion 5: Plan lifecycle management (Section 13.6)
- [x] Suggestion 6: Docker Desktop compatibility (Section 9.4.2)
- [x] Suggestion 7: Agent self-containment requirements (Section 2.1.1)
- [x] Suggestion 8: Infrastructure blocker escalation (Section 13.7)

## Propagation (COMPLETE)

- [x] 03-02.testing.instructions.md: Coverage ceiling + test seam refs
- [x] 02-03.observability.instructions.md: OTel constraints + Docker Desktop refs
- [x] 04-01.deployment.instructions.md: Docker Desktop + infrastructure blocker refs
- [x] 06-01.evidence-based.instructions.md: Coverage ceiling + blocker escalation refs
- [x] 01-02.beast-mode.instructions.md: Infrastructure blocker escalation ref
- [x] 06-02.agent-format.instructions.md: Agent self-containment checklist
- [x] implementation-planning.agent.md: ARCHITECTURE.md cross-reference table (18 refs)
- [x] implementation-execution.agent.md: Test seam + quality gates + infrastructure refs
