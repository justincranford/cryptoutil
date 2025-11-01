# Instruction Files Reorganization Plan

## Executive Summary

This document outlines the comprehensive reorganization of instruction files in `.github/instructions/` to eliminate duplication, improve clarity, and ensure content is in semantically appropriate files.

## Current Issues Identified

### 1. Massive Duplication
- **04-01.specialized-testing.instructions.md** contains complete duplicate of testing content from **02-02.testing.instructions.md**
- Linting guidelines appear in three files: 01-01, 02-03, 02-04
- Git workflow content duplicated between 01-01 and 04-04

### 2. Content Misplacement
- Coding patterns in 01-01 belong in 02-01
- Platform-specific command authorization lists in 01-01 belong in 04-03
- VS Code settings in 01-01 belong in 02-03
- Text encoding guidelines in 01-01 belong in 02-04

### 3. Files Needing Consolidation
- 01-01: Too broad, contains content belonging to 6+ other files
- 02-04: Should be THE authoritative source for ALL linting content
- 04-01: Should focus ONLY on act workflow testing, not duplicate general testing

## Reorganization Strategy

### Phase 1: Content Movement

#### From 01-01.copilot-customization to:
- **→ 02-01.coding**: Code patterns, conditional statement chaining
- **→ 02-03.golang**: VS Code Go development settings, Go-specific patterns
- **→ 02-04.linting**: Text file encoding, lint guidelines, magic values
- **→ 04-03.platform-specific**: Curl/wget rules, authorized commands, command restrictions
- **→ 04-04.git**: Git workflow, conventional commits, TODO docs maintenance, terminal command auto-approval

#### From 04-01.specialized-testing to:
- **→ DELETE**: All duplicated content (testing basics, fuzz testing, test organization)
- **→ KEEP**: Only act workflow testing specifics

#### From 02-03.golang to:
- **→ 02-04.linting**: Linter compliance section (consolidate ALL linting in one place)

### Phase 2: Refactoring Each File

#### 01-01.copilot-customization (SIGNIFICANTLY REDUCED)
**Keep ONLY:**
- General Copilot principles
- GitKraken prohibition
- Python/bash/powershell.exe prohibitions in chat
- Docker Compose secrets immutability rule
- Admin API patterns (HTTPS 127.0.0.1:9090)

**Remove:** Everything else (moved to appropriate files)

#### 02-01.coding (EXPANDED)
**Add from 01-01:**
- Code patterns (default values, pass-through calls)
- Conditional statement chaining
- Switch statement preferences

**Keep existing:** (Currently very minimal)

#### 02-02.testing (REFACTOR)
**Keep:**
- All existing testing content
- Remove duplication with 04-01

**Structure:**
- General testing practices
- Test file organization
- Fuzz testing (complete guidelines)
- Test concurrency
- Copilot testing guidelines
- cicd utility testing patterns
- Script testing requirements

#### 02-03.golang (REFACTOR)
**Add from 01-01:**
- VS Code Go development settings

**Remove:**
- Linter compliance section → move to 02-04

**Keep:**
- Go version consistency
- Project structure
- Application architecture
- Import alias conventions
- Crypto acronym exceptions
- Magic values management
- Dependency management
- Formatting standards (basic)
- Code patterns
- Build flags and linking

#### 02-04.linting (CONSOLIDATE ALL LINTING)
**Add from 01-01:**
- Text file encoding guidelines
- Lint critical guidelines
- Magic values management

**Add from 02-03:**
- Linter compliance section
- Automatic fixing with --fix
- wsl linter compliance
- godot comment period requirements
- detect-secrets inline allowlisting

**Structure:**
- Pre-commit hook documentation maintenance
- Pre-commit configuration guidelines
- Text file encoding
- Linter compliance (all auto-fixable and manual linters)
- Magic number detector (mnd) guidelines
- Code quality standards
- Resource cleanup
- Error wrapping

#### 02-05.security (MINOR REFACTOR)
**Current content is good, minor cleanup:**
- Reorganize for better flow
- Ensure no duplication with other files

#### 02-06.crypto (KEEP AS-IS)
- Already focused and concise
- No changes needed

#### 03-01.docker (REFACTOR FOR CLARITY)
**Keep all content, reorganize:**
- Cross-platform path requirements (critical section)
- Multi-stage build best practices
- Docker secrets best practices
- Networking considerations (localhost vs 127.0.0.1)
- Docker container guidelines
- Sidecar health checks
- Service port reference
- Configuration file requirements

#### 03-02.cicd (REFACTOR FOR CLARITY)
**Keep all content, reorganize:**
- CI/CD cost efficiency
- Workflow architecture overview
- Service connectivity verification patterns
- Configuration management
- Go module caching
- Artifact management
- Act workflow testing

#### 03-03.database (KEEP AS-IS)
- Already focused and concise
- Minor formatting improvements

#### 03-04.observability (KEEP AS-IS)
- Already focused and concise
- No changes needed

#### 04-01.specialized-testing (DRASTICALLY REDUCED)
**DELETE:**
- All duplicated testing content from 02-02

**KEEP ONLY:**
- Title update: "Instructions for act workflow testing"
- Act workflow testing with cmd/workflow utility
- Timing expectations
- Common mistakes specific to act
- CI/CD workflow testing patterns

#### 04-02.openapi (KEEP AS-IS)
- Already concise
- No changes needed

#### 04-03.platform-specific (SIGNIFICANTLY EXPANDED)
**Add from 01-01:**
- Curl/wget command usage rules (context-specific restrictions)
- Authorized commands for chat sessions reference
- Commands requiring manual authorization

**Keep existing:**
- PowerShell development
- Cross-platform script development
- Docker image pre-pull optimization

**Structure:**
- Platform-specific command restrictions (NEW from 01-01)
- Curl/wget usage rules (NEW from 01-01)
- Authorized commands reference (NEW from 01-01)
- PowerShell development
- Cross-platform script development
- Go script guidelines
- PowerShell/Bash script pairings
- Script testing
- Docker image pre-pull

#### 04-04.git (EXPANDED)
**Add from 01-01:**
- Git workflow (commit and push strategy)
- Conventional commits
- Terminal command auto-approval
- TODO docs maintenance

**Keep existing:**
- Pull request descriptions
- Documentation organization

**Structure:**
- Git operations and workflow
- Conventional commits
- Terminal command auto-approval
- Pull request descriptions
- Code review checklist
- PR size guidelines
- Documentation organization
- TODO docs maintenance (NEW from 01-01)

#### 04-05.dast (KEEP AS-IS)
- Already focused and concise
- No changes needed

## Phase 3: Validation

After reorganization:
1. Verify no duplication between files
2. Ensure each file contains ONLY relevant content
3. Check that all critical patterns are preserved
4. Validate that file descriptions accurately reflect content
5. Review for any missing content or gaps

## Implementation Notes

- Create backups before making changes
- Update one file at a time
- Test that instructions are still discoverable
- Update copilot-instructions.md table if needed
- Verify no broken cross-references between files

## Expected Benefits

1. **Reduced duplication**: 40-50% reduction in total instruction content
2. **Improved discoverability**: Content in semantically appropriate files
3. **Easier maintenance**: Changes only need to be made in one place
4. **Better Copilot performance**: Less token usage, more relevant context
5. **Clearer organization**: Each file has a single, well-defined purpose
