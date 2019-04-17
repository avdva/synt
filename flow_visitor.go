// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

import (
	"go/ast"
)

type flow []flowNode

type flowNode struct {
	statements []ast.Stmt
	branches   []flow
	defers     []ast.Stmt
}

func buildFlow(stmt *ast.BlockStmt) flow {
	var fn flowNode
	var result flow
	appendNode := func(statements ...ast.Stmt) {
		if len(fn.branches) > 0 || len(fn.statements) > 0 || len(fn.defers) > 0 {
			result = append(result, fn)
			fn = flowNode{}
		}
		if len(statements) > 0 {
			fn.statements = statements
		}
	}
	appendNodeOrStatements := func(statements ...ast.Stmt) {
		if len(fn.branches) == 0 && len(fn.defers) == 0 {
			fn.statements = append(fn.statements, statements...)
		} else {
			appendNode(statements...)
		}
	}
	for _, st := range stmt.List {
		switch typed := st.(type) {
		case *ast.IfStmt:
			appendNode()
			result = append(result, buildIfFlowNode(typed))
		case *ast.SwitchStmt, *ast.TypeSwitchStmt, *ast.GoStmt, *ast.SelectStmt, *ast.RangeStmt:
		case *ast.BlockStmt:
			appendNode()
			result = append(result, flowNode{branches: []flow{buildFlow(typed)}})
		case *ast.DeferStmt:
			appendNodeOrStatements(expressionsToStatemants(typed.Call.Args...)...)
			fn.defers = append(expressionsToStatemants(typed.Call), fn.defers...)
		default:
			appendNodeOrStatements(st)
		}
	}
	appendNode()
	return result
}

func buildIfFlowNode(stmt *ast.IfStmt) flowNode {
	var inits []ast.Stmt
	fn := flowNode{}
	for stmt != nil {
		inits = append(inits, stmt.Init)
		branchFlow := flow{
			{statements: expressionsToStatemants(stmt.Cond)},
		}
		bodyFlow := buildFlow(stmt.Body)
		branchFlow = append(branchFlow, bodyFlow...)
		fn.branches = append(fn.branches, branchFlow)
		switch typed := stmt.Else.(type) {
		case nil:
			stmt = nil
		case *ast.IfStmt:
			stmt = typed
		case *ast.BlockStmt:
			fn.branches = append(fn.branches, buildFlow(typed))
			stmt = nil
		case *ast.ExprStmt:
			fn.branches = append(fn.branches, flow{{statements: []ast.Stmt{typed}}})
			stmt = nil
		}
	}
	for i := range inits {
		fn.branches[i] = append(flow{{statements: inits[i:]}}, fn.branches[i]...)
	}
	return fn
}

func expressionsToStatemants(exprs ...ast.Expr) []ast.Stmt {
	var stmts []ast.Stmt
	for _, expr := range exprs {
		stmts = append(stmts, &ast.ExprStmt{X: expr})
	}
	return stmts
}
