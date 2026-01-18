# Copilot Configuration Enhancement - Task Breakdown

**Created**: 2026-01-18
**Plan Reference**: [copilot-enhancement-plan.md](copilot-enhancement-plan.md)
**Status**: Ready for Implementation

---

## Task Checklist

### Phase 1: Foundation (IMMEDIATE - Week 1)

#### 1.1 Create Common Development Prompts ⚠️ IN PROGRESS

**Description**: Create 10 reusable prompt files for frequent development tasks

**Files to Create**:
- [ ] `.github/prompts/code-review.prompt.md` - Code quality review with security/performance checks
- [ ] `.github/prompts/refactor-extract.prompt.md` - Extract method/class/package refactoring
- [ ] `.github/prompts/test-generate.prompt.md` - Generate table-driven tests with coverage
- [ ] `.github/prompts/test-property.prompt.md` - Generate property-based tests using gopter
- [ ] `.github/prompts/fix-bug.prompt.md` - Systematic bug investigation workflow
- [ ] `.github/prompts/optimize-performance.prompt.md` - Profile and optimize hot paths
- [ ] `.github/prompts/generate-docs.prompt.md` - Generate godoc, README, API documentation
- [ ] `.github/prompts/security-audit.prompt.md` - Security review checklist (OWASP, crypto)
- [ ] `.github/prompts/migration-generate.prompt.md` - Generate database migrations
- [ ] `.github/prompts/openapi-client.prompt.md` - Generate API client from OpenAPI spec

**Acceptance Criteria**:
- Each prompt has YAML frontmatter (description, name, agent, tools)
- Each prompt includes workflow steps with examples
- Each prompt references relevant instruction files
- Each prompt includes expected inputs/outputs
- All prompts registered in `.github/copilot-instructions.md`

**LOE**: 4 hours
**Dependencies**: None
**Blocking**: Tasks 1.2, 2.3
**Status**: ❌ Not Started

---

#### 1.2 Create Initial Collections ❌ NOT STARTED

**Description**: Organize prompts and instructions into 3 logical collections

**Files to Create**:
- [ ] `.github/collections/development-workflow.collection.yml` - Plan → Code → Test → Review → Deploy
- [ ] `.github/collections/cryptography.collection.yml` - Crypto, PKI, hashing, auth instructions
- [ ] `.github/collections/quality-gates.collection.yml` - Testing, linting, evidence-based completion

**Collection Structure**:
```yaml
---
name: Collection Name
description: Collection purpose and scope
items:
  prompts:
    - prompt-file-1
    - prompt-file-2
  instructions:
    - instruction-file-1
    - instruction-file-2
  agents:
    - agent-file-1
```

**Acceptance Criteria**:
- Collections use valid YAML schema
- All referenced prompts/instructions exist
- Collections cover 80% of common workflows
- Collections documented in copilot-instructions.md

**LOE**: 2 hours
**Dependencies**: Task 1.1 (prompts must exist first)
**Blocking**: None
**Status**: ❌ Not Started

---

#### 1.3 Add applyTo Patterns to Instructions ❌ NOT STARTED

**Description**: Add conditional application patterns to instruction files for targeted guidance

**Files to Modify** (28 instruction files):
- [ ] `02-03.https-ports.instructions.md` → `applyTo: "**/*server*.go|**/configs/**/*.yml"`
- [ ] `02-06.openapi.instructions.md` → `applyTo: "**/api/**/*.yaml|**/openapi*.yaml"`
- [ ] `03-01.coding.instructions.md` → `applyTo: "**/*.go"`
- [ ] `03-02.testing.instructions.md` → `applyTo: "**/*_test.go|**/*_bench_test.go"`
- [ ] `03-03.golang.instructions.md` → `applyTo: "**/*.go|**/go.mod|**/go.sum"`
- [ ] `03-04.database.instructions.md` → `applyTo: "**/repository/**/*.go|**/migrations/**/*.sql"`
- [ ] `03-05.sqlite-gorm.instructions.md` → `applyTo: "**/repository/**/*.go|**/*_provider.go"`
- [ ] `03-07.linting.instructions.md` → `applyTo: "**/*.go|**/.golangci.yml"`
- [ ] `04-01.github.instructions.md` → `applyTo: ".github/workflows/*.yml"`
- [ ] `04-02.docker.instructions.md` → `applyTo: "**/Dockerfile|**/docker-compose*.yml|**/compose*.yml"`
- [ ] `05-02.git.instructions.md` → `applyTo: "**"`
- [ ] `05-03.dast.instructions.md` → `applyTo: ".github/workflows/ci-dast.yml"`
- [ ] Continue for all 28 files...

**Pattern**:
```markdown
---
description: "Existing description"
applyTo: "**/*.go|**/go.mod"
---
# Existing Content
```

**Acceptance Criteria**:
- All instruction files have `applyTo` frontmatter
- Patterns correctly target relevant file types
- Testing shows reduced context for unrelated files
- Documentation updated with applyTo patterns

**LOE**: 3 hours
**Dependencies**: None (can run in parallel with 1.1)
**Blocking**: None
**Status**: ❌ Not Started

---

#### 1.4 Update VS Code Settings ❌ NOT STARTED

**Description**: Configure chat.promptFilesRecommendations and other Copilot settings

**File to Modify**: `.vscode/settings.json`

**Settings to Add**:
```json
{
  "chat.promptFilesRecommendations": [
    "plan-tasks-quizme",
    "code-review",
    "test-generate",
    "fix-bug",
    "refactor-extract"
  ],
  "chat.instructionsFilesLocations": [
    ".github/instructions"
  ],
  "github.copilot.editor.enableAutoCompletions": true,
  "github.copilot.chat.followUps": "always"
}
```

**Acceptance Criteria**:
- Settings valid JSON
- Prompt recommendations appear in chat
- Instruction files discovered automatically
- Settings documented in DEV-SETUP.md

**LOE**: 1 hour
**Dependencies**: Task 1.1 (prompts must exist)
**Blocking**: None
**Status**: ❌ Not Started

---

### Phase 2: Expansion (MEDIUM TERM - Weeks 2-3)

#### 2.1 Create Domain Expert Agents ❌ NOT STARTED

**Description**: Create 5 specialized agents for domain expertise

**Files to Create**:
- [ ] `.github/agents/expert.security.agent.md` - Security review, threat modeling, OWASP
- [ ] `.github/agents/expert.cryptography.agent.md` - Crypto implementation, FIPS compliance
- [ ] `.github/agents/expert.testing.agent.md` - Test strategy, coverage optimization
- [ ] `.github/agents/expert.performance.agent.md` - Profiling, optimization, benchmarking
- [ ] `.github/agents/expert.database.agent.md` - Schema design, query optimization

**Agent Structure**:
```markdown
---
description: "Agent description"
handoffs:
  - label: "Handoff Label"
    agent: target-agent
    prompt: "Context for handoff"
    send: true
tools:
  - semantic_search
  - grep_search
  - githubRepo
---

# Agent Name

## Expertise
- Domain 1
- Domain 2

## Workflow
1. Step 1
2. Step 2

## References
- Instruction files
- External resources
```

**Acceptance Criteria**:
- Each agent has clear domain boundaries
- Handoffs defined between related agents
- Tools appropriate for agent tasks
- References to relevant instruction files
- Agents documented in copilot-instructions.md

**LOE**: 8 hours
**Dependencies**: None
**Blocking**: None
**Status**: ❌ Not Started

---

#### 2.2 Implement Agent Skills Directory ❌ NOT STARTED

**Description**: Create skills directory structure with code generation templates

**Directories to Create**:
```
.claude/skills/
├── code-generation/
│   ├── skill.yaml
│   └── templates/
│       ├── service-template.go.tmpl
│       ├── repository-template.go.tmpl
│       └── handler-template.go.tmpl
├── test-generation/
│   ├── skill.yaml
│   └── templates/
│       ├── table-driven-test.go.tmpl
│       └── property-test.go.tmpl
└── migration-generation/
    ├── skill.yaml
    └── templates/
        ├── postgres-migration.sql.tmpl
        └── sqlite-migration.sql.tmpl
```

**Skill Configuration Example**:
```yaml
# .claude/skills/code-generation/skill.yaml
name: code-generation
description: Generate Go code from templates
version: 1.0.0
capabilities:
  - generate_service
  - generate_repository
  - generate_handler
resources:
  templates:
    - service-template.go.tmpl
    - repository-template.go.tmpl
```

**Acceptance Criteria**:
- Skills directory follows VS Code agent skills schema
- Templates use Go template syntax
- Skills reference cryptoutil patterns
- SKILLS.md created in workspace root
- Skills accessible to agents

**LOE**: 6 hours
**Dependencies**: None
**Blocking**: None
**Status**: ❌ Not Started

---

#### 2.3 Create Specialized Domain Prompts ❌ NOT STARTED

**Description**: Create 20+ prompts organized by technology domain

**Directory Structure**:
```
.github/prompts/
├── crypto/
│   ├── generate-key-pair.prompt.md
│   ├── implement-encryption.prompt.md
│   ├── implement-signing.prompt.md
│   └── audit-crypto-usage.prompt.md
├── database/
│   ├── design-schema.prompt.md
│   ├── optimize-query.prompt.md
│   ├── generate-migration.prompt.md
│   └── implement-repository.prompt.md
├── testing/
│   ├── generate-unit-tests.prompt.md
│   ├── generate-integration-tests.prompt.md
│   ├── generate-e2e-tests.prompt.md
│   ├── generate-benchmarks.prompt.md
│   └── improve-coverage.prompt.md
└── docs/
    ├── generate-readme.prompt.md
    ├── generate-api-docs.prompt.md
    ├── generate-architecture-docs.prompt.md
    └── generate-runbook.prompt.md
```

**Acceptance Criteria**:
- Each prompt follows YAML frontmatter structure
- Prompts reference domain-specific instructions
- Examples use cryptoutil patterns
- All prompts registered in copilot-instructions.md

**LOE**: 12 hours
**Dependencies**: Task 1.1 (build on common prompts)
**Blocking**: None
**Status**: ❌ Not Started

---

#### 2.4 Create Additional Collections ❌ NOT STARTED

**Description**: Create 5+ collections for specialized workflows

**Files to Create**:
- [ ] `.github/collections/go-development.collection.yml` - Go coding, testing, linting
- [ ] `.github/collections/docker-deployment.collection.yml` - Dockerfile, compose, K8s
- [ ] `.github/collections/database-operations.collection.yml` - Schema, migrations, repos
- [ ] `.github/collections/security-hardening.collection.yml` - Security, crypto, auth
- [ ] `.github/collections/ci-cd-pipeline.collection.yml` - GitHub Actions, workflows

**Acceptance Criteria**:
- Collections cover specific use cases
- No overlap between collections (clear boundaries)
- Collections include relevant prompts/instructions/agents
- Documentation includes when to use each collection

**LOE**: 3 hours
**Dependencies**: Tasks 1.1, 1.2, 2.1, 2.3 (need prompts/agents)
**Blocking**: None
**Status**: ❌ Not Started

---

### Phase 3: Advanced Integration (LONG TERM - Months 2-3)

#### 3.1 Research MCP Server Requirements ❌ NOT STARTED

**Description**: Research Model Context Protocol server integration for cryptoutil

**Research Areas**:
- [ ] Awesome Copilot MCP server capabilities
- [ ] Custom MCP server development requirements
- [ ] Integration with existing cryptoutil architecture
- [ ] Security implications of MCP server
- [ ] Performance considerations

**Deliverables**:
- Research document: `docs/copilot/mcp-server-research.md`
- Recommendation: Build custom vs. use existing
- Architecture proposal if building custom
- Security review of MCP integration

**Acceptance Criteria**:
- Comprehensive research covering all MCP aspects
- Clear recommendation with justification
- Architecture proposal if custom server recommended
- Security review completed
- Stakeholder review and approval

**LOE**: 8 hours
**Dependencies**: None
**Blocking**: Task 3.2 (must complete research first)
**Status**: ❌ Not Started

---

#### 3.2 Prototype Custom MCP Server ❌ NOT STARTED

**Description**: Build prototype MCP server for cryptoutil-specific operations

**Scope** (based on Task 3.1 research outcome):
- [ ] MCP server scaffold using Go
- [ ] Integration with KMS operations
- [ ] Integration with PKI operations
- [ ] Integration with test execution
- [ ] Integration with deployment tools

**Directories to Create**:
```
internal/mcp/
├── server/
│   ├── server.go
│   ├── handlers.go
│   └── config.go
├── cryptoutil-kms/
├── cryptoutil-pki/
├── cryptoutil-testing/
└── cryptoutil-deployment/
```

**Acceptance Criteria**:
- MCP server implements protocol specification
- Secure authentication/authorization
- Operations logged and audited
- Integration tests pass
- Documentation complete

**LOE**: 40 hours
**Dependencies**: Task 3.1 (research must be complete and approved)
**Blocking**: None
**Status**: ❌ Not Started

---

#### 3.3 Create Continuous Improvement Workflow ❌ NOT STARTED

**Description**: Implement feedback loops for Copilot configuration optimization

**Files to Create**:
- [ ] `.github/prompts/improve-copilot.prompt.md` - Analyze and improve configuration
- [ ] `scripts/analyze-copilot-usage.ps1` - Parse usage telemetry
- [ ] `docs/copilot/usage-analytics.md` - Usage patterns documentation

**Workflow**:
1. Collect Copilot usage telemetry
2. Analyze prompt activation frequency
3. Identify missing prompts for common tasks
4. Detect redundant or outdated patterns
5. Generate improvement recommendations
6. Implement and test improvements
7. Track effectiveness metrics

**Acceptance Criteria**:
- Telemetry collection working
- Analytics scripts generate reports
- Improvement workflow documented
- Metrics tracked over time
- Feedback incorporated monthly

**LOE**: 16 hours
**Dependencies**: Tasks 1.1-2.4 (need baseline to measure against)
**Blocking**: None
**Status**: ❌ Not Started

---

#### 3.4 Build Team Knowledge Repository ❌ NOT STARTED

**Description**: Create comprehensive knowledge base for Copilot customization

**Directory Structure**:
```
docs/copilot/
├── best-practices/
│   ├── prompt-engineering.md
│   ├── instruction-writing.md
│   ├── agent-design.md
│   └── collection-organization.md
├── examples/
│   ├── complex-prompts.md
│   ├── multi-agent-workflows.md
│   └── conditional-instructions.md
├── metrics/
│   ├── usage-statistics.md
│   ├── effectiveness-reports.md
│   └── improvement-tracking.md
└── troubleshooting/
    ├── common-issues.md
    ├── debugging-guide.md
    └── faq.md
```

**Acceptance Criteria**:
- All sections populated with content
- Examples from real cryptoutil usage
- Metrics tracked and updated monthly
- Troubleshooting covers common scenarios
- Documentation searchable and organized

**LOE**: 24 hours
**Dependencies**: Tasks 1.1-3.3 (need implementation experience)
**Blocking**: None
**Status**: ❌ Not Started

---

## Summary Statistics

### By Phase

**Phase 1** (Foundation):
- Tasks: 4
- Total LOE: 10 hours
- Status: 0/4 complete (0%)

**Phase 2** (Expansion):
- Tasks: 4
- Total LOE: 29 hours
- Status: 0/4 complete (0%)

**Phase 3** (Advanced):
- Tasks: 4
- Total LOE: 88 hours
- Status: 0/4 complete (0%)

### Overall Project

- **Total Tasks**: 12
- **Total LOE**: 127 hours (~3-4 weeks full-time)
- **Completion**: 0/12 (0%)
- **Blocking Issues**: None
- **Critical Path**: Phase 1 → Phase 2 → Phase 3 (sequential dependencies)

---

## Risk Assessment

### High Priority Risks

**Risk 1**: Instruction file size violations
- **Probability**: High
- **Impact**: Medium
- **Mitigation**: Split files in Task 1.3, track line counts

**Risk 2**: Collection schema compatibility
- **Probability**: Medium
- **Impact**: High
- **Mitigation**: Validate against VS Code examples, test thoroughly

**Risk 3**: MCP server complexity
- **Probability**: High
- **Impact**: High
- **Mitigation**: Thorough research (Task 3.1) before prototype

### Medium Priority Risks

**Risk 4**: Prompt discoverability
- **Probability**: Medium
- **Impact**: Medium
- **Mitigation**: Good naming, collections, recommendations settings

**Risk 5**: Agent handoff complexity
- **Probability**: Medium
- **Impact**: Low
- **Mitigation**: Clear documentation, test workflows

---

## Next Actions

1. **Answer QUIZME questions** in copilot-enhancement-plan.md
2. **Start Task 1.1** (Create Common Development Prompts)
3. **Commit progress** after each completed task
4. **Track metrics** as new prompts are used
5. **Iterate based on feedback** from team usage

---

**Last Updated**: 2026-01-18
**Status**: Ready for implementation
