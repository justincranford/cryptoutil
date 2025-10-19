---
description: "Instructions for terminal command auto-approval pattern management"
applyTo: "**"
---
# Terminal Auto-Approve Pattern Management

## Command Pattern Checking Workflow

When executing terminal commands through Copilot:

1. **Check Pattern Match**: Before executing any command, verify if it matches existing `chat.tools.terminal.autoApprove` patterns in `.vscode/settings.json`

2. **Track Unmatched Commands**: Maintain a list of commands that don't match any existing patterns during the current session

3. **End-of-Session Review**: After completing command execution tasks, if any commands were unmatched, ask the user if they would like to add new auto-approve patterns

4. **Pattern Recommendations**: For each unmatched command, provide specific recommendations:
   - **Auto-Enable (true)**: Safe, informational, or build commands that are commonly used in development workflows
   - **Auto-Disable (false)**: Destructive, potentially dangerous, or system-altering commands that require manual approval

## Recommendation Guidelines

### Auto-Enable Candidates
- Read-only operations (status, list, inspect, logs, history)
- Build and test commands (build, test, format, lint)
- Safe informational commands (version, info, df)
- Development workflow commands (fetch, status, diff)

### Auto-Disable Candidates
- Destructive operations (rm, delete, prune, reset, kill)
- Network operations (push, pull from remotes)
- System modifications (install, update, edit configurations)
- File system changes (create, update, delete files/directories)
- Container execution (exec, run interactive containers)

## Pattern Format
When suggesting new patterns, use the established regex format:
- `"/^command (subcommand1|subcommand2)/": true|false`
- Group related subcommands with alternation `(cmd1|cmd2|cmd3)`
- Use `^` for start anchor and appropriate word boundaries
- Include comments explaining the security rationale</content>
<parameter name="filePath">c:\Dev\Projects\cryptoutil\.github\instructions\terminal-auto-approve.instructions.md
