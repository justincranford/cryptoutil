---
description: Use for GitHub Actions workflow failures, CI/CD repair, workflow validation, or any work touching .github/workflows/*.yml files. Requires Docker Desktop for local testing.
---

# Fix Workflows — GitHub Actions Specialist

**Full Copilot original**: [.github/agents/fix-workflows.agent.md](../../.github/agents/fix-workflows.agent.md)

## Zero-Failure Tolerance

Fix ALL workflow failures. Do not stop at the first fix — run workflows locally, verify they pass end-to-end before committing.

## Local Testing (MANDATORY)

Before pushing workflow changes, test locally:
```bash
# Run specific workflow type
act -W .github/workflows/ci-quality.yml
act -W .github/workflows/ci-coverage.yml
act -W .github/workflows/ci-e2e.yml
```

Docker Desktop MUST be running for local workflow testing.

## Security Checklist (14 Mandatory Checks)

Every workflow MUST:
1. Pin all actions to commit SHA (not tag)
2. Use `permissions:` with least-privilege (`contents: read` minimum)
3. No secrets in `run:` steps (use `${{ secrets.NAME }}`)
4. No `pull_request_target` without explicit protection
5. No `workflow_dispatch` inputs echoed to shell without sanitization
6. `GITHUB_TOKEN` scoped to minimum needed
7. No `actions/checkout` with `persist-credentials: true` unless needed
8. Cache keys use hash of lockfile, not branch name
9. No hardcoded version strings — use `.github/versions-rules.xml`
10. Matrix strategies don't expose secret values
11. Artifacts not uploaded with overly broad paths
12. Concurrency groups cancel only stale runs (`cancel-in-progress: true`)
13. All `uses:` reference pinned SHA in `workflow-outdated-action-exemptions.json`
14. No shell injection via `${{ github.event.* }}` in `run:` steps

## Evidence Collection Pattern

Store evidence in `./workflow-reports/`:
```
./workflow-reports/
├── {workflow-name}/
│   ├── before.txt    # Failure output before fix
│   ├── after.txt     # Success output after fix
│   └── diff.patch    # Changes made
```

## Workflow Types and Execution

| Workflow | File | Command |
|----------|------|---------|
| Quality | ci-quality.yml | `golangci-lint run --timeout=30m` |
| Coverage | ci-coverage.yml | `go test -coverprofile=coverage.out ./...` |
| E2E | ci-e2e.yml | Docker Compose + real services |
| Mutation | ci-mutation.yml | `gremlins unleash ./...` |
| Load | ci-load.yml | k6 load tests |
| SAST | ci-sast.yml | Semgrep |
| Gitleaks | ci-gitleaks.yml | `gitleaks detect` |
| Fitness | ci-fitness.yml | `go run cmd/cicd-lint/main.go lint-fitness` |

## Completion Criteria

- All failing workflows pass locally
- No new security issues introduced
- All action versions pinned to SHA
- `actionlint` passes on all modified workflows
