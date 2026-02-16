# Quizme V1 - Deployment & Config Structure Decisions

**Created**: 2026-02-16
**Purpose**: Clarify architectural decisions before implementation

---

## Question 1: Service-Level Demo Compose Files

**Context**: Found `deployments/sm-kms/compose.demo.yml` which appears to be service-specific demo file. Need to determine if this violates patterns.

**Question**: Should service-level demo compose files exist, or should demos only be at suite/product level?

**A)** Delete `sm-kms/compose.demo.yml` - Demos should only be at suite/product level, not service level
**B)** Keep `sm-kms/compose.demo.yml` - Service-level demos are acceptable for standalone testing
**C)** Convert to profile in main compose.yml - Use Docker Compose profiles instead of separate files
**D)** Investigate if suite/product level demos exist first, then decide based on that
**E)** 

**Answer**: 

**Rationale**: This affects whether we create demo files for ALL services or centralize demos at higher levels.

---

## Question 2: ./configs/ Structure Design

**Context**: ./configs/ currently has loose structure (55 files, ad-hoc organization). Need to apply rigorous validation like ./deployments/.

**Question**: What rigid structure should ./configs/ follow?

**A)** Exact mirror of ./deployments/ - configs/{cryptoutil,PRODUCT,PRODUCT-SERVICE}/ matching deployments exactly
**B)** Hybrid approach - Suite/product/service dirs BUT keep profiles/ and policies/ subdirectories unique to configs
**C)** Minimal hierarchy - Only service-level, no suite/product aggregation (simpler for CLI)
**D)** Custom design - Different structure optimized specifically for CLI development workflows
**E)** 

**Answer**: 

**Rationale**: This determines restructuring scope (Phase 5) - Option A = massive migration, Option C = minimal changes.

---

## Question 3: Otel-Collector Config in Template

**Context**: Found three otel-collector-config.yaml files:
- `deployments/shared-telemetry/otel/otel-collector-config.yaml` (canonical source)
- `deployments/template/otel-collector-config.yaml` (possibly intentional example)
- `deployments/cipher-im/otel-collector-config.yaml` (likely duplicate)

**Question**: Should `deployments/template/otel-collector-config.yaml` be kept or deleted?

**A)** Delete - Only shared-telemetry should have otel-collector config (single canonical source)
**B)** Keep - Template directory should have example configs for reference
**C)** Keep but document as EXAMPLE ONLY with clear comments
**D)** Compare contents first - Delete only if identical to shared-telemetry version
**E)** 

**Answer**: 

**Rationale**: Template directory may intentionally have reference configs for developers copying patterns.

---

## Question 4: Implementation Priority

**Context**: This is a LARGE refactoring with 7 phases and ~35 hours of estimated work.

**Question**: Do you want me to start implementing immediately, or review plan first?

**A)** Start immediately - Execute all phases autonomously in sequence
**B)** Start with Phase 1 only - Do investigation, then pause for approval before cleanup
**C)** Review plan first - Wait for plan/tasks approval, then start from Phase 1
**D)** Prioritize specific phases - Start with high-priority phases (e.g., CICD refactoring first)
**E)** 

**Answer**: 

**Rationale**: Beast mode says "do all the work" but this is 35+ hours - want to confirm autonomous execution vs checkpoints.

---

## Question 5: Config Restructuring Scope

**Context**: ./configs/ has 55 files across 10 subdirectories. Restructuring means moving ALL files and updating ALL references.

**Question**: How aggressive should the config restructuring be?

**A)** Full restructuring - Move ALL 55 files to new structure, update ALL references, comprehensive migration
**B)** Incremental - Establish structure, move only new files, migrate old files opportunistically
**C)** Minimal - Keep existing files in place, only apply structure to NEW configs going forward
**D)** Phased approach - Critical services first (identity, cipher), others later
**E)** 

**Answer**: 

**Rationale**: Option A = high risk but maximum rigor, Option C = minimal disruption but less consistent.

---

## Instructions

1. Fill in **Answer**: field for each question with A, B, C, D, or E
2. If selecting E, provide your custom answer in the E option text
3. Delete this quizme-v1.md file will be deleted after answers merged into plan.md/tasks.md
4. Agent will proceed with implementation based on answers

---

## Notes

- This quizme covers only decisions that need USER input, not LLM research
- Agent will implement based on clear decisions:
  - Delete .gitkeep in non-empty dirs: PROCEED
  - Keep deployments/compose/compose.yml: PROCEED (E2E infrastructure)
  - Delete cipher-im/otel-collector-config.yaml if duplicate: PROCEED
  - Document template/config/ empty: PROCEED
