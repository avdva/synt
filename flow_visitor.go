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
	for _, st := range stmt.List {
		switch typed := st.(type) {
		case *ast.IfStmt:
			if len(fn.branches) > 0 || len(fn.statements) > 0 {
				result = append(result, fn)
			}
			fn = flowNode{}
			if typed.Init != nil {
				fn.statements = append(fn.statements, typed.Init)
			}
			ifFlow := flow{
				{statements: []ast.Stmt{&ast.ExprStmt{X: typed.Cond}}},
			}
			fn.branches = []flow{
				ifFlow,
			}
		case *ast.SwitchStmt, *ast.TypeSwitchStmt, *ast.GoStmt, *ast.SelectStmt, *ast.RangeStmt:
		case *ast.BlockStmt:
			if len(fn.branches) > 0 || len(fn.statements) > 0 {
				result = append(result, fn)
			}
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
