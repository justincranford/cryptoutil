# Tasks - Remaining Work (V4)

**Status**: 24 of 115 tasks remaining (20.9% incomplete) - Phases 6, 8 ONLY
**Last Updated**: 2026-01-28
**Priority Order**: KMS Modernization (Phase 6, intentionally LAST) → Template Mutation (Phase 8, DEFERRED)

**Completed Work**: See completed.md (91 of 115 tasks, 79.1%) - Phases 1, 1.5, 2, 3, 4, 5, 7 COMPLETE

**User Feedback**: Phase ordering updated - template quality first, services in architectural order, KMS last to leverage all learnings.

**Note**: Phase 1.5 added to address coverage gap. Phases 0.2, 0.3 to be added for violation remediation.

## Phase 6: KMS Modernization (LAST - Largest Migration)

**Objective**: Migrate KMS to service-template pattern, ≥95% coverage, ≥95% mutation
**Status**: ⏳ NOT STARTED - Tasks TBD after Phases 1-5
**Dependencies**: Phases 1-5 complete (all lessons learned, template proven)

**Note**: KMS is intentionally LAST - it's the largest service, most complex, and should benefit from all learnings from Phases 1-5. Detailed tasks will be defined after completing Phases 1-5.

**Placeholder Tasks**:
- Task 6.1: TBD - Plan KMS migration strategy
- Tasks 6.2-6.N: TBD - Implementation tasks


## Phase 8: Template Mutation Improvement (DEFERRED)

**Objective**: Address remaining template mutation (currently 98.91% efficacy)
**Status**: ⏳ DEFERRED
**Priority**: LOW (template already exceeds 98% ideal)

### Task 8.1: Analyze Remaining TLS Generator Mutation

**Status**: ⏳ DEFERRED
**Owner**: LLM Agent
**Dependencies**: Phase 1 complete
**Priority**: LOW

**Description**: Analyze remaining tls_generator.go mutation.

**Acceptance Criteria**:
- [ ] 8.1.1: Review gremlins output
- [ ] 8.1.2: Identify survived mutation type
- [ ] 8.1.3: Analyze killability
- [ ] 8.1.4: Document findings

**Files**:
- test-output/template-mutation-analysis/ (create)
