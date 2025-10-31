# Copilot Instruction File Prioritization Guidelines

**Document Purpose**: Guidelines for maintaining and reordering instruction files  
**Naming Convention**: Semantic names with ##-##. prefix (Tier-Priority format)

## VS Code Copilot Instruction Loading Behavior

According to [VS Code documentation](https://code.visualstudio.com/docs/copilot/customization/custom-instructions):

- **Multiple instruction files supported**: VS Code combines all `.instructions.md` files from `.github/instructions/`
- **No guaranteed order**: "no specific order is guaranteed" when loading multiple files
- **Automatic discovery**: Files are discovered automatically with `.instructions.md` extension
- **YAML frontmatter**: Use `applyTo` property to specify glob patterns for file matching
- **Load order influence**: Alphabetical ordering has strong influence (observed behavior, not documented)
- **Token budget constraints**: Practical limits exist based on model context window

## Naming Convention: ##-##. Prefix

**Purpose**: Explicit priority control through alphabetical ordering with semantic readability

- **First ##** = Tier number (01-04, priority level)
- **Second ##** = Priority within tier (01-05, load sequence)
- **Format**: `##-##.semantic-name.instructions.md`
- **Example**: `02-03.architecture.instructions.md` (Tier 2, Priority 3)

### Benefits
- **Explicit hierarchy**: Tier visible in filename
- **Flexible reordering**: Change priority by renaming file
- **Maintainable**: Easy to add/remove/reorder files
- **Self-documenting**: Structure clear from filenames
- **Semantic clarity**: Descriptive names improve readability over abstract codes
- **No linter conflicts**: Numeric prefixes avoid markdown tool interpretation issues

## Current Tier Structure

### TIER 1: FOUNDATION (Slot 1 - Always Loads First)

| File | Frequency | Impact | Content |
|------|-----------|--------|---------|
| `01-01.copilot-customization` | Every Session | Critical | Git ops, terminal patterns, curl/wget rules, fuzz testing, command authorization, git workflow, conventional commits, terminal auto-approval, TODO maintenance |

**Token Estimate**: ~2500 tokens

### TIER 2: CORE DEVELOPMENT (Slots 2-6 - Essential)

| File | Frequency | Impact | Content |
|------|-----------|--------|---------|
| `02-01.golang` | Every Go Edit | Critical | Go project structure, architecture, and coding standards |
| `02-02.testing` | Every Test | Critical | Test patterns, dependency mgmt, file organization, UUIDv7 concurrency |
| `02-03.security` | Crypto Project | Critical | Key hierarchy, IP allowlisting, rate limiting, TLS, secrets |
| `02-04.code-quality` | Every Go Edit | Critical | Linter compliance, wsl rules, godot rules, resource cleanup, pre-commit docs |
| `02-05.crypto` | Crypto Code | High | Cryptographic operations and CA/Browser Forum requirements |

**Combined Token Estimate**: ~6500 tokens

### TIER 3: HIGH PRIORITY (Slots 7-11 - Important)

| File | Frequency | Impact | Source Files |
|------|-----------|--------|--------------|
| `03-01.docker` | Infrastructure | High | Compose config, healthchecks, secrets, OTEL forwarding |
| `03-02.cicd` | Workflows | High | Standalone |
| `03-03.observability` | Telemetry | High | Standalone |
| `03-04.database` | DB Code | High | Standalone |
| `03-05.go-standards` | Go Code | High | imports + go-dependencies + formatting + conditional-chaining |

### TIER 4: MEDIUM PRIORITY (Slots 12-15 - Contextual)

| File | Source Files |
|------|--------------|
| `04-01.specialized-testing` | act-testing + localhost-vs-ip |
| `04-02.project-config` | openapi + magic-values + linting-exclusions |
| `04-03.platform-specific` | powershell + scripts + docker-prepull |
| `04-04.specialized-domains` | cabf + project-layout + pull-requests + documentation |

**Note**: `errors.instructions.md` removed (content in `02-04.code-quality`)

## Reordering Instructions

### To Change Priority Within Tier
Rename file with new second ## value:
```bash
git mv .github/instructions/02-01.golang.instructions.md .github/instructions/02-03.golang.instructions.md
git mv .github/instructions/02-04.code-quality.instructions.md .github/instructions/02-01.code-quality.instructions.md
```

### To Move File Between Tiers
Change first ## prefix:
```bash
git mv .github/instructions/02-05.crypto.instructions.md .github/instructions/03-05.crypto.instructions.md
```

### To Add New Instruction File
1. Create file with appropriate ##-##. prefix
2. Use `.instructions.md` extension
3. Add YAML frontmatter with `applyTo` and `description`
4. Git add and commit

### Best Practices
- Maintain sequential numbers within each tier (01, 02, 03, 04, 05)
- Leave gaps (e.g., 01, 03, 05) for easier insertion
- Update PRIORITIZATION-GUIDELINES.md after changes
- Update copilot-instructions.md cross-references
- Test with new structure before committing

## Token Optimization Targets (Optional Future Work)

Once consolidation is validated, consider further optimization:

| File | Current Est. | Target | Method |
|------|-------------|--------|--------|
| `01-01.copilot-customization` | ~2500 | ~2000 | Remove verbose decision trees, consolidate command examples |
| `02-04.code-quality` | ~1800 | ~1400 | Consolidate linter sections, remove redundant examples |
| `02-02.testing` | ~800 | ~650 | Simplify dependency management section |
| `03-01.docker` | ~1500 | ~1200 | Consolidate healthcheck patterns |
| `02-03.security` | ~700 | ~550 | Bullet-point only, remove verbose explanations |
| `02-01.golang` | ~400 | ~300 | Already concise, minimal reduction |

**Potential Token Reduction**: ~1800 tokens (~22% additional reduction)

**Approach**:
- Review and minimize redundant text
- Consolidate bullet points where possible
- Remove verbose explanations that can be inferred
- Preserve all critical guidance and patterns

## Final Outcomes

### Next Steps (Optional Enhancement)

**Token Optimization** (Not yet started):
- Review individual files for verbosity
- Target 20-30% reduction per file without losing critical information
- Focus on T1-P1-copilot-customization and T2-P1-code-quality (largest files)

**Testing & Validation** (Not yet started):
- Verify Copilot loads expected files in common workflows
- Test common development scenarios (Go editing, Docker operations, testing)
- Ensure no regression in instruction quality
- Validate ##-##. reordering mechanics

**Maintenance** (Ongoing):
- Use ##-##. reordering as project needs evolve
- Update PRIORITIZATION-GUIDELINES.md when structure changes
- Keep copilot-instructions.md cross-references current
- Monitor file sizes to prevent token budget issues

## Risk Assessment

| Risk | Mitigation |
|------|------------|
| Lose important guidance in consolidation | Cross-reference sections within consolidated files |
| Files become too large | Target <2000 tokens per file |
| Numeric prefix confusing | Use semantic names with ##-##. format (01-copilot, 02-quality, etc.) |
| Breaking existing references | Update copilot-instructions.md index table |
