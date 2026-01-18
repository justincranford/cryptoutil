# Copilot Configuration Enhancement - QUIZME v1

**Created**: 2026-01-18
**Plan Reference**: [copilot-enhancement-plan.md](copilot-enhancement-plan.md)
**Tasks Reference**: [copilot-enhancement-tasks.md](copilot-enhancement-tasks.md)
**Status**: Awaiting Answers

---

## Instructions for Answering

1. **Read full context** before answering (plan.md + tasks.md)
2. **Answer with letter** (A/B/C/D/E)
3. **Provide write-in** for E options
4. **Consider implications** of each choice
5. **Cross-reference** with existing cryptoutil patterns

---

## Section 1: Collections Organization Strategy

### Q1.1: Primary Organization Principle

**Context**: Collections can group prompts/instructions by workflow, technology, role, or hybrid approaches. Current cryptoutil structure has 28 instruction files, 11 prompts, 9 agents covering diverse domains.

**Question**: How should we primarily organize collections for cryptoutil?

**Options**:

**A)** By workflow stage (planning → coding → testing → review → deploy)
- **Pros**: Follows natural development lifecycle, easy for new developers
- **Cons**: Cross-cutting concerns (security, performance) span multiple stages
- **Example Collections**: 
  - planning-workflow (speckit, plan-tasks-quizme)
  - coding-workflow (code-review, refactor, test-generate)
  - deployment-workflow (docker, github-actions, compose)

**B)** By technology domain (Go, Docker, PostgreSQL, Cryptography)
- **Pros**: Good for technology specialists, clear boundaries
- **Cons**: Workflow context scattered across collections
- **Example Collections**:
  - go-development (coding, testing, linting Go)
  - cryptography (crypto, pki, hashes, authn)
  - infrastructure (docker, github, observability)

**C)** By service type (API, Auth, Database, Testing)
- **Pros**: Aligns with microservices architecture, service-oriented teams
- **Cons**: Overlaps with technology domains, less intuitive
- **Example Collections**:
  - api-development (openapi, server, client)
  - auth-security (authn, authz, security, crypto)
  - data-persistence (database, sqlite, migrations)

**D)** Hybrid approach (primary by workflow, secondary by domain)
- **Pros**: Most flexible, supports multiple mental models
- **Cons**: More complex, requires clear naming conventions
- **Example Collections**:
  - workflow.development (plan → code → test → review)
  - workflow.deployment (build → docker → compose → k8s)
  - domain.cryptography (all crypto-related regardless of workflow)
  - domain.database (all database-related regardless of workflow)

**E)** Write-in: ________________

**Answer**: ________________

**Follow-up Questions**:
- Should workflow-based collections include domain-specific prompts/instructions?
- How do we handle cross-cutting concerns (security, performance, logging)?
- Should we create meta-collections that combine other collections?

---

### Q1.2: Collection Granularity

**Context**: Collections can be broad (10+ items each) or narrow (3-5 items each). VS Code allows nested collections but no official guidance on optimal size.

**Question**: What granularity should collections target?

**Options**:

**A)** Broad collections (10-15 items each, fewer collections total)
- **Pros**: Fewer files to maintain, comprehensive coverage per collection
- **Cons**: May include unrelated items, harder to find specific resources
- **Example**: 
  - `backend-development.collection.yml` (15 items: Go, database, testing, API)

**B)** Narrow collections (3-5 items each, more collections total)
- **Pros**: Highly focused, easier discovery, better recommendations
- **Cons**: More files to maintain, may have gaps
- **Example**: 
  - `go-coding.collection.yml` (3 items: coding, golang, linting)
  - `go-testing.collection.yml` (4 items: testing, coverage, mutation, benchmarks)

**C)** Mixed granularity (broad for workflows, narrow for domains)
- **Pros**: Flexible approach matching different use cases
- **Cons**: Inconsistent organization, learning curve
- **Example**:
  - `development-workflow.collection.yml` (12 items - broad)
  - `cryptography.collection.yml` (4 items - narrow)

**D)** Dynamic based on usage (start broad, split when collections exceed threshold)
- **Pros**: Evolves with actual usage patterns
- **Cons**: Requires ongoing maintenance, may break muscle memory
- **Threshold**: Split when collection exceeds 10 items or low usage correlation

**E)** Write-in: ________________

**Answer**: ________________

**Follow-up Questions**:
- Should we track collection usage to optimize granularity?
- What's the maximum items before collection becomes unwieldy?
- Should related collections reference each other?

---

### Q1.3: Collection Naming Convention

**Context**: Collection names appear in VS Code UI and affect discoverability. Need balance between brevity and clarity.

**Question**: What naming convention should we use for collection files?

**Options**:

**A)** Descriptive kebab-case (e.g., `development-workflow.collection.yml`)
- **Pros**: Self-documenting, easy to understand at glance
- **Cons**: Can be verbose, may not sort logically
- **Examples**:
  - `backend-development-complete.collection.yml`
  - `cryptography-and-security.collection.yml`

**B)** Prefixed with category (e.g., `workflow.development.collection.yml`)
- **Pros**: Groups related collections, clear hierarchy
- **Cons**: May be redundant with file system organization
- **Examples**:
  - `workflow.planning.collection.yml`
  - `workflow.coding.collection.yml`
  - `domain.cryptography.collection.yml`
  - `domain.database.collection.yml`

**C)** Numbered by priority (e.g., `01-essential.collection.yml`)
- **Pros**: Explicit ordering, clear what beginners should use
- **Cons**: Subjective priority, may need renumbering
- **Examples**:
  - `01-getting-started.collection.yml`
  - `02-development-workflow.collection.yml`
  - `10-advanced-cryptography.collection.yml`

**D)** Short acronyms (e.g., `dev-wf.collection.yml`)
- **Pros**: Compact, fast to type
- **Cons**: Requires memorization, less discoverable
- **Examples**:
  - `dev-wf.collection.yml` (development workflow)
  - `crypto.collection.yml` (cryptography)
  - `qg.collection.yml` (quality gates)

**E)** Write-in: ________________

**Answer**: ________________

**Follow-up Questions**:
- Should collection names match instruction file naming patterns?
- Should we use singular or plural (e.g., `workflow` vs `workflows`)?
- How do we handle versioning of collections?

---

## Section 2: MCP Server Integration Priority

### Q2.1: MCP Server Development Timeline

**Context**: MCP (Model Context Protocol) servers enable advanced tool integration. Awesome Copilot has MCP server, but custom servers require significant development. cryptoutil has complex operations (KMS, PKI, testing, deployment).

**Question**: When should we prioritize MCP server development?

**Options**:

**A)** Immediate (start in Phase 1 alongside prompts/collections)
- **Pros**: Early adopter advantage, cutting-edge capabilities
- **Cons**: Unproven ecosystem, high risk, resource intensive
- **Rationale**: If cryptoutil operations are truly unique and MCP provides 10× value
- **Risk**: Ecosystem not mature, specification may change

**B)** Phase 2 (after prompts/collections complete, before advanced features)
- **Pros**: Balanced approach, proven patterns first, innovation second
- **Cons**: May miss early feedback opportunities
- **Rationale**: Establish baseline with prompts, then enhance with MCP
- **Dependencies**: Complete Tasks 1.1-2.4 first

**C)** Phase 3 (research now, implement later after ecosystem matures)
- **Pros**: Safe approach, learn from others' mistakes, stable ecosystem
- **Cons**: May lag behind competitors, miss early advantages
- **Rationale**: Current prompt/agent system sufficient for 80% of needs
- **Timeline**: 6-12 months research, 3-6 months implementation

**D)** Opportunistic (prototype when specific use case justifies it)
- **Pros**: Resource efficient, need-driven development
- **Cons**: Reactive rather than proactive, piecemeal integration
- **Rationale**: Build MCP server only when existing tools insufficient
- **Trigger**: Identify 3+ use cases that prompts/agents can't handle

**E)** Write-in: ________________

**Answer**: ________________

**Follow-up Questions**:
- What specific cryptoutil operations would benefit most from MCP?
- Are there existing MCP servers we could leverage instead of building custom?
- What's the ROI threshold to justify custom MCP server development?

---

### Q2.2: MCP Server Scope

**Context**: If we build custom MCP server, scope could range from narrow (single operation) to broad (entire cryptoutil API surface). Trade-off between focus and coverage.

**Question**: What scope should custom MCP server target?

**Options**:

**A)** Narrow - Single high-value operation (e.g., KMS only)
- **Pros**: Focused development, faster to market, easier to maintain
- **Cons**: Limited utility, may need multiple servers later
- **Example**: MCP server for cryptoutil-kms operations only
- **LOE**: 20-30 hours
- **Value**: Enables AI agents to manage encryption keys directly

**B)** Medium - Related operations (e.g., all cryptography: KMS + PKI + JOSE)
- **Pros**: Cohesive functionality, reasonable scope
- **Cons**: More complex, longer development
- **Example**: MCP server covering crypto operations across services
- **LOE**: 40-60 hours
- **Value**: Comprehensive crypto operations from AI agents

**C)** Broad - All cryptoutil operations (KMS, PKI, JOSE, Identity)
- **Pros**: Complete coverage, one-stop solution
- **Cons**: Very complex, long development, high maintenance
- **Example**: MCP server exposing entire cryptoutil API surface
- **LOE**: 80-120 hours
- **Value**: Full automation capabilities

**D)** Incremental - Start narrow, expand based on usage
- **Pros**: Validate value early, expand based on data
- **Cons**: May have inconsistent design across iterations
- **Example**: 
  - Phase 1: KMS operations (20h)
  - Phase 2: Add PKI if KMS MCP usage >50/week (30h)
  - Phase 3: Add JOSE if PKI MCP usage >50/week (20h)
- **Value**: Risk-managed expansion

**E)** Write-in: ________________

**Answer**: ________________

**Follow-up Questions**:
- Which cryptoutil operations are most frequently used in AI interactions?
- Are there security concerns with exposing certain operations via MCP?
- Should MCP server operations require additional authentication?

---

### Q2.3: MCP Server Technology Stack

**Context**: MCP servers can be implemented in various languages. cryptoutil is Go-based, but MCP ecosystem has examples in Python, TypeScript, Rust.

**Question**: What language/framework should we use for custom MCP server?

**Options**:

**A)** Go (match cryptoutil codebase)
- **Pros**: Team expertise, code reuse, consistent tooling
- **Cons**: Less MCP ecosystem support (most examples in Python/TS)
- **Dependencies**: Need Go MCP SDK or build from spec
- **Integration**: Can import cryptoutil packages directly

**B)** Python (most MCP examples and tooling)
- **Pros**: Rich MCP ecosystem, examples, community support
- **Cons**: Different language from cryptoutil, runtime dependencies
- **Dependencies**: MCP Python SDK, Python runtime
- **Integration**: Call cryptoutil APIs via HTTP/gRPC

**C)** TypeScript (VS Code native, good MCP support)
- **Pros**: Native VS Code integration, Node.js ecosystem
- **Cons**: Different language, JavaScript runtime needed
- **Dependencies**: MCP TypeScript SDK, Node.js runtime
- **Integration**: Call cryptoutil APIs via HTTP/gRPC

**D)** Polyglot (Go for core logic, Python/TS for MCP wrapper)
- **Pros**: Best of both worlds, leverage existing code
- **Cons**: More complex, two codebases to maintain
- **Architecture**: 
  - Go: Core cryptoutil operations library
  - Python/TS: MCP protocol wrapper calling Go library
- **Integration**: Expose Go library via C FFI or HTTP API

**E)** Write-in: ________________

**Answer**: ________________

**Follow-up Questions**:
- What's the maturity of MCP SDKs in each language?
- Are there performance implications for MCP protocol overhead?
- Should we contribute to Go MCP SDK if it doesn't exist?

---

## Section 3: Agent Specialization Depth

### Q3.1: Agent Specialization Philosophy

**Context**: Current cryptoutil has 9 SpecKit agents (workflow-focused). Could add domain experts (security, crypto, testing). Trade-off between specialization depth and maintenance overhead.

**Question**: How specialized should domain expert agents be?

**Options**:

**A)** Highly specialized (10+ domain experts covering narrow scopes)
- **Pros**: Deep expertise, precise recommendations, clear ownership
- **Cons**: High maintenance, more files, potential overlap
- **Examples**:
  - expert.cryptography.rsa
  - expert.cryptography.ecdsa
  - expert.cryptography.aes
  - expert.security.tls
  - expert.security.oauth
  - expert.testing.unit
  - expert.testing.integration
  - expert.testing.e2e
  - expert.database.postgres
  - expert.database.sqlite
- **Total Agents**: 19 (9 SpecKit + 10 experts)

**B)** Moderately specialized (5-6 broad domain experts)
- **Pros**: Balanced expertise vs. maintenance, covers major domains
- **Cons**: Less precise, agents may need sub-delegation
- **Examples**:
  - expert.security (all security topics)
  - expert.cryptography (all crypto topics)
  - expert.testing (all test types)
  - expert.performance (profiling, optimization)
  - expert.database (all database topics)
- **Total Agents**: 14 (9 SpecKit + 5 experts)

**C)** Minimally specialized (2-3 mega-agents covering broad areas)
- **Pros**: Simple, easy to maintain, clear boundaries
- **Cons**: Less targeted, may lack depth
- **Examples**:
  - expert.development (coding, testing, refactoring)
  - expert.operations (deployment, monitoring, security)
  - expert.quality (testing, linting, performance)
- **Total Agents**: 12 (9 SpecKit + 3 experts)

**D)** Dynamic specialization (agents spawn sub-agents as needed)
- **Pros**: Ultimate flexibility, context-aware expertise
- **Cons**: Complex orchestration, requires sophisticated handoff system
- **Examples**:
  - expert.coordinator (routes to specialized sub-agents)
    - Spawns expert.cryptography.rsa when RSA mentioned
    - Spawns expert.testing.property when property tests needed
- **Total Agents**: 10 (9 SpecKit + 1 coordinator)

**E)** Write-in: ________________

**Answer**: ________________

**Follow-up Questions**:
- How do we measure agent effectiveness to optimize specialization?
- Should agents declare expertise levels (beginner, intermediate, expert)?
- What's the overhead of agent handoffs vs. single comprehensive agent?

---

### Q3.2: Agent Handoff Strategy

**Context**: Agents can hand off to other agents for specialized tasks. Current SpecKit agents have defined handoff patterns. Need to decide handoff philosophy for domain experts.

**Question**: How should agents handle tasks outside their expertise?

**Options**:

**A)** Mandatory handoffs (agents MUST delegate outside expertise)
- **Pros**: Clear boundaries, enforces specialization, tracks expertise usage
- **Cons**: May require multiple handoffs, slower workflow
- **Example**: 
  - User: "Review this crypto code for security"
  - expert.security → expert.cryptography (crypto review) → expert.security (security review)
- **Rationale**: Ensures best expert handles each aspect

**B)** Optional handoffs (agents CAN delegate but try to handle first)
- **Pros**: Flexible, faster for simple tasks, fewer handoffs
- **Cons**: May provide suboptimal advice outside expertise
- **Example**:
  - User: "Review this crypto code for security"
  - expert.security handles both (security + basic crypto review)
  - Only hands off if crypto complexity exceeds threshold
- **Rationale**: Balance efficiency with expertise

**C)** No handoffs (each agent handles full scope, uses references)
- **Pros**: Simple, no handoff overhead, complete context
- **Cons**: May lack depth, agents need broader knowledge
- **Example**:
  - User: "Review this crypto code for security"
  - expert.security reads crypto instructions, provides integrated review
  - No handoffs needed
- **Rationale**: Instruction files provide sufficient expertise

**D)** Smart handoffs (agents assess complexity, delegate intelligently)
- **Pros**: Optimal efficiency, depth when needed
- **Cons**: Requires sophisticated agent logic
- **Example**:
  - User: "Review this crypto code for security"
  - expert.security analyzes complexity
  - If crypto simple: Handle internally using crypto instructions
  - If crypto complex: Hand off to expert.cryptography
- **Rationale**: Complexity-based delegation

**E)** Write-in: ________________

**Answer**: ________________

**Follow-up Questions**:
- Should handoffs be visible to users or transparent?
- Can agents hand off back to original agent after sub-task?
- How do we prevent infinite handoff loops?

---

### Q3.3: Agent Tool Access Patterns

**Context**: Agents can have different tool access levels. Some tools (semantic_search, grep_search) are broadly useful. Others (run_in_terminal) more sensitive.

**Question**: How should we manage tool access for domain expert agents?

**Options**:

**A)** Unrestricted (all agents can use all available tools)
- **Pros**: Maximum flexibility, agents self-sufficient
- **Cons**: Security risk, potential misuse, harder to audit
- **Tools**: semantic_search, grep_search, read_file, create_file, run_in_terminal, etc.
- **Example**: expert.security can run terminal commands to test vulnerabilities

**B)** Role-based (agents get tools matching their domain)
- **Pros**: Security by least privilege, clear boundaries
- **Cons**: May limit creativity, need tool categorization
- **Example**:
  - expert.security: semantic_search, grep_search, read_file (read-only tools)
  - expert.testing: semantic_search, read_file, runTests (testing tools)
  - expert.development: all tools (full access)

**C)** Tiered access (agents request elevated tools when needed)
- **Pros**: Secure by default, flexible when justified
- **Cons**: May interrupt workflow, requires approval mechanism
- **Example**:
  - Default: semantic_search, grep_search, read_file
  - Elevated (request required): create_file, replace_string_in_file, run_in_terminal
  - Agent requests: "Need run_in_terminal to verify security configuration"

**D)** Capability-based (agents declare required tools in frontmatter)
- **Pros**: Explicit requirements, easier to audit, clear capabilities
- **Cons**: Static allocation, may need updates
- **Example**:
```markdown
---
tools:
  - semantic_search  # REQUIRED: Find security patterns
  - grep_search      # REQUIRED: Search for vulnerabilities
  - read_file        # REQUIRED: Read code
  - create_file      # OPTIONAL: Generate security reports
---
```

**E)** Write-in: ________________

**Answer**: ________________

**Follow-up Questions**:
- Should tool usage be logged for security auditing?
- Can agents compose tools (e.g., semantic_search → read_file → analyze)?
- What tools should NEVER be available to agents (if any)?

---

## Section 4: Instruction File Size Management

### Q4.1: Handling Instruction Files Over 500 Lines

**Context**: cryptoutil coding standards mandate 500-line hard limit. Some instruction files may exceed this (02-01.architecture, 03-02.testing). Need strategy for compliance.

**Question**: How should we handle instruction files that exceed 500 lines?

**Options**:

**A)** Split into topic-specific files with `applyTo` patterns
- **Pros**: Compliant with standards, better targeting, smaller context
- **Cons**: More files, may break logical groupings
- **Example**:
  - `03-02.testing.instructions.md` (600 lines) →
    - `03-02.testing-unit.instructions.md` (applyTo: "**/*_test.go")
    - `03-02.testing-integration.instructions.md` (applyTo: "**/*_integration_test.go")
    - `03-02.testing-e2e.instructions.md` (applyTo: "**/test/e2e/**")
    - `03-02.testing-mutation.instructions.md` (applyTo: "**")

**B)** Keep as-is but add quick reference sections at top
- **Pros**: Maintains logical grouping, comprehensive reference
- **Cons**: Violates 500-line limit, large context size
- **Example**:
  - Add "Quick Reference" section at top (100 lines)
  - Keep full content below (500+ lines)
  - Total: 600+ lines but easier navigation

**C)** Extract examples/details to separate reference docs
- **Pros**: Compliant with limits, examples available when needed
- **Cons**: Information scattered, need cross-references
- **Example**:
  - `03-02.testing.instructions.md` (300 lines - core rules only)
  - `docs/testing/examples.md` (200 lines - examples)
  - `docs/testing/patterns.md` (100 lines - advanced patterns)

**D)** Use external links to detailed documentation
- **Pros**: Smallest instruction files, comprehensive docs elsewhere
- **Cons**: External dependency, may break links, context switching
- **Example**:
  - `03-02.testing.instructions.md` (200 lines - quick reference)
  - Links to: https://cryptoutil.readthedocs.io/testing (full docs)

**E)** Write-in: ________________

**Answer**: ________________

**Follow-up Questions**:
- Should we audit all instruction files for size compliance?
- What's acceptable overhead for quick reference sections (50, 100, 200 lines)?
- Should extracted content still be in `.github/` or moved to `docs/`?

---

### Q4.2: Quick Reference Format

**Context**: If we add quick reference sections to large instruction files, need consistent format. Should balance completeness with brevity.

**Question**: What format should quick reference sections use?

**Options**:

**A)** Table-based (compact, scannable, structured)
- **Pros**: Dense information, easy to scan, good for comparisons
- **Cons**: Limited detail, may not fit all content types
- **Example**:
```markdown
## Quick Reference: Test Coverage Targets

| Category | Packages | Minimum Coverage |
|----------|----------|------------------|
| Production | internal/{jose,identity,kms} | 95% |
| Infrastructure | internal/cmd/cicd/* | 98% |
| Utility | pkg/* | 98% |
```

**B)** List-based (flexible, detailed, hierarchical)
- **Pros**: More detail than tables, good for procedures
- **Cons**: Longer, less scannable
- **Example**:
```markdown
## Quick Reference: Test Execution Commands

- **Unit tests**: `go test ./... -cover -shuffle=on`
  - Concurrent execution (NEVER use -p=1)
  - Randomize order to catch dependencies
- **Race detection**: `go test -race -count=2 ./...`
  - Requires CGO_ENABLED=1
  - Run multiple iterations for probabilistic detection
```

**C)** Code snippet focused (show, don't tell)
- **Pros**: Immediately actionable, copy-paste ready
- **Cons**: Less explanation, may lack context
- **Example**:
```markdown
## Quick Reference: Database Compatibility

```go
// ✅ CORRECT: UUID cross-DB compatible
ID googleUuid.UUID `gorm:"type:text;primaryKey"`

// ❌ WRONG: Breaks SQLite
ID googleUuid.UUID `gorm:"type:uuid;primaryKey"`
```
```

**D)** Mixed format (tables for data, lists for procedures, code for examples)
- **Pros**: Most flexible, best tool for each content type
- **Cons**: May be inconsistent across files
- **Example**: Use format matching content (tables for comparisons, code for patterns, lists for workflows)

**E)** Write-in: ________________

**Answer**: ________________

**Follow-up Questions**:
- Should quick reference duplicate content from main body or summarize?
- How do we keep quick reference in sync when main content changes?
- Should quick reference link to full content sections?

---

### Q4.3: Instruction File Versioning

**Context**: Instruction files evolve over time. Need strategy for tracking changes, handling breaking changes, and maintaining backward compatibility.

**Question**: Should we version instruction files?

**Options**:

**A)** No versioning (single canonical version, updated in place)
- **Pros**: Simple, always current, no legacy versions
- **Cons**: Breaking changes may disrupt workflows, no rollback
- **Example**: `03-02.testing.instructions.md` updated directly when standards change

**B)** Git-based versioning (use git history, no file versions)
- **Pros**: Full change history, can rollback via git
- **Cons**: No explicit version markers, harder to reference specific version
- **Example**: Reference specific commit when needed (`03-02.testing.instructions.md` at commit abc123)

**C)** Semantic versioning in frontmatter (major.minor.patch)
- **Pros**: Explicit versions, breaking changes clear
- **Cons**: Maintenance overhead, need version bump process
- **Example**:
```markdown
---
description: "Testing instructions"
version: 2.1.0
applyTo: "**/*_test.go"
---
```

**D)** Dated versions with deprecation warnings
- **Pros**: Clear timeline, deprecation path
- **Cons**: Multiple versions in repo, migration needed
- **Example**:
  - `03-02.testing.instructions.md` (current)
  - `03-02.testing.2024-12.instructions.md` (deprecated, warnings added)

**E)** Write-in: ________________

**Answer**: ________________

**Follow-up Questions**:
- How do we communicate breaking changes in instructions?
- Should agents detect instruction version mismatches?
- Do we need migration guides when instructions change significantly?

---

## Section 5: Prompt Recommendation Strategy

### Q5.1: Default Prompt Recommendations

**Context**: VS Code `chat.promptFilesRecommendations` setting controls which prompts appear first in chat. Need balance between discoverability and noise.

**Question**: Which prompts should be recommended by default in chat?

**Options**:

**A)** SpecKit workflow only (methodology-focused)
- **Pros**: Emphasizes structured development, consistent with current practice
- **Cons**: Limits discoverability of utility prompts
- **Recommended**:
  - plan-tasks-quizme
  - speckit.plan
  - speckit.implement
  - speckit.clarify
- **Use Case**: When starting new feature work

**B)** Most frequently used (data-driven recommendations)
- **Pros**: Reflects actual usage, optimizes for efficiency
- **Cons**: Requires usage tracking, may vary by developer
- **Recommended** (example based on hypothetical usage):
  - code-review (used daily)
  - test-generate (used multiple times per day)
  - fix-bug (used as needed)
  - refactor-extract (used weekly)
- **Use Case**: General development workflow

**C)** Context-aware based on open files (smart recommendations)
- **Pros**: Most relevant to current work, reduces cognitive load
- **Cons**: Complex to implement, may require VS Code extension
- **Recommended** (examples):
  - Open `*.go` → Suggest: code-review, refactor-extract, test-generate
  - Open `*_test.go` → Suggest: test-generate, improve-coverage
  - Open `Dockerfile` → Suggest: docker-optimize, security-audit
  - Open `openapi*.yaml` → Suggest: openapi-client, generate-docs
- **Use Case**: Context-specific development

**D)** User-configurable favorites (personal preferences)
- **Pros**: Personalized experience, supports different work styles
- **Cons**: Inconsistent across team, no default guidance
- **Recommended**: Each developer configures own list in user settings
- **Use Case**: Power users with established workflows

**E)** Write-in: ________________

**Answer**: ________________

**Follow-up Questions**:
- Should recommendations change based on time of day or project phase?
- Can we A/B test different recommendation strategies?
- Should we limit recommendations to 5-7 to avoid overwhelming users?

---

### Q5.2: Prompt Discovery Mechanism

**Context**: Beyond default recommendations, need mechanism for users to discover available prompts. VS Code doesn't have built-in prompt catalog.

**Question**: How should users discover available prompts?

**Options**:

**A)** Documentation-based (maintain docs/copilot/prompt-catalog.md)
- **Pros**: Comprehensive reference, searchable, always available
- **Cons**: May get out of sync, requires manual updates
- **Example**:
```markdown
# Prompt Catalog

## Development Prompts
- **code-review**: Review code for quality, security, performance
- **refactor-extract**: Extract method/class/package
...
```

**B)** Collection-based (organize by collections, browse collections)
- **Pros**: Logical grouping, aligns with VS Code collections feature
- **Cons**: Requires understanding collection structure
- **Example**: User browses collections → selects "development-workflow" → sees all related prompts

**C)** Interactive prompt browser (custom VS Code webview/command)
- **Pros**: Best UX, searchable, filterable, sortable
- **Cons**: Requires VS Code extension development
- **Example**: Command palette → "Browse Copilot Prompts" → interactive search/filter UI

**D)** Auto-completion in chat (type `/` to see all prompts)
- **Pros**: Built-in VS Code feature, zero setup
- **Cons**: May be overwhelming if many prompts, no filtering
- **Example**: User types `/` in chat → sees alphabetical list of all prompts

**E)** Write-in: ________________

**Answer**: ________________

**Follow-up Questions**:
- Should we tag prompts with metadata (difficulty, LOE, dependencies)?
- Can we auto-generate prompt catalog from frontmatter?
- Should prompt descriptions include examples in frontmatter?

---

### Q5.3: Prompt Deprecation Strategy

**Context**: As prompts evolve, some may become obsolete or superseded. Need strategy for deprecating/removing prompts without breaking workflows.

**Question**: How should we handle deprecated prompts?

**Options**:

**A)** Immediate removal (delete deprecated prompts)
- **Pros**: Clean codebase, no legacy maintenance
- **Cons**: Breaks existing workflows, no migration path
- **Process**: Announce deprecation → remove after 1 sprint

**B)** Deprecation warnings (keep prompt but add warnings)
- **Pros**: Gives time to migrate, clear communication
- **Cons**: Clutters codebase with deprecated prompts
- **Example**:
```markdown
---
description: "[DEPRECATED] Use code-review.prompt.md instead"
deprecated: true
replacement: code-review
---

# Old Prompt (DEPRECATED)

**⚠️ WARNING**: This prompt is deprecated. Use `code-review` instead.
...
```

**C)** Version-based retention (keep N previous versions)
- **Pros**: Backward compatibility, rollback capability
- **Cons**: Multiple versions to maintain
- **Example**: Keep current + 2 previous versions (N=2)

**D)** Archive directory (move to `.github/prompts/archive/`)
- **Pros**: Available if needed, out of main discovery
- **Cons**: May accumulate cruft over time
- **Example**: `.github/prompts/archive/old-code-review.prompt.md`

**E)** Write-in: ________________

**Answer**: ________________

**Follow-up Questions**:
- Should we track prompt usage to identify candidates for deprecation?
- What's the minimum deprecation notice period (1 week, 1 month, 1 quarter)?
- Should deprecated prompts redirect to replacements automatically?

---

## Section 6: Implementation Priorities

### Q6.1: Phase 1 Task Ordering

**Context**: Phase 1 has 4 tasks (1.1-1.4) totaling 10 hours. Some tasks have dependencies. Need to optimize execution order.

**Question**: What order should Phase 1 tasks be executed?

**Options**:

**A)** Sequential dependency order (1.1 → 1.2 → 1.3 → 1.4)
- **Pros**: Respects dependencies, builds on previous work
- **Cons**: No parallelization, slower overall
- **Timeline**: Day 1 (1.1 4h), Day 2 (1.2 2h, 1.3 3h), Day 3 (1.4 1h)

**B)** Parallel where possible (1.1 + 1.3 parallel, then 1.2 → 1.4)
- **Pros**: Faster completion, maximizes parallelization
- **Cons**: Requires coordinating parallel work
- **Timeline**: Day 1 (1.1 4h + 1.3 3h parallel), Day 2 (1.2 2h, 1.4 1h)

**C)** Quick wins first (1.4 → 1.2 → 1.3 → 1.1)
- **Pros**: Early visible results, momentum building
- **Cons**: Violates dependencies, may need rework
- **Timeline**: Day 1 (1.4 1h, 1.2 2h, 1.3 3h), Day 2 (1.1 4h)

**D)** High value first (1.1 → 1.4 → 1.3 → 1.2)
- **Pros**: Delivers highest impact early
- **Cons**: May not respect dependencies
- **Timeline**: Day 1 (1.1 4h), Day 2 (1.4 1h, 1.3 3h), Day 3 (1.2 2h)

**E)** Write-in: ________________

**Answer**: ________________

**Follow-up Questions**:
- Can we validate task dependencies (are they hard or soft constraints)?
- Should we timebox tasks to maintain momentum?
- What's the incremental commit strategy (per task vs. per sub-task)?

---

### Q6.2: Success Metrics Definition

**Context**: Plan defines qualitative and quantitative success metrics. Need to prioritize which metrics to track initially and define measurement approach.

**Question**: Which success metrics should we track first?

**Options**:

**A)** Quantitative only (measurable, objective, data-driven)
- **Metrics**:
  - Prompt usage frequency (via telemetry)
  - Agent activation count
  - Time-to-complete tasks (before/after)
  - Code quality metrics (linting, coverage, mutation score)
- **Pros**: Objective, trackable, comparable
- **Cons**: May miss qualitative insights, requires instrumentation

**B)** Qualitative only (subjective, experiential, feedback-driven)
- **Metrics**:
  - Developer satisfaction surveys
  - Onboarding speed feedback
  - Code review comments on AI-generated code
  - Team feedback sessions
- **Pros**: Captures human experience, actionable insights
- **Cons**: Subjective, harder to quantify, may be biased

**C)** Balanced scorecard (mix of quantitative and qualitative)
- **Metrics**:
  - Quantitative: Prompt usage, time savings, coverage improvements
  - Qualitative: Satisfaction, onboarding feedback, code quality perception
- **Pros**: Comprehensive view, balanced perspective
- **Cons**: More complex tracking, potential metric overload

**D)** Leading indicators (predictive metrics)
- **Metrics**:
  - Prompt creation rate (are we building prompts proactively?)
  - Instruction file updates (are we refining based on learning?)
  - Agent handoff patterns (are agents collaborating effectively?)
  - Collection usage (are we grouping resources well?)
- **Pros**: Proactive, predictive, process-oriented
- **Cons**: May not correlate with outcomes, experimental

**E)** Write-in: ________________

**Answer**: ________________

**Follow-up Questions**:
- How often should we review metrics (weekly, monthly, quarterly)?
- What thresholds trigger intervention (e.g., if prompt usage <10/week)?
- Should we publish metrics to team dashboard?

---

### Q6.3: Continuous Improvement Cadence

**Context**: Copilot configuration should evolve based on feedback and usage. Need regular cadence for reviews and updates.

**Question**: How often should we review and improve Copilot configuration?

**Options**:

**A)** Weekly sprints (fast iteration, responsive to feedback)
- **Cadence**: Every Friday review week's usage, plan next week's improvements
- **Pros**: Rapid iteration, quick response to issues
- **Cons**: May be too frequent, not enough data per cycle
- **Activities**:
  - Review prompt usage telemetry
  - Gather developer feedback
  - Identify quick improvements
  - Implement and deploy

**B)** Bi-weekly reviews (balanced cadence)
- **Cadence**: Every other Friday review 2 weeks' data
- **Pros**: More data per cycle, sustainable cadence
- **Cons**: Slower response to urgent issues
- **Activities**:
  - Analyze 2-week trends
  - Prioritize improvements
  - Implement high-priority items
  - Document learnings

**C)** Monthly retrospectives (strategic improvements)
- **Cadence**: Last Friday of month for comprehensive review
- **Pros**: Strategic view, comprehensive analysis
- **Cons**: Slower feedback loop, may miss quick wins
- **Activities**:
  - Monthly metrics review
  - Team retrospective on Copilot effectiveness
  - Roadmap update
  - Major improvements implementation

**D)** Quarterly planning (long-term optimization)
- **Cadence**: End of quarter for strategic planning
- **Pros**: Aligns with project cycles, strategic focus
- **Cons**: Very slow feedback, may miss opportunities
- **Activities**:
  - Quarterly metrics analysis
  - Stakeholder reviews
  - Major initiatives planning
  - Budget allocation for tools/training

**E)** Write-in: ________________

**Answer**: ________________

**Follow-up Questions**:
- Should we have different cadences for different activities (daily vs. monthly)?
- Who should be involved in reviews (developers, leads, all hands)?
- Should urgent improvements bypass regular cadence?

---

## Section 7: Risk Mitigation

### Q7.1: Instruction File Size Violation Mitigation

**Context**: Risk identified that instruction files may violate 500-line limit. Need proactive mitigation strategy.

**Question**: How should we prevent/detect instruction file size violations?

**Options**:

**A)** Pre-commit hook (block commits with oversized files)
- **Pros**: Enforces limit automatically, prevents violations
- **Cons**: May be disruptive, requires override mechanism for edge cases
- **Implementation**: `.pre-commit-config.yaml` hook checking line counts

**B)** CI check (warn but allow merge)
- **Pros**: Visible but not blocking, allows exceptions
- **Cons**: May accumulate violations over time
- **Implementation**: GitHub Actions workflow checking file sizes

**C)** Monthly audit (review and split as needed)
- **Pros**: Non-disruptive, allows for considered refactoring
- **Cons**: Violations may exist for extended periods
- **Process**: Monthly scan for files >450 lines (early warning)

**D)** Automated splitting (tool automatically splits large files)
- **Pros**: Zero manual effort, always compliant
- **Cons**: May split at poor boundaries, needs review
- **Tool**: Script analyzing content, proposing splits

**E)** Write-in: ________________

**Answer**: ________________

**Follow-up Questions**:
- What's acceptable threshold before mitigation (450, 480, 500 lines)?
- Should we allow temporary exceptions with justification?
- How do we handle edge cases (e.g., one large table)?

---

### Q7.2: Collection Schema Compatibility Risk

**Context**: VS Code collection schema may evolve. Need strategy to ensure our collections remain compatible.

**Question**: How should we ensure collection schema compatibility?

**Options**:

**A)** Manual validation (check against VS Code examples before commit)
- **Pros**: Simple, no tooling needed
- **Cons**: Error-prone, depends on human diligence
- **Process**: Developer checks VS Code docs before committing collection changes

**B)** JSON Schema validation (validate against official schema)
- **Pros**: Automated, catches schema errors early
- **Cons**: Requires schema file, setup overhead
- **Implementation**: 
  - Obtain official VS Code collection schema
  - Add to `.github/schemas/collection.schema.json`
  - Validate in pre-commit hook

**C)** VS Code extension testing (test collections in real VS Code)
- **Pros**: Most realistic validation, catches UX issues
- **Cons**: Manual process, requires VS Code setup
- **Process**: Load collections in VS Code, verify they appear correctly

**D)** Automated E2E testing (test collection functionality)
- **Pros**: Comprehensive, catches integration issues
- **Cons**: Complex setup, requires VS Code automation
- **Implementation**: VS Code extension testing framework, automated collection loading

**E)** Write-in: ________________

**Answer**: ________________

**Follow-up Questions**:
- Should we version-pin VS Code for consistency?
- How do we handle VS Code updates that break collections?
- Should we contribute to VS Code schema documentation if unclear?

---

## Completion Status

**Total Questions**: 19
**Answered**: 0
**Remaining**: 19

**Next Steps**:
1. Answer all questions above
2. Document rationale for each answer
3. Update copilot-enhancement-plan.md with decisions
4. Update copilot-enhancement-tasks.md based on answers
5. Begin Phase 1 implementation

---

**Last Updated**: 2026-01-18
**Status**: Awaiting answers from stakeholders
