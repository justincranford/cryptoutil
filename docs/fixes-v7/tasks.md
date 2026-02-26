# Remaining Tasks - fixes-v7 Followup

**Source**: Distilled from archived fixes-v7 (archive2/)
**Status**: Active

## Phase 1: E2E Verification ✅ COMPLETE

- [x] Run sm-im E2E — PASS (committed da860dd8)
- [x] Run jose-ja E2E — PASS (18.274s, committed 6086fb29)
- [x] Run sm-kms E2E — PASS (41.609s, committed 6086fb29)
- [x] Run identity E2E — PASS (6.823s, 5 services, committed 6086fb29)
- [x] Deep research: identified 13 root causes (see plan.md for details)
- [x] Fix root cause #1: OTel docker detector (committed 7b3b78c2)
- [x] Fix root cause #2: ComposeManager profiles (committed da860dd8)
- [x] Fix root cause #3: CLI args os.Args bug (committed da860dd8)
- [x] Fix root cause #4: Missing SQLite database URL (committed 6086fb29)
- [x] Fix root cause #5: Port override via CLI flag (committed 6086fb29)
- [x] Fix root cause #6: browser_session_jwks test (committed da860dd8)
- [x] Fix root cause #7: Docker image caching --build (committed 6086fb29)
- [x] Fix root cause #8: "start" vs "server" subcommand (committed 6086fb29)
- [x] Fix root cause #9: sm-kms postgres hostname (committed 6086fb29)
- [x] Fix root cause #10: GLOB CHECK SQLite-only (committed 6086fb29)
- [x] Fix root cause #11: BLOB type PostgreSQL (committed 6086fb29)
- [x] Fix root cause #12: DROP TABLE FK cascade (committed 6086fb29)
- [x] Fix root cause #13: Identity unseal secrets too short (committed 6086fb29)
- [x] Rewrite 20 identity config files to flat kebab-case (committed 6086fb29)
- [x] All 62 deployment validators pass

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
