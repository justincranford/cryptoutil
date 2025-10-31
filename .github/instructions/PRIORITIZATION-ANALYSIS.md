# Copilot Instruction File Prioritization Analysis

**Analysis Date**: October 31, 2025  
**Current File Count**: 31 instruction files  
**Target**: Optimize for 6-file loading limit with strategic consolidation

## Context

GitHub Copilot Chat in VS Code loads approximately 6 instruction files per session based on:
- Alphabetical ordering (strong influence)
- Token budget constraints
- File size impact
- `copilot-instructions.md` (always prioritized)

## Categorization Matrix

### TIER 1: CRITICAL (Must Always Load - Slots 1-6)

| File | Size | Frequency | Impact | Rationale |
|------|------|-----------|--------|-----------|
| `copilot-customization` | Large | Every Session | Critical | Git ops, terminal patterns, curl/wget rules, fuzz testing, command authorization |
| `code-quality` | Large | Every Go Edit | Critical | Linter compliance, wsl rules, godot rules, resource cleanup, pre-commit docs |
| `testing` | Medium | Every Test | Critical | Test patterns, dependency mgmt, file organization, UUIDv7 concurrency, cicd_test.go |
| `architecture` | Small | Every Feature | Critical | Layered arch, config patterns, lifecycle, factory patterns, atomic ops |
| `security` | Medium | Crypto Project | Critical | Key hierarchy, IP allowlisting, rate limiting, TLS, secrets management |
| `docker` | Large | Infrastructure | Critical | Compose config, healthchecks, secrets, relative paths, otel forwarding |

**Combined Token Estimate**: ~6500 tokens (manageable for 6-file budget)

### TIER 2: HIGH PRIORITY (Important - Slots 7-12)

| File | Size | Frequency | Impact | Consolidation Candidate |
|------|------|-----------|--------|------------------------|
| `crypto` | Small | Crypto Code | High | Standalone (project-specific) |
| `cicd` | Medium | Workflows | High | Standalone (Go version critical) |
| `observability` | Medium | Telemetry Code | High | Standalone (OTLP architecture) |
| `database` | Small | DB Code | High | Standalone (GORM patterns) |
| `git` | Small | Git Ops | Medium | ✅ Merge into `copilot-customization` |
| `commits` | Small | Git Ops | Medium | ✅ Merge into `copilot-customization` |
| `terminal-auto-approve` | Medium | Terminal Ops | Medium | ✅ Merge into `copilot-customization` |
| `todo-maintenance` | Small | Maintenance | Low | ✅ Merge into `copilot-customization` |

**Consolidation Target**: `copilot-customization` ← `git` + `commits` + `terminal-auto-approve` + `todo-maintenance`

### TIER 3: MEDIUM PRIORITY (Contextual - Slots 13-18)

| File | Consolidation Target |
|------|---------------------|
| `imports` | ✅ **go-standards** ← `imports` + `go-dependencies` + `formatting` + `conditional-chaining` |
| `go-dependencies` | ✅ **go-standards** |
| `formatting` | ✅ **go-standards** |
| `conditional-chaining` | ✅ **go-standards** |
| `act-testing` | ✅ **specialized-testing** ← `act-testing` + `localhost-vs-ip` |
| `localhost-vs-ip` | ✅ **specialized-testing** |
| `openapi` | ✅ **project-config** ← `openapi` + `magic-values` + `linting-exclusions` |
| `magic-values` | ✅ **project-config** |
| `linting-exclusions` | ✅ **project-config** |
| `powershell` | ✅ **platform-specific** ← `powershell` + `scripts` + `docker-prepull` |
| `scripts` | ✅ **platform-specific** |
| `docker-prepull` | ✅ **platform-specific** |

### TIER 4: LOW PRIORITY (Specialized - Slots 19-20)

| File | Consolidation Target |
|------|---------------------|
| `cabf` | ✅ **specialized-domains** ← `cabf` + `project-layout` + `pull-requests` |
| `project-layout` | ✅ **specialized-domains** |
| `pull-requests` | ✅ **specialized-domains** |
| `documentation` | ✅ **specialized-domains** |
| `errors` | ⚠️ Already covered in `code-quality` - consider removing |

## Consolidation Strategy

### Phase 2 Plan: Reduce 31 → 18 Files

**New Structure**:

```
Tier 1 (Always Load - Prefix 01-06):
01-copilot-customization.instructions.md  ← Enhanced with git/commits/terminal/todo
02-code-quality.instructions.md
03-testing.instructions.md
04-architecture.instructions.md
05-security.instructions.md
06-docker.instructions.md

Tier 2 (High Priority - Prefix 07-12):
07-crypto.instructions.md
08-cicd.instructions.md
09-observability.instructions.md
10-database.instructions.md
11-go-standards.instructions.md          ← NEW: imports + go-deps + formatting + conditional
12-development-workflow.instructions.md   ← REMOVED (merged into 01)

Tier 3 (Medium Priority - Prefix 13-18):
13-specialized-testing.instructions.md    ← NEW: act-testing + localhost-vs-ip
14-project-config.instructions.md         ← NEW: openapi + magic-values + linting-exclusions
15-platform-specific.instructions.md      ← NEW: powershell + scripts + docker-prepull

Tier 4 (Low Priority - No Prefix):
specialized-domains.instructions.md       ← NEW: cabf + project-layout + pull-requests + documentation
```

**File Count**: 31 → 16 files (-48% reduction)

## Token Optimization Targets

| File | Current Est. | Target | Method |
|------|-------------|--------|--------|
| `copilot-customization` | ~2000 | ~1600 | Remove verbose decision trees, consolidate examples |
| `code-quality` | ~1800 | ~1400 | Consolidate linter sections, remove redundant examples |
| `testing` | ~800 | ~650 | Simplify dependency management section |
| `docker` | ~1500 | ~1200 | Consolidate healthcheck patterns |
| `security` | ~700 | ~550 | Bullet-point only, remove verbose explanations |
| `architecture` | ~400 | ~300 | Already concise, minimal reduction |

**Total Token Reduction**: ~1800 tokens (~22% reduction)

## Expected Outcomes

### With Alphabetical Prefixing Only (Phase 1)
- ✅ Guarantees Tier 1 (slots 1-6) always loads
- ⚠️ Remaining 25 files compete for context
- ⚠️ No token optimization

### With Full Consolidation (Phase 2)
- ✅ Guarantees Tier 1 always loads
- ✅ High probability Tier 2 (slots 7-12) loads
- ✅ 16 total files = less competition
- ✅ 22% token reduction = more room per file

### With Token Optimization (Phase 3)
- ✅ All benefits of Phase 2
- ✅ Even more files can fit in context
- ✅ Faster Copilot response times
- ✅ Lower API costs

## Risk Assessment

| Risk | Mitigation |
|------|------------|
| Lose important guidance in consolidation | Cross-reference sections within consolidated files |
| Files become too large | Target <2000 tokens per file |
| Alphabetical prefix confusing | Use semantic prefixes (01-copilot, 02-quality, etc.) |
| Breaking existing references | Update copilot-instructions.md index table |

## Recommendation

**Execute all three phases**:
1. ✅ Phase 1: Alphabetical prefixing (commits immediately)
2. ✅ Phase 2: Strategic consolidation (commits immediately)
3. ✅ Phase 3: Token optimization (commits immediately)

**Total Time**: ~45-60 minutes
**Risk**: Low (all changes are additive/renaming)
**Benefit**: High (guaranteed critical instruction loading)
