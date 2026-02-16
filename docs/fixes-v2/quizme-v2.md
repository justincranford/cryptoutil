# Qlearn Quiz - docs/fixes-v2 Implementation Clarifications v2

**Date**: 2025-02-16
**Purpose**: Resolve architectural unknowns discovered during deep analysis of plan.md/tasks.md

**Instructions**: Read each question, select A-E, write your choice in the **Answer:** field (leave BLANK if E)

---

## Question 1: File Listing Format (Phase 3 Task 3.1)

**Context**: Task 3.1 will create `deployments_all_files.txt` and `configs_all_files.txt` containing comprehensive file listings used by ValidateStructuralMirror.

**Question**: What format should these listing files use?

**A)** Simple newline-separated list of relative paths
```
deployments/cipher-im/compose.yml
deployments/cipher-im/Dockerfile
deployments/cipher-im/config/cipher-im-app-common.yml
```

**B)** Hierarchical with indentation showing directory structure
```
deployments/
  cipher-im/
    compose.yml
    Dockerfile
    config/
      cipher-im-app-common.yml
```

**C)** JSON with metadata for type/status tracking
```json
{
  "deployments/cipher-im/compose.yml": {"type": "compose", "status": "required"},
  "deployments/cipher-im/Dockerfile": {"type": "docker", "status": "required"}
}
```

**D)** Simple list with comment headers grouping by deployment type
```
# SUITE: cryptoutil
deployments/cryptoutil/compose.yml
# PRODUCT-SERVICE: cipher-im
deployments/cipher-im/compose.yml
```

**E)** 

**Answer**: 

**Rationale**: This format choice affects:
- ValidateStructuralMirror parsing complexity
- Human readability for manual audits
- Pre-commit hook performance
- Future extensibility (adding metadata)

---

## Question 2: Mirror Strictness and Infrastructure Handling (Phase 3 Task 3.2, Phase 5)

**Context**: Phase 5 creates "exact mirror" structure (Q2:A from quizme-v1), but edge cases need clarification:
- Infrastructure deployments (shared-postgres, shared-telemetry, shared-citus, compose) have NO application config files
- Template deployment is for creating new services, unclear if needs config/ counterpart
- Current configs/ has 55 files, deployments/ has ~36 config files + compose files

**Question**: How strict should the deployments ↔ configs mirroring be, and how to handle infrastructure/template?

**A)** Exact 1:1 for PRODUCT-SERVICE only, exclude infrastructure/template entirely
- sm-kms, cipher-im, jose-ja, pki-ca, identity-* MUST have exact mirror
- shared-postgres, compose, template EXCLUDED from mirror validation

**B)** Exact 1:1 for all, create placeholder config/ directories for infrastructure
- Create configs/shared-postgres/, configs/shared-telemetry/ with README.md explaining "no app config"
- Create configs/template/ with 4 template config files
- Create configs/compose/ (empty or README)

**C)** Deployments-driven strict, configs can have extras
- Every deployments/ directory MUST have configs/ counterpart
- configs/ CAN have extras (orphaned files) - handled separately in Question 5
- Infrastructure gets placeholder configs/ if they have ANY configurable settings

**D)** Bidirectional loose with explicit exception list
- Most directories MUST mirror
- Maintain exceptions list in code: []string{"shared-postgres", "shared-telemetry", "compose"}
- Template mirrors to configs/template/ (special case)

**E)** 

**Answer**: 

**Rationale**: This decision affects:
- Phase 3 Task 3.2 ValidateStructuralMirror implementation
- Phase 5 Task 5.3 config restructuring scope
- Whether 55 → ~36 migration is expected (orphans exist) or 55 → 55 (create placeholders)

---

## Question 3: Compose File Validation Scope (Phase 4 Task 4.1)

**Context**: Task 4.1 "Implement ValidateComposeFiles" currently in plan as "4h" with vague "schema, ports, health checks" acceptance criteria. Implementation complexity varies widely based on scope.

**Question**: What validations should ValidateComposeFiles perform?

**A)** Minimal: Schema validation only
- `docker compose config --quiet` for schema correctness
- Basic YAML parse to detect syntax errors
- ~2h implementation

**B)** Moderate: Schema + critical runtime issues
- Option A PLUS:
- Port conflict detection (overlapping host ports)
- Health check presence validation (all services MUST have health checks)
- ~4-5h implementation

**C)** Comprehensive: Schema + runtime + security
- Option B PLUS:
- Service dependency validation (depends_on chains correct)
- Secret reference validation (all secrets defined in compose secrets section)
- No hardcoded credentials in environment variables
- Bind mount security (no /run/docker.sock mounts)
- ~8-10h implementation

**D)** Staged approach: Implement A now, B in Phase 4, C deferred to future
- Task 4.1: Schema only (get working, unblock other tasks)
- Task 4.1b (add to tasks.md): Port conflicts + health checks
- FUTURE: Security validations (not in this plan)

**E)** 

**Answer**: 

**Rationale**: Affects implementation timeline (2h vs 10h), test complexity, and LOE accuracy for Phase 4.

---

## Question 4: Config File Validation Scope (Phase 4 Task 4.3)

**Context**: Task 4.3 "Implement ValidateConfigFiles" currently "3h" with "YAML structure, database URLs, bind addresses" but validation rules unspecified.

**Question**: What validations should ValidateConfigFiles perform?

**A)** Minimal: YAML syntax only
- Parse YAML to detect syntax errors
- No semantic validation (bind addresses, URLs could be nonsense)
- ~1-2h implementation

**B)** Moderate: Syntax + format validation
- Option A PLUS:
- Bind address format check (valid IPv4/IPv6, no typos like "127.0.0.l")
- Port number range check (1-65535, no reserved ports <1024 for non-privileged)
- Database URL format (postgres://user:pass@host:port/db structure)
- ~3-4h implementation

**C)** Comprehensive: Syntax + format + cross-reference
- Option B PLUS:
- Cross-reference with compose services (config refers to service names in compose.yml)
- Bind address policy enforcement (admin MUST be 127.0.0.1, public SHOULD be 0.0.0.0 in containers)
- Secret reference validation (database password references secrets, not inline)
- ~6-8h implementation

**D)** Staged approach: Implement A now, B in Phase 4, C deferred
- Task 4.3: YAML syntax only (unblock progress)
- Task 4.3b (add to tasks.md): Format validation
- FUTURE: Cross-reference validation (complex, needs schema definition)

**E)** 

**Answer**: 

**Rationale**: 
- Option C requires schema definition for config files (not documented in ARCHITECTURE.md)
- Affects Phase 4 timeline and whether schema definition is prerequisite work
- Impacts Phase 7 test complexity (mocking invalid configs)

---

## Question 5: Orphaned Config File Handling (Phase 5 Task 5.3)

**Context**: Phase 5 Task 5.3 migrates configs/ files to mirror deployments/ structure. Current state:
- configs/ has 55 files
- deployments/ has ~36 config files (excluding compose.yml, Dockerfile, secrets)
- Potential for configs/ files WITHOUT corresponding deployments/ directories (orphans)

**Question**: How should orphaned config files (exist in configs/ but no deployments/ counterpart) be handled?

**A)** Validation error - Refuse migration, force manual cleanup first
- Pre-migration check: Identify all orphans
- Fail with error message: "Found N orphaned configs, resolve manually before migration"
- User MUST delete orphans or create deployments/ directories before proceeding

**B)** Create placeholder deployments - Auto-generate missing deployment directories
- If configs/mystery-service/ exists but deployments/mystery-service/ doesn't:
- Create deployments/mystery-service/ with minimal compose.yml + README.md
- Log warning: "Created placeholder deployment for mystery-service"

**C)** Orphaned directory - Move to configs/orphaned/ for review
- Create configs/orphaned/ directory
- Move all orphans there during migration
- Create configs/orphaned/README.md documenting what to review
- Continue migration for valid configs

**D)** Best-effort with logging - Migrate valid, skip orphans
- Migrate configs that have deployments/ counterparts
- Skip configs without counterparts (leave in place)
- Log skipped files to test-output/phase5/orphaned-configs.txt
- Manual review after migration

**E)** 

**Answer**: 

**Rationale**: 
- Affects whether Phase 5 can proceed autonomously or requires user intervention
- Determines if pre-migration investigation task (count orphans) needs to be added to Phase 5
- Impacts rollback strategy if migration partially fails

---

## Instructions for Agent

**After user provides answers**:
1. Merge answers into plan.md "Executive Decisions" section
2. Update affected tasks in tasks.md with specific implementation details
3. Adjust LOE estimates based on scope decisions
4. Delete this quizme-v2.md file
5. Commit changes: `git add -A && git commit -m "docs(planning): merge quizme-v2 answers into plan/tasks"`
6. Begin Phase 0.5 execution autonomously (per quizme-v1 Q4:A)

**Continuous Execution Mandate**: Per quizme-v1 Q4:A ("STOP ASKING ME TO CONFIRM!!!"), after merging answers, proceed directly to Phase 0.5 execution WITHOUT asking permission.
