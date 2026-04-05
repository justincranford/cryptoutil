# cryptoutil — Claude Code Instructions

## Architecture Source of Truth

| Resource | Purpose |
|----------|---------|
| [docs/ENG-HANDBOOK.md](docs/ENG-HANDBOOK.md) | Canonical source for ALL architectural decisions, patterns, security, testing, deployment, and implementation guidelines (v2.0). Read relevant sections before making decisions. |
| [api/cryptosuite-registry/registry.yaml](api/cryptosuite-registry/registry.yaml) | Machine-readable registry: 10 PS-IDs, port assignments, migration number ranges per PS-ID. |
| [.github/copilot-instructions.md](.github/copilot-instructions.md) | Copilot instructions summary — Claude Code uses this file too. |

### Key ENG-HANDBOOK.md Sections

| Section | Topic |
|---------|-------|
| §1 | Executive summary, entity hierarchy (1 suite → 5 products → 10 PS-IDs) |
| §2 | Agent/skill catalog, architecture strategy, quality principles |
| §3 | Product suite architecture, port assignments |
| §5 | Service architecture, dual HTTPS endpoint pattern, builder pattern |
| §6 | Security: FIPS 140-3, PKI, barrier layer, TLS, key management |
| §7 | Data architecture, dual database strategy, multi-tenancy |
| §8 | API architecture, OpenAPI-first, dual path prefixes |
| §10 | Testing architecture: unit/integration/e2e/fuzz/benchmark/load/mutation |
| §11 | Quality strategy: ≥95% coverage production, ≥98% infrastructure |
| §13 | Deployment, @propagate documentation system |
| §14 | Development practices, Go patterns, import aliases |
| §14.11 | Claude Code autonomous execution modes (beast-mode, plan+execute, standard chat) |

## Instruction Files

Copilot instruction files auto-apply to all Claude Code work in this repo.

@.github/instructions/01-01.terminology.instructions.md
@.github/instructions/01-02.beast-mode.instructions.md
@.github/instructions/02-01.architecture.instructions.md
@.github/instructions/02-02.versions.instructions.md
@.github/instructions/02-03.observability.instructions.md
@.github/instructions/02-04.openapi.instructions.md
@.github/instructions/02-05.security.instructions.md
@.github/instructions/02-06.authn.instructions.md
@.github/instructions/03-01.coding.instructions.md
@.github/instructions/03-02.testing.instructions.md
@.github/instructions/03-03.golang.instructions.md
@.github/instructions/03-04.data-infrastructure.instructions.md
@.github/instructions/03-05.linting.instructions.md
@.github/instructions/04-01.deployment.instructions.md
@.github/instructions/05-01.cross-platform.instructions.md
@.github/instructions/05-02.git.instructions.md
@.github/instructions/06-01.evidence-based.instructions.md
@.github/instructions/06-02.agent-format.instructions.md

## Agents

Custom sub-agents for Claude Code live in [.claude/agents/](.claude/agents/).
Full Copilot originals: [.github/agents/](.github/agents/).

| Agent | When to Use |
|-------|-------------|
| [claude-beast-mode](.claude/agents/beast-mode.md) | Activate for continuous autonomous execution without interruptions or permission requests |
| [claude-fix-workflows](.claude/agents/fix-workflows.md) | GitHub Actions workflow repair and validation |
| [claude-implementation-execution](.claude/agents/implementation-execution.md) | Execute plan.md/tasks.md items autonomously with continuous tasks.md updates |
| [claude-implementation-planning](.claude/agents/implementation-planning.md) | Create/update plan.md + tasks.md + lessons.md scaffold before implementation |

## Skills (Slash Commands)

Copilot skills are available as Claude Code skills in [.claude/skills/](.claude/skills/).
Full Copilot originals: [.github/skills/](.github/skills/).

| Command | Purpose |
|---------|---------|
| `/test-table-driven` | Table-driven Go tests with `t.Parallel`, UUIDv7 test data, subtests |
| `/test-fuzz-gen` | `_fuzz_test.go` with build tags, seed corpus, 15s minimum |
| `/test-benchmark-gen` | `_bench_test.go` for crypto with `ResetTimer`, `SetBytes` |
| `/coverage-analysis` | Identify coverage gaps from coverprofile, categorize by type |
| `/fips-audit` | Detect FIPS 140-3 violations; approved algorithms only |
| `/openapi-codegen` | Generate oapi-codegen configs (server/model/client) + OpenAPI spec skeleton |
| `/migration-create` | Create numbered SQL migration files per registry.yaml ranges |
| `/new-service` | Create new PS-ID service from skeleton-template (9-step guide) |
| `/propagation-check` | Detect `@propagate`/`@source` drift between ENG-HANDBOOK.md and instruction files |
| `/contract-test-gen` | Cross-service contract compliance tests via TestMain pattern |
| `/fitness-function-gen` | New architecture fitness function linter in cicd_lint/lint_fitness/ |
| `/agent-scaffold` | New `.github/agents/NAME.agent.md` + `.claude/agents/NAME.md` dual canonical pair |
| `/instruction-scaffold` | New `.github/instructions/NN-NN.name.instructions.md` |
| `/skill-scaffold` | New `.github/skills/NAME/SKILL.md` with proper YAML frontmatter |
| `/sync-copilot-claude` | Audit/sync Copilot skills+agents with Claude skills+agents |
