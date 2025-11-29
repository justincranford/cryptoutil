# Grooming Session 3: Implementation Details

## Purpose

Session 3 dives into implementation specifics based on your Session 2 decisions:
- Self-documenting UI requirements
- Identity embedding mechanics
- Key hierarchy implementation
- Demo scenario walkthrough

**Instructions**: Mark selections with `[X]`. Add notes where helpful.

---

## Section 1: Self-Documenting UI (Q1-6)

### Q1. UI Framework

What UI framework/approach for the web interface?

- [x] React (most employers use)
- [ ] Vue.js (simpler learning curve)
- [ ] Svelte (modern, minimal)
- [ ] HTMX + Go templates (backend-focused)
- [ ] No preference - whatever ships fastest
- [ ] Start with Swagger UI only (defer custom UI)

**Notes**:

---

### Q2. Self-Documenting UI Elements

What makes a UI "self-documenting" for you? (Select all that apply)

- [x] Tooltips on every field
- [x] Inline help text below inputs
- [ ] Contextual sidebars with explanations
- [ ] Progressive disclosure (show more as needed)
- [x] Example values as placeholders
- [ ] Guided tours/wizards for first-time users
- [x] Just good labels and logical flow

**Notes**:

---

### Q3. Error Message Philosophy

How should errors appear in a self-documenting UI?

- [x] Inline validation with fix suggestions
- [x] Toast notifications with details
- [ ] Error pages with troubleshooting steps
- [ ] Modal dialogs explaining what went wrong
- [ ] All errors should prevent submission (no retry)
- [ ] Errors should auto-recover when possible

**Notes**:

---

### Q4. First-Time User Flow

When someone opens the UI for the first time, what should happen?

- [ ] Landing page with big "Get Started" button
- [ ] Immediate redirect to login
- [ ] Tutorial overlay showing key features
- [ ] Demo mode with pre-populated data visible
- [x] Choose your path (Admin vs User vs Developer)
- [ ] Just show the main dashboard

**Notes**:

---

### Q5. UI Navigation Structure

How should the UI be organized?

- [x] Sidebar navigation (Gmail style)
- [ ] Top navigation bar (GitHub style)
- [ ] Tab-based interface
- [ ] Single-page with sections (scroll)
- [ ] Command palette (Ctrl+K) focus
- [ ] No preference

**Notes**:

---

### Q6. Mobile Responsiveness

How important is mobile/tablet support?

- [ ] Critical - must work on mobile
- [x] Nice to have - basic responsive
- [ ] Desktop only is fine
- [ ] Start desktop, add responsive later

**Notes**:

---

## Section 2: Demo Scenario (Q7-11)

### Q7. Demo Starting Point

User runs `docker compose up`. What appears first?

- [x] Terminal output with "Open http://localhost:8080"
- [ ] Browser auto-opens to welcome page
- [ ] Terminal shows service status table
- [ ] Swagger UI links printed
- [ ] All of the above

**Notes**:

---

### Q8. Demo User Accounts

What pre-seeded accounts for demo?

- [x] Admin user (full access)
- [x] Regular user (limited access)
- [x] Service account (API only)
- [x] All three above
- [ ] Single demo user with configurable roles
- [ ] No accounts - create during demo

**Notes**:

---

### Q9. Demo Data - Keys

What pre-seeded keys for KMS demo?

- [ ] One key of each supported type
- [ ] Full three-tier hierarchy example
- [ ] Root key only (create others in demo)
- [x] Multiple hierarchies (different use cases)
- [ ] Empty - show key creation in demo

**Notes**:

---

### Q10. Demo Walk-Through Steps

Put these demo steps in your preferred order (number them 1-6):

- [1] Login with demo credentials
- [3] View existing keys in hierarchy
- [4] Generate a new content key
- [5] Use key for an operation (sign/encrypt)
- [6] View audit log of operations
- [2] Logout

**Order**:

**Notes**:

---

### Q11. Demo Reset

How should demo data be resettable?

- [x] `docker compose down -v && docker compose up` (volumes)
- [x] Reset button in admin UI
- [ ] CLI command `cryptoutil demo reset`
- [ ] Auto-reset on container restart
- [ ] Don't need reset - demo is stateless

**Notes**:

---

## Section 3: Identity Embedding (Q12-16)

### Q12. Embedding Use Case

When would someone embed Identity vs run standalone?

- [ ] Small deployments (single binary preferred)
- [ ] When auth is only needed for one service
- [ ] Development/testing scenarios
- [ ] All of the above
- [x] Unsure - need to think about this more

**Notes**:

---

### Q13. Embedded Identity Configuration

How should embedded Identity be configured?

- [ ] Code-based (Go struct initialization)
- [ ] Same config file as host app
- [ ] Separate identity config section
- [ ] Minimal defaults, optional overrides

**Notes**:

---

### Q14. Embedded vs Standalone Feature Parity

Should embedded Identity have all features of standalone?

- [x] Yes - 100% feature parity
- [ ] No - embedded is simplified subset
- [ ] Core auth only, no admin UI when embedded
- [ ] Configurable - enable/disable features

**Notes**:

---

### Q15. Embedded Identity Authentication

When Identity is embedded in KMS, how does KMS authenticate?

- [ ] Trust implicitly (same process)
- [ ] Internal tokens (machine-to-machine)
- [x] Same auth flow as external clients
- [ ] Configurable trust level

**Notes**:

---

### Q16. Identity Go API

What's the ideal Go API for embedding Identity?

- [x] Single function: `identity.New(config)` returns server
- [ ] Builder pattern: `identity.NewServer().WithStorage(...).Build()`
- [ ] Interface-based: implement `identity.Provider`
- [ ] Just export the Fiber app for mounting
- [ ] No preference

**Notes**:

---

## Section 4: Key Hierarchy Details (Q17-21)

### Q17. Key Derivation Strategy

How should the three-tier hierarchy work?

- [x] Each tier is independent keys (stored separately)
- [ ] Child keys derived from parent (HKDF/similar)
- [x] Parent keys encrypt child key material
- [ ] Combination (derived AND encrypted)
- [ ] Unsure - what's the best practice?

**Notes**:
Current implementation

---

### Q18. Root Key Protection

How should root keys be protected?

- [x] Encrypted with unseal secrets (current cryptoutil approach)
- [x] HSM integration (future)
- [ ] Password-derived key encryption
- [ ] Multiple custodian split (Shamir-like)
- [ ] All of the above as options

**Notes**:

---

### Q19. Key Rotation Scope

When rotating keys, what happens?

- [x] Only specific key rotates, children unchanged
- [ ] Key rotates, children re-encrypted with new key
- [ ] Full hierarchy rotation (cascade)
- [ ] Configurable per key type
- [x] Rotation not in MVP

**Notes**:

---

### Q20. Key Access Control

How granular should key access control be?

- [ ] Per-key permissions (read/use/admin)
- [ ] Per-tier permissions (root users, content users)
- [ ] Role-based (Admin, Operator, Auditor)
- [x] All of the above
- [ ] Simple: authenticated = full access (MVP)

**Notes**:

---

### Q21. Key Hierarchy UI Representation

How should the three-tier hierarchy appear in UI?

- [x] Tree view (expandable nodes)
- [ ] Nested cards (visual hierarchy)
- [ ] Flat list with tier badges
- [x] Diagram view (visual connections)
- [ ] Table with parent column

**Notes**:

---

## Section 5: Scope System Design (Q22-25)

### Q22. Scope Syntax

What format for scope strings?

- [ ] Colon-separated: `kms:keys:read`
- [ ] Slash-separated: `kms/keys/read`
- [ ] Dot-separated: `kms.keys.read`
- [ ] URN-style: `urn:cryptoutil:kms:keys:read`
- [x] OAuth2 standard: `read:keys`
- [ ] No preference

**Notes**:

---

### Q23. Scope Inheritance

Should scopes be hierarchical?

- [ ] Yes: `kms:keys` implies `kms:keys:read`, `kms:keys:write`
- [ ] No: each scope is explicit
- [ ] Wildcard support: `kms:keys:*`
- [x] Both hierarchical and wildcard

**Notes**:

---

### Q24. Default Scopes

What scopes should tokens have by default?

- [x] None (explicit grant only)
- [ ] Read-only access to own resources
- [ ] Configurable per client
- [ ] Profile-based defaults

**Notes**:

---

### Q25. Resource-Specific Scopes

Should scopes reference specific resources?

- [x] Yes: `kms:keys:uuid-1234:read`
- [ ] No: scopes are type-level only
- [ ] Separate mechanism for resource permissions
- [x] Both scope-level and resource-level

**Notes**:

---

## Summary Section

### Top 3 Insights from This Session

After answering, list the 3 most important realizations:

1.
2.
3.

### Decisions That Need External Research

List any answers where you want to see industry best practices:

1.
2.
3.

### Ready for Implementation?

Which areas feel ready to implement now?

- [ ] UI framework choice
- [ ] Demo scenario
- [ ] Identity embedding API
- [ ] Key hierarchy design
- [ ] Scope system
- [ ] None yet - need more refinement

---

**Status**: AWAITING ANSWERS
**Next Step**: Complete answers, then request implementation plan or Session 4
