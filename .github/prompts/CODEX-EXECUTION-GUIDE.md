# Codex Agent Autonomous Execution Guide

## Overview

This guide provides rock-solid instructions for executing large plan+tasks documents using Codex Agent with Claude Opus 4.5.

**Target Use Case**: JOSE-JA refactoring (10 phases, 142 tasks, ~8-20 hours estimated)

---

## Prerequisites

### 1. Ensure Clean Git State

```powershell
# Check status
git status

# Stash any uncommitted changes
git stash

# Create execution branch
git checkout -b jose-ja-refactor-codex
```

### 2. Verify Development Environment

```powershell
# Go version
go version  # Should be 1.25.5+

# Dependencies
go mod download
go mod tidy

# Baseline tests pass
go test ./...

# Linting clean
golangci-lint run --fix ./...
```

### 3. Document Baseline State

```powershell
# Generate baseline coverage
go test ./... -coverprofile=./test-output/baseline_coverage.out
go tool cover -html=./test-output/baseline_coverage.out -o ./test-output/baseline_coverage.html

# Count existing TODOs
grep -r "TODO\|FIXME" . --include="*.go" | wc -l
```

---

## Codex Agent Setup

### Access Codex Agent

1. Navigate to **Anthropic Console**: https://console.anthropic.com
2. Select **"Codex"** from navigation menu
3. Choose **Claude Opus 4.5** model
4. Set workspace to your cryptoutil project directory

### Configure Agent Settings

**Model**: Claude Opus 4.5
**Max Tokens**: Unlimited (or maximum available)
**Temperature**: 0.0 (deterministic execution)
**Workspace**: `c:\Dev\Projects\cryptoutil`
**Git Integration**: Enabled (for automatic commits)

---

## Execution Prompt

Copy this EXACT prompt into Codex Agent:

```
I need you to execute a complete refactoring plan autonomously.

STEP 1: Read and internalize these files completely:
- .github/prompts/autonomous-execution.prompt.md
- docs/jose-ja/JOSE-JA-REFACTORING-PLAN-V3.md
- docs/jose-ja/JOSE-JA-REFACTORING-TASKS-V3.md

STEP 2: Replace placeholders in the autonomous execution prompt:
- {{PLAN_FILE_PATH}} = docs/jose-ja/JOSE-JA-REFACTORING-PLAN-V3.md
- {{TASKS_FILE_PATH}} = docs/jose-ja/JOSE-JA-REFACTORING-TASKS-V3.md
- {{DETAILED_DOC_PATH}} = specs/002-cryptoutil/implement/DETAILED.md

STEP 3: Follow the autonomous execution prompt EXACTLY as written.

Requirements:
- Execute ALL 10 phases and ALL 142 tasks sequentially
- Apply ALL quality gates after EVERY task (build, lint, test, coverage)
- Commit after EVERY completed task (conventional commit format)
- Push every 5-10 commits so I can monitor progress
- Append timeline entry to DETAILED.md after each PHASE
- Do NOT ask questions, pause, or request confirmation
- Do NOT stop until ALL tasks complete OR I explicitly interrupt

BEGIN EXECUTION NOW.
```

---

## Monitoring Progress

### Real-Time Monitoring

**Option 1: Git Log (Local)**
```powershell
# Watch commits in real-time (run in separate terminal)
while ($true) {
    git log --oneline -20
    Start-Sleep -Seconds 30
}
```

**Option 2: GitHub Web (Remote)**
```powershell
# Ensure codex pushes to remote
git push -u origin jose-ja-refactor-codex

# Monitor at: https://github.com/<user>/cryptoutil/commits/jose-ja-refactor-codex
```

**Option 3: DETAILED.md Timeline**
```powershell
# Watch timeline updates
Get-Content specs/002-cryptoutil/implement/DETAILED.md | Select-String "^### 202"
```

### Progress Checkpoints

| Metric | Command |
|--------|---------|
| **Tasks Completed** | `git log --oneline --grep="refactor\|feat\|fix" \| wc -l` |
| **Current Phase** | `Get-Content specs/002-cryptoutil/implement/DETAILED.md \| Select-String "^### 202" \| Select-Object -Last 1` |
| **Build Status** | `go build ./...` |
| **Test Status** | `go test ./... -short` |
| **Coverage** | `go test ./... -coverprofile=coverage.out ; go tool cover -func=coverage.out` |

---

## Intervention Points

### When to Interrupt

**ONLY interrupt if**:
- Repeated failures on same task (>3 attempts)
- Test suite becomes persistently red
- Coverage drops below thresholds
- Obvious wrong direction

**HOW to interrupt**:
1. Stop Codex execution (button in UI)
2. Review git log and DETAILED.md
3. Identify blocker
4. Fix manually or provide guidance
5. Resume with modified prompt

### Resuming After Interruption

```
Continue autonomous execution from where you stopped.

Current state:
- Last completed: Phase X, Task Y.Z
- Next task: Phase X, Task Y.Z+1

Review recent commits to understand context.
Resume with full autonomous execution mode.
Do not re-do completed tasks.
```

---

## Validation After Completion

### Quality Verification

```powershell
# 1. All builds pass
go build ./...

# 2. All tests pass
go test ./...

# 3. Linting clean
golangci-lint run ./...

# 4. Coverage meets targets
go test ./... -coverprofile=final_coverage.out
go tool cover -func=final_coverage.out | grep "total:"
# Should show ≥95% for production, ≥98% for infrastructure

# 5. Mutation testing
gremlins unleash --tags=!integration

# 6. No new TODOs
$newTODOs = (grep -r "TODO\|FIXME" . --include="*.go" | wc -l)
echo "New TODOs: $($newTODOs - $baselineTODOs)"
```

### Git Verification

```powershell
# Review commit history
git log --oneline --graph

# Verify conventional commits
git log --oneline | grep -E "^[a-f0-9]+ (feat|fix|refactor|test|docs|style|chore)"

# Check for evidence-based commits
git log --grep="coverage" --grep="tests pass" --grep="build clean"
```

### Integration Testing

```powershell
# Start full stack
docker compose -f deployments/compose/compose.yml up -d

# Wait for health
Start-Sleep -Seconds 30

# Run E2E tests
go test ./test/e2e/... -v

# Cleanup
docker compose -f deployments/compose/compose.yml down
```

---

## Troubleshooting

### Codex Stops Unexpectedly

**Symptom**: Execution halts, asks questions, or requests confirmation

**Fix**: Restate the autonomous execution directive

```
You stopped execution. This violates the autonomous execution prompt.

Re-read .github/prompts/autonomous-execution.prompt.md

RESUME autonomous execution from current task.
Do NOT ask questions. Do NOT pause. Do NOT summarize.
```

### Build Failures

**Symptom**: `go build ./...` fails repeatedly

**Fix**: Check if codex is fixing or looping

```powershell
# Check recent commits
git log --oneline -10

# If looping (same fix repeated):
# 1. Stop codex
# 2. Review error
# 3. Fix manually
# 4. Commit
# 5. Resume codex
```

### Test Failures

**Symptom**: Tests fail after changes

**Fix**: Codex should auto-fix per quality gates, but verify:

```powershell
# Run tests with verbose output
go test ./... -v

# Check if tests were updated
git log --oneline --grep="test"

# If tests not updated, intervene
```

### Coverage Drops

**Symptom**: Coverage below thresholds (95%/98%)

**Fix**: Codex should auto-fix, but monitor:

```powershell
# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Open in browser to find gaps
start coverage.html
```

---

## Post-Execution

### Final Review

1. **Code Review**: Review key changes in GitHub PR
2. **Documentation**: Verify DETAILED.md timeline is complete
3. **Testing**: Run full test suite + mutation testing
4. **Merge**: Merge execution branch to main

### Cleanup

```powershell
# Squash commits if desired (optional)
git rebase -i main

# Merge to main
git checkout main
git merge jose-ja-refactor-codex

# Push
git push origin main

# Delete branch
git branch -d jose-ja-refactor-codex
git push origin --delete jose-ja-refactor-codex
```

---

## Expected Timeline

| Phase | Tasks | Estimated Time |
|-------|-------|----------------|
| 0 | 12 | 1-2 hours |
| 1 | 15 | 1-2 hours |
| 2 | 18 | 2-3 hours |
| 3 | 20 | 2-3 hours |
| 4 | 16 | 1-2 hours |
| 5 | 14 | 1-2 hours |
| 6 | 12 | 1-2 hours |
| 7 | 18 | 2-3 hours |
| 8 | 10 | 1 hour |
| 9 | 7 | 1 hour |
| **Total** | **142** | **~12-20 hours** |

**Note**: Codex can run continuously for days if needed. Time estimates assume no major blockers.

---

## Success Criteria

Execution is considered successful when:

✅ All 142 tasks completed
✅ All quality gates pass globally
✅ Coverage ≥95% production, ≥98% infrastructure
✅ Mutation score ≥85% production, ≥98% infrastructure
✅ All tests pass (zero skips)
✅ Build clean (zero errors)
✅ Linting clean (zero warnings)
✅ DETAILED.md timeline complete (10 phase entries)
✅ Git history shows evidence-based commits
✅ E2E tests pass

---

## Fallback Plan

If Codex proves unsuitable:

1. **Option A**: Use GitHub Copilot Workspace (cloud agent, different UI)
2. **Option B**: Break into smaller chunks (1-2 phases at a time)
3. **Option C**: Semi-autonomous (manual phase boundaries, autonomous tasks within)
4. **Option D**: Manual execution with AI assistance

---

## Contact & Support

**Questions**: Review this guide first
**Blockers**: Check Troubleshooting section
**Bugs in Prompt**: File issue in cryptoutil repo
**Codex Issues**: Contact Anthropic support

---

**Last Updated**: 2026-01-16
**Prompt Version**: v1.0
**Target Model**: Claude Opus 4.5
**Validated**: No (first execution)
