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
| search/fileSearch | [ ] | [ ] | [ ] | [ ] | [ ] | Search for files in the workspace using glob patterns; returns file paths |
| search/listDirectory | [ ] | [ ] | [ ] | [ ] | [ ] | List all files in a given directory in the workspace |
| search/textSearch | [ ] | [ ] | [ ] | [ ] | [ ] | Find literal or regex text matches across files in the workspace |
| search/usages | [x] | [x] | [x] | [x] | [x] | Combination of Find All References, Find Implementation, and Go to Definition |
| **vscode/** | | | | | | |
| vscode/askQuestions | [ ] | [ ] | [ ] | [ ] | [ ] | Ask the user clarifying questions via the interactive questions carousel |
| vscode/extensions | [x] | [x] | [x] | [x] | [x] | Search for and ask about VS Code extensions in the Marketplace |
| vscode/getProjectSetupInfo | [ ] | [ ] | [ ] | [ ] | [ ] | Provide instructions and configuration for scaffolding different project types |
| vscode/installExtension | [x] | [x] | [x] | [x] | [x] | Install a VS Code extension from the Marketplace |
| vscode/memory | [ ] | [ ] | [ ] | [ ] | [ ] | Read and write persistent agent memory across chat sessions *(u)* |
| vscode/runCommand | [ ] | [ ] | [ ] | [ ] | [ ] | Execute a VS Code command by ID |
| vscode/VSCodeAPI | [ ] | [ ] | [ ] | [ ] | [ ] | Ask about VS Code functionality and extension development APIs |
| **web/** | | | | | | |
| web/fetch | [x] | [x] | [x] | [x] | [x] | Fetch and return the content from a given URL |
| web/githubRepo | [x] | [x] | [x] | [x] | [x] | Access GitHub repository file content and metadata *(u)* |

## Notes

- **Updating agents**: After editing this table, reflect changes in the corresponding `tools:` list in each `.github/agents/<name>.agent.md` file.
- **Undocumented tools *(u)***: Not listed in the [VS Code agent tools docs](https://code.visualstudio.com/docs/copilot/agents/agent-tools) but observed as functional in agents.
- **Notebook tools**: `edit/editNotebook`, `execute/runNotebookCell`, `read/getNotebookSummary`, `read/readNotebookCellOutput` are only useful for agents that work with Jupyter notebooks.
- **MCP / extension tools**: Tools from installed MCP servers or VS Code extensions appear in the tool picker alongside built-in tools. Add rows here for any that should be selectively enabled per agent.
