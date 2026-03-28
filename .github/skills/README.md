# Copilot Skills

Skills provide targeted context for specific tasks in VS Code Copilot Chat.
Each skill lives in its own subdirectory with a `SKILL.md` file defining its
behavior. VS Code loads skill metadata (name + description) for discovery, then
loads the full `SKILL.md` body only when the skill is relevant or invoked.

## Structure

```
.github/skills/
└── skill-name/        # Directory name must match `name` in SKILL.md frontmatter
    └── SKILL.md       # Required: YAML frontmatter + instructions
    └── ...            # Optional: scripts, templates, examples
```

## Invoking Skills

Reference a skill using `/skill-name` as a slash command in Copilot Chat, or let
Copilot auto-load it when your request matches the skill description.

## Skill Catalogue

| Skill | Domain | Purpose | File |
|-------|--------|---------|------|
| `test-table-driven` | testing | Generate table-driven Go tests (t.Parallel, UUIDv7 data, subtests) | [SKILL.md](test-table-driven/SKILL.md) |
| `test-fuzz-gen` | testing | Generate `_fuzz_test.go` (15s fuzz time, corpus examples, build tags) | [SKILL.md](test-fuzz-gen/SKILL.md) |
| `test-benchmark-gen` | testing | Generate `_bench_test.go` (mandatory for crypto, reset timer pattern) | [SKILL.md](test-benchmark-gen/SKILL.md) |
| `coverage-analysis` | testing | Analyze coverprofile output, categorize uncovered lines, suggest tests | [SKILL.md](coverage-analysis/SKILL.md) |
| `fips-audit` | security | Detect FIPS 140-3 violations and provide fix guidance | [SKILL.md](fips-audit/SKILL.md) |
| `openapi-codegen` | api | Generate three oapi-codegen configs (server/model/client) + OpenAPI spec skeleton | [SKILL.md](openapi-codegen/SKILL.md) |
| `migration-create` | data | Create numbered golang-migrate SQL files (template 1001-1999, domain 2001+) | [SKILL.md](migration-create/SKILL.md) |
| `new-service` | architecture | Guide service creation from skeleton-template: copy, rename, register, migrate, test | [SKILL.md](new-service/SKILL.md) |
| `propagation-check` | docs | Detect @propagate/@source drift, generate corrected @source blocks | [SKILL.md](propagation-check/SKILL.md) |
| `agent-scaffold` | tooling | Create conformant `.github/agents/NAME.agent.md` with all mandatory sections | [SKILL.md](agent-scaffold/SKILL.md) |
| `instruction-scaffold` | tooling | Create conformant `.github/instructions/NN-NN.name.instructions.md` | [SKILL.md](instruction-scaffold/SKILL.md) |
| `skill-scaffold` | tooling | Create conformant `.github/skills/NAME/SKILL.md` with proper frontmatter | [SKILL.md](skill-scaffold/SKILL.md) |
| `contract-test-gen` | testing | Generate cross-service contract compliance tests (RunContractTests, ServiceServer, SetReady) | [SKILL.md](contract-test-gen/SKILL.md) |
| `fitness-function-gen` | testing | Generate architecture fitness functions for lint-fitness (Check, CheckInDir, registration) | [SKILL.md](fitness-function-gen/SKILL.md) |

## Skills vs Custom Instructions vs Agents

| Type | File | How Used | Scope |
|------|------|----------|-------|
| **Skills** | `.github/skills/NAME/SKILL.md` | `/skill-name` slash command or auto-load | On-demand, task-specific |
| **Instructions** | `.github/instructions/NN-NN.name.instructions.md` | Auto-applied by file pattern | Always-on (or pattern-based) |
| **Agents** | `.github/agents/NAME.agent.md` | `/agent-name` slash command | Specialized autonomous workflows |

Use **skills** for:
- Reusable templates and code generation patterns
- Task-specific guidance (test writing, crypto auditing, migrations)
- Capabilities that work across different sessions without loading always

Use **instructions** instead for:
- Always-on coding standards (architecture patterns, naming conventions)
- Rules that apply to every chat response (formatting, error handling)

Use `/skill-scaffold` to create new skills, `/instruction-scaffold` for new instructions,
`/agent-scaffold` for new agents.
