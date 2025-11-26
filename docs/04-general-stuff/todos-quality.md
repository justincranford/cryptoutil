# Cryptoutil Code Quality & Linting TODOs

**IMPORTANT**: Delete completed tasks immediately after completion to maintain a clean, actionable TODO list.

**Last Updated**: November 23, 2025
**Status**: Identity subsystem TODOs tracked in docs/02-identityV2/current/MASTER-PLAN.md

---

## 🟡 MEDIUM - Code Quality & Linting Enhancements

### Task CQ1: Identity Subsystem TODO Comments

- **Description**: Identity subsystem contains 40+ TODO comments representing incomplete features
- **Current State**: All non-identity TODOs completed; identity TODOs tracked as feature work in docs/02-identityV2/current/MASTER-PLAN.md
- **Identity Subsystem TODOs** (40+ items - tracked separately as they represent incomplete features):
  - OAuth 2.1 authorization code flow (16 TODOs - Task R01)
  - OIDC login/consent/logout/userinfo endpoints (11 TODOs - Task R02)
  - Client authentication security (5 TODOs - Task R04)
  - Session/token lifecycle management (6 TODOs - Task R05, R08)
  - Advanced authentication (WebAuthn, TOTP, OTP flows)
  - Repository integrations
- **Action Items**:
  - Follow remediation plan in docs/02-identityV2/current/MASTER-PLAN.md
  - Complete critical path tasks (R01-R03) before production
  - Track progress in COMPLETION-STATUS-REPORT.md
- **Files**: `internal/identity/**/*.go`
- **Expected Outcome**: Production-ready OAuth 2.1 / OIDC identity platform
- **Priority**: TRACKED IN IDENTITY V2 MASTER PLAN
- **Note**: godox linter disabled in favor of manual tracking in remediation plan

### Task CQ4: Investigate linters for EOL/maintenance mode dependencies

- **Description**: Research and evaluate tools that can detect dependencies in end-of-life or maintenance mode
- **Current State**: No automated detection of deprecated/unmaintained dependencies
- **Potential Tools to Investigate**:
  - `go-mod-outdated`: Shows outdated dependency versions (<https://github.com/psampaz/go-mod-outdated>)
  - `govulncheck`: Official Go vulnerability scanner (already in use)
  - Custom scripts to check GitHub repository status/README for maintenance warnings
  - Integration with dependency health services or APIs
- **Action Items**:
  - Research available Go tools for dependency lifecycle detection
  - Evaluate feasibility of integrating EOL detection into CI/CD pipeline
  - Consider custom linter or cicd command for maintenance mode checking
  - Document findings and recommend implementation approach
- **Files**: `.golangci.yml`, `internal/cmd/cicd/cicd.go` (potential new command)
- **Expected Outcome**: Automated detection of unmaintained dependencies to prevent security/technical debt
- **Priority**: LOW - Proactive maintenance improvement
- **Timeline**: Q2 2026
