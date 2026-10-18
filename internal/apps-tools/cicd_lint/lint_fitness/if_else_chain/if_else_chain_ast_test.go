// Copyright (c) 2025-2026 Justin Cranford.
package if_else_chain

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
)

func TestAssignmentHelpers(t *testing.T) {
	t.Parallel()

	require.False(t, stmtAssigns(nil, "value"))

	assignment := &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent("other"), ast.NewIdent("value")},
		Tok: token.ASSIGN,
	}
	require.True(t, stmtAssigns(assignment, "value"))
	require.False(t, stmtAssigns(assignment, "missing"))
	require.False(t, stmtAssigns(&ast.ExprStmt{X: ast.NewIdent("value")}, "value"))
	require.False(t, stmtAssigns(&ast.ReturnStmt{}, "value"))

	require.False(t, blockAssigns(nil, "value"))
	require.False(t, blockAssigns(&ast.BlockStmt{}, ""))
	require.False(t, blockAssigns(&ast.BlockStmt{List: []ast.Stmt{&ast.ExprStmt{X: ast.NewIdent("value")}}}, "value"))
	require.True(t, blockAssigns(&ast.BlockStmt{List: []ast.Stmt{assignment}}, "value"))

	require.Equal(t, "filters", rootIdent("filters.Name"))
	require.Equal(t, "value", rootIdent("value"))
	require.Equal(t, "err", extractRootVar("os.IsNotExist:err"))
	require.Equal(t, "filters", extractRootVar("filters.Name"))
}

func TestExitStatementHelpers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		stmt ast.Stmt
		want bool
	}{
		{name: "return", stmt: &ast.ReturnStmt{}, want: true},
		{name: "branch", stmt: &ast.BranchStmt{Tok: token.BREAK}, want: true},
		{name: "panic call", stmt: &ast.ExprStmt{X: &ast.CallExpr{Fun: ast.NewIdent("panic")}}, want: true},
		{name: "other call", stmt: &ast.ExprStmt{X: &ast.CallExpr{Fun: ast.NewIdent("log")}}},
		{name: "selector call", stmt: &ast.ExprStmt{X: &ast.CallExpr{Fun: &ast.SelectorExpr{X: ast.NewIdent("fmt"), Sel: ast.NewIdent("Print")}}}},
		{name: "non-call expression", stmt: &ast.ExprStmt{X: ast.NewIdent("value")}},
		{name: "other statement", stmt: &ast.DeclStmt{}},
		{name: "nested conditional", stmt: &ast.IfStmt{Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ReturnStmt{}}}, Else: &ast.BlockStmt{List: []ast.Stmt{&ast.ReturnStmt{}}}}, want: true},
		{name: "nested conditional without else", stmt: &ast.IfStmt{Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ReturnStmt{}}}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, isExitStmt(tt.stmt))
		})
	}

	require.False(t, bodyEndsWithExit(nil))
	require.False(t, bodyEndsWithExit(&ast.BlockStmt{}))
	require.True(t, bodyEndsWithExit(&ast.BlockStmt{List: []ast.Stmt{&ast.ReturnStmt{}}}))
	require.False(t, ifStmtAlwaysExits(&ast.IfStmt{}))
}

func TestCollectIdent_LogicalExpressionsAndNestedSelectors(t *testing.T) {
	t.Parallel()

	for _, op := range []token.Token{token.LAND, token.LOR} {
		expr := &ast.BinaryExpr{X: ast.NewIdent("left"), Op: op, Y: ast.NewIdent("right")}
		require.Empty(t, collectIdent(expr, ""))
	}

	nestedSelector := &ast.SelectorExpr{
		X:   &ast.SelectorExpr{X: ast.NewIdent("object"), Sel: ast.NewIdent("field")},
		Sel: ast.NewIdent("nested"),
	}
	require.Equal(t, "object.field.nested", collectIdent(nestedSelector, ""))

	nestedIndex := &ast.SelectorExpr{
		X: &ast.IndexExpr{
			X:     ast.NewIdent("settings"),
			Index: &ast.BasicLit{Kind: token.STRING, Value: `"up"`},
		},
		Sel: ast.NewIdent("nested"),
	}
	require.Equal(t, "settings.up.nested", collectIdent(nestedIndex, ""))
}

func TestCheckInDir_WrapsWalkFailure(t *testing.T) {
	t.Parallel()

	root := filepath.Join(t.TempDir(), "missing")
	err := CheckInDir(cryptoutilCmdCicdCommon.NewQuietLogger("if-else-chain-walk-failure"), root)
	require.ErrorContains(t, err, "failed to walk directory")
}

func TestCheck_Workspace(t *testing.T) {
	t.Parallel()

	require.NoError(t, Check(cryptoutilCmdCicdCommon.NewQuietLogger("if-else-chain-workspace")))
}
