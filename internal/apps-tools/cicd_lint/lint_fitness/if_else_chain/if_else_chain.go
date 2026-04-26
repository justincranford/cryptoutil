// Copyright (c) 2025 Justin Cranford

// Package if_else_chain detects consecutive if statements in Go source where the first
// block does NOT end with an exit statement (return, panic, continue, break, goto) AND
// both conditions reference the same primary variable/expression. Such pairs should be
// written as else if chains or switch statements for clarity and correctness.
//
// Exempt patterns (not flagged):
//   - The first if has an else clause (already chained).
//   - The first block ends with return, panic, break, continue, or goto.
//   - The two if conditions reference different primary identifiers (independent filter accumulation).
//   - Files marked "Code generated" (oapi-codegen, protoc, etc.) are skipped.
//   - Test files (_test.go) are skipped.
package if_else_chain

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Violation holds details about a detected consecutive-if violation.
type Violation struct {
	File    string
	Line    int
	Message string
}

// Check runs the if_else_chain linter from the current working directory.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir scans all non-generated Go source files under rootDir for
// consecutive if statements that should be else if chains.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking for consecutive if-statement chains that should be else if...")

	violations, err := FindViolationsInDir(rootDir)
	if err != nil {
		return fmt.Errorf("if-else-chain: failed to walk directory: %w", err)
	}

	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Fprintf(os.Stderr, "%s:%d: %s\n", v.File, v.Line, v.Message)
		}

		return fmt.Errorf("found %d consecutive-if violations: use else if or switch instead", len(violations))
	}

	logger.LogWithPrefix("if-else-chain", "✅ No consecutive-if violations found")

	return nil
}

// FindViolationsInDir returns all consecutive-if violations found under rootDir.
func FindViolationsInDir(rootDir string) ([]Violation, error) {
	var violations []Violation

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			switch d.Name() {
			case cryptoutilSharedMagic.CICDExcludeDirVendor, cryptoutilSharedMagic.CICDExcludeDirGit:
				return filepath.SkipDir
			}

			return nil
		}

		// Only check non-test Go source files.
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		fileViolations, err := CheckFile(path)
		if err != nil {
			return err
		}

		violations = append(violations, fileViolations...)

		return nil
	})
	if err != nil {
		return violations, fmt.Errorf("walking directory %s: %w", rootDir, err)
	}

	return violations, nil
}

// CheckFile parses a single Go source file and returns consecutive-if violations.
func CheckFile(filePath string) ([]Violation, error) {
	fset := token.NewFileSet()

	src, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", filePath, err)
	}

	// Skip generated files (e.g. oapi-codegen output).
	if isGeneratedFile(src) {
		return nil, nil
	}

	f, err := parser.ParseFile(fset, filePath, src, 0)
	if err != nil {
		// Skip files that don't parse (e.g. build-tag-excluded files in non-matching builds).
		return nil, nil //nolint:nilerr // intentional: skip unparseable files silently
	}

	var violations []Violation

	ast.Inspect(f, func(n ast.Node) bool {
		// Walk block statements (function bodies, if bodies, etc.).
		block, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}

		vs := checkBlock(fset, filePath, block)
		violations = append(violations, vs...)

		return true
	})

	return violations, nil
}

// checkBlock examines a list of statements for consecutive if pairs where
// the first if block does not end with an exit statement AND both conditions
// reference the same primary variable (indicating mutually exclusive intent).
//
// A pair is NOT flagged when:
//   - The first if has an else clause (already chained).
//   - The first block ends with return, panic, break, continue, or goto.
//   - The two conditions reference different primary identifiers.
//   - Between the two if statements, the primary variable is reassigned via := or =
//     (indicating the variable is an independent loop/scope variable, not a shared one).
func checkBlock(fset *token.FileSet, filePath string, block *ast.BlockStmt) []Violation {
	stmts := block.List

	var violations []Violation

	for i := 0; i+1 < len(stmts); i++ {
		first, ok1 := stmts[i].(*ast.IfStmt)
		second, ok2 := stmts[i+1].(*ast.IfStmt)

		if !ok1 || !ok2 {
			continue
		}

		// The first if has an else clause — already chained, not a violation.
		if first.Else != nil {
			continue
		}

		// Check whether the body of the first if ends with an exit statement.
		if bodyEndsWithExit(first.Body) {
			continue
		}

		// Only flag when both conditions reference the same primary identifier.
		// This avoids false positives on independent filter-accumulation patterns.
		firstIdent := primaryIdent(first.Cond)
		secondIdent := primaryIdent(second.Cond)

		if firstIdent == "" || secondIdent == "" || firstIdent != secondIdent {
			continue
		}

		// Skip when the Init clause of the second if reassigns the variable.
		// Pattern: if val, ok := ...; ok { ... } — ok is redeclared, so independent.
		// extractRootVar strips compound-key prefixes (e.g. "os.IsNotExist:err" → "err").
		if second.Init != nil && stmtAssigns(second.Init, extractRootVar(firstIdent)) {
			continue
		}

		// Skip when the first if body assigns to the primary variable.
		// Pattern: if len(args) == 0 { args = normalize(args) } then if IsHelp(args) { ... }
		// The first block mutates the variable — the second check needs to run after mutation,
		// so else if would be semantically incorrect.
		if blockAssigns(first.Body, extractRootVar(firstIdent)) {
			continue
		}

		pos := fset.Position(first.Pos())
		violations = append(violations, Violation{
			File:    filePath,
			Line:    pos.Line,
			Message: fmt.Sprintf("consecutive if statements on %q: first block does not exit; consider using else if or switch", firstIdent),
		})
	}

	return violations
}

// stmtAssigns returns true when stmt assigns to the named variable via := or =.
// Used to detect when a variable is redeclared between two consecutive if statements,
// making the second if independent of the first.
func stmtAssigns(stmt ast.Stmt, name string) bool {
	if stmt == nil {
		return false
	}

	switch s := stmt.(type) {
	case *ast.AssignStmt:
		for _, lhs := range s.Lhs {
			if ident, ok := lhs.(*ast.Ident); ok && ident.Name == name {
				return true
			}
		}
	case *ast.ExprStmt:
		return false
	}

	return false
}

// blockAssigns returns true when any statement in body assigns to the named variable.
// Used to detect when the first if block mutates the primary variable, making a
// subsequent if sequential (not alternative), so else if would be semantically wrong.
func blockAssigns(body *ast.BlockStmt, name string) bool {
	if body == nil || name == "" {
		return false
	}

	for _, stmt := range body.List {
		if stmtAssigns(stmt, name) {
			return true
		}
	}

	return false
}

// rootIdent returns the root (leftmost) identifier from a dotted path string.
// For "filters.A", returns "filters". For "x", returns "x".
// This is used to check whether a simple variable name is assigned in a block,
// even when the primary identifier was computed from a selector expression.
func rootIdent(ident string) string {
	for i, c := range ident {
		if c == '.' {
			return ident[:i]
		}
	}

	return ident
}

// extractRootVar extracts the assignable root variable from a potentially compound key.
// Compound keys produced by the CallExpr case have the form "funIdent:argIdent"
// (e.g. "os.IsNotExist:err"). When checking whether an init clause reassigns the
// tracked variable, the relevant name is the argument part ("err"), not the full key.
// For non-compound keys (plain idents or dotted selectors), this is equivalent to
// rootIdent.
func extractRootVar(key string) string {
	// Strip function-prefix from compound keys like "os.IsNotExist:err".
	if idx := strings.LastIndex(key, ":"); idx >= 0 {
		key = key[idx+1:]
	}

	return rootIdent(key)
}

// primaryIdent returns the canonical string key for the leftmost expression in a condition.
// For simple identifiers (x > 0, err != nil), it returns the name (e.g. "x", "err").
// For selector expressions (filters.A != nil), it returns the full dotted path (e.g. "filters.A").
// For call expressions (errors.Is(err, ...)), it returns the first argument's key.
// Returns "" when no identifier can be extracted or when the pattern is ambiguous.
//
// The goal is to match conditions that test the SAME variable, so that
//
//	if x > 0 { doA() }
//	if x < 10 { doB() }   // violation: same "x"
//
// is flagged, while
//
//	if filters.A != nil { db = db.Where(...) }
//	if filters.B != nil { db = db.Where(...) }  // ok: different fields
//
// is not.
func primaryIdent(expr ast.Expr) string {
	return collectIdent(expr, "")
}

// collectIdent recursively builds the canonical key for an expression.
func collectIdent(expr ast.Expr, suffix string) string {
	if expr == nil {
		return ""
	}

	switch e := expr.(type) {
	case *ast.Ident:
		if suffix != "" {
			return e.Name + "." + suffix
		}

		return e.Name
	case *ast.BinaryExpr:
		// For logical operators (&& and ||), both sides are independent sub-conditions.
		// We cannot extract a single primary variable without misidentifying compound
		// conditions (e.g. "s.DevMode && A" and "s.DevMode && B" are independent).
		// Return "" to skip matching — these pairs are not consecutive-if violations.
		if e.Op.String() == "&&" || e.Op.String() == "||" {
			return ""
		}

		return collectIdent(e.X, suffix)
	case *ast.UnaryExpr:
		return collectIdent(e.X, suffix)
	case *ast.ParenExpr:
		return collectIdent(e.X, suffix)
	case *ast.SelectorExpr:
		// Build the full dotted path: X.Sel
		sel := e.Sel.Name
		if suffix != "" {
			sel = sel + "." + suffix
		}

		return collectIdent(e.X, sel)
	case *ast.CallExpr:
		// Build a compound key that includes the full dotted function name and the first
		// argument's ident, joined by ":". This lets calls with the same argument but
		// different functions (e.g. patternA.MatchString(x) vs patternB.MatchString(x))
		// produce different keys, while calls with the same function AND same argument
		// (e.g. errors.Is(err, ErrA) vs errors.Is(err, ErrB)) remain the same key.
		var funIdent string

		switch f := e.Fun.(type) {
		case *ast.SelectorExpr:
			funIdent = collectIdent(f, "")
		case *ast.Ident:
			funIdent = f.Name
		}

		if len(e.Args) > 0 {
			argIdent := collectIdent(e.Args[0], "")
			if argIdent != "" {
				if funIdent != "" {
					return funIdent + ":" + argIdent
				}

				return argIdent
			}
		}

		return funIdent
	case *ast.IndexExpr:
		// Include string-literal index so dirs["up"] and dirs["down"] get distinct keys,
		// matching how struct field selectors (filters.A vs filters.B) are treated.
		if lit, ok := e.Index.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			key := strings.Trim(lit.Value, `"`)
			if suffix != "" {
				key = key + "." + suffix
			}

			return collectIdent(e.X, key)
		}

		return collectIdent(e.X, suffix)
	default:
		return ""
	}
}

// bodyEndsWithExit returns true when the last statement in body is an exit statement:
// return, panic call, break, continue, or goto.
func bodyEndsWithExit(body *ast.BlockStmt) bool {
	if body == nil || len(body.List) == 0 {
		return false
	}

	last := body.List[len(body.List)-1]

	return isExitStmt(last)
}

// isExitStmt returns true for statements that unconditionally transfer control
// out of the current block: return, break, continue, goto, and panic(...) calls.
func isExitStmt(stmt ast.Stmt) bool {
	switch s := stmt.(type) {
	case *ast.ReturnStmt:
		return true
	case *ast.BranchStmt:
		// break, continue, goto, fallthrough — all exit the current if-block.
		return true
	case *ast.ExprStmt:
		// Detect panic(...) calls.
		call, ok := s.X.(*ast.CallExpr)
		if !ok {
			return false
		}

		ident, ok := call.Fun.(*ast.Ident)

		return ok && ident.Name == "panic"
	case *ast.IfStmt:
		// A trailing if/else chain exits only if both branches exit.
		return ifStmtAlwaysExits(s)
	default:
		return false
	}
}

// ifStmtAlwaysExits returns true when an if/else chain is guaranteed to exit on every path.
func ifStmtAlwaysExits(s *ast.IfStmt) bool {
	if !bodyEndsWithExit(s.Body) {
		return false
	}

	switch e := s.Else.(type) {
	case *ast.BlockStmt:
		return len(e.List) > 0 && isExitStmt(e.List[len(e.List)-1])
	case *ast.IfStmt:
		return ifStmtAlwaysExits(e)
	default:
		// No else clause — cannot guarantee exit on all paths.
		return false
	}
}

// isGeneratedFile returns true when the first 512 bytes of src contain a standard
// "Code generated" marker (as used by oapi-codegen, protoc, etc.).
func isGeneratedFile(src []byte) bool {
	const checkBytes = 512

	if len(src) > checkBytes {
		src = src[:checkBytes]
	}

	return strings.Contains(string(src), "Code generated")
}
