# Agent Tool Matrix

Reference table for reviewing which tools are enabled per agent. Edit the cells to plan changes, then update the corresponding `.github/agents/*.agent.md` files.

**Agent files**: each column maps to `.github/agents/<name>.agent.md`
**Column keys**: `fix-wf` = `fix-workflows`, `impl-exec` = `implementation-execution`, `impl-plan` = `implementation-planning`

## Tool Assignments

| Tool | beast-mode | fix-wf | impl-exec | impl-plan | Description |
|------|:----------:|:------:|:---------:|:---------:|-------------|
| **agent/** | | | | | |
| agent/runSubagent | [x] | [x] | [x] | [x] | Run a task in an isolated subagent context; improves context management of the main agent thread |
| agent/searchSubagent | [ ] | [ ] | [ ] | [ ] | Search workspace codebase in an isolated subagent context; alternative to agent/runSubagent for pure search tasks |
| agent/switchAgent | [ ] | [ ] | [ ] | [ ] | Switch to a different agent by name; useful for declarative agent-to-agent handoffs |
| **browser/** | | | | | |
| browser | [ ] | [ ] | [ ] | [ ] | *(Experimental)* Interact with the integrated browser: navigate, read page, take screenshots, click, type. Requires `workbench.browser.enableChatTools` |
| **edit/** | | | | | |
| edit/applyPatch | [x] | [x] | [x] | [x] | Apply a patch/diff to files in the workspace |
| edit/createDirectory | [x] | [x] | [x] | [x] | Create a new directory in the workspace |
| edit/createFile | [x] | [x] | [x] | [x] | Create a new file in the workspace |
| edit/editFiles | [x] | [x] | [x] | [x] | Apply edits to files in the workspace |
| edit/editNotebook | [ ] | [ ] | [ ] | [ ] | Make edits to a Jupyter notebook |
| edit/insertEdit | [x] | [x] | [x] | [x] | Insert text at a specific position in a file; more granular than edit/editFiles |
| edit/multiReplaceString | [x] | [x] | [x] | [x] | Apply multiple find-and-replace string substitutions across files in one call *(u)* |
| edit/rename | [x] | [x] | [x] | [x] | Rename or move a file or directory in the workspace *(u)* |
| edit/replaceString | [x] | [x] | [x] | [x] | Find and replace a string occurrence in a specific file *(u)* |
| **execute/** | | | | | |
| execute/awaitTerminal | [x] | [x] | [x] | [x] | Wait for a running terminal command to finish and return its output *(u)* |
| execute/createAndRunTask | [x] | [x] | [x] | [x] | Create and run a new VS Code task in the workspace |
| execute/getTerminalOutput | [x] | [x] | [x] | [x] | Get the output from a terminal command running in the workspace |
| execute/killTerminal | [x] | [x] | [x] | [x] | Terminate a running terminal session or process *(u)* |
| execute/runInTerminal | [x] | [x] | [x] | [x] | Run a shell command in the integrated terminal |
| execute/runNotebookCell | [ ] | [ ] | [ ] | [ ] | Execute a notebook cell |
| execute/runTests | [x] | [x] | [x] | [x] | Invoke the VS Code test runner for a test file or suite *(u)* |
| execute/testFailure | [x] | [x] | [x] | [x] | Get unit test failure information; useful when running and diagnosing tests |
| **newWorkspace** | | | | | |
| newWorkspace | [ ] | [ ] | [ ] | [ ] | Create a new VS Code workspace |
| **read/** | | | | | |
| read/getNotebookSummary | [ ] | [ ] | [ ] | [ ] | Get the list of notebook cells and their details |
| read/problems | [x] | [x] | [x] | [x] | Add workspace issues and problems from the Problems panel as context; useful while fixing code or debugging |
| read/readFile | [x] | [x] | [x] | [x] | Read the content of a file in the workspace |
| read/readNotebookCellOutput | [ ] | [ ] | [ ] | [ ] | Read the output from a notebook cell execution |
| read/terminalLastCommand | [x] | [x] | [x] | [x] | Get the last run terminal command and its output |
| read/terminalSelection | [x] | [x] | [x] | [x] | Get the text currently selected in the integrated terminal |
| read/viewImage | [x] | [x] | [x] | [x] | Display an image file inline in the chat *(u)* |
| **search/** | | | | | |
| search/codebase | [x] | [x] | [x] | [x] | Perform a semantic code search across the workspace to find relevant context |
| search/changes | [x] | [x] | [x] | [x] | List source control changes (git diff / SCM history) |
| search/fileSearch | [x] | [x] | [x] | [x] | Search for files in the workspace using glob patterns; returns file paths |
| search/findTestFiles | [x] | [x] | [x] | [x] | Find test files associated with a given source file using VS Code test discovery |
| search/listDirectory | [x] | [x] | [x] | [x] | List all files in a given directory in the workspace |
| search/symbols | [x] | [x] | [x] | [x] | Search workspace symbols (functions, types, variables, constants) by name |
| search/textSearch | [x] | [x] | [x] | [x] | Find literal or regex text matches across files in the workspace |
| search/usages | [x] | [x] | [x] | [x] | Combination of Find All References, Find Implementation, and Go to Definition |
| **selection** | | | | | |
| selection | [x] | [x] | [x] | [x] | Get the current editor selection (only available when text is selected) |
| **todos** | | | | | |
| todos | [x] | [x] | [x] | [x] | Track implementation and progress of a chat request with a todo list |
| **vscode/** | | | | | |
| vscode/askQuestions | [ ] | [ ] | [ ] | [ ] | Ask the user clarifying questions via the interactive questions carousel |
| vscode/extensions | [x] | [x] | [x] | [x] | Search for and ask about VS Code extensions in the Marketplace |
| vscode/getProjectSetupInfo | [ ] | [ ] | [ ] | [ ] | Provide instructions and configuration for scaffolding different project types |
| vscode/installExtension | [x] | [x] | [x] | [x] | Install a VS Code extension from the Marketplace |
| vscode/listCodeUsages | [ ] | [ ] | [ ] | [ ] | List all usages of a symbol using VS Code's language intelligence (Find All References / Go to Definition) *(u)* ΓÇö **DISABLED**: redundant with `search/usages` which covers the same functionality |
| vscode/memory | [ ] | [ ] | [ ] | [ ] | Read and write persistent agent memory across chat sessions *(u)* |
| vscode/renameSymbol | [x] | [x] | [x] | [x] | Rename a symbol across the workspace using VS Code's language intelligence *(u)* |
| vscode/runCommand | [ ] | [ ] | [ ] | [ ] | Execute a VS Code command by ID |
| vscode/VSCodeAPI | [ ] | [ ] | [ ] | [ ] | Ask about VS Code functionality and extension development APIs |
| **web/** | | | | | |
| web/fetch | [x] | [x] | [x] | [x] | Fetch and return the content from a given URL |
| web/githubRepo | [x] | [x] | [x] | [x] | Access GitHub repository file content and metadata *(u)* |
| web/searchResults | [x] | [x] | [x] | [x] | Execute a web search and return the results |

## Extension Tool Assignments

Tools contributed by installed VS Code extensions. All currently `[ ]` (this is a Go project ΓÇö Java/Python/Mermaid tools not needed).

### Debugger for Java (`vscjava.vscode-java-debug`)

| Tool | beast-mode | fix-wf | impl-exec | impl-plan | Description |
|------|:----------:|:------:|:---------:|:---------:|-------------|
| debugJavaApplication | [ ] | [ ] | [ ] | [ ] | Launch or attach to a Java application in debug mode with automatic compilation and classpath resolution |
| debugStepOperation | [ ] | [ ] | [ ] | [ ] | Control program execution flow: stepIn (enter method calls), stepOut (exit current method), stepOver, continue, or pause |
| evaluateDebugExpression | [ ] | [ ] | [ ] | [ ] | Evaluate a Java expression in a specific thread's debug context |
| getDebugSessionInfo | [ ] | [ ] | [ ] | [ ] | Get information about the currently active Java debug session |
| getDebugStackTrace | [ ] | [ ] | [ ] | [ ] | Retrieve the call stack showing all method calls leading to the current execution point |
| getDebugThreads | [ ] | [ ] | [ ] | [ ] | List all threads in the debugged Java application with their IDs, names, and states |
| getDebugVariables | [ ] | [ ] | [ ] | [ ] | Inspect variables in a specific thread's stack frame: local variables, method parameters, and fields |
| removeJavaBreakpoints | [ ] | [ ] | [ ] | [ ] | Remove breakpoints: specific breakpoint by file and line, all breakpoints in a file, or all breakpoints |
| setJavaBreakpoint | [ ] | [ ] | [ ] | [ ] | Set a breakpoint at a specific line in Java source code to pause execution and enable debugging |
| stopDebugSession | [ ] | [ ] | [ ] | [ ] | Stop the active Java debug session when investigation is complete |

### Mermaid Chat Features (`vscode.mermaid-chat-features`)

| Tool | beast-mode | fix-wf | impl-exec | impl-plan | Description |
|------|:----------:|:------:|:---------:|:---------:|-------------|
| vscode.mermaid-chat-features/renderMermaidDiagram | [x] | [x] | [x] | [x] | Render a Mermaid diagram inline in chat |

### Python (`ms-python.python`)

| Tool | beast-mode | fix-wf | impl-exec | impl-plan | Description |
|------|:----------:|:------:|:---------:|:---------:|-------------|
| configurePythonEnvironment | [x] | [x] | [x] | [x] | Configure the Python environment for the workspace |
| create_virtual_environment | [x] | [x] | [x] | [x] | Create a new Python virtual environment |
| getPythonEnvironmentInfo | [x] | [x] | [x] | [x] | Get details about the currently selected Python environment |
| getPythonExecutableCommand | [x] | [x] | [x] | [x] | Get the Python executable path and command for running scripts |
| installPythonPackage | [x] | [x] | [x] | [x] | Install one or more Python packages using pip |
| selectEnvironment | [x] | [x] | [x] | [x] | Select a Python environment for the workspace |

## Notes

- **Updating agents**: After editing this table, reflect changes in the corresponding `tools:` list in each `.github/agents/<name>.agent.md` file.
- **Undocumented tools *(u)***: Not listed in the [VS Code agent tools docs](https://code.visualstudio.com/docs/copilot/agents/agent-tools) but observed as functional in agents.
- **Notebook tools**: `edit/editNotebook`, `execute/runNotebookCell`, `read/getNotebookSummary`, `read/readNotebookCellOutput` are only useful for agents that work with Jupyter notebooks.
- **Extension tools**: The tool ID to put in an agent's `tools:` list is the `toolReferenceName` from the extension's `package.json` (camelCase, as shown in the tools picker). For tools with no `toolReferenceName`, use the `name` field (snake_case). For `github.copilot-chat` extension tools, use `category/toolReferenceName` (infer category from the tool picker groupings: agent, browser, edit, execute, read, search, todo, vscode, web).
- **MCP tools**: Tools from installed MCP servers appear in the tool picker alongside built-in tools. Add rows here for any that should be selectively enabled per agent. MCP config files: `%APPDATA%\Code\User\mcp.json` (Windows) or `.vscode/mcp.json` (workspace-level).

## Tool Discovery Commands

To rediscover all available tools when VS Code or extensions are updated:

```powershell
# Scan all installed extensions for language model tools
Get-ChildItem ~/.vscode/extensions -Directory | ForEach-Object {
    $pkg = Join-Path $_.FullName "package.json"
    if (Test-Path $pkg) {
        $json = Get-Content $pkg -Raw | ConvertFrom-Json -ErrorAction SilentlyContinue
        if ($json.contributes.languageModelTools) {
            Write-Host "=== $($_.Name) ==="
            $json.contributes.languageModelTools | Select-Object name, toolReferenceName, description | Format-Table -AutoSize
        }
    }
}
```

See `docs/ARCHITECTURE.md` Section 2.1.6 for the complete tool discovery methodology (4 sources: documented built-ins, undocumented built-ins, extension tools, MCP server tools).

## MCP Server Configuration

MCP servers are configured in `.vscode/mcp.json` (workspace-level, shared via git). Tools from running MCP servers appear in the Copilot tool picker automatically. To force-include specific MCP tools in an agent's `tools:` frontmatter, list them by their exact tool name as exposed by the server.

Currently configured servers (see `.vscode/mcp.json`):

| Server | Type | Tools Exposed | Purpose |
|--------|------|---------------|---------|
| `github` | HTTP (remote) | issues, PRs, code search, notifications | Richer GitHub access than built-in `web/githubRepo`; auth via existing Copilot token |
| `playwright` | stdio (local) | browser navigation, click, type, screenshot | SPA (identity-spa) testing, DAST validation, /browser/** path testing |

**Adding MCP tools to agent frontmatter**: Once an MCP server is running, list its tool names (e.g. `list_issues`, `browser_navigate`) in the agent `tools:` array to force-include them. Run the MCP server once to discover available tool names from the tool picker.

**Additional servers to consider** (not yet configured):

| Server | Install |
|--------|---------|
| `grafana/mcp-grafana` | `mcp-grafana --url http://localhost:3000` |
| `@modelcontextprotocol/server-postgres` | `npx -y @modelcontextprotocol/server-postgres <DSN>` |
