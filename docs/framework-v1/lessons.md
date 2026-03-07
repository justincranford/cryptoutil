# Framework V1 - Lessons Learned

This file captures lessons from each phase, used as:
1. Memory for the entire plan.md / tasks.md execution
2. Input for Phase 8 knowledge propagation to ARCHITECTURE.md, agents, skills, instructions

---

## Phase 6: Cross-Service Contract Test Suite

### What Worked Well

**Contract test design**:
- `RunContractTests(t *testing.T, server ServiceServer)` entry point is elegant - one call tests all contracts
- Grouping into `RunHealthContracts`, `RunServerContracts`, `RunResponseFormatContracts` is clean
- Three separate files (health_contracts.go, server_contracts.go, response_contracts.go) keeps each focused
- Real HTTPS connections (not mocks) validate the actual running service

**Test integration**:
- Adding `TestServiceName_ContractCompliance` to existing integration test files is minimal friction
- All 3 non-integration services pass in < 1s each

### What Didn't Go Well

**HTTP keep-alive connections caused 90-second shutdown hang**:
- Root cause: fasthttp keeps `open` counter > 0 when keep-alive connections are open
- `ShutdownWithContext` loops until `open == 0`, which never happens with persistent connections
- Fix: `DisableKeepAlives: true` on `http.Transport` in all test HTTP clients
- **Lesson**: ALL test HTTP clients that call real servers MUST use `DisableKeepAlives: true`

**TestMain shutdown timeout double-multiplication bug**:
- `DefaultDataServerShutdownTimeout` is already `time.Duration` (= `5 * time.Second`)
- Several TestMain files multiplied it by `time.Second` again, creating ~158-year timeout
- Fix: Use `DefaultDataServerShutdownTimeout` directly without `* time.Second`
- **Lesson**: Magic constants that are already `time.Duration` MUST NOT be multiplied by `time.Second`

**Task naming drift**:
- Tasks 6.3 and 6.4 were planned as "Auth" and "Error Format" contracts
- Actually implemented as "Server Isolation" and "Response Format" contracts
- Auth contracts need auth middleware to be configured - not portable across all services
- **Lesson**: Auth contracts belong in service-specific tests, not cross-service contracts

**Design deviation - acceptable**:
- Original design did not account for keep-alive / shutdown hang
- Production requirement: `DisableKeepAlives: true` is idiomatic for test HTTP clients
- This is a good pattern to document in instructions

### Patterns to Propagate

1. **Contract test pattern**: `RunContractTests(t, server)` for cross-service behavioral consistency
2. **Test HTTP transport**: ALWAYS `DisableKeepAlives: true` in test HTTP transports
3. **Duration constant usage**: Magic constants of type `time.Duration` MUST NOT be multiplied by `time.Second`
4. **Integration-tagged contract tests**: sm-kms pattern with `//go:build integration` for PostgreSQL-dependent services

---

## Phase 5: Shared Test Infrastructure

### What Worked Well

- `MustStartAndWaitForDualPorts` pattern is clean - startup + port polling + panic-on-failure
- `SetupTestServer` helper encapsulates all heavyweight setup
- `TestServerResources` struct with `Shutdown()` method cleans up nicely
- Reusing shared `testDB`, `testSmIMServer` across tests eliminates per-test setup overhead

### What Didn't Go Well

- `SetReady(true)` must be explicitly called AFTER server starts - not automatic
- Not documented clearly enough in the helper API

### Patterns to Propagate

1. **`SetReady(true)` requirement**: Must always be called after `MustStartAndWaitForDualPorts` unless server manages it internally (e.g., via `SetupTestServer` helper)
2. **TestMain pattern**: One-time setup via TestMain, SharedResources struct, `defer resources.Shutdown()`

---

## Phases 1-4: Foundation

*(Will be filled in during Phase 7/8 retrospective)*
