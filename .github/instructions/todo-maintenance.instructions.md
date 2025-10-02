---
description: "Instructions for maintaining actionable TODO/task lists"
applyTo: "**/dast-todos.md"
---
# TODO List Maintenance Instructions

## Critical Requirements
- If a TODO/task list file contains a maintenance guideline requiring deletion of completed tasks, you MUST immediately remove those tasks as soon as they are finished
- Do NOT leave completed, obsolete, or irrelevant tasks in actionable lists—this is mandatory
- Always review files for completed items before ending sessions or marking work complete
- Historical context belongs in commit messages or durable docs, not actionable TODO lists

## Implementation Guidelines
- **When file is large with many completed tasks**: Use create_file to rewrite entire file with only active tasks (preferred method)
- **When text replacement fails**: Create clean version in new file, then replace original with Move-Item
- **Avoid complex replace_string_in_file operations**: Large text blocks often fail due to whitespace/formatting mismatches
- **Validate maintenance guideline compliance**: Check that file contains ONLY active, actionable tasks
- **Test approach first**: For complex cleanups, create clean version, verify content, then replace original

## Failure Recovery
- If replace_string_in_file fails repeatedly: Switch to create_file approach immediately
- If uncertain about exact text: Use read_file to verify current content before replacement
- If file structure is complex: Rewrite entire file with active tasks only
- Always ensure compliance with maintenance guideline regardless of method used
