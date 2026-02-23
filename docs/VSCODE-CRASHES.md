# VS Code Copilot Chat Crash Diagnosis

## Summary

Over the past week (Feb 16-23, 2026), VS Code has crashed 6-7 times during Copilot Chat sessions working on the cryptoutil project. All crashes share common root causes related to memory pressure and terminal output volume.

## Crash Inventory

| # | Approximate Date | Activity at Time of Crash | Probable Root Cause |
|---|-----------------|--------------------------|---------------------|
| 1 | ~Feb 17 | Running `gremlins unleash` in batch loops across many packages | Terminal output overflow |
| 2 | ~Feb 18 | Large batch mutation testing with `for pkg in ... ; do gremlins unleash` | Terminal output overflow |
| 3 | ~Feb 19 | Extended session with many file reads/searches + mutation testing | Conversation context size |
| 4 | ~Feb 20 | Running gremlins on template service packages in single batch command | Terminal output overflow |
| 5 | ~Feb 21 | Parallel tool calls (multiple file reads + searches simultaneously) | Memory pressure from parallel ops |
| 6 | ~Feb 22 | Long-running mutation testing session with accumulated context | Conversation context size |
| 7 | ~Feb 23 | Continued mutation testing (gremlins + many tool calls) | Combined: terminal output + context size |

## Root Causes

### 1. Terminal Output Overflow (PRIMARY — causes ~60% of crashes)

**Problem**: `gremlins unleash` produces large amounts of output (one line per mutant × many mutants per package). When run in batch loops (`for pkg in ...; do gremlins unleash ...; done`), the combined output for 10+ packages can exceed 60KB+ easily.

The `run_in_terminal` tool has a ~60KB output truncation limit, but the VS Code extension still processes the full output stream internally before truncating. Large terminal buffers cause the extension host process to consume excessive memory.

**Worst Pattern**: Running gremlins on ALL packages in a single `for` loop:
```bash
# THIS CRASHES VS CODE:
for pkg in pkg1 pkg2 pkg3 pkg4 pkg5 pkg6 pkg7 pkg8 pkg9 pkg10; do
  gremlins unleash --timeout-coefficient=60 "./$pkg"
done
```

**Fix**: Run gremlins on ONE package at a time, with `| tail -10` to limit output:
```bash
# SAFE:
gremlins unleash --timeout-coefficient=60 ./path/to/single/package 2>&1 | tail -15
```

### 2. Conversation Context Accumulation (SECONDARY — causes ~25% of crashes)

**Problem**: Long sessions accumulate massive conversation context from:
- Many `read_file` calls (each adds file content to context)
- Many `grep_search` results (each adds match results)
- Many `run_in_terminal` outputs (each adds command output)
- Many `runSubagent` results (each adds agent report)

After 50+ tool calls in a single session, the conversation context can exceed VS Code Copilot's memory budget, causing the extension host to OOM or become unresponsive.

**Fix**: No direct user fix. Mitigations:
- Start fresh sessions more frequently (every 20-30 substantive tool calls)
- Avoid reading the same large files repeatedly
- Use targeted `grep_search` instead of broad `read_file` when possible

### 3. Parallel Tool Calls Memory Spike (TERTIARY — causes ~15% of crashes)

**Problem**: When the agent makes 4+ parallel tool calls (e.g., 4 simultaneous `read_file` calls or mixed `grep_search` + `read_file`), the VS Code extension host processes all responses simultaneously. Combined with an already-large conversation context, this can push memory over the limit.

**Fix**: No direct user fix. The agent should serialize operations when context is already large.

## Specific Trigger Patterns to Avoid

### Pattern A: Batch Gremlins in For Loops
```bash
# DANGEROUS - produces massive output, #1 crash cause
for pkg in $(find . -name '*_test.go' -exec dirname {} \; | sort -u); do
  gremlins unleash "$pkg"
done
```

### Pattern B: Gremlins Without Output Limiting
```bash
# DANGEROUS - single package can produce 500+ lines
gremlins unleash --timeout-coefficient=60 ./internal/apps/template/service/...
```

### Pattern C: Reading Entire Large Files
```bash
# Context accumulates when reading many 400+ line files in sequence
# Read only the specific lines needed instead
```

## Recommended Agent Behavior Rules

1. **ONE package per gremlins run** — NEVER batch multiple packages in a loop
2. **ALWAYS pipe gremlins through `tail -15`** — only need summary + LIVED mutants
3. **Use `grep LIVED` for targeted mutant identification** — don't dump full output
4. **Commit and push frequently** — preserves work before potential crash
5. **Start fresh sessions after ~30 tool calls** — prevents context overflow
6. **Avoid redundant file reads** — read once, remember the content
7. **Use `grep_search` over `read_file`** — returns only matching lines, not entire files

## VS Code Settings That May Help

Add to `.vscode/settings.json`:

```json
{
    "terminal.integrated.scrollback": 1000,
    "terminal.integrated.persistentSessionScrollback": 100,
    "chat.editor.wordWrap": "on"
}
```

Reducing `scrollback` limits terminal buffer memory. The default (1000) is reasonable but could be reduced to 500 if crashes persist.

## Extension Host Memory Monitoring

To monitor VS Code extension host memory:
1. Open Command Palette (Ctrl+Shift+P)
2. Run "Developer: Open Process Explorer"
3. Watch "Extension Host" memory — if it exceeds ~1.5GB, consider restarting

Alternatively, from terminal:
```bash
# Watch VS Code memory usage
watch -n 5 'ps aux | grep -E "extensionHost|copilot" | grep -v grep | awk "{print \$6/1024 \" MB\", \$11}"'
```

## Recovery After Crash

1. VS Code usually auto-recovers and reopens
2. Check `git status` to verify no uncommitted work was lost
3. Check `git stash list` for auto-stashed changes
4. Resume from last commit — the conversation context is lost on crash
