# Refactor Learn Product and learn-im Service - Questions

This document contains questions that need to be answered before implementing the refactoring tasks outlined in REFACTOR-LEARN-IM.md.

## Pattern Selection

### Q1: Which command-line pattern should be adopted for the learn product?

**Context**: CMD-PATTERN.md describes three patterns (Suite, Product, Product-Service), with Product-Service recommended.

**A.** Suite Pattern - Single `cryptoutil` executable with `cryptoutil learn im server` command structure
**B.** Product Pattern - Separate `learn` executable with `learn im server` command structure
**C.** Product-Service Pattern (Recommended) - Separate `learn-im` executable with `learn-im server` command structure
**D.** Hybrid approach combining multiple patterns
**E.** Other: _________________________________

**Follow-up**: Should the learn product maintain compatibility with all three patterns for flexibility? ALL 3 PATTERNS MUST BE IMPLEMENTED.

---

## Directory Structure

### Q2: Should the internal command structure be created even if no product-level commands exist?

**Context**: Currently `internal/cmd/learn/` is empty. Other products have this structure.

**A.** Yes - Create `internal/cmd/learn/learn.go` with `im(args)` function for consistency
**B.** No - Skip since learn product only has one service (learn-im)
**C.** Wait - Create only when a second learn service is added
**D.** Depends on chosen pattern (required for Suite/Product, optional for Product-Service)
**E.** Other: _________________________________

A, ALL PRODUCTS MUST FOLLOW THE INTERNAL/CMD/PRODUCT/PRODUCT.GO STRUCTURE, EVEN IF THEY ONLY HAVE ONE SERVICE.
REASON IS CONSISTENCY, AND TO SIMPLIFY FUTURE ADDITIONS OF SERVICES UNDER THE PRODUCT.

---

### Q3: Should config files be created with the hierarchical structure?

**Context**: Pattern shows `configs/cryptoutil/config.yml`, `configs/learn/config.yml`, `configs/learn/im/config.yml`.

**A.** Yes - Create all three levels for consistency with other products
**B.** No - Use existing single config file approach
**C.** Partially - Create `configs/learn/im/config.yml` only
**D.** Merge - Combine learn and learn-im configs since only one service exists
**E.** Other: _________________________________

A YES

---

## Subcommand Support

### Q4: Which subcommands should be implemented for learn-im?

**Context**: Pattern shows: server, client, init, health, livez, readyz, shutdown.

**A.** All subcommands (server, client, init, health, livez, readyz, shutdown)
**B.** Core only (server, client, init)
**C.** Server-focused (server, health, livez, readyz, shutdown)
**D.** Minimal (server only - current state)
**E.** Other: _________________________________

A YES ALL

**Follow-up questions per subcommand**:

- **client**: What client operations should be supported? Message send/receive? User management? YES ALL OF THESE
- **init**: What initialization is needed? Database schema? Default users? Config generation? SEE KMS FOR REFERENCE, SQL/GORM MIGRATIONS, POSTGRES AND SQLITE, NO CGO
- **health**: Should this be a CLI wrapper for the admin HTTP endpoint? NO, PUBLIC API, NOT ADMIN, THIS NEEDS TO BE IMPLEMENTED IN SERVICE TEMPLATE, AND LEARN-IM NEEDS TO USE IT
- **livez**: Should this check if the process is running? CALL ADMIN LIVEZ API, IT IS A NO-OP, ALWAYS RETURNS 200, LEVERAGE IT FROM SERVICE TEMPLATE
- **readyz**: Should this verify database connectivity and dependencies? CALL ADMIN READYZ API, YES, LEVERAGE IT FROM SERVICE TEMPLATE
- **shutdown**: Should this trigger graceful shutdown via admin API or signal? CALL ADMIN SHUTDOWN API, LEVERAGE IT FROM SERVICE TEMPLATE

---

### Q5: Should the subcommands follow the admin API pattern or be independent?

**Context**: Admin API already provides `/admin/v1/livez`, `/admin/v1/readyz`, `/admin/v1/shutdown`.

**A.** CLI wrappers - Subcommands call the admin API endpoints (HTTP requests)
**B.** Independent - Subcommands implement logic directly (no HTTP)
**C.** Hybrid - Health checks via API, other commands independent
**D.** Duplicate - Both HTTP endpoints and CLI implementation
**E.** Other: _________________________________

A

---

## Implementation Details

### Q6: Should learn-im support both PostgreSQL and SQLite like the template?

**Context**: Current implementation uses SQLite in-memory. Service template supports both.

**A.** Yes - Add PostgreSQL support immediately
**B.** No - Keep SQLite-only (simpler for educational service)
**C.** Future - Plan for it but don't implement yet
**D.** Configurable - Support both, default to SQLite
**E.** Other: _________________________________

A

---

### Q7: What should the `server` subcommand do compared to current behavior?

**Context**: Current `learn-im` binary directly starts the server. Pattern suggests `learn-im server` structure.

**A.** No change - `learn-im` alone starts server (backward compatible)
**B.** Breaking change - Require `learn-im server` explicitly
**C.** Both - Default to server if no subcommand, but support explicit `server` subcommand
**D.** Deprecation path - Support both now, deprecate direct start later
**E.** Other: _________________________________

A, THERE ARE NO BREAKING CHANGES BECAUSE THIS PRODUCT HAS NEVER BEEN RELEASED YET

---

### Q8: Should the refactoring preserve backward compatibility?

**Context**: Changing command structure may break existing deployments, scripts, and documentation.

**A.** Yes - All existing commands must continue to work
**B.** No - Breaking changes acceptable for educational service
**C.** Partial - Support old commands with deprecation warnings
**D.** Version-based - Old binary stays as-is, new version has new structure
**E.** Other: _________________________________

E, AVOID BREAKING OTHER PRODUCTS!!! LEAVE THEM ALONE!

---

## Testing Strategy

### Q9: What level of testing is required for the refactored command structure?

**A.** Comprehensive - Unit tests for all subcommands, integration tests, E2E tests
**B.** Moderate - Unit tests for command parsing, basic integration tests
**C.** Minimal - Ensure server starts and basic functionality works
**D.** Existing - Rely on current test suite, no new tests for CLI structure
**E.** Other: _________________________________

A

---

### Q10: Should E2E tests be updated to use new command structure?

**Context**: Current E2E tests may use old binary invocation patterns.

**A.** Yes - Update all E2E tests immediately
**B.** No - Keep E2E tests using current structure
**C.** Both - Support both old and new command structures in tests
**D.** Gradual - Update E2E tests incrementally
**E.** Other: _________________________________

C

---

## Documentation Updates

### Q11: Which documentation needs updating?

**A.** All - README, API.md, Docker Compose files, deployment scripts, runbooks
**B.** Core - README and Docker Compose only
**C.** User-facing - README and deployment docs only
**D.** None - Keep docs describing current behavior
**E.** Other: _________________________________

D

---

## Integration with Orchestration Tools

### Q12: Should cryptoutil-compose, cryptoutil-demo, and cryptoutil-e2e be updated?

**Context**: These top-level orchestration tools are mentioned in CMD-PATTERN.md.

**A.** Yes - Update all three tools to support learn product
**B.** Partial - Update compose and demo only
**C.** Future - Plan for it but don't implement yet
**D.** No - Not needed for educational service
**E.** Other: _________________________________

A

---

### Q13: What Docker Compose patterns should be followed?

**A.** Existing - Keep current docker-compose.yml structure
**B.** Standardized - Move to `deployments/compose/learn/compose.yml` like other products
**C.** Both - Support both locations for compatibility
**D.** Simplified - Single compose file without product hierarchy
**E.** Other: _________________________________

B

---

## Migration Path

### Q14: Should there be a phased migration approach?

**A.** Big bang - Refactor everything at once
**B.** Phased - Step 1: Internal structure, Step 2: CLI, Step 3: Orchestration
**C.** Incremental - One subcommand at a time
**D.** Parallel - New structure alongside old until fully tested
**E.** Other: _________________________________

A

---

### Q15: What is the priority/urgency for this refactoring?

**A.** High - Blocks other work, do immediately
**B.** Medium - Important for consistency, do soon
**C.** Low - Nice to have, do when convenient
**D.** Optional - Educational service doesn't need full structure
**E.** Other: _________________________________

A

---

## Additional Considerations

### Q16: Should the learn product be renamed or repositioned?

**Context**: "learn" suggests educational/tutorial purpose. "im" is instant messaging.

**A.** Keep names - learn and learn-im are clear enough
**B.** Rename product - Use more descriptive name (e.g., "tutorial", "examples")
**C.** Rename service - Use more descriptive name (e.g., "messaging-demo")
**D.** Clarify purpose - Add documentation explaining educational nature
**E.** Other: _________________________________

A

---

### Q17: Should learn-im demonstrate additional service template features?

**Context**: Service template supports dual HTTPS, health checks, telemetry, etc.

**A.** Yes - Demonstrate all service template features
**B.** Selective - Pick subset most relevant for tutorials
**C.** Minimal - Keep simple for educational clarity
**D.** Advanced - Add features that production services don't have yet
**E.** Other: _________________________________

A

---

### Q18: How should configuration layering work?

**Context**: Pattern shows three config files: cryptoutil-wide, product-wide, service-specific.

**A.** Full layering - All three levels with proper merging/override logic
**B.** Two levels - Product and service only
**C.** Single level - Service config only
**D.** Environment-based - Different configs for dev/demo/e2e/prod
**E.** Other: _________________________________

A

---

## Notes

- All questions marked with "(Recommended)" align with CMD-PATTERN.md recommendations
- Questions should be answered before creating detailed implementation tasks
- Additional questions may emerge during implementation planning
