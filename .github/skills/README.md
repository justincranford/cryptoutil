# Copilot Skills

Skills provide targeted context for specific tasks in VS Code Copilot Chat. Reference a skill using `#skill-name` in any Copilot Chat message.

## File Naming Convention

- All skills use flat, kebab-case filenames: `SKILLNAME.md` in `.github/skills/`
- Skill names use hyphens: `test-table-driven`, `migration-create`
- Referenced in chat as: `#test-table-driven`, `#migration-create`

## Available Skills

| Skill | Purpose |
|-------|---------|
| `#test-table-driven` | Generate table-driven Go tests (t.Parallel, UUIDv7 data, subtests) |
| `#test-fuzz-gen` | Generate `_fuzz_test.go` with 15s fuzz time and corpus examples |
| `#test-benchmark-gen` | Generate `_bench_test.go` (mandatory for crypto operations) |
| `#migration-create` | Create numbered golang-migrate SQL files (template 1001+, domain 2001+) |
| `#coverage-analysis` | Analyze coverprofile output, categorize uncovered lines, suggest targeted tests |
| `#fips-audit` | Detect FIPS 140-3 violations and provide fix guidance |
| `#propagation-check` | Detect @propagate/@source drift, generate corrected @source blocks |
| `#openapi-codegen` | Generate oapi-codegen configs (server/model/client) + OpenAPI spec skeleton |
| `#agent-scaffold` | Create conformant `.github/agents/NAME.agent.md` with all mandatory sections |
| `#instruction-scaffold` | Create conformant `.github/instructions/NN-NN.name.instructions.md` |
| `#skill-scaffold` | Create conformant `.github/skills/NAME.md` |
| `#new-service` | Guide service creation from skeleton-template: copy, rename, register, migrate, test |

## Three Copilot Customization Types

VS Code Copilot has exactly 3 customization file types:

| Type | Pattern | Trigger |
|------|---------|---------|
| Instructions | `.github/instructions/*.instructions.md` | Always loaded |
| Agents | `.github/agents/*.agent.md` | `/agent-name` invocation |
| Skills | `.github/skills/*.md` | `#skill-name` in chat |

See [ARCHITECTURE.md Section 2.1.5](../../../docs/ARCHITECTURE.md#215-copilot-skills) for the complete skill catalogue.

## Creating New Skills

Use `#skill-scaffold` to create a new skill, or copy `SKILL-TEMPLATE.md`.
