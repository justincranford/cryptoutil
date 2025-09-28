---
description: "Instructions for conventional commit message formatting"
applyTo: "**"
---
# Conventional Commit Instructions

## Format
Use the Conventional Commits specification for all commit messages:
```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

## Commit Types
- **feat**: New feature for the user or a particular part of the application
- **fix**: Bug fix for the user, not a fix to a build script
- **docs**: Changes to documentation only
- **style**: Formatting changes that do not affect code meaning (white-space, formatting, missing semi-colons, etc)
- **refactor**: Code change that neither fixes a bug nor adds a feature
- **perf**: Code change that improves performance
- **test**: Adding missing tests or correcting existing tests
- **build**: Changes that affect the build system or external dependencies (example scopes: gulp, broccoli, npm)
- **ci**: Changes to CI configuration files and scripts (example scopes: GitHub Actions, Travis)
- **chore**: Other changes that don't modify src or test files
- **revert**: Reverts a previous commit

## Examples
```
feat: implement DAST security testing with OWASP ZAP and Nuclei
fix: resolve pre-commit hook violations in formatting
docs: enhance API documentation with usage examples
refactor: simplify cryptographic key generation logic
test: add comprehensive unit tests for barrier operations
ci: update GitHub Actions workflow for Go 1.25.1
chore: update dependencies to latest versions
```

## Guidelines
- Use lowercase for type and description
- Keep description under 72 characters for the first line
- Use imperative mood ("add" not "added" or "adds")
- Don't end the subject line with a period
- Use body to explain what and why vs. how
- Reference issues and pull requests when relevant

## Breaking Changes
For breaking changes, add `!` after the type or include `BREAKING CHANGE:` in the footer:
```
feat!: change API response format for elastic keys
```

## Benefits
- Automated semantic versioning
- Generated changelogs
- Clear project history
- Easy filtering of commit types
- Professional repository standards
