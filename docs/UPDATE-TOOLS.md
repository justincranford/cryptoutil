# Agent Tool Matrix

Reference table for reviewing which tools are enabled per agent. Edit the cells to plan changes, then update the corresponding `.github/agents/*.agent.md` files.

**Agent files**: each column maps to `.github/agents/<name>.agent.md`
**Column keys**: `fix-wf` = `fix-workflows`, `impl-exec` = `implementation-execution`, `impl-plan` = `implementation-planning`

## Tool Assignments

| Tool | beast-mode | doc-sync | fix-wf | impl-exec | impl-plan | Description |
|------|:----------:|:--------:|:------:|:---------:|:---------:|-------------|
| **agent/** | | | | | | |
| agent/runSubagent | [x] | [x] | [x] | [x] | [x] | Run a task in an isolated subagent context; improves context management of the main agent thread |
| **browser/** | | | | | | |
| browser | [ ] | [ ] | [ ] | [ ] | [ ] | *(Experimental)* Interact with the integrated browser: navigate, read page, take screenshots, click, type. Requires `workbench.browser.enableChatTools` |
| **edit/** | | | | | | |
| edit/createDirectory | [x] | [x] | [x] | [x] | [x] | Create a new directory in the workspace |
| edit/createFile | [x] | [x] | [x] | [x] | [x] | Create a new file in the workspace |
| edit/editFiles | [x] | [x] | [x] | [x] | [x] | Apply edits to files in the workspace |
| edit/editNotebook | [ ] | [ ] | [ ] | [ ] | [ ] | Make edits to a Jupyter notebook |
| edit/rename | [x] | [x] | [x] | [x] | [x] | Rename or move a file or directory in the workspace *(u)* |
| **execute/** | | | | | | |
| execute/awaitTerminal | [x] | [x] | [x] | [x] | [x] | Wait for a running terminal command to finish and return its output *(u)* |
| execute/createAndRunTask | [x] | [x] | [x] | [x] | [x] | Create and run a new VS Code task in the workspace |
| execute/getTerminalOutput | [x] | [x] | [x] | [x] | [x] | Get the output from a terminal command running in the workspace |
| execute/killTerminal | [x] | [x] | [x] | [x] | [x] | Terminate a running terminal session or process *(u)* |
| execute/runInTerminal | [x] | [x] | [x] | [x] | [x] | Run a shell command in the integrated terminal |
| execute/runNotebookCell | [ ] | [ ] | [ ] | [ ] | [ ] | Execute a notebook cell |
| execute/runTests | [x] | [x] | [x] | [x] | [x] | Invoke the VS Code test runner for a test file or suite *(u)* |
| execute/testFailure | [x] | [x] | [x] | [x] | [x] | Get unit test failure information; useful when running and diagnosing tests |
| **newWorkspace** | | | | | | |
| newWorkspace | [ ] | [ ] | [ ] | [ ] | [ ] | Create a new VS Code workspace |
| **read/** | | | | | | |
| read/getNotebookSummary | [ ] | [ ] | [ ] | [ ] | [ ] | Get the list of notebook cells and their details |
| read/problems | [x] | [x] | [x] | [x] | [x] | Add workspace issues and problems from the Problems panel as context; useful while fixing code or debugging |
| read/readFile | [x] | [x] | [x] | [x] | [x] | Read the content of a file in the workspace |
| read/readNotebookCellOutput | [ ] | [ ] | [ ] | [ ] | [ ] | Read the output from a notebook cell execution |
| read/terminalLastCommand | [x] | [x] | [x] | [x] | [x] | Get the last run terminal command and its output |
| read/terminalSelection | [x] | [x] | [x] | [x] | [x] | Get the text currently selected in the integrated terminal |
| read/viewImage | [x] | [x] | [x] | [x] | [x] | Display an image file inline in the chat *(u)* |
| **search/** | | | | | | |
| search/codebase | [x] | [x] | [x] | [x] | [x] | Perform a semantic code search across the workspace to find relevant context |
| search/changes | [x] | [x] | [x] | [x] | [x] | List source control changes (git diff / SCM history) |
| search/fileSearch | [x] | [x] | [x] | [x] | [x] | Search for files in the workspace using glob patterns; returns file paths |
| search/listDirectory | [x] | [x] | [x] | [x] | [x] | List all files in a given directory in the workspace |
| search/textSearch | [x] | [x] | [x] | [x] | [x] | Find literal or regex text matches across files in the workspace |
| search/usages | [x] | [x] | [x] | [x] | [x] | Combination of Find All References, Find Implementation, and Go to Definition |
| **selection** | | | | | | |
| selection | [ ] | [ ] | [ ] | [ ] | [ ] | Get the current editor selection (only available when text is selected) |
| **todos** | | | | | | |
| todos | [x] | [x] | [x] | [x] | [x] | Track implementation and progress of a chat request with a todo list |
| **vscode/** | | | | | | |
| vscode/askQuestions | [ ] | [ ] | [ ] | [ ] | [ ] | Ask the user clarifying questions via the interactive questions carousel |
| vscode/extensions | [x] | [x] | [x] | [x] | [x] | Search for and ask about VS Code extensions in the Marketplace |
| vscode/getProjectSetupInfo | [ ] | [ ] | [ ] | [ ] | [ ] | Provide instructions and configuration for scaffolding different project types |
| vscode/installExtension | [x] | [x] | [x] | [x] | [x] | Install a VS Code extension from the Marketplace |
| vscode/listCodeUsages | [x] | [x] | [x] | [x] | [x] | List all usages of a symbol using VS Code's language intelligence (Find All References / Go to Definition) *(u)* |
| vscode/memory | [ ] | [ ] | [ ] | [ ] | [ ] | Read and write persistent agent memory across chat sessions *(u)* |
| vscode/renameSymbol | [x] | [x] | [x] | [x] | [x] | Rename a symbol across the workspace using VS Code's language intelligence *(u)* |
| vscode/runCommand | [ ] | [ ] | [ ] | [ ] | [ ] | Execute a VS Code command by ID |
| vscode/VSCodeAPI | [ ] | [ ] | [ ] | [ ] | [ ] | Ask about VS Code functionality and extension development APIs |
| **web/** | | | | | | |
| web/fetch | [x] | [x] | [x] | [x] | [x] | Fetch and return the content from a given URL |
| web/githubRepo | [x] | [x] | [x] | [x] | [x] | Access GitHub repository file content and metadata *(u)* |

## Extension Tool Assignments

Tools contributed by installed VS Code extensions. All currently `[ ]` (this is a Go project — Java/Python/Mermaid tools not needed).

### Debugger for Java (`vscjava.vscode-java-debug`)

| Tool | beast-mode | doc-sync | fix-wf | impl-exec | impl-plan | Description |
|------|:----------:|:--------:|:------:|:---------:|:---------:|-------------|
| debugJavaApplication | [ ] | [ ] | [ ] | [ ] | [ ] | Launch or attach to a Java application in debug mode with automatic compilation and classpath resolution |
| debugStepOperation | [ ] | [ ] | [ ] | [ ] | [ ] | Control program execution flow: stepIn (enter method calls), stepOut (exit current method), stepOver, continue, or pause |
| evaluateDebugExpression | [ ] | [ ] | [ ] | [ ] | [ ] | Evaluate a Java expression in a specific thread's debug context |
| getDebugSessionInfo | [ ] | [ ] | [ ] | [ ] | [ ] | Get information about the currently active Java debug session |
| getDebugStackTrace | [ ] | [ ] | [ ] | [ ] | [ ] | Retrieve the call stack showing all method calls leading to the current execution point |
| getDebugThreads | [ ] | [ ] | [ ] | [ ] | [ ] | List all threads in the debugged Java application with their IDs, names, and states |
| getDebugVariables | [ ] | [ ] | [ ] | [ ] | [ ] | Inspect variables in a specific thread's stack frame: local variables, method parameters, and fields |
| removeJavaBreakpoints | [ ] | [ ] | [ ] | [ ] | [ ] | Remove breakpoints: specific breakpoint by file and line, all breakpoints in a file, or all breakpoints |
| setJavaBreakpoint | [ ] | [ ] | [ ] | [ ] | [ ] | Set a breakpoint at a specific line in Java source code to pause execution and enable debugging |
| stopDebugSession | [ ] | [ ] | [ ] | [ ] | [ ] | Stop the active Java debug session when investigation is complete |

### Mermaid Chat Features (`vscode.mermaid-chat-features`)

| Tool | beast-mode | doc-sync | fix-wf | impl-exec | impl-plan | Description |
|------|:----------:|:--------:|:------:|:---------:|:---------:|-------------|
| vscode.mermaid-chat-features/renderMermaidDiagram | [ ] | [ ] | [ ] | [ ] | [ ] | Render a Mermaid diagram inline in chat |

### Python (`ms-python.python`)

| Tool | beast-mode | doc-sync | fix-wf | impl-exec | impl-plan | Description |
|------|:----------:|:--------:|:------:|:---------:|:---------:|-------------|
| configurePythonEnvironment | [ ] | [ ] | [ ] | [ ] | [ ] | Configure the Python environment for the workspace |
| create_virtual_environment | [ ] | [ ] | [ ] | [ ] | [ ] | Create a new Python virtual environment |
| getPythonEnvironmentInfo | [ ] | [ ] | [ ] | [ ] | [ ] | Get details about the currently selected Python environment |
| getPythonExecutableCommand | [ ] | [ ] | [ ] | [ ] | [ ] | Get the Python executable path and command for running scripts |
| installPythonPackage | [ ] | [ ] | [ ] | [ ] | [ ] | Install one or more Python packages using pip |
| selectEnvironment | [ ] | [ ] | [ ] | [ ] | [ ] | Select a Python environment for the workspace |

## Notes

- **Updating agents**: After editing this table, reflect changes in the corresponding `tools:` list in each `.github/agents/<name>.agent.md` file.
- **Undocumented tools *(u)***: Not listed in the [VS Code agent tools docs](https://code.visualstudio.com/docs/copilot/agents/agent-tools) but observed as functional in agents.
- **Notebook tools**: `edit/editNotebook`, `execute/runNotebookCell`, `read/getNotebookSummary`, `read/readNotebookCellOutput` are only useful for agents that work with Jupyter notebooks.
- **Extension tools**: The tool ID to put in an agent's `tools:` list is the `toolReferenceName` from the extension's `package.json` (camelCase, as shown in the tools picker). For tools with no `toolReferenceName`, use the `name` field (snake_case).
- **MCP tools**: Tools from installed MCP servers appear in the tool picker alongside built-in tools. Add rows here for any that should be selectively enabled per agent.
