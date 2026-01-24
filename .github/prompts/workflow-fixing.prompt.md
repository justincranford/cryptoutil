# Workflow Fixing Prompt

## Objective

Systematically verify and fix all GitHub Actions workflows to ensure CI/CD health.

## Workflow Verification and Fixing Process

### Phase 1: Initial Assessment

1. **List all workflows and their latest runs**
   ```bash
   ls -la .github/workflows/
   # Identify all workflow files
   ```

2. **Check latest workflow runs for each workflow**
   - Use GitHub web interface: https://github.com/justincranford/cryptoutil/actions
   - Identify failing workflows
   - Document failure patterns

### Phase 2: Categorize Failures

Group failures by root cause:
- **Configuration issues**: YAML syntax, environment variables, secrets
- **Dependency issues**: Missing packages, version conflicts
- **Test failures**: Flaky tests, environment-specific failures
- **Docker issues**: Image build failures, compose issues
- **Timeout issues**: Long-running operations exceeding limits

### Phase 3: Fix Each Failed Workflow

For each failing workflow, follow this pattern:

#### 3.1 Analyze Failure

1. Download workflow logs (if available via API):
   ```bash
   # Use GitHub API or web interface to get logs
   ```

2. Identify error message and stack trace
3. Determine root cause category
4. Check if similar failure exists in other workflows

#### 3.2 Implement Fix

Based on failure category:

**Configuration Fixes:**
- Fix YAML syntax errors
- Add/update environment variables
- Configure GitHub secrets
- Update workflow triggers (paths, branches)

**Dependency Fixes:**
- Update package versions in workflow
- Add missing dependencies
- Fix version conflicts
- Update Go version, golangci-lint version, etc.

**Test Fixes:**
- Fix flaky tests (add retries, increase timeouts)
- Skip platform-specific tests appropriately
- Fix race conditions
- Add proper cleanup in tests

**Docker Fixes:**
- Fix Dockerfile syntax errors
- Update Docker Compose configurations
- Fix image build issues
- Add healthchecks
- Fix port conflicts

**Timeout Fixes:**
- Increase workflow timeout
- Optimize long-running operations
- Add caching
- Parallelize where possible

#### 3.3 Test Locally (if possible)

For workflows that can be tested locally:
```bash
# Using act for local workflow testing
act -W .github/workflows/ci-quality.yml

# Or test specific jobs
go test ./...
golangci-lint run
docker compose up
```

#### 3.4 Commit Fix

```bash
git add .github/workflows/WORKFLOW_NAME.yml
# OR fix source code causing test failures
git add path/to/fixed/file.go

git commit -m "fix(ci): fix WORKFLOW_NAME - DESCRIPTION"
# Examples:
# git commit -m "fix(ci): fix ci-coverage - add missing PostgreSQL service"
# git commit -m "fix(test): fix consent_expired test - SQLite datetime comparison"
# git commit -m "fix(docker): fix ci-e2e - update Dockerfile healthcheck syntax"
```

### Phase 4: Batch Push and Monitor

After fixing multiple workflows:

1. **Batch push all fixes**
   ```bash
   git push origin main
   ```

2. **Monitor workflow runs**
   - Watch workflows as they trigger
   - Check each as it finishes (don't wait for all)
   - Identify new failures

3. **Iterate**
   - For each new failure, go back to Phase 3
   - Continue until all workflows pass

### Phase 5: Final Verification

Once all workflows pass:

1. **Run comprehensive local tests**
   ```bash
   go test ./... -cover -shuffle=on
   golangci-lint run
   docker compose -f deployments/compose/compose.yml up -d
   ```

2. **Verify all workflows are green**
   - Check GitHub Actions page
   - Ensure no pending or failing runs

3. **Document any known issues**
   - Create issues for non-critical failures
   - Document workarounds
   - Update README or docs if needed

## Common Workflow Patterns

### CI Quality Workflow

Typical issues:
- golangci-lint version mismatch
- Import alias violations
- Format violations

Fixes:
- Update golangci-lint version
- Run `golangci-lint run --fix`
- Run `go run ./cmd/cicd format-go`

### CI Coverage Workflow

Typical issues:
- PostgreSQL service not available
- Timeout on long-running tests
- Coverage targets not met

Fixes:
- Add PostgreSQL service to workflow
- Increase timeout
- Fix tests to improve coverage

### CI E2E Workflow

Typical issues:
- Docker Compose failures
- Image build failures
- Healthcheck failures

Fixes:
- Fix Dockerfile syntax
- Update compose.yml
- Fix healthcheck commands

### CI DAST Workflow

Typical issues:
- Variable expansion in heredocs
- PostgreSQL connection failures
- Nuclei template errors

Fixes:
- Use `${VAR}` syntax (not `$VAR`)
- Verify PostgreSQL credentials
- Update Nuclei templates

## Workflow-Specific Fixes

### Known Issues

1. **consent_expired test failure**
   - **Root cause**: SQLite datetime comparison not working correctly on Linux
   - **Location**: `internal/identity/repository/orm/consent_decision_repository_test.go`
   - **Fix**: Update query to use SQLite-compatible datetime comparison OR skip test on Linux
   - **Commit**: `fix(test): fix consent_expired test - SQLite datetime comparison`

2. **E2E Docker Compose failures**
   - **Root cause**: Dockerfile healthcheck syntax error (`--start_period` should be `--start-period`)
   - **Location**: `deployments/*/Dockerfile.*`
   - **Fix**: Replace `--start_period=30s` with `--start-period=30s`
   - **Commit**: `fix(docker): fix healthcheck syntax - use --start-period`

3. **Import alias violations**
   - **Root cause**: Unaliased cryptoutil imports
   - **Fix**: Run `golangci-lint run --fix`
   - **Commit**: `style: fix import aliases`

## Automation Script

Create a script to automate workflow monitoring:

```bash
#!/bin/bash
# scripts/monitor-workflows.sh

# List all workflows
echo "=== All Workflows ==="
ls -1 .github/workflows/

echo ""
echo "=== Latest Workflow Runs ==="
# Use GitHub CLI or API to list runs
# gh run list --limit 20

echo ""
echo "=== Failed Workflows ==="
# Filter for failures
# gh run list --status failure --limit 10
```

## Success Criteria

All workflows must:
- ✅ Pass successfully
- ✅ Complete within timeout limits
- ✅ Have no flaky failures
- ✅ Be documented (if known issues exist)

## Final Steps

1. Commit all fixes
2. Push to main
3. Verify all workflows pass
4. Update this prompt with lessons learned
5. Document any permanent workarounds in project documentation

---

## Execution Template

When executing this prompt:

1. Start with Phase 1 (assessment)
2. Document all failures found
3. Proceed to Phase 3 for each failure
4. Batch push after fixing 5-10 workflows
5. Monitor and iterate
6. Continue until all workflows pass
7. Final verification
8. Update documentation

**CRITICAL: Do NOT stop between phases - continue autonomous execution until all workflows are fixed or blocked.**
