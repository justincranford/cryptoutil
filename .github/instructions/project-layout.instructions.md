---
description: "Instructions for Go project layout"
applyTo: "**"
---
# Go Project Layout Instructions

- Follow [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- Use `/cmd`, `/internal`, `/pkg`, `/api`, `/configs`, `/scripts`, `/docs`, `/deployments` as described in the repo
- Keep main apps minimal, logic in `/internal` or `/pkg`
- Test directories may contain non-Go code (e.g., Java Gatling performance tests in `/test/gatling/`)
- Avoid: logic in `/cmd`, `/src` at root, deep nesting
