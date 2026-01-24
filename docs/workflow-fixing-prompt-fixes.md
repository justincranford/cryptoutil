# Workflow Fixing Prompt - Fixes Applied

## Summary of Changes (2026-01-23)

### Issue 1: gopls Configuration Error ✅

**Problem:**
```
Cannot find the alternate tool gopls configured for gopls.
Please install it and reload this VS Code window.
```

**Solution:**
Created `scripts/fix-gopls.sh` to install gopls:
```bash
./scripts/fix-gopls.sh
```

This will:
1. Check if gopls is installed
2. Install it if missing: `go install golang.org/x/tools/gopls@latest`
3. Verify installation
4. Provide VS Code settings fix instructions

**Manual VS Code Settings Fix:**
If script doesn't resolve it, update `.vscode/settings.json`:
```json
{
  "go.alternateTools": {
    "gopls": "/home/q/go/bin/gopls"
  }
}
```

Or remove the `alternateTools` setting entirely and let VS Code use default gopls.

---

### Issue 2: Prompt File Not Appearing in Agent Dropdown ✅

**Problem:** Missing YAML frontmatter required by GitHub Copilot to recognize agent prompts.

**Fix Applied:** Added required frontmatter to top of file:
```yaml
---
description: Workflow Fixing Agent - Systematically verify and fix all GitHub Actions workflows
tools: ['extensions', 'codebase', 'usages', 'vscodeAPI', 'problems', 'changes', 'testFailure', 'terminalSelection', 'terminalLastCommand', 'fetch', 'search', 'runCommands', 'runTasks', 'editFiles']
---
```

**Result:** Prompt should now appear in GitHub Copilot agent dropdown list.

---

### Issue 3: Improvements from Beast Mode 3.1 ✅

**Key Improvements Added:**

1. **Autonomous Execution Directive**
   - Added explicit instruction to continue until all workflows fixed
   - No stopping between phases
   - Must complete entire todo list before yielding

2. **Todo List Management**
   - Added "How to Create a Todo List" section with format guidelines
   - Instructions to update todo list after each step
   - Display updated list to user after completions
   - Continue to next step instead of ending turn

3. **Internet Research Requirements**
   - Added explicit directive: "THE PROBLEM CANNOT BE SOLVED WITHOUT EXTENSIVE INTERNET RESEARCH"
   - Instructions to use `fetch_webpage` for Google searches
   - Recursive link following for comprehensive research
   - Research dependencies and best practices

4. **Communication Guidelines**
   - Added casual, friendly, professional tone examples
   - Clear structure: bullet points and code blocks
   - Avoid unnecessary explanations and filler
   - Write code directly to files

5. **Git Commit Guidelines**
   - Explicit rules: don't auto-commit without permission
   - Use conventional commit format
   - Make small, atomic commits

6. **Memory Management**
   - Added section on using `.github/instructions/memory.instruction.md`
   - Store recurring issues and workarounds
   - Front matter format for new entries

7. **Critical Autonomous Execution Rules**
   - 7 explicit rules for continuous execution
   - Handle "resume"/"continue" commands
   - Test rigorously and research extensively
   - Display updated todo lists

8. **Tool Specifications**
   - Added tools array in frontmatter
   - Specifies which VS Code/Copilot tools agent can use

---

## Testing the Fixes

### Test 1: Verify gopls Fix
```bash
# Run the fix script
./scripts/fix-gopls.sh

# Reload VS Code window
# Press Ctrl+Shift+P (or Cmd+Shift+P on Mac)
# Type: "Developer: Reload Window"
```

### Test 2: Verify Prompt Appears
1. Open GitHub Copilot Chat
2. Click on agent dropdown (@ icon)
3. Look for "Workflow Fixing Agent" in the list
4. Description should show: "Workflow Fixing Agent - Systematically verify and fix all GitHub Actions workflows"

### Test 3: Verify Autonomous Execution
1. Select "Workflow Fixing Agent" from dropdown
2. Ask: "Fix all failing workflows"
3. Agent should:
   - Create todo list
   - Check workflow statuses
   - Research issues with fetch_webpage
   - Fix workflows one by one
   - Update todo list after each fix
   - Continue until all done
   - NOT stop to ask permission between steps

---

## Key Differences from Original

| Aspect | Original | Improved |
|--------|----------|----------|
| **Frontmatter** | Missing | Added with description + tools |
| **Autonomous** | Basic instruction | 7 explicit execution rules |
| **Todo Lists** | Not mentioned | Full section with format/tracking |
| **Research** | Optional | MANDATORY with fetch_webpage |
| **Communication** | Formal | Casual, friendly, professional |
| **Git Commits** | Basic examples | Explicit rules + permission |
| **Memory** | Not mentioned | Full memory management section |
| **Tools** | Not specified | Explicit tools list in frontmatter |

---

## Next Steps

1. **Reload VS Code** to apply gopls fix
2. **Test the agent** by selecting it from dropdown
3. **Try it out** with: "Check all workflow statuses and fix any failures"
4. **Watch it work autonomously** - should not stop between phases

---

## Additional Notes

The improved prompt now follows GitHub Copilot agent best practices:
- Proper YAML frontmatter for agent recognition
- Clear autonomous execution expectations
- Structured workflow with todo tracking
- Research-first approach with fetch_webpage
- Memory for long-term learning
- Professional communication style

This makes it a true "agent" that can work independently to completion rather than requiring constant user guidance.
