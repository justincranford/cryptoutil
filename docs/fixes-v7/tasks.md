# Remaining Tasks - fixes-v7 Followup

**Source**: Distilled from archived fixes-v7 (archive2/)
**Status**: Active

## Phase 1: E2E Verification (Blocked — Docker Desktop)

- [ ] Run `go test -tags=e2e -timeout=30m ./internal/apps/sm/im/e2e/...`
- [ ] Verify sm-im E2E passes end-to-end with fixed OTel config
- [ ] If E2E fails: diagnose and fix (OTel config change: `detectors: [env, system]`)

## Phase 2: Propagation Infrastructure

### 2.1 Reference Validation Script

- [ ] Create `cicd lint-docs validate-propagation` subcommand
- [ ] Extract all `See [ARCHITECTURE.md Section X.Y]` refs from .github/instructions/*.md
- [ ] Extract all `See [ARCHITECTURE.md Section X.Y]` refs from .github/agents/*.md
- [ ] Resolve refs against actual ARCHITECTURE.md section headers
- [ ] Report broken links (ref to non-existent section)
- [ ] Report orphaned sections (high-impact sections with zero refs)
- [ ] Add tests for the validator (>=95% coverage)
- [ ] Add to pre-commit and CI/CD workflow

### 2.2 Section 14 Instruction Coverage

- [ ] Review ARCHITECTURE.md Section 14 content scope
- [ ] Add Operational Excellence content to existing instruction file OR create new file
- [ ] Add `See [ARCHITECTURE.md Section 14]` cross-references

### 2.3 ARCHITECTURE-INDEX.md Sync

- [ ] Compare ARCHITECTURE-INDEX.md against current ARCHITECTURE.md section headers
- [ ] Update if stale
- [ ] Consider adding index validation to `cicd lint-docs`

## Phase 3: Propagation Quality (Medium-Term)

### 3.1 Lint Propagation Coverage

- [ ] Extend `cicd lint-docs` to report section coverage percentage
- [ ] Classify sections as high/medium/low impact
- [ ] Set target: 60% of high-impact sections referenced
- [ ] Track coverage in CI/CD

### 3.2 Content Hash Staleness Detection

- [ ] Design hash storage format (inline in markers or separate file)
- [ ] Implement hash comparison in `cicd lint-docs`
- [ ] Flag stale `@source` blocks where ARCHITECTURE.md content changed
- [ ] Add to CI/CD workflow
