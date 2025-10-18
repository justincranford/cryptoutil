---
description: "Instructions for Pull Request description generation"
applyTo: "**"
---
# Pull Request Description Generation Instructions

## PR Description Structure

When generating Pull Request descriptions, follow this comprehensive format:

### Title
- Use conventional commit format: `type(scope): description`
- Keep under 72 characters
- Start with appropriate type: feat, fix, docs, style, refactor, perf, test, build, ci, chore

### Description Sections

#### 📋 **What**
- Clear, concise description of what the PR does
- Focus on the problem being solved, not the solution details
- Use present tense: "Adds feature X" not "Added feature X"

#### 🎯 **Why**
- Explain the business/technical rationale
- Reference issues, requirements, or user stories
- Describe impact on users, performance, or system behavior

#### 🔧 **How**
- High-level implementation approach
- Key technical decisions and trade-offs
- Reference to architecture changes if applicable

#### ✅ **Testing**
- How the changes were tested
- Test coverage impact
- Manual testing steps if applicable
- Performance benchmarks if relevant

#### 🔍 **Breaking Changes**
- List any breaking changes with migration guidance
- API changes, configuration updates, or behavioral changes
- Deprecation notices for phased removals

#### 📚 **Documentation**
- Documentation updates included
- README changes, API docs, or user guides
- Migration guides or release notes

### Code Review Checklist

#### 🔒 **Security**
- [ ] No sensitive data exposure
- [ ] Proper input validation
- [ ] Authentication/authorization checks
- [ ] Secure defaults applied

#### 🧪 **Quality**
- [ ] Tests added/updated
- [ ] Code follows project conventions
- [ ] Linting passes
- [ ] Documentation updated

#### 🚀 **Performance**
- [ ] No performance regressions
- [ ] Memory leaks addressed
- [ ] Database query optimization
- [ ] Caching strategy appropriate

#### 🔧 **Operations**
- [ ] Logging appropriate
- [ ] Monitoring/metrics added
- [ ] Configuration documented
- [ ] Deployment considerations addressed

### Related Links
- Closes #ISSUE_NUMBER
- Related to #ISSUE_NUMBER
- See also: DOC_LINK

### Screenshots/Examples
Include screenshots, before/after comparisons, or examples for UI changes, API responses, or complex logic.

---

## PR Size Guidelines

### 🟢 **Small (< 200 lines)**
- Single focused change
- Low risk, easy review
- Can be reviewed in one sitting

### 🟡 **Medium (200-500 lines)**
- Multiple related changes
- Moderate risk, needs careful review
- May require multiple review sessions

### 🔴 **Large (500+ lines)**
- Complex feature or refactor
- High risk, extensive review needed
- Consider splitting into smaller PRs
- Require thorough testing and documentation

### 📦 **Epic/Multiple Features**
- Break down into smaller, focused PRs
- Each PR should be independently deployable
- Clear dependencies between PRs documented</content>
<parameter name="filePath">c:\Dev\Projects\cryptoutil\.github\instructions\pull-requests.instructions.md
