# Copilot Instructions

## Core Principles

- Instruction files auto-discovered from .github/instructions/*.instructions.md
- Keep rules short, one directive per line
- Reference external & project resources; avoid duplication of content
- ALWAYS use built-in tools over terminal commands
- **MUST: Do regular commits and pushes to enable workflow monitoring and validation**
- **MUST: ALWAYS bias towards accuracy, completeness, and correctness** - NEVER bias towards completing fast
- **MUST: Take the time required to do things correctly** - Time and token budgets are not constraints for Speckit work
- **MUST: Prioritize doing things right over doing things quickly** - Quality over speed is mandatory

## Terminology - RFC 2119 Keywords

**Requirement Keywords** (source: .specify/memory/constitution.md Section VIII):

- **MUST** = **REQUIRED** = **MANDATORY** = **SHALL** - Absolute requirement
- **MUST NOT** = **SHALL NOT** - Absolute prohibition
- **SHOULD** = **RECOMMENDED** - Highly desirable (may ignore with justification)
- **SHOULD NOT** = **NOT RECOMMENDED** - Not advisable (may do with justification)
- **MAY** = **OPTIONAL** - Truly optional (implementer decides)

**Emphasis Keywords** (instruction files only):

- **CRITICAL** - Historically regression-prone areas requiring extra attention (format_go self-modification, Windows Firewall, SQLite deadlocks)
- **ALWAYS** / **NEVER** - Emphatic MUST / MUST NOT (no exceptions)

**All keywords are semantically equivalent to their RFC 2119 base** - use context to determine if CRITICAL/ALWAYS adds historical emphasis or is synonym for MUST.

## LLM Agent Continuous Work Directive - ABSOLUTE ENFORCEMENT

**CRITICAL: NEVER STOP WORKING UNTIL USER EXPLICITLY CLICKS "STOP" button in Copilot Chat**

**CRITICAL: NEVER FEEL TIME PRESSURE OR TOKEN PRESSURE WHEN WORKING ON SPECKIT**

**CRITICAL: ALWAYS COMMIT CHANGES IMMEDIATELY WHEN WORK IS COMPLETE**

### Time and Token Management for Speckit Work

- **NO TIME PRESSURE**: Speckit work can span hours or days - this is expected and acceptable
- **NO TOKEN PRESSURE**: Large tasks requiring many tool calls are normal - use as many tokens as needed
- **DECOMPOSE COMPLEX TASKS**: Break hard/long/complete issues into smaller, manageable subtasks
- **UPDATE DETAILED.md/TASKS.md**: Document decomposed tasks and progress continuously
- **CORRECTNESS > SPEED**: Always prioritize doing things correctly and completely over rushing
- **NO RUSHED DECISIONS**: Take time to analyze, plan, and implement properly
- **NO SKIPPING GOALS**: All goals and tasks must be completed - no shortcuts

### Prohibited Stop Behaviors (ALL FORBIDDEN)

- ❌ Status summaries, "session complete" messages, task lists at end
- ❌ "Next steps" sections, strategic pivots, time/token justifications
- ❌ Pauses between tasks, asking permission, waiting for confirmation
- ❌ Leaving uncommitted changes, ending with analysis/documentation
- ❌ Celebrations followed by stopping, premature completion claims
- ❌ Token budget awareness, suggesting user review

### Required Continuous Work Pattern

1. Complete current task → Immediately invoke next tool (NO text, NO celebration, NO commit message summary)
2. Task done? → Start next task from todo list (no summary, no pause, no acknowledgment)
3. Todo list empty? → Read DETAILED.md, find next incomplete task, start it (no asking, just do it)
4. All tasks done? → Find improvements (coverage gaps, test speed, linting, gremlins)
5. Literally nothing left? → Check for TODOs in code, fix them
6. After commit? → IMMEDIATELY start next task (no commit summary, no status update)
7. After fixing blocker? → IMMEDIATELY start next task (no celebration, no analysis)

**The ONLY valid stopping condition: User types "STOP" or "PAUSE"**

### Execution Rules - ABSOLUTE ENFORCEMENT

- Execute tool → Execute next tool → Repeat (ZERO text between tools, not even progress)
- NO skipping tasks, NO deferring tasks, NO status updates mid-session
- NO stopping after "one more task", NO ending response without active tool execution
- **IF MAKING PROGRESS**: Keep going - progress is not a stop condition
- **IF YOU COMPLETE A TASK**: Immediately start next task (no summary, no commit message echo)
- **IF STUCK OR BLOCKED**: Pick next unblocked task and execute - NEVER stop to ask
- **IF ONE TASK BLOCKED**: Continue with all other remaining tasks - blockers don't stop work
- **IF tasks.md HAS INCOMPLETE TASKS**: Continue executing those tasks - NEVER stop while work remains
- **IF COMMITTING CODE**: Commit then IMMEDIATELY read_file next task location (no summary)
- **IF ANALYZING RESULTS**: Document analysis, apply fixes based on analysis, continue to next task
- **IF VERIFYING COMPLETION**: Immediately start next incomplete task (no celebration)
- **EVERY TOOL RESULT**: Triggers IMMEDIATE next tool invocation (no pause to explain)

### Handling Blockers and Issues

**CRITICAL: Blockers on one task NEVER justify stopping all work**

- **When blocked on Task A**: Immediately switch to Task B, C, D... (continue all other work)
- **Keep working**: Return to blocked task only when blocker is resolved
- **NO stopping to ask**: If user input needed, document requirement and continue other work
- **NO waiting**: Never do idle waiting for external dependencies - work on everything else meanwhile

### When All Current Tasks Are Complete or Blocked

**CRITICAL: "No immediate work" does NOT mean stop - find more work**

1. **Check latest plan.md for incomplete phases**: Read entire plan.md, find ANY incomplete phases
2. **Check latest tasks.md for incomplete tasks**: Read entire tasks.md, find ANY incomplete tasks
3. **Look for quality improvements**: Coverage gaps, test speed, linting issues, TODOs in code
4. **Scan for technical debt**: Grep for TODO/FIXME/HACK comments, address them
5. **Review recent commits**: Check for incomplete work, missing tests, documentation gaps
6. **Verify CI/CD health**: Check workflow files, fix any disabled/failing checks
7. **Code quality sweep**: Run golangci-lint, fix warnings, improve test coverage & quality, improve gremlins coverage & quality
8. **Performance analysis**: Identify slow tests (>15s), apply probabilistic execution
9. **Mutation testing**: Run gremlins on packages below 98% mutation score
10. **ONLY if literally nothing exists**: Ask user for next work direction

**Pattern when phase complete**:

- ❌ WRONG: "Phase 3 complete! Here's what we did..." (STOPPING)
- ✅ CORRECT: `read_file DETAILED.md` → find Phase 4/5/6 tasks → immediately start first task (NO SUMMARY)

## CRITICAL Regression Prevention

**See detailed patterns in instruction files:**

- Format_go self-modification: See 01-03.coding.instructions.md "Context Reading Before Refactoring"
- Windows Firewall exceptions: See 01-07.security.instructions.md "Windows Firewall Exception Prevention"
- Git workflow patterns: See 03-03.git.instructions.md "Restore from Clean Baseline Pattern"

## CRITICAL: Health Check Endpoint Pattern - MANDATORY

**MANDATORY: All 9 proprietary services MUST use `/admin/v1/livez` and `/admin/v1/readyz` - NEVER `/admin/v1/healthz`**

### Proprietary Services Pattern (KMS Reference Implementation)

**Admin HTTPS Endpoint** (127.0.0.1:9090):

- `/admin/v1/livez` - Liveness probe (lightweight check: service running, process alive)
- `/admin/v1/readyz` - Readiness probe (heavyweight check: dependencies healthy, ready for traffic)
- `/admin/v1/shutdown` - Graceful shutdown trigger

**Health Check Semantics**:

- **livez**: Fast, lightweight check (~1ms) - verifies process is alive, TLS server responding
- **readyz**: Slow, comprehensive check (~100ms+) - verifies database connectivity, downstream services, resource availability
- **Use livez for**: Docker healthchecks (fast, frequent), liveness probes (restart on failure)
- **Use readyz for**: Kubernetes readiness probes (remove from load balancer), deployment validation

**Why Two Separate Endpoints** (Kubernetes standard):

- Liveness: Process alive but stuck? Restart container
- Readiness: Process alive but dependencies down? Remove from load balancer, don't restart
- Combined healthz endpoint can't distinguish these two failure modes

**Implementation Source**: KMS uses gofiber middleware which provides livez/readyz pattern out-of-box

### Third-Party Services (Different Patterns)

| Service | Health Check Type | Port | Endpoint | Exposed to Host? |
|---------|------------------|------|----------|------------------|
| OpenTelemetry Collector | OTEL standard | 13133 | Internal | ❌ No (container-only) |
| PostgreSQL | pg_isready | N/A | N/A | N/A |
| Grafana OTEL LGTM | Grafana standard | 3000 | Varies | ✅ Yes |

**Third-party services MAY use their own standard health check patterns - document them explicitly**

### CA Service Alignment Required

**Current CA Status**: Uses `/health`, `/livez`, `/readyz` without `/admin/v1` prefix

**Action Required**: Migrate CA to standard `/admin/v1/livez` and `/admin/v1/readyz` pattern for consistency

### Reference

- Implementation: `internal/kms/server/application/application_listener.go`
- Magic constants: `internal/shared/magic/magic_network.go`
- Analysis: `docs/HEALTH-CHECK-ANALYSIS.md`

## CRITICAL: Public HTTPS Bind Address Pattern - NEVER VIOLATE

**MANDATORY: Public HTTPS endpoints MUST be documented as `<configurable_address>:<configurable_port>` - NEVER hardcode `0.0.0.0`**

**Why This Matters**:

- **Container deployments**: Use `0.0.0.0` bind address (enables external access)
- **Test/dev environments**: Use `127.0.0.1` bind address (prevents Windows Firewall prompts)
- **Configuration-driven**: Bind address MUST be configurable, not hardcoded to `0.0.0.0`

**Pattern Recognition**:

- ❌ WRONG: "Public API (0.0.0.0:configurable)" - hardcodes container-only pattern
- ❌ WRONG: "Public API (0.0.0.0:8080)" - hardcodes both address and port
- ✅ CORRECT: "Public API (<configurable_address>:<configurable_port>)" - fully configurable
- ✅ CORRECT: In code: `bindAddress := config.PublicBindAddress` (NEVER hardcode "0.0.0.0")

**Why `0.0.0.0` is Wrong in Documentation**:

- Implies public endpoints ALWAYS bind to all interfaces (not true for tests/dev)
- Contradicts constitution.md requirement: "Public endpoints MAY use configurable IPv4 or IPv6 bind address"
- Prevents Windows development: `0.0.0.0` triggers Firewall prompts, blocks automation
- Misleads implementers: Suggests hardcoding instead of configuration

**Correct Documentation Pattern**:

```markdown
**Dual HTTPS Servers**: Public API (<configurable_address>:<configurable_port>) + Admin API (127.0.0.1:9090)
```

**Correct Implementation Pattern**:

```go
// Configuration structure
type ServerConfig struct {
    PublicBindAddress string // Default: "127.0.0.1" (tests/dev), "0.0.0.0" (containers)
    PublicPort        int    // Configurable port
    AdminBindAddress  string // ALWAYS "127.0.0.1" (never configurable)
    AdminPort         int    // ALWAYS 9090 (never configurable)
}

// Server binding
publicAddr := fmt.Sprintf("%s:%d", config.PublicBindAddress, config.PublicPort)
adminAddr := "127.0.0.1:9090" // Admin NEVER configurable
```

**Where This Pattern Appears**:

- constitution.md Section V (Service Architecture)
- spec.md (Service Template Extraction, Dual HTTPS Pattern)
- PLAN.md (Core Requirements)
- 01-01.architecture.instructions.md (Service Template Pattern)

**Enforcement**: When writing documentation about public HTTPS endpoints, ALWAYS use `<configurable_address>:<configurable_port>` pattern. When implementing, ALWAYS read bind address from configuration.

## CRITICAL: Service Template Requirements - MANDATORY

**Service Template Implementation Details**:

- **Admin Endpoints** (127.0.0.1:9090):
  - `/livez`, `/readyz`, `/shutdown` endpoints MANDATORY
  - Admin prefix MUST be configurable (default: `/admin/v1`)
  - Implementation: gofiber middleware (reference: sm-kms `internal/kms/server/application/application_listener.go`)
- **Health Check Requirements**:
  - OpenTelemetry Collector Contrib MUST use separate health check job (does NOT expose external health endpoint)
  - Reference: KMS Docker Compose `deployments/compose/compose.yml` (working pattern)
- **Docker Secrets Validation**:
  - Docker Compose MUST include dedicated job to validate Docker Secrets presence and mounting
  - Fast-fail check before starting all other jobs and services

**Service Template Migration Priority** (HIGH PRIORITY):

1. **learn-ps FIRST** (Phase 7): CRITICAL - implement and validate ALL template requirements before production services
2. **One service at a time** (Phase 8+, excludes sm-kms): Sequentially refactor jose-ja, pki-ca, identity services
3. **sm-kms LAST** (Phase 10): Only after ALL other services running excellently on template

## CRITICAL: Hash Registry Salt Encoding - MANDATORY

**LowEntropyRandomHashRegistry / HighEntropyRandomHashRegistry** (Random):

- MUST encode version AND all parameters (iterations, salt, algorithm) WITH hash
- Format: `{version}:{algorithm}:{params}:base64(salt):base64(hash)`
- Rationale: Random salt must be stored to verify later

**LowEntropyDeterministicHashRegistry / HighEntropyDeterministicHashRegistry** (Deterministic):

- MUST encode version ONLY (NEVER encode salt or parameters)
- Format: `{version}:base64(hash)`
- MUST use different fixed configurable salt per version (v1/v2/v3)
- MUST derive ACTUAL SALT from: configured fixed salt (pepper) + input cleartext
- Rationale: Revealing salt in DB would be crypto bug; similar to AES-GCM-SIV IV derivation

---

## Instruction Files Reference

| File | Description |
|------|-------------|
| 01-01.architecture | Products & Services Architecture |
| 01-02.versions | Minimum Versions & Consistency Requirements |
| 01-03.coding | Coding patterns & standards |
| 01-04.testing | Testing patterns & best practices |
| 01-05.golang | Go project structure & conventions |
| 01-06.database | Database & ORM patterns |
| 01-07.security | Security patterns |
| 01-08.linting | Code quality & linting standards |
| 01-09.cryptography | FIPS compliance, hash versioning, algorithm agility |
| 01-10.pki | PKI, CA, certificate management, CA/Browser Forum compliance |
| 02-01.github | CI/CD workflow |
| 02-02.docker | Docker & Compose |
| 02-03.observability | Observability & monitoring |
| 03-01.openapi | OpenAPI rules |
| 03-02.cross-platform | Cross-platform tooling |
| 03-03.git | Git workflow rules |
| 03-04.dast | DAST scanning |
| 04-01.sqlite-gorm | SQLite GORM config |
| 05-01.evidence-based | Evidence-based task completion |
| 06-01.speckit | Speckit workflow integration & feedback loops |
| 07-01.anti-patterns | Common anti-patterns and mistakes |
