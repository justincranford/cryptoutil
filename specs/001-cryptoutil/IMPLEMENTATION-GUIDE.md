# cryptoutil Quick Implementation Guide

**Use this for actual work. Ignore all other planning documents.**

## 🎯 Start Here: Critical Path (3-5 days)

### Day 1: Address Slow Test Packages (CRITICAL FOUNDATION)

```bash
# Critical: Fix test performance blocking efficient development
# Target: clientauth (168s), jose/server (94s), kms/client (74s)

# 1. Apply aggressive t.Parallel() to clientauth package
# 2. Split jose/server tests into parallel subtests
# 3. Mock KMS dependencies to reduce network roundtrips
# 4. Verify improvements with: go test ./internal/identity/authz/clientauth -v
```

### Day 2: Complete JOSE E2E Tests

```bash
# Priority: Get JOSE fully working end-to-end
cd internal/jose/server/
# Create comprehensive E2E test suite
# Target: Full API coverage with integration tests
```

### Day 3: Fix CI/CD Workflows (HIGH IMPACT)

```bash
# Check current workflow status
go run ./cmd/workflow

# Target these failing workflows:
# - ci-dast, ci-e2e, ci-load (highest impact)
# - ci-coverage, ci-race, ci-sast
# - ci-benchmark, ci-fuzz

# Work through failures systematically:
# 1. Run locally with Act: go run ./cmd/workflow -workflows=dast
# 2. Fix the specific failure
# 3. Verify fix with local run
# 4. Commit and push to verify in GitHub Actions
```

### Day 4: CA OCSP + Docker Integration

```bash
# OCSP responder for CA server
# JOSE Docker Compose integration
# Verify: docker compose up -d && curl -k https://localhost:8080/health
```

### Day 5: Coverage Improvements

```bash
# Focus on these packages to reach 95%:
go test ./internal/ca/handler/... -cover
go test ./internal/identity/auth/userauth/... -cover

# Target: Get both packages to ≥95% coverage
# Use runTests tool for efficient testing
```

---

## 🛠️ Key Commands

**Run Tests**:
```bash
# Use the runTests tool in VS Code Copilot
# OR manually: go test ./... -cover -shuffle=on
```

**Check Workflows**:
```bash
go run ./cmd/workflow
go run ./cmd/workflow -workflows=dast -inputs="scan_profile=quick"
```

**Coverage Check**:
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

**Linting**:
```bash
golangci-lint run --fix
```

---

## 📋 Track Progress

**Update `PROJECT-STATUS.md` as you complete tasks:**
- Change ❌ to ✅ when tasks complete
- Add evidence (commit hashes, test results)
- Keep this as the single source of truth

**Ignore These Files** (historical clutter):
- All `PROGRESS-ITERATION-*.md`
- All `CHECKLIST-ITERATION-*.md`
- All `plan-ITERATION-*.md`
- All `tasks-ITERATION-*.md`

---

## 🚨 Blockers & Workarounds

**EST serverkeygen**: Blocked on PKCS#7 library
- **Workaround**: Skip for now, project can complete without it
- **Resolution**: Use `go.mozilla.org/pkcs7` if/when needed

**Slow tests**: Some packages take 5-10 minutes
- **Workaround**: Use targeted runs: `go test ./specific/package -run=TestSpecific`
- **Reference**: `SLOW-TEST-PACKAGES.md` for timing info

---

## ✅ Success = Green CI/CD + Working Demos

**You're done when**:
1. `go run ./cmd/workflow` shows 11/11 passing ✅
2. `go run ./cmd/demo all` works without errors ✅
3. Coverage reports show ≥95% on target packages ✅

**Estimated effort**: 3-5 focused days

---

*Keep this guide handy. It's all you need to finish the project.*
