---
description: "Instructions for code quality and maintenance standards"
applyTo: "**"
---
# Code Quality Instructions

- Implement proper resource cleanup (defer statements for HTTP bodies, files, etc.)
- Maintain clear function boundaries (avoid high cyclomatic complexity)
- Remove unused code and parameters
- Validate input parameters in mapper and utility functions
- Wrap all external package errors with context using fmt.Errorf and %w verb to satisfy wrapcheck linter
- Follow maintenance guidelines in files: immediately remove completed/obsolete tasks from actionable lists
