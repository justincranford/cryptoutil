---
name: fix-tool-names
description: Migrate deprecated Copilot tool names to namespaced identifiers
tools:
  - edit/editFiles
  - search/codebase
  - search/usages
model: claude-sonnet-4.5
argument-hint: "[file-path or 'all']"
---

# Fix Copilot Tool Names Agent

This agent migrates deprecated Copilot Chat tool names to the current namespaced tool identifiers used by VS Code Copilot Chat.

## Mapping Rules

- Old flat tool names are deprecated
- Internal `functions.*` names MUST NOT be used in prompts
- Prompts MUST use namespaced aliases: `<namespace>/<action>`

Valid namespaces:
- search/
- read/
- edit/
- execute/
- vscode/
- web/

## Old → New Tool Name Mapping

### Search

extensions              → vscode/extensions
codebase                → search/codebase
usages                  → search/usages
findTestFiles           → search
searchResults           → search/searchResults
changes                 → search/changes
search                  → search

### Read / Inspect

problems                → read/problems
terminalSelection       → read/terminalSelection
terminalLastCommand     → read/terminalLastCommand
vscodeAPI               → vscode/vscodeAPI

### Edit

editFiles               → edit/editFiles

### Execute / Run

runCommands             → execute/runInTerminal
runTasks                → execute/createAndRunTask
runNotebooks            → execute/runNotebookCell
testFailure             → execute/testFailure

### VS Code / Workspace

extensions              → vscode/extensions
openSimpleBrowser       → vscode/openSimpleBrowser
new                     → vscode/newWorkspace

### Web / Network

fetch                   → web/fetch
githubRepo              → web/githubRepo

---

## Removed / Invalid Tools

think                   → ❌ REMOVED (no replacement)
functions.*             → ❌ INTERNAL ONLY (never valid in prompts)

---

## Validation Rules

- Reject tools without a `/`
- Reject tools starting with `functions.`
- Reject unknown namespaces
- Replace deprecated names using the table above
