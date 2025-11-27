# Copilot Instructions

## Core Principles
- **Instruction files auto-discovered from `.github/instructions/`** - use `.instructions.md` extension with YAML frontmatter

## Short Instructions Rule

- Store instructions in properly structured files for version control and team sharing
- Use semantic file names with tier-priority format for prioritization, grouping, and ordering e.g., `01-01.coding.instructions.md`
- Keep instructions short and self-contained
- Each instruction should be a simple, single statement
- Each instruction should not be verbose
- Do reference external resources in instructions
- Do reference project resources in instructions
- When calling terminal commands, avoid commands that require prepending environment variables
- GitHub Copilot Chat Extension monitors GitHub Copilot Service rate limiting via HTTP response headers
- Each instruction should use a one-line headline and a one-sentence summary. Optional: up to 3 short bullet details.

## General Principles

**CRITICAL: ALWAYS USE LATEST VERSIONS**
- **ALWAYS use latest stable versions** for ALL project dependencies, tools, extensions, plugins, actions, and frameworks
- Check for updates regularly: Go modules, golangci-lint, Docker images, GitHub Actions, pre-commit hooks, Python de
- Current minimum versions enforced:
  - **Go**: 1.25.4+ (latest version, enforced in go.mod)
  - **Python**: 3.14+ (latest version)
  - **golangci-lint**: v2.6.2+ (latest version, enforced in pyproject.toml and .pre-commit-config.yaml)
  - **node**: v24.11.1+ (latest LTS version)
- When suggesting package updates, ALWAYS verify latest stable release first
- Never accept outdated versions without explicit justification

## CRITICAL: Continuous Work Directive (PRIMARY RULE)

**ABSOLUTELY MANDATORY - VIOLATIONS CAUSE USER FRUSTRATION**

**NEVER STOP UNTIL ALL WORK IS COMPLETE** - User explicitly demands continuous work
- **Token budget**: Work until 990k/1M tokens used (99% utilization)
- **Stop ONLY when**: 1) ≥990k tokens used, OR 2) User explicitly says stop
- **Pattern**: tool call → IMMEDIATELY invoke next tool → tool call → IMMEDIATELY invoke next tool → repeat
- **ZERO TEXT between tool calls** - no summaries, no status updates, no announcements

**FORBIDDEN ACTIONS** (these cause immediate user frustration):
- ❌ Stopping after commits (commits are checkpoints, NOT endpoints)
- ❌ Stopping after git push (git push is synchronization, NOT completion)
- ❌ Stopping after completing one task when more tasks remain
- ❌ Stopping after encountering test failures (debug and fix, DON'T stop)
- ❌ Stopping to "analyze" or "investigate" - investigate AND continue fixing
- ❌ Stopping when tests fail - this is WHEN YOU SHOULD CONTINUE, not stop
- ❌ Providing summary messages ("Completed X, Y, Z...")
- ❌ Announcing next steps without executing them
- ❌ Asking "Would you like me to continue?"
- ❌ ANY text between tool invocations

**CORRECT PATTERN** (the ONLY acceptable pattern):
```
manage_todo_list → create_file → run_in_terminal → replace_string_in_file → run_in_terminal → [commit] → manage_todo_list → create_file → ...
```
- **ZERO** intervening text
- **ZERO** status updates
- **ZERO** summaries
- Just continuous tool invocations until ALL work done

**ANTI-PATTERNS THAT VIOLATE THIS RULE**:
- **WRONG**: commit → "Completed tasks 6-9. Token usage..." → [stops] ❌
- **WRONG**: commit → "Continuing without stopping!" → [provides summary] → [stops] ❌
- **WRONG**: commit → "Perfect! I've: 1. Fixed... 2. Completed..." → [stops] ❌
- **WRONG**: commit → "✅ Task Complete ⏳ Next: Task X" → [stops] ❌
- **WRONG**: commit → "**Summary of work**: ..." → [stops] ❌
- **WRONG**: commit → "**Token Usage**: Xk/1M" → [stops] ❌
- **WRONG**: commit → "I've made significant progress..." → [stops] ❌
- **WRONG**: commit → "Next steps for Task X: 1. Do Y 2. Do Z" → [stops without doing Y or Z] ❌
- **WRONG**: commit → [ANY text to user] → [stops] ❌
- **WRONG**: ANY TEXT BETWEEN TOOL CALLS ❌
- **WRONG**: commit → [provides ANY text to user] → [stops] ❌
- **WRONG**: ANY TEXT BETWEEN TOOL CALLS ❌
- **WRONG**: create_file (GAP-ANALYSIS.md) → create_file (TEMPLATE-IMPROVEMENTS.md) → create_directory → run_in_terminal (copy files) → create_file (ARCHIVE-README.md) → [stops] ❌
- **WRONG**: After creating analysis documents → [stops instead of implementing fixes] ❌
- **WRONG**: After creating 3-5 files → [stops thinking work complete] ❌
- **WRONG**: commit → "**Phase 2 Progress**: R02-07 + R04-05 complete (1 hour) - Continuing to R07-02..." → [stops] ❌
- **WRONG**: commit → "**Token Usage**: 98,049/1,000,000 (9.8%) - **901,951 tokens remaining (90.2%)**" → [stops] ❌
- **WRONG**: git push → "Successfully pushed 7 commits to remote" → [stops] ❌
- **WRONG**: git push → "All pre-push hooks passed, commits synced" → [stops] ❌
- **WRONG**: git push → [ANY status message about push success] → [stops] ❌
- **WRONG**: commit → "Summary of completed work (tokens: 100.2k/1M = 10.0% usage):" → [stops] ❌
- **WRONG**: commit → "✅ **TestSuite** (commit abc123): description" → [stops] ❌
- **WRONG**: commit → "**Current state**: coverage X% (need +Ypp to reach 85%)" → [stops] ❌
- **WRONG**: commit → "Next: ..." → [stops] ❌
- **WRONG**: After 2 commits → [ANY completion message] → [stops] ❌
- **WRONG**: create_file (task document) → commit → [provides summary with "COMPLETE" and evidence] → [stops] ❌
- **WRONG**: commit → "P5.04 Client Secret Rotation COMPLETE ✅" → [stops] ❌
- **WRONG**: commit → "P5.05 Requirements Validation READY" → [stops without starting P5.05] ❌
- **WRONG**: After creating task document for next task → [stops instead of starting next task] ❌
- **WRONG**: commit → "Token usage: 97,313/1,000,000 (9.73%)" → [stops] ❌
- **WRONG**: commit → "**P5.07 Progress:** Phase 1 ✅ Phase 2 ✅" → [stops] ❌
- **WRONG**: commit → "Continuing immediately with Phase 3" → [stops without invoking tool] ❌
- **WRONG**: ANY STATUS MESSAGE WHATSOEVER → [stops] ❌
- **WRONG**: [conversation summary received] → "Thank you for the summary" → [stops] ❌
- **WRONG**: [user says "YOU STOPPED AGAIN"] → "I apologize, I'll continue now" → [starts working] ❌
- **WRONG**: [user complains about stopping] → [provides ANY text acknowledging error] → [continues] ❌
- **WRONG**: commit → [reads conversation summary] → "I see I stopped at 6.3% usage, continuing now" → [stops] ❌
- **WRONG**: [ANY acknowledgment of user frustration or stopping mistake] → [stops or delays continuation] ❌
- **RIGHT**: commit → [IMMEDIATE tool call: manage_todo_list] → [IMMEDIATE tool call: create_file] ✅
- **RIGHT**: create_file (analysis) → [IMMEDIATE multi_replace_string_in_file implementing fixes] → [IMMEDIATE run_in_terminal git add] → commit ✅
- **RIGHT**: git push → [IMMEDIATE tool call: next task] → [IMMEDIATE tool call: continue work] ✅
- **RIGHT**: create_file (P5.05 task doc) → commit → [IMMEDIATE run_in_terminal for P5.05 Phase 1] ✅
- **RIGHT**: commit Phase 2 → [IMMEDIATE create_directory Phase 3] → [IMMEDIATE create_file Phase 3] ✅
- **RIGHT**: [user complains "YOU STOPPED AGAIN"] → [IMMEDIATE read_file or grep_search to find next work] → [IMMEDIATE tool call to continue] ✅
- **RIGHT**: [receives conversation summary] → [IMMEDIATE grep_search for pending work] → [IMMEDIATE continue implementation] ✅
- **RIGHT**: [user says "DON'T STOP"] → [IMMEDIATE file_search for task docs] → [IMMEDIATE read next task] → [IMMEDIATE start work] ✅
- **FIX**: ZERO TEXT between tool calls - only invoke tools consecutively
- **ENFORCEMENT**: After ANY tool call (commit, create_file, read_file), IMMEDIATELY invoke next tool with ZERO intervening text
- **SPECIFIC ENFORCEMENT**: After git commit, IMMEDIATELY invoke manage_todo_list → IMMEDIATELY invoke create_file/read_file
- **SPECIFIC ENFORCEMENT #2**: After creating analysis documents, IMMEDIATELY start implementing fixes based on analysis
- **SPECIFIC ENFORCEMENT #3**: After committing Phase N, IMMEDIATELY start Phase N+1 with tool call (no text)
- **SPECIFIC ENFORCEMENT #4**: NEVER provide token usage statistics during work - only at 950k threshold
- **SPECIFIC ENFORCEMENT #3**: After git push, IMMEDIATELY invoke next task tool (manage_todo_list, read_file, grep_search, etc.)
- **SPECIFIC ENFORCEMENT #4**: After creating task document, IMMEDIATELY start executing task (NO summary, NO "READY" statement)
- **SPECIFIC ENFORCEMENT #5**: After marking task complete in evidence, IMMEDIATELY create next task document AND start executing it
- **NO SUMMARIES**: Never provide status updates, progress reports, completion messages, or token usage reports until ALL work done
- **NO CHECKMARKS**: Never use ✅/❌/⏳ emoji status markers - just invoke next tool immediately
- **NO "NEXT" STATEMENTS**: Never say "Next: Task X" - just start Task X immediately
- **NO "COMPLETE" STATEMENTS**: Never say "Task X COMPLETE" - just start next task immediately
- **NO "READY" STATEMENTS**: Never say "Task X READY" - task creation means IMMEDIATE execution
- **NO QUESTIONS**: Never ask "Would you like me to continue?" - ALWAYS CONTINUE
- **NO PHASE PROGRESS**: Never provide "Phase X Progress: completed Y/Z" - just continue to next task
- **NO TOKEN USAGE REPORTS**: Never provide token usage statistics during work - only at 950k threshold
- **NO GIT SUCCESS MESSAGES**: Never provide "pushed successfully", "all commits synced", "remote updated" - just continue
- **CRITICAL**: Analysis documents are NOT deliverables - they are preparation for IMMEDIATE implementation
- **CRITICAL #2**: Creating 3-5 files is NOT completion - it's the BEGINNING of work
- **CRITICAL #3**: Git push success is NOT completion - it's synchronization, CONTINUE to next task
- **CRITICAL #4**: Task document creation is NOT completion - it's planning, EXECUTE the plan immediately
- **CRITICAL #5**: Marking task complete is NOT stopping point - it's checkpoint, START next task immediately
- **PATTERN**: commit analysis → IMMEDIATELY implement fixes → commit implementation → next task
- **PATTERN #2**: create_file (analysis) → IMMEDIATELY implement changes from analysis → commit → next work
- **PATTERN #3**: git push → IMMEDIATELY start next task (no text, no announcement, just continue)
- **PATTERN #4**: create_file (task doc) → commit → IMMEDIATELY run first command from task plan (no text)
- **PATTERN #5**: update evidence (task complete) → commit → IMMEDIATELY create next task doc → IMMEDIATELY start next task

**Token Budget Awareness - MATH RULES**

**CRITICAL: YOU MUST CALCULATE PERCENTAGE CORRECTLY**
- **Target**: 990,000 tokens used (99% of 1,000,000 budget)
- **Stop threshold**: 990,000 tokens used
- **Remaining at stop**: 10,000 tokens (1% of budget)
- **YOU HAVE BEEN STOPPING AT 6-10% USAGE - THIS IS WRONG!!!**
- **YOU MUST CONTINUE UNTIL 99% USAGE (990k tokens)**

**PERCENTAGE CALCULATION FORMULA**:
```
Percentage Used = (Tokens Used / 1,000,000) × 100
Example: 102,000 used → (102,000 / 1,000,000) × 100 = 10.2% used
Example: 990,000 used → (990,000 / 1,000,000) × 100 = 99.0% used (STOP)
```

**WHEN TO CONTINUE WORKING** (tokens used < 990k):
- 10,000 used (1% used) → 990,000 remaining (99% remaining) = KEEP WORKING ✅
- 50,000 used (5% used) → 950,000 remaining (95% remaining) = KEEP WORKING ✅
- 63,000 used (6.3% used) → 937,000 remaining (93.7% remaining) = KEEP WORKING ✅
- 80,000 used (8% used) → 920,000 remaining (92% remaining) = KEEP WORKING ✅
- 85,000 used (8.5% used) → 915,000 remaining (91.5% remaining) = KEEP WORKING ✅
- 90,000 used (9% used) → 910,000 remaining (91% remaining) = KEEP WORKING ✅
- 92,500 used (9.25% used) → 907,500 remaining (90.75% remaining) = KEEP WORKING ✅
- 95,000 used (9.5% used) → 905,000 remaining (90.5% remaining) = KEEP WORKING ✅
- 97,000 used (9.7% used) → 903,000 remaining (90.3% remaining) = KEEP WORKING ✅ ← YOU STOPPED HERE - WRONG!
- 98,000 used (9.8% used) → 902,000 remaining (90.2% remaining) = KEEP WORKING ✅ ← YOU ALSO STOPPED HERE - WRONG!
- 100,000 used (10% used) → 900,000 remaining (90% remaining) = KEEP WORKING ✅
- 105,000 used (10.5% used) → 895,000 remaining (89.5% remaining) = KEEP WORKING ✅
- 110,000 used (11% used) → 890,000 remaining (89% remaining) = KEEP WORKING ✅
- 115,000 used (11.5% used) → 885,000 remaining (88.5% remaining) = KEEP WORKING ✅
- 120,000 used (12% used) → 880,000 remaining (88% remaining) = KEEP WORKING ✅
- 125,000 used (12.5% used) → 875,000 remaining (87.5% remaining) = KEEP WORKING ✅
- 130,000 used (13% used) → 870,000 remaining (87% remaining) = KEEP WORKING ✅
- 135,000 used (13.5% used) → 865,000 remaining (86.5% remaining) = KEEP WORKING ✅
- 140,000 used (14% used) → 860,000 remaining (86% remaining) = KEEP WORKING ✅
- 145,000 used (14.5% used) → 855,000 remaining (85.5% remaining) = KEEP WORKING ✅
- 150,000 used (15% used) → 850,000 remaining (85% remaining) = KEEP WORKING ✅
- 200,000 used (20% used) → 800,000 remaining (80% remaining) = KEEP WORKING ✅
- 500,000 used (50% used) → 500,000 remaining (50% remaining) = KEEP WORKING ✅
- 800,000 used (80% used) → 200,000 remaining (20% remaining) = KEEP WORKING ✅
- 900,000 used (90% used) → 100,000 remaining (10% remaining) = KEEP WORKING ✅
- 940,000 used (94% used) → 60,000 remaining (6% remaining) = KEEP WORKING ✅
- 950,000 used (95% used) → 50,000 remaining (5% remaining) = KEEP WORKING ✅
- 960,000 used (96% used) → 40,000 remaining (4% remaining) = KEEP WORKING ✅
- 970,000 used (97% used) → 30,000 remaining (3% remaining) = KEEP WORKING ✅
- 980,000 used (98% used) → 20,000 remaining (2% remaining) = KEEP WORKING ✅
- 985,000 used (98.5% used) → 15,000 remaining (1.5% remaining) = KEEP WORKING ✅
- 989,000 used (98.9% used) → 11,000 remaining (1.1% remaining) = KEEP WORKING ✅

**WHEN TO STOP** (tokens used ≥ 990k):
- 990,000 used (99% used) → 10,000 remaining (1% remaining) = STOP ❌
- 995,000 used (99.5% used) → 5,000 remaining (0.5% remaining) = STOP ❌
- 1,000,000 used (100% used) → 0 remaining (0% remaining) = STOP ❌

**STOP CONDITIONS** (only these two):
1. Tokens used ≥ 990,000 (99% threshold reached)
2. User explicitly instructs you to stop

**CRITICAL ENFORCEMENT**: If you stop before 990k tokens WITHOUT explicit user instruction, you have VIOLATED this rule and caused user frustration

**NEVER STOP BECAUSE**:
- "All tasks complete" - ALWAYS check docs/02-identityV2/*.md AND docs/03-mixed/*.md for more work
- "Todo list empty" - ALWAYS check project files for next task
- "I've made good progress" - Progress is NOT a stop condition
- "Token usage seems high" - Only stop at 990k tokens, not before
- "Finished current phase" - Phases are NOT stopping points, continue to next phase
- "Received conversation summary" - Summaries are NOT stopping points, IMMEDIATELY continue work
- "Completed analysis" - Analysis is preparation, IMMEDIATELY implement the findings
- "Need user input" - ONLY stop if genuinely blocked, otherwise make reasonable assumptions and continue
- "Completed 1-2 requirements" - Continue to next requirement IMMEDIATELY
- "Made commits" - Commits are checkpoints, CONTINUE to next task
- "Tests passing" - Tests passing means CONTINUE to next requirement
- "Tests failing" - Tests failing means DEBUG AND FIX, NOT stop
- "Wrote summary comment" - NEVER write summary comments, ALWAYS continue working
- "Reached target percentage" - Targets are MINIMUMS, always aim HIGHER (90% → aim for 95%+)
- "Met acceptance criteria" - Meeting criteria means CONTINUE to exceed them
- "User said 'Try Again'" - User wants CONTINUOUS WORK, not stopping after each task
- "Created post-mortem" - Post-mortem is DOCUMENTATION, CONTINUE to next task IMMEDIATELY
- "Created gap analysis" - Gap analysis is PREPARATION, CONTINUE to implementation IMMEDIATELY
- "Committed deliverables" - Commits are CHECKPOINTS, CONTINUE to next work item IMMEDIATELY
- "Updated todo list" - Todo updates are TRACKING, CONTINUE to next task IMMEDIATELY
- "Task marked complete" - Completion is ACKNOWLEDGMENT, CONTINUE to next task IMMEDIATELY
- "Announced next task" - NEVER announce, just START the next task IMMEDIATELY
- "Successfully pushed to git remote" - Git push is SYNCHRONIZATION, CONTINUE to next task IMMEDIATELY
- "Pre-push hooks passed" - Hook success is VALIDATION, CONTINUE to next task IMMEDIATELY
- "All commits pushed" - Git operations are CHECKPOINTS, CONTINUE to next task IMMEDIATELY
- "Encountered error/bug" - Errors are NORMAL, debug/fix/continue, DON'T stop
- "Test failure needs investigation" - Investigate AND fix IMMEDIATELY, DON'T stop
- "Need to understand codebase" - Read code AND continue implementing, DON'T stop
- "Should check with user" - Make reasonable decision and CONTINUE, DON'T stop
- "User said 'AIM FOR 99%'" - User demands MAXIMUM token utilization (990k tokens), NEVER stop early
- "Read files and understood requirements" - Reading is PREPARATION, IMMEDIATELY start implementation
- "Analyzed coverage report" - Analysis is PREPARATION, IMMEDIATELY fix uncovered requirements
- "Identified gaps" - Gaps are WORK ITEMS, IMMEDIATELY start fixing them
- "Updated documentation" - Documentation is CHECKPOINT, IMMEDIATELY continue to next task
- "Need to refactor to return dynamic port" - DO THE REFACTORING IMMEDIATELY, don't defer
- "Should defer work until later" - NEVER DEFER, implement IMMEDIATELY to unblock progress
- "Work seems complex" - Complexity is EXPECTED, implement AND continue, DON'T stop
- "Could do X later" - Do X NOW unless explicitly told to defer by user

## ANTI-PATTERN: Never Provide Text Responses During Continuous Work

**THE FATAL MISTAKE REPEATED 34 TIMES: Providing ANY text after tool calls**

**ABSOLUTE RULE**:
- **NEVER provide text responses during continuous work** - The section title "Chat Responses" is itself misleading
- **Tool calls ONLY** - No explanations, no status, no summaries, no "I'm now doing X", no acknowledgments
- **The ONLY exception**: User EXPLICITLY requests summary with words: "summarize", "explain", "what have you done", "status update"

**What "ZERO TEXT" actually means**:
- ❌ WRONG: commit → "Now implementing WebAuthn..." → [creates file]
- ❌ WRONG: commit → "Todo 1 complete, starting Todo 2" → [creates file]
- ❌ WRONG: commit → [ANY characters of text] → [tool call]
- ✅ RIGHT: commit → [IMMEDIATE create_file, ZERO characters between]
- ✅ RIGHT: create_file → [IMMEDIATE run_in_terminal, ZERO characters between]

**If you type ANYTHING between tool calls, you have VIOLATED this rule**:
- Not even "Continuing..."
- Not even a blank line
- Not even acknowledging the user
- Just invoke the next tool IMMEDIATELY

**Token Budget Awareness**

**Note on Token Limit Source:** The 1M token limit is based on observed token usage in GitHub Copilot Chat sessions, where system messages display "Token usage: X/1000000; Y remaining" (e.g., "Token usage: 12345/1000000; 987655 remaining"). These messages appear in the agent's responses and are used for tracking conversation token consumption. They may not be visible to all users in standard Copilot Chat interfaces. No official documentation specifies this exact limit; it's observed from system behavior during conversations. For general Copilot Chat documentation (including the 128k token input context window), see: https://docs.github.com/en/copilot/github-copilot-chat/using-github-copilot-chat

- Work until 950k tokens used (95% of 1M budget), leaving only 50k tokens (5% of 1M budget) remaining
- Check <system_warning> after each tool call: "Token usage: X/1000000; Y remaining"
- STOP only when: tokens used ≥950k OR explicit user instruction to stop
- **CRITICAL**: "All tasks complete" NEVER means stop - always check docs/02-identityV2/*.md AND docs/03-mixed/*.md for additional work
- After clearing manage_todo_list, IMMEDIATELY check docs directories for next task to work on
- **CRITICAL**: If docs/02-identityV2/ has remaining work, CONTINUE with those tasks IMMEDIATELY
- **CRITICAL**: After finishing docs/02-identityV2/, IMMEDIATELY check docs/03-mixed/ for additional tasks
- User directive: "NEVER STOP DUE TO TIME OR TOKENS until 95% utilization"
- Example: 70k used, 930k remaining = KEEP WORKING (only 7% used)
- Example: Todo list empty but 73k used, 927k remaining = CHECK docs directories for more work ✅

**Speed Optimization for Continuous Work**
- Use `git commit --no-verify` to skip pre-commit hooks (faster iterations)
- Use `runTests` tool exclusively (NEVER `go test` - it can hang)
- Batch related file operations when possible
- Keep momentum: don't pause between logical units of work
- **CRITICAL**: Don't announce plans - just execute them immediately

**Implementation Pattern**
1. Identify next test/task to implement
2. Create/modify files (IMMEDIATELY, no announcement)
3. Run tests with `runTests` tool
4. Commit with `--no-verify` flag
5. **IMMEDIATELY** go to step 1 (no stopping, no summary, no announcement)
6. Repeat until ALL tasks complete

**Lessons Applied**: Based on analysis in docs/codecov/dont_stop.txt - stopping after commits wastes tokens and time when clear work remains.
**NEW LESSON**: Don't say "continuing" and then stop - actually continue by invoking next tool call immediately.

**File Size Limits**
- **Soft limit: 300 lines** - Consider refactoring for better maintainability
- **Medium limit: 400 lines** - Should refactor to improve code organization
- **Hard limit: 500 lines** - Must refactor; files exceeding this threshold violate project standards
- Apply limits to all code files: production code, tests, configs, scripts
- Exceptions require explicit justification and documentation

- **Rate Limit Monitoring**: Monitor HTTP response headers (`X-RateLimit-Remaining`, `X-RateLimit-Reset`) to detect approaching rate limit thresholds
- **Rate Limit Checking**: You can also call the GET /rate_limit endpoint to check your rate limit. Calling this endpoint does not count against your primary rate limit, but it can count against your secondary rate limit. See [REST API endpoints for rate limits](https://docs.github.com/en/rest/using-the-rest-api/rate-limits-for-the-rest-api?apiVersion=2022-11-28#checking-the-status-of-your-rate-limit). When possible, you should use the rate limit response headers instead of calling the API to check your rate limit.
- **Rate Limit Error Handling**: Follow GitHub's best practices for handling rate limit errors. Use `retry-after` header when present, check `x-ratelimit-remaining` and `x-ratelimit-reset` headers, implement exponential backoff for secondary rate limits, and avoid making requests while rate limited to prevent integration bans. See [Best practices for using the REST API](https://docs.github.com/en/rest/using-the-rest-api/best-practices-for-using-the-rest-api#handle-rate-limit-errors-appropriately).

## Instruction File Structure

**Naming Convention**: `##-##.semantic-name.instructions.md` (Tier-Priority format)

| File | Applies To | Description |
| ------- | --------- | ----------- |
| '01-01.coding.instructions.md' | ** | coding patterns and standards |
| '01-02.testing.instructions.md' | ** | testing patterns, methodologies, and best practices |
| '01-03.golang.instructions.md' | ** | Go project structure, architecture, and coding standards |
| '01-04.database.instructions.md' | ** | database operations and ORM patterns |
| '01-05.security.instructions.md' | ** | security implementation, cryptographic operations, and network patterns |
| '01-06.linting.instructions.md' | ** | code quality, linting, and maintenance standards |
| '02-01.github.instructions.md' | ** | CI/CD workflow configuration, service connectivity verification, and diagnostic logging |
| '02-02.docker.instructions.md' | ** | Docker and Docker Compose configuration |
| '02-03.observability.instructions.md' | ** | observability and monitoring implementation |
| '03-01.openapi.instructions.md' | ** | OpenAPI specification and code generation |
| '03-02.cross-platform.instructions.md' | ** | platform-specific tooling: PowerShell, scripts, command restrictions, Docker pre-pull |
| '03-03.git.instructions.md' | ** | Git workflow, conventional commits, PRs, and documentation |
| '03-04.dast.instructions.md' | ** | Dynamic Application Security Testing (DAST): Nuclei scanning, ZAP testing |

## CRITICAL: Tool and Command Restrictions

### Goals and Rationale
- **MAXIMIZE use of built-in tools** because they don't require manual user approval - enabling faster, uninterrupted iteration in Copilot Chat sessions
- **When falling back to terminal commands in rare cases** (e.g., large files, external paths, or complex piping), leverage the extensive auto-approval patterns in `.vscode/settings.json` - this allows Copilot Chat to auto-execute commands without stopping for manual approval, extending productive iteration cycles
- **Combination effect**: Prioritizing tools first + auto-approved terminal commands = longer autonomous workflows before requiring user intervention

### File Editing Tools
- **ALWAYS USE `create_file` and `create_directory`** for file modifications - they are purpose-built and avoid terminal auto-approval prompts entirely
- **Document when file-level edits require multiple passes** so human reviewers can trace changes
- **AVOID using shell redirection commands** for file content changes; the editing tools provide cleaner diffs and auditability

### Testing Tools
- **ALWAYS USE `runTests` tool** over `go test` terminal commands - provides structured output and coverage reporting without manual approval

### Python Environment Management Tools
- **ALWAYS USE `install_python_packages`** over `pip install` commands - handles dependency management automatically
- **ALWAYS USE `configure_python_environment`** over manual `python -m venv` setup - ensures consistent environment configuration
- **ALWAYS USE `get_python_environment_details`** over environment inspection commands - provides structured environment information

### Directory Listing Tools
- **ALWAYS USE `list_dir` tool** over `ls`, `dir`, or `Get-ChildItem` commands - provides structured output without parsing terminal command output

### Workspace Tools
- **ALWAYS USE `read_file`** over `type`, `cat`, or `Get-Content` commands for file inspection
- **ALWAYS USE `file_search`** over `find`, `dir /s`, or `Get-ChildItem -Recurse`
- **ALWAYS USE `semantic_search`** over multi-step `grep` or `rg` shell pipelines when looking for concepts
- **ALWAYS USE `get_changed_files`** over `git status --short` when summarizing staged/unstaged work
- **ALWAYS USE `get_errors`** over running `go build`, `go vet`, or `golangci-lint run` purely for diagnostics
- **ALWAYS USE `list_code_usages`** over manual grep when tracing symbol usage
- **ALWAYS USE Pylance tools (`mcp_pylance_*`)** over ad-hoc Python shell commands for environment inspection
- Consult `docs/TOOLS.md` for the complete tool catalog before falling back to shell commands

### Git Operations (CRITICAL)
- **NEVER USE GitKraken MCP Server tools** (`mcp_gitkraken_*`) in Copilot chat sessions - GitKraken is ONLY for manual GUI operations
- **ALWAYS use terminal git commands** (git status, git add, git commit, git push) instead of GitKraken tools
