# Issue Categories Tracker - v3

## Purpose
Track categories of issues to identify patterns and systemic problems requiring broader fixes.

## Session: 2025-01-24 - Documentation Clarification and Enhancement

### Category 1: Documentation Clarity

**Frequency**: 1 issue  
**Severity**: Medium  
**Pattern**: Conflating orthogonal system concerns in technical documentation

**Root Causes**:
- Insufficient verification of system architecture assumptions
- Documentation written during implementation without architectural review
- Lack of peer review for technical accuracy

**Specific Issues**:
- Issue #1: "SQLite + 0.0.0.0" terminology conflates database choice with bind address validation

**Broader Impact**:
- May affect multiple documentation files (plan.md, tasks.md, ARCHITECTURE.md, copilot instructions)
- Risk of propagating misunderstandings to future implementers
- Creates confusion about system design decisions

**Systematic Fixes**:
1. Search all documentation for similar conflations
2. Add architecture review step before finalizing technical docs
3. Create glossary of orthogonal system concerns:
   - Database type (SQLite vs PostgreSQL) - deployment choice
   - Bind address (127.0.0.1 vs 0.0.0.0) - networking requirement
   - Dev mode (security restriction) - local development safety
   - Container mode (networking requirement) - Docker deployment
4. Update documentation templates to prompt for orthogonality verification

**Prevention Strategies**:
- ALWAYS diagram system architecture before writing documentation
- ALWAYS list orthogonal concerns explicitly
- ALWAYS verify assumptions with code archaeology
- ALWAYS cross-check documentation against implementation

---

### Category 2: Process Improvement

**Frequency**: 1 issue  
**Severity**: Low  
**Pattern**: Missing workflows for documentation lifecycle management

**Root Causes**:
- No documented process for temporary documentation cleanup
- No systematic lessons learned extraction
- No audit trail for deletion decisions

**Specific Issues**:
- Issue #2: Missing process for extracting lessons before deleting maintenance docs

**Broader Impact**:
- Risk of losing valuable insights
- Inconsistent documentation quality across sessions
- No way to audit what was deleted and why

**Systematic Fixes**:
1. Create lessons learned extraction checklist
2. Define permanent homes for different lesson types:
   - Architecture lessons → ARCHITECTURE.md
   - Workflow lessons → copilot instructions
   - Tool-specific lessons → relevant prompt files
3. Establish deletion approval workflow:
   - Extract lessons → Map to permanent docs → Verify coverage → Document decision → Delete
4. Create deletion log tracking what was removed and why

**Prevention Strategies**:
- ALWAYS create extraction checklist before deleting docs
- ALWAYS verify lessons are captured in permanent locations
- ALWAYS document deletion decision with justification
- ALWAYS maintain deletion log for audit trail

---

### Category 3: Tooling Enhancement

**Frequency**: 1 issue  
**Severity**: Medium  
**Pattern**: Prompt files lack systematic session tracking and analysis workflows

**Root Causes**:
- Prompts evolved organically without formal requirements
- No specification for session tracking location
- No defined workflow for post-completion analysis

**Specific Issues**:
- Issue #3: Prompt files don't specify how/where to track issues during implementation

**Broader Impact**:
- Inconsistent session documentation across different prompts
- Difficult to identify patterns across multiple sessions
- Manual ad-hoc decisions about planning phase
- No standard location for session artifacts

**Systematic Fixes**:
1. Enhance all 3 prompt files with session tracking requirements:
   - `.github/prompts/workflow-fixing.prompt.md`
   - `.github/prompts/beast-mode-3.1.prompt.md`
   - `.github/prompts/autonomous-execution.prompt.md`
2. Standardize on docs/fixes-needed-plan-tasks-v#/ location
3. Define issues.md and categories.md templates
4. Specify analysis → plan → tasks → QUIZME workflow
5. Define criteria for creating QUIZME.md vs proceeding directly

**Prevention Strategies**:
- ALWAYS specify session tracking requirements in prompts
- ALWAYS use consistent location for session artifacts
- ALWAYS create issues.md and categories.md during implementation
- ALWAYS perform post-completion analysis before starting next task

---

## Category Summary

**Total Categories**: 3  
**Affected Issues**: 3 (1 per category)  

**Cross-Cutting Themes**:
1. **Documentation Quality**: Need better verification and review processes
2. **Workflow Formalization**: Need documented processes for common tasks
3. **Systematic Tracking**: Need consistent patterns for issue tracking and analysis

**Recommended Actions**:
1. Create documentation review checklist (architecture assumptions, orthogonality)
2. Create lessons learned extraction workflow
3. Enhance all prompt files with session tracking requirements
4. Establish standard session artifacts location and templates

---

## Pattern Analysis

### Documentation Accuracy Pattern
**Problem**: Technical documentation written during implementation lacks architectural verification  
**Solution**: Add architecture review step + orthogonality verification checklist

### Temporary Documentation Lifecycle Pattern
**Problem**: No systematic process for extracting lessons before deletion  
**Solution**: Create extraction checklist + permanent home mapping + deletion log

### Session Tracking Pattern
**Problem**: Ad-hoc decisions about where/how to track issues during sessions  
**Solution**: Standardize location (docs/fixes-needed-plan-tasks-v#/) + template files + workflow specification

---

## Next Steps

1. ✅ Created issues.md and categories.md for this session
2. 🔄 Fix Issue #1 (terminology confusion in plan.md and tasks.md)
3. ⏳ Fix Issue #2 (extract lessons from maintenance docs, then delete)
4. ⏳ Fix Issue #3 (enhance 3 prompt files with session tracking)
5. ⏳ Create plan.md and tasks.md for this session's fixes
6. ⏳ Implement tasks from plan
