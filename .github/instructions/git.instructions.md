---
description: "Instructions for git operations and workflow"
applyTo: "**"
---
# Git Operations Instructions

- **Commit vs Push**: Commit frequently for logical units of work; push only when ready for CI/CD and peer review
- **Pre-push hooks**: Run automatically before push to enforce quality gates (dependency checks, linting)
- **Dependency management**: Update dependencies incrementally with test validation between updates
- **Branch strategy**: Use feature branches for development; merge to main only after CI passes
- **Commit hygiene**: Keep commits atomic and focused; use conventional commit messages
- **Push readiness**: Ensure all pre-commit checks pass before pushing; resolve any hook failures
