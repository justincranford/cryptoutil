# V10 Implementation - Questions for User

**Version**: 1
**Created**: 2026-02-05
**Status**: Awaiting Answers

---

## Overview

This quizme file contains questions requiring user input to make informed decisions for V10 implementation.

**Instructions**: Mark your answer with `[x]` or fill in Choice E.

---

## Question 1: E2E Health Check Endpoint Standardization

**Issue**: Docker uses `/admin/api/v1/livez` but E2E uses `/service/api/v1/health`.

- [ ] **A**: Standardize ALL on `/admin/api/v1/livez`
- [ ] **B**: Standardize ALL on `/service/api/v1/health`
- [ ] **C**: Keep BOTH endpoints
- [ ] **D**: Deprecate service endpoint

**E**: Custom Answer: _______________

---

## Question 2: sm-kms cmd Structure Pattern

**Issue**: cipher-im has rich structure, sm-kms/jose-ja minimal.

- [ ] **A**: Rich structure for ALL services
- [ ] **B**: Minimal structure for ALL services
- [ ] **C**: Hybrid (main.go + Dockerfile in cmd/, compose in deployments/)
- [ ] **D**: Per-service decision (no standard)

**E**: Custom Answer: _______________

---

## Question 3: V9 Completion - 71% vs 100%

**Issue**: Plan claims 71% but grep shows all complete/skipped.

- [ ] **A**: V9 is 100% complete, skip Phase 4
- [ ] **B**: Skipped tasks count as incomplete
- [ ] **C**: Manual re-audit needed
- [ ] **D**: Deferred work goes to V11

**E**: Custom Answer: _______________

---

## Question 4: unsealkeysservice Duplication Audit

**Issue**: Verify no duplicate unseal logic in template.

- [ ] **A**: Quick comparison (15 min)
- [ ] **B**: Comprehensive line-by-line (45 min)
- [ ] **C**: Structural analysis - verify imports only (20 min)
- [ ] **D**: Full codebase audit all services (1-2 hours)

**E**: Custom Answer: _______________

---

## Question 5: V8 Incomplete Task (58/59)

**Issue**: One task incomplete, what if blocked?

- [ ] **A**: Complete in V10 regardless
- [ ] **B**: Mark complete if work done
- [ ] **C**: Create V11 for blocker
- [ ] **D**: Audit if truly incomplete (292 total tasks found)

**E**: Custom Answer: _______________

---

## Question 6: E2E Timeout Values

**Issue**: cipher-im times out at 90s.

- [ ] **A**: Aggressive (E2E 30s, Docker 5s)
- [ ] **B**: Conservative (E2E 120s, Docker 15s)
- [ ] **C**: Per-service based on startup profile
- [ ] **D**: Dynamic with retry backoff

**E**: Custom Answer: _______________

---

## Submit Instructions

Mark answers, notify agent, file will be deleted after merge.
