# Refactor Learn Product and learn-im Service

## Overview

This document outlines the tasks required to refactor the learn product and learn-im service to support ALL THREE command-line patterns defined in `docs/CMD-PATTERN.md`: Suite, Product, and Product-Service.

**Status**: ✅ APPROVED - All questions answered, ready for implementation

**Goal**:

- Support ALL 3 CLI patterns (Suite, Product, Product-Service)
- Full configuration hierarchy (cryptoutil/learn/im)
- All subcommands (server, client, init, health, livez, readyz, shutdown)
- PostgreSQL + SQLite support (no CGO)
- Comprehensive testing (unit, integration, E2E)
- All orchestration tools (compose, demo, e2e)
- NO breaking changes (learn product never released)
- This is the FIRST refactoring; all other products (identity, jose, pki, sm) will follow this pattern

---

## Prerequisites

- [x] Review `docs/CMD-PATTERN.md` for pattern specifications
- [x] Review `docs/REFACTOR-LEARN-IM-QUIZME.md` and answer all questions - **COMPLETED**
- [x] Analyze existing learn-im implementation in `cmd/learn-im/` and `internal/learn/`
- [x] Review service template requirements in `.github/instructions/02-02.service-template.instructions.md`

## Key Decisions from QUIZME Answers

**Q1**: ALL 3 CLI patterns (Suite, Product, Product-Service) - REQUIRED
**Q2**: internal/cmd/learn/learn.go structure - REQUIRED for consistency
**Q3**: Full config hierarchy (cryptoutil/learn/im) - REQUIRED
**Q4**: ALL subcommands (server, client, init, health, livez, readyz, shutdown) - REQUIRED
**Q5**: CLI wrappers calling admin API endpoints
**Q6**: PostgreSQL + SQLite support - REQUIRED
**Q7**: No breaking changes (learn-im defaults to server)
**Q8**: Avoid breaking OTHER products (leave them alone)
**Q9**: Comprehensive testing - REQUIRED
**Q10**: Support both old and new command structures in tests
**Q11**: No documentation updates (keep current behavior)
**Q12**: Update ALL orchestration tools (compose, demo, e2e) - REQUIRED
**Q13**: Move to deployments/compose/learn/compose.yml
**Q14**: Big bang approach (refactor everything at once)
**Q15**: High priority - do immediately
**Q16**: Keep names (learn, learn-im)
**Q17**: Demonstrate ALL service template features
**Q18**: Full configuration layering (3 levels with merging/override)

---

## Phase 1: Directory Structure and Internal Command Infrastructure

### Task 1.1: Create Internal Command Structure

**Status**: ❌ Not Started

**Description**: Create the internal command module that will handle all learn product commands.

**Files to Create**:

- `internal/cmd/learn/learn.go` - Main learn product command router
- `internal/cmd/learn/im.go` - Instant messaging service command handler (or integrate into learn.go)

**Implementation**:

```go
// internal/cmd/learn/learn.go
package learn

// Learn implements the learn product command router
func Learn(args []string) int {
    // Route to im(args) for instant messaging service
}

// im implements the instant messaging service subcommand handler
func im(args []string) int {
    // Handle subcommands: server, client, init, health, livez, readyz, shutdown
}
```

**Dependencies**: None

**Completion Criteria**:

- [ ] `internal/cmd/learn/learn.go` created with proper structure
- [ ] Command routing logic implemented
- [ ] Unit tests for command parsing (≥95% coverage)
- [ ] Conventional commit created

---

### Task 1.2: Create Configuration Hierarchy

**Status**: ❌ Not Started

**Description**: Create the configuration directory structure matching other products.

**Files to Create** (pending Q3 decision):

- `configs/learn/config.yml` - Product-level configuration (if needed)
- `configs/learn/im/config.yml` - Service-specific configuration
- Migration script to convert existing configs

**Implementation Notes**:

- Follow configuration layering pattern: cryptoutil → product → service
- Support config merging and override logic
- Maintain backward compatibility with existing config locations

**Dependencies**: Q3, Q18 answers

**Completion Criteria**:

- [ ] Directory structure created under `configs/learn/`
- [ ] Configuration files created with proper YAML structure
- [ ] Migration path documented for existing configs
- [ ] Conventional commit created

---

## Phase 2: Subcommand Implementation

### Task 2.1: Implement `server` Subcommand

**Status**: ❌ Not Started

**Description**: Refactor existing server startup logic into a proper `server` subcommand.

**Files to Modify**:

- `cmd/learn-im/main.go` - Update to call new command structure
- `internal/cmd/learn/learn.go` - Add server subcommand handler

**Implementation**:

- Extract server initialization from `main.go` to reusable function
- Add command-line flag parsing for server configuration
- Support both `learn-im` (backward compatible) and `learn-im server` (new pattern)

**Dependencies**: Task 1.1, Q7 decision

**Completion Criteria**:

- [ ] Server starts successfully with `learn-im server` command
- [ ] Backward compatibility maintained (if required by Q7)
- [ ] All existing server functionality preserved
- [ ] Unit tests updated (≥95% coverage)
- [ ] Integration tests pass
- [ ] Conventional commit created

---

### Task 2.2: Implement `client` Subcommand (Optional)

**Status**: ❌ Not Started

**Description**: Create client subcommand for learn-im service operations.

**Files to Create**:

- `internal/cmd/learn/client.go` - Client command implementation

**Implementation** (depends on Q4 follow-up answers):

- Message send/receive operations
- User management operations
- Configuration operations
- Help text and usage examples

**Dependencies**: Task 1.1, Q4 decision

**Completion Criteria**:

- [ ] Client subcommand implemented with defined operations
- [ ] Help text and examples provided
- [ ] Unit tests created (≥95% coverage)
- [ ] Integration tests with running server
- [ ] Conventional commit created

---

### Task 2.3: Implement `init` Subcommand (Optional)

**Status**: ❌ Not Started

**Description**: Create initialization subcommand for database setup and configuration generation.

**Files to Create**:

- `internal/cmd/learn/init.go` - Initialization command implementation

**Implementation** (depends on Q4 follow-up answers):

- Database schema initialization
- Default user creation
- Configuration file generation
- Idempotency (safe to run multiple times)

**Dependencies**: Task 1.1, Q4 decision

**Completion Criteria**:

- [ ] Init subcommand creates required database schema
- [ ] Default configuration generated if missing
- [ ] Idempotent operation (no errors on re-run)
- [ ] Unit tests created (≥95% coverage)
- [ ] Conventional commit created

---

### Task 2.4: Implement Health Check Subcommands (Optional)

**Status**: ❌ Not Started

**Description**: Create health, livez, and readyz subcommands.

**Files to Create**:

- `internal/cmd/learn/health.go` - Health check commands

**Implementation Options** (depends on Q5 decision):

**Option A - CLI Wrappers**:

- Make HTTP requests to `/admin/v1/livez`, `/admin/v1/readyz` endpoints
- Parse responses and return appropriate exit codes

**Option B - Independent**:

- Implement health check logic directly
- Check database connectivity, dependencies
- No HTTP dependency

**Dependencies**: Task 1.1, Q4, Q5 decisions

**Completion Criteria**:

- [ ] `health` subcommand implemented
- [ ] `livez` subcommand implemented (lightweight check)
- [ ] `readyz` subcommand implemented (comprehensive check)
- [ ] Proper exit codes (0=healthy, non-zero=unhealthy)
- [ ] Unit tests created (≥95% coverage)
- [ ] Conventional commit created

---

### Task 2.5: Implement `shutdown` Subcommand (Optional)

**Status**: ❌ Not Started

**Description**: Create graceful shutdown subcommand.

**Files to Create**:

- `internal/cmd/learn/shutdown.go` - Shutdown command implementation

**Implementation Options** (depends on Q5 decision):

**Option A - Admin API**:

- Make HTTP POST to `/admin/v1/shutdown`
- Wait for graceful shutdown completion

**Option B - Signal-Based**:

- Send SIGTERM signal to running process
- Monitor for process termination

**Dependencies**: Task 1.1, Q4, Q5 decisions

**Completion Criteria**:

- [ ] Shutdown triggers graceful termination
- [ ] Active connections drained (30s timeout)
- [ ] Resources released properly
- [ ] Unit tests created (≥95% coverage)
- [ ] Integration tests verify graceful behavior
- [ ] Conventional commit created

---

## Phase 3: Database Configuration (Optional)

### Task 3.1: Add PostgreSQL Support

**Status**: ❌ Not Started

**Description**: Add PostgreSQL support alongside existing SQLite implementation.

**Files to Modify**:

- `internal/learn/server/database.go` - Add PostgreSQL initialization
- Configuration files - Add database connection settings

**Implementation** (depends on Q6 decision):

- Follow pattern from service template
- Support both SQLite and PostgreSQL via configuration
- Maintain SQLite as default for simplicity
- Use GORM for database abstraction

**Dependencies**: Q6 decision

**Completion Criteria**:

- [ ] PostgreSQL initialization code added
- [ ] Database selection via configuration
- [ ] Both SQLite and PostgreSQL tested
- [ ] Connection pooling configured
- [ ] Migration scripts compatible with both databases
- [ ] Unit tests for both database types (≥95% coverage)
- [ ] Conventional commit created

---

## Phase 4: Testing Updates

### Task 4.1: Update Unit Tests

**Status**: ❌ Not Started

**Description**: Create/update unit tests for new command structure.

**Files to Create/Modify**:

- `internal/cmd/learn/learn_test.go` - Command routing tests
- `internal/cmd/learn/*_test.go` - Subcommand tests

**Test Coverage Requirements**:

- Command parsing and routing: ≥95%
- Subcommand execution: ≥95%
- Error handling: All error paths
- Flag validation: All flag combinations

**Dependencies**: Phase 2 tasks

**Completion Criteria**:

- [ ] Unit tests created for all new code
- [ ] Coverage ≥95% for internal/cmd/learn/
- [ ] All tests pass with `go test ./internal/cmd/learn/...`
- [ ] Table-driven test pattern used
- [ ] Conventional commit created

---

### Task 4.2: Update Integration Tests

**Status**: ❌ Not Started

**Description**: Update integration tests to use new command structure.

**Files to Modify**:

- Test helper functions
- Test fixtures
- Integration test suite

**Dependencies**: Phase 2 tasks, Q10 decision

**Completion Criteria**:

- [ ] Integration tests pass with new command structure
- [ ] Backward compatibility tested (if required by Q8)
- [ ] All service interactions verified
- [ ] Conventional commit created

---

### Task 4.3: Update E2E Tests

**Status**: ❌ Not Started

**Description**: Update end-to-end tests to use new command patterns.

**Files to Modify**:

- `internal/learn/e2e/learn_im_e2e_test.go`
- E2E test configurations
- Docker Compose files used in tests

**Dependencies**: Phase 2 tasks, Q10 decision

**Completion Criteria**:

- [ ] E2E tests updated to new command structure
- [ ] Both `/service/**` and `/browser/**` paths tested
- [ ] Docker Compose E2E scenarios pass
- [ ] Test timing <45s per E2E test
- [ ] Conventional commit created

---

## Phase 5: Documentation Updates

### Task 5.1: Update README and User Documentation

**Status**: ❌ Not Started

**Description**: Update all user-facing documentation.

**Files to Modify**:

- `cmd/learn-im/README.md`
- `cmd/learn-im/API.md`
- Main project `README.md` (if learn-im referenced)

**Documentation Updates**:

- Command syntax examples
- Configuration file locations
- Subcommand descriptions
- Migration guide for existing users
- Troubleshooting section

**Dependencies**: All implementation tasks, Q11 decision

**Completion Criteria**:

- [ ] README updated with new command syntax
- [ ] Examples demonstrate all subcommands
- [ ] Migration guide created
- [ ] API documentation reflects any changes
- [ ] Conventional commit created

---

### Task 5.2: Update Docker and Deployment Documentation

**Status**: ❌ Not Started

**Description**: Update Docker Compose files and deployment scripts.

**Files to Modify**:

- `cmd/learn-im/docker-compose.yml`
- `cmd/learn-im/docker-compose.dev.yml`
- `cmd/learn-im/Dockerfile`
- Deployment runbooks (if any)

**Implementation**:

- Update CMD in Dockerfile to use new command structure
- Update docker-compose.yml command overrides
- Document environment-specific configurations
- Update health check commands if needed

**Dependencies**: Phase 2 tasks, Q11, Q13 decisions

**Completion Criteria**:

- [ ] Dockerfile CMD updated
- [ ] Docker Compose files updated
- [ ] Health checks use new command structure
- [ ] All deployment documentation updated
- [ ] Conventional commit created

---

## Phase 6: Integration with Orchestration Tools (Optional)

### Task 6.1: Update cryptoutil-compose

**Status**: ❌ Not Started

**Description**: Add learn product support to cryptoutil-compose orchestration tool.

**Files to Create/Modify**:

- `internal/cmd/cryptoutil-compose/` - Add learn product handlers
- Compose configuration files

**Implementation**:

- `cryptoutil-compose learn up` - Start all learn services
- `cryptoutil-compose learn-im up` - Start learn-im service only
- `cryptoutil-compose learn down` - Stop all learn services
- `cryptoutil-compose learn status` - Show learn services status
- `cryptoutil-compose learn clean` - Clean up learn containers

**Dependencies**: Q12 decision, Phase 2 tasks

**Completion Criteria**:

- [ ] Orchestration commands implemented
- [ ] Integration with existing compose patterns
- [ ] Unit tests created (≥95% coverage)
- [ ] Manual testing verified
- [ ] Conventional commit created

---

### Task 6.2: Update cryptoutil-demo (Optional)

**Status**: ❌ Not Started

**Description**: Add learn product support to demonstration tool.

**Files to Create/Modify**:

- `internal/cmd/cryptoutil-demo/` - Add learn demo scenarios
- Demo configuration files

**Implementation**:

- Create demo scenarios showcasing learn-im features
- Integrate with existing demo framework
- Add demo data and test scenarios

**Dependencies**: Q12 decision, Task 6.1

**Completion Criteria**:

- [ ] Demo scenarios implemented
- [ ] Demo configurations created
- [ ] Manual demonstration tested
- [ ] Demo documentation created
- [ ] Conventional commit created

---

### Task 6.3: Update cryptoutil-e2e (Optional)

**Status**: ❌ Not Started

**Description**: Add learn product support to E2E testing tool.

**Files to Create/Modify**:

- `internal/cmd/cryptoutil-e2e/` - Add learn E2E test orchestration
- E2E test configurations

**Implementation**:

- `cryptoutil-e2e learn up` - Start learn services for E2E
- `cryptoutil-e2e learn full` - Run full E2E test suite
- `cryptoutil-e2e learn smoke` - Run smoke tests
- Integration with CI/CD workflows

**Dependencies**: Q12 decision, Task 6.1

**Completion Criteria**:

- [ ] E2E orchestration commands implemented
- [ ] Test configurations created
- [ ] CI/CD integration verified
- [ ] All E2E tests pass
- [ ] Conventional commit created

---

## Phase 7: Docker Compose Migration (Optional)

### Task 7.1: Move Docker Compose Files to Standard Location

**Status**: ❌ Not Started

**Description**: Move compose files from `cmd/learn-im/` to `deployments/compose/learn/`.

**Files to Move/Create**:

- `deployments/compose/learn/compose.yml` (from `cmd/learn-im/docker-compose.yml`)
- `deployments/compose/learn/compose.dev.yml` (from `cmd/learn-im/docker-compose.dev.yml`)
- `deployments/compose/learn/compose.e2e.yml` (new, for E2E tests)

**Implementation**:

- Maintain backward compatibility via symlinks (if required by Q13)
- Update all references in documentation
- Update CI/CD workflows

**Dependencies**: Q13 decision

**Completion Criteria**:

- [ ] Compose files moved to standard location
- [ ] Backward compatibility maintained (if required)
- [ ] All references updated
- [ ] CI/CD workflows pass
- [ ] Conventional commit created

---

## Phase 8: Final Validation and Cleanup

### Task 8.1: Comprehensive Testing

**Status**: ❌ Not Started

**Description**: Run full test suite to validate all changes.

**Test Execution**:

```bash
# Unit tests
go test ./internal/cmd/learn/... -cover -shuffle=on

# Integration tests
go test ./internal/learn/... -tags=integration

# E2E tests
go test ./internal/learn/e2e/... -tags=e2e

# Linting
golangci-lint run --fix
golangci-lint run

# Build verification
go build ./cmd/learn-im
```

**Dependencies**: All implementation tasks

**Completion Criteria**:

- [ ] All unit tests pass (≥95% coverage)
- [ ] All integration tests pass
- [ ] All E2E tests pass
- [ ] Linting passes with zero issues
- [ ] Build successful
- [ ] No new TODOs introduced

---

### Task 8.2: Update Project Documentation

**Status**: ❌ Not Started

**Description**: Update high-level project documentation.

**Files to Update**:

- Main `README.md` - Update learn product references
- `docs/CMD-PATTERN.md` - Add learn product examples (if missing)
- Architecture diagrams (if any)

**Dependencies**: All implementation tasks

**Completion Criteria**:

- [ ] Main README reflects new structure
- [ ] CMD-PATTERN examples verified/updated
- [ ] Architecture documentation updated
- [ ] Conventional commit created

---

### Task 8.3: Backward Compatibility Verification (If Required)

**Status**: ❌ Not Started

**Description**: Verify backward compatibility with existing deployments.

**Verification Steps**:

- Test old command syntax still works (if Q8 answer requires)
- Verify existing Docker Compose files work
- Check existing scripts and automation
- Document migration path for users

**Dependencies**: Q8 decision, all implementation tasks

**Completion Criteria**:

- [ ] Old command patterns verified (if required)
- [ ] Existing deployments tested
- [ ] Migration guide complete
- [ ] Breaking changes documented

---

## Phase 9: CI/CD Integration

### Task 9.1: Update GitHub Workflows

**Status**: ❌ Not Started

**Description**: Update CI/CD workflows to test new command structure.

**Files to Modify**:

- `.github/workflows/ci-test.yml`
- `.github/workflows/ci-e2e.yml`
- Other workflow files referencing learn-im

**Implementation**:

- Update test commands to use new structure
- Add tests for all subcommands
- Verify Docker builds with new CMD

**Dependencies**: Phase 2 tasks

**Completion Criteria**:

- [ ] CI workflows pass with new command structure
- [ ] All subcommands tested in CI
- [ ] Docker builds verified
- [ ] Conventional commit created

---

## Summary and Metrics

### Estimated Effort

- **Phase 1-2** (Core Structure): 8-16 hours
- **Phase 3** (PostgreSQL): 4-8 hours (optional)
- **Phase 4** (Testing): 8-12 hours
- **Phase 5** (Documentation): 4-6 hours
- **Phase 6** (Orchestration): 8-12 hours (optional)
- **Phase 7** (Migration): 2-4 hours (optional)
- **Phase 8-9** (Validation & CI): 4-6 hours

**Total**: 38-64 hours (depending on options selected)

### Dependencies

Many tasks depend on answers to questions in `REFACTOR-LEARN-IM-QUIZME.md`:

- Q3, Q18: Configuration structure
- Q4, Q5: Subcommand implementation scope
- Q6: Database support
- Q7, Q8: Backward compatibility
- Q10: E2E test updates
- Q11, Q13: Documentation scope
- Q12: Orchestration integration
- Q14: Migration approach

### Success Criteria

- [ ] All selected subcommands implemented and tested
- [ ] Configuration hierarchy created (if selected)
- [ ] Tests pass with ≥95% coverage (production), ≥98% (infrastructure)
- [ ] Documentation complete and accurate
- [ ] CI/CD workflows pass
- [ ] Backward compatibility maintained (if required)
- [ ] Zero linting issues
- [ ] No new TODOs without tracking

---

## Next Steps

1. **Review questions** in `REFACTOR-LEARN-IM-QUIZME.md`
2. **Answer all questions** to make architectural decisions
3. **Update this task list** based on answers (remove optional tasks not selected)
4. **Prioritize phases** based on Q15 (urgency) and Q14 (migration approach)
5. **Begin implementation** starting with Phase 1

---

## Notes

- This refactoring should follow SpecKit methodology (constitution → spec → plan → tasks → implement)
- Track progress in `specs/001-cryptoutil/implement/DETAILED.md` Section 2
- Create conventional commits for each completed task
- Maintain evidence-based completion (tests pass, coverage verified, linting clean)
- Consider creating a feature branch for this work
