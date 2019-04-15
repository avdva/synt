// Copyright 2017 Aleksandr Demakin. All rights reserved.

package synt

import (
	"go/ast"
)

type flow []flowNode

type flowNode struct {
	statements []ast.Stmt
	branches   []flow
}

func buildFlow(stmt *ast.BlockStmt) flow {
	var fn flowNode
	var result flow
	appendFn := func() {
		if len(fn.branches) > 0 || len(fn.statements) > 0 {
			result = append(result, fn)
			fn = flowNode{}
		}
	}
	for _, st := range stmt.List {
		switch typed := st.(type) {
		case *ast.IfStmt:
			appendFn()
			result = append(result, buildIfFlowNode(typed))
		case *ast.SwitchStmt, *ast.TypeSwitchStmt, *ast.GoStmt, *ast.SelectStmt, *ast.RangeStmt:
		case *ast.BlockStmt:
			appendFn()
			result = append(result, flowNode{branches: []flow{buildFlow(typed)}})
		case *ast.DeferStmt:
		default:
			if len(fn.branches) == 0 {
				fn.statements = append(fn.statements, st)
			} else {
				result = append(result, fn)
				fn = flowNode{statements: []ast.Stmt{st}}
			}
		}
	}
	if len(fn.branches) > 0 || len(fn.statements) > 0 {
		result = append(result, fn)
	}
	return result
}

func buildIfFlowNode(stmt *ast.IfStmt) flowNode {
	var inits []ast.Stmt
	fn := flowNode{}
	for stmt != nil {
		inits = append(inits, stmt.Init)
		branchFlow := flow{
			{statements: []ast.Stmt{&ast.ExprStmt{X: stmt.Cond}}},
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
